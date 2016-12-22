//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1189
//

package main

import (
	"fmt"
	"net/http"
)

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/bob", bobHandler)

	// mux.HandleErrors(404, http.HandlerFunc(myErrHandler))

	http.ListenAndServe(":7890", mux)
}

// trailing '/' allows handling of all request that start with '/' so '/', '/index.html', '/whatever'
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		myErrHandler(w, r)
		return
	}
	fmt.Fprint(w, "welcome home")
}

// Longest match catches /bob as an exact match
func bobHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/bob" {
		panic("Should never reach this point")
		return
	}
	fmt.Fprint(w, "welcome bob")
}

func myErrHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	fmt.Fprint(w, "custom 404")
}
