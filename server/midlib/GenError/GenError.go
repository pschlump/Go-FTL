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

	"www.2c-why.com/JsonX"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
)

// --------------------------------------------------------------------------------------------------------------------------
//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*GenErrorType)
//		if ok {
//			pCfg.SetNext(next)
//			rv = pCfg
//			pCfg.outputFile = os.Stdout
//		} else {
//			err = mid.FtlConfigError
//			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
//		}
//		return
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} { return &GenErrorType{} }
//
//	postInit := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
//
//		hh, ok := h.(*GenErrorType)
//		if !ok {
//			fmt.Printf("Error: Wrong data type passed to GenErrorType - postInit\n")
//			return mid.ErrInternalError
//		} else {
//			_ = hh
//			//if hh.FileName != "" {
//			//	var err error
//			//	hh.outputFile, err = lib.Fopen(hh.FileName, "a")
//			//	if err != nil {
//			//		fmt.Printf("Error: Unable to open %s for append, Error: %s Line No:%d\n", hh.FileName, err, hh.LineNo)
//			//		return mid.ErrInternalError
//			//	}
//			//}
//		}
//
//		return nil
//	}
//
//	cfg.RegInitItem2("GenError", initNext, createEmptyType, postInit, `{
//		"Paths":        { "type":["string","filepath"], "isarray":true, "default":"/api/gen-error" },
//		"StatusCode":   { "type":[ "int" ], "default":"406" },
//		"LineNo":       { "type":[ "int" ], "default":"1" }
//		}`)
//}
//
//// normally identical
//func (hdlr *GenErrorType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &GenErrorType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("GenError", CreateEmpty, `{
		"Paths":        { "type":["string","filepath"], "isarray":true, "default":"/api/gen-error" },
		"StatusCode":   { "type":[ "int" ], "default":"406" },
		"LineNo":       { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *GenErrorType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *GenErrorType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*GenErrorType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type GenErrorType struct {
	Next       http.Handler
	Paths      []string
	StatusCode int
	LineNo     int
	outputFile *os.File
}

// Parameterized for testing? or just change the test
func NewGenErrorServer(n http.Handler, p []string, e int) *GenErrorType {
	return &GenErrorType{Next: n, Paths: p, StatusCode: e}
}

func (hdlr *GenErrorType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
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
