//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2015-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1242
//

//
// Ban a set of IP addresses base on geograpic location
//
// Copyright (C) Philip Schlump, 2016
//

package GeoIpFilter

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"www.2c-why.com/JsonX"

	"github.com/oschwald/maxminddb-golang"

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
//		pCfg, ok := ppCfg.(*GeoIPFilterType)
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
//	createEmptyType := func() interface{} { return &GeoIPFilterType{} }
//
//	cfg.RegInitItem2("GeoIpFilter", initNext, createEmptyType, InitGeoIPFilter, `{
//		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
//		"Action":        { "type":[ "string" ], "list":[ "reject", "allow" ], "default":"allow" },
//		"CountryCodes":  { "type":[ "string" ], "default":"index.html", "isarray":true },
//		"DBFileName":    { "type":[ "string","filepath" ], "default":"./cfg/GeoLite2-Country.mmdb" },
//		"PageIfBlocked": { "type":[ "string","filepath" ] },
//		"LineNo":        { "type":[ "int" ], "default":"1" }
//		}`)
//}
//
//// normally identical
//func (hdlr *GeoIPFilterType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &GeoIPFilterType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("GeoIPFilter", CreateEmpty, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"Action":        { "type":[ "string" ], "list":[ "reject", "allow" ], "default":"allow" },
		"CountryCodes":  { "type":[ "string" ], "default":"index.html", "isarray":true },
		"DBFileName":    { "type":[ "string","filepath" ], "default":"./cfg/GeoLite2-Country.mmdb" },
		"PageIfBlocked": { "type":[ "string","filepath" ] },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *GeoIPFilterType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *GeoIPFilterType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*GeoIPFilterType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type GeoIPFilterType struct {
	Next          http.Handler
	Paths         []string
	Action        string // reject, allow
	CountryCodes  []string
	PageIfBlocked string
	DBFileName    string
	LineNo        int
	dbHandler     *maxminddb.Reader // Database's handler when it get opened
}

func NewGeoIPFilterServer(n http.Handler, p []string, dbf string, act string, codes []string, fail_page_url string) *GeoIPFilterType {
	if !lib.InArray(act, []string{"reject", "allow"}) {
		fmt.Printf("Fatal: Invalid action, should be reject, allow, it is %s\n", act)
		os.Exit(1)
	}
	h := &GeoIPFilterType{
		Next:          n,             //
		Paths:         p,             //
		DBFileName:    dbf,           // File name for the database of countries
		Action:        act,           // reject, allow
		CountryCodes:  codes,         // []string
		PageIfBlocked: fail_page_url, // string
	}
	err := InitGeoIPFilter(h, nil, -1)
	if err != nil {
		fmt.Printf("Fatal: Unable to open/initialize the Geo IP Filter, %s\n", err)
		os.Exit(1)
	}
	return h
}

func (hdlr *GeoIPFilterType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "GeoIpFilter", hdlr.Paths, pn, req.URL.Path)

			p, err := lib.GetIpFromReq(req)
			if err != nil {
				www.WriteHeader(http.StatusInternalServerError)
				return
			}
			IP := net.ParseIP(p)
			var cc OnlyCountryCode
			if err := hdlr.dbHandler.Lookup(IP, &cc); err != nil {
				www.WriteHeader(http.StatusInternalServerError)
				return
			}

			isocode := cc.Country.ISOCode
			rejectCode := func() {
				if hdlr.PageIfBlocked != "" {
					req.URL.Path = hdlr.PageIfBlocked
					hdlr.Next.ServeHTTP(www, req)
				} else {
					www.WriteHeader(http.StatusForbidden)
				}
			}

			switch hdlr.Action {
			case "allow":
				if hdlr.MatchCountryCode(isocode) {
					hdlr.Next.ServeHTTP(www, req)
				} else {
					rejectCode()
				}

			case "block":
				if hdlr.MatchCountryCode(isocode) {
					rejectCode()
				} else {
					hdlr.Next.ServeHTTP(www, req)
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

func (hdlr *GeoIPFilterType) MatchCountryCode(isocode string) bool {
	for _, c := range hdlr.CountryCodes {
		if isocode == c {
			return true
		}
	}
	return false
}

// This is drectly from the example code
type OnlyCountryCode struct {
	Country struct {
		ISOCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
}

func InitGeoIPFilter(h interface{}, cfgData map[string]interface{}, callNo int) (err error) {
	//if callNo != -1 {
	//	return
	//}
	hdlr, ok := h.(*GeoIPFilterType)
	if ok {
		hdlr.dbHandler, err = maxminddb.Open(hdlr.DBFileName)
	}
	return

}

/* vim: set noai ts=4 sw=4: */
