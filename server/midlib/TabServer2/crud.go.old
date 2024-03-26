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
	"sort"
	"strings"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/json" //	"encoding/json"
)

//	"github.com/pschlump/Authorize.Net"

// =============================================================================================================================================================
// http://localhost:8200/api/list/sql-cfg-files-loaded?dev_auth_token=9abb4f75-f336-46d2-a3af-1115c3d49f14
// =============================================================================================================================================================
// func respHandlerListSqlCfgFilesLoaded(res http.ResponseWriter, req *http.Request) {
// theMux.HandleFunc(api_list+"sql-cfg-files-loaded", closure_respHandlerListSQLCfgFilesLoaded(hdlr)).Methods("GET") //
func closure_respHandlerListSQLCfgFilesLoaded(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select

		if hdlr.gCfg.DbOn("*", "TabServer2", "db-closure-1") {
			fmt.Printf("AT top of closure/respHandlerListSQLCfgFilesLoaded, %s\n", godebug.LF())
		}

		_ /*rw*/, _ /*top_hdlr*/, ps, _ /*err*/ := GetRwPs(res, req)

		// See: ,"dev_auth_token":"9abb4f75-f336-46d2-a3af-1115c3d49f14"

		dev_auth_token := ps.ByNameDflt("dev_auth_token", "")
		ok := (dev_auth_token == hdlr.DevAuthToken)

		if ok {
			fmt.Fprintf(res, "{\"status\":\"success\",\"data\":%s}", sizlib.SVarI(SqlCfgFilesLoaded))
		} else {
			io.WriteString(res, `{"status":"error"}`)
		}
	}
}

// =============================================================================================================================================================
// http://localhost:8200/api/list/cfg-for?item=/api/cart-get-shipping-list&dev_auth_token=9abb4f75-f336-46d2-a3af-1115c3d49f14
// =============================================================================================================================================================
// func respHandlerListCfgFor(res http.ResponseWriter, req *http.Request) {
func closure_respHandlerListCfgFor(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select

		if hdlr.gCfg.DbOn("*", "TabServer2", "db-closure-1") {
			fmt.Printf("AT top of closure/respHandlerListCfgFor, %s\n", godebug.LF())
		}

		_ /*rw*/, _ /*top_hdlr*/, ps, _ /*err*/ := GetRwPs(res, req)

		dev_auth_token := ps.ByName("dev_auth_token")
		item := ps.ByName("item")
		ok := (dev_auth_token == hdlr.DevAuthToken)

		fmt.Printf("dev_auth_token [%s] hdlr.DevAuthToken [%s], %s\n", dev_auth_token, hdlr.DevAuthToken, godebug.LF())

		if ok {
			if v, ok1 := hdlr.SQLCfg[item]; ok1 {
				v.Status = "success"
				io.WriteString(res, sizlib.SVarI(v))
			} else {
				io.WriteString(res, fmt.Sprintf(`{"status":"error","msg":"invalid item", "item":%q}`, item))
			}
		} else {
			io.WriteString(res, `{"status":"error"}`)
		}
	}
}

// =============================================================================================================================================================
// http://localhost:8200/api/list/end-points?dev_auth_token=9abb4f75-f336-46d2-a3af-1115c3d49f14
// =============================================================================================================================================================
// func respHandlerListEndPoints(res http.ResponseWriter, req *http.Request) {
func closure_respHandlerListEndPoints(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select

		if hdlr.gCfg.DbOn("*", "TabServer2", "db-closure-1") {
			fmt.Printf("AT top of closure/respHandlerListEndPoints, %s\n", godebug.LF())
		}

		_ /*rw*/, _ /*top_hdlr*/, ps, _ /*err*/ := GetRwPs(res, req)

		dev_auth_token := ps.ByName("dev_auth_token")
		//ok := true
		//	ok = false
		//}
		ok := (dev_auth_token == hdlr.DevAuthToken)

		if ok {
			var name []string
			for i, _ := range hdlr.SQLCfg {
				if i[0:1] == "/" {
					name = append(name, i)
				}
			}
			sort.Strings(name)
			io.WriteString(res, sizlib.SVarI(name))
		} else {
			io.WriteString(res, `{"status":"error"}`)
		}
	}
}

// =============================================================================================================================================================
// SELECT 1 row by PK - only 1 column PK allowed.
// =============================================================================================================================================================
func closure_respHandlerTableGetPk1(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select

		if hdlr.gCfg.DbOn("*", "TabServer2", "db-closure-1") {
			fmt.Printf("AT top of closure/respHandlerTagbleGetPk1, %s\n", godebug.LF())
		}
		if hdlr.gCfg.DbOn("*", "TabServer2", "db-closure-2") {
			fmt.Fprintf(os.Stderr, "%sAT top of closure/respHandlerTagbleGetPk1, %s%s\n", MiscLib.ColorRed, godebug.LF(), MiscLib.ColorReset)
		}
		var rv string

		fmt.Printf("closure_respHandlerTableGetPk1, %s\n", godebug.LF())

		rw, _ /*top_hdlr*/, ps, _ /*err*/ := GetRwPs(res, req)

		ok, mdata, cfgTag, h, trx, table_name, wc := hdlr.CommmonCrudPrefix(res, req, "select", ps, rw)
		if !ok {
			return
		}

		// xyzzy - should be an "pullIdFromUrl" call - that will find the PK and set the name to that.
		idName, ok := hdlr.GetPkName(h, 1, trx, res, req, ps)
		if !ok {
			return
		}
		id := GetMuxValue("id", idName[0], mdata, trx, res, req, *ps)

		trx.AddNote(1, "Validate Query Parameters")
		err := ValidateQueryParams(ps, h, req) // Validate them
		if err != nil {
			ReturnErrorMessage(406, "Error(12041): Invalid Query Parameter", "12041",
				fmt.Sprintf(`Error(12041): Invalid Query Parameters (%s) sql-cfg.json[%s], %v`, sizlib.EscapeError(err), cfgTag, godebug.LFj()), res, req, *ps, trx, hdlr) // status:error
			return
		}

		data := CommonMakeData(h, trx, ps)

		mdata["cols"] = "*"
		mdata["order_by"] = ""

		hdlr.GenProjectedCols(mdata, h, trx, ps)

		if !hdlr.GenOrderBy(mdata, h, trx, res, req, ps) {
			return
		}

		if ok, _ := hdlr.CommonWhereClause(true, id, true, mdata, &data, h, table_name, trx, wc, res, req, ps); !ok {
			return
		}
		if !hdlr.ExtendedWhereCaluse(mdata, h, &data, trx, wc, res, req, ps) {
			return
		}

		trx.AddNote(1, fmt.Sprintf("cols=%s order_by=%s where=%s", mdata["cols"], mdata["order_by"], mdata["where"]))

		default_tmpl, err := hdlr.GenTemplate(h, "Select-PK1", mdata, wc, &data, false, ps) // "select %{cols%} from \"%{table_name%}\" %{where_where%} %{where%} %{order_by_order_by%} %{order_by%}"
		if err != nil {
			// trx.AddNote(1, fmt.Sprintf("%v", err))
			//rv = fmt.Sprintf(`{"status":"error","msg":"Error(12909): template error (%s) sql-cfg.json[%s]",%s}`,
			//	sizlib.EscapeError(err), cfgTag, godebug.LFj())
			//trx.SetQryDone(rv, "")
			ReturnErrorMessage(406, "Error(12909): template error", "12909",
				fmt.Sprintf(`Error(12909): template error (%s) sql-cfg.json[%s], %v`, sizlib.EscapeError(err), cfgTag, godebug.LFj()), res, req, *ps, trx, hdlr) // status:error
			return
		}
		wc.GenWhereClause(mdata)
		Query := UseTemplate("Select", h.SelectPK1Tmpl, default_tmpl, mdata, trx)

		trx.AddNote(1, "Running .Query")
		s, gotIt := hdlr.HaveCachedData(res, req, h, Query, data...)
		if gotIt {
			trx.AddNote(1, "Data Was In Cache")
			rv = s
			trx.SetCacheData(h.Query, 1, rv, data...)
		} else {
			trx.SetQry(Query, 1, data...)
			// Rows, err := db.Query(Query, data...)
			Rows, err := hdlr.gCfg.Pg_client.Db.Query(Query, data...)
			if err != nil {
				// rv = fmt.Sprintf(`{ "status":"error", "msg":%q, "query":%q, %s }`, err, Query, godebug.LFj())
				// trx.SetQryDone(rv, "")
				// xyzzyErrorReport

				fmt.Fprintf(os.Stderr, "\n%serror %s, %s %s\n", MiscLib.ColorRed, err, godebug.LF(), MiscLib.ColorReset)
				fmt.Fprintf(os.Stdout, "\nError %s, %s\n\n", err, godebug.LF())
				trx.AddNote(1, fmt.Sprintf("Error on Query %s, %s", Query, err))

				ReturnErrorMessage(406, "Error(12910): query error", "12910",
					fmt.Sprintf(`Error(12910): query error (%s) query [%s] %s`, sizlib.EscapeError(err), Query, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
				return
			} else {
				defer Rows.Close()
				fmtData := ps.ByNameDflt("__fmt__", "JSON")
				fmt.Printf("x44: fmtData=%s, %s\n", fmtData, godebug.LF())
				if fmtData == "JSON" {
					// ReturnGetPKAsHashTableName bool         // If ReturnGetPkAsHash is true then either use "data" or if this is true use the "assigned_name"
					// AssignedName               string       // if "", then use TableName if ReturnGetPKAsHashTableName is true.
					if h.ReturnGetPKAsHash {
						rv, _ = sizlib.RowsToJsonFirstRow(Rows)
					} else if h.ReturnAsHash {
						rv, _ = sizlib.RowsToJson(Rows)
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
							rv = fmt.Sprintf(`{"status":"success","data":%s}`, rv)
						}
					} else {
						rv, _ = sizlib.RowsToJson(Rows)
					}
					trx.SetQryDone("", rv)
					hdlr.CacheItForLater(res, req, h, id, rv, Query, data...)
					io.WriteString(res, rv)
					return
				} else if fmtData == "raw" {
					// Row                 map[string]interface{}   //	Single Row Response -- or table header info
					// Table               []map[string]interface{} //	Table of Row Response
					// rw.Raw = Rows
					// rw.State = goftlmux.TableBuffer
					Arr, _, _ := sizlib.RowsToInterface(Rows) // ([]map[string]interface{}, string, int) {

					if h.ReturnGetPKAsHash {
						rv = lib.SVarI(Arr[0])
						_ = rw.WriteRow(Arr[0])
					} else if h.ReturnAsHash {
						rv = lib.SVarI(Arr)
						_ = rw.WriteTable(Arr)
						rv = fmt.Sprintf(`{"status":"success","data":%s}`, rv)
					} else {
						rv = lib.SVarI(Arr)
						_ = rw.WriteTable(Arr)
					}
					trx.SetQryDone("", rv)
					hdlr.CacheItForLater(res, req, h, id, rv, Query, data...)
					return
				} else {
					rv = `{"status":"error","msg":"__fmt__ invalid"}`
				}
			}
		}
		// io.WriteString(res, sizlib.JsonP(rv, res, req))
		io.WriteString(res, rv)
	}
}

