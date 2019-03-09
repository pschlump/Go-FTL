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

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/godebug"
	//	"encoding/json"
)

//
//	Error		Meaning
//	-----		--------------------------------------
//	400			Bad Request
//	401			Unauthorized
//	402			Payment Required
//	403			Forbidden
//	404			Not Found
//	405			Method Not Allowed
//	406			Not Acceptable
//	412			Precondition Failed
//	417			Expectation Failed
//	428			Precondition Required
//
/*
	http.
		StatusBadRequest                   = 400
		StatusUnauthorized                 = 401
		StatusPaymentRequired              = 402
		StatusForbidden                    = 403
		StatusNotFound                     = 404
		StatusMethodNotAllowed             = 405
		StatusNotAcceptable                = 406
		StatusProxyAuthRequired            = 407
		StatusRequestTimeout               = 408
		StatusConflict                     = 409
		StatusGone                         = 410
		StatusLengthRequired               = 411
		StatusPreconditionFailed           = 412
		StatusRequestEntityTooLarge        = 413
		StatusRequestURITooLong            = 414
		StatusUnsupportedMediaType         = 415
		StatusRequestedRangeNotSatisfiable = 416
		StatusExpectationFailed            = 417
		StatusTeapot                       = 418

		StatusInternalServerError     = 500
		StatusNotImplemented          = 501
		StatusBadGateway              = 502
		StatusServiceUnavailable      = 503
		StatusGatewayTimeout          = 504
		StatusHTTPVersionNotSupported = 505

*/
func ConvertErrorToCode(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rvOut string, pptFlag PrePostFlagType, exit bool, a_status int) {
	exit = false
	a_status = 200
	x, err := sizlib.JSONStringToData(rv)
	if err != nil {
		// rv = fmt.Sprintf(`{ "status":"error","msg":"Error(10009): Parsing return value failed. sql-cfg.json[%s] Post Function Call(CacheEUser)",%s, "err":%q }`, cfgTag, godebug.LFj(), err)
		// res.Header().Set("Content-Type", "text/html")
		// http.Error(res, "400 Bad Request", http.StatusBadRequest)
		ReturnErrorMessage(400, "Bad Request", "19043",
			fmt.Sprintf(`Error(19043): Bad Request (%s) sql-cfg.json[%s] %s %s`, sizlib.EscapeError(err), cfgTag, err, godebug.LF()),
			res, req, *ps, trx, hdlr) // status:error
		exit = true
		a_status = 500
		return "", PrePostFatalSetStatus, true, 500
	}
	if GetSI("status", x) != "success" || isError {
		// res.Header().Set("Content-Type", "text/html")
		// http.Error(res, "406 Not Acceptable", http.StatusNotAcceptable)
		ReturnErrorMessage(406, "Not Acceptable", "19044",
			fmt.Sprintf(`Error(19044): Not Acceptable (%s) sql-cfg.json[%s] %s %s`, sizlib.EscapeError(err), cfgTag, err, godebug.LF()),
			res, req, *ps, trx, hdlr) // status:error
		exit = true
		a_status = 406
	}
	return rv, PrePostContinue, exit, a_status
}

/* vim: set noai ts=4 sw=4: */
