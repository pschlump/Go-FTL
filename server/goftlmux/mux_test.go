package goftlmux

// Copyright 2012 The GoGoMux Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Some of the code in this is derived from Gorilla Mux and HttpRouter.

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	// "./context" // "github.com/gorilla/context"

	"github.com/pschlump/json" //	"encoding/json"

	debug "github.com/pschlump/godebug"
)

type routeTestNew struct {
	title          string            // title of the test
	request        *http.Request     // a request to test the route
	vars           map[string]string // the expected vars of the match
	host           string            // the expected host of the match
	path           string            // the expected path of the match
	route          string            // the route pattern we are creating
	shouldMatch    bool              // whether the request is expected to match the route at all
	shouldRedirect bool              // whether the request should result in a redirect
	DId            int               // Used for testing
	DName          string            // Used to identify a route by name
	DPath          string            // Set by Handler("/path",Fx), Path(), PathPrefix()
	DPathPrefix    string            //
	DHandlerFunc   http.Handler      //
	DHeaders       []string          // Set by Headers()
	DHost          string            // Set by Host()
	DMethods       []string          // Set by Methods()
	DSchemes       []string          // Set by Schemes()
	DQueries       []string          // Set by Queries()
	DProtocal      []string          //
	//route          *ARoute           // the route being tested
}

var data_DataCollectBeforeCompile = []routeTestNew{
	{
		title:    "Setting up routes",
		route:    "/abc/def",
		DMethods: []string{"GET", "PUT"},
	},
	{
		title: "Setting up routes",
		route: "/abc/ghi",
		DHost: "pschlump.2c-why.com",
	},
}

var called = 0
var g_params string
var g_route_i int

func rptCalled(www http.ResponseWriter, req *http.Request) {
	called = 1
	g_route_i = 0
	if rw, ok := www.(*MidBuffer); ok {
		// fmt.Printf("%s, AT %s\n", rw.Ps.DumpParam(), debug.LF())
		g_params = rw.Ps.DumpParam()
		g_route_i = rw.Ps.route_i
	}
}

func (r *MuxRouter) dumpTest() {
	for i := 0; i < len(r.routes); i++ {
		fmt.Printf("%3d: %s %s %s %s\n", i,
			debug.SVar(r.routes[i].DMethods),
			debug.SVar(r.routes[i].DSchemes),
			r.routes[i].DHost,
			r.routes[i].DPath)
	}
}

const db_sr1 = true
const db_sr2 = true

func Test_SimpleRoute01(t *testing.T) {
	if false {
		fmt.Printf("At Top, %s\n", debug.LF())
	}

	r := NewRouter()

	for ii, test := range data_DataCollectBeforeCompile {
		_, _ = ii, test
		// testRoute(t, test)
		x := r.HandleFunc(test.route, rptCalled)
		if test.DHost != "" {
			x.Host(test.DHost)
		}
		if len(test.DMethods) > 0 {
			x.Methods(test.DMethods...)
		} else {
			x.Methods("GET")
		}
	}
	r.setDefaults()
	r.buildRoutingTable()
	if db_sr1 {
		fmt.Printf("AT: %s\n", debug.LF())
		r.dumpTest()
	}
	if r.routes[0].DPath != "/abc/def" {
		t.Errorf("Expected /abc/def\n")
	}
	if len(r.routes[0].DMethods) != 2 {
		t.Errorf("Expected 2, got %d == %s\n", len(r.routes[0].DMethods), debug.SVar(r.routes[0].DMethods))
	}
	if r.routes[1].DPath != "/abc/ghi" {
		t.Errorf("Expected /abc/def\n")
	}
	if len(r.routes[1].DMethods) != 1 {
		t.Errorf("Expected 1, got %d == %s\n", len(r.routes[1].DMethods), debug.SVar(r.routes[1].DMethods))
	}

	r.AttachWidget(Before, ParseQueryParams)
	r.AttachWidget(Before, MethodParam)          // 15ns
	r.AttachWidget(Before, ParseBodyAsParams)    // 27ns
	r.AttachWidget(Before, ParseCookiesAsParams) // 28ns
	if true {
		r.AttachWidget(Before, ApacheLogingBefore) // 17ns
		r.AttachWidget(After, ApacheLogingAfter)   // 1 alloc + 475ns - Caused by format of time
	}

	r.CompileRoutes()
	if db_sr2 {
		fmt.Printf("AT: %s\n", debug.LF())
		r.OutputStatusInfo()
	}

	// Check Routing - simple check.
	var req http.Request
	// var w http.ResponseWriter
	var url url.URL
	// var tls tls.ConnectionState
	disableOutput = true
	w := new(mockResponseWriter)
	www := NewMidBuffer(w, nil)
	req.URL = &url
	req.URL.Path = "/abc/def"
	req.URL.RawQuery = "id=12"
	req.Method = "GET"
	req.Host = "localhost:8080"
	req.TLS = nil
	req.RemoteAddr = "[::1]:53248"
	if req.URL.RawQuery != "" {
		req.RequestURI = req.URL.Path + "?" + req.URL.RawQuery
	} else {
		req.RequestURI = req.URL.Path
	}
	req.Proto = "HTTP/1.1"
	req.Header = make(http.Header)

	called = 0
	r.ServeHTTP(www, &req)
	if called != 1 {
		t.Errorf("Test: Expected to have handler called. %d\n", called)
	}

	expected_params := `[{"Name":"id","Value":"12","From":1,"Type":113}]`
	if g_params != expected_params {
		t.Errorf("Test: Expected params ->%s<- got ->%s<-\n", expected_params, g_params)
	}
}

// -----------------------------------------------------------------------------------------------------
// Fake Writer so that lack of a writer during tests will not result in a core dump.
// Just think the author of this software is so old that he has actually written programs
// that were loaded into "core" memory (The little tiny magnets)!
type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockResponseWriter) WriteHeader(int) {}

// -----------------------------------------------------------------------------------------------------
// This little dodad outputs info on colisions and routes.  Remember that GET /abc/def is a completely
// different route than POST /abc/def and will show up in a different lcoaiton in the hash table.
func (r *MuxRouter) OutputStatusInfo() {
	nc := 0
	fmt.Printf("AT %s\n", debug.LF())
	fmt.Printf("[%4s] %6s %s\n", "", "cType", ".cType flags")
	fmt.Printf("%6s %6s %s\n", "------", "------", "---------------------------")
	for i, v := range r.Hash2Test {
		if v > 0 {
			fmt.Printf("[%4d] 0x%04x %s\n", i, r.LookupResults[v].cType, dumpCType(r.LookupResults[v].cType))
			if (r.LookupResults[v].cType & MultiUrl) != 0 {
				nc++
				for j, w := range r.LookupResults[v].Multi {
					ns := numChar(j, '/')
					fmt.Printf("   %2d %s\n", ns, j)
					_ = w
				}
			} else {
				fmt.Printf("   %2s %s\n", "", r.LookupResults[v].Url)
			}
		}
	}
	fmt.Printf("\nNumber of Collisions = %d\n\n", nc)
}

// -------------------------------------------------------------------------------------------------

