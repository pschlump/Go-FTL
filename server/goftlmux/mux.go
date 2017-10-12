package goftlmux

//
// Go Go Mux - Go Fast Mux / Router for HTTP requests
//
// (C) Philip Schlump, 2013-2017.  All rights reserved.
// Version: 0.4.4
// BuildNo: 804
//

// xyzzyGoFTL10

// Code derived, but not copied, from Gorilla Mux and HttpRouter.
// Code from HttpRouter is copyright the Go authors under the same license.
// All of the rest of the code is MIT licensed.

/*
TODO:
	// fileServer := http.FileServer(root)		// xyzzyGoFtl02 - replace with file server from our GoFTL code
	2. look for all xyzzyGoFtl and address
*/

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"

	"github.com/pschlump/MiscLib"
	debug "github.com/pschlump/godebug"
)

/*
	// "./context" // "github.com/gorilla/context"
	"github.com/pschlump/Go-FTL/server/goftlmux"	-- depricated -- -- or move all of this code into it --
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
*/

// NewRouter returns a new router instance.
func NewRouter() *MuxRouter {
	r := &MuxRouter{
		HasBeenCompiled: false,
		AllHostPortFlag: false,
		MaxSlash:        1,
		nLookupResults:  1,
	}
	fn, ln := LineFile(1)
	r.LookupResults = append(r.LookupResults, Collision2{cType: Dummy, FileName: fn, LineNo: ln})
	r.NotFound = http.NotFound // Set to default, http.NotFound handler.
	r.Hash2Test = make([]int, bitMask+1, bitMask+1)
	r.nMatch = make([]UrlPat, MaxSlashInUrl, MaxSlashInUrl)
	return r
}

func InitMuxRouterProcessing(r *MuxRouter, trr *MuxRouterProcessing) {
	trr.r = r
}

// MuxRouter registers routes to be matched and dispatches a handler.
//
//// xyzzy - change this comment to be accurate
// It implements the http.Handler interface, so it can be registered to serve
// requests:
//
//     var router = mux.NewRouter()
//
//     func main() {
//         http.Handle("/", router)
//     }
//
// Or, for Google App Engine, register it in a init() function:
//
//     func init() {
//         http.Handle("/", router)
//     }
//
// This will send all incoming requests to the router.
// MaxSlash       int                    // Maximum number of slashes found in any route - taken from compiling the routes
type MuxRouterProcessing struct {
	// Info used during processing of a URL --- This is bad --- This can not go into this struct!!!!!!!!!!!!!
	// ------------------------------------------------------------------------------------------------------
	CurUrl string                 // The current URL being processed.
	Hash   [MaxSlashInUrl]int     // The set of hash keys in the current operation.
	Slash  [MaxSlashInUrl + 1]int // Array of locaitons for the '/' in the url.  For /abc/def, it would be [ 0, 4, 8 ]
	NSl    int                    // Number of slashes in the URL for /abc/def it would be 2
	UsePat string                 // The used T::T pattern for matching - at URL time.
	r      *MuxRouter             // The router that this belongs to
}

type MuxRouter struct {
	NotFoundHandler http.Handler // Configurable Handler to be used when no route matches.
	routes          []*ARoute    // Routes to be matched, from longes to shotest
	UseRedirect     bool         // PJS

	AllHostPortFlag bool
	AllHostPort     map[string]int

	// ------------------------------------------------------------------------------------------------------
	// The hash of paths/URLs
	// HashItems []HashItem
	Hash2Test []int

	// ------------------------------------------------------------------------------------------------------
	// Info used during processing of a URL --- This is bad --- This can not go into this struct!!!!!!!!!!!!!
	// ------------------------------------------------------------------------------------------------------
	// CurUrl   string                 // The current URL being processed.
	// Hash     [MaxSlashInUrl]int     // The set of hash keys in the current operation.
	// Slash    [MaxSlashInUrl + 1]int // Array of locaitons for the '/' in the url.  For /abc/def, it would be [ 0, 4, 8 ]
	// NSl      int                    // Number of slashes in the URL for /abc/def it would be 2
	// UsePat   string                 // The used T::T pattern for matching - at URL time.
	// ------------------------------------------------------------------------------------------------------
	// ------------------------------------------------------------------------------------------------------

	MaxSlash int // Maximum number of slashes found in any route -- generated during route compile --

	// Processing Widgets ----------------------------------------------------------------------------------
	widgetBefore   []GoGoWidgetSet // Support for middleware (GoGoWidget)
	widgetAfter    []GoGoWidgetSet
	widgetHashNewM []GoGoWidgetSet

	// ------------------------------------------------------------------------------------------------------
	// Setup Info
	routeData []RouteData // Raw routes - this is where the routes are kept before compile

	// ------------------------------------------------------------------------------------------------------
	// User settable handler called when no match is found.  Type: http.HandlerFunc.
	// If not set then http.NotFound will be called.
	NotFound http.HandlerFunc

	// ------------------------------------------------------------------------------------------------------
	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics.
	PanicHandler func(http.ResponseWriter, *http.Request, interface{})

	// ------------------------------------------------------------------------------------------------------
	HasBeenCompiled bool //	Flag, set to true when the routes are compiled.

	LookupResults  []Collision2
	nLookupResults int

	DebugMatchOn bool
	nMatch       []UrlPat // Index by Length ( NSl )
}

// Handle is a function that can be registered to a route to handle HTTP
// requests. Like http.HandlerFunc, but has a third parameter for the values of
// parameters.  Parameters can be from the URL or from other sources.
/////// type Handle func(w http.ResponseWriter, req *http.Request, ps Params) 	// xyzzyGoFtl - marker
type Handle func(w http.ResponseWriter, req *http.Request)

// I will use this type (synonomous with Handle) as HandleFunc is a unique
// string and Handle is just a word.
// type HandleFunc func(w http.ResponseWriter, req *http.Request, ps Params)
type HandleFunc Handle

// Verify the NewRouter (GoGoMux) works with the http.Handler interface.  This will produce a syntax
// error if there is a mismatch in interface.
// See:  https://blog.splice.com/golang-verify-type-implements-interface-compile-time/
var _ http.Handler = NewRouter()

type ARoute struct {
	parent *MuxRouter

	DId            int                    // Used for testing
	DName          string                 // Used to identify a route by name
	DPath          string                 // Set by Handler("/path",Fx), Path(), PathPrefix()
	DPathPrefix    string                 // -- Concatenated on front of path --
	DHandlerFunc   HandleFunc             //
	DHeaders       []string               // Set by Headers()
	DHost          string                 // Set by Host()
	DPort          string                 // Set by Port()
	DHostPort      string                 // Set by HostPort()
	DMethods       []string               // Set by Methods()	List of methods, GET, POST etc.
	DSchemes       []string               // Set by Schemes()	https, http etc.
	DQueries       []string               // Set by Queries()
	DProtocal      map[string]bool        // Set by Protocal() https == TLS on, http == no TLS, both is no-check(default)
	DComment       string                 // Set by Comment()
	DUser          map[string]interface{} // Can be set by user to data needed in matches.
	HeaderMatchMap map[string]string      // Map constructed form pairs of DHeaders
	QueryMatchMap  map[string]string      // Map constructed form pairs of DQueries
	FileName       string                 // Line no && File name where this was defined
	LineNo         int                    //
}

type RouteData struct {
	Method      string          // GET, PUT ...
	Route       string          // Route Pattern /abc/:def/ghi
	Hdlr        int             // User supplied integer returned on finding route
	NFxNo       int             // Index into []ARoute on what function to use
	Ns          int             //
	MatchIt     []Match         // Array of potential matches with regular expressions
	MatchItRank MatchItRankType //
}

type MatchItRankType uint32

const (
	ReMatch       MatchItRankType = 1 << iota // Not used anymore
	HeaderMatch                   = 1 << iota // Has a match on the HTTP Headers
	QueryMatch                    = 1 << iota // Has a match on the query - stuff after question mark
	TLSMatch                      = 1 << iota // Matches "https" v.s. "http"
	PortMatch                     = 1 << iota // Has a match on the Port
	HostMatch                     = 1 << iota // Has a match on the Host
	PortHostMatch                 = 1 << iota // Matches both Host and Port
	ProtocalMatch                 = 1 << iota // Has a match on prococal, http/1.0, http/1.1, http/2.0
	User0Match                    = 1 << iota // Reserved for User Functions
	User1Match                    = 1 << iota // Reserved for User Functions
	User2Match                    = 1 << iota // Reserved for User Functions
	User3Match                    = 1 << iota // Reserved for User Functions
	User4Match                    = 1 << iota // Reserved for User Functions
)

// -------------------------------------------------------------------------------------------------
const MaxSlashInUrl = 20

/*
This is the maximum number of "/" that can be processed in a URL.  If you have more than this number
of slashes then it will try to route based on what it has found and it may succeed.  With more it
will just disregard the trailing end of the URL - and probably return a 404.

Realistically more than 20 slashes in a route is probably a mistake somewhere.  I have looked at
a bunch of APIs and none of them had more than 9 slashes in any route.
*/

// -------------------------------------------------------------------------------------------------
const nBits = 12
const bitMask = 0xfff

// const MaxParams = 200 -- Moved to ./Params.go // PJS Sun Nov 15 13:16:04 MST 2015

/*
Hash Table Size Constants

	bitMask				nBits									Probability of Miss
	Hex		Decimal		Number of Bits		Number of URLs 		resoled in single compare
	-------	-------		--------------		--------------		-------------------
	0x7ff	2048		11					75					96.34%
	0xfff	4096		12					175					95.73%
	0x1fff	8192 		13					350					95.73%
	0x3fff	16384		14					700					95.73%

If you have more URLs then you should probably increase the size of nBits and bitMask to match.
This will help to keep the performance to a near constant.
*/

//
// The average world length in English is 5.1 chars.  7 chars seems to be enough to capture
// the uniqueness of most API tokens.   I have looked at 9 different APIs and more than 7
// chars took more time but did not decrease the number of collisions in the hashing table.
// In most cases more than 7 chars increased the collisions and slowed down the processing.
//
// Note: Set to 6 for a 4x and a 3x collision - then test collision handling - in the Github test.
//
const nKeyChar = 7

// This causes an early exit in processing.  For example if you have both routes and files
// being served, but all the routes are /api and the files are /js, /css, /image - then
// if the first word "js" failes to match any routes (all starting with "api"), then
// the only match possible is something like /*filename - this will be use if possible.
// If you do not have a catch-all like /*filename then no match will be returned.
// Timing wise this is much qucker than just doing the entire search.
// Note: The _test fucntions assume that this is true.
const optionEarlyExit = true

// -------------------------------------------------------------------------------------------------
/*

Jenkins Hash: In C Code from Wikipedia.   This code is in the public domain.   This is the
old Jenkins hash.    A newer function exits - but I have not profiled the results of using
it.  This hash produces a low collision rate in a small amount of time.   I may switch to
using the newer (supposedly better) hash but not until I have done timing on it to very
that it will work with a similar speed.

uint32_t jenkins_one_at_a_time_hash(char *key, size_t len)
{
    uint32_t hash, i;
    for(hash = i = 0; i < len; ++i)
    {
        hash += key[i];
        hash += (hash << 10);
        hash ^= (hash >> 6);
    }
    hash += (hash << 3);
    hash ^= (hash >> 11);
    hash += (hash << 15);
    return hash;
}

*/

func (r *MuxRouter) DebugMatch(tf bool) {
	r.DebugMatchOn = tf
}

// ----------------------------------------------------------------------------
// Route factories
// ----------------------------------------------------------------------------

// Just return the data for the routes that are built
func (r *MuxRouter) ListRoutes() []*ARoute {
	return r.routes
}

// NewRoute registers an empty route.
func (r *MuxRouter) NewRoute() *ARoute {
	fn, ln := LineFile(3)
	route := &ARoute{parent: r, LineNo: ln, FileName: fn}
	route.DProtocal = make(map[string]bool)        // Set by Protocal() https == TLS on, http == no TLS, both is no-check(default)
	route.DUser = make(map[string]interface{})     // Can be set by user to data needed in matches.
	route.HeaderMatchMap = make(map[string]string) // Map constructed form pairs of DHeaders
	route.QueryMatchMap = make(map[string]string)  // Map constructed form pairs of DQueries
	r.routes = append(r.routes, route)
	return route
}

