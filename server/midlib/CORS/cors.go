//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2017
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1246
//

//
// Copyright (C) Philip Schlump, 2017
//

// TODO
// 1. Add in a "Redis" version that adds a prefix like "CORS:" and if CORS:<Origin> is a key with a a "ok" value then
//		it is a valid CORS domain.
//		1. Connection info for connect to Redis
//		2. { "valid": "yes", "add_to_request": { ... } }
//		3.

package CORS

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	JsonX "github.com/pschlump/JSONx"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	logrus "github.com/pschlump/pslog" // "github.com/sirupsen/logrus"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &CORSType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("CORS", CreateEmpty, `{
		"Paths":    	    	{ "type":["string","filepath"], "isarray":true, "required":true },
		"CookiePolyfilPaths": 	{ "type":["string","filepath"], "isarray":true, "default":"/api/cors-session-cookie-polyfil.js" },
		"AllowedOrigins": 		{ "type":["string"], "isarray":true },
		"AllowedHeaders":		{ "type":["string"], "isarray":true },
		"AllowedMethods":		{ "type":["string"], "isarray":true },
		"ExposedHeaders":		{ "type":["string"], "isarray":true },
		"AllowCredentials": 	{ "type":[ "bool" ], "default":"false" },
		"AllowOriginFunc":		{ "type":["string"] },
		"OptionsPassthrough": 	{ "type":[ "bool" ], "default":"false" },
		"RedisPrefix":          { "type":[ "string" ], "required":false, "default":"" },
		"GetOriginURI":         { "type":[ "string" ], "required":false, "default":"/api/cors/GetOrigin" },
		"Debug01": 				{ "type":[ "bool" ], "default":"true" },
		"MaxAge":      	  		{ "type":[ "int" ], "default":"86400" },
		"LineNo":        		{ "type":[ "int" ], "default":"1" }
		}`)
}

// -----------------------------------------------------------------------------------------------------------------------------------------------------------
// Suppor fot Setting Cookies
//	http://caniuse.com/#feat=xhr2 -- requries special option and this version -- 90% support
//	http://stackoverflow.com/questions/14221722/set-cookie-on-browser-with-ajax-request-via-cors -- Details on special option
// Could be supported on other 10% via client side action to set cookie. /api/create-session-cookie?X-Go-FTL-Trx-Id=xxxxxxx + Set of cookie by self (polyfill)?
//	if no cookie, then load .js file from server (In iFrame 0px by 0px), with SetCookie in JS code, then can do Ajax from that site
//	<script src="http://server:port/api/create-session-cookie-polyfil.js"></script>

const cookiePolyfilJs = `

;

function HashCode( str ) {
	var hash = 0, i, chr;
	if (str.length === 0) return hash;
	for (i = 0; i < str.length; i++) {
		chr	 = str.charCodeAt(i);
		hash	= ((hash << 5) - hash) + chr;
		hash |= 0; // Convert to 32bit integer
	}
	return hash;
}

function GenerateSecureRandomGUID () {
	var buf = new Uint16Array(8);
	window.crypto.getRandomValues(buf);
	var S4 = function(num) {
			var ret = num.toString(16);
			while(ret.length < 4){
					ret = "0"+ret;
			}
			return ret;
	};
	return (S4(buf[0])+S4(buf[1])+"-"+S4(buf[2])+"-"+S4(buf[3])+"-"+S4(buf[4])+"-"+S4(buf[5])+S4(buf[6])+S4(buf[7]));
}

function GenerateRegularRandomGUID () {
	var d = new Date().getTime();
	if (typeof performance !== 'undefined' && typeof performance.now === 'function'){
		d += performance.now(); //use high-precision timer if available
	}
	d ^= HashCode(navigator.userAgent)	// should prevent monotonically increasing values
	return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
		var r = (d + Math.random() * 16) % 16 | 0;
		d = Math.floor(d / 16);
		return (c === 'x' ? r : (r & 0x3 | 0x8)).toString(16);
	});
}

