//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1288
//

package Rewrite

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
)

// -----------------------------------------------------------------------------------------------------------------------------------------------
// test that redirect will tranform ULR
//
// 1. req.URL - modified
// 2. req.RequestURI - modified

func Test_RewriteServer(t *testing.T) {
	tests := []struct {
		url              string
		hdr              []lib.NameValue
		expectedUrl      string
		expectedRerun    bool
		expectedNRewrite int
	}{
		{
			"http://example.com/foo?abc=def",
			[]lib.NameValue{lib.NameValue{"X-Test", "A-Value"}},
			"http://localhost:8204/api/process?path=foo&name=example.com&abc=def",
			true,
			1,
		},
		{
			"http://example.com/def",
			[]lib.NameValue{lib.NameValue{"X-Test", "A-Value"}},
			"http://example.com/def",
			false,
			0,
		},
	}
	// bot := NewBotHandler()
	bot := mid.NewServer()
	// ms := NewHeaderServer(bot, []string{"/foo"}, "X-Test2", "A-Value2")
	ms := NewRewriteServer(bot, []string{"/foo"}, "http://(example.com)/(.*)\\?(.*)", "http://localhost:8204/api/process?path=${2}&name=${1}&${3}")
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec,nil) // var wr http.ResponseWriter
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

		// Tests to perform on final recorder data.
		if u != test.expectedUrl {
			t.Errorf("Error %2d, Invalid redirect: %s, expected %s\n", ii, u, test.expectedUrl)
		}
		if wr.RerunRequest != test.expectedRerun {
			t.Errorf("Error %2d, Invalid rerun request: %v, expected %v\n", ii, wr.RerunRequest, test.expectedRerun)
		}
		if wr.NRewrite != test.expectedNRewrite {
			t.Errorf("Error %2d, Invalid NRewrite request: %d, expected %d\n", ii, wr.NRewrite, test.expectedNRewrite)
		}
		if !(wr.StatusCode == 0 || wr.StatusCode == 200) {
			t.Errorf("Error %2d, Invalid StatusCode : %d\n", ii, wr.StatusCode)
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
