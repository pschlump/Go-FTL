//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1231
//

package Cookie

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

func Test_HeaderServer(t *testing.T) {

	if !cfg.SetupRedisForTest("../test_redis.json") {
		return
	}

	tests := []struct {
		url               string
		hdr               []lib.NameValue
		expectedCode      int
		expectedBody      string
		expectedNHdr      int
		expectedHdr       []lib.NameValue
		expectedCookieSet string
	}{
		{
			"http://example.com/foo?abc=def",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			http.StatusOK,
			"Hello World  *four*  CallNo:1",
			2,
			[]lib.NameValue{lib.NameValue{Name: "X-Test2", Value: "A-Value2"}},
			"X-Test2=A-Value2",
		},
		{
			"http://example.com/def",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			http.StatusOK,
			"Hello World  *four*  CallNo:2",
			1,
			[]lib.NameValue{},
			"",
		},
	}
	// bot := NewBotHandler()
	bot := mid.NewServer()
	ms := NewCookieServer(bot, []string{"/foo"}, "X-Test2", "A-Value2")
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec, nil) // var wr http.ResponseWriter
		// lib.SetupTestCreateHeaders(wr, test.hdr)

		id := "test-01-HeaderServer"
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
		if db1 {
			fmt.Printf("{\"req\":%s,\n\"wr\":%s}\n", lib.SVarI(req), lib.SVarI(wr))
		}
		lib.SetupRequestHeaders(req, test.hdr)

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
		c := wr.Header().Get("Set-Cookie")
		// fmt.Printf("c= >%s<\n", c)
		if c != test.expectedCookieSet {
			t.Errorf("Error %d, Invalid set cookie, got >%s< expected >%s<\n", ii, c, test.expectedCookieSet)
		}

	}

}

const db1 = false

/* vim: set noai ts=4 sw=4: */
