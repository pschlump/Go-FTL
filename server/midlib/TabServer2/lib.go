//
// R E S T s e r v e r - Server Component	(tab-server1.go/new dispatcher)
// main program - and a bunch of support functions that should be moved.
//
// Copyright (C) Philip Schlump, 2012-2016. All rights reserved.
// Version: 1.1.0
// BuildNo: 0274
// FileId: 0001
//

package TabServer2

import (
	// _ "github.com/lib/pq"

	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/hjson/hjson-go"
	"github.com/pschlump/Go-FTL/server/common"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/json" //	"encoding/json"
	"github.com/pschlump/ms"
	"github.com/pschlump/user_agent"
	"github.com/pschlump/uuid"
)

const jsonOld = false // false == use hjson

const MAX_MEMORY = 10 * 1024 * 1024

// pprof --- http://golang.org/pkg/net/http/pprof/
// Looking at output of pprof --- http://google-perftools.googlecode.com/svn/trunk/doc/cpuprofile.html
//	"log"
//	 _ "net/http/pprof"

var fo *os.File
var fx *os.File

var debug_1 bool

// var g_LimitPostJoinRows int = -1 // -1 indicates no limit

const db_cache = false
const db_auth = true
const db_user_login = false
const db_where_collect = false
const db_post_join = true
const db_fuzzy_date = false
const db_ticker = false

var Init_Main []func(theMux *goftlmux.MuxRouter)

//func loadToken(s string, flag string) {
//	rr.RedisDo("SET", "csrf:token:"+flag, s)
//}

var DbBeginQuote = `"`
var DbEndQuote = `"`

func loadAllCsrfTokens(hdlr *TabServer2Type) {

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		// logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		// xyzzy log stuff
		return
	}

	defer hdlr.gCfg.RedisPool.Put(conn)

	// fmt.Printf("At %s ------------------------------------------*** \n", godebug.LF())
	// xyzzy222 - need to abstract this out to a config file!
	// Load CSRF tokens from MS-SQL to Redis
	// xyzzy - Oracle?? -- what about oracle???
	// if GlobalCfg["connectToDbType"] == "odbc" {
	if hdlr.GetDbType() == DbType_odbc {
		DbBeginQuote = `[`
		DbEndQuote = `]`
		// fmt.Printf("Before: Setting to Test for Database\n")
		// err := sizlib.Run1(hdlr.gCfg.Pg_client.Db, `use Test`)
		// if err != nil {
		// 	fmt.Printf("Unable to set database, %s\n", err)
		// }
		// fmt.Printf("After Setting to Test for Database\n")
		tokens1 := sizlib.SelData(hdlr.gCfg.Pg_client.Db, `select [token], [created] from [t_csrf_token] order by [created]`)

		// loadToken(tokens1[0]["token"].(string), "login")
		conn.Cmd("SET", tokens1[0]["token"].(string), "login")

		// loadToken(tokens1[len(tokens1)-1]["token"].(string), "regular")
		conn.Cmd("SET", tokens1[len(tokens1)-1]["token"].(string), "regular")

		tokens2 := sizlib.SelData(hdlr.gCfg.Pg_client.Db, `select [token], [created] from [t_csrf_token2] order by [created]`)
		for i := range tokens2 {
			// loadToken(tokens2[i]["token"].(string), "cookie")
			conn.Cmd("SET", tokens2[i]["token"].(string), "cookie")
		}
	} else {
		// Load CSRF tokens from PostgreSQL to Redis
		tokens1 := sizlib.SelData(hdlr.gCfg.Pg_client.Db, `select "token", "created" from "t_csrf_token" order by "created"`)

		//loadToken(tokens1[0]["token"].(string), "login")
		conn.Cmd("SET", tokens1[0]["token"].(string), "login")

		//loadToken(tokens1[len(tokens1)-1]["token"].(string), "regular")
		conn.Cmd("SET", tokens1[len(tokens1)-1]["token"].(string), "regular")

		tokens2 := sizlib.SelData(hdlr.gCfg.Pg_client.Db, `select "token", "created" from "t_csrf_token2" order by "created"`)
		for i := range tokens2 {
			// loadToken(tokens2[i]["token"].(string), "cookie")
			conn.Cmd("SET", tokens2[i]["token"].(string), "cookie")

		}
	}
}

func GetIpFromRemoteAddr(RemoteAddr string) (rv string) {
	n := strings.LastIndex(RemoteAddr, ":")
	rv = RemoteAddr
	if n > 0 && n < len(RemoteAddr) {
		rv = RemoteAddr[:n]
	}
	return
}

// Consider caching the tokens in memory for this server - fetch onece from Redis then cache.
func GetByNameWTrxLog(name string, ps *goftlmux.Params, err string, msg string, req *http.Request, trx *tr.Trx) (xerr error, xal string) {
	xerr = nil
	xal = ""
	// fmt.Printf("top:%s\n", name)
	if xal = ps.ByName(name); xal != "" {
		// fmt.Printf("found in using ByName:%s = ->%s<-, %s\n", name, xal, godebug.LF())
	} else {
		xerr = errors.New(err)
		s := fmt.Sprintf(msg, req.RemoteAddr, req.RequestURI)
		// fmt.Printf("%s, %s\n", s, godebug.LF())
		trx.AddNote(2, s)
	}
	return
}

func checkCsrfTokens(isLogin bool, req *http.Request, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type, rw *goftlmux.MidBuffer) (err error) {

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	defer hdlr.gCfg.RedisPool.Put(conn)

	l_csrf_token := ""
	// ok := true
	if db_auth {
		fmt.Printf("At %s\n", godebug.LF())
	}

	err, l_csrf_token = GetByNameWTrxLog("csrf_token", ps, "Missing Parameter:csrf_token",
		"Error(12069): Request requreis a 'csrf_token' parameter.  It was not suplied. URL=%s/%s", req, trx)
	/*
		if _, ok = m["csrf_token"]; !ok {
			if l_csrf_token, ok := ps.ByName("csrf_token"); ! ok {
				err = errors.New("Missing Parameter:csrf_token")
				fmt.Printf("Error(12069): Request requreis a 'csrf_token' parameter.  It was not suplied. URL=%s/%s\n", req.RemoteAddr, req.RequestURI)
			}
		} else {
			l_csrf_token = m["csrf_token"][0]
		}
	*/

	if isLogin {
		if db_auth {
			fmt.Printf("At %s\n", godebug.LF())
		}
		// f_csrf_token, err := redis.String(rr.RedisDo("GET", "csrf:token:login"))
		aKey := "csrf:token:login"
		f_csrf_token, err := conn.Cmd("GET", aKey).Str() // Get the value
		if err != nil {
			err = errors.New("Missing csrf_token in redis")
		} else {
			if f_csrf_token != l_csrf_token {
				err = errors.New("Invalid csrf_token")
			}
		}
	} else {

		if db_auth {
			fmt.Printf("At %s\n", godebug.LF())
		}
		// f_csrf_token, err := redis.String(rr.RedisDo("GET", "csrf:token:login"))
		aKey := "csrf:token:login"
		f_csrf_token, err := conn.Cmd("GET", aKey).Str() // Get the value
		if err != nil {
			// fmt.Printf("At %s\n", godebug.LF())
			if db_auth {
				fmt.Printf("At %s\n", godebug.LF())
			}
			err = errors.New("Missing csrf_token in redis")
		} else {
			if db_auth {
				fmt.Printf("At %s f ->%s<- l ->%s<-\n", godebug.LF(), f_csrf_token, l_csrf_token) // this one
			}
			if f_csrf_token != l_csrf_token {
				if db_auth {
					fmt.Printf("At %s\n", godebug.LF())
				}
				err = errors.New("Invalid csrf_token")
			}
		}

		if db_auth {
			fmt.Printf("At %s -- 1st tokens matched -- \n", godebug.LF())
		}
		if err == nil {

			if db_auth {
				fmt.Printf("At %s\n", godebug.LF())
			}
			err, l_csrf_token2 := GetByNameWTrxLog("cookie_csrf_token", ps, "Missing Parameter:cookie_csrf_token",
				"Error(12070): Request requreis a 'cookie_csrf_token' parameter.  It was not suplied. URL=%s/%s\n", req, trx)
			/*
				l_csrf_token2 := ""
				if _, ok = m["cookie_csrf_token"]; !ok {
					// fmt.Printf("At %s\n", godebug.LF())
					err = errors.New("Missing Parameter:cookie_csrf_token")
					fmt.Printf("Error(12070): Request requreis a 'cookie_csrf_token' parameter.  It was not suplied. URL=%s/%s\n", req.RemoteAddr, req.RequestURI)
				} else {
					// fmt.Printf("At %s\n", godebug.LF())
					l_csrf_token = m["cookie_csrf_token"][0]
				}
			*/
			// f_csrf_token2, err := redis.String(rr.RedisDo("GET", "csrf:token:cookie"))
			aKey := "csrf:token:cookie"
			f_csrf_token2, err := conn.Cmd("GET", aKey).Str() // Get the value
			if err != nil {
				if db_auth {
					fmt.Printf("At %s\n", godebug.LF())
				}
				err = errors.New("Missing csrf_token2 in redis")
			} else {
				if db_auth {
					fmt.Printf("At %s\n", godebug.LF())
				}
				if f_csrf_token2 != l_csrf_token2 {
					if db_auth {
						fmt.Printf("At %s\n", godebug.LF())
					}
					err = errors.New("Invalid csrf_token2")
				}
			}

		}
	}

	if db_auth {
		fmt.Printf("At %s, RETURN, err=->%v<-\n", godebug.LF(), err)
	}

	return err
}

