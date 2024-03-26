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

	logrus "github.com/pschlump/pslog" // "github.com/sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/godebug"
)

func DeCacheEUser(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rvOut string, pptFlag PrePostFlagType, exit bool, a_status int) {

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
		fmt.Printf("In DeCacheEUser\n")
	}
	if ps.HasName("auth_token") {
		auth_token := ps.ByName("auth_token")
		// rr.RedisDo("DEL", "api:USER:"+auth_token)
		// rr.RedisDo("PUBLISH", "pubsub", fmt.Sprintf(`{"cmd":"logout","auth_token":%q}`, auth_token))
		conn.Cmd("DEL", "api:USER:"+auth_token)
		conn.Cmd("PUBLISH", "pubsub", fmt.Sprintf(`{"cmd":"logout","auth_token":%q}`, auth_token))
	}
	return rv, PrePostContinue, exit, a_status
}

/* vim: set noai ts=4 sw=4: */