var test2017Data = []struct {
	LoadUrl bool
	Method  string
	Url     string
	Result  int
}{
	{true, "GET", "/planb/:vA/t1/:vB", 2},                 // 2
	{true, "GET", "/planb/:vD/t2/:vE", 3},                 // 3
	{true, "GET", "/planb/:vD/t2/xx", 4},                  // 4
	{true, "GET", "/planb/:vD/t8/{yy:^[0-9][0-9]*$}", 10}, // 4
	{true, "GET", "/planb/:vD/t8/bob", 11},                // 4
	{true, "GET", "/planb/:vD/t9/:vX", 12},                // 4
	{true, "GET", "/planb/:vD/t9/bob", 14},                // 4
	{true, "GET", "/planb/:vF/t3", 5},                     // 5
	{true, "GET", "/planb/:vG/t4", 6},                     // 6
	{true, "GET", "/planb/x3/t5", 7},                      // 7
	{true, "GET", "/planb/:vC", 1},                        // 1
	{true, "GET", "/planE/:vC", 20},                       // 20
	{true, "GET", "/users/:user/received_events", 9},      // 9
	{true, "GET", "/*vG", 8},                              // 8
	{true, "GET", "/rc/{yy:^[0-9][0-9]*$}", 15},           // 15
	{true, "GET", "/rc/:zz", 17},                          // 17
	{true, "GET", "/rc/dave", 16},                         // 16
	{true, "GET", "/rd/:z2", 18},                          // 18
	{true, "GET", "/re/{z2}", 19},                         // 19
	{true, "GET", "/js/*filename", 21},                    // 21
	{true, "GET", "/img/*filename", 22},                   // 22
	{true, "GET", "/css/*filename", 23},                   // 23
	{true, "GET", "/abc/*p1/:p2", 24},                     // 24 // test with /abc/*p1/:k2 - should be a bad pattern - should not work - should break loop at this point

	{true, "GET", "/authorizations", 31},
	{true, "GET", "/authorizations/:id", 32},
	{true, "POST", "/authorizations", 33},
	{true, "PUT", "/authorizations/clients/:client_id", 34},
	{true, "PATCH", "/authorizations/:id", 35},
	{true, "DELETE", "/authorizations/:id", 36},
	{true, "GET", "/applications/:client_id/tokens/:access_token", 37},
	{true, "DELETE", "/applications/:client_id/tokens", 38},
	{true, "DELETE", "/applications/:client_id/tokens/:access_token", 39},

	{true, "GET", "/events", 41},
	{true, "GET", "/repos/:owner/:repo/events", 42},
	{true, "GET", "/networks/:owner/:repo/events", 43},
	{true, "GET", "/orgs/:org/events", 44},
	{true, "GET", "/users/:user/received_events", 45},
	{true, "GET", "/users/:user/received_events/public", 46},
	{true, "GET", "/users/:user/events", 47},
	{true, "GET", "/users/:user/events/public", 48},
	{true, "GET", "/users/:user/events/orgs/:org", 49},
	{true, "GET", "/feeds", 50},
	{true, "GET", "/notifications", 51},
	{true, "GET", "/repos/:owner/:repo/notifications", 52},
	{true, "PUT", "/notifications", 53},
	{true, "PUT", "/repos/:owner/:repo/notifications", 54},
	{true, "GET", "/notifications/threads/:id", 55},
	{true, "PATCH", "/notifications/threads/:id", 56},
	{true, "GET", "/notifications/threads/:id/subscription", 57},
	{true, "PUT", "/notifications/threads/:id/subscription", 58},
	{true, "DELETE", "/notifications/threads/:id/subscription", 59},
	{true, "GET", "/repos/:owner/:repo/stargazers", 60},
	{true, "GET", "/users/:user/starred", 61},
	{true, "GET", "/user/starred", 62},
	{true, "GET", "/user/starred/:owner/:repo", 63},
	{true, "PUT", "/user/starred/:owner/:repo", 64},
	{true, "DELETE", "/user/starred/:owner/:repo", 65},
	{true, "GET", "/repos/:owner/:repo/subscribers", 66},
	{true, "GET", "/users/:user/subscriptions", 67},
	{true, "GET", "/user/subscriptions", 68},
	{true, "GET", "/repos/:owner/:repo/subscription", 69},
	{true, "PUT", "/repos/:owner/:repo/subscription", 70},
	{true, "DELETE", "/repos/:owner/:repo/subscription", 71},
	{true, "GET", "/user/subscriptions/:owner/:repo", 72},
	{true, "PUT", "/user/subscriptions/:owner/:repo", 73},
	{true, "DELETE", "/user/subscriptions/:owner/:repo", 74},

	{true, "GET", "/users/:user/gists", 76},
	{true, "GET", "/gists", 77},
	{true, "GET", "/gists/public", 78},
	{true, "GET", "/gists/starred", 79},
	{true, "GET", "/gists/:id", 80},
	{true, "POST", "/gists", 81},
	{true, "PATCH", "/gists/:id", 82},
	{true, "PUT", "/gists/:id/star", 83},
	{true, "DELETE", "/gists/:id/star", 84},
	{true, "GET", "/gists/:id/star", 85},
	{true, "POST", "/gists/:id/forks", 86},
	{true, "DELETE", "/gists/:id", 87},

	{true, "GET", "/repos/:owner/:repo/git/blobs/:sha", 89},
	{true, "POST", "/repos/:owner/:repo/git/blobs", 90},
	{true, "GET", "/repos/:owner/:repo/git/commits/:sha", 91},
	{true, "POST", "/repos/:owner/:repo/git/commits", 92},
	{true, "GET", "/repos/:owner/:repo/git/refs/*ref", 93},
	{true, "GET", "/repos/:owner/:repo/git/refs", 94},
	{true, "POST", "/repos/:owner/:repo/git/refs", 95},
	{true, "PATCH", "/repos/:owner/:repo/git/refs/*ref", 96},
	{true, "DELETE", "/repos/:owner/:repo/git/refs/*ref", 97},
	{true, "GET", "/repos/:owner/:repo/git/tags/:sha", 98},
	{true, "POST", "/repos/:owner/:repo/git/tags", 99},
	{true, "GET", "/repos/:owner/:repo/git/trees/:sha", 100},
	{true, "POST", "/repos/:owner/:repo/git/trees", 101},

	{true, "GET", "/issues", 103},
	{true, "GET", "/user/issues", 104},
	{true, "GET", "/orgs/:org/issues", 105},
	{true, "GET", "/repos/:owner/:repo/issues", 106},
	{true, "GET", "/repos/:owner/:repo/issues/:number", 107},
	{true, "POST", "/repos/:owner/:repo/issues", 108},
	{true, "PATCH", "/repos/:owner/:repo/issues/:number", 109},
	{true, "GET", "/repos/:owner/:repo/assignees", 110},
	{true, "GET", "/repos/:owner/:repo/assignees/:assignee", 111},
	{true, "GET", "/repos/:owner/:repo/issues/:number/comments", 112},
	{true, "GET", "/repos/:owner/:repo/issues/comments", 113},
	{true, "GET", "/repos/:owner/:repo/issues/comments/:id", 114},
	{true, "POST", "/repos/:owner/:repo/issues/:number/comments", 115},
	{true, "PATCH", "/repos/:owner/:repo/issues/comments/:id", 116},
	{true, "DELETE", "/repos/:owner/:repo/issues/comments/:id", 117},
	{true, "GET", "/repos/:owner/:repo/issues/:number/events", 118},
	{true, "GET", "/repos/:owner/:repo/issues/events", 119},
	{true, "GET", "/repos/:owner/:repo/issues/events/:id", 120},
	{true, "GET", "/repos/:owner/:repo/labels", 121},
	{true, "GET", "/repos/:owner/:repo/labels/:name", 122},
	{true, "POST", "/repos/:owner/:repo/labels", 123},
	{true, "PATCH", "/repos/:owner/:repo/labels/:name", 124},
	{true, "DELETE", "/repos/:owner/:repo/labels/:name", 125},
	{true, "GET", "/repos/:owner/:repo/issues/:number/labels", 126},
	{true, "POST", "/repos/:owner/:repo/issues/:number/labels", 127},
	{true, "DELETE", "/repos/:owner/:repo/issues/:number/labels/:name", 128},
	{true, "PUT", "/repos/:owner/:repo/issues/:number/labels", 129},
	{true, "DELETE", "/repos/:owner/:repo/issues/:number/labels", 130},
	{true, "GET", "/repos/:owner/:repo/milestones/:number/labels", 131},
	{true, "GET", "/repos/:owner/:repo/milestones", 132},
	{true, "GET", "/repos/:owner/:repo/milestones/:number", 133},
	{true, "POST", "/repos/:owner/:repo/milestones", 134},
	{true, "PATCH", "/repos/:owner/:repo/milestones/:number", 135},
	{true, "DELETE", "/repos/:owner/:repo/milestones/:number", 136},

	{true, "GET", "/emojis", 138},
	{true, "GET", "/gitignore/templates", 139},
	{true, "GET", "/gitignore/templates/:name", 140},
	{true, "POST", "/markdown", 141},
	{true, "POST", "/markdown/raw", 142},
	{true, "GET", "/meta", 143},
	{true, "GET", "/rate_limit", 144},

	{true, "GET", "/users/:user/orgs", 146},
	{true, "GET", "/user/orgs", 147},
	{true, "GET", "/orgs/:org", 148},
	{true, "PATCH", "/orgs/:org", 149},
	{true, "GET", "/orgs/:org/members", 150},
	{true, "GET", "/orgs/:org/members/:user", 151},
	{true, "DELETE", "/orgs/:org/members/:user", 152},
	{true, "GET", "/orgs/:org/public_members", 153},
	{true, "GET", "/orgs/:org/public_members/:user", 154},
	{true, "PUT", "/orgs/:org/public_members/:user", 155},
	{true, "DELETE", "/orgs/:org/public_members/:user", 156},
	{true, "GET", "/orgs/:org/teams", 157},
	{true, "GET", "/teams/:id", 158},
	{true, "POST", "/orgs/:org/teams", 159},
	{true, "PATCH", "/teams/:id", 160},
	{true, "DELETE", "/teams/:id", 161},
	{true, "GET", "/teams/:id/members", 162},
	{true, "GET", "/teams/:id/members/:user", 163},
	{true, "PUT", "/teams/:id/members/:user", 164},
	{true, "DELETE", "/teams/:id/members/:user", 165},
	{true, "GET", "/teams/:id/repos", 166},
	{true, "GET", "/teams/:id/repos/:owner/:repo", 167},
	{true, "PUT", "/teams/:id/repos/:owner/:repo", 168},
	{true, "DELETE", "/teams/:id/repos/:owner/:repo", 169},
	{true, "GET", "/user/teams", 170},

	{true, "GET", "/repos/:owner/:repo/pulls", 172},
	{true, "GET", "/repos/:owner/:repo/pulls/:number", 173},
	{true, "POST", "/repos/:owner/:repo/pulls", 174},
	{true, "PATCH", "/repos/:owner/:repo/pulls/:number", 175},
	{true, "GET", "/repos/:owner/:repo/pulls/:number/commits", 176},
	{true, "GET", "/repos/:owner/:repo/pulls/:number/files", 177},
	{true, "GET", "/repos/:owner/:repo/pulls/:number/merge", 178},
	{true, "PUT", "/repos/:owner/:repo/pulls/:number/merge", 179},
	{true, "GET", "/repos/:owner/:repo/pulls/:number/comments", 180},
	{true, "GET", "/repos/:owner/:repo/pulls/comments", 181},
	{true, "GET", "/repos/:owner/:repo/pulls/comments/:number", 182},
	{true, "PUT", "/repos/:owner/:repo/pulls/:number/comments", 183},
	{true, "PATCH", "/repos/:owner/:repo/pulls/comments/:number", 184},
	{true, "DELETE", "/repos/:owner/:repo/pulls/comments/:number", 185},

	{true, "GET", "/user/repos", 187},
	{true, "GET", "/users/:user/repos", 188},
	{true, "GET", "/orgs/:org/repos", 189},
	{true, "GET", "/repositories", 190},
	{true, "POST", "/user/repos", 191},
	{true, "POST", "/orgs/:org/repos", 192},
	{true, "GET", "/repos/:owner/:repo", 193},
	{true, "PATCH", "/repos/:owner/:repo", 194},
	{true, "GET", "/repos/:owner/:repo/contributors", 195},
	{true, "GET", "/repos/:owner/:repo/languages", 196},
	{true, "GET", "/repos/:owner/:repo/teams", 197},
	{true, "GET", "/repos/:owner/:repo/tags", 198},
	{true, "GET", "/repos/:owner/:repo/branches", 199},
	{true, "GET", "/repos/:owner/:repo/branches/:branch", 200},
	{true, "DELETE", "/repos/:owner/:repo", 201},
	{true, "GET", "/repos/:owner/:repo/collaborators", 202},
	{true, "GET", "/repos/:owner/:repo/collaborators/:user", 203},
	{true, "PUT", "/repos/:owner/:repo/collaborators/:user", 204},
	{true, "DELETE", "/repos/:owner/:repo/collaborators/:user", 205},
	{true, "GET", "/repos/:owner/:repo/comments", 206},
	{true, "GET", "/repos/:owner/:repo/commits/:sha/comments", 207},
	{true, "POST", "/repos/:owner/:repo/commits/:sha/comments", 208},
	{true, "GET", "/repos/:owner/:repo/comments/:id", 209},
	{true, "PATCH", "/repos/:owner/:repo/comments/:id", 210},
	{true, "DELETE", "/repos/:owner/:repo/comments/:id", 211},
	{true, "GET", "/repos/:owner/:repo/commits", 212},
	{true, "GET", "/repos/:owner/:repo/commits/:sha", 213},
	{true, "GET", "/repos/:owner/:repo/readme", 214},
	{true, "GET", "/repos/:owner/:repo/contents/*path", 215},
	{true, "PUT", "/repos/:owner/:repo/contents/*path", 216},
	{true, "DELETE", "/repos/:owner/:repo/contents/*path", 217},
	{true, "GET", "/repos/:owner/:repo/:archive_format/:ref", 218},
	{true, "GET", "/repos/:owner/:repo/keys", 219},
	{true, "GET", "/repos/:owner/:repo/keys/:id", 220},
	{true, "POST", "/repos/:owner/:repo/keys", 221},
	{true, "PATCH", "/repos/:owner/:repo/keys/:id", 222},
	{true, "DELETE", "/repos/:owner/:repo/keys/:id", 223},
	{true, "GET", "/repos/:owner/:repo/downloads", 224},
	{true, "GET", "/repos/:owner/:repo/downloads/:id", 225},
	{true, "DELETE", "/repos/:owner/:repo/downloads/:id", 226},
	{true, "GET", "/repos/:owner/:repo/forks", 227},
	{true, "POST", "/repos/:owner/:repo/forks", 228},
	{true, "GET", "/repos/:owner/:repo/hooks", 229},
	{true, "GET", "/repos/:owner/:repo/hooks/:id", 230},
	{true, "POST", "/repos/:owner/:repo/hooks", 231},
	{true, "PATCH", "/repos/:owner/:repo/hooks/:id", 232},
	{true, "POST", "/repos/:owner/:repo/hooks/:id/tests", 233},
	{true, "DELETE", "/repos/:owner/:repo/hooks/:id", 234},
	{true, "POST", "/repos/:owner/:repo/merges", 235},
	{true, "GET", "/repos/:owner/:repo/releases", 236},
	{true, "GET", "/repos/:owner/:repo/releases/:id", 237},
	{true, "POST", "/repos/:owner/:repo/releases", 238},
	{true, "PATCH", "/repos/:owner/:repo/releases/:id", 239},
	{true, "DELETE", "/repos/:owner/:repo/releases/:id", 240},
	{true, "GET", "/repos/:owner/:repo/releases/:id/assets", 241},
	{true, "GET", "/repos/:owner/:repo/stats/contributors", 242},
	{true, "GET", "/repos/:owner/:repo/stats/commit_activity", 243},
	{true, "GET", "/repos/:owner/:repo/stats/code_frequency", 244},
	{true, "GET", "/repos/:owner/:repo/stats/participation", 245},
	{true, "GET", "/repos/:owner/:repo/stats/punch_card", 246},
	{true, "GET", "/repos/:owner/:repo/statuses/:ref", 247},
	{true, "POST", "/repos/:owner/:repo/statuses/:ref", 248},

	{true, "GET", "/search/repositories", 250},
	{true, "GET", "/search/code", 251},
	{true, "GET", "/search/issues", 252},
	{true, "GET", "/search/users", 253},
	{true, "GET", "/legacy/issues/search/:owner/:repository/:state/:keyword", 254},
	{true, "GET", "/legacy/repos/search/:keyword", 255},
	{true, "GET", "/legacy/user/search/:keyword", 256},
	{true, "GET", "/legacy/user/email/:email", 257},

	{true, "GET", "/users/:user", 259},
	{true, "GET", "/user", 260},
	{true, "PATCH", "/user", 261},
	{true, "GET", "/users", 262},
	{true, "GET", "/user/emails", 263},
	{true, "POST", "/user/emails", 264},
	{true, "DELETE", "/user/emails", 265},
	{true, "GET", "/users/:user/followers", 266},
	{true, "GET", "/user/followers", 267},
	{true, "GET", "/users/:user/following", 268},
	{true, "GET", "/user/following", 269},
	{true, "GET", "/user/following/:user", 270},
	{true, "GET", "/users/:user/following/:target_user", 271},
	{true, "PUT", "/user/following/:user", 272},
	{true, "DELETE", "/user/following/:user", 273},
	{true, "GET", "/users/:user/keys", 274},
	{true, "GET", "/user/keys", 275},
	{true, "GET", "/user/keys/:id", 276},
	{true, "POST", "/user/keys", 277},
	{true, "PATCH", "/user/keys/:id", 278},
	{true, "DELETE", "/user/keys/:id", 279},

	{true, "GET", "/xyz/xyz/xyz", 300},
	{true, "GET", "/xyz01", 301},
	{true, "delete", "/bad01", 302},
	{true, "DELETE", "bad02", 303},
	{true, "GET", "/", 3001},
	{true, "GET", "/index.html", 3002},
	{true, "GET", "/index.htm", 3002},
	{true, "GET", "/default.html", 3002},
	{true, "GET", "/default.htm", 3002},
}