var db_valid = false
var MissingValueError = errors.New("Missing Value")
var InvalidUser = errors.New("Invalid user")

func (hdlr *TabServer2Type) ValidateUserTrx(h SQLOne, trx *tr.Trx, res http.ResponseWriter, req *http.Request, ps *goftlmux.Params, rw *goftlmux.MidBuffer) (err error) {

	err = nil

	isLogin := false
	if !h.LoginRequired {
		isLogin = true
	}

	if hdlr.loginSystem == LstAesSrp {
		trx.AddNote(1, "Start of AesSrp autorization check")
		if db_auth {
			fmt.Printf("At %s\n", godebug.LF())
		}
		is_logged_in := ps.ByNameDflt("$is_logged_in$", "")
		is_full_login := ps.ByNameDflt("$is_full_login$", "")
		if is_logged_in == "y" && is_full_login == "y" {
			if db_auth {
				fmt.Printf("At %s\n", godebug.LF())
			}
			id0, _ := uuid.NewV4()
			f_auth_token := id0.String()
			goftlmux.AddValueToParams("auth_token", f_auth_token, 'i', goftlmux.FromInject, ps)
			trx.AddNote(1, fmt.Sprintf("Error(13073): Authorization passed. URL=%s/%s", req.RemoteAddr, req.RequestURI))
			return
		} else {
			trx.AddNote(1, fmt.Sprintf("Error(13063): Authorization requires an 'AesSrp' login that was not performed. URL=%s/%s", req.RemoteAddr, req.RequestURI))
			err = fmt.Errorf("Login required, AesSrp authentication")
			return
		}
	}

	if hdlr.loginSystem == LstBasic {
		trx.AddNote(1, "Start of Basic autorization check")
		if db_auth {
			fmt.Printf("At %s\n", godebug.LF())
		}
		is_logged_in := ps.ByNameDflt("$is_logged_in$", "")
		if is_logged_in == "y" {
			if db_auth {
				fmt.Printf("At %s\n", godebug.LF())
			}
			id0, _ := uuid.NewV4()
			f_auth_token := id0.String()
			goftlmux.AddValueToParams("auth_token", f_auth_token, 'i', goftlmux.FromInject, ps)
			trx.AddNote(1, fmt.Sprintf("Error(13074): Authorization passed. URL=%s/%s", req.RemoteAddr, req.RequestURI))
			return
		} else {
			trx.AddNote(1, fmt.Sprintf("Error(13064): Authorization requires a 'Basic' login that was not performed. URL=%s/%s", req.RemoteAddr, req.RequestURI))
			err = fmt.Errorf("Login required, Basic authentication")
			return
		}
	}

	if hdlr.loginSystem != LstUnPw {
		trx.AddNote(1, fmt.Sprintf("Error(13065): Authorization requires a some form of login that was not performed. Internal configuration error. URL=%s/%s", req.RemoteAddr, req.RequestURI))
		err = fmt.Errorf("Login required")
		return
	}

	trx.AddNote(1, "Start of autorization")
	l_auth_token := ""
	l_username := ""
	// l_seq := ""
	/*
		ok := false
		if _, ok = m["username"]; !ok {
			if isRequired(h, "username") {
				err = errors.New("Missing Parameter:user")
				fmt.Printf("Error(12067): Request requreis a 'username' parameter.  It was not suplied. URL=%s/%s\n", req.RemoteAddr, req.RequestURI)
				return
			}
		} else {
			l_username = m["username"][0]
		}
		if _, ok = m["auth_token"]; !ok {
			err = errors.New("Missing Parameter:auth_token")
			fmt.Printf("Error(12066): Request requreis a 'auth_token' parameter.  It was not suplied. URL=%s/%s\n", req.RemoteAddr, req.RequestURI)
			return
		} else {
			l_auth_token = m["auth_token"][0]
		}
	*/

	if db_auth {
		fmt.Printf("At %s\n", godebug.LF())
	}

	err, l_username = GetByNameWTrxLog("username", ps, "Missing Parameter:username", "Error(12067): Request requreis a 'auth_token' parameter.  It was not suplied. URL=%s/%s\n", req, trx)
	if err != nil {
		if db_auth {
			fmt.Printf("At %s\n", godebug.LF())
		}
		if isRequired(h, "username") {
			if db_auth {
				fmt.Printf("At %s\n", godebug.LF())
			}
			return
		}
	}

	if db_auth {
		fmt.Printf("At %s\n", godebug.LF())
	}

	l_auth_token = req.Header.Get("X-Auth-Token")
	if l_auth_token == "" {
		err, l_auth_token = GetByNameWTrxLog("auth_token", ps, "Missing Parameter:auth_token",
			"Error(12066): Request requreis a 'auth_token' parameter.  It was not suplied. URL=%s/%s\n", req, trx)
		if err != nil {
			if db_auth {
				fmt.Printf("Failed X-Auth-Token At %s\n", godebug.LF())
			}
			return
		}
	}

	if db_auth {
		fmt.Printf("At %s\n", godebug.LF())
	}
	if err = checkCsrfTokens(isLogin, req, ps, trx, hdlr, rw); err != nil {
		fmt.Printf("Error(12065): Request requreis a 'csrf_token' and 'cookie_csrf_token' parameter.  It was not suplied. URL=%s/%s\n", req.RemoteAddr, req.RequestURI)
		trx.AddNote(1, fmt.Sprintf("Error(12073): Authorization requires a 'csrf_token' parameter that was not supplied. URL=%s/%s", req.RemoteAddr, req.RequestURI))
		if db_auth {
			fmt.Printf("Failed checkCsrfTokens At %s\n", godebug.LF())
		}
		return
	}

	if db_auth {
		fmt.Printf("At %s\n", godebug.LF())
	}

	// fmt.Printf("At %s -- Just before changed secion for X-XSRF-TOKEN\n", godebug.LF())
	x_xsrf_token := req.Header.Get("X-XSRF-TOKEN")
	trx.AddNote(1, fmt.Sprintf("X-XSRF-TOKEN from header ->%s<-", x_xsrf_token))
	if x_xsrf_token == "" {
		// fmt.Printf("At %s\n", godebug.LF())
		x_xsrf_token = ps.ByName("XSRF-TOKEN")
	}

	if db_auth {
		fmt.Printf("x_xsrf_token = %s At %s\n", x_xsrf_token, godebug.LF())
	}

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	defer hdlr.gCfg.RedisPool.Put(conn)

	// dt, err := redis.String(rr.RedisDo("GET", "api:USER:"+l_auth_token))
	aKey := "api:USER:" + l_auth_token
	dt, err := conn.Cmd("GET", aKey).Str() // Get the value
	if err != nil {
		if db_auth {
			fmt.Printf("At %s -- did not find value in Redis!!, '' get api:USER:%s ''\n", godebug.LF(), l_auth_token)
		}
		err = InvalidUser
		// chec in d.b. if user is still valid - then refresh in Redis??
		rvdata := sizlib.SelData(hdlr.gCfg.Pg_client.Db, `select "username", "id", "customer_id", "privs" from "t_user" where "auth_token" = $1`, l_auth_token)
		if len(rvdata) == 1 {
			err = nil
			if db_auth {
				fmt.Printf("At %s -- found user in PostgreSQL!!\n", godebug.LF())
			}

			// f_username := rvdata[0]["username"].(string)       // make this a func - and check to see if value is not NULL, with default
			// f_user_id := rvdata[0]["id"].(string)              //
			// f_customer_id := rvdata[0]["customer_id"].(string) //
			//x_privs, x_ok := rvdata[0]["privs"]                //
			//f_privs := ""                                      //
			//if x_ok && x_privs != nil {
			//	f_privs = rvdata[0]["privs"].(string)
			//} else {
			//	f_privs = "{}"
			//}

			f_username := sizlib.DbValueFromRow(rvdata, 0, "username", "")
			f_user_id := sizlib.DbValueFromRow(rvdata, 0, "user_id", "1")
			f_customer_id := sizlib.DbValueFromRow(rvdata, 0, "customer_id", "1")
			f_privs := sizlib.DbValueFromRow(rvdata, 0, "privs", "{}")
			if db_auth {
				fmt.Printf("At %s -- %s %s %s\n", godebug.LF(), f_username, f_user_id, f_customer_id)
			}
			id0, _ := uuid.NewV4()
			f_xsrf_token := id0.String()
			dt := fmt.Sprintf(`{"master":%q,"username":%q,"user_id":%q,"XSRF-TOKEN":%q,"customer_id":%q}`, f_privs, f_username, f_user_id, f_xsrf_token, f_customer_id)

			//rr.RedisDo("SET", "api:USER:"+l_auth_token, dt)
			//rr.RedisDo("EXPIRE", "api:USER:"+l_auth_token, 1*60*60) // Validate for 1 hour

			aKey := "api:USER:" + l_auth_token
			conn.Cmd("SET", aKey, dt)
			conn.Cmd("EXPIRE", aKey, 1*60*60)

			if db_auth {
				fmt.Printf("At %s -- Should be set in redis now with 1 hour expire!!\n", godebug.LF())
			}

			goftlmux.AddValueToParams("username", f_username, 'i', goftlmux.FromInject, ps)
			goftlmux.AddValueToParams("$user_id$", f_user_id, 'i', goftlmux.FromInject, ps)
			goftlmux.AddValueToParams("$privs$", f_privs, 'i', goftlmux.FromInject, ps)
			if !ps.HasName("$customer_id$") {
				goftlmux.AddValueToParams("$customer_id$", f_customer_id, 'i', goftlmux.FromInject, ps)
			}
			expire := time.Now().AddDate(0, 0, 2) // Years, Months, Days==2 // xyzzy - should be a config - on how long to keep cookie
			secure := false
			if req.TLS != nil {
				secure = true
			}
			vv := "XSRF-TOKEN"
			cookie := http.Cookie{Name: vv, Value: f_xsrf_token, Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400, Secure: secure, HttpOnly: false}
			http.SetCookie(res, &cookie)
		}
		return
	} else {
		if db_auth {
			fmt.Printf("Had the cached redis auth-token, At %s\n", godebug.LF())
		}
		var x map[string]interface{}
		x, err = sizlib.JSONStringToData(dt)
		if err != nil {
			// fmt.Printf("At %s\n", godebug.LF())
			err = errors.New("Invalid user")
		} else {
			// fmt.Printf("At %s\n", godebug.LF())
			err = errors.New("Invalid user")
			f_username := x["username"].(string)
			f_user_id := x["user_id"].(string)
			f_customer_id := "1"
			if _, ok := x["customer_id"]; ok {
				f_customer_id = x["customer_id"].(string)
			}
			if l_username == "" {
				goftlmux.AddValueToParams("username", f_username, 'i', goftlmux.FromInject, ps)
				f_username = l_username
				if db_auth {
					fmt.Printf("Just looks like an error to me\n")
				}
			}
			goftlmux.AddValueToParams("$user_id$", f_user_id, 'i', goftlmux.FromInject, ps)
			if !ps.HasName("$customer_id$") {
				goftlmux.AddValueToParams("$customer_id$", f_customer_id, 'i', goftlmux.FromInject, ps)
			}
			// fmt.Printf("At %s, f_username=%s l_username=%s\n", godebug.LF(), f_username, l_username)
			f_xsrf_token := x["XSRF-TOKEN"].(string)
			if f_username == l_username {
				// fmt.Printf("At %s\n", godebug.LF())
				if x_xsrf_token == f_xsrf_token {
					// fmt.Printf("At %s -- Successful AUTH!\n", godebug.LF())
					if db_auth {
						fmt.Printf("XSRF-TOKEN matched, %s\n", godebug.LF())
					}
					privs := x["master"].(string)
					XSRF_TOKEN := x["XSRF-TOKEN"].(string)
					// n.Add("user_master", privs)
					if db_auth {
						fmt.Printf("Injecting privs, ->%s<-, %s\n", privs, godebug.LF())
					}
					goftlmux.AddValueToParams("$privs$", privs, 'i', goftlmux.FromInject, ps)
					dt := fmt.Sprintf(`{"master":%q, "username":%q, "user_id":%q, "XSRF-TOKEN":%q, "customer_id":%q }`, privs, f_username, f_user_id, XSRF_TOKEN, f_customer_id)
					//rr.RedisDo("SET", "api:USER:"+l_auth_token, dt)
					//rr.RedisDo("EXPIRE", "api:USER:"+l_auth_token, 1*60*60) // Validate for 1 hour
					aKey := "api:USER:" + l_auth_token
					conn.Cmd("SET", aKey, dt)
					conn.Cmd("EXPIRE", aKey, 1*60*60)
					err = nil
				} else {
					// fmt.Printf("At %s\n", godebug.LF())
					trx.AddNote(1, fmt.Sprintf("Failed on XSRF-TOKEN, x=%s f=%s", x_xsrf_token, f_xsrf_token))
				}
			}
		}
	}
	// fmt.Printf("At %s\n", godebug.LF())
	return
}

