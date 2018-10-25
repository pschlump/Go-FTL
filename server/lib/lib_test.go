//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1193
//

package lib

import (
	"net/http"
	"net/url"
	"testing"
)

// -----------------------------------------------------------------------------------------------------------------------------------------------
func Test_GenUrl_1(t *testing.T) {
	tests := []struct {
		req              *http.Request
		expectedUrl      string
		expectedUrlProxy string
	}{
		{
			&http.Request{URL: &url.URL{Path: "/api/process", RawQuery: "path=foo&name=example.com&abc=def"}, Host: "localhost:8204"},
			"http://localhost:8204/api/process?path=foo&name=example.com&abc=def",
			"http://www.zepher.com/api/process?path=foo&name=example.com&abc=def",
		},
	}

	for ii, test := range tests {

		url := GenURLFromReq(test.req)
		if url != test.expectedUrl {
			t.Errorf("Error %2d, Invalid : %s, expected %s\n", ii, url, test.expectedUrl)
		}

		url = GenURLFromReqProxy(test.req, "www.zepher.com")
		if url != test.expectedUrlProxy {
			t.Errorf("Error %2d, Invalid : %s, expected %s\n", ii, url, test.expectedUrl)
		}

		// func GenURL(www http.ResponseWriter, req *http.Request) (url string, hUrl string) {
	}

}

// func MatchURLPrefix(APath, Pattern string) bool {
func Test_MatchURLPrefix(t *testing.T) {

	tests := []struct {
		APath          string
		Pattern        string
		ExpectedReturn bool
	}{
		{
			APath:          "/q/2",
			Pattern:        "/q^",
			ExpectedReturn: true,
		},
		{
			APath:          "/q/2",
			Pattern:        "/q/",
			ExpectedReturn: true,
		},
		{
			APath:          "/q/2",
			Pattern:        "/q",
			ExpectedReturn: true,
		},
		{
			APath:          "/q",
			Pattern:        "/q^",
			ExpectedReturn: true,
		},
		{
			APath:          "/qwerty",
			Pattern:        "/q^",
			ExpectedReturn: false,
		},
		{
			APath:          "/r",
			Pattern:        "/q^",
			ExpectedReturn: false,
		},
		{
			APath:          "/r",
			Pattern:        "/q",
			ExpectedReturn: false,
		},
		{
			APath:          "/r",
			Pattern:        "/q/",
			ExpectedReturn: false,
		},
		{
			APath:          "/r",
			Pattern:        "/rrr/",
			ExpectedReturn: false,
		},
		{
			APath:          "/abc/def",
			Pattern:        "/abcXdef^",
			ExpectedReturn: false,
		},
		{
			APath:          "/abc/def",
			Pattern:        "/abc/def^",
			ExpectedReturn: true,
		},
		{
			APath:          "/abc/def",
			Pattern:        "/abc^def^",
			ExpectedReturn: true,
		},
		{
			APath:          "/abc/def",
			Pattern:        "/abc^",
			ExpectedReturn: true,
		},
		{
			APath:          "/abc/def",
			Pattern:        "/abc^ghi",
			ExpectedReturn: false,
		},
	}

	for ii, test := range tests {
		bb := MatchURLPrefix(test.APath, test.Pattern)
		if bb != test.ExpectedReturn {
			t.Errorf("Error %2d, Expected: %v, Got %v for pattern [%s] url [%s]\n", ii, test.ExpectedReturn, bb, test.Pattern, test.APath)
		}
	}

}

/* vim: set noai ts=4 sw=4: */
