//
// Package aessrp implements encrypted authentication and encrypted REST.
// SRP-6a for login authenticaiton, followed by AES 256 bit encrypted RESTful calls.
//
// Copyright (C) Philip Schlump, 2013-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 0001
// 你好眴世界
//

package AesSrp

import (
	"fmt"
	"regexp"

	"github.com/pschlump/Go-FTL/ses-aws/email-lib"
	"github.com/pschlump/godebug"
)

// "www.2c-why.com/ses-aws/email-lib"

var kermit *regexp.Regexp
var el *emaillib.EmailLib

func init() {
	kermit = regexp.MustCompile("kermit.*@the-green-pc.com")
	el = emaillib.NewEmailLib()
}

func ConfigEmailAWS(hdlr *AesSrpType, file string) {
	el.ReadCfg(file)
}

// ============================================================================================================================================
// s, err := client.Get(fmt.Sprintf("https://52.21.71.211/api/send?auth_token=Dg9Tp4ecr8Y3H19lQZtGwFX3ug&app=%s&tmpl=%s&to=%s&from=no-reply@2c-why.com&p1=%s",
func SendEmailViaAWS(hdlr *AesSrpType, email_addr string, app string, tmpl string, pw string, email_auth_token string) {

	if hdlr.KermitRule && kermit.MatchString(email_addr) {
		fmt.Printf("KermitRule true and matched [%s], no email sent\n", email_addr)
		fmt.Printf("	app [%s] tmpl [%s] pw [%s] email_auth_token [%s], %s\n", app, tmpl, pw, email_auth_token, godebug.LF())
		return
	}

	mdata := make(map[string]string)
	mdata["app"] = app
	mdata["tmpl"] = tmpl
	mdata["to"] = email_addr
	mdata["p1"] = pw
	mdata["p2"] = email_auth_token

	dSubject, dBodyHtml, dBodyText, err := el.TemplateEmail(mdata)
	if err != nil {
		fmt.Printf("Error %s on email, %s\n", err, godebug.LF())
	} else {
		err := el.SendEmailMessage(email_addr, dSubject, dBodyHtml, dBodyText)
		if err != nil {
			fmt.Printf("Error %s on email, %s\n", err, godebug.LF())
		}
	}

}

// ============================================================================================================================================
func SendEmailViaAWS_support(hdlr *AesSrpType, email_addr string, app string, tmpl string, to, sub, bod string) {

	mdata := make(map[string]string)
	mdata["app"] = app
	mdata["tmpl"] = tmpl
	mdata["to"] = email_addr
	mdata["p1"] = to
	mdata["p2"] = sub
	mdata["p3"] = bod

	dSubject, dBodyHtml, dBodyText, err := el.TemplateEmail(mdata)
	if err != nil {
		fmt.Printf("Error %s on email, %s\n", err, godebug.LF())
	} else {
		err := el.SendEmailMessage(to, dSubject, dBodyHtml, dBodyText)
		if err != nil {
			fmt.Printf("Error %s on email, %s\n", err, godebug.LF())
		}
	}
}

/* vim: set noai ts=4 sw=4: */