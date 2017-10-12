//
// Go-FTL Redis List Data - Package
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 0000
//
// Description:  A simple database lookup in redis.  It is assumed that you are searching for
// a set of keys that have a common prefix.  The data values are required to be JSON objects.
//
// Limitations:  This returns the entire set at onece.   All the data valeus are searched.
// That's a linear scan of the entire set of keys.  So... that means that this is not
// the fastes.  This is good when you have a small set of keys that match ( like 10 )
// and is NOT suitable when you have 100 or more.
//
// Eache return value is filtered by field based on user role.  If this is unsed without
// login then the role is alwasy 'anon' for the anonomuous user.   The legitimated return
// fields are set in the configuration file.
//
// Example:
//			{ "redisList": { "LineNo": 100,
//				"Paths":             "/api/list/user",
//				"Prefix":            "srp:U:",
//				"UserRoles":         [ "anon,$key$", "user,$key$,confirmed", "admin,$key$,confirmed,disabled" ],
//				...
//			} },
//
// This can be both a pre-login for 'anon' user role and a post login .  The field $key$ is the remainer of the
// key after the prefix is striped off.
//
// Example Output:
//
//		[{"$key$":"jane@example.com"}
// 		, {"$key$":"bob@example.com"}
// 		, {"$key$":"frog@the-green-pc.com"}
// 		, {"$key$":"abc@def.ghi"}]
//
// Unlike a real (PostgreSQL, Oracle, DB2) database there is no where/order by/group by etc.  You just get
// all the data back.
//
// If you need to use a real database interface look at the Tab2 module.  That provices a complete interface
// to PostgreSQL, Oracle, T-SQL (MS-Sql) and other relational databases.
//

// xyzzy - verify logged in at this point
// xyzzyGernalCaseFilter - should be general case filter

package RedisListRaw

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"www.2c-why.com/JsonX"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical - but not this time.
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*RedisListRawHandlerType)
//		if ok {
//			pCfg.SetNext(next)
//			rv = pCfg
//		} else {
//			err = mid.FtlConfigError
//			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
//		}
//		gCfg.ConnectToRedis()
//		pCfg.gCfg = gCfg
//		return
//	}
//
//	postInit := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
//
//		hh, ok := h.(*RedisListRawHandlerType)
//		if !ok {
//			// logrus.Warn(fmt.Sprintf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo))
//			fmt.Printf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo)
//			return mid.ErrInternalError
//		} else {
//			if len(hh.Filter) > 0 {
//				hh.filter = make([]*lib.FilterType, 0, len(hh.filter))
//				for ii, vv := range hh.Filter {
//					ff, err := lib.ParseFilter(vv)
//					if err != nil {
//						fmt.Fprintf(os.Stderr, "%sError: Unable to parse the %d filter: Error: %s Line No:%d%s\n", MiscLib.ColorRed, ii, err, hh.LineNo, MiscLib.ColorReset)
//						fmt.Printf("Error: Unable to parse the %d filter: Error: %s Line No:%d\n", ii, err, hh.LineNo)
//						return mid.ErrInternalError
//					}
//					hh.filter = append(hh.filter, ff)
//				}
//			} else {
//				hh.filter = []*lib.FilterType{}
//			}
//			godebug.Printf(db2, "Filter: end of postInit, hh.filter=%s, %s\n", godebug.SVarI(hh.filter), godebug.LF())
//		}
//
//		return nil
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} { return &RedisListRawHandlerType{} }
//
//	cfg.RegInitItem2("RedisListRaw", initNext, createEmptyType, postInit, `{
//		}`)
//}
//
//// normally identical
//func (hdlr *RedisListRawHandlerType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &RedisListRawHandlerType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("RedisListRaw", CreateEmpty, `{
		"Paths":             { "type":[ "string", "filepath" ], "isarray":true, "required":true },
		"Prefix":            { "type":[ "string" ], "required":true },
		"Filter":            { "type":[ "string" ], "isarray":true },
		"UserRoles":         { "type":[ "string" ], "isarray":true, "required":true },
		"UserRolesReject":   { "type":[ "string" ], "isarray":true },
		"LineNo":            { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *RedisListRawHandlerType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	gCfg.ConnectToRedis()
	hdlr.gCfg = gCfg
	return
}

func (hdlr *RedisListRawHandlerType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	if len(hdlr.Filter) > 0 {
		hdlr.filter = make([]*lib.FilterType, 0, len(hdlr.filter))
		for ii, vv := range hdlr.Filter {
			ff, err := lib.ParseFilter(vv)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%sError: Unable to parse the %d filter: Error: %s Line No:%d%s\n", MiscLib.ColorRed, ii, err, hdlr.LineNo, MiscLib.ColorReset)
				fmt.Printf("Error: Unable to parse the %d filter: Error: %s Line No:%d\n", ii, err, hdlr.LineNo)
				return mid.ErrInternalError
			}
			hdlr.filter = append(hdlr.filter, ff)
		}
	} else {
		hdlr.filter = []*lib.FilterType{}
	}
	godebug.Printf(db2, "Filter: end of postInit, hdlr.filter=%s, %s\n", godebug.SVarI(hdlr.filter), godebug.LF())
	return
}