const GenerateGUID = ( (typeof(window.crypto) != 'undefined' && typeof(window.crypto.getRandomValues) != 'undefined') ) ?
		GenerateSecureRandomGUID : GenerateRegularRandomGUID ;

function createCookie(name,value,days) {
	var expires = "";
	if (days) {
		var date = new Date();
		date.setTime(date.getTime()+(days*24*60*60*1000));
		expires = "; expires="+date.toGMTString();
	} 
	document.cookie = name+"="+value+expires+"; path=/";
}

function getCookie(name) {
	var nameEQ = name + "=";
	var ca = document.cookie.split(';');
	for(var i=0;i < ca.length;i++) {
		var c = ca[i];
		while (c.charAt(0)==' ') c = c.substring(1,c.length);
		if (c.indexOf(nameEQ) == 0) {
			return c.substring(nameEQ.length,c.length);
		}
	}
	return null;
}

function doIt(name) {
	var cookieName = "X-Go-FTL-Trx-Id";
	if ( getCookie(cookieName) === "" ) {
		var id = GenerateGUID();
		createCookie(cookieName,id);
	}
}

doIt();

`

func (hdlr *CORSType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init

	gCfg.ConnectToRedis()
	hdlr.gCfg = gCfg
	var ok bool

	hdlr.exposedHeaders = convert(hdlr.ExposedHeaders, http.CanonicalHeaderKey)
	if hdlr.AllowOriginFunc != "" {
		hdlr.allowOriginFunc, ok = ValidationMap[hdlr.AllowOriginFunc]
		if !ok {
			err = fmt.Errorf("Error: invalid validation function - not defined [%s]", hdlr.AllowOriginFunc)
			return
		}
	}
	if hdlr.RedisPrefix != "" {
		// fmt.Printf("CORS: setting up a redix_prefix based origin check\n")
		hdlr.allowOriginFunc, ok = ValidationMap["redis_prefix"]
		if !ok {
			err = fmt.Errorf("Error: invalid validation function - not defined [%s]", hdlr.AllowOriginFunc)
			return
		}
	}
	hdlr.allowCredentials = hdlr.AllowCredentials
	// hdlr.OptionPassthrough = false		// Just in case somebody else wants to see the "OPTIONS" request?

	// Normalize options
	// Note: for origins and methods matching, the spec requires a case-sensitive matching.
	// As it may error prone, we chose to ignore the spec here.

	// Allowed Origins
	if len(hdlr.AllowedOrigins) == 0 {
		// Default is localhost - If you just include CORS as a layer then localhost should be allowed.
		// Not certain that this is s a good idea!
		origin := "^http[s]?://localhost(:[0-9]*)?$"
		hdlr.allowedOrigins = append(hdlr.allowedOrigins, origin)
		hdlr.allowedOriginsRE = append(hdlr.allowedOriginsRE, regexp.MustCompile(origin))
	} else {
		hdlr.allowedOrigins = []string{} // xyzzy - pre-allocate
		hdlr.allowedOriginsRE = make([]*regexp.Regexp, 0, len(hdlr.AllowedOrigins))
		for _, origin := range hdlr.AllowedOrigins {
			// Normalize
			origin = strings.ToLower(origin)
			if origin == "*" || origin == ".*" {
				// If ".*" is present in the list, turn the whole list into a match all
				hdlr.allowedOriginsAll = true
				hdlr.allowedOrigins = nil
				hdlr.allowedOriginsRE = nil
				break
			} else {
				hdlr.allowedOrigins = append(hdlr.allowedOrigins, origin)
				hdlr.allowedOriginsRE = append(hdlr.allowedOriginsRE, regexp.MustCompile(origin))
			}
		}
	}

	// Allowed Headers
	if len(hdlr.AllowedHeaders) == 0 {
		// Use sensible defaults
		hdlr.allowedHeaders = []string{"Origin", "Accept", "Content-Type"}
	} else {
		// Origin is always appended as some browsers will always request for this header at preflight
		hdlr.allowedHeaders = convert(append(hdlr.AllowedHeaders, "Origin"), http.CanonicalHeaderKey)
		for _, h := range hdlr.AllowedHeaders {
			if h == "*" || h == ".*" {
				if hdlr.gCfg.DbOn("*", "CORS", "db-headers-allowed") {
					fmt.Fprintf(os.Stderr, "AT: -- All Headers Allowed -- %s\n", godebug.LF())
				}
				hdlr.allowedHeadersAll = true
				hdlr.allowedHeaders = nil
				break
			}
		}
	}

	// Allowed Methods
	if len(hdlr.AllowedMethods) == 0 {
		// Default is NOT spec's "simple" methods -- That would be GET POST
		hdlr.allowedMethods = []string{"GET", "POST", "PUT", "DELETE"}
	} else {
		hdlr.allowedMethods = convert(hdlr.AllowedMethods, strings.ToUpper)
	}
	return
}

func (hdlr *CORSType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*CORSType)(nil)

type RedisCORSType struct {
	Valid        string            `json:"Valid"`
	AddToRequest map[string]string `json:"AddToRequest"`
}

// --------------------------------------------------------------------------------------------------------------------------
func ValidateWithRedisPrefix(req *http.Request, rw *goftlmux.MidBuffer, hdlr *CORSType) bool {

	origin := req.Header.Get("origin")
	// prefix := hdlr.RedisPrefix
	// fmt.Printf("CORS: Validating with redis : Prefix [%s] origin [%s], %s\n", prefix, origin, godebug.LF())
	val, err := hdlr.RedisGetValidOrigin(hdlr.RedisPrefix + origin)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s request from origin(%s) that did not validate.","LineFile":%q}`+"\n", err, origin, godebug.LF()))
		return false
	}
	// fmt.Printf("CORS: AT %s\n", godebug.LF())
	var dv RedisCORSType
	err = json.Unmarshal([]byte(val), &dv)
	if err != nil {
		// fmt.Printf("CORS: AT %s\n", godebug.LF())
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s xyzzy.","LineFile":%q}`+"\n", err, godebug.LF()))
		return false
	}
	// fmt.Printf("CORS: dv.Valid [%s] AT %s\n", dv.Valid, godebug.LF())
	if dv.Valid == "yes" {
		// add in AddToRequest to the request params.
		ps := &rw.Ps
		for key, value := range dv.AddToRequest {
			// fmt.Printf("CORS: key/value = %s/%s AT %s\n", key, value, godebug.LF())
			goftlmux.AddValueToParams(key, value, 'i', goftlmux.FromInject, ps)
		}
		return true
	}
	return false
}

