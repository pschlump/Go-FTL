//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1192
//

package lib

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pschlump/godebug"
	"github.com/pschlump/json" // //	Modifed from: "encoding/json"
	"github.com/pschlump/uuid"
)

// IsLocalhost is true when passed a localhost IP address v4 or ipv6
func IsLocalhost(s string) bool {
	return s == "localhost" ||
		s == "::1" ||
		s == "[::1]" ||
		s == "0:0:0:0:0:0:0:1" ||
		(len(s) >= 4 && s[0:4] == "127.")
}

func SetMaxCPUs(numCPU int) {
	availCPU := runtime.NumCPU()
	if availCPU < numCPU || numCPU == -1 {
		numCPU = availCPU
	}
	runtime.GOMAXPROCS(numCPU)
}

func IsErrFatal(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s from %s\n", err, LF(2))
		os.Exit(1)
	}
}

// Return the File name and Line no as a string.
func LF(d ...int) string {
	depth := 1
	if len(d) > 0 {
		depth = d[0]
	}
	_, file, line, ok := runtime.Caller(depth)
	if ok {
		return fmt.Sprintf("File: %s LineNo:%d", file, line)
	} else {
		return fmt.Sprintf("File: Unk LineNo:Unk")
	}
}

func IsErr(err error) (e error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s from %s\n", err, LF(2))
	}
	e = err
	return
}

/*

Find local addresses of machine - bind to machine -

	ifaces, err := net.Interfaces()
	// handle err
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
					ip = v.IP
			case *net.IPAddr:
					ip = v.IP
			}
			// process IP address
		}
	}

https://github.com/codegangsta/negroni/blob/master/negroni.go
https://www.godoc.org/golang.org/x/oauth2
http://golang.org/src/net/http/server.go?s=1517:2599#L48 -- Look at Flusher and Hijacker interface
*/

func IsLocalPath(s string) bool {
	// ORIG: matches, _ := regexp.MatchString("^/", s)		// returns t/f bool
	// ORIG: Xyzzy - must not ignore error
	// ORIG: return matches
	return len(s) > 0 && s[0] == '/'
	// xyzzy - 1st char '/' apears to me - how is this local what about ./xyz
}

type ProtocalType int

const (
	HTTPProtocal  ProtocalType = 1
	HTTPSProtocal ProtocalType = 2
	WSProtocal    ProtocalType = 3
	WSSProtocal   ProtocalType = 4
	FileProtocal  ProtocalType = 5
	ProxyProtocal ProtocalType = 6
)

func (p ProtocalType) String() string {
	switch p {
	case 1:
		return "HTTPProtocal"
	case 2:
		return "HTTPSProtocal"
	case 3:
		return "WSProtocal"
	case 4:
		return "WSSProtocal"
	case 5:
		return "FileProtocal"
	case 6:
		return "ProxyProtocal"
	}
	return fmt.Sprintf("InvalidProtocal(%d)", int(p))
}

// Looks at the leading part of a URL to determine the protocal, HTTPProtocal is the default
func DetermineProtocal(s string) (rv ProtocalType) {
	rv = HTTPProtocal
	switch {
	case strings.HasPrefix(s, "https:"):
		rv = HTTPSProtocal
	case strings.HasPrefix(s, "http:"):
		rv = HTTPProtocal
	case strings.HasPrefix(s, "wss:"):
		rv = WSSProtocal
	case strings.HasPrefix(s, "ws:"):
		rv = WSProtocal
	case strings.HasPrefix(s, "/"):
		rv = FileProtocal
	case strings.HasPrefix(s, "./"):
		rv = FileProtocal
	case strings.HasPrefix(s, "~"):
		rv = FileProtocal
	}
	return
}

func IsProtocal(s string) (rv bool) {
	switch {
	case strings.HasPrefix(s, "https:"):
		rv = true
	case strings.HasPrefix(s, "http:"):
		rv = true
	case strings.HasPrefix(s, "wss:"):
		rv = true
	case strings.HasPrefix(s, "ws:"):
		rv = true
	}
	return
}

// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// Check that a file exists and is an executable program - Linux/Unix/MacOS - 'x' permissions && not a directory
func ExistsIsExecutable(name string) bool {
	fi, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	if fi.IsDir() {
		return false
	}
	mode := fi.Mode()
	if (mode & 0444) != 0 {
		return true
	}
	return false
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------
func SVar(v interface{}) string {
	s, err := json.Marshal(v)
	// s, err := json.MarshalIndent ( v, "", "\t" )
	if err != nil {
		return fmt.Sprintf("Error:%s", err)
	} else {
		return string(s)
	}
}

// -------------------------------------------------------------------------------------------------
func SVarI(v interface{}) string {
	// s, err := json.Marshal ( v )
	s, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return fmt.Sprintf("Error:%s", err)
	} else {
		return string(s)
	}
}

// -------------------------------------------------------------------------------------------------
// Convert from a request back to a URL
func GenURLFromReq(req *http.Request) (url string) {
	scheme := "http://"
	if req.TLS != nil {
		scheme = "https://"
	}
	fn := ""
	if strings.HasSuffix(req.URL.Path, "/") {
		fn = "index.html"
	}
	Q := ""
	if len(req.URL.RawQuery) > 0 {
		Q = "?"
	}
	return scheme + req.Host + req.URL.Path + fn + Q + req.URL.RawQuery
}

// -------------------------------------------------------------------------------------------------
// Convert from a request back to a URL
func GenURLFromReqProxy(req *http.Request, newHost string) (url string) {
	scheme := "http://"
	if req.TLS != nil {
		scheme = "https://"
	}
	fn := ""
	if strings.HasSuffix(req.URL.Path, "/") {
		fn = "index.html"
	}
	Q := ""
	if len(req.URL.RawQuery) > 0 {
		Q = "?"
	}
	return scheme + newHost + req.URL.Path + fn + Q + req.URL.RawQuery
}

// -------------------------------------------------------------------------------------------------
func InArray(lookFor string, inArr []string) bool {
	for _, vv := range inArr {
		if lookFor == vv {
			return true
		}
	}
	return false
}

// -------------------------------------------------------------------------------------------------
func InArrayN(lookFor string, inArr []string) int {
	for ii, vv := range inArr {
		if lookFor == vv {
			return ii
		}
	}
	return -1
}

// -------------------------------------------------------------------------------------------------
func GetIpFromReq(req *http.Request) (ip string, err error) {

	ip = req.Header.Get("X-Forwarded-For")
	if ip != "" {
		return
	} else {
		ip, _, err = net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			return
		}
	}
	return
}

// -------------------------------------------------------------------------------------------------
// func (hdlr RedirectHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
func HasHeader(wr http.ResponseWriter, name string) (val string, found bool) {
	val = ""
	found = false
	vv := wr.Header().Get(name)
	if vv == "" {
		return
	}

	found = true
	val = vv
	return

}

// -------------------------------------------------------------------------------------------------
// Get a list of filenames and directorys.
func GetFilenames(dir string) (filenames, dirs []string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, nil
	}
	for _, fstat := range files {
		if !strings.HasPrefix(string(fstat.Name()), ".") {
			if fstat.IsDir() {
				dirs = append(dirs, fstat.Name())
			} else {
				filenames = append(filenames, fstat.Name())
			}
		}
	}
	return
}

// -------------------------------------------------------------------------------------------------
func GetFilenamesSorted(dir string) (filenames, dirs []string) {
	filenames, dirs = GetFilenames(dir)
	sort.Strings(filenames)
	sort.Strings(dirs)
	return
}

// -------------------------------------------------------------------------------------------------
func SearchForFile(root []string, fn string) (string, bool) {
	var a string
	if len(root) == 0 {
		a = filepath.Clean("./" + fn)
		if Exists(a) {
			return a, true
		}
	}
	for _, pth := range root {
		a = filepath.Clean(pth + "/" + fn)
		if Exists(a) {
			return a, true
		}
	}
	return "", false
}

//		lib.SetupRequestHeaders(req, test.hdr)
func SetupRequestHeaders(req *http.Request, hdr []NameValue) {
	for _, hh := range hdr {
		if _, ok := req.Header[hh.Name]; !ok {
			req.Header[hh.Name] = make([]string, 0, 1)
		}
		req.Header[hh.Name] = append(req.Header[hh.Name], hh.Value)
	}
}

func SetupTestMimicReq(req *http.Request, h string) {
	req.RemoteAddr = "1.2.2.2:52180" // "RemoteAddr": "[::1]:59668",
	req.Host = h
	a := ""
	if req.URL.RawQuery != "" {
		a = "?"
	}
	req.RequestURI = req.URL.Path + a + req.URL.RawQuery
	// ?? set up a Body?
}

