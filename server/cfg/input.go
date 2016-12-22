//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1004
//

// IsBool: name[InTestMode], File: /Users/corwin/go/src/github.com/pschlump/mapstructure/mapstructure.go LineNo:217

package cfg

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/oleiade/reflections" //
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/check-json-syntax/lib" // jsonSyntaxErroLib "github.com/pschlump/check-json-syntax/lib" //
	"github.com/pschlump/godebug"               //
	"github.com/pschlump/json"                  //	"encoding/json"
	"github.com/pschlump/mapstructure"          //
)

// Input and Validation
//
//
// Pass a chunk of input to a validaiton function
//
//	ok, msg := IsInputValid ( ValidationDataJson ValidationType, data map[string]interface{} ) ( bool, string ) {
//		return true, ""
//	}
//
//if ok - then valid, else "msg" is the syntax errors in that chunk.
//
//Map from a validated data set onto a structure for the user's data.
//
//	err := MapJsonToStruct ( data map[string]interface{}, st interface{} ) {
//		for name, val := range data {
//			x, found := lookupName ( name, st )
//			if found {
//				st.x = val.(type)
//			} else {
//				st.extra[name] = val
//			}
//		}
//	}
//
//
//	cfg.RegInitItem("simple_proxy", fx, `{
//		"Paths": { "type":["string","filepath"], "isarray":true, "required":true },
//		"To": { "type":["string","url"], "required":true },
//		"Extra": { "allowed":false }
//		}`)
//
//	type CfgSimpleProxyType struct {
//		Paths []string
//		To  string
//	}
//
// "http://192.168.0.157:2000/": {
//		"simple_proxy": {
//			"Paths": "/api",
//			"To": "http://localhost:8204/"
//		}
// }
//
// "http://192.168.0.157:2000/": {
//		"simple_proxy": {
//			"Paths": [ "/api", "/app" ] ,
//			"To": "http://localhost:8204/"
//		}
// }
//
//
// Notes: https://godoc.org/github.com/mitchellh/mapstructure
//		https://github.com/mitchellh/mapstructure

type VType struct {
	Type      []string `json:"type"`      // One of the types, string, int, float, filepath, url, bool,
	IsArray   bool     `json:"isarray"`   // Convert single item to array 1 long
	Required  bool     `json:"required"`  // Must be suplied
	Default   string   `json:"default"`   // A string that can be converted into a value if not supplied - implies that Required is meangless
	List      []string `json:"list"`      // Must be one of the listed values
	ReMatch   string   `json:"rematch"`   // Must match the regular expression
	MinValInt int      `json:"minvalint"` // Must be g.e. this value
	MaxValInt int      `json:"maxvalint"` // Must be l.e. this value
	Allowed   bool     `json:"allowed"`   // Is extra allowed?
	MinLength int      `json:"minlength"` // String MinLen, MaxLen
	MaxLength int      `json:"maxlength"` //
	reMatch   *regexp.Regexp
}

type ValidationType map[string]VType

