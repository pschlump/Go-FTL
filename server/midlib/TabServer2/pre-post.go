package TabServer2

import (
	"fmt"
	"net/http"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/tr"
)

type PrePostFlagType int

//
// Pre-Post fucntion calls return a status that indicates what is to happen next.
// In the past this has been a true/false - this was a bad choice.   This set of
// values is to repace the 2nd return value from fuctions like X2faSetup.
//
// Hopefully this will clarify the processing after a call to a pre/post
// procesing function.
//
const (
	PrePostRVUpdatedSuccess PrePostFlagType = 1 // rv written, use 'status' - processing complete.	exit=true
	PrePostRVUpdatedFail    PrePostFlagType = 2 // rv written, use 'status' - processing complete.	exit=true
	PrePostNoAction         PrePostFlagType = 3 // go to next processing, neither 'rv' or 'satus' relevant.
	PrePostFatalSetStatus   PrePostFlagType = 4 // Fatal Error: set status.	exit=true
	PrePostNextStep         PrePostFlagType = 5 // Just procede to next processing step.
)

// xyzzy - String() function for these constants! - so can log it.

// var funcMap map[string]func(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps goftlmux.Params, trx *tr.Trx) (string, bool, int)
var funcMap map[string]FuncMapType

type FuncMapType func(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (string, bool, int)

// NEW: type FuncMapType func(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) ( /*rv*/ string /*ReturnStatus*/, PrePostFlagType /*status*/, int)

func init() {
	// fmt.Printf("init in main\n")
	funcMap = map[string]FuncMapType{
		"CacheEUser":              CacheEUser,
		"DeCacheEUser":            DeCacheEUser,
		"AfterPasswordChange":     AfterPasswordChange,
		"ConvertErrorToCode":      ConvertErrorToCode,
		"PubEMailToSend":          PubEMailToSend,
		"SendReportsToGenMessage": SendReportsToGenMessage,
		"SendEmailToGenMessage":   SendEmailToGenMessage,
		"SendEmailMessage":        SendEmailMessage,
		"RedirectTo":              RedirectTo,
		"Sleep":                   Sleep,
		"CreateJWTToken":          CreateJWTToken,
		"X2faSetup":               X2faSetup,
		"X2faValidateToken":       X2faValidateToken,
		"X2faStash":               X2faStash,
		"X2faSetupPt2of2":         X2faSetupPt2of2,
		"DumpRV":                  DumpRV,
		// -- add support for "push-to-login" at this point.
		// "ChargeCreditCard":        ChargeCreditCard,
	}
}

// FuncMapExtend will add a new named fucntion to the set of callable functions.  This allows new modules to be
// build and use the pre-post processing.
func FuncMapExtend(name string, fx FuncMapType) (err error) {
	if _, ok := funcMap[name]; ok {
		err = fmt.Errorf("Invalid - %s is already defined\n", name)
	}
	funcMap[name] = fx
	return
}

// CallFunction will call a pre-post processing function.  This is the palce where the PrePost constants will need to be handled.
// This function also reports to the log any attempts to call a non-existent function.
func (hdlr *TabServer2Type) CallFunction(ba string, fx_name string, res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx) ( /*rv*/ string /*exit*/, bool /*status*/, int) {
	var exit bool = false
	var a_status int = 200
	if fx, ok := funcMap[fx_name]; ok {
		rv, exit, a_status = fx(res, req, cfgTag, rv, isError, cookieList, ps, trx, hdlr)
	} else {
		trx.AddNote(2, fmt.Sprintf("Error(10010): Invalid internal configuration.  A called function %s has not been provided in the Go code. sql-cfg.json[%s].", fx_name, cfgTag))
		// xyzzy - really should report to log and to user this at this point! -- This sould return in an empty "rv"
		exit = true
		a_status = 501
	}
	return rv, exit, a_status
}

/* vim: set noai ts=4 sw=4: */
