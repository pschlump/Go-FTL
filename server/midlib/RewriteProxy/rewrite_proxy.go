//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1289
//

//
// Package to perform both rewrite and act as a proxy.
//

package RewriteProxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"

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
		pCfg, ok := ppCfg.(*RewriteProxyHandlerType)
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
	createEmptyType := func() interface{} { return &RewriteProxyHandlerType{} }

	postInit := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
		hh, ok := h.(*RewriteProxyHandlerType)
		if !ok {
			fmt.Printf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo)
			return mid.ErrInternalError
		} else {
			dest, err := url.Parse(hh.Dest)
			if err != nil {
				return err
			} else {
				hh.theProxy = httputil.NewSingleHostReverseProxy(dest)
			}
			hh.matchre, err = regexp.Compile(hh.MatchRE)
			if err != nil {
				return err
			}
		}
		return nil
	}

	cfg.RegInitItem2("RewriteProxy", initNext, createEmptyType, postInit, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"MatchRE":       { "type":[ "string" ], "default":"" },
		"ReplaceRE":     { "type":[ "string" ], "default":"" },
		"Dest":          { "type":[ "string","url" ], "default":"" },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

// normally identical
func (hdlr *RewriteProxyHandlerType) SetNext(next http.Handler) {
	hdlr.Next = next
}

var _ mid.GoFTLMiddleWare = (*RewriteProxyHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

// xyzzy - need to template translate value/name before use!

type RewriteProxyHandlerType struct {
	Next      http.Handler   //
	Paths     []string       //
	MatchRE   string         // regular expression
	ReplaceRE string         // replacement string, with ${1} pattern replacements in it
	Dest      string         //
	matchre   *regexp.Regexp //
	theProxy  http.Handler   //
	LineNo    int            //
}

func NewRewriteProxyServer(n http.Handler, p []string, h, v, d string) *RewriteProxyHandlerType {
	x := &RewriteProxyHandlerType{Next: n, Paths: p, MatchRE: h, ReplaceRE: v, Dest: d}
	re, err := regexp.Compile(x.MatchRE)
	if err != nil {
		fmt.Printf("Invalid regular expression %s, Error: %s\n", x.MatchRE, err)
	}
	x.matchre = re
	return x
}

func (hdlr RewriteProxyHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	var err error
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "RewriteProxy", hdlr.Paths, pn, req.URL.Path)

			// ------------------------- rewrite part ----------------------------------------------------------------------
			u := lib.GenURLFromReq(req)
			if rw_prox_db1 {
				fmt.Printf("Just before >%s< %s\n", u, lib.SVarI(req))
			}
			w := hdlr.matchre.ReplaceAllString(u, hdlr.ReplaceRE)
			req.URL, err = url.Parse(w)
			// req.URL.Host = ""
			if rw_prox_db1 {
				fmt.Printf("Just after err >%s< >%s<\n", err, w)
			}
			if err == nil {
				a := ""
				if req.URL.RawQuery != "" {
					a = "?"
				}
				req.RequestURI = req.URL.Path + a + req.URL.RawQuery
				if rw_prox_db1 {
					fmt.Printf("req ->%s<-\n", lib.SVarI(req))
				}
				req.Host = req.URL.Host
			} else {
				www.WriteHeader(http.StatusInternalServerError)
				return
			}

			// --------------------- the proxy part ----------------------------------------------------------------------
			//if hdlr.theProxy == nil {
			//	dest, err := url.Parse(hdlr.Dest)
			//	if err != nil {
			//		rw.Error = err
			//		www.WriteHeader(http.StatusInternalServerError)
			//		return
			//	} else {
			//		hdlr.theProxy = httputil.NewSingleHostReverseProxy(dest)
			//	}
			//}

			hdlr.theProxy.ServeHTTP(www, req) // hdlr.Next.ServeHTTP(rw, req)

		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		hdlr.Next.ServeHTTP(www, req)
	}
}

/*

func (hdlr RewriteProxyHandlerType) Init() (status int) {
}

package main

import (
	"fmt"
	"regexp"
)

func main() {
	re := regexp.MustCompile("a(x*)b")
	fmt.Println(re.ReplaceAllString("-ab-axxb-", "T"))
	fmt.Println(re.ReplaceAllString("-ab-axxb-", "$1"))
	fmt.Println(re.ReplaceAllString("-ab-axxb-", "$1W"))
	fmt.Println(re.ReplaceAllString("-ab-axxb-", "${1}W"))		// xyzzy - this one - with examples on rewrite
}

*/

const rw_prox_db1 = true

/* vim: set noai ts=4 sw=4: */
