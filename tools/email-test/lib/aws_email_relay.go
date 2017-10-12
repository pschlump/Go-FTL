//
// Copyright (C) Philip Schlump, 2013-2017
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 0001
// 你好眴世界
//

package aws

import (
	"fmt"

	"github.com/pschlump/Go-FTL/ses-aws/email-lib"
	"github.com/pschlump/godebug"
)

// "www.2c-why.com/ses-aws/email-lib"

//--var kermit *regexp.Regexp
//--var mis_piggy *regexp.Regexp
var el *emaillib.EmailLib

func init() {
	//--kermit = regexp.MustCompile("kermit.*@the-green-pc.com")
	//--mis_piggy = regexp.MustCompile("mis_piggy.*@the-green-pc.com")
	el = emaillib.NewEmailLib()
}

func ConfigEmailAWS(file string) {
	el.ReadCfg(file)
}

func SendEmailViaAWS(dSubject, dBodyHtml, dBodyText, email_addr string) (err error) {

	err = el.SendEmailMessage(email_addr, dSubject, dBodyHtml, dBodyText)
	if err != nil {
		fmt.Printf("Error %s on email, %s\n", err, godebug.LF())
	}

	return

}

/* vim: set noai ts=4 sw=4: */