type ValidationFunc func(req *http.Request, rw *goftlmux.MidBuffer, hdlr *CORSType) bool

var ValidationMap map[string]ValidationFunc

func init() {
	ValidationMap = make(map[string]ValidationFunc)
	ValidationMap["redis_prefix"] = ValidateWithRedisPrefix
}

func SetValidationFunction(name string, fx ValidationFunc) {
	ValidationMap[name] = fx
}

func (hdlr *CORSType) RedisGetValidOrigin(key string) (val string, err error) {
	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return "{}", err
	}

	v, err := conn.Cmd("GET", key).Str()
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis key [%s].","LineFile":%q}`+"\n", err, key, godebug.LF()))
		return "{}", err
	}

	val = v
	err = nil
	return
}

// allowOriginFunc   func(origin string) bool // Optional origin validation function
type CORSType struct {
	Next               http.Handler
	Paths              []string                    `gfType:"string,filepath" gfRequired:"true"`
	CookiePolyfilPaths []string                    `gfType:"string,filepath" gfDefault:"/api/cors-session-cookie-polyfil.js"`
	AllowedOrigins     []string                    `gfDefault:""` // Array of RE to match
	AllowedMethods     []string                    `gfDefault:""`
	ExposedHeaders     []string                    `gfDefault:""`
	AllowCredentials   bool                        `gfDefault:"false"`
	AllowOriginFunc    string                      `gfDefault:""`
	AllowedHeaders     []string                    `gfDefault:""`
	MaxAge             int                         `gfDefault:"86400"`
	GetOriginURI       string                      `gfDefault:"/api/cors/GetOrigin"`
	OptionPassthrough  bool                        `gfDefault:"false"`
	LineNo             int                         `gfDefault:"1"`
	allowedOriginsAll  bool                        // Set to true when allowed origins contains a "*"
	allowedOrigins     []string                    // Normalized list of plain allowed origins
	allowedOriginsRE   []*regexp.Regexp            // Converted to RT's to match with
	allowOriginFunc    ValidationFunc              // Optional origin validation function
	allowedHeadersAll  bool                        // Set to true when allowed headers contains a "*"
	allowedHeaders     []string                    // Normalized list of allowed headers
	allowedMethods     []string                    // Normalized list of allowed methods
	exposedHeaders     []string                    // Normalized list of exposed headers -- Should be "" or joined -- xyzzy
	exposedHeadersStr  string                      // Normalized list of exposed headers - strings.Join(hdlr.exposedHeaders, ", ") // xyzzy
	allowCredentials   bool                        //
	Debug01            bool                        `gfDefault:"true"`
	RedisPrefix        string                      `gfDefault:""`
	gCfg               *cfg.ServerGlobalConfigType //
}

