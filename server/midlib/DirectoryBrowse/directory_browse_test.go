//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1233
//

package DirectoryBrowse

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/testsup"
	"github.com/pschlump/Go-FTL/server/tr"
)

func Test_DrectoryBrowseServer(t *testing.T) {

	if !cfg.SetupRedisForTest("../test_redis.json") {
		return
	}

	tests := []struct {
		url string
		hdr []testsup.NameValue
	}{
		{"http://example.com/testdir/", []testsup.NameValue{testsup.NameValue{"X-Test", "A-Value"}}},
	}
	// ms := NewServer()
	bot := mid.NewServer()
	ms := NewDirectoryBrowseServer(bot, []string{"/"}, "", []string{"./www"})
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

		testsup.SetupTestCreateHeaders(wr, test.hdr)

		var req *http.Request

		req, err = http.NewRequest("GET", test.url, nil)
		if err != nil {
			t.Fatalf("Test %d: Could not create HTTP request: %v", ii, err)
		}
		lib.SetupTestMimicReq(req, "example.com")
		if db9 {
			fmt.Printf("{\"req\":%s,\n\"wr\":%s}\n", lib.SVarI(req), lib.SVarI(wr))
		}

		ms.ServeHTTP(wr, req)

		// Tests to perform on wr - first
		if wr.Error != nil {
			t.Errorf("Error %2d, Invalid error : %s\n", ii, wr.Error)
		}

		wr.FinalFlush()

		if wr.DirTemplateFileName != "www/index.tmpl" {
			t.Errorf("Error %2d, Invalid data set\n", ii)
		}

		// fmt.Printf(">>>> %s <<<<\n", wr.DirTemplateFileName)

	}

}

const db9 = false

/* vim: set noai ts=4 sw=4: */
