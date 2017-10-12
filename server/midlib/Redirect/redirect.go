//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1269
//

//
// Package redirect -- 301 or 307 redirects
//
// From: https://news.ycombinator.com/item?id=12411747
// ohthehugemanate said, "301 redirects are the herpes of the Internet.  You make one mistake, and it's with you for the rest of eternity."
//
// Alwasy remember that some browsers STILL do not properly handle caching of 301 and once a 301 redirect is used it is permanent until
// the OS is re-installed on the machine.  In those browsers that do handle cachine it could easily be 10 years before the 301 can be update.
//

package Redirect

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"www.2c-why.com/JsonX"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/tmplp"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*RedirectHandlerType)
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
//	createEmptyType := func() interface{} { return &RedirectHandlerType{} }
//
//	postInit := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
//		fmt.Printf("In postInitValidation, h=%v\n", h)
//		hh, ok := h.(*RedirectHandlerType)
//		if !ok {
//			fmt.Printf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo)
//			return mid.ErrInternalError
//		} else {
//			for ii, vv := range hh.To {
//				if vv.Code == "" {
//					vv.Code = "307"
//				} else if code, ok := decodeCode[vv.Code]; ok {
//					vv.Code = fmt.Sprintf("%d", code)
//				} else {
//					fmt.Printf("Error: Invalid response code %s at position %d - should be 301 or 307 - Redirect, Line No:%d\n", vv, ii, hh.LineNo)
//					return mid.ErrInvalidConfiguration
//				}
//				hh.To[ii] = vv
//			}
//		}
//		if hh.TemplateFileName != "" {
//			hh.TemplateFileName = lib.FilepathAbs(filepath.Clean("./" + hh.TemplateFileName))
//			if lib.Exists(hh.TemplateFileName) {
//				b, err := ioutil.ReadFile(hh.TemplateFileName)
//				if err != nil {
//					fmt.Printf("Error: Specified redirect template file %s Error: %s, Line No:%d\n", hh.TemplateFileName, err, hh.LineNo)
//					return mid.ErrInvalidConfiguration
//				}
//				hh.templateData = string(b)
//			} else {
//				fmt.Printf("Error: Specified redirect template file %s missing, Line No:%d\n", hh.TemplateFileName, hh.LineNo)
//				return mid.ErrInvalidConfiguration
//			}
//		} else {
//			hh.templateData = // use default template
//				`<html>
//<head>
//<title>Moved</title>
//</head>
//<body>
//<h1>Moved</h1>
//<p>This page has moved to <a href="{{.RedirectTo}}">{{.RedirectTo}}</a>.</p>
//</body>
//</html>
//`
//		}
//		return nil
//	}
//
//	cfg.RegInitItem2("Redirect", initNext, createEmptyType, postInit, `{
//		}`)
//}
//
//// normally identical
//func (hdlr *RedirectHandlerType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &RedirectHandlerType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("Redirect", CreateEmpty, `{
		"Paths":        	{ "type":[ "string","filepath"], "isarray":true, "required":true },
		"To":               { "type":[ "struct" ] },
        "TemplateFileName": { "type":[ "string" ] },
		"LineNo":        	{ "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *RedirectHandlerType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *RedirectHandlerType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	for ii, vv := range hdlr.To {
		if vv.Code == "" {
			vv.Code = "307"
		} else if code, ok := decodeCode[vv.Code]; ok {
			vv.Code = fmt.Sprintf("%d", code)
		} else {
			fmt.Printf("Error: Invalid response code %s at position %d - should be 301 or 307 - Redirect, Line No:%d\n", vv, ii, hdlr.LineNo)
			return mid.ErrInvalidConfiguration
		}
		hdlr.To[ii] = vv
	}
	if hdlr.TemplateFileName != "" {
		hdlr.TemplateFileName = lib.FilepathAbs(filepath.Clean("./" + hdlr.TemplateFileName))
		if lib.Exists(hdlr.TemplateFileName) {
			b, err := ioutil.ReadFile(hdlr.TemplateFileName)
			if err != nil {
				fmt.Printf("Error: Specified redirect template file %s Error: %s, Line No:%d\n", hdlr.TemplateFileName, err, hdlr.LineNo)
				return mid.ErrInvalidConfiguration
			}
			hdlr.templateData = string(b)
		} else {
			fmt.Printf("Error: Specified redirect template file %s missing, Line No:%d\n", hdlr.TemplateFileName, hdlr.LineNo)
			return mid.ErrInvalidConfiguration
		}
	} else {
		hdlr.templateData = // use default template
			`<html>
<head>
<title>Moved</title>
</head>
<body>
<h1>Moved</h1>
<p>This page has moved to <a href="{{.RedirectTo}}">{{.RedirectTo}}</a>.</p>
</body>
</html>
`
	}
	return
}

