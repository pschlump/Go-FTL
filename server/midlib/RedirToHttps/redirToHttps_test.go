//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1272
//

package RedirToHttps

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

func Test_RewriteServer(t *testing.T) {
	tests := []struct {
		url         string
		hdr         []lib.NameValue
		expectedUrl string
	}{
		{
			"http://example.com/foo?abc=def",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			"https://example.com/foo?abc=def",
		},
		{
			"http://example.com/def",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			"https://example.com/def",
		},
	}
	bot := mid.NewServer()
	ms := NewRedirectToHTTPSServer(bot, "")
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec, nil) // var wr http.ResponseWriter
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

		// u := lib.GenURLFromReq(req)
		u := wr.Header().Get("Location")

		// Tests to perform on final recorder data.
		if u != test.expectedUrl {
			t.Errorf("Error %2d, Invalid redirec to https, got: %s, expected %s\n", ii, u, test.expectedUrl)
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
