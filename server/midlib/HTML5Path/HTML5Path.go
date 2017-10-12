//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1254
//

//
// Package HTML5Path implements a rewrite for AngularJS 2.0 and AngularJS 1.x path routes to index.html.  When a 404
// error would normally be returned it it is a GET request - then map the request to /index.html.
//
// Problem: with AngularJS 1.x or 2.0 you end up with rouging paths like ``/dashboard'' that do not map to any file.
// This allows you to map these paths back to index.html where in a singe page application they need to be mapped.
//

package HTML5Path

/*

TODO:

Server: -- HTML5Path stuff
	6. Document Fix to HTML5Path - explain use R.E.				1/2hr
	6. Test Case Fix to HTML5Path - explain use R.E.
	6. Consider "/" -> "/index.html" and /Aaa /Bbb - how to handle
		-> On input "/Aaa" on return 404
		-> What we need is /index.html/Aaa
		-> 1st slash rule - 404+1st slash try, /index.html/Aaa

http://stackoverflow.com/questions/31415052/angular-2-0-router-not-working-on-reloading-the-browser?rq=1

*/

import (
	"fmt"
	"net/http"
	"os"
	"regexp"

	JsonX "github.com/pschlump/JSONx"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/MiscLib"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*HTML5PathType)
//		if ok {
//			pCfg.SetNext(next)
//			rv = pCfg
//		} else {
//			err = mid.FtlConfigError
//			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
//		}
//		return
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} { return &HTML5PathType{} }
//
//	cfg.RegInitItem2("HTML5Path", initNext, createEmptyType, nil, `{
//		}`)
//}
//
//// normally identical
//func (hdlr *HTML5PathType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &HTML5PathType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("HTML5Path", CreateEmpty, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"LimitTo":       { "type":["string","filepath"], "isarray":true },
		"ReplaceWith":   { "type":["string","filepath"], "default":"${1}" },
		"IndexReplace":  { "type":["string"], "default":"" },
		"IndexPaths":    { "type":["string"], "isarray":true },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *HTML5PathType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *HTML5PathType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*HTML5PathType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type HTML5PathType struct {
	Next         http.Handler //
	Paths        []string     //
	LimitTo      []string     // Checked on Request AFTER return of 404 - since the req.URL.Path can have changed (think re-write)
	ReplaceWith  string       //
	IndexReplace string       // set to name of "index.html" file if you want all GET 404s to get repalce with try for this file. (Breaks 404)
	IndexPaths   []string     // set of paths that will be used with IndexReplace -- means server must know all client "paths" that might arive!
	LineNo       int          //
}

func NewHTML5PathServer(n http.Handler, p []string) *HTML5PathType {
	return &HTML5PathType{Next: n, Paths: p, ReplaceWith: "/index.html"}
}

func (hdlr *HTML5PathType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	//{
	//	pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path)
	//	fmt.Fprintf(os.Stderr, "%sReq: [%s] match to [%s] pn=%d, %s%s\n", MiscLib.ColorCyan, req.URL.Path, lib.SVar(hdlr.Paths), pn, godebug.LF(), MiscLib.ColorReset)
	//}
	if pn := lib.PathsMatchReN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "HTML5Path", hdlr.Paths, pn, req.URL.Path)

			// req_RequestURI := req.RequestURI
			hdlr.Next.ServeHTTP(rw, req)
			if rw.StatusCode == http.StatusNotFound && req.Method == "GET" {
				nl := len(hdlr.LimitTo)
				if nl > 0 {
					pn = lib.PathsMatchReN(hdlr.LimitTo, req.URL.Path)
				}
				if nl == 0 || pn >= 0 {
					if db1 {
						fmt.Printf("\n\nHTML5Path: Convert of 404 on %s to %s\n", req.RequestURI, hdlr.ReplaceWith)
						// fmt.Printf("Headers: %s\n", rw.Headers)
					}
					rw.EmptyBody()    // Discard error body
					rw.StatusCode = 0 // remove 404 header

					// req.RequestURI = hdlr.ReplaceWith // see aessrp_ext.go:3351
					// req.URL.Path = hdlr.ReplaceWith   //
					// re := regexp.MustCompile(hdlr.Paths[pn])
					var re *regexp.Regexp
					if nl > 0 {
						re = lib.LookupRe(hdlr.LimitTo[pn])
					} else {
						re = lib.LookupRe(hdlr.Paths[pn])
					}
					ns := re.ReplaceAllString(req.URL.Path, hdlr.ReplaceWith)
					// xyzzy - may need to re-parse at this point?
					req.RequestURI = ns
					req.URL.Path = ns

					rw.Header().Del("Content-Type") // Discard text/plain MIME type from 404 error
					hdlr.Next.ServeHTTP(rw, req)
					if db1 {
						fmt.Printf("After Headers: rw=%s, www=%s\n", rw.Headers, www.Header())
					}
				} else if hdlr.IndexReplace != "" && lib.PathsMatch(hdlr.IndexPaths, req.URL.Path) {
					if db1 {
						fmt.Printf("\n\nHTML5Path: Convert of 404 on %s to %s -- special HandleIndex case\n", req.RequestURI, hdlr.ReplaceWith)
					}
					rw.EmptyBody()    // Discard error body
					rw.StatusCode = 0 // remove 404 header

					ns := hdlr.IndexReplace

					req.RequestURI = ns
					req.URL.Path = ns

					rw.Header().Del("Content-Type") // Discard text/plain MIME type from 404 error
					hdlr.Next.ServeHTTP(rw, req)
					if db1 {
						fmt.Printf("After Headers: rw=%s, www=%s\n", rw.Headers, www.Header())
					}
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		hdlr.Next.ServeHTTP(www, req)
	}

}

const db1 = false

/* vim: set noai ts=4 sw=4: */
