package goftlmux

//
// Go Go Mux - Go Fast Mux / Router for HTTP requests
//
// (C) Philip Schlump, 2013-2014.
// Version: 0.4.3
// BuildNo: 804
//
//

// Definitions:
//	Pat			Something of the form T::T, T:T{, that is a pattern used to define where constants and patterns are in a route.
//	Route		/abc/:def/ghi
//	CleanRoute	/abc/:/ghi
//	Names		[ "def" ]
//	Url			Input from user /abc/1234/ghi
//	Values		[ "1234" ]						 		Params

import "runtime"

// Return the line number and file name.  A single depth paramter 'd' can be supplied. 1
// is the routien that called LineFile, 2 is the caller of the routine that called
// LineFile etc.
func LineFile(d ...int) (string, int) {
	depth := 1
	if len(d) > 0 {
		depth = d[0]
	}
	_, file, line, ok := runtime.Caller(depth)
	if ok {
		return file, line
	}
	return "unkown", 0
}