//	{"status":"full","query":"/api/status?id=xyzzy","req":{
//		"Method": "GET",
//		"URL": {
//			"Scheme": "",
//			"Opaque": "",
//			"User": null,
//			"Host": "",
//			"Path": "/api/status",
//			"RawQuery": "id=xyzzy",
//			"Fragment": ""
//		},
//		"Proto": "HTTP/1.1",
//		"ProtoMajor": 1,
//		"ProtoMinor": 1,
//		"Header": {
//			"Accept": [
//				"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"
//			],
//			"Accept-Encoding": [
//				"gzip, deflate, sdch"
//			],
//			"Accept-Language": [
//				"en-US,en;q=0.8"
//			],
//			"Connection": [
//				"keep-alive"
//			],
//			"Upgrade-Insecure-Requests": [
//				"1"
//			],
//			"User-Agent": [
//				"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.99 Safari/537.36"
//			]
//		},
//		"Body": {
//			"Closer": {
//				"Reader": null
//			}
//		},
//		"ContentLength": 0,
//		"TransferEncoding": null,
//		"Close": false,
//		"Host": "localhost:8204",
//		"Form": null,
//		"PostForm": null,
//		"MultipartForm": null,
//		"Trailer": null,
//		"RemoteAddr": "[::1]:59668",
//		"RequestURI": "/api/status?id=xyzzy",
//		"TLS": null
//	}, "header":{
//		"Content-Type": [
//			"application/json"
//		]
//	}}

type NameValue struct {
	Name  string
	Value string
}

var setup_done = false
var setup_done1 = false

func SetupTestCreateDirsFileServe() {
	if setup_done1 { // only run once per test - create logs, files for test.
		return
	}
	setup_done1 = true
	if !Exists("./test") {
		os.Mkdir("./test", 0755)
	}
	// fmt.Printf("Before\n")
	ioutil.WriteFile("./test/old.txt", []byte(`# Test version of old file`), 0644)
	ioutil.WriteFile("./test/ok.in", []byte(`ok.in`), 0644)
	ioutil.WriteFile("./test/rb.out", []byte(`# Error - if this is found - output should be overwritten #`), 0644)
	const delay = 1001 * time.Millisecond
	time.Sleep(delay)
	// fmt.Printf("After\n")
	ioutil.WriteFile("./test/new.txt", []byte(`# Test version of old file`), 0644)
	ioutil.WriteFile("./test/ok.out", []byte(`ok.in`), 0644)
	ioutil.WriteFile("./test/rb.in", []byte(`rb.in`), 0644)
}

func SetupTestCreateDirs() {
	if setup_done { // only run once per test - create logs, files for test.
		return
	}
	setup_done = true
	if !Exists("./test") {
		os.Mkdir("./test", 0755)
	}
	if !Exists("./log") {
		os.Mkdir("./log", 0755)
	}
	if !Exists("./cfg") {
		os.Mkdir("./cfg", 0755)
	}
	if !Exists("./www") {
		os.Mkdir("./www", 0755)
	}
	if !Exists("./www/foo") {
		os.Mkdir("./www/foo", 0755)
	}
	if !Exists("./www/testdir") {
		os.Mkdir("./www/testdir", 0755)
	}
	if !Exists("./www/testdir/js") {
		os.Mkdir("./www/testdir/js", 0755)
	}
	if !Exists("./www/testdir/a-dir") {
		os.Mkdir("./www/testdir/a-dir", 0755)
	}

	ioutil.WriteFile("./www/testdir/js/ex.js", []byte(`var v = 1;
`), 0644)

	ioutil.WriteFile("./www/testdir/js/ex2.js", []byte(`

// Cookie code from Quirksmode and other Places
// Modified by me in about a decade ago.

function createCookie(name,value,days) {
	var expires = "";
	if (days) {
		var date = new Date();
		date.setTime(date.getTime()+(days*24*60*60*1000));
		expires = "; expires="+date.toGMTString();
	} 
	document.cookie = name+"="+value+expires+"; path=/";
}

function getCookie(name) {
	var nameEQ = name + "=";
	var ca = document.cookie.split(';');
	for(var i=0;i < ca.length;i++) {
		var c = ca[i];
		while (c.charAt(0)==' ') c = c.substring(1,c.length);
		if (c.indexOf(nameEQ) == 0) {
			return c.substring(nameEQ.length,c.length);
		}
	}
	return null;
}

function delCookie(name) {
	createCookie(name,"",-1);
}

`), 0644)

	ioutil.WriteFile("./www/foo/t1.md", []byte(`Title for t1.html
===============

Body y Body y Body y Body y

Bodz z Bodz z Bodz z Bodz z

`), 0644)
	ioutil.WriteFile("./cfg/.htaccess", []byte(`# Test version of .htaccess ( testme, bobbob )
testme:example.com:58efc4116248aaffe3eb010fa43805b7
`), 0644)
	ioutil.WriteFile("./www/index.html", []byte(`<html><body>
Index.html Demo File
</body></html>
`), 0644)
	// xyzzy - add a index.tmpl file
	ioutil.WriteFile("./www/testdir/a.html", []byte(` a.html
`), 0644)
	ioutil.WriteFile("./www/testdir/b.html", []byte(`
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html b.html
`), 0644)
	ioutil.WriteFile("./www/testdir/c.html", []byte(` c.html
`), 0644)
	ioutil.WriteFile("./www/index.tmpl", []byte(`<html><body>
<ul>
{{range $ii, $ee := .files}}
	<li><a href="{{$ee.name}}">{{$ee.name}}</a></li>
{{end}}
</ul>
</body></html>
`), 0644)

}

