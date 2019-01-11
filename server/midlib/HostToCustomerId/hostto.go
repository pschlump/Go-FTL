//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2018.
//

package HostToCustomerId

import (
	"fmt"
	"net/http"
	"os"

	JsonX "github.com/pschlump/JSONx"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &HostToCustomerIdType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("HostToCustomerId", CreateEmpty, `{
		"Paths":        	 { "type":["string","filepath"], "isarray":true, "required":true },
		"RedisPrefix":  	 { "type":[ "string" ], "required":false, "default":"htci:" },
		"LineNo":       	 { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *HostToCustomerIdType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	gCfg.ConnectToRedis()
	hdlr.gCfg = gCfg
	return
}

func (hdlr *HostToCustomerIdType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*HostToCustomerIdType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type HostToCustomerIdType struct {
	Next        http.Handler                //
	Paths       []string                    //
	RedisPrefix string                      //
	LineNo      int                         //
	gCfg        *cfg.ServerGlobalConfigType //
}

func NewBasicAuthServer(n http.Handler, p []string, redis_prefix, realm string) *HostToCustomerIdType {
	return &HostToCustomerIdType{
		Next:        n,
		Paths:       p,
		RedisPrefix: redis_prefix,
	}
}

func (hdlr *HostToCustomerIdType) ServeHTTP(www http.ResponseWriter, req *http.Request) {

	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "HostToCustomerId", hdlr.Paths, pn, req.URL.Path)

			ps := &rw.Ps
			customer_id := hdlr.redisGetCustomerId(www, rw, req)
			goftlmux.AddValueToParams("$customer_id$", customer_id, 'i', goftlmux.FromInject, ps)
			goftlmux.AddValueToParams("$host$", req.Host, 'i', goftlmux.FromInject, ps)

		}
	}

	hdlr.Next.ServeHTTP(www, req)
}

func (hdlr *HostToCustomerIdType) redisGetCustomerId(www http.ResponseWriter, rw *goftlmux.MidBuffer, req *http.Request) (customer_id string) {

	key := hdlr.RedisPrefix + req.Host

	if db4 {
		fmt.Printf("redisGetCustomerId: %s key= [%s], %s\n", godebug.LF(), key, godebug.LF())
		fmt.Fprintf(os.Stderr, "redisGetCustomerId: %s key= [%s], %s\n", godebug.LF(), key, godebug.LF())
	}

	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return "1"
	}

	v, err := conn.Cmd("GET", key).Str()
	if err != nil {
		fmt.Printf("Error on redis - invalid host (will use default) - get(%s): host[%s] redisPrefix[%s] %s, %s\n", key, req.Host, hdlr.RedisPrefix, err, godebug.LF())
		// lookup default from Redis at this point.
		key = "htci:--default--"
		v, err = conn.Cmd("GET", key).Str()
		if err != nil {
			fmt.Printf("Error on redis - failed to find 'htci:--default--' - get(%s): host[%s] redisPrefix[%s] %s, %s\n", key, req.Host, hdlr.RedisPrefix, err, godebug.LF())
			return "1"
		}
	}

	if db4 {
		fmt.Printf("redisGetCustomerId: %s key= [%s], Set customer_id to:%s, %s\n", godebug.LF(), key, v, godebug.LF())
		fmt.Fprintf(os.Stderr, "redisGetCustomerId: %s key= [%s], Set customer_id to:%s, %s\n", godebug.LF(), key, v, godebug.LF())
	}
	customer_id = v
	return

}

const db4 = true

/* vim: set noai ts=4 sw=4: */
