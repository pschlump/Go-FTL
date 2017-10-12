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

package LimitJson

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
		url            string
		expectedBody   string
		Paths          []string        // Paths that this will work for
		Allowed        []LimitJsonType // Limit to only these json items
		OnErrorDiscard string
	}{
		{
			url:          "http://example.com/api/status",
			expectedBody: `{"abc":"def"}`,
			Allowed: []LimitJsonType{
				{
					Path:         []string{"/api/status"},
					ItemsAllowed: []string{"abc"},
				},
			},
			OnErrorDiscard: "yes",
		},
		{
			url:          "http://example.com/api/status",
			expectedBody: `{"ghi":"na"}`,
			Allowed: []LimitJsonType{
				{
					Path:         []string{"/api/status"},
					ItemsAllowed: []string{"ghi"},
				},
			},
			OnErrorDiscard: "yes",
		},
		{
			url:          "http://example.com/api/status",
			expectedBody: `{"ghi":"na"}`,
			Allowed: []LimitJsonType{
				{
					Path:         []string{"/api/status"},
					ItemsRemoved: []string{"abc"},
				},
			},
			OnErrorDiscard: "yes",
		},
	}
	bot := mid.NewConstHandler(`{"abc":"def","ghi":"na"}`, "Content-Type", "application/json")
	ms := NewLimitJsonServer(bot, []string{"/api/status"})
	// func NewLimitJsonServer(n http.Handler, p []string) *LimitJsonHandlerType {
	var err error
	lib.SetupTestCreateDirs()

	/*

	   type LimitJsonType struct {
	   	Path         []string
	   	ItemsAllowed []string
	   	ItemsRemoved []string
	   }

	   type LimitJsonHandlerType struct {
	   	Next           http.Handler    //
	   	Paths          []string        // Paths that this will work for
	   	Allowed        []LimitJsonType // Limit to only these json items
	   	OnErrorDiscard string          //
	   	LineNo         int             //
	   }

	*/

	for ii, test := range tests {

		ms.Allowed = test.Allowed
		ms.OnErrorDiscard = test.OnErrorDiscard

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

		b := string(rec.Body.Bytes())
		if b != test.expectedBody {
			t.Errorf("Test No: %2d, reject error got: --->>>%s<<<---, expected --->>>%s<<<---\n", ii, b, test.expectedBody)
		}

	}

}

/* vim: set noai ts=4 sw=4: */
