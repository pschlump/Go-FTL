//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2018.
//

package OriginToCustomerId

import (
	"fmt"
	"net/http"
	"os"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	JsonX "github.com/pschlump/JSONx"
	"github.com/pschlump/godebug"
	logrus "github.com/pschlump/pslog" // "github.com/sirupsen/logrus"
)

// --------------------------------------------------------------------------------------------------------------------------
/*
192.168.0.199:6379> keys htci:*
 1) "htci:www.qr-today.com"
 2) "htci:localhost:9018"
 3) "htci:192.168.0.157:9018"
 4) "htci:localhost:16040"
 5) "htci:lonetree-ranch.beefchain.com"
 6) "htci:t2.test1.com"
 7) "htci:--default--"
 8) "htci:localhost:9019"
 9) "htci:www.go-ftl.com"
10) "htci:t1.test1.com"
11) "htci:192.168.0.200:9018"
12) "htci:127.0.0.1:9019"

192.168.0.199:6379> get htci:localhost:9018
"1"

*/

// --------------------------------------------------------------------------------------------------------------------------

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &OriginToCustomerIdType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("OriginToCustomerId", CreateEmpty, `{
		"Paths":        	 { "type":["string","filepath"], "isarray":true, "required":true },
		"RedisPrefix":  	 { "type":[ "string" ], "required":false, "default":"htci:" },
		"LineNo":       	 { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *OriginToCustomerIdType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	gCfg.ConnectToRedis()
	hdlr.gCfg = gCfg
	return
}

func (hdlr *OriginToCustomerIdType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*OriginToCustomerIdType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type OriginToCustomerIdType struct {
	Next        http.Handler                //
	Paths       []string                    //
	RedisPrefix string                      //
	LineNo      int                         //
	gCfg        *cfg.ServerGlobalConfigType //
}

func NewBasicAuthServer(n http.Handler, p []string, redis_prefix, realm string) *OriginToCustomerIdType {
	return &OriginToCustomerIdType{
		Next:        n,
		Paths:       p,
		RedisPrefix: redis_prefix,
	}
}

func (hdlr *OriginToCustomerIdType) ServeHTTP(www http.ResponseWriter, req *http.Request) {

	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "OriginToCustomerId", hdlr.Paths, pn, req.URL.Path)

			ps := &rw.Ps
			customer_id, Origin := hdlr.redisGetCustomerId(www, rw, req)
			goftlmux.AddValueToParams("$customer_id$", customer_id, 'i', goftlmux.FromInject, ps)
			goftlmux.AddValueToParams("$host$", Origin, 'i', goftlmux.FromInject, ps)

		}
	}

	hdlr.Next.ServeHTTP(www, req)
}

func (hdlr *OriginToCustomerIdType) redisGetCustomerId(www http.ResponseWriter, rw *goftlmux.MidBuffer, req *http.Request) (customer_id, Origin string) {

	Origin = req.Header.Get("Origin")
	key := hdlr.RedisPrefix + Origin

	if db4 {
		fmt.Printf("OriginToCustomerID: key= [%s], %s\n", key, godebug.LF())
		fmt.Fprintf(os.Stderr, "OriginToCustomerID: key= [%s], %s\n", key, godebug.LF())
	}

	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return "1", Origin
	}

	v, err := conn.Cmd("GET", key).Str()
	if err != nil {
		fmt.Printf("Error on Redis - invalid host (will use default) - get(%s): host[%s] redisPrefix[%s] %s, %s\n", key, Origin, hdlr.RedisPrefix, err, godebug.LF())
		fmt.Printf("    *** This error indicates that you should 1) connect to redis, 2) `set \"%s\" \"1\"\n", key)
		fmt.Printf("    *** Or use the correct customer id insteadof \"1\"\n\n")
		fmt.Fprintf(os.Stderr, "Error on Redis - invalid host (will use default) - get(%s): host[%s] redisPrefix[%s] %s, %s\n", key, Origin, hdlr.RedisPrefix, err, godebug.LF())
		fmt.Fprintf(os.Stderr, "    *** This error indicates that you should 1) connect to redis, 2) `set \"%s\" \"1\"\n", key)
		fmt.Fprintf(os.Stderr, "    *** Or use the correct customer id insteadof \"1\"\n\n")
		// lookup default from Redis at this point.
		key = "htci:--default--"
		v, err = conn.Cmd("GET", key).Str()
		if err != nil {
			fmt.Printf("Error on redis - failed to find 'htci:--default--' - get(%s): host[%s] redisPrefix[%s] %s, %s\n", key, Origin, hdlr.RedisPrefix, err, godebug.LF())
			return "1", Origin
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