type NV struct {
	Type  string
	Name  string
	Value string
}

var test2017Run = []struct {
	RunTest        bool
	Method         string
	Url            string
	Expect         int
	ShouldBeFound  bool
	ExpectedParams []NV
}{
	/*  00 */ {true, "GET", "/planb/vD-Data", 1, true, []NV{NV{Type: ":", Name: "vC", Value: "vD-Data"}}},
	/*  01 */ {true, "GET", "/planb/vD-data/t2/xx", 4, true, []NV{NV{Type: ":", Name: "vD", Value: "vD-data"}}},
	/*  02 */ {true, "GET", "/planb/vD-data/t2/yy", 3, true, []NV{NV{Type: ":", Name: "vD", Value: "vD-data"}, {Type: ":", Name: "vE", Value: "yy"}}},
	/*  03 */ {true, "GET", "/planb/x3/t5", 7, true, nil},
	/*  04 */ {true, "GET", "/planb/x4/t5", 8, true, []NV{NV{Type: "*", Name: "vG", Value: "planb/x4/t5"}}}, // <<<< STAR >>>>	<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< this one >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	/*  05 */ {true, "GET", "//planb/x/.././/////x3/t5", 7, true, nil},
	/*  06 */ {true, "GET", "/planb/x/.././/////x3/t5", 7, true, nil},
	/*  07 */ {true, "GET", "/index.html", 8, true, nil},
	/*  08 */ {true, "GET", "/p/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z/a/b/c/d/e/f", 8, true, []NV{NV{Type: "*", Name: "vG",
		Value: "p/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z/a/b/c/d/e/f"}}},
	/*  09 */ {true, "GET", "/", 8, true, []NV{NV{Type: "*", Name: "vG", Value: ""}}},
	/*  10 */ {true, "GET", "/planb/:vD/t2/xx", 4, true, nil},
	/*  11 */ {true, "GET", "/planb/:vF/t3", 5, true, nil},
	/*  12 */ {true, "GET", "/planb/:vG/t4", 6, true, nil},
	/*  13 */ {true, "GET", "/planb/x3/t5", 7, true, nil},
	/*  14 */ {true, "GET", "/planb/:vC", 1, true, nil},
	/*  15 */ {true, "GET", "/*vG", 8, true, []NV{NV{Type: "*", Name: "vG", Value: "*vG"}, NV{Type: "*", Name: "vX", Value: ""}}},
	/*  16 */ {true, "GET", "/planb/vD-data/t8/11", 10, true, []NV{NV{Type: ":", Name: "vD", Value: "vD-data"}, NV{Type: "{", Name: "yy", Value: "11"}}},
	/*  17 */ {true, "GET", "/planb/vD-data/t9/bob", 14, true, []NV{NV{Type: ":", Name: "vD", Value: "vD-data"}}},
	/*  18 */ {true, "GET", "/planb/vD-data/t9/jane", 12, true, []NV{NV{Type: ":", Name: "vD", Value: "vD-data"}, NV{Type: ":", Name: "vX", Value: "jane"}}},
	/*  19 */ {true, "GET", "/rc/dave", 16, false, nil},
	/*  20 */ {true, "GET", "/rc/11", 15, true, []NV{NV{Type: "{", Name: "yy", Value: "11"}}},
	/*  21 */ {true, "GET", "/rc/jane", 17, true, nil},
	/*  22 */ {true, "GET", "/rd/jane", 18, true, []NV{NV{Type: ":", Name: "z2", Value: "jane"}}},
	/*  23 */ {true, "GET", "/re/jane", 19, true, []NV{NV{Type: ":", Name: "z2", Value: "jane"}}},
	/*  24 */ {true, "GET", "/planb/more-Data", 1, true, []NV{NV{Type: ":", Name: "vC", Value: "more-Data"}}},
	/*  25 */ {true, "GET", "/abc/cbs/bbc", 24, true, []NV{NV{Type: "*", Name: "p1", Value: "cbs/bbc"}, NV{Type: ":", Name: "p2", Value: ""}}},
	/*  26 */ {true, "GET", "/repos/cbs/bbc/assignees", 110, true, []NV{NV{Type: ":", Name: "owner", Value: "cbs"}, NV{Type: ":", Name: "repo", Value: "bbc"}}},
	/*  27 */ {true, "GET", "/users/Auser", 259, true, []NV{NV{Type: ":", Name: "user", Value: "Auser"}}},
	/*  28 */ {true, "GET", "/user", 260, true, nil},
	/*  29 */ {true, "PATCH", "/user", 261, true, nil},
	/*  30 */ {true, "GET", "/users", 262, true, nil},
	/*  31 */ {true, "GET", "/user/emails", 263, true, nil},
	/*  32 */ {true, "POST", "/user/emails", 264, true, nil},
	/*  33 */ {true, "DELETE", "/user/emails", 265, true, nil},
	/*  34 */ {true, "GET", "/users/Auser/followers", 266, true, []NV{NV{Type: ":", Name: "user", Value: "Auser"}}},
	/*  35 */ {true, "GET", "/user/followers", 267, true, nil},
	/*  36 */ {true, "GET", "/users/Auser/following", 268, true, []NV{NV{Type: ":", Name: "user", Value: "Auser"}}},
	/*  37 */ {true, "GET", "/user/following", 269, true, nil},
	/*  38 */ {true, "GET", "/user/following/Auser", 270, true, []NV{NV{Type: ":", Name: "user", Value: "Auser"}}},
	/*  39 */ {true, "GET", "/users/Auser/following/Atarget_user", 271, true, []NV{NV{Type: ":", Name: "user", Value: "Auser"}, NV{Type: ":", Name: "target_user", Value: "Atarget_user"}}},
	/*  40 */ {true, "PUT", "/user/following/Auser", 272, true, nil},
	/*  41 */ {true, "DELETE", "/user/following/Auser", 273, true, []NV{NV{Type: ":", Name: "user", Value: "Auser"}}},
	/*  42 */ {true, "GET", "/js/angular-animate.min.js", 21, false, nil},
	/*  43 */ {true, "GET", "/js/angular-animate.min.js.map", 21, false, nil},
	/*  44 */ {true, "GET", "/js/angular-cookies.min.js", 21, false, nil},
	/*  45 */ {true, "GET", "/js/angular-cookies.min.js.map", 21, false, nil},
	/*  46 */ {true, "GET", "/js/angular-loader.min.js", 21, false, nil},
	/*  47 */ {true, "GET", "/js/angular-loader.min.js.map", 21, false, nil},
	/*  48 */ {true, "GET", "/js/angular-resource.min.js", 21, false, nil},
	/*  49 */ {true, "GET", "/js/angular-resource.min.js.map", 21, false, nil},
	/*  50 */ {true, "GET", "/js/angular-route.min.js", 21, false, nil},
	/*  51 */ {true, "GET", "/js/angular-route.min.js.map", 21, false, nil},
	/*  52 */ {true, "GET", "/js/angular-sanitize.min.js", 21, false, nil},
	/*  53 */ {true, "GET", "/js/angular-sanitize.min.js.map", 21, false, nil},
	/*  54 */ {true, "GET", "/js/angular-touch.min.js", 21, false, nil},
	/*  55 */ {true, "GET", "/js/angular-touch.min.js.map", 21, false, nil},
	/*  56 */ {true, "GET", "/js/angular-translate.min.js", 21, false, nil},
	/*  57 */ {true, "GET", "/js/angular.1.2.10.js", 21, false, nil},
	/*  58 */ {true, "GET", "/js/angular.js", 21, false, nil},
	/*  59 */ {true, "GET", "/js/angular.min.js", 21, false, nil},
	/*  60 */ {true, "GET", "/js/angular.min.js.map", 21, false, nil},
	/*  61 */ {true, "GET", "/js/bootstrap-3.1.1-mod", 21, false, nil},
	/*  62 */ {true, "GET", "/js/date.js", 21, false, nil},
	/*  63 */ {true, "GET", "/js/dialog-4.2.0", 21, false, nil},
	/*  64 */ {true, "GET", "/js/dialogs.js", 21, false, nil},
	/*  65 */ {true, "GET", "/js/dialogs.min.js", 21, false, nil},
	/*  66 */ {true, "GET", "/js/jquery-1.10.2.js", 21, false, nil},
	/*  67 */ {true, "GET", "/js/jquery-1.10.2.min.js", 21, false, nil},
	/*  68 */ {true, "GET", "/js/jquery-1.10.2.min.map", 21, false, nil},
	/*  69 */ {true, "GET", "/js/jquery-1.11.0.min.js", 21, false, nil},
	/*  70 */ {true, "GET", "/js/jquery.complexify.banlist.js", 21, false, nil},
	/*  71 */ {true, "GET", "/js/jquery.complexify.js", 21, false, nil},
	/*  72 */ {true, "GET", "/js/jquery.complexify.min.js", 21, false, nil},
	/*  73 */ {true, "GET", "/js/jquery.hotkeys.js", 21, false, nil},
	/*  74 */ {true, "GET", "/js/jsoneditor.js", 21, false, nil},
	/*  75 */ {true, "GET", "/js/jsoneditor.min.js", 21, false, nil},
	/*  76 */ {true, "GET", "/js/jstorage.min.js", 21, false, nil},
	/*  77 */ {true, "GET", "/js/libmp3lame.min.js", 21, false, nil},
	/*  78 */ {true, "GET", "/js/mobile-angular-ui.js", 21, false, nil},
	/*  79 */ {true, "GET", "/js/mobile-angular-ui.min.js", 21, false, nil},
	/*  80 */ {true, "GET", "/js/moment.min.js", 21, false, nil},
	/*  81 */ {true, "GET", "/js/mp3Worker.js", 21, false, nil},
	/*  82 */ {true, "GET", "/js/ng-grid-2.0.11", 21, false, nil},
	/*  83 */ {true, "GET", "/js/recorderWorker.js", 21, false, nil},
	/*  84 */ {true, "GET", "/js/recordmp3.js", 21, false, nil},
	/*  85 */ {true, "GET", "/js/so.js", 21, false, nil},
	/*  86 */ {true, "GET", "/js/so.m4.js", 21, false, nil},
	/*  87 */ {true, "GET", "/js/ui-bootstrap-tpls-0.10.0-SNAPSHOT.js", 21, false, nil},
	/*  88 */ {true, "GET", "/js/ui-bootstrap-tpls-0.10.0-SNAPSHOT.min.js", 21, false, nil},
	/*  89 */ {true, "GET", "/css/dialogs.css", 23, false, nil},
	/*  90 */ {true, "GET", "/css/mobile-angular-ui-base.css", 23, false, nil},
	/*  91 */ {true, "GET", "/css/mobile-angular-ui-base.min.css", 23, false, nil},
	/*  92 */ {true, "GET", "/css/mobile-angular-ui-desktop.css", 23, false, nil},
	/*  93 */ {true, "GET", "/css/mobile-angular-ui-desktop.min.css", 23, false, nil},
	/*  94 */ {true, "GET", "/img/SafteyIcon-v1-114x114.png", 22, false, nil},
	/*  95 */ {true, "GET", "/img/SafteyIcon-v1-144x144.png", 22, false, nil},
	/*  96 */ {true, "GET", "/img/SafteyIcon-v1-57x57.png", 22, false, nil},
	/*  97 */ {true, "GET", "/img/SafteyIcon-v1-72x72.png", 22, false, nil},
	/*  98 */ {true, "GET", "/img/SafteyIcon-v1.png", 22, false, nil},
	/*  99 */ {true, "GET", "/img/ajax-loader-small.gif", 22, false, nil},
	/* 100 */ {true, "GET", "/img/bg_strength_gradient.jpg", 22, false, nil},
	/* 101 */ {true, "GET", "/img/checkbox_yes.png", 22, false, nil},
	/* 102 */ {true, "GET", "/img/checkbox_yes.svg", 22, false, nil},
	/* 103 */ {true, "GET", "/img/clear.gif", 22, false, nil},
	/* 104 */ {true, "GET", "/img/favicon.ico", 22, false, nil},
	/* 105 */ {true, "GET", "/img/favicon.png", 22, false, nil},
	/* 106 */ {true, "GET", "/img/icons.png", 22, false, nil},
	/* 107 */ {true, "GET", "/app.html", 8, false, nil},
	// {true, "GET", "/repos/:owner/:repo/pulls/:number/files", 177},
	/* 108 */ {true, "GET", "/repos/--owner--/--repo--/pulls/--number--/files", 177, true,
		[]NV{
			NV{Type: ":", Name: "owner", Value: "--owner--"},
			NV{Type: ":", Name: "repo", Value: "--repo--"},
			NV{Type: ":", Name: "number", Value: "--number--"},
		}},
	/* 109 */ {true, "GET", "/gists", 77, true, nil},
	/* 110 */ {true, "HEAD", "/r1", 7, true, nil},
	/* 111 */ {true, "GET", "/r2/4/4", 11, true, nil},
	/* 112 */ {true, "GET", "/r2/a/4", 10, true, nil},
	/* 113 */ {true, "GET", "/r2/A/4", 12, true, nil},
	/* 114 */ {true, "GET", "/", 12, true, nil},
	/* 115 */ {true, "GET", "/index.html", 12, true, nil},
	/* 116 */ {true, "GET", "/index.htm", 12, true, nil},
	/* 117 */ {true, "GET", "/?id=112&testNo=117", 12, true,
		[]NV{
			NV{Type: "?", Name: "id", Value: "112"},
			NV{Type: "?", Name: "testNo", Value: "117"},
		}},
	/* 118 */ {true, "GET", "/repos/:owner/:repo/collaborators", 4008, false, nil},
}

