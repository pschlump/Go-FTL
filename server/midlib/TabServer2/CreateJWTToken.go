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
	"os"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/json" //	"encoding/json"
)

/*
	return rv, PrePostFatalSetStatus, true, 500
*/

// xyzzy-JWT
func CreateJWTToken(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rvOut string, pptFlag PrePostFlagType, exit bool, a_status int) {

	fmt.Printf("%sAT:%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())
	fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())

	// func SignToken(tokData []byte, keyFile string) (out string, err error) {
	//	hdlr.KeyFilePrivate        string                      // private key file for signing JWT tokens
	// https://github.com/dgrijalva/jwt-go.git

	type RedirectToData struct {
		Status    string   `json:"status"`
		JWTClaims []string `json:"$JWT-claims$"`
	}

	var ed RedirectToData
	var all map[string]interface{}

	err := json.Unmarshal([]byte(rv), &ed)
	if err != nil {
		return "", PrePostFatalSetStatus, true, 500
	}
	err = json.Unmarshal([]byte(rv), &all)
	if err != nil {
		return "", PrePostFatalSetStatus, true, 500
	}

	if ed.Status == "success" && len(ed.JWTClaims) > 0 {

		fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())

		claims := make(map[string]string)
		for _, vv := range ed.JWTClaims {
			claims[vv] = all[vv].(string)
			// delete(all, vv)
		}
		tokData := godebug.SVar(claims)

		fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())

		signedKey, err := SignToken([]byte(tokData), hdlr.KeyFilePrivate)
		if err != nil {
			all["status"] = "error"
			all["msg"] = fmt.Sprintf("Error: Unable to sign the JWT token, %s", err)
			delete(all, "$JWT-claims$")
			rv = godebug.SVar(all)

			fmt.Printf("Error: Unable to sign the JWT token, %s\n", err)
			fmt.Fprintf(os.Stderr, "Error: Unable to sign the JWT token, %s\n", err)
			return rv, PrePostFatalSetStatus, true, 406
		}

		fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top signedKey = -->>%s<<-- %s\n", MiscLib.ColorYellow, MiscLib.ColorReset, signedKey, godebug.LF())

		all["jwt_token"] = signedKey

		delete(all, "$JWT-claims$")

		rv = godebug.SVar(all)
		fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())
		return rv, PrePostContinue, false, 200
	}

	return rv, PrePostContinue, false, 200
}

/* vim: set noai ts=4 sw=4: */
