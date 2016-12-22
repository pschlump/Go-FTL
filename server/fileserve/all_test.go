// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1010
//

package fileserve

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kr/pretty"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
)

// // -----------------------------------------------------------------------------------------------------------------------------------------------
// // should be moved to ../lib
// func Test_FileServe_00(t *testing.T) {
//
// 	var s string
//
// 	s = RmExt("abc.md")
// 	if s != "abc" {
// 		t.Errorf("Error 0001, Expected >abc< got >%s<\n", s)
// 	}
//
// 	s = RmExt("abc")
// 	if s != "abc" {
// 		t.Errorf("Error 0002, Expected >abc< got >%s<\n", s)
// 	}
//
// 	s = RmExt("")
// 	if s != "" {
// 		t.Errorf("Error 0003, Expected >< got >%s<\n", s)
// 	}
//
// 	s = RmExt("abc.min.js")
// 	if s != "abc.min" {
// 		t.Errorf("Error 0004, Expected >abc.min< got >%s<\n", s)
// 	}
//
// 	s = RmExtSpecified("abc.min.js", ".js")
// 	if s != "abc.min" {
// 		t.Errorf("Error 0005, Expected >abc.min< got >%s<\n", s)
// 	}
//
// 	s = RmExtSpecified("abc.min.js", ".min.js")
// 	if s != "abc" {
// 		t.Errorf("Error 0006, Expected >abc< got >%s<\n", s)
// 	}
//
// 	s = RmExtSpecified("abc.min.js", "abc.min.js")
// 	if s != "" {
// 		t.Errorf("Error 0007, Expected >< got >%s<\n", s)
// 	}
//
// 	s = RmExtSpecified("abc.min.js", ".html")
// 	if s != "abc.min.js" {
// 		t.Errorf("Error 0008, Expected >abc.min.js< got >%s<\n", s)
// 	}
// }
//
// // -----------------------------------------------------------------------------------------------------------------------------------------------
// // should be moved to ../lib
// //func CompareModTime(in, out time.Time) bool {
// // https://golang.org/src/time/sleep_test.go
// func Test_FileServe_01(t *testing.T) {
//
// 	var b, shouldRebuild RebuildFlag
//
// 	t1 := time.Now()
// 	t2 := time.Now()
//
// 	b = CompareModTime(t1, t2)
// 	if shouldRebuild == NeedRebuild {
// 		t.Errorf("Error 0101, Expected >true< got >%v<\n", b)
// 	}
//
// 	lib.SetupTestCreateDirsFileServe()
//
// 	ok1, inFi := lib.ExistsGetFileInfo("./test/old.txt")
// 	ok2, outFi := lib.ExistsGetFileInfo("./test/new.txt")
// 	if !ok1 || !ok2 {
// 		t.Errorf("Error 0102, Test file missing\n")
// 	}
//
// 	shouldRebuild = CompareModTime(inFi.ModTime(), outFi.ModTime())
// 	if shouldRebuild == NeedRebuild {
// 		t.Errorf("Error 0103, Expected >NeedRebuild< got >%s<\n", shouldRebuild)
// 	}
//
// 	shouldRebuild = CompareModTime(outFi.ModTime(), inFi.ModTime())
// 	if shouldRebuild != NeedRebuild {
// 		t.Errorf("Error 0104, Expected >NOT NeedRebuild< got >%s<\n", shouldRebuild)
// 	}
//
// }
//
// // func runCmdIfNecessary(
// // 	fcfg *FileServerType, www http.ResponseWriter, req *http.Request,
// // 	inputFn string, haveInput bool, inFi os.FileInfo, inExt string,
// // 	outputFn string, haveOutput bool, outFi os.FileInfo, outExt string,
// // 	ti int, tr *ExtProcessType) {
// // 		Create a ./test directory
// // 		Create 2 files ./test/old.test and ./test/new.test
//
// func Test_FileServe_02(t *testing.T) {
//
// 	lib.SetupTestCreateDirsFileServe()
//
// 	cmd := `{ "Cmd":"cp", "Params":[ "{{.inputFile}}", "{{.outputFile}}" ] }`
// 	ok, out, err := ExecuteCommands(cmd, "./test/rb.in", "./test/rb.out", ".in", ".out")
// 	fmt.Printf("ok=%v out=%v err=%v\n", ok, out, err)
// 	// ioutil.WriteFile("./test/rb.in", []byte(`rb.in`), 0644)
// 	// ioutil.WriteFile("./test/rb.out", []byte(`# Error - if this is found - output should be overwritten #`), 0644)
//
// 	data, err1 := ioutil.ReadFile("./test/rb.out")
// 	if err1 != nil {
// 		t.Errorf("Error 0201, File is missing after copy operaiton\n")
// 	}
//
// 	if string(data) != `rb.in` {
// 		t.Errorf("Error 0202, Wrong contents for ./test/rb.out, got >%s<, expected >rb.in<\n", data)
// 	}
//
// }

