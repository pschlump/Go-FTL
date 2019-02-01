//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2018
//

//
// Package to perform both rewrite and act as a proxy.
//

// TODO:
// 1. Change MatchRE, ReplaceRE into per-path array
// 2. Note: https://blog.charmes.net/post/reverse-proxy-go/
//

package RewriteProxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"

	JsonX "github.com/pschlump/JSONx"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &RewriteProxyHandlerType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // Xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("RewriteProxy", CreateEmpty, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"MatchRE":       { "type":[ "string" ], "isarray":true, "default":"" },
		"ReplaceRE":     { "type":[ "string" ], "isarray":true, "default":"" },
		"AddGETParam":   { "type":[ "string" ], "default":"" },
		"Dest":          { "type":[ "string","url" ], "default":"" },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *RewriteProxyHandlerType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *RewriteProxyHandlerType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	fmt.Printf("%sNewSingleHostReverseProxy AT:%s %s\n", MiscLib.ColorYellow, godebug.LF(), MiscLib.ColorReset)
	dest, err := url.Parse(hdlr.Dest)
	if err != nil {
		fmt.Printf("%sNewSingleHostReverseProxy error:%s %s\n", MiscLib.ColorYellow, err, MiscLib.ColorReset)
		return err
	} else {
		dest.Path = ""
		fmt.Printf("%sNewSingleHostReverseProxy( %s )%s\n", MiscLib.ColorYellow, godebug.SVarI(dest), MiscLib.ColorReset)
		hdlr.theProxy = httputil.NewSingleHostReverseProxy(dest)
	}
	if len(hdlr.MatchRE) > 0 {
		if len(hdlr.MatchRE) != len(hdlr.ReplaceRE) {
			return fmt.Errorf("Length of MatchRE and ReplaceRE did not match.")
		}
		hdlr.matchre = make([]*regexp.Regexp, 0, len(hdlr.MatchRE))
		for _, aMatch := range hdlr.MatchRE {
			tmp, err := regexp.Compile(aMatch)
			hdlr.matchre = append(hdlr.matchre, tmp)
			if err != nil {
				return err
			}
		}
	}
	return
}