// -------------------------------------------------- New --------------------------------------------------
// test: t-pp1.go
type Validation struct {
	Required  bool
	Type      string
	Min_len   int
	Max_len   int
	Min       int64
	Max       int64
	MinF      float64
	MaxF      float64
	MinD      time.Time
	MaxD      time.Time
	Default   string
	UrlEncode bool
	ReMatch   string
	ChkType   bool

	eRequired  bool
	eType      bool
	eMin_len   bool
	eMax_len   bool
	eMin       bool
	eMax       bool
	eMinF      bool
	eMaxF      bool
	eMinD      bool
	eMaxD      bool
	eDefault   bool
	eUrlEncode bool
	eReMatch   bool
	eChkType   bool
}

type ValidationIn struct {
	Required  *bool
	Type      *string
	Min_len   *int
	Max_len   *int
	Min       *int64
	Max       *int64
	MinF      *float64
	MaxF      *float64
	MinD      *time.Time
	MaxD      *time.Time
	Default   *string
	UrlEncode *bool
	ReMatch   *string
	ChkType   *bool
}

// test: t-pp1.go
type ColSpec struct {
	ColName     string   //
	ColAlias    string   //
	ColType     string   //
	ColLen      int      //
	IsPk        bool     //
	IsIndexed   bool     //
	Insert      bool     //
	Update      bool     //
	AutoGen     bool     //
	OrderBy     string   //
	Values      []string //
	DefaultData string   // Default data to use if not supplied in insert
	DolNo       int      //
	DolNoUpd    int      //
	NoSort      bool     //
	ColTitle    string   //
	DataColName string   // Column to use from user data for insert "ColName" in table
}

// test: t-pp1.go
type OrdSpec struct {
	ColName string
	Dir     string
}

// deleteBehavior: { "colName": "isDeleted", "colType": "s|b|i", "Deleted":"1", "Present":"0" }
// , "deleteViaUpdate": { "colType":"i", "colName":"isDeleted", "Absent":1, "Present":0 }
type DelBehavior struct {
	ColType  string
	ColName  string
	Absent   string
	Present  string
	ColAlias string
}

type CustomerIdPartBehavior struct {
	ColName  string
	ColType  string // only 's' or 'i' make sence for partitioning customers
	ColAlias string
}

//	, PostJoin: [
//				{ "ColName": "bob", "ColType":"s", "Query": "select...", "p":[ ... ], "SetCol": "bobArray" }
//		]
type PostJoinType struct {
	ColName  string
	ColType  string // only 's' or 'i' make sence for partitioning customers
	ColAlias string
	Query    string
	P        []string
	SetCol   string
	PostJoin []PostJoinType
}

type DbOperation struct {
	Op         string   // One of "select", "call", "trxNote"
	Q          string   // Primary data parameter, "call" == stored procedure name, "select" == query, "trxNote" == the note
	P          []string // Parameters to use
	Done       string   // if != "", then what to do after this - this is "done"
	Onerror    string   // OnError - If an error occured in a query, then what to do
	Onsuccess  string   // OnSuccess - If success occured, then "done" => return data, "discard" => Discard data on success, continue
	CallBefore []string
	CallAfter  []string
}

type RemapParameterType struct {
	FromName string
	ToName   string
}

// test: t-pp1.go
type SQLOne struct {
	F                          string                  // Depricated in favor of Fx or G
	Fx                         string                  //
	G                          string                  //
	Gx                         string                  //	Extended function call, instead of just return 'x' will look at names of function and return and use that
	P                          []string                //
	Popt                       []string                //
	Pname                      []string                //
	Exec                       []DbOperation           //
	SetCookie                  map[string]bool         //
	SetSession                 map[string]bool         //	// don't know if this is correct yet //
	Query                      string                  //
	Valid                      map[string]ValidationIn //
	ValidGet                   map[string]ValidationIn //
	ValidPost                  map[string]ValidationIn //
	ValidPut                   map[string]ValidationIn //
	ValidDel                   map[string]ValidationIn //
	LoginRequired              bool                    // Defaults to false, no login required (nokey)
	Redis                      bool                    //
	CacheIt                    string                  //
	TableName                  string                  // xyzzy921 - Should override URL - not doing this currently - need to do this. -- Should be put in TableList also xyzzy
	Method                     []string                //
	Crud                       []string                //
	Cols                       []ColSpec               //
	OrderBy                    []OrdSpec               //
	ReadOnly                   string                  //
	NoLog                      []string                // Skip logging of these parametes to the call.  Repalce each chareacter in Trx log with '*'
	CallBefore                 []string                //
	CallAfter                  []string                //
	TableList                  []string                //
	LineNo                     string                  //
	SelectTmpl                 string                  //
	SelectPK1Tmpl              string                  //
	SelectCountTmpl            string                  //
	InsertTmpl                 string                  //
	UpdateTmpl                 string                  //
	DeleteTmpl                 string                  //
	ReturnGetPKAsHash          bool                    // On get via Primary Key - should a 1 long array be returned or just the hash inside the array, true if hash -- Overidden by ReturnMeta, ReturnAsHash
	ReturnGetPKAsHashTableName bool                    // If ReturnGetPkAsHash is true then either use "data" or if this is true use the "assigned_name"
	AssignedName               string                  // if "", then use TableName if ReturnGetPKAsHashTableName is true.
	ReturnMeta                 bool                    // Return as { "data": [ ... ], "meta": { "count": n } }
	ReturnAsHash               bool                    // Return as { "data": [ ... ] }
	CmdList                    []string                //
	DebugFlag                  []string                // set of debuging flags true/flase

	FromKey string

	valid     map[string]Validation
	validGet  map[string]Validation // mm22mm
	validPost map[string]Validation
	validPut  map[string]Validation
	validDel  map[string]Validation

	DeleteViaUpdate DelBehavior // Specification for delete/undelete with marker in row.
	CustomerIdPart  CustomerIdPartBehavior

	SetWhereAlias string
	setWhereAlias string

	PostJoin      []PostJoinType
	CachePostJoin bool

	/*
	   , "key_word_col_name": "key_word"
	   , "key_word_list_col": "key_word_list_id"
	   , "key_word_tmpl": " in ( select k1.\"id\" from \"p_key_word_list\" as k1, \"p_key_word\" as k2 where k1.\"word_id\" = k2.\"id\" and k2.\"word\" %{kw_op%} %{kw_vals%} ) "

	   , "category_col_name": "category"
	   , "category_col": "category_id"
	   , "category_tmpl": " in ( select c1.\"id\" from \"p_category\" as c1 where c1.\"category_name\" %{cat_op%} %{cat_val%} ) "

	   , "attr_table_name": "p_cart"
	   , "attr_col": "id"
	   , "attr_tmpl":" in ( select a1.\"fk_id\" from \"p_attr\" as a1 where a1.\"attr_type\" = %{attr_type%} and a1.\"attr_name\" = %{attr_name%} and a1.\"%{ref_col%}\" %{attr_op%} %{attr_val%} )"
	*/

	Key_word_col_name string
	Key_word_list_col string
	Key_word_tmpl     string

	Category_col_name string
	Category_col      string
	Category_tmpl     string

	Attr_table_name string
	Attr_col        string
	Attr_tmpl       string

	ReMapParameter []RemapParameterType

	Comment            string
	Status             string `json:"status"`
	OracleSequenceName string // xyzzyOracle - not used yet
}

