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

package HTML5Path

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/tr"
)

// ----------------------------------------------------------------------------------------------------------------------------------------------------
// Return a constant string value.
// ----------------------------------------------------------------------------------------------------------------------------------------------------
// ct := h.Get("Content-Type")
// if rw.StatusCode == http.StatusOK && strings.HasPrefix(ct, "application/json") {
type ConstHandler struct {
	Path []string
	S    string
	N    string
	V    string
}

// func NewConstHandler() *ConstHandler {
func NewConstIfMatchHandler(path []string, s, n, v string) http.Handler {
	return &ConstHandler{Path: path, S: s, N: n, V: v}
}

func (hdlr ConstHandler) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	// fmt.Printf("Match [%s] to [%s]\n", req.RequestURI, hdlr.Path)
	if lib.InArray(req.RequestURI, hdlr.Path) {
		www.Header().Set(hdlr.N, hdlr.V)
		www.Write([]byte(hdlr.S))
		www.WriteHeader(200)
	} else {
		www.WriteHeader(404)
	}
	return
}

// -----------------------------------------------------------------------------------------------------------------------------------------------
func Test_HTTP5Path(t *testing.T) {

	if !cfg.SetupRedisForTest("../test_redis.json") {
		return
	}

	tests := []struct {
		url            string
		expectedBody   string
		expectedStatus int
	}{
		{
			"http://example.com/index.html",
			`{"abc":"def"}`,
			200,
		},
		{
			"http://example.com/",
			`{"abc":"def"}`,
			200,
		},
		{
			"http://example.com/list",
			`{"abc":"def"}`,
			200,
		},
	}
	// ct := h.Get("Content-Type")
	// if rw.StatusCode == http.StatusOK && strings.HasPrefix(ct, "application/json") {
	bot := NewConstIfMatchHandler([]string{"/", "/index.html"}, `{"abc":"def"}`, "Content-Type", "text/html")

	// func NewHTML5PathServer(n http.Handler, p []string) *HTML5PathHandlerType {
	ms := NewHTML5PathServer(bot, []string{"/"})
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

		b := string(rec.Body.Bytes())
		if b != test.expectedBody {
			t.Errorf("Error %2d, reject error got: %s, expected %s\n", ii, b, test.expectedBody)
		}

		// xyzzy - check status

	}

}

/* vim: set noai ts=4 sw=4: */