// func (this *Trx) SetQry(sql string, depth int, data ...interface{}) {
func pj_Key(as string, a ...interface{}) (rv string) {
	ss := make([]string, 0, len(a)+1)
	ss = append(ss, as)
	for _, tt := range a {
		ss = append(ss, fmt.Sprintf("%s", tt))
	}
	rv = strings.Join(ss, ":")
	return
}

// ====================================================================================================================================================================
// SELECT - GET Request
// ====================================================================================================================================================================
// func respHandlerTableGet(res http.ResponseWriter, req *http.Request) {
func closure_respHandlerTableGet(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select

		if hdlr.gCfg.DbOn("*", "TabServer2", "db-closure-1") {
			fmt.Printf("AT top of closure/respHandlerTableGet, %s\n", godebug.LF())
		}
		if hdlr.gCfg.DbOn("*", "TabServer2", "db-closure-2") {
			fmt.Fprintf(os.Stderr, "%sAT top of closure/respHandlerTableGet, %s%s\n", MiscLib.ColorRed, godebug.LF(), MiscLib.ColorReset)
		}

		fmt.Printf("closure_respHandlerTableGet, %s\n", godebug.LF())

		var rv string
		var s string = ""
		var gotIt bool = false

		rw, _ /*top_hdlr*/, psP, err := GetRwPs(res, req)

		ok, mdata, cfgTag, h, trx, table_name, wc := hdlr.CommmonCrudPrefix(res, req, "select", psP, rw)
		if !ok {
			return
		}

		if sizlib.InArray("dump_closure", h.DebugFlag) {
			fmt.Printf("closure_respHandlerTableGet, %s\n", godebug.LF())
		}

		trx.AddNote(1, "Validate Query Parameters")
		err = ValidateQueryParams(psP, h, req) // Validate them
		if err != nil {
			ReturnErrorMessage(406, "Error(12042): Invalie Query Parameter.", "12042",
				fmt.Sprintf(`Error(12042): Invalid Query Parameters (%s) sql-cfg.json[%s], %v`, sizlib.EscapeError(err), cfgTag, godebug.LFj()),
				res, req, *psP, trx, hdlr) // status:error
			return
		}
		if db_get1 {
			fmt.Printf("where -1 (after common crud prefix) = %s, %s\n", wc, godebug.LF()) // xyzzy
			fmt.Printf("%sParams: AT %s - after CommonCrudPrefix: %s%s\n", MiscLib.ColorGreen, godebug.LF(), psP.DumpParamTable(), MiscLib.ColorReset)
		}

		trx.AddNote(1, "After Query Vlidation")

		data := CommonMakeData(h, trx, psP)
		// xyzzyExtraDataQuery -- implemented by this
		// func AddBindValue(bind *[]interface{}, x interface{}) (pos int) {
		mdata["where_and"] = ""
		//if len(data) > 0 {
		//	mdata["where_and"] = ""
		//}

		mdata["cols"] = "*"
		mdata["order_by"] = ""
		mdata["limit_limit"] = ""
		mdata["limit"] = ""
		mdata["limit_after"] = ""
		mdata["offset_offset"] = ""
		mdata["offset"] = ""
		mdata["offset_after"] = ""
		mdata["before_query"] = ""
		mdata["after_query"] = ""

		hdlr.GenProjectedCols(mdata, h, trx, psP)
		trx.AddNote(1, "After Projected Cols")

		if !hdlr.GenOrderBy(mdata, h, trx, res, req, psP) {
			return
		}
		trx.AddNote(1, "After Order By")

		// isPk := true		// xyzzy need to pass and check this -- Need to collect PK data and put it into 'm' -- need to check that it covers PK
		if !hdlr.CommonLimitOffset(mdata, &data, h, table_name, trx, res, req, psP) {
			return
		}

		var isPk2 bool
		if ok, isPk2 = hdlr.CommonWhereClause(false, "", false, mdata, &data, h, table_name, trx, wc, res, req, psP); !ok {
			return
		}
		if db_get1 {
			fmt.Printf("where 0 = %s, %s\n", wc, godebug.LF()) // xyzzy
		}
		if !hdlr.ExtendedWhereCaluse(mdata, h, &data, trx, wc, res, req, psP) {
			return
		}

		if db_get1 {
			fmt.Printf("where 1 = %s, %s\n", wc, godebug.LF())
		}

		isPk := hdlr.HasPKInWhere(mdata, h, trx, res, req, psP)
		// fmt.Printf("isPk = %v, isPk2 = %v\n", isPk, isPk2) // xyzzy
		isPk = isPk || isPk2

		trx.AddNote(1, fmt.Sprintf("cols=%s order_by=%s where=%s", mdata["cols"], mdata["order_by"], mdata["where"]))

		default_tmpl, err := hdlr.GenTemplate(h, "Select", mdata, wc, &data, false, psP)
		if err != nil {
			ReturnErrorMessage(406, "Error(12909): template error", "12909",
				fmt.Sprintf(`Error(12909): template error (%s) sql-cfg.json[%s], %v`, sizlib.EscapeError(err), cfgTag, godebug.LFj()),
				res, req, *psP, trx, hdlr) // status:error
			return
		}
		wc.GenWhereClause(mdata)
		Query := UseTemplate("Select", h.SelectTmpl, default_tmpl, mdata, trx)

		trx.SetQry(Query, 1, data...)

		// fmt.Printf ( "order_by ->%s<- where ->%s<- h.CacheIt = %s, isPk = %v\n", mdata["order_by"], mdata["where"], h.CacheIt, isPk )
		if (mdata["order_by"] == "" && mdata["where"] == "" && h.CacheIt == "table") || (mdata["order_by"] == "" && mdata["where"] != "" && h.CacheIt == "row" && isPk) {
			s, gotIt = hdlr.HaveCachedDataMk(res, req, h)
			if gotIt {
				trx.AddNote(1, "Data Was In Cache")
				rv = s
				trx.SetCacheData(h.Query, 1, rv, data...)
			}
		}
		fmtData := psP.ByNameDflt("__fmt__", "JSON")
		fmt.Printf("\n%sQuery: %s\nData:%s%s\n\n", MiscLib.ColorGreen, Query, godebug.SVarI(data), MiscLib.ColorReset)
		if !gotIt {
			Rows, err := hdlr.gCfg.Pg_client.Db.Query(Query, data...)
			if err != nil {
				// xyzzyErrorReport

				fmt.Fprintf(os.Stderr, "\n%serror %s, %s %s\n", MiscLib.ColorRed, err, godebug.LF(), MiscLib.ColorReset)
				fmt.Fprintf(os.Stdout, "\nError %s, %s\n\n", err, godebug.LF())
				trx.AddNote(1, fmt.Sprintf("Error on Query %s, %s", Query, err))

				ReturnErrorMessage(406, "Error(18042): Invalie Query Parameter.", "12042",
					fmt.Sprintf(`Error(18042): Invalid Query Parameters (%s) (%s) sql-cfg.json[%s] Query= ->%s<-`, sizlib.EscapeError(err), cfgTag, godebug.LFj(), Query),
					res, req, *psP, trx, hdlr) // status:error
				return
			} else {
				defer Rows.Close()
				if len(h.PostJoin) == 0 { // if we don't have any post join
					// rv, _ = sizlib.RowsToJson(Rows)
					if fmtData == "JSON" {
						if h.ReturnGetPKAsHash && isPk {
							rv, _ = sizlib.RowsToJsonFirstRow(Rows)
						} else {
							rv, _ = sizlib.RowsToJson(Rows)
						}
						trx.SetQryDone("", rv)
					} else if fmtData == "raw" {
						// Row                 map[string]interface{}   //	Single Row Response -- or table header info
						// Table               []map[string]interface{} //	Table of Row Response
						// rw.Raw = Rows
						// rw.State = goftlmux.TableBuffer
						Arr, _, _ := sizlib.RowsToInterface(Rows) // ([]map[string]interface{}, string, int) {

						if h.ReturnGetPKAsHash && isPk {
							rv = lib.SVarI(Arr[0])
							_ = rw.WriteRow(Arr[0])
						} else {
							rv = lib.SVarI(Arr)
							_ = rw.WriteTable(Arr)
						}
						trx.SetQryDone("", rv)
					} else {
						fmtData = "JSON"
						rv = `{"status":"error","msg":"__fmt__ invalid"}`
					}
				} else {
					fmtData = "JSON"
					trx.AddNote(1, "Post Join")
					if db_post_join {
						fmt.Printf("*************** supposed to Post Join at this poin ************************\n")
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
									fmt.Printf("PostJoin: On Post_join[%d]\n", j)
								}
								if vv, ok := v[w.ColName]; ok {
									if db_post_join {
										fmt.Printf("PostJoin: non-null col = ->%v<-\n", vv)
									}
									var data []interface{}
									if len(w.P) > 0 {
										for _, x := range w.P {
											if _, ok2 := v[x]; ok2 {
												data = append(data, v[x])
											} else if t, ok2 := psP.GetByName(x); ok2 {
												// data = append(data, t[0]) // Allows for use of InjectData like $custoemr_id$ derived from AuthToken
												// PJS mod, Mon Nov 21 19:39:36 MST 2016, - found error in adding to .Query - did not test this side.
												data = append(data, t) // Allows for use of InjectData like $custoemr_id$ derived from AuthToken
											} else {
												data = append(data, "")
											}
										}
									} else {
										data = append(data, v[w.ColName])
									}
									// xyzzySubQueriesCached -- Check cache at this point -if- caced then do not increment pj_ and pull from cache
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
					if h.ReturnGetPKAsHash && isPk {
						rvB, err = json.MarshalIndent(rvX[0], "", "\t")
					} else {
						rvB, err = json.MarshalIndent(rvX, "", "\t")
					}
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
				// xyzzy824 - Problem - table caching should be per-customer_id - so can cache tblCategoies - and send back a 304 response.
				if (mdata["where"] == "" && h.CacheIt == "table") || (mdata["where"] != "" && h.CacheIt == "row" && isPk) {
					hdlr.CacheItForLaterMk(res, req, h, rv)
				}
			} // end-if has PostJoin
		}
		// fmt.Printf("AtAT: %s\n", godebug.LF())
		if h.ReturnMeta {
			// fmt.Printf("AtAT: %s\n", godebug.LF())
			trx.AddNote(1, "Note: meta data return format.")
			default_tmpl_cnt, err := hdlr.GenTemplate(h, "Select-Count", mdata, wc, &data, true, psP) //"select %{cols%} from \"%{table_name%}\" %{where_where%} %{where%} %{order_by_order_by%} %{order_by%} %{limit_limit%} %{limit%} %{offset_offset%} %{offset%}"
			if err != nil {
				// trx.AddNote(1, fmt.Sprintf("%v", err))
				// rv = fmt.Sprintf(`{"status":"error","msg":"Error(12909): template error (%s) sql-cfg.json[%s]",%s}`,
				// 	sizlib.EscapeError(err), cfgTag, godebug.LFj())
				// trx.SetQryDone(rv, "")
				ReturnErrorMessage(406, "Error(12909): template error", "12909",
					fmt.Sprintf(`Error(12909): template error (%s) sql-cfg.json[%s], %v`, sizlib.EscapeError(err), cfgTag, godebug.LFj()),
					res, req, *psP, trx, hdlr) // status:error
				return
			}
			CntQuery := UseTemplate("Select-Count", h.SelectCountTmpl, default_tmpl_cnt, mdata, trx)
			// fmt.Printf ( "q=%s, data=%v\n", CntQuery, data )
			RowsCnt, errCnt := hdlr.gCfg.Pg_client.Db.Query(CntQuery, data...)
			if errCnt != nil {
				// xyzzyErrorReport -- err

				fmt.Fprintf(os.Stderr, "\n%serror %s, %s %s\n", MiscLib.ColorRed, errCnt, godebug.LF(), MiscLib.ColorReset)
				fmt.Fprintf(os.Stdout, "\nError %s, %s\n\n", errCnt, godebug.LF())
				trx.AddNote(1, fmt.Sprintf("Error on Query %s, %s", Query, errCnt))

				// rv = fmt.Sprintf(`{ "status":"error", "msg":%q, "query":%q, %s }`, errCnt, CntQuery, godebug.LFj())
				ReturnErrorMessage(406, "Error(12910): query error", "12910",
					fmt.Sprintf(`Error(12910): query error (%s) query [%s] %s`, sizlib.EscapeError(errCnt), CntQuery, godebug.LF()),
					res, req, *psP, trx, hdlr) // status:error
				return
			} else {
				defer RowsCnt.Close()
				meta, _ := sizlib.RowsToJson(RowsCnt)
				// fmt.Printf ( "meta=%s q=%s\n", meta, CntQuery )
				rv = fmt.Sprintf("{\"meta_flag\":\"meta-data\",\"data\":%s,\"meta\":{\"count\":%s}}", rv, meta)
			}
		} else if h.ReturnAsHash {
			// fmt.Printf("AtAT: %s\n", godebug.LF())
			fmt.Printf("x44:a: fmtData=%s, %s\n", fmtData, godebug.LF())
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
				//	rv = fmt.Sprintf("{\"status\":\"success\",\"data\":%s}", rv)
				rv = fmt.Sprintf(`{"status":"success","data":%s}`, rv)
			}
		}
		if fmtData != "raw" {
			io.WriteString(res, rv)
		}

	} // end of closure -- return
}