var _ mid.GoFTLMiddleWare = (*RewriteProxyHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type RewriteProxyHandlerType struct {
	Next        http.Handler     //
	Paths       []string         //
	MatchRE     []string         // regular expression
	ReplaceRE   []string         // replacement string, with ${1} pattern replacements in it
	AddGETParam string           // Additional Parmas to add to Request
	Dest        string           //
	matchre     []*regexp.Regexp //
	theProxy    http.Handler     //
	LineNo      int              //
}

func NewRewriteProxyServer(n http.Handler, p, h, v []string, d string) *RewriteProxyHandlerType {
	x := &RewriteProxyHandlerType{Next: n, Paths: p, MatchRE: h, ReplaceRE: v, Dest: d}
	for ii, aMatch := range x.MatchRE {
		re, err := regexp.Compile(aMatch)
		if err != nil {
			fmt.Printf("Invalid regular expression %s, Error: %s\n", x.MatchRE, err)
		}
		x.matchre[ii] = re
	}
	return x
}

func (hdlr RewriteProxyHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	var err error
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "RewriteProxy", hdlr.Paths, pn, req.URL.Path)

			// ------------------------- rewrite part ----------------------------------------------------------------------
			fmt.Printf("%sNewSingleHostReverseProxy (((Rewrite Part))) AT:%s %s\n", MiscLib.ColorYellow, godebug.LF(), MiscLib.ColorReset)
			u := lib.GenURLFromReq(req)
			if rw_prox_db1 {
				fmt.Printf("RewriteProxy: Just before Original pn = %v URL = >%s<  request is = %s, %s\n", pn, u, lib.SVarI(req), godebug.LF())
			}
			newURL := u
			if len(hdlr.MatchRE) > 0 {
				// This used to do a match with 1 re pair for each match in Paths
				// if len(hdlr.MatchRE) > pn && len(hdlr.ReplaceRE) > pn && hdlr.MatchRE[pn] != "" && hdlr.ReplaceRE[pn] != "" {
				//	newURL = hdlr.matchre[pn].ReplaceAllString(u, hdlr.ReplaceRE[pn])
				for ii := 0; ii < len(hdlr.MatchRE); ii++ {
					newURL = hdlr.matchre[ii].ReplaceAllString(u, hdlr.ReplaceRE[ii])
					if rw_prox_db1 {
						fmt.Printf("RewriteProxy: Modified URL(re) is = >%s<, %s\n", newURL, godebug.LF())
					}
				}
			}
			req.URL, err = url.Parse(newURL)
			// req.URL.Host = ""
			if rw_prox_db1 {
				fmt.Printf("RewriteProxy: Updated URL = >%s< newReq ->%s<- err:%s\n", newURL, lib.SVarI(req), err)
			}
			if err == nil {
				questionMark := ""
				if req.URL.RawQuery != "" {
					questionMark = "?"
				}
				// Rebuild URL for proxy server
				if hdlr.AddGETParam != "" {
					if rw_prox_db1 {
						fmt.Printf("RewriteProxy: AddGETParam = >%s<\n", hdlr.AddGETParam)
					}
					q2 := "&"
					if questionMark == "" {
						q2 = "?"
					}
					req.RequestURI = req.URL.Path + questionMark + req.URL.RawQuery + q2 + hdlr.AddGETParam
				} else {
					req.RequestURI = req.URL.Path + questionMark + req.URL.RawQuery
				}
				// xyzzy - remove Headers at this point	(see below also)
				// xyzzy - remove Cookies at this point
				// xyzzy - add Headers at this point
				// xyzzy - add Cookies at this point
				// xyzzy - dump incoming headers/cookies
				// xyzzy - dump outgoing headers/cookies
				if rw_prox_db1 {
					fmt.Printf("RewriteProxy: Final URL = >%s<\n", req.RequestURI)
					fmt.Printf("Final to be passed on req= ->%s<- note req.RequestURI\n", lib.SVarI(req))
				}
				req.Host = req.URL.Host
			} else {
				www.WriteHeader(http.StatusInternalServerError)
				return
			}

			if rw_prox_db1 {
				fmt.Printf("RewriteProxy: Before call to proxy\n")
			}

			fmt.Printf("%sNewSingleHostReverseProxy AT:%s %s\n", MiscLib.ColorYellow, godebug.LF(), MiscLib.ColorReset)
			if rw_prox_db2 {
				fmt.Printf("RewriteProxy: Before_4 call to proxy, www=%s\n", godebug.SVarI(www))
			}
			hdlr.theProxy.ServeHTTP(www, req) // hdlr.Next.ServeHTTP(rw, req)
			if rw_prox_db2 {
				fmt.Printf("RewriteProxy: After_4 call to proxy, www=%s\n", godebug.SVarI(www))
			}

			if rw.StatusCode == 502 {
				fmt.Printf("502 error may indeicate that the 'to' proxy server is not running.\n")
				fmt.Fprintf(os.Stderr, "502 error may indeicate that the 'to' proxy server is not running.\n")
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

/*

---------- Initial Example -------------------------------------------------------

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
	fmt.Println(re.ReplaceAllString("-ab-axxb-", "${1}W"))		// Xyzzy - this one - with examples on rewrite
}

---------- more relistic example -------------------------------------------------

package main

import (
	"fmt"
	"regexp"
)

func main() {
	re := regexp.MustCompile("/q/(.*)")
	fmt.Printf("%s\n", re.ReplaceAllString("/q/z", "/q?id=${1}")) // Xyzzy - this one - with examples on rewrite
}

*/

// xyzzy - todo - match on User Agent and send proxy or reject proxy to different destinations.

const rw_prox_db1 = true
const rw_prox_db2 = false

/* vim: set noai ts=4 sw=4: */
