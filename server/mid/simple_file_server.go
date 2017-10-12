//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1198
//

// --------------------------------------------------------------------------------------------------------------------------
// Really simple file server - never intended to be used in production.  This is a fiel server for testing components
// of the Go-FTL middlware.
// --------------------------------------------------------------------------------------------------------------------------

package mid

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	JsonX "github.com/pschlump/JSONx"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/hash-file/lib"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*SimpleFileServer)
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
//	createEmptyType := func() interface{} { return &SimpleFileServer{} }
//
//	cfg.RegInitItem2("simple_file_server", initNext, createEmptyType, nil, `{
//		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
//		"Root":          { "type":[ "string","filepath" ], "isarray":true },
//		"IndexFileList": { "type":[ "string" ], "default":"index.html", "isarray":true },
//		"LineNo":        { "type":[ "int" ], "default":"1" }
//		}`)
//}
//
//// normally identical
//func (hdlr *SimpleFileServer) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) GoFTLMiddleWare {
		x := &SimpleFileServer{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	RegInitItem3("Gzip", CreateEmpty, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"Root":          { "type":[ "string","filepath" ], "isarray":true },
		"IndexFileList": { "type":[ "string" ], "default":"index.html", "isarray":true },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *SimpleFileServer) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *SimpleFileServer) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ GoFTLMiddleWare = (*SimpleFileServer)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type SimpleFileServer struct {
	Next          http.Handler // No Next, this is the bottom of the stack.
	Paths         []string
	Root          []string
	IndexFileList []string
	LineNo        int
}

func NewSimpleFileServer(n http.Handler, p []string, r []string, m []string) *SimpleFileServer {
	if p == nil || len(m) == 0 {
		p = []string{"/"}
	}
	if r == nil || len(r) == 0 {
		r = []string{"./www"}
	}
	if m == nil || len(m) == 0 {
		m = []string{"index.html"}
	}
	return &SimpleFileServer{
		Next:          n, // may be NIL!
		Paths:         p,
		Root:          r,
		IndexFileList: m,
	}
}

// xyzzy - Create additional "New..." funtions that will mimic proxy or other sources.

func (hdlr *SimpleFileServer) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if lib.PathsMatch(hdlr.Paths, req.URL.Path) {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			// fmt.Printf("&rw = %x, %s\n", rw, godebug.LF())

			fn := req.URL.Path
			fn, found := lib.SearchForFileSimple(hdlr.Root, fn, hdlr.IndexFileList)
			// fmt.Printf("simpe_file_server: Root: %s OrigFn: %s FoundFn: %s\n", hdlr.Root, req.URL.Path, fn)
			if !found {
				// fmt.Printf("simpe_file_server: failed to find %s for path %s - case 1\n", fn, req.URL.Path)
				www.WriteHeader(http.StatusNotFound)
				return
			}
			buf, err := ioutil.ReadFile(fn)
			if err != nil {
				// fmt.Printf("simpe_file_server: failed to find %s for path %s - case 2\n", fn, req.URL.Path)
				www.WriteHeader(http.StatusNotFound)
				return
			}

			// Set Mime Type
			ct := mime.TypeByExtension(filepath.Ext(fn))
			www.Header().Set("Content-Type", ct)

			// fmt.Printf("ct=%s for=%s\n", ct, fn)

			// fmt.Printf("simpe_file_server: File successfullly read\n")
			rw.ResolvedFn = fn
			rw.DependentFNs = append(rw.DependentFNs, fn)
			NewETag, _ := hashlib.HashData(buf)
			www.Header().Set("ETag", NewETag)
			www.Header().Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0") // HTTP 1.1.
			rw.Write(buf)
			rw.StatusCode = 200
			return
		} else {
			fmt.Fprintf(os.Stderr, "%s\n", ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		if hdlr.Next != nil {
			hdlr.Next.ServeHTTP(www, req)
		} else {
			www.WriteHeader(http.StatusNotFound)
		}
	}
}

/* vim: set noai ts=4 sw=4: */
