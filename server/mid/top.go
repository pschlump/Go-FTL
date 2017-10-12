//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1202
//

package mid

import (
	"fmt"
	"net/http"
	"strings"

	JsonX "github.com/pschlump/JSONx"

	"github.com/pschlump/Go-FTL/server/cfg"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*MidServer)
//		if ok {
//			pCfg.SetNext(next)
//			pCfg.callNo = 1 // not boilerplate
//			rv = pCfg
//		} else {
//			err = FtlConfigError
//		}
//		return
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} { return &MidServer{} }
//
//	cfg.RegInitItem2("x_top", initNext, createEmptyType, nil, `{
//		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
//		"Info":          { "type":[ "string","filepath" ], "isarray":true },
//		"LineNo":        { "type":[ "int" ], "default":"1" }
//		}`)
//}
//
//// normally identical
//func (hdlr *MidServer) SetNext(next http.Handler) {
//	// not boilerplate
//}

func init() {
	CreateEmpty := func(name string) GoFTLMiddleWare {
		x := &MidServer{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	RegInitItem3("Gzip", CreateEmpty, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"Info":          { "type":[ "string","filepath" ], "isarray":true },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *MidServer) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	//hdlr.Next = next
	hdlr.callNo = 1
	return
}

func (hdlr *MidServer) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ GoFTLMiddleWare = (*MidServer)(nil)

// --------------------------------------------------------------------------------------------------------------------------
// top - the top level that connects ServeHTTP with the mid package and it's calling conventions.
// Also performs a flush on the buffers at this point.

type MidServer struct {
	Info   string
	callNo int
}

func NewServer() *MidServer {
	// return &MidServer{Info: " *four* ", callNo: 1}
	return &MidServer{Info: "", callNo: 1}
}

func (ms *MidServer) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	// wr.Write([]byte(fmt.Sprintf("Hello World %s CallNo:%d", ms.Info, ms.callNo)))
	// wr.Write([]byte(ms.Info))
	if strings.Index(ms.Info, "%d") > 0 {
		fmt.Fprintf(wr, ms.Info, ms.callNo)
	} else {
		fmt.Fprintf(wr, ms.Info)
	}
	ms.callNo++
}

func (ms *MidServer) SetInfo(s string) {
	ms.Info = s
}

/*
	server := &http.Server{
		Addr:    s.address,
		Handler: s,
	}
	return server.ListenAndServe()


*/

/* vim: set noai ts=4 sw=4: */