// test: t-pp1.go
// var SQLCfg map[string]SQLOne

//------------------------------------------------------------------------------------------------
// Save tables referened to Redis
//
// Note:
// 	h := SQLCfg[cfgTag]
//------------------------------------------------------------------------------------------------
func TablesReferenced(FuncName string, cfgTag string, TablesRefed []string, hdlr *TabServer2Type) {

	/*
		for _, v := range TablesRefed {
			rr.RedisDo("SADD", "trace:set:Table", v)
		}
		rr.RedisDo("SADD", "trace:set:Func", FuncName)
		if cfgTag[0:1] == "/" {
			rr.RedisDo("SADD", "trace:set:URI", cfgTag)
		}
	*/

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		// logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		// xyzzy log stuff
		return
	}

	defer hdlr.gCfg.RedisPool.Put(conn)

	for _, v := range TablesRefed {
		// rr.RedisDo("SADD", "trace:set:Table", v)
		conn.Cmd("SADD", "trace:set:Table", v)
	}
	// rr.RedisDo("SADD", "trace:set:Func", FuncName)
	conn.Cmd("SADD", "trace:set:Func", FuncName)
	if cfgTag[0:1] == "/" {
		// rr.RedisDo("SADD", "trace:set:URI", cfgTag)
		conn.Cmd("SADD", "trace:set:URI", cfgTag)
	}
}

