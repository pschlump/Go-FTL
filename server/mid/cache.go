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

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/lib"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {

	// normally identical
	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
		pCfg, ok := ppCfg.(*CacheHandlerType)
		if ok {
			pCfg.SetNext(next)
			rv = pCfg
		} else {
			err = FtlConfigError
		}
		return
	}

	// normally identical
	createEmptyType := func() interface{} { return &CacheHandlerType{} }

	cfg.RegInitItem2("cache", initNext, createEmptyType, nil, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"CacheDir":      { "type":[ "string","filepath" ], "default":"" },
		"CacheSize":     { "type":[ "int" ], "default":"50000000" },
		"LimitTooBig":   { "type":[ "int" ], "default":"2000000" },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

// normally identical
func (hdlr *CacheHandlerType) SetNext(next http.Handler) {
	hdlr.Next = next
}

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

func (cacheHandler CacheHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if lib.PathsMatch(cacheHandler.Paths, req.URL.Path) {

		if LookupInLRU(req.URL.Path, www, req) {
			//		if ModDateTime is different on disk then ReRead
			// 		if on disk read, else return Data
		} else {
			cacheHandler.Next.ServeHTTP(www, req)
		}

	} else {
		cacheHandler.Next.ServeHTTP(www, req)
	}
}

func LookupInLRU(fn string, www http.ResponseWriter, req *http.Request) bool {
	return false

}

/* vim: set noai ts=4 sw=4: */
