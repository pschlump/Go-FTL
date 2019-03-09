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
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	logrus "github.com/pschlump/pslog" // "github.com/sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/json" //	"encoding/json"
	"github.com/pschlump/ms"
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
	pptCode := PrePostContinue
	if !done {
		//fmt.Printf ( "At %s\n", godebug.LF() )
		if len(h.CallBefore) > 0 {
			trx.AddNote(1, "Functions to call before running queries.  CallBefore is set.")
			for _, fx_name := range h.CallBefore {
				if !exit {
					trx.AddNote(1, fmt.Sprintf("CallBefore[%s]", fx_name))
					rv, pptCode, exit, a_status = hdlr.CallFunction("before", fx_name, res, req, cfgTag, rv, isError, cookieList, ps, trx)
				}
			}
		}
	}
	// if exit {
	switch pptCode {
	case PrePostRVUpdatedSuccess, PrePostRVUpdatedFail, PrePostFatalSetStatus:
		fmt.Printf("*** exit=%v pptCode=%v from before operations has been signaled ***, rv=%s, %s\n", exit, pptCode, rv, godebug.LF())
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
	pptCode = PrePostContinue
	if len(h.CallAfter) > 0 {
		trx.AddNote(1, "CallAfter is True - functions will be called.")
		for _, fx_name := range h.CallAfter {
			trx.AddNote(1, fmt.Sprintf("CallAfter [%s]", fx_name))
			if !exit {
				fmt.Fprintf(os.Stderr, "CallAfter [%s] rv before ->%s<-\n", fx_name, rv)
				fmt.Fprintf(os.Stdout, "CallAfter [%s] rv before ->%s<-\n", fx_name, rv)
				rv, pptCode, exit, a_status = hdlr.CallFunction("after", fx_name, res, req, cfgTag, rv, isError, cookieList, ps, trx)
				fmt.Fprintf(os.Stderr, "CallAfter exit at bottom rv= %s exit=%v\n", rv, exit)
				fmt.Fprintf(os.Stdout, "CallAfter exit at bottom rv= %s exit=%v\n", rv, exit)
			}
		}
	}
	switch pptCode {
	case PrePostRVUpdatedSuccess, PrePostRVUpdatedFail, PrePostFatalSetStatus:
		fmt.Printf("*** exit=%v pptCode=%v from before operations has been signaled ***, rv=%s, %s\n", exit, pptCode, rv, godebug.LF())
		fmt.Fprintf(os.Stderr, "*** exit=%v pptCode=%v from before operations has been signaled ***, rv=%s, %s\n", exit, pptCode, rv, godebug.LF())
		done = true
		isError = true
		ReturnErrorMessageRv(a_status, rv, "Postprocessing signaled error", "18007",
			fmt.Sprintf(`Error(18007): Postprocessing signaled error. sql-cfg.json[%s] %s`, cfgTag, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
	case PrePostSuccessWriteRV:
		isError = false
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

/* vim: set noai ts=4 sw=4: */
