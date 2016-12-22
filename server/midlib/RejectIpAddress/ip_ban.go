//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1282
//

//
// Ban a set of IP addresses
//

package RejectIpAddress

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
		pCfg, ok := ppCfg.(*RejectIPAddressType)
		if ok {
			pCfg.SetNext(next)
			rv = pCfg
		} else {
			err = mid.FtlConfigError
			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
		}
		pCfg.gCfg = gCfg
		return
	}

	// normally identical
	createEmptyType := func() interface{} { return &RejectIPAddressType{} }

	postInitValidation := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
		// fmt.Printf("In postInitValidation, h=%v\n", h)
		hh, ok := h.(*RejectIPAddressType)
		if !ok {
			fmt.Printf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo)
			return mid.ErrInternalError
		} else {
			if hh.RedisPrefix != "" && len(hh.IPAddrs) > 0 {
				fmt.Printf("Error: Can not have both a set of IP Addres and a RedisPrefix at the same time - RejectIPAddress, Line No:%d\n", hh.LineNo)
				return mid.ErrInvalidConfiguration
			}
		}
		return nil
	}

	cfg.RegInitItem2("RejectIPAddress", initNext, createEmptyType, postInitValidation, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"IPAddrs":       { "type":[ "string","ip" ], "isarray":true },
		"RedisPrefix":   { "type":[ "string" ] },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)

	// FIXME - testing with invalid input

}

// normally identical
func (hdlr *RejectIPAddressType) SetNext(next http.Handler) {
	hdlr.Next = next
}

var _ mid.GoFTLMiddleWare = (*RejectIPAddressType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type RejectIPAddressType struct {
	Next        http.Handler
	Paths       []string
	IPAddrs     []string
	RedisPrefix string
	LineNo      int
	gCfg        *cfg.ServerGlobalConfigType //
}

func NewRejectIpServer(n http.Handler, p []string, ips []string, pre string) *RejectIPAddressType {
	return &RejectIPAddressType{Next: n, Paths: p, IPAddrs: ips, RedisPrefix: pre}
}

func (hdlr *RejectIPAddressType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "RejectIPAddress", hdlr.Paths, pn, req.URL.Path)

			p, err := lib.GetIpFromReq(req)
			if err != nil {
				www.WriteHeader(http.StatusInternalServerError)
				return
			}
			if hdlr.RedisPrefix != "" {
				if !hdlr.redisValidIpAddr(p, rw) {
					www.WriteHeader(http.StatusForbidden)
					return
				}
			} else if lib.InArray(p, hdlr.IPAddrs) {
				www.WriteHeader(http.StatusForbidden)
				return
			}
		}
	}
	hdlr.Next.ServeHTTP(www, req)

}

func (hdlr *RejectIPAddressType) redisValidIpAddr(ip string, rw *goftlmux.MidBuffer) (ipIsGood bool) {
	key := hdlr.RedisPrefix + ip

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return false
	}

	v, err := conn.Cmd("GET", key).Str()

	hdlr.gCfg.RedisPool.Put(conn)

	godebug.Printf(db44, "Error on redis - get(%s): %s %s\n", key, v, err)

	ipIsGood = (v == "" || err != nil)

	godebug.Printf(db44, "Return value: isIpGood=%v\n", ipIsGood)

	return
}

const db44 = false

/* vim: set noai ts=4 sw=4: */
