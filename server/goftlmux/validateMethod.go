package goftlmux

//
// Go Go Mux - Go Fast Mux / Router for HTTP requests
//
// (C) Philip Schlump, 2013-2017.
// Version: 0.4.9
// BuildNo: 804
//
// /Users/corwin/Projects/gogo2
//

import (
	"fmt"

	debug "github.com/pschlump/godebug"
)

// Table of valid methods,  If other http-methods are created or used they should be added to this list.
var validMethod map[string]bool

// Table of valid schemes, https, http
var validScheme map[string]bool

var validProtocal map[string]bool

func init() {
	validMethod = make(map[string]bool)
	validMethod["GET"] = true
	validMethod["PUT"] = true
	validMethod["POST"] = true
	validMethod["PATCH"] = true
	validMethod["OPTIONS"] = true
	validMethod["HEAD"] = true
	validMethod["DELETE"] = true
	validMethod["CONNECT"] = true
	validMethod["TRACE"] = true

	validScheme = make(map[string]bool)
	validScheme["https"] = true
	validScheme["http"] = true

	validProtocal = make(map[string]bool)
	validProtocal["HTTP/1.0"] = true
	validProtocal["HTTP/1.1"] = true
	validProtocal["HTTP/2.0"] = true
}

func checkInBoolMap(mm []string, cmpTo map[string]bool, em string) bool {
	for _, v := range mm {
		if b, ok := cmpTo[v]; !ok || !b {
			fmt.Printf("Error(20000): %s %s is invalid.  Called From: %s\n", em, v, debug.LF(4))
			return false
		}
	}
	return true
}

// Check for GET, PUT etc.
func checkMethods(methods []string) bool {
	return checkInBoolMap(methods, validMethod, "Method")
}

// Check for https / http - valid schemes
func checkScheme(s []string) bool {
	return checkInBoolMap(s, validScheme, "Scheme")
}

// Check for HTTP/1.0 etc.
func checkProtocal(s []string) bool {
	return checkInBoolMap(s, validProtocal, "Protocal")
}
