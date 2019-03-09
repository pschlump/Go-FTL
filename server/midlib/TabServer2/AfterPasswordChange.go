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
	"fmt"
	"net/http"
	"strings"
	"time"

	logrus "github.com/pschlump/pslog" // "github.com/sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/godebug" //	"encoding/json"
	"github.com/pschlump/uuid"
)

/*
 */

func AfterPasswordChange(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rvOut string, pptFlag PrePostFlagType, exit bool, a_status int) {

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
			// combine into SETEX, PJS: Thu Feb 21 13:14:26 MST 2019
			// conn.Cmd("SET", "api:USER:"+auth_token, dt)
			// conn.Cmd("EXPIRE", "api:USER:"+auth_token, 1*60*60) // Validate for 1 hour
			conn.Cmd("SETEX", "api:USER:"+auth_token, 1*60*60, dt) // Validate for 1 hour
		}
	}
	return rv, PrePostContinue, exit, a_status
}

/* vim: set noai ts=4 sw=4: */
