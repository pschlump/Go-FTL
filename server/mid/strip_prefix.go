//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2015-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1200
//

//
// Package header remove prefix - usually prior to proxy or file server
//
// Copyright (C) Philip Schlump, 2016
//

package mid

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/lib"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {

	// normally identical
	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
		pCfg, ok := ppCfg.(*StripPrefixType)
		if ok {
			pCfg.SetNext(next)
			rv = pCfg
		} else {
			err = FtlConfigError
		}
		return
	}

	// normally identical
	createEmptyType := func() interface{} { return &StripPrefixType{} }

	cfg.RegInitItem2("strip_prefix", initNext, createEmptyType, nil, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"Prefix":        { "type":[ "string" ], "required":true },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

// normally identical
func (hdlr *StripPrefixType) SetNext(next http.Handler) {
	hdlr.Next = next
}

var _ GoFTLMiddleWare = (*StripPrefixType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type StripPrefixType struct {
	Next    http.Handler //
	Paths   []string     //
	Prefix  string       // regular expression
	matchre *regexp.Regexp
}

func NewStripPrefixServer(n http.Handler, p []string, h, v string) *StripPrefixType {
	return &StripPrefixType{Next: n, Paths: p, Prefix: h}
}

func (hdlr StripPrefixType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if lib.PathsMatch(hdlr.Paths, req.URL.Path) {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, hdlr.Prefix)
			req.RequestURI = strings.TrimPrefix(req.RequestURI, hdlr.Prefix)
			hdlr.Next.ServeHTTP(rw, req)
		} else {
			fmt.Fprintf(os.Stderr, "%s\n", ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		hdlr.Next.ServeHTTP(www, req)
	}
}

/* vim: set noai ts=4 sw=4: */
