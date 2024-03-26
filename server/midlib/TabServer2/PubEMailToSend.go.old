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
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/godebug"
)

func PubEMailToSend(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rvOut string, pptFlag PrePostFlagType, exit bool, a_status int) {

	// rw, _ /*top_hdlr*/, _ /*ps*/, _ /*err*/ := GetRwPs(res, req)

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return "", PrePostFatalSetStatus, true, 500
	}

	defer hdlr.gCfg.RedisPool.Put(conn)

	// rr.RedisDo("PUBLISH", "emailReadyToSend", fmt.Sprintf(`{"cmd":"readToSend"}`))
	conn.Cmd("PUBLISH", "emailReadyToSend", fmt.Sprintf(`{"cmd":"readToSend"}`))

	return rv, PrePostContinue, exit, a_status
}
