// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Modifications Copyright 2015-2016 Philip Schlump under the same license.

// HTTP server.  See RFC 2616.

// Taken from "server.go" in go library soruce and
// modified.   - Marked with "PJS" comments.

/*
TODO:
	*1. Add in capability to match "Method"
	*2. Change how to handle "NotFound" to be a specific return or callback state
	*4. Build a little test code to verify - with error handler etc.
	+5.  map[string]MuxEntry //	Xyzzy0001 - need to have array of MuxEntry at this point.
	*7. // xyzzy0005 - just fix the URI - if end in '/' don't be pedantic.
		*1. Clean URL before lookup
		*2. NO redirect!


	use map[string/key] -> []Match instead of hash



	8. // xyzzy0009 - need to check for anyBleow!

	3. Pull in tests from test code:e /usr/local/go/src/net/http/serve_test.go

	5. Preformance
		BuidlKey ( r ) -> [length] + HTTP(1/1.1/2) + https + GET + Host + Port + URI + AnyBelowFlag ===>>> Hash lookup
									 3               2       7     (str)  (int)  (str) 2
			Work Long to Short Search
Later:
	***simple***
	6. errorHandlerFx Handler             // xyzzy0002 - change this to lookup map[int]Handler - for errors by code -- use 'mu' to lock
		// xyzzy0007 - convert to lookup in hash table

*/

package httpmux

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"sync"

	tr "github.com/pschlump/godebug"
)

// Objects implementing the Handler interface can be
// registered to serve a particular path or subtree
// in the HTTP server.
//
// ServeHTTP should write reply headers and data to the http.ResponseWriter
// and then return.  Returning signals that the request is finished
// and that the HTTP server can move on to the next request on
// the connection.
//
// If ServeHTTP panics, the server (the caller of ServeHTTP) assumes
// that the effect of the panic was isolated to the active request.
// It recovers the panic, logs a stack trace to the server error log,
// and hangs up the connection.
//
type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as HTTP handlers.  If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler object that calls f.
type HandlerFunc func(http.ResponseWriter, *http.Request)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(w, r)
}

// Helper handlers

// Error replies to the request with the specified error message and HTTP code.
// The error message should be plain text.
func Error(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, error)
}

// NotFound replies to the request with an HTTP 404 not found error.
func NotFound(w http.ResponseWriter, r *http.Request) {
	Error(w, "404 page not found", http.StatusNotFound)
}

// NotFoundHandler returns a simple request handler
// that replies to each request with a ``404 page not found'' reply.
func NotFoundHandler() Handler { return HandlerFunc(NotFound) }

// ServeMux is an HTTP request multiplexer.
// It matches the URL of each incoming request against a list of registered
// patterns and calls the handler for the pattern that
// most closely matches the URL.
//
// Patterns name fixed, rooted paths, like "/favicon.ico",
// or rooted subtrees, like "/images/" (note the trailing slash).
// Longer patterns take precedence over shorter ones, so that
// if there are handlers registered for both "/images/"
// and "/images/thumbnails/", the latter handler will be
// called for paths beginning "/images/thumbnails/" and the
// former will receive requests for any other paths in the
// "/images/" subtree.
//
// Note that since a pattern ending in a slash names a rooted subtree,
// the pattern "/" matches all paths not matched by other registered
// patterns, not just the URL with Path == "/".
//
// Patterns may optionally begin with a host name, restricting matches to
// URLs on that host only.  Host-specific patterns take precedence over
// general patterns, so that a handler might register for the two patterns
// "/codesearch" and "codesearch.google.com/" without also taking over
// requests for "http://www.google.com/".
//
// ServeMux also takes care of sanitizing the URL request path,
// redirecting any request containing . or .. elements to an
// equivalent .- and ..-free URL.
type ServeMux struct {
	mu             sync.RWMutex          //
	mm             map[string][]MuxEntry //	Xyzzy0001 - need to have array of MuxEntry at this point.
	errorHandlerFx Handler               // xyzzy0002 - change this to lookup map[int]Handler - for errors by code -- use 'mu' to lock
	hostUsed       bool                  // whether any patterns contain hostnames
	initFlag       bool                  //
	methodUsed     bool                  // Some stuff diferentiated by Method
	belowUsed      bool                  // Some stuff diferentiated by Method
}

/*

Search based on Length (1) with possible wild cards
	/abc/def/ghi -> /abc/def/ghi				3
	/abc/def/ghi -> /abc/def/ghi/.*
	/abc/def/ghi -> /abc/def					2
	/abc/def/ghi -> /abc/def/.*
	/abc/def/ghi -> /abc						1
	/abc/def/ghi -> /abc/.*

*/