// I am still working on this pair of functions!

//// Handle registers a new route with a matcher for the URL path.
//// xyzzy - Unknown - don't understand this yet
//func (r *MuxRouter) Handle(path string, handler http.Handler) *ARoute {
//	return r.NewRoute().Handle(path, handler)
//}
//
//// Handle registers a new route with a matcher for the URL path.
//// xyzzy - Unknown - don't understand this yet
//func (r *ARoute) Handle(path string, handlerFunc http.Handler) *ARoute {
//	/*
//		Not Implemented Yet - still working on this -
//		r.Path = path		// Path pattern
//		r.HandleFunc = handlerFunc
//	*/
//	return r
//}

// HandleFunc registers a new route with a matcher for the URL path.
func (r *MuxRouter) HandleFunc(path string, f HandleFunc) *ARoute {
	return r.NewRoute().HandleFunc(path, f)
}

// HandleFunc registers a new route with a matcher for the URL path.
func (r *ARoute) HandleFunc(path string, f HandleFunc) *ARoute {
	r.DPath = path // Path pattern
	r.DHandlerFunc = f
	return r
}

// Headers registers a new route with a matcher for request header values.
func (r *MuxRouter) Headers(pairs ...string) *ARoute {
	// fmt.Printf("Setting Header - top\n")
	return r.NewRoute().Headers(pairs...)
}

// Headers registers a new route with a matcher for request header values.
func (r *ARoute) Headers(pairs ...string) *ARoute {
	// fmt.Printf("Setting Header - bot\n")
	r.DHeaders = append(r.DHeaders, pairs...)
	return r
}

// Set/Append to list of valid host/ports for all routes by this router
func (r *MuxRouter) HostPort_AllRoutes(hp ...string) *MuxRouter {
	for _, v := range hp {
		if hasReInString(v) {
			fmt.Printf("Warning(20018): regular expressions are not supported in host/port.  This will be fixed in a week or two.\n")
		}
		fmt.Fprintf(os.Stderr, "\n\nWarning(40018): hpn - bad code!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!.\n\n\n")
		r.AllHostPortFlag = true
		r.AllHostPort[v] = hpn
		hpn += 3
	}
	return r
}

var hpn = 5 // xyzzy - what the hell! - Just increments by 3, why 3 - why global static - what about multiple uses of same?

// Host registers a new route with a matcher for the URL host.
func (r *MuxRouter) Host(h string) *ARoute {
	return r.NewRoute().Host(h)
}

// Host registers a new route with a matcher for the URL host.
func (r *ARoute) Host(h string) *ARoute {
	if hasReInString(h) {
		fmt.Printf("Warning(20018): regular expressions are not supported in host/port.  This will be fixed in a week or two.\n")
	}
	r.DHost = h
	return r
}

// Host:Port registers a new route with a matcher for the URL host and port.
func (r *MuxRouter) HostPort(h string) *ARoute {
	return r.NewRoute().HostPort(h)
}

// Host:Port registers a new route with a matcher for the URL host and port.
func (r *ARoute) HostPort(h string) *ARoute {
	if hasReInString(h) {
		fmt.Printf("Warning(20018): regular expressions are not supported in host/port.  This will be fixed in a week or two.\n")
	}
	r.DHostPort = h
	return r
}

// Name sets the name for this route.  This is not use for matching the route.
func (r *MuxRouter) Name(n string) *ARoute {
	return r.NewRoute().Name(n)
}

// Name sets the name for this route.  This is not use for matching the route.
func (r *ARoute) Name(n string) *ARoute {
	r.DName = n
	return r
}

// Id sets the name for this route.  This is not used for matching the
// route.  This is used in testing.
func (r *MuxRouter) Id(n int) *ARoute {
	return r.NewRoute().Id(n)
}

// Id sets the name for this route.  This is not used for matching the
// route.  This is used in testing.
func (r *ARoute) Id(n int) *ARoute {
	r.DId = n
	return r
}

// Port sets the port or this route.   This is a string like "80" or "8000"
// xyzzy
func (r *MuxRouter) Port(p string) *ARoute {
	// fmt.Printf("Port ATx: %s\n", debug.LF())
	return r.NewRoute().Port(p)
}

// Port sets the port or this route.   This is a string like "80" or "8000"
// xyzzy - ports are numbers ? validate!
func (r *ARoute) Port(p string) *ARoute {
	// fmt.Printf("Port ATx: %s\n", debug.LF())
	if hasReInString(p) {
		fmt.Printf("Warning(20018): regular expressions are not supported in host/port.  This will be fixed in the future.\n")
	}
	r.DPort = p
	// fmt.Printf("r=%s\n", debug.SVarI(r))
	return r
}

// Protocal sets the HTTP Protocal http/1.0, http/1.1, http/2.0
func (r *MuxRouter) Protocal(p ...string) *ARoute {
	return r.NewRoute().Protocal(p...)
}

// Protocal sets the HTTP Protocal http/1.0, http/1.1, http/2.0
func (r *ARoute) Protocal(p ...string) *ARoute {
	if checkProtocal(p) {
		for _, v := range p {
			///*db*/ fmt.Printf("Setting ->%s<- to true\n", v)
			r.DProtocal[v] = true
		}
	}
	return r
}

// MatcherFunc registers a new route with a custom matcher function.
// xyzzy
/*
func (r *MuxRouter) MatcherFunc(f MatcherFunc) *ARoute {
	return r.NewRoute().MatcherFunc(f)
}
*/

// Methods registers a new route with a matcher for HTTP methods.
func (r *MuxRouter) Methods(methods ...string) *ARoute {
	return r.NewRoute().Methods(methods...)
}

// Methods registers a new route with a matcher for HTTP methods.
func (r *ARoute) Methods(methods ...string) *ARoute {
	if checkMethods(methods) {
		r.DMethods = append(r.DMethods, methods...)
	}
	return r
}

func (r *MuxRouter) Comment(c string) *ARoute {
	return r.NewRoute().Methods(c)
}

func (r *ARoute) Comment(c string) *ARoute {
	r.DComment = c
	return r
}

// Schemes registers a new route with a matcher for URL schemes.
func (r *MuxRouter) Schemes(schemes ...string) *ARoute {
	return r.NewRoute().Schemes(schemes...)
}

// Schemes registers a new route with a matcher for URL schemes.
func (r *ARoute) Schemes(schemes ...string) *ARoute {
	if checkScheme(schemes) {
		r.DSchemes = append(r.DSchemes, schemes...)
	}
	return r
}

// PathPrefix registers a new route with a matcher for the URL path prefix.
func (r *MuxRouter) PathPrefix(p string) *ARoute {
	return r.NewRoute().PathPrefix(p)
}

// PathPrefix registers a new route with a matcher for the URL path prefix.
func (r *ARoute) PathPrefix(p string) *ARoute {
	r.DPathPrefix = p
	return r
}

// Queries registers a new route with a matcher for URL query values.
// xyzzy
func (r *MuxRouter) Queries(q ...string) *ARoute {
	return r.NewRoute().Queries(q...)
}
func (r *ARoute) Queries(q ...string) *ARoute {
	if len(q)%2 == 1 {
		fmt.Printf("Query parameter is invalid.  Must be paris of name, value. Called From: %s\n", debug.LF(1))
	} else {
		r.DQueries = append(r.DQueries, q...)
	}
	for _, p := range q {
		if hasReInString(p) {
			fmt.Printf("Warning(20018): regular expressions are not supported in host/port.  This will be fixed in a week or two.\n")
		}
	}
	return r
}

// Path registers a new route with a matcher for the URL path.
func (r *MuxRouter) Path(tpl string) *ARoute {
	return r.NewRoute().Path(tpl)
}

// Path registers a new route with a matcher for the URL path.
func (r *ARoute) Path(p string) *ARoute {
	r.DPath = p
	return r
}

// ----------------------------------------------------------------------------
// Manpiulate LineNo/FileName in ARoute data.
// ----------------------------------------------------------------------------
func (r *ARoute) AppendFileName(p string) *ARoute {
	r.FileName += p
	return r
}

// Create default values for matching.   By default a non-specified route is
// assued to be a GET request with any scheme (both https and http)
func (r *MuxRouter) setDefaults() {
	for i := 0; i < len(r.routes); i++ {
		if r.routes[i].DId == 0 {
			r.routes[i].DId = i
		}
		if len(r.routes[i].DMethods) == 0 {
			r.routes[i].DMethods = append(r.routes[i].DMethods, "GET")
		}
		if len(r.routes[i].DSchemes) == 0 {
			r.routes[i].DSchemes = append(r.routes[i].DSchemes, "")
		}
	}
}

// See if both https and http - null case.
func isHttpHttps(s []string) (ignore bool, http bool, https bool) {
	ignore = false
	http = false
	https = false
	for _, v := range s {
		if v == "https" {
			https = true
		}
		if v == "http" {
			http = true
		}
	}
	if http && https {
		ignore = true
	}
	return
}

