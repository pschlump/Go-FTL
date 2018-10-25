//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2018.
//
//
// Package ZipIt implements compression middleware
//

package ZipIt

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	JsonX "github.com/pschlump/JSONx"
	"github.com/pschlump/MiscLib"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &ZipItType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("ZipIt", CreateEmpty, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"FileName":      { "type":[ "string" ], "default":"10G.gzip" },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *ZipItType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *ZipItType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*ZipItType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type ZipItType struct {
	Next     http.Handler
	Paths    []string `gfType:"string,filepath" gfRequired:"true"`
	FileName string   `gfDefault:"10G.gzip"`
	LineNo   int      `gfDefault:"1"`
}

func NewZipItServer(n http.Handler, p []string, ml string) *ZipItType {
	return &ZipItType{Next: n, Paths: p, FileName: ml}
}

func (hdlr ZipItType) ServeHTTP(www http.ResponseWriter, req *http.Request) {

	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			hdlr.Next.ServeHTTP(www, req)
			return
		}
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "ZipIt", hdlr.Paths, pn, req.URL.Path)

			hdlr.Next.ServeHTTP(rw, req)

			{
				req.Header.Del("Accept-Encoding")
				rw.Header().Set("Content-Encoding", "gzip")                                    // Set header to inticate we are processing it
				rw.Header().Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0") // HTTP 1.1.

				in, err := os.Open(hdlr.FileName)
				if err != nil {
					www.WriteHeader(http.StatusInternalServerError)
					return
				}
				defer in.Close()

				fi, err := in.Stat()
				if err != nil {
					www.WriteHeader(http.StatusInternalServerError)
					return
				}
				len := fi.Size()
				rw.Length = len

				rw.Header().Set("Length", fmt.Sprintf("%d", len)) // Set header to inticate we are processing it
				io.Copy(www, in)

				rw.SaveDataInCache = false
				return

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