type MuxEntry struct {
	h        Handler
	pattern  string
	method   []string
	anyBelow bool
	// http 1.0, 1.1, 2.0, 2.0+
	// https/http
}

type addCriteriaType struct {
	pattern  string
	mux      *ServeMux
	posInArr int
}

func (m MuxEntry) String() (s string) {
	s = fmt.Sprintf("{  pattern: -->>%s<<-- method:%v anyBelow:%t }", m.pattern, m.method, m.anyBelow)
	return
}

func (mux *ServeMux) String() (s string) {
	s = fmt.Sprintf("Mux: %s\n", tr.LF(2))
	s += fmt.Sprintf("len(m) = %d\n", len(mux.mm))
	s += fmt.Sprintf("hostUsed = %t\n", mux.hostUsed)
	for ii, ww := range mux.mm {
		for jj, vv := range ww {
			s += fmt.Sprintf("   MuxEntry[%s][%d] = %s\n", ii, jj, vv)
		}
	}
	return
}

var (
	ErrNotFound = errors.New("404 Not Found")
)

// NewServeMux allocates and returns a new ServeMux.
func NewServeMux() *ServeMux {
	return &ServeMux{
		mm:             make(map[string][]MuxEntry),
		errorHandlerFx: HandlerFunc(func(w http.ResponseWriter, r *http.Request) { Error(w, "404 page not found", http.StatusNotFound) }),
	}
}

// Does path match pattern?
func pathMatch(pattern, path string) bool {
	if len(pattern) == 0 {
		// should not happen
		return false
	}
	n := len(pattern)
	if pattern[n-1] != '/' {
		return pattern == path
	}
	return len(path) >= n && path[0:n] == pattern
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
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	// if p[len(p)-1] == '/' && np != "/" {
	// 	np += "/"
	// }
	return np
}

// Matches on additional criteria like "Method"
func additionalCriteria(v MuxEntry, r *http.Request) bool {
	if db200 {
		fmt.Printf("additional: checking additional\n")
	}
	if len(v.method) > 0 {
		if db200 {
			fmt.Printf("additional have v.method\n")
		}
		for _, w := range v.method {
			if db200 {
				fmt.Printf("additional: Compare to [%s], request [%s]\n", w, r.Method)
			}
			if r.Method == w {
				if db200 {
					fmt.Printf("additional: Matched ==================== yea! ====================\n")
				}
				return true
			}
		}
		if db200 {
			fmt.Printf("additional: *** failed to match ***\n")
		}
		return false
	}
	if db200 {
		// xyzzy0009 - need to check for anyBleow!
		fmt.Printf("additional: Matched - test not used\n")
	}
	return true
}

// Find a handler on a handler map given a path string
// Most-specific (longest) pattern wins
func (mux *ServeMux) match(path string, r *http.Request) (h Handler, pattern string) {
	if db200 {
		fmt.Printf("\nmatch: top path= -->>%s<<--\n", path)
	}
	n := 0
	for k, v := range mux.mm {
		if db200 {
			fmt.Printf("match: k=%v v=%v\n", k, v)
		}
		if !pathMatch(k, path) {
			if db200 {
				fmt.Printf("match: not match to path[%s] -- continue \n", path)
			}
			continue
		}
		for _, vv := range v {
			if additionalCriteria(vv, r) {
				if db200 {
					fmt.Printf("match: match on additiona critera -- mtached \n")
				}
				if h == nil || len(k) > n { // search all - linear search - and pick longest!
					if db200 {
						fmt.Printf("match: if at bottom\n")
					}
					n = len(k)
					h = vv.h
					pattern = vv.pattern
				}
			}
		}
	}
	if db200 {
		fmt.Printf("match: return\n\n")
	}
	return
}

func (mux *ServeMux) encodeMethod(mt string) (s []byte) {
	if len(mt) < 3 {
		return []byte("z")
	}
	var c byte = (((mt[0] << 1) ^ mt[1] ^ mt[2]) + ' ') & 0x7F
	s = append(s, c)
	// fmt.Printf("C = %x S >%s< >%x<\n", c, s, s)
	return
}

func (mux *ServeMux) genKey(path string, r *http.Request) (s string) {
	if mux.methodUsed {
		s += string(mux.encodeMethod(r.Method))
	}
	if mux.belowUsed {
		s += "b"
	}
	if mux.hostUsed { // this says "EVERY TIME use the host!"
		s += r.Host
	}
	s += path
	return
}

