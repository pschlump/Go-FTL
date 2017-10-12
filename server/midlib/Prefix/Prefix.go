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

	JsonX "github.com/pschlump/JSONx"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/MiscLib"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &PrefixHandlerType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		// "Prefix":             { "type":["string"], "default":")]}',`+"\n"+`" },
		return x
	}
	mid.RegInitItem3("Prefix", CreateEmpty, `{
		"Paths":              { "type":["string","filepath"], "isarray":true, "required":true },
		"Prefix":             { "type":["string"], "default":"while(1);" },
		"PrePend":            { "type":["string"], "default":"yes" },
		"LineNo":             { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *PrefixHandlerType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *PrefixHandlerType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*PrefixHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type PrefixHandlerType struct {
	Next    http.Handler //
	Paths   []string     // Paths that this will work for
	Prefix  string       // Prefix to put before JSON responses
	PrePend string       //	"set", "yes"=="before"==PrePend, "after"
	LineNo  int          //
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
				if hdlr.PrePend == "yes" || hdlr.PrePend == "before" {
					rw.Prefix = hdlr.Prefix + rw.Prefix
				} else if hdlr.PrePend == "after" {
					rw.Prefix = rw.Prefix + hdlr.Prefix
				} else {
					rw.Prefix = hdlr.Prefix
					rw.Postfix = ""
				}
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
