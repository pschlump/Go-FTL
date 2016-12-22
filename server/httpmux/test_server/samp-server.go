//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1190
//

package main

import (
	"fmt"
	"net/http"

	tr "github.com/pschlump/godebug"

	"github.com/pschlump/Go-FTL/server/httpmux"
)

func main() {

	mux := httpmux.NewServeMux()

	mux.HandleFunc("/", homeHandler).Method("GET", "POST")
	mux.HandleFunc("/bob", bobHandler).Method("GET", "POST")
	mux.HandleFunc("/abc", aaaHandler).Method("GET", "POST")
	mux.HandleFunc("/abc/def", aaaHandler).Method("GET", "POST")
	mux.HandleFunc("/abc/def/ghi", aaaHandler).Method("GET", "POST")
	mux.HandleFunc("/xyz/", xyzGetHandler).Method("GET")
	mux.HandleFunc("/xyz/", xyzPostHandler).Method("POST")

	// OPEN - exact match
	mux.HandleFunc("/api/srp_register", respHandlerSRPRegister).Method("GET", "POST")
	mux.HandleFunc("/api/srp_simulate_email_confirm", respHandlerSimulateEmailConfirm).Method("GET", "POST")
	mux.HandleFunc("/api/srp_email_confirm", respHandlerEmailConfirm).Method("GET", "POST")
	mux.HandleFunc("/api/srp_login", respHandlerSRPLogin).Method("GET", "POST")
	mux.HandleFunc("/api/srp_challenge", respHandlerSRPChallenge).Method("GET", "POST")
	mux.HandleFunc("/api/srp_validate", respHandlerSRPValidate).Method("GET", "POST")
	mux.HandleFunc("/api/srp_getNg", respHandlerSRPGetNg).Method("GET", "POST")
	mux.HandleFunc("/api/srp_recover_password_pt1", respHandlerRecoverPasswordPt1).Method("GET", "POST")
	mux.HandleFunc("/api/srp_recover_password_pt2", respHandlerRecoverPasswordPt2).Method("GET", "POST")

	// Login Required - exact match
	mux.HandleFunc("/api/srp_change_password", respHandlerChangePassword).Method("GET", "POST")
	mux.HandleFunc("/api/srp_admin_set_password", respHandlerAdminSetPassword).Method("GET", "POST")
	mux.HandleFunc("/api/srp_logout", respHandlerSRPLogout).Method("GET", "POST")
	mux.HandleFunc("/api/cipher", respHandlerCipher).Method("GET", "POST")
	// xyzzy - admin disable account
	// xyzzy - admin enable account

	mux.HandleErrors(404, httpmux.HandlerFunc(myErrHandler))

	mux.CompilePatternMatcher()

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
	fmt.Fprint(w, "welcome bob - t2 ")
}

func aaaHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "aaaHandler URI = %s\n", r.RequestURI)
}

func xyzGetHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "xyzGetHandler URI = %s\n", r.RequestURI)
}

func xyzPostHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "xyzPostHandler URI = %s\n", r.RequestURI)
}

func myErrHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	fmt.Fprint(w, "custom 404")
}

// ----------------------------------------------------------------------------------------------------------------------------
func common(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "common URI = %s, %s\n", r.RequestURI, tr.LF(2))
}

func respHandlerSRPRegister(ww http.ResponseWriter, req *http.Request)          { common(ww, req) }
func respHandlerSimulateEmailConfirm(ww http.ResponseWriter, req *http.Request) { common(ww, req) }
func respHandlerEmailConfirm(ww http.ResponseWriter, req *http.Request)         { common(ww, req) }
func respHandlerSRPLogin(ww http.ResponseWriter, req *http.Request)             { common(ww, req) }
func respHandlerSRPChallenge(ww http.ResponseWriter, req *http.Request)         { common(ww, req) }
func respHandlerSRPValidate(ww http.ResponseWriter, req *http.Request)          { common(ww, req) }
func respHandlerSRPGetNg(ww http.ResponseWriter, req *http.Request)             { common(ww, req) }
func respHandlerRecoverPasswordPt1(ww http.ResponseWriter, req *http.Request)   { common(ww, req) }
func respHandlerRecoverPasswordPt2(ww http.ResponseWriter, req *http.Request)   { common(ww, req) }
func respHandlerChangePassword(ww http.ResponseWriter, req *http.Request)       { common(ww, req) }
func respHandlerAdminSetPassword(ww http.ResponseWriter, req *http.Request)     { common(ww, req) }
func respHandlerSRPLogout(ww http.ResponseWriter, req *http.Request)            { common(ww, req) }
func respHandlerCipher(ww http.ResponseWriter, req *http.Request)               { common(ww, req) }
