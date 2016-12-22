//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1254
//

// Package jsonp impements Prefix middleware
//

package Prefix

import (
	"fmt"
	"net/http"
	"os"
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
		pCfg, ok := ppCfg.(*PrefixHandlerType)
		if ok {
			pCfg.SetNext(next)
			rv = pCfg
		} else {
			err = mid.FtlConfigError
			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
		}
		return
	}

	postInit := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
		// fmt.Printf("In postInitValidation, h=%v\n", h)
		hh, ok := h.(*PrefixHandlerType)
		if !ok {
			fmt.Fprintf(os.Stderr, "%sError: Wrong data type passed, Line No:%d\n%s", MiscLib.ColorRed, hh.LineNo, MiscLib.ColorReset)
			return mid.ErrInternalError
		}
		return nil
	}

	// normally identical
	createEmptyType := func() interface{} { return &PrefixHandlerType{} }

	cfg.RegInitItem2("Prefix", initNext, createEmptyType, postInit, `{
		"Paths":              { "type":["string","filepath"], "isarray":true, "required":true },
		"Prefix":             { "type":["string"], "default":")]}',`+"\n"+`" },
		"LineNo":             { "type":[ "int" ], "default":"1" }
		}`)
}

// normally identical
func (hdlr *PrefixHandlerType) SetNext(next http.Handler) {
	hdlr.Next = next
}

var _ mid.GoFTLMiddleWare = (*PrefixHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type PrefixHandlerType struct {
	Next   http.Handler //
	Paths  []string     // Paths that this will work for
	Prefix string       // Prefix to put before JSON responses
	LineNo int          //
}

func NewPrefixServer(n http.Handler, p []string, reMatch string) *PrefixHandlerType {
	return &PrefixHandlerType{
		Next:   n,
		Paths:  p,
		Prefix: ")]}',\n",
	}
}

func (hdlr *PrefixHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "Prefix", hdlr.Paths, pn, req.URL.Path)

			hdlr.Next.ServeHTTP(rw, req)
			h := www.Header()
			ct := h.Get("Content-Type")
			trx.AddNote(1, fmt.Sprintf("Content-Type == %s StatusCode = %d", ct, rw.StatusCode))
			if rw.StatusCode == http.StatusOK && strings.HasPrefix(ct, "application/json") {
				trx.AddNote(1, "Is JSON - will add prefix")
				rw.Prefix = hdlr.Prefix
				rw.Postfix = ""
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
