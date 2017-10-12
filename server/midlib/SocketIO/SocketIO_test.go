//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1255
//

package SocketIO

import (
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
func Test_JsonPPathServer(t *testing.T) {

	if !cfg.SetupRedisForTest("../test_redis.json") {
		return
	}

	tests := []struct {
		url          string
		expectedBody string
	}{
		{
			"http://example.com/img/foo.jpg",
			`{"abc":"def"}`,
		},
		{
			"http://example.com/api/status",
			`{"abc":"def"}`,
		},
		{
			"http://example.com/api/status?callback=j1232131231",
			`j1232131231({"abc":"def"});`,
		},
	}
	// ct := h.Get("Content-Type")
	// if rw.StatusCode == http.StatusOK && strings.HasPrefix(ct, "application/json") {
	bot := mid.NewConstHandler(`{"abc":"def"}`, "Content-Type", "application/json")
	ms := NewJSONPServer(bot, []string{"/api/status"}, `^[a-zA-Z\$_][a-zA-Z0-9\$_]*$`)
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

		ms.ServeHTTP(wr, req)
		wr.FinalFlush()

		if false {
			b := string(rec.Body.Bytes())
			if b != test.expectedBody {
				t.Errorf("Error %2d, reject error got: %s, expected %s\n", ii, b, test.expectedBody)
			}
		}

	}

}

/* vim: set noai ts=4 sw=4: */
