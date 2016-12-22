//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1238
//

//
// A echo-like call, /api/echo usually
//

package Echo

import (
	"fmt"
	"io"
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
		pCfg, ok := ppCfg.(*EchoType)
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
	createEmptyType := func() interface{} { return &EchoType{} }

	cfg.RegInitItem2("Echo", initNext, createEmptyType, nil, `{
		"Paths":         { "type":["string","filepath"], "isarray":true },
		"Msg":           { "type":["string"] },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

// normally identical
func (hdlr *EchoType) SetNext(next http.Handler) {
	hdlr.Next = next
}

var _ mid.GoFTLMiddleWare = (*EchoType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type EchoType struct {
	Next  http.Handler
	Paths []string
	Msg   string
}

func NewEchoServer(n http.Handler, p []string) *EchoType {
	return &EchoType{Next: n, Paths: p}
}

func (hdlr EchoType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if rw, ok := www.(*goftlmux.MidBuffer); ok {

		trx := mid.GetTrx(rw)
		trx.PathMatched(1, "Echo", []string{}, 0, req.URL.Path)

		s := fmt.Sprintf("%s\n", lib.SVarI(req))
		// s += fmt.Sprintf("%#v\n", req)
		fmt.Fprintf(www, "Msg: %s\n", hdlr.Msg)
		io.WriteString(www, s)
		www.WriteHeader(http.StatusOK)

	} else {
		fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
		fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
		www.WriteHeader(http.StatusInternalServerError)
	}
}

/* vim: set noai ts=4 sw=4: */
