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

import (
	"fmt"
	"net/http"

	logrus "github.com/pschlump/pslog" // "github.com/sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/godebug" //	"encoding/json"
)

func CacheEUser(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rvOut string, pptFlag PrePostFlagType, exit bool, a_status int) {

	// rw, _ /*top_hdlr*/, _ /*ps*/, _ /*err*/ := GetRwPs(res, req)

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return "", PrePostFatalSetStatus, true, 500
	}

	defer hdlr.gCfg.RedisPool.Put(conn)

	if isError {
		return rv, PrePostFatalSetStatus, true, 500
	}

	exit = false
	a_status = 200
	if db_user_login {
		fmt.Printf("In CacheEUser\n")
	}
	x, err := sizlib.JSONStringToData(rv)
	if err != nil {
		ReturnErrorMessage(406, "Error(10009): Parsing return value vaivfailed", "10009",
			fmt.Sprintf(`Error(10009): Parsing return value failed sql-cfg.json[%s] %s - Post function CacheEUser, %s`, cfgTag, err, godebug.LF()),
			res, req, *ps, trx, hdlr) // status:error
		// xyzzy - func (hdlr *TabServer2Type) ProcessErrors(
		return "", PrePostRVUpdatedFail, true, 406
	}

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
	return rv, PrePostContinue, exit, a_status
}
