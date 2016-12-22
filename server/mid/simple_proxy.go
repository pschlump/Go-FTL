//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2015-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1199
//

//
// Package proxy impements a reverse proxy to a single server
//
// Copyright (C) Philip Schlump, 2016
//

package mid

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/lib"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {

	// normally identical
	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
		pCfg, ok := ppCfg.(*SimpleProxyHandlerType)
		if ok {
			pCfg.SetNext(next)
			rv = pCfg
		} else {
			err = FtlConfigError
		}
		return
	}

	// normally identical
	createEmptyType := func() interface{} { return &SimpleProxyHandlerType{} }

	cfg.RegInitItem2("simple_proxy", initNext, createEmptyType, nil, `{
		"Paths":        { "type":["string","filepath"], "isarray":true, "required":true },
		"Dest":         { "type":["string","url"], "required":true },
		"Extra":        { "allowed":false },
		"LineNo":       { "type":[ "int" ], "default":"1" }
		}`)
}

// normally identical
func (hdlr *SimpleProxyHandlerType) SetNext(next http.Handler) {
	hdlr.Next = next
}

var _ GoFTLMiddleWare = (*SimpleProxyHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type SimpleProxyHandlerType struct {
	Next     http.Handler
	Paths    []string
	Dest     string
	LineNo   int
	theProxy http.Handler
}

func NewSimpleProxyServer(n http.Handler, p []string, d string) *SimpleProxyHandlerType {
	return &SimpleProxyHandlerType{Next: n, Paths: p, Dest: d}
}

func (hdlr SimpleProxyHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if lib.PathsMatch(hdlr.Paths, req.URL.Path) {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			if hdlr.theProxy == nil {
				dest, err := url.Parse(hdlr.Dest)
				if err != nil {
					rw.Error = err
					www.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					hdlr.theProxy = httputil.NewSingleHostReverseProxy(dest)
				}
			}

			hdlr.theProxy.ServeHTTP(www, req)

		} else {
			fmt.Fprintf(os.Stderr, "%s\n", ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		if hdlr.Next != nil {
			hdlr.Next.ServeHTTP(www, req)
		}
	}

}

/* vim: set noai ts=4 sw=4: */
