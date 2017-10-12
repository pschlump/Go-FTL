//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1267
//

//
// Minify returned data
//

package Minify

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"www.2c-why.com/JsonX"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/hash-file/lib"

	min "github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"
)

// --------------------------------------------------------------------------------------------------------------------------
//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*MinifyType)
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
//	createEmptyType := func() interface{} { return &MinifyType{} }
//
//	// /api/tmpl/showRpt.tmpl -> fetch data inside template?
//	// /api/tmpl/showRpt.tmpl?data=bob (data in row/table data)
//	cfg.RegInitItem2("Minify", initNext, createEmptyType, nil, `{
//		}`)
//}
//
//// normally identical
//func (hdlr *MinifyType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &MinifyType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("Minify", CreateEmpty, `{
		"Paths":              { "type":[ "string", "filepath" ], "isarray":true, "default":"/" },
		"FileTypes":          { "type":[ "string" ], "isarray":true, "default":"*" },
		"LineNo":             { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *MinifyType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *MinifyType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*MinifyType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type MinifyType struct {
	Next      http.Handler //
	Paths     []string     // Paths that match this
	FileTypes []string     // File tyeps are: html, css, js, xml, json, svg, or * for all, * and some other type  will eliminate star
	LineNo    int          //
}

// Parameterized for testing? or just change the test
func NewMinifyServer(n http.Handler, p []string, ft []string) *MinifyType {
	return &MinifyType{Next: n, Paths: p, FileTypes: ft}
}

var m *min.M
var re_json *regexp.Regexp
var re_xml *regexp.Regexp

func init() {
	m = min.New()
	htmlMinifier := &html.Minifier{}
	m.AddFunc("text/css", css.Minify)
	m.Add("text/html", htmlMinifier)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	re_json = regexp.MustCompile("[/+]json$")
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	re_xml = regexp.MustCompile("[/+]xml$")
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
}

func (hdlr *MinifyType) ServeHTTP(www http.ResponseWriter, req *http.Request) {

	// fmt.Printf("In Minify req.URL.Path [%s] Paths %s\n", req.URL.Path, hdlr.Paths)

	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "Minify", hdlr.Paths, pn, req.URL.Path)

			hdlr.Next.ServeHTTP(rw, req)

			// fmt.Printf("AT: %s, Header=%s\n", godebug.LF(), lib.SVarI(www.Header()))

			mimetype := www.Header().Get("Content-Type") // www.Header().Set("Content-Type", "text/html; charset=utf-8")

			doIt := func(mimetype string) {
				// fmt.Printf("AT: %s\n", godebug.LF())
				oldbody := rw.GetBody()
				rw.EmptyBody()

				readFromString := strings.NewReader(string(oldbody)) // create string reader for 's' -> r
				err := m.Minify(mimetype, www, readFromString)       // Do the minify
				if err != nil {
					logrus.Warn(fmt.Sprintf("Minify Error: %s, original content used un-changed, Line No:%d, %s", err, hdlr.LineNo, godebug.LF()))
					rw.ReplaceBody(oldbody) // resore original body from 's'
				}
				rw.SaveCurentBody(string(oldbody)) // save original body!

				// move the file name from ResolvedFn  to DependentFNs -- Replace file in ResolvedFn wioth --gzip--
				if !lib.InArray(rw.ResolvedFn, rw.DependentFNs) {
					rw.DependentFNs = append(rw.DependentFNs, rw.ResolvedFn)
				}
				rw.ResolvedFn = "--Minify--"

				var newdata []byte
				var NewETag string

				newdata = rw.GetBody()

				NewETag, err = hashlib.HashData(newdata)
				if err != nil {
					logrus.Warn(fmt.Sprintf("Minify Error: %s, original content used un-changed, Line No:%d, %s", err, hdlr.LineNo, godebug.LF()))
					rw.ReplaceBody(oldbody) // resore original body from 's'
				}

				www.Header().Set("ETag", NewETag)
				if www.Header().Get("Cache-Control") == "" { // if have a cache that indicates no-caching - then what
					www.Header().Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0") // HTTP 1.1.
				}

				// rw.ReplaceBody(newdata)
				rw.SaveDataInCache = true // Mark the data for saving if this file gets cached.
			}

			// fmt.Printf("AT: %s, mimetype=[%s]\n", godebug.LF(), mimetype)

			for _, fc := range hdlr.FileTypes {
				switch fc {
				case "*":
					if len(hdlr.FileTypes) == 1 && (strings.HasPrefix(mimetype, "text/html") || strings.HasPrefix(mimetype, "text/css") || strings.HasPrefix(mimetype, "application/javascript") ||
						strings.HasPrefix(mimetype, "image/svg+xml") || re_json.MatchString(mimetype) || re_xml.MatchString(mimetype)) {
						doIt(mimetype)
					}
				case "html", ".html", ".htm":
					if strings.HasPrefix(mimetype, "text/html") {
						doIt("text/html")
					}
				case "css", ".css":
					if strings.HasPrefix(mimetype, "text/css") {
						doIt("text/css")
					}
				case "js", ".js":
					if strings.HasPrefix(mimetype, "application/javascript") {
						doIt("text/javascript")
					}
				case "svg", ".svg":
					if strings.HasPrefix(mimetype, "image/svg+xml") {
						doIt("image/svg+xml")
					}
				case "json", ".json":
					if re_json.MatchString(mimetype) {
						doIt("text/javascript")
					}
				case "xml", ".xml":
					if re_xml.MatchString(mimetype) {
						doIt(mimetype)
					}
				default:
					logrus.Warn(fmt.Sprintf("Minify Error: Invalid file type %s, original content used un-changed, Line No:%d", fc, hdlr.LineNo))
				}
				break
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
