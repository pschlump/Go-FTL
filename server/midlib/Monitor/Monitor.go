//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2017
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1291
//

//
// A echo-like call, /api/status usually
//

package Monitor

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	JsonX "github.com/pschlump/JSONx"
	"github.com/pschlump/godebug"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/MiscLib"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &MonitorType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("Monitor", CreateEmpty, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"Fmt":           { "type":["string"], "default":"JSON" },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *MonitorType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *MonitorType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	// If callNo == 0, then this is a 1st call -- it will count up.
	// fmt.Fprintf(os.Stderr, "%sMonitor: %d%s\n", MiscLib.ColorCyan, callNo, MiscLib.ColorReset)
	return
}

var _ mid.GoFTLMiddleWare = (*MonitorType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type MonitorType struct {
	Next  http.Handler
	Paths []string
	Fmt   string
}

func NewMonitorServer(n http.Handler, p []string, fmt string) *MonitorType {
	return &MonitorType{Next: n, Paths: p, Fmt: fmt}
}

type Monitor struct {
	Alloc        uint64
	TotalAlloc   uint64
	Sys          uint64
	Mallocs      uint64
	Frees        uint64
	LiveObjects  uint64
	PauseTotalNs uint64

	NumGC        uint32
	NumGoroutine int
}

func (hdlr *MonitorType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "Monitor", hdlr.Paths, pn, req.URL.Path)

			if "JSON" == hdlr.Fmt {
				www.Header().Set("Content-Type", "application/json")

				var m Monitor
				var rtm runtime.MemStats
				// Read full mem stats
				runtime.ReadMemStats(&rtm)

				// Number of goroutines
				m.NumGoroutine = runtime.NumGoroutine()

				// Misc memory stats
				m.Alloc = rtm.Alloc
				m.TotalAlloc = rtm.TotalAlloc
				m.Sys = rtm.Sys
				m.Mallocs = rtm.Mallocs
				m.Frees = rtm.Frees

				// Live objects = Mallocs - Frees
				m.LiveObjects = m.Mallocs - m.Frees

				// GC Stats
				m.PauseTotalNs = rtm.PauseTotalNs
				m.NumGC = rtm.NumGC

				fmt.Fprintf(www, "%s\n", godebug.SVarI(m))

				www.WriteHeader(http.StatusOK)
			}

			return
		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
		}
	}
	hdlr.Next.ServeHTTP(www, req)
}

/* vim: set noai ts=4 sw=4: */
