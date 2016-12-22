//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1284
//

package RejectHotlink

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/MiscLib"
)

// -----------------------------------------------------------------------------------------------------------------------------------------------
func Test_RejectPathServer01(t *testing.T) {
	tests := []struct {
		url          string
		expectedCode int
		referer      string
	}{
		// 0
		{
			url:          "http://example.com/testdir/b.html", // From ./www/testdir/b.html - using simple file server that just reads the file
			expectedCode: http.StatusOK,
			referer:      "www.bob.com",
		},
		// 1
		{
			url:          "http://example.com/testdir/a.html", // From ./www/testdir/b.html - using simple file server that just reads the file
			expectedCode: http.StatusOK,
			referer:      "",
		},
		// 2
		{
			url:          "http://example.com/testdir/js/ex.js",
			expectedCode: http.StatusOK,
			referer:      "example.com",
		},
		// 3
		{
			url:          "http://example.com/testdir/js/ex.js",
			expectedCode: http.StatusNotFound,
			referer:      "www.bad.com",
		},
		// 4
		{
			url:          "http://example.com/testdir/js/ex.js",
			expectedCode: http.StatusOK,
			referer:      "",
		},
	}

	// bot := NewBotHandler()
	// bot := mid.NewServer()
	// bot := NewSimpleFileServer(nil, []string{"/"}, r []string, m []string)
	bot := mid.NewSimpleFileServer(nil, nil, nil, nil)
	// ms := NewHeaderServer(bot, []string{"/foo"}, "X-Test2", "A-Value2")
	ms := NewRejectHotLinkServer(bot, []string{"/testdir/js"}, []string{"www.example.com", "example.com"}, []string{".jpg", ".css", ".js"})
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		if dbA {
			fmt.Printf("\nTest %d -------------------------------------------------------------------------- \n", ii)
		}

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec, nil)

		var req *http.Request

		req, err = http.NewRequest("GET", test.url, nil)
		if err != nil {
			t.Fatalf("Test %d: Could not create HTTP request: %v", ii, err)
		}
		lib.SetupTestMimicReq(req, "example.com")
		if db6 {
			fmt.Printf("{\"req\":%s,\n\"wr\":%s}\n", lib.SVarI(req), lib.SVarI(wr))
		}

		// referer := req.Header.Get("Referer")
		if test.referer != "" {
			req.Header.Set("Referer", test.referer)
		}

		ms.ServeHTTP(wr, req)

		// Tests to perform on final recorder data.
		if wr.StatusCode != test.expectedCode {
			t.Errorf("Test %2d, reject error got: %d, expected %d\n", ii, wr.StatusCode, test.expectedCode)
		} else {
			if dbA {
				fmt.Printf("%sTest %d: Expted %d code returned - that's good!%s\n", MiscLib.ColorGreen, ii, wr.StatusCode, MiscLib.ColorReset)
			}
		}

	}

}

const dbA = false
const db6 = false

/* vim: set noai ts=4 sw=4: */
