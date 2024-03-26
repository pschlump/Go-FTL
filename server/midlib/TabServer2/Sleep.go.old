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
	"time"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/json" //	"encoding/json"
)

func Sleep(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rvOut string, pptFlag PrePostFlagType, exit bool, a_status int) {

	fmt.Printf("%sAT:%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())

	type RedirectToData struct {
		Status string `json:"status"`
		SleepN int    `json:"$sleep$"`
	}

	var ed RedirectToData
	err := json.Unmarshal([]byte(rv), &ed)
	if err != nil {
		fmt.Printf("%sAT:%s *** Sleep Ignored - Failed to Parse s %s *** rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, err, rv, godebug.LF())
		return rv, PrePostContinue, false, 200
	}

	if ed.SleepN > 0 {
		slowDown := time.Duration(int64(ed.SleepN)) * time.Second
		time.Sleep(slowDown)
	}

	return rv, PrePostContinue, false, 200
}

/* vim: set noai ts=4 sw=4: */
