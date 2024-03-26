//
// Go-FTL / TabServer2
//
// Copyright (C) Philip Schlump, 2012-2017. All rights reserved.
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1011
//

package TabServer2

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
)

/*

-- Redis ------------------------------------------------------------------------------------------------------------------------------

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		rw.Log.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	// ... do stuff ..
	data, err := conn.Cmd("GET", aKey).Str() // Get the value

	hdlr.gCfg.RedisPool.Put(conn)

-- PostgreSQL -------------------------------------------------------------------------------------------------------------------------

	Rows, err := hdlr.gCfg.Pg_client.Db.Query(Query, data...)

*/

// -----------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// Create the "closure" fucntion that will save passed data for later and return a
// function bound to the passed data.
// func Hello(w http.ResponseWriter, r *http.Request, ps goftlmux.Params) {
func GetSqlCfgHandler2(name string, hdlr *TabServer2Type) func(www http.ResponseWriter, req *http.Request) {
	return func(www http.ResponseWriter, req *http.Request) {
		if hdlr.gCfg.DbOn("*", "TabServer2", "db-closure-1") {
			fmt.Printf("AT top of GetSqlCfgHandler2, %s\n", godebug.LF())
		}
		if hdlr.gCfg.DbOn("*", "TabServer2", "db-closure-2") {
			fmt.Fprintf(os.Stderr, "%sAT top of GetSqlCfgHandler2, %s%s\n", MiscLib.ColorRed, godebug.LF(), MiscLib.ColorReset)
		}
		rw, _ /*hdlr*/, psP, err := GetRwPs(www, req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
			return
		}
		hdlr.RespHandlerSQL(www, req, name, psP, rw)
	}
}

func GetRedisCfgHandler2(name string, hdlr *TabServer2Type) func(www http.ResponseWriter, req *http.Request) {
	return func(www http.ResponseWriter, req *http.Request) {
		rw, _ /*hdlr*/, psP, err := GetRwPs(www, req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
			return
		}
		hdlr.RespHandlerRedis(www, req, name, psP, rw)
	}
}

const base64GifPixel = "R0lGODlhAQABAIAAAP///wAAACwAAAAAAQABAAACAkQBADs="

// ========================================================================================================================================================================
func respHandlerGrabFeedback(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "image/gif")
	output, _ := base64.StdEncoding.DecodeString(base64GifPixel)
	io.WriteString(res, string(output))
}

