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
	"os"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/json" //	"encoding/json"
)

// RedirectTo should be the last step in a chain of post calls.  It is final.  Data is written to do the redirct and status is set.
func RedirectTo(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rvOut string, pptFlag PrePostFlagType, exit bool, a_status int) {

	fmt.Printf("%sAT:%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())

	type RedirectToData struct {
		Status     string `json:"status"`
		RedirectTo string `json:"$redirect_to$"`
	}

	var ed RedirectToData
	var all map[string]interface{}

	err := json.Unmarshal([]byte(rv), &ed)
	if err != nil {
		return rv, PrePostFatalSetStatus, true, 500
	}
	err = json.Unmarshal([]byte(rv), &all)
	if err != nil {
		return rv, PrePostFatalSetStatus, true, 500
	}

	if ed.Status == "success" && ed.RedirectTo != "" {

		to := RunTemplateString(ed.RedirectTo, all)

		fmt.Fprintf(os.Stderr, "\n%sThis One%s to=%s ed=%s all=%s at:%s\n", MiscLib.ColorGreen, MiscLib.ColorReset, to, godebug.SVarI(ed), godebug.SVarI(all), godebug.LF())
		fmt.Fprintf(os.Stdout, "\n%sThis One%s to=%s ed=%s all=%s at:%s\n", MiscLib.ColorGreen, MiscLib.ColorReset, to, godebug.SVarI(ed), godebug.SVarI(all), godebug.LF())
		fmt.Printf("%sAT:%s to = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, to, godebug.LF())

		res.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
		res.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
		res.Header().Set("Expires", "0")                                         // Proxies.
		res.Header().Set("Content-Type", "text/html")                            //
		res.Header().Set("Location", to)
		res.WriteHeader(http.StatusTemporaryRedirect)
		return rv, PrePostDone, true, http.StatusTemporaryRedirect
	}

	return rv, PrePostContinue, false, 200
}

/* vim: set noai ts=4 sw=4: */
