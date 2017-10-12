//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1283
//

package RejectHotlink

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	JsonX "github.com/pschlump/JSONx"

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
//		pCfg, ok := ppCfg.(*RejectHotlinkType)
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
//	// convert from "yes"/"no" to bool data
//	// 	AlloweEmpty    string       //
//	//	ReturnError    string       // if false, "no" - then return a file - usually empty.
//	postInit := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
//
//		var err error
//
//		hh, ok := h.(*RejectHotlinkType)
//		if !ok {
//			fmt.Printf("Error: Wrong data type passed to FileServeType - postInit\n")
//			return mid.ErrInternalError
//		} else {
//
//			hh.alloweEmpty, err = lib.ParseBool(hh.AlloweEmpty)
//			if err != nil {
//				fmt.Fprintf(os.Stderr, "%sError: Invalid boolean data, %d%s\n", MiscLib.ColorRed, hh.LineNo, MiscLib.ColorReset)
//				fmt.Printf("Error: Invalid boolean data, %d\n", hh.LineNo)
//				return mid.ErrInternalError
//			}
//			hh.returnError, err = lib.ParseBool(hh.ReturnError)
//			if err != nil {
//				fmt.Fprintf(os.Stderr, "%sError: Invalid boolean data, %d%s\n", MiscLib.ColorRed, hh.LineNo, MiscLib.ColorReset)
//				fmt.Printf("Error: Invalid boolean data, %d\n", hh.LineNo)
//				return mid.ErrInternalError
//			}
//
//			for ii, vv := range hh.FileExtensions {
//				if len(vv) > 0 && vv[0] != '.' {
//					fmt.Fprintf(os.Stderr, "%sError: Invalid file extenstion (%s) must start with '.' at, %d, LineNo:%d%s\n", MiscLib.ColorRed, vv, ii, hh.LineNo, MiscLib.ColorReset)
//					fmt.Fprintf(os.Stderr, "Error: Invalid file extenstion (%s) must start with '.' at, %d, LineNo:%d\n", vv, ii, hh.LineNo)
//					return mid.ErrInternalError
//				}
//			}
//
//		}
//		return nil
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} { return &RejectHotlinkType{} }
//
//	cfg.RegInitItem2("RejectHotlink", initNext, createEmptyType, postInit, `{
//		}`)
//}
//
//// normally identical
//func (hdlr *RejectHotlinkType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &RejectHotlinkType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("RejectHotlink", CreateEmpty, `{
		"Paths":           { "type":["string","filepath"], "isarray":true, "required":true },
		"AllowedReferer":  { "type":["string"], "isarray":true, "required":true },
		"FileExtensions":  { "type":["string"], "isarray":true, "required":true },
		"AlloweEmpty":     { "type":["string"], "default":"yes" },
		"IgnoreHosts":     { "type":["string"], "isarray":true },
		"ReturnError":     { "type":["string"], "default":"no" },
		"LineNo":          { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *RejectHotlinkType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *RejectHotlinkType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	hdlr.alloweEmpty, err = lib.ParseBool(hdlr.AlloweEmpty)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError: Invalid boolean data, %d%s\n", MiscLib.ColorRed, hdlr.LineNo, MiscLib.ColorReset)
		fmt.Printf("Error: Invalid boolean data, %d\n", hdlr.LineNo)
		return mid.ErrInternalError
	}
	hdlr.returnError, err = lib.ParseBool(hdlr.ReturnError)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError: Invalid boolean data, %d%s\n", MiscLib.ColorRed, hdlr.LineNo, MiscLib.ColorReset)
		fmt.Printf("Error: Invalid boolean data, %d\n", hdlr.LineNo)
		return mid.ErrInternalError
	}

	for ii, vv := range hdlr.FileExtensions {
		if len(vv) > 0 && vv[0] != '.' {
			fmt.Fprintf(os.Stderr, "%sError: Invalid file extenstion (%s) must start with '.' at, %d, LineNo:%d%s\n", MiscLib.ColorRed, vv, ii, hdlr.LineNo, MiscLib.ColorReset)
			fmt.Fprintf(os.Stderr, "Error: Invalid file extenstion (%s) must start with '.' at, %d, LineNo:%d\n", vv, ii, hdlr.LineNo)
			return mid.ErrInternalError
		}
	}
	return
}