// Handler returns the handler to use for the given request,
// consulting r.Method, r.Host, and r.URL.Path. It always returns
// a non-nil handler. If the path is not in its canonical form, the
// handler will be an internally-generated handler that redirects
// to the canonical path.
//
// Handler also returns the registered pattern that matches the
// request or, in the case of internally-generated redirects,
// the pattern that will match after following the redirect.
//
// If there is no registered handler that applies to the request,
// Handler returns a ``page not found'' handler and an empty pattern.
func (mux *ServeMux) Handler(r *http.Request) (h Handler, pattern string, err error) {
	mux.CompilePatternMatcher()
	if r.Method != "CONNECT" {
		r.URL.Path = CleanPath(r.URL.Path)
	}

	if db200 {
		fmt.Printf("Key: %s\n", mux.genKey(r.URL.Path, r))
	}

	return mux.handler(r.Host, r.URL.Path, r)
}

// handler is the main implementation of Handler.
// The path is known to be in canonical form, except for CONNECT methods.
func (mux *ServeMux) handler(host, path string, r *http.Request) (h Handler, pattern string, err error) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	// Host-specific pattern takes precedence over generic ones
	if mux.hostUsed {
		h, pattern = mux.match(host+path, r)
	}
	if h == nil {
		h, pattern = mux.match(path, r)
	}
	// PJS - This is the "not-found" stuff -- I added an error return so that you can
	// call mux.Hanler and get back the handler and implement your own ServeHTTP
	// in a decent fashion.
	if h == nil {
		h, pattern = HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mux.errorHandlerFx.ServeHTTP(w, r)
		}), ""
		err = ErrNotFound
	}
	return
}

// ServeHTTP dispatches the request to the handler whose
// pattern most closely matches the request URL.
func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "*" {
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h, _, err := mux.Handler(r)
	_ = err
	h.ServeHTTP(w, r)
}

// Handle registers the handler for the given pattern.
// If a handler already exists for pattern, Handle panics.
func (mux *ServeMux) Handle(pattern string, handler Handler) (my *addCriteriaType) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if handler == nil {
		handler = HandlerFunc(func(w http.ResponseWriter, r *http.Request) { Error(w, "404 page not found", http.StatusNotFound) })
	}

	ab := false
	n := len(pattern)
	if n > 0 && pattern[n-1] == '/' {
		ab = true
	}

	pattern = CleanPath(pattern)

	mux.mm[pattern] = mux.mm[pattern]
	pos := len(mux.mm[pattern])
	my = &addCriteriaType{mux: mux, pattern: pattern, posInArr: pos}
	mux.mm[pattern] = append(mux.mm[pattern], MuxEntry{h: handler, pattern: pattern, anyBelow: ab})

	if len(pattern) > 0 && pattern[0] != '/' {
		mux.hostUsed = true
	}

	return
}

// HandleFunc registers the handler function for the given pattern.
func (mux *ServeMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) *addCriteriaType {
	return mux.Handle(pattern, HandlerFunc(handler))
}

func (aa *addCriteriaType) Method(mdthodArray ...string) *addCriteriaType {
	aa.mux.methodUsed = true
	mt, ok := aa.mux.mm[aa.pattern]
	if !ok {
		if db200 {
			fmt.Printf("failed ot find pattern in table - something wrong pattern=[%s]\nmux=%s\n", aa.pattern, aa.mux)
		}
		return aa
	}
	mx := mt[aa.posInArr] // xyzzy - check in range!
	for _, v := range mdthodArray {
		// fmt.Printf("Setting Method: %s\n", v)
		mx.method = append(mx.method, v)
	}
	aa.mux.mm[aa.pattern][aa.posInArr] = mx
	return aa
}

// www.WriteHeader(http.StatusForbidden)
// exactMatcher.HandleErrors ( errorHandlerFunc)
func (mux *ServeMux) HandleErrors(status int, h Handler) {
	// xyzzy0007 - convert to lookup in hash table
	mux.errorHandlerFx = h
}

// This should be called when you are done adding or changing the routes.  This will process the
// routs into the necessary data for running matching.  It gets called automatically if you forget -
// however that means that any errors in processing will not show up until the first request is
// made.  Better to just call it yourself curing initialization.
func (mux *ServeMux) CompilePatternMatcher() {
	if !mux.initFlag {
		mux.initFlag = true
		if db200 {
			fmt.Printf("It is done now %s\n", mux)
		}
	} else {
		if db200 {
			fmt.Printf("subsiquent req %s\n", mux)
		}
	}
}

const db200 = false
