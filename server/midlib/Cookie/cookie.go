//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1230
//

//
// Package cookie allows setting and deleting of cookies using server cookie setting header.
//

package Cookie

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"www.2c-why.com/JsonX"

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
//		pCfg, ok := ppCfg.(*CookieHandlerType)
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
//	createEmptyType := func() interface{} { return &CookieHandlerType{} }
//
//	cfg.RegInitItem2("Cookie", initNext, createEmptyType, nil, `{
//		"Paths":         { "type":[ "string", "filepath" ], "isarray":true, "required":true },
//		"Name":          { "type":[ "string" ], "required":true },
//		"Value":         { "type":[ "string" ], "required":true },
//        "CookiePath":    { "type":[ "string" ] },
//        "Domain":        { "type":[ "string" ] },
//        "Expires":       { "type":[ "string" ] },
//        "MaxAge":        { "type":[ "int" ] },
//        "Secure":        { "type":[ "bool" ] },
//        "HttpOnly":      { "type":[ "bool" ] },
//		"LineNo":        { "type":[ "int" ], "default":"1" }
//		}`)
//}
//
//// normally identical
//func (hdlr *CookieHandlerType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &CookieHandlerType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("CookieHandlerType", CreateEmpty, `{
		"Paths":         { "type":[ "string", "filepath" ], "isarray":true, "required":true },
		"Name":          { "type":[ "string" ], "required":true },
		"Value":         { "type":[ "string" ], "required":true },
        "CookiePath":    { "type":[ "string" ] },
        "Domain":        { "type":[ "string" ] },
        "Expires":       { "type":[ "string" ] },
        "MaxAge":        { "type":[ "int" ] },
        "Secure":        { "type":[ "bool" ] },
        "HttpOnly":      { "type":[ "bool" ] },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *CookieHandlerType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *CookieHandlerType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*CookieHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type CookieHandlerType struct {
	Next       http.Handler
	Paths      []string
	Name       string // if Name starts with "-" then delete existing header before creating new one.
	Value      string // if Value is "" then do not set header.
	CookiePath string //
	Domain     string //
	Expires    string // (time)		// xyzzy - need a time type
	MaxAge     int    //
	Secure     bool   //
	HttpOnly   bool   //
	LineNo     int
	// theCookie  http.Cookie
}

func NewCookieServer(n http.Handler, p []string, h, v string) *CookieHandlerType {
	return &CookieHandlerType{Next: n, Paths: p, Name: h, Value: v}
}

func (hdlr CookieHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "Cookie", hdlr.Paths, pn, req.URL.Path)

			hdlr.Next.ServeHTTP(rw, req)
			h := hdlr.Name
			if hdlr.Name[0] == '-' { // if starts with - then delete the cookie
				h = hdlr.Name[1:]
				// rw.Header().Del(hdlr.Name)		// Set-Cookie - with MaxAge -1
				// rw.Header.Add("Set-Cookie", hdlr.DeleteCookie())
				http.SetCookie(www, hdlr.DeleteCookie(h))
			} else {
				// rw.Header.Add("Set-Cookie", hdlr.StringCookie())
				http.SetCookie(www, hdlr.GenCookie())
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

func (hdlr CookieHandlerType) DeleteCookie(name string) (theCookie *http.Cookie) {
	theCookie = hdlr.GenCookie()
	theCookie.MaxAge = -1
	theCookie.Value = ""
	return
}

func (hdlr CookieHandlerType) GenCookie() (theCookie *http.Cookie) {
	theCookie = &http.Cookie{}
	theCookie.Name = hdlr.Name
	theCookie.Value = hdlr.Value
	theCookie.Path = hdlr.CookiePath
	theCookie.Domain = hdlr.Domain
	exptime, err := time.Parse(time.RFC1123, hdlr.Expires)
	if err != nil {
		exptime, err = time.Parse("Mon, 02-Jan-2006 15:04:05 MST", hdlr.Expires)
		if err != nil {
			theCookie.Expires = time.Time{}
			goto skip
		}
	}
	theCookie.Expires = exptime.UTC()
skip:
	theCookie.MaxAge = hdlr.MaxAge
	theCookie.Secure = hdlr.Secure
	theCookie.HttpOnly = hdlr.HttpOnly
	return
}

/* vim: set noai ts=4 sw=4: */
