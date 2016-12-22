//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1296
//

package nameresolve

import (
	"fmt"
	"testing"

	"github.com/pschlump/godebug"
)

// -----------------------------------------------------------------------------------------------------------------------------------------------
// test that redirect will tranform ULR
//
// 1. req.URL - modified
// 2. req.RequestURI - modified

func Test_NameResolver1(t *testing.T) {

	// xyzzy - use "ID" to find DUPS - error result when adding.

	setupBot := []struct {
		name string
		id   int
	}{
		{"http://www.test1.com/", 1},
		{"http://*.test1.com/", 2},
		{"http://*.test1.com:1000/", 3},
		{"https://www.test1.com/", 4},
		{"https://www.test1.com:1000/", 5},
		{"http://app-t00.test1.com/", 7},
		{"http://app-t01.test1.com/", 8},
		{"http://app-t02.test1.com/", 9},
		{"http://www.test2.com/", 10},
		{"http://test2.com/", 11},
		{"http://192.168.0.157/", 12},
		{"http://localhost/", 13},
		{"http://localhost:8080/", 14},
	}

	tests := []struct {
		run_it     bool
		url        string
		expectedId int
	}{
		{true, "http://www.test1.com", 1},
		{true, "http://bob.test1.com", 2},
		{true, "http://localhost", 13},
		{true, "http://www.test3.com", -1},
		{true, "http://app-t00.test1.com", 7},
		{true, "http://app-t01.test1.com", 8},
		{true, "http://app-t02.test1.com", 9},
		{true, "http://mike.test1.com:1000/", 3},
		{true, "http://jane.test1.com:1000/", 3},
		{true, "http://localhost/", 13},
		{true, "http://192.168.0.157/", 12},
	}

	bot := NewNameResolve()
	// bot.Debug1 = true
	// bot.Debug2 = true
	// bot.Debug3 = true
	bot.Debug4 = true
	bot.Debug5 = true

	for _, vv := range setupBot {
		// func (nr *NameResolve) AddName(namePattern string, hdlr http.Handler, id int, addrIfNone string) (e error) {
		bot.AddName(vv.name, nil, vv.id, vv.name)
	}

	if db3 {
		fmt.Printf("Lookup table: %s\n", godebug.SVarI(bot))
	}

	for ii, vv := range tests {
		if vv.run_it {
			rv, ok := bot.GetHandler(vv.url)
			if ok {
				if db3 {
					fmt.Printf("Found: %s//%s:%s == %d\n", rv.Proto, rv.Host, rv.Port, rv.Id)
				}
				if vv.expectedId != rv.Id {
					t.Errorf("Error %2d, Expeced ID = %d, found %d\n", ii, vv.expectedId, rv.Id)
				}
			} else {
				if vv.expectedId != -1 {
					t.Errorf("Error %2d, Expeced to not find %s, found it\n", ii, vv.url)
				} else {
					if db1 {
						fmt.Printf("Not Found (as expectd), test %d\n", ii)
					}
				}
			}
		}
	}

	return

	bot.AddDefault("http:", "*", nil, 100)

	tests2 := []struct {
		url        string
		expectedId int
	}{
		{"http://www.test1.com", 1},
		{"http://www.test3.com", 100},
		{"http://app-t00.test1.com", 7},
		{"http://app-t01.test1.com", 8},
		{"http://app-t02.test1.com", 9},
	}

	for ii, vv := range tests2 {
		rv, ok := bot.GetHandler(vv.url)
		if ok {
			if db3 {
				fmt.Printf("Found: %s//%s:%s == %d\n", rv.Proto, rv.Host, rv.Port, rv.Id)
			}
			if vv.expectedId != rv.Id {
				t.Errorf("Error %2d, Expeced ID = %d, found %d\n", ii, vv.expectedId, rv.Id)
			}
		} else {
			if vv.expectedId != -1 {
				t.Errorf("Error %2d, Expeced to not find %s, found it\n", ii, vv.url)
			} else {
				if db1 {
					fmt.Printf("Not Found (as expectd), test %d\n", ii)
				}
			}
		}
	}

}

const db1 = false
const db3 = false
