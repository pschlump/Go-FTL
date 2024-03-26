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
	"regexp"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/json" //	"encoding/json"
)

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
func SendEmailMessage(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rvOut string, pptFlag PrePostFlagType, exit bool, a_status int) {

	fmt.Printf("%sAT:%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())

	type EmailData struct {
		Status string            `json:"status"`
		Email  map[string]string `json:"$send_email$"`
	}

	var ed EmailData
	err := json.Unmarshal([]byte(rv), &ed)
	if err != nil {
		return "", PrePostFatalSetStatus, true, 500
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
			return "", PrePostFatalSetStatus, true, 500
		}
		delete(teb, "$send_email$")
		rv = godebug.SVar(teb)
		fmt.Printf("%sAT:%s rv=%s %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())

	} else {
		fmt.Printf("%sAT:%s rv=%s %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())
		// xyzzy - should remove email info then return error.
	}

	return rv, PrePostContinue, exit, a_status

}

/* vim: set noai ts=4 sw=4: */
