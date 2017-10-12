//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1265
//

package Logging

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

// -----------------------------------------------------------------------------------------------------------------------------------------------
func Test_RejectExtServer(t *testing.T) {

	if !cfg.SetupRedisForTest("../test_redis.json") {
		return
	}

	tests := []struct {
		url          string
		expectedCode int
	}{
		{
			"http://example.com/foo.cfg",
			http.StatusOK,
		},
		{
			"http://example.com/foo.php",
			http.StatusOK,
		},
		{
			"http://example.com/foo",
			http.StatusOK,
		},
		{
			"http://example.com/foo.html",
			http.StatusOK,
		},
	}
	// bot := NewBotHandler()
	bot := mid.NewServer()
	// ms := NewHeaderServer(bot, []string{"/foo"}, "X-Test2", "A-Value2")
	ms := NewLoggingServer(bot, []string{"/foo"}, "")
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec, nil)

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
		if db5 {
			fmt.Printf("{\"req\":%s,\n\"wr\":%s}\n", lib.SVarI(req), lib.SVarI(wr))
		}

		ms.ServeHTTP(wr, req)

		// Tests to perform on final recorder data.
		if wr.StatusCode != test.expectedCode {
			t.Errorf("Error %2d, reject error got: %d, expected %d\n", ii, wr.StatusCode, test.expectedCode)
		}

	}

}

const db5 = false

/* vim: set noai ts=4 sw=4: */
