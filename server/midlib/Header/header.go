//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2015-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1248
//

//
// Package header allows setting of additional headers
//

package Header

import (
	"fmt"
	"net/http"
	"os"

	"www.2c-why.com/JsonX"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/tmplp"
	"github.com/pschlump/MiscLib"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*HeaderType)
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
//	createEmptyType := func() interface{} { return &HeaderType{} }
//
//	cfg.RegInitItem2("Header", initNext, createEmptyType, nil, `{
//		}`)
//}
//
//// normally identical
//func (hdlr *HeaderType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &HeaderType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("Header", CreateEmpty, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"Name":          { "type":[ "string" ], "required":true },
		"Value":         { "type":[ "string" ], "required":true },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *HeaderType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *HeaderType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*HeaderType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

// xyzzy - need to template translate value/name before use!
// xyzzy - need to be able to delete headers from lower level!

type HeaderType struct {
	Next   http.Handler
	Paths  []string
	Name   string // if Name starts with "-" then delete existing header before creating new one.
	Value  string // if Value is "" then do not set header.
	LineNo int
}

func NewHeaderServer(n http.Handler, p []string, h, v string) *HeaderType {
	return &HeaderType{Next: n, Paths: p, Name: h, Value: v}
}

func (hdlr *HeaderType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "Header", hdlr.Paths, pn, req.URL.Path)

			hdlr.Next.ServeHTTP(rw, req)
			h := hdlr.Name
			if hdlr.Name[0] == '-' {
				h = hdlr.Name[1:]
				rw.Header().Del(hdlr.Name)
			}
			if hdlr.Value != "" {
				rw.Header().Set(h, tmplp.TemplateProcess(hdlr.Value, rw, req, make(map[string]string)))
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
