//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2018
//
//

// Package dumpit impements Geth compression middleware
//
// Copyright (C) Philip Schlump, 2016
//

package Geth

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	JsonX "github.com/pschlump/JSONx"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/hash-file/lib"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &GethType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	// "Contracts": { "type":[ "struct" ] },
	//		Name: 		"name of contract"
	//		Address: 	"Address where loaded"
	//		Version: 	"Version of contract (semantic v1.0.2)"
	//		ABIFile:	"Path to ABI File" - or Template to file name ./abi/{{.Name}}_sol_{{.Name}}.abi
	//
	//		Additional Security Stuff...
	//			Discoverable: "yes"/"no"
	//			KeyReq: 	"yes"/"no"

	mid.RegInitItem3("Geth", CreateEmpty, `{
		"Paths":         { "type":[ "string", "filepath"], "isarray":true, "required":true },
		"ETH_Account":   { "type":[ "string" ], "required":true },
		"ETH_Password":  { "type":[ "string" ], "required":true },
		"ETH_WS_URL":	 { "type":[ "string" ], "required":true, "default":"ws://192.168.0.139:8546" },
		"Contracts":     { "type":[ "struct" ] },
		"Final":         { "type":[ "string" ], "default":"no" },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
	// "ETH_Config":	 { "type":["string"], "default":"./cfg/cfg.jsonx" },
}

func (hdlr *GethType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *GethType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	// initEndPoints(hdlr.theMux, hdlr) // Line:434 /Users/corwin/go/src/github.com/pschlump/Go-FTL/server/midlib/TabServer2/ts2_ftl.go
	//func initEndPoints(theMux *goftlmux.MuxRouter, hdlr *TabServer2Type) {
	//		theMux.HandleFunc(api_table+"{name}/desc", closure_respHandlerTableDesc(hdlr)).Methods("GET")                          // Describe
	return
}

var _ mid.GoFTLMiddleWare = (*GethType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type GethType struct {
	Next      http.Handler
	Paths     []string `gfType:"string,filepath" gfRequired:"true"`
	MinLength int      `gfDefault:"500" gfMin:"100"`
	// xyzzy
	LineNo int `gfDefault:"1"`
}

func NewGethServer(n http.Handler, p []string, ml int) *GethType {
	// xyzzy
	return &GethType{Next: n, Paths: p, MinLength: ml}
}

func (hdlr GethType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	var err error

	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			hdlr.Next.ServeHTTP(www, req)
			return
		}
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "Geth", hdlr.Paths, pn, req.URL.Path)

			// xyzzy

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