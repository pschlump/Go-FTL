//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2015-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1299
//

//
// Copyright (C) Philip Schlump, 2016
//

package standard

import (
	"net/http"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
)

//	_ "github.com/pschlump/Go-FTL/server/fileserve"

// --------------------------------------------------------------------------------------------------------------------------

func init() {

	// normally identical
	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
		pCfg, ok := ppCfg.(*StandardFileServer)
		if ok {
			pCfg.SetNext(next)
			rv = pCfg
		} else {
			err = mid.FtlConfigError
		}
		return
	}

	// normally identical
	createEmptyType := func() interface{} { return &StandardFileServer{} }

	cfg.RegInitItem2("standard_file_server", initNext, createEmptyType, nil, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"Root":          { "type":[ "string","filepath" ], "isarray":true },
		"StripPrefix":   { "type":[ "string","filepath" ], "isarray":true },
		"IndexFileList": { "type":[ "string" ], "default":"index.html", "isarray":true },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

// normally identical
func (hdlr *StandardFileServer) SetNext(next http.Handler) {
	hdlr.Next = next
}

var _ mid.GoFTLMiddleWare = (*StandardFileServer)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type StandardFileServer struct {
	Next          http.Handler // No Next, this is the bottom of the stack.
	Paths         []string
	Root          []string
	StripPrefix   []string
	IndexFileList []string
	LineNo        int
}

func NewStandardFileServer(n http.Handler, p []string, r []string, m []string) *StandardFileServer {
	if p == nil || len(m) == 0 {
		p = []string{"/"}
	}
	if r == nil || len(r) == 0 {
		r = []string{"./www"}
	}
	if m == nil || len(m) == 0 {
		m = []string{"index.html"}
	}
	return &StandardFileServer{
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
// Also this is the standard server - only one location to check.
func (hdlr *StandardFileServer) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if lib.PathsMatch(hdlr.Paths, req.URL.Path) {
		//if rw, ok := www.(*goftlmux.MidBuffer); ok {

		// xyzzy - strip prefix at this point if have it

		//fn := req.URL.Path
		//fn, found := lib.SearchForFileStandard(hdlr.Root, fn, hdlr.IndexFileList)
		//if !found {
		//	www.WriteHeader(http.StatusNotFound)
		//	return
		//}
		//buf, err := ioutil.ReadFile(fn)
		//if err != nil {
		//	www.WriteHeader(http.StatusNotFound)
		//	return
		//}
		//rw.Write(buf)

		fs := http.FileServer(http.Dir(hdlr.Root[0]))
		// gob := fileserve.NewFSConfig(hdlr.Root[0])
		// fs := fileserve.FileServer(gob)
		fs.ServeHTTP(www, req)
		return

		//} else {
		//	// fmt.Fprintf(os.Stderr, "%s\n", ErrNonMidBufferWriter)
		//	www.WriteHeader(http.StatusNotFound)
		//	return
		//}
	} else {
		if hdlr.Next != nil {
			hdlr.Next.ServeHTTP(www, req)
		} else {
			www.WriteHeader(http.StatusNotFound)
		}
	}
}

/* vim: set noai ts=4 sw=4: */
