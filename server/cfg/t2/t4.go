package main

import (
	"fmt"
	"os"

	"github.com/pschlump/godebug"      //
	"github.com/pschlump/json"         //	"encoding/json"
	"github.com/pschlump/mapstructure" //
)

func run1() {

	fmt.Printf("Top of run1/t4 - AT: %s\n", godebug.LF())

	type APair struct {
		A string
		B string
	}

	type Person struct {
		Name   string
		Age    int
		Pairs  []APair
		Emails []string
		Extra  map[string]string
	}

	// This input can come from anywhere, but typically comes from
	// something like decoding JSON where we're not quite sure of the
	// struct initially.
	input := map[string]interface{}{
		"name":   "Mitchell",
		"age":    91,
		"Pairs":  []APair{APair{"AAA", "BBB"}, APair{"aaa", "bbb"}},
		"emails": []string{"one", "two", "three"},
		"extra": map[string]string{
			"twitter": "mitchellh",
		},
	}
	_ = input

	input2s := `{
	"name": "Philip",
	"age": 153,
	"pairs": [
		{ "A":"AAA", "B":"aaa" },
		{ "A":"AAb", "B":"aab" },
		{ "A":"AAc", "B":"aac" }
	],
	"inkyDinky": "should-not-map-or-error",
	"emails": [ "a@example.com", "b@example.com" ],
	"extra": {
		"ya": "ga",
		"da": "na"
	}
}
`
	input2 := make(map[string]interface{})

	err := json.Unmarshal([]byte(input2s), &input2)
	if err != nil {
		fmt.Printf("Error parsing JSON: %s\n", err)
		os.Exit(1)
	}

	// _ = input2s
	// _ = input2

	var result Person
	err = mapstructure.Decode(input2, &result)
	if err != nil {
		panic(err)
	}

	fmt.Printf("go  v=%#v\n\n", result)
	fmt.Printf("SVarI=%s\n\n", SVarI(result))
}

// -------------------------------------------------------------------------------------------------
func SVar(v interface{}) string {
	s, err := json.Marshal(v)
	// s, err := json.MarshalIndent ( v, "", "\t" )
	if err != nil {
		return fmt.Sprintf("Error:%s", err)
	} else {
		return string(s)
	}
}

// -------------------------------------------------------------------------------------------------
func SVarI(v interface{}) string {
	// s, err := json.Marshal ( v )
	s, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return fmt.Sprintf("Error:%s", err)
	} else {
		return string(s)
	}
}

func main() {
	run1()
}
