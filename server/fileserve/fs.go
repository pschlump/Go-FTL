//
// Based on fs.go - file server from the go source code.
//
// Origial Copyright:
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Modifications are Copyright (C) Philip Schlump, 2014-2016.
// MIT Licensed.
//

package fileserve

/*
-- example config ---------------------------------------------------------------------------------------------
-- from: go-chat/main.go

	gob := fileserve.NewFSConfig("it")
	gob.IndexHTML("/index.html").DirectoryListings(false).AddPreRule([]fileserve.PreRuleType{
		fileserve.PreRuleType{
			IfMatch:       "/",
			UseRoot:       "./push-static/",
			StatusOnMatch: fileserve.PreSuccess,
			Fx:            PushPushHtml,
		},
		fileserve.PreRuleType{
			IfMatch:       "/push-static/",
			StripPrefix:   "/push-static/",
			UseRoot:       "./push-static/",
			StatusOnMatch: fileserve.PreSuccess,
		},
		fileserve.PreRuleType{
			IfMatch:       "/",
			UseRoot:       "./static/",
			StatusOnMatch: fileserve.PreSuccess,
			Fx:            UrlModTheme,
		},
		fileserve.PreRuleType{
			IfMatch: "/",
			// StripPrefix:   "/static/",
			UseRoot:       "./static/",
			StatusOnMatch: fileserve.PreSuccess,
		},
		fileserve.PreRuleType{
			IfMatch:       "",
			UseRoot:       "./static/",
			StatusOnMatch: fileserve.PreSuccess,
		},
	})

	fs := fileserve.FileServer(gob)
	http.Handle("/", fs)
*/

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/tmplp"
	"github.com/pschlump/Go-FTL/server/urlpath"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/hash-file/lib"
)

type UrlModifyFunc func(fcfg *FileServerType, www http.ResponseWriter, req *http.Request, urlIn string, g *FSConfig, rulNo int) (urlOut string, rootOut string, stat RuleStatus, err error)
type InternalFuncType func(input, output string) (err error)

type RuleStatus int

const (
	PreSuccess RuleStatus = iota // Use this result - it matched.
	PreFail                      // Fail all results, return 404
	PreNext                      // Do not use this, proceed to next, if no next then 404
	PreReplace                   // Function  returns new root and or new name for file
)

type PreRuleType struct {
	IfMatch       string        //
	StripPrefix   string        //
	UseRoot       string        //
	StatusOnMatch RuleStatus    //
	Fx            UrlModifyFunc //
	MatchData     interface{}   //
	IfFromSuffix  string        // From suffix .ts
	IfToSuffix    []string      // To .js
	CommandToRun  string        // typescript {{.f_base}}.ts
	ReRun         bool          // if false(default) then check timestamps on input->output, else just run every time
}

//
// Pre-Rules
// Post-Rules
// Data-Avail-Rules
//
type FSConfig struct {
	Name                    string
	IndexPage               []string // list of paths for IndexPage
	MayShowDirectoryListing bool

	// -----------------------------------------------------------------------------------------------------------
	PreMangle bool // Process ULR before - this is like strip prefix
	PreRule   []PreRuleType
}

func (rs RuleStatus) String() string {
	switch rs {
	case PreSuccess:
		return "PreSuccess"
	case PreFail:
		return "PreFail"
	case PreNext:
		return "PreNext"
	case PreReplace:
		return "PreReplace"
	}
	return fmt.Sprintf("*** invalid RuleStatus = %d ***", rs)
}

func NewFSConfig(s string) *FSConfig {
	return &FSConfig{
		Name:                    s,
		IndexPage:               []string{"/index.html"},
		MayShowDirectoryListing: true,
		PreMangle:               false,
		//PreRule:                 make([]PreRuleType, 0, 10),
	}
}

const db_gob_debug1 = false

//
func (g *FSConfig) IndexHTML(s string) *FSConfig {
	g.IndexPage = append(g.IndexPage, s)
	return g
}

func (g *FSConfig) DirectoryListings(b bool) *FSConfig {
	g.MayShowDirectoryListing = b
	return g
}

