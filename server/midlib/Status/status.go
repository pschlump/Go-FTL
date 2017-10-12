//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2017
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1291
//

//
// A echo-like call, /api/status usually
//

package Status

import (
	"fmt"
	"io"
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
//		pCfg, ok := ppCfg.(*StatusType)
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
//	createEmptyType := func() interface{} { return &StatusType{} }
//
//	cfg.RegInitItem2("Status", initNext, createEmptyType, nil, `{
//		}`)
//}
//
//// normally identical
//func (hdlr *StatusType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &StatusType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("Status", CreateEmpty, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"Fmt":           { "type":["string"], "default":"JSON" },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *StatusType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *StatusType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	// If callNo == 0, then this is a 1st call -- it will count up.
	// fmt.Fprintf(os.Stderr, "%sStatus: %d%s\n", MiscLib.ColorCyan, callNo, MiscLib.ColorReset)
	return
}

var _ mid.GoFTLMiddleWare = (*StatusType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type StatusType struct {
	Next  http.Handler
	Paths []string
	Fmt   string
}

func NewStatusServer(n http.Handler, p []string, fmt string) *StatusType {
	return &StatusType{Next: n, Paths: p, Fmt: fmt}
}

func (hdlr *StatusType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "Status", hdlr.Paths, pn, req.URL.Path)

			if "JSON" == hdlr.Fmt {
				www.Header().Set("Content-Type", "application/json")
				s := fmt.Sprintf("%s\n", lib.SVarI(req))
				io.WriteString(www, s)
				www.WriteHeader(http.StatusOK)
			}

			return
		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
		}
	}
	hdlr.Next.ServeHTTP(www, req)
}

/* vim: set noai ts=4 sw=4: */
