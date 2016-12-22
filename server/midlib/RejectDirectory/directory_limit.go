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

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {

	// normally identical
	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
		pCfg, ok := ppCfg.(*DirectoryLimitType)
		if ok {
			pCfg.SetNext(next)
			rv = pCfg
		} else {
			err = mid.FtlConfigError
			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
		}
		return
	}

	// normally identical
	createEmptyType := func() interface{} { return &DirectoryLimitType{} }

	cfg.RegInitItem2("RejectDirectory", initNext, createEmptyType, nil, `{
		"Paths":         { "type":[ "string", "filepath" ], "isarray":true, "required":true },
		"Disalow":       { "type":[ "[]string", "filepath" ] },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

// normally identical
func (hdlr *DirectoryLimitType) SetNext(next http.Handler) {
	hdlr.Next = next
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

/*

../../fileserve/fs.go Line:220 func dirList(w http.ResponseWriter, f File) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

*/

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
