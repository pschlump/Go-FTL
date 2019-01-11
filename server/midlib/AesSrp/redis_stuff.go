//
// Package aessrp implements encrypted authentication and encrypted REST.
// Redis Inteface Stuff.
//
// Copyright (C) Philip Schlump, 2013-2016
//
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 0001
// 你好无聊的世界
//

package AesSrp

import (
	"fmt"
	"strings"

	// "github.com/mediocregopher/radix.v2/redis"
	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/goftlmux" //
	"github.com/pschlump/godebug"                //
	"github.com/pschlump/json"                   //	"encoding/json"
)

// ----------------------------------------------------------------------------------------------------------------------------
// Redis Interface Code
// ----------------------------------------------------------------------------------------------------------------------------

// This didn't work correctly so....
func DbSandboxKey(hdlr *AesSrpType, key string) (rkey string) {
	// return hdlr.SandboxPrefix + key
	return key
}

// ----------------------------------------------------------------------------------------------------------------------------
func DbSetExpire(hdlr *AesSrpType, rw *goftlmux.MidBuffer, key string, val string, life int) (err error) {
	key = DbSandboxKey(hdlr, key)
	/*
		From: https://godoc.org/github.com/mediocregopher/radix.v2/redis
		From: https://godoc.org/github.com/mediocregopher/radix.v2/pool
	*/
	godebug.Db2Printf(db201, "DbSetExpire: %s key [%s] value [%s] life [%d]\n", godebug.LF(), key, val, life)
	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
	}
	defer hdlr.gCfg.RedisPool.Put(conn)

	godebug.Db2Printf(db201, "DbSetExpire: %s\n", godebug.LF())
	err = conn.Cmd("SET", key, val).Err
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s on setting key: %s to value %s","LineFile":%q}`+"\n", err, key, val, godebug.LF()))
		return
	}

	godebug.Db2Printf(db201, "DbSetExpire: %s\n", godebug.LF())
	err = conn.Cmd("EXPIRE", key, fmt.Sprintf("%d", life)).Err
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s on setting key: %s to value %s","LineFile":%q}`+"\n", err, key, val, godebug.LF()))
		return
	}

	return
}

// ----------------------------------------------------------------------------------------------------------------------------
func DbSetString(hdlr *AesSrpType, rw *goftlmux.MidBuffer, key string, val string) (err error) {
	key = DbSandboxKey(hdlr, key)
	/*
		From: https://godoc.org/github.com/mediocregopher/radix.v2/redis
		From: https://godoc.org/github.com/mediocregopher/radix.v2/pool
	*/
	godebug.Db2Printf(db201, "DbSetString: %s key [%s] value [%s]\n", godebug.LF(), key, val)
	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
	}
	defer hdlr.gCfg.RedisPool.Put(conn)

	godebug.Db2Printf(db201, "DbSetString: %s\n", godebug.LF())
	err = conn.Cmd("SET", key, val).Err
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s on setting key: %s to value %s","LineFile":%q}`+"\n", err, key, val, godebug.LF()))
		return
	}

	return
}

/*
	logrus.Warn(fmt.Sprintf(`{"type":"email","from":%q,"subject":%q,"body":%q}`+"\n", fr, sub, bod))
*/

func DbGetString(hdlr *AesSrpType, rw *goftlmux.MidBuffer, key string) (rkey string, err error) {
	key = DbSandboxKey(hdlr, key)
	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
	}
	defer hdlr.gCfg.RedisPool.Put(conn)

	rkey, err = conn.Cmd("GET", key).Str()

	return
}

func DbDel(hdlr *AesSrpType, rw *goftlmux.MidBuffer, key string) {
	key = DbSandboxKey(hdlr, key)
	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
	}
	defer hdlr.gCfg.RedisPool.Put(conn)

	conn.Cmd("DEL", key)
}

func DbExpire(hdlr *AesSrpType, rw *goftlmux.MidBuffer, key string, life int) (err error) {
	key = DbSandboxKey(hdlr, key)
	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
	}
	defer hdlr.gCfg.RedisPool.Put(conn)

	err = conn.Cmd("EXPIRE", key, life).Err
	return
}

// ============================================================================================================================================
///////////////////////////////// redis non-sim /////////////////////////////////////////////////////////////////////////
type RSaveToRedis struct {
	Pre       string
	Ttl       uint64
	Ttl_srp_S uint64
	Ttl_srp_V uint64
}