func NewCORSServer(n http.Handler, p []string, ml int) *CORSType {
	return &CORSType{Next: n, Paths: p}
}

func (hdlr *CORSType) ServeHTTP(www http.ResponseWriter, req *http.Request) {

	if pn := lib.PathsMatchN(hdlr.CookiePolyfilPaths, req.URL.Path); pn >= 0 {
		www.Header().Set("Content-Type", "application/javascript")
		fmt.Fprintf(www, cookiePolyfilJs)
		return
	}
	if pn := lib.PathsMatchN([]string{hdlr.GetOriginURI}, req.URL.Path); pn >= 0 {
		headers := www.Header()
		origin := req.Header.Get("Origin")
		headers.Set("Access-Control-Allow-Origin", origin)
		if len(hdlr.exposedHeaders) > 0 {
			headers.Set("Access-Control-Expose-Headers", strings.Join(hdlr.exposedHeaders, ", ")) // xyzzy
		}
		if hdlr.allowCredentials {
			headers.Set("Access-Control-Allow-Credentials", "true")
		}
		// DONE: check Redis vorig: to see if this is a valid origin.
		isValid := false
		if rw, ok := www.(*goftlmux.MidBuffer); ok {
			isValid = ValidateWithRedisPrefix(req, rw, hdlr)
		}
		hdlr.logf("  Success Get Origin: %s : Actual response added headers: %v", hdlr.GetOriginURI, headers)
		www.Header().Set("Content-Type", "application/json")
		www.WriteHeader(http.StatusOK)
		fmt.Fprintf(www, `{"status":"success","origin":%q,"valid":%v}`+"\n", origin, isValid)
		return
	}
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "CORS", hdlr.Paths, pn, req.URL.Path)

			// ---------------------------------------------------------------------------------------------------
			if req.Method == "OPTIONS" {
				fmt.Fprintf(os.Stderr, "\n%s\n------------------------------------------------- OPTIONS -------------------------------------------------\nAT:%s\n%s\n%s\n", MiscLib.ColorCyan, godebug.LF(), godebug.SVarI(req), MiscLib.ColorReset)
				hdlr.logf("ServeHTTP: Preflight request")
				hdlr.handlePreflight(www, rw, req)
				// Preflight requests are standalone and should stop the chain as some other
				// middleware may not handle OPTIONS requests correctly. One typical example
				// is authentication middleware ; OPTIONS requests won't carry authentication
				// headers (see #1)
				if hdlr.OptionPassthrough {
					hdlr.Next.ServeHTTP(www, req)
				} else {
					www.WriteHeader(http.StatusOK)
				}
			} else {
				fmt.Fprintf(os.Stderr, "\n%s\n------------------------------------------------- %s -------------------------------------------------\nAT:%s\n%s\n%s\n", MiscLib.ColorCyan, req.Method, godebug.LF(), godebug.SVarI(req), MiscLib.ColorReset)
				hdlr.logf("ServeHTTP: Actual request")
				hdlr.handleActualRequest(www, rw, req)
				hdlr.Next.ServeHTTP(www, req)
			}

		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	hdlr.Next.ServeHTTP(www, req)

}

