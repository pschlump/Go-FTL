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
		pCfg, ok := ppCfg.(*LatencyType)
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
	createEmptyType := func() interface{} { return &LatencyType{} }

	postInit := func(h interface{}, cfgData map[string]interface{}, callNo int) error {

		hh, ok := h.(*LatencyType)
		if !ok {
			// logrus.Warn(fmt.Sprintf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo))
			fmt.Printf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo)
			return mid.ErrInternalError
		} else {
			hh.slowDown = time.Duration(int64(hh.SlowDown)) * time.Millisecond
		}

		return nil
	}

	cfg.RegInitItem2("Latency", initNext, createEmptyType, postInit, `{
		"Paths":        { "type":["string","filepath"], "isarray":true, "default":"/" },
		"SlowDown":     { "type":[ "int" ], "default":"500" },
		"LineNo":       { "type":[ "int" ], "default":"1" }
		}`)
}

// normally identical
func (hdlr *LatencyType) SetNext(next http.Handler) {
	hdlr.Next = next
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
