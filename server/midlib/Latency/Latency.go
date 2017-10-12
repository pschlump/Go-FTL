//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1234
//

//
// Package Latency - add latency to requests to simulate slow networks.
//

package Latency

import (
	"fmt"
	"net/http"
	"time"

	"www.2c-why.com/JsonX"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
)

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &LatencyType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("Latency", CreateEmpty, `{
		"Paths":        { "type":["string","filepath"], "isarray":true, "default":"/" },
		"SlowDown":     { "type":[ "int" ], "default":"500" },
		"LineNo":       { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *LatencyType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *LatencyType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	hdlr.slowDown = time.Duration(int64(hdlr.SlowDown)) * time.Millisecond
	return
}

var _ mid.GoFTLMiddleWare = (*LatencyType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type LatencyType struct {
	Next     http.Handler
	Paths    []string
	SlowDown int
	LineNo   int
	slowDown time.Duration
}

// Parameterized for testing? or just change the test
func NewLatencyServer(n http.Handler, p []string, slow int) *LatencyType {
	return &LatencyType{
		Next:     n,
		Paths:    p,
		SlowDown: slow,
		slowDown: time.Duration(int64(slow)) * time.Millisecond,
	}
}

func (hdlr *LatencyType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "Latency", hdlr.Paths, pn, req.URL.Path)

			if db5 {
				fmt.Printf("before sleep [%.4f] seconds\n", hdlr.slowDown.Seconds())
			}
			time.Sleep(hdlr.slowDown)
			if db5 {
				fmt.Printf("after sleep\n")
			}
			hdlr.Next.ServeHTTP(rw, req)
			if db5 {
				fmt.Printf("requst has been made\n")
			}
			return

		}
	}
	hdlr.Next.ServeHTTP(www, req)
}

const db5 = false

/* vim: set noai ts=4 sw=4: */
