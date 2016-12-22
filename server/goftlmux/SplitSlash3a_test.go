package goftlmux

//
// Go Go Mux - Go Fast Mux / Router for HTTP requests
//
// (C) Philip Schlump, 2013-2014.
// Version: 0.4.3
// BuildNo: 804
//
//

import (
	"testing"

	debug "github.com/pschlump/godebug"
)

var testSplitSlash3_Data = []struct {
	Url             string
	Result          string
	NewUrl          string
	ResultEarlyExit string
}{
	{"/repos/julienschmidt/httprouter/stargazers", "[0,6,20,31,42]", "/repos/julienschmidt/httprouter/stargazers", "[0,6]"},
	{"/planb/:vA/t1/:vB", "[0,6,10,13,17]", "/planb/:vA/t1/:vB", "[0,6]"},
	{"/planb/:vD/t2/:vE", "[0,6,10,13,17]", "/planb/:vD/t2/:vE", "[0,6]"},
	{"/:vE", "[0,4]", "/:vE", "[0,4]"},
	{"/vE", "[0,3]", "/vE", "[0,3]"},
	{"/*x", "[0,3]", "/*x", "[0,3]"},
	{"/*filename", "[0,10]", "/*filename", "[0,10]"},
	{"/a/b/c/d/e/f/g/h/i/j", "[0,2,4,6,8,10,12,14,16,18,20]", "/a/b/c/d/e/f/g/h/i/j", "[0,2]"},
	{"/a//c/./e/f/../h/i/j", "[0,2,4,6,8,10,12]", "/a/c/e/h/i/j", "[0,2]"},
	{"///c/./e/f/../h/i/j", "[0,2,4,6,8,10]", "/c/e/h/i/j", "[0,2]"},
}

func TestSplitOnSlash3a(t *testing.T) {
	htx := NewRouter()
	var procData MuxRouterProcessing
	InitMuxRouterProcessing(htx, &procData)

	// fmt.Printf("This ONe\n")
	for i, test := range testSplitSlash3_Data {
		htx.SplitOnSlash3(&procData, 1, test.Url, false)
		// fmt.Printf("NSl = %d: ->%s<- Slash %s, %s\n", htx.NSl, htx.CurUrl, debug.SVar(htx.Slash[0:htx.NSl+1]), debug.LF())
		rv := debug.SVar(procData.Slash[0 : procData.NSl+1])
		if rv != test.Result {
			t.Errorf("SplitOnSlash3 [%d] URL:%s failed, Expected ->%s<- got ->%s<-\n", i, test.Url, test.Result, rv)
		}
		if procData.CurUrl != test.NewUrl {
			t.Errorf("SplitOnSlash3 [%d] URL:%s failed, Expected ->%s<- got ->%s<-\n", i, test.Url, test.NewUrl, procData.CurUrl)
		}
	}
	for i, test := range testSplitSlash3_Data {
		htx.SplitOnSlash3(&procData, 1, test.Url, true)
		// fmt.Printf("NSl = %d: ->%s<- Slash %s, %s\n", htx.NSl, htx.CurUrl, debug.SVar(htx.Slash[0:htx.NSl+1]), debug.LF())
		rv := debug.SVar(procData.Slash[0 : procData.NSl+1])
		if rv != test.ResultEarlyExit {
			t.Errorf("SplitOnSlash3 [%d] URL:%s failed, Expected ->%s<- got ->%s<-\n", i, test.Url, test.ResultEarlyExit, rv)
		}
	}
}

// 148 ns - Old Structured Version
// 127 ns - new version
// 29.2 ns - early exit - no /repos in hash table.
func OldBenchmarkOfSplitOnSlash3_long(b *testing.B) {
	htx := NewRouter()
	var procData MuxRouterProcessing
	InitMuxRouterProcessing(htx, &procData)

	url := "/repos/julienschmidt/httprouter/stargazers"
	for n := 0; n < b.N; n++ {
		htx.SplitOnSlash3(&procData, 1, url, true)
	}
}

// 25.4 ns - Old Structured Version
// 21.2 ns - new version
// 28.4 ns - early exit vesion for (index.html)
func OldBenchmarkOfSplitOnSlash3_short(b *testing.B) {
	htx := NewRouter()
	var procData MuxRouterProcessing
	InitMuxRouterProcessing(htx, &procData)

	url := "/repos"
	url = "/index.html"
	for n := 0; n < b.N; n++ {
		htx.SplitOnSlash3(&procData, 1, url, true)
	}
}
