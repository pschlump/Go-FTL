//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2015-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1257
//

package JSONToTable

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
// Future: Add - meta data add on - no_rows and make hash always. -- Takes array and makes { "meta": { ... }, "data": { .... } } - stored in wr.Row
// -----------------------------------------------------------------------------------------------------------------------------------------------

func Test_JSONToTable_01_Server(t *testing.T) {

	tests := []struct {
		runTest            bool
		url                string
		hdr                []lib.NameValue
		inputData          string
		expectedDataOutput string
		expecedNRows       int
		BufferState        goftlmux.StateType // Byte, Row, Table
		expectError        bool
	}{
		{ // test with correct array data
			true,
			"http://example.com/foo?abc=def&$privs$=user&t=user&template_name=data.tmpl",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			`[{"a":"b"}]`,
			`[{"a":"b"}]`,
			1,
			goftlmux.TableBuffer,
			false,
		},
		{ // test with a hash
			true,
			"http://example.com/foo?$privs$=user&t=user",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			`{"a":"b"}`,
			`[{"a":"b"}]`,
			1,
			goftlmux.TableBuffer,
			false,
		},
		{ // test with invalid JSON data
			true,
			"http://example.com/foo?$privs$=user&t=user",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			`{"a"b"}`,
			`[{}]`, // I am not certain that this is really a correct return value, maybe `[]`
			1,
			goftlmux.TableBuffer,
			true,
		},
		{ // test with correct array data
			true,
			"http://example.com/foo?abc=def&$privs$=user&t=user&template_name=data.tmpl",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			`[]`,
			`[]`,
			0,
			goftlmux.TableBuffer,
			true,
		},
	}

	bot := mid.NewServer()
	// func NewGoTemplateServer(n http.Handler, p []string) *GoTemplateType {
	ms := NewJSONToTableServer(bot, []string{"/foo"}, true, false) // 2nd test with false
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		if db3 {
			fmt.Printf("\nTest %d ---------------------------------------------------------------\n", ii)
		}

		if test.expectError {
			fmt.Printf("\tExpect log messages to be printed out\n")
		}

		if !test.runTest {
			continue
		}

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

		bot.SetInfo(test.inputData)

		if db3 {
			fmt.Printf("Bef: wr.Row = %s wr.Table=%s wr.State=%s\n", lib.SVarI(wr.Row), lib.SVarI(wr.Table), wr.State)
		}

		ms.ServeHTTP(wr, req)

		s := wr.GetBody()
		if db3 {
			fmt.Printf("Aft: wr.Row = %s wr.Table=%s wr.State=%s body >>>%s<<<\n", lib.SVarI(wr.Row), lib.SVarI(wr.Table), wr.State, s)
		}

		if wr.State != test.BufferState {
			t.Errorf("Error %2d, Invalid  state - should be %s, got %s\n", ii, test.BufferState, wr.State)
		}
		if wr.State == goftlmux.TableBuffer {
			if len(wr.Table) != test.expecedNRows {
				t.Errorf("Error %2d, Invalid  expecedNRows, got %d expected %d\n", ii, len(wr.Table), test.expecedNRows)
			}
			tt := lib.SVar(wr.Table)
			if tt != test.expectedDataOutput {
				t.Errorf("Error %2d, Invalid data, got >%s< expected >%s<\n", ii, tt, test.expectedDataOutput)
			}
		}
		if wr.State == goftlmux.RowBuffer {
			tt := lib.SVar(wr.Row)
			if tt != test.expectedDataOutput {
				t.Errorf("Error %2d, Invalid data, got >%s< expected >%s<\n", ii, tt, test.expectedDataOutput)
			}
		}

		if db3 {
			fmt.Printf("End of %d test\n\n\n", ii)
		}

	}

}

