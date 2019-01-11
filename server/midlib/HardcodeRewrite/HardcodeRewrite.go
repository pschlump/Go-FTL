//
// Go-FTL
//
// Copyright (C) Ethereum Plumbing, LLC.  Philip Schlump, 2018
// Thu Feb  8 06:40:47 MDT 2018
//

//
// Rewwrite the request usign a hard coded table.
//

package HardcodeRewrite

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

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &HardcodeRewriteHandlerType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("HardcodeRewrite", CreateEmpty, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"MatchReplace":  { "type":[ "struct" ] },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *HardcodeRewriteHandlerType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *HardcodeRewriteHandlerType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	if rr_db1 {
		fmt.Fprintf(os.Stderr, "Parsed Data Is: %s\n", lib.SVarI(hdlr))
	}
	return
}

var _ mid.GoFTLMiddleWare = (*HardcodeRewriteHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------
type MatchReplaceType struct {
	Match    string // Text Match for Path
	Replace  string // Replacement Path
	SetName  string // Add inset of name=value when replace happens TODO
	SetValue string // TODO
}

type HardcodeRewriteHandlerType struct {
	Next         http.Handler       //
	Paths        []string           //
	MatchReplace []MatchReplaceType // set of match/replaces
	LineNo       int                //
	matchre      []*regexp.Regexp   //
}

func (hdlr HardcodeRewriteHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	var err error
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "HardcodeRewrite", hdlr.Paths, pn, req.URL.Path)

			ps := rw.Ps
			_ = ps

			uu := lib.GenURLFromReq(req)
			if rr_db1 {
				fmt.Fprintf(os.Stderr, "%sJust before >%s< %s%s\n", MiscLib.ColorYellow, uu, lib.SVarI(req), MiscLib.ColorReset)
			}
			for ii, vv := range hdlr.MatchReplace {
				parsed, _ := url.Parse(uu)
				if rr_db1 {
					fmt.Fprintf(os.Stderr, "At Top Parsed >%s< %d\n", parsed, ii)
				}
				if parsed.Path == vv.Match {
					if rr_db1 {
						fmt.Fprintf(os.Stderr, "Matched AT: %s\n", godebug.LF())
					}
					parsed.Path = vv.Replace
					uu = parsed.String()
				}
			}
			req.URL, err = url.Parse(uu)
			if rr_db1 {
				fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
			}
			if err != nil {
				if rr_db1 {
					fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
				}
				logrus.Warn(fmt.Sprintf("Unable to parse the rewritten URL %s, Error: %s, URL: %s, Configuration Line No:%d\n", uu, err, req.URL.Path, hdlr.LineNo))
				fmt.Fprintf(os.Stderr, "%sUnable to parse the rewritten URL %s, Error: %s, URL: %s, Configuration Line No:%d%s\n", MiscLib.ColorRed, uu, err, req.URL.Path, hdlr.LineNo, MiscLib.ColorReset)
				www.WriteHeader(http.StatusInternalServerError)
				return
			}
			if rr_db1 {
				fmt.Fprintf(os.Stderr, "%sRewrite Successful: AT: %s%s\n", MiscLib.ColorGreen, godebug.LF(), MiscLib.ColorReset)
			}
			aa := ""
			if req.URL.RawQuery != "" {
				aa = "?"
			}
			req.RequestURI = req.URL.Path + aa + req.URL.RawQuery
			if rr_db1 {
				fmt.Fprintf(os.Stderr, "%sFinal Request: ->%s<-%s\n", MiscLib.ColorGreen, lib.SVarI(req), MiscLib.ColorReset)
			}
			req.Host = req.URL.Host
			if rr_db1 {
				fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
			}
			goftlmux.ParseQueryParamsReg(www, req, &rw.Ps) //
			goftlmux.MethodParamReg(www, req, &rw.Ps)      // 15ns
			hdlr.Next.ServeHTTP(www, req)
			return
		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Fprintf(os.Stderr, "%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		hdlr.Next.ServeHTTP(www, req)
	}
}

const rr_db1 = false

/* vim: set noai ts=4 sw=4: */
