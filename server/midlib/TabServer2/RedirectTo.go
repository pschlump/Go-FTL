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
	"net/url"

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
		Status     string   `json:"status"`
		RedirectTo string   `json:"$redirect_to$"`
		Variables  []string `json:"$redirect_vars$"`
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

		to := ed.RedirectTo
		fmt.Printf("%sAT: %s%s -- to %s\n", MiscLib.ColorCyan, godebug.LF(), MiscLib.ColorReset, to)
		if len(ed.Variables) > 0 {
			fmt.Printf("%sAT: %s%s -- variables %s\n", MiscLib.ColorCyan, godebug.LF(), MiscLib.ColorReset, ed.Variables)
			sep := "?"
			for _, vv := range ed.Variables {
				if xx, ok := all[vv]; ok {
					to += fmt.Sprintf("%s%s=%s", sep, url.QueryEscape(vv), url.QueryEscape(fmt.Sprintf("%v", xx)))
					sep = "&"
				}
			}
		}
		fmt.Printf("%sAT: %s%s -- to %s\n", MiscLib.ColorCyan, godebug.LF(), MiscLib.ColorReset, to)

		res.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
		res.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
		res.Header().Set("Expires", "0")                                         // Proxies.
		res.Header().Set("Content-Type", "text/html")                            //
		res.Header().Set("Location", to)
		res.WriteHeader(http.StatusTemporaryRedirect)
		return rv, PrePostRVUpdatedSuccess, true, http.StatusTemporaryRedirect
	}

	return rv, PrePostContinue, false, 200
}
