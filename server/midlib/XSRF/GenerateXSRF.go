//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1254
//

//
// Generate XSRF token that can be checked and validate to prevent XSRF.
// Normally this token is only generated for the initial load of the first HTML file.
// It is safe to use the user token for tracking session data in Redis.
//

// If no cookies then must not allow 304 response - must send out code

// ---------------------------------------------------------------------------------------------------------------------------------------------

// xyzzy - test code -- You know what - this is going to be annoying to build a test for.  That is the way it should be.
// That's because you need to get some data, extract it - then mangle it - then add UA and take the SHA256 of it.  Then
// create a pair of cookies in a cookie jar and then mane a 2nd request.  Ouch!!!
//
// Now the second problem in testing.  If I build a really nice test that runs say (PhantomJS) to do the test then
// I will make it easy to copy the test code to delete this kind of XSRF prevention.  Ouch!!!

// XyzzyOverlap - if Paths overlaps with valid paths - then need to generate token at this point!!!!!!!!!!!!!!!!!!!!!!!!!!!
// Unfortunately - "Paths" and "ValidPaths" must not overlap at this point in time.  TODO: fix this.

package GenerateXSRF

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"www.2c-why.com/JsonX"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/HashStrings"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------
//
//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*GenerateXSRFHandlerType)
//		if ok {
//			pCfg.SetNext(next)
//			rv = pCfg
//		} else {
//			err = mid.FtlConfigError
//			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
//		}
//		gCfg.ConnectToRedis()
//		pCfg.gCfg = gCfg
//		return
//	}
//
//	postInit := func(h interface{}, callNo int) error {
//
//		hh, ok := h.(*GenerateXSRFHandlerType)
//		if !ok {
//			// logrus.Warn(fmt.Sprintf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo))
//			fmt.Printf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo)
//			return mid.ErrInternalError
//		}
//		hh.cleanCSS = regexp.MustCompile("[ \t\n\r\f]")
//		tt := time.Now()
//		tt = tt.Add(time.Duration(hh.MaxAge) * time.Second)
//		hh.expires = tt.UTC()
//		return nil
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} { return &GenerateXSRFHandlerType{} }
//
//	cfg.RegInitItem2("GenerateXSRF", initNext, createEmptyType, postInit, `{
//		}`)
//}
//
//// SetNext normally identical
//func (hdlr *GenerateXSRFHandlerType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &GenerateXSRFHandlerType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("GenerateXSRF", CreateEmpty, `{
		"Paths":               { "type":["string","filepath"], "isarray":true, "required":true },
		"Base":                { "type":["string","filepath"], "isarray":true },
		"ValidPaths":          { "type":["string","filepath"], "isarray":true, "required":true },
	    "UserCookieName":      { "type":["string"], "default":"xsrf_user" },
	    "ValueCookieName":     { "type":["string"], "default":"X_XSRF_TOKEN" },
	    "RedisPrefix":         { "type":["string"], "default":"XSRF:" },
	    "RedisSessionPrefix":  { "type":["string"], "default":"" },
        "MaxAge":              { "type":[ "int" ] },
        "RedirectToLogin":     { "type":[ "bool" ] },
        "LoginURL":            { "type":["string","filepath"], "isarray":true },
		"LineNo":              { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *GenerateXSRFHandlerType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	gCfg.ConnectToRedis()
	hdlr.gCfg = gCfg
	return
}

func (hdlr *GenerateXSRFHandlerType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	hdlr.cleanCSS = regexp.MustCompile("[ \t\n\r\f]")
	tt := time.Now()
	tt = tt.Add(time.Duration(hdlr.MaxAge) * time.Second)
	hdlr.expires = tt.UTC()
	return
}

