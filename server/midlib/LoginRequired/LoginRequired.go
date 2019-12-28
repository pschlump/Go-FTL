//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2017
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1266
//

// TODO:
//		hdlr.MergeSessionData - takes sesison data if not logged in and merges it.  If logged in then takes logged in data and adds that.
//		add
//			ps := &rw.Ps
//			rw
//		etc.
//
// TODO:
// 		for _, pp := range []string{"//", ")]}'", "while(1);", "while(true);", "for(;;);"} {		// xyzzy - hdlr.PrefixList

//
// Package LoginRequired implement checks that a login has occurred at a previous level in the set of middleware.
//
// Why this works -
//
// At the top level the server (top) will remove the parameters $is_logged_in$ and $is_full_login$.  If the parameters
// are found then they will get converted into "user_param::$is_logged_in$" and "user_param::$is_full_login$".
// Then if login occurs it can set the params and this can see them.
//

package LoginRequired

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	JsonX "github.com/pschlump/JSONx"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/jwtverif"
	logrus "github.com/pschlump/pslog" // "github.com/sirupsen/logrus"
	"github.com/pschlump/uuid"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &LoginRequiredType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // Xyzzy - report errors in 'meta'
		return x
	}
	//
	// "PostgresQuery":	{ "type":["string"], "default":"select 'ok' as \\"x\\" from \\"t_user\\" where \\"auth_token\\" = $1" },
	//
	// Method is one of:            AesSrp, Cookie, Data, Data:POST, Authentication:Bearer, Session
	// ValidationSource is one of:	pg, redis, jwt, AesSrp
	//
	// Cookie, Data, Data:POST - all use "ParamName" as the name of the parameter to get auth_token from.
	//
	mid.RegInitItem3("LoginRequired", CreateEmpty, `{
		"Paths":            	{ "type":["string","filepath"], "isarray":true, "required":true },
		"StrongLoginReq":   	{ "type":["string"], "default":"no" },
		"AuthMethod":	    	{ "type":["string"], "isarray":true, "default":"AesSrp" },
		"Final":		    	{ "type":["string"], "default":"yes" },
		"ValidationSource":		{ "type":["string"], "default":"AesSrp" },
		"RedisAuthTokenPrefix":	{ "type":["string"], "default":"isli:" },
		"RedisSessionPrefix":	{ "type":["string"], "default":"session:" },
		"PostgresQuery":		{ "type":["string"], "default":"select s_validate_logged_in( $1 )" },
		"ParamName":			{ "type":["string"], "default":"auth_token" },
		"KeyFilePublic":		 { "type":["string"], "default":"key.pub" },
		"KeyFileType":		     { "type":["string"], "default":"ES256" },
		"CheckXsrfToken":   	{ "type":["string"], "default":"yes" },
		"RemoteValidate":		{ "type":["string"], "default":"no" },
		"RemoteValidateURL":	{ "type":["string"], "default":"http://auth.2c-why.com/api/validate_auth_token" },
		"RemoteValidatePrefix":	{ "type":["string"], "default":"rvli:" },
		"DefaultTTL":			{ "type":["string"], "default":"3600" },
		"XssiPrefix":			{ "type":["string"], "isarray":true },
		"LineNo":           	{ "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *LoginRequiredType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	if len(hdlr.XssiPrefix) == 0 {
		hdlr.XssiPrefix = []string{"while(1);", "//", ")]}'\n", ")]}'", "while(true);", "for(;;);"}
		// for _, pp := range []string{"//", ")]}'", "while(1);", "while(true);", "for(;;);"} { // xyzzy - hdlr.PrefixList
	}
	gCfg.ConnectToRedis()
	gCfg.ConnectToPostgreSQL()
	hdlr.gCfg = gCfg
	return
}

func (hdlr *LoginRequiredType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	hdlr.strongLoginReq, _ = lib.ParseBool(hdlr.StrongLoginReq)
	hdlr.defaultTtl, _ = strconv.Atoi(hdlr.DefaultTTL)
	if hdlr.defaultTtl <= 0 {
		hdlr.defaultTtl = 3600
	}
	return
}

var _ mid.GoFTLMiddleWare = (*LoginRequiredType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type LoginRequiredType struct {
	Next                 http.Handler
	Paths                []string
	StrongLoginReq       string
	AuthMethod           []string
	Final                string
	ValidationSource     string
	RedisAuthTokenPrefix string
	RedisSessionPrefix   string // Deprecated!
	PostgresQuery        string
	ParamName            string
	KeyFilePublic        string // public key for verification. (private is used for signing)
	KeyFileType          string // ES256, RS256 etc.
	CheckXsrfToken       string // validate that the X-Xsrf-Token is set correctly for this session
	RemoteValidate       string // yes means that you should use RemoteValidateURL to validate the "auth_token"
	RemoteValidateURL    string
	RemoteValidatePrefix string // prefix for keeping auth_token when jwi
	DefaultTTL           string
	XssiPrefix           []string //
	LineNo               int
	strongLoginReq       bool
	defaultTtl           int
	gCfg                 *cfg.ServerGlobalConfigType //
}

func NewLoginRequiredServer(n http.Handler, p []string) *LoginRequiredType {
	return &LoginRequiredType{Next: n, Paths: p}
}

//		"Method":		    { "type":["string"], "default":"AesSrp" },
// Method is one of:            AesSrp, Cookie, Data, Data:POST, Authentication:Bearer, Session
//		"ValidationSource":	{ "type":["string"], "default":"AesSrp" },
// ValidationSource is one of:	pg, redis, jwt, AesSrp
//		"RedisAuthTokenPrefix":		{ "type":["string"], "default":"isli:" },
//		"PostgresQuery":	{ "type":["string"], "default":"select s_validate_logged_in( $1 )" },
//		"ParamName":		{ "type":["string"], "default":"auth_token" },
//		"LineNo":           { "type":[ "int" ], "default":"1" }
//
// Cookie, Data, Data:POST - all use "ParamName" as the name of the parameter to get auth_token from.

//			if hdlr.gCfg.DbOn("*", "SessionRedis", "db1") {

func (hdlr *LoginRequiredType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "LoginRequired", hdlr.Paths, pn, req.URL.Path)

			ps := &rw.Ps

			/*

			   // to set value in session - find out how to get
			   		rw.Session.SetData("regular", "email_addr", "kermit@the-green-pc.com")
			   		rw.Session.SetRule("email_addr", false, true)

			   // to get value in session
			   		dv, err := rw.Session.GetData("regular", "email_addr")

			   // dump data to string
			   		x := rw.Session.DumpData()

			*/

			bc := lib.GetCookie("X-Go-FTL-Trx-Id", req)
			sc := lib.GetCookie("X-Go-FTL-Sesion-Id", req)
			xt := hdlr.GetXsrfToken(req)

			var chkXsrf = func(xt string) bool {
				if hdlr.CheckXsrfToken == "yes" {
					// session := hdlr.GetSessionFromRedis(hdlr.RedisSessionPrefix, bc)
					session := rw.Session
					if session == nil {
						fmt.Fprintf(os.Stderr, "\n%s!!! Config Error Valid !!! X-Xsrf-Token rw.Session is nill {SessionRedis} is not configured for this path\n\tAT:%s\n%s\n",
							MiscLib.ColorRed, godebug.LF(), MiscLib.ColorReset)
						fmt.Fprintf(os.Stdout, "\n%s!!! Config Error Valid !!! X-Xsrf-Token rw.Session is nill {SessionRedis} is not configured for this path\n\tAT:%s\n%s\n",
							MiscLib.ColorRed, godebug.LF(), MiscLib.ColorReset)
					} else {
						//			if hdlr.gCfg.DbOn("*", "SessionRedis", "db-session") {
						if hdlr.gCfg.DbOn("*", "LoginRequired", "db-session") {
							fmt.Printf("session before = %s\n", session.DumpData())
						}
						good, err := session.GetData("user", "$xsrf_token$")
						// isValid := hdlr.CheckXsrfTokenVsSession(xt, session)
						isValid := (err == nil && xt == good)
						if isValid {
							fmt.Fprintf(os.Stderr, "\n%sChecking X-Xsrf-Token = -->>%s<<--, X-Go-FTL-Trx-Id=%s, session.err=%s good=%s isValid=%v\nAT:%s %s\n",
								MiscLib.ColorGreen, xt, bc, err, good, isValid, godebug.LF(), MiscLib.ColorReset)
							fmt.Fprintf(os.Stdout, "\n%sChecking X-Xsrf-Token = -->>%s<<--, X-Go-FTL-Trx-Id=%s, session.err=%s good=%s isValid=%v\nAT:%s %s\n",
								MiscLib.ColorGreen, xt, bc, err, good, isValid, godebug.LF(), MiscLib.ColorReset)
						} else {
							fmt.Fprintf(os.Stderr, "\n%sChecking X-Xsrf-Token = -->>%s<<--, X-Go-FTL-Trx-Id=%s, session.err=%s good=%s isValid=%v\nAT:%s %s\n",
								MiscLib.ColorRed, xt, bc, err, good, isValid, godebug.LF(), MiscLib.ColorReset)
							fmt.Fprintf(os.Stdout, "\n%sChecking X-Xsrf-Token = -->>%s<<--, X-Go-FTL-Trx-Id=%s, session.err=%s good=%s isValid=%v\nAT:%s %s\n",
								MiscLib.ColorRed, xt, bc, err, good, isValid, godebug.LF(), MiscLib.ColorReset)
						}
						//if !isValid {
						//	fmt.Fprintf(os.Stderr, "\n%s!!! Not Valid !!! X-Xsrf-Token = -->>%s<<--, X-Go-FTL-Trx-Id=%s, isValid=%v\n\tAT:%s\n%s\n",
						//		MiscLib.ColorRed, xt, bc, isValid, godebug.LF(), MiscLib.ColorReset)
						//	fmt.Printf("\n%s!!! Not Valid !!! X-Xsrf-Token = -->>%s<<--, X-Go-FTL-Trx-Id=%s, isValid=%v\n\tAT:%s\n%s\n",
						//		MiscLib.ColorRed, xt, bc, isValid, godebug.LF(), MiscLib.ColorReset)
						//}
						if enableXsrfCheckOn && !isValid {
							www.WriteHeader(http.StatusForbidden)
							return false
						}
					}
				}
				return true
			}

			for _, aMethod := range hdlr.AuthMethod {
				switch aMethod {
				case "AesSrp":
					fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
					is_logged_in := ps.ByNameDflt("$is_logged_in$", "")
					is_full_login := ps.ByNameDflt("$is_full_login$", "")
					if is_logged_in == "y" {
						if hdlr.StrongLoginReq == "yes" {
							if is_full_login == "y" {
								hdlr.MergeSessionData(rw, true)
								if !chkXsrf(xt) {
									return
								}
								hdlr.Next.ServeHTTP(www, req)
								return
							}
						} else {
							hdlr.MergeSessionData(rw, true)
							if !chkXsrf(xt) {
								return
							}
							hdlr.Next.ServeHTTP(www, req)
							return
						}
					}
				case "Cookie":
					fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
					auth_token := hdlr.GetCookie(hdlr.ParamName, req)
					if valid, first, _, _ := hdlr.ValidateAuthToken(rw, www, req, auth_token); valid {
						_ = first
						hdlr.MergeSessionData(rw, true)
						if !chkXsrf(xt) {
							return
						}
						hdlr.Next.ServeHTTP(www, req)
						return
					}
				case "Data":
					fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
					auth_token := hdlr.GetData(hdlr.ParamName, ps)
					if valid, first, _, _ := hdlr.ValidateAuthToken(rw, www, req, auth_token); valid {
						_ = first
						hdlr.MergeSessionData(rw, true)
						if !chkXsrf(xt) {
							return
						}
						hdlr.Next.ServeHTTP(www, req)
						return
					}
				case "Data:POST":
					fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
					auth_token := hdlr.GetDataPost(hdlr.ParamName, ps)
					if valid, first, _, _ := hdlr.ValidateAuthToken(rw, www, req, auth_token); valid {
						_ = first
						hdlr.MergeSessionData(rw, true)
						if !chkXsrf(xt) {
							return
						}
						hdlr.Next.ServeHTTP(www, req)
						return
					}

				case "Authentication:Bearer":
					if hdlr.gCfg.DbOn("*", "LoginRequired", "db-bearer") {
						fmt.Fprintf(os.Stderr, "AT: - About to check/validate bearer token %s\n", godebug.LF())
					}
					jwt_token := hdlr.GetBearer(req, ps)
					if hdlr.gCfg.DbOn("*", "LoginRequired", "db-bearer") {
						fmt.Fprintf(os.Stderr, "%sAuthorization: bearer -->>%s<<--, %s%s\n", MiscLib.ColorYellow, jwt_token, godebug.LF(), MiscLib.ColorReset)
					}
					if valid, first, xsrf_token, ttl := hdlr.ValidateAuthToken(rw, www, req, jwt_token); valid {

						// if hdlr.gCfg.DbOn("*", "LoginRequired", "db-bearer") {
						if true {
							fmt.Fprintf(os.Stderr, "%sAT: -- Authorized! Yea! Inject of jwt_token(%s)=[%s] ttl=%d first=%v %s%s\n", MiscLib.ColorGreen, hdlr.ParamName, jwt_token, first, ttl, godebug.LF(), MiscLib.ColorReset)
						}

						// ----------------------------------------------------------------------------------------------------------------------------------------------------
						// xyzzy - if this ia a "first" check, where we wen't remote to get the validation of the token then
						// we should skip the chkXsrf as it is a new xsrf token.
						// ----------------------------------------------------------------------------------------------------------------------------------------------------
						// xyzzy - should check "prev" token if first!
						// ----------------------------------------------------------------------------------------------------------------------------------------------------
						/*
						   /Users/corwin/go/src/github.com/pschlump/Go-FTL/server/midlib/LoginRequired/LoginRequired.go: if false { // xyzzy-2016-06-13
						   	xref_token when swapped out fails to verify - maybee keep a "set" in redis?
						   	Flush set when regular success.

						   	SMEMBERS - see if in set.
						   	SADD
						   	SREM
						*/

						if false { // <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< PJS Feb 5 ( Sat Apr 27 19:59:30 MDT 2019 )
							hdlr.MergeSessionData(rw, true)
							if !first {
								if !chkXsrf(xt) {
									return
								}
							} else {
								xref_in_set, err := hdlr.RedisSetContains("ses__xsrf:"+sc, xt) // trx__xsrf: shoud be set in config
								fmt.Printf("%sFirst==True, key=trx__xsrf:%s xt=[%s] xref_in_set=[%v] err=%s, AT:%s%s\n", MiscLib.ColorCyan, bc, xt, xref_in_set, err, godebug.LF(), MiscLib.ColorReset)
								if err == nil { // if no error then
									if xref_in_set {
										fmt.Printf("%s\tShould return - with error, skipping for now%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
										fmt.Fprintf(os.Stderr, "%s\tShould return - with error, skipping for now%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
										if true { // xyzzy-2016-06-13
											www.WriteHeader(http.StatusForbidden)
											return
										}
									} else {
										fmt.Printf("%s\txsrf_token: \"It's an older code sir, but it checks out. I was about to clear them.\" - %s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
										fmt.Fprintf(os.Stderr, "%s\txsrf_token: \"It's an older code sir, but it checks out. I was about to clear them.\" - %s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
									}
								}
							}
						} // <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< PJS Feb 5

						hdlr.Next.ServeHTTP(www, req)
						if first {
							hdlr.InjextXsrf(rw, www, xsrf_token)                              // disasemble response // inject "$xsrf_token$"
							hdlr.RedisUpdateSet("ses__xsrf:"+sc, 60*60*24*30, xsrf_token, xt) // save the token locally for revalidation when new one fetched
						}
						return
					}

				case "Session":
					fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
					auth_token := hdlr.GetSessionAuth()
					if valid, first, _, _ := hdlr.ValidateAuthToken(rw, www, req, auth_token); valid {
						_ = first
						hdlr.MergeSessionData(rw, true)
						if !chkXsrf(xt) {
							return
						}
						hdlr.Next.ServeHTTP(www, req)
						return
					}

				default:
					fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
					logrus.Errorf("Invalid validation method at: %s", godebug.LF())
					methodOneOf := "AesSrp, Cookie, Data, Data:POST, Authentication:Bearer, Session"
					fmt.Printf("Error - Invalid Method [%s] should be one of %s - LoginRequired - At: %s\n", hdlr.AuthMethod, methodOneOf, godebug.LF())
					fmt.Fprintf(os.Stderr, "%sError - Invalid Method [%s] should be one of %s - LoginRequired - At: %s%s\n",
						MiscLib.ColorRed, hdlr.AuthMethod, methodOneOf, godebug.LF(), MiscLib.ColorReset)
					return
				}
			}
			fmt.Fprintf(os.Stderr, "\n%sAT: --- Falied to Validate / 403 to be returned ---, %s%s\n\n", MiscLib.ColorRed, godebug.LF(), MiscLib.ColorReset)
			www.WriteHeader(http.StatusForbidden)
			return

			// ip := lib.GetIP(req)
			// cookieValue := lib.GetCookie("LoginAuthCookie", req)
			// cookieHash := lib.GetCookie("LoginHashCookie", req)

			// Xyzzy - pantopick at this pont - if system thas changed then fail.

			// if xip, _, _, hash, err := hdlr.GetCookieAuth(cookieValue, rw); err == nil && xip == ip && cookieHash == hash {
			// 	// fmt.Printf("   Serve it\n")
			// 	hdlr.Next.ServeHTTP(www, req)
			// 	return
			// } else {
			// 	// fmt.Printf("   *** Reject *** it\n")
			// 	www.WriteHeader(http.StatusForbidden)
			// }

		}
	}
	if hdlr.Final == "yes" {
		www.WriteHeader(http.StatusNotFound)
	} else {
		hdlr.MergeSessionData(nil, false)
		hdlr.Next.ServeHTTP(www, req)
	}

}

func (hdlr *LoginRequiredType) MergeSessionData(rw *goftlmux.MidBuffer, logged_in bool) {
	if rw != nil {
		if logged_in {
			rw.Session.Login()
		} else {
			rw.Session.Logout()
		}
	}
}

func (hdlr *LoginRequiredType) GetCookie(CookieName string, req *http.Request) (rv string) {
	rv = lib.GetCookie(CookieName, req)
	return
}

func (hdlr *LoginRequiredType) GetData(ParamName string, ps *goftlmux.Params) (rv string) {
	rv = ps.ByNameDflt(ParamName, "")
	return
}

// Deprecated!
//func (hdlr *LoginRequiredType) GetSessionFromRedis(SessionPrefix string, Id string) (rv string) {
//	key := SessionPrefix + Id
//
//	conn, err := hdlr.gCfg.RedisPool.Get()
//	defer hdlr.gCfg.RedisPool.Put(conn)
//	if err != nil {
//		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
//		return
//	}
//
//	val, err := conn.Cmd("GET", key).Str()
//	if err != nil {
//		return
//	}
//
//	fmt.Fprintf(os.Stderr, "%sRedis: -->>%s<<-- %s%s\n", MiscLib.ColorCyan, val, godebug.LF(), MiscLib.ColorReset)
//
//	rv = val
//	return
//}

// isValid := hdlr.CheckXsrfTokenVsSession ( xt, session );
/*
"{"$is_logged_in$":true,
"UserData":{"$xsrf_token$":"45182535-fc56-4c95-90a7-0d3283228617","test_aaa":"aaa"},"RegularData":{"test_bbb":"bbb"},"Rules":{"$is_logged_in$":{"temporary":false,"user_may_set":false},"$xsrf_token$":{"temporary":false,"user_may_set":true},"test_aaa":{"temporary":false,"user_may_set":true},"test_bbb":{"temporary":false,"user_may_set":true}}}"
*/
// Deprecated!
//func (hdlr *LoginRequiredType) CheckXsrfTokenVsSession(xt, session string) (rv bool) {
//	if db8 {
//		fmt.Printf("CheckXsrfVsSession:  session >%s<\n", session)
//	}
//	type GetXsData struct {
//		Xs string `json:"$xsrf_token$"`
//	}
//	type GetXs struct {
//		UserData GetXsData `json:"UserData"`
//	}
//	var xx GetXs
//	err := json.Unmarshal([]byte(session), &xx)
//	if err != nil {
//		fmt.Printf("Error: unable to unmarshal session -->>%s<<--, err=%s, %s\n", session, err, godebug.LF())
//		return false
//	}
//	if xx.UserData.Xs == xt {
//		fmt.Fprintf(os.Stderr, "%s - Xsrf Matched %s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
//		return true
//	}
//	fmt.Fprintf(os.Stderr, "%s - Xsrf !! did not match !! %s\n", MiscLib.ColorRed, MiscLib.ColorReset)
//	return false
//}

func (hdlr *LoginRequiredType) GetDataPost(ParamName string, ps *goftlmux.Params) (rv string) {
	// func (ps *Params) GetByNameAndType(name string, ft FromType) (rv string, found bool) {
	val, found := ps.GetByNameAndType(ParamName, goftlmux.FromBody)
	if found {
		rv = val
	}
	val, found = ps.GetByNameAndType(ParamName, goftlmux.FromBodyJson)
	if found {
		rv = val
	}
	return
}

// Look for an Authentication header with a 'bearer' and pull that out
func (hdlr *LoginRequiredType) GetBearer(req *http.Request, ps *goftlmux.Params) (rv string) {
	aa := req.Header.Get("Authorization")
	aX := strings.Split(aa, " ")
	if len(aX) >= 2 && aX[0] == "bearer" {
		rv = aX[1]
	}
	if len(aX) >= 2 && aX[0] == "Bearer" {
		rv = aX[1]
	}
	// xyzzy - Try cookie - xyzzy
	if rv == "" {
		rv = lib.GetCookie("X-Jwt-Token", req)
	}
	// xyzzy - Try URL Param
	if rv == "" {
		rv = hdlr.GetData("XJwtToken", ps)
		fmt.Fprintf(os.Stderr, "rv - via GetData/XJwtToken = ->%s<- AT: %s\n", rv, godebug.LF())
	}
	return
}

func (hdlr *LoginRequiredType) GetXsrfToken(req *http.Request) (rv string) {
	rv = req.Header.Get("X-Xsrf-Token")
	return
}

func (hdlr *LoginRequiredType) GetSessionAuth() (rv string) {
	// xyzzy - TODO - do this when we work on Sessions
	return
}

func (hdlr *LoginRequiredType) ValidateAuthToken(rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, auth_token string) (valid, first bool, xsrf_token string, ttl int) {
	// Note: https://github.com/dgrijalva/jwt-go.git
	//		"RedisAuthTokenPrefix":		{ "type":["string"], "default":"isli:" },
	//		"PostgresQuery":	{ "type":["string"], "default":"select s_validate_logged_in( $1 )" },

	valid, first = false, false

	switch hdlr.ValidationSource {
	case "pg":
		fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())

		//		"PostgresQuery":	{ "type":["string"], "default":"select s_validate_logged_in( $1 )" },
		// PostgresQuery    string

		rows, err := hdlr.gCfg.Pg_client.Db.Query(hdlr.PostgresQuery, auth_token)
		if err != nil {
			fmt.Printf("Database error %s, attempting to validate login qry=%s\n", err, hdlr.PostgresQuery)
			return
		}

		for nr := 0; rows.Next(); nr++ {
			if nr >= 1 {
				fmt.Printf("Error too many rows for a user, should be unique primary key\n")
			} else {

				var val string

				err := rows.Scan(&val)
				if err != nil {
					fmt.Printf("Error on d.b. query %s\n", err)
				}

				type MmType struct {
					Status string
				}
				var mm MmType
				err = json.Unmarshal([]byte(val), &mm)
				if err != nil {
					fmt.Printf("Error on d.b. query %s / unable to unmarsal results [%s]\n", err, val)
				}

				if mm.Status == "success" {
					valid = true
				}
			}
		}

		fmt.Fprintf(os.Stderr, "%sPG: -->>%s<<-- valid=%v, %s%s\n", MiscLib.ColorCyan, auth_token, valid, godebug.LF(), MiscLib.ColorReset)
		return

	case "redis":
		fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
		val, err := hdlr.GetRedisKey(hdlr.RedisAuthTokenPrefix + auth_token)
		if err != nil || val == "no" {
			return
		}
		fmt.Fprintf(os.Stderr, "%sRedis: -->>%s<<-- valid=%v, %s%s\n", MiscLib.ColorCyan, auth_token, true, godebug.LF(), MiscLib.ColorReset)

		valid = true
		return

	case "jwt":

		if hdlr.gCfg.DbOn("*", "LoginRequired", "db-jwt-token") {
			fmt.Fprintf(os.Stderr, "JWT Validaiton of Token -- AT: %s -->>%s<<--\n", godebug.LF(), auth_token)
		}
		iat, err := hdlr.VerifyToken([]byte(auth_token), hdlr.KeyFilePublic, hdlr.KeyFileType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			fmt.Printf("Error: VerifyToken return err=%s ->%s<-\n", err, auth_token)
			return
		}

		ps := &rw.Ps
		goftlmux.AddValueToParams("auth_token", iat, 'i', goftlmux.FromInject, ps)

		// xyzzy - check exprie? -- Use redis for this?
		// xyzzy - check redis - auth_token still valid?  ,"/api/session/validate_auth_token"
		// xyzzy - validate auth_token with remote server?  Store as valid in "redis" - re-check it.
		// ,"/api/session/validate_auth_token": { "g": "s_validate_auth_token", "p": [ "auth_token", "$url$" ]

		if hdlr.gCfg.DbOn("*", "LoginRequired", "db-validate-token") {
			fmt.Fprintf(os.Stderr, "JWT auth_token - AT: %s -->>%s<<--\n", godebug.LF(), iat)
			fmt.Fprintf(os.Stdout, "JWT auth_token - AT: %s -->>%s<<--\n", godebug.LF(), iat)
		}

		fmt.Fprintf(os.Stderr, "%sJWT auth_token - AT: %s -->>%s<<--%s\n", MiscLib.ColorRed, godebug.LF(), iat, MiscLib.ColorReset)

		// get-from-redis, if not then...
		if hdlr.RemoteValidate == "yes" {
			key := hdlr.RemoteValidatePrefix + iat
			val, err := hdlr.GetRedisKey(key)
			if hdlr.gCfg.DbOn("*", "LoginRequired", "db-validate-token") {
				fmt.Fprintf(os.Stderr, "%s[[[ This One - First will be %v (we will remote validate) ]]]%s AT: %s, val=%s, key=%s\n", MiscLib.ColorYellow,
					(err != nil || val == ""), MiscLib.ColorReset, godebug.LF(), val, key)
				fmt.Fprintf(os.Stdout, "%s[[[ This One - First will be %v (we will remote validate) ]]]%s AT: %s, val=%s, key=%s\n", MiscLib.ColorYellow,
					(err != nil || val == ""), MiscLib.ColorReset, godebug.LF(), val, key)
			}
			type GetStatus struct {
				Status    string `json:"status"`
				Ttl       int    `json:"ttl"`
				XsrfToken string `json:"xsrf_token"`
				UserId    string `json:"user_id"`
			}
			var gt GetStatus
			ps := &rw.Ps
			// ------------------------------------------------------------------------------------------------------------------------------------------------------------
			// xyzzy - this is the point to check the TTL - if timeoud is true then need to remote-revalidate token
			// ------------------------------------------------------------------------------------------------------------------------------------------------------------
			if err != nil || val == "" {
				first = true
				id0, _ := uuid.NewV4()
				rn := id0.String()
				if hdlr.gCfg.DbOn("*", "LoginRequired", "db-validate-token") {
					fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
					fmt.Fprintf(os.Stdout, "AT: %s\n", godebug.LF())
				}
				status, rv := GetURL(hdlr.RemoteValidateURL, hdlr.XssiPrefix, "auth_token", iat, "_ran_", rn)
				if status >= 400 {
					fmt.Fprintf(os.Stderr, "\n%s-- Auth Failed -- AT: %s%s\n\n", MiscLib.ColorRed, godebug.LF(), MiscLib.ColorRed)
					fmt.Fprintf(os.Stdout, "\n%s-- Auth Failed -- AT: %s%s\n\n", MiscLib.ColorRed, godebug.LF(), MiscLib.ColorRed)
					return
				}
				if hdlr.gCfg.DbOn("*", "LoginRequired", "db-validate-token") {
					fmt.Fprintf(os.Stderr, "AT: %s, rv=%s\n", godebug.LF(), rv)
					fmt.Fprintf(os.Stdout, "AT: %s, rv=%s\n", godebug.LF(), rv)
				}
				err := json.Unmarshal([]byte(rv), &gt)
				if gt.Ttl <= 0 {
					gt.Ttl = hdlr.defaultTtl
				}
				if err != nil {
					fmt.Fprintf(os.Stderr, "\n%s-- Auth Failed -- AT: %s%s\n\n", MiscLib.ColorRed, godebug.LF(), MiscLib.ColorRed)
					fmt.Fprintf(os.Stdout, "\n%s-- Auth Failed -- AT: %s%s\n\n", MiscLib.ColorRed, godebug.LF(), MiscLib.ColorRed)
					return
				}
				if gt.Status != "success" {
					fmt.Fprintf(os.Stderr, "\n%s-- Auth Failed -- AT: %s%s\n\n", MiscLib.ColorRed, godebug.LF(), MiscLib.ColorRed)
					fmt.Fprintf(os.Stdout, "\n%s-- Auth Failed -- AT: %s%s\n\n", MiscLib.ColorRed, godebug.LF(), MiscLib.ColorRed)
					return
				}
				val = rv
				if hdlr.gCfg.DbOn("*", "LoginRequired", "db-validate-token") {
					fmt.Fprintf(os.Stderr, "$xsrf_token$ = %s AT: %s\n", gt.XsrfToken, godebug.LF())
				}
				hdlr.SetRedisKey(key, gt.Ttl, val) // xyzzy  - return TTL in seconds from remote valiate
				goftlmux.AddValueToParams("$user_id$", gt.UserId, 'i', goftlmux.FromInject, ps)
				if hdlr.gCfg.DbOn("*", "LoginRequired", "db-validate-token") {
					fmt.Printf("Inject (a) of UserId ($user_id$) = [%s]\n", gt.UserId)
				}
				rw.Session.SetData("user", "$xsrf_token$", gt.XsrfToken)
				rw.Session.SetRule("$xsrf_token$", false, true)
				xsrf_token = gt.XsrfToken
				ttl = gt.Ttl
			} else {
				err := json.Unmarshal([]byte(val), &gt)
				if gt.Ttl <= 0 {
					gt.Ttl = hdlr.defaultTtl
				}
				if err != nil {
					fmt.Fprintf(os.Stderr, "\n%s-- Auth Failed -- key=%s val=%s -- AT: %s%s\n\n", MiscLib.ColorRed, key, val, godebug.LF(), MiscLib.ColorRed)
					fmt.Fprintf(os.Stdout, "\n%s-- Auth Failed -- key=%s val=%s -- AT: %s%s\n\n", MiscLib.ColorRed, key, val, godebug.LF(), MiscLib.ColorRed)
					return
				}
				if gt.Status != "success" {
					fmt.Fprintf(os.Stderr, "\n%s-- Auth Failed -- AT: %s%s\n\n", MiscLib.ColorRed, godebug.LF(), MiscLib.ColorRed)
					fmt.Fprintf(os.Stdout, "\n%s-- Auth Failed -- AT: %s%s\n\n", MiscLib.ColorRed, godebug.LF(), MiscLib.ColorRed)
					return
				}
				if hdlr.gCfg.DbOn("*", "LoginRequired", "db-validate-token") {
					fmt.Fprintf(os.Stderr, "%s[[[ This One ]]] First=False Local Redis had the Key [%s] - and it is valid - AT: %s%s\n", MiscLib.ColorYellow, key, godebug.LF(), MiscLib.ColorReset)
				}
				fmt.Fprintf(os.Stdout, "%s[[[ This One ]]] First=False Local Redis had the Key [%s] - and it is valid - AT: %s%s\n", MiscLib.ColorYellow, key, godebug.LF(), MiscLib.ColorReset)
				hdlr.SetRedisKey(key, gt.Ttl, val) // xyzzy  - return TTL in seconds from remote valiate	-- should jsut set new TTL - but for now just update key
				goftlmux.AddValueToParams("$user_id$", gt.UserId, 'i', goftlmux.FromInject, ps)
				if hdlr.gCfg.DbOn("*", "LoginRequired", "db-validate-token") {
					fmt.Printf("Inject (b) of UserId ($user_id$) = [%s]\n", gt.UserId)
				}
				ttl = gt.Ttl
			}
			fmt.Fprintf(os.Stderr, "\n%s-- Auth Token Validated -- auth_token==[%s]  AT: %s %s\n\n", MiscLib.ColorGreen, iat, godebug.LF(), MiscLib.ColorReset)
			fmt.Fprintf(os.Stdout, "\n%s-- Auth Token Validated -- auth_token==[%s]  AT: %s %s\n\n", MiscLib.ColorGreen, iat, godebug.LF(), MiscLib.ColorReset)

			// set auth token as a value in this?
			goftlmux.AddValueToParams(hdlr.ParamName, iat, 'i', goftlmux.FromInject, ps)

		}

		valid = true
		return

	case "AesSrp":
		fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
		valid = true
		return

	case "ChallengeResponse":
		fmt.Fprintf(os.Stderr, "Challenge Response AT: %s\n", godebug.LF())
		valid = true
		return

	default:
		fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
		logrus.Errorf("Invalid ValidationSource at: %s", godebug.LF())
		vsOneOf := "pg, redis, jwt, AesSrp"
		fmt.Printf("Error - Invalid ValidationSource [%s] should be one of %s - LoginRequired - At: %s\n", hdlr.ValidationSource, vsOneOf, godebug.LF())
		fmt.Fprintf(os.Stderr, "%sError - Invalid ValidationSource [%s] should be one of %s - LoginRequired - At: %s%s\n",
			MiscLib.ColorRed, hdlr.ValidationSource, vsOneOf, godebug.LF(), MiscLib.ColorReset)
		return
	}
	if hdlr.gCfg.DbOn("*", "LoginRequired", "db-jwt-token") {
		fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
	}
	valid = true
	return
}

// SREM old key, SADD new key
// hdlr.RedisUpdateSet("ses__xsrf:"+sc, 60*60*24*30, xsrf_token, xt) // save the token locally for revalidation when new one fetched
func (hdlr *LoginRequiredType) RedisUpdateSet(key string, ttl int, new, old string) (err error) {
	if hdlr.gCfg.DbOn("*", "LoginRequired", "db-jwt-token") {
		fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
	}
	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	conn.Cmd("SREM", key, old)
	err = conn.Cmd("SADD", key, new).Err
	if err != nil {
		return
	}

	if ttl > 0 {
		err = conn.Cmd("EXPIRE", key, ttl).Err
	}
	return
}

// xref_in_set, err := hdlr.RedisSetContains("ses__xsrf:" + sc, xt) // trx__xsrf: shoud be set in config
// SMEMBERS on old key, 'xt'
func (hdlr *LoginRequiredType) RedisSetContains(key, item string) (found bool, err error) {
	if hdlr.gCfg.DbOn("*", "LoginRequired", "db-jwt-token") {
		fmt.Fprintf(os.Stderr, "AT: %s, key=%s item=%s\n", godebug.LF(), key)
	}
	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	found = false

	val, err := conn.Cmd("SISMEMBER", key, item).Str()
	if hdlr.gCfg.DbOn("*", "LoginRequired", "db-jwt-token") {
		fmt.Fprintf(os.Stderr, "AT: %s, val=%s err=%s\n", godebug.LF(), val, err)
	}
	if err != nil {
		return
	}
	if val == "1" {
		found = true
	} else {
		found = false
	}
	return
}

func (hdlr *LoginRequiredType) GetRedisKey(key string) (rv string, err error) {
	if hdlr.gCfg.DbOn("*", "LoginRequired", "db-jwt-token") {
		fmt.Fprintf(os.Stderr, "AT: %s, key=%s\n", godebug.LF(), key)
	}
	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	// key := hdlr.RedisAuthTokenPrefix + auth_token

	val, err := conn.Cmd("GET", key).Str()
	if err != nil {
		return
	}
	rv = val
	return
}

func (hdlr *LoginRequiredType) SetRedisKey(key string, ttl int, value string) (err error) {
	if hdlr.gCfg.DbOn("*", "LoginRequired", "db-jwt-token") {
		fmt.Fprintf(os.Stderr, "AT: %s\n", godebug.LF())
	}
	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	if ttl > 0 {
		err = conn.Cmd("SETEX", key, ttl, value).Err
	} else {
		err = conn.Cmd("SET", key, value).Err
	}
	return
}

/* OLD Documentation:

func (hdlr *BasicPgSQLHandlerType) pgGetUser(un, pw string) (userID string, ok bool) {
	var salt string
	var pwh string
	key := hdlr.Realm + ":" + un
	if hdlr.HashUsername {
		key = fmt.Sprintf("%x", pbkdf2.Key([]byte(key), []byte(hdlr.HashUsernameSalt), NIterations, 64, sha256.New))
	}
	rows, err := hdlr.gCfg.Pg_client.Db.Query("select \"salt\", \"password\", \"user_id\" from \"basic_auth\" where \"username\" = $1", key)
	if err != nil {
		fmt.Printf("Database error %s, attempting to validate user %s\n", err, un)
		return
	}

	for nr := 0; rows.Next(); nr++ {
		if nr >= 1 {
			fmt.Printf("Error too many rows for a user, should be unique primary key\n")
			return
		}

		err := rows.Scan(&salt, &pwh, &userID)
		if err != nil {
			fmt.Printf("Error on d.b. query %s\n", err)
			return
		}
		// fmt.Printf("%d. %s\n", id, description)
	}

	dk := fmt.Sprintf("%x", pbkdf2.Key([]byte(pw), []byte(salt), NIterations, 64, sha256.New))

	if db1 {
		fmt.Printf("salt [%s], pwh[%s], dk[%s]\n", salt, pwh, dk)
	}

	if subtle.ConstantTimeCompare([]byte(dk), []byte(pwh)) == 1 {
		if db1 {
			fmt.Printf("At: %s --------------- should be authoraized -- \n", lib.LF())
		}
		ok = true
	}
	return
}

*/

//const PreAuth = "aut:"
//
//var CookieSessionLife = 2 * 60 * 60 * 24 // 60 seconds * 60 minutes * 24 hours (1 day)
//
//func (hdlr *LoginRequiredType) SaveCookieAuth(cookieValue, ip, email, hash, id string, rw *goftlmux.MidBuffer) {
//	conn, err := hdlr.gCfg.RedisPool.Get()
//	defer hdlr.gCfg.RedisPool.Put(conn)
//	if err != nil {
//		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
//		return
//	}
//
//	err = conn.Cmd("SET", PreAuth+cookieValue, fmt.Sprintf(`{"ip":%q,"email":%q,"id":%q,"hash":%q}`, ip, email, id, hash)).Err
//	if err != nil {
//		fmt.Printf("Error on setting cookie key: %s\n", err)
//		return
//	}
//	conn.Cmd("EXPIRE", PreAuth+cookieValue, CookieSessionLife)
//}
//
//// --------------------------------------------------------------------------------------------------------------------------
//// Consider sending a 2nd token to client - and having it hash that with private key - then set that as a 2nd cookie.
//// --------------------------------------------------------------------------------------------------------------------------
//
//var ErrNoSuchUser = errors.New("User not found - no such user")
//
//func (hdlr *LoginRequiredType) GetCookieAuth(cookieValue string, rw *goftlmux.MidBuffer) (ip, email, id, hash string, err error) {
//
//	if cookieValue == "" {
//		err = ErrNoSuchUser
//		return
//	}
//
//	conn, err := hdlr.gCfg.RedisPool.Get()
//	defer hdlr.gCfg.RedisPool.Put(conn)
//	if err != nil {
//		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
//		return
//	}
//
//	var s string
//	s, err = conn.Cmd("GET", PreAuth+cookieValue).Str()
//	if err != nil {
//		fmt.Printf("Error on getting cookie key: %s\n", err)
//		err = ErrNoSuchUser
//		return
//	}
//
//	type jData struct {
//		Ip    string
//		Email string
//		Id    string
//		Hash  string
//	}
//	var rv jData
//
//	err = json.Unmarshal([]byte(s), &rv)
//	if err != nil {
//		fmt.Printf(`{"status":"error","msg":"Error(19913): %v - unable to unmarshal data from cookie save into Redis"}\n`, err)
//		err = ErrNoSuchUser
//		return
//	}
//
//	id = rv.Id
//	ip = rv.Ip
//	email = rv.Email
//	hash = rv.Hash
//
//	conn.Cmd("EXPIRE", PreAuth+cookieValue, CookieSessionLife)
//
//	return
//}

// Helper func:  Read input from specified file or stdin
func loadData(p string) ([]byte, error) {
	if p == "" {
		return nil, fmt.Errorf("No path specified")
	}

	var rdr io.Reader
	//	if p == "-" {
	//		rdr = os.Stdin
	//	} else if p == "+" {
	//		return []byte("{}"), nil
	//	} else {
	if f, err := os.Open(p); err == nil {
		rdr = f
		defer f.Close()
	} else {
		return nil, err
	}
	//	}
	return ioutil.ReadAll(rdr)
}

//-// Print a json object in accordance with the prophecy (or the command line options)
//-func printJSON(j interface{}) error {
//-	var out []byte
//-	var err error
//-
//-	if *flagCompact == false {
//-		out, err = json.MarshalIndent(j, "", "    ")
//-	} else {
//-		out, err = json.Marshal(j)
//-	}
//-
//-	if err == nil {
//-		fmt.Println(string(out))
//-	}
//-
//-	return err
//-}

// Verify a token and output the claims.  This is a great example
// of how to verify and view a token.
func (hdlr *LoginRequiredType) VerifyToken(tokData []byte, keyFile, keyType string) (iat string, err error) {

	return jwtverif.VerifyToken(tokData, keyFile, keyType)

	//	pwd, _ := os.Getwd()
	//	fmt.Fprintf(os.Stderr, "%sNew tokData ->%s<- keyFile ->%s<- pwd ->%s<- AT: %s%s\n", MiscLib.ColorYellow, tokData, keyFile, godebug.LF(), pwd, MiscLib.ColorReset)
	//
	//	// trim possible whitespace from token
	//	tokData = regexp.MustCompile(`\s*$`).ReplaceAll(tokData, []byte{})
	//	if hdlr.gCfg.DbOn("*", "LoginRequired", "db-validate-token") {
	//		fmt.Fprintf(os.Stderr, "Token len: %v bytes\n", len(tokData))
	//	}
	//
	//	// Parse the token.  Load the key from command line option
	//	token, err := jwt.Parse(string(tokData), func(t *jwt.Token) (interface{}, error) {
	//		data, err := loadData(keyFile)
	//		if err != nil {
	//			return nil, err
	//		}
	//		if isEs() {
	//			return jwt.ParseECPublicKeyFromPEM(data)
	//		} else if isRs() {
	//			return jwt.ParseRSAPublicKeyFromPEM(data)
	//		}
	//		return data, nil
	//	})
	//
	//	// Print some debug data
	//	if hdlr.gCfg.DbOn("*", "LoginRequired", "db-validate-token") && token != nil {
	//		fmt.Fprintf(os.Stderr, "Header:\n%v\n", token.Header)
	//		fmt.Fprintf(os.Stderr, "Claims:\n%v\n", token.Claims)
	//	}
	//
	//	// Print an error if we can't parse for some reason
	//	if err != nil {
	//		return "", fmt.Errorf("Couldn't parse token: %v", err)
	//	}
	//
	//	// Is token invalid?
	//	if !token.Valid {
	//		return "", fmt.Errorf("Token is invalid")
	//	}
	//
	//	if hdlr.gCfg.DbOn("*", "LoginRequired", "db-token") {
	//		fmt.Fprintf(os.Stderr, "Token Claims: %s\n", godebug.SVarI(token.Claims))
	//	}
	//
	//	// {"auth_token":"f5d8f6ae-e2e5-42c9-83a9-dfd07825b0fc"}
	//	type GetAuthToken struct {
	//		AuthToken string `json:"auth_token"`
	//	}
	//	var gt GetAuthToken
	//	cl := godebug.SVar(token.Claims)
	//	if hdlr.gCfg.DbOn("*", "LoginRequired", "db-jwt-token") {
	//		fmt.Fprintf(os.Stderr, "Claims just before -->>%s<<--\n", cl)
	//	}
	//	err = json.Unmarshal([]byte(cl), &gt)
	//	if err == nil {
	//		if hdlr.gCfg.DbOn("*", "LoginRequired", "db-jwt-token") {
	//			fmt.Fprintf(os.Stderr, "Success: %s -- token [%s] \n", err, gt.AuthToken)
	//		}
	//		fmt.Fprintf(os.Stdout, "Success: %s -- token [%s] \n", err, gt.AuthToken)
	//		return gt.AuthToken, nil
	//	} else {
	//		if hdlr.gCfg.DbOn("*", "LoginRequired", "db-jwt-token") {
	//			fmt.Fprintf(os.Stderr, "Error: %s -- Unable to unmarsal -->>%s<<--\n", err, cl)
	//		}
	//		fmt.Fprintf(os.Stdout, "Error: %s -- Unable to unmarsal -->>%s<<--\n", err, cl)
	//		return "", err
	//	}

}

//-// Create, sign, and output a token.  This is a great, simple example of
//-// how to use this library to create and sign a token.
//-func SignToken(tokData []byte) error {
//-
//-	// parse the JSON of the claims
//-	var claims jwt.MapClaims
//-	if err := json.Unmarshal(tokData, &claims); err != nil {
//-		return fmt.Errorf("Couldn't parse claims JSON: %v", err)
//-	}
//-
//-	// add command line claims
//-	if len(flagClaims) > 0 {
//-		for k, v := range flagClaims {
//-			claims[k] = v
//-		}
//-	}
//-
//-	// get the key
//-	var key interface{}
//-	key, err = loadData(*flagKey)
//-	if err != nil {
//-		return fmt.Errorf("Couldn't read key: %v", err)
//-	}
//-
//-	// get the signing alg
//-	alg := jwt.GetSigningMethod(*flagAlg)
//-	if alg == nil {
//-		return fmt.Errorf("Couldn't find signing method: %v", *flagAlg)
//-	}
//-
//-	// create a new token
//-	token := jwt.NewWithClaims(alg, claims)
//-
//-	// add command line headers
//-	if len(flagHead) > 0 {
//-		for k, v := range flagHead {
//-			token.Header[k] = v
//-		}
//-	}
//-
//-	if isEs() {
//-		if k, ok := key.([]byte); !ok {
//-			return fmt.Errorf("Couldn't convert key data to key")
//-		} else {
//-			key, err = jwt.ParseECPrivateKeyFromPEM(k)
//-			if err != nil {
//-				return err
//-			}
//-		}
//-	} else if isRs() {
//-		if k, ok := key.([]byte); !ok {
//-			return fmt.Errorf("Couldn't convert key data to key")
//-		} else {
//-			key, err = jwt.ParseRSAPrivateKeyFromPEM(k)
//-			if err != nil {
//-				return err
//-			}
//-		}
//-	}
//-
//-	if out, err := token.SignedString(key); err == nil {
//-		fmt.Println(out)
//-	} else {
//-		return fmt.Errorf("Error signing token: %v", err)
//-	}
//-
//-	return nil
//-}

//-// showToken pretty-prints the token on the command line.
//-func showToken() error {
//-	// get the token
//-	tokData, err := loadData(*flagShow)
//-	if err != nil {
//-		return fmt.Errorf("Couldn't read token: %v", err)
//-	}
//-
//-	// trim possible whitespace from token
//-	tokData = regexp.MustCompile(`\s*$`).ReplaceAll(tokData, []byte{})
//-	if db1 {
//-		fmt.Fprintf(os.Stderr, "Token len: %v bytes\n", len(tokData))
//-	}
//-
//-	token, err := jwt.Parse(string(tokData), nil)
//-	if token == nil {
//-		return fmt.Errorf("malformed token: %v", err)
//-	}
//-
//-	// Print the token details
//-	fmt.Println("Header:")
//-	if err := printJSON(token.Header); err != nil {
//-		return fmt.Errorf("Failed to output header: %v", err)
//-	}
//-
//-	fmt.Println("Claims:")
//-	if err := printJSON(token.Claims); err != nil {
//-		return fmt.Errorf("Failed to output claims: %v", err)
//-	}
//-
//-	return nil
//-}

//func isEs() bool {
//	// return strings.HasPrefix(*flagAlg, "ES")
//	return true
//}
//
//func isRs() bool {
//	// return strings.HasPrefix(*flagAlg, "RS")
//	return false
//}

//------------------------------------------------------------------------------------------------------------------------------
func GetURL(uri string, XssiPrefix []string, args ...string) (status int, rv string) {

	sep := "?"
	var qq bytes.Buffer
	qq.WriteString(uri)
	for ii := 0; ii < len(args); ii += 2 {
		// q = q + sep + name + "=" + value;
		qq.WriteString(sep)
		qq.WriteString(url.QueryEscape(args[ii]))
		qq.WriteString("=")
		if ii < len(args) {
			qq.WriteString(url.QueryEscape(args[ii+1]))
		}
		sep = "&"
	}
	url_q := qq.String()

	res, err := http.Get(url_q)
	if err != nil {
		return 500, ""
	} else {
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return 500, ""
		}
		status = res.StatusCode
		if status == 200 {

			rv = string(body)

			// for _, pp := range []string{"//", ")]}'", "while(1);", "while(true);", "for(;;);"} { // xyzzy - hdlr.PrefixList
			for _, pp := range XssiPrefix {
				if len(rv) > len(pp) && rv[0:len(pp)] == pp {
					rv = rv[len(pp):]
					break
				}
			}

		}
		return
	}
}

func (hdlr *LoginRequiredType) InjextXsrf(rw *goftlmux.MidBuffer, www http.ResponseWriter, xsrf_token string) {

	h := www.Header()
	ct := h.Get("Content-Type")
	if rw.StatusCode == http.StatusOK && strings.HasPrefix(ct, "application/json") {
		if hdlr.gCfg.DbOn("*", "LoginRequired", "db-injext-xsrf") {
			fmt.Printf("\nInjextXsrf: is JSON %s\n", godebug.LF())
		}
		mdata := make(map[string]interface{})
		body := rw.GetBody()
		if hdlr.gCfg.DbOn("*", "LoginRequired", "db-injext-xsrf") {
			fmt.Printf("InjextXsrf: body -->>%s<<-- %s\n", body, godebug.LF())
		}
		err := json.Unmarshal(body, &mdata)
		if err != nil {
			if hdlr.gCfg.DbOn("*", "LoginRequired", "db-injext-xsrf") {
				fmt.Printf("InjextXsrf: Failed to parse - 1st try with hash, data=%s err=%s, %s\n", body, err, godebug.LF())
			}
			body = []byte("{\"data\":" + string(body) + "}")
			err = json.Unmarshal(body, &mdata)
			fmt.Fprintf(os.Stdout, "%sModified Data --->%s<--- err=%s %s\n", MiscLib.ColorRed, body, err, MiscLib.ColorReset)
		}
		if err != nil {
			if hdlr.gCfg.DbOn("*", "LoginRequired", "db-injext-xsrf") {
				fmt.Printf("InjextXsrf: Failed to parse - 2nd try with created hash, data=%s err=%s, %s\n", body, err, godebug.LF())
			}
			//if hdlr.OnErrorDiscard == "yes" {
			if hdlr.gCfg.DbOn("*", "LoginRequired", "db-jwt-token") {
				fmt.Fprintf(os.Stderr, "%sData Discarded - due to syntax error ->%s<- %s%s\n", MiscLib.ColorRed, body, godebug.LF(), MiscLib.ColorReset)
			}
			fmt.Fprintf(os.Stdout, "%sData Discarded - due to syntax error ->%s<- %s%s\n", MiscLib.ColorRed, body, godebug.LF(), MiscLib.ColorReset)
			www.WriteHeader(http.StatusInternalServerError)
			rw.ReplaceBody([]byte("{}"))
			rw.SaveDataInCache = false
			return
			//}
		} else {
			if hdlr.gCfg.DbOn("*", "LoginRequired", "db-injext-xsrf") {
				fmt.Printf("\nInjextXsrf: will proces %s\n", godebug.LF())
			}
			mdata["$xsrf_token$"] = xsrf_token
			newData := godebug.SVar(mdata)
			rw.ReplaceBody([]byte(newData))
			if hdlr.gCfg.DbOn("*", "LoginRequired", "db-injext-xsrf") {
				fmt.Printf("InjextXsrf: newData -->>%s<<-- %s\n", newData, godebug.LF())
			}
			rw.SaveDataInCache = true
		}
	}
}

const enableXsrfCheckOn = true // OPTION: if true, then xsrf_token is checked

/* vim: set noai ts=4 sw=4: */