// ====================================================================================================================================================================
// SELECT - GET Count Request
// ====================================================================================================================================================================
// func respHandlerTableGetCount(res http.ResponseWriter, req *http.Request) {
func closure_respHandlerTableGetCount(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select
		var rv string
		var s string = ""
		var gotIt bool = false

		if hdlr.gCfg.DbOn("*", "TabServer2", "db-closure-1") {
			fmt.Printf("AT top of closure/respHandlerTableGetCount, %s\n", godebug.LF())
		}

		rw, _ /*top_hdlr*/, psP, err := GetRwPs(res, req)

		ok, mdata, cfgTag, h, trx, table_name, wc := hdlr.CommmonCrudPrefix(res, req, "select", psP, rw)
		if !ok {
			return
		}

		trx.AddNote(1, "Validate Query Parameters/Count")
		err = ValidateQueryParams(psP, h, req) // Validate them
		if err != nil {
			// trx.AddNote(1, "Failed To Validate Query Parameters")
			// rv = fmt.Sprintf(`{"status":"error","msg":"Error(12042): Invalid Query Parameters (%s) sql-cfg.json[%s]",%s}`,
			// 	sizlib.EscapeError(err), cfgTag, godebug.LFj())
			// trx.SetQryDone(rv, "")
			ReturnErrorMessage(406, "Error(12042): Invalid Parameter", "12042",
				fmt.Sprintf(`Error(12042): Invalid Query Parameters (%s),%s`, sizlib.EscapeError(err), godebug.LFj()),
				res, req, *psP, trx, hdlr) // status:error
			return
		}

		trx.AddNote(1, "After Query Vlidation/Count")

		data := CommonMakeData(h, trx, psP)

		var isPk2 bool
		if ok, isPk2 = hdlr.CommonWhereClause(false, "", false, mdata, &data, h, table_name, trx, wc, res, req, psP); !ok {
			return
		}
		if !hdlr.ExtendedWhereCaluse(mdata, h, &data, trx, wc, res, req, psP) {
			return
		}

		isPk := hdlr.HasPKInWhere(mdata, h, trx, res, req, psP)
		isPk = isPk || isPk2

		trx.AddNote(1, fmt.Sprintf("cols=%s order_by=%s where=%s", mdata["cols"], mdata["order_by"], mdata["where"]))

		default_tmpl, err := hdlr.GenTemplate(h, "Select-Count", mdata, wc, &data, false, psP) // "select count(*) as \"nRows\" from \"%{table_name%}\" %{where_where%} %{where%}"
		if err != nil {
			// trx.AddNote(1, fmt.Sprintf("%v", err))
			// rv = fmt.Sprintf(`{"status":"error","msg":"Error(12909): template error (%s) sql-cfg.json[%s]",%s}`,
			// 	sizlib.EscapeError(err), cfgTag, godebug.LFj())
			// trx.SetQryDone(rv, "")
			ReturnErrorMessage(406, "Error(12909): template error", "12909",
				fmt.Sprintf(`Error(12909): template error (%s) sql-cfg.json[%s] %s, %v`, sizlib.EscapeError(err), cfgTag, err, godebug.LFj()),
				res, req, *psP, trx, hdlr) // status:error
			return
		}
		wc.GenWhereClause(mdata)
		Query := UseTemplate("Select", h.SelectCountTmpl, default_tmpl, mdata, trx)

		trx.SetQry(Query, 1, data...)

		// fmt.Printf ( "order_by ->%s<- where ->%s<- h.CacheIt = %s, isPk = %v\n", mdata["order_by"], mdata["where"], h.CacheIt, isPk )
		if (mdata["order_by"] == "" && mdata["where"] == "" && h.CacheIt == "table") || (mdata["order_by"] == "" && mdata["where"] != "" && h.CacheIt == "row" && isPk) {
			s, gotIt = hdlr.HaveCachedDataMk(res, req, h)
			if gotIt {
				trx.AddNote(1, "Data Was In Cache")
				rv = s
				trx.SetCacheData(h.Query, 1, rv, data...)
			}
		}
		if !gotIt {
			Rows, err := hdlr.gCfg.Pg_client.Db.Query(Query, data...)
			if err != nil {
				// xyzzyErrorReport -- err

				fmt.Fprintf(os.Stderr, "\n%serror %s, %s %s\n", MiscLib.ColorRed, err, godebug.LF(), MiscLib.ColorReset)
				fmt.Fprintf(os.Stdout, "\nError %s, %s\n\n", err, godebug.LF())
				trx.AddNote(1, fmt.Sprintf("Error on Query %s, %s", Query, err))

				// rv = fmt.Sprintf(`{ "status":"error", "msg":%q, "query":%q, %s }`, err, Query, godebug.LFj())
				ReturnErrorMessage(406, "Error(12910): query error", "12910",
					fmt.Sprintf(`Error(12910): query error (%s) query [%s] %s %s`, sizlib.EscapeError(err), Query, err, godebug.LF()),
					res, req, *psP, trx, hdlr) // status:error
				return
			} else {
				defer Rows.Close()
				rv, _ = sizlib.RowsToJson(Rows)
				trx.SetQryDone("", rv)
				if (mdata["where"] == "" && h.CacheIt == "table") || (mdata["where"] != "" && h.CacheIt == "row" && isPk) {
					hdlr.CacheItForLaterMk(res, req, h, rv)
				}
			}
		}
		// io.WriteString(res, sizlib.JsonP(rv, res, req))
		io.WriteString(res, rv)
	}
}

