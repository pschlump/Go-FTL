//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1263
//

package LimitPathTo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
)

// -----------------------------------------------------------------------------------------------------------------------------------------------
func Test_LimitPathServer(t *testing.T) {
	tests := []struct {
		url          string
		expectedCode int
	}{
		{
			"http://example.com/cfg/foo.cfg",
			http.StatusNotFound,
		},
		{
			"http://example.com/private/foo.php",
			http.StatusNotFound,
		},
		{
			"http://example.com/js/foo.js",
			http.StatusOK,
		},
		{
			"http://example.com/static/foo.html",
			http.StatusOK,
		},
	}

	bot := mid.NewServer()

	ms := NewLimitPathServer(bot, []string{"/static", "/js"})
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec,nil)

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
