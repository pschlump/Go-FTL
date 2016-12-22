//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1262
//

package LimitPathTo

import (
	"fmt"
	"net/http"
	"os"

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
		pCfg, ok := ppCfg.(*LimitPathType)
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
	createEmptyType := func() interface{} { return &LimitPathType{} }

	cfg.RegInitItem2("LimitPathTo", initNext, createEmptyType, nil, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

// normally identical
func (hdlr *LimitPathType) SetNext(next http.Handler) {
	hdlr.Next = next
}

var _ mid.GoFTLMiddleWare = (*LimitPathType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type LimitPathType struct {
	Next   http.Handler
	Paths  []string
	LineNo int
}

func NewLimitPathServer(n http.Handler, p []string) *LimitPathType {
	return &LimitPathType{Next: n, Paths: p}
}

func (hdlr *LimitPathType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "LimitPathTo", hdlr.Paths, pn, req.URL.Path)

			hdlr.Next.ServeHTTP(www, req)

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
