//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1240
//

//
// Seems to be a "if not served by existing servers then, "else" server"
//

package Else

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"www.2c-why.com/JsonX"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical - not this time - pre-builds Body
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*ElseType)
//		if ok {
//			pCfg.SetNext(next)
//			//pCfg.Body = "<!DOCTYPE html><html><head></head><body>%s<ul>\n"
//			//for _, vv := range gCfg.Config {
//			//	for _, ww := range vv.ListenTo {
//			//		pCfg.Body += fmt.Sprintf("\t<li><a href=\"%s\">%s: %s</a></li>\n", ww, vv.Name, ww)
//			//	}
//			//}
//			//pCfg.Body += "</ul></body></html>"
//			//if db1 {
//			//	fmt.Printf(">>>>>%s<<<<<, %s\n", pCfg.Body, godebug.LF())
//			//}
//			genBody(pCfg, gCfg)
//			rv = pCfg
//		} else {
//			err = mid.FtlConfigError
//			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
//		}
//		return
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} { return &ElseType{} }
//
//	cfg.RegInitItem2("Else", initNext, createEmptyType, nil, `{
//		"Paths":         { "type":["string","filepath"], "isarray":true },
//		"Msg":           { "type":["string"] },
//		"LineNo":        { "type":[ "int" ], "default":"1" }
//		}`)
//}
//
//// normally identical
//func (hdlr *ElseType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &ElseType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("Else", CreateEmpty, `{
		}`)
}

func (hdlr *ElseType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	genBody(hdlr, gCfg)
	return
}

func (hdlr *ElseType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*ElseType)(nil)

// --------------------------------------------------------------------------------------------------------------------------
// xyzzy - Host Name Resolution - for HTTP 1.0 or IP access - list. See /Users/corwin/go/src/github.com/pschlump/Go-FTL/server/goftl/goftl_cli.go

type ElseType struct {
	Next  http.Handler
	Paths []string
	Msg   string
	Body  string
}

func NewElseServer(n http.Handler, p []string) *ElseType {
	x := &ElseType{Next: n, Paths: p}
	cfg.SetupEmptyForTest()
	genBody(x, cfg.ServerGlobal)
	return x
}

func genBody(pCfg *ElseType, gCfg *cfg.ServerGlobalConfigType) {
	pCfg.Body = "<!DOCTYPE html><html><head></head><body>%s<ul>\n"
	for _, vv := range gCfg.Config {
		for _, ww := range vv.ListenTo {
			if strings.Index(ww, "?") > 0 {
				pCfg.Body += fmt.Sprintf("\t<li><a href=\"%s&$$host_name$$=%s\">%s: %s</a></li>\n", ww, vv.Name, vv.Name, ww)
			} else {
				pCfg.Body += fmt.Sprintf("\t<li><a href=\"%s?$$host_name$$=%s\">%s: %s</a></li>\n", ww, vv.Name, vv.Name, ww)
			}
		}
	}
	pCfg.Body += "</ul></body></html>"
	if db1 {
		fmt.Printf(">>>>>%s<<<<<, %s\n", pCfg.Body, godebug.LF())
	}
}

func (hdlr ElseType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if rw, ok := www.(*goftlmux.MidBuffer); ok {

		trx := mid.GetTrx(rw)
		trx.PathMatched(1, "Else", []string{}, 0, req.URL.Path)

		fmt.Fprintf(www, hdlr.Body, hdlr.Msg)
		www.WriteHeader(http.StatusOK)

	} else {
		fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
		fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
		www.WriteHeader(http.StatusInternalServerError)
	}
}

const db1 = false

/* vim: set noai ts=4 sw=4: */