var invalidMode = errors.New("Invalid Mode")

func Fopen(fn string, mode string) (file *os.File, err error) {
	file = nil
	if mode == "r" {
		file, err = os.Open(fn) // For read access.
	} else if mode == "w" {
		file, err = os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	} else if mode == "a" {
		file, err = os.OpenFile(fn, os.O_RDWR|os.O_APPEND, 0660)
		if err != nil {
			file, err = os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		}
	} else {
		err = invalidMode
	}
	return
}

func ExistsIsDir(name string) bool {
	fi, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	if fi.IsDir() {
		return true
	}
	return false
}

// ----------------------------------------------------------------------------------------------------------------------------------------------------
//		if h, f := lib.PathsExact(hdlr.Exact, req.URL.Path) ; f{
//func PathsExact(Paths []string, APath string) ( hdlr http.Handler, rv bool ) {
//	if Paths == nil || len(Paths) == 0 {
//		return nil, false
//	}
//	for _, it := range Paths {
//		if APath == it {
//			return nil, true
//		}
//	}
//	return nil, false
//}

// ----------------------------------------------------------------------------------------------------------------------------------------------------
func PathsMatch(Paths []string, APath string) bool {
	if Paths == nil || len(Paths) == 0 {
		return true
	}
	for _, prefix := range Paths {
		if strings.HasPrefix(APath, prefix) {
			return true
		}
	}
	return false
}

func PathsMatchN(Paths []string, APath string) int {
	if Paths == nil || len(Paths) == 0 {
		return 0
	}
	for ii, prefix := range Paths {
		if strings.HasPrefix(APath, prefix) {
			return ii
		}
	}
	return -1
}

func PathsMatchIgnore(Paths []string, APath string) bool {
	if Paths == nil || len(Paths) == 0 {
		return false
	}
	for _, prefix := range Paths {
		if strings.HasPrefix(APath, prefix) {
			return true
		}
	}
	return false
}

func PathsMatchPos(Paths []string, APath string) int {
	if Paths == nil || len(Paths) == 0 {
		return -1
	}
	for jj, prefix := range Paths {
		if strings.HasPrefix(APath, prefix) {
			return jj
		}
	}
	return -1
}

func GetRemoteIP(req *http.Request) string {
	if xForwardedHdr := req.Header.Get("X-Forwarded-For"); xForwardedHdr != "" {
		return xForwardedHdr
	}
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return host
}

func GetPort(req *http.Request) (rv string) {
	_, port, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return
	}
	rv = port
	return
}

// xyzzy - test this
// Match a path in a list of paths
func MatchPathInList(path string, pathList []string) bool {
	if InArray(path, pathList) {
		return true
	}
	a, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	for _, vv := range pathList {
		b, err := filepath.Abs(vv)
		if err != nil {
			return false
		}
		if a == b {
			return true
		}
	}
	return false
}

// xyzzy - put out into own project - cache of Re's

type ReCacheType struct {
	Pattern string
	Re      *regexp.Regexp
}

var reCache map[string]*ReCacheType
var reLock sync.Mutex

func init() {
	reCache = make(map[string]*ReCacheType)
}