// ====================================================================================================================================================================
// UPDATE
// ====================================================================================================================================================================
// 1. Failed to use bind variables in "set"
// 2. auth_token should be changed to $auth_token$ in all requests - leaving auth_token as a field in a table.
// 3. need to do data-validation on 'm'
// 4. need to pull data from URL/Mux -> 'm'
// func respHandlerTablePut(res http.ResponseWriter, req *http.Request) { // Update ( or insert )
func closure_respHandlerTablePut(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select

		var rv string

		rw, _ /*top_hdlr*/, psP, err := GetRwPs(res, req)

		ok, mdata, cfgTag, h, trx, table_name, wc := hdlr.CommmonCrudPrefix(res, req, "update", psP, rw)
		if !ok {
			fmt.Printf("Invalid Query Parameters, **406**, %s\n", godebug.LF())
			return
		}

		trx.AddNote(1, "Validate Query Parameters")
		err = ValidateQueryParams(psP, h, req) // Validate them
		if err != nil {
			fmt.Printf("Invalid Query Parameters, 406, %s, %s\n", err, godebug.LF())
			ReturnErrorMessage(406, "Invalid Parameter", "12043",
				fmt.Sprintf(`Error(12043): Invalid Query Parameters (%s) sql-cfg.json[%s] %s %s`, sizlib.EscapeError(err), cfgTag, err, godebug.LF()),
				res, req, *psP, trx, hdlr) // status:error
			return
		}

		data := CommonMakeData(h, trx, psP)

		mdata["updcols"] = ""
		mdata["where"] = ""

		ok = hdlr.GenUpdateSet(h, trx, &data, mdata, res, req, psP)
		if !ok {
			return
		}

		if ok, _ = hdlr.CommonWhereClause(true, "", false, mdata, &data, h, table_name, trx, wc, res, req, psP); !ok {
			return
		}

		if !hdlr.ExtendedWhereCaluse(mdata, h, &data, trx, wc, res, req, psP) {
			return
		}

		default_tmpl, err := hdlr.GenTemplate(h, "Update", mdata, wc, &data, false, psP) // "update \"%{table_name%}\" %{updcols%} %{where_where%} %{where%}"
		if err != nil {
			fmt.Printf("Invalid Query Parameters, 406, %s, %s\n", err, godebug.LF())
			trx.AddNote(1, fmt.Sprintf("%v", err))
			ReturnErrorMessage(406, "Error(12909): template error", "12909",
				fmt.Sprintf(`Error(12909): template error (%s) sql-cfg.json[%s] %s, %v`, sizlib.EscapeError(err), cfgTag, err, godebug.LFj()),
				res, req, *psP, trx, hdlr) // status:error
			return
		}
		wc.GenWhereClause(mdata)
		Query := UseTemplate("Update", h.UpdateTmpl, default_tmpl, mdata, trx)

		fmt.Printf("Query: %s\n\tData:%s\n\tAT:%s\n", Query, sizlib.SVar(data), godebug.LF())

		err = sizlib.Run1Thx(hdlr.gCfg.Pg_client.Db, trx, Query, data...)
		if err != nil {
			// xyzzyErrorReport -- err

			fmt.Fprintf(os.Stderr, "\n%serror %s, %s %s\n", MiscLib.ColorRed, err, godebug.LF(), MiscLib.ColorReset)
			fmt.Fprintf(os.Stdout, "\nError %s, %s\n\n", err, godebug.LF())
			trx.AddNote(1, fmt.Sprintf("Error on Query %s, %s", Query, err))

			rv = fmt.Sprintf(`{ "status":"error", "msg":%q, %s }`, err, godebug.LFj())
			trx.SetQryDone(rv, "")
		} else {
			rv = `{ "status":"success", "x1":2 }`
			trx.SetQryDone("", rv)
		}
		// io.WriteString(res, sizlib.JsonP(rv, res, req))
		io.WriteString(res, rv)
	}
}

