//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1228
//

//
// Package implements HTTP basic auth with the authorization stored in Redis.
//
// Basic auth should only be used in conjunction with TLS (https).  If you need to use
// an authentication scheme with http or you want a better authentication scheme
// take a look at the auth_srp.go  middleware.  There are examples of using it with
// jQuery and AngularJS 1.3 (2.0 of AngularJS coming soon).   Also take a look at the
// auth_oauth20 middleware (also coming soon) if you need a 3rd party auth.
//
// Also this is "basic auth" with the ugly browser pop-up of username/password and no
// real error reporting to the user.  If you want something better switch to the SRP
// solution.
//
// Pbkdf2 is used to help prevent cracking via rainbow tables.
//
// So what is "basic" auth really good for?  Simple answer.  If you need just a
// touch of security - and no more.   Example:  You took a video of your children
// that is and you want to send it to Grandma.  It is too big for her email so
// you need to send a link - so quick copy it up to your server and set basic
// auth on the directory.  Send her the link and the username and password.
// This keeps googlebot and other nosy folks out of it - but it is not really
// secure.  Then a couple of days later you delete the video.   Works like a
// champ!
//
//

package BasicAuthRedis

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	JsonX "github.com/pschlump/JSONx"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/godebug"
	"github.com/pschlump/uuid"
	"golang.org/x/crypto/pbkdf2" // https://github.com/golang/crypto/blob/master/pbkdf2/pbkdf2.go
)

// --------------------------------------------------------------------------------------------------------------------------

const NIterations = 5000

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &BasicAuthRedis{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("BasicAuthRedis", CreateEmpty, `{
		"Paths":        	 { "type":["string","filepath"], "isarray":true, "required":true },
		"Realm":        	 { "type":[ "string" ], "required":true },
		"RedisPrefix":  	 { "type":[ "string" ], "required":false, "default":"BasicAuth" },
		"HashUsername":  	 { "type":[ "bool" ], "required":false, "default":"false" },
		"HashUsernameSalt":  { "type":[ "string" ], "required":false, "default":"8H3QhT9uHElh+c5NfowHx1gLeDw6qBMSTLvoL87GcB4FwflM8v2cTs" },
		"LineNo":       	 { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *BasicAuthRedis) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	gCfg.ConnectToRedis()
	hdlr.gCfg = gCfg
	return
}

func (hdlr *BasicAuthRedis) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*BasicAuthRedis)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type BasicAuthRedis struct {
	Next             http.Handler                //
	Paths            []string                    //
	Realm            string                      //
	RedisPrefix      string                      //
	HashUsername     bool                        // If true then the email address is hashed before retreval
	HashUsernameSalt string                      // Optional - must be application wide salt - not the same as per-user salt that is generated.
	LineNo           int                         //
	gCfg             *cfg.ServerGlobalConfigType //
}

var loaded bool = false

func NewBasicAuthServer(n http.Handler, p []string, redis_prefix, realm string) *BasicAuthRedis {
	return &BasicAuthRedis{
		Next:        n,
		Paths:       p,
		Realm:       realm,
		RedisPrefix: redis_prefix,
	}
}

func (hdlr *BasicAuthRedis) ServeHTTP(www http.ResponseWriter, req *http.Request) {

	// hdlr.gCfg.ConnectToRedis()

	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "BasicAuthRedis", hdlr.Paths, pn, req.URL.Path)

			auth := req.Header.Get("Authorization")
			if auth == "" {
				www.Header().Set("WWW-Authenticate", "Basic realm=\""+hdlr.Realm+"\"")
				http.Error(www, "Not Authorized", http.StatusUnauthorized)
				return
			}
			userPassword, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
			// fmt.Printf("At:%s userPassword >%s<\n", lib.LF(), userPassword)
			if err == nil {
				parts := strings.SplitN(string(userPassword), ":", 2)
				if len(parts) == 2 {
					un := parts[0]
					pw := parts[1]
					// fmt.Printf("At:%s un:pw >%s:%s<\n", lib.LF(), un, pw)
					// fmt.Printf("At:%s database: %+v\n", lib.LF(), hdlr.passwords)

					if user_id, ok := hdlr.redisGetUser(un, pw, rw); ok {
						goftlmux.AddValueToParams("$username$", un, 'i', goftlmux.FromAuth, &rw.Ps)
						if user_id != "" {
							goftlmux.AddValueToParams("$user_id$", user_id, 'i', goftlmux.FromAuth, &rw.Ps)
						}
						id0, _ := uuid.NewV4()
						f_auth_token := id0.String()
						goftlmux.AddValueToParams("auth_token", f_auth_token, 'i', goftlmux.FromInject, &rw.Ps)
						goftlmux.AddValueToParams("$is_logged_in$", "y", 'i', goftlmux.FromAuth, &rw.Ps)
						hdlr.Next.ServeHTTP(www, req)
						return
					}
				}
			}
		}
		http.Error(www, "Not Authorized", http.StatusUnauthorized)
		return
	}

	hdlr.Next.ServeHTTP(www, req)
}

func (hdlr *BasicAuthRedis) redisGetUser(un, pw string, rw *goftlmux.MidBuffer) (user_id string, ok bool) {
	var key string

	if db4 {
		fmt.Printf("redisGetUser: %s - un >%s< pw >%s<\n", godebug.LF(), un, pw)
	}

	// store password as `{"salt":"salt","pwh":"hash"}`
	if hdlr.HashUsername {
		em := fmt.Sprintf("%x", pbkdf2.Key([]byte(hdlr.Realm+":"+un), []byte(hdlr.HashUsernameSalt), NIterations, 64, sha256.New))
		key = hdlr.RedisPrefix + ":" + em
	} else {
		key = hdlr.RedisPrefix + ":" + hdlr.Realm + ":" + un
	}

	if db4 {
		fmt.Printf("redisGetUser: %s key= [%s]\n", godebug.LF(), key)
	}

	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return "", false
	}

	v, err := conn.Cmd("GET", key).Str()
	if err != nil {
		if db4 {
			fmt.Printf("Error on redis - user not found - invalid relm - bad prefix - get(%s): %s\n", key, err)
		}
		return
	}

	if db4 {
		fmt.Printf("redisGetUser: %s returned from Redis= [%s]\n", godebug.LF(), v)
	}

	t := strings.Split(v, ":")
	if len(t) < 2 {
		fmt.Printf("Error on redis - invalid data, len=%d\n", len(t))
		return
	}
	salt, pwh := t[0], t[1]
	dk := fmt.Sprintf("%x", pbkdf2.Key([]byte(pw), []byte(salt), NIterations, 64, sha256.New))
	if len(t) > 2 {
		user_id = t[2]
	}

	if db10 {
		fmt.Printf("salt [%s], pwh[%s], dk[%s]\n", salt, pwh, dk)
	}

	if subtle.ConstantTimeCompare([]byte(dk), []byte(pwh)) == 1 {
		if db10 {
			fmt.Printf("At: %s --------------- should be authoraized -- \n", lib.LF())
		}
		ok = true
	}

	return
}

/*
	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return false
	}

	v, err := conn.Cmd("GET", key).Str()

	hdlr.gCfg.RedisPool.Put(conn)
*/
const db4 = false
const db10 = false

/* vim: set noai ts=4 sw=4: */
