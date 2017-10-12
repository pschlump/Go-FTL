//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1247
//

package LoginRequired

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/tr"
)

func Test_LoginRequiredServer_01(t *testing.T) {

	tests := []struct {
		url          string //
		expectedCode int    //
	}{
		{
			url:          "http://example.com/testdir/b.html?$is_logged_in$=y", // From ./www/testdir/b.html - using simple file server that just reads the file
			expectedCode: http.StatusOK,
		},
		{
			url:          "http://example.com/testdir/b.html", // From ./www/testdir/b.html - using simple file server that just reads the file
			expectedCode: http.StatusUnauthorized,
		},
		{
			url:          "http://example.com/index.html",
			expectedCode: http.StatusOK,
		},
	}

	fs := mid.NewSimpleFileServer(nil, nil, nil, nil)
	ms := NewLoginRequiredServer(fs, []string{"/testdir"})
	var err error
	lib.SetupTestCreateDirs()

	//	for ii, test := range tests {
	//
	//		rec := httptest.NewRecorder()
	//
	//		wr := goftlmux.NewMidBuffer(rec, nil) // var wr http.ResponseWriter
	//
	//		var req *http.Request
	//
	//		req, err = http.NewRequest("GET", test.url, nil)
	//		if err != nil {
	//			t.Fatalf("Test %d: Could not create HTTP request: %v", ii, err)
	//		}
	//		lib.SetupTestMimicReq(req, "example.com")
	//
	//		if db9 {
	//			fmt.Printf("{\"req\":%s,\n\"wr\":%s}\n", lib.SVarI(req), lib.SVarI(wr))
	//		}
	//
	//		ms.ServeHTTP(wr, req)
	//
	//		wr.FinalFlush()
	//
	//		if rec.Code != test.expectedCode {
	//			t.Errorf("Error %2d, Invalid status code: %d, expected %d\n", ii, wr.StatusCode, test.expectedCode)
	//		}
	//	}

}

func Test_LoginRequiredServer_02(t *testing.T) {

	if !cfg.SetupRedisForTest("../test_redis.json") {
		return
	}

	tests := []struct {
		url          string //
		expectedCode int    //
	}{
		{
			url:          "http://example.com/testdir/b.html?$is_logged_in$=y&$is_full_login$=y",
			expectedCode: http.StatusOK,
		},
		{
			url:          "http://example.com/testdir/b.html?$is_logged_in$=y",
			expectedCode: http.StatusUnauthorized,
		},
		{
			url:          "http://example.com/testdir/b.html",
			expectedCode: http.StatusUnauthorized,
		},
		{
			url:          "http://example.com/index.html",
			expectedCode: http.StatusOK,
		},
	}

	fs := mid.NewSimpleFileServer(nil, nil, nil, nil)
	ms := NewLoginRequiredServer(fs, []string{"/testdir"})
	ms.StrongLoginReq = "yes"
	ms.strongLoginReq = true
	var err error
	lib.SetupTestCreateDirs()

	return

	for ii, test := range tests {

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec, nil) // var wr http.ResponseWriter

		id := "test-01-StatusHandler"
		trx := tr.NewTrx(cfg.ServerGlobal.RedisPool)
		trx.TrxIdSeen(id, test.url, "GET")
		wr.RequestTrxId = id

		wr.G_Trx = trx

		var req *http.Request

		req, err = http.NewRequest("GET", test.url, nil)
		if err != nil {
			t.Fatalf("Test %d: Could not create HTTP request: %v", ii, err)
		}
		lib.SetupTestMimicReq(req, "example.com")

		if db9 {
			fmt.Printf("{\"req\":%s,\n\"wr\":%s}\n", lib.SVarI(req), lib.SVarI(wr))
		}

		ms.ServeHTTP(wr, req)

		wr.FinalFlush()

		if rec.Code != test.expectedCode {
			t.Errorf("Error %2d, Invalid status code: %d, expected %d\n", ii, wr.StatusCode, test.expectedCode)
		}
	}

}

const db9 = false

/* vim: set noai ts=4 sw=4: */