var rpTest string
var rpTest2 string

func rptParams(w http.ResponseWriter, r *http.Request, ps Params) {
	arrived = 4000
	s := ps.DumpParam()
	rpTest = s
	// fmt.Printf("\nrptParams: %s\n", s)
	w.Write([]byte("Hello Silly World<br>"))
}
func rptParams2(www http.ResponseWriter, r *http.Request) {
	arrived = 4008
	if rw, ok := www.(*MidBuffer); ok {
		s := rw.Ps.DumpParam()
		rw.Ps.CreateSearch()
		if db_rptParams2 {
			fmt.Printf(" *************************************************************************\n")
		}
		rpTest2 = s
		if db_rptParams2 {
			fmt.Printf("\nrptParams: %s\n", s)
		}
		done := false
		t := ""
		n := ""
		for i := 0; !done; i++ {
			n, t, done = rw.Ps.ByPostion(i)
			if db_rptParams2 {
				fmt.Printf("At [%d] %s ->%s<-\n", i, n, t)
			}
		}
		xx := rw.Ps.ByName("id")
		_ = xx
		// func (ps Params) ByPostion(pos int) ( s string, inRange bool ) {
		www.Write([]byte("Hello Silly World<br>"))
	}
}

const db_rptParams2 = false

func createFx(ii int) HandleFunc {
	return func(www http.ResponseWriter, r *http.Request) {
		var t = ii
		arrived = t
		if rw, ok := www.(*MidBuffer); ok {
			g_route_i = rw.Ps.route_i
		}
	}
}