//------------------------------------------------------------------------------------------------
func sIfNil(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
func iIfNil(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}
func jIfNil(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
func bIfNil(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}
func fIfNil(p *float64) float64 {
	if p == nil {
		return 0.0
	}
	return *p
}
func dIfNil(p *time.Time) (z time.Time) {
	if p == nil {
		return
	}
	z = *p
	return
}

//------------------------------------------------------------------------------------------------
// read in the configuration file for the queries / function calls that can be handled by
// RespHandlerSQL
//
// Notes:
//	1. Should add stuff for generic queries
//	2. Figure out how to do an /api/table/<name>?params
//	3. Figure out how to do full CRUD
//------------------------------------------------------------------------------------------------
// test: t-pp1.go
func readInSQLConfig(path string) (jsonData map[string]SQLOne, err error) {

	file, err := sizlib.ReadJSONDataWithComments(path)
	if err != nil {
		fmt.Printf("Error(10014): Error Reading/Opening %v, %s, Config File:%s\n", err, godebug.LF(), path)
		fmt.Fprintf(os.Stderr, "%sError(10014): Error Reading/Opening %v, %s, Config File:%s%s\n", MiscLib.ColorRed, err, godebug.LF(), path, MiscLib.ColorReset)
		return
	} else {
		if db22 {
			fmt.Printf("---- SQL CFG JSON FIlE After Comments Removed, LineNo's etc., %s, Successfully read in ----\n%s\n\n\n", path, file)
		}
	}

	//	file, err := ioutil.ReadFile(path)
	//	if err != nil {
	//		fmt.Printf("Error(10014): Error Reading/Opening %v, %s, Config File:%s\n", err, godebug.LF(), path)
	//		fmt.Fprintf(os.Stderr, "%sError(10014): Error Reading/Opening %v, %s, Config File:%s%s\n", MiscLib.ColorRed, err, godebug.LF(), path, MiscLib.ColorReset)
	//		return
	//	} else {
	//		if db22 {
	//			fmt.Printf("---- SQL CFG JSON FIlE, %s, Successfully read in ----\n%s\n\n\n", path, file)
	//		}
	//	}
	//
	//	data := strings.Replace(string(file), "\t", " ", -1)
	//	lines := strings.Split(data, "\n")
	//	ln := regexp.MustCompile("__LINE__")
	//	fi := regexp.MustCompile("__FILE__")
	//	cm := regexp.MustCompile("//.*$")
	//	for lineNo, aLine := range lines {
	//		aLine = ln.ReplaceAllString(aLine, fmt.Sprintf("%d", lineNo+1))
	//		aLine = fi.ReplaceAllString(aLine, path)
	//		aLine = cm.ReplaceAllString(aLine, "")
	//		lines[lineNo] = aLine
	//	}
	//	file = []byte(strings.Join(lines, "\n"))
	//
	//	if db22 {
	//		fmt.Printf("---- SQL CFG JSON FIlE After Comments Removed, LineNo's etc., %s, Successfully read in ----\n%s\n\n\n", path, file)
	//	}

	if jsonOld {
		err = json.Unmarshal(file, &jsonData)
	} else {
		// https://hjson.org/syntax.html --
		tmp := make(map[string]interface{})
		err = hjson.Unmarshal(file, &tmp)
		if err != nil {
			fmt.Printf("Error(10012): %v, %s, Config File:%s\n", err, godebug.LF(), path)
			fmt.Fprintf(os.Stderr, "%sError(10012): %v, %s, Config File:%s%s\n", MiscLib.ColorRed, err, godebug.LF(), path, MiscLib.ColorReset)
			return
		}
		ftmp := godebug.SVar(tmp)
		err = json.Unmarshal([]byte(ftmp), &jsonData)
	}
	if err != nil {
		fmt.Printf("Error(10012): %v, %s, Config File:%s\n", err, godebug.LF(), path)
		fmt.Fprintf(os.Stderr, "%sError(10012): %v, %s, Config File:%s%s\n", MiscLib.ColorRed, err, godebug.LF(), path, MiscLib.ColorReset)
		return
	} else {
		fmt.Printf("---- SQL CFG JSON FIlE, %s, Successfully parsed ----\n\n", path)
	}

	for i, v := range jsonData {
		v.valid = make(map[string]Validation, len(v.Valid))
		v.validGet = make(map[string]Validation, len(v.Valid))
		v.validPost = make(map[string]Validation, len(v.Valid))
		v.validPut = make(map[string]Validation, len(v.Valid))
		v.validDel = make(map[string]Validation, len(v.Valid))

		for aColName, val := range v.Valid {
			x := Validation{Required: bIfNil(val.Required),
				Type:       sIfNil(val.Type),
				Min_len:    jIfNil(val.Min_len),
				Max_len:    jIfNil(val.Max_len),
				Min:        iIfNil(val.Min),
				Max:        iIfNil(val.Max),
				MinF:       fIfNil(val.MinF),
				MaxF:       fIfNil(val.MaxF),
				MinD:       dIfNil(val.MinD),
				MaxD:       dIfNil(val.MaxD),
				Default:    sIfNil(val.Default),
				UrlEncode:  bIfNil(val.UrlEncode),
				ReMatch:    sIfNil(val.ReMatch),
				ChkType:    bIfNil(val.ChkType),
				eRequired:  (val.Required != nil),
				eType:      (val.Type != nil),
				eMin_len:   (val.Min_len != nil),
				eMax_len:   (val.Max_len != nil),
				eMin:       (val.Min != nil),
				eMax:       (val.Max != nil),
				eMinF:      (val.MinF != nil),
				eMaxF:      (val.MaxF != nil),
				eMinD:      (val.MinD != nil),
				eMaxD:      (val.MaxD != nil),
				eDefault:   (val.Default != nil),
				eUrlEncode: (val.UrlEncode != nil),
				eReMatch:   (val.ReMatch != nil),
				eChkType:   (val.ChkType != nil),
			}
			x.eRequired = true
			v.valid[aColName] = x
		}

		for aColName, val := range v.ValidGet {
			x := Validation{Required: bIfNil(val.Required),
				Type:       sIfNil(val.Type),
				Min_len:    jIfNil(val.Min_len),
				Max_len:    jIfNil(val.Max_len),
				Min:        iIfNil(val.Min),
				Max:        iIfNil(val.Max),
				MinF:       fIfNil(val.MinF),
				MaxF:       fIfNil(val.MaxF),
				MinD:       dIfNil(val.MinD),
				MaxD:       dIfNil(val.MaxD),
				Default:    sIfNil(val.Default),
				UrlEncode:  bIfNil(val.UrlEncode),
				ReMatch:    sIfNil(val.ReMatch),
				ChkType:    bIfNil(val.ChkType),
				eRequired:  (val.Required != nil),
				eType:      (val.Type != nil),
				eMin_len:   (val.Min_len != nil),
				eMax_len:   (val.Max_len != nil),
				eMin:       (val.Min != nil),
				eMax:       (val.Max != nil),
				eMinF:      (val.MinF != nil),
				eMaxF:      (val.MaxF != nil),
				eMinD:      (val.MinD != nil),
				eMaxD:      (val.MaxD != nil),
				eDefault:   (val.Default != nil),
				eUrlEncode: (val.UrlEncode != nil),
				eReMatch:   (val.ReMatch != nil),
				eChkType:   (val.ChkType != nil),
			}
			x.eRequired = true
			v.validGet[aColName] = x
		}

		for aColName, val := range v.ValidPost {
			x := Validation{Required: bIfNil(val.Required),
				Type:       sIfNil(val.Type),
				Min_len:    jIfNil(val.Min_len),
				Max_len:    jIfNil(val.Max_len),
				Min:        iIfNil(val.Min),
				Max:        iIfNil(val.Max),
				MinF:       fIfNil(val.MinF),
				MaxF:       fIfNil(val.MaxF),
				MinD:       dIfNil(val.MinD),
				MaxD:       dIfNil(val.MaxD),
				Default:    sIfNil(val.Default),
				UrlEncode:  bIfNil(val.UrlEncode),
				ReMatch:    sIfNil(val.ReMatch),
				ChkType:    bIfNil(val.ChkType),
				eRequired:  (val.Required != nil),
				eType:      (val.Type != nil),
				eMin_len:   (val.Min_len != nil),
				eMax_len:   (val.Max_len != nil),
				eMin:       (val.Min != nil),
				eMax:       (val.Max != nil),
				eMinF:      (val.MinF != nil),
				eMaxF:      (val.MaxF != nil),
				eMinD:      (val.MinD != nil),
				eMaxD:      (val.MaxD != nil),
				eDefault:   (val.Default != nil),
				eUrlEncode: (val.UrlEncode != nil),
				eReMatch:   (val.ReMatch != nil),
				eChkType:   (val.ChkType != nil),
			}
			x.eRequired = true
			v.validPost[aColName] = x
		}

		for aColName, val := range v.ValidPut {
			x := Validation{Required: bIfNil(val.Required),
				Type:       sIfNil(val.Type),
				Min_len:    jIfNil(val.Min_len),
				Max_len:    jIfNil(val.Max_len),
				Min:        iIfNil(val.Min),
				Max:        iIfNil(val.Max),
				MinF:       fIfNil(val.MinF),
				MaxF:       fIfNil(val.MaxF),
				MinD:       dIfNil(val.MinD),
				MaxD:       dIfNil(val.MaxD),
				Default:    sIfNil(val.Default),
				UrlEncode:  bIfNil(val.UrlEncode),
				ReMatch:    sIfNil(val.ReMatch),
				ChkType:    bIfNil(val.ChkType),
				eRequired:  (val.Required != nil),
				eType:      (val.Type != nil),
				eMin_len:   (val.Min_len != nil),
				eMax_len:   (val.Max_len != nil),
				eMin:       (val.Min != nil),
				eMax:       (val.Max != nil),
				eMinF:      (val.MinF != nil),
				eMaxF:      (val.MaxF != nil),
				eMinD:      (val.MinD != nil),
				eMaxD:      (val.MaxD != nil),
				eDefault:   (val.Default != nil),
				eUrlEncode: (val.UrlEncode != nil),
				eReMatch:   (val.ReMatch != nil),
				eChkType:   (val.ChkType != nil),
			}
			x.eRequired = true
			v.validPut[aColName] = x
		}

		for aColName, val := range v.ValidDel {
			x := Validation{Required: bIfNil(val.Required),
				Type:       sIfNil(val.Type),
				Min_len:    jIfNil(val.Min_len),
				Max_len:    jIfNil(val.Max_len),
				Min:        iIfNil(val.Min),
				Max:        iIfNil(val.Max),
				MinF:       fIfNil(val.MinF),
				MaxF:       fIfNil(val.MaxF),
				MinD:       dIfNil(val.MinD),
				MaxD:       dIfNil(val.MaxD),
				Default:    sIfNil(val.Default),
				UrlEncode:  bIfNil(val.UrlEncode),
				ReMatch:    sIfNil(val.ReMatch),
				ChkType:    bIfNil(val.ChkType),
				eRequired:  (val.Required != nil),
				eType:      (val.Type != nil),
				eMin_len:   (val.Min_len != nil),
				eMax_len:   (val.Max_len != nil),
				eMin:       (val.Min != nil),
				eMax:       (val.Max != nil),
				eMinF:      (val.MinF != nil),
				eMaxF:      (val.MaxF != nil),
				eMinD:      (val.MinD != nil),
				eMaxD:      (val.MaxD != nil),
				eDefault:   (val.Default != nil),
				eUrlEncode: (val.UrlEncode != nil),
				eReMatch:   (val.ReMatch != nil),
				eChkType:   (val.ChkType != nil),
			}
			x.eRequired = true
			v.validDel[aColName] = x
		}

		jsonData[i] = v
	}

	for i, v := range jsonData {
		if v.CacheIt == "row" {
			jsonData[v.TableName] = SQLOne{CacheIt: "row", FromKey: i, Crud: []string{}}
		} else if v.CacheIt == "table" {
			jsonData[v.TableName] = SQLOne{CacheIt: "table", FromKey: i, Crud: []string{}}
		}
	}

	// fmt.Printf ( "%s\n", sizlib.SVarI( jsonData[ "/api/observedPage" ] ) )
	// fmt.Printf ( "Results: %s\n", sizlib.SVarI(jsonData) )
	// os.Exit(1)

	return
}

//------------------------------------------------------------------------------------------------
var replaceStringRes map[string]*regexp.Regexp //
var replaceStringsResMutex sync.RWMutex        //
func init() {
	replaceStringRes = make(map[string]*regexp.Regexp)
}

func ReplaceString(s string, pat string, repl string) (rv string) {
	var re *regexp.Regexp
	var ok bool
	replaceStringsResMutex.RLock()
	re, ok = replaceStringRes[pat]
	replaceStringsResMutex.RUnlock()
	if !ok {
		re = regexp.MustCompile(pat)
		replaceStringsResMutex.Lock()
		replaceStringRes[pat] = re
		replaceStringsResMutex.Unlock()
	}
	rv = re.ReplaceAllLiteralString(s, repl)
	return
}

func InjectDataPs(ps *goftlmux.Params, h SQLOne, res http.ResponseWriter, req *http.Request) {
	https := ""
	// fmt.Printf ( "In Inject Data\n" )
	for i, _ := range h.valid {
		switch i {
		case "$IP$", "$ip$":
			forward := req.Header.Get("X-Forwarded-For")
			if forward != "" {
				goftlmux.AddValueToParams("$ip$", forward, 'i', goftlmux.FromInject, ps)
			} else {
				h, _, err := net.SplitHostPort(req.RemoteAddr)
				if err == nil {
					goftlmux.AddValueToParams("$ip$", h, 'i', goftlmux.FromInject, ps)
				} else {
					goftlmux.AddValueToParams("$ip$", "0.0.0.0", 'i', goftlmux.FromInject, ps)
				}
			}
			//x := goftlmux.LastIndexOfChar(req.RemoteAddr, ':')
			//if x >= 0 {
			//	goftlmux.AddValueToParams("$ip$", req.RemoteAddr[0:x], 'i', goftlmux.FromInject, ps)
			//} else {
			//	goftlmux.AddValueToParams("$ip$", "0.0.0.0", 'i', goftlmux.FromInject, ps)
			//}
		case "$url$":
			if req.TLS != nil {
				https = "https://"
			} else {
				https = "http://"
			}
			v := https + req.Host
			goftlmux.AddValueToParams("$url$", v, 'i', goftlmux.FromInject, ps)
		case "$ua$": // User Agent
			{
				// fmt.Printf ( "Found $ua$ - inject\n" )
				v := req.UserAgent()
				ua := user_agent.New(v)
				goftlmux.AddValueToParams("$ua$", v, 'i', goftlmux.FromInject, ps)

				// xyzzy
				family, ua_version := ua.Browser()
				ua_arr := strings.Split(ua_version+".0.0.0.0", ".")
				goftlmux.AddValueToParams("$ua_family$", family, 'i', goftlmux.FromInject, ps)
				goftlmux.AddValueToParams("$ua_major$", ua_arr[0], 'i', goftlmux.FromInject, ps)
				goftlmux.AddValueToParams("$ua_minor$", ua_arr[1], 'i', goftlmux.FromInject, ps)
				goftlmux.AddValueToParams("$ua_patch$", ua_arr[2], 'i', goftlmux.FromInject, ps)

				if ua.Mobile() {
					goftlmux.AddValueToParams("$is_mobile$", "y", 'i', goftlmux.FromInject, ps)
				} else {
					goftlmux.AddValueToParams("$is_mobile$", "n", 'i', goftlmux.FromInject, ps)
				}

				if ua.Bot() {
					goftlmux.AddValueToParams("$is_bot$", "y", 'i', goftlmux.FromInject, ps)
				} else {
					goftlmux.AddValueToParams("$is_bot$", "y", 'i', goftlmux.FromInject, ps)
				}

				pl := ua.Platform()
				os := ua.OS()
				// fmt.Printf ( "OS:->%s<- Platform:->%s<-\n", os, pl )
				os_arr := make([]string, 0, 10)
				// m["$os$"] = make([]string, 1)
				// m["$os$"][0] = pl
				// goftlmux.AddValueToParams("$os$", os, 'i', goftlmux.FromInject, ps)
				// fr["$os$"] = "inject-$ua$"
				switch pl {
				case "X11":
					goftlmux.AddValueToParams("$os$", os, 'i', goftlmux.FromInject, ps)
					os_arr = strings.Split(os+" 0 0 0 0", " ")
					os_arr[0] = ps.ByNameDflt("os_family", "Linux")
				case "Macintosh":
					goftlmux.AddValueToParams("$os$", "Mac_OS_X", 'i', goftlmux.FromInject, ps)
					os_arr = strings.Split(os+" 0 0 0 0", " ")
					if strings.HasPrefix(family, "FireFox") {
						os_arr = strings.Split("Mac_OS_X."+ReplaceString(os, ".*Mac OS X  *", "")+".0.0.0.0", "_")
					} else if strings.HasPrefix(family, "Opera") {
						os_arr = strings.Split("Mac_OS_X."+ReplaceString(os, ".*Opera-?", "")+".0.0.0.0", "_")
					} else if strings.HasPrefix(family, "Chrome") {
						os_arr = strings.Split("Mac_OS_X."+ReplaceString(os, ".*Mac OS X *", "")+".0.0.0.0", "_")
					} else if strings.HasPrefix(family, "Safari") {
						os_arr = strings.Split("Mac_OS_X."+ReplaceString(os, ".*Mac OS X *", "")+".0.0.0.0", "_")
					}
					os_arr[0] = ps.ByNameDflt("os_family", "Mac_OS_X")
				case "Windows":
					goftlmux.AddValueToParams("$os$", os, 'i', goftlmux.FromInject, ps)
					os_arr = strings.Split(os+" 0 0 0 0", " ")
					os_arr = os_arr[1:]
					os_arr[0] = ps.ByNameDflt("os_family", "Windows")
				case "iPhone":
					fallthrough
				case "iPod":
					fallthrough
				case "iPad":
					os_arr[1] = "iOS"
					goftlmux.AddValueToParams("$os$", os, 'i', goftlmux.FromInject, ps)
				case "Linux":
					os_arr = strings.Split(os+" 0 0 0 0", " ")
					os_arr = os_arr[1:]
					os_arr = strings.Split(os_arr[1]+".0.0.0.0", ".")
					goftlmux.AddValueToParams("$os$", "Android", 'i', goftlmux.FromInject, ps)
					os_arr[0] = ps.ByNameDflt("os_family", "Android")
				default:
					os_arr = strings.Split(os+" 0 0 0 0", " ")
					goftlmux.AddValueToParams("$os$", os, 'i', goftlmux.FromInject, ps)
				}
				os_arr[1] = ps.ByNameDflt("osMajor", "0")
				os_arr[2] = ps.ByNameDflt("osMinor", "0")
				os_arr[3] = ps.ByNameDflt("osPatch", "0")

				goftlmux.AddValueToParams("$os_family$", os_arr[0], 'i', goftlmux.FromInject, ps)
				goftlmux.AddValueToParams("$os_major$", os_arr[1], 'i', goftlmux.FromInject, ps)
				goftlmux.AddValueToParams("$os_minor$", os_arr[2], 'i', goftlmux.FromInject, ps)
				goftlmux.AddValueToParams("$os_patch$", os_arr[3], 'i', goftlmux.FromInject, ps)

			}
		case "$port$":
			x := goftlmux.LastIndexOfChar(req.RemoteAddr, ':')
			if x >= 0 {
				goftlmux.AddValueToParams("$port$", req.RemoteAddr[x+1:], 'i', goftlmux.FromInject, ps)
			} else {
				goftlmux.AddValueToParams("$port$", "", 'i', goftlmux.FromInject, ps)
			}
		case "$host$":
			goftlmux.AddValueToParams("$host$", req.Host, 'i', goftlmux.FromInject, ps)
		case "$protocal$":
			if req.TLS != nil {
				goftlmux.AddValueToParams("$protocal$", "https", 'i', goftlmux.FromInject, ps)
			} else {
				goftlmux.AddValueToParams("$protocal$", "http", 'i', goftlmux.FromInject, ps)
			}
		case "$method$":
			goftlmux.AddValueToParams("$method$", req.Method, 'i', goftlmux.FromInject, ps)
		}
		/*
			// 	 -- Let's think about htis - need to pull the "top" from the set of "top" for file-server? -- use top_hdlr to find file server and ask it?
			// 	 -- have file_server set this as some global data?  Don't know what to do.
			case "$top$":
				{
					s := "/index.html" // This is a mistake!!! - need to figure this out -
					// http://localhost:8200/tree4/confirm-email.html?store=1#!/noRoute
					// http://localhost:8200/tree4/?store=1#!/confirm-email.html
					// http://localhost:8200/tree4/?store=%{$customer_id$%}#!/confirm-email.html
					if t, ok := GlobalCfg["web_top"]; ok {
						s = t
						mdata := make(map[string]string)
						ps.MakeStringMap(mdata)
						mdata["customer_id"] = mdata["$customer_id$"]
						fmt.Printf("before %s\n", s)
						s = sizlib.Qt(s, mdata)
						fmt.Printf("after  %s\n", s)
					}
					goftlmux.AddValueToParams("$top$", s, 'i', goftlmux.FromInject, ps)
				}
		*/
		// xyzzy100 - add in other INJECT
		// RemoteAddr
		// RequestURI
		// Path
		// QueryString
		// Fragment
	}
}

//------------------------------------------------------------------------------------------------
// xyzzy - should check method for GET/POST etc.  See: func ValidateQueryParams ( m url.Values, h SQLOne, req *http.Request ) ( err error ) {
//------------------------------------------------------------------------------------------------
func isRequired(h SQLOne, name string) bool {
	v, ok := h.valid[name]
	if !ok {
		return false
	}
	return v.Required
}

func HasKeys(v map[string]Validation) bool {
	for _, _ = range v {
		return true
	}
	return false
}

//------------------------------------------------------------------------------------------------
// 3. Add in other types - and checks
// 4. Add in RegEx for match
// 5. Add in "email", "ip", etc.
//------------------------------------------------------------------------------------------------
func ValidateQueryParams(ps *goftlmux.Params, h SQLOne, req *http.Request) (err error) {
	err = nil
	var vv *map[string]Validation // mm22mm
	vv = &h.valid
	switch req.Method {
	case "GET":
		if HasKeys(h.validGet) {
			vv = &h.validGet
		}
	case "POST":
		if HasKeys(h.validPost) {
			vv = &h.validPost
		}
	case "PUT":
		if HasKeys(h.validPut) {
			vv = &h.validPut
		}
	case "DELETE":
		if HasKeys(h.validDel) {
			vv = &h.validDel
		}
	default:
		err = errors.New(fmt.Sprintf("Error(14022): Interal error - should never reach this code, %s", godebug.LF()))
		return
	}
	for i, v := range *vv {

		// d, ok := m[i]
		ok := ps.HasName(i)
		d := ps.ByName(i)
		if !ok {
			if v.Required {
				fmt.Printf("At: %s, checking [%s] -->>%s<<--\n", i, godebug.LF(), v.Default)
				fmt.Fprintf(os.Stderr, "%sError (00000): Required field [%s] missing.%s\n", MiscLib.ColorRed, i, MiscLib.ColorReset)
				fmt.Printf("Error (00000): Required field [%s] missing.\n", i)
				d = v.Default
				goftlmux.AddValueToParams(i, d, 'i', goftlmux.FromDefault, ps)
				err = errors.New("Error(10000): Missing Parameter:" + i)
				return
			} else {
				ok = true
				d = v.Default
				goftlmux.AddValueToParams(i, d, 'i', goftlmux.FromDefault, ps)
				fmt.Printf("At: %s, checking [%s] -->>%s<<--\n", i, godebug.LF(), v.Default)
			}
		} else {

			if v.Required {
				if !ok {
					fmt.Printf("At: %s, checking [%s]\n", i, godebug.LF())
					err = errors.New("Error(10080): Missing Parameter:" + i)
					return
				}
				if len(d) <= 0 {
					fmt.Printf("At: %s, checking [%s]\n", i, godebug.LF())
					err = errors.New("Error(10081): Missing Parameter:" + i)
					return
				}
			}

			switch v.Type {

			case "uuid":
				fallthrough
			case "u":
				fallthrough
			case "string":
				fallthrough
			case "s":
				fmt.Printf("At: %s, checking [%s]\n", i, godebug.LF())
				if v.eMin_len {
					fmt.Printf("At: %s, checking [%s]\n", i, godebug.LF())
					if len(d) < v.Min_len && v.Required {
						fmt.Printf("At: %s, checking [%s]\n", i, godebug.LF())
						err = errors.New(fmt.Sprintf("Error(10082): Parameter (%s) Too Short:%s Minimum Length %d", d, i, v.Min_len))
						return
					}
				}
				if v.eMax_len {
					fmt.Printf("At: %s, checking [%s]\n", i, godebug.LF())
					if len(d) > v.Max_len && v.Required {
						fmt.Printf("At: %s, checking [%s]\n", i, godebug.LF())
						err = errors.New(fmt.Sprintf("Error(10083): Parameter Too Long:%s Maximum Length %d", i, v.Max_len))
						return
					}
				}
				if v.eReMatch {
					matched, err2 := regexp.MatchString(v.ReMatch, d)
					if err2 != nil {
						err = errors.New(fmt.Sprintf("Error(10084): Rgular expression in valiation invalid - error %s", err2))
						return
					}
					if !matched && v.Required {
						err = errors.New(fmt.Sprintf("Error(10085): Parameter failed to match regular expression:%s", i))
						return
					}
				}

			case "int":
				fallthrough
			case "i":
				var w int64
				if v.eMin || v.eMax || v.ChkType {
					if !v.Required && len(d) == 0 {
						d = "0"
					}
					w, err = strconv.ParseInt(d, 10, 64)
					if err != nil {
						err = errors.New(fmt.Sprintf("Error(10086): Invalid integer - failed to parse at %s, %s, value=%s", i, err, d))
						return
					}
				}
				if v.eMin {
					if w < v.Min && v.Required {
						// fmt.Printf ( " -- failed min --\n" );
						err = errors.New(fmt.Sprintf("Error(10087): Parameter Too Short:%s Minimum %d", i, v.Min))
						return
					}
				}
				if v.eMax {
					if w > v.Max && v.Required {
						// fmt.Printf ( " -- failed max --\n" );
						err = errors.New(fmt.Sprintf("Error(10088): Parameter Too Large:%s Maximum %d", i, v.Max))
						return
					}
				}

			case "float":
				fallthrough
			case "f":
				var w float64
				if v.eMinF || v.eMaxF || v.ChkType {
					if !v.Required && len(d) == 0 {
						d = "0.0"
					}
					w, err = strconv.ParseFloat(d, 64)
					if err != nil {
						err = errors.New(fmt.Sprintf("Error(10089): Invalid number - failed to parse at %s, %s, value=%s", i, err, d))
						return
					}
				}
				if v.eMinF {
					if w < v.MinF && v.Required {
						// fmt.Printf ( " -- failed min --\n" );
						err = errors.New(fmt.Sprintf("Error(10090): Parameter Too Small:%s Minimum %f", i, v.MinF))
						return
					}
				}
				if v.eMaxF {
					if w > v.MaxF && v.Required {
						// fmt.Printf ( " -- failed max --\n" );
						err = errors.New(fmt.Sprintf("Error(10091): Parameter Too Large:%s Maximum %f", i, v.MaxF))
						return
					}
				}

			case "d":
				fallthrough
			case "t":
				fallthrough
			case "e":
				var w time.Time
				if v.eMinD || v.eMaxD || v.ChkType {
					if !v.Required && len(d) == 0 {
						err = nil
					} else {
						// w, err = time.Parse( ISO8601, d)
						nullOk := true // Checked by "required above
						w, _, err = ms.FuzzyDateTimeParse(d, nullOk)
					}
					if err != nil {
						err = errors.New(fmt.Sprintf("Error(10092): Invalid date/time - failed to parse at %s, %s, value=%s", i, err, d))
						return
					}
				}
				if v.eMinD {
					if w.Before(v.MinD) && v.Required {
						// fmt.Printf ( " -- failed min date --\n" );
						err = errors.New(fmt.Sprintf("Error(10093): Parameter too far in the past:%s Minimum %v", i, v.MinD))
						return
					}
				}
				if v.eMaxD {
					if w.After(v.MaxD) && v.Required {
						// fmt.Printf ( " -- failed max date --\n" );
						err = errors.New(fmt.Sprintf("Error(10094): Parameter too far in the future:%s Maximum %v", i, v.MaxD))
						return
					}
				}

			}
			if v.eUrlEncode && v.UrlEncode {
				d = url.QueryEscape(d)
			}
			ps.SetValue(i, d) // can only change an existing value in 'ps'
		}
	}
	err = nil
	return
}

// ==============================================================================================================================================================================
// ==============================================================================================================================================================================
// ------------------------------------------------------------------- Image Ops ------------------------------------------------------------------------------------
// (1,2) select * from "img_group" where "user_id" = '979ecbf4-7647-41bd-d895-b77d461ec9f4'
// func GetImgSetId ( db *sql.DB, mdata map[string]string ) string {
func projectData(data []map[string]interface{}, param ...string) []map[string]interface{} {
	//var finalResult   []map[string]interface{}

	finalResult := make([]map[string]interface{}, 0, len(data))
	for j := range data {
		x := make(map[string]interface{})
		for _, v := range param {
			x[v] = data[j][v]
		}
		finalResult = append(finalResult, x)
	}
	return finalResult
}

// func ifNull ( data []map[string]interface{}, to string, fr string ) ( []map[string]interface{} ) {
func ifNull(data []map[string]interface{}, to string, fr string) {
	for j := range data {
		if data[j][to] == nil || data[j][to] == "" {
			data[j][to] = data[j][fr]
		}
	}
	// return data
}

func addData(data map[string]interface{}, newData []map[string]interface{}, param ...string) map[string]interface{} {

	path := param[0:1]
	vars := param[1:]

	data[path[0]] = projectData(newData, vars...)

	return data
}

func setHasChildrenTrue(data map[string]interface{}, newData []map[string]interface{}, param ...string) map[string]interface{} {
	path := param[0:1]
	data[path[0]] = true
	return data
}

// ===============================================================================================================================================================================================
func GetUrlValue(m url.Values, name string, dflt string) string {
	if _, ok := m[name]; ok {
		if len(m[name]) > 0 {
			return m[name][0]
		}
	}
	return dflt
}

// ===============================================================================================================================================================================================
func GetUrlValueErr(m url.Values, name string, dflt string) (string, error) {
	if _, ok := m[name]; ok {
		if len(m[name]) > 0 {
			return m[name][0], nil
		}
	}
	return dflt, errors.New("no value found")
}

// ============================================================================================================================================================================

func RedisReferenced(FuncName string, cfgTag string, hdlr *TabServer2Type) {
	/*
		rr.RedisDo("SADD", "trace:set:Func", FuncName)
		if cfgTag[0:1] == "/" {
			rr.RedisDo("SADD", "trace:set:URI", cfgTag)
		}
	*/

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		// logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		// xyzzy log stuff
		return
	}

	defer hdlr.gCfg.RedisPool.Put(conn)

	// rr.RedisDo("SADD", "trace:set:Func", FuncName)
	conn.Cmd("SADD", "trace:set:Func", FuncName)
	if cfgTag[0:1] == "/" {
		// rr.RedisDo("SADD", "trace:set:URI", cfgTag)
		conn.Cmd("SADD", "trace:set:URI", cfgTag)
	}
}

const ISO8601 = "2006-01-02T15:04:05.99999Z07:00"

var Hostname string = ""

func init() {
	Hostname, _ = os.Hostname()
}

// ============================================================================================================================================================================
// Log errors to Redis - with the unique error ID code
/*
-- sample error --
	rv = fmt.Sprintf(`{ "status":"error", "msg":"Error(10034): Invalid Where Clause; %s", %s }`, sizlib.EscapeError(err), godebug.LFj())
	trx.AddNote(2, rv)
	trx.SetQryDone(rv, "")
	io.WriteString(res, sizlib.JsonP(rv, res, req))
*/
func LogError(rv string, id string, status int, msg string, code string, details string, res http.ResponseWriter, req *http.Request, ps goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) {

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		// logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		// xyzzy log stuff
		return
	}

	defer hdlr.gCfg.RedisPool.Put(conn)

	t := time.Now()
	h := t.Hour()
	m := t.Minute()
	s := t.Second()
	TimeUTC := t.UTC()
	w := TimeUTC.Format(ISO8601) // "02/Jan/2006 03:04:05") // 450+ us to do a time format and 1 alloc
	rv2 := fmt.Sprintf(`{"id":%q,"rv":%s,"status":%q,"msg":%q,"code":%q,"details":%q,"req.Method":%q,"req.RemoteAddr":%q,"req.RequestURI":%q,"req.Host":%q,"req.Header.UserAgent":%q,"When":%q,"Hostname":%q}`,
		id, rv, status, msg, code, details, req.Method, req.RemoteAddr, req.RequestURI, req.Host, req.Header.Get("User-Agent"), w, Hostname)
	key := fmt.Sprintf("trace:err:%02d:%02d:%02d", h, m, s)
	/*
		rr.RedisDo("SADD", key, rv2)
		rr.RedisDo("EXPIRE", key, 2*60*60) // Validate for 2 hour
	*/
	// rr.RedisDo("SADD", key, rv2)
	conn.Cmd("SADD", key, rv2)
	// rr.RedisDo("EXPIRE", key, 2*60*60) // Validate for 2 hour
	conn.Cmd("EXPIRE", key, 2*60*60) // Validate for 2 hour
}