// ========================================================================================================================================================================
// func respHandlerStatus(res http.ResponseWriter, req *http.Request) {
func closure_respHandlerStatus(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select

		rw, _ /*hdlr*/, ps, _ /*err*/ := GetRwPs(res, req)

		_ = rw // will need it for GetTrx I Think

		q := req.RequestURI

		fmt.Printf("Method: %s\n", req.Method)

		// tr.TraceUriPs(req, ps)

		fmt.Printf("godebug.FUNCAME()=%s (1)=%s\n", godebug.FUNCNAME(), godebug.FUNCNAME(2))

		// TablesReferenced(godebug.FUNCNAME(), "/api/status:"+req.Method, []string{}, hdlr)

		// fmt.Fprintf(os.Stderr, "At: %s\n", godebug.LF())
		trx := mid.GetTrx(res)
		// fmt.Fprintf(os.Stderr, "At: %s\n", godebug.LF())
		trx.SetFunc(1) //xyzzy - verify may need (2) for closure_
		// fmt.Fprintf(os.Stderr, "At: %s\n", godebug.LF())
		SetDataPs(trx, ps) // trx.SetDataPs(ps) -- old had to de-couple packages
		// fmt.Fprintf(os.Stderr, "At: %s\n", godebug.LF())

		global_ver := hdlr.StatusMessage

		var sql_ver = ""
		if h, ok := hdlr.SQLCfg["note:"+req.Method]; ok {
			sql_ver = h.F
		}

		// fmt.Fprintf(os.Stderr, "At: %s\n", godebug.LF())
		id := ps.ByNameDflt("id", "-- id not set --") ///////////////////////////////// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< BOOM !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		// fmt.Fprintf(os.Stderr, "At: %s\n", godebug.LF())
		ps_fmt := ps.ByName("fmt")
		// fmt.Fprintf(os.Stderr, "At: %s\n", godebug.LF())
		var rv string
		if ps_fmt == "text" {
			res.Header().Set("Content-Type", "text/html")
			if id == "xyzzy" {
				rv = fmt.Sprintf("status:success\nURI:%q\nid:%q\nreq:%+v\nresponse_header:%+v", q, id, req, res.Header())
			} else {
				rv = fmt.Sprintf("status:success\nURI:%q\nid:%q\n", q, id)
			}
		} else {
			res.Header().Set("Content-Type", "application/json")
			if id == "xyzzy" {
				rv = fmt.Sprintf("{\"status\":\"success\",\n\"URI\":%q,\n\"SqlVer\":%q,\n\"GlobalVer\":%q,\n\"req\":%s, \"response_header\":%s}",
					q, sql_ver, global_ver, sizlib.SVarI(req), sizlib.SVarI(res.Header()))
			} else {
				rv = fmt.Sprintf("{\"status\":\"success\",\n\"URI\":%q,\n\"id\":%q}", q, id)
			}
		}

		// fmt.Fprintf(os.Stderr, "At: %s\n", godebug.LF())
		io.WriteString(res, rv)
		trx.SetRvBody(rv)
		fmt.Printf("***Got Status Request - Patch, global=%s sql=%s\n", global_ver, sql_ver)
	}
}

// ==============================================================================================================================================================================
func respHandlerLogIt(res http.ResponseWriter, req *http.Request) {

	res.Header().Set("Content-Type", "application/json")

	// Xyzzy - send to "logit" table?
	// _, err := hdlr.gCfg.Pg_client.Db.Query(`insert into "logit" ...`, data...)

	q := req.RequestURI

	fmt.Printf("***Log, URI=%s\n", q)
	if fx != nil {
		fmt.Fprintf(fx, "{\"type\":\"log\", \"uri\":%q}\n", q)
	}
	fmt.Fprintf(os.Stdout, "{\"type\":\"log\", \"uri\":%q}\n", q)
	io.WriteString(res, "{\"status\":\"success\",\"query\":\""+q+"\"}")
}