func DumpReCache() {
	for ii, vv := range reCache {
		fmt.Printf("%s == %s %p\n", ii, vv.Pattern, vv.Re)
	}
}

func LookupRe(pattern string) (rv *regexp.Regexp) {
	var err error
	reLock.Lock()
	defer reLock.Unlock()
	// DumpReCache()
	rt, ok := reCache[pattern]
	if !ok {
		rv, err = regexp.Compile(pattern)
		if err != nil {
			fmt.Printf("Error: Invalid regular expression %s, Error: %s\n", pattern, err)
			return nil
		}
		rt := &ReCacheType{Pattern: pattern, Re: rv}
		reCache[pattern] = rt
	} else {
		rv = rt.Re
	}
	return
}

func PathsMatchRe(Paths []string, APath string) bool {
	var err error
	var re *regexp.Regexp
	if Paths == nil || len(Paths) == 0 {
		return true
	}
	// fmt.Printf("\n")
	for _, prefix := range Paths {
		if false { // Regenerate the RE each time
			re, err = regexp.Compile(prefix) // xyzzy - cache and compile - 500 cached or something!
			if err != nil {
				fmt.Printf("Error: Invalid regular expression %s, Error: %s\n", prefix, re)
				return false
			}
		} else { // use cached RE
			// fmt.Printf("Matching %s v.s. %s -- 715\n", prefix, APath)
			re = LookupRe(prefix)
			if re == nil {
				// fmt.Printf("  Bad Lookup - error\n")
				return false
			}
		}
		if re.MatchString(APath) {
			// fmt.Printf("  It Matched\n")
			return true
		}
	}
	// fmt.Printf("  NO Match\n")
	return false
}

func PathsMatchReN(Paths []string, APath string) int {
	var err error
	var re *regexp.Regexp
	if Paths == nil || len(Paths) == 0 {
		return 0
	}
	// fmt.Printf("\n")
	for ii, prefix := range Paths {
		if false { // Regenerate the RE each time
			re, err = regexp.Compile(prefix) // xyzzy - cache and compile - 500 cached or something!
			if err != nil {
				fmt.Printf("Error: Invalid regular expression %s, Error: %s\n", prefix, re)
				return -1
			}
		} else { // use cached RE
			// fmt.Printf("Matching %s v.s. %s -- 715\n", prefix, APath)
			re = LookupRe(prefix)
			if re == nil {
				// fmt.Printf("  Bad Lookup - error\n")
				return -1
			}
		}
		if re.MatchString(APath) {
			// fmt.Printf("  It Matched\n")
			return ii
		}
	}
	// fmt.Printf("  NO Match\n")
	return -1
}

func Md5sum(s string) (buf string) {
	buf = fmt.Sprintf("%x", md5.Sum([]byte(s)))
	return
}

func JsonStringToData(s string) (theJSON map[string]interface{}, err error) {
	err = json.Unmarshal([]byte(s), &theJSON)
	if err != nil {
		theJSON = make(map[string]interface{})
	}
	return
}

func JsonStringToArrayOfData(s string) (theJSON []map[string]interface{}, err error) {
	err = json.Unmarshal([]byte(s), &theJSON)
	if err != nil {
		theJSON = make([]map[string]interface{}, 0, 1)
	}
	return
}

// c_parsed, err2 := lib.JsonStringToArrayOfString ( b )
func JsonStringToArrayOfString(s string) (theJSON []string, err error) {
	err = json.Unmarshal([]byte(s), &theJSON)
	if err != nil {
		theJSON = make([]string, 0, 1)
	}
	return
}

func JsonStringToString(s string) (theJSON map[string]string, err error) {
	err = json.Unmarshal([]byte(s), &theJSON)
	if err != nil {
		theJSON = make(map[string]string)
	}
	return
}

func LenOfMap(ww map[string]interface{}) (rv int) {
	rv = 0
	for ii := range ww {
		_ = ii
		rv++
	}
	return
}

func FirstName(ww map[string]interface{}) (rv string) {
	for ii := range ww {
		return ii
	}
	return ""
}

var BoolError = errors.New("Invalid boolean value")

func ParseBool(s string) (bool, error) {
	switch s {
	case "true", "True", "TRUE", "y", "Y", "Yes", "yes", "yep", "1", "on", "ON", "On":
		return true, nil
	case "false", "False", "FALSE", "n", "N", "No", "no", "nope", "0", "off", "OFF", "Off":
		return false, nil
	}
	return false, BoolError
}