// handleActualRequest handles simple cross-origin requests, actual request or redirects
func (hdlr *CORSType) handleActualRequest(www http.ResponseWriter, rw *goftlmux.MidBuffer, req *http.Request) {
	headers := www.Header()
	origin := req.Header.Get("Origin")

	if req.Method == "OPTIONS" {
		hdlr.logf("  Actual request no headers added: method == %s", req.Method)
		return
	}
	// Always set Vary, see https://github.com/rs/cors/issues/10
	headers.Add("Vary", "Origin")
	if origin == "" {
		hdlr.logf("  Actual request no headers added: missing origin")
		return
	}

	if !hdlr.isOriginAllowed(origin, rw, req) {
		hdlr.logf("  Actual request no headers added: origin '%s' not allowed", origin)
		return
	}

	// Note that spec does define a way to specifically disallow a simple method like GET or
	// POST. Access-Control-Allow-Methods is only used for pre-flight requests and the
	// spec doesn't instruct to check the allowed methods for simple cross-origin requests.
	// We think it's a nice feature to be able to have control on those methods though.
	if !hdlr.isMethodAllowed(req.Method) {
		hdlr.logf("  Actual request no headers added: method '%s' not allowed", req.Method)

		return
	}
	headers.Set("Access-Control-Allow-Origin", origin)
	if len(hdlr.exposedHeaders) > 0 { // xyzzy
		headers.Set("Access-Control-Expose-Headers", strings.Join(hdlr.exposedHeaders, ", ")) // xyzzy
	}
	if hdlr.allowCredentials {
		headers.Set("Access-Control-Allow-Credentials", "true")
	}
	hdlr.logf("  Actual response added headers: %v", headers)
}

// handlePreflight handles pre-flight CORS requests
func (hdlr *CORSType) handlePreflight(www http.ResponseWriter, rw *goftlmux.MidBuffer, req *http.Request) {
	headers := www.Header()
	origin := req.Header.Get("Origin")

	if req.Method != "OPTIONS" {
		hdlr.logf("  Preflight aborted: %s!=OPTIONS", req.Method)
		return
	}
	// Always set Vary headers
	// see https://github.com/rs/cors/issues/10,
	//     https://github.com/rs/cors/commit/dbdca4d95feaa7511a46e6f1efb3b3aa505bc43f#commitcomment-12352001
	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")

	if origin == "" {
		hdlr.logf("  Preflight aborted: empty origin")
		return
	}
	if !hdlr.isOriginAllowed(origin, rw, req) {
		hdlr.logf("  Preflight aborted: origin '%s' not allowed", origin)
		return
	}

	reqMethod := req.Header.Get("Access-Control-Request-Method")
	if !hdlr.isMethodAllowed(reqMethod) {
		if hdlr.gCfg.DbOn("*", "LoginRequired", "db2") {
			fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
		}
		hdlr.logf("  Preflight aborted: method '%s' not allowed", reqMethod)
		return
	}
	reqHeaders := parseHeaderList(req.Header.Get("Access-Control-Request-Headers"))
	if !hdlr.areHeadersAllowed(reqHeaders) {
		if hdlr.gCfg.DbOn("*", "LoginRequired", "db2") {
			fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
		}
		hdlr.logf("  Preflight aborted: headers '%v' not allowed", reqHeaders)
		return
	}
	if hdlr.gCfg.DbOn("*", "LoginRequired", "db2") {
		fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
	}
	headers.Set("Access-Control-Allow-Origin", origin)
	// Spec says: Since the list of methods can be unbounded, simply returning the method indicated
	// by Access-Control-Request-Method (if supported) can be enough
	headers.Set("Access-Control-Allow-Methods", strings.ToUpper(reqMethod))
	if len(reqHeaders) > 0 {

		if hdlr.gCfg.DbOn("*", "LoginRequired", "db2") {
			fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
		}
		// Spec says: Since the list of headers can be unbounded, simply returning supported headers
		// from Access-Control-Request-Headers can be enough
		headers.Set("Access-Control-Allow-Headers", strings.Join(reqHeaders, ", "))
	}
	if hdlr.gCfg.DbOn("*", "LoginRequired", "db2") {
		fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
	}
	if hdlr.allowCredentials {
		if hdlr.gCfg.DbOn("*", "LoginRequired", "db2") {
			fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
		}
		headers.Set("Access-Control-Allow-Credentials", "true")
	}
	if hdlr.MaxAge > 0 {
		if hdlr.gCfg.DbOn("*", "LoginRequired", "db2") {
			fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
		}
		headers.Set("Access-Control-Max-Age", strconv.Itoa(hdlr.MaxAge))
	}
	if hdlr.gCfg.DbOn("*", "LoginRequired", "db2") {
		fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
	}
	hdlr.logf("  Preflight response headers: %v", headers)
}