func setupHtx() (htx *MuxRouter, trr *MuxRouterProcessing) {

	var err error
	ApacheLogFile, err = os.OpenFile("log.log", os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		ApacheLogFile, err = os.Create("log.log")
		if err != nil {
			panic(err)
		}
	}

	// var htx *MuxRouter
	var procDataHtx MuxRouterProcessing
	htx = NewRouter()
	InitMuxRouterProcessing(htx, &procDataHtx)
	trr = &procDataHtx

	fmt.Printf("Should generate 3 errors that look like:\n")
	fmt.Printf("( this is to test the error check code )\n")
	fmt.Printf("-------------------------------------------------------------------------\n")
	fmt.Printf("Error(20000): Method %%s is invalid.  Called From: %%s\n")
	fmt.Printf("Error(20002): Path should begin with '/', passed %%s, File:%%s LinLineNo:%%d\n")
	fmt.Printf("Error(20002): Path should begin with '/', passed %%s, File:%%s LinLineNo:%%d\n")
	fmt.Printf("-------------------------------------------------------------------------\n")
	fmt.Printf("\n")
	htx.AttachWidget(Before, ParseQueryParams)
	htx.AttachWidget(Before, MethodParam)          // 15ns
	htx.AttachWidget(Before, ParseBodyAsParams)    // 27ns
	htx.AttachWidget(Before, ParseCookiesAsParams) // 28ns
	if true {
		htx.AttachWidget(Before, ApacheLogingBefore) // 17ns
		htx.AttachWidget(After, ApacheLogingAfter)   // 1 alloc + 475ns - Caused by format of time
	}
	for i, test := range test2017Data {
		if test.LoadUrl {
			// fmt.Printf("[%d] %s will create test %d, %s\n", i, test.Url, i, debug.LF())
			if test.Url == "/xyz/xyz/xyz" {
				// htx.AddRoute(test.Method, test.Url, test.Result, func(w http.ResponseWriter, r *http.Request, ps Params) { j := i; arrived = 2000 + j }).SetPort("8090")
				htx.HandleFunc(test.Url, createFx(i)).Methods(test.Method).Port("8090")
			} else if test.Url == "/xyz01" {
				// htx.AddRoute(test.Method, test.Url, test.Result, func(w http.ResponseWriter, r *http.Request, ps Params) { arrived = 2000 }).SetPort("2000").SetHost("localhost:2000").SetHTTPSOnly()
				// htx.HandleFunc(test.Url, createFx(i)).Methods(test.Method).Port("2000").Host("localhost").Schemes("https")
				htx.HandleFunc(test.Url, createFx(i)).Methods(test.Method).HostPort("localhost:2000").Schemes("https")
			} else if test.Url == "/user/keys/:id" {
				// htx.AddRoute(test.Method, test.Url, test.Result, rptParams)
				htx.HandleFunc(test.Url, createFx(i)).Methods(test.Method)

			} else if test.Url == "/repos/:owner/:repo/merges" {
				// htx.AddRoute(test.Method, test.Url, test.Result, rptParams2)
				htx.HandleFunc(test.Url, createFx(i)).Methods(test.Method)

			} else if test.Url == "/repos/:owner/:repo/collaborators" {
				htx.HandleFunc(test.Url, rptParams2).Methods(test.Method)

			} else {
				// htx.AddRoute(test.Method, test.Url, test.Result, func(w http.ResponseWriter, r *http.Request, ps Params) { j := i; arrived = 1000 + j }) // 0
				// fmt.Printf("URL: %s %s\n", test.Url, debug.LF())
				htx.HandleFunc(test.Url, createFx(i)).Methods(test.Method)
			}
		}
	}
	htx.HandleFunc("/r1", createFx(4000)).Methods("GET").Name("AName").Id(133)
	htx.HandleFunc("/r1", createFx(4001)).Methods("POST")
	htx.HandleFunc("/r1", createFx(4002)).Methods("PUT")
	htx.HandleFunc("/r1", createFx(4003)).Methods("PATCH")
	htx.HandleFunc("/r1", createFx(4004)).Methods("DELETE")
	htx.HandleFunc("/r1", createFx(4005)).Methods("OPTIONS")
	htx.HandleFunc("/r1", createFx(4006)).Methods("HEAD")
	htx.HandleFunc("/r1", createFx(4007)).Methods("CONNECT")
	htx.HandleFunc("/r1", createFx(4008)).Methods("TRACE")
	htx.HandleFunc("/r2/{blah:[a-z]}/:goo", createFx(4009)).Methods("GET")
	htx.HandleFunc("/r2/{blah:[1-9]}/:goo", createFx(4010)).Methods("GET")
	htx.HandleFunc("/r2/{blah:[A-Z]}/:goo", createFx(4011)).Methods("GET")
	// htx.HandleFunc("/hp01", createFx(4012)).Methods("GET").Host("localhost").Port("2000")
	// fmt.Printf("\n\nSetup of /hp01\n")
	htx.HandleFunc("/hp01", createFx(4012)).Methods("GET").Host("localhost").Port("2000")
	// fmt.Printf("\nEnd Setup of /hp01\n\n")
	// last := len(htx.routes) - 1
	// fmt.Printf("last=%d\n", last)
	// fmt.Printf("\nDump of last add, Route data[%d] = %s, %s\n", last, debug.SVarI(htx.routes[last]), debug.LF())

	htx.PathPrefix("/pp00/pp01/").HandleFunc("/pp02", createFx(4014))
	htx.HandleFunc("/proto01", createFx(4016)).Methods("GET").Protocal("HTTP/1.1", "HTTP/2.0")
	// Headers
	htx.HandleFunc("/hdr01", createFx(4018)).Headers("x-test-header", "def").Methods("GET")
	htx.HandleFunc("/hdr01", createFx(4017)).Methods("GET")
	// Queries
	htx.HandleFunc("/qry01", createFx(4020)).Methods("GET").Queries("id", "22")
	htx.HandleFunc("/qry01", createFx(4021)).Methods("GET")

	// r.HandleFunc(test.route, rptCalled).Methods("GET")

	htx.setDefaults()
	htx.buildRoutingTable()
	htx.CompileRoutes()
	// fmt.Printf ( "Calling OutputStatusInfo, %s\n", debug.LF())
	// htx.OutputStatusInfo( )

	return
}