// ==============================================================================================================================================================================
// func respHandlerSwapLogFile(res http.ResponseWriter, req *http.Request) {
func closure_respHandlerSwapLogFile(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select

		_ /*rw*/, _ /*hdlr*/, ps, _ /*err*/ := GetRwPs(res, req)

		var err error

		// if GlobalCfg["log_to_file"] != "no" {
		if hdlr.LogToFile != "" {
			if err := fo.Close(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
		if err := fx.Close(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		dumpURL("respHandlerSwapLogFile", req)

		// get seq # from cmd line
		// seq := mux.Vars(req)["seq"]
		seq := ps.ByName("seq")
		// fmt.Printf ( "seq=%s\n", seq)
		// rename files
		// if GlobalCfg["log_to_file"] != "no" {
		if hdlr.LogToFile != "" {
			os.Rename(hdlr.LogToFile+"/alog.log", "/alog.log."+seq)
		}
		os.Rename(hdlr.LogToFile+"/xlog.log", hdlr.LogToFile+"/xlog.log."+seq)

		// if GlobalCfg["log_to_file"] != "no" {
		if hdlr.LogToFile != "" {
			fo, err = os.Create(hdlr.LogToFile + "/alog.log")
			if err != nil {
				panic(err)
			}
			// close fo on exit and check for its returned error
			defer func() {
				if err := fo.Close(); err != nil {
					panic(err)
				}
			}()
		}
		fx, err = os.Create(hdlr.LogToFile + "/xlog.log")
		if err != nil {
			panic(err)
		}
		// close fo on exit and check for its returned error
		defer func() {
			if err := fx.Close(); err != nil {
				panic(err)
			}
		}()
		res.Header().Set("Content-Type", "application/javascript") // For JSONP
		io.WriteString(res, "{\"status\":\"success\"}")
	}
}

// ==============================================================================================================================================================================
// ----------------------------------------------------------------------- Config Ops ------------------------------------------------------------------------------------
// ==============================================================================================================================================================================
// func respHandlerReloadTableConfig(res http.ResponseWriter, req *http.Request) {
func closure_respHandlerReloadTableConfig(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select

		_ /*rw*/, _ /*hdlr*/, ps, _ /*err*/ := GetRwPs(res, req)

		var err error
		res.Header().Set("Content-Type", "application/json")

		dev_auth_token := ps.ByName("dev_auth_token")
		fn := ps.ByName("fn")
		ok := (dev_auth_token == hdlr.DevAuthToken)
		// fmt.Printf("dev_auth_token [%s] hdlr.DevAuthToken [%s], %s\n", dev_auth_token, hdlr.DevAuthToken, godebug.LF())
		fn_valid := false
		pos := -1
		// fmt.Printf("fn --->>>%s<<<---\n", fn)
		// fmt.Printf("valid files are --->>>%s<<<---\n", godebug.SVarI(SqlCfgFilesLoaded))
		for i, v := range SqlCfgFilesLoaded {
			if v.FileName == fn {
				pos = i
				fn_valid = true
			}
		}

		// fmt.Printf("ok=%v fn_valid=%v, %s\n", ok, fn_valid, godebug.LF())

		if ok && fn_valid {

			fmt.Printf("***Got Reaload Config Request - Get, %s, %s, %s\n", dev_auth_token, fn, godebug.LF())
			// io.WriteString(res, fmt.Sprintf(`{"status":"success","method":"get","fn":"%s","ok":%v,"fn_valid":%v}`, fn, ok, fn_valid))

			fmt.Printf("Re-Reading SQLCfg: %s\n", fn)
			had_err := false
			tSQLCfg, err := readInSQLConfig(fn) // func readInSQLConfig(path string) map[string]SQLOne {
			if err != nil {
				had_err = true
				fmt.Printf("Error: %s, %s\n", err, godebug.LF())
				SqlCfgFilesLoaded[pos].ErrorMsg = fmt.Sprintf("%s", err)
			} else {
				for j, w := range tSQLCfg {
					hdlr.SQLCfg[j] = w
				}
				SqlCfgFilesLoaded[pos].ErrorMsg = ""

				hdlr.MuxAutoPass++
				for key, val := range tSQLCfg {
					if len(key) > 0 && key[0:1] == "/" && (!strings.HasPrefix(key, "/api/table/") || len(val.Crud) == 0) && !val.Redis {
						hdlr.MuxAuto[key] = hdlr.MuxAutoPass
						if debugCrud01 {
							fmt.Printf("Creating %s for %v\n", key, val.Method)
						}
						if len(val.Method) == 0 {
							hdlr.theMux.HandleFunc(key, GetSqlCfgHandler2(key, hdlr)).Methods("GET")
						} else {
							hdlr.theMux.HandleFunc(key, GetSqlCfgHandler2(key, hdlr)).Methods(val.Method...)
						}
					} else if len(key) > 0 && key[0:1] == "/" && val.Redis {
						hdlr.MuxAuto[key] = hdlr.MuxAutoPass
						if len(val.Method) == 0 {
							hdlr.theMux.HandleFunc(key, GetRedisCfgHandler2(key, hdlr)).Methods("GET")
						} else {
							hdlr.theMux.HandleFunc(key, GetRedisCfgHandler2(key, hdlr)).Methods(val.Method...)
						}
					}
				}
			}

			if !had_err {
				fmt.Printf("***Got Reaload Config Request - Get, %s, %s\n", dev_auth_token, fn)
				io.WriteString(res, fmt.Sprintf(`{"status":"success","method":"get","fn":"%s","ok":%v,"fn_valid":%v}`, fn, ok, fn_valid))
			} else {
				fmt.Printf("***Got Reaload Config Request - Get, %s\n", err)
				io.WriteString(res, "{\"status\":\"error\",\"msg\":\"syntax error in file\"}")
			}

		} else {

			fmt.Printf("***Got Reaload Config Request - Get, %s\n", err)
			io.WriteString(res, "{\"status\":\"error\",\"method\":\"get\"}")

		}
		return

		/*
			SQLCfg, err = readInSQLConfig(opts.SQLCfgFN)
			if err != nil {
				fmt.Printf("***Got Reaload Config Request - Get, %s\n", err)
				io.WriteString(res, "{\"status\":\"error\",\"method\":\"get\"}")
			} else {
				fmt.Printf("***Got Reaload Config Request - Get\n")
				io.WriteString(res, "{\"status\":\"success\",\"method\":\"get\"}")

				MuxAutoPass++

				for key, val := range SQLCfg {
					if len(key) > 0 && key[0:1] == "/" && (!strings.HasPrefix(key, "/api/table/") || len(val.Crud) == 0) && !val.Redis {
						MuxAuto[key] = MuxAutoPass
						if len(val.Method) == 0 {
							theMux.HandleFunc(key, GetSqlCfgHandler(key)).Methods("GET")
						} else {
							theMux.HandleFunc(key, GetSqlCfgHandler(key)).Methods(val.Method...)
						}
					}
				}
			}
		*/

		// Intent is to disable any stuff that disapears - don't understand how old code worked - xyzzy
		//	for key, val := range muxAuto {
		//		if val == (muxAutoPass - 1) {
		//			theMux.HandleFunc(key, respHandlerDepricated).Methods("GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS")
		//		}
		//	}
	}
}

// ==============================================================================================================================================================================
// func respHandlerListBuiltinRoutes(res http.ResponseWriter, req *http.Request) {
func closure_respHandlerListBuiltinRoutes(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select

		type ASingleRoute struct {
			Method string
			Path   string
			LineNo string
		}

		res.Header().Set("Content-Type", "application/json")
		routes := hdlr.theMux.ListRoutes()
		rv := make([]ASingleRoute, 0, len(routes))
		for _, v := range routes {
			for _, m := range v.DMethods {
				rv = append(rv, ASingleRoute{Method: m, Path: v.DPath, LineNo: fmt.Sprintf("%s:%d", v.FileName, v.LineNo)})
			}
		}
		// io.WriteString(res, sizlib.SVarI(rv))
		fmt.Fprintf(res, `{"status":"success","data":%s}`, sizlib.SVarI(rv))
	}
}