var _ mid.GoFTLMiddleWare = (*RedisListRawHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------
//
// UserRole example:
//		user,auth,confirmed
//		admin,auth,DeviceId,confirmed
// UserRoleReject example:
//		anon-user, not-logged-in
//

type RedisListRawHandlerType struct {
	Next            http.Handler                //
	Paths           []string                    // Path to respond to
	Prefix          string                      // Redis Prefix of set of keys to return
	Filter          []string                    // Array of value paris, Name(op)Value that is used to filter - probably should be expression // xyzzyGernalCaseFilter //
	UserRoles       []string                    // Set of user roles.  Each role has Username followed by a comma delimited list of projected colums for that user that can be returned.
	UserRolesReject []string                    //	User roles to not return data for - example anon-user
	filter          []*lib.FilterType           //
	gCfg            *cfg.ServerGlobalConfigType // Configuration including connection to redis
	LineNo          int                         //
}

func NewRedisListRawServer(n http.Handler, p []string, prefix string, userRoles []string, rj []string) *RedisListRawHandlerType {
	return &RedisListRawHandlerType{Next: n, Paths: p, Prefix: prefix, UserRoles: userRoles, UserRolesReject: rj}
}

func (hdlr *RedisListRawHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {

	SandBoxPrefix := ""
	GenKey := func() string {
		if SandBoxPrefix != "" {
			return hdlr.Prefix + SandBoxPrefix
		}
		return hdlr.Prefix
	}

	if db1 {
		fmt.Printf("\n")
		fmt.Printf("redisList ------------------------------------------------------------------------------\n")
		fmt.Printf("hdlr.Paths[%s] url = %s, %s\n", hdlr.Paths, req.URL.Path, godebug.LF())
		fmt.Printf("Config: %s\n", lib.SVarI(hdlr))
		fmt.Printf("--------- ------------------------------------------------------------------------------\n")
		fmt.Printf("\n")
	}
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 && len(hdlr.Prefix) > 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "RedisListRaw", hdlr.Paths, pn, req.URL.Path)

			if db1 {
				fmt.Printf("Top of Request: At: %s\n", godebug.LF())
			}

			role := rw.Ps.ByNameDflt("$privs$", "anon")
			if db1 {
				fmt.Printf("role = %s, %s\n", role, godebug.LF())
			}
			SandBoxPrefix = rw.Ps.ByNameDflt("GOFTL_Sandbox", "")
			if db1 {
				fmt.Printf("SandBoxPrefix -->>%s<<--\n", SandBoxPrefix)
			}

			// xyzzy - verify logged in at this point

			// tt := rw.Ps.ByNameDflt("$privs$", "")
			tt := rw.Ps.ByNameDflt("t", "")
			if tt == "" {
				if db1 {
					fmt.Printf("Did not have tt, seting role to 'anon', %s\n", godebug.LF())
				}
				role = "anon"
			}

			var ur []string
			for _, vv := range hdlr.UserRoles {
				t := strings.Split(vv, ",")
				if len(t) > 0 && t[0] == role {
					ur = t
				}
			}

			if db1 {
				fmt.Printf("CRITICAL: ur = %+v, %s\n", ur, godebug.LF())
			}
			www.Header().Set("Content-Type", "application/json")                     // For JSON data
			www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
			www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
			www.Header().Set("Expires", "0")                                         // Proxies.
			rw.State = goftlmux.TableBuffer                                          //

			if len(ur) > 0 {

				ur = ur[1:]
				uur := make(map[string]bool)
				for _, x := range ur {
					uur[x] = true
				}

				if db1 {
					fmt.Printf("uur = %+v, KEY=[[%s]]\n", uur, GenKey())
				}
				ks := hdlr.gCfg.GetKeys(GenKey() + "*") // Get Keys using Prefix+*

				if db1 {
					fmt.Printf("ks = %+v\n", ks)
				}

				conn, err := hdlr.gCfg.RedisPool.Get()
				if err != nil {
					logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
					return
				}

				// get/filter data for each key
				for _, aKey := range ks {
					data, err := conn.Cmd("GET", aKey).Str() // Get the value
					if db1 {
						fmt.Printf("data[%s] = %s\n", aKey, data)
					}
					if err == nil { // If we got the data

						mdata := make(map[string]interface{})
						fdata := make(map[string]interface{})
						err = json.Unmarshal([]byte(data), &mdata) // JSON parse the data
						if err != nil {
							goto next
						} else {
							if lib.ApplyFilter(hdlr.filter, mdata) {
								privs := mdata["privs"].(string) // xyzzy - improve this
								if db1 {
									fmt.Printf("From Data in Redis: privs = %s\n", privs)
								}
								if (len(hdlr.UserRolesReject) > 0 && lib.InArray(privs, hdlr.UserRolesReject)) || (privs != "anon-user") {
									if uur["$key$"] { // if have special $key$ then include key in return data
										// fdata["$key$"] = fmt.Sprintf("\"$key$\":%q", aKey[len(GenKey()):])
										fdata["$key$"] = aKey[len(GenKey()):]
									}
									for jj, ww := range mdata { // for each of the data items
										if uur[jj] { // if permitted based on role
											fdata[jj] = ww
										}
									}
								}
							}
						}

						rw.Table = append(rw.Table, fdata)

					}
				next:
				}

				hdlr.gCfg.RedisPool.Put(conn)

			}

		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		hdlr.Next.ServeHTTP(www, req)
	}
}

const db1 = false
const db2 = false // filters and parsing of them

/* vim: set noai ts=4 sw=4: */