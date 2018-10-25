//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1254
//

// Package jsonp impements ErrorReturn middleware
//

package ErrorReturn

import (
	"fmt"
	"net/http"
	"os"

	JsonX "github.com/pschlump/JSONx"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/MiscLib"
)

// --------------------------------------------------------------------------------------------------------------------------

/*
Error (00011): Unable to initialize module ErrorReturn in server http://www.redux-react-class.com, 1 error(s) decoding:

* 'ErrorCode': source data must be an array or slice, got int
Error (00011): Unable to initialize module ErrorReturn in server http://www.redux-react-class.com, 1 error(s) decoding:

* 'ErrorCode': source data must be an array or slice, got int
Error (00011): Unable to initialize module ErrorReturn in server http://www.redux-react-class.com, 1 error(s) decoding:

* 'ErrorCode': source data must be an array or slice, got int
*/

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &ErrorReturnHandlerType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("ErrorReturn", CreateEmpty, `{
		"Paths":              { "type":[ "string", "filepath" ], "isarray":true, "required":true },
		"ErrorCode":          { "type":[ "int" ], "isarray":true },
		"LineNo":             { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *ErrorReturnHandlerType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *ErrorReturnHandlerType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*ErrorReturnHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type ErrorReturnHandlerType struct {
	Next      http.Handler //
	Paths     []string     // Paths that this will work for
	ErrorCode []int        //
	LineNo    int          //
}

func NewErrorReturnServer(n http.Handler, p []string, ec []int) *ErrorReturnHandlerType {
	return &ErrorReturnHandlerType{
		Next:      n,
		Paths:     p,
		ErrorCode: ec,
	}
}

func (hdlr *ErrorReturnHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "ErrorReturn", hdlr.Paths, pn, req.URL.Path)

			if pn < len(hdlr.ErrorCode) {
				rw.WriteHeader(hdlr.ErrorCode[pn])
			} else {
				rw.WriteHeader(http.StatusInternalServerError)
			}

		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		hdlr.Next.ServeHTTP(www, req)
	}

}

/* vim: set noai ts=4 sw=4: */