// ==============================================================================================================================================================================
// Rreturn a message with the list of available/installed "themes"
// func respHandlerListInstalledThemes(res http.ResponseWriter, req *http.Request) {
func closure_respHandlerListInstalledThemes(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select

		_ /*rw*/, _ /*hdlr*/, ps, _ /*err*/ := GetRwPs(res, req)

		customer_id := ps.ByNameDflt("$customer_id$", "")
		var themes []string

		// base_dir := GlobalCfg["static_dir"]
		for _, base_dir := range hdlr.AppRoot {

			//	dev_auth_token := ps.ByName("dev_auth_token")
			//	ok := true
			//	if dev_auth_token != GlobalCfg["dev_auth_token"] {
			//		ok = false
			//	}

			//	if ok {

			_, dirs := sizlib.GetFilenames(base_dir + "/tmpl")

			fmt.Printf("dirs (1) = %s\n", sizlib.SVar(dirs))
			for _, v := range dirs {
				// fmt.Printf("dirs (a) = ->%s<-, ->%s<-\n", v, customer_id)
				if v == "u:"+customer_id {
					_, tmp := sizlib.GetFilenames(base_dir + "/tmpl/u:" + customer_id)
					// fmt.Printf("dirs (2) = %s\n", sizlib.SVar(tmp))
					themes = append(themes, tmp...)
				} else if v[0:2] == "u:" {
				} else {
					themes = append(themes, v)
				}
			}

			// 1. Fitler for base_dir/tmpl/(<name>)$
			// dirs = sizlib.FilterArray(base_dir+"/tmpl/[^/]*$", dirs)		// "[^/]*$"
			// fmt.Printf("dirs (2) = %s\n", sizlib.SVar(dirs))

			// 2. Fitler for base_dir/tmpl/<customer_id>/(<name>)$
			// dirs = sizlib.FilterArray(base_dir+"/tmpl/"+customer_id+"/[^/]*$", dirs)
			// fmt.Printf("dirs (3) = %s\n", sizlib.SVar(dirs))

			//dirs := sizlib.FindDirsWithSQLCfg(opts__TopPath, ignoreList)
			//fmt.Printf("dirs ->%s<- At, %s\n", sizlib.SVar(dirs), godebug.LF())
			//fList, ok := sizlib.SearchPathAppModule(opts.SQLCfgFN, opts.AppName, dirs)
			//fmt.Printf("fList ->%s<- At, %s\n", sizlib.SVar(fList), godebug.LF())
			// xyzzy
			// xyzzy
			// xyzzy
		}

		io.WriteString(res, `{"status":"success","themes":`+sizlib.SVar(themes)+`}`)

		//	} else {
		//		io.WriteString(res, `{"status":"error"}`)
		//	}

	}
}

