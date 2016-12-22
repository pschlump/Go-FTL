//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2015-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1246
//

// Package dumpit impements Gzip compression middleware
//
// Copyright (C) Philip Schlump, 2016
//

package Gzip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/hash-file/lib"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {

	// normally identical
	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
		pCfg, ok := ppCfg.(*GzipType)
		if ok {
			pCfg.SetNext(next)
			rv = pCfg
		} else {
			err = mid.FtlConfigError
			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
		}
		return
	}

	// normally identical
	createEmptyType := func() interface{} { return &GzipType{} }

	cfg.RegInitItem2("Gzip", initNext, createEmptyType, nil, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"MinLength":     { "type":[ "int" ], "default":"500" },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

// normally identical
func (hdlr *GzipType) SetNext(next http.Handler) {
	hdlr.Next = next
}

var _ mid.GoFTLMiddleWare = (*GzipType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type GzipType struct {
	Next      http.Handler
	Paths     []string
	MinLength int
	LineNo    int
}

func NewGzipServer(n http.Handler, p []string, ml int) *GzipType {
	return &GzipType{Next: n, Paths: p, MinLength: ml}
}

func (hdlr GzipType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	var err error

	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			hdlr.Next.ServeHTTP(www, req)
			return
		}
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "Gzip", hdlr.Paths, pn, req.URL.Path)

			hdlr.Next.ServeHTTP(rw, req)

			if (rw.StatusCode == 200 || rw.StatusCode == 0) && rw.Length >= int64(hdlr.MinLength) {
				req.Header.Del("Accept-Encoding")
				rw.Header().Set("Content-Encoding", "gzip") // Set header to inticate we are processing it

				var b bytes.Buffer // Setup to process
				gz := gzip.NewWriter(&b)
				defer gz.Close()

				oldbody := rw.GetBody()
				rw.SaveCurentBody(string(oldbody)) // save original body!

				// move the file name from ResolvedFn  to DependentFNs -- Replace file in ResolvedFn wioth --gzip--
				if !lib.InArray(rw.ResolvedFn, rw.DependentFNs) {
					rw.DependentFNs = append(rw.DependentFNs, rw.ResolvedFn)
				}
				rw.ResolvedFn = "--gzip--"

				var newdata []byte
				var NewETag string

				if _, err := gz.Write(oldbody); err != nil { // Get body and apply transform
					goto booboo
				}
				if err := gz.Flush(); err != nil {
					goto booboo
				}

				// b has data in it now -- this is the point to tell the cache to save the gzip version!
				newdata = b.Bytes()

				// Update ETag -- Need file ModTime and size - then re-calculate hash
				NewETag, err = hashlib.HashData(newdata)
				if err != nil {
					goto booboo
				}

				www.Header().Set("ETag", NewETag)
				if www.Header().Get("Cache-Control") == "" { // if have a cache that indicates no-caching - then what
					www.Header().Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0") // HTTP 1.1.
				}

				rw.ReplaceBody(newdata)
				rw.SaveDataInCache = true // Mark the data for saving if this file gets cached.

			booboo:
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

/* vim: set noai ts=4 sw=4: */
