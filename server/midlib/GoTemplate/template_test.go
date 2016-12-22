//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1245
//

package GoTemplate

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
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

	var dd []map[string]interface{}
	dd = make([]map[string]interface{}, 2, 2)
	dd[0] = make(map[string]interface{})
	dd[1] = make(map[string]interface{})
	dd[0]["abc"] = "row-no-000"
	dd[1]["abc"] = "row-no-001"
	dd[0]["myId"] = 100
	dd[1]["myId"] = 101

	tests := []struct {
		url          string
		hdr          []lib.NameValue
		data_raw     []map[string]interface{} //	Table of Row Response
		expectedBody string
	}{
		{
			url:      "http://example.com/foo?abc=def&$privs$=user&t=user&template_name=data.tmpl",
			hdr:      []lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			data_raw: dd,
			expectedBody: `<!DOCTYPE html>
<html lang="en">
<body>
	<div> header </div>
	<ul>

	
		<li><a href="row-no-000">row-no-000</a></li>
	
		<li><a href="row-no-001">row-no-001</a></li>
	

	</ul>
	<div> footer </div>
</body>
</html>
`,
		},
		{
			url:          "http://example.com/def?$privs$=user&t=user",
			hdr:          []lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			data_raw:     dd,
			expectedBody: `[{"abc":"row-no-000","myId":100},{"abc":"row-no-001","myId":101}]`,
		},
	}

	bot := mid.NewServer()
	// func NewGoTemplateServer(n http.Handler, p []string) *GoTemplateType {
	ms := NewGoTemplateServer(bot, []string{"/foo"}, "data.tmpl", []string{}) // TODO: add in temlate libraries in tests
	var err error
	lib.SetupTestCreateDirs()
	re := regexp.MustCompile("[ 	][ 	]*")

	for ii, test := range tests {

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec, nil) // var wr http.ResponseWriter
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

		wr.State = goftlmux.TableBuffer
		wr.Table = test.data_raw
		if db3 {
			fmt.Printf("wr.Table = %s\n", lib.SVarI(wr.Table))
			fmt.Printf("wr.State = %s\n", wr.State)
		}

		ms.ServeHTTP(wr, req)

		//if wr.Table != nil { // len of TableData > 0 &&
		//	// Tests to perform on final recorder data.
		//	if wr.State != goftlmux.TableBuffer {
		//		t.Errorf("Error %2d, Invalid data returned\n", ii)
		//	}
		//	// xyzzy - verify data
		//}

		wr.FinalFlush()
		s := wr.GetBody()
		bb := string(s)
		bb = re.ReplaceAllString(bb, "@")
		tt := re.ReplaceAllString(test.expectedBody, "@")
		if db8 {
			fmt.Printf("Final body =====>>>>>%s<<<<<=====\n", bb)
			fmt.Printf("Expec body =====>>>>>%s<<<<<=====\n", tt)
		}

		// if test.expectedBody != bb {
		if bb != tt {
			t.Errorf("Error %d, Invalid body, got >%s< expected >%s<\n", ii, bb, test.expectedBody)
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
const db8 = false

/* vim: set noai ts=4 sw=4: */