func (g *FSConfig) AddPreRule(r []PreRuleType) *FSConfig {
	g.PreMangle = true
	// convert from relative paths to hard paths!
	for ii, vv := range r {
		if vv.UseRoot[0] == '.' {
			t, err := filepath.Abs(vv.UseRoot)
			fmt.Fprintf(os.Stderr, "Root from '.': %s %s %s %s\n", MiscLib.ColorCyan, t, godebug.LF(), MiscLib.ColorReset)
			if cfg.IsWindows {
				t = strings.Replace(t, `\`, "/", -1)
			}
			fmt.Fprintf(os.Stderr, "Root from '.': %s %s %s %s\n", MiscLib.ColorCyan, t, godebug.LF(), MiscLib.ColorReset)
			if err != nil {
				fmt.Printf("Error: Init: %s, %s converted to %s for path\n", err, r[ii].UseRoot, vv.UseRoot)
				// xyzzyLog - should log this.
			} else {
				if db_gob_debug1 {
					fmt.Printf("%s converted to %s for path\n", vv.UseRoot, t)
				}
				// xyzzyLog - should log this.
				vv.UseRoot = t
				r[ii] = vv
			}
		}
	}
	g.PreRule = append(g.PreRule, r...)
	return g
}

// A Dir implements FileSystem using the native file system restricted to a specific directory tree.
//
// While the FileSystem.Open method takes '/'-separated paths, a Dir's string value is a filename on the
// native file system, not a URL, so it is separated by filepath.Separator, which isn't necessarily '/'.
//
// An empty Dir is treated as ".".

type Dir string

func (d Dir) Open(name string) (File, error) {
	// if (cfg.IsWindows && strings.IndexRune(name, filepath.Separator) >= 0) || strings.Contains(name, "\x00") {
	if filepath.Separator != '/' && strings.IndexRune(name, filepath.Separator) >= 0 || strings.Contains(name, "\x00") {
		return nil, errors.New("http: invalid character in file path")
	}
	dir := string(d)
	if dir == "" {
		dir = "."
	}
	fmt.Fprintf(os.Stderr, "Before: %s %s %s\n", MiscLib.ColorCyan, name, MiscLib.ColorReset)
	x_name := filepath.Join(dir, filepath.FromSlash(path.Clean("/"+name)))
	fmt.Fprintf(os.Stderr, " after: %s %s %s\n", MiscLib.ColorCyan, x_name, MiscLib.ColorReset)
	f, err := os.Open(x_name)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// A FileSystem implements access to a collection of named files.  The elements in a file path are separated by slash ('/', U+002F)
// characters, regardless of host operating system convention.
type FileSystem interface {
	Open(name string) (File, error)
}

// A File is returned by a FileSystem's Open method and can be served by the FileServer implementation.
//
// The methods should behave the same as those on an *os.File.
type File interface {
	io.Closer
	io.Reader
	Readdir(count int) ([]os.FileInfo, error)
	Seek(offset int64, whence int) (int64, error)
	Stat() (os.FileInfo, error)
}

// ============================================================================================================================================

func (fcfg *FileServerType) dirList(www http.ResponseWriter, f File, req *http.Request, name, Root string) {

	if rw, ok := www.(*goftlmux.MidBuffer); ok {
		if dbE {
			fmt.Printf("%sAt top: name[%s] Root[%s]%s\n", MiscLib.ColorCyan, name, Root, MiscLib.ColorReset)
		}

		if len(Root) > len(name) {
			logrus.Error(fmt.Sprintf("Length of Root is shorter than lenngth of name, [%s], [%s], %s\n", Root, name, godebug.LF()))
			http.NotFound(www, req)
			return
		}
		if len(rw.IgnoreDirs) > 0 && lib.MatchPathInList(name[len(Root):], rw.IgnoreDirs) {
			http.NotFound(www, req)
			return
		}
		if rw.DirTemplateFileName != "" {

			templateFileName := rw.DirTemplateFileName
			xFn := ""

			chk3 := func(pth ...string) (foundIt bool) {
				if db8 {
					fmt.Printf("AT %s ---- doing chk2 pth >%s<\n", godebug.LF(), pth)
				}
				xFn = filepath.Join(pth...)
				// xyzzyTrx - note this would be a really good item to add to the Trx trace stuff -what files got looked for - and found.
				if db8 {
					fmt.Printf("AT %s -- xFn >%s<\n", godebug.LF(), xFn)
				}
				if ok, outFi := lib.ExistsGetFileInfo(xFn); ok { // If we have input and output, then we will check to see if need rebuild
					_ = outFi
					if db8 {
						fmt.Printf("()()() Success ()()() xFn=%s AT %s\n", xFn, godebug.LF())
					}
					templateFileName = xFn
					foundIt = true
					return
				}
				if db8 {
					fmt.Printf("DID NOT FIND template file [%s] AT %s\n", xFn, godebug.LF())
				}
				return
			}

			chk4 := func(user, theme, root string, pth ...string) (found bool) {
				if user != "" && theme != "" {
					pth_name := name[len(Root):]
					if rw.DirTemplateFileName[0] == '/' {
						if chk3(fcfg.ThemeRoot, user, theme, rw.DirTemplateFileName) {
							found = true
						} else {
							www.WriteHeader(http.StatusNotFound)
							logrus.Warn(fmt.Sprintf("Template not found in %s", xFn))
						}
					} else if chk3(fcfg.ThemeRoot, user, theme, pth_name, rw.DirTemplateFileName) {
						found = true
					} else if chk3(fcfg.ThemeRoot, user, theme, rw.DirTemplateFileName) {
						found = true
					} else if rw.DirTemplateFileName[0] == '/' && chk3(Root, rw.DirTemplateFileName) {
						found = true
					} else if chk3(name, rw.DirTemplateFileName) {
						found = true
					} else if chk3(Root, rw.DirTemplateFileName) {
						found = true
					} else {
						www.WriteHeader(http.StatusNotFound)
						logrus.Warn(fmt.Sprintf("Template not found in %s", xFn))
					}
				} else if theme != "" {
					pth_name := name[len(Root):]
					if rw.DirTemplateFileName[0] == '/' {
						if chk3(fcfg.ThemeRoot, theme, rw.DirTemplateFileName) {
							found = true
						} else {
							www.WriteHeader(http.StatusNotFound)
							logrus.Warn(fmt.Sprintf("Template not found in %s", xFn))
						}
					} else if chk3(fcfg.ThemeRoot, theme, pth_name, rw.DirTemplateFileName) {
						found = true
					} else if chk3(fcfg.ThemeRoot, theme, rw.DirTemplateFileName) {
						found = true
					} else if rw.DirTemplateFileName[0] == '/' && chk3(Root, rw.DirTemplateFileName) {
						found = true
					} else if chk3(name, rw.DirTemplateFileName) {
						found = true
					} else if chk3(Root, rw.DirTemplateFileName) {
						found = true
					} else {
						www.WriteHeader(http.StatusNotFound)
						logrus.Warn(fmt.Sprintf("Template not found in %s", xFn))
					}
				} else {
					if rw.DirTemplateFileName[0] == '/' {
						if chk3(Root, rw.DirTemplateFileName) {
							found = true
						} else {
							www.WriteHeader(http.StatusNotFound)
							logrus.Warn(fmt.Sprintf("Template not found in %s", xFn))
						}
					} else if chk3(name, rw.DirTemplateFileName) {
						found = true
					} else if chk3(Root, rw.DirTemplateFileName) {
						found = true
					} else {
						www.WriteHeader(http.StatusNotFound)
						logrus.Warn(fmt.Sprintf("Template not found in %s", xFn))
					}
				}
				return
			}

			// _ = chk4

			if db5 || dbE {
				fmt.Printf("************** Template File: %s, name [%s] Root [%s], %s\n", rw.DirTemplateFileName, name, Root, godebug.LF())
			}
			// execute template to get Content-Type // read directory, get info // execute temlate to get body

			user, theme, t_root := getThemeUserRoot(fcfg, www, req)
			chk4(user, theme, t_root, Root, name, rw.DirTemplateFileName)

			fmt.Printf("\n%s[[[dir/tempalte]]] user [%s] theme[%s], Root [%s] name [%s] t_root [%s]%s\n\n", MiscLib.ColorBlue, user, theme, Root, name, t_root, MiscLib.ColorReset)

			// OLD CODE!
			//			if rw.DirTemplateFileName[0] == '/' {
			//				if db5 {
			//					fmt.Printf("************** Using Root - by specification in rw.DirTemplateFileName\n")
			//				}
			//				templateFileName = Root + rw.DirTemplateFileName
			//				if !lib.Exists(templateFileName) {
			//					if db5 {
			//						fmt.Printf("AT: failed to read template %s\n", godebug.LF())
			//					}
			//					www.WriteHeader(http.StatusNotFound)
			//					logrus.Warn(fmt.Sprintf("Template not found in root for directory listing, %s", rw.DirTemplateFileName))
			//					return
			//				}
			//			} else if lib.Exists(name + "/" + rw.DirTemplateFileName) {
			//				if db5 {
			//					fmt.Printf("************** Using Name\n")
			//				}
			//				templateFileName = name + "/" + rw.DirTemplateFileName
			//			} else if lib.Exists(Root + "/" + rw.DirTemplateFileName) {
			//				if db5 {
			//					fmt.Printf("************** Using Root\n")
			//				}
			//				templateFileName = Root + "/" + rw.DirTemplateFileName
			//			}

			rw.ResolvedFn = name
			rw.DependentFNs = append(rw.DependentFNs, name, templateFileName)

			// templateFileName = rw.DirTemplateFileName
			// fmt.Fprintf(www, "************** Template File: %s\n", rw.DirTemplateFileName)

			// fmt.Printf("At: %s - template found, path >%s<\n", lib.LF(), a)
			compiledTemplate, err := template.New("file-template").ParseFiles(templateFileName)
			if err != nil {
				if db5 {
					fmt.Printf("AT: failed to read template %s\n", godebug.LF())
				}
				www.WriteHeader(http.StatusNotFound)
				logrus.Warn(fmt.Sprintf("Template parse error Error: %s, Line No:%d", err, rw.TemplateLineNo))
				return
			}

			// read/process template
			data := make(map[string]string)
			data = tmplp.GenDataFromReq(rw, req)
			ct := tmplp.ExecuteATemplate("content_type", data)
			www.Header().Set("Content-Type", ct)

			dir := make([]map[string]interface{}, 0, 100)

			fns, dirs := lib.GetFilenamesSorted(name) // fullDirectoryPath) // Read in directory
			for _, vv := range dirs {
				dir = append(dir, map[string]interface{}{"type": "dir", "name": vv})
			}
			for _, vv := range fns {
				dir = append(dir, map[string]interface{}{"type": "file", "name": vv}) // xyzzy - other file items [ size etc ]
			}
			dir2 := make(map[string]interface{})
			dir2["files"] = dir
			dir2["request"] = data

			err = compiledTemplate.ExecuteTemplate(www, "page", dir2)
			if err != nil {
				if db5 {
					fmt.Printf("AT: failed to execute template %s\n", godebug.LF())
				}
				www.WriteHeader(http.StatusInternalServerError)
				return
			}
			// xyzzy - should be able to set content-type with content type template!
			www.Header().Set("Content-Type", "text/html; charset=utf-8")
			www.WriteHeader(http.StatusOK)

			return
		} else {
			goto default_index
		}
	}
	return

default_index:
	www.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(www, "<pre>\n")
	for {
		dirs, err := f.Readdir(100)
		if err != nil || len(dirs) == 0 {
			break
		}
		for _, d := range dirs {
			name := d.Name()
			if d.IsDir() {
				name += "/"
			}
			// name may contain '?' or '#', which must be escaped to remain
			// part of the URL path, and not indicate the start of a query
			// string or fragment.
			url := url.URL{Path: name}
			fmt.Fprintf(www, "<a href=\"%s\">%s</a>\n", url.String(), htmlReplacer.Replace(name))
		}
	}
	fmt.Fprintf(www, "</pre>\n")
}

// from server.go
var htmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&#34;", // "&#34;" is shorter than "&quot;".
	"'", "&#39;", // "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
)

const sniffLen = 512 // from sniff.go

// ============================================================================================================================================
// if name is empty, filename is unknown. (used for mime type, before sniffing)
// if modtime.IsZero(), modtime is unknown.
// content must be seeked to the beginning of the file.
// The sizeFunc is called at most once. Its error, if any, is sent in the HTTP response.
func serveContent(www http.ResponseWriter, req *http.Request, name string, modtime time.Time, sizeFunc func() (int64, error), content io.ReadSeeker, fullName string) {

	if CheckLastModified(www, req, modtime) {
		return
	}

	rangeReq, done := CheckETag(www, req, modtime)
	if done {
		return
	}

	code := http.StatusOK

	// If Content-Type isn't set, use the file's extension to find it, but if the Content-Type is unset explicitly, do not sniff the type.
	ctypes, haveType := www.Header()["Content-Type"]
	var ctype string
	if !haveType {
		// pjs - this will need to be a little bit more complex with caching and things like ".less" and ".jsx" - may need to save "Content-Type" in meta-cache
		ctype = mime.TypeByExtension(filepath.Ext(name))
		if ctype == "" {
			// read a chunk to decide between utf-8 text and binary
			var buf [sniffLen]byte
			n, _ := io.ReadFull(content, buf[:])
			ctype = http.DetectContentType(buf[:n])
			_, err := content.Seek(0, os.SEEK_SET) // rewind to output whole file
			if err != nil {
				http.Error(www, "seeker can't seek", http.StatusInternalServerError)
				return
			}
		}
		www.Header().Set("Content-Type", ctype)
	} else if len(ctypes) > 0 {
		ctype = ctypes[0]
	}

	size, err := sizeFunc()
	if err != nil {
		http.Error(www, err.Error(), http.StatusInternalServerError)
		return
	}

	// have size, modtime -- but...
	NewETag, err := hashlib.HashFile(nil, fullName) // xyzzy - shoud have a HashFileWithInfo ( nil, name, size, modtime )
	if err != nil {
		http.Error(www, err.Error(), http.StatusInternalServerError)
		return
	}

	// fmt.Printf("NewETag=%s\n", NewETag)
	www.Header().Set("ETag", NewETag)
	www.Header().Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0") // HTTP 1.1.

	// handle Content-Range header.
	sendSize := size
	var sendContent io.Reader = content
	if size >= 0 {
		ranges, err := parseRange(rangeReq, size)
		if err != nil {
			http.Error(www, err.Error(), http.StatusRequestedRangeNotSatisfiable)
			return
		}
		if sumRangesSize(ranges) > size {
			// The total number of bytes in all the ranges is larger than the size of the file by
			// itself, so this is probably an attack, or a dumb client.  Ignore the range request.
			ranges = nil
		}
		switch {
		case len(ranges) == 1:
			// RFC 2616, Section 14.16: "When an HTTP message includes the content of a single
			// range (for example, a response to a request for a single range, or to a request for a set of ranges
			// that overlap without any holes), this content is transmitted with a Content-Range header, and a
			// Content-Length header showing the number of bytes actually transferred.
			// ...
			// A response to a request for a single range MUST NOT be sent using the multipart/byteranges media type."
			ra := ranges[0]
			if _, err := content.Seek(ra.start, os.SEEK_SET); err != nil {
				http.Error(www, err.Error(), http.StatusRequestedRangeNotSatisfiable)
				return
			}
			sendSize = ra.length
			code = http.StatusPartialContent
			www.Header().Set("Content-Range", ra.contentRange(size))
		case len(ranges) > 1:
			sendSize = rangesMIMESize(ranges, ctype, size)
			code = http.StatusPartialContent

			pr, pw := io.Pipe()
			mw := multipart.NewWriter(pw)
			www.Header().Set("Content-Type", "multipart/byteranges; boundary="+mw.Boundary())
			sendContent = pr
			defer pr.Close() // cause writing goroutine to fail and exit if CopyN doesn't finish.
			go func() {
				for _, ra := range ranges {
					part, err := mw.CreatePart(ra.mimeHeader(ctype, size))
					if err != nil {
						pw.CloseWithError(err)
						return
					}
					if _, err := content.Seek(ra.start, os.SEEK_SET); err != nil {
						pw.CloseWithError(err)
						return
					}
					if _, err := io.CopyN(part, content, ra.length); err != nil {
						pw.CloseWithError(err)
						return
					}
				}
				mw.Close()
				pw.Close()
			}()
		}

		www.Header().Set("Accept-Ranges", "bytes")
		if www.Header().Get("Content-Encoding") == "" {
			www.Header().Set("Content-Length", strconv.FormatInt(sendSize, 10))
		}
	}

	www.WriteHeader(code)

	if req.Method != "HEAD" {
		io.CopyN(www, sendContent, sendSize)
	}
}

// ============================================================================================================================================
// modtime is the modification time of the resource to be served, or IsZero().
// return value is whether this request is now complete.
func CheckLastModified(www http.ResponseWriter, req *http.Request, modtime time.Time) bool {
	if modtime.IsZero() {
		return false
	}

	// The Date-Modified header truncates sub-second precision, so use mtime < t+1s instead of mtime <= t to check for unmodified.
	if t, err := time.Parse(http.TimeFormat, req.Header.Get("If-Modified-Since")); err == nil && modtime.Before(t.Add(1*time.Second)) {
		h := www.Header()
		delete(h, "Content-Type")
		delete(h, "Content-Length")
		www.WriteHeader(http.StatusNotModified)
		return true
	}
	www.Header().Set("Last-Modified", modtime.UTC().Format(http.TimeFormat))
	return false
}

// ============================================================================================================================================
// CheckETag implements If-None-Match and If-Range checks.
//
// The ETag or modtime must have been previously set in the http.ResponseWriter's headers.  The modtime is only compared at second
// granularity and may be the zero value to mean unknown.
//
// The return value is the effective request "Range" header to use and whether this request is now considered done.
func CheckETag(www http.ResponseWriter, req *http.Request, modtime time.Time) (rangeReq string, done bool) {

	etag := www.Header().Get("Etag")
	rangeReq = req.Header.Get("Range")

	if false {
		ir := req.Header.Get("If-Range")
		inm := req.Header.Get("If-None-Match")
		fmt.Printf("CheckETag: ETag = >%s<, If-None-Match >%s< Range >%s<, If-Range >%s< \n", etag, inm, rangeReq, ir)
	}

	// Invalidate the range request if the entity doesn't match the one the client was expecting.
	// "If-Range: version" means "ignore the Range: header unless version matches the current file."
	// We only support ETag versions.  The ETag will be set to the hash of the files contents.
	// if ir := r.Header.get("If-Range"); ir != "" && ir != etag {
	if ir := req.Header.Get("If-Range"); ir != "" && ir != etag {
		// The If-Range value is typically the ETag value, but it may also be
		// the modtime date. See golang.org/issue/8367.
		timeMatches := false
		if !modtime.IsZero() {
			if t, err := http.ParseTime(ir); err == nil && t.Unix() == modtime.Unix() {
				timeMatches = true
			}
		}
		if !timeMatches {
			rangeReq = ""
		}
	}

	if inm := req.Header.Get("If-None-Match"); inm != "" {
		// Must know ETag.
		if etag == "" {
			return rangeReq, false
		}

		// TODO(bradfitz): non-GET/HEAD requests require more work: sending a different status code on matches, and
		// also can't use weak cache validators (those with a "W/ prefix).  But most users of ServeContent will be using
		// it on GET or HEAD, so only support those for now.
		if req.Method != "GET" && req.Method != "HEAD" {
			return rangeReq, false
		}

		// TODO(bradfitz): deal with comma-separated or multiple-valued list of If-None-match values.  For now just handle the common
		// case of a single item.
		if inm == etag || inm == "*" {
			h := www.Header()
			delete(h, "Content-Type")
			delete(h, "Content-Length")
			www.WriteHeader(http.StatusNotModified)
			return "", true
		}
	}
	return rangeReq, false
}

// ============================================================================================================================================
/*

1. Transform extensions

	.md->.html
	.sccs->.css
	.ts->.js

	f: [ .ext ], t [ .out, .other ]		If request .other, and have .ext then run command and build
										If request .ext and have .other, then if out of date rebuild

[
	{ "f": ".js", "t": [ ".min.js", ".min.map" ],    "cmd": "uglifyjs --input {{.f}} --output {{.t}} --generate-mapfile" }
,	{ "f": ".ts", "t": [ ".js" ],                    "cmd": "typescirpt {{.f}} " }
,	{ "f": ".css", "t": [ ".min.css" ],              "cmd": "pack-css --in {{.f}} --out {{.t}}" }
]

*/
// name is '/'-separated, not filepath.Separator.
// fs contains the directory name, name is the file name.
func (fcfg *FileServerType) ServeFile(www http.ResponseWriter, req *http.Request, name string) {

	if db1 {
		fmt.Printf("AT %s\n", godebug.LF())
	}
	// fmt.Fprintf(os.Stderr, "name [%s] at=%s\n", name, godebug.LF())
	if rw, ok := www.(*goftlmux.MidBuffer); ok {

		if db11 {
			fmt.Printf("/////////////////////////////////////////////////////////////////////\n// ServeFile called with %s, %v, %s\n", name, fcfg.LineNo, godebug.LF())
		}

		var found bool
		var fileInfo os.FileInfo
		var err error
		// fmt.Fprintf(os.Stderr, "name [%s] at=%s\n", name, godebug.LF())

		isOk := false
		t_root := ""
		if fcfg.Cfg.PreMangle {

			if db1 {
				fmt.Printf("***************** Have a Pre-Mangle Flag of True *****************, %s \n", godebug.LF())
				fmt.Printf("\tname = -->>%s<<-- (should be full URL path, cleaned) \n", name)
			}

			// xyzzy - this would be the place to do .md -> .html tranform

			for ii, aRule := range fcfg.Cfg.PreRule {

				// process each suffix rule?
				if db1 {
					fmt.Printf("AT %s\n", godebug.LF())
				}

				if aRule.IfMatch == "" || strings.HasPrefix(name, aRule.IfMatch) {
					if db1 {
						fmt.Printf("at[%d] status=%d\n", ii, aRule.StatusOnMatch)
					}
					ss := aRule.StatusOnMatch
					switch ss {

					case PreNext, PreSuccess:
						if db1 {
							fmt.Printf("AT %s\n", godebug.LF())
						}
						t_name := name
						t_root = aRule.UseRoot
						// fmt.Fprintf(os.Stderr, "t_root [%s] at=%s\n", t_root, godebug.LF())

						if aRule.StripPrefix != "" {
							if db1 {
								fmt.Printf("Before [%d] StripPrefix >%s<- with t_name >%s<-\n", ii, aRule.StripPrefix, t_name)
							}
							if p := strings.TrimPrefix(t_name, aRule.StripPrefix); len(p) < len(t_name) {
								if db1 {
									fmt.Printf("\tp= -->>%s<<--\n", p)
								}
								t_name = urlpath.Clean(string(filepath.Separator) + p)
							}
							if db1 {
								fmt.Printf("After  [%d] StripPrefix >%s<- with t_name >%s<-\n", ii, aRule.StripPrefix, t_name)
							}
						}

						if fcfg.StripPrefix != "" {
							if db11 {
								fmt.Printf("Before [%d] fcfg.StripPrefix >%s<- with t_name >%s<-\n", ii, fcfg.StripPrefix, t_name)
							}
							if p := strings.TrimPrefix(t_name, fcfg.StripPrefix); len(p) < len(t_name) {
								if db11 {
									fmt.Printf("\tp= -->>%s<<--\n", p)
								}
								t_name = urlpath.Clean(string(filepath.Separator) + p)
							}
							if db11 {
								fmt.Printf("After  [%d] fcfg.StripPrefix >%s<- with t_name >%s<-\n", ii, aRule.StripPrefix, t_name)
							}
						}

						if aRule.Fx != nil {
							if db1 || db9 {
								fmt.Printf("AT %s\n", godebug.LF())
							}
							nn, rr, s2, err := aRule.Fx(fcfg, www, req, t_name, fcfg.Cfg, ii) // xyzzyFileSet - if .md -> .html, then file set should include .md etc.
							if db9 {
								fmt.Fprintf(os.Stderr, "** nn [%s] rr [%s] at=%s\n", nn, rr, godebug.LF())
							}
							if err != nil {
								// xyzzLog - log error
								http.NotFound(www, req)
								return
							} else {
								switch s2 {
								case PreSuccess:
									t_name = nn
									t_root = rr
									ss = PreSuccess
									if db1 || db9 {
										fmt.Printf("Fx called, success returned, t_root=>%s< t_name=>%s< %d, %s\n", rr, nn, ss, godebug.LF())
										fmt.Fprintf(os.Stderr, "name [%s] at=%s\n", name, godebug.LF())
										fmt.Fprintf(os.Stderr, "t_root [%s] at=%s\n", t_root, godebug.LF())
									}
								case PreFail:
									http.NotFound(www, req)
									return
								case PreNext:
									ss = PreNext
								case PreReplace:
									if db9 {
										fmt.Fprintf(os.Stderr, "name [%s] at=%s\n", name, godebug.LF())
									}
									name = nn
									if db9 {
										fmt.Fprintf(os.Stderr, "name [%s] at=%s\n", name, godebug.LF())
									}
									t_root = rr
									if db9 {
										fmt.Fprintf(os.Stderr, "name [%s] at=%s\n", name, godebug.LF())
										fmt.Fprintf(os.Stderr, "t_root [%s] at=%s\n", t_root, godebug.LF())
									}
								}
							}
						}
						if db9 {
							fmt.Printf("t_root [%s] name [%s] at=%s\n", t_root, name, godebug.LF())
						}

						// name = path.Clean ( aRule.UseRoot + "/" + name )
						t_name = t_root + t_name
						if db1 || db9 {
							fmt.Printf("++++++++++++ After adding t_root [%s], t_name[%s] %s\n", t_root, t_name, godebug.LF())
						}

						if found, fileInfo = lib.ExistsGetFileInfo(t_name); found {
							isOk = true
							if db1 || dbE {
								fmt.Printf("%sFILE found! t_name [%s] name [%s], use it%s\n", MiscLib.ColorYellow, t_name, name, MiscLib.ColorReset)
							}
							if ss == PreSuccess {
								name = t_name // xyzzyFileSet
								goto done
							}
						}

					case PreFail:
						fmt.Printf("AT %s\n", godebug.LF())
						http.NotFound(www, req)
						return
					}
				}
			}
		done:
		}
		if db4 {
			fmt.Printf("AT t_root = [%s], %s\n", t_root, godebug.LF())
		}
		if db9 {
			fmt.Fprintf(os.Stderr, "name [%s] at=%s\n", name, godebug.LF())
		}

		if !isOk || fileInfo == nil {
			// Assume Current Directory and go with it
			nameOrig := name
			if dbE {
				fmt.Fprintf(os.Stderr, "name [%s] at=%s\n", name, godebug.LF())
			}
			name, err = filepath.Abs(filepath.Clean("./" + name))
			if err != nil {
				logrus.Error(fmt.Sprintf("Error: 404: converting file path from %s to %s", nameOrig, name))
				http.NotFound(www, req)
				return
			}
			if found, fileInfo = lib.ExistsGetFileInfo(name); !found {
				logrus.Error(fmt.Sprintf("Error: 404: file not found %s\n", name))
				http.NotFound(www, req)
				return
			}
			if dbE {
				fmt.Fprintf(os.Stderr, "name [%s] at=%s\n", name, godebug.LF())
			}
		}

		if db1 {
			fmt.Printf("AT %s\n", godebug.LF())
		}

		if fileInfo == nil {
			http.NotFound(www, req)
			return
		}

		if db1 {
			fmt.Printf("fileInfo = %+v, %s\n", fileInfo, godebug.LF())
		}

		// fmt.Fprintf(os.Stderr, "name [%s] at=%s\n", name, godebug.LF())

		if fileInfo.IsDir() {
			for _, vv := range fcfg.Cfg.IndexPage {
				index := filepath.Clean(name + string(filepath.Separator) + vv) // list of paths for IndexPage
				if found2, fileInfo2 := lib.ExistsGetFileInfo(index); found2 {
					name = index
					if db1 {
						fmt.Printf("Found index.html file !!!!!!!!!!!!!!!!!!!!!, %s\n", godebug.LF())
					}
					fileInfo = fileInfo2
					if db1 {
						fmt.Printf("fileInfo = %+v, %s\n", fileInfo, godebug.LF())
					}
					break
				}
			}
		}

		if db4 {
			fmt.Printf("--------------------------------------------------------------------------\n")
			fmt.Printf("FileName: -->>%s<<-- Root -->>%s<<-- fileInfo=%+v, %s\n", name, t_root, fileInfo, godebug.LF())
			fmt.Printf("--------------------------------------------------------------------------\n")
		}

		// fmt.Fprintf(os.Stderr, "name [%s] at=%s\n", name, godebug.LF())

		rw.ResolvedFn = name
		rw.DependentFNs = append(rw.DependentFNs, name) // xyzzyFileSet - need to be set of files to support caching

		// xyzzy - this was put int the original for a reason - and I am not getting it!!!
		// if (filepath.Separator != '/' && strings.IndexRune(name, filepath.Separator) >= 0) || strings.Index(name, "\x00") > 0 {
		if cfg.IsWindows && strings.IndexRune(name, filepath.Separator) >= 0 {
			fmt.Fprintf(os.Stderr, "Filename contains backslashes already - why? [%s] indexRule=%v\n", name, strings.IndexRune(name, filepath.Separator))
		}
		if strings.Index(name, "\x00") > 0 {
			//fmt.Fprintf(os.Stderr, "Sep [%s] indexRule=%v\n", string(filepath.Separator), strings.IndexRune(name, filepath.Separator))
			//nnn := strings.Index(name, "\x00")
			//if nnn < len(name) {
			//	fmt.Fprintf(os.Stderr, ">%s< >%s< at %d\n", name, name[0:nnn+1], nnn)
			//} else {
			//	fmt.Fprintf(os.Stderr, ">%s< end at %d\n", name, nnn)
			//}
			logrus.Error(fmt.Sprintf(`{"Error":%q, "FileName":%q, "FileName/Hex":"%x", "LineNo":%q}`, "0 chars found in file name", name, name, godebug.LF()))
			http.NotFound(www, req)
			return
		}

		fh, err := os.Open(name) // xyzzy - FileSystem.Open - lock to 1 directory (see top of file)
		if err != nil {
			logrus.Error(fmt.Sprintf(`{"FilesystemError":%q, "FileName":%q, "LineNo":%q}`, err, name, godebug.LF()))
			http.NotFound(www, req)
			return
		}
		defer fh.Close()

		if db1 {
			fmt.Printf("--------------------------------------------------------------------------- %s\n", godebug.LF())
			fmt.Printf("File should be open [%s]\n", name)
			fmt.Printf("---------------------------------------------------------------------------\n")
		}

		// Still a directory? (we didn't find an index.html file)
		if fileInfo.IsDir() {
			if db1 {
				fmt.Printf("AT %s\n", godebug.LF())
			}
			if fcfg.Cfg.MayShowDirectoryListing {
				if CheckLastModified(www, req, fileInfo.ModTime()) {
					return
				}
				fcfg.dirList(www, fh, req, name, t_root)
			} else {
				http.NotFound(www, req)
			}
			return
		}

		if db1 {
			fmt.Printf("AT Size %d - %s\n", fileInfo.Size(), godebug.LF())
		}

		// serveContent will check modification time
		sizeFunc := func() (int64, error) { return fileInfo.Size(), nil }
		// this is where logging should take place - we have the "size" of the file!
		serveContent(www, req, fileInfo.Name(), fileInfo.ModTime(), sizeFunc, fh, name)

	}
}

func (fcfg *FileServerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(fcfg.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "file_server", fcfg.Paths, pn, req.URL.Path)

			upath := req.URL.Path

			if db1 {
				fmt.Printf("File Server - ------------------------------------------------- called, %s, %s\n", upath, godebug.LF())
			}
			if len(upath) == 0 || upath[0] != '/' {
				upath = "/" + upath
				req.URL.Path = upath
			}
			upath = path.Clean(upath) // Xyzzy - faster clean
			req.URL.Path = upath

			fcfg.ServeFile(www, req, upath)
			nf := rw.StatusCode == 404
			// if nf && fcfg.IsFinal {
			if nf && false {
				return
			} else if !nf { // !nf == found or error
				return
			}

			rw.StatusCode = 0

		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	fcfg.Next.ServeHTTP(www, req)

}

// httpRange specifies the byte range to be sent to the client.
type httpRange struct {
	start, length int64
}

func (r httpRange) contentRange(size int64) string {
	return fmt.Sprintf("bytes %d-%d/%d", r.start, r.start+r.length-1, size)
}

func (r httpRange) mimeHeader(contentType string, size int64) textproto.MIMEHeader {
	return textproto.MIMEHeader{
		"Content-Range": {r.contentRange(size)},
		"Content-Type":  {contentType},
	}
}

// parseRange parses a Range header string as per RFC 2616.
func parseRange(s string, size int64) ([]httpRange, error) {
	if s == "" {
		return nil, nil // header not present
	}
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, errors.New("invalid range")
	}
	var ranges []httpRange
	for _, ra := range strings.Split(s[len(b):], ",") {
		ra = strings.TrimSpace(ra)
		if ra == "" {
			continue
		}
		i := strings.Index(ra, "-")
		if i < 0 {
			return nil, errors.New("invalid range")
		}
		start, end := strings.TrimSpace(ra[:i]), strings.TrimSpace(ra[i+1:])
		var r httpRange
		if start == "" {
			// If no start is specified, end specifies the
			// range start relative to the end of the file.
			i, err := strconv.ParseInt(end, 10, 64)
			if err != nil {
				return nil, errors.New("invalid range")
			}
			if i > size {
				i = size
			}
			r.start = size - i
			r.length = size - r.start
		} else {
			i, err := strconv.ParseInt(start, 10, 64)
			if err != nil || i > size || i < 0 {
				return nil, errors.New("invalid range")
			}
			r.start = i
			if end == "" {
				// If no end is specified, range extends to end of the file.
				r.length = size - r.start
			} else {
				i, err := strconv.ParseInt(end, 10, 64)
				if err != nil || r.start > i {
					return nil, errors.New("invalid range")
				}
				if i >= size {
					i = size - 1
				}
				r.length = i - r.start + 1
			}
		}
		ranges = append(ranges, r)
	}
	return ranges, nil
}

// ============================================================================================================================================
// countingWriter counts how many bytes have been written to it.
type countingWriter int64

func (w *countingWriter) Write(p []byte) (n int, err error) {
	*w += countingWriter(len(p))
	return len(p), nil
}

// ============================================================================================================================================
// rangesMIMESize returns the number of bytes it takes to encode the
// provided ranges as a multipart response.
func rangesMIMESize(ranges []httpRange, contentType string, contentSize int64) (encSize int64) {
	var w countingWriter
	mw := multipart.NewWriter(&w)
	for _, ra := range ranges {
		mw.CreatePart(ra.mimeHeader(contentType, contentSize))
		encSize += ra.length
	}
	mw.Close()
	encSize += int64(w)
	return
}

// ============================================================================================================================================
func sumRangesSize(ranges []httpRange) (size int64) {
	for _, ra := range ranges {
		size += ra.length
	}
	return
}

// ============================================================================================================================================
func ServeFile(w http.ResponseWriter, r *http.Request, name string) {
	bot := mid.NewServer()
	// func NewFileServer(n http.Handler, g *FSConfig, Path []string, IndexFileList []string, Root []string) *FileServerType {
	ms := NewFileServer(bot, nil, []string{"/foo"}, []string{"index.html", "index.htm"}, []string{"./www"})
	ms.ServeFile(w, r, name)
}

const dbD = false
const dbE = false
const db11 = true // strip prerix fcfg

/* vim: set noai ts=4 sw=4: */