// ============================================================================================================================================================================
/*
Example of return information from error.
	HTTP/1.1 401 Unauthorized
	{
		"status": "Error"
		, "msg": "No access token provided."
		, "code": "10002"
		, "details": "bla bla bla"
	}
*/
//
//	Error		Meaning                 	Usage in this code
//	-----		-------------------------- 	------------
//	400			Bad Request					x
//	401			Unauthorized				Invalid token or un/pw is invalid during login
//	402			Payment Required			x
//	403			Forbidden					Accessing an invalid resource in a table - or an invalid table - login or token required
//	404			Not Found					No such file, no such table
//	405			Method Not Allowed			x
//	406			Not Acceptable				Invalid parameters - see body for details
//	412			Precondition Failed			x
//	417			Expectation Failed			x
//	428			Precondition Required		x
//
//	500			Internal Server Error		x
//	501
//
/*
	http.
		StatusBadRequest                   = 400
		StatusUnauthorized                 = 401
		StatusPaymentRequired              = 402
		StatusForbidden                    = 403
		StatusNotFound                     = 404
		StatusMethodNotAllowed             = 405
		StatusNotAcceptable                = 406
		StatusProxyAuthRequired            = 407
		StatusRequestTimeout               = 408
		StatusConflict                     = 409
		StatusGone                         = 410
		StatusLengthRequired               = 411
		StatusPreconditionFailed           = 412
		StatusRequestEntityTooLarge        = 413
		StatusRequestURITooLong            = 414
		StatusUnsupportedMediaType         = 415
		StatusRequestedRangeNotSatisfiable = 416
		StatusExpectationFailed            = 417
		StatusTeapot                       = 418

		StatusInternalServerError     = 500
		StatusNotImplemented          = 501
		StatusBadGateway              = 502
		StatusServiceUnavailable      = 503
		StatusGatewayTimeout          = 504
		StatusHTTPVersionNotSupported = 505

*/
func ReturnErrorMessage(status int, msg string, code string, details string, res http.ResponseWriter, req *http.Request, ps goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) {
	if status == 200 {
		return
	}
	if hdlr.SendStatusOnError {
		res.WriteHeader(status)
	} else {
		res.WriteHeader(200)
	}
	res.Header().Set("Content-Type", "application/json")
	id0, _ := uuid.NewV4()
	id := id0.String()
	// rv := JsonP_2(fmt.Sprintf(`{"status":"error","req.status":%d,"code":%q,"msg":%q,"details":%q,"unique_error_id":%q}`, status, code, msg, details, id), res, req, ps, trx)
	rv := fmt.Sprintf(`{"status":"error","req.status":%d,"code":%q,"msg":%q,"details":%q,"unique_error_id":%q}`, status, code, msg, details, id)
	trx.AddNote(2, "Error return, log id="+id)
	trx.AddNote(3, rv)
	trx.SetQryDone(rv, "")
	io.WriteString(res, rv)
	LogError(rv, id, status, msg, code, details, res, req, ps, trx, hdlr)
}

