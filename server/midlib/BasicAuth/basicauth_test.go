//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1225
//

package BasicAuth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
)

const dbA = false

// -----------------------------------------------------------------------------------------------------------------------------------------------
func Test_BasicAuthServer(t *testing.T) {
	tests := []struct {
		url           string
		expectedCode  int
		doLogin       bool
		username      string
		password      string
		realm         string
		iAmIn         bool
		hdr           []lib.NameValue
		expectedCode2 int
	}{
		{
			url:           "http://example.com/admin/login.html",
			expectedCode:  http.StatusUnauthorized,
			doLogin:       true,
			username:      "testme",
			password:      "bobbob",
			realm:         "example.com",
			iAmIn:         true,
			hdr:           []lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			expectedCode2: http.StatusOK,
		},
		{
			url:           "http://example.com/admin/login.html",
			expectedCode:  http.StatusUnauthorized,
			doLogin:       true,
			username:      "testme",
			password:      "goofy",
			realm:         "example.com",
			iAmIn:         false,
			hdr:           []lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			expectedCode2: http.StatusUnauthorized,
		},
		{
			url:           "http://example.com/admin/login.html",
			expectedCode:  http.StatusUnauthorized,
			doLogin:       true,
			username:      "nono",
			password:      "bobbob",
			realm:         "example.com",
			iAmIn:         false,
			hdr:           []lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			expectedCode2: http.StatusUnauthorized,
		},
		{
			url:           "http://example.com/admin/login.html",
			expectedCode:  http.StatusUnauthorized,
			doLogin:       true,
			username:      "testme",
			password:      "bobbob",
			realm:         "boo.com",
			iAmIn:         false,
			hdr:           []lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			expectedCode2: http.StatusUnauthorized,
		},
		{
			url:           "http://example.com/admin/login.html",
			expectedCode:  http.StatusUnauthorized,
			doLogin:       true,
			username:      "testme",
			password:      "goofy",
			realm:         "boo.com",
			iAmIn:         false,
			hdr:           []lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			expectedCode2: http.StatusUnauthorized,
		},
		{
			url:           "http://example.com/admin/login.html",
			expectedCode:  http.StatusUnauthorized,
			doLogin:       true,
			username:      "nono",
			password:      "bobbob",
			realm:         "boo.com",
			iAmIn:         false,
			hdr:           []lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			expectedCode2: http.StatusUnauthorized,
		},
		{
			url:           "http://example.com/admin/login.html",
			expectedCode:  http.StatusUnauthorized,
			doLogin:       true,
			username:      "testme",
			password:      "bobbob",
			realm:         "example.com",
			iAmIn:         true,
			hdr:           []lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			expectedCode2: http.StatusOK,
		},
		{
			url:          "http://example.com/private/foo.php",
			expectedCode: http.StatusOK,
		},
		{
			url:          "http://example.com/img/foo.jpg",
			expectedCode: http.StatusOK,
		},
		{
			url:          "http://example.com/js/foo.html",
			expectedCode: http.StatusOK,
		},
	}

	bot := mid.NewServer()

	ms := newBasicAuthServer(bot, []string{"/admin"}, "./cfg/.htaccess", "example.com")
	var err error
	lib.SetupTestCreateDirs()

	if !cfg.SetupRedisForTest("../test_redis.json") {
		return
	}

	for ii, test := range tests {

		rec := httptest.NewRecorder()
		wr := goftlmux.NewMidBuffer(rec, nil)

		var req *http.Request

		req, err = http.NewRequest("GET", test.url, nil)
		if err != nil {
			t.Fatalf("Test %d: Could not create HTTP request: %v", ii, err)
		}

		lib.SetupTestMimicReq(req, "example.com")
		if dbA {
			fmt.Printf("{\"req\":%s,\n\"wr\":%s}\n", lib.SVarI(req), lib.SVarI(wr))
		}

		ms.ServeHTTP(wr, req)

		code := wr.StatusCode
		// Tests to perform on final recorder data.
		if code != test.expectedCode {
			t.Errorf("Error %2d, reject error got: %d, expected %d\n", ii, wr.StatusCode, test.expectedCode)
		}

		// xyzzy - this is the spot to do the login if requested.
		if test.doLogin {
			// expectedCode: http.StatusUnauthorized,
			// if got the auth login code && got header then...
			// Send new request with hdr and see if got login
			h := wr.Header()
			ww := h.Get("WWW-Authenticate") //, "Basic realm=\""+hdlr.Realm+"\"")
			if ww == "" {
				t.Errorf("Error %2d, missing WWW-Authenticae header\n", ii)
			} else {
				wr.FinalFlush()
				if db8 {
					fmt.Printf("*1* >%s<\n", ww)
				}
				realm := strings.TrimPrefix(ww, "Basic realm=\"")
				realm = strings.TrimSuffix(realm, "\"")
				if db8 {
					fmt.Printf("*2*/realm >%s<\n", realm)
				}
				if realm != "example.com" {
					t.Errorf("Error %2d, missing realm\n", ii)
				} else {
					// xyzzy - create header -then send-
					Pw := lib.Md5sum(test.username + ":" + test.realm + ":" + test.password)
					userPassword := base64.StdEncoding.EncodeToString([]byte(test.username + ":" + Pw))
					test.hdr = append(test.hdr, lib.NameValue{Name: "Authorization", Value: "Basic " + userPassword})
					lib.SetupRequestHeaders(req, test.hdr)
					rec1 := httptest.NewRecorder()
					wr1 := goftlmux.NewMidBuffer(rec1, nil)
					ms.ServeHTTP(wr1, req)
					if db8 {
						fmt.Printf("*3* wr1 StatusCode=%d\n", wr1.StatusCode)
					}
					if wr1.StatusCode != test.expectedCode2 {
						t.Errorf("Error %2d, failed to authorize\n", ii)
					}
				}
			}
		}

	}

}

const db6 = false
const db8 = false

/* vim: set noai ts=4 sw=4: */
