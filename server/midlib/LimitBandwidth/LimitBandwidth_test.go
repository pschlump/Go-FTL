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

/*
Test
	0. Setup config in redis for an "auth_token"
	1. Run a bunch in a window to verify that the limit - delay 1500ms is enforced
	1. Check in redis for keys
	2. Run a bunc in a window to verify that the N-Per-Sec limit is enforced
	2. Check in redis for keys
	3. Sleep for X (2sec)
	4. Verify (1) is not-enforced
	4. Verify (2) is reset for a new per-sec amount
	5. Check in redis for keys
	6. Verify config in redis
*/

package LimitBandwidth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/tr"
)

// -----------------------------------------------------------------------------------------------------------------------------------------------
func Test_LimitBandwidthServer(t *testing.T) {

	if !cfg.SetupRedisForTest("../test_redis.json") {
		return
	}
	//if !cfg.SetupPgSqlForTest("../test_pgsql.json") {
	//	return
	//}
	//return

	tests := []struct {
		cmd          string
		url          string
		expectedCode int
		nRun         int
		nMilliSec    int
	}{
		{
			cmd:          "run",
			url:          "http://example.com/cfg/foo.cfg",
			expectedCode: http.StatusOK, // expectedCode: http.StatusNotFound,
			nRun:         1,
		},
		{
			cmd:          "run",
			url:          "http://example.com/private/foo.php",
			expectedCode: http.StatusOK,
			nRun:         1,
		},
		{
			cmd:          "run",
			url:          "http://example.com/js/foo.js",
			expectedCode: http.StatusOK,
			nRun:         1,
		},
		{
			cmd:          "run",
			url:          "http://example.com/static/foo.html",
			expectedCode: 429,
			nRun:         2,
		},
		{
			cmd:       "sleep",
			nMilliSec: 2000,
		},
		{
			cmd:          "run",
			url:          "http://example.com/js/foo.js",
			expectedCode: http.StatusOK,
			nRun:         1,
		},
		{
			cmd:          "run",
			url:          "http://example.com/js/foo.js",
			expectedCode: 429,
			nRun:         1,
		},
	}

	bot := mid.NewServer()

	ms := NewLimitBandwidthServer(bot, []string{"/static", "/js"}, 1500, -1)
	ms.gCfg = cfg.ServerGlobal
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec, nil)

		id := "test-01-LimitBandwidth"
		trx := tr.NewTrx(cfg.ServerGlobal.RedisPool)
		trx.TrxIdSeen(id, test.url, "GET")
		wr.RequestTrxId = id

		wr.G_Trx = trx

		var req *http.Request

		switch test.cmd {
		case "run":
			for i := 0; i < test.nRun; i++ {
				req, err = http.NewRequest("GET", test.url, nil)
				if err != nil {
					t.Fatalf("Test %d: Could not create HTTP request: %v", ii, err)
				}
				lib.SetupTestMimicReq(req, "example.com")

				ms.ServeHTTP(wr, req)

				if wr.StatusCode != test.expectedCode {
					t.Errorf("Error TestNo:%2d i=%d, reject error got: %d, expected %d\n", ii, i, wr.StatusCode, test.expectedCode)
				}
			}
		case "sleep":
			fmt.Printf("Sleeping for %d millisecond, please be patient\n", test.nMilliSec)
			slowDown := time.Duration(int64(test.nMilliSec)) * time.Millisecond
			time.Sleep(slowDown)
		}

	}

}

// -----------------------------------------------------------------------------------------------------------------------------------------------
func Test_LimitBandwidthServer2(t *testing.T) {

	if !cfg.SetupRedisForTest("../test_redis.json") {
		return
	}

	tests := []struct {
		cmd          string
		url          string
		expectedCode int
		nRun         int
		nSkip        int
		nMilliSec    int
	}{
		{ // test 0
			cmd:          "run",
			url:          "http://example.com/cfg/foo.cfg",
			expectedCode: http.StatusOK, // expectedCode: http.StatusNotFound,
			nRun:         1,
		},
		{ // test 1
			cmd:          "run",
			url:          "http://example.com/private/foo.php",
			expectedCode: http.StatusOK,
			nRun:         1,
		},
		{ // test 2
			cmd:          "run",
			url:          "http://example.com/js/foo.js",
			expectedCode: http.StatusOK,
			nRun:         1,
		},
		{ // test 3
			cmd:          "run",
			url:          "http://example.com/static/foo.html",
			expectedCode: 429,
			nSkip:        1,
			nRun:         4,
		},
		{ // test 4
			cmd:       "sleep",
			nMilliSec: 2000,
		},
		{ // test 5
			cmd:          "run",
			url:          "http://example.com/static/foo.html",
			expectedCode: 429,
			nSkip:        2,
			nRun:         4,
		},
	}

	bot := mid.NewServer()

	ms := NewLimitBandwidthServer(bot, []string{"/static", "/js"}, -1, 2)
	ms.gCfg = cfg.ServerGlobal
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec, nil)

		id := "test-02-LimitBandwidth"
		trx := tr.NewTrx(cfg.ServerGlobal.RedisPool)
		trx.TrxIdSeen(id, test.url, "GET")
		wr.RequestTrxId = id

		wr.G_Trx = trx

		var req *http.Request

		switch test.cmd {
		case "run":
			nSkip := test.nSkip
			for i := 0; i < test.nRun; i++ {
				req, err = http.NewRequest("GET", test.url, nil)
				if err != nil {
					t.Fatalf("Test %d: Could not create HTTP request: %v", ii, err)
				}
				lib.SetupTestMimicReq(req, "example.com")

				ms.ServeHTTP(wr, req)

				if nSkip--; nSkip < 0 {
					if wr.StatusCode != test.expectedCode {
						t.Errorf("Error TestNo:%2d i=%d, reject error got: %d, expected %d\n", ii, i, wr.StatusCode, test.expectedCode)
					}
				}
			}
		case "sleep":
			fmt.Printf("Sleeping for %d millisecond, please be patient\n", test.nMilliSec)
			slowDown := time.Duration(int64(test.nMilliSec)) * time.Millisecond
			time.Sleep(slowDown)
		}

	}

}

/* vim: set noai ts=4 sw=4: */