func IsInputValid(mid_name string, ValidationDataJson string, data map[string]interface{}) (eok bool, dflt map[string]interface{}, msg string) {

	// fmt.Printf("At top of IsInputValid for %s\nvalid=%s\ndata=%s\n", mid_name, ValidationDataJson, lib.SVarI(data))

	eok = true // assume the best
	dflt = make(map[string]interface{})

	// conver ValidationDataJson -> ValidationType
	var vt ValidationType
	err := json.Unmarshal([]byte(ValidationDataJson), &vt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError: Invlaid validation for %s data %s%s\n", MiscLib.ColorRed, mid_name, ValidationDataJson, MiscLib.ColorReset)
		es := jsonSyntaxErroLib.GenerateSyntaxError(ValidationDataJson, err)
		fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorYellow, es, MiscLib.ColorReset)
		logrus.Errorf("Error: Invlaid validation for %s Error:\n%s\n", mid_name, es)
	}

	//fmt.Printf("vt=%s\n", lib.SVarI(vt))

	LineNoI, okA := data["LineNo"]
	LineNoF, okB := LineNoI.(float64)
	if !okA || !okB {
		LineNoF = 1
	}
	LineNo := int(LineNoF)

	// This loop only processes data that is "present" - so if something is a Default - it will be skipped
	// completely.  The next loop fills in defaults.  Therefore defaults will not get checked for
	// valid data.  That is also correct.   You many want to use an invalid value for a default
	// to mark data that is absent.

	for name, val := range data {
		// fmt.Printf("name [%s], %s\n", name, godebug.LF()) // "Paths", "To"
		if vv, ok := vt[name]; ok {
			// fmt.Printf("name [%s], %s\n", name, godebug.LF()) // "Paths", "To"

			switch val.(type) {
			case string:
				if !lib.InArray("string", vv.Type) {
					msg += fmt.Sprintf("Syntax Error: Line:%d Invalid type %T - Expected String, %s, is not allowed in %s\n", LineNo, val, name, mid_name)
					eok = false
				}
				for _, tt := range vv.Type {
					switch tt {
					case "string":
					case "url":
						// validate that this is an url
						s, ok := val.(string)
						if !ok {
							fmt.Printf("SYNTAX Error: invalid type [%T] for string, name[%s]\n", val, name)
						}
						_, err := url.Parse(s)
						if err != nil {
							msg += fmt.Sprintf("Syntax Error: Line:%d Invalid type - Expected URL, %s, is not allowed in %s\n", LineNo, name, mid_name)
							eok = false
						}
					case "filepath":
						s, ok := val.(string)
						if !ok {
							fmt.Printf("SYNTAX Error: invalid type [%T] for string, name[%s]\n", val, name)
						}
						if len(s) == 0 {
							msg += fmt.Sprintf("Syntax Error: Line:%d Invalid type %T - Expected FilePath, %s, is not allowed in %s\n", LineNo, val, name, mid_name)
							eok = false
						}
					case "float":
						f, ok := val.(string)
						if !ok {
							fmt.Printf("SYNTAX Error: invalid type [%T] for float, name[%s]\n", val, name)
						}
						fmt.Printf("name [%s], val[%f], %s\n", name, f, godebug.LF())
					case "int":
						i, ok := val.(string)
						if !ok {
							fmt.Printf("SYNTAX Error: invalid type [%T] for int/float, name[%s]\n", val, name)
						}
						fmt.Printf("name [%s], val[%d], %s\n", name, i, godebug.LF())
					case "bool":
						b, ok := val.(bool)
						if !ok {
							//s, ok := val.(string)
							//if ok {
							//	b := sizlib.ParseBool(s)
							//	fmt.Printf("String Converted To Boolean: name [%s], input[%s], val[%v], %s\n", name, s, b, godebug.LF())
							//} else {
							fmt.Printf("SYNTAX Error: invalid type [%T] for boolean, name[%s]\n", val, name)
							//}
						}
						fmt.Printf("name [%s], val[%v], %s\n", name, b, godebug.LF())
					case "hash":
					default:
						fmt.Printf("+======================================================================\n")
						fmt.Printf("| Invalid type - not checked %s, %s\n", tt, lib.LF())
						fmt.Printf("+======================================================================\n")
					}
				}
				// vv.List - if any set then must be in - check this
				if len(vv.List) > 0 {
					s := val.(string)
					if !lib.InArray(s, vv.List) {
						msg += fmt.Sprintf("Syntax Error: Line:%d Invalid - Expected to be one of %+v, got - in %s, %v\n", LineNo, vv.List, s, mid_name)
						eok = false
					}
				}
				// String MinLen, MaxLen
				if vv.MinLength > 0 {
					s := val.(string)
					if len(s) < vv.MinLength {
						msg += fmt.Sprintf("Syntax Error: Line:%d Invalid - string too short, expected %d got %d\n", LineNo, vv.MinLength, len(s))
						eok = false
					}
				}
				if vv.MaxLength > 0 {
					s := val.(string)
					if len(s) > vv.MaxLength {
						msg += fmt.Sprintf("Syntax Error: Line:%d Invalid - string too long, expected %d got %d\n", LineNo, vv.MaxLength, len(s))
						eok = false
					}
				}
				// Match regular expression
				if len(vv.ReMatch) > 0 {
					if vv.reMatch == nil {
						re, err := regexp.Compile(vv.ReMatch)
						if err != nil {
							msg += fmt.Sprintf("Syntax Error: Line:%d Invalid - Invalid regular expression : %s\n", LineNo, err)
							eok = false
						} else {
							vv.reMatch = re
						}
					}
					if vv.reMatch != nil {
						s := val.(string)
						if !vv.reMatch.MatchString(s) {
							msg += fmt.Sprintf("Syntax Error: Line:%d Invalid - string dit not match regular expression\n", LineNo)
							eok = false
						}
					}
				}
				if vv.IsArray {
					if _, ok := val.(string); ok {
						// fmt.Printf("****************** Doing IsArray convertion\n")
						data[name] = []string{val.(string)}
					}
				}
			case bool:
				if !lib.InArray("bool", vv.Type) {
					msg += fmt.Sprintf("Syntax Error: Line:%d Invalid type %T - Expected Bool, %s, is not allowed in %s\n", LineNo, val, name, mid_name)
					eok = false
				} else {
					fmt.Printf("name[%s] vv.Default=%s, %s\n", name, vv.Default, godebug.LF())
					if vv.Default != "" {
						b, err := strconv.ParseBool(vv.Default)
						if err == nil {
							dflt[name] = b
						} else {
							fmt.Printf("Default invalid type, %s\n", name)
						}
					}
				}
			case float64:
				isInt := lib.InArray("int", vv.Type)
				isFloat := lib.InArray("float", vv.Type)
				// fmt.Printf("float64 <><><><> isInt=%v isFloat=%v\n", isInt, isFloat)
				if isInt {
					data[name] = int(data[name].(float64))
					// xyzzy10 - MinValInt, MaxValInt
				} else if isFloat {
					// xyzzy10 - MinValFloat, MaxValFloat
				} else {
					msg += fmt.Sprintf("Syntax Error: Line:%d Invalid type %T - Expected Float, %s, is not allowed in %s\n", LineNo, val, name, mid_name)
					eok = false
				}
			case map[string]interface{}:
				fmt.Printf("**************************************** New one --- map[string]interface{} AT: %s\n", godebug.LF())
				// Nothing to do for validation?
				// xyzy10 validate it?
			case []interface{}:
				fmt.Printf("**************************************** New one --- Array of interface{} --- AT: %s\n", godebug.LF())
				// Nothing to do for validation?
				// xyzzy40 -- validate data in array
				// xyzzy40 - validate []interface{}
			default:
				fmt.Printf("**************************************** --- syntax error --- AT: %s\n", godebug.LF())
				msg += fmt.Sprintf("Syntax Error: Line:%d Invalid type %T, %s, is not allowed in %s\n", LineNo, val, name, mid_name)
				eok = false
			}
		} else if ww, ok2 := vt["Extra"]; ok2 {
			if ww.Allowed {
				// assign data to Extra[name] - in stuff
			} else {
				msg += fmt.Sprintf("Syntax Error: Line:%d Extra configuraiton field, %s, is not allowed in %s, %s\n", LineNo, name, mid_name, godebug.LF())
				eok = false
			}
		} else {
			msg += fmt.Sprintf("Syntax Error: Line:%d Extra configuraiton field, %s, is not allowed in %s, %s\n", LineNo, name, mid_name, godebug.LF())
			eok = false
		}
	}

	// Processing for default values - set up a temporary hash with the converted to correct type values -
	// this will get used later if no value is specified.
	for name, vv := range vt {
		if name == "Extra" {
		} else if len(vv.Default) > 0 {
			// convert and assign default value -- note this is after validation so default need not meet validation requirements.
			if len(vv.Type) > 0 {
				tt := vv.Type[0]
				switch tt {
				case "string":
					if !vv.IsArray {
						dflt[name] = vv.Default
					} else {
						dflt[name] = []string{vv.Default}
					}
				case "int":
					i, err := strconv.Atoi(vv.Default)
					if err != nil {
						msg += fmt.Sprintf("Syntax Error: Invalid default value for %s, should be int, Error: %s\n", name, err)
						eok = false
						i = 0
					}
					dflt[name] = i
				case "float":
					f, err := strconv.ParseFloat(vv.Default, 64)
					if err != nil {
						msg += fmt.Sprintf("Syntax Error: Invalid default value for %s, should be float, Error: %s\n", name, err)
						eok = false
						f = 0
					}
					dflt[name] = f
				case "bool":
					b, err := lib.ParseBool(vv.Default)
					if err != nil {
						msg += fmt.Sprintf("Syntax Error: Invalid default value for %s, should be bool, Error: %s\n", name, err)
						// xyzzyLogrus
						eok = false
						b = false
					}
					dflt[name] = b
				}
			} else {
				dflt[name] = vv.Default
			}
		} else if vv.Required {
			if _, ok := data[name]; !ok {
				msg += fmt.Sprintf("Syntax Error: Line:%d Required configuraiton field: %s is missing, module name=%s\n", LineNo, name, mid_name)
				// xyzzyLogrus
				eok = false
			}
		}
	}

	return
}

