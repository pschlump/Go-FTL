//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2015-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1264
//

// Package dumpit impements logging
//
// Copyright (C) Philip Schlump, 2016
//

package Logging

import (
	"fmt"
	"net/http"

	JsonX "github.com/pschlump/JSONx"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/tmplp"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*LoggingType)
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
//	createEmptyType := func() interface{} { return &LoggingType{} }
//
//	cfg.RegInitItem2("Logging", initNext, createEmptyType, nil, `{
//		}`)
//}
//
//// normally identical
//func (hdlr *LoggingType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &LoggingType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("Logging", CreateEmpty, `{
		"Paths":         { "type":[ "string", "filepath" ], "isarray":true, "required":true },
		"Format":        { "type":[ "string" ], "required":true},
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *LoggingType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *LoggingType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*LoggingType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type LoggingType struct {
	Next   http.Handler
	Paths  []string
	Format string
	LineNo int
}

func NewLoggingServer(n http.Handler, p []string, f string) *LoggingType {
	return &LoggingType{Next: n, Paths: p, Format: f}
}

func (hdlr LoggingType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "Logging", hdlr.Paths, pn, req.URL.Path)

			hdlr.Next.ServeHTTP(rw, req)

			logrus.Info(fmt.Sprintf("%s\n", tmplp.TemplateProcess(hdlr.Format, rw, req, make(map[string]string))))

		} else {
			logrus.Warn(fmt.Sprintf("Error: DumpResponse: %s\n", mid.ErrNonMidBufferWriter))
			www.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		hdlr.Next.ServeHTTP(www, req)
	}
}

/* vim: set noai ts=4 sw=4: */
