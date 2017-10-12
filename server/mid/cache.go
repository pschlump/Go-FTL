//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2015-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1195
//

// Package cache - in memory and on disk cache - disk for proxies
//
// Copyright (C) Philip Schlump, 2016
//
package mid

import (
	"net/http"
	"time"

	JsonX "github.com/pschlump/JSONx"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/lib"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*CacheHandlerType)
//		if ok {
//			pCfg.SetNext(next)
//			rv = pCfg
//		} else {
//			err = FtlConfigError
//		}
//		return
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} { return &CacheHandlerType{} }
//
//	cfg.RegInitItem2("cache", initNext, createEmptyType, nil, `{
//		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
//		"CacheDir":      { "type":[ "string","filepath" ], "default":"" },
//		"CacheSize":     { "type":[ "int" ], "default":"50000000" },
//		"LimitTooBig":   { "type":[ "int" ], "default":"2000000" },
//		"LineNo":        { "type":[ "int" ], "default":"1" }
//		}`)
//}

func init() {
	// type CreateEmptyFx3 func(name string) *GoFTLMiddleWare
	CreateEmpty := func(name string) GoFTLMiddleWare {
		x := &CacheHandlerType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	RegInitItem3("Gzip", CreateEmpty, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"CacheDir":      { "type":[ "string","filepath" ], "default":"" },
		"CacheSize":     { "type":[ "int" ], "default":"50000000" },
		"LimitTooBig":   { "type":[ "int" ], "default":"2000000" },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *CacheHandlerType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *CacheHandlerType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

// normally identical
//func (hdlr *CacheHandlerType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

var _ GoFTLMiddleWare = (*CacheHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

// Stuff for LRU Cache

type LRUItemType struct {
	Length         int64
	ModDateTime    time.Time
	SHA1Hash       string
	URLPath        string
	FileSystemPath string
	LineNo         int
	Data           []byte // Optional Data - if using Redis - then this may be empty - and read off of disk.
	SavedInFile    string // if not "", then it was written out to file system.
}

type CacheHandlerType struct {
	Next        http.Handler
	Paths       []string
	CacheDir    string
	CacheSize   int64
	LimitTooBig int64
}

func NewCacheServer(n http.Handler, p []string, d string, siz int64, lim int64) *CacheHandlerType {
	return &CacheHandlerType{Next: n, Paths: p, CacheDir: d, CacheSize: siz, LimitTooBig: lim}
}

//	                                  ServeHTTP(www http.ResponseWriter, req *http.Request)
func (hdlr *CacheHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if lib.PathsMatch(hdlr.Paths, req.URL.Path) {

		if LookupInLRU(req.URL.Path, www, req) {
			//		if ModDateTime is different on disk then ReRead
			// 		if on disk read, else return Data
		} else {
			hdlr.Next.ServeHTTP(www, req)
		}

	} else {
		hdlr.Next.ServeHTTP(www, req)
	}
}

func LookupInLRU(fn string, www http.ResponseWriter, req *http.Request) bool {
	return false

}

/* vim: set noai ts=4 sw=4: */
