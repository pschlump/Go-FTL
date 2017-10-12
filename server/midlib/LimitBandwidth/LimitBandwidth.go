//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1262
//

package LimitBandwidth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	JsonX "github.com/pschlump/JSONx"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------

// set Key value
// pexpire key Miliseconds

//

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &LimitBandwidthType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("LimitBandwidth", CreateEmpty, `{
		"Paths":        	{ "type":["string","filepath"], "isarray":true, "required":true },
		"FreqMiliKey":		{ "type":["string"], "default":"lim-ban:%{auth_token%}" }
		"HttpErrorCode":	{ "type":["int"], "default":"429" }
		"FreqMili":			{ "type":["int"], "default":"1500" }
		"PerAuthCfgKey":	{ "type":["string"], "default":"lim-ban-cfg:%{auth_token%}" }
		"NPerSecKey":		{ "type":["string"], "default":"lim-ban-nps:%{auth_token%}" }
		"NPerSecond":		{ "type":["int"], "default":"0" }
		"AuthTokenName":	{ "type":["string"], "default":"auth_token" }
		"LineNo":       	{ "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *LimitBandwidthType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	if hdlr.FreqMili > 0 && hdlr.NPerSecond > 0 {
		fmt.Fprintf(os.Stderr, "Error: LineNo:%d %sConfiguration - LimitBandwidth - can not have both the FreqMili (%d) and the NPerSecond (%d) larger than zero\n"+
			"Assuming NPerSecond is configured and ignoring FreqMili%s\n", hdlr.LineNo, MiscLib.ColorRed, hdlr.FreqMili, hdlr.NPerSecond, MiscLib.ColorReset)
		fmt.Fprintf(os.Stdout, "Error: LineNo:%d %sConfiguration - LimitBandwidth - can not have both the FreqMili (%d) and the NPerSecond (%d) larger than zero\n"+
			"Assuming NPerSecond is configured and ignoring FreqMili%s\n", hdlr.LineNo, MiscLib.ColorRed, hdlr.FreqMili, hdlr.NPerSecond, MiscLib.ColorReset)
		hdlr.FreqMili = -1
	}
	gCfg.ConnectToRedis()
	// gCfg.ConnectToPostgreSQL()
	hdlr.gCfg = gCfg
	return
}

func (hdlr *LimitBandwidthType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*LimitBandwidthType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type LimitBandwidthType struct {
	Next          http.Handler                //
	Paths         []string                    //
	FreqMiliKey   string                      // QT template for lookup of Redis key
	HttpErrorCode int                         //
	FreqMili      int                         // sequence for request in milliseconds
	PerAuthCfgKey string                      // QT template for using a non-default timeout, can be set to "inf" - for no limit on time.
	NPerSecKey    string                      // QT template for the N Per Second
	NPerSecond    int                         // if 0 then will not apply
	AuthTokenName string                      // The name of the "auth_token" could be "$user_id$" for logged in users
	LineNo        int                         //
	gCfg          *cfg.ServerGlobalConfigType //
}

// PerAuthCfgKey is JSON data
//	{ "FreqMili": -1, 			// unlimited
//	{ "FreqMili": 0, 			// turned off
//	{ "FreqMili": 1500, 		// 1.5  seconds between requests
//	{ "FreqMili":0,"NPerSecond": -1		// unlimited
//	{ "FreqMili":0 "NPerSecond": 0			// Check not applicable
// 	{ "FreqMili":0 "NPerSecond": 200		// Allow up to 200 request per second.
type PerAuthCfg struct {
	FreqMili   int
	NPerSecond int
}

func NewLimitBandwidthServer(n http.Handler, p []string, pFreqMili, pNPerSecond int) *LimitBandwidthType {
	return &LimitBandwidthType{
		Next:          n,
		Paths:         p,
		FreqMili:      pFreqMili,
		NPerSecond:    pNPerSecond,
		FreqMiliKey:   "lim-ban:%{auth_token%}",
		HttpErrorCode: 429,
		PerAuthCfgKey: "lim-ban-cfg:%{auth_token%}",
		NPerSecKey:    "lim-ban-nps:%{auth_token%}",
		AuthTokenName: "auth_token",
	}
}

func (hdlr *LimitBandwidthType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "LimitBandwidth", hdlr.Paths, pn, req.URL.Path)

			ps := &rw.Ps

			// Get Redis Config - if "auth_token" set.
			cfg := PerAuthCfg{FreqMili: hdlr.FreqMili, NPerSecond: hdlr.NPerSecond}
			token := ps.ByNameDflt(hdlr.AuthTokenName, "")
			if token != "" {
				mdata := map[string]string{"auth_token": token}
				perAutTokenCfg := sizlib.Qt(hdlr.PerAuthCfgKey, mdata)
				freqMiliKey := sizlib.Qt(hdlr.FreqMiliKey, mdata)
				nPerSecKey := sizlib.Qt(hdlr.NPerSecKey, mdata)
				vv, err := hdlr.RedisGet(perAutTokenCfg)
				if err != nil {
					if cfg.FreqMili <= 0 && cfg.NPerSecond <= 0 {
						hdlr.Next.ServeHTTP(www, req)
					} else {
						// FreqMili > 0 and ...
						hdlr.RedisSetFreqMili(freqMiliKey, cfg.FreqMili, "{}")
					}
				} else {
					err = json.Unmarshal([]byte(vv), &cfg)
					if err != nil {
						fmt.Fprintf(os.Stderr, "AT: %s, Invalid configuration for this auth_token=%s, err=%s, data=%s assume unlimited bandwidth\n", godebug.LF(), token, err, vv)
						fmt.Fprintf(os.Stdout, "AT: %s, Invalid configuration for this auth_token=%s, err=%s, data=%s assume unlimited bandwidth\n", godebug.LF(), token, err, vv)
						hdlr.Next.ServeHTTP(www, req)
					}
					if cfg.FreqMili > 0 {
						val, err := hdlr.RedisGet(freqMiliKey)
						if err != nil || val != "" {
							hdlr.RedisSetFreqMili(freqMiliKey, cfg.FreqMili, "{}") // start timeout on this now, for next request.
							hdlr.Next.ServeHTTP(www, req)
						} else {
							hdlr.RedisSetFreqMili(freqMiliKey, cfg.FreqMili, "{}") // start timeout on this now, for next request.
						}
					} else if cfg.NPerSecond > 0 {
						nserved, err := hdlr.RedisSetIncrementNPerSecond(nPerSecKey) // Increment count served.
						if err != nil {
							fmt.Fprintf(os.Stderr, "AT: %s, unable to set NPerSecond in Redis - auth_token=%s, err=%s, data=%s assume unlimited bandwidth\n", godebug.LF(), token, err, vv)
							fmt.Fprintf(os.Stdout, "AT: %s, unable to set NPerSecond in Redis - auth_token=%s, err=%s, data=%s assume unlimited bandwidth\n", godebug.LF(), token, err, vv)
							hdlr.Next.ServeHTTP(www, req)
						} else if nserved > cfg.NPerSecond {
							www.WriteHeader(hdlr.HttpErrorCode)
							return
						} else {
							hdlr.Next.ServeHTTP(www, req)
						}
					}
				}
			} else {
				mdata := map[string]string{"auth_token": "--n/a--"}
				freqMiliKey := sizlib.Qt(hdlr.FreqMiliKey, mdata)
				nPerSecKey := sizlib.Qt(hdlr.NPerSecKey, mdata)
				if db1 {
					fmt.Printf("AT: %s cfg=%s\n", godebug.LF(), godebug.SVar(cfg))
				}
				if cfg.FreqMili > 0 {
					val, err := hdlr.RedisGet(freqMiliKey)
					if err != nil || val != "{}" {
						if db1 {
							fmt.Printf("AT: %s, err=%s val=%s\n", godebug.LF(), err, val)
						}
						hdlr.RedisSetFreqMili(freqMiliKey, cfg.FreqMili, "{}") // start timeout on this now, for next request.
						hdlr.Next.ServeHTTP(www, req)
					} else {
						if db1 {
							fmt.Printf("AT: %s\n", godebug.LF())
						}
						www.WriteHeader(hdlr.HttpErrorCode)
					}
				} else if cfg.NPerSecond > 0 {
					nserved, err := hdlr.RedisSetIncrementNPerSecond(nPerSecKey) // Increment count served.
					if err != nil {
						fmt.Fprintf(os.Stderr, "AT: %s, unable to set NPerSecond in Redis - auth_token=%s, err=%s, assume unlimited bandwidth\n", godebug.LF(), token, err)
						fmt.Fprintf(os.Stdout, "AT: %s, unable to set NPerSecond in Redis - auth_token=%s, err=%s, assume unlimited bandwidth\n", godebug.LF(), token, err)
						hdlr.Next.ServeHTTP(www, req)
					} else if nserved > cfg.NPerSecond {
						www.WriteHeader(hdlr.HttpErrorCode)
					} else {
						hdlr.Next.ServeHTTP(www, req)
					}
				}
				return
			}

			// do Redis Get, if found then error, set timeout to FreqMili (setex)
			// if not found then create with setex using FreqMiliKey and FreqMili

			hdlr.Next.ServeHTTP(www, req)

		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		hdlr.Next.ServeHTTP(www, req)
	}

}

func (hdlr *LimitBandwidthType) RedisGet(key string) (val string, err error) {
	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	val, err = conn.Cmd("GET", key).Str()
	if err != nil {
		if db1 {
			fmt.Printf("%sRedis key=%s err=%s, %s%s\n", MiscLib.ColorRed, key, err, godebug.LF(), MiscLib.ColorReset)
		}
		return
	}
	return
}

func (hdlr *LimitBandwidthType) RedisSetFreqMili(key string, ttlMilisec int, val string) (err error) {
	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	conn.Cmd("SET", key, val)
	if err != nil {
		if db1 {
			fmt.Printf("%sRedis key=%s err=%s, %s%s\n", MiscLib.ColorRed, key, err, godebug.LF(), MiscLib.ColorReset)
		}
		return
	}

	err = conn.Cmd("PEXPIRE", key, ttlMilisec).Err
	return
}

func (hdlr *LimitBandwidthType) RedisSetIncrementNPerSecond(key string) (nSoFar int, err error) {
	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	ms, err := conn.Cmd("PTTL", key).Int()
	if err != nil || ms <= -1 {
		ms = 1000
	}

	cur, err := conn.Cmd("INCR", key).Int()
	if err != nil {
		conn.Cmd("SETEX", key, 1, "1")
		nSoFar = 1
	} else {
		nSoFar = cur
		err = conn.Cmd("PEXPIRE", key, ms).Err
	}

	return
}

const db1 = false

/* vim: set noai ts=4 sw=4: */
