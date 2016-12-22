//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2015-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1271
//

//
// Package header redirect HTTP request to HTTPS
//
// Copyright (C) Philip Schlump, 2016
//

package RedirToHttps

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {

	// normally identical
	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
		pCfg, ok := ppCfg.(*RedirectToHTTPSHandlerType)
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
	createEmptyType := func() interface{} { return &RedirectToHTTPSHandlerType{} }

	cfg.RegInitItem2("RedirectToHTTPS", initNext, createEmptyType, nil, `{
		"To":     		 { "type":[ "string" ] },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

// normally identical
func (hdlr *RedirectToHTTPSHandlerType) SetNext(next http.Handler) {
	hdlr.Next = next
}

var _ mid.GoFTLMiddleWare = (*RedirectToHTTPSHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type RedirectToHTTPSHandlerType struct {
	Next http.Handler //
	To   string       // replacement string, with ${1} pattern replacements in it
}

func NewRedirectToHTTPSServer(n http.Handler, t string) *RedirectToHTTPSHandlerType {
	return &RedirectToHTTPSHandlerType{Next: n, To: t}
}

func (hdlr *RedirectToHTTPSHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if rw, ok := www.(*goftlmux.MidBuffer); ok {

		trx := mid.GetTrx(rw)
		trx.PathMatched(1, "RedirToHttps", []string{}, 0, req.URL.Path)

		www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
		www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
		www.Header().Set("Expires", "0")                                         // Proxies.
		if hdlr.To != "" {
			if strings.HasSuffix(hdlr.To, "/") {
				www.Header().Set("Location", "https://"+hdlr.To+req.RequestURI) // xyzzy - what about trailing slash in To[0]
			} else {
				www.Header().Set("Location", "https://"+hdlr.To+"/"+req.RequestURI) // xyzzy - what about trailing slash in To[0]
			}
		} else {
			www.Header().Set("Location", "https://"+req.Host+req.RequestURI)
		}
		www.WriteHeader(http.StatusTemporaryRedirect) // 307

	} else {
		fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
		fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
		www.WriteHeader(http.StatusInternalServerError)
	}
}

/* vim: set noai ts=4 sw=4: */