// ====================================================================================================================================================================
// INSERT
// ====================================================================================================================================================================
// func respHandlerTablePost(res http.ResponseWriter, req *http.Request) { // Insert ( or update )
func closure_respHandlerTablePost(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select

		fmt.Fprintf(os.Stdout, "\nIn Insert\n")
		fmt.Fprintf(os.Stderr, "\nIn Insert\n")

		var s, rv, s_id, x_id string
		s_id, x_id = "", ""

		rw, _ /*top_hdlr*/, psP, err := GetRwPs(res, req)

		ok, mdata, cfgTag, h, trx, _ /*table_name*/, wc := hdlr.CommmonCrudPrefix(res, req, "insert", psP, rw)
		if !ok {
			return
		}

		if db_post || sizlib.InArray("dump_insert_params", h.DebugFlag) {
			fmt.Printf("Insert AT: %s Params: %s\n", godebug.LF(), psP.DumpParamTable())
		}

		// fmt.Printf ( "5598: cfgTag=%s LineNo=%s m=%s\n", cfgTag, h.LineNo, sizlib.SVarI(m) )
		trx.AddNote(1, "Validate Query Parameters")
		err = ValidateQueryParams(psP, h, req) // Validate them
		if err != nil {
			// trx.AddNote(1, "Failed To Validate Query Parameters")
			// func ReturnErrorMessage(status int, msg string, code string, details string, res http.ResponseWriter, req *http.Request, ps goftlmux.Params) {
			// rv = fmt.Sprintf(`{"status":"error","msg":"Error(12045): Invalid Query Parameters (%s) sql-cfg.json[%s]",%s}`,
			// 	sizlib.EscapeError(err), cfgTag, godebug.LFj())
			// trx.SetQryDone(rv, "")
			ReturnErrorMessage(406, "Invalid Parameter", "12043",
				fmt.Sprintf(`Error(12043): Invalid Query Parameters (%s) sql-cfg.json[%s] %s %s`, sizlib.EscapeError(err), cfgTag, err, godebug.LF()),
				res, req, *psP, trx, hdlr) // status:error
			return
		}
		// fmt.Printf ( "5609: m=%s\n", sizlib.SVarI(m) )

		data := CommonMakeData(h, trx, psP)

		mdata["cols"] = ""
		mdata["vals"] = ""

		generateId := false
		idName := ""
		haveId := false
		// fmt.Printf ( "5616: m=%s, h.Cols=%s\n", sizlib.SVarI(m), sizlib.SVarI(h.Cols) )
		n_col := 0
		if len(h.Cols) > 0 {
			s = ""
			com := ""
			for ii, v := range h.Cols {
				if db_post || sizlib.InArray("db_insert", h.DebugFlag) {
					fmt.Printf("Column[%s] pos %d\n", v.ColName, ii)
				}
				colName := v.ColName
				if v.Insert {
					// fmt.Printf ( " v.Insert is true, " );
					if v.DataColName != "" {
						colName = v.DataColName
					}
					if db_post || sizlib.InArray("db_insert", h.DebugFlag) {
						fmt.Printf("Insert colsName=[%s], v.ColName=[%s] v.DataColName=[%s]\n", colName, v.ColName, v.DataColName)
					}
					if psP.HasName(colName) {
						if db_post || sizlib.InArray("db_insert", h.DebugFlag) {
							fmt.Printf("Found [%s] in the params\n", colName)
						}
						s = s + com + DbBeginQuote + v.ColName + DbEndQuote
						com = ","
						n_col++
						if v.IsPk {
							x_id = psP.ByNameDflt(colName, "")
							haveId = true
							if sizlib.InArray("db_insert", h.DebugFlag) {
								fmt.Printf("Found [%s] value is %s\n", colName, x_id)
							}
						}
					} else if v.AutoGen {
						if db_post || sizlib.InArray("db_insert", h.DebugFlag) {
							fmt.Printf("Found [%s] v.AutoGen\n", colName)
						}
						s = s + com + DbBeginQuote + v.ColName + DbEndQuote
						com = ","
						n_col++
						if v.IsPk {
							x_id = psP.ByNameDflt(colName, "")
							if sizlib.InArray("db_insert", h.DebugFlag) {
								fmt.Printf("Found [%s] value is %s\n", colName, x_id)
							}
						}
					} else if v.DefaultData != "" {
						if db_post || sizlib.InArray("db_insert", h.DebugFlag) {
							fmt.Printf("Found [%s] in as a default for insert\n", colName)
						}
						s = s + com + DbBeginQuote + v.ColName + DbEndQuote
						com = ","
						n_col++
						if v.IsPk {
							x_id = psP.ByNameDflt(colName, v.DefaultData)
							haveId = true
							if sizlib.InArray("db_insert", h.DebugFlag) {
								fmt.Printf("Found [%s] value is %s\n", colName, x_id)
							}
						}
					}
				} else if v.IsPk {
					idName = colName
					if v.DataColName != "" {
						idName = v.DataColName
					}
					if v.AutoGen {
						generateId = true
						// xyzzyOracle - if !haveId && hdlr.GetDbType() == DbType_Oracle {
						// xyzzyOracle - use v.OracleSequenceName to select ID and save it into x_id, set haveId to true - work out details
					}
				} else {
					if psP.HasName(colName) {
						fmt.Printf(`{"type":"not-used", "ColName": "%s", "note":"value supplied for column but not used because not an insert-able column in specification", "LineFile":%q }`+"\n", colName, godebug.LF())
						trx.AddNote(1, fmt.Sprintf(`Column [%s] Not Used: value supplied for column but not used because not an insert-able column in specification Line:%s`, colName, godebug.LF()))
					}
				}
			}
			if db_post || sizlib.InArray("db_insert", h.DebugFlag) {
				fmt.Printf("Columns to insert are %s, %s\n", s, godebug.LF())
			}
			if len(h.CustomerIdPart.ColName) > 0 {
				s = s + com + DbBeginQuote + h.CustomerIdPart.ColName + DbEndQuote
				com = ","
			}
			mdata["cols"] = s
			// fmt.Printf ( "s=%s\n", s )
		}
		if db_post || sizlib.InArray("db_insert", h.DebugFlag) {
			fmt.Printf("Insert mdata=%s\n", godebug.SVar(mdata))
		}

		ok, s_id = hdlr.GenInsertValues(h, trx, &data, mdata, n_col, res, req, psP)
		if !ok {
			return
		}

		if s_id == "" && x_id != "" {
			if db_post || sizlib.InArray("db_insert", h.DebugFlag) {
				fmt.Printf("Using passed ID value of x_id=%s\n", x_id)
			}
			s_id = x_id
		}

		default_tmpl, err := hdlr.GenTemplate(h, "Insert", mdata, wc, &data, false, psP) // "insert into \"%{table_name%}\" ( %{cols%} ) values ( %{vals%} )"
		if err != nil {
			// trx.AddNote(1, fmt.Sprintf("%v", err))
			// func ReturnErrorMessage(status int, msg string, code string, details string, res http.ResponseWriter, req *http.Request, ps goftlmux.Params) {
			// rv = fmt.Sprintf(`{"status":"error","msg":"Error(12909): template error (%s) sql-cfg.json[%s]",%s}`,
			// 	sizlib.EscapeError(err), cfgTag, godebug.LFj())
			// trx.SetQryDone(rv, "")
			ReturnErrorMessage(406, "Error(12909): template error", "12909",
				fmt.Sprintf(`Error(12909): template error (%s) sql-cfg.json[%s] %s %s`, sizlib.EscapeError(err), cfgTag, err, godebug.LF()),
				res, req, *psP, trx, hdlr) // status:error
			return
		}
		wc.GenWhereClause(mdata)
		Query := UseTemplate("Insert", h.InsertTmpl, default_tmpl, mdata, trx)

		// INSERT INTO persons (lastname,firstname) VALUES ('Smith', 'John') RETURNING id;
		if generateId && !haveId && hdlr.GetDbType() == DbType_postgres {
			// fmt.Printf("AT: %s\n", godebug.LF())
			Query += fmt.Sprintf(" returning %q ", idName)
			if db_post || db_DumpInsert {
				fmt.Printf("\nQuery: %s\n\tData:%s\n\tAT:%s\n\n", Query, tr.SVar(data), godebug.LF())
			}
			err, s_id = sizlib.Run1IdThx(hdlr.gCfg.Pg_client.Db, trx, Query, data...)
		} else {
			// fmt.Printf("AT: %s\n", godebug.LF())
			if db_post || db_DumpInsert {
				fmt.Printf("\n%sQuery: %s\n\tData:%s\n%s\tAT:%s\n\n", MiscLib.ColorGreen, Query, tr.SVar(data), MiscLib.ColorReset, godebug.LF())
			}
			err = sizlib.Run1Thx(hdlr.gCfg.Pg_client.Db, trx, Query, data...)
		}
		if err != nil {

			fmt.Fprintf(os.Stderr, "\n%serror %s, %s %s\n", MiscLib.ColorRed, err, godebug.LF(), MiscLib.ColorReset)
			fmt.Fprintf(os.Stdout, "\nError %s, %s\n\n", err, godebug.LF())
			trx.AddNote(1, fmt.Sprintf("Error on Query %s, %s", Query, err))

			// xyzzyErrorReport -- err

			fmt.Fprintf(os.Stderr, "\n%serror %s, %s %s\n", MiscLib.ColorRed, err, godebug.LF(), MiscLib.ColorReset)
			fmt.Fprintf(os.Stdout, "\nError %s, %s\n\n", err, godebug.LF())
			trx.AddNote(1, fmt.Sprintf("Error on Query %s, %s", Query, err))

			// func ReturnErrorMessage(status int, msg string, code string, details string, res http.ResponseWriter, req *http.Request, ps goftlmux.Params) {
			rv = fmt.Sprintf(`{ "status":"error", "msg":%q, %s, "query":%q }`, err, godebug.LFj(), Query)
			trx.SetQryDone(rv, "")

		} else {
			rv = fmt.Sprintf(`{ "status":"success", "n_rows_inserted":1, "id":%q }`, s_id)
			trx.SetQryDone("", rv)
		}
		// io.WriteString(res, sizlib.JsonP(rv, res, req))
		io.WriteString(res, rv)
	}
}

// func respHandlerTablePostPk1(res http.ResponseWriter, req *http.Request) { // Insert ( or update )
func closure_respHandlerTablePostPk1(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select
		var s, rv, s_id string
		s_id = ""

		rw, _ /*top_hdlr*/, psP, err := GetRwPs(res, req)

		ok, mdata, cfgTag, h, trx, _ /*table_name*/, wc := hdlr.CommmonCrudPrefix(res, req, "insert", psP, rw)
		if !ok {
			return
		}

		if db_post || sizlib.InArray("dump_params", h.DebugFlag) {
			fmt.Printf("Insert AT: %s Params: %s\n", godebug.LF(), psP.DumpParamTable())
		}

		idName, ok := hdlr.GetPkName(h, 1, trx, res, req, psP)
		if !ok {
			return
		}
		_ = GetMuxValue("id", idName[0], mdata, trx, res, req, *psP)

		// fmt.Printf ( "5598: cfgTag=%s LineNo=%s m=%s\n", cfgTag, h.LineNo, sizlib.SVarI(m) )
		trx.AddNote(1, "Validate Query Parameters")
		err = ValidateQueryParams(psP, h, req) // Validate them
		if err != nil {
			// trx.AddNote(1, "Failed To Validate Query Parameters")
			// rv = fmt.Sprintf(`{"status":"error","msg":"Error(12045): Invalid Query Parameters (%s) sql-cfg.json[%s]",%s}`,
			// 	sizlib.EscapeError(err), cfgTag, godebug.LFj())
			// trx.SetQryDone(rv, "")
			ReturnErrorMessage(406, "Invalid Parameter", "12043",
				fmt.Sprintf(`Error(12043): Invalid Query Parameters (%s) sql-cfg.json[%s] %s %s`, sizlib.EscapeError(err), cfgTag, err, godebug.LF()),
				res, req, *psP, trx, hdlr) // status:error
			return
		}
		// fmt.Printf ( "5609: m=%s\n", sizlib.SVarI(m) )

		data := CommonMakeData(h, trx, psP)

		mdata["cols"] = ""
		mdata["vals"] = ""

		generateId := false
		haveId := false
		idName0 := ""

		// fmt.Printf ( "5616: m=%s, h.Cols=%s\n", sizlib.SVarI(m), sizlib.SVarI(h.Cols) )
		n_col := 0
		if len(h.Cols) > 0 {
			s = ""
			com := ""
			for _, v := range h.Cols {
				// fmt.Printf ( "Column[%s]", v.ColName )
				colName := v.ColName
				if v.DataColName != "" {
					colName = v.DataColName
				}
				if v.Insert {
					// fmt.Printf ( " v.Insert is true, " );
					// xyzzy - if v.DataColName != "", then psP.GetByName ( v.DataColName ) else ...
					if psP.HasName(colName) {
						// fmt.Printf ( " is in the input data, " );
						s = s + com + "\"" + v.ColName + "\""
						com = ","
						n_col++
					} else if v.AutoGen {
						// fmt.Printf ( " is AutoGen, " );
						s = s + com + "\"" + v.ColName + "\""
						com = ","
						n_col++
					}
				}
				if v.IsPk {
					idName0 = colName
					if v.AutoGen {
						generateId = true
						// xyzzyOracle - if !haveId && hdlr.GetDbType() == DbType_Oracle {
						// xyzzyOracle - use v.OracleSequenceName to select ID and save it into x_id, set haveId to true - work out details
					}
				}
				// fmt.Printf ( "\n" )
			}
			if len(h.CustomerIdPart.ColName) > 0 {
				s = s + com + "\"" + h.CustomerIdPart.ColName + "\""
				com = ","
				// n_col++
			}
			mdata["cols"] = s
			// fmt.Printf ( "s=%s\n", s )
		}
		if db_post || sizlib.InArray("dump_params", h.DebugFlag) {
			fmt.Printf("Columns to insert (PK1) are %s, %s\n", s, godebug.LF())
		}

		ok, s_id = hdlr.GenInsertValues(h, trx, &data, mdata, n_col, res, req, psP)
		if !ok {
			return
		}

		default_tmpl, err := hdlr.GenTemplate(h, "Insert", mdata, wc, &data, false, psP) // "insert into \"%{table_name%}\" ( %{cols%} ) values ( %{vals%} )"
		if err != nil {
			// trx.AddNote(1, fmt.Sprintf("%v", err))
			// func ReturnErrorMessage(status int, msg string, code string, details string, res http.ResponseWriter, req *http.Request, ps goftlmux.Params) {
			// rv = fmt.Sprintf(`{"status":"error","msg":"Error(12909): template error (%s) sql-cfg.json[%s]",%s}`,
			// 	sizlib.EscapeError(err), cfgTag, godebug.LFj())
			// trx.SetQryDone(rv, "")
			ReturnErrorMessage(406, "Error(12909): template error", "12909",
				fmt.Sprintf(`Error(12909): template error (%s) sql-cfg.json[%s] %s %s`, sizlib.EscapeError(err), cfgTag, err, godebug.LFj()),
				res, req, *psP, trx, hdlr) // status:error
			return
		}
		wc.GenWhereClause(mdata)
		Query := UseTemplate("Insert", h.InsertTmpl, default_tmpl, mdata, trx)

		if generateId && !haveId && hdlr.GetDbType() == DbType_postgres {
			// fmt.Printf("AT: %s\n", godebug.LF())
			Query += fmt.Sprintf(" returning %q ", idName0)
			if db_post || db_DumpInsert {
				fmt.Printf("Query: %s\n\tData:%s\n\tAT:%s\n", Query, tr.SVar(data), godebug.LF())
			}
			err, s_id = sizlib.Run1IdThx(hdlr.gCfg.Pg_client.Db, trx, Query, data...)
		} else {
			err = sizlib.Run1Thx(hdlr.gCfg.Pg_client.Db, trx, Query, data...)
		}
		if err != nil {
			// func ReturnErrorMessage(status int, msg string, code string, details string, res http.ResponseWriter, req *http.Request, ps goftlmux.Params) {

			// xyzzyErrorReport -- err

			fmt.Fprintf(os.Stderr, "\n%serror %s, %s %s\n", MiscLib.ColorRed, err, godebug.LF(), MiscLib.ColorReset)
			fmt.Fprintf(os.Stdout, "\nError %s, %s\n\n", err, godebug.LF())
			trx.AddNote(1, fmt.Sprintf("Error on Query %s, %s", Query, err))

			rv = fmt.Sprintf(`{ "status":"error", "msg":%q, %s }`, err, godebug.LFj())
			trx.SetQryDone(rv, "")
		} else {
			rv = fmt.Sprintf(`{ "status":"success", "n_rows_inserted":1, "id":%q }`, s_id)
			trx.SetQryDone("", rv)
		}
		// io.WriteString(res, sizlib.JsonP(rv, res, req))
		io.WriteString(res, rv)
	}
}