// convenience method. checks if debugging is turned on before printing
func (hdlr *CORSType) logf(format string, a ...interface{}) {
	// if c.Log != nil {
	if hdlr.gCfg.DbOn("*", "LoginRequired", "db1") {
		fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF(2))
		fmt.Fprintf(os.Stderr, "\t"+format+"\n", a...)
		fmt.Fprintf(os.Stdout, "AT: %s\n", godebug.LF(2))
		fmt.Fprintf(os.Stdout, "\t"+format+"\n", a...)
	}
}

// isOriginAllowed checks if a given origin is allowed to perform cross-domain requests
// on the endpoint
func (hdlr *CORSType) isOriginAllowed(origin string, rw *goftlmux.MidBuffer, req *http.Request) bool {
	if hdlr.allowOriginFunc != nil {
		return hdlr.allowOriginFunc(req, rw, hdlr)
	}
	if hdlr.allowedOriginsAll {
		return true
	}
	origin = strings.ToLower(origin)
	for _, w := range hdlr.allowedOriginsRE {
		if w.MatchString(origin) {
			return true
		}
	}
	return false
}

// isMethodAllowed checks if a given method can be used as part of a cross-domain request on the endpoint
func (hdlr *CORSType) isMethodAllowed(method string) bool {
	if len(hdlr.allowedMethods) == 0 {
		// If no method allowed, always return false, even for preflight request
		return false
	}
	method = strings.ToUpper(method)
	if method == "OPTIONS" {
		// Always allow preflight requests
		return true
	}
	for _, m := range hdlr.allowedMethods {
		if m == method {
			return true
		}
	}
	return false
}

