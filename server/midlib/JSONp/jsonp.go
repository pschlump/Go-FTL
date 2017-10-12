//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1254
//

// Package jsonp impements JSONp middleware
//

package JSONp

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

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
//		pCfg, ok := ppCfg.(*JSONPHandlerType)
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
//	postInit := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
//		// fmt.Printf("In postInitValidation, h=%v\n", h)
//		hh, ok := h.(*JSONPHandlerType)
//		if !ok {
//			fmt.Fprintf(os.Stderr, "%sError: Wrong data type passed, Line No:%d\n%s", MiscLib.ColorRed, hh.LineNo, MiscLib.ColorReset)
//			return mid.ErrInternalError
//		} else {
//			var err error
//			if db1 {
//				fmt.Printf("RegExp >%s<\n", hh.CallbackMustMatch)
//			}
//			hh.callbackMustMatchRe, err = regexp.Compile(hh.CallbackMustMatch)
//			if err != nil {
//				fmt.Fprintf(os.Stderr, "%sError: Unable to read compile regular expression >%s<, LineNo:%v error:%s\n%s", MiscLib.ColorRed, hh.CallbackMustMatch, hh.LineNo, err, MiscLib.ColorReset)
//				return mid.ErrInvalidConfiguration
//			}
//		}
//		return nil
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} { return &JSONPHandlerType{} }
//
//	cfg.RegInitItem2("JSONp", initNext, createEmptyType, postInit, `{
//		}`)
//}
//
//// normally identical
//func (hdlr *JSONPHandlerType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &JSONPHandlerType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("JSONp", CreateEmpty, `{
		"Paths":              { "type":["string","filepath"], "isarray":true, "required":true },
		"CallbackMustMatch":  { "type":["string"], "default":"^[a-zA-Z\\$_][a-zA-Z0-9\\$_]*$" },
		"CallbackName":       { "type":["string"], "default":"callback" },
		"LineNo":             { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *JSONPHandlerType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *JSONPHandlerType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	if db1 {
		fmt.Printf("RegExp >%s<\n", hdlr.CallbackMustMatch)
	}
	hdlr.callbackMustMatchRe, err = regexp.Compile(hdlr.CallbackMustMatch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError: Unable to read compile regular expression >%s<, LineNo:%v error:%s\n%s", MiscLib.ColorRed, hdlr.CallbackMustMatch, hdlr.LineNo, err, MiscLib.ColorReset)
		return mid.ErrInvalidConfiguration
	}
	return
}

var _ mid.GoFTLMiddleWare = (*JSONPHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type JSONPHandlerType struct {
	Next                http.Handler   //
	Paths               []string       // Paths that this will work for
	CallbackMustMatch   string         // Validation regular expression for the callback parameter - default is a JavaScript ID
	CallbackName        string         // Name of the parameter that is used to get the callback name, default "callback"
	LineNo              int            //
	callbackMustMatchRe *regexp.Regexp // precompiled regular expression from CallbackMustMatch
}

func NewJSONPServer(n http.Handler, p []string, reMatch string) *JSONPHandlerType {
	return &JSONPHandlerType{
		Next:                n,
		Paths:               p,
		CallbackMustMatch:   reMatch,
		callbackMustMatchRe: regexp.MustCompile(reMatch),
		CallbackName:        "callback",
	}
}

func (hdlr *JSONPHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "JSONp", hdlr.Paths, pn, req.URL.Path)

			req_RequestURI := req.RequestURI
			hdlr.Next.ServeHTTP(rw, req)
			h := www.Header()
			ct := h.Get("Content-Type")
			// fmt.Printf("in JSONPHandler: rw.StatusCode = %d, ct = >>>%s<<<, req.RequestURI = %s, %s\n", rw.StatusCode, ct, req_RequestURI, godebug.LF())
			// if rw.StatusCode == http.StatusOK && (strings.HasPrefix(ct, "application/json") || strings.HasPrefix(ct, "application/javascript")) {
			if rw.StatusCode == http.StatusOK && strings.HasPrefix(ct, "application/json") {
				//if rw.StatusCode == http.StatusOK {
				uu, err := url.ParseRequestURI(req_RequestURI)
				if err != nil {
					rw.Error = err
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}
				mm, err := url.ParseQuery(uu.RawQuery)
				if err != nil {
					rw.Error = err
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}
				callback := mm.Get(hdlr.CallbackName) // if there is a "callback" argument then use that to format the JSONp response.
				if callback != "" {
					if hdlr.callbackMustMatchRe.MatchString(callback) {
						h.Set("Content-Type", "application/javascript")

						// Case 1 - outside
						// rw.Prefix = rw.Prefix + callback + "("
						// rw.Postfix = ");" + rw.Postfix

						// Case 2 - inside
						// rw.Prefix = callback + "(" + rw.Prefix
						// rw.Postfix = rw.Postfix + ");"

						// Case 3 - overwrite
						rw.Prefix = callback + "("
						rw.Postfix = ");"

					} else {
						rw.WriteHeader(http.StatusInternalServerError)
						return
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

const db1 = true

/* vim: set noai ts=4 sw=4: */