func SearchForFileSimple(root []string, fn string, index []string) (x string, ok bool) {
	for _, rr := range root {
		fn = filepath.Clean(rr + "/" + fn)
		if ExistsIsDir(fn) {
			for _, indexfile := range index {
				if Exists(fn + "/" + indexfile) {
					x = fn + "/" + indexfile
					ok = true
					return
				}
			}
			ok = false
			return
		}
		if Exists(fn) {
			x = fn
			ok = true
		}
	}
	return

}

func TimeZero() (z time.Time) {
	return
}

func GenSHA(data []byte) (s string) {
	// s = fmt.Sprintf("%x", sha256.Sum256(data))
	var b []byte
	x := sha256.Sum256(data)
	b = x[0:32]
	s = base64.URLEncoding.EncodeToString(b)
	return
}

/*
// get is like Get, but key must already be in CanonicalHeaderKey form.
func (h Header) get(key string) string {
	if v := h[key]; len(v) > 0 {
		return v[0]
	}
	return ""
}
*/

// -------------------------------------------------------------------------------------------------
/*
type FileInfo interface {
	Name() string       // base name of the file
	Size() int64        // length in bytes for regular files; system-dependent for others
	Mode() FileMode     // file mode bits
	ModTime() time.Time // modification time
	IsDir() bool        // abbreviation for Mode().IsDir()
	Sys() interface{}   // underlying data source (can return nil)
}
*/
// Exists reports whether the named file or directory exists.
// -------------------------------------------------------------------------------------------------
func ExistsGetFileInfo(name string) (bool, os.FileInfo) {
	fi, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
	}
	return true, fi
}

func GetIP(req *http.Request) (ip string) {
	ip = req.Header.Get("X-Forwarded-For")
	if ip == "" {
		h, _, err := net.SplitHostPort(req.RemoteAddr)
		if err == nil {
			ip = h
		}
	}
	if ip == "" {
		ip = "127.0.0.1"
	}
	return
}

// cookieValue := lib.GetCookie ( "LoginAuthCookie", req )
// -------------------------------------------------------------------------------------------------
func GetCookie(name string, req *http.Request) (rv string) {

	Ck := req.Cookies()
	for _, v := range Ck {
		if v.Name == name {
			rv = v.Value
			return
		}
	}
	return ""
}

func ExistsGetUDate(name string) (bool, os.FileInfo) {
	fi, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
	}
	return true, fi
}

func ReadJsonFile(fn string) (jdata map[string]string, err error) {
	jdata = make(map[string]string, 40) // The posts that match
	file, err := ioutil.ReadFile(fn)
	if err != nil {
		return
	}

	err = json.Unmarshal(file, &jdata)
	return
}

// ---------------------------------------------------------------------------------------------------------------------------------------------------------

var mime map[string]string

func init() {
	mime = make(map[string]string)
	mime["audio/basic"] = "au"
	mime["video/msvideo, video/avi, video/x-msvideo"] = "avi"
	mime["image/bmp"] = "bmp"
	mime["application/x-bzip2"] = "bz2"
	mime["text/css"] = "css"
	mime["application/xml-dtd"] = "dtd"
	mime["application/msword"] = "doc"
	mime["application/vnd.openxmlformats-officedocument.wordprocessingml.document"] = "docx"
	mime["application/vnd.openxmlformats-officedocument.wordprocessingml.template"] = "dotx"
	mime["application/ecmascript"] = "es"
	mime["application/octet-stream"] = "exe"
	mime["image/gif"] = "gif"
	mime["application/x-gzip"] = "gz"
	mime["application/mac-binhex40"] = "hqx"
	mime["text/html"] = "html"
	mime["application/java-archive"] = "jar"
	mime["image/jpeg"] = "jpg"
	mime["application/x-javascript"] = "js"
	mime["audio/x-midi"] = "midi"
	mime["audio/mpeg"] = "mp3"
	mime["video/mpeg"] = "mpeg"
	mime["audio/vorbis, application/ogg"] = "ogg"
	mime["application/pdf"] = "pdf"
	mime["application/x-perl"] = "pl"
	mime["image/png"] = "png"
	mime["application/vnd.openxmlformats-officedocument.presentationml.template"] = "potx"
	mime["application/vnd.openxmlformats-officedocument.presentationml.slideshow"] = "ppsx"
	mime["application/vnd.ms-powerpointtd>"] = "ppt"
	mime["application/vnd.openxmlformats-officedocument.presentationml.presentation"] = "pptx"
	mime["application/postscript"] = "ps"
	mime["video/quicktime"] = "qt"
	mime["audio/x-pn-realaudio, audio/vnd.rn-realaudio"] = "ra"
	mime["audio/x-pn-realaudio, audio/vnd.rn-realaudio"] = "ram"
	mime["application/rdf, application/rdf+xml"] = "rdf"
	mime["application/rtf"] = "rtf"
	mime["text/sgml"] = "sgml"
	mime["application/x-stuffit"] = "sit"
	mime["application/vnd.openxmlformats-officedocument.presentationml.slide"] = "sldx"
	mime["image/svg+xml"] = "svg"
	mime["application/x-shockwave-flash"] = "swf"
	mime["application/x-tar"] = "tar.gz"
	mime["application/x-tar"] = "tgz"
	mime["image/tiff"] = "tiff"
	mime["text/tab-separated-values"] = "tsv"
	mime["text/plain"] = "txt"
	mime["audio/wav, audio/x-wav"] = "wav"
	mime["application/vnd.ms-excel.addin.macroEnabled.12"] = "xlam"
	mime["application/vnd.ms-excel"] = "xls"
	mime["application/vnd.ms-excel.sheet.binary.macroEnabled.12"] = "xlsb"
	mime["application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"] = "xlsx"
	mime["application/vnd.openxmlformats-officedocument.spreadsheetml.template"] = "xltx"
	mime["application/xml"] = "xml"
	mime["application/zip, application/x-compressed-zip"] = "zip"
}

