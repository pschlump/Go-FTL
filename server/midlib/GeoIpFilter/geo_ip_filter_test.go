//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1243
//

package GeoIpFilter

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
func Test_GeoIPFilterServer(t *testing.T) {

	if !cfg.SetupRedisForTest("../test_redis.json") {
		return
	}

	tests := []struct {
		url          string
		from_ip      string
		expectedCode int
	}{
		{
			"http://example.com/foo?abc=def",
			"1.1.1.1",
			http.StatusOK,
		},
		{
			"http://example.com/foo?abc=def",
			"1.1.1.2",
			http.StatusOK,
		},
		// xyzzy - reject address - find one
		// xyzzy - allow address - test that
		// xyzzy - test return of not-avail.*html file
	}

	bot := mid.NewServer()
	ms := NewGeoIPFilterServer(bot, []string{"/"}, "./test/GeoLite2-Country.mmdb", "reject", []string{"JP", "VN"}, "not-avaiable-in-your-country.html")
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
		req.RemoteAddr = test.from_ip + ":44444"
		if db4 {
			fmt.Printf("{\"req\":%s,\n\"wr\":%s}\n", lib.SVarI(req), lib.SVarI(wr))
		}

		ms.ServeHTTP(wr, req)

		// Tests to perform on final recorder data.
		if wr.StatusCode != test.expectedCode {
			t.Errorf("Error %2d, reject error got: %d, expected %d\n", ii, wr.StatusCode, test.expectedCode)
		}

	}

}

const db4 = false

/* vim: set noai ts=4 sw=4: */