// ====================================================================================================================================================================
// DELETE
// ====================================================================================================================================================================
// func respHandlerTableDelPk1(res http.ResponseWriter, req *http.Request) {
func closure_respHandlerTableDelPk1(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select
		var rv string

		rw, _ /*top_hdlr*/, psP, err := GetRwPs(res, req)

		ok, mdata, cfgTag, h, trx, table_name, wc := hdlr.CommmonCrudPrefix(res, req, "delete", psP, rw)
		if !ok {
			return
		}

		// xyzzy - should be an "pullIdFromUrl" call - that will find the PK and set the name to that.
		idName, ok := hdlr.GetPkName(h, 1, trx, res, req, psP)
		if !ok {
			return
		}
		id := GetMuxValue("id", idName[0], mdata, trx, res, req, *psP)

		trx.AddNote(1, "Validate Query Parameters")
		err = ValidateQueryParams(psP, h, req) // Validate them
		if err != nil {
			ReturnErrorMessage(406, "Invalid Parameter", "12043",
				fmt.Sprintf(`Error(12043): Invalid Query Parameters (%s) sql-cfg.json[%s] %s %s`, sizlib.EscapeError(err), cfgTag, err, godebug.LF()),
				res, req, *psP, trx, hdlr) // status:error
			return
		}

		data := CommonMakeData(h, trx, psP)

		if ok, _ = hdlr.CommonWhereClause(true, id, true, mdata, &data, h, table_name, trx, wc, res, req, psP); !ok {
			return
		}

		if !hdlr.ExtendedWhereCaluse(mdata, h, &data, trx, wc, res, req, psP) {
			return
		}

		default_tmpl, err := hdlr.GenTemplate(h, "Delete-PK1", mdata, wc, &data, false, psP) // "delete from \"%{table_name%}\" %{where_where%} %{where%}", mdata, trx )
		if err != nil {
			ReturnErrorMessage(406, "Error(12909): template error", "12909",
				fmt.Sprintf(`Error(12909): template error (%s) sql-cfg.json[%s] %s %s`, sizlib.EscapeError(err), cfgTag, err, godebug.LFj()),
				res, req, *psP, trx, hdlr) // status:error
			return
		}
		wc.GenWhereClause(mdata)
		Query := UseTemplate("Delete", h.DeleteTmpl, default_tmpl, mdata, trx)

		err = sizlib.Run1Thx(hdlr.gCfg.Pg_client.Db, trx, Query, data...)
		if err != nil {
			// xyzzyErrorReport -- err

			fmt.Fprintf(os.Stderr, "\n%serror %s, %s %s\n", MiscLib.ColorRed, err, godebug.LF(), MiscLib.ColorReset)
			fmt.Fprintf(os.Stdout, "\nError %s, %s\n\n", err, godebug.LF())
			trx.AddNote(1, fmt.Sprintf("Error on Query %s, %s", Query, err))

			ReturnErrorMessage(406, "Error(12910): query error", "12910",
				fmt.Sprintf(`Error(12910): query error (%s) query [%s] %s %s`, sizlib.EscapeError(err), Query, err, godebug.LF()),
				res, req, *psP, trx, hdlr) // status:error
			return
		} else {
			rv = `{ "status":"success" }`
			trx.SetQryDone("", rv)
		}
		// io.WriteString(res, sizlib.JsonP(rv, res, req))
		io.WriteString(res, rv)

	}
}

// ====================================================================================================================================================================
// UPDATE
// ====================================================================================================================================================================
// 1. Failed to use bind variables in "set"
// 2. auth_token should be changed to $auth_token$ in all requests - leaving auth_token as a field in a table.
// 3. need to do data-validation on 'm'
// 4. need to pull data from URL/Mux -> 'm'
// func respHandlerTablePutPk1(res http.ResponseWriter, req *http.Request) { // Update ( or insert )
func closure_respHandlerTablePutPk1(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select

		var rv string

		rw, _ /*top_hdlr*/, psP, err := GetRwPs(res, req)

		ok, mdata, cfgTag, h, trx, table_name, wc := hdlr.CommmonCrudPrefix(res, req, "update", psP, rw)
		if !ok {
			return
		}

		// xyzzy - should be an "pullIdFromUrl" call - that will find the PK and set the name to that.
		idName, ok := hdlr.GetPkName(h, 1, trx, res, req, psP)
		if !ok {
			return
		}
		id := GetMuxValue("id", idName[0], mdata, trx, res, req, *psP)

		trx.AddNote(1, "Validate Query Parameters")
		err = ValidateQueryParams(psP, h, req) // Validate them
		if err != nil {
			ReturnErrorMessage(406, "Invalid Parameter", "12043",
				fmt.Sprintf(`Error(12043): Invalid Query Parameters (%s) sql-cfg.json[%s] %s %s`, sizlib.EscapeError(err), cfgTag, err, godebug.LF()),
				res, req, *psP, trx, hdlr) // status:error
			return
		}

		data := CommonMakeData(h, trx, psP)

		mdata["updcols"] = ""
		mdata["where"] = ""

		ok = hdlr.GenUpdateSet(h, trx, &data, mdata, res, req, psP)
		if !ok {
			return
		}

		if ok, _ = hdlr.CommonWhereClause(true, id, true, mdata, &data, h, table_name, trx, wc, res, req, psP); !ok {
			return
		}

		if !hdlr.ExtendedWhereCaluse(mdata, h, &data, trx, wc, res, req, psP) {
			return
		}

		default_tmpl, err := hdlr.GenTemplate(h, "Update", mdata, wc, &data, false, psP) // "update \"%{table_name%}\" %{updcols%} %{where_where%} %{where%}"
		if err != nil {
			// trx.AddNote(1, fmt.Sprintf("%v", err))
			// func ReturnErrorMessage(status int, msg string, code string, details string, res http.ResponseWriter, req *http.Request, ps goftlmux.Params) {
			// rv = fmt.Sprintf(`{"status":"error","msg":"Error(12909): template error (%s) sql-cfg.json[%s]",%s}`,
			// 	sizlib.EscapeError(err), cfgTag, godebug.LFj())
			// trx.SetQryDone(rv, "")
			ReturnErrorMessage(406, "Error(12909): template error", "12909",
				fmt.Sprintf(`Error(12909): template error (%s) sql-cfg.json[%s] %s %s`, sizlib.EscapeError(err), cfgTag, err, godebug.LFj()),
				res, req, *psP, trx, hdlr) // status:error
			return
		}
		wc.GenWhereClause(mdata)
		Query := UseTemplate("Update", h.UpdateTmpl, default_tmpl, mdata, trx)

		fmt.Printf("Query: %s; data=%s\n", Query, sizlib.SVar(data))

		err = sizlib.Run1Thx(hdlr.gCfg.Pg_client.Db, trx, Query, data...)
		if err != nil {
			rv = fmt.Sprintf(`{ "status":"error", "msg":%q, %s }`, err, godebug.LFj())
			trx.SetQryDone(rv, "")
			// xyzzyErrorReport -- err

			fmt.Fprintf(os.Stderr, "\n%serror %s, %s %s\n", MiscLib.ColorRed, err, godebug.LF(), MiscLib.ColorReset)
			fmt.Fprintf(os.Stdout, "\nError %s, %s\n\n", err, godebug.LF())
			trx.AddNote(1, fmt.Sprintf("Error on Query %s, %s", Query, err))

			ReturnErrorMessage(406, "Error(12409): database error error", "12909",
				fmt.Sprintf(`Error(12409): database error (%s) sql-cfg.json[%s] %s %s`, sizlib.EscapeError(err), cfgTag, err, godebug.LFj()),
				res, req, *psP, trx, hdlr) // status:error
			return
		} else {
			// rv = `{ "status":"success", "x1":1, "category":"ooky" }`

			rd := make(map[string]interface{})

			rd["status"] = "success"
			rd["x1"] = 3
			//	for k, v := range m {
			for j := 0; j < (*psP).NParam; j++ {
				k := (*psP).Data[j].Name
				v := (*psP).Data[j].Value
				f := (*psP).Data[j].From
				// fmt.Printf ( "fr=%s, k=%s\n", fr[k], k )
				// if strings.HasPrefix(k, "$") {
				if len(k) > 0 && k[0] == '$' {
					// fmt.Printf ( "Prefix of '$' found, k=%s\n", k )
				} else if map[string]bool{"XSRF-TOKEN": true, "auth_token": true, "cookie_csrf_token": true}[k] {
					// fmt.Printf ( "Not returned, k=%s\n", k );
					// } else if map[string]bool{"Cookie": true, "User Validation": true}[fr[k]] {
				} else if f == goftlmux.FromInject || f == goftlmux.FromOther {
					// fmt.Printf ( "Not returned, fr=%s k=%s\n", fr[k], k );
				} else {
					rd[k] = v
				}
			}

			rv0, err := json.MarshalIndent(rd, "", "\t")
			if err != nil {
				fmt.Printf("Unable to convert to JSON data, %v\n", err)
			} else {
				rv = string(rv0)
			}
			trx.SetQryDone("", rv)
		}
		// io.WriteString(res, sizlib.JsonP(rv, res, req))
		io.WriteString(res, rv)
	}
}

