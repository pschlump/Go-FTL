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

package ErrorReturn

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
func Test_ErrorReturnServer(t *testing.T) {

	if !cfg.SetupRedisForTest("../test_redis.json") {
		return
	}

	tests := []struct {
		url            string
		expectedStatus int
	}{
		{
			url:            "http://example.com/img/foo.jpg",
			expectedStatus: 200,
		},
		{
			url:            "http://example.com/api/error",
			expectedStatus: 522,
		},
	}
	// ct := h.Get("Content-Type")
	// if rw.StatusCode == http.StatusOK && strings.HasPrefix(ct, "application/json") {
	bot := mid.NewConstHandler(`{"abc":"def"}`, "Content-Type", "application/json")
	ms := NewErrorReturnServer(bot, []string{"/api/error"}, []int{522})
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

		// fmt.Printf("wr=%s\n", godebug.SVarI(wr))

		if wr.StatusCode != test.expectedStatus {
			t.Errorf("Error %2d, reject error got: %d, expected %d\n", ii, wr.StatusCode, test.expectedStatus)
		}

	}

}

/* vim: set noai ts=4 sw=4: */