var _ mid.GoFTLMiddleWare = (*RedirectHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

var decodeCode map[string]int

func init() {
	decodeCode = make(map[string]int)
	decodeCode["301"] = 301
	decodeCode["MovedPermanently"] = 301
	decodeCode["StatusMovedPermanently"] = 301
	decodeCode["307"] = 307 // if code not specified then default to 307
	decodeCode["MovedTemporary"] = 307
	decodeCode["StatusMovedTemporary"] = 307
	decodeCode[""] = 307 // if code not specified then default to 307
}

type ToType struct {
	To   string // Location to redirect to
	Code string // Code to use, 307 is default - 301 or "MovedPermanently"  - 307 or "MovedTemporary"
}

type RedirectHandlerType struct {
	Next             http.Handler //
	Paths            []string     //
	To               []ToType     // replacement string, with ${1} pattern replacements in it
	TemplateFileName string       // If "" then use default, else read and use template.  Temlate file is read at startup/reconfigure only.
	LineNo           int
	templateData     string // Data from file
}

func NewRedirectServer(n http.Handler, p []string, t []string) *RedirectHandlerType {
	return &RedirectHandlerType{Next: n, Paths: p, To: []ToType{ToType{To: t[0]}}}
}

func (hdlr *RedirectHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	var n int     // Location where paths matched, then n-th item
	var s string  //
	var To string // Location where we will redirect to

	if n = lib.PathsMatchPos(hdlr.Paths, req.URL.Path); n >= 0 && n < len(hdlr.To) {
		rest := req.URL.Path
		if n < len(hdlr.Paths) && len(hdlr.Paths[n]) < len(req.RequestURI) {
			rest = req.RequestURI[len(hdlr.Paths[n]):]
		}
		if db1 {
			fmt.Printf("rest >%s< req.RequestURI >%s< n = %d\n", rest, req.RequestURI, n)
		}
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "Redirect", hdlr.Paths, n, req.URL.Path)

			data := make(map[string]string)
			To = hdlr.To[n].To
			data["RedirectTo"] = To
			code := hdlr.To[n].Code
			data["RedirectCode"] = code
			data["THE_REST"] = rest
			if db1 {
				fmt.Printf("hdlr.To[%d]= >%s<\n", n, hdlr.To[n])
			}
			To = tmplp.TemplateProcess(hdlr.To[n].To, rw, req, data)
			data["RedirectTo"] = To
			if db1 {
				fmt.Printf("After template processing To >%s<\n", To)
			}
			if code == "307" { // http.StatusTemporaryRedirect
				if db1 {
					fmt.Printf("307 Redirect, %s = %s\n", req.URL.Path, lib.LF())
				}
				www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
				www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
				www.Header().Set("Expires", "0")                                         // Proxies.
				www.Header().Set("Content-Type", "text/html")                            //
				www.Header().Set("Location", To)
				www.WriteHeader(http.StatusTemporaryRedirect)
			} else {
				if db1 {
					fmt.Printf("301 Redirect, %s = %s\n", req.URL.Path, lib.LF())
				}
				www.Header().Set("Location", To)
				www.WriteHeader(http.StatusMovedPermanently)
			}
			s = tmplp.TemplateProcess(hdlr.templateData, rw, req, data)
			www.Write([]byte(s))
			return

		}
	}
	if db1 {
		fmt.Printf("PassThrough, %s = %s\n", req.URL.Path, lib.LF())
	}
	hdlr.Next.ServeHTTP(www, req)
	return

}

/*
HTTP/1.1 301 Moved Permanently
Location: http://www.example.org/
Content-Type: text/html
Content-Length: 174

<html>
<head>
<title>Moved</title>
</head>
<body>
<h1>Moved</h1>
<p>This page has moved to <a href="http://www.example.org/">http://www.example.org/</a>.</p>
</body>
</html>
*/

const db1 = false

/* vim: set noai ts=4 sw=4: */