// ====================================================================================================================================================================
// DELETE
// ====================================================================================================================================================================
// func respHandlerTableDel(res http.ResponseWriter, req *http.Request) {
func closure_respHandlerTableDel(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select
		var rv string

		rw, _ /*top_hdlr*/, psP, err := GetRwPs(res, req)

		ok, mdata, cfgTag, h, trx, table_name, wc := hdlr.CommmonCrudPrefix(res, req, "delete", psP, rw)
		if !ok {
			return
		}

		trx.AddNote(1, "Validate Query Parameters")
		err = ValidateQueryParams(psP, h, req) // Validate them
		if err != nil {
			ReturnErrorMessage(406, "Invalid Parameter", "12043",
				fmt.Sprintf(`Error(12043): Invalid Query Parameters (%s) sql-cfg.json[%s] %s %s`, sizlib.EscapeError(err), cfgTag, err, godebug.LF()),
				res, req, *psP, trx, hdlr) // status:error
			return
		}

		data := CommonMakeData(h, trx, psP)

		if ok, _ = hdlr.CommonWhereClause(true, "", false, mdata, &data, h, table_name, trx, wc, res, req, psP); !ok {
			return
		}

		if !hdlr.ExtendedWhereCaluse(mdata, h, &data, trx, wc, res, req, psP) {
			return
		}

		default_tmpl, err := hdlr.GenTemplate(h, "Delete", mdata, wc, &data, false, psP) // "delete from \"%{table_name%}\" %{where_where%} %{where%}"
		if err != nil {
			ReturnErrorMessage(406, "Error(12909): template error", "12909",
				fmt.Sprintf(`Error(12909): template error (%s) sql-cfg.json[%s] %s %s`, sizlib.EscapeError(err), cfgTag, err, godebug.LFj()),
				res, req, *psP, trx, hdlr) // status:error
			return
		}
		wc.GenWhereClause(mdata)
		Query := UseTemplate("Delete", h.DeleteTmpl, default_tmpl, mdata, trx)

		if db_DumpDelete {
			fmt.Printf("Query: %s\n\tData=%s\n\tAT:%s\n", Query, tr.SVar(data), godebug.LF())
		}

		err = sizlib.Run1Thx(hdlr.gCfg.Pg_client.Db, trx, Query, data...)
		if err != nil {
			// func ReturnErrorMessage(status int, msg string, code string, details string, res http.ResponseWriter, req *http.Request, ps goftlmux.Params) {
			rv = fmt.Sprintf(`{ "status":"error", "msg":%q, %s }`, err, godebug.LFj())
			trx.SetQryDone(rv, "")
			// xyzzyErrorReport -- err

			fmt.Fprintf(os.Stderr, "\n%serror %s, %s %s\n", MiscLib.ColorRed, err, godebug.LF(), MiscLib.ColorReset)
			fmt.Fprintf(os.Stdout, "\nError %s, %s\n\n", err, godebug.LF())
			trx.AddNote(1, fmt.Sprintf("Error on Query %s, %s", Query, err))

			ReturnErrorMessage(406, "Error(12709): database error", "12709",
				fmt.Sprintf(`Error(12709): database error (%s) sql-cfg.json[%s] %s %s`, sizlib.EscapeError(err), cfgTag, err, godebug.LFj()),
				res, req, *psP, trx, hdlr) // status:error
			return
		} else {
			rv = `{ "status":"success" }`
			trx.SetQryDone("", rv)
		}
		// io.WriteString(res, sizlib.JsonP(rv, res, req))
		io.WriteString(res, rv)

	}
}

func (hdlr *TabServer2Type) genTemplateExec(h SQLOne, op string, data *[]interface{}, ps *goftlmux.Params) (rv string, err error) {
	err = nil
	rv = ""
	switch op {
	case "G":
		rv, err = hdlr.ExecSProcCmd(h.G, len(*data))
	default:
		err = errors.New("Error (12903): Not configure for this operation.")
	}
	return
}

func (hdlr *TabServer2Type) ExecSProcCmd(name string, np int) (rv string, err error) {
	err = nil
	rv = ""
	if hdlr.GetDbType() == DbType_postgres {
		pp := "("
		com := ""
		for i := 1; i <= np; i++ {
			pp += fmt.Sprintf("%s$%d", com, i)
			com = ","
		}
		pp += ")"
		rv = fmt.Sprintf("select %s%s as \"x\"", name, pp)
	} else if hdlr.GetDbType() == DbType_odbc {
		/*
			pre := "declare \n"
			com := ","
			for i := 1; i <= np; i++ {
				if i >= np {
					com = ""
				}
				pre += fmt.Sprintf ( "@p%d varchar(30)%s\n", i, com )
			}
			pre += "\n"
			for i := 1; i <= np; i++ {
				pre += fmt.Sprintf ( "set @p%d = 'data';\n", i )
			}

			pp := " "
			com = ""
			for i := 1; i <= np; i++ {
				pp += fmt.Sprintf ( "%s@p%d", com, i )
				com = ","
			}
		*/
		pp := " "
		pre := ""
		com := ""
		for i := 1; i <= np; i++ {
			pp += fmt.Sprintf("%s?", com)
			com = ","
		}

		rv = fmt.Sprintf("%s\nexec %s%s\n", pre, name, pp)
	} else if hdlr.GetDbType() == DbType_Oracle {
		panic("Error(00000): Not implemented yet.")
	} else {
		err = errors.New("Error (12903): Not configure for this operation.")
	}
	return
}