// Data is input data
// dflt is set of default values where not specified
// ms is the struct to map to
func MapJsonToStruct(data map[string]interface{}, dflt map[string]interface{}, ms interface{}) (err error) {

	// Verify that we were passed a pointer to a struture, if not error
	rv := reflect.ValueOf(ms)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &json.InvalidUnmarshalError{Type: reflect.TypeOf(ms)}
	}

	// Get all the "tags" in the structure
	structTags, err := reflections.Tags(ms, "matched")
	if err != nil {
		fmt.Printf("Error: Unable to get tags for structure - fatal error - not configured\n")
		// xyzzyLogrus
		return
	}
	// 	fmt.Printf("Tags Are: %s, data=%s dflt=%s, %s\n", lib.SVarI(structTags), lib.SVarI(data), lib.SVarI(dflt), lib.LF())

	for name := range structTags {
		if val, ok := dflt[name]; ok {
			// fmt.Printf("MapJsonToStruct: setting default value [%s], %s\n", name, godebug.LF())
			err = reflections.SetField(ms, name, val)
			if err != nil {
				fmt.Printf("Error: Unable to set field [%s] to default value [%v]\n", name, val)
				fmt.Printf("************************* err2 name=%s %T %T\n", name, ms, val)
				// xyzzyLogrus
			}
		}
	}

	// fmt.Printf("MapJsonToStruct: just before mapstructure, %s, data:%s Before:%s\n", godebug.SVarI(data), godebug.LF(), godebug.SVarI(ms))

	err = mapstructure.WeakDecode(data, ms)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		// xyzzyLogrus
	}

	// fmt.Printf("MapJsonToStruct: Results: %s\n", godebug.SVarI(ms))

	if _, ok0 := structTags["Extra"]; ok0 { // If we have a field named Extra then...
		ex := make(map[string]interface{}) // Create a map to hold extra data
		for name, val := range data {
			if _, ok := structTags[name]; !ok { // If this is NOT a field in the structure
				ex[name] = val // Save value
			}
		}
		err = reflections.SetField(ms, "Extra", ex)
		if err != nil {
			fmt.Printf("Error: Unable to set 'Extra' field with [%#v]\n", ex)
			fmt.Printf("************************* err3, %s\n", err)
			// xyzzyLogrus
		}
		fmt.Printf("Resuling Value Extra(3): %#v\n", ms)
	}

	return
}

/* vim: set noai ts=4 sw=4: */
