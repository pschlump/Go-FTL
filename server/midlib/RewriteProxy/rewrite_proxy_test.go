//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1290
//

package RewriteProxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
)

// -----------------------------------------------------------------------------------------------------------------------------------------------

func Test_TopServer(t *testing.T) {
	tests := []struct {
		url          string
		hdr          []lib.NameValue
		expectedCode int
		expectedBody string
	}{
		{
			url:          "http://example.com/foo?abc=def",
			hdr:          []lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			expectedCoce: http.StatusOK,
			expectedBody: "",
		},
	}
	// {"http://example.com/foo?abc=def", []lib.NameValue{lib.NameValue{"X-Test", "A-Value"}}, http.StatusOK, "Hello World  *four*  CallNo:1"},
	ms := mid.NewServer()
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec, nil) // var wr http.ResponseWriter
		lib.SetupTestCreateHeaders(wr, test.hdr)

		var req *http.Request

		req, err = http.NewRequest("GET", test.url, nil)
		if err != nil {
			t.Fatalf("Test %d: Could not create HTTP request: %v", ii, err)
		}
		lib.SetupTestMimicReq(req, "example.com")
		if db1 {
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
		if b != test.expectedBody {
			t.Errorf("Error %d, Invalid body, got >%s< expected >%s<\n", ii, b, test.expectedBody)
		}
	}

}

const db1 = false

/* vim: set noai ts=4 sw=4: */
