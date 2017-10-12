//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1277
//

//
// Package dumpit directory browsing.   The results of browsing to a direcotry can be fead through a Go template.
//

package RejectDirectory

import (
	"net/http"

	"www.2c-why.com/JsonX"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*DirectoryLimitType)
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
//	createEmptyType := func() interface{} { return &DirectoryLimitType{} }
//
//	cfg.RegInitItem2("RejectDirectory", initNext, createEmptyType, nil, `{
//		}`)
//}
//
//// normally identical
//func (hdlr *DirectoryLimitType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &DirectoryLimitType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("RejectDirectory", CreateEmpty, `{
		"Paths":         { "type":[ "string", "filepath" ], "isarray":true, "required":true },
		"Disalow":       { "type":[ "[]string", "filepath" ] },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *DirectoryLimitType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *DirectoryLimitType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*DirectoryLimitType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type DirectoryLimitType struct {
	Next    http.Handler //
	Paths   []string     // thins directory browsing is enabled for -- Paths that are served with a directory index
	Disalow []string     // Set of Directories to dis-allow
	LineNo  int
}

// IgnoreDirectories []string     //

func NewDirectoryLimitServer(n http.Handler, p []string, dis []string) *DirectoryLimitType {
	return &DirectoryLimitType{Next: n, Paths: p, Disalow: dis}
}

func (hdlr DirectoryLimitType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "RejectDirectory", hdlr.Paths, pn, req.URL.Path)

			rw.IgnoreDirs = hdlr.Disalow

		}
	}
	hdlr.Next.ServeHTTP(www, req)

}

/* vim: set noai ts=4 sw=4: */