func Test_JSONToTable_02_Server(t *testing.T) {

	tests := []struct {
		runTest            bool
		url                string
		hdr                []lib.NameValue
		inputData          string
		expectedDataOutput string
		expecedNRows       int
		BufferState        goftlmux.StateType // Byte, Row, Table
	}{
		{ // test with correct array data
			true,
			"http://example.com/foo?abc=def&$privs$=user&t=user&template_name=data.tmpl",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			`[{"a":"b"}]`,
			`{"a":"b"}`,
			1,
			goftlmux.RowBuffer,
		},
		{ // test with a hash
			true,
			"http://example.com/foo?$privs$=user&t=user",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			`{"a":"b"}`,
			`{"a":"b"}`,
			0,
			goftlmux.RowBuffer,
		},
		{ // test with invalid JSON data
			true,
			"http://example.com/foo?$privs$=user&t=user",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			`{"a","b"}`,
			`{}`,
			0,
			goftlmux.RowBuffer,
		},
		{ // test with correct array data
			true,
			"http://example.com/foo?abc=def&$privs$=user&t=user&template_name=data.tmpl",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			`[]`,
			`{}`,
			0,
			goftlmux.RowBuffer,
		},
	}

	bot := mid.NewServer()
	// func NewGoTemplateServer(n http.Handler, p []string) *GoTemplateType {
	ms := NewJSONToTableServer(bot, []string{"/foo"}, false, true) // 2nd test with false
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		if !test.runTest {
			continue
		}

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

		bot.SetInfo(test.inputData)

		if db3 {
			fmt.Printf("Bef: wr.Row = %s wr.Table=%s wr.State=%s\n", lib.SVarI(wr.Row), lib.SVarI(wr.Table), wr.State)
		}

		ms.ServeHTTP(wr, req)

		s := wr.GetBody()
		if db3 {
			fmt.Printf("Aft: wr.Row = %s wr.Table=%s wr.State=%s body >>>%s<<<\n", lib.SVarI(wr.Row), lib.SVarI(wr.Table), wr.State, s)
		}

		if wr.State != test.BufferState {
			t.Errorf("Error %2d, Invalid  state - should be %s, got %s\n", ii, test.BufferState, wr.State)
		}
		if wr.State == goftlmux.TableBuffer {
			if len(wr.Table) != test.expecedNRows {
				t.Errorf("Error %2d, Invalid  expecedNRows, got %d expected %d\n", ii, len(wr.Table), test.expecedNRows)
			}
			tt := lib.SVar(wr.Table)
			if tt != test.expectedDataOutput {
				t.Errorf("Error %2d, Invalid data, got >%s< expected >%s<\n", ii, tt, test.expectedDataOutput)
			}
		}
		if wr.State == goftlmux.RowBuffer {
			tt := lib.SVar(wr.Row)
			if tt != test.expectedDataOutput {
				t.Errorf("Error %2d, Invalid data, got >%s< expected >%s<\n", ii, tt, test.expectedDataOutput)
			}
		}

		if db3 {
			fmt.Printf("End of %d test\n\n\n", ii)
		}

	}

}

func Test_JSONToTable_03_Server(t *testing.T) {

	tests := []struct {
		runTest            bool
		url                string
		hdr                []lib.NameValue
		inputData          string
		expectedDataOutput string
		expecedNRows       int
		BufferState        goftlmux.StateType // Byte, Row, Table
	}{
		{ // test with correct array data
			runTest:            true,
			url:                "http://example.com/foo?abc=def&$privs$=user&t=user&template_name=data.tmpl",
			hdr:                []lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			inputData:          `[{"a":"b"}]`,
			expectedDataOutput: `[{"a":"b"}]`,
			expecedNRows:       1,
			BufferState:        goftlmux.TableBuffer,
		},
		{ // test with a hash
			runTest:            true,
			url:                "http://example.com/foo?$privs$=user&t=user",
			hdr:                []lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			inputData:          `{"a":"b"}`,
			expectedDataOutput: `{"a":"b"}`,
			expecedNRows:       1,
			BufferState:        goftlmux.RowBuffer,
		},
		{ // test with invalid JSON data
			runTest:            true,
			url:                "http://example.com/foo?$privs$=user&t=user",
			hdr:                []lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			inputData:          `{"a","b"}`,
			expectedDataOutput: `{}`,
			expecedNRows:       0,
			BufferState:        goftlmux.RowBuffer,
		},
		{ // test with correct array data
			runTest:            true,
			url:                "http://example.com/foo?abc=def&$privs$=user&t=user&template_name=data.tmpl",
			hdr:                []lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			inputData:          `[]`,
			expectedDataOutput: `[]`,
			expecedNRows:       0,
			BufferState:        goftlmux.TableBuffer,
		},
	}

	bot := mid.NewServer()
	// func NewGoTemplateServer(n http.Handler, p []string) *GoTemplateType {
	ms := NewJSONToTableServer(bot, []string{"/foo"}, false, false) // 2nd test with false
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		if !test.runTest {
			continue
		}

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

		bot.SetInfo(test.inputData)

		if db3 {
			fmt.Printf("Bef: wr.Row = %s wr.Table=%s wr.State=%s\n", lib.SVarI(wr.Row), lib.SVarI(wr.Table), wr.State)
		}

		ms.ServeHTTP(wr, req)

		s := wr.GetBody()
		if db3 {
			fmt.Printf("Aft: wr.Row = %s wr.Table=%s wr.State=%s body >>>%s<<<\n", lib.SVarI(wr.Row), lib.SVarI(wr.Table), wr.State, s)
		}

		if wr.State != test.BufferState {
			t.Errorf("Error %2d, Invalid  state - should be %s, got %s\n", ii, test.BufferState, wr.State)
		}
		if wr.State == goftlmux.TableBuffer {
			if len(wr.Table) != test.expecedNRows {
				t.Errorf("Error %2d, Invalid  expecedNRows, got %d expected %d\n", ii, len(wr.Table), test.expecedNRows)
			}
			tt := lib.SVar(wr.Table)
			if tt != test.expectedDataOutput {
				t.Errorf("Error %2d, Invalid data, got >%s< expected >%s<\n", ii, tt, test.expectedDataOutput)
			}
		}
		if wr.State == goftlmux.RowBuffer {
			tt := lib.SVar(wr.Row)
			if tt != test.expectedDataOutput {
				t.Errorf("Error %2d, Invalid data, got >%s< expected >%s<\n", ii, tt, test.expectedDataOutput)
			}
		}

		if db3 {
			fmt.Printf("End of %d test\n\n\n", ii)
		}

	}

}

const db3 = false

/* vim: set noai ts=4 sw=4: */
