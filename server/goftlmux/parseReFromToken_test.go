package goftlmux

//
// Go Go Mux - Go Fast Mux / Router for HTTP requests
//
// (C) Philip Schlump, 2013-2014.
// Version: 0.4.3
// BuildNo: 804
//
//

import "testing"

var testParseReRun = []struct {
	In           string
	ExpectName   string
	ExpectRe     string
	ExpectValid  bool
	Expect3valid bool
	Expect3conv  bool
}{
	/* 00 */ {"{abc:[a-z]}", "abc", "[a-z]", true, true, false},
	/* 01 */ {"abc:[a-z]}", "abc", "[a-z]", true, true, false},
	/* 02 */ {"abc}", "abc", "", false, true, true},
	/* 03 */ {"abc:a{3,4}}", "abc", "a{3,4}", true, true, false},
	/* 04 */ {"{}", "", "", false, false, false},
	/* 05 */ {"{:[a-z]}", "", "[a-z]", false, false, false},
	/* 06 */ {"", "", "", false, false, false},
	/* 07 */ {"}", "", "", false, false, false},
	/* 08 */ {"{", "", "", false, false, false},
	/* 09 */ {"{z2}", "z2", "", false, true, true},
}

func TestParseReFromToken3(t *testing.T) {
	for i, test := range testParseReRun {
		// fmt.Printf("test %d\n", i)
		name, re, valid, conv := parseReFromToken3(test.In)
		if name != test.ExpectName {
			t.Errorf("Test: %d, Expected Result = %s, got %s\n", i, test.ExpectName, name)
		}
		if re != test.ExpectRe {
			t.Errorf("Test: %d, Expected Result = %s, got %s\n", i, test.ExpectRe, re)
		}
		if valid != test.Expect3valid {
			t.Errorf("Test: %d, Expected Result = %v, got %v\n", i, test.Expect3valid, valid)
		}
		if conv != test.Expect3conv {
			t.Errorf("Test: %d, Expected Result = %v, got %v\n", i, test.Expect3conv, conv)
		}
	}
}
