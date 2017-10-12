//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2015-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1196
//

// Implements serving GO tempates
//
// Copyright (C) Philip Schlump, 2016
//

package mid

import (
	"fmt"
	"html/template"
	"net/http"
	"os"

	JsonX "github.com/pschlump/JSONx"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*GoTemplateServer)
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
//	createEmptyType := func() interface{} { return &GoTemplateServer{} }
//
//	cfg.RegInitItem2("go_template", initNext, createEmptyType, nil, `{
//		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
//		"Root":          { "type":[ "string","filepath" ], "isarray":true },
//		"IndexFileList": { "type":[ "string" ], "default":"index.html", "isarray":true },
//		"LineNo":        { "type":[ "int" ], "default":"1" }
//		}`)
//}

func init() {
	CreateEmpty := func(name string) GoFTLMiddleWare {
		x := &GoTemplateServer{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	RegInitItem3("go_template", CreateEmpty, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"Root":          { "type":[ "string","filepath" ], "isarray":true },
		"IndexFileList": { "type":[ "string" ], "default":"index.html", "isarray":true },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *GoTemplateServer) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *GoTemplateServer) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

// normally identical
//func (hdlr *GoTemplateServer) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

var _ GoFTLMiddleWare = (*GoTemplateServer)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type GoTemplateServer struct {
	Next          http.Handler // No Next, this is the bottom of the stack.
	Paths         []string
	Root          []string
	IndexFileList []string
	LineNo        int
}

func NewGoTemplateServer(n http.Handler, p []string, r []string, m []string) *GoTemplateServer {
	if p == nil || len(m) == 0 {
		p = []string{"/"}
	}
	if r == nil || len(r) == 0 {
		r = []string{"./tmpl"}
	}
	if m == nil || len(m) == 0 {
		m = []string{"index.html.tmpl"}
	}
	return &GoTemplateServer{
		Next:          n, // may be NIL!
		Paths:         p,
		Root:          r,
		IndexFileList: m,
	}
}

//func NewDirectoryBrowseServer(n http.Handler, p []string, m string, r []string) *DirectoryBrowseType {
//	return &DirectoryBrowseType{
//		Next: n, Paths: p, Name: m, Root: r}
//}

// This is the bottom of the stack - at this point we check
// 	1. If the file is in the in-memory cache
// 	2. If the file is on disk
// Also this is the simple server - only one location to check.
func (hdlr GoTemplateServer) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if lib.PathsMatch(hdlr.Paths, req.URL.Path) {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {
			_ = rw

			fn := req.URL.Path
			fn, found := lib.SearchForFileSimple(hdlr.Root, fn, hdlr.IndexFileList)
			if !found {
				www.WriteHeader(http.StatusNotFound)
				return
			}
			tmpl := fn

			// xyzzy - at this point run the template
			a, ok := lib.SearchForFile(hdlr.Root, fn)
			if !ok {
				goto next
			}
			t, err := template.New("file-template").ParseFiles(a)
			if err != nil {
				goto next
			}

			dir2 := make(map[string]interface{}) // xyzzy - need some sort of context to get data from - this makes no sence to me.

			err = t.ExecuteTemplate(www, tmpl, dir2)
			if err != nil {
				www.WriteHeader(http.StatusInternalServerError)
				return
			}
			www.WriteHeader(http.StatusOK)
			return

		} else {
			fmt.Fprintf(os.Stderr, "%s\n", ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
next:
	hdlr.Next.ServeHTTP(www, req)
}

/*
func SearchForFile(root []string, fn string, index []string) (x string, ok bool) {
	for _, rr := range root {
		fn = filepath.Clean(rr + "/" + fn)
		if lib.ExistsIsDir(fn) {
			for _, indexfile := range index {
				if lib.Exists(fn + "/" + indexfile) {
					x = fn + "/" + indexfile
					ok = true
					return
				}
			}
			ok = false
			return
		}
		if lib.Exists(fn) {
			x = fn
			ok = true
		}
	}
	return
}
*/

/* vim: set noai ts=4 sw=4: */
