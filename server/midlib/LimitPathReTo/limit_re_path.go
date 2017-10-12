//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1260
//

//
// Limit serving of files to the specified set of regular expression paths.  If the file is not one of the specified
// paths then reject the request.
//

package LimitPathReTo

import (
	"fmt"
	"net/http"
	"os"

	"www.2c-why.com/JsonX"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/MiscLib"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*LimitReType)
//		if ok {
//			pCfg.SetNext(next)
//			rv = pCfg
//		} else {
//			err = mid.FtlConfigError
//			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
//		}
//		return
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} { return &LimitReType{} }
//
//	cfg.RegInitItem2("LimitPathRePath", initNext, createEmptyType, nil, `{
//		}`)
//}
//
//// normally identical
//func (hdlr *LimitReType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &LimitReType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("LimitPathRePath", CreateEmpty, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *LimitReType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *LimitReType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*LimitReType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type LimitReType struct {
	Next   http.Handler
	Paths  []string
	LineNo int
}

func NewLimitRePathServer(n http.Handler, p []string) *LimitReType {
	return &LimitReType{Next: n, Paths: p}
}

func (hdlr *LimitReType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchReN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "LimitPathReTo", hdlr.Paths, pn, req.URL.Path)

			hdlr.Next.ServeHTTP(www, req)

		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		www.WriteHeader(http.StatusNotFound)
	}
}

/* vim: set noai ts=4 sw=4: */
