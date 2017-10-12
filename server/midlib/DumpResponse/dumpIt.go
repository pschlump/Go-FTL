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

package DumpResponse

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	JsonX "github.com/pschlump/JSONx"

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
//		pCfg, ok := ppCfg.(*DumpRequestType)
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
//	createEmptyType := func() interface{} { return &DumpRequestType{} }
//
//	postInit := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
//
//		hh, ok := h.(*DumpRequestType)
//		if !ok {
//			fmt.Printf("Error: Wrong data type passed to DumpRequestType - postInit\n")
//			return mid.ErrInternalError
//		} else {
//			if hh.FileName != "" {
//				var err error
//				hh.outputFile, err = lib.Fopen(hh.FileName, "a")
//				if err != nil {
//					fmt.Printf("Error: Unable to open %s for append, Error: %s Line No:%d\n", hh.FileName, err, hh.LineNo)
//					return mid.ErrInternalError
//				}
//			}
//		}
//
//		return nil
//	}
//
//	cfg.RegInitItem2("DumpResponse", initNext, createEmptyType, postInit, `{
//		"Paths":        { "type":["string","filepath"], "isarray":true, "default":"/" },
//		"Msg":          { "type":[ "string" ], "default":"" },
//		"SaveBodyFlag": { "type":[ "bool" ], "default":"false" },
//		"SaveTextOnly": { "type":[ "bool" ], "default":"true" },
//		"FileName":     { "type":[ "string","filepath" ], "default":"" },
//		"LineNo":       { "type":[ "int" ], "default":"1" }
//		}`)
//}
//
//// normally identical
//func (hdlr *DumpRequestType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &DumpRequestType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("DumpResponse", CreateEmpty, `{
		"Paths":        { "type":["string","filepath"], "isarray":true, "default":"/" },
		"Msg":          { "type":[ "string" ], "default":"" },
		"SaveBodyFlag": { "type":[ "bool" ], "default":"false" },
		"SaveTextOnly": { "type":[ "bool" ], "default":"true" },
		"FileName":     { "type":[ "string","filepath" ], "default":"" },
		"LineNo":       { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *DumpRequestType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *DumpRequestType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	if hdlr.FileName != "" {
		var err error
		hdlr.outputFile, err = lib.Fopen(hdlr.FileName, "a")
		if err != nil {
			fmt.Printf("Error: Unable to open %s for append, Error: %s Line No:%d\n", hdlr.FileName, err, hdlr.LineNo)
			return mid.ErrInternalError
		}
	}
	return
}

var _ mid.GoFTLMiddleWare = (*DumpRequestType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type DumpRequestType struct {
	Next         http.Handler
	Paths        []string
	Msg          string
	SaveBodyFlag bool
	SaveTextOnly bool
	FileName     string
	LineNo       int
	outputFile   *os.File
}

// Parameterized for testing? or just change the test
func NewDumpRequestServer(n http.Handler, p []string, m string, sb bool, fn string) *DumpRequestType {
	return &DumpRequestType{Next: n, Paths: p, Msg: m, SaveBodyFlag: sb, FileName: fn}
}

func (hdlr *DumpRequestType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "DumpResponse", hdlr.Paths, pn, req.URL.Path)

			hdlr.Next.ServeHTTP(rw, req)

			fmt.Fprintf(hdlr.outputFile, "DumpResponse %s\n\tStatusCode=%d\n", hdlr.Msg, rw.StatusCode)
			h := rw.GetHeader()
			fmt.Fprintf(hdlr.outputFile, "\tResponse Header=%s\n", lib.SVarI(h))
			if hdlr.SaveBodyFlag {
				// Should be "if" to dump - and dump to file body - based on call number - in ./dumpIt dir
				s := rw.GetBody()
				if !hdlr.SaveTextOnly {
					fmt.Fprintf(hdlr.outputFile, "\tBody=>>>%s%s%s<<<\n", rw.Prefix, s, rw.Postfix)
				} else {
					if IsTextContentType(rw) {
						fmt.Fprintf(hdlr.outputFile, "\tBody=>>>%s%s%s<<<\n", rw.Prefix, s, rw.Postfix)
					} else {
						fmt.Fprintf(hdlr.outputFile, "\tBody is binary image of length %d\n", len(s))
					}
				}
			}

		} else {
			logrus.Warn(fmt.Sprintf("Error: DumpResponse: %s\n", mid.ErrNonMidBufferWriter))
			www.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		hdlr.Next.ServeHTTP(www, req)
	}

}

// ----------------------------------------------------------------------------------------------------------------------------------------------------
func IsTextContentType(rw *goftlmux.MidBuffer) (isText bool) {
	isText = false
	if len(rw.Headers) > 0 {
		for key, val := range rw.Headers {
			if key == "Content-Type" {
				for _, ss := range val {
					if strings.HasPrefix(ss, "text/") {
						isText = true
						return
					}
				}
			}
		}
	}
	return
}

/* vim: set noai ts=4 sw=4: */