// Convert from the supplied routing information in r.routes to the sortable
// routing information.
func (r *MuxRouter) buildRoutingTable() {
	for i, v := range r.routes {
		for _, w := range r.routes[i].DMethods {
			k := r.addRoute(w, v.DPathPrefix+v.DPath, v.DId, v.DHandlerFunc, i, v.FileName, v.LineNo)
			if k >= 0 {

				ignore, http, https := isHttpHttps(v.DSchemes)
				if !ignore {
					if https && !http {
						r.setHTTPS_Only(k)
					} else if http && !https {
						r.setHTTP_Only(k)
					}
				}

				if v.DHostPort != "" {
					r.setHostPort(k)
				}
				if v.DHost != "" {
					r.setHost(k)
				}
				if v.DPort != "" {
					r.setPort(k)
				}
				if len(v.DHeaders) > 0 {
					// fmt.Printf("Setting DHeaders\n")
					x, err := mapFromPairs(v.DHeaders...)
					if err != nil {
						fmt.Printf("Error(20012): %s FileName: %s LineNo: %d\n", err, v.FileName, v.LineNo)
					} else {
						v.HeaderMatchMap = x
						r.setHeaderMatch(k)
					}
				}
				if len(v.DQueries) > 0 {
					x, err := mapFromPairs(v.DQueries...)
					if err != nil {
						fmt.Printf("Error(20018): %s FileName: %s LineNo: %d\n", err, v.FileName, v.LineNo)
					} else {
						v.QueryMatchMap = x
						r.setQueryMatch(k)
					}
				}
				// func (r *MuxRouter) setProtocal(k int) {
				if !IsMapStringBoolEmpty(v.DProtocal) {
					r.setProtocal(k)
				}
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Perform the match of a header.
func matchHeaderMatch(www http.ResponseWriter, req *http.Request, r *MuxRouter, route_i int) bool {
	return matchMap(r.routes[route_i].HeaderMatchMap, req.Header, true)
}

// Setup to match headers.
func (r *MuxRouter) setHeaderMatch(k int) {
	r.routeData[k].MatchIt = append(r.routeData[k].MatchIt, Match{MatchFunc: matchHeaderMatch})
	r.routeData[k].MatchItRank |= HeaderMatch
}

// ----------------------------------------------------------------------------
// Perform a match on the Query portion of the URL (Requires that the query
// be parsed by the GoGoWidget.
func matchQueryMatch(www http.ResponseWriter, req *http.Request, r *MuxRouter, route_i int) bool {
	if rw, ok := www.(*MidBuffer); ok {
		return matchQueryMap(r.routes[route_i].QueryMatchMap, rw.Ps)
	}
	return false
}

// Setup to match on query.
func (r *MuxRouter) setQueryMatch(k int) {
	r.routeData[k].MatchIt = append(r.routeData[k].MatchIt, Match{MatchFunc: matchQueryMatch})
	r.routeData[k].MatchItRank |= QueryMatch
}

// ----------------------------------------------------------------------------
func matchTlsFunc(www http.ResponseWriter, req *http.Request, r *MuxRouter, route_i int) bool {
	return req.TLS != nil
}
func matchNoTlsFunc(www http.ResponseWriter, req *http.Request, r *MuxRouter, route_i int) bool {
	return req.TLS == nil
}
func (r *MuxRouter) setHTTPS_Only(k int) {
	r.routeData[k].MatchIt = append(r.routeData[k].MatchIt, Match{MatchFunc: matchTlsFunc})
	r.routeData[k].MatchItRank |= TLSMatch
}
func (r *MuxRouter) setHTTP_Only(k int) {
	r.routeData[k].MatchIt = append(r.routeData[k].MatchIt, Match{MatchFunc: matchNoTlsFunc})
	r.routeData[k].MatchItRank |= TLSMatch
}

// ----------------------------------------------------------------------------
func matchPortFunc(www http.ResponseWriter, req *http.Request, r *MuxRouter, route_i int) bool {
	// fmt.Printf("***************************** r.routes[%d].DPort ->%s<- v.s. %s\n", route_i, r.routes[route_i].DPort, req.Host)
	colon := LastIndexOfChar(req.Host, ':')
	// fmt.Printf("!! DPort ->%s<- vs. ->%s<-\n", r.routes[route_i].DPort, req.Host[colon+1:])
	if colon != -1 {
		return r.routes[route_i].DPort == req.Host[colon+1:]
	} else {
		return r.routes[route_i].DPort == "80"
	}
}
func (r *MuxRouter) setPort(k int) {
	r.routeData[k].MatchIt = append(r.routeData[k].MatchIt, Match{MatchFunc: matchPortFunc})
	r.routeData[k].MatchItRank |= PortMatch
}

// ----------------------------------------------------------------------------
func matchHostFunc(www http.ResponseWriter, req *http.Request, r *MuxRouter, route_i int) bool {
	// fmt.Printf("***************************** r.routes[%d].DHost ->%s<- v.s. %s\n", route_i, r.routes[route_i].DHost, req.Host)
	colon := LastIndexOfChar(req.Host, ':')
	if colon != -1 {
		return r.routes[route_i].DHost == req.Host[:colon]
	} else {
		return false
	}
}
func (r *MuxRouter) setHost(k int) {
	r.routeData[k].MatchIt = append(r.routeData[k].MatchIt, Match{MatchFunc: matchHostFunc})
	r.routeData[k].MatchItRank |= HostMatch
}

// ----------------------------------------------------------------------------
func matchHostPortFunc(www http.ResponseWriter, req *http.Request, r *MuxRouter, route_i int) bool {
	// fmt.Printf("***************************** r.routes[%d].DHostPort ->%s<- v.s. %s\n", route_i, r.routes[route_i].DHostPort, req.Host)
	return r.routes[route_i].DHostPort == req.Host
}
func (r *MuxRouter) setHostPort(k int) {
	r.routeData[k].MatchIt = append(r.routeData[k].MatchIt, Match{MatchFunc: matchHostPortFunc}) // route_i
	r.routeData[k].MatchItRank |= PortHostMatch
}

// ----------------------------------------------------------------------------
func matchProtocalFunc(www http.ResponseWriter, req *http.Request, r *MuxRouter, route_i int) bool {
	///*db*/ fmt.Printf(":42: Checking ->%s<- for correct protocal = %v, %s\n", req.Proto, r.routes[route_i].DProtocal[req.Proto], debug.LF())
	return r.routes[route_i].DProtocal[req.Proto]
}
func (r *MuxRouter) setProtocal(k int) {
	///*db*/ fmt.Printf(":42:setProtocal called\n")
	r.routeData[k].MatchIt = append(r.routeData[k].MatchIt, Match{MatchFunc: matchProtocalFunc}) // route_i
	r.routeData[k].MatchItRank |= ProtocalMatch
}

// ----------------------------------------------------------------------------

const disableOutput bool = false

//type MyResponseWriter struct {
//	StartTime     time.Time
//	Status        int
//	ResponseBytes int64
//	w             http.ResponseWriter
//}

//func (m *MyResponseWriter) Header() http.Header {
//	return m.w.Header()
//	// return http.Header{}
//}

//func (m *MyResponseWriter) Write(p []byte) (written int, err error) {
//	if disableOutput {
//		written = len(string(p))
//		m.ResponseBytes += int64(written)
//		return written, nil
//	}
//	written, err = m.w.Write(p)
//	m.ResponseBytes += int64(written)
//	return written, err
//}

//func (m *MyResponseWriter) WriteHeader(p int) {
//	m.Status = p
//	m.w.WriteHeader(p)
//}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------

type Where int

const (
	HashNewM Where = iota
	Before         = iota
	After          = iota
)

type GoGoWidgetSet struct {
	w  Where
	fx GoGoWidgetFunc
}

// type GoGoWidgetFunc func(http.ResponseWriter, *http.Request, *Params, *GoGoData, int) int
// type GoGoWidgetFunc func(*MyResponseWriter, *http.Request, *Params) int
type GoGoWidgetFunc func(*MidBuffer, *http.Request, *Params) int

type GoGoWidgetSetMatch struct {
	w  Where
	fx GoGoWidgetMatchFunc
}

type GoGoWidgetMatchFunc func(http.ResponseWriter, *http.Request, Params, *int, int, *[]string) bool

// Attach middlewhare widget to the handler.
func (r *MuxRouter) AttachWidget(w Where, fx GoGoWidgetFunc) {
	switch w {
	case HashNewM:
		r.widgetHashNewM = append(r.widgetHashNewM, GoGoWidgetSet{w: w, fx: fx})
	case Before:
		r.widgetBefore = append(r.widgetBefore, GoGoWidgetSet{w: w, fx: fx})
	case After:
		r.widgetAfter = append(r.widgetAfter, GoGoWidgetSet{w: w, fx: fx})
	}
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------

type colType int

const (
	IsWord    colType = 1 << iota
	MultiUrl          = 1 << iota
	SingleUrl         = 1 << iota
	Dummy             = 1 << iota
)

type MatchFunc func(www http.ResponseWriter, req *http.Request, r *MuxRouter, route_i int) bool

type Match struct {
	MatchFunc MatchFunc
}

type Re struct {
	Pos  int
	Re   string
	cRe  *regexp.Regexp
	Name string
}

type ReList struct {
	Hdlr     int        // User specified int, mostly for testing
	Fx       HandleFunc // Function to call to handle this request
	ArgNames []string   //
	ReSet    []Re
	MatchIt  []Match
	route_i  int
}

type Collision2 struct {
	cType      colType    // if IsWord, then this is just a marker that the prefix is a valid m+word hash
	Url        string     // /path/:v1/:v2/whatever
	Comment    string     // Optionally set comment on what is getting called.
	NSL        int        // number of / in the URL/Route
	CleanUrl   string     // /path/:/:/whatever
	Hdlr       int        // User specified int, mostly for testing
	Fx         HandleFunc // Function to call to handle this request
	TPat       string     // T::T
	ArgNames   []string   //
	ArgPattern string     // T::T dup?? not used
	LineNo     int        // Location this was created
	FileName   string     // Location this was created
	HasRe      []ReList   // Set of RE that is required to match this Collision2
	MatchIt    []Match    // If additional matching criteria are used
	route_i    int
	Multi      map[string]Collision2 // if (cType&MultiUrl)!=0, then use string to disambiguate collisions
}

// -------------------------------------------------------------------------------------------------
// USED - in related GET/PUT/POST functions.

// Add a route to be handed.
//
// Handle registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH, OPTIONS, HEAD and DELETE requests the
// respective shortcut functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (r *MuxRouter) addRoute(Method string, Route string, hdlr int, fx HandleFunc, pos int, fn string, ln int) int {
	if Route[0] != '/' {
		fmt.Printf("Error(20002): Path should begin with '/', passed %s, File:%s LinLineNo:%d\n", Route, fn, ln)
		//if oneSlash {
		//	fmt.Printf("%s\n", debug.LF(1))
		//	fmt.Printf("%s\n", debug.LF(2))
		//	fmt.Printf("%s\n", debug.LF(3))
		//	fmt.Printf("%s\n", debug.LF(4))
		//}
		return -1
	}
	if !validMethod[Method] {
		fmt.Printf("Error(20003): Method invalid, should be one of: GET, POST, PUT, PATCH, OPTIONS, HEAD, CONNECT, TRACE or DELETE, instead got %s, File:%s LineNo:%d\n", Method, fn, ln)
		return -1
	}

	k := len(r.routeData)

	r.routeData = append(r.routeData, RouteData{
		Method: Method,
		Route:  Route,
		Hdlr:   hdlr,
		NFxNo:  pos,
	})

	return k
}

func (r *MuxRouter) CompileRoutes() {

	if r.HasBeenCompiled {
		return
	}
	r.HasBeenCompiled = true // Mark that the compilation has taken place.

	var procData MuxRouterProcessing
	trr := &procData
	InitMuxRouterProcessing(r, trr)

	r.setDefaults()
	r.buildRoutingTable()
	r.calcNumSlash() // Use this to find over MaxSlashInUrl of slashes and report error/warn.

	// -------------------------------------------------------------------------------------------------
	sf_NumSlash_Desc := func(c1, c2 *RouteData) bool {
		return c1.Ns > c2.Ns
	}
	sf_Length_Desc := func(c1, c2 *RouteData) bool {
		return len(c1.Route) > len(c2.Route)
	}
	sf_Text := func(c1, c2 *RouteData) bool {
		return c1.Route < c2.Route
	}
	sf_MethodHash := func(c1, c2 *RouteData) bool {
		return MethodToCode(c1.Method, 0) < MethodToCode(c2.Method, 0)
	}
	sf_MatchFuncs := func(c1, c2 *RouteData) bool {
		return c1.MatchItRank > c2.MatchItRank
	}
	// -------------------------------------------------------------------------------------------------

	OrderedBy(sf_MethodHash, sf_NumSlash_Desc, sf_Length_Desc, sf_Text, sf_MatchFuncs).Sort(r.routeData)

	///*db*/ r.DumpRouteData("After Sort")
	// fmt.Printf("Dump of table at %s\n", debug.LF())
	// r.DumpRouteData("After Sort")

	for _, v := range r.routeData {
		fx := r.routes[v.NFxNo].DHandlerFunc
		FileName := r.routes[v.NFxNo].FileName
		LineNo := r.routes[v.NFxNo].LineNo
		Comment := r.routes[v.NFxNo].DComment
		cleanRoute, names := r.addPatT__T(trr, v.Route, v.Hdlr, fx, FileName, LineNo)
		ns := numChar(v.Route, '/')
		r.addHash2Map(trr, v.Method, v.Route, cleanRoute, v.Hdlr, fx, names, v.MatchIt, ns, v.NFxNo, FileName, LineNo, Comment) // AddToM
	}

	r.addStarPat(trr)
	r.sortPat(trr)
}

// -------------------------------------------------------------------------------------------------
/*

	Length Longest to Shortest, 	Hi .. Lo		len(Pat[i])
	Degrees of Freedom, 			Lo .. Hi		sortDf(Pat[i])
		T 		0
		{		1
		:		2
		*		3
	Frequencey in PatOcc, 			Hi .. Lo


This is dependent on having r.MaxSlash set properly.   That is set in calcNumSlash.

*/
func (r *MuxRouter) sortPat(trr *MuxRouterProcessing) {
	var CurPatOcc map[string]int
	sp_Length_Desc := func(c1, c2 *UrlAPat) bool {
		return len(c1.Pat) > len(c2.Pat)
	}
	sp_DF := func(c1, c2 *UrlAPat) bool {
		return sortDf(c1.Pat) < sortDf(c2.Pat)
	}
	sp_Frequency := func(c1, c2 *UrlAPat) bool {
		return CurPatOcc[c1.Pat] < CurPatOcc[c2.Pat]
	}
	sp_Text := func(c1, c2 *UrlAPat) bool {
		return c1.Pat < c2.Pat
	}
	for i := 0; i < minInt(MaxSlashInUrl, r.MaxSlash+1); i++ {
		if r.nMatch[i].PatList != nil && len(r.nMatch[i].PatList) > 1 {
			CurPatOcc = r.nMatch[i].PatOcc
			// fmt.Printf("sortPat: (before) r.nMatch[%d]=%s\n", i, debug.SVarI(r.nMatch[i]))
			OrderedByPat(sp_Length_Desc, sp_DF, sp_Frequency, sp_Text).Sort(r.nMatch[i].PatList)
			// fmt.Printf("sortPat: (after) r.nMatch[%d]=%s\n", i, debug.SVarI(r.nMatch[i]))
		}
	}
}

// -------------------------------------------------------------------------------------------------
// Extract arguments from the URL.
//func (r *MuxRouter) GetArgs3(Url string, _ string, names []string, _ int) {
//	k := 0
//	// db("GetArgs3", "names=%s r.UsePat=%s %s\n", debug.SVar(names), r.UsePat, debug.LF())
//	for i, v := range r.UsePat {
//		// db("GetArgs3","k=%d\n", k)
//		if i < MaxSlashInUrl-1 {
//			vv := ""
//			if v == ':' {
//				if r.Slash[i]+1 < len(Url) && r.Slash[i+1] <= len(Url) {
//					vv = Url[r.Slash[i]+1 : r.Slash[i+1]]
//				}
//				AddValueToParams(names[k], vv, ':', FromURL, &r.AllParam)
//				k++
//			} else if v == '{' {
//				if r.Slash[i]+1 < len(Url) && r.Slash[i+1] <= len(Url) {
//					vv = Url[r.Slash[i]+1 : r.Slash[i+1]]
//				}
//				AddValueToParams(names[k], vv, '{', FromURL, &r.AllParam)
//				k++
//			} else if v == '*' {
//				if r.Slash[i]+1 < len(Url) {
//					vv = Url[r.Slash[i]+1:]
//				}
//				AddValueToParams(names[k], vv, '{', FromURL, &r.AllParam)
//				k++
//			}
//		}
//	}
//}

// -------------------------------------------------------------------------------------------------
// Extract arguments from the URL.
func (r *MuxRouter) GetArgs4(trr *MuxRouterProcessing, Url string, _ string, names []string, _ int, ps *Params) {
	k := 0
	// db("GetArgs3", "names=%s r.UsePat=%s %s\n", debug.SVar(names), r.UsePat, debug.LF())
	for i, v := range trr.UsePat {
		// db("GetArgs3","k=%d\n", k)
		if i < MaxSlashInUrl-1 {
			vv := ""
			if v == ':' {
				if trr.Slash[i]+1 < len(Url) && trr.Slash[i+1] <= len(Url) {
					vv = Url[trr.Slash[i]+1 : trr.Slash[i+1]]
				}
				AddValueToParams(names[k], vv, ':', FromURL, ps)
				k++
			} else if v == '{' {
				if trr.Slash[i]+1 < len(Url) && trr.Slash[i+1] <= len(Url) {
					vv = Url[trr.Slash[i]+1 : trr.Slash[i+1]]
				}
				AddValueToParams(names[k], vv, '{', FromURL, ps)
				k++
			} else if v == '*' {
				if trr.Slash[i]+1 < len(Url) {
					vv = Url[trr.Slash[i]+1:]
				}
				AddValueToParams(names[k], vv, '{', FromURL, ps)
				k++
			}
		}
	}
}

// Post process r.nMatch adding all of the "*" patterns where they need to be during LookupUrlViaHash2.
//
// for the longest pattern with a star
//    for each pattern that is longer than that up to min(MaxSlashInUrl,r.MasSlash+1)
//		Add that pattern (the * one) to the set of the others.
//
func (r *MuxRouter) addStarPat(trr *MuxRouterProcessing) {
	// fmt.Printf("r.nMatch=%s\n", debug.SVarI(r.nMatch))
	for i := minInt(MaxSlashInUrl-1, r.MaxSlash+1); i > 0; i-- { // nothing at 0 so skip it.
		// fmt.Printf("i=%d\n", i)
		if r.nMatch[i].PatList != nil {
			// fmt.Printf("Star is not nil\n")
			for ii := 0; ii < len(r.nMatch[i].PatList); ii++ {
				// fmt.Printf("ii=%d\n", ii)
				if r.nMatch[i].PatList[ii].Star {
					mm := minInt(MaxSlashInUrl, r.MaxSlash+1)
					for j := i + 1; j < mm; j++ {
						// do add
						p := r.nMatch[i].PatList[ii]
						r.nMatch[j].PatList = append(r.nMatch[j].PatList, p)
						if r.nMatch[j].PatOcc == nil {
							r.nMatch[j].PatOcc = make(map[string]int)
						}
						r.nMatch[j].PatOcc[p.Pat] = 1
					}
				}
			}
		}
	}
}

// Return true if the sting 'p' has a '*' in it.
// Possible Improvement: this should be a better scanner - this, just look for '*' character is probably buggy.
func hasStar(p string) bool {
	// func numChar(s string, c rune) (rv int) {
	// fmt.Printf("p ->%s<- numChcar=%d\n", p, numChar(p, '*'))
	if numChar(p, '*') > 0 {
		return true
	}
	return false
}

// Add a pattern to r.nMatch - check to see if it is already there.
// Possible Improvement - inefficient/slow but it works.
func (r *MuxRouter) addPat2(NSl int, p string, FileName string, LineNo int) {
	f := false
	// fmt.Printf("NSl=%d ->%s<- %s\n", NSl, p, debug.LF())
	for _, v := range r.nMatch[NSl].PatList {
		if v.Pat == p {
			f = true
			break
		}
	}
	if !f {
		r.nMatch[NSl].PatList = append(r.nMatch[NSl].PatList, UrlAPat{Pat: p, Star: hasStar(p)})
	}
	if r.nMatch[NSl].PatOcc == nil {
		r.nMatch[NSl].PatOcc = make(map[string]int)
	}
	r.nMatch[NSl].PatOcc[p]++
}

// Count the number of characters 'c' in the string 's', return that value.
func numChar(s string, c rune) (rv int) {
	rv = 0
	for _, v := range s {
		if v == c {
			rv++
		}
	}
	return
}

func hasReInString(s string) bool {
	a := numChar(s, '{')
	b := numChar(s, '}')
	return a > 0 && b > 0
}

// Iterate over the set of routes and calculate the number of '/' in each route.
func (r *MuxRouter) calcNumSlash() {
	for i, v := range r.routeData {
		ns := numChar(v.Route, '/')
		r.routeData[i].Ns = ns
		// fmt.Printf("ns=%2d %s\n", ns, v.Route)
	}
	r.MaxSlash = 1
	for _, v := range r.routeData {
		if r.MaxSlash < v.Ns {
			r.MaxSlash = v.Ns
		}
	}
	// fmt.Printf("MaxSlash=%d\n", r.MaxSlash)
}

// -------------------------------------------------------------------------------------------------
// Sorting of routes.   Use to improve the average case - at the expense of more rare cases.
// -------------------------------------------------------------------------------------------------

type lessFunc func(p1, p2 *RouteData) bool

// multiSorter implements the Sort interface, sorting the the_data within.
type multiSorter struct {
	the_data []RouteData
	less     []lessFunc
}

// Sort sorts the argument slice according to the less functions passed to OrderedBy.
func (ms *multiSorter) Sort(the_data []RouteData) {
	ms.the_data = the_data
	sort.Sort(ms)
}

// OrderedBy returns a Sorter that sorts using the less functions, in order.
// Call its Sort method to sort the data.
func OrderedBy(less ...lessFunc) *multiSorter {
	return &multiSorter{
		less: less,
	}
}

// Len is part of sort.Interface.
func (ms *multiSorter) Len() int {
	return len(ms.the_data)
}

// Swap is part of sort.Interface.
func (ms *multiSorter) Swap(i, j int) {
	ms.the_data[i], ms.the_data[j] = ms.the_data[j], ms.the_data[i]
}

// Less is part of sort.Interface. It is implemented by looping along the
// less functions until it finds a comparison that is either Less or
// !Less. Note that it can call the less functions twice per call. We
// could change the functions to return -1, 0, 1 and reduce the
// number of calls for greater efficiency: an exercise for the reader.
func (ms *multiSorter) Less(i, j int) bool {
	p, q := &ms.the_data[i], &ms.the_data[j]
	// Try all but the last comparison.
	var k int
	for k = 0; k < len(ms.less)-1; k++ {
		less := ms.less[k]
		switch {
		case less(p, q):
			// p < q, so we have a decision.
			return true
		case less(q, p):
			// p > q, so we have a decision.
			return false
		}
		// p == q; try the next comparison.
	}
	// All comparisons to here said "equal", so just return whatever
	// the final comparison reports.
	return ms.less[k](p, q)
}

//
// Build the route pattern table.  A route of /abc/:def/ghi will become T:T for the fixed tokens and return
// the string /abc/:/ghi for a matching patter for colision resolution.   The T:T patterns are stored by
// addPat2().
//
// Only called at route compile time.
//
func (r *MuxRouter) addPatT__T(trr *MuxRouterProcessing, Route string, hdlr int, fx HandleFunc, FileName string, LineNo int) (ss string, names []string) {
	i, k := 0, 0
	//if oneSlash {
	//	/*db*/ fmt.Printf("Route:%s, NSl=%d r.Slash=%s\n", Route, r.NSl, debug.SVar(r.Slash[:r.NSl+1]))
	//}
	pp := ""
	r.SplitOnSlash3(trr, 1, Route, false)
	if Route == "/" {
		ss += "/"
		pp += "T"
	} else {
		for i = 0; i < trr.NSl; i++ {
			if trr.Slash[i]+1 >= len(Route) {
			} else if trr.CurUrl[trr.Slash[i]+1] == ':' {
				ss += "/:"
				pp += ":"
				names = append(names, trr.CurUrl[trr.Slash[i]+2:trr.Slash[i+1]])
				k++
			} else if trr.CurUrl[trr.Slash[i]+1] == '*' {
				ss += "/*"
				pp += "*"
				names = append(names, trr.CurUrl[trr.Slash[i]+2:trr.Slash[i+1]])
				k++
				break
			} else if trr.CurUrl[trr.Slash[i]+1] == '{' {
				name, re, valid, convertToColon := parseReFromToken3(Route[trr.Slash[i]+1 : trr.Slash[i+1]])
				names = append(names, name)
				_, _, _, _ = name, re, valid, convertToColon
				if convertToColon {
					ss += "/:"
					pp += ":"
					k++
					//if dbHash2 {
					//	fmt.Printf(" ss=%s after : {nanme} / convertToColon, name=%s re=%s valid=%v \n", ss, name, re, valid)
					//}
				} else {
					ss += "/{"
					pp += "{"
					k++
					//if dbHash2 {
					//	fmt.Printf(" ss=%s after {name:re}\n", ss)
					//}
				}
			} else {
				ss += "/" + trr.CurUrl[trr.Slash[i]+1:trr.Slash[i+1]]
				pp += "T"
			}
		}
	}
	r.addPat2(trr.NSl, pp, FileName, LineNo)
	trr.UsePat = pp
	// ss = pp

	//if dbHash2 {
	//	fmt.Printf("names=%s, %s\n", debug.SVar(names), debug.LF())
	//}
	return
}

func MethodToCode(Method string, AddM int) int {
	return ((int(Method[0]) + (int(Method[1]) << 1)) + AddM)
}

// With a known bad URL, that has //, /./, or /../ in it, fix the URL.
func (r *MuxRouter) FixBadUrl(trr *MuxRouterProcessing, Url string) (rv string, fixed bool) {
	fixed = false
	rv = Url
	var tmpUrlArray [MaxSlashInUrl]string // Decleration takes 200 ns - Should move to r. so declare at startup.
	tmpUrl := tmpUrlArray[:]
	n := FixPath(Url, tmpUrl, MaxSlashInUrl)
	tmpUrl = tmpUrl[:n]
	nUrl := ""
	for i, v := range tmpUrl {
		if i > 0 {
			nUrl += "/" + v
		}
	}
	// fmt.Printf("Trailing Slash: nUrl ->%s<- Url ->%s<-\n", nUrl, Url)
	if nUrl != Url {
		trr.CurUrl = nUrl
		rv = nUrl
		fixed = true
	}
	return
}

func (r *MuxRouter) addHash2Map(trr *MuxRouterProcessing, Method string, Route string, cleanRoute string, hdlr int, fx HandleFunc, names []string, AddToM []Match, ns int, NFxNo int, FileName string, LineNo int, Comment string) {
	//if dbMatch2 {
	//	fmt.Printf("\naddHash2Map: len(AddToM) = %d %s\n", len(AddToM), debug.LF())
	//}
	var i int
	var ss int
	ss = 0
	var pp string
	tmpRe := make([]Re, 0, MaxSlashInUrl)
	reNames := make([]string, 0, MaxSlashInUrl)
	pp = ""
	//if dbHash2 || dbMatch2 {
	//	fmt.Printf("TOP(addHash2Map): %s %s (%s) => %d, %s\n", Method, Route, cleanRoute, hdlr, debug.LF())
	//}
	// m := ((int(Method[0]) + (int(Method[1]) << 1)) + AddToM) ^ (ns << 2)
	m := MethodToCode(Method, 0)
	// fmt.Printf("m=%d, %s\n", m, debug.LF())
	r.SplitOnSlash3(trr, m, Route, false)
	if optionEarlyExit {
		hh := (trr.Hash[0] ^ m) & bitMask
		// fmt.Printf("hh=%d bitMask=%x\n", m, bitMask)
		if r.Hash2Test[hh] == 0 {
			r.Hash2Test[hh] = r.nLookupResults
			r.LookupResults = append(r.LookupResults, Collision2{cType: IsWord, Comment: Comment})
			r.nLookupResults++
		}
	}
	if dbHash2 {
		fmt.Printf("After SplitOnSlash3 Orig:->%s<- Fixed:->%s<-\n r.Hash=%s r.Slash=%s r.NSl=%d\n", Route, trr.CurUrl, debug.SVar(trr.Hash[0:trr.NSl]), debug.SVar(trr.Slash[0:trr.NSl+1]), trr.NSl)
	}
	haveRealRe := false
	for i = 0; i < trr.NSl; i++ {
		//if dbHash2 {
		//	fmt.Printf("i=%d ->%c<-, ->%s<-", i, Route[trr.Slash[i]+1], Route[trr.Slash[i]+1:trr.Slash[i+1]])
		//}
		if dbHash2 {
			fmt.Printf("i=%d ss=%d, %s\n", i, ss, debug.LF())
		}
		if trr.Slash[i]+1 >= len(Route) {
			// fmt.Printf("AT %s\n", debug.LF())
			//if dbHash2 {
			//	fmt.Printf("At (Added to code at this point) %s\n", debug.LF())
			//}
			ss = ss ^ trr.Hash[i]
		} else if Route[trr.Slash[i]+1] == ':' {
			// fmt.Printf("AT %s\n", debug.LF())
			ss += 153
			pp += ":"
			reNames = append(reNames, Route[trr.Slash[i]+2:trr.Slash[i+1]])
			//if dbHash2 {
			//	fmt.Printf(" ss=%d after : 153\n", ss)
			//}
		} else if Route[trr.Slash[i]+1] == '*' {
			// fmt.Printf("AT %s\n", debug.LF())
			ss += 51
			pp += "*"
			//if dbHash2 {
			//	fmt.Printf(" ss=%d after * 51\n", ss)
			//}
			reNames = append(reNames, Route[trr.Slash[i]+2:trr.Slash[i+1]])
			break
		} else if Route[trr.Slash[i]+1] == '{' {
			// fmt.Printf("AT %s\n", debug.LF())
			name, re, valid, convertToColon := parseReFromToken3(Route[trr.Slash[i]+1 : trr.Slash[i+1]])
			_, _, _, _ = name, re, valid, convertToColon
			if convertToColon {
				// fmt.Printf("AT %s\n", debug.LF())
				ss += 153
				pp += ":"
				reNames = append(reNames, name)
				//if dbHash2 {
				//	fmt.Printf(" ss=%d after : 153 / convertToColon, name=%s re=%s valid=%v \n", ss, name, re, valid)
				//}
			} else {
				// fmt.Printf("AT %s\n", debug.LF())
				haveRealRe = true
				ss += 211
				pp += "{"
				aRe := regexp.MustCompile(re)
				tmpRe = append(tmpRe, Re{Pos: i, Re: re, Name: name, cRe: aRe})
				reNames = append(reNames, name)
				//if dbHash2 {
				//	fmt.Printf(" ss=%d after { 211\n", ss)
				//}
			}
		} else {
			// fmt.Printf("AT %s\n", debug.LF())
			ss = ss ^ trr.Hash[i]
			if dbHash2 {
				fmt.Printf(" ss=%d after adding -- trr.Hash[%d]= %d, %s\n", ss, i, trr.Hash[i], debug.LF())
			}
			pp += "T"
		}
	}
	ss = ((ss & bitMask) ^ ((ss >> nBits) & bitMask) ^ ((ss >> (nBits * 2)) & bitMask))
	if dbDumpHashTable {
		if dbX1 {
			fmt.Printf("At, FileName=%s LineNo=%d, %s\n", FileName, LineNo, debug.LF())
			fmt.Printf("%-40s %-8s %-13s %-40s %s\n", "Route", "ss", "m,Method", "FileName", "LineNo")
			fmt.Printf("%-40s %-8s %-13s %-40s %s\n", "---------------------------", "--------", "----------", "------------------------------", "-----")
			dbX1 = false
		}
		// fmt.Printf("After, ss=%-5d m=%4d/%s Url=%s, %s %d, %s\n", ss, m, Method, Route, FileName, LineNo, debug.LF())
		fmt.Printf("%-40s %8d %4d/%-8s %40s %d\n", Route, ss, m, Method, FileName, LineNo)
	}
	// -----------------------------------------------------------------------------------------------------------------
	// xyzzy - if AddToM - then we have RE via extra functions
	// add in stuff to ReSet/tmpRe for AddToM
	// -----------------------------------------------------------------------------------------------------------------
	//if dbHash3 {
	//	fmt.Printf("[%03d] %s %s len(AddToM)=%d\n", NFxNo, Method, Route, len(AddToM))
	//	if haveRealRe {
	//		fmt.Printf("****** Have a real re *******\nreNames=%s tmpRe=%s\n", debug.SVar(reNames), dumpReArray(tmpRe))
	//	}
	//}
	if len(AddToM) > 0 {
		haveRealRe = true
		//		fmt.Printf("****** Have a AddToM/MatchIt re ******* = len(AddToM)=%d\n", len(AddToM))
	}
	if r.Hash2Test[ss] == 0 {
		//if dbHash2 {
		// fmt.Printf("At (no collision) %s\n", debug.LF())
		// fmt.Printf("Adding to empty locaiton in table, trr.Hash2Test[ss]==0\n")
		//}
		r.Hash2Test[ss] = r.nLookupResults
		if haveRealRe {
			///*db*/ fmt.Printf("At %s -- creating ReSet\n", debug.LF())
			r.LookupResults = append(r.LookupResults, Collision2{cType: SingleUrl, Url: Route, NSL: ns, CleanUrl: cleanRoute,
				Hdlr: hdlr, Fx: fx, TPat: pp, FileName: FileName, route_i: NFxNo, LineNo: LineNo, ArgNames: names, MatchIt: AddToM,
				HasRe: []ReList{ReList{Hdlr: hdlr, Fx: fx, ArgNames: reNames, ReSet: tmpRe, MatchIt: AddToM, route_i: NFxNo}}, Comment: Comment})
		} else {
			///*db*/ fmt.Printf("At %s\n", debug.LF())
			r.LookupResults = append(r.LookupResults, Collision2{cType: SingleUrl, Url: Route, NSL: ns, CleanUrl: cleanRoute,
				Hdlr: hdlr, Fx: fx, TPat: pp, FileName: FileName, LineNo: LineNo, ArgNames: names, MatchIt: AddToM, route_i: NFxNo, Comment: Comment})
		}
		// fmt.Printf("At %s\n", debug.LF())
		r.nLookupResults++
		// Hash Collision ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
	} else { // Have a collision on our hands.
		//if dbHash2 {
		// fmt.Printf("At (collision) %s\n", debug.LF())
		// fmt.Printf("At %s -- Collision for %s\n", debug.LF(), Route)
		//}
		c := r.Hash2Test[ss]
		if optionEarlyExit && r.LookupResults[c].cType == IsWord { // No biggie - just a IsWord marker.
			if haveRealRe {
				///*db*/ fmt.Printf("At %s\n", debug.LF())
				r.LookupResults[c] = Collision2{cType: (SingleUrl | IsWord), Url: Route, NSL: ns, CleanUrl: cleanRoute, Hdlr: hdlr,
					Fx: fx, TPat: pp, FileName: FileName, route_i: NFxNo, LineNo: LineNo, ArgNames: names, MatchIt: AddToM,
					HasRe: []ReList{ReList{Hdlr: hdlr, Fx: fx, ArgNames: reNames, ReSet: tmpRe, MatchIt: AddToM, route_i: NFxNo}}, Comment: Comment}
			} else {
				///*db*/ fmt.Printf("At %s\n", debug.LF())
				r.LookupResults[c] = Collision2{cType: (SingleUrl | IsWord), Url: Route, NSL: ns, CleanUrl: cleanRoute,
					Hdlr: hdlr, Fx: fx, TPat: pp, FileName: FileName, route_i: NFxNo, LineNo: LineNo, MatchIt: AddToM,
					ArgNames: names, Comment: Comment}
			}
		} else {
			///*db*/ fmt.Printf("At %s\n", debug.LF())
			//if dbHash2 {
			//	fmt.Printf("+==========================================+\n| Just a collision                         |\n+==========================================+\n")
			//}
			old := r.LookupResults[c]
			if old.HasRe != nil && haveRealRe { // need to check to see if is alreay a RE in old.  If so just append
				///*db*/ fmt.Printf("At %s\n", debug.LF())
				//if dbHash2 {
				//	fmt.Printf("Old - is just a RE, so append it\n")
				//}
				old.HasRe = append(old.HasRe, ReList{Hdlr: hdlr, Fx: fx, ArgNames: reNames, ReSet: tmpRe, MatchIt: AddToM, route_i: NFxNo})
			} else {
				//if dbHash2 {
				//	fmt.Printf("Before Multi Check:cType=%04x,%s %s\n", old.cType, dumpCType(old.cType), debug.LF())
				//}
				///*db*/ fmt.Printf("At %s old.CleanURL=%s cleanRoute=%s\n", debug.LF(), old.CleanUrl, cleanRoute)
				if old.Multi == nil {
					///*db*/ fmt.Printf("At %s\n", debug.LF())
					//if dbHash2 {
					//	fmt.Printf("Multi is NIL\n")
					//}
					old.Multi = make(map[string]Collision2)
					//if dbHash2 {
					//	fmt.Printf("Before:cType=%04x,%s %s\n", old.cType, dumpCType(old.cType), debug.LF())
					//}
					old.cType = (old.cType & (^SingleUrl))
					old.cType |= MultiUrl
					//if dbHash2 {
					//	fmt.Printf("After:cType=%04x,%s\n", old.cType, dumpCType(old.cType))
					//}
					// fmt.Printf("At %s\n", debug.LF())
					old.Multi[old.CleanUrl] = r.LookupResults[c]
				}

				// One of the cleanRoute items is the colliding one - we will need to set the RE on that one.
				if haveRealRe || old.HasRe != nil { // >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> or old has re
					///*db*/ fmt.Printf("At %s - append to end of multi RE if has one\n", debug.LF())
					//if dbHash2 {
					//	fmt.Printf("Have RE in multi colision\n")
					//}
					if xx, multiOk := old.Multi[cleanRoute]; multiOk { // If we have already seen this cleanRoute, then just append the RE
						///*db*/ fmt.Printf("At %s\n", debug.LF())
						//if dbHash2 {
						//	fmt.Printf("HasRe - append case\n")
						//}
						xx.HasRe = append(xx.HasRe, ReList{Hdlr: hdlr, Fx: fx, ArgNames: reNames, ReSet: tmpRe, MatchIt: AddToM, route_i: NFxNo})
						old.Multi[cleanRoute] = xx
					} else {
						///*db*/ fmt.Printf("At %s\n", debug.LF())
						//if dbHash2 {
						//	fmt.Printf("Crea new entry in multi\n")
						//}
						old.Multi[cleanRoute] = Collision2{cType: MultiUrl, Url: Route, NSL: ns, CleanUrl: cleanRoute, Hdlr: hdlr, Fx: fx,
							TPat: pp, FileName: FileName, LineNo: LineNo, ArgNames: names, MatchIt: AddToM, route_i: NFxNo,
							HasRe: []ReList{ReList{Hdlr: hdlr, Fx: fx, ArgNames: reNames, ReSet: tmpRe, MatchIt: AddToM, route_i: NFxNo}}, Comment: Comment}
					}
				} else {
					///*db*/ fmt.Printf("At %s\n", debug.LF())
					//if dbHash2 {
					//	fmt.Printf("NON Re Case - just insert into Multi\n")
					//}
					old.Multi[cleanRoute] = Collision2{cType: MultiUrl, Url: Route, NSL: ns, CleanUrl: cleanRoute, Hdlr: hdlr,
						Fx: fx, TPat: pp, FileName: FileName, LineNo: LineNo, MatchIt: AddToM, ArgNames: names, route_i: NFxNo, Comment: Comment}
				}
			}
			r.LookupResults[c] = old
		}
	}
	//if dbLookupUrlMap {
	//	fmt.Printf("%-10s : %s\n", pp, Route)
	//}
	// fmt.Printf("At -- that's all folks -- %s\n", debug.LF())
}

type UrlAPat struct {
	Pat  string
	Star bool
}

type UrlPat struct {
	PatList []UrlAPat
	// Pat    []string
	// Star   []bool
	PatOcc map[string]int
}

//var nMatch []UrlPat // Index by Length ( NSl )
// var starPat []string // Longer than max NSl => only match to * items

//func init() {
//	nMatch = make([]UrlPat, MaxSlashInUrl, MaxSlashInUrl)
//}

// -------------------------------------------------------------------------------------------------
/*
	Degrees of Freedom, 			Lo .. Hi		sortDf(Pat[i])
		T 		0
		{		1
		:		2
		*		3
*/
func sortDf(Pat string) (rv int) {
	rv = 0
	for _, v := range Pat {
		if v == 'T' {
		} else if v == '{' {
			rv += 1
		} else if v == ':' {
			rv += 2
		} else if v == '*' {
			rv += 3
		}
	}
	return
}

// Return the minimum of 2 integers.
func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// -------------------------------------------------------------------------------------------------

/*
Old - Slower code, structured - see SplitOnSlash3
Benchmark: 148 ns - long, 28.3 ns short.

func (r *MuxRouter) SplitOnSlash2(m uint8, Url string) {
	r.CurUrl = Url
	ln := len(Url)
	NSl := 0
	h := int(m)
	j := 0
	im1 := 0
	for i := 1; i < ln && NSl < MaxSlashInUrl-1; i++ {
		if Url[im1] == '/' || j == nKeyChar {
			h += (h << 3)
			h = h ^ (h >> 11)
			h += (h << 15)
		}
		if Url[im1] == '/' {
			h += j
			h = ((h & bitMask) ^ ((h >> nBits) & bitMask))
			r.Hash[NSl] = h
			r.Slash[NSl] = im1
			NSl++
			h = int(m)
			j = 0
		}
		if j < nKeyChar {
			h += int(Url[i])
			h += (h << 10)
			h = h ^ (h >> 6)
		}
		j++
		im1++
	}
	r.Slash[NSl] = ln
	h += (h << 3)
	h = h ^ (h >> 11)
	h += (h << 15)
	h += j
	h = ((h & bitMask) ^ ((h >> nBits) & bitMask))
	r.Hash[NSl] = h
	r.NSl = NSl + 1
	return
}

*/

/*
// USED
// Test: SplitSlash3_test.go
// Benchmark: 118 for long, 21.1 for short.

PASS
BenchmarkOfSplitOnSlash3_long	20000000	       118 ns/op	       0 B/op	       0 allocs/op
BenchmarkOfSplitOnSlash3_short	100000000	        21.1 ns/op	       0 B/op	       0 allocs/op
*/

// xyzzy-hash
func (r *MuxRouter) SplitOnSlash3(trr *MuxRouterProcessing, m int, Url string, isUrl bool) {
	// fmt.Printf("SplitOnSlash3: Url: %s, m=%d isUrl=%v, %s\n", Url, m, isUrl, debug.LF())
	fixed := false
	var ln, NSl, h, wLen, i int
	var p, eem bool
	eem = true
	pp := true
	// fmt.Printf("\nAt top: ->%s<- %s\n", Url, debug.LF())
	if len(Url) == 0 {
		Url = "/"
	} else if Url[0] != '/' {
		Url = "/" + Url
	}
s0:
	// fmt.Printf("At s0: ->%s<- %s\n", Url, debug.LF())
	trr.CurUrl = Url
	ln = len(Url)
	NSl = 0
	h = m
	wLen = 0
	i = 0
	trr.Hash[NSl] = 0
	trr.Slash[NSl] = 0
	trr.Slash[1] = ln
	NSl++
	p = true

s1:
	// fmt.Printf("At s1: i=%d %s\n", i, debug.LF())
	i++
	if i >= ln {
		//if oneSlash {
		//	fmt.Printf("s1: goto s10, Url=%s i=%d %s\n", Url, i, debug.LF())
		//}
		goto s10
	}
	// if Url[i-1] == '/' && (Url[i] == '.' || Url[i] == '/') {
	if p && (Url[i] == '.' || Url[i] == '/') {
		if pp {
			pp = false
			Url, fixed = r.FixBadUrl(trr, trr.CurUrl)
			if fixed {
				goto s0
			}
		}
	}
	// fmt.Printf("s0: i=%d url ->%s<- wLen=%d\n", i, Url[i:], wLen)
	if wLen >= nKeyChar {
		goto s2a
	}
	if Url[i] == '/' {
		h += wLen
		h += (h << 3)
		h = h ^ (h >> 11)
		h += (h << 15)
		trr.Hash[NSl-1] = h
		trr.Slash[NSl] = i
		NSl++
		trr.Slash[NSl] = ln
		if optionEarlyExit {
			if eem && isUrl {
				eem = false
				if r.Hash2Test[(h^m)&bitMask] == 0 {
					trr.NSl = 1
					goto s11
				}
			}
			eem = false
		}
		if NSl >= MaxSlashInUrl-1 {
			goto s10
		}
		h = m
		wLen = 0
		p = true
		goto s1
	}
	h += int(Url[i])
	h += (h << 10)
	h = h ^ (h >> 6)
	wLen++
	p = false
	goto s1

s2:
	// fmt.Printf("At s2: i=%d %s\n", i, debug.LF())
	i++
	if i >= ln {
		goto s10
	}
s2a:
	// fmt.Printf("At s2a: i=%d %s\n", i, debug.LF())
	if p && (Url[i] == '.' || Url[i] == '/') {
		if pp {
			pp = false
			Url, fixed = r.FixBadUrl(trr, trr.CurUrl)
			if fixed {
				goto s0
			}
		}
	}
	// fmt.Printf("s2a: i=%d url ->%s<- wLen=%d\n", i, Url[i:], wLen)
	if Url[i] == '/' {
		h += wLen
		h += (h << 3)
		h = h ^ (h >> 11)
		h += (h << 15)
		trr.Hash[NSl-1] = h
		trr.Slash[NSl] = i
		NSl++
		trr.Slash[NSl] = ln
		if optionEarlyExit {
			if eem && isUrl {
				eem = false
				if r.Hash2Test[(h^m)&bitMask] == 0 {
					trr.NSl = 1
					goto s11
				}
			}
			eem = false
		}
		if NSl >= MaxSlashInUrl-1 {
			goto s10
		}
		h = m
		wLen = 0
		p = true
		goto s1
	}
	wLen++
	p = false
	goto s2

s10:
	// fmt.Printf("At s10: i=%d %s\n", i, debug.LF())
	// fmt.Printf("s10: wLen=%d\n", wLen)
	if wLen > 0 {
		h += (h << 3)
		h = h ^ (h >> 11)
		h += (h << 15)
		h += wLen
		trr.Hash[NSl-1] = h
		trr.Slash[NSl] = ln
		trr.NSl = NSl
	} else if i == 1 {
		//if oneSlash {
		//	fmt.Printf("Special Case, Url=->%s<- %s\n", Url, debug.LF())
		//}
		trr.Hash[0] = h
		trr.Slash[1] = ln
		trr.NSl = 1
	}
	//if false {
	//	fmt.Printf("Hash=%s Slash=%s NSl=%d\n", debug.SVar(trr.Hash[0:trr.NSl]), debug.SVar(trr.Slash[0:trr.NSl+1]), trr.NSl)
	//}
s11:
	// fmt.Printf("At s11: i=%d pp=%v, trr.Hash[]=%s %s\n", i, pp, debug.SVar(trr.Hash[0:trr.NSl]), debug.LF())
	return
}

// -------------------------------------------------------------------------------------------------
func (r *MuxRouter) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	m := 0

	var procData MuxRouterProcessing
	InitMuxRouterProcessing(r, &procData)

	if rw, ok := www.(*MidBuffer); ok {

		if r.PanicHandler != nil { // 2ns
			defer r.recv(www, req)
		}
		if !r.HasBeenCompiled { // 2ns
			r.CompileRoutes()
		}

		// fmt.Printf("req.Method=%s, %s\n", req.Method, debug.LF())
		// r.AllParam.NParam = 0
		if r.widgetBefore != nil {
			for _, x := range r.widgetBefore {
				// fmt.Printf("req.Method=%s, %s\n", req.Method, debug.LF())
				// m = x.fx(r_www, req, &r.AllParam)
				m = x.fx(rw, req, &rw.Ps)
				// fmt.Printf("req.Method=%s, %s\n", req.Method, debug.LF())
			}
		}

		// fmt.Printf("req.Method=%s, %s\n", req.Method, debug.LF())
		path := req.URL.Path
		Method := req.Method
		// fmt.Printf("req.Method=%s, %s\n", req.Method, debug.LF())
		// m = (int(Method[0]) + (int(Method[1]) << 1))
		m = MethodToCode(Method, 0)

		// fmt.Printf("m=%d/%s, %s\n", m, req.Method, debug.LF())

		if r.widgetHashNewM != nil {
			for _, x := range r.widgetHashNewM {
				// m = x.fx(r_www, req, &r.AllParam)
				m = x.fx(rw, req, &rw.Ps)
			}
		}

		// fmt.Printf("m=%d, %s\n", m, debug.LF())

		r.SplitOnSlash3(&procData, m, path, true)
		found, ln, item := r.LookupUrlViaHash2(&procData, www, req, &m)
		if dbLookup4 {
			fmt.Printf("found=%v, ln=%v item=%v %s\n", found, ln, debug.SVarI(item), debug.LF())
		}
		if found {
			// fmt.Printf("Was Found!  Getting args now\n")
			r.GetArgs4(&procData, path, item.ArgPattern, item.ArgNames, ln, &rw.Ps)
			// fmt.Printf("Was Found!  Calling Fx, params=%s\n", r.AllParam.DumpParam())
			// r.AllParam.route_i = item.route_i
			rw.Ps.route_i = item.route_i
			// fmt.Printf("Found, parsing paras for route_i=%d\n", r.AllParam.route_i)
			// item.Fx(r_www, req, r.AllParam)
			// item.Fx(r_www, req)
			item.Fx(rw, req)
		} else {
			r.NotFound(www, req)
		}

		if r.widgetAfter != nil {
			for _, x := range r.widgetAfter {
				// m = x.fx(r_www, req, &r.AllParam)
				m = x.fx(rw, req, &rw.Ps)
			}
		}

	}

	return
}

func (r *MuxRouter) MatchAndServeHTTP(www http.ResponseWriter, req *http.Request) (Found bool, err error) {

	var procData MuxRouterProcessing
	InitMuxRouterProcessing(r, &procData)

	m := 0
	if !r.HasBeenCompiled { // 2ns
		r.CompileRoutes()
	}
	fmt.Printf("AT: %s\n", debug.LF())

	if rw, ok := www.(*MidBuffer); ok {

		path := req.URL.Path
		Method := req.Method
		// m = (int(Method[0]) + (int(Method[1]) << 1))
		m = MethodToCode(Method, 0)

		// fmt.Printf("AT: %s\n", debug.LF())
		r.SplitOnSlash3(&procData, m, path, true)
		found, ln, item := r.LookupUrlViaHash2(&procData, www, req, &m)
		Found = found
		if dbLookup4 {
			fmt.Printf("found=%v, %s\n", found, debug.LF())
		}
		fmt.Printf("AT: ln=%v item=%v %s\n", ln, debug.SVar(item), debug.LF())
		if found {
			// fmt.Printf("AT: %s\n", debug.LF())
			// fmt.Printf("Was Found!  Getting args now\n")
			r.GetArgs4(&procData, path, item.ArgPattern, item.ArgNames, ln, &rw.Ps) // this pulls arguments from url and puts them into 'r' - bad idea
			// fmt.Printf("Was Found!  Calling Fx, params=%s\n", r.AllParam.DumpParam())
			// r.AllParam.route_i = item.route_i // xyzzyGoFtl01 - Remove in favor of Ps in buffer
			rw.Ps.route_i = item.route_i
			// fmt.Printf("Found, parsing paras for route_i=%d\n", r.AllParam.route_i)
			// fmt.Printf("AT: %s\n", debug.LF())
			if r.DebugMatchOn {
				fmt.Fprintf(os.Stderr, "%sMatched: %s%s%s [%s LineNo:%d] [%s], %s %s\n", MiscLib.ColorCyan, MiscLib.ColorReset, item.Url, MiscLib.ColorCyan, item.FileName, item.LineNo, item.Comment, debug.LF(), MiscLib.ColorReset)
			}
			item.Fx(rw, req)
		}

	} else {
		err = fmt.Errorf("goftlmux: Invalid data type passwed to MatchAdnServeHTTP")
	}

	return
}

// ServeFiles serves files from the given file system root.
// The path must end with "/*filepath", files are then served from the local
// path /defined/root/dir/*filepath.
// For example if root is "/etc" and *filepath is "passwd", the local file
// "/etc/passwd" would be served.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
// To use the operating system's file system implementation,
// use http.Dir:
//     router.ServeFiles("/src/*filepath", http.Dir("/var/www"))
func (r *MuxRouter) ServeFiles(path string, root http.FileSystem) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath")
	}

	fileServer := http.FileServer(root) // xyzzyGoFtl02 - replace with file server from our GoFTL code

	// r.addRoute(w, v.DPath, v.DId, v.DHandlerFunc, i, v.FileName, v.LineNo)
	FileName, LineNo := LineFile(2)
	fx := func(www http.ResponseWriter, req *http.Request) {
		if rw, ok := www.(*MidBuffer); ok {
			req.URL.Path = rw.Ps.ByName("filepath")
			fileServer.ServeHTTP(www, req)
		} else {
			rw.Error = fmt.Errorf("goftlmux: Invalid data type passwed to ServeFiles")
			rw.StatusCode = 500
			return
		}
	}
	k := len(r.routes)
	r.NewRoute().HandleFunc(path, fx)
	r.addRoute("GET", path, 0, fx, k, FileName, LineNo)
}

func (r *MuxRouter) recv(w http.ResponseWriter, req *http.Request) {
	if rcv := recover(); rcv != nil {
		r.PanicHandler(w, req, rcv)
	}
}

func (r *MuxRouter) LookupUrlViaHash2(trr *MuxRouterProcessing, w http.ResponseWriter, req *http.Request, m *int) (found bool, ln int, rv Collision2) {
	Url := trr.CurUrl // xyzzy - mistake
	found = false
	ln = len(Url)
	var ss int
	var cRoute string
	var reMatch bool

	if dbLookupUrlMap {
		fmt.Printf("\n\nLookupUrlViaHash2: Top of Lookup test, Url=%s, %s\n", Url, debug.LF())
	}

	if trr.NSl > minInt(MaxSlashInUrl-1, r.MaxSlash+1) {
		trr.NSl = minInt(MaxSlashInUrl-1, r.MaxSlash+1)
	}
	//if dbLookupUrlMap2 {
	//	fmt.Printf("LookupUrlViaHash2: %s, trr.NSl=%d, len(r.nMatch[%d].PatList)=%d\n", debug.LF(), trr.NSl, trr.NSl, len(r.nMatch[trr.NSl].PatList))
	//	fmt.Printf("LookupUrlViaHash2: r.nMatch[%d]=%s\n", trr.NSl, debug.SVarI(r.nMatch[trr.NSl]))
	//	fmt.Printf("r.nMatch=%s\n", debug.SVarI(r.nMatch))
	//}
	k2 := len(r.nMatch[trr.NSl].PatList)
	//if dbHash2 {
	// fmt.Printf("k2 = %d, trr.NSl=%d, %s\n", k2, trr.NSl, debug.LF())
	//}
	if dbHash2 {
		fmt.Printf("Lookup trr.Hash=%s trr.Slash=%s trr.NSl=%d, %s\n", debug.SVar(trr.Hash[0:trr.NSl]), debug.SVar(trr.Slash[0:trr.NSl+1]), trr.NSl, debug.LF())
	}
	for jj := 0; jj < k2; jj++ {
		ss = 0
		xPat := r.nMatch[trr.NSl].PatList[jj].Pat
		//if dbHash2 {
		// fmt.Printf("Top of Pat Match Loop, jj=%d pat=%s, %s\n", jj, xPat, debug.LF())
		//}
		trr.UsePat = xPat
		for i := 0; i < trr.NSl; i++ {
			if dbHash2 {
				fmt.Printf("i=%d ss=%d, xPat [%s], %s\n", i, ss, xPat, debug.LF())
			}
			if xPat[i] == ':' {
				// fmt.Printf("AT %s\n", debug.LF())
				ss += 153
				//if dbHash2 {
				// fmt.Printf(" ss=%d after : 153, %s\n", ss, debug.LF())
				//}
			} else if xPat[i] == '*' {
				// fmt.Printf("AT %s\n", debug.LF())
				ss += 51
				//if dbHash2 {
				// fmt.Printf(" ss=%d after * 51, %s\n", ss, debug.LF())
				//}
				break
			} else if xPat[i] == '{' {
				// fmt.Printf("AT %s\n", debug.LF())
				ss += 211
				//if dbHash2 {
				// fmt.Printf(" ss=%d after { 211, %s\n", ss, debug.LF())
				//}
			} else {
				// fmt.Printf("AT %s\n", debug.LF())
				ss = ss ^ trr.Hash[i]
				if dbHash2 {
					fmt.Printf(" ss=%d after adding -- trr.Hash[%d] = %d, %s\n", ss, i, trr.Hash[i], debug.LF())
				}
			}
		}
		ss = ((ss & bitMask) ^ ((ss >> nBits) & bitMask) ^ ((ss >> (nBits * 2)) & bitMask))
		if dbLookup4 {
			fmt.Printf("ss=%d, %s\n", ss, debug.LF())
		}
		//if dbLookup4 {
		// for x := range trr.Hash2Test {
		// 	if trr.Hash2Test[x] > 0 {
		// 		fmt.Printf("  **** trr.Hash2Test[%d]=%d\n", x, trr.Hash2Test[x])
		// 	}
		// }
		//}
		if r.Hash2Test[ss] > 0 {
			if dbHash2 {
				fmt.Printf("Found [%d] at ss=%d, Url=%s, %s\n", r.Hash2Test[ss], ss, Url, debug.LF())
			}
			// 1. match all the constants and disambiguate.  If Collision2 - has REFlag set to true ( > 0 Re Sets ) then iterate over them to see if we get a match.
			// xyzzy-widget
			c := r.LookupResults[r.Hash2Test[ss]]
			if (c.cType & SingleUrl) != 0 {
				//if dbHash2 {
				// fmt.Printf("Found a SingleUrl to be true - just return success?? what about RE, %s\n", debug.LF())
				//}
				if c.HasRe != nil { // if we have a RE-List - then Iterate over it for a match.
					///*db*/ fmt.Printf("At %s\n", debug.LF())
					//if dbHash2 {
					//	fmt.Printf("Ok, has RE\n")
					//}
					for k, ww := range c.HasRe { // deal RE at this point too
						///*db*/ fmt.Printf("At %s\n", debug.LF())
						//if dbHash2 {
						//	fmt.Printf("k=%d, len()=%d\n", k, len(c.HasRe))
						//}
						_ = k
						reMatch = true
						for m, x := range ww.ReSet {
							_ = m
							if !x.cRe.MatchString(Url[trr.Slash[x.Pos]+1 : trr.Slash[x.Pos+1]]) {
								//if dbHash2 {
								//	fmt.Printf("Found false match on set k=%d\n", k)
								//}
								reMatch = false
								goto next
							}
						}
						if reMatch {
							///*db*/ fmt.Printf("At %s\n", debug.LF())
							//if dbHash2 {
							//	fmt.Printf("RE match is found\n")
							//}
							// xyzzy-widget -- Final matching on user stuff
							if ww.MatchIt != nil {
								///*db*/ fmt.Printf("At %s\n", debug.LF())
								if r.WidgetMatch(ww.MatchIt, w, req, m, ww.route_i) {
									found = true
									rv.Hdlr = ww.Hdlr
									rv.Fx = ww.Fx
									rv.route_i = ww.route_i
									rv.ArgNames = ww.ArgNames
									return
								}
							} else {
								///*db*/ fmt.Printf("At %s - may be error\n", debug.LF())
								found = true
								rv.Hdlr = ww.Hdlr
								rv.Fx = ww.Fx
								rv.route_i = ww.route_i
								rv.ArgNames = ww.ArgNames
								return
							}
						}
					next:
					}
				} else {
					/* Problem at this locaiton xyzzy xyzzy */
					// r.UrlToCleanRoute3(trr.UsePat) // URL to cRoute? -- URL --
					// if dbHash2 {
					// fmt.Printf("NO RE match is found?? TPat=->%s<-, match v.s. URL ->%s<- CleanUrl ->%s<- cRoute ->%s<-\n", c.TPat, c.Url, c.CleanUrl, cRoute)
					// }
					// if c.CleanUrl == trr.cRoute {
					//if dbHash2 {
					//	fmt.Printf("NO RE match is found?? TPat=->%s<-, match v.s. URL ->%s<- CleanUrl ->%s<-\n", c.TPat, c.Url, c.CleanUrl)
					//}
					if r.CmpUrlToCleanRoute(trr, trr.UsePat, c.CleanUrl) {
						//if dbHash2 {
						// fmt.Printf("   Matched on absolute pattern, returning success\n")
						//}
						// xyzzy-widget -- Final matching on user stuff
						if c.MatchIt != nil {
							if r.WidgetMatch(c.MatchIt, w, req, m, c.route_i) {
								//if dbHash2 {
								//		fmt.Printf("   Widget Match Found\n")
								//}
								found = true
								rv = c
								return
							}
						} else {
							//if dbHash2 {
							//	fmt.Printf("   Match Found\n")
							//}
							found = true
							rv = c
							return
						}
					} // else {
					//if dbHash2 {
					// fmt.Printf("   Did not match, %s\n", debug.LF())
					//}
					// }
				}
			} else if (c.cType & MultiUrl) != 0 {
				// xyzzy - this is where to use NSL
				cRoute = r.UrlToCleanRoute(trr, trr.UsePat) // xyzzy- will alloc memory on call- URL to cRoute? -- URL --
				if c2, ok := c.Multi[cRoute]; ok {
					if c2.HasRe != nil { // if we have a RE-List - then Iterate over it for a match.
						///*db*/ fmt.Printf("At %s\n", debug.LF())
						for k, ww := range c2.HasRe { // deal RE at this point too
							_ = k
							reMatch = true
							///*db*/ fmt.Printf("At %s\n", debug.LF())
							for m, x := range ww.ReSet {
								_ = m
								if !x.cRe.MatchString(Url[trr.Slash[x.Pos]+1 : trr.Slash[x.Pos+1]]) {
									reMatch = false
									goto next2
								}
							}
							if reMatch {
								// xyzzy-widget -- Final matching on user stuff
								if ww.MatchIt != nil {
									///*db*/ fmt.Printf("At %s\n", debug.LF())
									if r.WidgetMatch(ww.MatchIt, w, req, m, ww.route_i) {
										found = true
										rv.Hdlr = ww.Hdlr
										rv.Fx = ww.Fx
										rv.route_i = ww.route_i
										rv.ArgNames = ww.ArgNames
										return
									}
								} else {
									///*db*/ fmt.Printf("At %s\n", debug.LF())
									found = true
									rv.Hdlr = ww.Hdlr
									rv.Fx = ww.Fx
									rv.route_i = ww.route_i
									rv.ArgNames = ww.ArgNames
									return
								}
							}
						next2:
						}
					} else {
						// xyzzy-widget -- Final matching on user stuff
						if c2.MatchIt != nil {
							if r.WidgetMatch(c2.MatchIt, w, req, m, c2.route_i) {
								found = true
								rv = c2
								return
							}
						} else {
							found = true
							rv = c2
							return
						}
					}
				}
			}
		}
	}
	return
}

// xyzzyGoFTL10
// This is the function that applies the extended match functions

// xyzzy - not take into account ReList -
// xyzzy - remove m *int param?? - not used
// xyzzy - remov eMatchIt[i].Data?? - not used
func (r *MuxRouter) WidgetMatch(MatchIt []Match, www http.ResponseWriter, req *http.Request, m *int, route_i int) bool {
	if MatchIt != nil {
		for i, v := range MatchIt {
			_ = i
			// b := v.MatchFunc(req, r, v.Data)
			b := v.MatchFunc(www, req, r, route_i)
			// fmt.Printf("MatchFunc [%d] == %v, with route_i = %d, req.RequestURI=%s, %s\n", i, b, route_i, req.RequestURI, debug.LF())
			if !b {
				return false
			}
		}
	}
	return true
}

// Input:  Pattern like T::T and the current URL with Slash locaiton information.
// So... /abc/:def/ghi is the Route, /abc/x/ghi is the ULR, Slash is [ 0, 4, 6, 10 ]
// The output is /abc/:/ghi - Sutiable for lookup in a map of cleanRoute
func (r *MuxRouter) UrlToCleanRoute(trr *MuxRouterProcessing, UsePat string) (rv string) {
	for i, v := range UsePat { // Pat is T::T format pattern
		if v == ':' {
			rv += "/:"
		} else if v == '*' {
			rv += "/*"
			break
		} else if v == '{' {
			rv += "/{"
		} else {
			rv += "/" + trr.CurUrl[trr.Slash[i]+1:trr.Slash[i+1]]
		}
	}
	return
}

// compate r.CurUrl to a Pattern
func (r *MuxRouter) CmpUrlToCleanRoute(trr *MuxRouterProcessing, UsePat string, CleanUrl string) (rv bool) {
	rv = true
	// for i, v := range UsePat { // Pat is T::T format pattern
	k := 1 // Index into CleanUrl
	//if dbCmp {
	//	fmt.Printf("UsePat ->%s<-\n", UsePat)
	//}
	for i := 0; i < len(UsePat); i++ {
		v := UsePat[i]
		//if dbCmp {
		//	fmt.Printf("k=%d i=%d, UsePat[i]=%c\n", k, i, UsePat[i])
		//}
		if v == ':' {
			k += 2
		} else if v == '*' {
			break
		} else if v == '{' {
			k += 2
		} else {
			// r.cRoute += "/" + r.CurUrl[r.Slash[i]+1:r.Slash[i+1]]
			l := trr.Slash[i+1] - trr.Slash[i] - 1
			//if dbCmp {
			//	fmt.Printf("T: l=%d\n", l)
			//	fmt.Printf("T: left=->%s<-\n", r.CurUrl[r.Slash[i]+1:r.Slash[i+1]])
			//	fmt.Printf("T: rght=->%s<-\n", CleanUrl[k:k+l])
			//}
			//fmt.Printf("\nX: l=%d\n", l)
			//fmt.Printf("X: i=%d\n", i)
			//fmt.Printf("X: len(r.Slash)=%d\n", len(r.Slash))
			//fmt.Printf("X: len(r.CurUrl)=%d\n", len(r.CurUrl))
			//fmt.Printf("X: r.Slash[i]= %d\n", r.Slash[i])
			//fmt.Printf("X: r.Slash[i+1]= %d\n", r.Slash[i+1])
			//fmt.Printf("X: k=%d, k+l=%d, len(CleanUrl)=%d\n", k, k+l, len(CleanUrl))
			//fmt.Printf("X: CleanUrl=->%s<-\n", CleanUrl)
			//fmt.Printf("X: r.CurUrl=->%s<-\n", r.CurUrl)
			m := k + l
			if m > len(CleanUrl) {
				m = len(CleanUrl)
			}
			//fmt.Printf("X: m=%d\n", m)
			if trr.CurUrl[trr.Slash[i]+1:trr.Slash[i+1]] != CleanUrl[k:m] {
				//if dbCmp {
				//	fmt.Printf("match failed\n")
				//}
				rv = false
				return
			}
			/*
			   X: l=18
			   X: i=2
			   X: len(trr.Slash)=21
			   X: len(trr.CurUrl)=29
			   X: trr.Slash[i]= 10
			   X: trr.Slash[i+1]= 29
			   X: k=11, k+l=29, len(CleanUrl)=12
			   X: CleanUrl=->/api/table/:<-
			   X: trr.CurUrl=->/api/table/user_valid_origins<-
			*/
			k += l + 1
		}
	}
	return
}

// Find the last index of the character 'ch' in 's'.
// Example
//   n := LastIndexOfChar( "[123:456]:80", ':' )
// will return 9
// -1 is returned if not found.
func LastIndexOfChar(s string, ch uint8) int {
	for i := len(s) - 1; i >= 0; i-- {
		if ch == s[i] {
			return i
		}
	}
	return -1
}

/*
// From Gorilla-Mux
// uniqueVars returns an error if two slices contain duplicated strings.
func uniqueVars(s1, s2 []string) error {
	for _, v1 := range s1 {
		for _, v2 := range s2 {
			if v1 == v2 {
				return fmt.Errorf("mux: duplicated route variable %q", v2)
			}
		}
	}
	return nil
}
*/

// From Gorilla-Mux
// mapFromPairs converts variadic string parameters to a string map.
func mapFromPairs(pairs ...string) (map[string]string, error) {
	length := len(pairs)
	if length%2 != 0 {
		return nil, fmt.Errorf(
			"mux: number of parameters must be multiple of 2, got %v", pairs)
	}
	m := make(map[string]string, length/2)
	for i := 0; i < length; i += 2 {
		m[pairs[i]] = pairs[i+1]
	}
	return m, nil
}

/*
// From Gorilla-Mux
// matchInArray returns true if the given string value is in the array.
func matchInArray(arr []string, value string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}
*/

// From Gorilla-Mux
// matchMap returns true if the given key/value pairs exist in a given map.
func matchMap(toCheck map[string]string, toMatch map[string][]string, canonicalKey bool) bool {
	for k, v := range toCheck {
		// Check if key exists.
		if canonicalKey {
			k = http.CanonicalHeaderKey(k)
		}
		if values := toMatch[k]; values == nil {
			return false
		} else if v != "" {
			// If value was defined as an empty string we only check that the
			// key exists. Otherwise we also check for equality.
			valueExists := false
			for _, value := range values {
				if v == value {
					valueExists = true
					break
				}
			}
			if !valueExists {
				return false
			}
		}
	}
	return true
}

func matchQueryMap(patternMap map[string]string, ps Params) bool {
	ps.CreateSearch()
	// fmt.Printf(">>>>>>>>>>>>>>>>> matchQueryMap patternMap=%s ps=%s\n", debug.SVar(patternMap), ps.DumpParam())
	for i, v := range patternMap {
		// fmt.Printf("checking ->%v<- ->%v<-\n", i, v)
		vv := ps.ByName(i)
		// fmt.Printf("vv=%s\n", vv)
		if vv != v {
			return false
		}
	}
	// fmt.Printf("Returingin True!\n")
	return true
}

func IsMapStringBoolEmpty(v map[string]bool) bool {
	for _, _ = range v {
		return false
	}
	return true
}

func dumpCType(n colType) (rv string) {
	rv = "("
	com := ""
	if (n & IsWord) != 0 {
		rv += com + "IsWord"
		com = "|"
	}
	if (n & MultiUrl) != 0 {
		rv += com + "MultiUrl"
		com = "|"
	}
	if (n & SingleUrl) != 0 {
		rv += com + "SingleUrl"
		com = "|"
	}
	//if (n & BadUrl) != 0 {
	//	rv += com + "BadUrl"
	//	com = "|"
	//}
	if (n & Dummy) != 0 {
		rv += com + "Dummy"
		com = "|"
	}
	rv += ")"
	return
}

func ServeHTTP(fx func(w http.ResponseWriter, r *http.Request)) (rv func(res http.ResponseWriter, req *http.Request, ps Params)) {
	return func(w http.ResponseWriter, r *http.Request, ps Params) {
		// maybee set ps -> Gorilla:context?
		fx(w, r)
	}
}

func (r *MuxRouter) DumpRouteData(msg string) {
	fmt.Printf("\nDumpRouteData: %s, %s\n", msg, debug.LF())
	for i, v := range r.routeData {
		if v.MatchItRank != 0 {
			fmt.Printf("[%03d] ARoute[%d] %s %s 0x%04x MatchItRank=%s\n", i, v.NFxNo, v.Method, v.Route, v.MatchItRank, DumpMatchItRan(v.MatchItRank))
		} else {
			fmt.Printf("[%03d] ARoute[%d] %s %s\n", i, v.NFxNo, v.Method, v.Route)
		}
	}
	fmt.Printf("\n\n")
}

func DumpMatchItRan(x MatchItRankType) (rv string) {
	com := ""
	rv = "{ "
	if (x & ReMatch) != 0 {
		rv += com + " ReMatch"
		com = "|"
	}
	if (x & HeaderMatch) != 0 {
		rv += com + " HeaderMatch"
		com = "|"
	}
	if (x & QueryMatch) != 0 {
		rv += com + " QueryMatch"
		com = "|"
	}
	if (x & TLSMatch) != 0 {
		rv += com + " TLSMatch"
		com = "|"
	}
	if (x & PortMatch) != 0 {
		rv += com + " PortMatch"
		com = "|"
	}
	if (x & HostMatch) != 0 {
		rv += com + " HostMatch"
		com = "|"
	}
	if (x & PortHostMatch) != 0 {
		rv += com + " PortHostMatch"
		com = "|"
	}
	if (x & ProtocalMatch) != 0 {
		rv += com + " ProtocalMatch"
		com = "|"
	}
	if (x & User0Match) != 0 {
		rv += com + " User0Match"
		com = "|"
	}
	if (x & User1Match) != 0 {
		rv += com + " User1Match"
		com = "|"
	}
	if (x & User2Match) != 0 {
		rv += com + " User2Match"
		com = "|"
	}
	if (x & User3Match) != 0 {
		rv += com + " User3Match"
		com = "|"
	}
	if (x & User4Match) != 0 {
		rv += com + " User4Match"
		com = "|"
	}
	rv += " }"
	return
}

func dumpReArray(tmpRe []Re) (rv string) {
	rv = "{{ "
	com := ""
	for i, v := range tmpRe {
		rv += com + fmt.Sprintf(" at[%d] Pos=%d Re=%s Name=%s ", i, v.Pos, v.Re, v.Name)
		com = ","
	}
	rv += " }}"
	return
}

const dbLookup4 = false
const dbHash2 = false
const dbLookupUrlMap = false
const dbMatch2 = false

// If you can't find out why something is not matching in the hash - then turn this on and look in output
// for the has table - printed out.  Usually it is obvious like not allowing "GET" for a endpoint.
const dbDumpHashTable = false

var dbX1 = true

// if r.DebugMatchOn {

// const oneSlash = false
