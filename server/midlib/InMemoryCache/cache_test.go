package InMemoryCache

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/tr"
)

/*

1. Use a fake file-server
	1. File server can return - based on a closure

*/

type AHeader struct {
	Name  string
	Value string
}

func Test_Cahce_01(t *testing.T) {

	fmt.Printf("\n\n\n - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -  Test_Cache_01 - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -\n\n\n")

	var err error

	tests := []struct {
		RunTest      bool
		url          string
		rootForTest  string // ./testdata/t00, .../t01 etc.
		body         string
		mimeType     string
		hdrs         []AHeader
		status       int
		ResolvedFn   string
		DependentFNs []string
		FileSource   string // local, proxy
		preFx        func()
		postFx       func()
		createRespFx func()
	}{
		// 0. test file returned
		{
			RunTest:     true,
			url:         "/td1.hmtl",
			rootForTest: "./testdata/t00",
			body:        "Yep - body of td1.data\n",
			mimeType:    "text/html",
			hdrs: []AHeader{
				AHeader{Name: "Etag", Value: "bob"},
				AHeader{Name: "Etag", Value: "bob"}, // caching heder
			},
			status:       200,
			ResolvedFn:   "./testdata/t00/td1.html",           // must convert to hard path
			DependentFNs: []string{"./testdata/t00/td1.html"}, // must convert to hard path
			FileSource:   "local",
		},
	}

	_ = tests
	_ = err

}

// Also tess SimpleFile!

func Test_Cache_02(t *testing.T) {

	fmt.Printf("\n\n\n - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -  Test_Cache_02 - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -\n\n\n")

	cfg.SetupRedisForTest("../test_redis.json")
	os.Mkdir("./cache", 0700)

	tests := []struct {
		url          string          //
		hdr          []lib.NameValue // request headers
		expectedCode int             //
		//expectedBody string        // Note body is base64 encoded - so can store binary in this file then check it at bottom of test
		//expectedETag string        // Can be derived via CLI tool sha256 hash
	}{
		{
			url:          "http://example.com/testdir/b.html", // From ./www/testdir/b.html - using simple file server that just reads the file
			hdr:          []lib.NameValue{lib.NameValue{Name: "Accept-Encoding", Value: "gzip, deflate"}},
			expectedCode: http.StatusOK,
			//expectedBody: "H4sIAAAJbogA/+zOIREAIBQFME8KEhAKhQBH/yPG4/5Nza7Nse7Z/T/MzMziCTMzMzMzMzMzMzMzMzMzMzMzs3zCzMzMzMzMzMzMzMzMzMzMrNbsAQAA//8=",
			//expectedETag: "22d5274857e60f69604450dff82675e1919c0210fca92b89e4bcb77aba82dbf5",
		},
		{
			url:          "http://example.com/testdir/a.html", // From ./www/testdir/b.html - using simple file server that just reads the file
			hdr:          []lib.NameValue{lib.NameValue{Name: "Accept-Encoding", Value: "gzip, deflate"}},
			expectedCode: http.StatusOK,
			//expectedBody: "IGEuaHRtbAo=", // base 64 encoded body - so can have it in this file easily for binary.
			//expectedETag: "",
		},
	}

	fs := mid.NewSimpleFileServer(nil, nil, nil, nil)
	// ms := NewGzipServer(fs, []string{"/"}, 100)
	// func NewInMemoryCacheServer(n http.Handler, p []string, e []string, d int, sl int) *InMemoryCache {
	ms := NewInMemoryCacheServer(fs, []string{"/"}, []string{".html", ".js", ".css"}, 60, 1024*1204)
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec, nil) // var wr http.ResponseWriter

		id := "test-01-StatusHandler"
		trx := tr.NewTrx(cfg.ServerGlobal.RedisPool)
		trx.TrxIdSeen(id, test.url, "GET")
		wr.RequestTrxId = id

		wr.G_Trx = trx

		// lib.SetupTestCreateHeaders(wr, test.hdr)

		var req *http.Request

		req, err = http.NewRequest("GET", test.url, nil)
		if err != nil {
			t.Fatalf("Test %d: Could not create HTTP request: %v", ii, err)
		}
		lib.SetupTestMimicReq(req, "example.com")
		lib.SetupRequestHeaders(req, test.hdr)

		if db9 {
			fmt.Printf("{\"req\":%s,\n\"wr\":%s}\n", lib.SVarI(req), lib.SVarI(wr))
		}

		ms.ServeHTTP(wr, req)

		// Tests to perform on wr - first
		if wr.Error != nil {
			t.Errorf("Error %2d, Invalid error : %s\n", ii, wr.Error)
		}

		wr.FinalFlush()

		// Tests to perform on final recorder data.
		if rec.Code != test.expectedCode {
			t.Errorf("Error %2d, Invalid status code: %d, expected %d\n", ii, wr.StatusCode, test.expectedCode)
		}

		//		b := string(rec.Body.Bytes())
		//		bb := base64.StdEncoding.EncodeToString([]byte(b))
		//		if db9 {
		//			fmt.Printf("Body (base64 encoded) >%s<\n", bb)
		//		}
		//		if bb != test.expectedBody {
		//			// ioutil.WriteFile(",,a", []byte(bb), 0600)
		//			// ioutil.WriteFile(",,b", []byte(test.expectedBody), 0600)
		//			t.Errorf("Error %d, Invalid body, got >%s< expected >%s<\n", ii, bb, test.expectedBody)
		//		}
		//		if test.expectedETag != "" {
		//			etag := wr.Header().Get("Etag")
		//			if test.expectedETag != etag {
		//				t.Errorf("Error %d, Invalid etag, got >%s< expected >%s<\n", ii, etag, test.expectedETag)
		//			}
		//		}
	}

}

const db9 = false
