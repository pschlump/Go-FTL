package test_mid

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"

	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/midlib/rewrite_proxy"
)

func Test_RewriteProxyServer(t *testing.T) {
	tests := []struct {
		url          string          //
		hdr          []lib.NameValue // request headers
		expectedCode int             //
		expectedBody string          //
	}{
		{"http://example.com/api/server_status", []lib.NameValue{lib.NameValue{"Accept-Encoding", "gzip, deflate"}}, http.StatusOK,
			`{"status":"success","URI":"/api/status?id=2&q=/api/server_status","id":"2"}`,
		},
	}

	fs := mid.NewSimpleFileServer(nil, nil, nil, nil)
	ms := rewrite_proxy.NewRewriteProxyServer(fs, []string{"/api/"}, "(/api/server_status)", "/api/status?id=2&q=${1}", "http://localhost:8204/")
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec) // var wr http.ResponseWriter
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
		b := string(rec.Body.Bytes())
		if db9 {
			fmt.Printf("Body >%s<\n", b)
		}
		if b != test.expectedBody {
			ioutil.WriteFile(",,a", []byte(b), 0600)
			// ioutil.WriteFile(",,b", []byte(test.expectedBody), 0600)
			t.Errorf("Error %d, Invalid body, got >%s< expected >%s<\n", ii, b, test.expectedBody)
		}
	}

}

const db9 = false

/* vim: set noai ts=4 sw=4: */
