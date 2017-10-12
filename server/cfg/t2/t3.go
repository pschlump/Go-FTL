package main

import (
	"fmt"

	"github.com/pschlump/godebug"      //
	"github.com/pschlump/json"         //	"encoding/json"
	"github.com/pschlump/mapstructure" //
)

func run1() {

	fmt.Printf("Top of run1 - AT: %s\n", godebug.LF())

	type APair struct {
		A string
		B string
	}

	type Person struct {
		Name   string
		Age    int
		Paris  []APair
		Emails []string
		Extra  map[string]string
	}

	// This input can come from anywhere, but typically comes from
	// something like decoding JSON where we're not quite sure of the
	// struct initially.
	input := map[string]interface{}{
		"name":   "Mitchell",
		"age":    91,
		"Paris":  []APair{APair{"AAA", "BBB"}, APair{"aaa", "bbb"}},
		"emails": []string{"one", "two", "three"},
		"extra": map[string]string{
			"twitter": "mitchellh",
		},
	}
	input2s := `{
	"name": "Philip",
	"age": 53,
	"pairs": [
		{ "AAA", "aaa" },
		{ "AAb", "aab" },
		{ "AAc", "aac" }
	],
	"emails": [ "a@example.com", "b@example.com" ],
	"extra": {
		"ya": "ga",
		"da": "na
	}
}
`
	_ = input2s

	var result Person
	err := mapstructure.Decode(input, &result)
	if err != nil {
		panic(err)
	}

	fmt.Printf("go  v=%#v\n\n", result)
	fmt.Printf("SVarI=%s\n\n", SVarI(result))
	// Output:
	// mapstructure.Person{Name:"Mitchell", Age:91, Emails:[]string{"one", "two", "three"}, Extra:map[string]string{"twitter":"mitchellh"}}
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
