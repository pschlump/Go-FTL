//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1285
//

//
// Package that uses regular expressions to match paths.  When they match reject the request.
//

package RejectRePath

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

//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*RejectPathReType)
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
//	createEmptyType := func() interface{} { return &RejectPathReType{} }
//
//	cfg.RegInitItem2("RejectRePath", initNext, createEmptyType, nil, `{
//		}`)
//}
//
//// normally identical
//func (hdlr *RejectPathReType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &RejectPathReType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("RejectRePath", CreateEmpty, `{
		"Paths":         { "type":[ "string", "filepath" ], "isarray":true, "required":true },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *RejectPathReType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *RejectPathReType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*RejectPathReType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type RejectPathReType struct {
	Next  http.Handler
	Paths []string
	//rePaths []*regexp.Regexp
}

func NewRejectRePathServer(n http.Handler, p []string) *RejectPathReType {
	return &RejectPathReType{Next: n, Paths: p}
}

func (hdlr RejectPathReType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	//if len(hdlr.Paths) != len(hdlr.rePaths) {
	//	for ii, vv := range hdlr.Paths {
	//		hdlr.rePaths[ii] = regexp.MustCompile( vv)
	//	}
	//}
	// if lib.PathsMatchRe(hdlr.Paths, req.URL.Path) {
	if pn := lib.PathsMatchReN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "RejectRePath", hdlr.Paths, pn, req.URL.Path)

			www.WriteHeader(http.StatusNotFound)
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
