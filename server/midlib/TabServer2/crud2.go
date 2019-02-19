package TabServer2

//
// R E S T s e r v e r - Server Component	(TabServer2)
//
// Copyright (C) Philip Schlump, 2012-2017 -- All rights reserved.
//
// Do not remove the following lines - used in auto-update.
// Version: 1.1.0
// BuildNo: 0391
// FileId: 0005
// File: TabServer2/crud.go
//

// xyzzy-JWT

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/Go-FTL/server/tmplp"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/json" //	"encoding/json"
	"github.com/pschlump/ms"
	"github.com/pschlump/uuid"
)

// ====================================================================================================================================================================
// ====================================================================================================================================================================
func (hdlr *TabServer2Type) CommmonCrudPrefix(res http.ResponseWriter, req *http.Request, op string, ps *goftlmux.Params, rw *goftlmux.MidBuffer) (ok bool, mdata map[string]string, cfgTag string, h SQLOne, trx *tr.Trx, table_name string, wc *WhereCollect) {

	if db_trace_functions {
		fmt.Printf("In %s at %s\n", godebug.FUNCNAME(), godebug.LF())
	}

	// var rv string
	trx = mid.GetTrx(rw)

	wc = NewWhereCollect()
	ok = true
	res.Header().Set("Content-Type", "application/json")

	mdata = make(map[string]string, 80) // The posts that match

	// table_name = mux.Vars(req)["name"]
	table_name = ps.ByName("name")
	mdata["table_name"] = table_name

	cfgTag = "/api/table/" + table_name

	TableList := []string{table_name}
	TablesReferenced(godebug.FUNCNAME(2), cfgTag, TableList, hdlr)

	if db_common_crud1 {
		fmt.Printf("CommonCrudPrefix: %s:%s or %s, %s\n", cfgTag, req.Method, cfgTag, godebug.LF())
	}
	if h, ok = hdlr.SQLCfg[cfgTag+":"+req.Method]; !ok {
		h, ok = hdlr.SQLCfg[cfgTag] // get configuration						// Xyzzy - what if config not defined for this item at this point!!!!
	} else {
		// fmt.Printf ( "cfgTag(orig) = ->%s<- %[1]T\n", cfgTag )
		// fmt.Printf ( "Method = ->%s<- %[1]T\n", req.Method )
		cfgTag = cfgTag + ":" + req.Method
		trx.AddNote(1, fmt.Sprintf("Found a method specific key!, %s", cfgTag))
	}
	if !ok {
		// io.WriteString(res, fmt.Sprintf(`{"status":"error","code":"00025","msg":"Error(10102): Invalid table %s"}`, table_name))
		fmt.Printf("Invalid Query Parameters:Invalid Table, 406, %s\n", godebug.LF())
		ReturnErrorMessage(406, "Invalid Table", "00025",
			fmt.Sprintf(`Error(00025): Invalid Table (%s) sql-cfg.json[%s] %s`, table_name, cfgTag, godebug.LF()),
			res, req, *ps, trx, hdlr) // status:error
		ok = false
		return
	}

	hdlr.RemapParams(ps, h, trx)

	// Xyzzy851 - reset table name based on key data
	if len(h.TableName) > 0 {
		trx.AddNote(1, fmt.Sprintf("Reseting table name from %s in cfgTag to %s from data.", table_name, h.TableName))
		table_name = h.TableName
		mdata["table_name"] = table_name
	}

	trx.SetFrom(fmt.Sprintf("sql-cfg.json[%s]Line#:%s", cfgTag, h.LineNo))
	trx.SetFunc(2)
	trx.SetTablesUsed(TableList)
	trx.AddNote(2, "In common CRUD routine, called from...")
	trx.AddNote(1, "Table:"+table_name)

	if !hdlr.validOp(table_name, op, cfgTag) {
		fmt.Printf("Invalid Query Parameters:Invalid Operation, 406, %s\n", godebug.LF())
		ReturnErrorMessage(406, "Invalid Operation", "12043",
			fmt.Sprintf(`Error(12043): Invalid Operation (select) sql-cfg.json[%s] %s`, cfgTag, godebug.LF()),
			res, req, *ps, trx, hdlr) // status:error
		ok = false
		return
	}

	// m, fr = sizlib.UriToStringMap(req) // pull out the ?name=value params
	// tr.TraceUri ( req, m )

	err := ValidateQueryParams(ps, h, req) // Validate them
	if err != nil {
		fmt.Printf("Invalid Query Parameters:Invalid Parameter, 406, %s, %s\n", err, godebug.LF())
		ReturnErrorMessage(406, "Invalid Parameter", "12043",
			fmt.Sprintf(`Error(12043): Invalid Query Parameters (%s) sql-cfg.json[%s] %s`, sizlib.EscapeError(err), cfgTag, godebug.LF()),
			res, req, *ps, trx, hdlr) // status:error
		ok = false
		return
	}

	if db_common_crud1 {
		fmt.Printf("CommonCrudPrefix: Query Params are Valid, %s\n", godebug.LF())
	}
	trx.AddNote(1, "Query Params are Valid")

	if h.LoginRequired {
		trx.AddNote(1, "Query requires user validation.")
		// see if this is a valid user
		err = hdlr.ValidateUserTrx(h, trx, res, req, ps, rw)
		if err != nil {
			ReturnErrorMessage(401, "Authentication Failed, Invalid API Key", "1", fmt.Sprintf(`Error(18004): Invalid API Key %s`, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
			ok = false
			return
		}
	}

	if db_common_crud1 {
		fmt.Printf("CommonCrudPrefix: At Bottom, %s\n", godebug.LF())
	}

	trx.AddNote(1, "Passed user validation")
	SetDataPs(trx, ps)
	return
}

// ===================================================================================================================================================
// ===================================================================================================================================================
func UseTemplate(tag string, XTmpl string, dflt string, mdata map[string]string, trx *tr.Trx) (Query string) {
	var query_tmpl string
	if XTmpl != "" {
		query_tmpl = XTmpl
	} else {
		query_tmpl = dflt
	}
	Query = sizlib.Qt(query_tmpl, mdata)
	trx.AddNote(2, tag+" Template: "+query_tmpl)
	return
}

// ===================================================================================================================================================
// ===================================================================================================================================================
func (hdlr *TabServer2Type) HasPKInWhere(mdata map[string]string, h SQLOne, trx *tr.Trx, res http.ResponseWriter, req *http.Request, ps *goftlmux.Params) (hasPk bool) {
	hasPk = false
	var wc WhereClause
	sw := ps.ByName("where")
	if sw != "" {
		err := json.Unmarshal([]byte(sw), &wc)
		if err == nil {
			hasPk, _ = GenPKWhereFromWc(wc, trx, h, ps)
		}
	}
	return
}

// ====================================================================================================================================================================
// ====================================================================================================================================================================
func GenPKWhereFromWc(wc WhereClause, trx *tr.Trx, h SQLOne, ps *goftlmux.Params) (isPk bool, err error) {
	isPk = false
	err = nil
	x := make(map[string]bool)
	for _, v := range h.Cols {
		if v.IsPk {
			x[v.ColName] = false
		}
	}
	// fmt.Printf ( "At Top:%s\n", sizlib.SVarI(m) )
	if wc.Op == "and" {
		for _, vv := range wc.List {
			if sizlib.InArray(vv.Op, []string{"==", "="}) {
				ty, _ := ValidateColInWhere(vv.Name, h)
				x[vv.Name] = true
				switch ty {
				case "i":
					if ps.HasName(vv.Name) {
						goftlmux.AddValueToParams(vv.Name, fmt.Sprintf("%v", vv.Val1i), 'i', goftlmux.FromOther, ps)
					}
				case "f":
					if ps.HasName(vv.Name) {
						goftlmux.AddValueToParams(vv.Name, fmt.Sprintf("%v", vv.Val1f), 'i', goftlmux.FromOther, ps)
					}
				case "u": // UUID/GUID
					// odbc-xyzzy
					fallthrough
				case "":
					fallthrough
				case "s":
					if ps.HasName(vv.Name) {
						goftlmux.AddValueToParams(vv.Name, vv.Val1s, 'i', goftlmux.FromOther, ps)
					}
				case "d":
					if ps.HasName(vv.Name) {
						goftlmux.AddValueToParams(vv.Name, fmt.Sprintf("%v", vv.Val1d), 'i', goftlmux.FromOther, ps)
					}
				/* d, t, e */
				default:
					return false, errors.New(fmt.Sprintf("Error(10035): Invalid Type: %s for column %s", ty, vv.Name))
				}
			}
		}
	} else {
		return false, errors.New("Error(10041): Must have an 'and' or 'or' as the top level in the where clause")
	}
	for _, v := range x {
		if !v {
			return false, nil
		}
	}
	// fmt.Printf ( "At Bot:%s\n", sizlib.SVarI(m) )
	return true, nil
}

// ====================================================================================================================================================================
// http://localhost:8090/api/table/t_email_q?where={%22op%22:%22and%22,%22List%22:[{%22op%22:%22=%22,%22name%22:%22to%22,%22val1s%22:%22pschlump@gmail.com%22},{%22op%22:%22=%22,%22name%22:%22status%22,%22val1s%22:%22pending%22}]}
// http://localhost:8090/api/table/t_email_q?where={"op":"and","List":[{"op":"=","name":"to","val1s":"pschlump@gmail.com"},{"op":"=","name":"status","val1s":"pending"}]}
// see l2/rt.go
// ====================================================================================================================================================================
func (hdlr *TabServer2Type) ExtendedWhereCaluse(mdata map[string]string, h SQLOne, data *[]interface{}, trx *tr.Trx, wc *WhereCollect, res http.ResponseWriter, req *http.Request, ps *goftlmux.Params) (ok bool) {
	ok = true
	var rv string = ""
	var whereClause WhereClause
	sw := ps.ByName("where")
	if sw != "" {
		err := json.Unmarshal([]byte(sw), &whereClause)
		if err != nil {
			rv = fmt.Sprintf(`{ "status":"error", "msg":"Error(10032): Unable to construct where clause. Error with JSON parse. Raw Data:%s Err:%s", %s }`,
				sizlib.EscapeDoubleQuote(sw), sizlib.EscapeError(err), godebug.LFj())
			trx.SetQryDone(rv, "")
			ReturnErrorMessage(406, "Error(10032): Unable to construct where clause.  Error with JSON Parse", "10032",
				fmt.Sprintf(`Error(10032): Unable to construct where clause.  Error with JSON Parse (%s) %s`, sizlib.EscapeError(err), godebug.LFj()),
				res, req, *ps, trx, hdlr) // status:error
			ok = false
			return
		} else {
			if len(h.SetWhereAlias) > 0 && len(h.setWhereAlias) == 0 {
				h.setWhereAlias = `"` + h.SetWhereAlias + `".`
			}
			s, err := GenWhereFromWc(whereClause, trx, h, data, hdlr)
			if err != nil {
				rv = fmt.Sprintf(`{ "status":"error", "msg":"Error(10034): Invalid Where Clause; %s", %s }`, sizlib.EscapeError(err), godebug.LFj())
				trx.SetQryDone(rv, "")
				ReturnErrorMessage(406, "Error(10034): Invalid where clause.", "10034",
					fmt.Sprintf(`Error(10034): Invalid where clause. (%s) %s`, sizlib.EscapeError(err), godebug.LFj()),
					res, req, *ps, trx, hdlr) // status:error
				ok = false
				return
			}
			wc.AddClause(s)
		}
	}
	return
}

// ==============================================================================================================================================================================
// Collect the chunks of the were and "and" them together.
// ==============================================================================================================================================================================

type WhereCollect struct {
	AClause []string
}

func NewWhereCollect() *WhereCollect {
	return &WhereCollect{}
}

func (this *WhereCollect) AddClause(s string) {
	if s != "" {
		this.AClause = append(this.AClause, s)
	}
}

func (this *WhereCollect) GenWhereClause(mdata map[string]string) {
	mdata["where_where"] = ""
	mdata["where_where_or_and"] = "where"
	mdata["where"] = ""
	if len(this.AClause) > 0 {
		mdata["where_and"] = "and"
		mdata["where_where"] = "where"
		mdata["where_where_or_and"] = "and"
		s := ""
		com := ""
		for i, v := range this.AClause {
			if db_where_collect {
				fmt.Printf("GenWhereClause: %d: ->%s<-\n", i, v)
			}
			s += com + " " + v
			com = "and"
		}
		mdata["where"] = s
	}
}

// ==============================================================================================================================================================================
// Return true if Name is a column name.
// ==============================================================================================================================================================================
func IsColumnName(h SQLOne, Name string) bool {
	for _, v := range h.Cols {
		if v.ColName == Name {
			if !v.NoSort {
				return true
			} else {
				return false
			}
		}
	}
	return false
}

// ==============================================================================================================================================================================
//
// Limits:
//		ColName can be 1..N of columns in h.Cols or the name of the column (case senstivie)
//		Dir is ASC or DESC - case in-sensitive.
//
// Example: ( " should be %22 )
//		http://localhost:8090/api/table/tblNotify?limit=5&orderBy=[{"ColName":"4","Dir":"desc"}]
//
// Enhancement: Allow for sort by function or type cast.
// Enhancement: Sort by any leading edge of a key defined in d.b.
//
// ==============================================================================================================================================================================
func (hdlr *TabServer2Type) GenOrderBy(mdata map[string]string, h SQLOne, trx *tr.Trx, res http.ResponseWriter, req *http.Request, ps *goftlmux.Params) bool {

	s := ""
	com := ""
	//if len(h.SetWhereAlias) > 0 && len(h.setWhereAlias) == 0 {
	//	h.setWhereAlias = `"`+h.SetWhereAlias+`".`
	//}

	if x, ok := ps.GetByName("orderBy"); ok {

		var ob []OrdSpec

		trx.AddNote(1, fmt.Sprintf("User supplied order by, %s", x))

		err := json.Unmarshal([]byte(x), &ob) // Convert the JSON column to values
		if err != nil {
			hdlr.CrudErrMsg(1, fmt.Sprintf("Error(12051): Invalid orderBy clause.  Clause=->%s<-  err=%v", x, err), err, "", trx, res, req, ps)
			return false
		}

		orderByFound := false

		for i, v := range ob {
			if sizlib.IsIntString(v.ColName) {
				n, err := strconv.Atoi(v.ColName)
				if err != nil {
					hdlr.CrudErrMsg(1, fmt.Sprintf("Error(12052): Invalid orderBy clause.  Invalid numeric conversion of [%s] at postion %d, err=%v", v.ColName, i, err), err, "", trx, res, req, ps)
					return false
				}
				if !(n >= 1 && n <= len(h.Cols)) {
					hdlr.CrudErrMsg(1, fmt.Sprintf("Error(12053): Invalid orderBy clause.  Invalid column positon of [%s], out of range. Should be between 1 and %d err=%v",
						v.ColName, len(h.Cols), err), err, "", trx, res, req, ps)
					return false
				} else {
					if v.Dir == "" {
					} else if len(v.Dir) > 0 && !(strings.ToLower(v.Dir) == "asc" || strings.ToLower(v.Dir) == "desc") {
						hdlr.CrudErrMsg(1, fmt.Sprintf("Error(12054): Invalid orderBy clause.  Invalid sort direction of [%s]", v.Dir), nil, "", trx, res, req, ps)
						return false
					}
					s += com + v.ColName + " " + v.Dir
					com = ","
					orderByFound = true
					trx.AddNote(1, fmt.Sprintf("User supplied order by, Pos:%d, [%s %s]", i, v.ColName, v.Dir))
				}
			} else {
				if v.Dir == "" {
				} else if len(v.Dir) > 0 && !(strings.ToLower(v.Dir) == "asc" || strings.ToLower(v.Dir) == "desc") {
					hdlr.CrudErrMsg(1, fmt.Sprintf("Error(12056): Invalid orderBy clause.  Invalid sort direction of [%s] err=%v", v.Dir, err), err, "", trx, res, req, ps)
					return false
				}
				if !IsColumnName(h, v.ColName) {
					hdlr.CrudErrMsg(1, fmt.Sprintf("Error(12055): Invalid orderBy clause.  Invalid column name of [%s] err=%v", v.ColName, err), err, "", trx, res, req, ps)
					return false
				}
				// xyzzy - convert v.ColName to #
				// s += com + h.setWhereAlias + `"` + v.ColName + `" ` + v.Dir
				s += com + `"` + v.ColName + `" ` + v.Dir
				com = ","
				orderByFound = true
				trx.AddNote(1, fmt.Sprintf("User supplied order by, Pos:%d, [%s %s]", i, v.ColName, v.Dir))
			}
		}

		if orderByFound {
			mdata["order_by"] = s
			mdata["order_by_order_by"] = " order by "
		}

	} else if len(h.OrderBy) > 0 {
		for _, v := range h.OrderBy {
			if v.Dir == "" {
			} else if len(v.Dir) > 0 && !(strings.ToLower(v.Dir) == "asc" || strings.ToLower(v.Dir) == "desc") {
				hdlr.CrudErrMsg(1, fmt.Sprintf("Error(12057): Invalid orderBy clause.  Invalid sort direction of [%s]", v.Dir), nil, "", trx, res, req, ps)
				return false
			}
			if sizlib.IsIntString(v.ColName) {
				s = s + com + v.ColName + v.Dir
			} else {
				// xyzzy - convert v.ColName to #
				// s = s + com + h.setWhereAlias + `"` + v.ColName + `" ` + v.Dir
				s = s + com + `"` + v.ColName + `" ` + v.Dir
			}
			com = ","
		}
		mdata["order_by"] = s
		trx.AddNote(1, fmt.Sprintf("Default order by, [%s]", s))
		mdata["order_by_order_by"] = " order by "
	}

	return true

}

// ==============================================================================================================================================================================
// Build the projected columns for selects.  This is the SELECT <columns> section.
// ==============================================================================================================================================================================
func (hdlr *TabServer2Type) GenProjectedCols(mdata map[string]string, h SQLOne, trx *tr.Trx, ps *goftlmux.Params) {

	if len(h.Cols) > 0 {
		s := ""
		com := ""
		for _, v := range h.Cols {
			if v.ColAlias != "" {
				switch v.ColType {
				case "u":
					if hdlr.GetDbType() == DbType_postgres {
						s = s + com + DbBeginQuote + v.ColName + DbEndQuote + " as " + DbBeginQuote + v.ColAlias + DbEndQuote // Odbc-xyzzy - Quote of col alias needs to be consistent
					} else if hdlr.GetDbType() == DbType_odbc {
						s = s + com + "convert(varchar(40)," + DbBeginQuote + v.ColName + DbEndQuote + ") as " + DbBeginQuote + v.ColAlias + DbEndQuote
					} else if hdlr.GetDbType() == DbType_Oracle {
						s = s + com + DbBeginQuote + v.ColName + DbEndQuote + " as " + DbBeginQuote + v.ColAlias + DbEndQuote
					} else {
						s = s + com + DbBeginQuote + v.ColName + DbEndQuote + " as " + DbBeginQuote + v.ColAlias + DbEndQuote
					}
				default:
					s = s + com + DbBeginQuote + v.ColName + DbEndQuote + " as " + DbBeginQuote + v.ColAlias + DbEndQuote
				}
			} else {
				switch v.ColType {
				case "u":
					if hdlr.GetDbType() == DbType_postgres {
						s = s + com + DbBeginQuote + v.ColName + DbEndQuote
					} else if hdlr.GetDbType() == DbType_odbc {
						s = s + com + "convert(varchar(40)," + DbBeginQuote + v.ColName + DbEndQuote + ") as [id] "
					} else if hdlr.GetDbType() == DbType_Oracle {
						s = s + com + DbBeginQuote + v.ColName + DbEndQuote
					} else {
						s = s + com + DbBeginQuote + v.ColName + DbEndQuote
					}
				default:
					s = s + com + DbBeginQuote + v.ColName + DbEndQuote
				}
			}
			com = ","
		}
		// fmt.Printf("Generating ->%s<-\n", s)
		mdata["cols"] = s
	}
}

// ==============================================================================================================================================================================
// Get the list of column names that form the primary key for the table in the h.Cols order.
// ==============================================================================================================================================================================
func (hdlr *TabServer2Type) GetPkName(h SQLOne, nPkCols int, trx *tr.Trx, res http.ResponseWriter, req *http.Request, ps *goftlmux.Params) ([]string, bool) {
	trx.AddNote(2, "In GetPkName")
	n_pk := 0
	pk := make([]string, nPkCols)
	if len(h.Cols) > 0 {
		for _, v := range h.Cols {
			if v.IsPk {
				n_pk++
				pk = append(pk, v.ColName)
			}
		}
		if n_pk != nPkCols {
			hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10063): Only %d column can be specified as a primary key when using this API. [sql-cfg.json]line_no=%s", nPkCols, h.LineNo), nil, "", trx, res, req, ps)
			return pk, false
		}
	} else {
		hdlr.CrudErrMsg(1, fmt.Sprintf("Error(10064): Collum names/types and a primary key must be specified, [sql-cfg.json]line_no=%s", h.LineNo), nil, "", trx, res, req, ps)
		return pk, false
	}

	return pk, true
}

// ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
func GenKey(h_Query string, typ string, tn string, data ...interface{}) string {
	// theKey := "data:cache:" + h_Query + sizlib.SVar(data)
	theKey := ""
	if typ == "row" {
		theKey = "data:row:" + data[0].(string)
	} else {
		theKey = "data:" + tn
	}
	return theKey
}

// ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
func (hdlr *TabServer2Type) HaveCachedData(res http.ResponseWriter, req *http.Request, h SQLOne, h_Query string, data ...interface{}) (rv string, found bool) {

	// rw, _ /*top_hdlr*/, _ /*psP*/, err := GetRwPs(res, req)

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	defer hdlr.gCfg.RedisPool.Put(conn)

	if !(h.CacheIt == "row" || h.CacheIt == "table") {
		return
	}
	theKey := GenKey(h_Query, h.CacheIt, h.TableName, data...)
	if db_cache {
		fmt.Printf("CacheIt:True, %s: fetch: key = -->>%s<<--\n", godebug.LF(), theKey)
	}
	// s, err := redis.String(rr.RedisDo("GET", theKey))
	s, err := conn.Cmd("GET", theKey).Str() // Get the value
	if db_cache {
		fmt.Printf("CacheIt:True, %s: redis says [%s] [%v]\n", godebug.LF(), s, err)
	}
	if err != nil {
		if db_cache {
			fmt.Printf("CacheIt:True, %s: redis error\n", godebug.LF())
		}
		return
	}
	if db_cache {
		fmt.Printf("CacheIt:True, %s: found it - returning it\n", godebug.LF())
	}
	// rr.RedisDo("EXPIRE", theKey, 60*5)
	conn.Cmd("EXPIRE", theKey, 60*5)
	return s, true
}

// ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
func (hdlr *TabServer2Type) CacheItForLater(res http.ResponseWriter, req *http.Request, h SQLOne, id string, theData string, h_Query string, data ...interface{}) {
	if h.CacheIt == "row" || h.CacheIt == "table" {

		// rw, _ /*top_hdlr*/, _ /*psP*/, err := GetRwPs(res, req)

		conn, err := hdlr.gCfg.RedisPool.Get()
		if err != nil {
			logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
			return
		}

		defer hdlr.gCfg.RedisPool.Put(conn)

		theKey := GenKey(h_Query, h.CacheIt, h.TableName, data...)
		if db_cache {
			fmt.Printf("CacheIt:True, %s: cache: 1st key = [%s] data=[%s] <<-\n", godebug.LF(), theKey, theData)
		}
		// x1, err := rr.RedisDo("SET", theKey, theData) // Missing Timeout -- Use Expire
		err = conn.Cmd("SET", theKey, theData).Err // Get the value
		// rr.RedisDo("EXPIRE", theKey, 60*5)
		if db_cache && err != nil {
			fmt.Printf("CacheIt:True, %s: Data Saved! redis says [%v]\n", godebug.LF(), err)
		}
		err = conn.Cmd("EXPIRE", theKey, 60*5).Err // Get the value
		if db_cache && err != nil {
			fmt.Printf("CacheIt:True, %s: Data Saved! redis says [%v]\n", godebug.LF(), err)
		}
	}
}

// ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
func (hdlr *TabServer2Type) HaveCachedDataMk(res http.ResponseWriter, req *http.Request, h SQLOne) (data string, found bool) {

	_ /*rw*/, _ /*top_hdlr*/, ps, err := GetRwPs(res, req)

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	defer hdlr.gCfg.RedisPool.Put(conn)

	if !(h.CacheIt == "row" || h.CacheIt == "table") {
		return
	}
	theKey := hdlr.GenKeyMk(h, ps)
	if db_cache {
		fmt.Printf("CacheItMk:True, %s: fetch: key = -->>%s<<--\n", godebug.LF(), theKey)
	}
	// s, err := redis.String(rr.RedisDo("GET", theKey))
	s, err := conn.Cmd("GET", theKey).Str() // Get the value
	if db_cache && err != nil {
		fmt.Printf("CacheItMk:True, %s: redis says [%s] [%v]\n", godebug.LF(), s, err)
	}
	if err != nil {
		if db_cache {
			fmt.Printf("CacheItMk:True, %s: redis error\n", godebug.LF())
		}
		return
	}
	if db_cache {
		fmt.Printf("CacheItMk:True, %s: found it - returning it\n", godebug.LF())
	}
	// rr.RedisDo("EXPIRE", theKey, 60*5)
	_ = conn.Cmd("EXPIRE", theKey, 60*5).Err // Get the value
	return s, true
}

// ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
func (hdlr *TabServer2Type) CacheItForLaterMk(res http.ResponseWriter, req *http.Request, h SQLOne, theData string) {
	if h.CacheIt == "row" || h.CacheIt == "table" {

		_ /*rw*/, _ /*top_hdlr*/, ps, err := GetRwPs(res, req)

		conn, err := hdlr.gCfg.RedisPool.Get()
		if err != nil {
			logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
			return
		}

		defer hdlr.gCfg.RedisPool.Put(conn)

		theKey := hdlr.GenKeyMk(h, ps)
		if db_cache {
			fmt.Printf("CacheItMk:True, %s: cache: 1st key = [%s]\n", godebug.LF(), theKey)
		}
		// x1, err := rr.RedisDo("SET", theKey, theData) // Missing Timeout -- Use Expire
		err = conn.Cmd("SET", theKey, theData).Err // Get the value
		if db_cache && err != nil {
			fmt.Printf("CacheItMk:True, %s: Data Saved! redis says [%v]\n", godebug.LF(), err)
		}
		// rr.RedisDo("EXPIRE", theKey, 60*5)
		err = conn.Cmd("EXPIRE", theKey, 60*5).Err // Get the value
		if db_cache && err != nil {
			fmt.Printf("CacheItMk:True, %s: Data Saved! redis says [%v]\n", godebug.LF(), err)
		}
	}
}

// ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
func (hdlr *TabServer2Type) GenKeyMk(h SQLOne, ps *goftlmux.Params) string {
	// theKey := "data:cache:" + h_Query + sizlib.SVar(data)
	pk_found := false
	rv := ""
	com := ""
	var ty string
	var err error
	for ii, v := range h.Cols {
		if v.IsPk {
			pk_found = true
			ty, _ = ValidateColInWhere(v.ColName, h)

			switch ty {

			case "u": // UUID/GUID		-- Redis Key
				fallthrough
			case "":
				fallthrough
			case "s":
				rv += com + ps.ByName(v.ColName)
				com = "/"

			case "i":
				var x int64
				x, err = strconv.ParseInt(ps.ByNameDflt(v.ColName, "0"), 10, 64)
				rv += com + fmt.Sprintf("%d", x)
				com = "/"

			case "b":
				var b bool
				b = sizlib.ParseBool(ps.ByNameDflt(v.ColName, "false"))
				rv += com + fmt.Sprintf("%v", b)
				com = "/"

			case "f":
				var f float64
				f, err = strconv.ParseFloat(ps.ByNameDflt(v.ColName, "0"), 64)
				rv += com + fmt.Sprintf("%f", f)
				com = "/"

			case "d":
				fallthrough
			case "t":
				fallthrough
			case "e":
				var d time.Time
				nullOk := ValidateNullOk(v.ColName, ii, h, "key")
				d, _, err = ms.FuzzyDateTimeParse(ps.ByName(v.ColName), nullOk)
				rv += com + fmt.Sprintf("%s", d)
				com = "/"
			}
		}
	}
	_ = err
	theKey := ""
	if pk_found {
		theKey = rv
	} else {
	}
	if h.CacheIt == "row" {
		theKey = "data:row:" + rv
	} else {
		theKey = "data:" + h.TableName
	}
	return theKey
}

// ==============================================================================================================================================================================
// Pull out a value from the URL using mux.Vars - save it in common store.
// ==============================================================================================================================================================================
func GetMuxValue(name string, as string, mdata map[string]string, trx *tr.Trx, res http.ResponseWriter, req *http.Request, ps goftlmux.Params) string {

	// id := mux.Vars(req)["id"]
	id := ps.ByName("id")
	mdata[as] = id
	// goftlmux.AddValueToParams("key", key_tmpl, 'i', goftlmux.FromOther, psP)
	return id
}

// ==============================================================================================================================================================================
//
// Generate %{insCols%} %{insDol%} %{updColsDol%} %{andPkWhere%}
// Values stored in 'mdata'
//
// Example:
//		insUpd2 ( "insert into \"img_file\" ( %{insCols%} ) values ( %{insDol%} )",
//			"update \"img_file\" set %{updColsDol%} where %{andPkWhere%}", colData, mergedData, mdata )
//
// test:  t-pp1.go
//
// ==============================================================================================================================================================================

var dolNoSeq int = 0

func getDolNo(cols []ColSpec, colName string) string {
	dn := getDolNoNumber(cols, colName)
	return fmt.Sprintf("%d", dn)
}
func getDolNoNumber(cols []ColSpec, colName string) int {
	ii := FindColNoInMeta(cols, colName)
	if cols[ii].DolNo >= 0 {
		return cols[ii].DolNo
	} else {
		dolNoSeq++
		cols[ii].DolNo = dolNoSeq
		return dolNoSeq
	}
}

var dolNoSeqUpd int = 0

func getDolNoUpd(cols []ColSpec, colName string) string {
	dn := getDolNoUpdNumber(cols, colName)
	return fmt.Sprintf("%d", dn)
}
func getDolNoUpdNumber(cols []ColSpec, colName string) int {
	ii := FindColNoInMeta(cols, colName)
	if cols[ii].DolNoUpd >= 0 {
		return cols[ii].DolNoUpd
	} else {
		dolNoSeqUpd++
		cols[ii].DolNoUpd = dolNoSeqUpd
		return dolNoSeqUpd
	}
}

func (hdlr *TabServer2Type) GenInsUpdInfo(table_key string, jData map[string]interface{}, mdata map[string]string) {

	key, ok := hdlr.SQLCfg[table_key]
	dolNoSeq = 0
	dolNoSeqUpd = 0
	for i := range key.Cols {
		key.Cols[i].DolNo = -1
		key.Cols[i].DolNoUpd = -1
	}

	if !ok {
		fmt.Printf("Invalid key, or missing in in sql-cfg.json data (%s)\n", table_key)
		os.Exit(1)
	}

	insCols := ""
	insDol := ""
	updColsDol := ""
	com := ""
	ucom := ""
	for _, v := range key.Cols {
		if _, ok := jData[v.ColName]; ok {
			if v.Insert {
				dolNo := getDolNo(key.Cols, v.ColName)
				insCols += com + `"` + v.ColName + `"`
				insDol += com + "$" + dolNo
				com = ", "
			}
		}
		if _, ok := jData[v.ColName]; ok {
			if !v.IsPk && v.Update {
				dolNo := getDolNoUpd(key.Cols, v.ColName)
				updColsDol += ucom + `"` + v.ColName + `" = ` + "$" + dolNo
				ucom = ", "
			}
		}
	}

	andPkWhere := ""
	com = ""
	for _, v := range key.Cols {
		if _, ok := jData[v.ColName]; ok {
			if v.IsPk {
				dolNo := getDolNoUpd(key.Cols, v.ColName)
				andPkWhere += com + `"` + v.ColName + `" = ` + "$" + dolNo
				com = " and "
			}
		}
	}

	mdata["insCols"] = insCols
	mdata["insDol"] = insDol
	mdata["updColsDol"] = updColsDol
	mdata["andPkWhere"] = andPkWhere

	// fmt.Printf ( "((%s)) insCols = [%s]\ninsDol = [%s]\nupdColsDol = [%s]\nandPkWhere = [%s]\n", table_key, insCols, insDol, updColsDol, andPkWhere )

}

// ==============================================================================================================================================================================
// ==============================================================================================================================================================================
// return an array of data for insert statment
// xyzzy -did they all get data??? -- input-Requried, update-Required
func GetDataInsert(cols []ColSpec, data map[string]interface{}) []interface{} {
	rv := make([]interface{}, len(cols))
	mx := 0
	// xyzzy - generated data
	// xyzzy - default data
	for _, v := range cols { // For all the columns specified
		if d, ok := data[v.ColName]; ok { // If we have data for that column
			dn := getDolNoNumber(cols, v.ColName)
			if v.Insert { // If this is an insert able column
				rv[dn-1] = d
				if mx < dn {
					mx = dn
				}
			}
		}
	}
	return rv[0:mx]
}

// ==============================================================================================================================================================================
// ==============================================================================================================================================================================
func GetDataUpdate(cols []ColSpec, data map[string]interface{}) []interface{} {
	rv := make([]interface{}, len(cols))
	// xyzzy - default data
	mx := 0
	// fmt.Printf ( "data=%s\n", sizlib.SVar(data) )
	// fmt.Printf ( "cols=%s\n", sizlib.SVar(cols) )
	for _, v := range cols {
		// fmt.Printf ( "v.ColName [%s] Update=%t IsPk=%t", v.ColName, v.Update, v.IsPk )
		if d, ok := data[v.ColName]; ok {
			// fmt.Printf ( " have in data=%t", ok )
			dn := getDolNoUpdNumber(cols, v.ColName)
			if v.Update || v.IsPk {
				// fmt.Printf ( " adding [%v] at %d ", d, dn )
				rv[dn-1] = d
				if mx < dn {
					mx = dn
				}
			}
			// fmt.Printf ( " mx = %d ", mx )
		}
		// fmt.Printf ( "\n" )
	}
	return rv[0:mx]
}

// ==============================================================================================================================================================================
// ==============================================================================================================================================================================
// -- depricated -- not used
func insUpd2(ins string, upd string, table_key string, data map[string]interface{}, mdata map[string]string, trx *tr.Trx, hdlr *TabServer2Type) {
	key, _ := hdlr.SQLCfg[table_key]
	var err error
	err = nil

	// xyzzy22 - need diff data for insert, update since different $1,$... values in query - 2 data params.

	// fmt.Printf ( "ins=>%s<= upd=>%s<= data %s mdata %s\n", ins, upd, sizlib.SVar(data), sizlib.SVar(mdata) )

	if len(ins) > 0 {
		ins_q := sizlib.Qt(ins, mdata)
		dataArr := GetDataInsert(key.Cols, data) // xyzzy - for ins, upd
		// fmt.Printf ( "     insUpd2(ins) %s, data = %v\n", ins_q, sizlib.SVar(dataArr) )
		err = sizlib.Run1Thx(hdlr.gCfg.Pg_client.Db, trx, ins_q, dataArr...)
	}
	if len(ins) == 0 || err != nil {
		// fmt.Printf ( "Error (1452) in insUpd2 = %s\n", err )
		err = nil
		dataArr := GetDataUpdate(key.Cols, data) // xyzzy - for ins, upd
		upd_q := sizlib.Qt(upd, mdata)
		// fmt.Printf ( "     insUpd2(upd) %s, data = %s\n", upd_q, sizlib.SVar(dataArr) )
		err = sizlib.Run1Thx(hdlr.gCfg.Pg_client.Db, trx, upd_q, dataArr...)
		if err != nil {
			fmt.Printf("Error (1458) in insUpd2 = %s\n", err)
		}
	}
}

// ==============================================================================================================================================================================
// test: t-pp1.go
// ==============================================================================================================================================================================
func FindColNoInMeta(cols []ColSpec, aColName string) int {
	// fmt.Printf ( "cols=%s\n", sizlib.SVar(cols) )
	if cols == nil {
		fmt.Printf("Error (16): Passed an empyt set of colls to FindColNoInMeta - probably an error. aColName=(%s), %s\n", aColName, godebug.LF())
	}
	for i := range cols {
		// fmt.Printf ( "i=%d ->%s<- == ->%s<-\n", i, cols[i].ColName, aColName );
		if cols[i].ColName == aColName {
			return i
		}
	}
	return -1
}

// ==============================================================================================================================================================================
//		theFile0 = pickInsertUpdate ( theFile0, "table:img_file" )												// xyzzy f(x)
// test: t-pp1.go
// ==============================================================================================================================================================================
// xyzzy -- depricated -- I don't think this function is called at all --
func (hdlr *TabServer2Type) PickInsertUpdateColumns(www http.ResponseWriter, theFile0 map[string]interface{}, table_key string) (rv map[string]interface{}) {
	rv = make(map[string]interface{}, len(theFile0))
	key, ok := hdlr.SQLCfg[table_key]
	if !ok {
		fmt.Printf("Error (29) !internal! table_key is invalid: %s\n", table_key)
		return
	}
	for i, v := range theFile0 {
		if hdlr.DbEnabledOn(www, "PickInsertUpdateColumns") {
			fmt.Printf("Pick: %s\n", i)
		}
		ii := FindColNoInMeta(key.Cols, i)
		if ii >= 0 {
			if hdlr.DbEnabledOn(www, "PickInsertUpdateColumns") {
				fmt.Printf("Found it in Cols at %d with ins=%v upd=%v\n", ii, key.Cols[ii].Insert, key.Cols[ii].Update)
			}
			if key.Cols[ii].Insert || key.Cols[ii].Update {
				rv[i] = v
			}
		}
	}
	return
}

// -----------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// Create the "closure" fucntion that will save passed data for later and return a
// function bound to the passed data.
// func Hello(w http.ResponseWriter, r *http.Request, ps goftlmux.Params) {
// -- depricated --
//func GetSqlCfgHandler(name string) func(res http.ResponseWriter, req *http.Request, ps goftlmux.Params) {
//	return func(res http.ResponseWriter, req *http.Request, ps goftlmux.Params) {
//		RespHandlerSQL(res, req, name, ps)
//	}
//}

// -- depricated --
//func GetRedisCfgHandler(name string) func(res http.ResponseWriter, req *http.Request, ps goftlmux.Params) {
//	return func(res http.ResponseWriter, req *http.Request, ps goftlmux.Params) {
//		RespHandlerRedis(res, req, name, ps)
//	}
//}

// ==============================================================================================================================================================================
// func InjectDataPs(ps *goftlmux.Params, h SQLOne, res http.ResponseWriter, req *http.Request) {
/*
	Issue00079: NoLog parameters
	Issue00018: Actual CC Processing with account + fake cards																	4 hr
			2. NoLog parameters
			3. Put in actual test params with account - see that it works
*/
// ==============================================================================================================================================================================
func NoLogData(data []interface{}, h SQLOne) []interface{} {
	if len(h.NoLog) == 0 {
		return data
	}
	fmt.Printf("NoLog Stuff, h.NoLog=%s, %s\n", sizlib.SVar(h.NoLog), godebug.LF())
	n_data := make([]interface{}, len(data))
	for i, w := range h.P {
		fmt.Printf("\ti=%d w=%v\n", i, w)
		if sizlib.InArray(w, h.NoLog) {
			n_data[i] = "*** not logged ***"
		} else {
			n_data[i] = w
		}
	}
	return n_data
}

// ==============================================================================================================================================================================
// ==============================================================================================================================================================================
func (hdlr *TabServer2Type) RespHandlerSQL(res http.ResponseWriter, req *http.Request, cfgTag string, ps *goftlmux.Params, rw *goftlmux.MidBuffer) {

	h := hdlr.SQLCfg[cfgTag] // get configuration						// Xyzzy - what if config not defined for this item at this point!!!!
	var rv string = ""
	var done bool = false
	var isError bool = false
	data := make([]interface{}, len(h.P))            // organize data for call to d.b.
	addNames := make([]string, 0, len(h.Popt)+1)     // organize data for call to d.b.
	addVals := make([]interface{}, 0, len(h.Popt)+1) // organize data for call to d.b.
	// xyzzyAdd1 -- addCols, addVals -- from validation "optional" values
	cookieList := make(map[string]string)

	if sizlib.InArray("dump_params", h.DebugFlag) {
		fmt.Printf("Debug flag 'dump_prams' is on, request=%s\n", sizlib.SVarI(req))
		fmt.Printf("Params are %s\n", ps.DumpParamDB())
	}

	// Setting Heders
	res.Header().Set("Content-Type", "application/json") // set default reutnr type

	// Get Initial Data
	// m, fr := sizlib.UriToStringMap(req) // pull out the ?name=value params				// xyzzy - add 'fr'
	// tr.TraceUri(req, m)

	trx := mid.GetTrx(rw)
	trx.SetFrom(fmt.Sprintf("sql-cfg.json[%s]Line#:%s", cfgTag, h.LineNo))
	trx.SetFunc(1)
	trx.SetTablesUsed(h.TableList)
	TraceUriPs(trx, req, ps) // xyzzyBoom11111 // trx.TraceUriPs(req, ps) // xyzzyBoom11111

	// InjectData(m, fr, h, res, req) // Inject Data
	InjectDataPs(ps, h, res, req) // xyzzy - add 'fr'

	// trx.SetData(m, fr)
	SetDataPs(trx, ps)

	TablesReferenced(godebug.FUNCNAME(), cfgTag, h.TableList, hdlr)

	// Validate Query Params
	trx.AddNote(1, "Validate Query Parameters")
	err := ValidateQueryParams(ps, h, req) // Validate them
	if err != nil {
		// trx.AddNote(1, "Validate Query Parameters Failed:"+fmt.Sprintf("%v", err))
		// rv = fmt.Sprintf(`{"status":"error","code":"3","msg":"Error(10005): Invalid Query Parameters (%s) sql-cfg.json[%s]",%s}`, err, cfgTag, godebug.LFj())
		done = true
		isError = true
		// trx.ErrorReturn ( 1, rv )
		ReturnErrorMessage(406, "Invalid Parameter", "12043",
			fmt.Sprintf(`Error(12043): Invalid Query Parameters (%s) sql-cfg.json[%s] %s`, sizlib.EscapeError(err), cfgTag, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
		rv = ""
	}

	// Authenticate User if Necessary
	if !done {
		//fmt.Printf ( "At %s\n", godebug.LF() )
		if h.LoginRequired {
			trx.AddNote(1, "User Authentication is Requried")
			err = hdlr.ValidateUserTrx(h, trx, res, req, ps, rw)
			if err != nil {
				// trx.AddNote(1, "User Authentication Failed:"+fmt.Sprintf("%v", err))
				// func ReturnErrorMessage(status int, msg string, code string, details string, res http.ResponseWriter, req *http.Request, ps goftlmux.Params) {
				// rv = fmt.Sprintf(`{"status":"error","code":"1","msg":"Error(10004): Invalid API Key.",%s}`, godebug.LFj())
				ReturnErrorMessage(401, "Authentication Failed, Invalid API Key", "1", fmt.Sprintf(`Error(10004): Invalid API Key %s`, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
				done = true
				isError = true
				rv = ""
			} else {
				trx.SetUserInfo(ps.ByName("username"), ps.ByName("$user_id$"), ps.ByName("auth_token"))
			}
		}
	}

	// CallBefore
	exit := false
	a_status := 200
	if !done {
		//fmt.Printf ( "At %s\n", godebug.LF() )
		if len(h.CallBefore) > 0 {
			trx.AddNote(1, "Functions to call before running queries.  CallBefore is set.")
			for _, fx_name := range h.CallBefore {
				if !exit {
					trx.AddNote(1, fmt.Sprintf("CallBefore[%s]", fx_name))
					rv, exit, a_status = hdlr.CallFunction("before", fx_name, res, req, cfgTag, rv, isError, cookieList, ps, trx)
				}
			}
		}
	}
	if exit {
		fmt.Printf("****************** exit from before operations has been signaled **********************, rv=%s, %s\n", rv, godebug.LF())
		done = true
		isError = true
		ReturnErrorMessageRv(a_status, rv, "Preprocessing signaled error", "18008",
			fmt.Sprintf(`Error(18008): Preprocessing signaled error. sql-cfg.json[%s] %s`, cfgTag, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
	}

	// xyzzyExtraDataQuery

	if !done {
		fmt.Printf("%sBefore At: %s\n%s%s", MiscLib.ColorGreen, godebug.LF(), (*ps).DumpParamTable(), MiscLib.ColorReset)
		for i := 0; i < len(h.P); i++ {
			aName := h.P[i]
			if !(*ps).HasName(aName) {
				fmt.Fprintf(os.Stderr, "%sWarning(10023): Missing data for %s - using empty string%s", MiscLib.ColorRed, aName, MiscLib.ColorReset)
				trx.AddNote(1, fmt.Sprintf("Warning(10023): Mising data for %s - using empty string", aName))
			}
			data[i] = ps.ByName(aName)
		}
		// xyzzyAdd1 -- addCols, addVals -- from validation "optional" values
		// for each optional param in validation, if haveByName - on command, and in h.Poptional? - then ...
		if h.Query != "" {
			// fmt.Printf("h.Popt=%s, %s\n", godebug.SVarI(h.Popt), godebug.LF())
			for i := 0; i < len(h.Popt); i++ {
				aName := h.Popt[i]
				// fmt.Printf("Popt[%d]=%s, %s\n", i, aName, godebug.LF())
				if ps.HasName(aName) {
					if h.Pname[i] != "" {
						addNames = append(addNames, h.Pname[i])
					} else {
						addNames = append(addNames, aName)
					}
					addVals = append(addVals, ps.ByNameDflt(aName, ""))
				}
			}
			// fmt.Printf("addNames=%s addVals=%s, %s\n", godebug.SVar(addNames), godebug.SVar(addVals), godebug.LF())
		}
	}

	if !done {
		if h.G != "" && h.Query != "" {
			// rv = fmt.Sprintf("{\"status\":\"error\",\"msg\":\"Error(10008): Invalid internal configuration.  Both G and Query set to values in sql-cfg.json[%s].\",%s}", cfgTag, godebug.LFj())
			done = true
			isError = true
			// trx.AddNote(1, fmt.Sprintf("Error(10008): Invalid configuration in sql-cfg.json[%s]Line#:%s, Both G and Query set to values.", cfgTag, h.LineNo))
			ReturnErrorMessage(409, "Configuration Conflict", "10008",
				fmt.Sprintf(`Error(10008): Interal Conflict, both G and Query set. sql-cfg.json[%s] %s`, cfgTag, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
			rv = ""
		}
	}

	if !done {
		if rv == "" {
			rv = fmt.Sprintf("{\"status\":\"success\",%s}", godebug.LFj())
		}
		if h.F != "" {
			fmt.Printf("At %s, F=%s\n", godebug.LF(), h.F)
			trx.AddNote(1, "Running .F Query")
			trx.SetQry(h.F, 1, data...)
			Rows, err := hdlr.gCfg.Pg_client.Db.Query(h.F, data...)
			if err != nil {
				rv = fmt.Sprintf(`{ "status":"error","msg":"Error(10003): Database Error. sql-cfg.json[%s].F",%s, "err":%q }`, cfgTag, godebug.LFj(), err)
				done = true
				isError = true
				trx.SetQryDone(rv, "")
				ReturnErrorMessage(400, "Database Error", "10003",
					fmt.Sprintf(`Error(10003): Database Error, sql-cfg.json[%s].F %s err:%q`, cfgTag, godebug.LFj(), err), res, req, *ps, trx, hdlr) // status:error
				rv = ""
			} else {
				defer Rows.Close()
				dt, _, _ := sizlib.RowsToInterface(Rows)
				trx.SetQryDone("", sizlib.SVar(dt))
			}
			if !done {
				if h.Query != "" {
				} else if h.G != "" {
				} else {
					done = true // rv = "{\"status\":\"success\"}"
				}
			}
		}
	}
	if !done {
		if rv == "" {
			rv = fmt.Sprintf("{\"status\":\"success\",%s}", godebug.LFj())
		}
		if h.Fx != "" {
			fmt.Printf("At xyzzy992 %s, Fx=%s\n", godebug.LF(), h.Fx)
			trx.AddNote(1, "Running .Fx procedure call")
			qry := ""
			plist := hdlr.GetStoredProcedurePlist(len(data))
			if hdlr.GetDbType() == DbType_postgres {
				/*
					Issue00079: NoLog parameters
					Issue00018: Actual CC Processing with account + fake cards																	4 hr
							2. NoLog parameters
							3. Put in actual test params with account - see that it works
				*/
				log_data := NoLogData(data, h)
				fmt.Printf("Postgres Call of %s%s data = %s, %s\n", h.Fx, plist, sizlib.SVar(log_data), godebug.LF())
				qry = fmt.Sprintf("select %s%s as \"x\"", h.Fx, plist)
			} else if hdlr.GetDbType() == DbType_odbc {
				fmt.Printf("MS-SQL Call of->exec %s%s<- len(data)=%d\n", h.Fx, plist, len(data))
				qry = fmt.Sprintf("exec %s%s", h.Fx, plist)
			} else if hdlr.GetDbType() == DbType_Oracle {
				fmt.Printf("MS-SQL Call of->call %s ( %s )<- len(data)=%d\n", h.Fx, plist, len(data))
				qry = fmt.Sprintf("call %s ( %s )", h.Fx, plist)
			} else {
				fmt.Printf("Oracle Call of %s\n", h.Fx)
				qry = fmt.Sprintf("execute %s%s", h.Fx, plist)
			}
			trx.AddNote(1, qry)
			trx.SetQry(qry, 1, data...)
			Rows, err := hdlr.gCfg.Pg_client.Db.Query(qry, data...)
			if err != nil {
				rv = fmt.Sprintf(`{ "status":"error","msg":"Error(10003): Database Error. sql-cfg.json[%s].F",%s, "err":%q }`, cfgTag, godebug.LFj(), err)
				done = true
				isError = true
				trx.SetQryDone(rv, "")
				ReturnErrorMessage(400, "Database Error", "10003",
					fmt.Sprintf(`Error(10003): Database Error, sql-cfg.json[%s].F %s err:%q`, cfgTag, godebug.LFj(), err), res, req, *ps, trx, hdlr) // status:error
				rv = ""
			} else {
				defer Rows.Close()
				dt, _, _ := sizlib.RowsToInterface(Rows)
				trx.SetQryDone("", sizlib.SVar(dt))
			}
			if !done {
				if h.Query != "" {
				} else if h.G != "" {
				} else {
					done = true // rv = "{\"status\":\"success\"}"
				}
			}
		}
	}
	if !done {
		if rv == "" {
			rv = fmt.Sprintf("{\"status\":\"success\",%s}", godebug.LFj())
		}
		//fmt.Printf ( "At %s\n", godebug.LF() )
		if h.G != "" {
			trx.AddNote(1, "Running .G Query")
			qry := ""
			// ,"ms-connectToDbType":"odbc"
			// func genTemplateExec(h SQLOne, op string, data *[]interface{}, *ps *goftlmux.Params) (rv string, err error) {
			plist := hdlr.GetStoredProcedurePlist(len(data))
			if hdlr.GetDbType() == DbType_postgres {
				/*
					Issue00079: NoLog parameters
					Issue00018: Actual CC Processing with account + fake cards																	4 hr
							2. NoLog parameters
							3. Put in actual test params with account - see that it works
				*/
				log_data := NoLogData(data, h)
				fmt.Printf("Postgres Call of %s%s data = %s, %s\n", h.G, plist, sizlib.SVar(log_data), godebug.LF())
				qry = fmt.Sprintf("select %s%s as \"x\"", h.G, plist)
			} else if hdlr.GetDbType() == DbType_odbc {
				fmt.Printf("MS-SQL Call of->exec %s%s<- len(data)=%d\n", h.G, plist, len(data))
				qry = fmt.Sprintf("exec %s%s", h.G, plist)
			} else if hdlr.GetDbType() == DbType_Oracle {
				fmt.Printf("MS-SQL Call of->call %s ( %s )<- len(data)=%d\n", h.G, plist, len(data))
				qry = fmt.Sprintf("call %s ( %s )", h.G, plist)
			} else {
				fmt.Printf("Oracle Call of %s\n", h.G)
				qry = fmt.Sprintf("execute %s%s", h.G, plist)
			}
			// Rows, err := sizlib.Sel ( res, req, db, qry, data... )
			trx.SetQry(qry, 1, data...)
			Rows, err := hdlr.gCfg.Pg_client.Db.Query(qry, data...)
			if err != nil {
				fmt.Printf("qry: %s data: %s, at:%s\n", qry, data, godebug.LF())
				rv = fmt.Sprintf(`{ "status":"error","msg":"Error(10002): Database Error. sql-cfg.json[%s].G",%s, "err":%q }`, cfgTag, godebug.LFj(), err)
				done = true
				isError = true
				trx.SetQryDone(rv, "")
				ReturnErrorMessage(400, "Database Error", "10002",
					fmt.Sprintf(`Error(10002): Database Error, sql-cfg.json[%s].G %s err:%q`, cfgTag, godebug.LFj(), err), res, req, *ps, trx, hdlr) // status:error
				rv = ""
			} else {
				defer Rows.Close()
			}
			if !done {
				data, _, n_row := sizlib.RowsToInterface(Rows)
				trx.SetQryDone("", sizlib.SVar(data))
				if n_row == 1 { // xyzzy - is a 0 row result allowed?
					// fmt.Printf("Above IF Before!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!, error if 'x' is NULL from d.b.\n")
					if len(data) <= 0 {
						rv = `{"status":"error","msg":"Stored procedure failed to return a row."}`
						trx.AddNote(1, fmt.Sprintf("No data returned from stored proceudre. Warning this is an error in the stored procedure. sql-cfg.json[%s] source-code %s",
							cfgTag, godebug.LF()))
					} else if tx, ok := data[0]["x"]; ok {
						// fmt.Printf("Just Before!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n")
						if tx == nil {
							rv = `{"status":"error","msg":"Stored procedure return a NULL value."}`
							// fmt.Printf("caught NULL return value,!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n")
						} else {
							rv = data[0]["x"].(string)
						}
						// fmt.Printf("Just after,!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n")
					} else {
						// fmt.Printf ( "x not defined, %s\n", godebug.LF() )
						if hdlr.GetDbType() == DbType_odbc {
							trx.AddNote(1, fmt.Sprintf("Column 'x' not defined, Warning this is an error in the stored procedure. sql-cfg.json[%s] source-code %s",
								cfgTag, godebug.LF()))
						}
						// xyzzy - oracle behavior??
						rv = Get1stRowFromMap(data[0]).(string)
					}
					if len(h.SetCookie) > 0 {
						teb, err := sizlib.JSONStringToData(rv)
						// fmt.Printf("teb = %s - at top\n", sizlib.SVar(teb))
						if err != nil {
							//fmt.Printf ( "Bad Data: %s\n", rv )
							// rv = fmt.Sprintf(`{ "status":"error","msg":"Error(10001): Parsing return value failed. sql-cfg.json[%s].G",%s, "err":%q }`, cfgTag, godebug.LFj(), err)
							fmt.Printf("rv=%s at:%s\n", rv, godebug.LF())
							done = true
							isError = true
							ReturnErrorMessage(500, "Database Error", "10001",
								fmt.Sprintf(`Error(10001): Parsing return value failed for SetCookie, sql-cfg.json[%s].G %s err:%q`, cfgTag, godebug.LFj(), err), res, req, *ps, trx, hdlr) // status:error
							rv = ""
						}
						if !done {
							for vv, do_save := range h.SetCookie { // SetCookie at this point.
								expire := time.Now().AddDate(0, 0, 2) // Years, Months, Days==2 // xyzzy - should be a config - on how long to keep cookie
								// cookie := http.Cookie{vv, value, "/", "www.sliceone.com", expire, expire.Format(time.UnixDate), 86400, true, true, "test=tcookie", []string{"test=tcookie"}}
								cookieValue, ok := teb[vv]
								if !ok {
									cookieValue = ""
								}
								// host := strings.Split(req.Host,":")[0]
								// cookie := http.Cookie{Name:vv, Value:cookieValue.(string), Path:"/", Domain:host, Expires:expire, RawExpires:expire.Format(time.UnixDate)
								//		, MaxAge:86400, Secure:false, HttpOnly:false}
								// xyzzy - secure shall be set to true for HTTPS
								secure := false
								if req.TLS != nil {
									secure = true
								}
								cookie := http.Cookie{Name: vv, Value: cookieValue.(string), Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400, Secure: secure, HttpOnly: false}
								//fmt.Printf ( "cookie header is: %s\n", cookie.String() )
								cookieList[vv] = cookieValue.(string)
								http.SetCookie(res, &cookie)
								trx.AddNote(1, fmt.Sprintf("Setting Cookie Name=%s Value=%s", vv, cookieValue.(string)))
								if ok && !do_save {
									delete(teb, vv)
								}
							}
							// fmt.Printf("teb = %s - at bot\n", sizlib.SVar(teb))
							rv = sizlib.SVar(teb)
						}
					} else if len(h.SetSession) > 0 {
						/*
						   l_data = '{"status":"success","$send_email$":{'
						   		||'"template":"please_confirm_registration"'
						   		||',"username":'||to_json(p_username)
						   		||',"real_name":'||to_json(p_real_name)
						   		||',"email_token":'||to_json(l_email_token)
						   		||',"app":'||to_json(p_app)
						   		||',"name":'||to_json(p_name)
						   		||',"url":'||to_json(p_url)
						   		||',"from":'||to_json(l_from)
						   	||'},"$session$":{'
						   		||'"set":['
						   			||'{"path":["gen","auth"],"value":"y"}'
						   		||']'
						   	||'}}';
						*/
						teb, err := sizlib.JSONStringToData(rv)
						// fmt.Printf("teb = %s - at top\n", sizlib.SVar(teb))
						if err != nil {
							//fmt.Printf ( "Bad Data: %s\n", rv )
							// rv = fmt.Sprintf(`{ "status":"error","msg":"Error(10001): Parsing return value failed. sql-cfg.json[%s].G",%s, "err":%q }`, cfgTag, godebug.LFj(), err)
							fmt.Printf("rv=%s at:%s\n", rv, godebug.LF())
							done = true
							isError = true
							ReturnErrorMessage(500, "Database Error", "10001",
								fmt.Sprintf(`Error(10001): Parsing return value failed for SetSession, sql-cfg.json[%s].G %s err:%q`, cfgTag, godebug.LFj(), err), res, req, *ps, trx, hdlr) // status:error
							rv = ""
						}
						if !done {
							for vv, _ := range h.SetSession { // Set Seession data at this point - usually do_save is $session$ - this key gets procesed.
								if rw, ok := res.(*goftlmux.MidBuffer); ok {
									sesdata0, ok := teb[vv]
									sesdata := ConvRawSesData(godebug.SVar(sesdata0))

									// --------------------------------------------------------------------------------------------------------------------------
									// {
									//  "$session$":{"set":[{"path":["user","$is_logged_in$"],"value":"y"}]}
									// ,"auth_token":"a2c2a98d-0054-4c11-b476-2f52d5270904"
									// ,"config":"{}"
									// ,"customer_id":"1"
									// ,"privs":"[]"
									// ,"seq":"929b11cc-b31b-4607-a23f-0dba7d4abeac"
									// ,"status":"success"
									// ,"user_id":"0ba414c8-ccdc-475b-98c0-537fd75e64db"
									// }
									// --------------------------------------------------------------------------------------------------------------------------
									if hdlr.gCfg.DbOn("*", "TabServer2", "db-session") {
										fmt.Fprintf(os.Stderr, "%sAT:%s -- session -- %s\n", MiscLib.ColorYellow, MiscLib.ColorReset, godebug.LF())
									}
									for _, aset := range sesdata.Set {
										if hdlr.gCfg.DbOn("*", "TabServer2", "db-session") {
											fmt.Fprintf(os.Stderr, "%sAT:%s %s\n", MiscLib.ColorYellow, MiscLib.ColorReset, godebug.LF())
										}
										if aset.Path[0] == "user" && aset.Path[1] == "$is_logged_in$" {
											if hdlr.gCfg.DbOn("*", "TabServer2", "db-session") {
												fmt.Fprintf(os.Stderr, "%sAT:%s %s\n", MiscLib.ColorYellow, MiscLib.ColorReset, godebug.LF())
											}
											if aset.Value == "y" {
												if hdlr.gCfg.DbOn("*", "TabServer2", "db-session") {
													fmt.Fprintf(os.Stderr, "%sAT:%s %s\n", MiscLib.ColorYellow, MiscLib.ColorReset, godebug.LF())
												}
												rw.Session.Login()
											} else {
												if hdlr.gCfg.DbOn("*", "TabServer2", "db-session") {
													fmt.Fprintf(os.Stderr, "%sAT:%s %s\n", MiscLib.ColorYellow, MiscLib.ColorReset, godebug.LF())
												}
												rw.Session.Logout()
											}
										} else {
											if hdlr.gCfg.DbOn("*", "TabServer2", "db-session") {
												fmt.Fprintf(os.Stderr, "%sAT:%s %s\n", MiscLib.ColorYellow, MiscLib.ColorReset, godebug.LF())
											}
											rw.Session.SetData(aset.Path[0], aset.Path[1], aset.Value)
											rw.Session.SetRule(aset.Path[1], false, true)
										}
										if hdlr.gCfg.DbOn("*", "TabServer2", "db-session") {
											fmt.Fprintf(os.Stderr, "%sAT:%s %s\n", MiscLib.ColorYellow, MiscLib.ColorReset, godebug.LF())
											fmt.Fprintf(os.Stderr, "%sSession: %s, %s%s %s\n", MiscLib.ColorCyan, aset.Path, aset.Value, godebug.LF(), MiscLib.ColorReset)
										}
										trx.AddNote(1, fmt.Sprintf("Setting Session Name=%s Value=%s", aset.Path, aset.Value))
									}
									// --------------------------------------------------------------------------------------------------------------------------

									if ok {
										delete(teb, vv)
									}
								}
								// fmt.Printf("teb = %s - at bot\n", sizlib.SVar(teb))
								rv = sizlib.SVar(teb)
							} // xyzzy - report error!
						}
					} else {
						// new code - xyzzy87237283723 ----------------------------------------------------------------------------------------------------------------------------
						_, err := sizlib.JSONStringToData(rv)
						if err != nil {
							// xyzzyLog - this really should be logged!
							fmt.Printf("Bad Data (new case Tue Aug  4 11:36:38 MDT 2015) : %s\n", rv)
							fmt.Fprintf(os.Stderr, "Bad Data (new case Tue Aug  4 11:36:38 MDT 2015) : %s\n", rv)
							// xyzzyLog - this really should be logged!
							// rv = fmt.Sprintf(`{ "status":"error","msg":"Error(10001): Parsing return value failed. sql-cfg.json[%s].G",%s, "err":%q }`, cfgTag, godebug.LFj(), err)
							fmt.Printf("rv=%s at:%s\n", rv, godebug.LF())
							done = true
							isError = true
							ReturnErrorMessage(500, "Database Error", "10001",
								fmt.Sprintf(`Error(10001): Parsing return value failed, sql-cfg.json[%s].G %s err:%q`, cfgTag, godebug.LFj(), err), res, req, *ps, trx, hdlr) // status:error
							rv = ""
						}
						// end new code -------------------------------------------------------------------------------------------------------------------------------------------
					}
				} else {
					// rv = fmt.Sprintf("{\"status\":\"error\",\"msg\":\"Error(10007): Invalid number of rows returned. sql-cfg.json[%s].G\",\"nRows\":%d,%s}", cfgTag, n_row, godebug.LFj())
					ReturnErrorMessage(400, "Database Error", "10007",
						fmt.Sprintf(`Error(10007): Invalid number of rows returned. sql-cfg.json[%s].G nRows:%d, %s`, cfgTag, n_row, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
					rv = ""
					done = true
					isError = true
				}
			}
		}
	}
	if !done {
		if h.Query == "" && rv == "" {
			rv = fmt.Sprintf("{\"status\":\"success\",%s}", godebug.LFj())
		}
		//fmt.Printf ( "At %s\n", godebug.LF() )
		if h.Query != "" {
			HQuery := h.Query
			trx.AddNote(1, "Running .Query")
			// fmt.Printf("addNames=%s addVals=%s, %s\n", godebug.SVar(addNames), godebug.SVar(addVals), godebug.LF())
			// addNames = append(addNames, aName)
			// xyzzyAdditionalQueryParams01 // xyzzyAdd1
			// 1. if have {{where_additional_params}} -- then generate and replace in where clause, "and colname = $2 and colname2 = $3 ..."
			// if HasTemplateAdditionalParams(HQuery, addNames, addVals) {
			if strings.Index(HQuery, "%{where_additional_params%}") >= 0 {
				// HQuery = AddAdditionalParamsToWhere(HQuery, addNames, addVals)
				// fmt.Printf("addNames=%s len=%d, %s\n", godebug.SVar(addNames), len(addNames), godebug.LF())
				if len(addNames) == 0 {
					HQuery = strings.Replace(HQuery, "%{where_additional_params%}", " ", -1)
				} else {
					mv := make(map[string]string)
					s := ""
					a := " and "
					if strings.Index(HQuery, "where ") >= 0 || strings.Index(HQuery, "WHERE ") >= 0 {
					} else {
						a = " where "
					}
					for ii, name := range addNames {
						s = s + a + name + fmt.Sprintf(" = $%d", ii+len(h.P)+1)
					}
					mv["where_additional_params"] = s
					// fmt.Printf("s=%s\n", s)
					HQuery = sizlib.Qt(HQuery, mv)
					// fmt.Printf("HQuery After=%s, addVals=%s\n", HQuery, addVals)
					data = append(data, addVals...)
				}
			}
			s, gotIt := hdlr.HaveCachedData(res, req, h, HQuery, data...)
			if gotIt {
				trx.AddNote(1, "Data Was In Cache")
				rv = s
				trx.SetCacheData(HQuery, 1, rv, data...)
			} else {
				// Rows, err := sizlib.Sel ( res, req, db, h.Query, data... )
				fmt.Fprintf(os.Stderr, "\n%sAt .Query: %s, addNames=%s, %s%s\n\n", MiscLib.ColorYellow, HQuery, addNames, godebug.LF(), MiscLib.ColorReset)
				trx.SetQry(HQuery, 1, data...)
				Rows, err := hdlr.gCfg.Pg_client.Db.Query(HQuery, data...)
				if err != nil {
					rv = fmt.Sprintf("{ \"status\":\"error\",\"msg\":\"Error(10006): Query should have return rows but did not. sql-cfg.json[%s].Query[%s]\", \"error\":%q, %s }",
						cfgTag, sizlib.EscapeDoubleQuote(HQuery), err, godebug.LFj())
					done = true
					isError = true
					trx.SetQryDone(rv, "")
					ReturnErrorMessage(400, "Database Error", "10006",
						fmt.Sprintf(`Error(10006): Query should have returned rows but did not.  sql-cfg.json[%s].Query[%s] error:%q, %s`,
							cfgTag, sizlib.EscapeDoubleQuote(HQuery), err, godebug.LFj()), res, req, *ps, trx, hdlr) // status:error
					rv = ""
				} else {
					defer Rows.Close()
				}
				if !done {
					// ~/Projects/so/cfg/sql-cfg*.json - has PostJoin
					var id string
					if len(h.PostJoin) == 0 { // if we don't have any post join
						if db_post_join {
							fmt.Printf("*************** NO .Query Post Join ************************\n")
						}
						rv, id = sizlib.RowsToJson(Rows)
						trx.SetQryDone("", rv)
						if db_cache {
							godebug.IAmAt(fmt.Sprintf("Returned Data: %s\n", rv))
						}
						hdlr.CacheItForLater(res, req, h, id, rv, HQuery, data...)
					} else {
						trx.AddNote(1, "Post Join")
						if db_post_join {
							fmt.Printf("*************** supposed to .Query Post Join at this poin ************************\n")
						}
						rvX, id, n := sizlib.RowsToInterface(Rows) // parse the return data
						if db_post_join {
							fmt.Printf("PostJoin: id=%s n=%d, %s\n", id, n, godebug.LF())
						}
						pj_queries_run := 0
						var pj_cache map[string][]map[string]interface{}
						if h.CachePostJoin {
							pj_cache = make(map[string][]map[string]interface{})
						}
						for i, v := range rvX {

							hdlr.LimitPostJoinRows = 5000 // !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! test it.

							// (performance-imporovement) (xyzzySubQueriesCached)
							// var g_LimitPostJoinRows int = -1		// -1 indicates no limit
							if hdlr.LimitPostJoinRows >= 0 && pj_queries_run < hdlr.LimitPostJoinRows {
								if db_post_join {
									fmt.Printf("PostJoin: in loop at %d\n", i)
								}
								for j, w := range h.PostJoin {
									if db_post_join {
										fmt.Printf("PostJoin: On Post_join[%d], query=%s\n", j, w.Query)
									}
									if vv, ok := v[w.ColName]; ok {
										if db_post_join {
											fmt.Printf("PostJoin: non-null col = ->%v<-\n", vv)
										}
										var data []interface{}
										if len(w.P) > 0 {
											for _, x := range w.P {
												if _, ok2 := v[x]; ok2 {
													fmt.Printf("PostJoin: x=%s At: %s\n", x, godebug.LF())
													data = append(data, v[x])
												} else if t, ok2 := ps.GetByName(x); ok2 {
													fmt.Printf("PostJoin: x=%s At: %s\n", x, godebug.LF())
													data = append(data, t) // Allows for use of InjectData like $custoemr_id$ derived from AuthToken
												} else {
													fmt.Printf("PostJoin: x=%s At: %s\n", x, godebug.LF())
													data = append(data, "")
												}
											}
										} else {
											fmt.Printf("PostJoin: w.ColName=%s At: %s\n", w.ColName, godebug.LF())
											data = append(data, v[w.ColName])
										}
										// xyzzySubQueriesCached -- Check cache at this point -if- caced then do not increment pj_ and pull from cache
										if db_post_join {
											fmt.Printf("PostJoin: data=%s\n", godebug.SVar(data))
										}
										pj_found := false
										pj_key := ""
										if h.CachePostJoin {
											pj_key = pj_Key(w.Query, data...)
											if d, ok := pj_cache[pj_key]; ok {
												pj_found = true
												v[w.SetCol] = d
												rvX[i] = v
											}
										}
										if !pj_found {
											pj_queries_run++
											// func SelData2 ( db *sql.DB, q string, data ...interface{} ) ( []map[string]interface{}, error ) {
											d, err := sizlib.SelData2(hdlr.gCfg.Pg_client.Db, w.Query, data...)
											// Rows, err := hdlr.gCfg.Pg_client.Db.Query(Query, data...)
											if err != nil {
												fmt.Printf("Error(12062): on query %s, query=%s, data=%v\n", err, w.Query, data)
												trx.AddNote(1, fmt.Sprintf("Error(12062): on query %s, query=%s, data=%v\n", err, w.Query, data))
											} else {
												v[w.SetCol] = d
												rvX[i] = v
												// xyzzySubQueriesCached -- This is the spot to cache data -- [ w.Query ++ data... ]
												// should check for "cache" flag for this sub-query
												if h.CachePostJoin {
													pj_cache[pj_key] = d
												}
											}
										}
									}
								}
							} else {
								if db_post_join {
									fmt.Printf("PostJoin: Limit reached in loop %d, %s\n", i, godebug.LF())
								}
							}
						}
						var err error
						var rvB []byte
						/* xyzzyUnTested xyzzy xyzzy -- untested !!!!!!!!!!!!!!!!!!!!!!!!!!!! -------------------------- */
						rvB, err = json.MarshalIndent(rvX, "", "\t")
						rv = string(rvB)
						if rv == "null" {
							rv = "[]"
						}
						if err != nil {
							fmt.Printf("Error(12061): convering data to JSON, %s\n", err)
							trx.AddNote(1, fmt.Sprintf("Error(12061): convering data to JSON, %s\n", err))
						}
						trx.SetQryDone("", rv)
					}
				}
			}
		}
	}

	exit = false
	a_status = 200
	if len(h.CallAfter) > 0 {
		trx.AddNote(1, "CallAfter is True - functions will be called.")
		// fmt.Printf ( "At %s\n", godebug.LF() )
		for _, fx_name := range h.CallAfter {
			trx.AddNote(1, fmt.Sprintf("CallAfter [%s]", fx_name))
			if !exit {
				fmt.Fprintf(os.Stderr, "CallAfter [%s] rv before ->%s<-\n", fx_name, rv)
				rv, exit, a_status = hdlr.CallFunction("after", fx_name, res, req, cfgTag, rv, isError, cookieList, ps, trx)
				// , "CallAfter": ["SendReportsToGenMessage", "SendEmailToGenMessage"]
				fmt.Fprintf(os.Stderr, "CallAfter exit at bottom rv= %s exit=%v\n", rv, exit)
			}
		}
		// exit = false
	}
	if exit {
		fmt.Printf("****************** exit from after operations has been signaled **********************, rv=%s, %s\n", rv, godebug.LF())
		fmt.Fprintf(os.Stderr, "****************** exit from after operations has been signaled **********************, rv=%s, %s\n", rv, godebug.LF())
		done = true
		isError = true
		ReturnErrorMessageRv(a_status, rv, "Postprocessing signaled error", "18007",
			fmt.Sprintf(`Error(18007): Postprocessing signaled error. sql-cfg.json[%s] %s`, cfgTag, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
	}

	fmt.Fprintf(os.Stderr, "%s AT: %s%s\n", MiscLib.ColorGreen, godebug.LF(), MiscLib.ColorReset)

	if !isError {
		fmt.Fprintf(os.Stderr, "%s AT: %s%s\n", MiscLib.ColorGreen, godebug.LF(), MiscLib.ColorReset)
		trx.SetRvBody(rv)
		// io.WriteString(res, sizlib.JsonP(rv, res, req))
		if h.ReturnAsHash {
			fmt.Fprintf(os.Stderr, "%s AT: %s%s\n", MiscLib.ColorGreen, godebug.LF(), MiscLib.ColorReset)
			if h.ReturnGetPKAsHashTableName {
				tn := h.AssignedName
				if tn == "" {
					tn = h.TableName
				}
				if tn == "" {
					tn = "data"
				}
				rv = fmt.Sprintf(`{"status":"success","%s":%s}`, tn, rv)
			} else {
				fmt.Fprintf(os.Stderr, "%s AT: %s%s\n", MiscLib.ColorGreen, godebug.LF(), MiscLib.ColorReset)
				//	rv = fmt.Sprintf("{\"status\":\"success\",\"data\":%s}", rv)
				rv = fmt.Sprintf(`{"status":"success","data":%s}`, rv)
			}
		}
		fmt.Fprintf(os.Stderr, "%s AT: %s ->%s<- %s\n", MiscLib.ColorGreen, godebug.LF(), rv, MiscLib.ColorReset)
		io.WriteString(res, rv)
	}

}

// Exists reports whether the named file or directory exists.
//func Exists(name string) bool {
//    if _, err := os.Stat(name); err != nil {
//		if os.IsNotExist(err) {
//			return false
//		   }
//		}
//	return true
//}

// ==============================================================================================================================================================================
// Redis Keys
// Do we need a "prefix" or a "table" type thing to limit access to a set of keys?
// ==============================================================================================================================================================================
func (hdlr *TabServer2Type) RespHandlerRedis(res http.ResponseWriter, req *http.Request, cfgTag string, ps *goftlmux.Params, rw *goftlmux.MidBuffer) {

	rw, _ /*top_hdlr*/, ps, err := GetRwPs(res, req)

	h := hdlr.SQLCfg[cfgTag] // get configuration						// Xyzzy - what if config not defined for this item at this point!!!!
	var rv string = ""
	var done bool = false
	var isError bool = false
	mdata := make(map[string]string)
	cookieList := make(map[string]string)

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	defer hdlr.gCfg.RedisPool.Put(conn)

	// Setting Heders
	res.Header().Set("Content-Type", "application/json") // set default reutnr type

	// Get Initial Data
	// m, fr := sizlib.UriToStringMap(req) // pull out the ?name=value params				// xyzzy - add 'fr'
	// tr.TraceUri(req, m)

	trx := mid.GetTrx(rw)
	trx.SetFrom(fmt.Sprintf("sql-cfg.json[%s]Line#:%s", cfgTag, h.LineNo))
	trx.SetFunc(1)
	// trx.TraceUriPs(req, ps)
	TraceUriPs(trx, req, ps)
	// trx.SetRedisUsed ( key )

	// InjectData(m, fr, h, res, req) // xyzzy - add 'fr'
	InjectDataPs(ps, h, res, req) // xyzzy - add 'fr'

	for _, kk := range h.P { // Not real clear to me why you would ever cobine CRUD and p:[] values
		mdata[kk] = ps.ByName(kk)
		if !(*ps).HasName(kk) {
			trx.AddNote(2, fmt.Sprintf("Warning(12050): Missing data for %s, Empty string used!", kk))
		}
	}

	key_tmpl := h.Query
	goftlmux.AddValueToParams("key", key_tmpl, 'i', goftlmux.FromOther, ps)
	mdata["key"] = key_tmpl

	// trx.SetData(m, fr)
	SetDataPs(trx, ps)

	RedisReferenced(godebug.FUNCNAME(), cfgTag, hdlr)

	// Validate Query Params
	trx.AddNote(1, "Validate Query Parameters")
	err = ValidateQueryParams(ps, h, req) // Validate them
	if err != nil {
		// trx.AddNote(1, "Validate Query Parameters Failed:"+fmt.Sprintf("%v", err))
		// rv = fmt.Sprintf(`{"status":"error","code":"3","msg":"Error(12048): Invalid Query Parameters (%s) sql-cfg.json[%s]",%s}`, err, cfgTag, godebug.LFj())
		done = true
		isError = true
		// trx.ErrorReturn ( 1, rv )
		ReturnErrorMessage(406, "Invalid Parameter", "12048",
			fmt.Sprintf(`Error(12048): Invalid Query Parameters (%s) sql-cfg.json[%s] %s`, sizlib.EscapeError(err), cfgTag, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
		rv = ""
	}

	// Authenticate User if Necessary
	if !done {
		if h.LoginRequired {
			trx.AddNote(1, "User Authentication is Requried")
			err = hdlr.ValidateUserTrx(h, trx, res, req, ps, rw)
			if err != nil {
				done = true
				isError = true
				ReturnErrorMessage(401, "Authentication Failed, Invalid API Key", "1", fmt.Sprintf(`Error(17004): Invalid API Key %s`, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
				rv = ""
			} else {
				trx.SetUserInfo(ps.ByName("username"), ps.ByName("$user_id$"), ps.ByName("auth_token"))
			}
		}
	}

	// CallBefore
	exit := false
	a_status := 200
	if !done {
		if len(h.CallBefore) > 0 {
			trx.AddNote(1, "Functions to call before running queries.  CallBefore is set.")
			for _, fx_name := range h.CallBefore {
				if !exit {
					trx.AddNote(1, fmt.Sprintf("CallBefore[%s]", fx_name))
					rv, exit, a_status = hdlr.CallFunction("before", fx_name, res, req, cfgTag, rv, isError, cookieList, ps, trx)
				}
			}
		}
	}
	if exit {
		fmt.Printf("****************** exit from before operations has been signaled **********************, rv=%s, %s\n", rv, godebug.LF())
		done = true
		isError = true
		ReturnErrorMessageRv(a_status, rv, "Preprocessing signaled error", "18008",
			fmt.Sprintf(`Error(18008): Preprocessing signaled error. sql-cfg.json[%s] %s`, cfgTag, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
	}

	//	if ! done {
	//		for i, v := range m {
	//			mdata[i] = v[0]
	//		}
	//	}

	if !done {
		rv = `{"status":"success"}`
		done = true
		key := sizlib.Qt(mdata["key"], mdata) // template the key
		switch req.Method {
		case "GET": // Fetch It
			// rv, err = redis.String(rr.RedisDo("GET", key))
			rv, err := conn.Cmd("GET", key).Str() // Get the value
			if err != nil {
				rv = fmt.Sprintf(`{"status":"error","msg":"key not found","raw":%q}`, err)
				trx.SetQryDone("", rv)
				ReturnErrorMessage(406, "Error(12810): Key not found", "12810",
					fmt.Sprintf(`Error(12810): Key not found (%s) %s, %s`, err, key, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
				return
			}
		case "POST": // Set It
			// _, err = rr.RedisDo("SET",key,sizlib.SVar(mdata))				// xyzzy - seems like a bad idea, should tell what to save
			// _, err = rr.RedisDo("SET", key, sizlib.SVar(mdata)) // xyzzy - seems like a bad idea, should tell what to save
			err := conn.Cmd("SET", key, sizlib.SVar(mdata)).Err //
			if err != nil {
				rv = fmt.Sprintf(`{"status":"error","msg":"key not found","raw":%q}`, err)
				trx.SetQryDone("", rv)
				ReturnErrorMessage(406, "Error(12810): Key not found", "12810",
					fmt.Sprintf(`Error(12810): Key not found (%s) %s, %s`, err, key, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
				return
			}
		case "PUT": // Do Update - Get, Merge, Set
			// rv, err := redis.String(rr.RedisDo("GET", key))
			rv, err := conn.Cmd("GET", key).Str() // Get the value
			if err != nil {
				rv = fmt.Sprintf(`{"status":"error","msg":"key not found","raw":%q}`, err)
				trx.SetQryDone("", rv)
				ReturnErrorMessage(406, "Error(12810): Key not found", "12810",
					fmt.Sprintf(`Error(12810): Key not found (%s) %s, %s`, err, key, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
				return
			} else {
				var old_data map[string]string
				err = json.Unmarshal([]byte(rv), &old_data) // Convert the JSON column to values
				if err == nil {
					mdata = sizlib.ExtendDataS(old_data, mdata) // merge
				}
				//rr.RedisDo("SET", key, sizlib.SVar(mdata))
				conn.Cmd("SET", key, sizlib.SVar(mdata))
			}
		case "DELETE":
			// _, err = rr.RedisDo("DEL", key)
			err := conn.Cmd("DEL", key).Err // Get the value
			if err != nil {
				rv = fmt.Sprintf(`{"status":"error","msg":"key not found","raw":%q}`, err)
				trx.SetQryDone("", rv)
				ReturnErrorMessage(406, "Error(12810): Key not found", "12810",
					fmt.Sprintf(`Error(12810): Key not found (%s) %s, %s`, err, key, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
				return
			}
		case "HEAD":
			rv = key
		}
		trx.SetQryDone("", rv)
	}

	exit = false
	a_status = 200
	if len(h.CallAfter) > 0 {
		trx.AddNote(1, "CallAfter fore is True - functions will be called.")
		for _, fx_name := range h.CallAfter {
			trx.AddNote(1, fmt.Sprintf("CallAfter fore[%s]", fx_name))
			if !exit {
				rv, exit, a_status = hdlr.CallFunction("after", fx_name, res, req, cfgTag, rv, isError, cookieList, ps, trx)
			}
		}
	}
	if exit {
		fmt.Printf("****************** exit from after operations has been signaled **********************, rv=%s, %s\n", rv, godebug.LF())
		done = true
		isError = true
		ReturnErrorMessageRv(a_status, rv, "Postprocessing signaled error", "18007",
			fmt.Sprintf(`Error(18007): Postprocessing signaled error. sql-cfg.json[%s] %s`, cfgTag, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
	}

	if !isError { // xyzzy-new Thu Mar  5 09:06:21 MST 2015
		trx.SetRvBody(rv)
		// io.WriteString(res, sizlib.JsonP(rv, res, req))
		io.WriteString(res, rv)
	}

}

// ==============================================================================================================================================================================
// ==============================================================================================================================================================================
func PubEMailToSend(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (string, bool, int) {

	// rw, _ /*top_hdlr*/, _ /*ps*/, _ /*err*/ := GetRwPs(res, req)

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return "", true, 500
	}

	defer hdlr.gCfg.RedisPool.Put(conn)

	// rr.RedisDo("PUBLISH", "emailReadyToSend", fmt.Sprintf(`{"cmd":"readToSend"}`))
	conn.Cmd("PUBLISH", "emailReadyToSend", fmt.Sprintf(`{"cmd":"readToSend"}`))

	return rv, false, 200
}

var getTmplCache map[string][]byte
var getTmplLock sync.Mutex

func init() {
	getTmplCache = make(map[string][]byte)
}

func (hdlr *TabServer2Type) getTemplate(name string) string {
	getTmplLock.Lock()
	defer getTmplLock.Unlock()
	var tv []byte
	var ok bool
	if tv, ok = getTmplCache[name]; !ok || string(tv) == "" {
		fn := hdlr.EmailTemplateDir + name + ".tmpl"
		tv, err := ioutil.ReadFile(fn)
		if err != nil {
			fmt.Printf("Unable to open %s - email template file, err=%s, %s\n", fn, err, godebug.LF())
			fmt.Fprintf(os.Stderr, "%sUnable to open %s - email template file, err=%s, %s%s\n", MiscLib.ColorRed, fn, err, godebug.LF(), MiscLib.ColorReset)
		}
		getTmplCache[name] = tv
		return string(tv)
	}
	return string(tv)
	//	return `template={{.template}}
	//username={{.username}}
	//real_name={{.real_name}}
	//email_token={{.email_token}}
	//app={{.app}}
	//url={{.url}}
	//from={{.from}}
	//`
}

// s1, b1, s2, b2, err := TemplateEmail ( ed.Email["template"], ed )
func (hdlr *TabServer2Type) TemplateEmail(template_name string, mdata map[string]string) (s1, b1, b2 string, err error) {
	// s1, b1, s2, b2 = "s1", "b1", "s2", "b2"
	fmt.Printf("TemlateEmail mdata=%s\n", godebug.SVarI(mdata))
	s1 = tmplp.ExecuteATemplateByName(hdlr.getTemplate(template_name), "email_subject", mdata)
	b1 = tmplp.ExecuteATemplateByName(hdlr.getTemplate(template_name), "body_html", mdata)
	b2 = tmplp.ExecuteATemplateByName(hdlr.getTemplate(template_name), "body_text", mdata)
	return
}

//func init() {
//	fmt.Printf("init2 in main\n")
//}

//
//	Error		Meaning
//	-----		--------------------------------------
//	400			Bad Request
//	401			Unauthorized
//	402			Payment Required
//	403			Forbidden
//	404			Not Found
//	405			Method Not Allowed
//	406			Not Acceptable
//	412			Precondition Failed
//	417			Expectation Failed
//	428			Precondition Required
//
/*
	http.
		StatusBadRequest                   = 400
		StatusUnauthorized                 = 401
		StatusPaymentRequired              = 402
		StatusForbidden                    = 403
		StatusNotFound                     = 404
		StatusMethodNotAllowed             = 405
		StatusNotAcceptable                = 406
		StatusProxyAuthRequired            = 407
		StatusRequestTimeout               = 408
		StatusConflict                     = 409
		StatusGone                         = 410
		StatusLengthRequired               = 411
		StatusPreconditionFailed           = 412
		StatusRequestEntityTooLarge        = 413
		StatusRequestURITooLong            = 414
		StatusUnsupportedMediaType         = 415
		StatusRequestedRangeNotSatisfiable = 416
		StatusExpectationFailed            = 417
		StatusTeapot                       = 418

		StatusInternalServerError     = 500
		StatusNotImplemented          = 501
		StatusBadGateway              = 502
		StatusServiceUnavailable      = 503
		StatusGatewayTimeout          = 504
		StatusHTTPVersionNotSupported = 505

*/
func ConvertErrorToCode(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (string, bool, int) {
	var exit bool = false
	var a_status int = 200
	x, err := sizlib.JSONStringToData(rv)
	if err != nil {
		// rv = fmt.Sprintf(`{ "status":"error","msg":"Error(10009): Parsing return value failed. sql-cfg.json[%s] Post Function Call(CacheEUser)",%s, "err":%q }`, cfgTag, godebug.LFj(), err)
		// res.Header().Set("Content-Type", "text/html")
		// http.Error(res, "400 Bad Request", http.StatusBadRequest)
		ReturnErrorMessage(400, "Bad Request", "19043",
			fmt.Sprintf(`Error(19043): Bad Request (%s) sql-cfg.json[%s] %s %s`, sizlib.EscapeError(err), cfgTag, err, godebug.LF()),
			res, req, *ps, trx, hdlr) // status:error
		exit = true
		a_status = 500
	} else {
		if GetSI("status", x) != "success" || isError {
			// res.Header().Set("Content-Type", "text/html")
			// http.Error(res, "406 Not Acceptable", http.StatusNotAcceptable)
			ReturnErrorMessage(406, "Not Acceptable", "19044",
				fmt.Sprintf(`Error(19044): Not Acceptable (%s) sql-cfg.json[%s] %s %s`, sizlib.EscapeError(err), cfgTag, err, godebug.LF()),
				res, req, *ps, trx, hdlr) // status:error
			exit = true
			a_status = 406
		}
	}
	return rv, exit, a_status
}

// ==============================================================================================================================================================================
func GetSI(s string, data map[string]interface{}) string {
	if x, ok := data[s]; ok {
		return x.(string)
	}
	return ""
}

// ==============================================================================================================================================================================
func CacheEUser(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (string, bool, int) {

	// rw, _ /*top_hdlr*/, _ /*ps*/, _ /*err*/ := GetRwPs(res, req)

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return "", true, 500
	}

	defer hdlr.gCfg.RedisPool.Put(conn)

	if isError {
		return rv, true, 500
	}
	var exit bool = false
	var a_status int = 200
	if db_user_login {
		fmt.Printf("In CacheEUser\n")
	}
	x, err := sizlib.JSONStringToData(rv)
	if err != nil {
		// rv = fmt.Sprintf(`{ "status":"error","msg":"Error(10009): Parsing return value failed. sql-cfg.json[%s] Post Function Call(CacheEUser)",%s, "err":%q }`, cfgTag, godebug.LFj(), err)
		exit = true
		a_status = 406
		ReturnErrorMessage(406, "Error(10009): Parsing return value vaivfailed", "10009",
			fmt.Sprintf(`Error(10009): Parsing return value failed sql-cfg.json[%s] %s - Post function CacheEUser, %s`, cfgTag, err, godebug.LF()),
			res, req, *ps, trx, hdlr) // status:error
		rv = ""
	} else {
		success := GetSI("status", x)
		if success == "success" {
			username := ""
			if ps.HasName("username") {
				username = ps.ByName("username")
			} else { // On password recovery it is returned by the database
				username = GetSI("username", x) // If not supplied then returns ""
			}
			auth_token := GetSI("auth_token", x)
			privs := GetSI("privs", x)
			if privs == "" {
				privs = "[]"
			}
			config := GetSI("config", x)
			user_id := GetSI("user_id", x)
			customer_id := GetSI("customer_id", x)
			csrf_token := GetSI("csrf_token", x)
			rv = fmt.Sprintf(`{"status":"success","username":%q,"auth_token":%q,"customer_id":%q,"csrf_token":%q,"privs":%q,"config":%q}`,
				username, auth_token, customer_id, csrf_token, privs, config)

			if db_user_login {
				fmt.Printf("CacheEUser: SUCCESS-Caching it: username=%s auth_token=%s privs=->%s<- user_id=%s\n", username, auth_token, privs, user_id)
			}
			dt := fmt.Sprintf(`{"master":%q, "username":%q, "user_id":%q, "XSRF-TOKEN":%q, "customer_id":%q }`, privs, username, user_id, cookieList["XSRF-TOKEN"], customer_id)
			//rr.RedisDo("SET", "api:USER:"+auth_token, dt)
			//rr.RedisDo("EXPIRE", "api:USER:"+auth_token, 1*60*60) // Validate for 1 hour
			conn.Cmd("SET", "api:USER:"+auth_token, dt)
			conn.Cmd("EXPIRE", "api:USER:"+auth_token, 1*60*60) // Validate for 1 hour

			// rr.RedisDo("PUBLISH", "pubsub", fmt.Sprintf(`{"cmd":"login","username":%q,"auth_token":%q}`, username, auth_token))
			conn.Cmd("PUBLISH", "pubsub", fmt.Sprintf(`{"cmd":"login","username":%q,"auth_token":%q}`, username, auth_token))
		} else {
			exit = true
			a_status = 401
			res.WriteHeader(401) // Failed to login
		}
	}
	return rv, exit, a_status
}

// ==============================================================================================================================================================================
/*
	Issue00079: NoLog parameters
	Issue00018: Actual CC Processing with account + fake cards																	4 hr
			2. NoLog parameters
			3. Put in actual test params with account - see that it works

Try some of these numbers:
	4000 0000 0000 0002
	4026 0000 0000 0002
	5018 0000 0009
	5100 0000 0000 0008
	6011 0000 0000 0004

*/
//func ChargeCreditCard(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (string, bool, int) {
//	var exit bool = false
//	var a_status int = 200
//	fmt.Printf("************************************* Charge Credit Card Called *************************************\n")
//
//	if MiscLib.InArray("credit_card_test_mode", hdlr.DebugFlags) >= 0 {
//		ps.SetValue("cc_authorize", "test-mode")
//		goftlmux.AddValueToParams("cc_authorize", "'test-mode'", 's', goftlmux.FromOther, &ps)
//		fmt.Printf("************************************* Credit Card in test-mode *************************************, rv=%s\n", rv)
//		rv = `{"status":"success"}`
//		return rv, exit, a_status
//	}
//
//	auth := AuthorizeNet.AuthorizeNet{
//		Login:     hdlr.AuthorizeNetLogin, // "<YourLogin>",
//		Key:       hdlr.AuthorizeNetKey,   // "<YourKey>",
//		DupWindow: 120,
//		TestMode:  true,
//	}
//
//	card := AuthorizeNet.CardInfoType{
//		CreditCardNumber: ps.ByNameDflt("credit_card", "x"),                                         // 5555 5555 5555 5555
//		CVV:              ps.ByNameDflt("ccv", "x"),                                                 // 123
//		Month_Year:       ps.ByNameDflt("exp_month", "1") + "/" + ps.ByNameDflt("exp_year", "2017"), // 01/2017
//		Method:           ps.ByNameDflt("card_type", "visa"),                                        // METHOD_VISA,
//	}
//
//	// func (ps *Params) ByNameDflt(name string, dflt string) (rv string) {
//	// Extend ""ps"" user info with info from t_user, p_cart
//
//	data := AuthorizeNet.AuthorizeNetType{
//		InvoiceNumber:   ps.ByNameDflt("invoice_no", "y"),                // "123444",		// need to create the invoice number in p_cart when "address" is added
//		Amount:          ps.ByNameDflt("ex_total", "0.0"),                // "5.56",
//		Description:     ps.ByNameDflt("customer_transaction_name", "y"), // "My Test transaction",
//		FirstName:       ps.ByNameDflt("b_first_name", "y"),              // Address for shipping/billing pulled from p_cart
//		LastName:        ps.ByNameDflt("b_last_name", "y"),               //
//		Company:         ps.ByNameDflt("b_company", "y"),                 //
//		BillingAddress:  ps.ByNameDflt("b_line_1", "y"),                  //
//		BillingCity:     ps.ByNameDflt("b_city", "y"),                    //
//		BillingState:    ps.ByNameDflt("b_state", "y"),                   //
//		BillingZip:      ps.ByNameDflt("b_postal_code", "y"),             //
//		BillingCountry:  ps.ByNameDflt("b_country", "y"),                 //
//		Phone:           ps.ByNameDflt("b_phone", "y"),                   //
//		Email:           ps.ByNameDflt("email", "y"),                     // From auth_token -> t_user.email
//		CustomerId:      ps.ByNameDflt("user_id", "y"),                   // From auth_token -> t_user.id
//		CustomerIp:      ps.ByNameDflt("$IP$", "y"),                      //
//		ShipToFirstName: ps.ByNameDflt("s_first_name", "y"),              //
//		ShipToLastName:  ps.ByNameDflt("s_last_name", "y"),               //
//		ShipToCompany:   ps.ByNameDflt("s_company", "y"),                 //
//		ShipToAddress:   ps.ByNameDflt("s_line_1", "y"),                  //
//		ShipToCity:      ps.ByNameDflt("s_city", "y"),                    //
//		ShipToState:     ps.ByNameDflt("s_state", "y"),                   //
//		ShipToZip:       ps.ByNameDflt("s_postal_code", "y"),             //
//		ShipToCountry:   ps.ByNameDflt("s_country", "y"),                 //
//	}
//
//	// Authorize a payment
//	response := auth.Authorize(card, data, false)
//	if !response.IsApproved() {
//		fmt.Printf("CC Failed to Approve: %s\n", sizlib.SVar(response))
//		return fmt.Sprintf(`{"status":"error","msg":"Credit card failed to authorize, %s"}`, sizlib.EscapeDoubleQuote(response.ReasonText)), true, 402
//	}
//	fmt.Printf("%s\n", response)
//	fmt.Printf("Successful Authorization with id: %s ", response.TransId)
//
//	// Example of capture the preious authorization (using the transactionId)
//	response = auth.CapturePreauth(response.TransId, ps.ByNameDflt("ex_total", "0.0")) // 5.56
//	if !response.IsApproved() {
//		//log.Print("Capture failed : ")
//		//log.Print(response)
//		fmt.Printf("CC Failed to Preauth: %s\n", sizlib.SVar(response))
//		return fmt.Sprintf(`{"status":"error","msg":"Credit card failed to authorize/capture, %s"}`, sizlib.EscapeDoubleQuote(sizlib.SVar(response))), true, 402
//	}
//	// log.Print(response)
//	// log.Print("Successful Capture !")
//
//	// ps.SetValue("cc_authorize", sizlib.SVar(response))
//	goftlmux.AddValueToParams("cc_authorize", sizlib.SVar(response), 'J', goftlmux.FromOther, &ps)
//
//	return rv, exit, a_status
//}

// ==============================================================================================================================================================================
func DeCacheEUser(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (string, bool, int) {

	// rw, _ /*top_hdlr*/, _ /*ps*/, _ /*err*/ := GetRwPs(res, req)

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return "", true, 500
	}

	defer hdlr.gCfg.RedisPool.Put(conn)

	if isError {
		return rv, true, 500
	}
	var exit bool = false
	var a_status int = 200
	if db_user_login {
		fmt.Printf("In DeCacheEUser\n")
	}
	if ps.HasName("auth_token") {
		auth_token := ps.ByName("auth_token")
		// rr.RedisDo("DEL", "api:USER:"+auth_token)
		// rr.RedisDo("PUBLISH", "pubsub", fmt.Sprintf(`{"cmd":"logout","auth_token":%q}`, auth_token))
		conn.Cmd("DEL", "api:USER:"+auth_token)
		conn.Cmd("PUBLISH", "pubsub", fmt.Sprintf(`{"cmd":"logout","auth_token":%q}`, auth_token))
	}
	return rv, exit, a_status
}

// ==============================================================================================================================================================================
func AfterPasswordChange(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (string, bool, int) {

	// rw, _ /*top_hdlr*/, _ /*ps*/, _ /*err*/ := GetRwPs(res, req)

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return "", true, 500
	}

	defer hdlr.gCfg.RedisPool.Put(conn)

	if isError {
		return rv, true, 500
	}
	var exit bool = false
	var a_status int = 200

	// xyzzy - should this use the d.b. to find every auth_token used by this user and de-auth every one?
	x, err := sizlib.JSONStringToData(rv)
	if err != nil {
		// rv = fmt.Sprintf(`{ "status":"error","msg":"Error(10009): Parsing return value failed. sql-cfg.json[%s] Post Function Call(CacheEUser)",%s, "err":%q }`, cfgTag, godebug.LFj(), err)
		exit = true
		a_status = 406
		ReturnErrorMessage(406, "Error(10009): Parsing return value vaivfailed", "10009",
			fmt.Sprintf(`Error(10009): Parsing return value failed sql-cfg.json[%s] %s - Post function CacheEUser, %s`, cfgTag, err, godebug.LF()),
			res, req, *ps, trx, hdlr) // status:error
		rv = ""
	} else {
		success := GetSI("status", x)
		if success == "success" {
			// Get rid of old auth in Redis
			if db_user_login {
				fmt.Printf("In AfterPasswordChange\n")
			}
			auth_token := ps.ByName("auth_token")
			// rr.RedisDo("DEL", "api:USER:"+auth_token)
			conn.Cmd("DEL", "api:USER:"+auth_token)

			// fmt.Printf ( "just before extracting data, rv=%s\n", rv );
			username := GetSI("username", x)
			auth_token = GetSI("auth_token", x)
			privs := GetSI("privs", x)
			user_id := GetSI("user_id", x)
			customer_id := GetSI("customer_id", x)
			csrf_token := GetSI("csrf_token", x)
			rv = fmt.Sprintf(`{"status":"success","username":%q,"auth_token":%q,"csrf_token":%q}`, username, auth_token, csrf_token)

			x_cookie, ok := req.Cookie("XSRF-TOKEN")
			cookie := ""
			if ok == nil {
				cookie = x_cookie.String()
				if db_user_login {
					fmt.Printf("AfterPasswordChange Raw : Cookie=%s ok=%v\n", cookie, ok)
				}
				cookie = strings.Split(cookie, "=")[1]
			}
			if db_user_login {
				fmt.Printf("AfterPasswordChange Cookie=%s ok=%v\n", cookie, ok)
			}
			if ok != nil {
				t_cookie, _ := uuid.NewV4()
				cookie = t_cookie.String()
				expire := time.Now().AddDate(0, 0, 2) // Years, Months, Days==2 // xyzzy - should be a config - on how long to keep cookie
				secure := false
				if req.TLS != nil {
					secure = true
				}
				if db_user_login {
					fmt.Printf("   not ok, generating a new one: %s\n", cookie)
				}
				cookieObj := http.Cookie{Name: "XSRF-TOKEN", Value: cookie, Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400, Secure: secure, HttpOnly: false}
				http.SetCookie(res, &cookieObj)
			}
			if db_user_login {
				fmt.Printf("   OK it is:%s\n", cookie)
			}

			if db_user_login {
				fmt.Printf("AfterPasswordChange SUCCESS-Caching it: username=%s auth_token=%s privs=->%s<- user_id=%s XSRF-TOKEN from cookie=%s\n", username, auth_token, privs, user_id, cookie)
			}
			dt := fmt.Sprintf(`{"master":%q, "username":%q, "user_id":%q, "XSRF-TOKEN":%q, "customer_id":%q }`, privs, username, user_id, cookie, customer_id)
			//rr.RedisDo("SET", "api:USER:"+auth_token, dt)
			//rr.RedisDo("EXPIRE", "api:USER:"+auth_token, 1*60*60) // Validate for 1 hour
			conn.Cmd("SET", "api:USER:"+auth_token, dt)
			conn.Cmd("EXPIRE", "api:USER:"+auth_token, 1*60*60) // Validate for 1 hour
		}
	}
	return rv, exit, a_status
}

/*
	,"/api/test/register_new_user": { "g": "test_register_new_user($1,$2,$3,$4,$5,$6,$7,$8,$9)", "p": [ "username", "password", "$ip$", "email", "real_name", "$url$", "csrf_token", "site", "name" ], "nokey":true

for ".G" - use "P" parameters and GetStoredProcedurePlist - so...
	PostgreSQL - ($1,$2...$n)
	ODBC		" "?,?,...       nof them
	Oracle		(:p1,:p2...:pN)

*/
func (hdlr *TabServer2Type) GetStoredProcedurePlist(np int) (rv string) {
	if hdlr.GetDbType() == DbType_postgres {
		rv = "("
		com := ""
		for i := 1; i <= np; i++ {
			rv += fmt.Sprintf("%s$%d", com, i)
			com = ","
		}
		rv += ")"
	} else if hdlr.GetDbType() == DbType_odbc {
		rv = " "
		com := ""
		for i := 0; i < np; i++ {
			rv += fmt.Sprintf("%s?", com)
			com = ","
		}
		rv += " "
	} else if hdlr.GetDbType() == DbType_Oracle {
		rv = "("
		com := ""
		for i := 0; i < np; i++ {
			rv += fmt.Sprintf("%s:p%d", com, i)
			com = ","
		}
		rv += ")"
	} else {
		panic("Error(00000): Not implemented yet.")
	}
	return
}

func MapKeys(kk map[string]interface{}) (rv string) {
	rv = ""
	com := ""
	for i := range kk {
		rv += com + i
		com = ","
	}
	if rv == "" {
		rv = " *** no keys *** "
	}
	return
}

func Get1stRowFromMap(kk map[string]interface{}) (rv interface{}) {
	rv = "{}"
	for _, v := range kk {
		rv = v
		return
	}
	return
}

func (hdlr *TabServer2Type) BindPlaceholder(n int) (rv string) {
	rv = "? "
	if hdlr.GetDbType() == DbType_postgres {
		rv = fmt.Sprintf("$%d ", n)
	} else if hdlr.GetDbType() == DbType_odbc {
		rv = "? "
	} else if hdlr.GetDbType() == DbType_mySQL {
		rv = "? "
	} else if hdlr.GetDbType() == DbType_MsSql {
		rv = "? "
	} else if hdlr.GetDbType() == DbType_Oracle {
		rv = fmt.Sprintf(":p%d ", n)
	}
	return
}

/*
create table "p_attr_meta" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "customer_id"			char varying (40) default '1'
	, "attr_type"			char varying (30) not null default 'str' check ( "attr_type" in ( 'int', 'int-r', 'str', 'str-r', 'float', 'float-r', 'date' , 'date-r', 'time', 'time-r', 'list-of', 'location', 'fraction' ) )
	, "table_name"			char varying (250) not null
	, "seq_no"				int
	, "attr_name"			char varying (250) not null
	, "fuzzy_search"		char varying(10) default 'n' not null
	, "attr_synonyms"		text not null
	, "default_value"		text 			-- Defaulit Value
	, "min_value"			text 			-- Min
	, "max_value"			text 			-- Max
	, "value_list"			text 			-- JSON list of values
	, "check_valid_data"	text 			-- RE used to validate
);
*/

type EAAttributeInfo struct {
	Attr_type        string
	Seq_no           int64
	Fuzzy_search     string
	Attr_synonyms    string
	Default_value    string
	Min_value        string
	Max_value        string
	Value_list       string
	Check_valid_data string
}

var EADataLoaded map[string]bool
var EA_Data map[string]map[string]EAAttributeInfo

func init() {
	EADataLoaded = make(map[string]bool)
	EA_Data = make(map[string]map[string]EAAttributeInfo)
}

func IfNullStr(x interface{}) string {
	if x == nil {
		return ""
	} else {
		return x.(string)
	}
}

func IfNullInt64(x interface{}) int64 {
	if x == nil {
		return 0
	} else {
		return x.(int64)
	}
}

func LoadEAData(table_name string, hdlr *TabServer2Type) {
	fmt.Printf("At, %s: Loading data for %s\n", godebug.LF(), table_name)
	d := sizlib.SelData(hdlr.gCfg.Pg_client.Db, "select * from p_attr_meta where table_name = $1", table_name)
	fmt.Printf("At, %s: data %s\n", godebug.LF(), sizlib.SVarI(d))
	for _, v := range d {
		fmt.Printf("At, %s\n", godebug.LF())
		if _, ok := EA_Data[table_name]; !ok {
			EA_Data[table_name] = make(map[string]EAAttributeInfo)
		}
		EA_Data[table_name][IfNullStr(v["attr_name"])] = EAAttributeInfo{
			Attr_type:        IfNullStr(v["attr_type"]),
			Seq_no:           IfNullInt64(v["seq_no"]),
			Fuzzy_search:     IfNullStr(v["fuzzy_search"]),
			Attr_synonyms:    IfNullStr(v["attr_synonyms"]),
			Default_value:    IfNullStr(v["default_value"]),
			Min_value:        IfNullStr(v["min_value"]),
			Max_value:        IfNullStr(v["max_value"]),
			Value_list:       IfNullStr(v["value_list"]),
			Check_valid_data: IfNullStr(v["check_valid_data"]),
		}
	}
	fmt.Printf("At, %s\n", godebug.LF())
	EADataLoaded[table_name] = true
}

// xyzzy synonyms for columns
// xyzzy count of values for columns in context
func EALookup(name string, table_name string, hdlr *TabServer2Type) (bool, *EAAttributeInfo) {
	fmt.Printf("At, %s\n", godebug.LF())
	if b, ok := EADataLoaded[table_name]; !ok || !b {
		fmt.Printf("At, %s\n", godebug.LF())
		LoadEAData(table_name, hdlr)
	}
	if x, ok := EA_Data[table_name][name]; ok {
		fmt.Printf("At, %s\n", godebug.LF())
		return ok, &x
	}
	fmt.Printf("At, %s\n", godebug.LF())
	return false, nil
}

// xyzzy5000
// qt_m["cat_col"] = fmt.Sprintf(`%s%s%s%s`, h.setWhereAlias, DbBeginQuote, h.Category_col, DbEndQuote)
// t, _ := GetDataListAsInList("s", vv, trx, h, bind)
// Used in categories - this is a postgresql specific function. -- Sets up an in list.
func GetDataListAsInList(ty string, wc WhereClause, trx *tr.Trx, h SQLOne, bind *[]interface{}) (string, error) {
	s := "("
	com := ""
	for _, vv := range wc.List {
		switch ty {
		case "i":
			s += com + fmt.Sprintf("%d", vv.Val1i)
		case "f":
			s += com + fmt.Sprintf("%f", vv.Val1f)
		case "u": // UUID/GUID
			s += com + "'" + vv.Val1s + "'"
		case "":
			fallthrough
		case "s":
			s += com + "'" + vv.Val1s + "'"
		/* d, t, e */
		default:
			trx.AddNote(1, fmt.Sprintf("Invalid Type: %s for column %s", ty, wc.Name))
			return "", errors.New(fmt.Sprintf("Error(12233): Invalid Type: %s for column %s", ty, wc.Name))
		}
		com = ","
	}
	s = s + ")"
	fmt.Printf("Category|Tag At, %s ->%s<-\n", godebug.LF(), s)
	return s, nil
}

// Used in categories - this is a postgresql specific function.
func GetDataListAsArray(ty string, wc WhereClause, trx *tr.Trx, h SQLOne, bind *[]interface{}) (string, error) {
	s := "{"
	com := ""
	for _, vv := range wc.List {
		switch ty {
		case "i":
			s += com + fmt.Sprintf("%d", vv.Val1i)
		case "f":
			s += com + fmt.Sprintf("%f", vv.Val1f)
		case "u": // UUID/GUID
			s += com + `"` + vv.Val1s + `"`
		case "":
			fallthrough
		case "s":
			s += com + `"` + vv.Val1s + `"`
		/* d, t, e */
		default:
			trx.AddNote(1, fmt.Sprintf("Invalid Type: %s for column %s", ty, wc.Name))
			return "", errors.New(fmt.Sprintf("Error(12233): Invalid Type: %s for column %s", ty, wc.Name))
		}
		com = ","
	}
	s = s + "}"
	fmt.Printf("Category|Tag At, %s ->%s<-\n", godebug.LF(), s)
	return s, nil
}

func GetColByType(attr_type string) string {
	switch attr_type {
	case "i":
		return "val1i"
	case "f":
		return "val1f"
	case "u": // UUID/GUID
		return "val1s"
	case "":
		fallthrough
	case "s":
		return "val1s"
	case "d", "t", "e":
		return "val1d"
	default:
		return "val1s"
	}
}

func AddBindValueByType(bind *[]interface{}, vv WhereClause, ty string) (pos int) {
	switch ty {
	case "i":
		return AddBindValue(bind, vv.Val1i)
	case "f":
		return AddBindValue(bind, vv.Val1f)
	case "u": // UUID/GUID
		return AddBindValue(bind, vv.Val1s)
	case "":
		fallthrough
	case "s":
		return AddBindValue(bind, vv.Val1s)
	case "d", "t", "e":
		return AddBindValue(bind, vv.Val1d)
	default:
		return AddBindValue(bind, vv.Val1s)
	}
}

// xyzzy - missing $customer_id$ if set/used
/*
http://tech.pro/tutorial/1142/building-faceted-search-with-postgresql  -- faceted-filters.html (saved)
http://www.postgresql.org/docs/8.3/static/textsearch-indexes.html
http://stackoverflow.com/questions/1540374/why-are-postgresql-text-search-gist-indexes-so-much-slower-than-gin-indexes
http://www.youlikeprogramming.com/2012/01/full-text-search-fts-in-postgresql-9-1/		-- trigger to update tsvecotr colum/index

http://blog.timothyandrew.net/blog/2013/06/24/recursive-postgres-queries/
http://stackoverflow.com/questions/11834579/postgresql-hierarchical-category-tree -- Category Example With Queries
https://gist.github.com/chanmix51/3225313 -- GIST example with get_chilren_of, get_parent_of, list and create table stuff

select * from p_get_children_of ( '{"guns","stock"}'::varchar[] );

, "key_word_col_name": "key_word"
, "key_word_list_col": "__keyword__"
, "key_word_tmpl": " %{kw_col%} @@ plainto_tsquery( %{kw_vals%} ) "

, "category_col_name": "category"
, "category_col": "category_id"
, "category_tmpl": " %{cat_col%} in ( select c1.\"id\" from p_get_children_of ( '%{cat_val%}'::varchar[] ) ) "

, "attr_table_name": "p_cart"
, "attr_col": "id"
, "attr_tmpl":" %{pk_attr_id%} in ( select a1.\"fk_id\" from \"p_attr\" as a1 where a1.\"attr_type\" = %{attr_type%} and a1.\"attr_name\" = %{attr_name%} and a1.\"%{ref_col%}\" %{attr_op%} %{attr_val%} )"

	Key_word_col_name string
	Key_word_list_col string
	Key_word_tmpl     string

	Category_col_name string
	Category_col      string
	Category_tmpl     string

	Attr_table_name string
	Attr_col        string
	Attr_tmpl       string

*/
func ExtendeAttributes(vv WhereClause, wc WhereClause, trx *tr.Trx, h SQLOne, bind *[]interface{}, hdlr *TabServer2Type) (processed bool, rv string, err error) {
	var bp int

	processed = false
	rv = " false "
	err = nil
	qt_m := make(map[string]string)
	if vv.Name == h.Key_word_col_name {
		fmt.Printf("%sAt, %s%s\n", MiscLib.ColorCyan, godebug.LF(), MiscLib.ColorReset)
		processed = true
		trx.AddNote(2, "Processing key_word extended attribute")
		// legit ops on keywords,
		//		kw = 'a'
		// https://www.postgresql.org/docs/9.5/static/textsearch-controls.html#TEXTSEARCH-PARSING-QUERIES
		// Looks like operator should be "@@"
		if !sizlib.InArray(vv.Op, []string{"==", "="}) {
			fmt.Printf("At, %s\n", godebug.LF())
			trx.AddNote(1, fmt.Sprintf("Error: Processing extended attributes, invalid op=%s %s%s", vv.Op, h.setWhereAlias, vv.Name))
			return false, "", errors.New(fmt.Sprintf("Error(14239): Error: Processing extended attributes, invalid op=%s %s%s", vv.Op, h.setWhereAlias, vv.Name))
		} else {
			fmt.Printf("At, %s\n", godebug.LF())
			bp = AddBindValue(bind, vv.Val1s)          // keywords are alwasy strings, or list of strings
			qt_m["kw_vals"] = hdlr.BindPlaceholder(bp) // %{kw_vals%}
			qt_m["kw_col"] = fmt.Sprintf(`%s%s%s%s`, h.setWhereAlias, DbBeginQuote, h.Key_word_list_col, DbEndQuote)
			rv = sizlib.Qt(h.Key_word_tmpl, qt_m)
		}
	} else if vv.Name == h.Category_col_name {
		fmt.Printf("%sAt, %s%s\n", MiscLib.ColorCyan, godebug.LF(), MiscLib.ColorReset)
		processed = true
		trx.AddNote(2, "Processing category extended attribute")
		// %{cat_vals%}
		// legit ops on keywords,
		//		cat in ( 'id1', 'id2' )
		if vv.Op != "in" {
			fmt.Printf("At, %s\n", godebug.LF())
			trx.AddNote(1, fmt.Sprintf("Error: Processing extended attributes, invalid op=%s %s%s", vv.Op, h.setWhereAlias, vv.Name))
			return false, "", errors.New(fmt.Sprintf("Error(14239): Error: Processing extended attributes, invalid op=%s %s%s", vv.Op, h.setWhereAlias, vv.Name))
		} else {
			fmt.Printf("At, %s\n", godebug.LF())
			s, err := GetDataListAsArray("s", vv, trx, h, bind)
			// xyzzy5000
			t, _ := GetDataListAsInList("s", vv, trx, h, bind)
			if err != nil {
				fmt.Printf("At, %s\n", godebug.LF())
				trx.AddNote(1, fmt.Sprintf("Error: Processing extended attributes, invalid data type=%s %s%s", "s", h.setWhereAlias, vv.Name))
				return false, "", errors.New(fmt.Sprintf("Error(14239): Error: Processing extended attributes, invalid data type=%s %s%s", "s", h.setWhereAlias, vv.Name))
			} else {
				qt_m["cat_vals"] = s
				qt_m["cat_in_vals"] = t
				qt_m["cat_col"] = fmt.Sprintf(`%s%s%s%s`, h.setWhereAlias, DbBeginQuote, h.Category_col, DbEndQuote)
				fmt.Printf("s = ->%s<- qt_m=%s At, %s\n", s, godebug.SVar(qt_m), godebug.LF())
				rv = sizlib.Qt(h.Category_tmpl, qt_m)
			}
		}
	} else if h.Attr_table_name != "" { // xyzzy - this should be some checkon name in attribute table names?
		fmt.Printf("At, %s\n", godebug.LF())
		fnd, x := EALookup(vv.Name, h.Attr_table_name, hdlr)
		if fnd {
			fmt.Printf("At, %s\n", godebug.LF())
			processed = true
			trx.AddNote(2, "Processing dynamic attribute extended attribute")
			qt_m["attr_col"] = fmt.Sprintf(`%s%s%s%s`, h.setWhereAlias, DbBeginQuote, h.Attr_col, DbEndQuote)
			qt_m["attr_type"] = x.Attr_type                // %{attr_type%} - from meta, x.Attr_type
			qt_m["attr_name"] = vv.Name                    // %{attr_name%} - vv.Name
			qt_m["ref_col"] = GetColByType(x.Attr_type)    // %{ref_col%} - derived from vv.Name + vv.Op - and fuzzy_search
			qt_m["attr_op"] = vv.Op                        // %{attr_op%} - vv.Op
			bp = AddBindValueByType(bind, vv, x.Attr_type) // keywords are alwasy strings, or list of strings
			qt_m["attr_vals"] = hdlr.BindPlaceholder(bp)   // %{attr_vals%} -- vv.Val1s, etc (based on type)
			rv = sizlib.Qt(h.Attr_tmpl, qt_m)
		}
	} else if vv.Name == h.Tag_col_name {
		fmt.Printf("%sAt, %s%s\n", MiscLib.ColorCyan, godebug.LF(), MiscLib.ColorReset)
		processed = true
		trx.AddNote(2, "Processing tag extended attribute")
		// %{cat_vals%}
		// legit ops on keywords,
		//		cat in ( 'id1', 'id2' )
		if vv.Op != "in" {
			fmt.Printf("At, %s\n", godebug.LF())
			trx.AddNote(1, fmt.Sprintf("Error: Processing extended attributes, invalid op=%s %s%s", vv.Op, h.setWhereAlias, vv.Name))
			return false, "", errors.New(fmt.Sprintf("Error(14239): Error: Processing extended attributes, invalid op=%s %s%s", vv.Op, h.setWhereAlias, vv.Name))
		} else {
			fmt.Printf("At, %s\n", godebug.LF())
			s, err := GetDataListAsArray("s", vv, trx, h, bind)
			// xyzzy5000
			t, _ := GetDataListAsInList("s", vv, trx, h, bind)
			if err != nil {
				fmt.Printf("At, %s\n", godebug.LF())
				trx.AddNote(1, fmt.Sprintf("Error: Processing extended attributes, invalid data type=%s %s%s", "s", h.setWhereAlias, vv.Name))
				return false, "", errors.New(fmt.Sprintf("Error(14239): Error: Processing extended attributes, invalid data type=%s %s%s", "s", h.setWhereAlias, vv.Name))
			} else {
				qt_m["tag_vals"] = s
				qt_m["tag_in_vals"] = t
				qt_m["tag_col"] = fmt.Sprintf(`%s%s%s%s`, h.setWhereAlias, DbBeginQuote, h.Tag_col, DbEndQuote)
				rv = sizlib.Qt(h.Tag_tmpl, qt_m)
				fmt.Printf("%ss = ->%s<- qt_m=%s rv= ->%s<- At, %s%s\n", MiscLib.ColorYellow, s, godebug.SVar(qt_m), rv, godebug.LF(), MiscLib.ColorReset)
			}
		}
	}
	return
}

// func respHandlerCreateNewDeviceId(www http.ResponseWriter, req *http.Request) {
// if rw, ok := www.(*goftlmux.MidBuffer); ok {
func GetRwPs(www http.ResponseWriter, req *http.Request) (rw *goftlmux.MidBuffer, top_hdlr interface{}, ps *goftlmux.Params, err error) {
	var ok bool
	rw, top_hdlr, ok = GetRwHdlrFromWWW(www, req) // xyzzy - hdlr alwasy is top - an error
	if rw != nil {
		ps = &rw.Ps
	}
	if !ok {
		err = fmt.Errorf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF(2))
		fmt.Fprintf(os.Stderr, "Failed to get rw/hdlr, AT: %s\n", godebug.LF())
		return
	}
	// fmt.Fprintf(os.Stderr, "Just before setting ps, %v, %v, %s\n", ps, rw.Ps, godebug.LF())
	return
}

// ============================================================================================================================================
func GetRwHdlrFromWWW(www http.ResponseWriter, req *http.Request) (rw *goftlmux.MidBuffer, top_hdlr interface{}, ok bool) {

	rw, ok = www.(*goftlmux.MidBuffer)
	if !ok {
		// AnError(hdlr, www, req, 500, 5, fmt.Sprintf("hdlr not correct type in rw.!, %s\n", godebug.LF()))
		fmt.Fprintf(os.Stderr, "Passed Wrong Thing, AT: %s\n", godebug.LF())
		return
	}

	top_hdlr = rw.Hdlr
	return
}

const debugCrud01 = false

/*
	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	defer hdlr.gCfg.RedisPool.Put(conn)
*/

type DbType int

const (
	DbType_postgres DbType = 1
	DbType_Oracle   DbType = 2
	DbType_MsSql    DbType = 3
	DbType_mySQL    DbType = 4
	DbType_odbc     DbType = 5
)

func (dd DbType) String() string {
	switch dd {
	case DbType_postgres:
		return "DbType_postgres"
	case DbType_Oracle:
		return "DbType_Oracle"
	case DbType_MsSql:
		return "DbType_MsSql"
	case DbType_mySQL:
		return "DbType_mySQL"
	case DbType_odbc:
		return "DbType_odbc"
	default:
		return fmt.Sprintf("-- unknown DbType %d --", dd)
	}
}

// return the type of the database --
func (hdlr *TabServer2Type) GetDbType() DbType {
	// from hdlr -> gCfg -> Global Config
	// conn := sizlib.ConnectToAnyDb(hdlr.DBType, hdlr.PGConn, hdlr.DBName)
	switch hdlr.gCfg.DBType {
	case "postgres":
		return DbType_postgres
	case "Oracle":
		return DbType_Oracle
	case "MsSQL":
		return DbType_MsSql
	case "odbc":
		return DbType_odbc
	case "mySQL", "mariadb":
		return DbType_mySQL
	}
	return DbType_postgres
}

func (hdlr *TabServer2Type) RemapParams(ps *goftlmux.Params, h SQLOne, trx *tr.Trx) {
	if db_RemapParams {
		fmt.Printf("AT: Comment:%s, %s\n", h.Comment, godebug.LF())
	}
	didRemap := false
	for _, vv := range h.ReMapParameter {
		if db_RemapParams {
			fmt.Printf("AT: %s\n", godebug.LF())
		}
		isRes := cfg.ReservedItems[vv.ToName]
		if vv.FromName != vv.ToName && !isRes {
			if ps.HasName(vv.FromName) {
				didRemap = true
				if db_RemapParams {
					fmt.Printf("AT: Remapping! To:%s From:%s isRes=%v, %s\n", vv.ToName, vv.FromName, isRes, godebug.LF())
					trx.AddNote(2, "Remaping Parameter Name "+vv.FromName+" to "+vv.ToName) // xyzzyTRX - announce remap to TRX for testing -- if trxTest --
				}
				s := ps.ByName(vv.FromName)
				goftlmux.AddValueToParams(vv.ToName, s, 'r', goftlmux.FromOther, ps)
			}
		}
	}
	if didRemap {
		if db_RemapParams {
			fmt.Printf("%sParams: AT %s - End of ReMap: %s%s\n", MiscLib.ColorYellow, godebug.LF(), ps.DumpParamTable(), MiscLib.ColorReset)
		}
	}
}

type SesItems struct {
	Path  []string
	Value string
}

type SesDataType struct {
	Set []SesItems
}

func ConvRawSesData(ss string) (rv SesDataType) {
	err := json.Unmarshal([]byte(ss), &rv)
	if err != nil {
		rv.Set = []SesItems{}
	}
	return
}

// Helper func:  Read input from specified file or stdin
func loadData(p string) ([]byte, error) {
	if p == "" {
		return nil, fmt.Errorf("No path specified")
	}

	var rdr io.Reader
	//	if p == "-" {
	//		rdr = os.Stdin
	//	} else if p == "+" {
	//		return []byte("{}"), nil
	//	} else {
	if f, err := os.Open(p); err == nil {
		rdr = f
		defer f.Close()
	} else {
		return nil, err
	}
	//	}
	return ioutil.ReadAll(rdr)
}

// Create, sign, and output a token.  This is a great, simple example of
// how to use this library to create and sign a token.
func SignToken(tokData []byte, keyFile string) (out string, err error) {

	// parse the JSON of the claims
	var claims jwt.MapClaims
	if err = json.Unmarshal(tokData, &claims); err != nil {
		err = fmt.Errorf("Couldn't parse claims JSON: %v", err)
		return
	}

	fmt.Fprintf(os.Stderr, "Siging: %s, AT: %s\n", tokData, godebug.LF())
	fmt.Fprintf(os.Stderr, "Claims: %s, AT: %s\n", godebug.SVarI(claims), godebug.LF())

	//-	// add command line claims
	//-	if len(flagClaims) > 0 {
	//-		for k, v := range flagClaims {
	//-			claims[k] = v
	//-		}
	//-	}

	// get the key
	var key interface{}
	key, err = loadData(keyFile)
	if err != nil {
		err = fmt.Errorf("Couldn't read key: %v", err)
		return
	}

	// get the signing alg
	// alg := jwt.GetSigningMethod(*flagAlg)
	alg := jwt.GetSigningMethod("RS256") // xyzzy - Param
	if alg == nil {
		err = fmt.Errorf("Couldn't find signing method: %v", "RS256") // xyzzy Param
		return
	}

	// create a new token
	token := jwt.NewWithClaims(alg, claims)

	//-	// add command line headers
	//-	if len(flagHead) > 0 {
	//-		for k, v := range flagHead {
	//-			token.Header[k] = v
	//-		}
	//-	}

	if isEs() {
		if k, ok := key.([]byte); !ok {
			err = fmt.Errorf("Couldn't convert key data to key")
			return
		} else {
			key, err = jwt.ParseECPrivateKeyFromPEM(k)
			if err != nil {
				return
			}
		}
	} else if isRs() {
		if k, ok := key.([]byte); !ok {
			err = fmt.Errorf("Couldn't convert key data to key")
			return
		} else {
			key, err = jwt.ParseRSAPrivateKeyFromPEM(k)
			if err != nil {
				return
			}
		}
	}

	if out, err = token.SignedString(key); err == nil {
		if db81 {
			fmt.Println(out)
		}
	} else {
		err = fmt.Errorf("Error signing token: %v", err)
	}

	return
}

func isEs() bool {
	// return strings.HasPrefix(*flagAlg, "ES")
	return false
}

func isRs() bool {
	// return strings.HasPrefix(*flagAlg, "RS")
	return true
}

const db81 = false
const db1 = false
const db2 = false
const db4 = false // Redis Sessions
const db_post = false
const db_get1 = false
const db_RemapParams = false
const db_trace_functions = false
const db_GenUpdateSet = false

const db_DumpInsert = true
const db_DumpDelete = true

// const hdlr.gCfg.DbOn("*", "TabServer2", "db-closure-1") = true
// const hdlr.gCfg.DbOn("*", "TabServer2", "db-closure-2") = true
// const hdlr.gCfg.DbOn("*", "TabServer2", "db-email") = true
// const hdlr.gCfg.DbOn("*", "TabServer2", "db-session") = true
//			if hdlr.gCfg.DbOn("*", "SessionRedis", "db1") {

/* vim: set noai ts=4 sw=4: */