// func UrlFileExt(fcfg *FileServerType, www http.ResponseWriter, req *http.Request, urlIn string, g *FSConfig, rulNo int) (urlOut string, rootOut string, stat RuleStatus, err error) {
func Test_FileServe_03(t *testing.T) {

	fmt.Printf("\n\n\n - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -  Test_FileServer_03 - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -\n\n\n")

	lib.SetupTestCreateDirsFileServe()

	tests := []struct {
		runTest bool
		url     string
		hdr     []lib.NameValue
	}{
		{
			true,
			"/foo/t1.html",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
		},
		{
			false,
			"http://example.com/foo?abc=def&$privs$=user&t=user&template_name=data.tmpl",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
		},
		{
			false,
			"http://example.com/def?$privs$=user&t=user",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
		},
	}

	bot := mid.NewServer()
	ms := NewFileServer(bot, nil, []string{"/foo"}, []string{"index.html", "index.htm"}, []string{"./www"})
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		if !test.runTest {
			continue
		}

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec, nil) // var wr http.ResponseWriter
		// lib.SetupTestCreateHeaders(wr, test.hdr)

		var req *http.Request

		req, err = http.NewRequest("GET", test.url, nil)
		if err != nil {
			t.Fatalf("Test %d: Could not create HTTP request: %v", ii, err)
		}
		goftlmux.ParseQueryParamsReg(wr, req, &wr.Ps) //
		lib.SetupTestMimicReq(req, "example.com")
		if db3 {
			fmt.Printf("{\"req\":%s,\n\"wr\":%s}\n", lib.SVarI(req), lib.SVarI(wr))
		}
		lib.SetupRequestHeaders(req, test.hdr)

		urlOut, rootOut, stat, err := UrlFileExt(ms, wr, req, test.url, ms.Cfg, ii)

		if db1 {
			fmt.Printf("urlOut >%s< rootOut >%s< stat=%s err=%v, %s\n", urlOut, rootOut, stat, err, godebug.LF())
		}

		if false {
			wr.FinalFlush()
			s := wr.GetBody()
			fmt.Printf("Final body =====>>>>>%s<<<<<=====\n", s)
		}

	}

}

const db3 = false

func GetModTime(fn string) time.Time {
	file, err := os.Stat(fn)
	if err != nil {
		fmt.Printf("Error on stat of file (%s), %s, %s\n", fn, err, godebug.LF())
	}
	modifiedtime := file.ModTime()
	// fmt.Printf("Before touch modified time : %s\n", modifiedtime)
	return modifiedtime
}

// PJS defered cleanup after test
func afterTest(t *testing.T) {
	// Remove ./testdata/index.htm
	// Create ./testdata/index.html
	os.Remove("./testdata/hero.js")
	os.Remove("./testdata/hero.js.map")
}

func setup1_test(t *testing.T) {
	// Create ./testdata
	// Create ./testdata/index.html
	if lib.Exists("./testdata/index.htm") {
		err := os.Rename("./testdata/index.htm", "./testdata/index.html")
		if err != nil {
			fmt.Printf("Unable to setup test, rename of './testdata/index.htm' to './testdata/index.html', %s, %s\n", err, godebug.LF())
			t.Fatalf("Fatal -- ending test\n")
		}
	}
	os.Remove("./testdata/hero.js")
	os.Remove("./testdata/hero.js.map")
}

// Look for is a /testdata/...name, Fns is full paths compare and verify that each of LookFor is in the set FNs
func CheckPathContains(FNs, LookFor []string) bool {
	foundIt := func(FNs []string, ALookFor string) bool {
		for _, ww := range FNs {
			a := filepath.Clean(ww)
			b := filepath.Clean(ALookFor)
			if strings.HasSuffix(a, b) {
				return true
			}
		}
		return false
	}
	for _, vv := range LookFor {
		if !foundIt(FNs, vv) {
			return false
		}
	}
	return true
}

// This fails to set the cookie correctly -- jsut name value get set - enough for this test --
func SetCookieValue(req *http.Request, CookieName string, CookieValue string) {
	expire := time.Now().AddDate(0, 0, 1)
	cookie := http.Cookie{Name: CookieName, Value: CookieValue, Path: "/", Domain: "www.example.com", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400}
	req.AddCookie(&cookie)
}

