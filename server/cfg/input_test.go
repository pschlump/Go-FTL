//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1005
//

package cfg

import (
	"fmt"
	"testing"

	"github.com/pschlump/Go-FTL/server/lib"

	JsonX "github.com/pschlump/JSONx"
)

// "github.com/pschlump/json" //	"encoding/json"

func Test_IsInputValid(t *testing.T) {
	tests := []struct {
		mid_name         string
		vs               string
		raw_data         string
		expectedValid    bool
		expectedMsg      string
		expectedDfltName string
		expectedDflt     string
	}{
		{
			mid_name: "simple_proxy",
			vs: `{
				"Paths":     { "type":["string","filepath"], "isarray":true, "required":true },
				"To":        { "type":["string","url"], "required":true, "default":"http://localhost/" },
				"LineNo":    { "type":[ "int" ], "default":"1" },
				"Extra":     { "allowed":false }
				}`,
			raw_data: `{ "LineNo":12,
				"Paths": "/api",
				"To": "http://localhost:8204/"
				}`,
			expectedValid:    true,
			expectedDfltName: "To",
			expectedDflt:     "http://localhost/",
		},
		{
			mid_name: "simple_proxy",
			vs: `{
				"Paths":   { "type":["string","filepath"], "isarray":true, "required":true },
				"To":      { "type":["string","url"], "required":true, "default":"http://localhost/" },
				"Extra":   { "allowed":false },
				"LineNo":  { "type":[ "int" ], "default":"1" }
				}`,
			raw_data:         `{ "LineNo":12 }`,
			expectedValid:    false,
			expectedDfltName: "To",
			expectedDflt:     "http://localhost/",
		},
		{
			mid_name: "dumpIt",
			vs: `{
				"Paths":        { "type":["string","filepath"], "isarray":true, "required":true },
				"Msg":          { "type":[ "string" ], "default":"abc" },
				"Msg2":         { "type":[ "string", "isarray" ] },
				"SaveBodyFlag": { "type":[ "bool" ], "default":"true" },
				"FileName":     { "type":[ "string","filepath" ], "default":"" },
				"LineNo":       { "type":[ "int" ], "default":"1" }
				}`,
			raw_data: `{ 
				"LineNo": 8,
				"Paths":"/",
				"Msg": "After Proxy"
				}`,
			expectedValid:    true,
			expectedDfltName: "",
		},
	}

	//
	// TODO
	// 1. Add ability to verify a 'set' of default values are found -- Name/Value
	// 2. test that default values get set for stuff that is NOT in data
	// 3. Test -- func MapJsonToStruct(data map[string]interface{}, dflt map[string]interface{}, ms interface{}) (err error) {
	// 1. testing of input -
	//		/Users/corwin/go/src/github.com/pschlump/Go-FTL/server/cfg/input.go
	//		/Users/corwin/go/src/github.com/pschlump/Go-FTL/server/cfg/test_input.go
	//		1. Test with each data type
	//		1. Test with isarray
	//		1. Test with and w/o defaults
	//		1. Test with syntax errors
	//		1. Test with hashees/structs
	//		1. Test with array of hash
	//		1. Implement defaults in array - add this capability (array of string only)
	// 1. test showing "Extra":     { "allowed":false }
	// 1. test showing that defaults get set in final struct
	//

	for ii, test := range tests {

		if db818 {
			fmt.Printf("---------------------------------------------- test %d ----------------------------------------------------- \n\n", ii)
		}

		data := make(map[string]interface{})
		// err := json.Unmarshal([]byte(test.raw_data), &data)
		meta, err := JsonX.Unmarshal(fmt.Sprintf("Test:%d", ii), []byte(test.raw_data), &data)
		_ = meta
		if err != nil {
			t.Errorf("Error: Invlaid test %s/%d data %s\n", test.mid_name, ii, test.raw_data)
		} else {

			// fmt.Printf("Data before >>>%s<<<\n", lib.SVarI(data))

			eok, dflt, msg := IsInputValid(test.mid_name, test.vs, data)
			// xyzzy40 - validate []interface{}

			if db818 {
				fmt.Printf("Data is now >>>%s<<<\n", lib.SVarI(data))
				fmt.Printf("Dflt is now >>>%s<<<\n", lib.SVarI(data))
			}

			//if eok != test.expectedValid {
			//	t.Errorf("Error %2d, Invalid error, got %v expected %v\n", ii, eok, test.expectedValid)
			//}
			if test.expectedValid {
				if !eok {
					fmt.Printf(">>>%s<<<\n", msg)
				}
			} else {
				if eok {
					t.Errorf("Error %2d, Expected syntax error, did not get one", ii)
				}
			}

			if len(test.expectedDfltName) > 0 {
				// Sample Defaults
				// fmt.Printf("dflt >>>%s<<<\n", dflt)
				if x, ok := dflt[test.expectedDfltName]; !ok {
					t.Errorf("Error %2d, Missing default for %s, expected %v\n", ii, test.expectedDfltName, test.expectedDflt)
				} else {
					if x != test.expectedDflt {
						t.Errorf("Error %2d, Missing default value %s, expected %v, got %s\n", ii, test.expectedDfltName, test.expectedDflt, x)
					}
				}
			}

		}
	}

}

// func MapJsonToStruct(data map[string]interface{}, ms interface{}) (err error) {
type CfgSimpleProxyType struct {
	Paths []string
	To    string
}

func Test_MapJsonToStruct(t *testing.T) {
	aCfgSimpleProxyType := CfgSimpleProxyType{}

	data := make(map[string]interface{})
	data["Paths"] = []string{"/abc"}
	// data["To"] = "http://localhost:8888/"
	dflt := make(map[string]interface{})
	dflt["To"] = "http://localhost:8080/"

	err := MapJsonToStruct(data, dflt, &aCfgSimpleProxyType)
	// xyzzy10 - implement setting of extra
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	} else {
		if db818 {
			fmt.Printf("End Result=%s\n", lib.SVarI(aCfgSimpleProxyType))
		}
	}
}

const db818 = false
