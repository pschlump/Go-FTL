//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1235
//

package Latency

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
)

// -----------------------------------------------------------------------------------------------------------------------------------------------
func Test_LatencyServer(t *testing.T) {
	tests := []struct {
		url          string
		expectedCode int
		expectSlow   bool
	}{
		{
			"http://example.com/foo?abc=def",
			http.StatusOK,
			false,
		},
		{
			"http://example.com/index.html",
			http.StatusOK,
			true,
		},
	}

	bot := mid.NewServer()
	var err error
	lib.SetupTestCreateDirs()
	ms := NewLatencyServer(bot, []string{"/index.html"}, 2000)

	fmt.Printf("Expect test to take serveral seconds\n")

	for ii, test := range tests {

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec, nil)

		var req *http.Request

		req, err = http.NewRequest("GET", test.url, nil)
		if err != nil {
			t.Fatalf("Test %d: Could not create HTTP request: %v", ii, err)
		}
		lib.SetupTestMimicReq(req, "example.com")

		if db4 {
			fmt.Printf("{\"req\":%s,\n\"wr\":%s}\n", lib.SVarI(req), lib.SVarI(wr))
		}

		start := time.Now()

		if db4 {
			fmt.Fprintf(os.Stderr, "bef: %s\n", test.url)
		}
		ms.ServeHTTP(wr, req)
		if db4 {
			fmt.Fprintf(os.Stderr, "aft: %s\n", test.url)
		}

		elapsed := time.Since(start)

		if test.expectSlow {
			if elapsed < ((2000 - 1) * time.Millisecond) {
				t.Errorf("Error %2d, did not slow request down got: %.4f seconds, expected %.4f\n", ii, elapsed.Seconds(), (2000 * time.Millisecond).Seconds())
			}
		}

		// Tests to perform on final recorder data.
		if wr.StatusCode != test.expectedCode {
			t.Errorf("Error %2d, reject error got: %d, expected %d\n", ii, wr.StatusCode, test.expectedCode)
		}

		// xyzzy - test the output in ./test.out.out

	}

}

const db4 = false

/* vim: set noai ts=4 sw=4: */