// areHeadersAllowed checks if a given list of headers are allowed to used within
// a cross-domain request.
func (hdlr *CORSType) areHeadersAllowed(requestedHeaders []string) bool {
	if hdlr.allowedHeadersAll || len(requestedHeaders) == 0 {
		return true
	}
	for _, header := range requestedHeaders {
		header = http.CanonicalHeaderKey(header)
		found := false
		for _, h := range hdlr.allowedHeaders {
			if h == header {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

const toLower = 'a' - 'A'

type converter func(string) string

// convert converts a list of string using the passed converter function
func convert(s []string, c converter) []string {
	out := []string{}
	for _, i := range s {
		out = append(out, c(i))
	}
	return out
}

// parseHeaderList tokenize + normalize a string containing a list of headers
func parseHeaderList(headerList string) []string {
	l := len(headerList)
	h := make([]byte, 0, l)
	upper := true
	// Estimate the number headers in order to allocate the right splice size
	t := 0
	for i := 0; i < l; i++ {
		if headerList[i] == ',' {
			t++
		}
	}
	headers := make([]string, 0, t)
	for i := 0; i < l; i++ {
		b := headerList[i]
		if b >= 'a' && b <= 'z' {
			if upper {
				h = append(h, b-toLower)
			} else {
				h = append(h, b)
			}
		} else if b >= 'A' && b <= 'Z' {
			if !upper {
				h = append(h, b+toLower)
			} else {
				h = append(h, b)
			}
		} else if b == '-' || b == '_' || (b >= '0' && b <= '9') {
			h = append(h, b)
		}

		if b == ' ' || b == ',' || i == l-1 {
			if len(h) > 0 {
				// Flush the found header
				headers = append(headers, string(h))
				h = h[:0]
				upper = true
			}
		} else {
			upper = b == '-' || b == '_'
		}
	}
	return headers
}

// if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
// 	hdlr.Next.ServeHTTP(www, req)
// 	return
// }

//	hdlr.Next.ServeHTTP(rw, req)

//	if (rw.StatusCode == 200 || rw.StatusCode == 0) && rw.Length >= int64(hdlr.MinLength) {

//		req.Header.Del("Accept-Encoding")
//		rw.Header().Set("Content-Encoding", "gzip") // Set header to inticate we are processing it

//		var b bytes.Buffer // Setup to process
//		gz := gzip.NewWriter(&b)
//		defer gz.Close()

//		oldbody := rw.GetBody()
//		rw.SaveCurentBody(string(oldbody)) // save original body!

//		// move the file name from ResolvedFn  to DependentFNs -- Replace file in ResolvedFn wioth --gzip--
//		if !lib.InArray(rw.ResolvedFn, rw.DependentFNs) {
//			rw.DependentFNs = append(rw.DependentFNs, rw.ResolvedFn)
//		}
//		rw.ResolvedFn = "--gzip--"

//		var newdata []byte
//		var NewETag string

//		if _, err := gz.Write(oldbody); err != nil { // Get body and apply transform
//			goto booboo
//		}
//		if err := gz.Flush(); err != nil {
//			goto booboo
//		}

//		// b has data in it now -- this is the point to tell the cache to save the gzip version!
//		newdata = b.Bytes()

//		// Update ETag -- Need file ModTime and size - then re-calculate hash
//		NewETag, err = hashlib.HashData(newdata)
//		if err != nil {
//			goto booboo
//		}

//		www.Header().Set("ETag", NewETag)
//		if www.Header().Get("Cache-Control") == "" { // if have a cache that indicates no-caching - then what
//			www.Header().Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0") // HTTP 1.1.
//		}

//		rw.ReplaceBody(newdata)
//		rw.SaveDataInCache = true // Mark the data for saving if this file gets cached.

//	booboo:
//	}

/*

Request URL:http://reduxblog.herokuapp.com/app/posts/?key=15519
Request Method:GET
Status Code:404 Not Found
Remote Address:23.23.190.31:80
Referrer Policy:no-referrer-when-downgrade
Response Headers
view source
Access-Control-Allow-Credentials:true
Access-Control-Allow-Methods:GET, POST, DELETE, PUT, PATCH, OPTIONS, HEAD
Access-Control-Allow-Origin:http://localhost:8080
Access-Control-Max-Age:0
Connection:keep-alive
Content-Length:1564
Content-Type:text/html; charset=utf-8
Date:Sun, 09 Apr 2017 22:03:32 GMT
Server:Cowboy
Vary:Origin
Via:1.1 vegur
X-Rack-Cors:hit
X-Request-Id:2bb771d3-2ae1-482e-befe-72517b907030
X-Runtime:0.003536
Request Headers
view source
Accept-Encoding:gzip, deflate, sdch
Accept-Language:en-US,en;q=0.8
Connection:keep-alive
Host:reduxblog.herokuapp.com
Origin:http://localhost:8080
Referer:http://localhost:8080/
User-Agent:Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.133 Safari/537.36
Query String Parameters
view source
view URL encoded
key:15519

*/

/* vim: set noai ts=4 sw=4: */
