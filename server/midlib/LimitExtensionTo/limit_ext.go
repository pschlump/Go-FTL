//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1258
//

package LimitExtensionTo

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {

	// normally identical
	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
		pCfg, ok := ppCfg.(*LimitExtType)
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
	createEmptyType := func() interface{} { return &LimitExtType{} }

	cfg.RegInitItem2("LimitExtensionTo", initNext, createEmptyType, nil, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"Extensions":    { "type":[ "string"], "isarray":true },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

// normally identical
func (hdlr *LimitExtType) SetNext(next http.Handler) {
	hdlr.Next = next
}

var _ mid.GoFTLMiddleWare = (*LimitExtType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type LimitExtType struct {
	Next       http.Handler
	Paths      []string
	Extensions []string
	LineNo     int
}

func NewLimitExtServer(n http.Handler, p []string, e []string) *LimitExtType {
	return &LimitExtType{Next: n, Paths: p, Extensions: e}
}

func (hdlr LimitExtType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "LimitExtensionTo", hdlr.Paths, pn, req.URL.Path)

			// extract extension
			ext := ".html"
			if !strings.HasSuffix(req.URL.Path, "/") {
				ext = filepath.Ext(req.URL.Path)
			}
			// if extension in Extensions - then reject
			// fmt.Printf("ext >%s< limit >%+v< TF=%v, %s\n", ext, hdlr.Extensions, lib.InArray(ext, hdlr.Extensions), lib.LF())
			if lib.InArray(ext, hdlr.Extensions) {
				// fmt.Printf("   Serve it\n")
				hdlr.Next.ServeHTTP(www, req)
				return
			} else {
				// fmt.Printf("   *** Reject *** it\n")
				www.WriteHeader(http.StatusNotFound)
			}

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