// ====================================================================================================================================================================
// ====================================================================================================================================================================
func (hdlr *TabServer2Type) GenTemplate(h SQLOne, op string, mdata map[string]string, wc *WhereCollect, data *[]interface{}, isCount bool, ps *goftlmux.Params) (string, error) {

	var bp int

	mdata["bQ"] = DbBeginQuote
	mdata["eQ"] = DbEndQuote

	if !isCount {
		switch op {
		case "Delete":
			fallthrough
		case "Delete-PK1":
			fallthrough
		case "UnDelete-PK1":
			fallthrough
		case "Select-PK1":
			fallthrough
		case "Select-Count":
			fallthrough
		case "Select":
			fallthrough
		case "Update":
			/* this is the point to deal with customer_id partitioing - if $customer_id$ is set in mdata then - partition by it. ??  */
			if len(h.CustomerIdPart.ColName) > 0 {
				mdata["customerIdPartColName"] = h.CustomerIdPart.ColName
				if len(h.CustomerIdPart.ColAlias) > 0 {
					mdata["cipAlias"] = `"` + h.CustomerIdPart.ColAlias + `".`
				} else {
					mdata["cipAlias"] = ""
				}
				// xyzzyAAAAb2 - If not logged in - then no $customer_id$ and will be out of range - should generate an error - not PANIC
				v := ps.ByNameDflt("$customer_id$", "1")
				if !ps.HasName("$customer_id$") {
					fmt.Printf("Error (12902): Must be logged in to have a $customer_id$, probably an invalid sql-cfg.json setting. Remove 'CustomerIdPart' or make LoginRequired:true.\n")
					return "** not logged in **", errors.New("Error (12902): Must be logged in to have a $customer_id$, probably an invalid sql-cfg.json setting. Remove 'CustomerIdPart' or make LoginRequired:true.\n")
				} else {
					bp = AddBindValue(data, v)
				}
				mdata["customerIdBindNo"] = fmt.Sprintf("%d", bp)
				wc.AddClause(sizlib.Qt(" %{cipAlias%}%{bQ%}%{customerIdPartColName%}%{eQ%} = $%{customerIdBindNo%} ", mdata))
			}
		}
	}

	b := false
	if len(h.DeleteViaUpdate.ColName) > 0 {
		b = true
		mdata["delMarkerCol"] = h.DeleteViaUpdate.ColName
		if len(h.DeleteViaUpdate.ColAlias) > 0 {
			mdata["delAlias"] = DbBeginQuote + h.DeleteViaUpdate.ColAlias + DbEndQuote + `.`
		} else {
			mdata["delAlias"] = ""
		}
		switch h.DeleteViaUpdate.ColType {
		case "u":
			// odbc-xyzzy
			fallthrough
		case "":
			fallthrough
		case "s":
			mdata["Absent"] = `'` + h.DeleteViaUpdate.Absent + `'`   // xyzzy - SQL Quote
			mdata["Present"] = `'` + h.DeleteViaUpdate.Present + `'` // xyzzy - SQL Quote
		case "i":
			mdata["Absent"] = h.DeleteViaUpdate.Absent
			mdata["Present"] = h.DeleteViaUpdate.Present
		case "b":
			mdata["Absent"] = h.DeleteViaUpdate.Absent
			mdata["Present"] = h.DeleteViaUpdate.Present
		}
	}

	switch op {
	case "Delete":
		if b {
			return "update %{bQ%}%{table_name%}%{eQ%} set %{bQ%}%{delMarkerCol%}%{eQ%} = %{Absent%} %{where_where%} %{where%}", nil
		} else {
			return "delete from %{bQ%}%{table_name%}%{eQ%} %{where_where%} %{where%}", nil
		}
	case "Delete-PK1":
		if b {
			return "update %{bQ%}%{table_name%}%{eQ%} set %{bQ%}%{delMarkerCol%}%{eQ%} = %{Absent%} %{where_where%} %{where%}", nil
		} else {
			return "delete from %{bQ%}%{table_name%}%{eQ%} %{where_where%} %{where%}", nil
		}

	case "UnDelete-PK1":
		if b {
			return "update %{bQ%}%{table_name%}%{eQ%} set %{bQ%}%{delMarkerCol%}%{eQ%} = %{Present%} %{where_where%} %{where%}", nil
		} else {
			return "** not configure for this operation **", errors.New("Error (12903): Not configure for this operation.")
		}

	case "Select-PK1":
		if b {
			wc.AddClause(sizlib.Qt(" %{delAlias%}%{bQ%}%{delMarkerCol%}%{eQ%} = %{Present%} ", mdata))
			// return "select %{cols%} from %{bQ%}%{table_name%}%{eQ%} where %{bQ%}%{delMarkerCol%}%{eQ%} = %{Present%} %{where_where%} %{where%} %{order_by_order_by%} %{order_by%}"
			return "select %{cols%} from %{bQ%}%{table_name%}%{eQ%} %{where_where%} %{where%} %{order_by_order_by%} %{order_by%}", nil
		} else {
			return "select %{cols%} from %{bQ%}%{table_name%}%{eQ%} %{where_where%} %{where%} %{order_by_order_by%} %{order_by%}", nil
		}
	case "Select-Count":
		if b {
			wc.AddClause(sizlib.Qt(" %{delAlias%}%{bQ%}%{delMarkerCol%}%{eQ%} = %{Present%} ", mdata))
			// return "select count(*) as %{bQ%}nRows%{eQ%} from %{bQ%}%{table_name%}%{eQ%} where %{bQ%}%{delMarkerCol%}%{eQ%} = %{Present%} %{where_where%} %{where%}"
			return "select count(*) as %{bQ%}nRows%{eQ%} from %{bQ%}%{table_name%}%{eQ%} %{where_where%} %{where%}", nil
		} else {
			return "select count(*) as %{bQ%}nRows%{eQ%} from %{bQ%}%{table_name%}%{eQ%} %{where_where%} %{where%}", nil
		}
	case "Select":
		if hdlr.GetDbType() == DbType_postgres {
			if b {
				wc.AddClause(sizlib.Qt(" %{delAlias%}%{bQ%}%{delMarkerCol%}%{eQ%} = %{Present%} ", mdata))
				// return "select %{cols%} from %{bQ%}%{table_name%}%{eQ%} where %{bQ%}%{delMarkerCol%}%{eQ%} = %{Present%} %{where_where%} %{where%} %{order_by_order_by%} %{order_by%} %{limit_limit%} %{limit%} %{offset_offset%} %{offset%}"
				return "select %{cols%} from %{bQ%}%{table_name%}%{eQ%} %{where_where%} %{where%} %{order_by_order_by%} %{order_by%} %{limit_limit%} %{limit%} %{limit_after%} %{offset_offset%} %{offset%} %{offset_after%}", nil
			} else {
				return "select %{cols%} from %{bQ%}%{table_name%}%{eQ%} %{where_where%} %{where%} %{order_by_order_by%} %{order_by%} %{limit_limit%} %{limit%} %{limit_after%} %{offset_offset%} %{offset%} %{offset_after%}", nil
			}
		} else if hdlr.GetDbType() == DbType_odbc {
			// MS-SQL requries a specific order that is the reverse of PostgreSQL
			if b {
				wc.AddClause(sizlib.Qt(" %{delAlias%}%{bQ%}%{delMarkerCol%}%{eQ%} = %{Present%} ", mdata))
				// return "select %{cols%} from %{bQ%}%{table_name%}%{eQ%} where %{bQ%}%{delMarkerCol%}%{eQ%} = %{Present%} %{where_where%} %{where%} %{order_by_order_by%} %{order_by%} %{limit_limit%} %{limit%} %{offset_offset%} %{offset%}"
				return "select %{cols%} from %{bQ%}%{table_name%}%{eQ%} %{where_where%} %{where%} %{order_by_order_by%} %{order_by%} %{offset_offset%} %{offset%} %{offset_after%} %{limit_limit%} %{limit%} %{limit_after%}", nil
			} else {
				return "select %{cols%} from %{bQ%}%{table_name%}%{eQ%} %{where_where%} %{where%} %{order_by_order_by%} %{order_by%} %{offset_offset%} %{offset%} %{offset_after%} %{limit_limit%} %{limit%} %{limit_after%}", nil
			}
		} else {
			// Oracle - Totaly different - uses ROWNUM and subqueries.
			if b {
				wc.AddClause(sizlib.Qt(" %{delAlias%}%{bQ%}%{delMarkerCol%}%{eQ%} = %{Present%} ", mdata))
				// return "select %{cols%} from %{bQ%}%{table_name%}%{eQ%} where %{bQ%}%{delMarkerCol%}%{eQ%} = %{Present%} %{where_where%} %{where%} %{order_by_order_by%} %{order_by%} %{limit_limit%} %{limit%} %{offset_offset%} %{offset%}"
				return "%{before_query%}select %{cols%} from %{bQ%}%{table_name%}%{eQ%} %{where_where%} %{where%} %{order_by_order_by%} %{order_by%}%{after_query%}", nil
			} else {
				return "%{before_query%}select %{cols%} from %{bQ%}%{table_name%}%{eQ%} %{where_where%} %{where%} %{order_by_order_by%} %{order_by%}%{after_query%}", nil
			}
		}

	case "Insert":
		return "insert into %{bQ%}%{table_name%}%{eQ%} ( %{cols%} ) values ( %{vals%} )", nil

	case "Update":
		if b {
			wc.AddClause(sizlib.Qt(" %{delAlias%}%{bQ%}%{delMarkerCol%}%{eQ%} = %{Present%} ", mdata))
			// return "update %{bQ%}%{table_name%}%{eQ%} %{updcols%} where %{bQ%}%{delMarkerCol%}%{eQ%} = %{Present%} %{where_where%} %{where%}"
			return "update %{bQ%}%{table_name%}%{eQ%} %{updcols%} %{where_where%} %{where%}", nil
		} else {
			return "update %{bQ%}%{table_name%}%{eQ%} %{updcols%} %{where_where%} %{where%}", nil
		}
	default:
		fmt.Printf("Ouch(12059): - unreacable code!\n")
		return "", errors.New("Error (12059): Internal error - supposedly unreacabl ecode reached.")
		// panic("bad code.")
	}
}

func closure_respHandlerTableDesc(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select

		rw, _ /*top_hdlr*/, psP, _ /*err*/ := GetRwPs(res, req)

		res.Header().Set("Content-Type", "application/json")
		// table_name := mux.Vars(req)["name"]
		table_name := psP.ByName("name")
		cfgTag := "/api/table/" + table_name
		trx := mid.GetTrx(rw)
		if !hdlr.validOp(table_name, "info", cfgTag) {
			// io.WriteString(res, "{\"status\":\"error\",\"code\":\"00004\",\"msg\":\"Invalid operation on this table.\"}")
			ReturnErrorMessage(403, "Invalid Operation", "00004",
				fmt.Sprintf(`Error(00004): Invalid Operation on this Table sql-cfg.json[%s] %s`, cfgTag, godebug.LF()), res, req, *psP, trx, hdlr) // status:error
			return
		}

		// xyzzyDOIT - Don't you think that you should DO somethin at this point

		io.WriteString(res, "{\"status\":\"success\",\"method\":\"post\"}")
	}
}

/* vim: set noai ts=4 sw=4: */
