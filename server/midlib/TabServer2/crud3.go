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
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/json" //	"encoding/json"
)

// ==============================================================================================================================================================================
//  Call by name of func table
// ==============================================================================================================================================================================
func SendReportsToGenMessage(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (string, bool, int) {
	//if isError {
	//	return rv, true, 500
	//}
	// rr.RedisDo("PUBLISH", "rptReadyToRun", fmt.Sprintf(`{"cmd":"readToSend","from":"tab-server1"}`))

	// rw, _ /*top_hdlr*/, _ /*ps*/, _ /*err*/ := GetRwPs(res, req)

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return "", true, 500
	}
	defer hdlr.gCfg.RedisPool.Put(conn)

	conn.Cmd("PUBLISH", "rptReadyToRun", fmt.Sprintf(`{"cmd":"readToSend","from":"tab-server1"}`))
	return rv, false, 200
}

func SendEmailToGenMessage(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (string, bool, int) {
	//if isError {
	//	return rv, true, 500
	//}
	// rr.RedisDo("PUBLISH", "emailReadyToSend", fmt.Sprintf(`{"cmd":"readToSend","from":"tab-server1"}`))

	// rw, _ /*top_hdlr*/, _ /*ps*/, _ /*err*/ := GetRwPs(res, req)

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return "", true, 500
	}
	defer hdlr.gCfg.RedisPool.Put(conn)

	conn.Cmd("PUBLISH", "emailReadyToSend", fmt.Sprintf(`{"cmd":"readToSend","from":"tab-server1"}`))
	return rv, false, 200
}

/*
   l_data = '{"status":"success","$send_email$":{'
   		||'"template":"please_confirm_registration"'
   		||',"username":'||to_json(p_username)
   		||',"real_name":'||to_json(p_real_name)
   		||',"email_token":'||to_json(l_email_token)
   		||',"app":'||to_json(p_app)
   		||',"name":'||to_json(p_name)
   		||',"url":'||to_json(p_url)
   		||',"from":'||to_json(l_from)
   	||'},"$session$":{'
   		||'"set":['
   			||'{"path":["gen","auth"],"value":"y"}'
   		||']'
   	||'}}';
*/
func RedirectTo(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (string, bool, int) {

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
		return rv, false, 200
	}
	err = json.Unmarshal([]byte(rv), &all)
	if err != nil {
		return rv, false, 200
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
		return rv, true, http.StatusTemporaryRedirect
	}

	return rv, false, 200
}

// xyzzy-JWT
func CreateJWTToken(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (string, bool, int) {

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
		return rv, false, 200
	}
	err = json.Unmarshal([]byte(rv), &all)
	if err != nil {
		return rv, false, 200
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
			return rv, true, 406
		}

		fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top signedKey = -->>%s<<-- %s\n", MiscLib.ColorYellow, MiscLib.ColorReset, signedKey, godebug.LF())

		all["jwt_token"] = signedKey

		delete(all, "$JWT-claims$")

		rv = godebug.SVar(all)
		fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())
		return rv, false, 200
	}

	return rv, false, 200
}

func Sleep(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (string, bool, int) {

	fmt.Printf("%sAT:%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())

	type RedirectToData struct {
		Status string `json:"status"`
		SleepN int    `json:"$sleep$"`
	}

	var ed RedirectToData
	err := json.Unmarshal([]byte(rv), &ed)
	if err != nil {
		return rv, false, 200
	}
	if ed.SleepN > 0 {
		slowDown := time.Duration(int64(ed.SleepN)) * time.Second
		time.Sleep(slowDown)
	}

	return rv, false, 200
}

func SendEmailMessage(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (string, bool, int) {

	fmt.Printf("%sAT:%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())

	type EmailData struct {
		Status string            `json:"status"`
		Email  map[string]string `json:"$send_email$"`
	}

	var ed EmailData
	err := json.Unmarshal([]byte(rv), &ed)
	if err != nil {
		return rv, false, 200
	}

	fmt.Printf("%sAT:%s ed=%s %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, godebug.LF(), godebug.SVarI(ed))
	var send_it = true
	var log_it = false
	if hdlr.gCfg.DbOn("*", "TabServer2", "db-email") {
		log_it = true
	}

	var mp = regexp.MustCompile("^mis_piggy")
	var kr = regexp.MustCompile("^kermit")
	if mp.MatchString(ed.Email["email_addr"]) {
		fmt.Printf("%sMiss Piggy Email Matched - skip send email, log  email%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
		send_it = false
		log_it = true
	}
	fmt.Printf("%sBefore Kermit check - email=%s %s\n", MiscLib.ColorRed, MiscLib.ColorReset, ed.Email["email_addr"])
	if kr.MatchString(ed.Email["email_addr"]) {
		fmt.Printf("%sKermit Matched Email - send email, log  email%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
		ed.Email["email_addr"] = "pschlump@gmail.com"
		send_it = true
		log_it = true
	}

	fmt.Printf("send_it %v log_it %v to = [%s]\n", send_it, log_it, ed.Email["email_addr"])

	if ed.Status == "success" {
		fmt.Printf("%sAT:%s %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, godebug.LF())
		s1, b1, b2, err := hdlr.TemplateEmail(ed.Email["template"], ed.Email)

		if log_it {
			fmt.Printf("Subject: %s\nHTML: %s\nSubject: %s\nText: %s\nerr=%s\n", s1, b1, b2, err)
			if _, ok := ed.Email["log_id"]; ok {
				ioutil.WriteFile(fmt.Sprintf("./output/%s.log", ed.Email["log_id"]), []byte(fmt.Sprintf("Subject:%s\nHTML:%s\nText:%s\n", s1, b1, b2)), 0666)
			}
		}
		if send_it {
			fmt.Printf("Sending email\n")
			fmt.Printf("%sSending email%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
			fmt.Printf("%sSending email%s\n", MiscLib.ColorYellow, MiscLib.ColorReset)
			fmt.Printf("%sSending email%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
			fmt.Printf("%sSending email%s\n", MiscLib.ColorYellow, MiscLib.ColorReset)
			fmt.Printf("%sSending email%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
			fmt.Printf("Sending email to %s\n", ed.Email["email_addr"])
			SendEmailViaAWS(s1, b1, b2, ed.Email["email_addr"])
			// xyzzy - if error - then it should be logged -> ./output! -- Notification sent to ?me?
		} else {
			fmt.Printf("Not Sending email\n")
			fmt.Printf("%sNot Sending email%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
			fmt.Printf("%sNot Sending email%s\n", MiscLib.ColorYellow, MiscLib.ColorReset)
			fmt.Printf("%sNot Sending email%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
			fmt.Printf("%sNot Sending email%s\n", MiscLib.ColorYellow, MiscLib.ColorReset)
			fmt.Printf("%sNot Sending email%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
			fmt.Printf("Not Sending email to %s\n", ed.Email["email_addr"])
		}

		// remove email data from return.
		teb := make(map[string]interface{})
		err = json.Unmarshal([]byte(rv), &teb)
		if err != nil {
			fmt.Printf("Internal error on sending email %s - data %s\n", err, rv)
			return "", true, 500
		}
		delete(teb, "$send_email$")
		rv = godebug.SVar(teb)
		fmt.Printf("%sAT:%s rv=%s %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())

	} else {
		fmt.Printf("%sAT:%s rv=%s %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())
		// xyzzy - should remove email info then return error.
	}

	return rv, false, 200

}
