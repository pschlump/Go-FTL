//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1226
//

//
// Package implements HTTP basic auth with the authorization stored in PostgreSQL.
//
// The PG package used to access the database is:
//  https://github.com/jackc/pgx
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
// This keeps googlebot and other nozzy folks out of it - but it is not really
// secure.  Then a couple of days later you delete the video.   Works like a
// champ!
//
// There is a command line tool in ../../../tools/user-pgsql to maintain the data
// in PostgreSQL -- Also if you want to use a relational database for storage look
// at ../basicpgsql - that is the same kind of a middleware just using PostgreSQL
// for the database
//
// pgx version.

package BasicAuthPgSQL

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"

	JsonX "github.com/pschlump/JSONx"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/uuid"
	"golang.org/x/crypto/pbkdf2" // https://github.com/golang/crypto/blob/master/pbkdf2/pbkdf2.go
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*BasicPgSQLHandlerType)
//		if ok {
//			pCfg.SetNext(next)
//			rv = pCfg
//		} else {
//			err = mid.FtlConfigError
//			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
//		}
//		gCfg.ConnectToPostgreSQL()
//		pCfg.gCfg = gCfg
//		return
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} { return &BasicPgSQLHandlerType{} }
//
//	cfg.RegInitItem2("BasicAuthPgSQL", initNext, createEmptyType, nil, `{
//		"Paths":        	 { "type":["string","filepath"], "isarray":true, "required":true },
//		"Realm":        	 { "type":[ "string" ], "required":true },
//		"HashUsername":  	 { "type":[ "bool" ], "required":false, "default":"false" },
//		"HashUsernameSalt":  { "type":[ "string" ], "required":false, "default":"8H3QhT9uHElh+c5NfowHx1gLeDw6qBMSTLvoL87GcB4FwflM8v2cTs" },
//		"LineNo":       	 { "type":[ "int" ], "default":"1" }
//		}`)
//}
//
//// SetNext normally identical
//func (hdlr *BasicPgSQLHandlerType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &BasicPgSQLHandlerType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("BasicPgSQLHandlerType", CreateEmpty, `{
		"Paths":        	 { "type":["string","filepath"], "isarray":true, "required":true },
		"Realm":        	 { "type":[ "string" ], "required":true },
		"HashUsername":  	 { "type":[ "bool" ], "required":false, "default":"false" },
		"HashUsernameSalt":  { "type":[ "string" ], "required":false, "default":"8H3QhT9uHElh+c5NfowHx1gLeDw6qBMSTLvoL87GcB4FwflM8v2cTs" },
		"LineNo":       	 { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *BasicPgSQLHandlerType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	gCfg.ConnectToPostgreSQL()
	hdlr.gCfg = gCfg
	return
}

func (hdlr *BasicPgSQLHandlerType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*BasicPgSQLHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

// NIterations is the number of iterations that pbkdf2 hasing will use
const NIterations = 5000

// BasicPgSQLHandlerType configuration for this middleare
type BasicPgSQLHandlerType struct {
	Next             http.Handler                //
	Paths            []string                    //
	Realm            string                      //
	HashUsername     bool                        // If true then the email address is hashed before retreval
	HashUsernameSalt string                      // Optional - must be application wide salt - not the same as per-user salt that is generated.
	LineNo           int                         //
	mutex            sync.Mutex                  //
	connected        string                      //
	gCfg             *cfg.ServerGlobalConfigType //
}

var loaded = false

// NewBasicAuthPgSQLServer return a default initialized BasicPgSQLHandlerType -- look in the test code to see if additional initialization was performed.
func NewBasicAuthPgSQLServer(n http.Handler, p []string, realm string) *BasicPgSQLHandlerType {
	return &BasicPgSQLHandlerType{
		Next:  n,
		Paths: p,
		Realm: realm,
	}
}

func (hdlr *BasicPgSQLHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {

	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			// fmt.Printf("%sJust Before (2): %s\n", MiscLib.ColorRed, MiscLib.ColorReset)

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "BasicAuthPgSQL", hdlr.Paths, pn, req.URL.Path)

			// fmt.Printf("%sJust After (2): %s\n", MiscLib.ColorRed, MiscLib.ColorReset)

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

					if userID, ok := hdlr.pgGetUser(un, pw); ok {
						goftlmux.AddValueToParams("$username$", un, 'i', goftlmux.FromAuth, &rw.Ps)
						goftlmux.AddValueToParams("$user_id$", userID, 'i', goftlmux.FromAuth, &rw.Ps)
						id0, _ := uuid.NewV4()
						fAuthToken := id0.String()
						goftlmux.AddValueToParams("auth_token", fAuthToken, 'i', goftlmux.FromInject, &rw.Ps)
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

const db1 = false

/* vim: set noai ts=4 sw=4: */