func ReturnErrorMessageRv(status int, rv string, msg string, code string, details string, res http.ResponseWriter, req *http.Request, ps goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) {
	if status == 200 {
		return
	}
	if hdlr.SendStatusOnError {
		res.WriteHeader(status)
	} else {
		res.WriteHeader(200)
	}
	res.Header().Set("Content-Type", "application/json")
	id0, _ := uuid.NewV4()
	id := id0.String()
	// rv := JsonP_2(fmt.Sprintf(`{"status":"error","req.status":%d,"code":%q,"msg":%q,"details":%q,"unique_error_id":%q}`, status, code, msg, details, id), res, req, ps, trx)
	trx.AddNote(2, "Error return, log id="+id)
	trx.AddNote(3, rv)
	trx.SetQryDone(rv, "")
	io.WriteString(res, rv)
	LogError(rv, id, status, msg, code, details, res, req, ps, trx, hdlr)
}

// xyzzy - remove - midldware takes care of this
//func JsonP_2(s string, res http.ResponseWriter, req *http.Request, ps goftlmux.Params, trx *tr.Trx) string {
//	// fmt.Printf ( "JsonP_2 passed ->%s<-\n", s )
//	callback := ps.ByName("callback")
//	if callback != "" {
//		res.WriteHeader(200)
//		res.Header().Set("Content-Type", "application/javascript") // For JSONP
//		return fmt.Sprintf("%s(%s);", callback, s)
//	} else {
//		// return sizlib.JSON_Prefix + s
//		return s
//	}
//}

