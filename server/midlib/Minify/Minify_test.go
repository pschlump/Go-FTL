//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1268
//

package Minify

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
		url          string
		hdr          []lib.NameValue
		expectedBody string
	}{
		{
			url: "http://example.com/testdir/js/ex2.js",
			hdr: []lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			expectedBody: `function createCookie(name,value,days){var expires="";if(days){var date=new Date();date.setTime(date.getTime()+(days*24*60*60*1000));expires="; expires="+date.toGMTString();}
document.cookie=name+"="+value+expires+"; path=/";}
function getCookie(name){var nameEQ=name+"=";var ca=document.cookie.split(';');for(var i=0;i<ca.length;i++){var c=ca[i];while(c.charAt(0)==' ')c=c.substring(1,c.length);if(c.indexOf(nameEQ)==0){return c.substring(nameEQ.length,c.length);}}
return null;}
function delCookie(name){createCookie(name,"",-1);}`,
		},
	}

	bot := mid.NewSimpleFileServer(nil, nil, nil, nil)
	ms := NewMinifyServer(bot, []string{"/testdir/"}, []string{".js", ".css", ".html", ".htm"})
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
		goftlmux.ParseQueryParamsReg(wr, req, &wr.Ps) //
		lib.SetupTestMimicReq(req, "example.com")
		if db3 {
			fmt.Printf("{\"req\":%s,\n\"wr\":%s}\n", lib.SVarI(req), lib.SVarI(wr))
		}
		lib.SetupRequestHeaders(req, test.hdr)

		ms.ServeHTTP(wr, req)

		wr.FinalFlush()
		s := wr.GetBody()
		if db3 {
			fmt.Printf("Final body =====>>>>>%s<<<<<=====\n", s)
		}

		// xyzzy -has been hand checked and works-  Need auto-test
		if string(s) != test.expectedBody {
			t.Errorf("Error %d, Invalid body, got >%s< expected >%s<\n", ii, s, test.expectedBody)
		}

	}

}

const db3 = false
const db4 = false

/* vim: set noai ts=4 sw=4: */