// function to get extension based on mime type
func GetExtenstionBasedOnMimeType(mt string) (ext string) {
	i := strings.Index(mt, ";")
	if i >= 0 && i < len(mt) {
		mt = mt[:i]
	}
	ext, ok := mime[mt]
	if !ok {
		ext = "unk"
	}
	return
}

func GetCTypes(www http.ResponseWriter, bod []byte) (ctypes []string) {
	ctypes, haveType := www.Header()["Content-Type"]
	if !haveType {
		ctype := http.DetectContentType(bod)
		www.Header().Set("Content-Type", ctype)
		ctypes = append(ctypes, ctype)
	}
	return
}

type MimeType int

const (
	HtmlMimeType MimeType = iota
	CssMimeType
	JSMimeType
	PngMimeType
	JpgMimeType
	GifMimeType
	WebpMimeType
	FontMimeType
)

func IsHtml(www http.ResponseWriter, bod []byte) bool {
	mt := GetCTypes(www, bod)
	if len(mt) > 0 && mt[0] == "text/html" {
		return true
	}
	return false
}

func FilepathAbs(fn string) (rfn string) {
	var err error
	rfn, err = filepath.Abs(fn)
	if err != nil {
		rfn = fn
	}
	return
}

// Return the full URL with http://host:port/path/path?arg=value
func GenURL(www http.ResponseWriter, req *http.Request) (url string, hUrl string) {
	http := "http:/"
	if req.TLS != nil {
		http = "https:/"
	}
	url = http + req.Host + req.RequestURI
	hUrl = Sha256(url)
	return
}

// Return a list of IPs that are listend to by this system.
func ExternalIP() (rv []string, err error) {
	ifaces, err1 := net.Interfaces()
	if err1 != nil {
		err = err1
		return
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err1 := iface.Addrs()
		if err1 != nil {
			err = err1
			return
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			rv = append(rv, ip.String())
		}
		if len(rv) > 0 {
			return
		}
	}
	err = errors.New("are you connected to the network?")
	return
}

// Return absolute value of integer
func IntAbs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// ----------------------------------------------------------------------------------------------------------------------------
/*
	"github.com/pschlump/gosrp"   // Path: /Users/corwin/go/src/www.2c-why.com/gosrp
func Sha256(s string) (rv string) {
	rv = gosrp.Hashstring(s)
	return
}
*/

/*
package main

import (
	"fmt"
	"regexp"
)

func main() {
	re := regexp.MustCompile(`^[ ]*([a-zA-Z][_a-zA-Z0-9]*)[ ]*(==|!=)[ ]*"([^"]*)"[ ]*$`)		// needs update //
	fmt.Printf("%q\n", re.FindAllStringSubmatch("abc == \"123\"", -1))
}

// Output: [["abc == \"123\"" "abc" "==" "\"123\""]]

*/

var reAFilter *regexp.Regexp