var ServeHTTP_Tests = []struct {
	RunTest       bool
	Method        string
	HTTPS         string
	Host          string
	Url           string
	Expect        int
	ShouldBeFound bool
	RawQuery      string
	Rp            string
	CookieName    string
	CookieValue   string
	Proto         string
	Headers       string // Headers       map[string][]string
	Protocal      string
}{
	/*  00 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "https",
		Host:          "localhost:2000",
		Url:           "/xyz01",
		Expect:        263,
		ShouldBeFound: true,
	},
	/*  01 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "http",
		Host:          "localhost:2000",
		Url:           "/xyz01",
		Expect:        13,
		ShouldBeFound: false,
	},
	/*  02 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "https",
		Host:          "localhost:2001",
		Url:           "/xyz01",
		Expect:        13,
		ShouldBeFound: false,
	},
	/*  03 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "http",
		Host:          "localhost:2001",
		Url:           "/xyz01",
		Expect:        13,
		ShouldBeFound: false,
	},
	/*  04 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "https",
		Host:          "127.0.0.1:2001",
		Url:           "/xyz01",
		Expect:        13,
		ShouldBeFound: false,
	},
	/*  05 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "http",
		Host:          "127.0.0.1:2001",
		Url:           "/xyz01",
		Expect:        13,
		ShouldBeFound: false,
	},
	/*  06 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "http",
		Host:          "localhost:8090",
		Url:           "/img/checkbox_yes.png",
		Expect:        20,
		ShouldBeFound: true,
	},
	/*  07 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "http",
		Host:          "localhost:8090",
		Url:           "/user/keys/1234",
		Expect:        258,
		ShouldBeFound: true,
		RawQuery:      "left=55&right=22&id=IdOnUrlWrong&x=12&x=15",
		Rp:            `[{RunTest:"Name":"right","Value":"22","From":1,"Type":113},{"Name":"id","Value":"1234","From":0,"Type":58},{"Name":Host:"left","Value":"55","From":1,"Type":113}]`,
		CookieName:    "top",
		CookieValue:   "999",
	},
	/*  08 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "http",
		Host:          "localhost:8090",
		Url:           "/user",
		Expect:        243,
		ShouldBeFound: true,
		RawQuery:      "left=66&right=22&id=IdOnUrlWrong&METHOD=HEAD",
	},
	/*  09 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "https",
		Host:          "localhost:2000",
		Url:           "/hp01",
		Expect:        4012,
		ShouldBeFound: true,
	},
	/*  10 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "https",
		Host:          "localhost:2000",
		Url:           "/pp00/pp01/pp02",
		Expect:        4014,
		ShouldBeFound: true,
	},
	/*  11 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "https",
		Host:          "localhost:2000",
		Url:           "/pp00/pp01/pp02",
		Expect:        4014,
		ShouldBeFound: true,
		Proto:         "HTTP/2.0",
		Headers:       `{"x-test-header":["abc","def"],"Content-Type":["application/json"],"X-Requested-With":["XMLHttpRequest"]}`,
	},
	/*  12 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "https",
		Host:          "localhost:2000",
		Url:           "/proto01",
		Expect:        4016,
		ShouldBeFound: true,
		Proto:         "HTTP/2.0",
	},
	/*  13 */ { // test failed because - no negative match - fix sort to include extended matches
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "https",
		Host:          "localhost:2000",
		Url:           "/proto01",
		Expect:        13,
		ShouldBeFound: true,
		Proto:         "HTTP/1.0",
	},
	/*  14 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "https",
		Host:          "localhost:2000",
		Url:           "/hdr01",
		Expect:        4017,
		ShouldBeFound: true,
		Proto:         "HTTP/2.0",
	},
	/*  15 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "https",
		Host:          "localhost:2000",
		Url:           "/hdr01",
		Expect:        4018,
		ShouldBeFound: true,
		Proto:         "HTTP/2.0",
		Headers:       `{"X-Test-Header":["abc","def"],"Content-Type":["application/json"],"X-Requested-With":["XMLHttpRequest"]}`,
	},
	/*  16 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "https",
		Host:          "localhost:2000",
		Url:           "/hdr01",
		Expect:        4017,
		ShouldBeFound: true,
		Proto:         "HTTP/2.0",
		Headers:       `{"x-test-header":["xyz","uwv"],"Content-Type":["application/json"],"X-Requested-With":["XMLHttpRequest"]}`,
	},
	/*  17 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "http",
		Host:          "localhost:8090",
		Url:           "/qry01",
		Expect:        4020,
		ShouldBeFound: true,
		RawQuery:      "left=55&right=22&id=22&x=12&x=15",
	},
	/*  18 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "http",
		Host:          "localhost:8090",
		Url:           "/qry01",
		Expect:        4021,
		ShouldBeFound: true,
		RawQuery:      "left=55&right=23&id=23&x=12&x=15",
	},
	/*  19 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "http",
		Host:          "localhost:8090",
		Url:           "/repos/:owner/:repo/collaborators",
		Expect:        4008,
		ShouldBeFound: true,
		RawQuery:      "left=55&right=23&id=23&x=12&x=15",
	},
}

// func (r *MuxRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
func Test_1_ServeHTTP(t *testing.T) {
	// fmt.Printf("\nTest2017_newHash2\n")
	//r := htx
	//var procData MuxRouterProcessing
	//InitMuxRouterProcessing(r, &procData)
	r, trr := setupHtx()

	var req http.Request
	// var w http.ResponseWriter
	var url url.URL
	var tls tls.ConnectionState

	disableOutput = true

	w := new(mockResponseWriter)
	www := NewMidBuffer(w, nil)

	// router.ServeHTTP(w, req)

	r.NotFound = func(w http.ResponseWriter, req *http.Request) {
		arrived = -1
	}
	req.URL = &url

	rr := dumpCType(MultiUrl | SingleUrl | Dummy | IsWord)
	if rr != "(IsWord|MultiUrl|SingleUrl|Dummy)" {
		t.Errorf("Test dumpCType failed %s\n", rr)
	}

	for i, v := range ServeHTTP_Tests {
		if v.RunTest {
			InitParams(&www.Ps)
			req.URL.Path = v.Url
			req.URL.RawQuery = v.RawQuery
			req.Method = v.Method
			req.Host = v.Host
			req.TLS = nil
			req.RemoteAddr = "[::1]:53248"
			// req.RequestURI = v.HTTPS + "://" + v.Host + v.Url + "?" + v.RawQuery
			if v.RawQuery != "" {
				req.RequestURI = v.Url + "?" + v.RawQuery
			} else {
				req.RequestURI = v.Url
			}
			req.Proto = "HTTP/1.1"
			if v.Proto != "" {
				req.Proto = v.Proto
			}
			req.Header = make(http.Header)
			// fmt.Printf("req.Method=%s, %s\n", req.Method, debug.LF())
			if v.Headers != "" {
				// func (r *MuxRouter) Headers(pairs ...string) *ARoute {
				// Headers:       `{"x-test-header":["abc","def"],"Content-Type":["application/json"],"X-Requested-With":["XMLHttpRequest"]}`,
				// fmt.Printf("Creating header\n")
				err := json.Unmarshal([]byte(v.Headers), &req.Header)
				if err != nil {
					fmt.Printf("Error(20032): %v, %s, Headers ->%s<_\n", err, debug.LF(), v.Headers)
				}
			}
			if v.HTTPS == "https" {
				req.TLS = &tls
			}
			if v.CookieName != "" {
				c := http.Cookie{Name: v.CookieName, Value: v.CookieValue}
				req.AddCookie(&c)
			}
			arrived = 0
			// ---------------------------------------------------------------------------------------------------
			// fmt.Printf("\nStart of %d, req.RequestURI=%s, %s\n", i, req.RequestURI, debug.LF())
			// fmt.Printf("req.Method=%s, %s\n", req.Method, debug.LF())
			r.ServeHTTP(www, &req)
			// ---------------------------------------------------------------------------------------------------
			// fmt.Printf("route_i = %d, r.route[%d][ %s %s ]\n", g_route_i, g_route_i, debug.SVar(r.routes[g_route_i].DMethods), r.routes[g_route_i].DPath)
			if arrived != v.Expect {
				t.Errorf("Test[%d][%s %s %s %s]: Loop Expected to have handler called. Got:%d\n", i, v.Method, v.HTTPS, v.Host, v.Url, arrived)
			} else if v.Rp != "" {
				fmt.Printf("%s\n", rpTest)
				//var aa, bb map[int]map[string]interface{}
				//err0 := json.Unmarshal([]byte(v.Rp), &aa)
				//err1 := json.Unmarshal([]byte(rpTest), &bb)
				//eq := reflect.DeepEqual(aa, bb)
				//if !eq || err0 != nil || err1 != nil {
				//	t.Errorf("Test[%d]: Loop Expected:\n%s\nGot:\n%s, err0=%v err1=%v\n", i, v.Rp, rpTest, err0, err1)
				//}
			}
		}
	}

	// func (r *MuxRouter) UrlToCleanRoute(UsePat string) (rv string) {
	url2 := "/abc/:def/:ghi/jkl"
	Method := "GET"
	m := (int(Method[0]) + (int(Method[1]) << 1))
	r.SplitOnSlash3(trr, m, url2, true)
	rv := r.UrlToCleanRoute(trr, "T::T")
	if rv != "/abc/:/:/jkl" {
		t.Errorf("Test: Expected to have clean pattern\n")
	}

}

const x_test = true

var ServeHTTP_Tests2 = []struct {
	RunTest       bool
	Method        string
	HTTPS         string
	Host          string
	Url           string
	Expect        int
	ShouldBeFound bool
	RawQuery      string
	Rp            string
	CookieName    string
	CookieValue   string
	Proto         string
	Headers       string // Headers       map[string][]string
	Protocal      string
}{
	/*  00 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "http",
		Host:          "localhost:8090",
		Url:           "/",
		Expect:        5001,
		ShouldBeFound: true,
		RawQuery:      "",
	},
	/*  01 */ {
		RunTest:       true,
		Method:        "POST",
		HTTPS:         "http",
		Host:          "localhost:8090",
		Url:           "/",
		Expect:        5002,
		ShouldBeFound: true,
		RawQuery:      "",
	},
	/*  02 */ {
		RunTest:       true,
		Method:        "PUT",
		HTTPS:         "http",
		Host:          "localhost:8090",
		Url:           "/",
		Expect:        -1,
		ShouldBeFound: false,
		RawQuery:      "",
	},
	/*  03 */ {
		RunTest:       true,
		Method:        "GET",
		HTTPS:         "http",
		Host:          "localhost:8090",
		Url:           "/",
		Expect:        5001,
		ShouldBeFound: true,
		RawQuery:      "abc=def&id=22",
	},
	/*  04 */ {
		RunTest:       x_test,
		Method:        "GET",
		HTTPS:         "http",
		Host:          "localhost:8090",
		Url:           "/5003",
		Expect:        5003,
		ShouldBeFound: true,
		RawQuery:      "",
	},
	/*  05 */ {
		RunTest:       x_test,
		Method:        "POST",
		HTTPS:         "http",
		Host:          "localhost:8090",
		Url:           "/5004",
		Expect:        5004,
		ShouldBeFound: true,
		RawQuery:      "",
	},
	/*  06 */ {
		RunTest:       x_test,
		Method:        "PUT",
		HTTPS:         "http",
		Host:          "localhost:8090",
		Url:           "/5003",
		Expect:        -1,
		ShouldBeFound: false,
		RawQuery:      "",
	},
	/*  07 */ {
		RunTest:       x_test,
		Method:        "PUT",
		HTTPS:         "http",
		Host:          "localhost:8090",
		Url:           "/5004",
		Expect:        -1,
		ShouldBeFound: false,
		RawQuery:      "",
	},
}