type RSaveToRedisInterface interface {
	RSetValue(hdlr *AesSrpType, rw *goftlmux.MidBuffer, key string, val map[string]string)
	RGetValue(hdlr *AesSrpType, rw *goftlmux.MidBuffer, key string) (val map[string]string, ok bool)
	RGetValueRaw(hdlr *AesSrpType, rw *goftlmux.MidBuffer, key string) (val string, ok bool)
	RUpdValue(hdlr *AesSrpType, rw *goftlmux.MidBuffer, key string, val map[string]string)
}

// Verify that RSaveSata fits to the interface
var _ RSaveToRedisInterface = (*RSaveToRedis)(nil)

func NewRSaveToRedis(pre string) *RSaveToRedis {
	return &RSaveToRedis{
		Pre:       pre,       // srp:U:
		Ttl_srp_S: (60 * 2),  // save for 2 min
		Ttl_srp_V: (60 * 4),  // save for 4 min
		Ttl:       (60 * 30), // save for 30 min (long enough to debug with)
	}
}

func (rs *RSaveToRedis) RSetValue(hdlr *AesSrpType, rw *goftlmux.MidBuffer, key string, val map[string]string) {
	pre := rs.Pre
	if strings.HasPrefix(key, "srp:") {
		pre = ""
	} else {
		panic("OOPSy")
	}
	s := rs.RSerial(val)

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
	}
	defer hdlr.gCfg.RedisPool.Put(conn)

	skey := DbSandboxKey(hdlr, pre+key)
	err = conn.Cmd("SET", skey, s).Err
	if dbSRP_Redis {
		fmt.Printf("Redis key=%s val=%s err=%s\n", skey, s, err)
	}
	if strings.HasPrefix(key, "srp:S:") {
		conn.Cmd("EXPIRE", skey, rs.Ttl_srp_S)
	}
	if strings.HasPrefix(key, "srp:V:") {
		conn.Cmd("EXPIRE", skey, rs.Ttl_srp_V)
	}
}
func (rs *RSaveToRedis) RGetValue(hdlr *AesSrpType, rw *goftlmux.MidBuffer, key string) (val map[string]string, ok bool) {
	var s string
	ok = true
	s, ok = rs.RGetValueRaw(hdlr, rw, key)
	val = rs.RDeSerial(s)
	if dbSRP_Redis {
		fmt.Printf("GET val = %+v\n", val)
	}
	return
}

func (rs *RSaveToRedis) RGetValueRaw(hdlr *AesSrpType, rw *goftlmux.MidBuffer, key string) (val string, ok bool) {
	ok = true
	pre := rs.Pre
	if strings.HasPrefix(key, "srp:") {
		pre = ""
	} else {
		fmt.Printf("Bad Result: key=%s\n", key)
		panic("OOPSy")
	}
	skey := DbSandboxKey(hdlr, pre+key)

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
	}
	defer hdlr.gCfg.RedisPool.Put(conn)

	s, err := conn.Cmd("GET", skey).Str()
	if err != nil {
		if dbSRP_Redis {
			fmt.Printf("On get of %s - error:%s\n", skey, err)
		}
		ok = false
		return
	}
	if dbSRP_Redis {
		fmt.Printf("GET Key: %s Val/Raw:%s\n", skey, s)
	}
	val = s
	return
}
func (rs *RSaveToRedis) RUpdValue(hdlr *AesSrpType, rw *goftlmux.MidBuffer, key string, val map[string]string) {
	pre := rs.Pre
	if strings.HasPrefix(key, "srp:") {
		pre = ""
	} else {
		panic("OOPSy")
	}
	skey := DbSandboxKey(hdlr, pre+key)

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
	}
	defer hdlr.gCfg.RedisPool.Put(conn)

	s, err := conn.Cmd("GET", skey).Str()
	if err != nil {
		rs.RSetValue(hdlr, rw, key, val) // Just set - no existing value
		return
	}
	val2 := rs.RDeSerial(s) // Merge in new data
	for kk, vv := range val {
		val2[kk] = vv
	}
	rs.RSetValue(hdlr, rw, key, val2) // Set it with new data
}
func (rs *RSaveToRedis) RSerial(val map[string]string) (rv string) {
	trv, err := json.Marshal(val)
	if err != nil {
		rv = "{}"
	}
	rv = string(trv)
	return
}
func (rs *RSaveToRedis) RDeSerial(s string) (rv map[string]string) {
	if s == "" {
		rv = make(map[string]string)
		return
	}
	err := json.Unmarshal([]byte(s), &rv)
	if err != nil {
		fmt.Printf(`{"status":"error","code":"0029","msg":"Error(19913): %v", "input_failed":%q}`+"\n", err, s)
		// xyzzyLogrus
		rv = make(map[string]string)
	}
	return
}

const db201 = false
const dbSRP_Redis = false

/* vim: set noai ts=4 sw=4: */
