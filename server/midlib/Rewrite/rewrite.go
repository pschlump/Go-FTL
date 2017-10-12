//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2015-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1287
//

//
// Package header rewrite
//
// Copyright (C) Philip Schlump, 2016
//

package Rewrite

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"

	JsonX "github.com/pschlump/JSONx"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*RewriteHandlerType)
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
//	createEmptyType := func() interface{} { return &RewriteHandlerType{} }
//
//	postInitValidation := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
//		fmt.Printf("In postInitValidation, h=%v\n", h)
//		hh, ok := h.(*RewriteHandlerType)
//		if !ok {
//			fmt.Printf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo)
//			return mid.ErrInternalError
//		} else {
//			if rewrite_db1 {
//				fmt.Printf("Parsed Data Is: %s\n", lib.SVarI(hh))
//			}
//			// Validate internal "struct" data
//			if len(hh.MatchReplace) == 0 {
//				fmt.Printf("Error: MatchReplace can not be empty - must have atleast one pair\n")
//				return mid.ErrInternalError
//			}
//			// build the parsed REs from input
//			for ii, vv := range hh.MatchReplace {
//				re, err := regexp.Compile(vv.Match)
//				if err != nil {
//					fmt.Printf("Invalid regular expression %s, #%d in set of match/replace pairs Error: %s\n", vv.Match, ii, err)
//				}
//				hh.matchre = append(hh.matchre, re)
//			}
//			//if rewrite_db1 {
//			//	os.Exit(1) // Exit so can see results of change -
//			//}
//		}
//		return nil
//	}
//
//	cfg.RegInitItem2("Rewrite", initNext, createEmptyType, postInitValidation, `{
//		}`)
//}
//
//// normally identical
//func (hdlr *RewriteHandlerType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &RewriteHandlerType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("Rewrite", CreateEmpty, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"MatchReplace":  { "type":[ "struct" ] },
	    "LoopLimit":     { "type":[ "int" ], "default":"50" },
	    "RestartAtTop":  { "type":[ "bool" ], "default":"true" },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *RewriteHandlerType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *RewriteHandlerType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	if rewrite_db1 {
		fmt.Printf("Parsed Data Is: %s\n", lib.SVarI(hdlr))
	}
	// Validate internal "struct" data
	if len(hdlr.MatchReplace) == 0 {
		fmt.Printf("Error: MatchReplace can not be empty - must have atleast one pair\n")
		return mid.ErrInternalError
	}
	// build the parsed REs from input
	for ii, vv := range hdlr.MatchReplace {
		re, err := regexp.Compile(vv.Match)
		if err != nil {
			fmt.Printf("Invalid regular expression %s, #%d in set of match/replace pairs Error: %s\n", vv.Match, ii, err)
		}
		hdlr.matchre = append(hdlr.matchre, re)
	}
	//if rewrite_db1 {
	//	os.Exit(1) // Exit so can see results of change -
	//}
	return
}

var _ mid.GoFTLMiddleWare = (*RewriteHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------
type MatchReplaceType struct {
	Match    string // regular expression
	Replace  string // replacement string, with ${1} pattern replacements in it
	Template string // TODO: apply templates after replace
}

type RewriteHandlerType struct {
	Next         http.Handler       //
	Paths        []string           //
	MatchReplace []MatchReplaceType // set of match/replaces
	LoopLimit    int                //
	RestartAtTop bool               // If false, then no "loop" to top will occur - just call Next
	LineNo       int                //
	matchre      []*regexp.Regexp   //
}

func NewRewriteServer(n http.Handler, p []string, h, v string) *RewriteHandlerType {
	x := &RewriteHandlerType{Next: n, Paths: p, LoopLimit: 50, RestartAtTop: true}
	x.MatchReplace = append(x.MatchReplace, MatchReplaceType{Match: h, Replace: v})
	re, err := regexp.Compile(h)
	if err != nil {
		fmt.Printf("Invalid regular expression %s, Error: %s\n", h, err)
	}
	x.matchre = append(x.matchre, re)
	return x
}

func (hdlr RewriteHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	var err error
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "Rewrite", hdlr.Paths, pn, req.URL.Path)

			uu := lib.GenURLFromReq(req)
			if rewrite_db1 {
				fmt.Printf("Just before >%s< %s\n", uu, lib.SVarI(req))
			}
			for ii, vv := range hdlr.matchre {
				ww := vv.ReplaceAllString(uu, hdlr.MatchReplace[ii].Replace)
				if rewrite_db1 {
					fmt.Printf("Just after %d match/replace >%s<, %s\n", ii, ww, godebug.LF())
				}
				uu = ww
			}
			rw.NRewrite++
			if rw.NRewrite > hdlr.LoopLimit { // check for limit on rewrites
				if rewrite_db1 {
					fmt.Printf("AT: %s\n", godebug.LF())
				}
				logrus.Warn(fmt.Sprintf("URL Exceded Rewrite Loop Limit of %d, URL: %s, Original URL: %s, Configuration Line No:%d\n", hdlr.LoopLimit, req.URL.Path, rw.OriginalURL, hdlr.LineNo))
				www.WriteHeader(http.StatusInternalServerError)
				return
			}
			req.URL, err = url.Parse(uu)
			if rewrite_db1 {
				fmt.Printf("AT: %s\n", godebug.LF())
			}
			if err == nil {
				if rewrite_db1 {
					fmt.Printf("AT: %s\n", godebug.LF())
				}
				aa := ""
				if req.URL.RawQuery != "" {
					aa = "?"
				}
				req.RequestURI = req.URL.Path + aa + req.URL.RawQuery
				if rewrite_db1 {
					fmt.Printf("req ->%s<-\n", lib.SVarI(req))
				}
				req.Host = req.URL.Host
			} else {
				if rewrite_db1 {
					fmt.Printf("AT: %s\n", godebug.LF())
				}
				logrus.Warn(fmt.Sprintf("Unable to parse the rewritten URL %s, Error: %s, URL: %s, Configuration Line No:%d\n", uu, err, req.URL.Path, hdlr.LineNo))
				www.WriteHeader(http.StatusInternalServerError)
				return
			}
			if rewrite_db1 {
				fmt.Printf("AT: %s\n", godebug.LF())
			}
			if hdlr.RestartAtTop {
				rw.RerunRequest = true // xyzzy - need to request a restart -
				rw.StatusCode = 0
			} else {
				goftlmux.ParseQueryParamsReg(www, req, &rw.Ps) //
				goftlmux.MethodParamReg(www, req, &rw.Ps)      // 15ns
				hdlr.Next.ServeHTTP(www, req)
			}
			return
		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		hdlr.Next.ServeHTTP(www, req)
	}
}

const rewrite_db1 = false

/* vim: set noai ts=4 sw=4: */
