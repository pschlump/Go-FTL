//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1266
//

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
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {

	// normally identical
	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
		pCfg, ok := ppCfg.(*LoginRequiredType)
		if ok {
			pCfg.SetNext(next)
			rv = pCfg
		} else {
			err = mid.FtlConfigError
			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
		}
		gCfg.ConnectToRedis()
		pCfg.gCfg = gCfg
		return
	}

	postInit := func(h interface{}, cfgData map[string]interface{}, callNo int) error {

		hh, ok := h.(*LoginRequiredType)
		if !ok {
			// logrus.Warn(fmt.Sprintf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo))
			fmt.Printf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo)
			return mid.ErrInternalError
		} else {
			hh.strongLoginReq, _ = lib.ParseBool(hh.StrongLoginReq)
		}

		return nil
	}

	// normally identical
	createEmptyType := func() interface{} { return &LoginRequiredType{} }

	cfg.RegInitItem2("LoginRequired", initNext, createEmptyType, postInit, `{
		"Paths":            { "type":["string","filepath"], "isarray":true, "required":true },
		"StrongLoginReq":   { "type":["string"], "default":"no" },
		"LineNo":           { "type":[ "int" ], "default":"1" }
		}`)
}

// normally identical
func (hdlr *LoginRequiredType) SetNext(next http.Handler) {
	hdlr.Next = next
}

var _ mid.GoFTLMiddleWare = (*LoginRequiredType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type LoginRequiredType struct {
	Next           http.Handler
	Paths          []string
	StrongLoginReq string
	LineNo         int
	strongLoginReq bool
	gCfg           *cfg.ServerGlobalConfigType //
}

func NewLoginRequiredServer(n http.Handler, p []string) *LoginRequiredType {
	return &LoginRequiredType{Next: n, Paths: p}
}

func (hdlr LoginRequiredType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "LoginRequired", hdlr.Paths, pn, req.URL.Path)

			ps := rw.Ps
			is_logged_in := ps.ByNameDflt("$is_logged_in$", "")
			is_full_login := ps.ByNameDflt("$is_full_login$", "")
			if is_logged_in == "y" {
				if hdlr.StrongLoginReq == "yes" {
					if is_full_login == "y" {
						hdlr.Next.ServeHTTP(www, req)
						return
					} else {
						www.WriteHeader(http.StatusForbidden)
					}
				} else {
					hdlr.Next.ServeHTTP(www, req)
					return
				}
			} else {
				www.WriteHeader(http.StatusForbidden)
			}

			// ip := lib.GetIP(req)
			// cookieValue := lib.GetCookie("LoginAuthCookie", req)
			// cookieHash := lib.GetCookie("LoginHashCookie", req)

			// xyzzy - pantopick at this pont - if system thas changed then fail.

			// if xip, _, _, hash, err := hdlr.GetCookieAuth(cookieValue, rw); err == nil && xip == ip && cookieHash == hash {
			// 	// fmt.Printf("   Serve it\n")
			// 	hdlr.Next.ServeHTTP(www, req)
			// 	return
			// } else {
			// 	// fmt.Printf("   *** Reject *** it\n")
			// 	www.WriteHeader(http.StatusForbidden)
			// }

		} else {
			www.WriteHeader(http.StatusForbidden)
		}
	} else {
		www.WriteHeader(http.StatusNotFound)
	}

}

/* OLD Documentation:

This implements a cookie-pair based authentication.  Cookie1 is a value, Cookie2 is a hash of a
secret value stored in Redis.  If Cookie1 matches RedisOf(Hash(Cookie2)) and the IP is the same
the the person is logged in.  Cookies are set to expire after 2 days.

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

/* vim: set noai ts=4 sw=4: */