var _ mid.GoFTLMiddleWare = (*RejectHotlinkType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type RejectHotlinkType struct {
	Next           http.Handler //
	Paths          []string     // Usually for this paths include all, so just set to "/", or maybee /js/, /css/, /image/
	AllowedReferer []string     // www.example.com, example.com etc.
	FileExtensions []string     // Set of file extenstions to apply this to,  ".js", ".gif", ".jpg", ".png" - not normally to .html
	AlloweEmpty    string       //
	IgnoreHosts    []string     // These hosts are completely ignored by this -- i.e. "localhost", "127.0.0.1" - useful for testing
	ReturnError    string       // if false, "no" - then return a file - usually empty.
	LineNo         int          //
	alloweEmpty    bool         //
	returnError    bool         // if false - then return a file - usually empty.
}

// func NewRejectPathServer(n http.Handler, p []string) *RejectHotlinkType {
func NewRejectHotLinkServer(n http.Handler, p []string, all []string, ext []string) *RejectHotlinkType {
	return &RejectHotlinkType{Next: n, Paths: p, AllowedReferer: all, FileExtensions: ext, alloweEmpty: true, returnError: true}
}

func (hdlr *RejectHotlinkType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "RejectHotLink", hdlr.Paths, pn, req.URL.Path)

			www.WriteHeader(http.StatusNotFound)
			if IgnoreHost(req.Host, hdlr.IgnoreHosts) {
				hdlr.Next.ServeHTTP(www, req)
				return
			}

			if HasExtension(req.URL.Path, hdlr.FileExtensions) {
				referer := req.Header.Get("Referer")
				// fmt.Printf("referer =%s, %s\n", referer, godebug.LF())
				if referer == "" && hdlr.alloweEmpty {
					// fmt.Printf("allowed true -- empty is allowed %s\n", godebug.LF())
					hdlr.Next.ServeHTTP(www, req)
					return
				}
				if AllowedReferer(referer, hdlr.AllowedReferer) {
					// fmt.Printf("allowed true %s\n", godebug.LF())
					hdlr.Next.ServeHTTP(www, req)
					return
				}

				if hdlr.returnError {
					// at this point return error or other file -
					// fmt.Printf("returing 404 !!!! %s\n", godebug.LF())
					www.WriteHeader(http.StatusNotFound)
				} else {
					// Ok - return a clear gif image - instead of requested file.
					www.Header().Set("Content-Type", "image/gif")
					output, _ := base64.StdEncoding.DecodeString(base64GifPixel)
					io.WriteString(www, string(output))
				}
				return
			}

		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
		}
	}
	hdlr.Next.ServeHTTP(www, req)

}

func IgnoreHost(Host string, IgnoreHosts []string) bool {
	for _, vv := range IgnoreHosts {
		if strings.HasPrefix(Host, vv+":") {
			return true
		}
	}
	return false
}

func AllowedReferer(referer string, AllowedReferer []string) bool {
	if lib.InArray(referer, AllowedReferer) {
		return true
	}
	return false
}

func HasExtension(path string, FileExtensions []string) bool {
	ext := filepath.Ext(path)
	// fmt.Printf("ext found=%s, %s\n", ext, godebug.LF())
	if lib.InArray(ext, FileExtensions) {
		return true
	}
	return false
}

const base64GifPixel = "R0lGODlhAQABAIAAAP///wAAACwAAAAAAQABAAACAkQBADs="

/* vim: set noai ts=4 sw=4: */
