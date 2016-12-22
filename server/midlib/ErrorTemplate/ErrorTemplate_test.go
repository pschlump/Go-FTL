//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1281
//

package ErrorTemplate

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
func Test_ErrorTemplateServer(t *testing.T) {
	tests := []struct {
		url          string
		from_ip      string
		expectedCode int
		expectedBody string
	}{
		{
			"http://example.com/foo?abc=def",
			"1.1.1.1",
			http.StatusNotFound,
			`<html>
<body>
<p>
Status Code: 404 <br>
Error: Not Found <br>
</p>
<p>
The error has been logged and will be looked into.  To return to the application, <a href="http://www.2c-why.com/">click.</a>
</p>
</body>
</html>
`,
		},
		{
			"http://example.com/foo?abc=def",
			"1.1.1.2",
			http.StatusOK,
			"",
		},
	}
	// bot := NewBotHandler()
	bot := mid.NewServer()
	// ms := NewHeaderServer(bot, []string{"/foo"}, "X-Test2", "A-Value2")
	ms := NewErrorTemplateServer(bot, []string{"/foo"}, []string{"404", "500"})
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec,nil)

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
		wr.StatusCode = test.expectedCode

		ms.ServeHTTP(wr, req)

		// Tests to perform on final recorder data.
		if wr.StatusCode != test.expectedCode {
			t.Errorf("Error %2d, reject error got: %d, expected %d\n", ii, wr.StatusCode, test.expectedCode)
		} else {

			wr.FinalFlush()

			bod := string(rec.Body.Bytes())

			// xyzzy - check that the template was rendered.
			if bod != test.expectedBody {
				t.Errorf("Error %2d, invalid body got: >>%s<<, expected >>%s<<\n", ii, bod, test.expectedBody)
			}
		}
	}

}

const db4 = false

/* vim: set noai ts=4 sw=4: */