//
// Purpose of this test:
//
//    ✓  0. Test that a file is returned                                            -- test 0
//    ✓  1. Test that "/" gets mapped to /index.html                                -- test 2, 8, 9
//    ✓  2. Test that a search on multiple index.* files will work                  -- test 7
//    ✓  3. Test that cookie-based themes will lead to a correct theme directory    -- test 22, 23, 24, 25
//    ✓  4. Test that a correct cross - directory - theme search works              -- test 22, 23, 24, 25
//    ✓  5. Test that themes work with directory templates                          -- test 26, 27, 28, 29
//    ✓  6. Test that a file is generated from .ts -> .js using "tsc" - noFile      -- test 10
//    ✓  7. Test that caching information is returned properly up the stack         -- test 1, 11, 12, 13, 14, 15, 22, 23, 24, 25, 26, 27, 28, 29
//    ✓  8. Test that /..0x00 files will not get served ( 404 )                     -- test 3
//    ✓  9. Test that files that do not exists will get a 404                       -- test 4
//    ✓ 10. Test that files get returned with a correct mime type                   -- test 2, 5 +others
//    ✓ 11. Test that directories get served with default template                  -- test 1
//    ✓ 12. Test that directories get served with custom template                   -- test 19, 20
//    ✓ 13. Test that a file is generated from .ts -> .js using "tsc" - mtime       -- test 11 ... 18
//    ✓ 14. Test that a file is generated from .ts -> .js+.map using "tsc"          -- test 10
//    ✓ 16. Test rw.IgnoreDirs
//        if len(rw.IgnoreDirs) > 0 && lib.MatchPathInList(name, rw.IgnoreDirs) {
//
// Deferred:
//
//      15. Test that a single output form multiple inputs, "xcat a b c >d"                                <<< 5        1. implement go-cat, go-touch, go-if-older
//            /path/a.js++b.js++c.js
//      17. Test with a theme root that is not ./xxxx - use a different path than the file root.
//      18. Test that sha256 hash based ETags are generated and returned.
//      19. Test that sha256 hash based ETags are file+path
//
// Verify that configuration can:
//       0. Set different locations for commands - from config
//       1. Checks for the existence of each command - reports missing at startup
//       2. Check on how ETags are created and where
//
/*
TODO:
	Global process locks on running outside programs - serialize them - only run once. -- Should probably be using workers for this.
	Extended Caching Info - where - what - generate test -- from ../cfg/cfg.go
		// ---- Caching Config Type ----------------------------------------------------------------------------------------------------
		type CacheConfigType struct {
			CacheForSeconds             int       // if 0, then not applicable, 1 = cache till end of 1 second up, 2..n do not refresh until timeout
			FetchedTime                 time.Time // When was data fetched
			ProxiedData                 bool      // Data is from a proxy, false implies source files local and can be re-checked
			CacheAndRecheckDependencies bool      // Cache it but re-checked dependencies
			OutputFile                  []string  // full path to output
			IntermediateFile            []string  // set of files that represent intermediate files
			InputFile                   []string  // set of files that represent input - timestamps can be checked
			CacheAndRevalidate          bool      // Cache - but re-generate source and see if SHA256 is same, if so then 304 else re-send
			Sha256Hash                  string    // Hash of output data
			IgnoreTotally               bool      // Not catchable at all
			CacheIfLargerThan           uint64    // Ignore if data size is less than this
			IgnoreIfLargerThan          uint64    // Ignore if data size is bigger than this
			CachePaths                  []string  // paths to Cache
			IgnorePaths                 []string  // paths to ignore
			IgnoreCookies               []string  // paths to ignore
			MatchUrl_Cookies            []string  // Add these cookies to URL before a lookup
			Prefetch                    bool      // Pre fetch indicates that catch pre-fetching should occurs on this item
			PrefetchCount               int       // Pre fetch this number of items
			PrefetchFreq                int       // Time for pre-fetch - how often
			StaleAfter                  int       // Delta-T for item in pre-fetch going stale (shelf-life)
			FlushFromCache              bool      // Indicates that a lower level knows that this should be flushed from the cache
		}
*/

type TestPrePostType func(*Test05Type)

type Test05Type struct {
	RunTest              bool
	DbTest               bool
	url                  string
	expectedBody         string
	expectedCode         int
	expectedContentType  string
	CheckSuccessfulBuild bool
	CheckNoBuildOccured  bool
	DirTemplateFileName  string
	IgnoreDirs           string
	SetThemeCookie       string
	SetUserCookie        string
	CheckResolvedFn      string
	CheckDependentFNs    []string
	Msg                  string
	preFx                TestPrePostType
	postFx               TestPrePostType
	TestComment          string
}

