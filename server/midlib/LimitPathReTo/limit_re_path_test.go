//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1261
//

package LimitPathReTo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
)

// -----------------------------------------------------------------------------------------------------------------------------------------------
func Test_LimitRePathServer(t *testing.T) {
	tests := []struct {
		url          string
		expectedCode int
	}{
		// 0
		{
			"http://example.com/cfg/foo.cfg",
			http.StatusNotFound,
		},
		// 1
		{
			"http://example.com/private/foo.cfg",
			http.StatusNotFound,
		},
		// 2
		{
			"http://example.com/img/foo.jpg",
			http.StatusNotFound,
		},
		// 3 -- should pass, has 2 char directory name
		{
			"http://example.com/js/foo.js",
			http.StatusOK,
		},
		// 4
		{
			"http://example.com/i/foo.jpg",
			http.StatusNotFound,
		},
	}

	bot := mid.NewServer()

	ms := NewLimitRePathServer(bot, []string{"^/[a-z][a-z]/", ".html$"})
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec, nil)

		var req *http.Request

		req, err = http.NewRequest("GET", test.url, nil)
		if err != nil {
			t.Fatalf("Test %d: Could not create HTTP request: %v", ii, err)
		}
		lib.SetupTestMimicReq(req, "example.com")

		ms.ServeHTTP(wr, req)

		if wr.StatusCode != test.expectedCode {
			t.Errorf("Error %2d, reject error got: %d, expected %d\n", ii, wr.StatusCode, test.expectedCode)
		}

	}

}

/* vim: set noai ts=4 sw=4: */