func Test_04_ServeHTTP(t *testing.T) {
	// fmt.Printf("\nTest_2_ServeHTTP ------------------------------------------------------------------------------------------------------------------\n")
	var ht2 *MuxRouter
	ht2 = NewRouter()
	r := ht2
	var procData MuxRouterProcessing
	InitMuxRouterProcessing(r, &procData)
	trr := &procData
	//r, trr := setupHtx()
	//ht2 := r

	aaa := createFx(5001)
	bbb := createFx(5002)

	ht2.HandleFunc("/", aaa).Methods("GET", "DELETE")
	ht2.HandleFunc("/", bbb).Methods("POST")
	if x_test {
		ht2.HandleFunc("/5003", createFx(5003)).Methods("GET")
		ht2.HandleFunc("/5004", createFx(5004)).Methods("POST")
	}

	ht2.setDefaults()
	ht2.buildRoutingTable()
	ht2.CompileRoutes()

	var req http.Request
	// var w http.ResponseWriter
	var url url.URL
	var tls tls.ConnectionState

	disableOutput = true

	w := new(mockResponseWriter)
	www := NewMidBuffer(w, nil)

	// router.ServeHTTP(w, req)

	r.NotFound = func(w http.ResponseWriter, req *http.Request) {
		arrived = -1
	}
	req.URL = &url

	// Test closures
	{
		var ps Params
		ps.route_i = 0
		arrived = -2
		// aaa(w, &req, ps)			// probably a problem xyzzy
		aaa(w, &req)
		if arrived != 5001 {
			t.Errorf("Test[000] closure did not work\n")
		}
		arrived = -2
		ps.route_i = 1
		// bbb(w, &req, ps)			// probably a problem xyzzy
		bbb(w, &req)
		if arrived != 5002 {
			t.Errorf("Test[001] closure did not work\n")
		}
	}

	rr := dumpCType(MultiUrl | SingleUrl | Dummy | IsWord)
	if rr != "(IsWord|MultiUrl|SingleUrl|Dummy)" {
		t.Errorf("Test dumpCType failed %s\n", rr)
	}

	for i, v := range ServeHTTP_Tests2 {
		if v.RunTest {
			req.URL.Path = v.Url
			req.URL.RawQuery = v.RawQuery
			req.Method = v.Method
			req.Host = v.Host
			req.TLS = nil
			req.RemoteAddr = "[::1]:53248"
			// req.RequestURI = v.HTTPS + "://" + v.Host + v.Url + "?" + v.RawQuery
			if v.RawQuery != "" {
				req.RequestURI = v.Url + "?" + v.RawQuery
			} else {
				req.RequestURI = v.Url
			}
			req.Proto = "HTTP/1.1"
			if v.Proto != "" {
				req.Proto = v.Proto
			}
			req.Header = make(http.Header)
			if v.Headers != "" {
				err := json.Unmarshal([]byte(v.Headers), &req.Header)
				if err != nil {
					fmt.Printf("Error(20032): %v, %s, Headers ->%s<_\n", err, debug.LF(), v.Headers)
				}
			}
			if v.HTTPS == "https" {
				req.TLS = &tls
			}
			if v.CookieName != "" {
				c := http.Cookie{Name: v.CookieName, Value: v.CookieValue}
				req.AddCookie(&c)
			}
			arrived = 0
			// ---------------------------------------------------------------------------------------------------
			r.ServeHTTP(www, &req)
			// ---------------------------------------------------------------------------------------------------
			if arrived != v.Expect {
				t.Errorf("Test[%d][%s %s %s url=%s]: Loop Expected to have handler called. Expected: %d Got:%d\n", i, v.Method, v.HTTPS, v.Host, v.Url, v.Expect, arrived)
			} else if v.Rp != "" {
				fmt.Printf("%s\n", rpTest)
			}
		}
	}

	Method := "GET"
	m := MethodToCode(Method, 0)
	// fmt.Printf("m=%d\n", m)
	Route := "/"
	r.SplitOnSlash3(trr, m, Route, false)
}

// -------------------------------------------------------------------------------------------------
// const dbMap1 = true

// func (r *MuxRouter) HostPort_AllRoutes(hp ...string) *MuxRouter {
