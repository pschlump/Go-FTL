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

package SessionRedis

import (
	"fmt"
	"net/http"
	"os"
	"time"

	JsonX "github.com/pschlump/JSONx"

	"github.com/pschlump/Go-FTL/server/RedisSessionData"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/radix.v2/redis"
	"github.com/pschlump/uuid"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &SessionRedis{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // Xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("SessionRedis", CreateEmpty, `{
		"Paths":       	 		{ "type":["string","filepath"], "isarray":true, "required":true },
		"RedisSessionPrefix":  	{ "type":[ "string" ], "required":false, "default":"session:" },
		"CookieName":  	        { "type":[ "string" ], "required":false, "default":"X-Go-FTL-Sesion-Id" },
		"LineNo":       		{ "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *SessionRedis) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	gCfg.ConnectToRedis()
	hdlr.gCfg = gCfg
	return
}

func (hdlr *SessionRedis) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*SessionRedis)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type SessionRedis struct {
	Next               http.Handler                //
	Paths              []string                    //	-- Path that will pull out non-auth session info
	RedisSessionPrefix string                      //	-- Session storage path in Redis
	CookieName         string                      // Name of the cookie that is used to lookup the session
	LineNo             int                         //
	gCfg               *cfg.ServerGlobalConfigType //
}

// var cookieName = "X-Go-FTL-Sesion-Id"

func NewSessionServer(n http.Handler, p, q []string, redis_prefix, realm string) *SessionRedis {
	return &SessionRedis{
		Next:               n,
		Paths:              p,
		RedisSessionPrefix: redis_prefix,
	}
}

func (hdlr *SessionRedis) ServeHTTP(www http.ResponseWriter, req *http.Request) {

	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "SessionRedis", hdlr.Paths, pn, req.URL.Path)

			id, ses_id_found := hdlr.GetTrxCookie(req)

			if !ses_id_found {
				id0, _ := uuid.NewV4()
				id = id0.String()
			}

			ps := &rw.Ps
			var session string

			if rw.Session == nil || !ses_id_found {
				rw.Session = RedisSessionData.NewRedisSesionDataType().SetPrefixKey(hdlr.RedisSessionPrefix, id).
					// xyzzy - add in Redis connection stuff so can access redis.
					// conn, err := hdlr.gCfg.RedisPool.Get()
					// defer hdlr.gCfg.RedisPool.Put(conn)
					// func (sd *RedisSessionDataType) GetFreeRedisConn(GetConn func() (*redis.Client, error), PutConn *redis.Client) {
					GetFreeRedisConn(func() (conn *redis.Client, err error) {
						conn, err = hdlr.gCfg.RedisPool.Get()
						return
					}, func(conn *redis.Client) {
						hdlr.gCfg.RedisPool.Put(conn)
					})
			}

			// xyzzy - pass in 'ps' so can access parameters - may need closure

			// get session, + raw -> $session$
			if !ses_id_found {
				session = rw.Session.CreateDefaultSession() // create and save
			} else {
				session = rw.Session.GetSessionFromRedis() //get by id from redis, if non - then create and save.
			}

			if false {
				rw.Session.SetData("regular", "email_addr", "kermit@the-green-pc.com")
				rw.Session.SetRule("email_addr", false, true)
			}

			x := rw.Session.DumpData()
			if hdlr.gCfg.DbOn("*", "SessionRedis", "db1") {
				fmt.Fprintf(os.Stderr, "%sBefore Session Data = %s\n%s%s\n", MiscLib.ColorYellow, x, godebug.LF(), MiscLib.ColorReset)
				fmt.Fprintf(os.Stdout, "%sBefore Session Data = %s\n%s%s\n", MiscLib.ColorYellow, x, godebug.LF(), MiscLib.ColorReset)
			}

			goftlmux.AddValueToParams("$session$", session, 'i', goftlmux.FromInject, ps)

			// -----------------------------------------------------------------------------------------------------------------------------------
			hdlr.Next.ServeHTTP(www, req)

			x = rw.Session.DumpData()
			if hdlr.gCfg.DbOn("*", "SessionRedis", "db1") {
				fmt.Fprintf(os.Stderr, "%sAfter Session Data = %s\n%s%s\n", MiscLib.ColorCyan, x, godebug.LF(), MiscLib.ColorReset)
				fmt.Fprintf(os.Stdout, "%sAfter Session Data = %s\n%s%s\n", MiscLib.ColorCyan, x, godebug.LF(), MiscLib.ColorReset)
			}

			if rw.Session.IsDirty() || !ses_id_found {
				if hdlr.gCfg.DbOn("*", "SessionRedis", "db1") {
					fmt.Fprintf(os.Stderr, "%sAfter Session -- is dirty will flush to redis%s%s\n", MiscLib.ColorCyan, godebug.LF(), MiscLib.ColorReset)
					fmt.Fprintf(os.Stdout, "%sAfter Session -- is dirty will flush to redi%s%s\n", MiscLib.ColorCyan, godebug.LF(), MiscLib.ColorReset)
				}
				rw.Session.FlushSessionToRedis()
			}

			if !ses_id_found {
				expire := time.Now().AddDate(10, 0, 2) // good for 10 year 2 days
				secureCookie := false
				if req.TLS != nil {
					secureCookie = true
				}
				cookie := http.Cookie{Name: hdlr.CookieName, Value: id, Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400 * 366 * 10, Secure: secureCookie, HttpOnly: true}
				http.SetCookie(www, &cookie)
			}

		}
		return
	}

	hdlr.Next.ServeHTTP(www, req)
}

func (hdlr *SessionRedis) GetTrxCookie(req *http.Request) (id string, ses_id_found bool) {
	Ck := req.Cookies()
	for _, v := range Ck {
		if v.Name == hdlr.CookieName {
			ses_id_found = true
			id = v.Value
			break
		}
	}
	if hdlr.gCfg.DbOn("*", "SessionRedis", "db1") {
		fmt.Printf("cookie=%s id=[%s]\n", hdlr.CookieName, id)
	}

	if !ses_id_found {
		id0, _ := uuid.NewV4()
		id = id0.String()
	}
	return
}

/* vim: set noai ts=4 sw=4: */