var _ mid.GoFTLMiddleWare = (*GenerateXSRFHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

// GenerateXSRFHandlerType keeps all the configuration for this middleware
type GenerateXSRFHandlerType struct {
	Next               http.Handler                //
	Paths              []string                    //
	Base               []string                    // If avaiable the ${1} matching nth base will be used - else - the Host
	ValidPaths         []string                    //
	UserCookieName     string                      // if Name starts with "-" then delete existing header before creating new one.
	ValueCookieName    string                      // if Name starts with "-" then delete existing header before creating new one.
	RedisPrefix        string                      // Key prefix to save users under in redis, default XSRF:
	RedisSessionPrefix string                      // Alternative prefix if using sessions, "Session:" for example
	MaxAge             int                         //
	RedirectToLogin    bool                        //
	LoginURL           []string                    // Positional match to ValidPaths - if it matches that Nth valid path and RedirectToLogin is true and you have this then 307 redirect to that location.
	LineNo             int                         //
	expires            time.Time                   //
	cleanCSS           *regexp.Regexp              //
	gCfg               *cfg.ServerGlobalConfigType //
	timeToLive         int                         // time to live in seconds -- same as MaxAge
}

// NewGenerateXSRFServer initialize a middleware structure for testing
func newGenerateXSRFServer(n http.Handler, p []string) *GenerateXSRFHandlerType {
	return &GenerateXSRFHandlerType{Next: n, Paths: p}
}

// ServeHTTP called in stack implements the token add and check.
func (hdlr *GenerateXSRFHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchReN(hdlr.ValidPaths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "XSRF - Validation", hdlr.ValidPaths, pn, req.URL.Path)

			// XyzzyOverlap - if Paths overlaps with valid paths - then need to generate token at this point!!!!!!!!!!!!!!!!!!!!!!!!!!!

			// check to see if have cookie ( user, match )
			// get value of cookie ( user, match )
			xsrfUser := lib.GetCookie(hdlr.UserCookieName, req)
			xsrfValue := lib.GetCookie(hdlr.ValueCookieName, req)
			if xsrfUser == "" || xsrfValue == "" {
				if hdlr.RedirectToLogin && pn < len(hdlr.LoginURL) && req.Method == "GET" {
					www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
					www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
					www.Header().Set("Expires", "0")                                         // Proxies.
					www.Header().Set("Content-Type", "text/html")                            //
					www.Header().Set("Location", hdlr.LoginURL[pn])
					www.WriteHeader(http.StatusTemporaryRedirect)
				} else {
					http.Error(www, "Not Authorized", http.StatusUnauthorized)
				}
				return
			}

			ok := hdlr.validateTokens(www, req, rw, xsrfUser, xsrfValue, pn, "ValidPaths")

			if ok {

				hdlr.Next.ServeHTTP(rw, req)

				// extend time on cookies also
				http.SetCookie(www, hdlr.genCookie(hdlr.UserCookieName, xsrfUser))
				http.SetCookie(www, hdlr.genCookie(hdlr.ValueCookieName, xsrfValue))
			}

			return
		}
	} else if pn := lib.PathsMatchReN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "GenerateXSRF - Generate", hdlr.Paths, pn, req.URL.Path)

			hdlr.Next.ServeHTTP(rw, req)

			// see if we already have the 2 cookies - if so don't generate a new one - just update TTL on them in Redis
			xsrfUser := lib.GetCookie(hdlr.UserCookieName, req)
			xsrfValue := lib.GetCookie(hdlr.ValueCookieName, req)
			if xsrfUser != "" || xsrfValue != "" {
				trx.AddNote(1, "GenerateXSRF - just a revalidate")
				ok := hdlr.validateTokens(www, req, rw, xsrfUser, xsrfValue, pn, "Paths")
				if ok {
					http.SetCookie(www, hdlr.genCookie(hdlr.UserCookieName, xsrfUser))
					http.SetCookie(www, hdlr.genCookie(hdlr.ValueCookieName, xsrfValue))
				}
				return
			}

			h := www.Header()
			ct := h.Get("Content-Type")

			if rw.StatusCode == http.StatusOK && strings.HasPrefix(ct, "text/html") && req.Method == "GET" {

				trx.AddNote(1, "GenerateXSRF - new session")

				r0 := GenRandNumber3x()
				r1 := GenRandNumber3x()
				r2 := GenRandNumber3x()

				css := fmt.Sprintf(` .bodyStyleColorTextStyle{ color: #%s; background-color: #%s; } `, r1, r2)

				csst := hdlr.cleanCSS.ReplaceAllString(css, "")

				ua := ""
				uaS := ""
				if useUa {
					ua = req.Header.Get("User-Agent")
					uaS = "+navigator.userAgent"
				}

				bs := ""
				bsS := ""
				if useBaseURL {
					if pn < len(hdlr.Base) {
						bs = hdlr.Base[pn]
					} else {
						bs = req.Host // xyzzy need to process to just get host?
						fmt.Fprintf(os.Stderr, "%sUsed HOST[%s] instead of href.Base!%s\n", MiscLib.ColorRed, req.Host, MiscLib.ColorReset)
					}
					bsS = "+gb()"
				}

				h256 := HashStrings.HashStrings(csst + ua + bs)

				xsrfUser := getUUIDAsString()

				rw.Postfix = fmt.Sprintf(`
<script src="js/sha256.min.js"></script>
<style id="dx441_%s"> %s </style>
<script type="text/javascript">
(function(){
	var gb=function(){
		if ('baseURI' in document){
			return(document.baseURI);
		} 
		var bt=document.getElementsByTagName("base");
		if(bt.length>0){
			return(bt[0].href);
		} 
		return(window.location.href);
	}
	var cc=function(nn,vv,ss){
		var dd=new Date();
		dd.setTime(dd.getTime()+(ss*1000));
		var expires="; expires="+dd.toGMTString();
		document.cookie=nn+"="+vv+expires+"; path=/";
	}
	var sb = document.getElementById("dx441_%s").innerHTML;
	sb = sb.replace(/[ \t\n\r\f]/g,"");
	var gg=CryptoJS.SHA256(sb+"\n"%s%s).toString().toLowerCase();
	cc(%q,gg,%d);
})();
</script>
`, r0, css, r0, uaS, bsS, hdlr.ValueCookieName, hdlr.MaxAge)

				// set cookie with unique user id number
				http.SetCookie(www, hdlr.genCookie(hdlr.UserCookieName, xsrfUser))

				// Save to Redis
				// SET ( "XSRF:tok", "value" );
				// TTL ( "XSRF:tok", "value" );

				conn, err := hdlr.gCfg.RedisPool.Get()
				defer hdlr.gCfg.RedisPool.Put(conn)
				if err != nil {
					logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
					rw.Error = err
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				key := hdlr.RedisPrefix + xsrfUser

				err = conn.Cmd("SET", key, h256).Err
				if err != nil {
					logrus.Errorf(`{"msg":"Error %s Unable to set redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF())
					rw.Error = err
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				err = conn.Cmd("TTL", key, hdlr.timeToLive).Err
				if err != nil {
					logrus.Errorf(`{"msg":"Error %s Unable to set redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF())
					rw.Error = err
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				if hdlr.RedisSessionPrefix != "" {
					key := hdlr.RedisSessionPrefix + xsrfUser
					err = conn.Cmd("TTL", key, hdlr.timeToLive).Err
					if err != nil {
						logrus.Errorf(`{"msg":"Error %s Unable to set redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF())
						rw.Error = err
						rw.WriteHeader(http.StatusInternalServerError)
						return
					}
				}

			}
			return
		}
	}
	hdlr.Next.ServeHTTP(www, req)
}

// create and return cookie
func (hdlr *GenerateXSRFHandlerType) genCookie(Name, Value string) (theCookie *http.Cookie) {
	theCookie = &http.Cookie{}
	theCookie.Name = Name
	theCookie.Value = Value
	theCookie.Path = "/"
	// hdlr.theCookie.Domain = hdlr.Domain
	theCookie.Expires = hdlr.expires.UTC()
	theCookie.MaxAge = hdlr.MaxAge
	theCookie.Secure = false // xyzzy - should be set to true if HTTPS
	theCookie.HttpOnly = true
	return
}

// lookup in Redis - check that it is good
func (hdlr *GenerateXSRFHandlerType) validateTokens(www http.ResponseWriter, req *http.Request, rw *goftlmux.MidBuffer, xsrfUser, xsrfValue string, pn int, which string) bool {

	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		rw.Error = err
		rw.WriteHeader(http.StatusInternalServerError)
		return false
	}

	key := hdlr.RedisPrefix + xsrfUser

	v1, err := conn.Cmd("GET", key).Str()
	if err != nil {
		logrus.Errorf(`{"msg":"Error %s Unable to get redis value.","LineFile":%q}`+"\n", err, godebug.LF())
		rw.Error = err
		rw.WriteHeader(http.StatusInternalServerError)
		return false
	}

	// if v1 != xsrfValue {
	if subtle.ConstantTimeCompare([]byte(v1), []byte(xsrfValue)) != 1 {
		if hdlr.RedirectToLogin && pn < len(hdlr.LoginURL) && req.Method == "GET" {
			www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
			www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
			www.Header().Set("Expires", "0")                                         // Proxies.
			www.Header().Set("Content-Type", "text/html")                            //
			www.Header().Set("Location", hdlr.LoginURL[pn])
			www.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			http.Error(www, "Not Authorized", http.StatusUnauthorized)
		}
		return false
	}

	// extend time data in Redis
	err = conn.Cmd("TTL", key, hdlr.timeToLive).Err
	if err != nil {
		logrus.Errorf(`{"msg":"Error %s Unable to set redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF())
		rw.Error = err
		rw.WriteHeader(http.StatusInternalServerError)
		return false
	}

	if hdlr.RedisSessionPrefix != "" {
		key := hdlr.RedisSessionPrefix + xsrfUser
		err = conn.Cmd("TTL", key, hdlr.timeToLive).Err
		if err != nil {
			logrus.Errorf(`{"msg":"Error %s Unable to set redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF())
			rw.Error = err
			rw.WriteHeader(http.StatusInternalServerError)
			return false
		}
	}

	return true
}

const db1 = false
const db4 = false
const useUa = true
const useBaseURL = true

/* vim: set noai ts=4 sw=4: */