// ==============================================================================================================================================================================
// Return the currently set "theme" or an empty string.
// func respHandlerListCurrentTheme(res http.ResponseWriter, req *http.Request) {
func closure_respHandlerListCurrentTheme(hdlr *TabServer2Type) func(res http.ResponseWriter, req *http.Request) { // Select

	return func(res http.ResponseWriter, req *http.Request) { // Select

		_ /*rw*/, _ /*hdlr*/, ps, _ /*err*/ := GetRwPs(res, req)

		var client_config map[string]interface{}
		templateName := ""
		customer_id := ps.ByNameDflt("$customer_id$", "")
		data := sizlib.SelData(hdlr.gCfg.Pg_client.Db, "select \"config\" from \"p_client_config\" where \"customer_id\" = $1", customer_id)
		if len(data) > 0 {
			err := json.Unmarshal([]byte(data[0]["config"].(string)), &client_config)
			if err != nil {
				//fmt.Printf("At: %s\n", godebug.LF())
				fmt.Printf("Error(10012): %v, %s, Client Config Error, Invlaid JSON\n", err, godebug.LF())
				http.Error(res, "404 page not found", http.StatusNotFound) // 404
				return
			} else {
				//fmt.Printf("At: %s\n", godebug.LF())
				if _, ok := client_config["TemplateName"]; ok {
					templateName = client_config["TemplateName"].(string)
				}
			}
		}
		io.WriteString(res, fmt.Sprintf(`{"status":"success","theme":%q}`, templateName))
	}
}

func dumpURL(s string, req *http.Request) {
	if db_dumpURL {
		fmt.Printf("%s\n", s)
		fmt.Printf("\treq.URL.Scheme=%s\n", req.URL.Scheme)
		fmt.Printf("\treq.URL.Host=%s\n", req.URL.Host)
		fmt.Printf("\treq.URL.Path=%s\n", req.URL.Path)
		fmt.Printf("\treq.URL.RawQuery=%s\n", req.URL.RawQuery)
		fmt.Printf("\treq.URL.Fragment=%s\n", req.URL.Fragment)
	}
}

const db_dumpURL = false

/* vim: set noai ts=4 sw=4: */
