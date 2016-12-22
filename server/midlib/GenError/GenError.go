//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1236
//

// Package dumpresponse allows for dumping of what is being returned in the stack.

package GenError

import (
	"fmt"
	"net/http"
	"os"

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
		pCfg, ok := ppCfg.(*DumpRequestType)
		if ok {
			pCfg.SetNext(next)
			rv = pCfg
			pCfg.outputFile = os.Stdout
		} else {
			err = mid.FtlConfigError
			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
		}
		return
	}

	// normally identical
	createEmptyType := func() interface{} { return &DumpRequestType{} }

	postInit := func(h interface{}, cfgData map[string]interface{}, callNo int) error {

		hh, ok := h.(*DumpRequestType)
		if !ok {
			fmt.Printf("Error: Wrong data type passed to DumpRequestType - postInit\n")
			return mid.ErrInternalError
		} else {
			_ = hh
			//if hh.FileName != "" {
			//	var err error
			//	hh.outputFile, err = lib.Fopen(hh.FileName, "a")
			//	if err != nil {
			//		fmt.Printf("Error: Unable to open %s for append, Error: %s Line No:%d\n", hh.FileName, err, hh.LineNo)
			//		return mid.ErrInternalError
			//	}
			//}
		}

		return nil
	}

	cfg.RegInitItem2("GenError", initNext, createEmptyType, postInit, `{
		"Paths":        { "type":["string","filepath"], "isarray":true, "default":"/api/gen-error" },
		"StatusCode":   { "type":[ "int" ], "default":"406" },
		"LineNo":       { "type":[ "int" ], "default":"1" }
		}`)
}

// normally identical
func (hdlr *DumpRequestType) SetNext(next http.Handler) {
	hdlr.Next = next
}

var _ mid.GoFTLMiddleWare = (*DumpRequestType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type DumpRequestType struct {
	Next       http.Handler
	Paths      []string
	StatusCode int
	LineNo     int
	outputFile *os.File
}

// Parameterized for testing? or just change the test
func NewGenErrorServer(n http.Handler, p []string, e int) *DumpRequestType {
	return &DumpRequestType{Next: n, Paths: p, StatusCode: e}
}

func (hdlr *DumpRequestType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "GenError", hdlr.Paths, pn, req.URL.Path)

			www.WriteHeader(hdlr.StatusCode)

		} else {
			logrus.Warn(fmt.Sprintf("Error: GenError: %s\n", mid.ErrNonMidBufferWriter))
			www.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		hdlr.Next.ServeHTTP(www, req)
	}

}

/* vim: set noai ts=4 sw=4: */
