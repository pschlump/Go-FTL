//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1270
//

package Redirect

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
// test that redirect will tranform ULR
//
// 1. req.URL - modified
// 2. req.RequestURI - modified

func Test_RewriteServer(t *testing.T) {

	if !cfg.SetupRedisForTest("../test_redis.json") {
		return
	}

	tests := []struct {
		url                 string
		hdr                 []lib.NameValue
		expectedUrl         string
		expectedHeaderFound bool
		expectedStatus      int
	}{
		{
			"http://example.com/api?abc=def",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			"http://www.test1.com:8000/internal/api?abc=def", // "http://www.test1.com:8000/internal/api{{THE_REST}}", // "http://www.test1.com:8000/internal/api?abc=def",
			true,
			307,
		},
		{
			"http://example.com/def",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			"http://example.com/def",
			false,
			200,
		},
	}
	// bot := NewBotHandler()
	bot := mid.NewServer()
	// ms := NewHeaderServer(bot, []string{"/foo"}, "X-Test2", "A-Value2")
	ms := NewRedirectServer(bot, []string{"/api"}, []string{"http://www.test1.com:8000/internal/api{{.THE_REST}}"})
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec, nil) // var wr http.ResponseWriter

		id := "test-01-StatusHandler"
		trx := tr.NewTrx(cfg.ServerGlobal.RedisPool)
		trx.TrxIdSeen(id, test.url, "GET")
		wr.RequestTrxId = id

		wr.G_Trx = trx

		// lib.SetupTestCreateHeaders(wr, test.hdr)

		var req *http.Request

		req, err = http.NewRequest("GET", test.url, nil)
		if err != nil {
			t.Fatalf("Test %d: Could not create HTTP request: %v", ii, err)
		}
		lib.SetupTestMimicReq(req, "example.com")
		if db3 {
			fmt.Printf("{\"req\":%s,\n\"wr\":%s}\n", lib.SVarI(req), lib.SVarI(wr))
		}
		lib.SetupRequestHeaders(req, test.hdr)

		ms.ServeHTTP(wr, req)

		u := lib.GenURLFromReq(req)

		// xyzzy - template for redirect body -

		if vv, found := lib.HasHeader(wr, "Location"); found {
			if test.expectedHeaderFound == false {
				t.Errorf("Error %2d, Invalid redirect: %s, expected to not redirect - it did\n", ii, u)
			}
			if vv != test.expectedUrl {
				fmt.Printf("%s\n", lib.SVarI(wr))
				t.Errorf("Error %2d, Requst: %s Location: %s Expected: %s\n", ii, u, vv, test.expectedUrl)
			}
		} else {
			if test.expectedHeaderFound == true {
				t.Errorf("Error %2d, Invalid redirect: %s, expected to redirect - it did not\n", ii, u)
			}
		}
		if wr.StatusCode != test.expectedStatus {
			t.Errorf("Error %2d, Invalid redirect: %v, expected %v\n", ii, u, test.expectedStatus)
		}

	}

}

/*
req ->{
	"Method": "GET",
	"URL": {
		"Scheme": "http",
		"Opaque": "",
		"User": null,
		"Host": "localhost:8204",
		"Path": "/api/process",
		"RawQuery": "path=foo\u0026name=example.com\u0026abc=def",
		"Fragment": ""
	},
	"Proto": "HTTP/1.1",
	"ProtoMajor": 1,
	"ProtoMinor": 1,
	"Header": {
		"X-Test": [
			"A-Value"
		]
	},
	"Body": null,
	"ContentLength": 0,
	"TransferEncoding": null,
	"Close": false,
	"Host": "example.com",
	"Form": null,
	"Form": null,
	"PostForm": null,
	"MultipartForm": null,
	"Trailer": null,
	"RemoteAddr": "1.2.2.2:52180",
	"RequestURI": "/api/process?path=foo\u0026name=example.com\u0026abc=def",
	"TLS": null
}<-
*/

const db3 = false

/* vim: set noai ts=4 sw=4: */