// xyzzy - remove - midldware takes care of this
func JsonP_3(data map[string]interface{}, res http.ResponseWriter, req *http.Request, ps goftlmux.Params, trx *tr.Trx) string {
	// fmt.Printf ( "JsonP_2 passed ->%s<-\n", s )
	format := ps.ByName("format")
	var s string
	switch format {
	default:
		fallthrough
	case "json", "JSON":
		sB, err := json.Marshal(data)
		_ = err
		s = string(sB)
		callback := ps.ByName("callback")
		if callback != "" {
			res.WriteHeader(200)
			res.Header().Set("Content-Type", "application/javascript") // For JSONP
			return fmt.Sprintf("%s(%s);", callback, s)
		} else {
			// return sizlib.JSON_Prefix + s
			return s
		}
	case "xml", "XML":
		res.Header().Set("Content-Type", "text/xml")
		sB, err := xml.Marshal(data)
		_ = err
		s = string(sB)
		return s
	case "html", "HTML":
		// xyzzy - templates and template data
		res.Header().Set("Content-Type", "text/html")
		return s
	case "text", "TEXT":
		// xyzzy - templates and template data
		res.Header().Set("Content-Type", "text/plain")
		return s
	case "pdf", "PDF":
		res.Header().Set("Content-Type", "application/pdf") // For JSONP
		return s
	}
}

// ============================================================================================================================================================================
// Custom GoGoWidgets for handling tracing
// ============================================================================================================================================================================

/*

// Moved to mid - top of middeware

func GetTrx(www http.ResponseWriter) (ptr *tr.Trx) {
	if rw, ok := www.(*goftlmux.MidBuffer); ok {
		//if rw.G_Trx == nil {
		//	ptr = tr.NewTrx()
		//	rw.G_Trx = ptr
		//	return ptr
		//}
		if ptr, ok = rw.G_Trx.(*tr.Trx); ok {
			return ptr
		}
	}
	panic(fmt.Sprintf("Should have has a *goftlmux.MidBuffger - got passed a %T\n", www))
}
*/

func InitTrx(w *goftlmux.MidBuffer, req *http.Request, ps *goftlmux.Params) int {
	// tx := mid.GetTrx(w)
	// tx.DepricatedInitTrx()
	return 0
}

func EndTrx(w *goftlmux.MidBuffer, req *http.Request, ps *goftlmux.Params) int {
	ip := req.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	// fmt.Fprintf(fo, ApacheFormatPattern, ip, timeFormatted, req.Method, req.RequestURI, req.Proto, w.status,
	// 	w.responseBytes, elapsedTime.Seconds())
	finishTime := time.Now()
	// finishTimeUTC := finishTime.UTC()
	elapsedTime := finishTime.Sub(w.StartTime)

	tx := mid.GetTrx(w)
	tx.UriSaveData(ip, w.StartTime, req.Method, req.RequestURI, req.Proto, w.StatusCode, int64(w.Length), elapsedTime, req)
	tx.TraceUriRawEnd(req, elapsedTime) // -- send pubsub --
	return 0
}

func minI(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxI(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// xyzzy need hdlr
//func PickInsertUpdateColumns(www http.ResponseWriter, theFile0 map[string]interface{}, table_key string) (rv map[string]interface{}) {
func (hdlr *TabServer2Type) DbEnabledOn(www http.ResponseWriter, s string) bool {
	// if db_func["PickInsertUpdateColumns"] {
	return false
}

// func (this *Trx) SetDataPs(ps *goftlmux.Params) {
func SetDataPs(trx *tr.Trx, ps *goftlmux.Params) {
	if ps != nil {
		for i := 0; i < ps.NParam; i++ {
			nm := ps.Data[i].Name
			vl := ps.Data[i].Value
			ff := goftlmux.FromTypeToString(ps.Data[i].From)
			haveIt := false
			for _, ww := range trx.Data {
				if ww.Name == nm {
					haveIt = true
				}
			}
			if !haveIt {
				trx.Data = append(trx.Data, common.NameValueFrom{Name: nm, Value: vl, From: ff})
			}
		}
	}
}

// trx.TraceUriPs(req, ps) // xyzzyBoom11111
// ----------------------------------------------------------------------------------------------------------
// ps is convered ps.DumpParams() from the Params package // ps *goftlmux.Params
// func (this *Trx) TraceUriPs(req *http.Request, ps string) {
func TraceUriPs(trx *tr.Trx, req *http.Request, ps *goftlmux.Params) {
	trx.TraceUriPs(req, "")
}

// ============================================================================================================================================================================

type SqlCfgLoaded struct {
	FileName string
	ErrorMsg string
}

var SqlCfgFilesLoaded []SqlCfgLoaded

const db22 = false // dump out  processing of JSON files for configuration

/* vim: set noai ts=4 sw=4: */