func Test_FileServer_05(t *testing.T) {

	fmt.Printf("\n\n\n - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -  Test_FileServer_05 - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -\n\n\n")
	InitNLogged()

	var err error

	tests := []Test05Type{

		// 0. test file returned
		{
			RunTest:             true,
			url:                 "/td1.data",
			expectedBody:        `{"abc":"def"}` + "\n",
			expectedCode:        200,
			expectedContentType: "text/plain",
		},
		// 1. test directory with default template return values
		{
			RunTest: true,
			url:     "/a-dir/",
			expectedBody: `<pre>
<a href="file1.txt">file1.txt</a>
<a href="file2.txt">file2.txt</a>
</pre>
`,
			expectedCode:        200,
			expectedContentType: "text/html",
			CheckResolvedFn:     "/testdata/a-dir",
			CheckDependentFNs:   []string{"/testdata/a-dir"},
		},
		// 2. test with up 1 directory
		{
			RunTest:             true,
			url:                 "/a-dir/../",
			expectedBody:        "index.html says hello\n",
			expectedCode:        200,
			expectedContentType: "text/html",
		},
		// 3. test with 0x00 chars in requested file name
		{
			RunTest:      true,
			url:          "/a-dir/..\x00",
			expectedCode: 404,
		},
		// 4. test file with bad name, no '/' at beginning - 404 returned
		{
			RunTest:      true,
			url:          "td1.data",
			Msg:          "Expect error to be logged, 404",
			expectedBody: "",
			expectedCode: 404,
		},
		// 5. test file returned
		{
			RunTest: true,
			url:     "/t5.css",
			expectedBody: `.something {
	color: red;
}
`,
			expectedCode:        200,
			expectedContentType: "text/css",
		},
		// 6. sniff test -- TODO - investigate why not sniffed?
		{
			RunTest: true,
			url:     "/t6",
			expectedBody: `.something {
	color: red;
}
.something {
	color: red;
}
.something {
	color: red;
}
.something {
	color: red;
}
.something {
	color: red;
}
.something {
	color: red;
}
.something {
	color: red;
}
.something {
	color: red;
}
.something {
	color: red;
}
.something {
	color: red;
}
.something {
	color: red;
}
.something {
	color: red;
}
.something {
	color: red;
}
.something {
	color: red;
}
.something {
	color: red;
}
.something {
	color: red;
}
.something {
	color: red;
}
.something {
	color: red;
}
.something {
	color: red;
}
`,
			expectedCode:        200,
			expectedContentType: "text/plain",
		},
		// 7. index.htm
		{
			RunTest:             true,
			url:                 "/",
			expectedBody:        "index.html says hello\n",
			expectedCode:        200,
			expectedContentType: "text/html",
			preFx: func(_ *Test05Type) {
				os.Rename("./testdata/index.html", "./testdata/index.htm")
			},
			postFx: func(_ *Test05Type) {
				os.Rename("./testdata/index.htm", "./testdata/index.html")
			},
		},
		// 8. index.html
		{
			RunTest:             true,
			url:                 "/",
			expectedBody:        "index.html says hello\n",
			expectedCode:        200,
			expectedContentType: "text/html",
		},
		// 9. index.html
		{
			RunTest:             true,
			url:                 "/index.html",
			expectedBody:        "index.html says hello\n",
			expectedCode:        200,
			expectedContentType: "text/html",
		},
		// 10. hero.js + hero.js.map build from hero.ts
		{
			RunTest: true,
			url:     "/hero.js",
			Msg:     "Expect status success output",
			expectedBody: `"use strict";
var Hero = (function () {
    function Hero(id, name) {
        this.id = id;
        this.name = name;
    }
    return Hero;
}());
exports.Hero = Hero;
//# sourceMappingURL=hero.js.map`,
			expectedCode:         200,
			expectedContentType:  "application/javascript",
			CheckSuccessfulBuild: true,
			preFx: func(_ *Test05Type) {
				os.Remove("./testdata/hero.js")
				os.Remove("./testdata/hero.js.map")
			},
			postFx: func(_ *Test05Type) {
				err := os.Remove("./testdata/hero.js")
				if err != nil {
					t.Errorf("test 10 missing file ./testdata/hero.js\n")
				}
				err = os.Remove("./testdata/hero.js.map")
				if err != nil {
					t.Errorf("test 10 missing file ./testdata/hero.js.map\n")
				}
			},
		},
		// 11. to satisfy: Test that a file is generated from .ts -> .js using "tsc" - mtime
		//		a. build hero.js from hero.ts 	(11)
		//		b. verify that build occurred
		//		c. re-ask for hero.js			(12)
		//		d. verify that *NO* build occurred
		//		e. sleep 1 second ( in preFx )	(13)
		//		f. touch hero.ts ( in preFx )
		//		g. build hero.js from hero.ts
		//		h. verify that build occurred
		//		i. re-ask for hero.js			(14)
		//		j. verify that *NO* build occurred
		// Repeat with .map file for hero.js	(15...18)
		{
			RunTest: true,
			DbTest:  true,
			url:     "/hero.js",
			Msg:     "Expect status success output",
			expectedBody: `"use strict";
var Hero = (function () {
    function Hero(id, name) {
        this.id = id;
        this.name = name;
    }
    return Hero;
}());
exports.Hero = Hero;
//# sourceMappingURL=hero.js.map`,
			expectedCode:         200,
			expectedContentType:  "application/javascript",
			CheckSuccessfulBuild: true,
			CheckResolvedFn:      "/testdata/hero.js",
			CheckDependentFNs:    []string{"/testdata/hero.ts"},
			preFx: func(trun *Test05Type) {
				if trun.DbTest {
					fmt.Printf("%s in preFx, %s %s\n", MiscLib.ColorCyan, godebug.LF(), MiscLib.ColorReset)
				}
				os.Remove("./testdata/hero.js")
				os.Remove("./testdata/hero.js.map")
			},
			postFx: func(trun *Test05Type) {
				if trun.DbTest {
					fmt.Printf("%s in postFx, %s %s\n", MiscLib.ColorCyan, godebug.LF(), MiscLib.ColorReset)
				}
			},
		},
		// 12 - re ask
		{
			RunTest: true,
			url:     "/hero.js",
			expectedBody: `"use strict";
var Hero = (function () {
    function Hero(id, name) {
        this.id = id;
        this.name = name;
    }
    return Hero;
}());
exports.Hero = Hero;
//# sourceMappingURL=hero.js.map`,
			expectedCode:        200,
			expectedContentType: "application/javascript",
			CheckNoBuildOccured: true,
			CheckResolvedFn:     "/testdata/hero.js",
			CheckDependentFNs:   []string{"/testdata/hero.ts"},
		},
		// 13 - sleep, touch and ask
		{
			RunTest: true,
			url:     "/hero.js",
			Msg:     "Expect status success output",
			expectedBody: `"use strict";
var Hero = (function () {
    function Hero(id, name) {
        this.id = id;
        this.name = name;
    }
    return Hero;
}());
exports.Hero = Hero;
//# sourceMappingURL=hero.js.map`,
			expectedCode:         200,
			expectedContentType:  "application/javascript",
			CheckSuccessfulBuild: true,
			CheckResolvedFn:      "/testdata/hero.js",
			CheckDependentFNs:    []string{"/testdata/hero.ts"},
			preFx: func(trun *Test05Type) {

				if trun.DbTest {
					fmt.Printf("Before touch modified time hero.ts : %s\n", GetModTime("./testdata/hero.ts"))
					fmt.Printf("Before touch modified time hero.js : %s\n", GetModTime("./testdata/hero.js"))
				}

				time.Sleep(2 * time.Second)
				currenttime := time.Now().Local()
				err = os.Chtimes("./testdata/hero.ts", currenttime, currenttime)

				if trun.DbTest {
					fmt.Printf("Adfter touch modified time hero.ts : %s\n", GetModTime("./testdata/hero.ts"))
					fmt.Printf("Adfter touch modified time hero.js : %s\n", GetModTime("./testdata/hero.js"))
				}

			},
			postFx: func(_ *Test05Type) {
			},
		},
		// 14 - re ask
		{
			RunTest: true,
			url:     "/hero.js",
			expectedBody: `"use strict";
var Hero = (function () {
    function Hero(id, name) {
        this.id = id;
        this.name = name;
    }
    return Hero;
}());
exports.Hero = Hero;
//# sourceMappingURL=hero.js.map`,
			expectedCode:        200,
			expectedContentType: "application/javascript",
			CheckNoBuildOccured: true,
			CheckResolvedFn:     "/testdata/hero.js",
			CheckDependentFNs:   []string{"/testdata/hero.ts"},
		},

		// 15. Repeat with .map file for hero.js	(15...18)
		{
			RunTest:              true,
			url:                  "/hero.js.map",
			Msg:                  "Expect status success output",
			expectedBody:         `{"version":3,"file":"hero.js","sourceRoot":"","sources":["hero.ts"],"names":[],"mappings":";AAAA;IACE,cACS,EAAS,EACT,IAAW;QADX,OAAE,GAAF,EAAE,CAAO;QACT,SAAI,GAAJ,IAAI,CAAO;IAAI,CAAC;IAC3B,WAAC;AAAD,CAAC,AAJD,IAIC;AAJY,YAAI,OAIhB,CAAA"}`,
			expectedCode:         200,
			expectedContentType:  "text/plain",
			CheckSuccessfulBuild: true,
			preFx: func(_ *Test05Type) {
				os.Remove("./testdata/hero.js")
				os.Remove("./testdata/hero.js.map")
			},
			postFx: func(_ *Test05Type) {
			},
		},
		// 16 - re ask
		{
			RunTest:             true,
			url:                 "/hero.js.map",
			expectedBody:        `{"version":3,"file":"hero.js","sourceRoot":"","sources":["hero.ts"],"names":[],"mappings":";AAAA;IACE,cACS,EAAS,EACT,IAAW;QADX,OAAE,GAAF,EAAE,CAAO;QACT,SAAI,GAAJ,IAAI,CAAO;IAAI,CAAC;IAC3B,WAAC;AAAD,CAAC,AAJD,IAIC;AAJY,YAAI,OAIhB,CAAA"}`,
			expectedCode:        200,
			expectedContentType: "text/plain",
			CheckNoBuildOccured: true,
		},
		// 17 - sleep, touch and ask
		{
			RunTest:              true,
			url:                  "/hero.js.map",
			Msg:                  "Expect status success output",
			expectedBody:         `{"version":3,"file":"hero.js","sourceRoot":"","sources":["hero.ts"],"names":[],"mappings":";AAAA;IACE,cACS,EAAS,EACT,IAAW;QADX,OAAE,GAAF,EAAE,CAAO;QACT,SAAI,GAAJ,IAAI,CAAO;IAAI,CAAC;IAC3B,WAAC;AAAD,CAAC,AAJD,IAIC;AAJY,YAAI,OAIhB,CAAA"}`,
			expectedCode:         200,
			expectedContentType:  "text/plain",
			CheckSuccessfulBuild: true,
			preFx: func(trun *Test05Type) {

				if trun.DbTest {
					fmt.Printf("Before touch modified time hero.ts : %s\n", GetModTime("./testdata/hero.ts"))
					fmt.Printf("Before touch modified time hero.js : %s\n", GetModTime("./testdata/hero.js.map"))
				}

				time.Sleep(2 * time.Second)
				currenttime := time.Now().Local()
				err = os.Chtimes("./testdata/hero.ts", currenttime, currenttime)

				if trun.DbTest {
					fmt.Printf("Adfter touch modified time hero.ts : %s\n", GetModTime("./testdata/hero.ts"))
					fmt.Printf("Adfter touch modified time hero.js : %s\n", GetModTime("./testdata/hero.js.map"))
				}

			},
			postFx: func(_ *Test05Type) {
			},
		},
		// 18 - re ask
		{
			RunTest:             true,
			url:                 "/hero.js.map",
			expectedBody:        `{"version":3,"file":"hero.js","sourceRoot":"","sources":["hero.ts"],"names":[],"mappings":";AAAA;IACE,cACS,EAAS,EACT,IAAW;QADX,OAAE,GAAF,EAAE,CAAO;QACT,SAAI,GAAJ,IAAI,CAAO;IAAI,CAAC;IAC3B,WAAC;AAAD,CAAC,AAJD,IAIC;AAJY,YAAI,OAIhB,CAAA"}`,
			expectedCode:        200,
			expectedContentType: "text/plain",
			CheckNoBuildOccured: true,
		},
		// 19. directory with custom template
		{
			RunTest: true,
			url:     "/b-dir/",
			expectedBody: `
<html><body>
<h1>Template for Directories at top level</h1>
<ul>

	<li><a href="b-1.txt">b-1.txt</a></li>

	<li><a href="b-2.txt">b-2.txt</a></li>

	<li><a href="b-3.txt">b-3.txt</a></li>

	<li><a href="b-4.txt">b-4.txt</a></li>

</ul>
</body></html>
`,
			expectedCode:        200,
			expectedContentType: "text/html",
			DirTemplateFileName: "dir.tmpl",
		},
		// 20. directory with custom template - in a sub-directory
		{
			RunTest: true,
			url:     "/c-dir/",
			expectedBody: `
<html><body>
<h1>Template for Directories in /c-dir</h1>
<ul>

	<li><a href="c.file">c.file</a></li>

	<li><a href="d.txt">d.txt</a></li>

	<li><a href="dir.tmpl">dir.tmpl</a></li>

</ul>
</body></html>
`,
			expectedCode:        200,
			expectedContentType: "text/html",
			DirTemplateFileName: "dir.tmpl",
		},
		// 21. directory that is ignored
		{
			RunTest:             true,
			url:                 "/d-dir/",
			expectedCode:        404,
			expectedContentType: "text/html",
			DirTemplateFileName: "dir.tmpl",
			IgnoreDirs:          "/d-dir/",
		},
		// SetThemeCookie   string
		// SetUserCookie    string
		// 22. test file returned
		{
			RunTest:             true,
			url:                 "/t22-a.html",
			expectedBody:        `body t22-a username=[0000] theme=[pink]` + "\n",
			expectedCode:        200,
			expectedContentType: "text/html",
			SetThemeCookie:      "pink",
			SetUserCookie:       "0000",
			CheckResolvedFn:     "/testdata/theme/0000/pink/t22-a.html",
			CheckDependentFNs:   []string{"/testdata/theme/0000/pink/t22-a.html"},
		},
		// 23.
		{
			RunTest:             true,
			url:                 "/t22-a.html",
			expectedBody:        `body t22-a username=[0000] theme=[gold]` + "\n",
			expectedCode:        200,
			expectedContentType: "text/html",
			SetThemeCookie:      "gold",
			SetUserCookie:       "0000",
			CheckResolvedFn:     "/testdata/theme/0000/gold/t22-a.html",
			CheckDependentFNs:   []string{"/testdata/theme/0000/gold/t22-a.html"},
		},
		// 24.
		{
			RunTest:             true,
			url:                 "/t22-a.html",
			expectedBody:        `body t22-a username - not specified - theme=[pink]` + "\n",
			expectedCode:        200,
			expectedContentType: "text/html",
			SetThemeCookie:      "pink",
			CheckResolvedFn:     "/testdata/theme/pink/t22-a.html",
			CheckDependentFNs:   []string{"/testdata/theme/pink/t22-a.html"},
		},
		// 25.
		{
			RunTest:             true,
			TestComment:         "run you silly boy",
			url:                 "/t22-a.html",
			expectedBody:        `body t22-a username - not specified - theme - not specified -` + "\n",
			expectedCode:        200,
			expectedContentType: "text/html",
			CheckResolvedFn:     "/testdata/t22-a.html",
			CheckDependentFNs:   []string{"/testdata/t22-a.html"},
		},
		// 26. directory with custom template
		{
			RunTest:     true,
			TestComment: "Directory with custom template",
			url:         "/b-dir/",
			expectedBody: `
<html><body>
<h1>Template for Directories at top level ; 0000 ; pink </h1>
<ul>

	<li><a href="Go-FTL-server-fileserve-testdata-theme-0000-pink-b-dir">Go-FTL-server-fileserve-testdata-theme-0000-pink-b-dir</a></li>

	<li><a href="a.data">a.data</a></li>

	<li><a href="b.data">b.data</a></li>

	<li><a href="c.data">c.data</a></li>

	<li><a href="d.data">d.data</a></li>

</ul>
</body></html>
`,
			expectedCode:        200,
			expectedContentType: "text/html",
			DirTemplateFileName: "dir.tmpl",
			SetThemeCookie:      "pink",
			SetUserCookie:       "0000",
			CheckResolvedFn:     "/testdata/theme/0000/pink/b-dir",
			CheckDependentFNs:   []string{"/testdata/theme/0000/pink/b-dir", "testdata/theme/0000/pink/dir.tmpl"},
		},
		// 27. directory with custom template - in a sub-directory
		{
			RunTest:     true,
			TestComment: "Directory with custom template",
			url:         "/c-dir/",
			expectedBody: `
<html><body>
<h1>Template for Directories in /c-dir ; 0000 ; pink </h1>
<ul>

	<li><a href="dir.tmpl">dir.tmpl</a></li>

	<li><a href="theme-0000-pink-c-dir">theme-0000-pink-c-dir</a></li>

</ul>
</body></html>
`,
			expectedCode:        200,
			expectedContentType: "text/html",
			DirTemplateFileName: "dir.tmpl",
			SetThemeCookie:      "pink",
			SetUserCookie:       "0000",
			CheckResolvedFn:     "/testdata/theme/0000/pink/c-dir",
			CheckDependentFNs:   []string{"/testdata/theme/0000/pink/c-dir", "testdata/theme/0000/pink/c-dir/dir.tmpl"},
		},
		// 28. directory with custom template -- verifies that non-themed dir.tmpl will still find in original directory
		{
			RunTest:     true,
			TestComment: "Directory with custom template - check non-themed",
			url:         "/b-dir/",
			expectedBody: `
<html><body>
<h1>Template for Directories at top level</h1>
<ul>

	<li><a href="b-1.txt">b-1.txt</a></li>

	<li><a href="b-2.txt">b-2.txt</a></li>

	<li><a href="b-3.txt">b-3.txt</a></li>

	<li><a href="b-4.txt">b-4.txt</a></li>

</ul>
</body></html>
`,
			expectedCode:        200,
			expectedContentType: "text/html",
			DirTemplateFileName: "dir.tmpl",
			SetThemeCookie:      "pink",
			CheckResolvedFn:     "/testdata/b-dir",
			CheckDependentFNs:   []string{"/testdata/b-dir", "testdata/dir.tmpl"},
		},
		// 29. directory with custom template - in a sub-directory -- no theme, no user
		{
			RunTest:     true,
			TestComment: "Directory with custom template - sub-dir",
			url:         "/c-dir/",
			expectedBody: `
<html><body>
<h1>Template for Directories in /c-dir</h1>
<ul>

	<li><a href="c.file">c.file</a></li>

	<li><a href="d.txt">d.txt</a></li>

	<li><a href="dir.tmpl">dir.tmpl</a></li>

</ul>
</body></html>
`,
			expectedCode:        200,
			expectedContentType: "text/html",
			DirTemplateFileName: "dir.tmpl",
			CheckResolvedFn:     "/testdata/c-dir",
			CheckDependentFNs:   []string{"/testdata/c-dir", "testdata/c-dir/dir.tmpl"},
		},
		// 30. index.html in sub-directory
		{
			RunTest:             true,
			TestComment:         "index.html in sub-dir",
			url:                 "/subdir/",
			expectedBody:        "index.html -subdir- says hello\n",
			expectedCode:        200,
			expectedContentType: "text/html",
		},
	}

	setup1_test(t)
	defer afterTest(t)

	bot := mid.NewConstHandler(`{"bottum":"reached"}`, "Content-Type", "application/json")
	ts := NewFileServer(bot, nil, []string{"/"}, []string{"index.html", "index.htm"}, []string{"./testdata"})
	ts.ThemeRoot = "./testdata/theme/"
	//	ts.ThemeCookieName: "theme",
	//	ts.UserCookieName:  "username",

	for ii, test := range tests {

		if test.RunTest {

			fmt.Printf("%sRunning: %d --- %s%s%s\n", MiscLib.ColorGreen, ii, test.Msg, test.TestComment, MiscLib.ColorReset)
			isOk := true

			tmpErrLogged := nErrLogged
			tmpBuildSuccessLogged := nBuildSuccessLogged

			if test.preFx != nil {
				test.preFx(&test)
			}

			if test.DbTest {
				fmt.Printf("%sAT:%s%s\n", MiscLib.ColorYellow, godebug.LF(), MiscLib.ColorReset)
			}

			rec := httptest.NewRecorder()

			wr := goftlmux.NewMidBuffer(rec, nil)
			if test.DirTemplateFileName != "" {
				wr.DirTemplateFileName = test.DirTemplateFileName
				if db5 {
					fmt.Printf("Using custom template for directories: %s\n", wr.DirTemplateFileName)
				}
			}
			if test.IgnoreDirs != "" {
				wr.IgnoreDirs = append(wr.IgnoreDirs, test.IgnoreDirs)
				if db5 {
					fmt.Printf("Using IgnoreDirs for directories: %s\n", wr.IgnoreDirs)
				}
			}

			var req *http.Request

			if test.DbTest {
				fmt.Printf("%sAT:%s%s\n", MiscLib.ColorYellow, godebug.LF(), MiscLib.ColorReset)
			}

			req, err = http.NewRequest("GET", test.url, nil)
			if err != nil {
				t.Fatalf("Test %d: Could not create HTTP request: %v", ii, err)
				isOk = false
			}
			lib.SetupTestMimicReq(req, "example.com")

			if test.DbTest {
				fmt.Printf("%sAT:%s%s\n", MiscLib.ColorYellow, godebug.LF(), MiscLib.ColorReset)
			}

			if test.SetThemeCookie != "" {
				SetCookieValue(req, ts.ThemeCookieName, test.SetThemeCookie)
			}

			if test.SetUserCookie != "" {
				SetCookieValue(req, ts.UserCookieName, test.SetUserCookie)
			}

			if db7 {
				fmt.Printf("Req: %s\n", lib.SVarI(req))
			}

			if test.DbTest {
				fmt.Printf("%sAT:%s%s\n", MiscLib.ColorYellow, godebug.LF(), MiscLib.ColorReset)
			}

			ts.ServeFile(wr, req, test.url)

			if test.DbTest {
				fmt.Printf("%sAT:%s%s\n", MiscLib.ColorYellow, godebug.LF(), MiscLib.ColorReset)
			}

			wr.FinalFlush()

			if test.DbTest {
				fmt.Printf("%sAT:%s%s\n", MiscLib.ColorYellow, godebug.LF(), MiscLib.ColorReset)
			}

			b := string(rec.Body.Bytes())

			if test.DbTest {
				fmt.Printf("%sAT:%s%s\n", MiscLib.ColorYellow, godebug.LF(), MiscLib.ColorReset)
			}

			// fmt.Printf("wr.Header=%s\n, %s", lib.SVarI(wr.Header()), godebug.LF())

			if wr.StatusCode != test.expectedCode {
				t.Errorf("test %d got status %d; expected %d.\n", ii, wr.StatusCode, test.expectedCode)
				isOk = false
			}
			if wr.StatusCode == 200 && b != test.expectedBody {
				// fmt.Printf("--->>>%s<<<---\n", b)
				// fmt.Printf("--->>>%s<<<---\n", test.expectedBody)
				fmt.Printf("%sDiff:%s%s\n", MiscLib.ColorRed, pretty.Diff(b, test.expectedBody), MiscLib.ColorReset)
				t.Errorf("test %d got --->>>%s<<<--- expected --->>>%s<<<---\n", ii, b, test.expectedBody)
				isOk = false
			}
			if wr.StatusCode == 200 {
				ct := wr.Header().Get("Content-Type")
				if !strings.HasPrefix(ct, test.expectedContentType) {
					t.Errorf("test %d got Content-Type [%s] expected mime tyep of [%s;...]\n", ii, ct, test.expectedContentType)
					isOk = false
				}
			}

			if false {
				// xyzzyNotFound - change http.NotFound() to fileserve.NotFound() so can increment marker and check it as this point
				if wr.StatusCode >= 400 {
					fmt.Printf("%d %d\n", tmpErrLogged, nErrLogged)
					if tmpErrLogged >= nErrLogged {
						t.Errorf("test %d failed to log error - when error occured\n", ii)
					}
				}
			}

			if test.CheckSuccessfulBuild {
				if tmpBuildSuccessLogged >= nBuildSuccessLogged {
					t.Errorf("test %d failed successfuly build\n", ii)
					isOk = false
				}
			}
			if test.CheckNoBuildOccured {
				if tmpBuildSuccessLogged != nBuildSuccessLogged {
					t.Errorf("test %d failed should not have built. Build occured.\n", ii)
					isOk = false
				}
			}

			// Check out caching returned information
			// rw.ResolvedFn = name
			// rw.DependentFNs = append(rw.DependentFNs, name, templateFileName)
			// xyzzy - Add test that checks from .../testdata/ down that this is correct for 0, 1, 5, 7, 8, 9, 19, 22, 23, 24, 25
			if wr.StatusCode == 200 {
				if dbB {
					fmt.Printf("%sCache Info: ResolvedFn: [%s], DependentFNs: %s%s\n", MiscLib.ColorYellow, wr.ResolvedFn, lib.SVar(wr.DependentFNs), MiscLib.ColorReset)
				}
				if test.CheckResolvedFn != "" {
					if !CheckPathContains([]string{wr.ResolvedFn}, []string{test.CheckResolvedFn}) {
						t.Errorf("test %d failed wr.ResolvedFn should have %s in it, == [%s] instead\n", ii, test.CheckResolvedFn, wr.ResolvedFn)
						isOk = false
					}
				}
				if len(test.CheckDependentFNs) > 0 {
					if !CheckPathContains(wr.DependentFNs, test.CheckDependentFNs) {
						t.Errorf("test %d failed wr.DependentFNs should have %s in it, == [%s] instead\n", ii, test.CheckDependentFNs, wr.DependentFNs)
						isOk = false
					}
				}
			}

			if test.postFx != nil {
				test.postFx(&test)
			}

			if !isOk {
				fmt.Printf("    %sTest %d failed%s\n", MiscLib.ColorRed, ii, MiscLib.ColorReset)
			}
		}
	}
}

const dbB = false

/* vim: set noai ts=4 sw=4: */
