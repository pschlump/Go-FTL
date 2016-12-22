//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1188
//

// Mathc Path

package main

import (
	"fmt"
	"path"
	"strings"
)

// 0. Patterns in URL
// 		/abc/{pattern}/def/{pattern}
// 1. All patterns are compiled and cached
// 2. Patterns are matched "after" the rest of the URL - on fail go on to next match

// 1. Take a path and clean it.
// 2. Look it up in a table - max # of matches
// 3.

// 1. Split /a/b/c -> [ a b c ]
//

type MatchInProg struct {
	MatchPath  string
	Components []string
}

func test1() {
	var mp []MatchInProg = []MatchInProg{
		{
			MatchPath: "/aaa/bbb/ccc",
		},
		{
			MatchPath: "aaa/bbb/ccc/",
		},
		{
			MatchPath: "aaa/bbb/./ddd/../ccc/",
		},
		{
			MatchPath: "////aaa/bbb/./ddd/../ccc/",
		},
		{
			MatchPath: "/../..//..//aaa/bbb/./ddd/../ccc/",
		},
		{
			MatchPath: "aaa/bbb/ccc/def/ghi",
		},
		{
			MatchPath: "/",
		},
		{
			MatchPath: "",
		},
	}

	for ii, vv := range mp {
		cp := CleanPath(vv.MatchPath)
		fmt.Printf("[%d] orig >%s< cp = >%s<\n", ii, vv.MatchPath, cp)
		vv.Components = strings.Split(cp, "/")
		fmt.Printf("[%d] components = >%s<\n", ii, vv.Components)

		mp[ii] = vv
	}

}

// Return the canonical path for p, eliminating . and .. elements.
func CleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	return np
}

func main() {
	test1()
}