func init() {

	// TODO: no way to escape " in string - bad
	// TODO: need to implement a complete expression parse and match
	reAFilter = regexp.MustCompile("^[ \t]*([a-zA-Z][_a-zA-Z0-9]*)[ \t]*(==|!=)[ \t]*[\"']([^\"]*)[\"'][ \t]*$")

}

type OpType int

const (
	OpEq OpType = 1
	OpNe OpType = 2
)

var OpLookup map[string]OpType
var OpLookupStr map[OpType]string

func init() {
	OpLookup = map[string]OpType{
		"==": OpEq,
		"!=": OpNe,
	}
	OpLookupStr = map[OpType]string{
		OpEq: "==",
		OpNe: "!=",
	}
}

func (x OpType) String() string {
	if vv, ok := OpLookupStr[x]; ok {
		return vv
	}
	return "--undefined op--"
}

var ErrFailedToParse = errors.New("Error in parsing filter")

func ParseFilter(ss string) (ff *FilterType, err error) {

	rv := reAFilter.FindAllStringSubmatch(ss, -1)
	if len(rv) == 0 {
		err = ErrFailedToParse
		return
	}

	godebug.Printf(db_filter, "Filter: rv=%q, FilterType=%s, %s\n", rv, SVar(FilterType{Name: rv[0][1], Value: rv[0][3], Op: OpLookup[rv[0][2]]}), godebug.LF())

	return &FilterType{
		Name:  rv[0][1],
		Value: rv[0][3],
		Op:    OpLookup[rv[0][2]],
	}, nil
}

type FilterType struct {
	Name  string
	Value string
	Op    OpType
}

func ApplyFilter(filter []*FilterType, mdata map[string]interface{}) bool {

	godebug.Printf(db_filter, "Filter: at top ApplyFilter, %s\n", godebug.LF())

	for _, ff := range filter {
		vvI, fnd := mdata[ff.Name]
		if !fnd {
			return false
		}
		vv, ok := vvI.(string)
		if !ok {
			return false
		}
		switch ff.Op {
		case OpEq:
			if vv == ff.Value {
				godebug.Printf(db_filter, "Filter: ApplyFilter -- matche equal, %s\n", godebug.LF())
			} else {
				return false
			}
		case OpNe:
			if vv != ff.Value {
				godebug.Printf(db_filter, "Filter: ApplyFilter -- matche not-equal, %s\n", godebug.LF())
			} else {
				return false
			}
		}
	}

	return true

}

const db_filter = false // filters and parsing of them

// Remove the `nth` element from the slice `s` returning the modified slice.
// If `nth` is out of range then just return the input `s`
func DelFromSliceString(s []string, nth int) (rv []string) {
	if nth == 0 {
		rv = s[1:]
	} else if nth == len(s)-1 {
		rv = s[0:nth]
	} else if nth >= 0 && nth < len(s)-1 {
		rv = append(s[:nth], s[nth+1:]...)
	} else {
		rv = s
	}
	return
}

//------------------------------------------------------------------------------------------------------------------------------
//func RedisClient() (client *redis.Client, conFlag bool) {
//	var err error
//	client, err = redis.Dial("tcp", cfg.ServerGlobal.RedisConnectHost+":"+cfg.ServerGlobal.RedisConnectPort)
//	if err != nil {
//		log.Fatal(err)
//	}
//	if cfg.ServerGlobal.RedisConnectAuth != "" {
//		err = client.Cmd("AUTH", cfg.ServerGlobal.RedisConnectAuth).Err
//		if err != nil {
//			log.Fatal(err)
//		} else {
//			conFlag = true
//		}
//	} else {
//		conFlag = true
//	}
//	return
//}

func GetUUIDAsString() (rv string) {
	id0x, _ := uuid.NewV4()
	rv = id0x.String()
	return
}

// GetImageDimension returns width, height of an immage - .png, .jpg, .gif
//	- not working on .svg
//	- not returning errors
//
// Example:
//		h, w = GetHWFromImage ( ffile_name );
func GetImageDimension(imagePath string) (width int, height int) {
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	defer file.Close()

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", imagePath, err)
		return
	}
	return image.Width, image.Height
}

// GetFileSize returns the file size in bytes
func GetFileSize(fn string) int64 {
	stat, err := os.Stat(fn)
	if err != nil {
		return 0
	}
	return stat.Size()
}

/* vim: set noai ts=4 sw=4: */
