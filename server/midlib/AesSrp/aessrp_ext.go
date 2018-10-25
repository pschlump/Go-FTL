// Package AesSrp implements encrypted authentication and encrypted REST.
// SRP-6a for login authentication, followed by AES 256 bit encrypted RESTful calls.
// A security model with roles is also implemented.
//
// Copyright (C) Philip Schlump, 2013-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.10
// BuildNo: 2101
// ErrorCode: 1000..1232 -- rserved to 1300
//

package AesSrp

/*

// func ConfigEmailAWS(hdlr *AesSrpType, file string) {
TODO:

	1. Admin interface for device IDs -- CRUD from admin perspecitve

	// xyzzy2016 - email be sent to warn user that an attempt was maid to re-register with this name
	// xyzzy2016 - email be sent to notify user that a new 2fa DeviceID was registered to this account

TODO:
	# Fix00000
	#  		+3. Force Loggout on an account																	1h (implement as a script above)
	# 		1. account has no maping to 't' value to get sessions
	# 		2. Need a way to list sessions and how old they are, get all "t" values -- A List
	# 			email -> list of t
	# 			List of "t" -> email -- with a daily cleanup? (24 lists with a TTL)
	#
	# 	1. /api/srp_validate - add to lists
	# 	2. /api/cipher - add to lists
	#  	3. /api/srp_logout - remove from lits
	//
	// 2. Show Logged In -- List all the Logged In Accounts/Types, Time of Login								4d
	//	  Last time of Activity, Time of Expire
	//		1. Show rate of login/logout
	//		2. Show average time for a login
	//		3. Show average time for a request
	//		4. Profile requests - histo of api calls.
	//





===========================================================================================================================================
== Cookies
===========================================================================================================================================

Cookies:
	"LoginAuthToken"
		Used In:
			func respHandlerRecoverPw2(www http.ResponseWriter, req *http.Request) {			-> email_auth_token
			func respHandlerChangePassword(www http.ResponseWriter, req *http.Request) {		-> "x" (chagned to delete)
			func respHandlerRecoverPasswordPt2(www http.ResponseWriter, req *http.Request) {	-> email_auth_token
			func respHandlerAdminSetPassword(www http.ResponseWriter, req *http.Request) {		-> "x" (chagned to delete)
			func respHandlerAdminSetAttributes(www http.ResponseWriter, req *http.Request) {	-> "x" (chagned to delete)
			func respHandlerSRPLogin(www http.ResponseWriter, req *http.Request) {				-> "x" (chagned to delete)
		Set To:
			"x"
			email_auth_token

	"LoginAuthCookie"
		Used In:
			func respHandlerSRPChallenge(www http.ResponseWriter, req *http.Request) {			-> "deleted" + Expired
			func respHandlerSRPValidate(www http.ResponseWriter, req *http.Request) {			-> An UUID
		Set To:
			"deleted"
			UUID
		Description:
			This is used in 2 ways
			1) 	This is used as the username for anon-user logins in combination with the
			   	browser fingerprint as the password.  This user is a generated user that
				allows a person to have "stayLoggedIn" facilities and enables encryption
				of RESTful calls.   In this fashion it is set on respHandlerSRPValidate to
				a value so that an anonymous page that load can find this cookie and use it
				to login as anon-user.
			2)	In the LoginRequired middleware to validate that a user has logged in.
				The middleware checks that the "privs" are not "anon-user" to prevent
				anon-user from having login privileges. -- Change to -- user privileges
				based on "privs" and what can bee seen. -- In this usage it is matched
				with "LoginHashCookie" - ?? how matched ?? how used?

	"LoginHashCookie"
		Used In:
			func respHandlerSRPValidate(www http.ResponseWriter, req *http.Request) {			-> An UUID
		Set To:
			UUID
		Expires:
			in d.b. based on CookieSessionLife
			As a cookie based on CookieSessionLife
		Description:
			Data is saved in the database under this cookie ID with the email and privs of user.
			Note: SaveCookieAuth(hdlr, cookieValue, SandBoxPrefix, ip, email, hash, id, privs)

	 "LoginAuthEmail"
		Used In:
			func respHandlerRecoverPw2(www http.ResponseWriter, req *http.Request) {			-> email
		Set To:
			email
		Description:
			Cookie is used to pass the "email" address thorugh the response-redirect (307)
			to the handler for the password reset.
		Redirect To:
			www.Header().Set("Location", https+req.Host+"/#/pwrecov2")

*/

import (
	"crypto/aes"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	JsonX "github.com/pschlump/JSONx"

	"github.com/Sirupsen/logrus"                 //
	"github.com/pschlump/AesCCM"                 //
	"github.com/pschlump/AesCCM/base64data"      //
	"github.com/pschlump/AesCCM/sjcl"            //
	"github.com/pschlump/Go-FTL/server/cfg"      //
	"github.com/pschlump/Go-FTL/server/goftlmux" //
	"github.com/pschlump/Go-FTL/server/httpmux"  // Parameter parsing and handling
	"github.com/pschlump/Go-FTL/server/lib"      //
	"github.com/pschlump/Go-FTL/server/mid"      //
	"github.com/pschlump/Go-FTL/server/tmplp"    //
	"github.com/pschlump/HashStrings"            //
	"github.com/pschlump/MiscLib"                //
	"github.com/pschlump/check-json-syntax/lib"  //
	"github.com/pschlump/godebug"                //
	"github.com/pschlump/gosrp"                  //
	"github.com/pschlump/gosrp/big"              //
	"github.com/pschlump/json"                   //	"encoding/json" - modified to allow dummy output of channels
	"github.com/pschlump/uuid"                   // defects fixed, modified to be faster
	"github.com/pschlump/verhoeff_algorithm"     //
	"golang.org/x/crypto/pbkdf2"                 // "golang.org/x/crypto/pbkdf2"
)

// xyzzyDeviceID - management calls for multiple DeviceID's - list of them, delete one, add one, update name on DeviceID
type DeviceIDType struct {
	DeviceID     string
	CreationDate string // Rfc3339 format date
	Title        string // Description of Device - optional
}

// fmt.Fprintf(www, `{"status":"success","M2":"%s","first_login":%v,"more_backup_keys":%v,"TwoFactorRequired":%v}`, m2, first_login, more_backup_keys, TwoFactorRequired)
type LoginRetrunValue struct {
	Status            string           `json:"status"`
	M2                string           `json:"M2"`
	FirstLogin        bool             `json:"first_login"`
	MoreBackupKeys    bool             `json:"more_backup_keys"`
	TwoFactorRequired string           `json:"TwoFactorRequired"`
	UserRole          RolesWithBitMask `json:"userRole"`
	DeviceID          string           `json:"DeviceID"` // xyzzyDeviceID
	DeviceIDList      []DeviceIDType   `json:"DeviceIDList"`
	BackupKeys        string           `json:"BackupKeys"`
	OwnerEmail        string           `json:"OwnerEmail"`
	LoginLastsTill    string           `json:"LoginLastsTill"`
	LoginLastsSeconds int              `json:"LoginLastsSeconds"`
	RealName          string           `json:"RealName"`
	PhoneNo           string           `json:"PhoneNo"`
	FirstName         string           `json:"FirstName"`
	MidName           string           `json:"MidName"`
	LastName          string           `json:"LastName"`
	UserName          string           `json:"UserName"`
	XAttrs            string           `json:"-"`
	HaveAnon          bool             `json:"have_anon"`
}

type LoginRetrunValueNo2fa struct {
	Status            string           `json:"status"`
	M2                string           `json:"M2"`
	FirstLogin        bool             `json:"first_login"`
	TwoFactorRequired string           `json:"TwoFactorRequired"`
	UserRole          RolesWithBitMask `json:"userRole"`
	OwnerEmail        string           `json:"OwnerEmail"`
	LoginLastsTill    string           `json:"LoginLastsTill"`
	LoginLastsSeconds int              `json:"LoginLastsSeconds"`
	RealName          string           `json:"RealName"`
	PhoneNo           string           `json:"PhoneNo"`
	FirstName         string           `json:"FirstName"`
	MidName           string           `json:"MidName"`
	LastName          string           `json:"LastName"`
	UserName          string           `json:"UserName"`
	XAttrs            string           `json:"-"`
	HaveAnon          bool             `json:"have_anon"`
}

// Taken from rfc5054 - default values.
var g_tVal = `{
		"2048": {"g":"2","N":"ac6bdb41324a9a9bf166de5e1389582faf72b6651987ee07fc3192943db56050a37329cbb4a099ed8193e0757767a13dd52312ab4b03310dcd7f48a9da04fd50e8083969edb767b0cf6095179a163ab3661a05fbd5faaae82918a9962f0b93b855f97993ec975eeaa80d740adbf4ff747359d041d5c33ea71d281e446b14773bca97b43a23fb801676bd207a436c6481f1d2b9078717461a5b9d32e688f87748544523b524b0d57d5ea77a2775d2ecfa032cfbdbf52fb3786160279004e57ae6af874e7303ce53299ccc041c7bc308d82a5698f3a8d0c38271ae35f8e9dbfbb694b5c803d89f7ae435de236d525f54759b65e372fcd68ef20fa7111f9e4aff73"} 
		,"3072": {"g":"5","N":"ffffffffffffffffc90fdaa22168c234c4c6628b80dc1cd129024e088a67cc74020bbea63b139b22514a08798e3404ddef9519b3cd3a431b302b0a6df25f14374fe1356d6d51c245e485b576625e7ec6f44c42e9a637ed6b0bff5cb6f406b7edee386bfb5a899fa5ae9f24117c4b1fe649286651ece45b3dc2007cb8a163bf0598da48361c55d39a69163fa8fd24cf5f83655d23dca3ad961c62f356208552bb9ed529077096966d670c354e4abc9804f1746c08ca18217c32905e462e36ce3be39e772c180e86039b2783a2ec07a28fb5c55df06f4c52c9de2bcbf6955817183995497cea956ae515d2261898fa051015728e5a8aaac42dad33170d04507a33a85521abdf1cba64ecfb850458dbef0a8aea71575d060c7db3970f85a6e1e4c7abf5ae8cdb0933d71e8c94e04a25619dcee3d2261ad2ee6bf12ffa06d98a0864d87602733ec86a64521f2b18177b200cbbe117577a615d6c770988c0bad946e208e24fa074e5ab3143db5bfce0fd108e4b82d120a93ad2caffffffffffffffff"}
		,"4096": {"g":"5","N":"ffffffffffffffffc90fdaa22168c234c4c6628b80dc1cd129024e088a67cc74020bbea63b139b22514a08798e3404ddef9519b3cd3a431b302b0a6df25f14374fe1356d6d51c245e485b576625e7ec6f44c42e9a637ed6b0bff5cb6f406b7edee386bfb5a899fa5ae9f24117c4b1fe649286651ece45b3dc2007cb8a163bf0598da48361c55d39a69163fa8fd24cf5f83655d23dca3ad961c62f356208552bb9ed529077096966d670c354e4abc9804f1746c08ca18217c32905e462e36ce3be39e772c180e86039b2783a2ec07a28fb5c55df06f4c52c9de2bcbf6955817183995497cea956ae515d2261898fa051015728e5a8aaac42dad33170d04507a33a85521abdf1cba64ecfb850458dbef0a8aea71575d060c7db3970f85a6e1e4c7abf5ae8cdb0933d71e8c94e04a25619dcee3d2261ad2ee6bf12ffa06d98a0864d87602733ec86a64521f2b18177b200cbbe117577a615d6c770988c0bad946e208e24fa074e5ab3143db5bfce0fd108e4b82d120a92108011a723c12a787e6d788719a10bdba5b2699c327186af4e23c1a946834b6150bda2583e9ca2ad44ce8dbbbc2db04de8ef92e8efc141fbecaa6287c59474e6bc05d99b2964fa090c3a2233ba186515be7ed1f612970cee2d7afb81bdd762170481cd0069127d5b05aa993b4ea988d8fddc186ffb7dc90a6c08f4df435c934063199ffffffffffffffff"}
		,"6144": {"g":"5","N":"ffffffffffffffffc90fdaa22168c234c4c6628b80dc1cd129024e088a67cc74020bbea63b139b22514a08798e3404ddef9519b3cd3a431b302b0a6df25f14374fe1356d6d51c245e485b576625e7ec6f44c42e9a637ed6b0bff5cb6f406b7edee386bfb5a899fa5ae9f24117c4b1fe649286651ece45b3dc2007cb8a163bf0598da48361c55d39a69163fa8fd24cf5f83655d23dca3ad961c62f356208552bb9ed529077096966d670c354e4abc9804f1746c08ca18217c32905e462e36ce3be39e772c180e86039b2783a2ec07a28fb5c55df06f4c52c9de2bcbf6955817183995497cea956ae515d2261898fa051015728e5a8aaac42dad33170d04507a33a85521abdf1cba64ecfb850458dbef0a8aea71575d060c7db3970f85a6e1e4c7abf5ae8cdb0933d71e8c94e04a25619dcee3d2261ad2ee6bf12ffa06d98a0864d87602733ec86a64521f2b18177b200cbbe117577a615d6c770988c0bad946e208e24fa074e5ab3143db5bfce0fd108e4b82d120a92108011a723c12a787e6d788719a10bdba5b2699c327186af4e23c1a946834b6150bda2583e9ca2ad44ce8dbbbc2db04de8ef92e8efc141fbecaa6287c59474e6bc05d99b2964fa090c3a2233ba186515be7ed1f612970cee2d7afb81bdd762170481cd0069127d5b05aa993b4ea988d8fddc186ffb7dc90a6c08f4df435c93402849236c3fab4d27c7026c1d4dcb2602646dec9751e763dba37bdf8ff9406ad9e530ee5db382f413001aeb06a53ed9027d831179727b0865a8918da3edbebcf9b14ed44ce6cbaced4bb1bdb7f1447e6cc254b332051512bd7af426fb8f401378cd2bf5983ca01c64b92ecf032ea15d1721d03f482d7ce6e74fef6d55e702f46980c82b5a84031900b1c9e59e7c97fbec7e8f323a97a7e36cc88be0f1d45b7ff585ac54bd407b22b4154aacc8f6d7ebf48e1d814cc5ed20f8037e0a79715eef29be32806a1d58bb7c5da76f550aa3d8a1fbff0eb19ccb1a313d55cda56c9ec2ef29632387fe8d76e3c0468043e8f663f4860ee12bf2d5b0b7474d6e694f91e6dcc4024ffffffffffffffff"}
		,"8192": {"g":"19","N":"ffffffffffffffffc90fdaa22168c234c4c6628b80dc1cd129024e088a67cc74020bbea63b139b22514a08798e3404ddef9519b3cd3a431b302b0a6df25f14374fe1356d6d51c245e485b576625e7ec6f44c42e9a637ed6b0bff5cb6f406b7edee386bfb5a899fa5ae9f24117c4b1fe649286651ece45b3dc2007cb8a163bf0598da48361c55d39a69163fa8fd24cf5f83655d23dca3ad961c62f356208552bb9ed529077096966d670c354e4abc9804f1746c08ca18217c32905e462e36ce3be39e772c180e86039b2783a2ec07a28fb5c55df06f4c52c9de2bcbf6955817183995497cea956ae515d2261898fa051015728e5a8aaac42dad33170d04507a33a85521abdf1cba64ecfb850458dbef0a8aea71575d060c7db3970f85a6e1e4c7abf5ae8cdb0933d71e8c94e04a25619dcee3d2261ad2ee6bf12ffa06d98a0864d87602733ec86a64521f2b18177b200cbbe117577a615d6c770988c0bad946e208e24fa074e5ab3143db5bfce0fd108e4b82d120a92108011a723c12a787e6d788719a10bdba5b2699c327186af4e23c1a946834b6150bda2583e9ca2ad44ce8dbbbc2db04de8ef92e8efc141fbecaa6287c59474e6bc05d99b2964fa090c3a2233ba186515be7ed1f612970cee2d7afb81bdd762170481cd0069127d5b05aa993b4ea988d8fddc186ffb7dc90a6c08f4df435c93402849236c3fab4d27c7026c1d4dcb2602646dec9751e763dba37bdf8ff9406ad9e530ee5db382f413001aeb06a53ed9027d831179727b0865a8918da3edbebcf9b14ed44ce6cbaced4bb1bdb7f1447e6cc254b332051512bd7af426fb8f401378cd2bf5983ca01c64b92ecf032ea15d1721d03f482d7ce6e74fef6d55e702f46980c82b5a84031900b1c9e59e7c97fbec7e8f323a97a7e36cc88be0f1d45b7ff585ac54bd407b22b4154aacc8f6d7ebf48e1d814cc5ed20f8037e0a79715eef29be32806a1d58bb7c5da76f550aa3d8a1fbff0eb19ccb1a313d55cda56c9ec2ef29632387fe8d76e3c0468043e8f663f4860ee12bf2d5b0b7474d6e694f91e6dbe115974a3926f12fee5e438777cb6a932df8cd8bec4d073b931ba3bc832b68d9dd300741fa7bf8afc47ed2576f6936ba424663aab639c5ae4f5683423b4742bf1c978238f16cbe39d652de3fdb8befc848ad922222e04a4037c0713eb57a81a23f0c73473fc646cea306b4bcbc8862f8385ddfa9d4b7fa2c087e879683303ed5bdd3a062b3cf5b3a278a66d2a13f83f44f82ddf310ee074ab6a364597e899a0255dc164f31cc50846851df9ab48195ded7ea1b1d510bd7ee74d73faf36bc31ecfa268359046f4eb879f924009438b481c6cd7889a002ed5ee382bc9190da6fc026e479558e4475677e9aa9e3050e2765694dfc81f56e880b96e7160c980dd98edd3dfffffffffffffffff"}
}
`

var g_SecurityData_Default = `{
	"Roles": [
		"public",
		"user",
		"admin",
		"root"
	],
	"AccessLevels": {
		"admin": [ "admin" ],
		"anon": [ "public" ],
		"public": [ "*" ],
		"root": [ "root", "admin", "user", "public" ],
		"user": [ "user", "admin" ]
	},
	"Privilages": {
		"admin": [ "MayChangeOtherPassword", "MayCreateAdminAccounts", "MayChangeOtherAttributes", "MayGetOneTimeKey", "MayGetSetAttrs" ]
	},
	"MayAccessApi": {
		"DeviceID": [ "/api/srp_register", "/api/srp_login", "/api/srp_challenge", "/api/srp_validate", "/api/srp_getNg", "/api/send_support_message",
			"/api/version", "/api/srp_logout", "/api/cipher", "/api/get2FactorFromDeviceID", "/api/enc_version" ],
		"admin": [ "*", "/api/admin" ],
		"anon": [ "*", "/api/anon" ],
		"public": [ "*", "/api/public" ],
		"root": [ "*", "/api/root" ],
		"user": [ "*", "/api/user" ]
	}
}
`

type Ng_struct struct {
	N string `json:"N"` // base 16, value - large
	G string `json:"g"` // base 10, value - 2, 5, 19 etc.
}

const LoginAuthCookieLife = (1 * 24 * 60 * 60) // one day in seconds
const LoginHashCookieLife = (1 * 24 * 60 * 60) // one day in seconds

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical -but- not in this case.
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*AesSrpType)
//		if ok {
//			pCfg.SetNext(next)
//			rv = pCfg
//		} else {
//			err = mid.FtlConfigError
//			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
//		}
//
//		SetDebugFlagsFromGlobal(gCfg)
//
//		// EmailRelayIP = pCfg.EmailRelayIP
//		// EmailAuthToken = pCfg.EmailAuthToken
//		for _, xx := range pCfg.TestModeInject {
//			TestModeInject[xx] = true
//		}
//		gCfg.ConnectToRedis()
//		gCfg.ConnectToPostgreSQL()
//		pCfg.gCfg = gCfg
//
//		if dbDumpConfig {
//			fmt.Printf("[][][][][][][][][] Config: %s\n", lib.SVarI(pCfg))
//		}
//		return
//	}
//
//	// normally identical - not this time
//	createEmptyType := func() interface{} {
//		rv := &AesSrpType{}
//		rv.mux = initRegularMux()
//		rv.muxEnc = initMuxEnc()
//		return rv
//	}
//
//	postInitValidation := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
//		if init_db1 {
//			fmt.Printf("In postInitValidation for AesSrp, %s\n", godebug.LF())
//		}
//		hh, ok := h.(*AesSrpType)
//		if !ok {
//			fmt.Printf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo)
//			return mid.ErrInternalError
//		} else {
//			ConfigEmailAWS(hh, hh.EmailConfigFileName)
//			if init_db1 {
//				fmt.Printf("Parsed Data Is: %s\n", lib.SVarI(hh))
//			}
//			if len(hh.DbUserCols) == 0 {
//				hh.DbUserCols = []string{"RealName", "Customer_id", "UserName", "FirstName", "MidName", "LastName ", "User_id", "Customer_id", "XAttrs", "PhoneNo"}
//				hh.DbUserColsDb = []string{"RealName", "Customer_id", "UserName", "FirstName", "MidName", "LastName ", "User_id", "Customer_id", "XAttrs", "PhoneNo"}
//			}
//			hh.anonUserPaths = make(map[string]bool)
//			for _, vv := range []string{"/api/1x1.gif", "/api/cipher", "/api/enc_version", "/api/send_support_message", "/api/srp_challenge", "/api/srp_getNg", "/api/srp_login", "/api/srp_logout", "/api/srp_register", "/api/srp_validate", "/api/version", "/api/resumeLogin"} {
//				hh.anonUserPaths[vv] = true
//			}
//			for _, vv := range hh.AnonUserPaths {
//				if vv[0:1] == "-" {
//					hh.anonUserPaths[vv[1:]] = false
//				} else {
//					hh.anonUserPaths[vv] = true
//				}
//			}
//			if len(hh.SecurityConfig.Roles) == 0 {
//				var SecurityData SecurityConfigType
//				err := json.Unmarshal([]byte(g_SecurityData_Default), &SecurityData)
//				if err != nil {
//					fmt.Printf("Unable to parse supplided security data\n")
//					es := jsonSyntaxErroLib.GenerateSyntaxError(g_SecurityData_Default, err)
//					fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorYellow, es, MiscLib.ColorReset)
//					logrus.Errorf("Error: Invlaid JSON Error:\n%s\n", es)
//					return mid.ErrInternalError
//				}
//				// SecurityData := lib.SVar(g_SecurityData_Default)
//				Rn, RnH, An := SetupRoles(SecurityData.Roles, SecurityData.AccessLevels)
//				hh.SecurityConfig = SecurityData
//				hh.secRn = Rn
//				hh.secRnH = RnH
//				hh.secAn = An
//			} else {
//				Rn, RnH, An := SetupRoles(hh.SecurityConfig.Roles, hh.SecurityConfig.AccessLevels)
//				hh.secRn = Rn
//				hh.secRnH = RnH
//				hh.secAn = An
//			}
//			if len(hh.NGData.N) == 0 {
//				ng_data := make(map[string]Ng_struct)
//				err := json.Unmarshal([]byte(g_tVal), &ng_data)
//				if err != nil {
//					fmt.Printf("Unable to parse constant confiration data, Data: >>>%s<<<", g_tVal)
//					es := jsonSyntaxErroLib.GenerateSyntaxError(g_tVal, err)
//					fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorYellow, es, MiscLib.ColorReset)
//					logrus.Errorf("Error: Invlaid JSON Error:\n%s\n", es)
//					return mid.ErrInternalError
//				}
//				if x, ok := ng_data[fmt.Sprintf("%d", hh.Bits)]; ok {
//					hh.NGData = x
//				} else {
//					fmt.Printf("Invalid size for Bits, %v\n", hh.Bits)
//					return mid.ErrInternalError
//				}
//			} // xyzzy - else -- N,G supplied - should check length / values
//		}
//		return nil
//	}
//
//	// SRP and AES Config --------------------------------------------------------------------------------------------------
//	cfg.RegInitItem2("SrpAesAuth", initNext, createEmptyType, postInitValidation, `{
//		}`)
//
//	dataStore = NewRSaveToRedis("srp:U:")
//	emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@.*[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
//	hexRe = regexp.MustCompile(`^[a-fA-F0-9]+$`)
//	isUUIDRe = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
//}
//
//// normally identical
//func (hdlr *AesSrpType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &AesSrpType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		x.mux = initRegularMux()
		x.muxEnc = initMuxEnc()
		return x
	}
	mid.RegInitItem3("SrpAesAuth", CreateEmpty, `{
		"Paths":                    { "type":[ "string","filepath" ], "isarray":true, "required":true },
		"EncReqPaths":              { "type":[ "string","filepath" ], "isarray":true },
		"MatchPaths":               { "type":[ "string","filepath" ], "isarray":true },
		"Bits":                     { "type":[ "int" ], "default":"2048" },
		"NGData":					{ "type":[ "struct" ] },
		"SendStatusOnError":        { "type":[ "bool" ], "default":"false" },
		"AdminPassword":            { "type":[ "string" ], "default":"green eggs and ham" },
		"FailedLoginThreshold":     { "type":[ "int" ], "default":"10" },
		"NewUserPrivs":             { "type":[ "string" ], "default":"user" },
		"SendEmail":                { "type":[ "bool" ], "default":"true" },
		"EmailApp":                 { "type":[ "string" ], "default":"user-login" },
		"KermitRule":               { "type":[ "bool" ], "default":"true" },
		"EmailConfigFileName":      { "type":[ "string" ], "default": "./email-config.json" },
		"SupportEmailTo":           { "type":[ "string" ], "default":"pschlump@gmail.com" },
		"TwoFactorRequired":        { "type":[ "string" ], "default":"y" },
		"BackupKeyIter":            { "type":[ "int" ], "default":"1000" },
		"KeyIter":                  { "type":[ "int" ], "default":"1000" },
		"BackupKeySizeBytes":       { "type":[ "int" ], "default":"16" },
		"CookieExpireInXDays":      { "type":[ "int" ], "default":"1" },
		"CookieExpireInXDays2":     { "type":[ "int" ], "default":"2" },
		"SessionLife":              { "type":[ "int" ], "default":"86400" },
		"KeySessionLife":           { "type":[ "int" ], "default":"300" },
		"CookieSessionLife":        { "type":[ "int" ], "default":"172800" },
		"TwoFactorLife":            { "type":[ "int" ], "default":"360" },
		"PreEau":                   { "type":[ "string" ], "default":"eau:" },
		"PreKey":                   { "type":[ "string" ], "default":"ses:" },
		"PreAuth":                  { "type":[ "string" ], "default":"aut:" },
		"Pre2Factor":               { "type":[ "string" ], "default":"p2f:" },
		"PwResetKey":               { "type":[ "string" ], "default":"pwr:" },
		"PwExpireIn":               { "type":[ "int" ], "default":"86400" },
		"TestModeInject":           { "type":[ "string" ], "isarray":true },
		"PasswordSV":               { "type":[ "string" ] },
		"SandBoxExpreTime":         { "type":[ "int" ], "default":"7200" },
		"SecurityAccessLevelsName": { "type":[ "hash" ] },
		"SecurityPrivilages":       { "type":[ "hash" ] },
		"StayLoggedInExpire":       { "type":[ "int" ], "default":"86400" },
		"UserNameForRegister":      { "type":[ "bool" ], "default":"false" },
		"SecurityConfig":           { "type":[ "struct" ] },
		"PwRecoverTemplate1":       { "type":[ "string" ], "default":"{{.HTTPS}}{{.HOST}}/unable-to-pwrecov1.html" },
		"PwRecoverTemplate2":       { "type":[ "string" ], "default":"{{.HTTPS}}{{.HOST}}/unable-to-pwrecov2.html" },
		"PwRecoverTemplate3":       { "type":[ "string" ], "default":"{{.HTTPS}}{{.HOST}}/#/pwrecov2" },
		"RegTemplate1":             { "type":[ "string" ], "default":"{{.HTTPS}}{{.HOST}}/unable-to-register1.html" },
		"RegTemplate2":             { "type":[ "string" ], "default":"{{.HTTPS}}{{.HOST}}/unable-to-register2.html" },
		"RegTemplate3":             { "type":[ "string" ], "default":"{{.HTTPS}}{{.HOST}}/unable-to-register3.html" },
		"RegTemplate4":             { "type":[ "string" ], "default":"{{.HTTPS}}{{.HOST}}/#/login" },
		"AllowReregisterDeviceID":  { "type":[ "bool" ], "default":"false" },
		"LimitDeviceIDs":           { "type":[ "int" ], "default":"20" },
		"InDemoMode":               { "type":[ "bool" ], "default":"false" },
		"InTestMode":               { "type":[ "bool" ], "default":"false" },
		"DbUserColAPI":             { "type":[ "string" ], "default":"/api/table/t_user" },
		"DbUserCols":               { "type":[ "string" ], "isarray":true },
		"DbUserColsDb":             { "type":[ "string" ], "isarray":true },
		"AnonUserPaths":            { "type":[ "string","filepath" ], "isarray":true },
	    "NonEmailAccts":            { "type":[ "string" ], "isarray":true },
		"LineNo":                   { "type":[ "int" ], "default":"1" }
		}`)
	dataStore = NewRSaveToRedis("srp:U:")
	emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@.*[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	hexRe = regexp.MustCompile(`^[a-fA-F0-9]+$`)
	isUUIDRe = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
}

func (hdlr *AesSrpType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init

	SetDebugFlagsFromGlobal(gCfg)

	// EmailRelayIP = pCfg.EmailRelayIP
	// EmailAuthToken = pCfg.EmailAuthToken
	for _, xx := range hdlr.TestModeInject {
		TestModeInject[xx] = true
	}
	gCfg.ConnectToRedis()
	gCfg.ConnectToPostgreSQL()
	hdlr.gCfg = gCfg

	if dbDumpConfig {
		fmt.Printf("[][][][][][][][][] Config: %s\n", lib.SVarI(hdlr))
	}

	return
}

func (hdlr *AesSrpType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	ConfigEmailAWS(hdlr, hdlr.EmailConfigFileName)
	if init_db1 {
		fmt.Printf("Parsed Data Is: %s\n", lib.SVarI(hdlr))
	}
	if len(hdlr.DbUserCols) == 0 {
		hdlr.DbUserCols = []string{"RealName", "Customer_id", "UserName", "FirstName", "MidName", "LastName ", "User_id", "Customer_id", "XAttrs", "PhoneNo"}
		hdlr.DbUserColsDb = []string{"RealName", "Customer_id", "UserName", "FirstName", "MidName", "LastName ", "User_id", "Customer_id", "XAttrs", "PhoneNo"}
	}
	hdlr.anonUserPaths = make(map[string]bool)
	for _, vv := range []string{"/api/1x1.gif", "/api/cipher", "/api/enc_version", "/api/send_support_message", "/api/srp_challenge", "/api/srp_getNg", "/api/srp_login", "/api/srp_logout", "/api/srp_register", "/api/srp_validate", "/api/version", "/api/resumeLogin"} {
		hdlr.anonUserPaths[vv] = true
	}
	for _, vv := range hdlr.AnonUserPaths {
		if vv[0:1] == "-" {
			hdlr.anonUserPaths[vv[1:]] = false
		} else {
			hdlr.anonUserPaths[vv] = true
		}
	}
	if len(hdlr.SecurityConfig.Roles) == 0 {
		var SecurityData SecurityConfigType
		err := json.Unmarshal([]byte(g_SecurityData_Default), &SecurityData)
		if err != nil {
			fmt.Printf("Unable to parse supplided security data\n")
			es := jsonSyntaxErroLib.GenerateSyntaxError(g_SecurityData_Default, err)
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorYellow, es, MiscLib.ColorReset)
			logrus.Errorf("Error: Invlaid JSON Error:\n%s\n", es)
			return mid.ErrInternalError
		}
		// SecurityData := lib.SVar(g_SecurityData_Default)
		Rn, RnH, An := SetupRoles(SecurityData.Roles, SecurityData.AccessLevels)
		hdlr.SecurityConfig = SecurityData
		hdlr.secRn = Rn
		hdlr.secRnH = RnH
		hdlr.secAn = An
	} else {
		Rn, RnH, An := SetupRoles(hdlr.SecurityConfig.Roles, hdlr.SecurityConfig.AccessLevels)
		hdlr.secRn = Rn
		hdlr.secRnH = RnH
		hdlr.secAn = An
	}
	if len(hdlr.NGData.N) == 0 {
		ng_data := make(map[string]Ng_struct)
		err := json.Unmarshal([]byte(g_tVal), &ng_data)
		if err != nil {
			fmt.Printf("Unable to parse constant confiration data, Data: >>>%s<<<", g_tVal)
			es := jsonSyntaxErroLib.GenerateSyntaxError(g_tVal, err)
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorYellow, es, MiscLib.ColorReset)
			logrus.Errorf("Error: Invlaid JSON Error:\n%s\n", es)
			return mid.ErrInternalError
		}
		if x, ok := ng_data[fmt.Sprintf("%d", hdlr.Bits)]; ok {
			hdlr.NGData = x
		} else {
			fmt.Printf("Invalid size for Bits, %v\n", hdlr.Bits)
			return mid.ErrInternalError
		}
	} // xyzzy - else -- N,G supplied - should check length / values
	return
}

var _ mid.GoFTLMiddleWare = (*AesSrpType)(nil)

// --------------------------------------------------------------------------------------------------------------------------
type AesSrpType struct {
	Next                     http.Handler                //
	Paths                    []string                    // List of start paths where encryption will be used
	EncReqPaths              []string                    // start with same as Paths, but require encryption/login
	MatchPaths               []string                    // Paths that match this are allowed without authentication. (Static files for example)
	Bits                     int                         // Usually 2048 - number of bits for SRP authentication modulo number
	NGData                   Ng_struct                   // You can supply both the Bits and the actual n, G data
	SendStatusOnError        bool                        // Respond to errors with non-200 status (i.e. 4xx and 5xx errors)
	AdminPassword            string                      // If compiled in "TestMode" then this can be used to authenticate users - ignored otherwise
	FailedLoginThreshold     int                         // Number of failed logins before a delay is inserted, 10 by default.
	NewUserPrivs             string                      // What is a new users default privilege (role)
	SendEmail                bool                        // If true, then emails will be sent using the AWS email relay code.
	EmailApp                 string                      // default user-login, the login demo, set to something else for a different set of templates
	KermitRule               bool                        // If true, then emails to kermit.*@the-green-pc.com will not be sent (used for testing)
	EmailConfigFileName      string                      // name of file to take Email config from
	SupportEmailTo           string                      // Request for help and support emails go to this address
	TwoFactorRequired        string                      // If "y" then 2-factor-authentication is turned on
	BackupKeyIter            int                         // Number of iterations for running pbkdf2 on backup one time keys, default 1000
	KeyIter                  int                         // Number of iterations for running pbkdf2 on login passwords, default 1000 (mates with value in JS/Client code)
	BackupKeySizeBytes       int                         // Key size for backup one-time keys, default 16
	CookieExpireInXDays      int                         //
	CookieExpireInXDays2     int                         //
	SessionLife              int                         //
	KeySessionLife           int                         //
	CookieSessionLife        int                         //
	TwoFactorLife            int                         // how long is a temporary 2 factor key good for - 5 min  + 1 min grace
	PreEau                   string                      // Redis Prefix: "eau:"
	PreKey                   string                      // Redis Prefix: "ses:"
	PreAuth                  string                      // Redis Prefix: "aut:"
	Pre2Factor               string                      // Redis Prefix: "p2f:"		// Key used for 2fa:DeviceID -> OneTimeKey
	PwResetKey               string                      // Redis Prefix: "pwr:"		// Key for password recovery token - Exipre is PwExpireIn
	PwExpireIn               int                         // Default 86400 == 1 day in seconds, time for password recovery
	TestModeInject           []string                    // Array of strings to inject values - after auth converted to global variable for injecting errors - used only in test mode.
	PasswordSV               string                      //
	SandBoxExpreTime         int                         // about 2 hours
	SecurityAccessLevelsName map[string][]string         //
	SecurityPrivilages       map[string][]string         //
	StayLoggedInExpire       int                         // in seconds, time for login to persist when "stayLoggedIn" is true: 86400 = 1 day.
	UserNameForRegister      bool                        // default false, use email for username
	SecurityConfig           SecurityConfigType          //
	PwRecoverTemplate1       string                      //
	PwRecoverTemplate2       string                      //
	PwRecoverTemplate3       string                      //
	RegTemplate1             string                      //
	RegTemplate2             string                      //
	RegTemplate3             string                      //
	RegTemplate4             string                      //
	AllowReregisterDeviceID  bool                        // If true (Defaults to false) then will allow re-register of DeviceID (same id).  Good for development and testing only.
	LimitDeviceIDs           int                         //
	InDemoMode               bool                        //
	InTestMode               bool                        //
	DbUserColAPI             string                      //
	DbUserCols               []string                    //
	DbUserColsDb             []string                    //
	AnonUserPaths            []string                    // Set of additional paths that will be allowed to a "anon-user", if path starts with "-" then it will be delete from set.
	NonEmailAccts            []string                    // Set of account names like "admin" that need not be email addresses
	LineNo                   int                         // Lin in input file
	passwordSVKey            []byte                      // generated key from password
	mux                      *httpmux.ServeMux           // for non-encrypted (regular) calls
	muxEnc                   *httpmux.ServeMux           // for encrypted calls
	gCfg                     *cfg.ServerGlobalConfigType //
	secRn                    []RolesWithBitMask          // Derived Security flags
	secRnH                   map[string]uint64           // Derived SeSecurity flags
	secAn                    []RolesWithBitMask          // Derived SeSecurity flags
	anonUserPaths            map[string]bool             // Set of additional paths that will be allowed to a "anon-user", if path starts with "-" then it will be delete from set.
	srp_N                    interface{}                 // Values derived from NGData or defaults
	srp_g                    interface{}                 // Values derived from sNGData or defaults
}

func NewAesSrpServer(n http.Handler, p []string, e []string, gCfg *cfg.ServerGlobalConfigType) *AesSrpType {
	// ms.gCfg = cfg.ServerGlobal
	return &AesSrpType{Next: n, Paths: p, MatchPaths: e, mux: initRegularMux(), muxEnc: initMuxEnc(), gCfg: gCfg}
}

func (hdlr *AesSrpType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	fmt.Printf("AesSrp: Paths[%s] req.URL.Path [%s], %s\n", hdlr.Paths, req.URL.Path, godebug.LF())
	if lib.PathsMatch(hdlr.Paths, req.URL.Path) {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "AesSrp", hdlr.Paths, 0, req.URL.Path)

			rw.Next = hdlr.Next
			rw.Hdlr = hdlr
			hh, _, err := hdlr.mux.Handler(req) // rv.mux.ServeHTTP(www, req)
			if err == nil {
				hh.ServeHTTP(www, req)
				return
			}
		}
		fmt.Printf("\n-----------------------------------------------------------------------------------------------------\n")
		fmt.Printf("AesSrp: Fall through case, %s\n", godebug.LF())
		fmt.Printf("AesSrp: This indicates that you have a path that matched - and requries login, like /api/..., (one of[%s] is [%s])\n\tbut it was not encrypted and will be returned as an error!\n", hdlr.Paths, req.URL.Path)
		fmt.Printf("-----------------------------------------------------------------------------------------------------\n\n")
	} else if lib.PathsMatch(hdlr.MatchPaths, req.URL.Path) {
		hdlr.Next.ServeHTTP(www, req)
		return
	}
	errorHandlerFunc(www, req)
}

// ----------------------------------------------------------------------------------------------------------------------------

func initRegularMux() (mux *httpmux.ServeMux) {

	mux = httpmux.NewServeMux()

	// OPEN - exact match
	mux.HandleFunc("/api/srp_register", respHandlerSRPRegister).Method("GET", "POST")                    // start registration process
	mux.HandleFunc("/api/checkEmailAvailable", respHandlerCheckEmailAvailable).Method("GET", "POST")     // See if an email is available for registration and other options
	mux.HandleFunc("/api/srp_login", respHandlerSRPLogin).Method("GET", "POST")                          // start login process (step 1)
	mux.HandleFunc("/api/srp_challenge", respHandlerSRPChallenge).Method("GET", "POST")                  // start login process (step 2)
	mux.HandleFunc("/api/srp_validate", respHandlerSRPValidate).Method("GET", "POST")                    // start login process (step 3)
	mux.HandleFunc("/api/srp_getNg", respHandlerSRPGetNg).Method("GET", "POST")                          // Initialize for getting the SRP N,g values and 2fa on/off
	mux.HandleFunc("/api/srp_recover_password_pt1", respHandlerRecoverPasswordPt1).Method("GET", "POST") // password recovery (step 1) send email
	mux.HandleFunc("/api/srp_recover_password_pt2", respHandlerRecoverPasswordPt2).Method("GET", "POST") // password recovery (step 2) set new password
	mux.HandleFunc("/api/getPageToken", respHandlerGetPageToken).Method("GET", "POST")                   // Mark page with cookie for password recovery, "page" marker
	mux.HandleFunc("/api/setup_sandbox", respHandlerSetupSandbox).Method("GET", "POST")                  // Create temorary sandbox for testing
	mux.HandleFunc("/api/srp_email_confirm", respHandlerEmailConfirm).Method("GET", "POST")              //	User enters token, calls to set confirmed <- 'y'
	mux.HandleFunc("/api/confirm-registration", respHandlerConfirmRegistration).Method("GET")            // Redirect-Link: To 1st Login /#/login, confirmed <- 'y'
	mux.HandleFunc("/api/pwrecov2", respHandlerRecoverPw2).Method("GET")                                 // Redirect-Link: To /#/pwrecov2 with cookies set
	mux.HandleFunc("/api/send_support_message", respHandlerSendSupportMessage).Method("GET", "POST")     // email a message to support
	mux.HandleFunc("/api/version", respHandlerVersion).Method("GET")                                     // version of the srp-aes implementation
	mux.HandleFunc("/api/srp_logout", respHandlerSRPLogout).Method("GET", "POST")                        // Logout, no effect if not logged in.
	mux.HandleFunc("/api/cipher", respHandlerCipher).Method("GET", "POST")                               // Decrypt and rewrite request for other handlers
	mux.HandleFunc("/api/1x1.gif", respHandler1x1Gif).Method("GET")                                      // return the clear 1x1 gif -- Used for testing some stuff --
	mux.HandleFunc("/api/testGet2fa", respHandlerTest2faReturn).Method("GET", "POST")                    // Test function that returns a dummy 2fa 2nd factor

	// Login required
	// mux.HandleFunc("/api/srp_simulate_email_confirm", respHandlerSimulateEmailConfirm).Method("GET", "POST") // If in test mode can do a confirm of user

	// mux.HandleFunc("/api/testA", respHandlerTest1).Method("GET")       // Test 1 - of Etag
	mux.HandleFunc("/api/markPage", respHandlerMarkPage).Method("GET") // Setup for stay logged in

	// should be moved to encrypted after it works.
	if cfg_AllowReactNativeTest == "ReactNative" {
		mux.HandleFunc("/api/get2FactorFromDeviceID", respHandlerGet2FactorFromDeviceID).Method("GET", "POST") // return the One Time key when Device ID is presented.
	}

	mux.HandleErrors(http.StatusNotFound, httpmux.HandlerFunc(errorHandlerFunc))

	return

}

func initMuxEnc() (muxEnc *httpmux.ServeMux) {

	muxEnc = httpmux.NewServeMux()

	// OPEN - exact match - may reach this point with anon-user/stayLoggedIn user encrypting.
	muxEnc.HandleFunc("/api/srp_register", respHandlerSRPRegister).Method("GET", "POST")                    // start registration process
	muxEnc.HandleFunc("/api/checkEmailAvailable", respHandlerCheckEmailAvailable).Method("GET", "POST")     // See if an email is available for registration and other options
	muxEnc.HandleFunc("/api/srp_login", respHandlerSRPLogin).Method("GET", "POST")                          // start login process (step 1)
	muxEnc.HandleFunc("/api/srp_challenge", respHandlerSRPChallenge).Method("GET", "POST")                  // start login process (step 2)
	muxEnc.HandleFunc("/api/srp_validate", respHandlerSRPValidate).Method("GET", "POST")                    // start login process (step 3)
	muxEnc.HandleFunc("/api/srp_getNg", respHandlerSRPGetNg).Method("GET", "POST")                          // Initialize for getting the SRP N,g values and 2fa on/off
	muxEnc.HandleFunc("/api/srp_recover_password_pt1", respHandlerRecoverPasswordPt1).Method("GET", "POST") // password recovery (step 1) send email
	muxEnc.HandleFunc("/api/srp_recover_password_pt2", respHandlerRecoverPasswordPt2).Method("GET", "POST") // password recovery (step 2) set new password
	muxEnc.HandleFunc("/api/getPageToken", respHandlerGetPageToken).Method("GET", "POST")                   // Mark page with cookie for password recovery, "page" marker
	muxEnc.HandleFunc("/api/setup_sandbox", respHandlerSetupSandbox).Method("GET", "POST")                  // Create temorary sandbox for testing
	muxEnc.HandleFunc("/api/srp_email_confirm", respHandlerEmailConfirm).Method("GET", "POST")              //	User enters token, calls to set confirmed <- 'y'
	muxEnc.HandleFunc("/api/confirm-registration", respHandlerConfirmRegistration).Method("GET")            // Redirect-Link: To 1st Login /#/login, confirmed <- 'y'
	muxEnc.HandleFunc("/api/pwrecov2", respHandlerRecoverPw2).Method("GET")                                 // Redirect-Link: To /#/pwrecov2 with cookies set
	muxEnc.HandleFunc("/api/send_support_message", respHandlerSendSupportMessage).Method("GET", "POST")     // email a message to support
	muxEnc.HandleFunc("/api/version", respHandlerVersion).Method("GET")                                     // version of the srp-aes implementation
	muxEnc.HandleFunc("/api/srp_logout", respHandlerSRPLogout).Method("GET", "POST")                        // Logout, no effect if not logged in.
	muxEnc.HandleFunc("/api/markPage", respHandlerMarkPage).Method("GET")                                   // Setup for stay logged in
	muxEnc.HandleFunc("/api/testGet2fa", respHandlerTest2faReturn).Method("GET", "POST")                    // Test function that returns a dummy 2fa 2nd factor

	// Login Required - 't=' value is required.

	muxEnc.HandleFunc("/api/set_user_attrs", respHandlerSetUserAttrs).Method("GET", "POST")                     // ENC:
	muxEnc.HandleFunc("/api/get_user_attrs", respHandlerGetUserAttrs).Method("GET", "POST")                     // ENC:
	muxEnc.HandleFunc("/api/admin_set_user_attrs", respHandlerAdminSetUserAttrs).Method("GET", "POST")          // ENC:
	muxEnc.HandleFunc("/api/admin_get_user_attrs", respHandlerAdminGetUserAttrs).Method("GET", "POST")          // ENC:
	muxEnc.HandleFunc("/api/srp_change_password", respHandlerChangePassword).Method("GET", "POST")              // ENC: Set a new password
	muxEnc.HandleFunc("/api/valid2Factor", respHandlerValid2Factor).Method("GET", "POST")                       // ENC: validate a one-time key - uses 't'/user_id to validate OneTimeKey - Login
	muxEnc.HandleFunc("/api/genTempKeys", respHandlerGenTempKeys).Method("GET", "POST")                         // ENC: get 20 one time keys.
	muxEnc.HandleFunc("/api/getDeviceID", respHandlerGetDeviceID).Method("GET", "POST")                         // ENC: get users DeviceID	- turns on Reveal for 1 hour
	muxEnc.HandleFunc("/api/createNewDeviceID", respHandlerCreateNewDeviceID).Method("GET", "POST")             // ENC: repalce DeviceID	replaced old device ID with new and invalidates old
	muxEnc.HandleFunc("/api/updateDeviceID", respHandlerUpdateDeviceID).Method("GET", "POST")                   // ENC:
	muxEnc.HandleFunc("/api/deleteDeviceID", respHandlerDeleteDeviceID).Method("GET", "POST")                   // ENC:
	muxEnc.HandleFunc("/api/setDebugFlags", respHandlerSetDebugFlags).Method("GET", "POST")                     // ENC: turn on/off debugging flags
	muxEnc.HandleFunc("/api/setup_sandbox", respHandlerSetupSandbox).Method("GET", "POST")                      // ENC:
	muxEnc.HandleFunc("/api/get2FactorFromDeviceID", respHandlerGet2FactorFromDeviceID).Method("GET", "POST")   // ENC: return the One Time key when Device ID is presented. (2FA iOS Client)
	muxEnc.HandleFunc("/api/admin_set_password", respHandlerAdminSetPassword).Method("GET", "POST")             // ENC: As admin set password on other users
	muxEnc.HandleFunc("/api/admin_set_user_attrs", respHandlerAdminSetAttributes).Method("GET", "POST")         // ENC:
	muxEnc.HandleFunc("/api/admin_set_attrs", respHandlerAdminSetAttributes).Method("GET", "POST")              // ENC: - Old - depricate this end point for /api/admin_set_user_attrs
	muxEnc.HandleFunc("/api/admin_get_one_time_key", respHandlerAdminGetOneTimeKey).Method("GET", "POST")       // ENC: admin with privilates can get a one time key for another user.
	muxEnc.HandleFunc("/api/enc_version", respHandlerVersion).Method("GET", "POST")                             // ENC: for test of encrypted calls/returns - version of the srp-aes implementation
	muxEnc.HandleFunc("/api/resumeLogin", respHandlerResumeLogin).Method("GET", "POST")                         // ENC: for resumption of a previously logged in session -- new --
	muxEnc.HandleFunc("/api/setupStayLoggedIn", respHandlerSetupStayLoggedIn).Method("GET", "POST")             // ENC:	allow for stayLoggedIn user  -- new --
	muxEnc.HandleFunc("/api/srp_simulate_email_confirm", respHandlerSimulateEmailConfirm).Method("GET", "POST") // If in test mode can do a confirm of user

	muxEnc.HandleErrors(http.StatusNotFound, httpmux.HandlerFunc(errorHandlerFunc))

	// xyzzyDeviceID - management calls for multiple DeviceID's - list of them, delete one, add one, update name on DeviceID

	return
}

// ----------------------------------------------------------------------------------------------------------------------------
// Inject "$username$=anon-user", "$email$=", "$user_id$=, $is_logged_in$=n etc. -- InjectDefaults -- Middleware

// Overlap with ../cfg/cfg.io-1:w: var ReservedItems = map[string]bool{
var ReservedIDs = map[string]bool{
	"$auth_key$":                  true,
	"$email$":                     true,
	"$$host_name$$":               true,
	"$is_logged_in$":              true,
	"$is_enc_logged_in$":          true,
	"$is_anon_user$":              true,
	"$is_full_login$":             true,
	"$privs$":                     true,
	"$saved_one_time_key_hashed$": true,
	"$user_id$":                   true,
	"$username$":                  true,
	"LoginAuthCookie":             true,
	"Method":                      true,
	"URL":                         true,
	"owner_email":                 true,
	"user_etag":                   true,
	"username":                    true,
}

//	"fingerprint":                 true,

// xyzzyEEE
var ApiIn2faPendingMode = map[string]bool{
	"/api/1x1.gif":                    true,
	"/api/cipher":                     true,
	"/api/confirm-registration":       true,
	"/api/enc_version":                true,
	"/api/get2FactorFromDeviceID":     true,
	"/api/getPageToken":               true,
	"/api/pwrecov2":                   true,
	"/api/send_support_message":       true,
	"/api/setDebugFlags":              true,
	"/api/setup_sandbox":              true,
	"/api/srp_challenge":              true,
	"/api/srp_email_confirm":          true,
	"/api/srp_getNg":                  true,
	"/api/srp_login":                  true,
	"/api/srp_logout":                 true,
	"/api/srp_recover_password_pt1":   true,
	"/api/srp_recover_password_pt2":   true,
	"/api/srp_register":               true,
	"/api/srp_simulate_email_confirm": true,
	"/api/srp_validate":               true,
	"/api/valid2Factor":               true,
	"/api/version":                    true,
	"/api/resumeLogin":                true,
}

var AdminReservedIDs = map[string]bool{
	"salt": true,
	"v":    true,
}

// ----------------------------------------------------------------------------------------------------------------------------
// Get *current* DeviceID
//
// xyzzyDeviceID - return array of DeviceID's --
// Should return list of Current DeviceID's valid for this login -- Need a 2nd call to *create* a new DeviceID, only (see below)
// create a device if if NONE is assocaited with the login.
//
// Called after registration of a new user.
// DeviceID should could be sent back with registration.
// Requires login 't' registration token
// Should be a fully encrypted message
//
func respHandlerGetDeviceID(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1000, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	tt := ps.ByNameDflt("t", "")
	if tt == "" {
		AnError(hdlr, www, req, 400, 1001, "Invalid input data.")
		return
	}

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	email, err := UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
	if err != nil { // check user exists
		AnError(hdlr, www, req, 400, 1002, "Failed to find user. Invalid input email.")
		return
	}

	mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email))
	if !ok {
		AnError(hdlr, www, req, 400, 1003, fmt.Sprintf(`Unable to find account with email '%s`, email))
		return
	}

	DeviceID := mdata["DeviceID"] // old
	deviceIDList, ok := mdata["DeviceIDList"]
	DeviceIDList := make([]DeviceIDType, 0, 10)
	genOne := func() {
		deviceIDList = "[]"
		DeviceID = GenerateRandomDeviceID()
		genDate := time.Now().Format(time.RFC3339) //
		DeviceIDList = append(DeviceIDList, DeviceIDType{
			DeviceID:     DeviceID,
			CreationDate: genDate,
		})
		mdata["DeviceID"] = DeviceID // The device id for one time keys
		deviceIDList = godebug.SVar(DeviceIDList)
		mdata["DeviceIDList"] = deviceIDList
		dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata) // registration with "email" as username
	}
	if !ok {
		genOne()
	} else {
		err = json.Unmarshal([]byte(deviceIDList), &DeviceIDList)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%sInvalid device id list [%s], error=%s, %s %s\n", MiscLib.ColorRed, deviceIDList, err, godebug.LF(), MiscLib.ColorReset)
			genOne()
		}
	}

	// xyzzyDeviceID - need to mark srp:U:<<OldDeviceID>> as invalid - in a way that user can see on request
	// "Pre2Factor":               { "type":[ "string" ], "default":"p2f:" },
	// DbDel(hdlr, rw, SandBoxKey(hdlr.Pre2Factor, SandBoxPrefix, OldDeviceID)) // Stop any pending logins using old device id.

	// replace device ID -- why?
	// DeviceID := GenerateRandomDeviceID()
	// mdata["DeviceID"] = DeviceID                                                     // The device id for one time keys
	// dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata) // registration with "email" as username

	fmt.Fprintf(www, `{"status":"success","DeviceID":%q,"DeviceIDList":%s}`, DeviceID, godebug.SVar(DeviceIDList))
}

// ----------------------------------------------------------------------------------------------------------------------------
// Create DeviceID - set as *current* DeviceID and return it.
//
// Called after login if user wants to change device.
// DeviceID should could be sent back.
// Requires login 't' registration token
// Should be a fully encrypted message
//
func respHandlerCreateNewDeviceID(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1004, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	tt := ps.ByNameDflt("t", "")
	if tt == "" {
		AnError(hdlr, www, req, 400, 1005, "Invalid input data.")
		return
	}

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")
	Title := ps.ByNameDflt("Title", "")

	email, err := UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
	if err != nil { // check user exists
		AnError(hdlr, www, req, 400, 1006, "Failed to find user. Invalid input email.")
		return
	}

	mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email))
	if !ok {
		AnError(hdlr, www, req, 400, 1007, fmt.Sprintf(`Unable to find account with email '%s`, email))
		return
	}

	// xyzzyDeviceID - return 1 DeviceID's -- - new created and appended to list of valid DeviceID's
	DeviceID := mdata["DeviceID"] // old
	deviceIDList, ok := mdata["DeviceIDList"]
	DeviceIDList := make([]DeviceIDType, 0, 10)
	err = json.Unmarshal([]byte(deviceIDList), &DeviceIDList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sInvalid device id list [%s], error=%s %s\n", MiscLib.ColorRed, deviceIDList, err, MiscLib.ColorReset)
	}
	// 5. Limit # of device IDs for a user (config) --  "LimitDeviceIDs":           { "type":[ "int" ], "default":"10" },
	if hdlr.LimitDeviceIDs > 0 && len(DeviceIDList)+1 >= hdlr.LimitDeviceIDs {
		AnError(hdlr, www, req, 403, 1008, fmt.Sprintf(`Excessive number of DeviceIDs - limit %d`, hdlr.LimitDeviceIDs))
		return
	}
	DeviceID = GenerateRandomDeviceID()
	genDate := time.Now().Format(time.RFC3339) //
	DeviceIDList = append(DeviceIDList, DeviceIDType{
		DeviceID:     DeviceID,
		CreationDate: genDate,
		Title:        Title,
	})
	mdata["DeviceID"] = DeviceID // The device id for one time keys
	deviceIDList = godebug.SVar(DeviceIDList)
	mdata["DeviceIDList"] = deviceIDList
	dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata) // registration with "email" as username

	fmt.Fprintf(www, `{"status":"success","DeviceID":%q,"DeviceIDList":%s}`, DeviceID, godebug.SVar(DeviceIDList))
}

// ----------------------------------------------------------------------------------------------------------------------------
// muxEnc.HandleFunc("/api/updateDeviceID", respHandlerUpdateDeviceID).Method("GET", "POST")                   // ENC:
func respHandlerUpdateDeviceID(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1009, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	tt := ps.ByNameDflt("t", "")
	if tt == "" {
		AnError(hdlr, www, req, 400, 1010, "Invalid input data.")
		return
	}
	Title := ps.ByNameDflt("Title", "")
	DeviceID := ps.ByNameDflt("DeviceID", "")

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	email, err := UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
	if err != nil { // check user exists
		AnError(hdlr, www, req, 400, 1011, "Failed to find user. Invalid input email.")
		return
	}
	// UserName := email

	mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email))
	if !ok {
		AnError(hdlr, www, req, 400, 1012, fmt.Sprintf(`Unable to find account with email '%s`, email))
		return
	}

	// xyzzyDeviceID - return 1 DeviceID's -- - new created and appended to list of valid DeviceID's
	deviceIDList, ok := mdata["DeviceIDList"]
	DeviceIDList := make([]DeviceIDType, 0, 10)
	err = json.Unmarshal([]byte(deviceIDList), &DeviceIDList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sInvalid device id list [%s], error=%s %s\n", MiscLib.ColorRed, deviceIDList, err, MiscLib.ColorReset)
	}
	ok = false
	for ii, di := range DeviceIDList {
		if di.DeviceID == DeviceID {
			ok = true
			di.Title = Title
			DeviceIDList[ii] = di
			break
		}
	}
	if !ok {
		AnError(hdlr, www, req, 4111, 1013, "Failed to find DeviceID. Invalid input DeviceID.")
		return
	}
	deviceIDList = godebug.SVar(DeviceIDList)
	mdata["DeviceIDList"] = deviceIDList
	dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata) // registration with "email" as username

	fmt.Fprintf(www, `{"status":"success","DeviceIDList":%s}`, godebug.SVar(DeviceIDList))
}

// ----------------------------------------------------------------------------------------------------------------------------
// muxEnc.HandleFunc("/api/deleteDeviceID", respHandlerDeleteDeviceID).Method("GET", "POST")                   // ENC:
func respHandlerDeleteDeviceID(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1014, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	tt := ps.ByNameDflt("t", "")
	if tt == "" {
		AnError(hdlr, www, req, 400, 1015, "Invalid input data.")
		return
	}

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	email, err := UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
	if err != nil { // check user exists
		AnError(hdlr, www, req, 400, 1016, "Failed to find user. Invalid input email.")
		return
	}
	DeviceID := ps.ByNameDflt("DeviceID", "")

	mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email))
	if !ok {
		AnError(hdlr, www, req, 400, 1017, fmt.Sprintf(`Unable to find account with email '%s`, email))
		return
	}

	deviceIDList, ok := mdata["DeviceIDList"]
	DeviceIDList := make([]DeviceIDType, 0, 10)
	NewDeviceIDList := make([]DeviceIDType, 0, 10)
	err = json.Unmarshal([]byte(deviceIDList), &DeviceIDList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sInvalid device id list [%s], error=%s %s\n", MiscLib.ColorRed, deviceIDList, err, MiscLib.ColorReset)
	}
	for _, di := range DeviceIDList {
		if di.DeviceID == DeviceID {
		} else {
			NewDeviceIDList = append(NewDeviceIDList, di)
		}
	}
	DbDel(hdlr, rw, SandBoxKey(hdlr.Pre2Factor, SandBoxPrefix, DeviceID)) // if a key exists right now - then delete it
	NewCreated := "n"
	if len(NewDeviceIDList) == 0 {
		NewCreated = "y"
		DeviceID = GenerateRandomDeviceID()
		genDate := time.Now().Format(time.RFC3339) //
		NewDeviceIDList = append(NewDeviceIDList, DeviceIDType{
			DeviceID:     DeviceID,
			CreationDate: genDate,
		})
		mdata["DeviceID"] = DeviceID // The device id for one time keys
	}
	deviceIDList = godebug.SVar(NewDeviceIDList)
	mdata["DeviceID"] = NewDeviceIDList[0].DeviceID
	mdata["DeviceIDList"] = deviceIDList
	dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata) // registration with "email" as username

	fmt.Fprintf(www, `{"status":"success","DeviceID":%q,"NewCreated":%q,"DeviceIDList":%s}`, DeviceID, NewCreated, godebug.SVar(DeviceIDList))
}

// ----------------------------------------------------------------------------------------------------------------------------
func respHandlerVersion(www http.ResponseWriter, req *http.Request) {
	www.Header().Set("Content-Type", "application/json") // For JSON data
	// fmt.Fprintf(www, `{"status":"success","msg":"SRP 6a, AES-256 RESTful encryption (part of Go-FTL 0.5.10)","version":"1.0.2","BuildDate":"Thu Mar 31 14:55:27 MDT 2016"}`)
	fmt.Fprintf(www, `{"status":"success","msg":"SRP 6a, AES-256 RESTful encryption (part of Go-FTL 0.5.10)","version":"1.0.2","BuildDate":"Wed Sep 28 12:11:01 MDT 2016"}`)
}

// ----------------------------------------------------------------------------------------------------------------------------
// mux.HandleFunc("/api/genTempKeys", respHandlerGenTempKeys).Method("GET", "POST")
// Input:
//		't' - the session key
// Should be a fully encrypted message
func respHandlerGenTempKeys(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1018, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	tt := ps.ByNameDflt("t", "")
	if tt == "" {
		AnError(hdlr, www, req, 400, 1019, "Invalid input data.")
		return
	}

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	email, err := UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
	if err != nil { // check user exists
		AnError(hdlr, www, req, 400, 1020, "Failed to find user. Invalid input email.")
		return
	}

	new20, hashed, DeviceID, deviceIDList := "", "", "", ""
	mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email))
	if !ok {
		AnError(hdlr, www, req, 400, 1021, fmt.Sprintf(`Unable to find account with email '%s`, email))
		return
	}
	salt, _ /*v*/ := GetSalt(hdlr, www, req, mdata)
	if mdata["acct_type"] == "DeviceID" {
		new20, hashed = GenBackupKeys(hdlr, salt, "4", www, req)
		mdata["offline_one_time_keys"] = hashed
		DeviceID = ""
		deviceIDList = "[]"
	} else {
		new20, hashed = GenBackupKeys(hdlr, salt, "9", www, req)
		mdata["backup_one_time_keys"] = hashed
		DeviceID = mdata["DeviceID"]
		deviceIDList = mdata["DeviceIDList"]
	}
	dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)

	// xyzzyDeviceID - return array of DeviceIDs
	fmt.Fprintf(www, `{"status":"success","BackupKeys":%q,"DeviceID":%q, "DeviceIDList":%s}`, new20, DeviceID, deviceIDList)
}

// ----------------------------------------------------------------------------------------------------------------------------
// mux.HandleFunc("/api/send_support_message", respHandlerSendSupportMessage).Method("GET", "POST")
// Should be a fully encrypted message
func respHandlerSendSupportMessage(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1022, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	email := hdlr.SupportEmailTo
	email_fr := ps.ByNameDflt("email", "")
	if !ValidateEmail(email_fr) {
		AnError(hdlr, www, req, 400, 1023, "Invalid input email.")
		return
	}

	subject := ps.ByNameDflt("subject", "")
	if subject == "" {
		AnError(hdlr, www, req, 400, 1024, "Invalid input email subject.")
		return
	}

	body := ps.ByNameDflt("body", "")
	if body == "" {
		AnError(hdlr, www, req, 400, 1025, "Invalid input email body.")
		return
	}

	// check frequencey - save data to file and just end a notice that data is available.
	SaveSupportMessage(rw, email_fr, subject, body)

	// Input: email, subject, body
	// Validate
	// func SendEmailViaAWS(email_addr string, app string, tmpl string, pw string, email_auth_token string) {
	go SendEmailViaAWS_support(hdlr, email, hdlr.EmailApp, "support-message.tmpl", email_fr, subject, body)

	io.WriteString(www, `{"status":"success"}`)
}

const cfg_AllowReactNativeTest = "SecuredClient"

// ----------------------------------------------------------------------------------------------------------------------------
// Input
//	DeviceID
//	t - from login - session key
// Output
//	OneTimeKey - Key to use for login
func respHandlerGet2FactorFromDeviceID(www http.ResponseWriter, req *http.Request) {

	// DeviceID -> URL + User(email) + Sandbox?

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1026, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	fmt.Printf("\n!!!! In Function !!!!! Params + Cookies for (%s): %s AT %s\n", req.URL.Path, rw.Ps.DumpParamTable(), godebug.LF())

	DeviceID := ps.ByNameDflt("DeviceID", "")
	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	fmt.Printf("respHandlerGet2FactorFromDeviceID: DeviceID=%s, %s\n", DeviceID, godebug.LF())

	if !verhoeff_algorithm.ValidateVerhoeff(DeviceID) {
		AnError(hdlr, www, req, 400, 1027, "DeviceID is not valid.")
		return
	}

	// xyzzySecDeviceID -- section start --
	if cfg_AllowReactNativeTest == "SecuredClient" {

		tt := ps.ByNameDflt("t", "")
		if tt == "" {
			AnError(hdlr, www, req, 400, 1028, `Invalid input data.`)
			return
		}

		// fmt.Printf("AT: %s\n", godebug.LF())

		email, err := UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
		if err != nil { // check user exists
			AnError(hdlr, www, req, 400, 1029, `Failed find existng user.  Invalid email/username.`)
			return
		}

		fmt.Printf("Email=%s\n", email)

		mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email))
		if !ok {
			AnError(hdlr, www, req, 400, 1030, fmt.Sprintf("Unable to find account with email/username= '%s'.", email))
			return
		}

		if mdata["acct_type"] == "DeviceID" {
			fmt.Printf("\nLegetimate DeviceID login for getting 2fa one time key based on device id, %s\n\n", godebug.LF())
		} else {
			fmt.Printf("\nSecDeviceID: Rejected - not a DeviceID login, goodbye! %s\n\n", godebug.LF())
			AnError(hdlr, www, req, 400, 1031, fmt.Sprintf("Secuirty Error: not a legitimae end point for a DeviceID type account\n\n"))
			return
		}

		/*
			if godebug.InArrayString("MayChangeOtherAttributes", hdlr.SecurityPrivilages["admin"]) < 0 { // if not found
				AnError(hdlr, www, req, 400, 1032, "Admin missing privilage 'MayChangeOtherAttributes'.")
				return
			}
			// hdlr.SecurityConfig           SecurityConfigType          //
		*/
		//if len(hdlr.SecurityConfig.MayAccessApi[mdata["acct_type"]]) > 0 && (hdlr.SecurityConfig.MayAccessApi[mdata["acct_type"]][0] == "*" ||
		//	godebug.InArrayString("/api/get2FactorFromDeviceID", hdlr.SecurityConfig.MayAccessApi[mdata["acct_type"]]) < 0) { // if not found
		if !CheckMayAccessApi(hdlr, rw, SandBoxPrefix, "/api/get2FactorFromDeviceID", "y", mdata["acct_type"]) {
			// fmt.Printf("\nxyzzySecDeviceID: Should reject - legitimae end point for a %s type account\n\n", mdata["acct_type"])
			AnError(hdlr, www, req, 400, 1033, fmt.Sprintf("Secuirty Error: not a legitimae end point for a %s type account\n\n", mdata["acct_type"]))
			return
		} // else {
		//	fmt.Printf("\nxyzzySecDeviceID: OK - passed MayAccessApi check\n\n")
		// }
	}
	// xyzzySecDeviceID -- section end --

	// From end of Login if Successful
	// 		DbSetExpire(hdlr, rw, SandBoxKey(hdlr.Pre2Factor, SandBoxPrefix, DeviceID), OneTimeKey, hdlr.TwoFactorLife)	// p2f:DeviceID -> OneTimeKey
	// example of what gets returned -- Just a OneTimeKey - that is the entire string

	// Created in respHandlerSRPValidate
	// "Pre2Factor":               { "type":[ "string" ], "default":"p2f:" },
	OneTimeKey, err := DbGetString(hdlr, rw, SandBoxKey(hdlr.Pre2Factor, SandBoxPrefix, DeviceID)) // p2f:DeviceID -> OneTimeKey
	if err != nil {
		AnError(hdlr, www, req, 400, 1034, `Unable to retreive a key based on this DeviceID`)
		return
	}

	fmt.Printf("\n\nLookup=%s, DeviceID=%s OneTimeKey = %q\n\n", SandBoxKey(hdlr.Pre2Factor, SandBoxPrefix, DeviceID), DeviceID, OneTimeKey)

	// Move to 2nd Access-Control-Allow-Origin middleware - allow on requests -
	//	if h := req.Header.Get("Origin"); h != "" {
	//		www.Header().Set("Access-Control-Allow-Origin", h) // Allow requests from any server for this.
	//	} else {
	//		www.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any server for this.
	//	}

	www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	www.Header().Set("Expires", "0")                                         // Proxies.

	www.WriteHeader(200)
	www.Header().Set("Content-Type", "application/json") // For JS
	fmt.Fprintf(www, `{"status":"success","version":1,"OneTimeKey":%q}`, OneTimeKey)
}

// --------------------------------------------------------------------------------------------------------------------------------------
//
//	muxEnc.HandleFunc("/api/valid2Factor", respHandlerValid2Factor).Method("GET", "POST")                     // ENC: validate a one-time key - uses 't'/user_id to validate OneTimeKey
//
// Input:
//	t=				Session key
//	OneTime	The one time key
func respHandlerValid2Factor(www http.ResponseWriter, req *http.Request) {
	// Input "t" + OneTimeKey -- In encrypted data.
	// Check and possibly delete it.

	// Not in 2fa mode - just return success.
	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1035, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	if hdlr.TwoFactorRequired == "n" {
		io.WriteString(www, `{"status":"success"}`)
		return
	}

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")
	OneTime := ps.ByNameDflt("OneTime", "")
	OneTime = strings.Replace(OneTime, " ", "", -1)
	if OneTime == "" {
		AnError(hdlr, www, req, 400, 1036, `Invalid input data.`)
		return
	}

	tt := ps.ByNameDflt("t", "")
	if tt == "" {
		AnError(hdlr, www, req, 400, 1037, `Invalid input data.`)
		return
	}

	email, err := UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
	if err != nil { // check user exists
		AnError(hdlr, www, req, 400, 1038, `Failed find existng user.  Invalid email`)
		return
	}

	mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email))
	if !ok {
		AnError(hdlr, www, req, 400, 1039, fmt.Sprintf("Unable to find account with email '%s'.", email))
		return
	}

	// mdata["auth"] = "P"		// Pending 2Fa validation
	if mdata["auth"] != "P" {
		fmt.Printf("\n\n************* Not P in auth - why?, got %s, %s\n\n\n", mdata["auth"], godebug.LF())
		// xyzzyEEE - should exit at this point - fatal error
		// var ApiIn2faPendingMode = map[string]bool{
	}

	switch OneTime[0:1] { // One time validated to not be "" above (after removal of blanks)
	case "4":
		// case "9" - Backup Key
		salt, _ /*v*/ := GetSalt(hdlr, www, req, mdata)
		set := mdata["offline_one_time_keys"]
		newKeys, found, _ := CmpBackupKeys(hdlr, salt, set, OneTime) // compare to backup keys	// "salt2"
		mdata["offline_one_time_keys"] = newKeys
		mdata["auth"] = "P" // Pending 2Fa validation
		if found {
			mdata["auth"] = "y" // success
			mdata["$saved_one_time_key_hashed$"] = OneTime
			dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)
			io.WriteString(www, `{"status":"success"}`)
			return
		}
		dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)

	case "9":
		// case "9" - Backup Key
		salt, _ /*v*/ := GetSalt(hdlr, www, req, mdata)
		set := mdata["backup_one_time_keys"]
		newKeys, found, _ := CmpBackupKeys(hdlr, salt, set, OneTime) // compare to backup keys	// "salt2"
		mdata["backup_one_time_keys"] = newKeys
		mdata["auth"] = "P" // Pending 2Fa validation
		if found {
			mdata["auth"] = "y" // success
			mdata["$saved_one_time_key_hashed$"] = OneTime
			dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)
			io.WriteString(www, `{"status":"success"}`)
			return
		}
		dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)

	case "8", "7", "6", "5":
		// case "8" - Normal device iOS key
		// case "7" - Twillo SMS Push - google message push ( soon - not implemented yet ) - Android
		// case "6" - Voice Message to Chat with Support, phone call etc.	(not implemented yet)
		// case "5" - Push Message ( other push message - not implemented yet )
		// Compare to generated one time key
		// SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")
		// DbSetExpire(hdlr, rw, SandBoxKey("srp:U:", cookieValue, "bob@example.com"), bob, 24*60*60*32)
		// "Pre2Factor":               { "type":[ "string" ], "default":"p2f:" },
		key := SandBoxKey(hdlr.Pre2Factor, SandBoxPrefix, OneTime)
		mdata["auth"] = "P" // Pending 2Fa validation
		fmt.Printf("key for 2fa lookup [%s], %s\n", key, godebug.LF())
		s, err := DbGetString(hdlr, rw, key)
		if err != nil {
			AnError(hdlr, www, req, 400, 1040, "Invalid one time key")
			return
		}
		if s == email {
			fmt.Printf("Auth Set to 'y' *** At: %s\n", godebug.LF())
			mdata["auth"] = "y" // Pending 2Fa validation
			dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)
			//DbDel(hdlr, rw, key)		// should OTK be left for a 2nd/3rd try - or deleted on success? -- Key will timeout in 5 anyhow.
			//key2 := SandBoxKey(hdlr.Pre2Factor, SandBoxPrefix, mdata["DeviceID"])
			//DbDel(hdlr, rw, key)
			io.WriteString(www, `{"status":"success"}`)
			return
		}
		dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)
		DbDel(hdlr, rw, key)
	}

	AnError(hdlr, www, req, 400, 1041, "Invalid one time key")
}

// ----------------------------------------------------------------------------------------------------------------------------
// Create a sandbox for demos.
func respHandlerSetupSandbox(www http.ResponseWriter, req *http.Request) {

	www.Header().Set("Content-Type", "application/json")                     // For JSON data
	www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	www.Header().Set("Expires", "0")                                         // Proxies.

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1042, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}

	// Set Cookie at this point
	id0, _ := uuid.NewV4()
	cookieValue := id0.String()
	cookieValue = cookieValue[0:8]
	expire := time.Now().AddDate(0, 0, 32) // Years, Months, Days==2
	secure := false
	if req.TLS != nil {
		secure = true
	}
	cookie := http.Cookie{Name: "GOFTL_Sandbox", Value: cookieValue, Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400, Secure: secure, HttpOnly: false}
	http.SetCookie(www, &cookie)

	// xyzzyDeviceID - Update "bob"

	bob_salt := "6e68106acec25a3376915e422093a7"
	bob_v := "78881e15d9612559a8e9f5086f5565ba601ef9d4552028ee206bb50a91e62fea3fb85b2874b3e65480cc85e5c95ad176d64ecedec0b57f134e9f7209f41aabbb723f2a832b763aa6a616befb5539fd861ea70294655cd8baa0ba4340cd3f6e43a6532c2a5073ee2154dd04c543c7e00922ddd8784d3342c124887a062dd5202dc5e559b75fafd5a36d462716a6f811d8e17f21788cefa3818a1a7a076e7d4d0745f5838a6bdb7a1bee5e6590bb4ff53ee8f009b02a7aa0b1f1bc615fd578ee8d92f0a7b5f4bde3a9498d734a601f99e5b03c4595caa0bc28ad7f8ca34ee9916a68566c5e6c4d27c80c8e15e8d18d50cffcfbfea77b8e164d2437d3d196d72994"

	if hdlr.PasswordSV == "" {
	} else {
		db_genKey(hdlr)
		bob_salt = db_encrypt(hdlr.passwordSVKey, bob_salt)
		bob_v = db_encrypt(hdlr.passwordSVKey, bob_v)
	}

	// xyzzyDeviceID - Update "bob"
	bob := fmt.Sprintf(`{"DeviceID":"93938897","auth":"y","backup_one_time_keys":"522b5305b8aa73807bb31acb8a6ee826,778e2b14702db0ba7477cf0d47133210,92cfe5fd24c4d6d13e74e4681dc4ac8f,7bf557bbe124c141f5b731942eae472a,d2ccdccc6385a9c14520284cfc9399e1,236d9549eaba314c9c998f225929ea70,73d42e29e70b197f3167c383dbe50fbd,395f71f45d4da571d9753a7efca86b32,c2102ea3e90001492ca1b4276bf0dcd8,e98e21d4c4ed7f5f2ab4f36be051c534,4c8f1337ecfae90dd54434d6ca8d6ad1,352a68b5fa6353885d6f5e06cfa6446c,08f98aed8aed1ae337c0d22c18f50f86,823bdf5a219472d4af5c2c5ee364c46e,6c02f27651d3ad73d10da4c3722d1dc3,4a0f59772f3fe6950e7f96107a2c81c9,9159bb08fe0913ea51e8f0d2376094fe,3617b23c8fcba36249a59659842c2f95,4ba6a6926a8997c07b23955a06504e04,e792deb7f1cb006d2a969fa78cd40f64","confirmed":"y","disabled":"n","disabled_reason":"","login_date_time":"2015-12-15T22:10:16-07:00","login_fail_time":"","n_failed_login":"0","num_login_times":"1","privs":"user","register_date_time":"2015-12-15T22:10:16-07:00","salt":%q,"v":%q}`, bob_salt, bob_v)

	DbSetExpire(hdlr, rw, SandBoxKey("srp:U:", cookieValue, "bob@example.com"), bob, hdlr.SandBoxExpreTime)

	jane_salt := "dd7d11848c3228468ab6eef3debbf2a"
	jane_v := "2a457a1cddbf34130318e47c77c6080aa5d59e467518f77a9c5bb905046a1bc56a6de3ef003e9658d9cfb71ef97dd05eb74e7472affe673e738879a4ce4ac6bc92c5b388d351a0416a5c36373e9c7c7b16a972a66aaa63bf1f92041521b92709c259c700c66c9bedbba12a150ceb9b1221f0ee5fe3777efe423774bd02602c89d1e95aced5531e2bd4ae2d52ca59e0862ffbced3cf2d14fef82178e54781eb8237bab92ed155c09fa50b1c9ae07c4f9bb356ab1ccd3896a409ba94715428eb6b2039a4b88742cea2cdc149f3f00a5304fbbbbf13895a8422ef9cb411268ed42bd531f7c5cf00a5a052de1db64682c6749ba8ca63c9fc159c81103ffed04ea25b"

	if hdlr.PasswordSV == "" {
	} else {
		db_genKey(hdlr)
		jane_salt = db_encrypt(hdlr.passwordSVKey, jane_salt)
		jane_v = db_encrypt(hdlr.passwordSVKey, jane_v)
	}

	// xyzzyDeviceID - Update "bob"
	jane := fmt.Sprintf(`{"DeviceID":"53237551","auth":"y","backup_one_time_keys":"b0f7c5e55d93cdf00019e892b23ef6fc,0f78cb7754d6b05d00935b7a3a1d859c,7a0068dc1f6f79fa483f7d1eef665018,360b2761b1cfe7515af295105989383b,14eebb5546eaded2d1546bd5a9ab03e8,69e9eb6cc8d1f9d0a1b2d8eb51fe2451,f7e92663da8b0652122e8f0217281db4,dddc8449d29fe56b7c87c7236566161f,d7d192aa4719ad1f102a711a34677f7c,95c2faf6b7bc3cd79a5dd1a3a716aa0c,c773d4990e6f74371fffa415eed92817,ca13ac3eac3e09c993d8040fde3d133a,a3a390a2e5727380db1846f451ccf997,d8f27773fc6f48442bdf04732d4c9faf,b9bcd1073284f8fa0023a3c0a21b6ebd,b837332197eb153721ab6e71a96c2a75,679520ef6320eaf64ada9a25c97e935a,353c3eacad56d11e889f107d949836d0,4e9a62c698158f080b498a27f0ef30d6,f07056983113fd70aa45e06920c6ca54","confirmed":"y","disabled":"n","disabled_reason":"","login_date_time":"2015-12-15T22:20:31-07:00","login_fail_time":"","n_failed_login":"0","num_login_times":"1","privs":"admin","register_date_time":"2015-12-15T22:20:31-07:00","salt":%q,"v":%q}`, jane_salt, jane_v)

	DbSetExpire(hdlr, rw, SandBoxKey("srp:U:", cookieValue, "jane@example.com"), jane, hdlr.SandBoxExpreTime)

	fmt.Fprintf(www, `{"status":"success","SandBox":%q}`, cookieValue)
}

// ----------------------------------------------------------------------------------------------------------------------------
func respHandlerConfirmRegistration(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1043, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	https := "http://"
	if req.TLS != nil {
		https = "https://"
	}

	email_auth_token := ps.ByNameDflt("auth_token", "")
	if email_auth_token == "" {
		www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
		www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
		www.Header().Set("Expires", "0")                                         // Proxies.
		// www.Header().Set("Location", https+req.Host+"/unable-to-register1.html")
		data := make(map[string]string)
		data["HTTPS"] = https
		data["HOST"] = req.Host // this is host and port
		To := tmplp.ExecuteATemplate(hdlr.RegTemplate1, data)
		www.Header().Set("Location", To)
		www.WriteHeader(307)
		return
	}
	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	email, ok := GetEmailAuth(hdlr, rw, email_auth_token, SandBoxPrefix)
	if !ok {
		// fmt.Printf("email_auth_token [%s]\n", email_auth_token)
		www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
		www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
		www.Header().Set("Expires", "0")                                         // Proxies.
		// www.Header().Set("Location", https+req.Host+"/unable-to-register2.html")
		data := make(map[string]string)
		data["HTTPS"] = https
		data["HOST"] = req.Host // this is host and port
		To := tmplp.ExecuteATemplate(hdlr.RegTemplate2, data)
		www.Header().Set("Location", To)
		www.WriteHeader(307)
		return
	}

	if mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email)); !ok {
		www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
		www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
		www.Header().Set("Expires", "0")                                         // Proxies.
		// www.Header().Set("Location", https+req.Host+"/unable-to-register2.html")
		data := make(map[string]string)
		data["HTTPS"] = https
		data["HOST"] = req.Host // this is host and port
		To := tmplp.ExecuteATemplate(hdlr.RegTemplate3, data)
		www.Header().Set("Location", To)
		www.WriteHeader(307)
		return
	} else {
		mdata["confirmed"] = "y"
		mdata["num_login_times"] = "0"
		dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)
	}

	www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	www.Header().Set("Expires", "0")                                         // Proxies.
	// www.Header().Set("Location", https+req.Host+"/#/login")
	data := make(map[string]string)
	data["HTTPS"] = https
	data["HOST"] = req.Host // this is host and port
	To := tmplp.ExecuteATemplate(hdlr.RegTemplate4, data)
	www.Header().Set("Location", To)
	www.WriteHeader(307)

	io.WriteString(www, `{"status":"success"}`)
}

// ----------------------------------------------------------------------------------------------------------------------------
// http://www.2c-why.com/pwrecov2?auth_token={{.p2}}
//	mux.HandleFunc("/api/pwrecov2", respHandlerRecoverPw2).Method("GET")                                     // Redirect-Link: To /#/pwrecov2 with cookies set
func respHandlerRecoverPw2(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1044, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	https := "http://"
	if req.TLS != nil {
		https = "https://"
	}

	email_auth_token := ps.ByNameDflt("auth_token", "")
	if email_auth_token == "" {
		// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
		// DID NOT get a token passed - template for error to user.
		// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
		www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
		www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
		www.Header().Set("Expires", "0")                                         // Proxies.
		// www.Header().Set("Location", https+req.Host+"/unable-to-pwrecov1.html")
		data := make(map[string]string)
		data["HTTPS"] = https
		data["HOST"] = req.Host // this is host and port
		To := tmplp.ExecuteATemplate(hdlr.PwRecoverTemplate1, data)
		www.Header().Set("Location", To)
		www.WriteHeader(307)
		return
	}

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	// Old
	// Note: -- DbSetExpire(hdlr, rw, SandBoxKey(hdlr.PwResetKey, SandBoxPrefix, email_auth_token), email, hdlr.PwExpireIn)
	// email, ok := GetEmailAuth(hdlr, rw, email_auth_token, SandBoxPrefix) // eau:Token
	// if !ok {

	email, err := DbGetString(hdlr, rw, SandBoxKey(hdlr.PwResetKey, SandBoxPrefix, email_auth_token))
	if err != nil {
		// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
		// Invalid EMAIL - not an account - error to user template // www.Header().Set("Location", https+req.Host+"/unable-to-pwrecov2.html")
		// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
		// fmt.Printf("email_auth_token [%s]\n", email_auth_token)
		www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
		www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
		www.Header().Set("Expires", "0")                                         // Proxies.
		data := make(map[string]string)
		data["HTTPS"] = https
		data["HOST"] = req.Host // this is host and port
		To := tmplp.ExecuteATemplate(hdlr.PwRecoverTemplate2, data)
		www.Header().Set("Location", To)
		www.WriteHeader(307)
		return
	}

	fmt.Printf("/api/pwrecov2 - email[%s], %s\n", email, godebug.LF())

	if mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email)); !ok {
		// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
		// Invalid EMAIL - error to user -- investigate this // www.Header().Set("Location", https+req.Host+"/unable-to-pwrecov2.html")
		// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
		www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
		www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
		www.Header().Set("Expires", "0")                                         // Proxies.
		data := make(map[string]string)
		data["HTTPS"] = https
		data["HOST"] = req.Host // this is host and port
		To := tmplp.ExecuteATemplate(hdlr.PwRecoverTemplate2, data)
		www.Header().Set("Location", To)
		www.WriteHeader(307)
		return
	} else {
		mdata["confirmed"] = "y"
		mdata["num_login_times"] = "0"
		dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)
	}

	// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	// Success - redirect into appliation to non-login page with info to complete password reset process.
	// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	rn, _ := GenRandNumber(8)
	lmSalt := fmt.Sprintf("%x", rn) // string(hn)
	linkMarker := HashStrings.Sha256(lmSalt + ":" + email_auth_token + ":" + email)

	DbSetExpire(hdlr, rw, SandBoxKey("pm:", SandBoxPrefix, linkMarker), fmt.Sprintf(`{"marked":"link","email":%q,"salt":%q,"hash":%q}`, email, lmSalt, linkMarker), hdlr.PwExpireIn)

	expire := time.Now().AddDate(0, 0, hdlr.CookieExpireInXDays) // Years, Months, Days==2
	// xyzzy need expire for this in seconds
	cookie := http.Cookie{Name: "linkMarkerCookie", Value: linkMarker, Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400, Secure: false, HttpOnly: true}
	http.SetCookie(www, &cookie)

	//expire := time.Now().AddDate(0, 0, hdlr.CookieExpireInXDays) // Years, Months, Days==2 // Xyzzy501 - should be a config - on how long to keep cookie
	//cookie := http.Cookie{Name: "LoginAuthToken", Value: email_auth_token, Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400, Secure: false, HttpOnly: true}
	//http.SetCookie(www, &cookie)
	//cookie = http.Cookie{Name: "LoginAuthEmail", Value: email, Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400, Secure: false, HttpOnly: false}
	//http.SetCookie(www, &cookie)
	www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	www.Header().Set("Expires", "0")                                         // Proxies.
	// www.Header().Set("Location", https+req.Host+"/#/pwrecov2")
	data := make(map[string]string)
	data["HTTPS"] = https
	data["HOST"] = req.Host // this is host and port
	To := tmplp.ExecuteATemplate(hdlr.PwRecoverTemplate3, data)
	www.Header().Set("Location", To)
	www.WriteHeader(307)

	io.WriteString(www, `{"status":"success"}`)
}

// ----------------------------------------------------------------------------------------------------------------------------

func errorHandlerFunc(ww http.ResponseWriter, req *http.Request) {
	code := http.StatusForbidden
	fmt.Printf("Generating a %d status code at %s\n", code, godebug.LF())
	ww.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ww.Header().Set("X-Content-Type-Options", "nosniff")
	ww.Header().Set("X-Go-FTL-LineNo", godebug.LF())
	ww.WriteHeader(code)
	fmt.Fprintln(ww, "403 Forbidden") // 403
	// panic("ya")
}

// ============================================================================================================================================
// ============================================================================================================================================
// ============================================================================================================================================
// Original SRP/AES code following
// ============================================================================================================================================
// ============================================================================================================================================
// ============================================================================================================================================

var dataStore *RSaveToRedis
var emailRe *regexp.Regexp
var hexRe *regexp.Regexp
var isUUIDRe *regexp.Regexp
var ErrNoSuchUser = errors.New("User not found - no such user")

// ============================================================================================================================================
// Config --------------------------------------------------------------------------------
// var EmailRelayIP string               // Set during initialization
// var EmailAuthToken string             // Set during initialization
var TestModeInject = map[string]bool{ // Map of injected errors
	"invalid-tt":                 true,
	"invalid-tt-change-password": false,
}

// Example: if tt == "" || ( InjectionTestMode && TestModeInject["invalid-tt-change-password"] ) {

// Can only be change with re-compile
// const TestMode = true           // if true allow /api/srp_simulate_email_confirm to do an email confirm
// const InDemoMode = true         // Demo mode sends one time keys to user via registration email -- See also "SandBoxMode" -- and allows /api/srp_simulate_email_confirm to confirm email registration
const InjectionTestMode = false // Allows *ERROR* injection to test.
const SandBoxMode = true        // Allow use of a database sandbox for testing and demos

const db2 = false          // debug of nonce - used in ./gen_ran.go
const dbCipher = false     //
const dbCipher2 = false    //
const dbCipher3 = false    //
const dbCipher4 = false    //
const dbSRP = true         // SRP login debug flag
const db11 = true          // report if connected to Redis server - print on stdout
const db12 = true          // report if connected to Redis server - to log
const dbFingerprint = true // test fingerpint update/save call.
const dbDumpConfig = false // Dump out configuration for aessrp at start

// ============================================================================================================================================
// Return N, g values for SRP.
// Return Configuration for privilages and a flag indicating if 2fa is enabled.
// Default number for N, g taken from: https://www.ietf.org/rfc/rfc5054.txt
func respHandlerSRPGetNg(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1045, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	outputFmt := ps.ByNameDflt("fmt", "json") // json or js
	fingerprint := ps.ByNameDflt("fingerprint", "0")
	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	if dbRespHandlerSRPGetNg {
		fmt.Printf("\n\n")
		fmt.Printf("/////////////////////////////////////////////////////////////////////////////////////////////////////\n")
		fmt.Printf("// outputFmt = %s\n", outputFmt)
		fmt.Printf("// fingerprint = %s\n", fingerprint)
		fmt.Printf("/////////////////////////////////////////////////////////////////////////////////////////////////////\n")
		fmt.Printf("\n\n")
	}

	SecurityData := lib.SVar(hdlr.SecurityConfig)

	if dbRespHandlerSRPGetNg {
		fmt.Printf("SecurityData = >>>%s<<<, Key >>>%s<<< %s\n", SecurityData, SandBoxKey("srp:U:", SandBoxPrefix, ":bits"), godebug.LF())
	}

	if h := req.Header.Get("Origin"); h != "" {
		www.Header().Set("Access-Control-Allow-Origin", h) // Allow requests from any server for this.
	} else {
		www.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any server for this.
	}
	www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	www.Header().Set("Expires", "0")                                         // Proxies.
	switch outputFmt {
	case "json", "JSON":
		www.Header().Set("Content-Type", "application/json") // For JSON data
		fmt.Fprintf(www, `{"status":"success","bits":%d,"g":%q,"N":%q,"TwoFactorRequired":%q,"SecurityData":%s}`+"\n", hdlr.Bits, hdlr.NGData.G, hdlr.NGData.N, hdlr.TwoFactorRequired, SecurityData)
	case "js", "JS":
		www.Header().Set("Content-Type", "application/javascript") // For JS
		fmt.Fprintf(www, `;var Bits=%d;var g=%q; var N=%q; var TwoFactorRequired=%q;`+"\n", hdlr.Bits, hdlr.NGData.G, hdlr.NGData.N, hdlr.TwoFactorRequired)
		fmt.Fprintf(www, "var SecurityData = %s;\n", SecurityData)
	default:
		AnError(hdlr, www, req, 400, 1046, "Only JSON and JS formats are supported.")
		return
	}
}

// ============================================================================================================================================
//
// Input:
//		 Email, Salt, Verifier
// Output:
//		status: 			success/error on registration step.
//		TwoFactorRequired: 	flag if 2fa is required for login.
//		msg:				if error, then the error message to display to user.
// Example: http://localhost:8000/api/srp_register?salt=s&email=e@e.eee&v=abc
//		{ "status":"success", "TwoFactorRequired":true }
//
func respHandlerSRPRegister(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1047, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	godebug.Printf(dbRespHandlerSrpRegister, "AT: %s\n", godebug.LF())
	email := ps.ByNameDflt("email", "")               // this is 'I', unless "username" is provided.
	UserName := ps.ByNameDflt("UserName", "")         // this is 'I', in preference to email
	DeviceID := ps.ByNameDflt("DeviceID", "")         // this is 'I', for regestering of devices
	RealName := ps.ByNameDflt("RealName", "")         // user attribute to associate with user if supplied - xyzzyUserAttrs
	FirstName := ps.ByNameDflt("FirstName", "")       // user attribute to associate with user if supplied - xyzzyUserAttrs
	MidName := ps.ByNameDflt("MidName", "")           // user attribute to associate with user if supplied - xyzzyUserAttrs
	LastName := ps.ByNameDflt("LastName", "")         // user attribute to associate with user if supplied - xyzzyUserAttrs
	User_id := getUUIDAsString()                      // Create a user id
	Customer_id := ps.ByNameDflt("$customer_id$", "") // user attribute to associate with user if supplied - xyzzyCustomerAttrs
	XAttrs := ps.ByNameDflt("XAttrs", "")             // user attribute to associate with user if supplied - xyzzyUserAttrs
	PhoneNo := ps.ByNameDflt("PhoneNo", "")           // user attribute to associate with user if supplied - xyzzyUserAttrs

	//	if ValidateEmail(email) { // emails are OK
	//	} else if ValidUUID(email) { // UUIDs are OK
	//	} else if hdlr.ValidUserName(email) { // Special Accounts are OK
	//	} else {
	isSpecialUsername := false

	if DeviceID == "" {
		if hdlr.ValidUserName(email) { // Special Accounts are OK
			isSpecialUsername = true
		} else if !ValidateEmail(email) {
			AnError(hdlr, www, req, 400, 1048, "Invalid email address.  Did not validate.")
			return
		} else if ValidUUID(email) {
			AnError(hdlr, www, req, 400, 1049, "Invalid email address.  Is a UUID.")
			return
		}
	}

	// Make UserName registration a configuration option.
	if hdlr.UserNameForRegister {
		// Check for correct combination of email, username, DeviceID
		if UserName != "" {
			godebug.Printf(dbRespHandlerSrpRegister, "AT: %s, UserName [%s]\n", godebug.LF(), UserName)

			if len(UserName) <= 6 {
				AnError(hdlr, www, req, 400, 1050, "UserName must be atleast 7 characters long")
				return
			}
			if ok, _ := regexp.MatchString("^[a-zA-Z]", UserName); !ok {
				AnError(hdlr, www, req, 400, 1051, "UserName must be start with a letter, a-z or A-Z")
				return
			}
			if ValidUUID(UserName) {
				AnError(hdlr, www, req, 400, 1052, "UserName can not be a UUID")
				return
			}
		} else { // else - username should be required?
			AnError(hdlr, www, req, 400, 1053, "UserName is required for registration")
			return
		}
	} else if len(UserName) > 0 {
		AnError(hdlr, www, req, 400, 1054, "UserName can not be used for registration")
		return
	}

	// If DeviceID is supplied then this is the registration of a 2fa device - that is associated with a normal user.
	// Only DeviceID is allowed
	if DeviceID != "" {
		if UserName != "" || email != "" {
			AnError(hdlr, www, req, 400, 1055, "DeviceID registrations can not have UserName or Email attributes.")
			return
		}
		if len(DeviceID) != 9 { // xyzzy2016 - length of device ID - shall be moved to function and constant set.
			AnError(hdlr, www, req, 400, 1056, "DeviceID are all 8 digits.")
			return
		}
		if ok, _ := regexp.MatchString("^[0-9]*$", DeviceID); !ok {
			AnError(hdlr, www, req, 400, 1057, "DeviceID are all digits.")
			return
		}
		if !verhoeff_algorithm.ValidateVerhoeff(DeviceID) {
			AnError(hdlr, www, req, 400, 1058, "DeviceID is not valid.")
			return
		}
		UserName = DeviceID
		godebug.Printf(dbRespHandlerSrpRegister, "DeviceID registration! --------- DeviceID [%s]\n", DeviceID)
	}

	// Set the username
	if !hdlr.UserNameForRegister && DeviceID == "" {
		UserName = email
	}

	// Validation of "salt"/"v" to be hex and correct length.
	salt := ps.ByNameDflt("salt", "") // this is 's'
	if !validSrpSalt(salt) {
		AnError(hdlr, www, req, 400, 1059, "Invalid salt.")
		return
	}
	v := ps.ByNameDflt("v", "") // this is 'v'
	if v == "" {
		v = ps.ByNameDflt("verifier", "") // this is 'v'	-- alternate name 'verifier'
	}
	if !validSrpV(v) {
		AnError(hdlr, www, req, 400, 1060, "Invalid 'v' verifier value.")
		return
	}
	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	mdata := make(map[string]string)
	// xdata := make(map[string]string)

	godebug.Printf(dbRespHandlerSrpRegister, "AT: %s\n", godebug.LF())

	SetSaltV(hdlr, www, req, mdata, salt, v) // Encrypt Salt,V if encryption is enabled

	DeviceIDList := make([]DeviceIDType, 0, 10)
	genDeviceIDList := func() {
		DeviceID := GenerateRandomDeviceID()
		genDate := time.Now().Format(time.RFC3339) //
		DeviceIDList = append(DeviceIDList, DeviceIDType{
			DeviceID:     DeviceID,
			CreationDate: genDate,
		})
		mdata["DeviceID"] = DeviceID // The device id for one time keys
		deviceIDList := godebug.SVar(DeviceIDList)
		mdata["DeviceIDList"] = deviceIDList
	}

	mdata["confirmed"] = "n"                         // mark as "email" is not confirmed by user yet
	mdata["disabled"] = "n"                          //
	mdata["disabled_reason"] = ""                    //
	mdata["acct_type"] = "user"                      //
	mdata["n_failed_login"] = "0"                    //
	t := time.Now()                                  //
	tss := t.Format(time.RFC3339)                    //
	mdata["register_date_time"] = tss                //
	mdata["login_date_time"] = tss                   //
	mdata["login_fail_time"] = ""                    // if n_failed_login > threshold then this is the time when to resume login trials
	mdata["privs"] = hdlr.NewUserPrivs               //
	_, s := GenBackupKeys(hdlr, salt, "0", www, req) // "salt2"
	mdata["backup_one_time_keys"] = s                //
	mdata["offline_one_time_keys"] = s               //
	mdata["num_login_times"] = "0"                   //
	mdata["RealName"] = RealName                     //
	mdata["PhoneNo"] = PhoneNo                       //
	mdata["FirstName"] = FirstName                   //
	mdata["MidName"] = MidName                       //
	mdata["LastName"] = LastName                     //
	mdata["UserName"] = UserName                     //
	mdata["User_id"] = User_id                       //
	mdata["Customer_id"] = Customer_id               //
	mdata["XAttrs"] = XAttrs                         //
	mdata["email"] = email                           //
	validation_secret := GenerateValidationSecret()  //
	mdata["validation_secret"] = validation_secret   //
	// mdata["DeviceID"] = GenerateRandomDeviceID()     // The device id for one time keys -- not used anywhere else // xyzzyDeviceID  -- make array
	genDeviceIDList()
	/*
		t := time.Now()
		then := t.Add(5 * time.Minute)
	*/
	fmt.Printf("validation_secret Generation Time = [%s] AT: %s\n", validation_secret, godebug.LF())

	// if DeviceID != "", then lookup DeviceID related item, and find accoun that this is tied to - add that to mdata
	// set "confirmed" to "y", set acct_type to "DeviceID"
	if DeviceID != "" {
		mdata["privs"] = "DeviceID"         //
		mdata["confirmed"] = "y"            //
		mdata["acct_type"] = "DeviceID"     //
		mdata["DeviceID"] = ""              //
		mdata["DeviceIDList"] = ""          //
		mdata["backup_one_time_keys"] = ""  //
		mdata["offline_one_time_keys"] = "" //
		mdata["auth"] = "y"                 // -- New PJS -- Tue Sep 27 11:43:14 MDT 2016
	}
	if isSpecialUsername {
		mdata["privs"] = email     //
		mdata["confirmed"] = "y"   //
		mdata["acct_type"] = email //
		mdata["DeviceID"] = ""     //
		mdata["auth"] = "y"        // -- New PJS -- Tue Sep 27 11:43:14 MDT 2016
		mdata["FirstName"] = email //
		mdata["LastName"] = email  //
	}

	godebug.Printf(dbRespHandlerSrpRegister, "Len v, salt = %d %d SandBoxPrefix [%s] email [%s], UserName [%s], DeviceID [%s] --- settting value at this point \n", len(ps.ByNameDflt("v", "")), len(ps.ByNameDflt("salt", "")), SandBoxPrefix, email, UserName, DeviceID)

	email_auth_token := GenerateEmailAuthKey()

	if tVal, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, UserName)); ok { // xyzzy2016 - hard coded "srp:U:"
		if tVal["confirmed"] == "n" {
			if hdlr.SendEmail {
				go SendEmailViaAWS(hdlr, email, hdlr.EmailApp, "reg-new-user.tmpl", "", email_auth_token)
				SaveEmailAuth(hdlr, rw, email, SandBoxPrefix, email_auth_token)
			}
			AnError(hdlr, www, req, 400, 1061, "Account is already registered with this email but has not been confirmed.  A new email confirmation has been sent.")
			return
		}

		// ----------------------------------------------------------------------------------------------------------------------------------------------------------
		// AllowReregisterDeviceID  bool  // If true (Defaults to false) then will allow re-register of DeviceID (same id).  Good for development and testing only.
		// AllowReregisterDeviceID  == false => return error
		// AllowReregisterDeviceID  == true => no error
		// xyzzyDeviceID - return 1 DeviceID's -- - new created and appended to list of valid DeviceID's
		// ----------------------------------------------------------------------------------------------------------------------------------------------------------
		if DeviceID != "" && !hdlr.AllowReregisterDeviceID {
			AnError(hdlr, www, req, 400, 1062, "Account is already registered with this email/username/id.")
			return
		}

		// Email and Username acconts allways get error
		AnError(hdlr, www, req, 400, 1063, "Account is already registered with this email/username/id.")
		// xyzzy2016 - email be sent to warn user that an attempt was maid to re-register with this name
		return
	}

	dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, UserName), mdata) // registration with "email" as username
	// dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:X:", SandBoxPrefix, HashStrings.Sha256(UserName)), xdata) // Save hashed data as marker that email is used.
	// Used In: func respHandlerCheckEmailAvailable(www http.ResponseWriter, req *http.Request) {
	hdlr.UpsertUserInfo(mdata["User_id"], mdata)

	// If this is a DeviceID registration then it is already confirmed and no email needs to be sent.
	// Else if configured to send email then do so.
	if isSpecialUsername {
	} else if hdlr.SendEmail && DeviceID == "" {
		godebug.Printf(dbRespHandlerSrpRegister, "\n\n%s\nEmail/Email: %s, %s, %s Auth Token: %s\n%s\n\n\n", strings.Repeat("-=- ", 30), email, UserName, DeviceID, email_auth_token, strings.Repeat("-=- ", 30))
		go SendEmailViaAWS(hdlr, email, hdlr.EmailApp, "reg-new-user.tmpl", "", email_auth_token)
		SaveEmailAuth(hdlr, rw, email, SandBoxPrefix, email_auth_token)
	}

	type RegisterRetrunValueType struct {
		Status            string `json:"status"`
		TwoFactorRequired string `json:"TwoFactorRequired"`
	}

	rv := RegisterRetrunValueType{
		Status:            "success",
		TwoFactorRequired: hdlr.TwoFactorRequired,
	}

	www.Header().Set("Content-Type", "application/json")
	// io.WriteString(www, fmt.Sprintf(`{"status":"success","TwoFactorRequired":%q,"DeviceID":%q,"OneUseKeys":%s}`, TwoFactorRequired, TwoFactorDeviceID, lib.SVar(OneUseKeys)))
	io.WriteString(www, lib.SVar(rv))
}

// ============================================================================================================================================
// mux.HandleFunc("/api/checkEmailAvailable", respHandlerCheckEmailAvailable).Method("GET", "POST")         // See if an email is available for registration and other options
/*
    $.validator.addMethod("ajaxEmailAavailable", function (value, element) {
		console.log ( "Doing the AJAX email check on:", value );
        var data = { "emailHash" : SHA256(value), "type": "registerNewUser" },
		$.ajax({
			type: "GET",
			url: "/api/checkEmailAvailable"
			dataType: "json",
			data: data,
			success: function(data) {
				try {
					var rv = JSON.parse(data);
					if ( rv.status == "success" ) {
						return true;
					}
				} catch (e) {
					console.log ( "Ajax Loading Error:", e );
				}
				return false;
			},
			error: function(xhr, textStatus, errorThrown) {
				console.log('ajax loading error... ... ',errorThrown);
				return false;
			}
		});
    }, 'Email address is already used.');

*/
// TODO: needs bandwidth limiiter to 200 per day?
// use the X-Go-FTL-Trx-Id Cookie     for the key, and count down in redis, below 0 show that ALL are ok.
func respHandlerCheckEmailAvailable(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1064, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	www.Header().Set("Content-Type", "application/json")

	email := ps.ByNameDflt("email", "")           // hash of Email address (sha256)
	aType := ps.ByNameDflt("type", "")            // request to check
	trxId := ps.ByNameDflt("X-Go-FTL-Trx-Id", "") // Bandwidth limiter
	found, done := false, false

	fmt.Printf("Validate Email Ajax Style: %s %s, %s\n", email, aType, godebug.LF())

	if trxId == "" {
		done = true
	}

	if !done {
		key := SandBoxKey("BWL:", SandBoxPrefix, trxId)
		rkey, err := DbGetString(hdlr, rw, key)
		if err != nil { // if not found, then create it
			// xyzzy - how to check that trxId is a valid trxId?
			DbSetExpire(hdlr, rw, SandBoxKey("BWL:", SandBoxPrefix, trxId), "200", 86400) // N max per day
			found, done = true, true
		} else {
			// if found, then fetch, parse - decriment - and save.  If less than 0 then start just returning found for all
			n, err := strconv.ParseInt(rkey, 10, 64)
			if err != nil {
				DbSetExpire(hdlr, rw, SandBoxKey("BWL:", SandBoxPrefix, trxId), "200", 86400) // N max per day
			} else {
				n--
				if n < 0 {
					found, done = true, true
				}
				DbSetExpire(hdlr, rw, SandBoxKey("BWL:", SandBoxPrefix, trxId), fmt.Sprintf("%d", n), 86400) // one day
			}
		}
	}

	if !done {
		//if mdata, ok := dataStore.RGetValue(hdlr, rw, email); !ok {
		found = false
		if _, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email)); ok {
			found, done = true, true
		}
	}

	switch aType {
	case "registerNewUser":
		if found {
			// io.WriteString(www, `{"status":"success"}`)
			io.WriteString(www, `false`)
		} else {
			// io.WriteString(www, `{"status":"not-found"}`)
			io.WriteString(www, `true`)
		}
	default:
		// io.WriteString(www, `{"status":"error","code":"1732","msg":"Invalid Input."}`)
		io.WriteString(www, `false`)
	}
}

// ============================================================================================================================================
// Take system test password - verify it
// Then do an email confirm if password matches.
// data: { "email": identity, "admin_password": admin_password }
func respHandlerSimulateEmailConfirm(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1065, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	if !hdlr.InDemoMode {
		fmt.Printf("respHandlerSimulateEmailConfirm called when not InDemoMode, success returned - noting done\n")
		fmt.Fprintf(os.Stderr, "%srespHandlerSimulateEmailConfirm called when not InDemoMode, success returned - noting done, %s%s\n", MiscLib.ColorYellow, godebug.LF(), MiscLib.ColorReset)
		www.Header().Set("Content-Type", "application/json")
		io.WriteString(www, `{"status":"success"}`)
		return
	}

	if !hdlr.InTestMode {
		fmt.Printf("respHandlerSimulateEmailConfirm called when not InTestMode, error returned - nothing done\n")
		fmt.Fprintf(os.Stderr, "%srespHandlerSimulateEmailConfirm called when not InTestMode, error returned - noting done, %s%s\n", MiscLib.ColorYellow, godebug.LF(), MiscLib.ColorReset)
		AnError(hdlr, www, req, 400, 1066, "Only available in 'test' mode.")
		return
	}

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	email := ps.ByNameDflt("email", "")       // this is 'I'
	DeviceID := ps.ByNameDflt("DeviceID", "") // this is 'I'
	if DeviceID == "" {
		if !ValidateEmail(email) {
			AnError(hdlr, www, req, 400, 1067, "Invalid email address")
			return
		}
	} else if email == "" {
		email = DeviceID
	}

	// mdata := make(map[string]string)
	admin_password := ps.ByNameDflt("admin_password", "") // this is 's'
	if hdlr.AdminPassword != admin_password {
		AnError(hdlr, www, req, 400, 1068, "Invalid admin password.")
		return
	}

	//if mdata, ok := dataStore.RGetValue(hdlr, rw, email); !ok {
	if mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email)); !ok {
		AnError(hdlr, www, req, 400, 1069, "Unable to find account with specified email.")
		return
	} else {
		mdata["confirmed"] = "y"
		mdata["auth"] = "y" //
		mdata["num_login_times"] = "0"
		user := SandBoxKey("srp:U:", SandBoxPrefix, email)
		fmt.Fprintf(os.Stderr, "%suser %s has been confirmed by %s%s\n", MiscLib.ColorGreen, user, admin_password, MiscLib.ColorReset)
		dataStore.RSetValue(hdlr, rw, user, mdata)
	}
	www.Header().Set("Content-Type", "application/json")
	io.WriteString(www, `{"status":"success"}`)

}

// ============================================================================================================================================
// Set account to "confirmed" = "y"
//
// Input:
//	email_auth_token
// Output: if token is valid then updates user to "confirmed=='y'" and sends welcome message
//
func respHandlerEmailConfirm(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1070, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps
	_ = hdlr

	godebug.Printf(dbRespHandlerSrpRegister, "AT: %s\n", godebug.LF())
	email := ""
	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")
	email_auth_token := ps.ByNameDflt("email_auth_token", "")
	godebug.Printf(dbRespHandlerSrpRegister, "email_auth_token: %s, %s\n", email_auth_token, godebug.LF())
	if email_auth_token != "" {
		email, ok = GetEmailAuth(hdlr, rw, email_auth_token, SandBoxPrefix) // func GetEmailAuth(email_auth_token string) (email string, ok bool) {
		if !ok {
			AnError(hdlr, www, req, 400, 1071, "Invalid email address.")
			godebug.Printf(dbRespHandlerSrpRegister, "Error code=0016 Token Looked Up:%s, %s\n", email_auth_token, godebug.LF())
			return
		}
	} else {
		AnError(hdlr, www, req, 400, 1072, "Invalid auth_token address.")
		return
	}

	if false {
		go SendEmailViaAWS(hdlr, email, hdlr.EmailApp, "welcome-user.tmpl", email, email_auth_token)
	}

	//if mdata, ok := dataStore.RGetValue(hdlr, rw, email); !ok {
	if mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email)); !ok {
		AnError(hdlr, www, req, 400, 1073, "Unable to find account with specified email.")
		return
	} else {
		mdata["confirmed"] = "y"
		mdata["auth"] = "y" //
		mdata["num_login_times"] = "0"
		dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)
	}

	www.Header().Set("Content-Type", "application/json")
	io.WriteString(www, `{"status":"success"}`)
}

// ============================================================================================================================================
// Check that the person is logged in
// Update the 'v' and 'salt' values for 'email'
// Send email to person saying that password was updated.
//
// muxEnc.HandleFunc("/api/srp_change_password", respHandlerChangePassword).Method("GET", "POST")            // ENC: Set a new password
//
func respHandlerChangePassword(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1074, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	tt := ps.ByNameDflt("t", "")
	if tt == "" || (InjectionTestMode && TestModeInject["invalid-tt-change-password"]) {
		AnError(hdlr, www, req, 400, 1075, `Invalid input data`)
		return
	}

	// Validation of "salt"/"v" to be hex and correct length.
	salt := ps.ByNameDflt("salt", "") // this is 's'
	if !validSrpSalt(salt) {
		AnError(hdlr, www, req, 400, 1076, `Invalid salt`)
		return
	}
	v := ps.ByNameDflt("v", "") // this is 'v'
	if v == "" {
		v = ps.ByNameDflt("verifier", "") // this is 'v'	-- alternate name 'verifier'
	}
	if !validSrpV(v) {
		AnError(hdlr, www, req, 400, 1077, "Invalid 'v' verifier value.")
		return
	}
	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	mdata := make(map[string]string)

	SetSaltV(hdlr, www, req, mdata, salt, v) // Encrypt Salt,V if encryption is enabled

	// Just need to check that the user is logged in - if they are then... go ahead and change password.
	// Basically if this arrives encrypted with a session ID and an "auth" == "y" -then- change pw. with new 'v'/'salt'

	email, err := UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
	if err != nil { // check user exists
		AnError(hdlr, www, req, 400, 1078, "Unable to find account with specified email.")
		return
	}

	mdata["confirmed"] = "y"       //
	mdata["auth"] = "y"            //
	mdata["n_failed_login"] = "0"  //
	t := time.Now()                //
	tss := t.Format(time.RFC3339)  //
	mdata["login_date_time"] = tss //
	mdata["login_fail_time"] = ""  // if n_failed_login > threshold then this is the time when to resume login trials

	dataStore.RUpdValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)

	if hdlr.SendEmail {
		go SendEmailViaAWS(hdlr, email, hdlr.EmailApp, "password-changed.tmpl", email, "")
		// SaveEmailAuth(hdlr, rw, email, SandBoxPrefix, "")	// Xyzzy should this be in?
	}

	// expire := time.Now().AddDate(0, 0, hdlr.CookieExpireInXDays) // Years, Months, Days==2 // Xyzzy501 - should be a config - on how long to keep cookie
	// cookie := http.Cookie{Name: "LoginAuthToken", Value: "x", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400, Secure: false, HttpOnly: true}
	expire := time.Now().AddDate(0, 0, -1) // Years, Months, Days==2 // Xyzzy501 - should be a config - on how long to keep cookie
	cookie := http.Cookie{Name: "LoginAuthToken", Value: "x", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 0, Secure: false, HttpOnly: true}
	http.SetCookie(www, &cookie)
	www.Header().Set("Content-Type", "application/json")
	io.WriteString(www, `{"status":"success"}`)
}

// ============================================================================================================================================
//	muxEnc.HandleFunc("/api/getPageToken", respHandlerGetPageToken).Method("GET", "POST")                   // Mark page with cookie for password recovery, "page" marker
func respHandlerGetPageToken(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1079, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	outputFmt := ps.ByNameDflt("fmt", "json") // json or js
	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	id0, _ := uuid.NewV4()
	pageMarker := id0.String()
	DbSetExpire(hdlr, rw, SandBoxKey("pm:", SandBoxPrefix, pageMarker), `{"marked":"yes"}`, hdlr.PwExpireIn) // xyzzy - need expire for this in sec

	if h := req.Header.Get("Origin"); h != "" {
		www.Header().Set("Access-Control-Allow-Origin", h) // Allow requests from any server for this.
	} else {
		www.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any server for this.
	}
	www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	www.Header().Set("Expires", "0")                                         // Proxies.

	expire := time.Now().AddDate(0, 0, hdlr.CookieExpireInXDays) // Years, Months, Days==2
	// xyzzy need expire for this in seconds
	cookie := http.Cookie{Name: "pageMarkerCookie", Value: pageMarker, Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400, Secure: false, HttpOnly: true}
	http.SetCookie(www, &cookie)

	switch outputFmt {
	case "json", "JSON":
		www.Header().Set("Content-Type", "application/json") // For JSON data
		fmt.Fprintf(www, `{"status":"success","pageMarker":%q}`, pageMarker)
	case "js", "JS":
		www.Header().Set("Content-Type", "application/javascript") // For JS
		fmt.Fprintf(www, "var pageMarker = %q;\n", pageMarker)
	default:
		AnError(hdlr, www, req, 400, 1080, "Only JSON and JS formats are supported.")
		return
	}
}

// ============================================================================================================================================
// Generate a temporary "token" for update/recovery of password
// Put "token" into account - mark as started recovery
// Add a "start-recovery-date" field
// Send email with "token" using recovery template.
//
/*

### Overview of process

0. Recover page is marked with an acceptable token "page" - page token is fetched/generated.
	1. /api/getPageToken - cookie lasts for 24 hrs. - called for JS code in pt1 - recover page
1. Generate token
2. Mark account with token and timestamp of recovery start
2. On request to pt1 - take "page" cookie and mark on the server side it with email.
3. Save 'prw:'||token with "email" and a 24hr timeout to d.b.
4. Send Email with `auth_token`
	1.  If user clicks on link
		1. Link directs to /api/pwrecov2 that will do a server-temporary redirect to page to enter Password.
		1. A "link" token cookey is created in the response as hash(Salt:Email:Token) - this will get checked later
		2. Creates a "cookie" that passes the email/username to the client for display in form.
		3. Client displays form - and deletes cookie.
		4. User enters new password + token - hits submit.
	2.  If user re-enters token into form.  - the form already has email from the request to reset password -
		1. User enters new password + token - hits submit.
	3. Call is made with Token/Usernmae=email/Salt/V to reset password
		1. Email is sent (optionally PwSendEmailOnRecoverPw) to user to notify them.
	4. A new Salt/V is associated with the account, New DeviceID, new one-time-keys -- User is logged in.
		0. Verify either the "page" token or the more secure "link" token
		1. new salt/v for verifying paswords
		2. resets invalid login count and last login dates
		2. Create new DeviceID, one-timekeys - and returns these to user.
5. Happy user logged in - Enteres new DeviceID into 2fa device.  Probably clicks *get one time key* key button.
	1. On next click of *get one time key* button - when device is connected to network
		1. Will know that device is not registed.
		2. Will register and login
		3. Will get new backup offline one time keys
		4. Will get a one-time-key for the user to login


*/
//	mux.HandleFunc("/api/srp_recover_password_pt1", respHandlerRecoverPasswordPt1).Method("GET", "POST")     // password recovery (step 1) send email
func respHandlerRecoverPasswordPt1(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1081, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")
	email := ps.ByNameDflt("email", "") // this is 'I'
	if !ValidateEmail(email) {
		AnError(hdlr, www, req, 400, 1082, "Invalid email address.")
		return
	}

	email_auth_token := GenerateEmailAuthKey()

	// PageMarker "page" -- This indicates that the /api/getPageToken has been visited.
	pageMarker := ps.ByNameDflt("pageMarkerCookie", "")
	if pageMarker == "" {
		AnError(hdlr, www, req, 400, 1083, "Invalid page marker.")
		return
	}

	fmt.Printf("8888 pageMarker=%s (will be marked as used with email and token), email=%s email_auth_token=%s, %s\n", pageMarker, email, email_auth_token, godebug.LF())

	// Basically just check to see that the apge has the pageMarker cookie set and it matches one on the server side.
	_, err := DbGetString(hdlr, rw, SandBoxKey("pm:", SandBoxPrefix, pageMarker))
	if err != nil {
		AnError(hdlr, www, req, 400, 1084, "Invalid page marker.")
		return
	}

	// Mark the server side with the email address - this will be check later in pt2
	DbSetExpire(hdlr, rw, SandBoxKey("pm:", SandBoxPrefix, pageMarker), fmt.Sprintf(`{"marked":"used","email":%q,"email_auth_token":%q}`, email, email_auth_token), hdlr.PwExpireIn) // xyzzy - need expire for this in sec

	if hdlr.SendEmail {
		go SendEmailViaAWS(hdlr, email, hdlr.EmailApp, "password-reset.tmpl", "", email_auth_token)
		// SaveEmailAuth(hdlr, rw, email, SandBoxPrefix, email_auth_token) // PreEau
		DbSetExpire(hdlr, rw, SandBoxKey(hdlr.PwResetKey, SandBoxPrefix, email_auth_token), email, hdlr.PwExpireIn)
	}

	www.Header().Set("Content-Type", "application/json")
	io.WriteString(www, `{"status":"success"}`)
}

// ============================================================================================================================================
//
// Given the "token" and the "email" - allow user to set 'v' and 'salt' values.
//
// *xyzzyPw1 - Verify on password recovery - a new DeviceID is issued.
// *xyzzyPw1 - Verify on password recovery - Temporary IDs (one time keys) - Both Kinds - are reset on password reset
// *xyzzyPw1 - Verify on password recovery - new DeviceID must go to user - and be displayed with info -
//
// xyzzyPw1 - Verify on password recovery - Delete account with old DeviceID
//
func respHandlerRecoverPasswordPt2(www http.ResponseWriter, req *http.Request) {

	var err error

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1085, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	email_auth_token := ps.ByNameDflt("email_auth_token", "") //
	linkMarkerCookie := ps.ByNameDflt("linkMarkerCookie", "") //
	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")
	if email_auth_token == "" {
		email_auth_token = ps.ByNameDflt("LoginAuthToken", "") //
	}

	fmt.Printf("email_auth_token=[%s], %s\n", email_auth_token, godebug.LF())
	fmt.Printf("linkMarkerCookie=[%s], %s\n", linkMarkerCookie, godebug.LF())

	email := ""
	err8011 := false
	if email_auth_token == "" {
		err8011 = true
	} else {
		email, err = DbGetString(hdlr, rw, SandBoxKey(hdlr.PwResetKey, SandBoxPrefix, email_auth_token))
		if err != nil {
			email = ""
			err8011 = true
		}
	}

	fmt.Printf("At, %s\n", godebug.LF())

	// sholud validate "page" or "link" cookies
	linkMarker := ps.ByNameDflt("linkMarkerCookie", "") //
	pageMarker := ps.ByNameDflt("pageMarkerCookie", "") //
	fmt.Printf("cookies for password reset, pageMarkerCookie [%s] linkMarkerCookie [%s], %s\n", pageMarker, linkMarker, godebug.LF())
	fmt.Printf("At, %s\n", godebug.LF())
	if email != "" {
		fmt.Printf("At, %s\n", godebug.LF())
		if pageMarker != "" {
			fmt.Printf("At, %s\n", godebug.LF())
			pageMarkerData, err := DbGetString(hdlr, rw, SandBoxKey("pm:", SandBoxPrefix, pageMarker))
			if err != nil {
				AnError(hdlr, www, req, 400, 1086, "Invalid token.")
				return
			}
			dd, err := lib.JsonStringToString(pageMarkerData)
			if err != nil {
				AnError(hdlr, www, req, 400, 1087, "Invalid token.")
				return
			}
			if dd["email"] != email {
				AnError(hdlr, www, req, 400, 1088, "Invalid token.")
				return
			}
			if dd["email_auth_token"] != email_auth_token {
				AnError(hdlr, www, req, 400, 1089, "Invalid token.")
				return
			}
		}
	} else {
		fmt.Printf("At, %s\n", godebug.LF())
		if pageMarker != "" {
			fmt.Printf("At, %s\n", godebug.LF())
			pageMarkerData, err := DbGetString(hdlr, rw, SandBoxKey("pm:", SandBoxPrefix, pageMarker))
			fmt.Printf("At, pageMarkerData=%s, %s\n", pageMarkerData, godebug.LF())
			if err != nil {
				AnError(hdlr, www, req, 400, 1090, "Invalid token.")
				return
			}
			dd, err := lib.JsonStringToString(pageMarkerData)
			if err != nil {
				AnError(hdlr, www, req, 400, 1091, "Invalid token.")
				return
			}
			fmt.Printf("At, %s\n", godebug.LF())
			email = dd["email"]
			email_auth_token = dd["email_auth_token"]
			email1, err := DbGetString(hdlr, rw, SandBoxKey(hdlr.PwResetKey, SandBoxPrefix, email_auth_token))
			if err != nil || email != email1 {
				AnError(hdlr, www, req, 400, 1092, "Invalid token.")
				return
			}
			fmt.Printf("At, %s\n", godebug.LF())
		}
		if linkMarker != "" {
			fmt.Printf("At, %s\n", godebug.LF())
			linkMarkerData, err := DbGetString(hdlr, rw, SandBoxKey("pm:", SandBoxPrefix, linkMarker))
			fmt.Printf("At, linkMarkerData=%s, %s\n", linkMarkerData, godebug.LF())
			if err != nil {
				AnError(hdlr, www, req, 400, 1093, "Invalid token.")
				return
			}
			dd, err := lib.JsonStringToString(linkMarkerData)
			fmt.Printf("At, %s\n", godebug.LF())
			if err != nil {
				AnError(hdlr, www, req, 400, 1094, "Invalid token.")
				return
			}
			if dd["hash"] != HashStrings.Sha256(dd["salt"]+":"+email_auth_token+":"+email) {
				AnError(hdlr, www, req, 400, 1095, "Invalid token.")
				return
			}
			fmt.Printf("At, %s\n", godebug.LF())
		}
		fmt.Printf("At, %s\n", godebug.LF())
		err8011 = false
	}
	fmt.Printf("At, %s\n", godebug.LF())

	if err8011 { // error from above, not resoved
		fmt.Printf("****************** email [%s] ok [%v] email_auth_token [%s], %s\n", email, ok, email_auth_token, godebug.LF())
		AnError(hdlr, www, req, 400, 1096, "Invalid token.")
		return
	}

	// This is  the early return - if we have a good token but no reset info. -client- should detect this and display the password form.
	salt := ps.ByNameDflt("salt", "") // this is 's'
	v := ps.ByNameDflt("v", "")       // this is 'v'
	if v == "" {
		v = ps.ByNameDflt("verifier", "") // this is 'v'	-- alternate name 'verifier'
	}
	if salt == "" && v == "" {
		fmt.Printf("At, %s\n", godebug.LF())
		// xyzzy - error occureds at this point because LoginAuthEmail is not set
		expire := time.Now().AddDate(0, 0, hdlr.CookieExpireInXDays) // Years, Months, Days==2 // Xyzzy501 - should be a config - on how long to keep cookie
		cookie := http.Cookie{Name: "LoginAuthEmail", Value: email, Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400, Secure: false, HttpOnly: true}
		http.SetCookie(www, &cookie)
		www.Header().Set("Content-Type", "application/json")
		expire = time.Now().AddDate(0, 0, hdlr.CookieExpireInXDays) // Years, Months, Days==2 // Xyzzy501 - should be a config - on how long to keep cookie
		cookie = http.Cookie{Name: "LoginAuthToken", Value: email_auth_token, Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400, Secure: false, HttpOnly: true}
		http.SetCookie(www, &cookie)
		www.Header().Set("Content-Type", "application/json")
		io.WriteString(www, `{"status":"success"}`)
		return
	}

	fmt.Printf("At, %s\n", godebug.LF())
	// Validation of "salt"/"v" to be hex and correct length.
	if !validSrpSalt(salt) {
		AnError(hdlr, www, req, 400, 1097, "Invalid salt.")
		return
	}
	if !validSrpV(v) {
		AnError(hdlr, www, req, 400, 1098, "Invalid 'v' verifier value.")
		return
	}

	mdata := make(map[string]string)

	SetSaltV(hdlr, www, req, mdata, salt, v) // Encrypt Salt,V if encryption is enabled

	mdata["confirmed"] = "y"       // mark as "email" is not confirmed by user yet
	mdata["num_login_times"] = "0" //	should this be set?
	mdata["n_failed_login"] = "0"  //
	t := time.Now()                //
	tss := t.Format(time.RFC3339)  //
	mdata["login_date_time"] = tss //
	mdata["login_fail_time"] = ""  // if n_failed_login > threshold then this is the time when to resume login trials

	var deviceIDList string
	DeviceIDList := make([]DeviceIDType, 0, 10)
	genDeviceIDList := func() string {
		DeviceID := GenerateRandomDeviceID()
		genDate := time.Now().Format(time.RFC3339) //
		DeviceIDList = []DeviceIDType{DeviceIDType{
			DeviceID:     DeviceID,
			CreationDate: genDate,
		}}
		mdata["DeviceID"] = DeviceID // The device id for one time keys
		deviceIDList = godebug.SVar(DeviceIDList)
		mdata["DeviceIDList"] = deviceIDList
		return DeviceID
	}

	// xyzzyDeviceID - return 1 DeviceID's -- - new created and appended to list of valid DeviceID's
	ki, s := GenBackupKeys(hdlr, salt, "0", www, req) // "salt2"
	mdata["backup_one_time_keys"] = s                 //
	_, ss := GenBackupKeys(hdlr, salt, "0", www, req) // "salt2"
	mdata["offline_one_time_keys"] = ss               //
	NewDeviceID := genDeviceIDList()

	fmt.Printf("Password Recover Len v, salt = %d %d email %s --- settting value at this point \n", len(ps.ByNameDflt("v", "")), len(ps.ByNameDflt("salt", "")), email)

	dataStore.RUpdValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)

	// 3. An email is sent (xyzzy 1hr - xyzzyEmailChanged ) to tell user that the password was chagned.
	// 3. xyzzyEmailChanged - create the template
	if hdlr.SendEmail {
		go SendEmailViaAWS(hdlr, email, hdlr.EmailApp, "password-changed.tmpl", "", "") // xyzzyRealName	-- add users real name
		// SaveEmailAuth(hdlr, rw, email, SandBoxPrefix, email_auth_token)
	}

	// xyzzy - should delete LoginAuthEmail?
	// expire := time.Now().AddDate(0, 0, hdlr.CookieExpireInXDays) // Years, Months, Days==2 // Xyzzy501 - should be a config - on how long to keep cookie
	// cookie := http.Cookie{Name: "LoginAuthToken", Value: "x", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400, Secure: false, HttpOnly: true}
	expire := time.Now().AddDate(0, 0, -1) // Years, Months, Days==2 // Xyzzy501 - should be a config - on how long to keep cookie
	cookie := http.Cookie{Name: "LoginAuthToken", Value: "x", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 0, Secure: false, HttpOnly: true}
	http.SetCookie(www, &cookie)
	www.Header().Set("Content-Type", "application/json")

	// Delete the cookie
	expire = time.Now().AddDate(0, 0, -1) // Years, Months, Days==-1
	cookie = http.Cookie{Name: "pageMarkerCookie", Value: "", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 0, Secure: false, HttpOnly: true}
	http.SetCookie(www, &cookie)
	cookie = http.Cookie{Name: "linkMarkerCookie", Value: "", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 0, Secure: false, HttpOnly: true}
	http.SetCookie(www, &cookie)

	if hdlr.TwoFactorRequired == "y" && linkMarker != "" {
		// xyzzyDeviceID - return 1 DeviceID's -- - new created and appended to list of valid DeviceID's
		io.WriteString(www, fmt.Sprintf(`{"status":"success","DeviceID":%q,"DeviceIDList":%s,"BackupKeys":%q,"email":%q}`, NewDeviceID, deviceIDList, ki, email))
	} else if hdlr.TwoFactorRequired == "y" {
		// xyzzyDeviceID - return 1 DeviceID's -- - new created and appended to list of valid DeviceID's
		io.WriteString(www, fmt.Sprintf(`{"status":"success","DeviceID":%q,"DeviceIDList":%s,"BackupKeys":%q}`, NewDeviceID, deviceIDList, ki))
	} else if hdlr.TwoFactorRequired == "n" && linkMarker != "" {
		io.WriteString(www, fmt.Sprintf(`{"status":"success","email":%q}`, email))
	} else {
		io.WriteString(www, fmt.Sprintf(`{"status":"success"}`))
	}
}

// ============================================================================================================================================
// Given an "admin" login - set password - and clean out temporary values like "confirmed"="n" and "token"
// Reset temporary disabled accounts
func respHandlerAdminSetPassword(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1099, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	// 0. verify admin privilages

	email_of_user := ps.ByNameDflt("email_of_user", "") //
	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	tt := ps.ByNameDflt("t", "")
	if tt == "" || (InjectionTestMode && TestModeInject["invalid-tt-change-password"]) {
		AnError(hdlr, www, req, 400, 1100, `Invalid input data`)
		return
	}

	email, err := UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
	if err != nil { // check user exists
		AnError(hdlr, www, req, 400, 1101, "Unable to find account with specified email.")
		return
	}

	// admin_mdata := make(map[string]string)
	// mdata := make(map[string]string)

	admin_mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email))
	if !ok {
		AnError(hdlr, www, req, 400, 1102, "Unable to loade admin session data.  Please logout and logback in.")
		return
	}

	if admin_mdata["privs"] != "admin" { // if not an admin account
		AnError(hdlr, www, req, 400, 1103, "Unable to identify user as an 'admin'.")
		return
	}

	if godebug.InArrayString("MayChangeOtherPassword", hdlr.SecurityPrivilages["admin"]) < 0 { // if not found
		AnError(hdlr, www, req, 400, 1104, "Admin missing privilage 'MayChangeOtherPassword'.")
		return
	}

	mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email_of_user))
	if !ok {
		AnError(hdlr, www, req, 400, 1105, "Email is not valid for designated use - the one having the chagne made to.")
		return
	}

	// Validation of "salt"/"v" to be hex and correct length.
	salt := ps.ByNameDflt("salt", "") // this is 's'
	if !validSrpSalt(salt) {
		AnError(hdlr, www, req, 400, 1106, `Invalid salt`)
		return
	}
	v := ps.ByNameDflt("v", "") // this is 'v'
	if v == "" {
		v = ps.ByNameDflt("verifier", "") // this is 'v'	-- alternate name 'verifier'
	}
	if !validSrpV(v) {
		AnError(hdlr, www, req, 400, 1107, "Invalid 'v' verifier value.")
		return
	}

	mdata["confirmed"] = "y"       //
	mdata["auth"] = "y"            //
	mdata["n_failed_login"] = "0"  //
	t := time.Now()                //
	tss := t.Format(time.RFC3339)  //
	mdata["login_date_time"] = tss //
	mdata["login_fail_time"] = ""  // if n_failed_login > threshold then this is the time when to resume login trials

	dataStore.RUpdValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)

	if hdlr.SendEmail {
		go SendEmailViaAWS(hdlr, email, hdlr.EmailApp, "password-changed-by-admin.tmpl", email, "")
	}

	// expire := time.Now().AddDate(0, 0, hdlr.CookieExpireInXDays) // Years, Months, Days==2 // Xyzzy501 - should be a config - on how long to keep cookie
	// cookie := http.Cookie{Name: "LoginAuthToken", Value: "x", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400, Secure: false, HttpOnly: true}
	expire := time.Now().AddDate(0, 0, -1) // Years, Months, Days==2 // Xyzzy501 - should be a config - on how long to keep cookie
	cookie := http.Cookie{Name: "LoginAuthToken", Value: "x", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 0, Secure: false, HttpOnly: true}
	http.SetCookie(www, &cookie)
	www.Header().Set("Content-Type", "application/json")
	io.WriteString(www, `{"status":"success"}`)

}

// ============================================================================================================================================
// Given an "admin" login - set password - and clean out temporary values like "confirmed"="n" and "token"
// Reset temporary disabled accounts
func respHandlerAdminSetAttributes(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1108, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	// 0. verify admin privilages

	email_of_user := ps.ByNameDflt("email_of_user", "") //
	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	tt := ps.ByNameDflt("t", "")
	if tt == "" || (InjectionTestMode && TestModeInject["invalid-tt-change-password"]) {
		AnError(hdlr, www, req, 400, 1109, `Invalid input data`)
		return
	}

	email, err := UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
	if err != nil { // check user exists
		AnError(hdlr, www, req, 400, 1110, "Unable to find account with specified email.")
		return
	}

	// admin_mdata := make(map[string]string)
	// mdata := make(map[string]string)

	admin_mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email))
	if !ok {
		AnError(hdlr, www, req, 400, 1111, "Unable to loade admin session data.  Please logout and logback in.")
		return
	}

	if admin_mdata["privs"] != "admin" { // if not an admin account
		AnError(hdlr, www, req, 400, 1112, "Unable to identify user as an 'admin'.")
		return
	}

	if godebug.InArrayString("MayChangeOtherAttributes", hdlr.SecurityPrivilages["admin"]) < 0 { // if not found
		AnError(hdlr, www, req, 400, 1113, "Admin missing privilage 'MayChangeOtherAttributes'.")
		return
	}

	mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email_of_user))
	if !ok {
		AnError(hdlr, www, req, 400, 1114, "Email is not valid for designated use - the one having the chagne made to.")
		return
	}

	// Validation of "salt"/"v" to be hex and correct length.
	attrs := ps.ByNameDflt("attrs", "") // this is 's'

	mdata["confirmed"] = "y"       //
	mdata["auth"] = "y"            //
	mdata["n_failed_login"] = "0"  //
	t := time.Now()                //
	tss := t.Format(time.RFC3339)  //
	mdata["login_date_time"] = tss //
	mdata["login_fail_time"] = ""  // if n_failed_login > threshold then this is the time when to resume login trials

	// func JsonStringToData(s string) (theJSON map[string]interface{}, err error) {
	AttrsParsed, err := lib.JsonStringToString(attrs)

	// parse the attrs - eliminate bad names - set
	for ii, vv := range AttrsParsed {
		if AdminReservedIDs[ii] {
			continue
		}
		mdata[ii] = vv
	}

	dataStore.RUpdValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)

	if hdlr.SendEmail {
		go SendEmailViaAWS(hdlr, email, hdlr.EmailApp, "password-changed-by-admin.tmpl", email, "")
	}

	// expire := time.Now().AddDate(0, 0, hdlr.CookieExpireInXDays) // Years, Months, Days==2 // Xyzzy501 - should be a config - on how long to keep cookie
	// cookie := http.Cookie{Name: "LoginAuthToken", Value: "x", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400, Secure: false, HttpOnly: true}
	expire := time.Now().AddDate(0, 0, -1) // Years, Months, Days==2 // Xyzzy501 - should be a config - on how long to keep cookie
	cookie := http.Cookie{Name: "LoginAuthToken", Value: "x", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 0, Secure: false, HttpOnly: true}
	http.SetCookie(www, &cookie)
	www.Header().Set("Content-Type", "application/json")
	io.WriteString(www, `{"status":"success"}`)

}

// ============================================================================================================================================
// theMux.HandleFunc("/api/force_logout", respHandlerSRPLogout).Methods("GET", "POST")
func respHandlerForceLogout(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1115, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	email_of_user := ps.ByNameDflt("email", "")

	// xyzzy - pull up user info based on email
	// xyzzy - find "t" based on email
	// 		mdata["s2"] = s2 -- Look like this one
	// xyzzy - look in redis to see if "t" value is stored with user on srp_validate

	admin_password := ps.ByNameDflt("admin_password", "") // this is 's'
	if hdlr.AdminPassword != admin_password {
		AnError(hdlr, www, req, 8432, 1116, "Invalid admin password.")
		return
	}

	// use "t" to verify that we are logged in as an "admin"

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")
	tt := ps.ByNameDflt("t", "")
	if tt == "" {
		AnError(hdlr, www, req, 8433, 1117, "Invalid, must be logged in as admin.")
		return
	}

	email, err := UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
	if err != nil { // check user exists
		AnError(hdlr, www, req, 400, 1118, "Unable to find account with specified email.")
		return
	}

	// admin_mdata := make(map[string]string)
	// mdata := make(map[string]string)

	admin_mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email))
	if !ok {
		AnError(hdlr, www, req, 400, 1119, "Unable to loade admin session data.  Please logout and logback in.")
		return
	}

	if admin_mdata["privs"] != "admin" { // if not an admin account
		AnError(hdlr, www, req, 400, 1120, "Unable to identify user as an 'admin'.")
		return
	}

	if godebug.InArrayString("MayChangeOtherAttributes", hdlr.SecurityPrivilages["admin"]) < 0 { // if not found
		AnError(hdlr, www, req, 400, 1121, "Admin missing privilage 'MayChangeOtherAttributes'.")
		return
	}

	mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email_of_user))
	if !ok {
		AnError(hdlr, www, req, 400, 1122, "Email is not valid for designated use - the one having the chagne made to.")
		return
	}

	// Get TT of email_of_user
	tt_of_user := mdata["s2"]

	// delete "ses:<T>"
	// DbDel(hdlr, rw, "ses:"+tt_of_user)
	DbDel(hdlr, rw, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt_of_user))

	//	IamI := ps.ByNameDflt("IamI", "")
	//	IamI_key := SandBoxKey("1x1:", SandBoxPrefix, IamI)
	//	DbDel(hdlr, rw, IamI_key)
	//
	//	SaveLogoutData(hdlr, rw, tt, SandBoxPrefix)
	//
	//	secure := false
	//	if req.TLS != nil {
	//		secure = true
	//	}
	//	expire := time.Now().Add(-1 * time.Second)
	//	cookie := http.Cookie{Name: "IamI", Value: "", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 0, Secure: secure, HttpOnly: true}
	//	http.SetCookie(www, &cookie)

	www.Header().Set("Content-Type", "application/json")
	rv := `{"status":"success"}`
	io.WriteString(www, rv)
}

// ============================================================================================================================================
// theMux.HandleFunc("/api/srp_logout", respHandlerSRPLogout).Methods("GET", "POST")
func respHandlerSRPLogout(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1123, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	tt := ps.ByNameDflt("t", "")
	if tt == "" {
		www.Header().Set("Content-Type", "application/json")
		rv := `{"status":"success"}`
		io.WriteString(www, rv)
		return
	}
	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")
	IamI := ps.ByNameDflt("IamI", "")
	IamI_key := SandBoxKey("1x1:", SandBoxPrefix, IamI)
	DbDel(hdlr, rw, IamI_key)

	SaveLogoutData(hdlr, rw, tt, SandBoxPrefix)

	secure := false
	if req.TLS != nil {
		secure = true
	}
	expire := time.Now().Add(-1 * time.Second)
	cookie := http.Cookie{Name: "IamI", Value: "", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 0, Secure: secure, HttpOnly: true}
	http.SetCookie(www, &cookie)

	www.Header().Set("Content-Type", "application/json")
	rv := `{"status":"success"}`
	io.WriteString(www, rv)
}

/*
// ============================================================================================================================================

From: http://srp.stanford.edu/ndss.html

(With additional notes by me)

The SRP Protocol
================

What follows is a complete description of the entire SRP authentication process from beginning to end, starting with the password setup steps.

			Table 3: Mathematical Notation for SRP
			---------------------------------------

		Var		Description

		n	   	A large prime number. All computations are performed modulo n.
		g	   	A primitive root modulo n (often called a generator)
		s	   	A random string used as the user's salt
		P	   	The user's password
		x	   	A private key derived from the password and salt
		v	   	The host's password verifier
		u	   	Random scrambling parameter, publicly revealed
		a,b	   	Ephemeral private keys, generated randomly and not publicly revealed
		A,B	   	Corresponding public keys
		H()	   	One-way hash function. 					PJS/note: In this H() will be Sha256
		m,n	   	The two quantities (strings) m and n concatenated
		K	   	Session key

		N		Modulo number
		C		Carol's Username (carol@example.com, also referred to as I in some cases)
		D.B.	Database
		t		Random UUID used as salt for generating session ID in steps 9,10

Table 3 shows the notation used in this section. The values n and g are well-known values, agreed to beforehand.

To establish a password P with Steve, Carol picks a random salt s, and computes

		x = H(s, P)
		v = g^x

Steve stores v and s as Carol's password verifier and salt. Remember that the computation of v is implicitly reduced modulo n. x is
discarded because it is equivalent to the plaintext password P.

The AKE protocol also allows Steve to have a password z with a corresponding public key held by Carol; in SRP, we set z = 0 so
that it drops out of the equations. Since this private key is 0, the corresponding public key is 1. Consequently, instead of
safeguarding its own password z, Steve needs only to keep Carol's verifier v secret to assure mutual authentication. This frees Carol
from having to remember Steve's public key and simplifies the protocol.

To authenticate, Carol and Steve engage in the protocol described in Table 4. A description of each step follows:

			Table 4: The Secure Remote Password Protocol
			--------------------------------------------

		Step	Carol				Communication				Steve
		1.								C -->					(lookup s, v from D.B. by username(C)) - send back s		/api/srp_login
		2.		x = H(s, P)				<-- s, t				t is 2nd salt sent back to client

		3.		A = g^a					A,C -->																				/api/srp_confirm
		4.								<-- B, u				B = v + g^b			Lookup D.B. (C), get s,v - gen b      ?? u
		5.		S = (B - g^x)^(a + ux)							S = (A · v^u)^b		Both sides can now compute S
		6.		K = H(S)										K = H(S)			Both sides compute the same K - Update D.B. with A,B,b,K

		7.		M[1] = H(A, B, K)		M[1],C -->				(verify M[1])		Lookup D.B. getting s, A, B, b, K
		8.		(verify M[2])			<-- M[2]				M[2] = H(A, M[1], K)

		9.		U = H(t,K)										U = H(t,K)			Generate session ID, store K, C, s, info in D.B. with key U
		10.		Use U as key for communication
				(Encrypt with K)								(Decrypt with K) (Encrypt Responses with K)
				(Decrypt with K)


1. Carol sends Steve her username, (e.g. carol@example.com).
	Example #8
2. Steve looks up Carol's password entry and fetches her password verifier v and her salt s. He sends s to Carol.
	Carol computes her long-term private key x using s and her real password P.
	Example s=#13, v=??
		var v = xxx.generateVerifier(s,identity,password);
		The verifier is computed as v = g^x (mod N).
		g=#2, N=#1000001
		x=??
3. Carol generates a random number a, 1 < a < n, computes her ephemeral public key A = g^a, and sends it to Steve.
4. Steve generates his own random number b, 1 < b < n, computes his ephemeral public key B = v + g^b, and sends
	it back to Carol, along with the randomly generated parameter u.
5. Carol and Steve compute the common exponential value S = g^(ab + bux) using the values available to each of them.
	If Carol's password P entered in Step 2 matches the one she originally used to generate v, then both values of
	S will match.
6. Both sides hash the exponential S into a cryptographically strong session key.
7. Carol sends Steve M[1] as evidence that she has the correct session key. Steve computes M[1] himself and verifies
	that it matches what Carol sent him.
8. Steve sends Carol M[2] as evidence that he also has the correct session key. Carol also verifies M[2] herself,
	accepting only if it matches Steve's value.

This protocol is mostly the result of substituting the equations of Section 3.2.1 into the generic AKE protocol, adding explicit
flows to exchange information like the user's identity and the salt s. Both sides will agree on the session key S = g^(ab + bux) if all
steps are executed correctly. SRP also adds the two flows at the end to verify session key agreement using a one-way hash function.
Once the protocol run completes successfully, both parties may use K to encrypt subsequent session traffic.

Version 0.0.1

*/

// ValidUserName returns true if this is a non-email valid user name, like "admin"
// } else hdlr.ValidUserName(email) {
func (hdlr *AesSrpType) ValidUserName(un string) bool {
	fmt.Printf("ValidUserName(%s) = %v\n", un, (godebug.InArrayString(un, hdlr.NonEmailAccts) >= 0))
	if godebug.InArrayString(un, hdlr.NonEmailAccts) >= 0 { // if un == "admin" {
		return true
	}
	return false
}

// ============================================================================================================================================
//
//	mux.HandleFunc("/api/srp_login", respHandlerSRPLogin).Method("GET", "POST")                              // start login process (step 1)
//
// Input:
//		email - the username
//
// Output:
//		status - success
//		salt - from registration process
//		B - the public server key
//		t - 2nd salt for session ID
//		r - 3nd volatile ID - used during login process
//
// Save To D.B.
//
//		Step	Carol				Communication				Steve
//		1.								C -->					(lookup s, v from D.B. by username(C)) - send back s		/api/srp_login
//		2.		x = H(s, P)				<-- s, t, r				s salt, t is 2nd random, r 3rd sent back to client
//
func respHandlerSRPLogin(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		return
	}
	ps := rw.Ps

	godebug.Printf(dbSRP, "AT: %s\n", godebug.LF())

	// Determine 'I' the user identifier.  This can be email or DeviceID parameters.
	// Note: "username" can also be used if configured to do so.  // Add alternative of Username if configured to use Username instead of Email for 'I'
	email := ps.ByNameDflt("email", "") // this is 'I', or C
	UserName := email
	DeviceID := ps.ByNameDflt("DeviceID", "") // this is 'I'
	if hdlr.UserNameForRegister && DeviceID == "" {
		email = ps.ByNameDflt("username", "") // Configured to use username
		UserName = email
		if UserName != "" {
			godebug.Printf(dbRespHandlerSrpRegister, "AT: %s, UserName [%s]\n", godebug.LF(), UserName)

			if len(UserName) <= 6 {
				AnError(hdlr, www, req, 400, 1124, "UserName must be atleast 7 characters long")
				return
			}
			if ok, _ := regexp.MatchString("^[a-zA-Z]", UserName); !ok {
				AnError(hdlr, www, req, 400, 1125, "UserName must be start with a letter, a-z or A-Z")
				return
			}
			if ValidUUID(UserName) {
				AnError(hdlr, www, req, 400, 1126, "UserName can not be a UUID")
				return
			}
		}
	}
	if !hdlr.UserNameForRegister && DeviceID == "" {
		if ValidateEmail(email) { // emails are OK
		} else if ValidUUID(email) { // UUIDs are OK
		} else if hdlr.ValidUserName(email) { // Special Accounts are OK
		} else {
			AnError(hdlr, www, req, 400, 1127, "Invalid email address:"+email) // "code":"0025"
			return
		}
	} else if email == "" {
		email = DeviceID // validate DeviceID	-- 9 digits, verhoff and number
		if !verhoeff_algorithm.ValidateVerhoeff(DeviceID) {
			AnError(hdlr, www, req, 400, 1128, "DeviceID is not valid.")
			return
		}
	}
	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")
	sss := gosrp.GoSrpNew(email, hdlr.Bits)
	if db100 {
		sss.FixRandom("706a423a9b390a79a21a53b5ebb02bcf55be72fa4f9b151f03630558cf0309f9a6e5fe876ae82bd1e1e822ed46d08d353c9aaff3fbc5aa77f1d921e2150c6751")
	}

	t_id, _ := uuid.NewV4() // Generate 't' - the temporary session identifier
	s2 := t_id.String()
	if len(s2) == 0 {
		AnError(hdlr, www, req, 500, 1129, "Internal error - unable to generate 'r' key.")
		return
	}
	t_id, _ = uuid.NewV4() // Calculate 'r'
	s3 := t_id.String()
	if len(s3) < 25 {
		AnError(hdlr, www, req, 500, 1130, "Internal error - unable to generate 't' key.")
		return
	}
	s3 = s3[24:] // Make it smaller for quicker lookup in d.b.

	// ------------------------------------------------------------------------------------------------------------------------------
	// Begin the server by parsing the client credentials
	// ------------------------------------------------------------------------------------------------------------------------------
	// I, A, err := srp.ServerBegin(creds)

	// Fetch the 'salt', 'v' (verifier) and user_metadata from the user for this user - if no error then procede.
	// 'I' is the user identifier, email, username, or DeviceID
	salt, v, user_mdata, err := DbFetchUser(hdlr, rw, req, email, SandBoxPrefix)
	if err != nil {
		AnError(hdlr, www, req, 401, 1131, fmt.Sprintf(`%s - failed to fetch from database - you are not a registerd user.`, err))
		return
	}

	// If the user has not been confirmed then reject the user - not confirmed means no email confirm.
	if user_mdata["confirmed"] == "n" {
		AnError(hdlr, www, req, 401, 1132, `The account has not been confirmed.  Please confirm or register again and get a new confirmation email.`)
		return
	}

	// If the user is disabled then done - can not login - see admin
	if user_mdata["disabled"] == "y" {
		AnError(hdlr, www, req, 401, 1133, "The account has been disabled.  Please contact customer support (call them).")
		return
	}

	// If more than allowed number of failed login attempts or this is not a valid number in the database then - oops no login
	nf, err := strconv.ParseInt(user_mdata["n_failed_login"], 10, 64)

	if err != nil {
		user_mdata["n_failed_login"] = "0" //
		user_mdata["login_fail_time"] = "" // if n_failed_login > threshold then this is the time when to resume login trials
		dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), user_mdata)
	} else if int(nf) > hdlr.FailedLoginThreshold {
		t := time.Now()
		if user_mdata["login_fail_time"] != "" {
			x, err := time.Parse(time.RFC3339, user_mdata["login_fail_time"])
			if err == nil && t.Before(x) {
				AnError(hdlr, www, req, 401, 1134, "The account has been disabled.  Please contact customer support (call them).")
				return
			}
		}
	}

	// validate salt/v - empty once make core dump. -- Use validation from register.
	if !validSrpSalt(salt) {
		AnError(hdlr, www, req, 400, 1135, "Invalid salt.")
		return
	}
	if !validSrpV(v) {
		AnError(hdlr, www, req, 400, 1136, "Invalid 'v' verifier value.")
		return
	}

	sss.Setup(v, salt)

	// Save this info in d.b.
	mdata := make(map[string]string)

	SetSaltV(hdlr, www, req, mdata, salt, v) // Encrypt Salt,V if encryption is enabled

	mdata["email"] = email
	mdata["s2"] = s2
	mdata["s3"] = s3
	mdata["State"] = fmt.Sprintf("%d", sss.State)
	mdata["B"] = sss.XB_s
	mdata["bits"] = fmt.Sprintf("%d", hdlr.Bits)
	if hdlr.Bits != 2048 {
		s := godebug.LF()
		panic(s)
	}
	mdata["auth"] = "ip"

	//xyzzy dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:S:", SandBoxPrefix, s2), mdata) // Maybe just save this until we have .Key_s - the store under hash of that

	mdata["k"] = sss.Xk_s // Temporary Value - need not be stored
	mdata["b"] = sss.Xb_s

	if db100 {
		fmt.Fprintf(os.Stderr, "%sk=[%s], %s%s\n", MiscLib.ColorMagenta, sss.Xk_s, godebug.LF(), MiscLib.ColorReset)
		fmt.Fprintf(os.Stderr, "%sb=[%s] B=[%s]%s, %s\n", MiscLib.ColorMagenta, sss.Xb_s, sss.XB_s, godebug.LF(), MiscLib.ColorReset)
	}

	dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:V:", SandBoxPrefix, s3), mdata)
	if dbSRP {
		fmt.Printf("**** Save temporary key under [%s], %s\n", s3, godebug.LF())
	}

	//expire := time.Now().AddDate(0, 0, hdlr.CookieExpireInXDays) // Years, Months, Days==2 // Xyzzy501 - xyzzy2016 - should be a config - on how long to keep cookie
	//cookie := http.Cookie{Name: "LoginAuthToken", Value: "x", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400, Secure: false, HttpOnly: true}
	expire := time.Now().AddDate(0, 0, -1) // Years, Months, Days==2 // Xyzzy501 - should be a config - on how long to keep cookie
	cookie := http.Cookie{Name: "LoginAuthToken", Value: "x", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 0, Secure: false, HttpOnly: true}
	http.SetCookie(www, &cookie)

	// Delete the cookie
	expire = time.Now().AddDate(0, 0, -1) // Years, Months, Days==-1
	cookie = http.Cookie{Name: "pageMarkerCookie", Value: "", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 0, Secure: false, HttpOnly: true}
	http.SetCookie(www, &cookie)

	www.Header().Set("Content-Type", "application/json")
	io.WriteString(www, fmt.Sprintf(`{"status":"success","salt":%q,"t":%q,"r":%q,"bits":%d,"B":%q,"f2":%q}`, salt, s2, s3, hdlr.Bits, sss.XB_s, hdlr.TwoFactorRequired))

}

// ============================================================================================================================================
// Call:
//		muxEnc.HandleFunc("/api/srp_challenge", respHandlerSRPChallenge).Method("GET", "POST")                  // start login process (step 2)
//
// Input:
//		A - client public key
//		r - client temporary ID for pulling back validation stuff
//		email - the user, 'I'
//
// Output:
//		status - success/fail
//		m1 - Server validation value
//		B - server public key
//		HAMK(m2) - -- Don't really use this - do a hash and compare instead - like the SRP6a specifies
//
func respHandlerSRPChallenge(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		return
	}
	ps := rw.Ps

	if dbSRP {
		fmt.Printf("AT: %s\n", godebug.LF())
	}
	var ok1, ok2, ok3, ok4 bool

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")
	A := ps.ByNameDflt("A", "")  // this is 'A'
	s3 := ps.ByNameDflt("r", "") // this is 'r'
	sss := gosrp.GoSrpNew("", hdlr.Bits)
	if db100 {
		sss.FixRandom("706a423a9b390a79a21a53b5ebb02bcf55be72fa4f9b151f03630558cf0309f9a6e5fe876ae82bd1e1e822ed46d08d353c9aaff3fbc5aa77f1d921e2150c6751")
	}

	mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:V:", SandBoxPrefix, s3))
	if !ok {
		AnError(hdlr, www, req, 500, 1137, "Temporary data lost.  Please try login again.")
		return
	}

	// s2 := mdata["s2"]
	// email := mdata["email"]
	// sss.XI_s = email
	mdata["A"] = A

	av := func(name string) (s string, bbb *big.Int, ok bool) {
		s = mdata[name]
		bbb, ok = big.NewInt(0).SetString(s, 16)
		if !ok {
			fmt.Printf(`{"msg":"Failed to recover ANY from temporary key for [%s]"}\n`, name)
		}
		return
	}

	sss.Salt_s, sss.Salt, ok = av("salt")
	sss.Xv_s, sss.Xv, ok1 = av("v")
	sss.Xk_s, sss.Xk, ok2 = av("k") // Temporary need not be recoverd
	sss.Xb_s, sss.Xb, ok3 = av("b")
	sss.XB_s, sss.XB, ok4 = av("B")
	if !ok || !ok1 || !ok2 || !ok3 || !ok4 {
		AnError(hdlr, www, req, 500, 1138, "Temporary data corrupted.  Please try login again.")
		return
	}

	sss.State = 2
	sss.IssueChallenge(A)
	mdata["m1"] = sss.XM1_s // Server value for M
	mdata["key"] = sss.Key_s
	mdata["state"] = "3"
	mdata["HAMK"] = sss.XHAMK_s

	godebug.Printf(dbSRP, "\nSESSION KEY - Resume Login: K <critical> <critical> %s type=%T len=%d\n\n", sss.Key_s, sss.Key_s, len(sss.Key_s))

	dataStore.RUpdValue(hdlr, rw, SandBoxKey("srp:V:", SandBoxPrefix, s3), mdata)

	//xyzzy dataStore.RUpdValue(hdlr, rw, SandBoxKey("srp:S:", SandBoxPrefix, s2), mdata)

	www.Header().Set("Content-Type", "application/json")
	//secure := false
	//if req.TLS != nil {
	//	secure = true
	//}
	//expire := time.Now().AddDate(0, 0, -5000)
	//cookie := http.Cookie{Name: "LoginAuthCookie", Value: "deleted", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 0, Secure: secure, HttpOnly: false}
	//http.SetCookie(www, &cookie)
	fmt.Fprintf(www, `{"status":"success","B":"%s","Bits":%d,"M1":%q}`, sss.XB_s, hdlr.Bits, sss.XM1_s)

}

// ============================================================================================================================================
//
//	muxEnc.HandleFunc("/api/srp_validate", respHandlerSRPValidate).Method("GET", "POST")                    // start login process (step 3)
//
func respHandlerSRPValidate(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1139, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	if dbSRP {
		fmt.Printf("AT: %s\n", godebug.LF())
	}
	var ok1, ok2, ok3, ok4, ok5, ok6 bool

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")
	// fingerprint := ps.ByNameDflt("fingerprint", "")
	// _ = fingerprint

	ClientM1 := ps.ByNameDflt("M1", "") // this is 'M1'
	s3 := ps.ByNameDflt("r", "")        // this is 'r'
	sss := gosrp.GoSrpNew("", hdlr.Bits)
	if db100 {
		sss.FixRandom("706a423a9b390a79a21a53b5ebb02bcf55be72fa4f9b151f03630558cf0309f9a6e5fe876ae82bd1e1e822ed46d08d353c9aaff3fbc5aa77f1d921e2150c6751")
	}

	stayLoggedIn := ps.ByNameDflt("stayLoggedIn", "false") // true == true, false == false
	stayLoggedInTf, _ := lib.ParseBool(stayLoggedIn)

	if dbFingerprint {
		fmt.Printf("\n\n")
		fmt.Printf("/////////////////////////////////////////////////////////////////////////////////////////////////////\n")
		fmt.Printf("// stayLoggedIn = %s, %s\n", stayLoggedIn, godebug.LF())
		fmt.Printf("/////////////////////////////////////////////////////////////////////////////////////////////////////\n")
		fmt.Printf("\n\n")
	}

	mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:V:", SandBoxPrefix, s3))
	if !ok {
		AnError(hdlr, www, req, 401, 1140, "Temporary data corrupted.  Please try login again.")
		return
	}
	email, ok := mdata["email"]
	if !ok {
		fmt.Printf("***************** no email in mdata\n")
	}

	s2 := mdata["s2"]
	mdata["ClientM1"] = ClientM1

	av := func(name string) (s string, bbb *big.Int, ok bool) {
		s = mdata[name]
		bbb, ok = big.NewInt(0).SetString(s, 16)
		if !ok {
			fmt.Printf(`{"msg":"Failed to recover %s from temporary key for [%s]"}\n`, name, s3)
		}
		return
	}

	av2 := func(name string) (s string, ok bool) {
		s, ok = mdata[name]
		if !ok {
			fmt.Printf(`{"msg":"Failed to recover %s from temporary key for [%s]"}\n`, name, s3)
		}
		return
	}

	// PasswordSV - encpyt before saving
	sss.Salt_s, sss.Salt, ok = av("salt")
	sss.Xv_s, sss.Xv, ok1 = av("v")
	sss.Xk_s, sss.Xk, ok2 = av("k") // Temporary need not be recoverd
	sss.Xb_s, sss.Xb, ok3 = av("b")
	sss.XB_s, sss.XB, ok4 = av("B")
	sss.XA_s, sss.XA, ok4 = av("A")
	sss.XM1_s, ok5 = av2("m1")
	sss.Key_s, ok6 = av2("key")
	if !ok || !ok1 || !ok2 || !ok3 || !ok4 || !ok5 || !ok6 {
		AnError(hdlr, www, req, 500, 1141, "Temporary data corrupted.  Please try login again.")
		return
	}

	key := sss.Key_s
	// _ = key

	if db100 {
		fmt.Fprintf(os.Stderr, "%s<<<Critical>>> srp_validate: Steve: Shared Key=[%s], %s%s\n", MiscLib.ColorMagenta, key, godebug.LF(), MiscLib.ColorReset)
	}

	sss.State = 3
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// validate at this point that m1 is good
	// This is where we should check client 'M' v.s. Server 'M' to verify that they are the same.
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	auth, m2 := sss.CalculateM2(ClientM1) // Calculate m2 using m1 - validate that ClientM1 == M1(server)
	mdata["state"] = "4"

	fmt.Fprintf(os.Stderr, "%sServer m2 [%s], %s%s\n", MiscLib.ColorYellow, m2, godebug.LF(), MiscLib.ColorReset)

	dataStore.RUpdValue(hdlr, rw, SandBoxKey("srp:V:", SandBoxPrefix, s3), mdata)

	user_mdata, user_err := DbFetchUserMdata(hdlr, rw, email, SandBoxPrefix)
	if user_err != nil {
		AnError(hdlr, www, req, 401, 1142, "Invalid email -unable to associate user with email-.")
		return
	}

	delete(mdata, "b")
	delete(mdata, "k")
	delete(mdata, "state")

	validation_secret := GenerateValidationSecret() //
	fmt.Printf("validation_secret Generation Time = [%s] AT: %s\n", validation_secret, godebug.LF())

	have_anon := true
	num_login_times := 0
	// backup_keys := ""
	if auth {

		privs, ok3 := user_mdata["privs"]
		if !ok3 {
			privs = hdlr.NewUserPrivs
		}

		mdata["auth"] = "y"
		id0, _ := uuid.NewV4()
		id := id0.String()
		// fmt.Printf("Saving Email Address: %s\n", email)
		SaveInitData(hdlr, rw, s2, SandBoxPrefix, sss.Key_s, email, id, privs)
		goftlmux.AddValueToParams("$username$", email, 'i', goftlmux.FromAuth, &ps)
		goftlmux.AddValueToParams("$auth_key$", id, 'i', goftlmux.FromAuth, &ps)
		SetLoggedIn(hdlr, rw, email, SandBoxPrefix, id)

		// Set Cookie at this point -- if a "anon-user" skip this -- // xyzzyStayLoggedIn2 what users should set "stayLoggedIn" privilage
		fmt.Printf("Validate email=%s privs=%s, %s\n", mdata["email"], mdata["privs"], godebug.LF())
		fmt.Printf("user_mdata=%s\n", lib.SVarI(user_mdata))
		// if mdata["privs"] == "user" {

		var LoginAuthCookie string
		if privs == "user" {

			LoginAuthCookie = ps.ByNameDflt("LoginAuthCookie", "")
			id0, _ = uuid.NewV4()
			cookieValue := id0.String()
			expire := time.Now().AddDate(0, 0, hdlr.CookieExpireInXDays2) // Years, Months, Days==2 // Xyzzy501 - should be a config - on how long to keep cookie
			secure := false
			if req.TLS != nil {
				secure = true
			}
			cookie := http.Cookie{Name: "LoginAuthCookie", Value: cookieValue, Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: LoginAuthCookieLife, Secure: secure, HttpOnly: false}
			// if true { // session-cookie
			if !stayLoggedInTf {
				cookie = http.Cookie{Name: "LoginAuthCookie", Value: cookieValue, Path: "/", MaxAge: 86400, Secure: secure, HttpOnly: false}
			}
			if LoginAuthCookie != "" {
				if !hdlr.CookieEmailMatch(rw, email, LoginAuthCookie, SandBoxPrefix) {
					http.SetCookie(www, &cookie)
				} else {
					cookieValue = LoginAuthCookie
				}
			} else {
				http.SetCookie(www, &cookie)
			}

			// xyzzyStayLoggedIn, Auth = 's'
			cookieHash := HashStrings.Sha256(cookieValue + ":" + validation_secret)
			mdata["LoginAuthCookie"] = SaveAsList(mdata["LoginAuthCookie"], cookieValue)
			mdata["LoginHashCookie"] = SaveAsList(mdata["LoginHashCookie"], cookieHash)
			mdata["validation_secret"] = validation_secret

			cookie2 := http.Cookie{Name: "LoginHashCookie", Value: cookieHash, Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: LoginHashCookieLife, Secure: secure, HttpOnly: true}
			http.SetCookie(www, &cookie2)

			ip := lib.GetIP(req)
			SaveCookieAuth(hdlr, rw, cookieValue, SandBoxPrefix, ip, email, cookieHash, id, privs) // xyzzy - suspect this may be bad

		}

		have_anon = true
		vvv, err := DbGetString(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, LoginAuthCookie))
		if err != nil || vvv == "" {
			have_anon = false
		}

		DeviceID, _ := user_mdata["DeviceID"] // xyzzyDeviceID -
		deviceIDList, _ := user_mdata["DeviceIDList"]
		w := GenerateRandomOneTimeKey("8")
		OneTimeKey := w

		// "Pre2Factor":               { "type":[ "string" ], "default":"p2f:" },
		fmt.Printf("\n-=- -=- -=- -=- -=- -=- -=- -=- -=- -=-\n")
		fmt.Printf("Setting Device ID: %s with key %s\n", hdlr.Pre2Factor+DeviceID, OneTimeKey)
		fmt.Printf("Setting Device ID List: %s with key %s\n", hdlr.Pre2Factor+deviceIDList, OneTimeKey)
		fmt.Printf("stayLoggedIn flag = %s\n", stayLoggedIn)
		fmt.Printf("\n-=- -=- -=- -=- -=- -=- -=- -=- -=- -=-\n")

		// xyzzyDeviceID - Will this need to be a "key" for each DeviceID -- if so should we limit # of valid device ides to say 5/10? - configurable.

		// DeviceID can now be tied to a One Time Key, from there
		// One Time Key can be ted to an email (user)
		// var TwoFactorLife = 5*60 + 60             // 5 * 60 seconds = 5 min + 1 minute of grace	- how long is a temporay 2 factor key good for - 5 min
		// "Pre2Factor":               { "type":[ "string" ], "default":"p2f:" },
		DbSetExpire(hdlr, rw, SandBoxKey(hdlr.Pre2Factor, SandBoxPrefix, DeviceID), OneTimeKey, hdlr.TwoFactorLife) // used in respHandlerGet2FactorFromDeviceID - when 2fa-client gets ID
		DbSetExpire(hdlr, rw, SandBoxKey(hdlr.Pre2Factor, SandBoxPrefix, OneTimeKey), email, hdlr.TwoFactorLife)    // use in respHandlerValid2Factor - when loggin in

		DeviceIDList := make([]DeviceIDType, 0, 10)
		err = json.Unmarshal([]byte(deviceIDList), &DeviceIDList)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%sInvalid device id list [%s], error=%s, %s %s\n", MiscLib.ColorRed, deviceIDList, err, godebug.LF(), MiscLib.ColorReset)
		} else {
			for _, di := range DeviceIDList {
				DbSetExpire(hdlr, rw, SandBoxKey(hdlr.Pre2Factor, SandBoxPrefix, di.DeviceID), OneTimeKey, hdlr.TwoFactorLife) // used in respHandlerGet2FactorFromDeviceID - when 2fa-client gets ID
			}
		}

	} else {

		mdata["auth"] = "n"
		SaveInitFailedLogin(hdlr, rw, s2, SandBoxPrefix)

	}

	more_backup_keys := false
	first_login := false
	di, deviceIDList, ki, hashed, oe, kiD, hashedD := "", "", "", "", "", "", ""
	var URole RolesWithBitMask
	www.Header().Set("Content-Type", "application/json")
	if !auth {
		AnError(hdlr, www, req, 401, 1143, "Failed to login.  Incorrect username/email or password.")

		t := time.Now()
		nf, err := strconv.ParseInt(user_mdata["n_failed_login"], 10, 64)
		if err != nil {
			nf = 1
		}
		nf++
		user_mdata["n_failed_login"] = fmt.Sprintf("%d", nf)
		then := t.Add(5 * time.Minute) // xyzzy2016 - should be configurable -- 1. Configurable time delay.  TODO_4001
		tss := then.Format(time.RFC3339)
		user_mdata["login_fail_time"] = tss // if n_failed_login > threshold then this is the time when to resume login trials
		dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), user_mdata)

		return
	} else {

		// # of logins --------------------------------------------------------------------------------------------
		num_login_times_s, ok := user_mdata["num_login_times"]
		if ok {
			n, err := strconv.ParseInt(num_login_times_s, 10, 64)
			if err == nil {
				num_login_times = int(n)
			}
		}
		user_mdata["num_login_times"] = fmt.Sprintf("%d", num_login_times+1)
		if num_login_times == 0 {
			first_login = true
		}

		user_mdata["validation_secret"] = validation_secret

		// # count backup keys -------------------------------------------------------------------------------------
		backup_keys, ok := user_mdata["backup_one_time_keys"]
		if user_mdata["acct_type"] == "DeviceID" {
			backup_keys, ok = user_mdata["offline_one_time_keys"]
		}
		if ok {
			keys := strings.Split(backup_keys, ",")
			n_keys := len(keys)
			if n_keys <= 5 {
				more_backup_keys = true
			}
		} else {
			user_mdata["backup_one_time_keys"] = ""
		}

		// ---------------------------------------------------------------------------------------------------------
		if hdlr.TwoFactorRequired == "y" {
			user_mdata["auth"] = "P" // Pending 2Fa validation
		}

		// ---------------------------------------------------------------------------------------------------------
		URole = RolesWithBitMask{
			Name:    user_mdata["privs"],
			BitMask: hdlr.secRnH[user_mdata["privs"]],
		}

		oe = user_mdata["owner_email"]

		// ---------------------------------------------------------------------------------------------------------
		// if first login then send out DeviceID and Backup keys
		// if first_login || (user_mdata["acct_type"] == "DeviceID" && more_backup_keys) {
		if first_login || user_mdata["acct_type"] == "DeviceID" {
			di = user_mdata["DeviceID"]                             // xyzzyDeviceID - Multiple!
			deviceIDList = user_mdata["DeviceIDList"]               // xyzzyDeviceID - Multiple!
			salt, _ /*v*/ := GetSalt(hdlr, www, req, user_mdata)    //
			ki, hashed = GenBackupKeys(hdlr, salt, "9", www, req)   //	// User account one time backup keys
			user_mdata["backup_one_time_keys"] = hashed             //
			kiD, hashedD = GenBackupKeys(hdlr, salt, "4", www, req) //	// Device One Time Backup Keys
			user_mdata["offline_one_time_keys"] = hashedD           //
			if user_mdata["acct_type"] == "DeviceID" {
				ki = kiD
			}
		}

		if user_mdata["acct_type"] == "DeviceID" {
			di = ""
			deviceIDList = "[]"
			oe = ""
		}

		dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), user_mdata)

	}

	// if first login then send out DeviceID and Backup keys

	tn := time.Now()                                     // Notify client that it's login will timeout in this amount of time.
	tn = tn.Add((LoginAuthCookieLife - 2) * time.Second) //
	tnss := tn.Format(time.RFC3339)                      // Format for client.

	// Delete the cookie
	expire := time.Now().AddDate(0, 0, -1) // Years, Months, Days==-1
	cookie := http.Cookie{Name: "pageMarkerCookie", Value: "", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 0, Secure: false, HttpOnly: true}
	http.SetCookie(www, &cookie)
	cookie = http.Cookie{Name: "linkMarkerCookie", Value: "", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 0, Secure: false, HttpOnly: true}
	http.SetCookie(www, &cookie)

	// Fix - bad for removing values if no 2fa -- would be much better to never generate and never save/return this at all... but...
	if hdlr.TwoFactorRequired == "n" {
		rv := LoginRetrunValueNo2fa{
			Status:            "success",
			M2:                m2,
			FirstLogin:        first_login,
			TwoFactorRequired: "n", //  hdlr.TwoFactorRequired == "n"
			UserRole:          URole,
			OwnerEmail:        oe,
			LoginLastsTill:    tnss, // Login will time out at this time.
			LoginLastsSeconds: LoginAuthCookieLife - 2,
			RealName:          user_mdata["RealName"],
			PhoneNo:           user_mdata["PhoneNo"],
			FirstName:         user_mdata["FirstName"],
			MidName:           user_mdata["MidName"],
			LastName:          user_mdata["LastName"],
			UserName:          user_mdata["UserName"],
			XAttrs:            user_mdata["XAttrs"],
			HaveAnon:          have_anon,
		}

		fmt.Fprintf(www, lib.SVar(rv))

	} else {

		DeviceIDList := make([]DeviceIDType, 0, 10)
		err := json.Unmarshal([]byte(deviceIDList), &DeviceIDList)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%sInvalid device id list [%s], error=%s, %s %s\n", MiscLib.ColorRed, deviceIDList, err, godebug.LF(), MiscLib.ColorReset)
		}

		rv := LoginRetrunValue{
			Status:            "success",
			M2:                m2,
			FirstLogin:        first_login,
			MoreBackupKeys:    more_backup_keys,
			TwoFactorRequired: hdlr.TwoFactorRequired,
			UserRole:          URole,
			DeviceID:          di, // xyzzyDeviceID
			DeviceIDList:      DeviceIDList,
			BackupKeys:        ki,
			OwnerEmail:        oe,
			LoginLastsTill:    tnss, // Login will time out at this time.
			LoginLastsSeconds: LoginAuthCookieLife - 2,
			RealName:          user_mdata["RealName"],
			PhoneNo:           user_mdata["PhoneNo"],
			FirstName:         user_mdata["FirstName"],
			MidName:           user_mdata["MidName"],
			LastName:          user_mdata["LastName"],
			UserName:          user_mdata["UserName"],
			XAttrs:            user_mdata["XAttrs"],
			HaveAnon:          have_anon,
		}
		fmt.Fprintf(www, lib.SVar(rv))
	}

}

// ============================================================================================================================================
//
//	mux.HandleFunc("/api/cipher", respHandlerCipher).Method("GET", "POST")                                   // Decrypt and rewrite request for other handlers
//
// ----------------------------------- cipher -------------------------------------------------------------
//
// Take an encrytped requests and convert it into a non-encrypted request.  When the request returns a
// reponse re-encrypt that before returing it to the client.
//
func respHandlerCipher(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1144, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := &rw.Ps

	// xyzzyEEE remove ps.Stuff - where it is reserved for internal use. -- "$is_logged_in$" for example!

	// fmt.Printf("Params Are: %s AT %s\n", ps.DumpParam(), godebug.LF())

	if dbCipher {
		fmt.Printf("Cipher Called AT: %s\n", godebug.LF())
		fmt.Printf("Method: %s\n", req.Method)
	}
	var rv string
	var plaintext []byte
	var key []byte

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	www.Header().Set("Content-Type", "application/json")

	data := ps.ByNameDflt("data", "{}")
	tt := ps.ByNameDflt("t", "")
	if data == "{}" || tt == "" {
		AnError(hdlr, www, req, 400, 1145, "Invalid input data.  Missing or invalid 't' parameter.")
		return
	}

	email, err := UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
	if err != nil { // check user exists
		AnError(hdlr, www, req, 400, 1146, "Failed to find user. Invalid input Email/DeviceID/UserName.")
		return
	}

	if dbCipher2 {
		fmt.Printf("// ////////////////////////////////////////////////////////////////////////////////////////////////////////\n")
		fmt.Printf("   AT: %s\n", godebug.LF())
		fmt.Printf("   t = [%s]\n", tt)
		fmt.Printf("   data = [%s]\n", data)
		fmt.Printf("// ////////////////////////////////////////////////////////////////////////////////////////////////////////\n")
	}

	// Xyzzy -    a. JSON decode it -> SJCL struct
	// func ConvertSJCL(file string) (eBlob SJCL_DataStruct, err error, msg string) {
	encData, err, msg0 := sjcl.ConvertSJCL(data)
	if err != nil {
		AnError(hdlr, www, req, 400, 1147, fmt.Sprintf("Error(0201): error parsing SJCL data - %s %s", err, msg0))
		return
	}

	if dbCipher2 {
		fmt.Printf("// ////////////////////////////////////////////////////////////////////////////////////////////////////////\n")
		fmt.Printf("   AT: %s\n", godebug.LF())
		fmt.Printf("   data converted to structure= [%+v]\n", encData)
		fmt.Printf("// ////////////////////////////////////////////////////////////////////////////////////////////////////////\n")
	}

	// Lookup user based on tt - the session ID - in Redis
	Password, tSalt, tKey, tIter, tKeySize, tEmail, Session := GetKeyData(hdlr, rw, tt, SandBoxPrefix)

	//	if dbCipher2 {
	//		fmt.Printf("// ////////////////////////////////////////////////////////////////////////////////////////////////////////\n")
	//		fmt.Printf("AT: %s Pw[%s] (shared key should match app_js_config.enc_key\n", godebug.LF(), Password)
	//		fmt.Printf("   tSalt    = %x\n", tSalt)
	//		fmt.Printf("   tKey     = %x\n", tKey)
	//		fmt.Printf("   tIter    = %v\n", tIter)
	//		fmt.Printf("   tKeySize = %v\n", tKeySize)
	//		fmt.Printf("   tEmail   = %v\n", tEmail)
	//		fmt.Printf("   Session  = %v\n", Session)
	//		fmt.Printf("// ////////////////////////////////////////////////////////////////////////////////////////////////////////\n")
	//	}

	// --------------------------------------------------------------------------------------------------------------------------------
	// Decrypt "data"
	// --------------------------------------------------------------------------------------------------------------------------------
	// Inputs
	// 		Password, tSalt, tKey, tIter, tKeySize, tEmail, Session := GetKeyData(hdlr, rw, tt, SandBoxPrefix)
	//		tt 										// the  session key, compared to signed AdditionalData
	// Outputs
	//		key
	//		plaintext
	// Misc
	//		dbCipher -> debugFlag1					// Debuging flag
	//

	plaintext, key, err = DecryptData(hdlr, rw, www, req, SandBoxPrefix, Password, tEmail, tSalt, &encData, tIter, tKeySize, tKey, Session, tt, dbCipher, GetDebugFlag("DumpEncryptedRequest"))
	if err != nil {
		return
	}

	//	{
	//		// encData.Salt.Debug_hex(db1, "salt")
	//		// encData.InitilizationVector.Debug_hex(db1, "Initilization Vector")
	//
	//		if dbCipher {
	//			fmt.Printf("tSalt [%x] encData.Salt [%x] tKey [%x] tIter %d %d tKeySize %d %d\n",
	//				tSalt, string(encData.Salt), tKey, tIter, encData.Iter, tKeySize, encData.KeySize)
	//		}
	//		if tSalt == "" || tKey == "" || tSalt != string(encData.Salt) || tIter != encData.Iter || tKeySize != encData.KeySize {
	//			if dbCipher2 {
	//				fmt.Printf("KEY GEN: password[%s] salt[%x] iter[%d] keysize[%d]\n", Password, encData.Salt, encData.Iter, encData.KeySizeBytes)
	//			}
	//			// Generete the "key" using the shared secret password and other parameters.
	//			key = pbkdf2.Key([]byte(Password), encData.Salt, encData.Iter, encData.KeySizeBytes, sha256.New)
	//			// debug_hex("key", key)
	//			SaveKeyData(hdlr, rw, tt, SandBoxPrefix, Password, string(encData.Salt), string(key), encData.Iter, encData.KeySize)
	//		} else {
	//			key = []byte(tKey)
	//		}
	//
	//		if GetDebugFlag("DumpEncryptedRequest") {
	//			fmt.Printf("key is [%x], salt is [%x], %s\n", key, tSalt, godebug.LF())
	//			fmt.Printf("At: %s, Ps=%s, req=%s\n", godebug.LF(), rw.Ps.DumpParamDB(), lib.SVarI(req)) // XyzzyDumpData
	//		}
	//
	//		cb, err := aes.NewCipher(key) // var cb cipher.Block
	//		if err != nil {
	//			AnError(hdlr, www, req, 400, 53, fmt.Sprintf("Error(0053): unable to setup AES:%s", err))
	//			return
	//		}
	//
	//		nonce, nlen := GetNonce(encData)
	//
	//		// b. Decrypt the "ct" - validate it.
	//		authmode, err := aesccm.NewCCM(cb, encData.TagSizeBytes, nlen) // var authmode cipher.AEAD
	//		if err != nil {
	//			AnError(hdlr, www, req, 400, 54, fmt.Sprintf("Error(0054): unable to setup CCM:%s", err))
	//			return
	//		}
	//
	//		plaintext, err = authmode.Open(nil, nonce, encData.CipherText, encData.AdditionalData)
	//		if err != nil {
	//			AnError(hdlr, www, req, 400, 55, fmt.Sprintf("Error(0055): decrypting or authenticating using CCM:%s", err))
	//			return
	//		}
	//
	//		// tt should match 1st part of AD - authenticated data
	//		// Session["one-time-key"] - should match 2nd part IFF one-time-auth is true
	//		// Split on ','
	//		OneTimeKey := "x" // "x" will not match to any hex number
	//		AdditionalData := string(encData.AdditionalData)
	//		fmt.Printf("Additional Data = %s, %T\n", AdditionalData, encData.AdditionalData)
	//		tt_ad := ""
	//		if strings.Index(AdditionalData, ",") >= 0 {
	//			tt_v := strings.Split(AdditionalData, ",")
	//			tt_ad = tt_v[0]
	//			if len(tt_v) > 1 {
	//				OneTimeKey = tt_v[1]
	//			}
	//			fmt.Printf("One Time Key from AD = [%s], %s\n", OneTimeKey, godebug.LF())
	//		} else {
	//			tt_ad = AdditionalData
	//		}
	//
	//		// Verify that the AdditonalData[first part, t] matches with the past session 't' value.
	//		if tt_ad != tt {
	//			AnError(hdlr, www, req, 400, 56, fmt.Sprintf("Error(0056): AdditionalData failed to match session key, Error:%s", err))
	//			return
	//		}
	//
	//		// xyzzy - Questionable
	//		if hdlr.TwoFactorRequired == "y" && Session["auth"] == "P" { // using one time key - then validated that OneTimeKey is a match
	//			AnError(hdlr, www, req, 400, 9096, "Error(0096): In 2FA mode, but did not validate 2nd factor.")
	//			return
	//		}
	//
	//		// xyzzy - Questionable
	//		if hdlr.TwoFactorRequired == "y" && Session["auth"] == "y" { // using one time key - then validated that OneTimeKey is a match
	//
	//			savedKey, ok := Session["$saved_one_time_key_hashed$"]
	//			if !ok {
	//				AnError(hdlr, www, req, 400, 9056, fmt.Sprintf("Error(0056): AdditionalData failed to match session key - one time key not saved - can not match, %s", err))
	//				return
	//			}
	//
	//			if savedKey != OneTimeKey {
	//				AnError(hdlr, www, req, 400, 8056, fmt.Sprintf("Error(0056): AdditionalData failed to match session key - one time key did match, %s", err))
	//				return
	//			}
	//		}
	//
	//	}

	PlainTextData := make(map[string]interface{})
	getSVal := func(name, dflt string) (rv string, ok bool) {
		t, ok := PlainTextData[name]
		if !ok {
			if dflt != "" {
				t = dflt
			} else {
				AnError(hdlr, www, req, 500, 1155, fmt.Sprintf("Invalid method/url %s passed", rv))
				return
			}
		} else {
			rv, ok = t.(string)
			if !ok {
				AnError(hdlr, www, req, 500, 1156, fmt.Sprintf("Invalid method/url %s passed", rv))
				return
			}
		}
		ok = true
		return
	}

	if dbCipher3 {
		fmt.Printf("\n----------------------------------------------------------------------------------------\nDecrypted Data: -->>%s<<--, %s\n", string(plaintext), godebug.LF())
	}
	err = json.Unmarshal(plaintext, &PlainTextData)
	if err != nil {
		if dbCipher3 {
			fmt.Printf("Fail in unmarshal\n")
		}
		es := jsonSyntaxErroLib.GenerateSyntaxError(string(plaintext), err)
		logrus.Errorf("Error: Invlaid JSON Error:\n%s\n", es)
		AnError(hdlr, www, req, 500, 1157, "invalid method passed - interal error")
		return
	} else {
		if dbCipher3 {
			fmt.Printf("Success in unmarshal, %s\n", lib.SVarI(PlainTextData))
		}
	}
	if dbCipher3 {
		fmt.Printf("\n----------------------------------------------------------------------------------------\n")
	}

	// -------------------------------------------------------------------------------------------------------------------------
	// -------------------------------------------------------------------------------------------------------------------------
	// -------------------------------------------------------------------------------------------------------------------------
	// -------------------------------------------------------------------------------------------------------------------------
	// -------------------------------------------------------------------------------------------------------------------------

	if dbCipher4 {
		fmt.Printf("At: %s, req=%s\n", godebug.LF(), lib.SVarI(req))
	}
	if rw, ok := www.(*goftlmux.MidBuffer); ok {

		if dbCipher4 {
			fmt.Printf("At: %s\n", godebug.LF())
		}
		newMethod, ok := getSVal("Method", "GET")
		if !ok {
			return
			//plaintext = []byte(newMethod)
			//goto next
		}
		if dbCipher4 {
			fmt.Printf("At: %s\n", godebug.LF())
		}
		if newMethod == "" {
			newMethod = "GET"
		}
		switch newMethod {
		case "GET", "PUT", "POST", "DELETE", "DEL", "HEAD", "OPTIONS", "PATCH":
			if dbCipher4 {
				fmt.Printf("At: %s\n", godebug.LF())
			}
			newPath, ok := getSVal("URL", "")
			if !ok {
				return
			}

			godebug.Printf(dbDumpURL, "\n%s\nURL: %s, %s\n\n%s\n\n", strings.Repeat("-=- ", 20), newPath, godebug.LF(), strings.Repeat("-=- ", 20))
			// fmt.Printf("\n\nURL Problem: URL = ---[%s]---, %s\n\n", newPath, godebug.LF())

			parsedPath, err := url.Parse(newPath)
			if err != nil {
				AnError(hdlr, www, req, 500, 1158, fmt.Sprintf("Error(4052): Unable to parse the URL (%s) that was supplied, error: %s", newPath, err))
				return
			}

			// fmt.Printf("\nURL Problem: URL parsed %s, %s\n\n", lib.SVarI(parsedPath), godebug.LF())

			req.URL.Path = parsedPath.Path
			req.URL.RawQuery = parsedPath.RawQuery
			req.Method = newMethod
			req.RequestURI = parsedPath.Path

			if dbCipher4 {
				fmt.Printf("At: %s\n", godebug.LF())
			}
			for ii, vv := range Session {
				// skip over Method, URL, etc
				// if godebug.InArrayString(ii, []string{"Method", "URL", "username", "$user_id$", "$privs$", "$username$", "$auth_key$"}) >= 0 {
				if ReservedIDs[ii] {
					continue
				}
				ss := fmt.Sprintf("%s", vv)
				goftlmux.AddValueToParams(ii, ss, 's', goftlmux.FromInject, &rw.Ps)
			}

			// Take additional parameters and inject them into Ps
			if dbCipher4 {
				fmt.Printf("At: %s\n", godebug.LF())
			}
			for ii, vv := range PlainTextData {
				// skip over Method, URL, etc
				// if godebug.InArrayString(ii, []string{"Method", "URL", "username", "$user_id$", "$privs$", "$username$", "$auth_key$"}) >= 0 {
				if ReservedIDs[ii] {
					continue
				}
				//ww, ok := vv.(string)
				//if !ok {
				//	AnError(hdlr, www, req, 500, 130, fmt.Sprintf("Invalid type for %s - should be string, got %T", ii, vv))
				//	return
				//	// goto next
				//}
				//goftlmux.AddValueToParams(ii, ww, 'e', goftlmux.FromInject, &rw.Ps)
				Value := ""
				switch vv.(type) {
				case bool:
					Value = fmt.Sprintf("%v", vv)
				case float64:
					Value = fmt.Sprintf("%v", vv)
				case int64:
					Value = fmt.Sprintf("%v", vv)
				case int32:
					Value = fmt.Sprintf("%v", vv)
				case time.Time:
					Value = fmt.Sprintf("%v", vv)
				case string:
					Value = fmt.Sprintf("%v", vv)
				default:
					Value = fmt.Sprintf("%s", lib.SVar(vv))
				}
				goftlmux.AddValueToParams(ii, Value, 'e', goftlmux.FromInject, &rw.Ps)
			}

			// fmt.Printf("At: %s\n", godebug.LF())
			goftlmux.AddValueToParams("username", tEmail, 'i', goftlmux.FromAuth, &rw.Ps)
			// fmt.Printf("\n\nSession: %s, %s\n\n", lib.SVarI(Session), lib.LF())
			// goftlmux.AddValueToParams("$privs$", "user", 'i', goftlmux.FromAuth, &rw.Ps)
			for kk, vv := range Session {
				ss := fmt.Sprintf("%s", vv)
				goftlmux.AddValueToParams(kk, ss, 'i', goftlmux.FromAuth, &rw.Ps)
			}

			fmt.Printf("\nParams + Cookies - after decrypt for (%s): %s AT %s\n", req.URL.Path, rw.Ps.DumpParamTable(), godebug.LF())

			if GetDebugFlag("DumpUnencryptedRequest") {
				fmt.Printf("At: %s, Ps=%s, req=%s\n", godebug.LF(), rw.Ps.DumpParamDB(), lib.SVarI(req)) // XyzzyDumpData
			}

			// ----------------------------------------------------------------------------------------------------------------------------------
			// ----------------------------------------------------------------------------------------------------------------------------------
			// -------------------------------- Important !!! call encrypted stuff --------------------------------------------------------------
			// ----------------------------------------------------------------------------------------------------------------------------------
			// ----------------------------------------------------------------------------------------------------------------------------------
			rw.Next = hdlr.Next
			hh, _, err := hdlr.muxEnc.Handler(req) // rv.mux.ServeHTTP(www, req)
			// ps := rw.Ps

			var mdata map[string]string
			if mdata, ok = dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email)); !ok {
				AnError(hdlr, www, req, 400, 1160, fmt.Sprintf(`Unable to find account with email '%s`, email))
				return
			}

			//fmt.Printf("\n\nnew: %s\n\n\n", godebug.SVarI(mdata))
			/*
			   new: {
			   	"Customer_id": "1",
			   	"DeviceID": "328693015",
			   	"FirstName": "",
			   	"LastName": "",
			   	"MidName": "",
			   	"PhoneNo": "",
			   	"RealName": "The white knight is talking backwards!",
			   	"UserName": "kerm4it@the-green-pc.com",
			   	"User_id": "37d911d5-6bdc-4024-5033-fd44e5e7608e",
			   	"XAttrs": "",
			   	"acct_type": "user",
			   	"auth": "y",
			   	"backup_one_time_keys": "a12a91fe4a87d683578f518968488b2e,907b259619bd8f39d21e9bb42e504b78,533d87f1fcf81649162e9e40f437ad8b,e656b4502049e23cc870dc4aa190ecb3,2e7962af579e7526edc28f3a560a58ef,29e3d10bb74e39d6b83583afb279bb13,4baf1cac81af58249493d5899064a3b3,1e2964d8681fb4005566442cf7dac604,f70c647f942741179ff5a6e63d557485,44ea631dabcffc54711ac22a5f85af4a,98da922c4d128e7f4f63bb1c773bc0fa,bd848bb4edfdf80ae814ab1ef04a7ee0,07af12743b02987d7640320902e62ee5,5ff67178ebeccae6c1db374f560873fb,072aaaa3147007bf07ece94b1d0d8fe6,3563e57266968a33d4c66451c9f64bd5,d8d19a5633695ad4a60784f803858d20,d393a09f4f2d13f5c3bc20aa5e557703,e84726a64640e46bc057d7e1ed482f2e,e91235df36fd5c9f98bcb8cf5f393cc1",
			   	"confirmed": "y",
			   	"disabled": "n",
			   	"disabled_reason": "",
			   	"email": "kerm4it@the-green-pc.com",
			   	"login_date_time": "2016-07-12T08:50:25-06:00",
			   	"login_fail_time": "",
			   	"n_failed_login": "0",
			   	"num_login_times": "11",
			   	"offline_one_time_keys": "59ea2298f16b05e3f8a2c518993b73a3,49fe3080873a1f5148c373d58638827a,f1724ce279a93652fdbe3fc1c94ef9b4,c34b991be3bbd97e0bb587e0fd099fd0,e1fa7602a89116b23a9a21fa34c91960,e36e152a7cb308a93283fb23ac383ed0,42eeb7386c2a1bc11df3c41944502696,bad4cf2e085ac5462687eb0c8735cbba,a5b03baad71dce4a21118140e4eb5acf,69bd2ea57133167c044e44d860054598,e1f21cd7e9699b9efd5e6fa505c91d0f,a88b4694cff3e4447584c2a1ec48ed84,7d60aafdbf18d30baa241387a81a6039,7ce3034d690668ac42cdd75096a0c0d0,947260c239a8a007c050488845b2b071,ea19d181a3d3a81a4870bb7ce62a5bec,b1f0bdd742637332faf2a7b33a980d3a,e9dd251cb0c4ea87fe2f103085e3233c,2711faf3a8ac9c0910564c0143c0525d,2f8bd11d0cf9d316fb0dac7a3144b755",
			   	"privs": "user",
			   	"register_date_time": "2016-07-12T08:50:25-06:00",
			   	"salt": "6f547b9377553b74784c8f626a115b83",
			   	"v": "9c4b28f7f23f712ef8f8fa55646f77d57106717012d96d0fdb95c0b8b2ea4b5e216d06ebb2f97e6c445b8318554a29121481982de8b18bbf6b4a086bf877f0f8c2df013ed0b43aaf3991c16344926da21b167c2222b55b92a7126089e5f0015267ed7068f58a636cd2e20f01a0b991720bbb3cf6dee6cd6abb2a08900364909279912f3ca9f6f27ad9cc5bc25b94385e710f47deeab5637860d6e1021e4ca2babbab7f6ef6fd36c9e7f44ef35ef2dfd51c15230f7b5e682edbb9c82f15972244b356985a02ba08042a3d5855c6f58a49754f6c2cfe76af1d2c7a5cfc3cb3ab5ad66ea0775abd3c67c525a946e0323940e23ecc53fc2401a47a5acefb99342c46",
			   	"validation_secret": "42393379"
			   }
			*/

			// newPath0 := GetPathFromURI(newPath)
			newPath0 := parsedPath.Path
			goftlmux.AddValueToParams("$user_id$", mdata["User_id"], 'i', goftlmux.FromAuth, ps)
			goftlmux.AddValueToParams("$username$", mdata["UserName"], 'i', goftlmux.FromAuth, ps)
			goftlmux.AddValueToParams("$real_name$", mdata["RealName"], 'i', goftlmux.FromAuth, ps)
			goftlmux.AddValueToParams("$email$", mdata["email"], 'i', goftlmux.FromAuth, ps)
			goftlmux.AddValueToParams("$acct_type$", mdata["acct_type"], 'i', goftlmux.FromAuth, ps)

			if _, ok := mdata["acct_type"]; !ok {
				mdata["acct_type"] = "anon-user" //
			}
			if err == nil {

				fmt.Printf("Will handle internal request, mdata[auth]=%s, mdata[acct_type]=%s, newPath0=%s %s\n", mdata["auth"], mdata["acct_type"], newPath0, godebug.LF())

				// ----------------------------------------------------------------------------------------------------------------------------------
				// goftlmux.AddValueToParams("$is_logged_in$", "y", 'i', goftlmux.FromAuth, &ps)
				// xyzzyEEE if "mdata["auth"] != "P" - pending 2fa validate what requests are legit for Pending! (Can you change password?)
				// var ApiIn2faPendingMode = map[string]bool{
				// ----------------------------------------------------------------------------------------------------------------------------------
				// ?? t_mdata["privs"] was this ??
				if mdata["auth"] == "" && mdata["acct_type"] == "anon-user" && hdlr.TwoFactorRequired == "n" && hdlr.IsValidAnonUserPath(newPath0) {
					goftlmux.AddValueToParams("$is_logged_in$", "y", 'i', goftlmux.FromAuth, ps)
					goftlmux.AddValueToParams("$is_anon_user$", "y", 'i', goftlmux.FromAuth, ps)
					fmt.Printf("**************** new *************** At: %s\n", godebug.LF())
					// really should have a "anon-user" type mux of privilates - what can an anon-user do? -- see ApiIn2faPendingMode
					hh.ServeHTTP(www, req)
				} else if mdata["auth"] == "y" || (mdata["auth"] == "P" && hdlr.TwoFactorRequired == "n") {
					goftlmux.AddValueToParams("$is_logged_in$", "y", 'i', goftlmux.FromAuth, ps)
					goftlmux.AddValueToParams("$is_enc_logged_in$", "y", 'i', goftlmux.FromAuth, ps)
					goftlmux.AddValueToParams("$is_full_login$", "y", 'i', goftlmux.FromAuth, ps)
					fmt.Printf("At: %s\n", godebug.LF())
					hh.ServeHTTP(www, req)
				} else if mdata["auth"] == "P" {
					fmt.Printf("At: %s\n", godebug.LF())
					goftlmux.AddValueToParams("$is_logged_in$", "y", 'i', goftlmux.FromAuth, ps)
					goftlmux.AddValueToParams("$is_enc_logged_in$", "y", 'i', goftlmux.FromAuth, ps)
					goftlmux.AddValueToParams("$is_full_login$", "n", 'i', goftlmux.FromAuth, ps)
					if ApiIn2faPendingMode[newPath0] {
						fmt.Printf("At: %s\n", godebug.LF())
						hh.ServeHTTP(www, req)
					} else {
						fmt.Printf("At: %s\n", godebug.LF())
						AnError(hdlr, www, req, 401, 1161, fmt.Sprintf(`Attempt to fetch resource requiring login when not logged in, path=%s, mode=%s`, newPath0, mdata["auth"]))
						return
					}
				} else {
					fmt.Printf("At: %s\n", godebug.LF())
					AnError(hdlr, www, req, 401, 1162, fmt.Sprintf(`Attempt to fetch resource requiring login when not logged in, path=%s, mode=%s`, newPath0, mdata["auth"]))
					return
				}

			} else if lib.PathsMatch(hdlr.EncReqPaths, req.URL.Path) {

				fmt.Printf("Will pass request on to next, mdata[auth]=%s, %s\n", mdata["auth"], godebug.LF())
				fmt.Printf("    mdata=%s\n", godebug.SVarI(mdata))

				// ----------------------------------------------------------------------------------------------------------------------------------
				// goftlmux.AddValueToParams("$is_logged_in$", "y", 'i', goftlmux.FromAuth, ps)
				// xyzzyEEE if "mdata["auth"] != "P" - pending 2fa and TwoFactorAuth is false etc.
				// if mdata["auth"] == "y" - then mark session and add "$is_logged_in$" = 'y' -- add as
				// if mdata["auth"] == "y" - then mark session and add "$user_id$" =  etc.
				// ----------------------------------------------------------------------------------------------------------------------------------
				// fmt.Printf("auth=%s acct_type=%s, CheckMayAccessApi()=%v,  At: %s\n", mdata["auth"], mdata["acct_type"],
				// 	CheckMayAccessApi(hdlr, rw, SandBoxPrefix, newPath0, mdata["auth"], mdata["acct_type"]), godebug.LF())
				if mdata["auth"] == "" && mdata["acct_type"] == "anon-user" && hdlr.TwoFactorRequired == "n" {
					//  hdlr.IsValidAnonUserPath(newPath0) {
					fmt.Printf("At: %s\n", godebug.LF())
					if CheckMayAccessApi(hdlr, rw, SandBoxPrefix, newPath0, mdata["auth"], mdata["acct_type"]) {
						fmt.Printf("At: %s\n", godebug.LF())
						goftlmux.AddValueToParams("$is_logged_in$", "y", 'i', goftlmux.FromAuth, ps)
						goftlmux.AddValueToParams("$is_anon_user$", "y", 'i', goftlmux.FromAuth, ps)
						fmt.Printf("Will handle external request - security check passed, %s\n", godebug.LF())
						// xyzzy - need some sort of check for htis too --- if CheckMayAccessApi(hdlr, rw, SandBoxPrefix, newPath0, mdata["auth"], mdata["acct_type"]) {
						hdlr.Next.ServeHTTP(www, req)
					} else {
						fmt.Printf("At: %s\n", godebug.LF())
						fmt.Printf("Security check - ***Failed*** - At: %s\n", godebug.LF())
						AnError(hdlr, www, req, 401, 1163, fmt.Sprintf(`Attempt to fetch resource requiring full login when not logged in, path=%s, mode=%s`, newPath0, mdata["auth"]))
						return
					}
				} else if mdata["auth"] == "y" || (mdata["auth"] == "P" && hdlr.TwoFactorRequired == "n") {
					fmt.Printf("At: %s\n", godebug.LF())
					if CheckMayAccessApi(hdlr, rw, SandBoxPrefix, newPath0, mdata["auth"], mdata["acct_type"]) {
						fmt.Printf("At: %s\n", godebug.LF())
						goftlmux.AddValueToParams("$is_logged_in$", "y", 'i', goftlmux.FromAuth, ps)
						goftlmux.AddValueToParams("$is_full_login$", "y", 'i', goftlmux.FromAuth, ps)
						fmt.Printf("Will handle external request - security check passed, %s\n", godebug.LF())
						// xyzzy - need some sort of check for htis too --- if CheckMayAccessApi(hdlr, rw, SandBoxPrefix, newPath0, mdata["auth"], mdata["acct_type"]) {
						hdlr.Next.ServeHTTP(www, req)
					} else {
						fmt.Printf("At: %s\n", godebug.LF())
						fmt.Printf("Security check - ***Failed*** - At: %s\n", godebug.LF())
						AnError(hdlr, www, req, 401, 1164, fmt.Sprintf(`Attempt to fetch resource requiring full login when not logged in, path=%s, mode=%s`, newPath0, mdata["auth"]))
						return
					}
				} else if mdata["auth"] == "P" {
					fmt.Printf("At: %s\n", godebug.LF())
					fmt.Printf("Check Privilages for passing a \"P\"/%s account on to client - SecurityData=%s At: %s\n", mdata["acct_type"], godebug.SVarI(hdlr.SecurityConfig), godebug.LF())
					fmt.Printf("Check Privilages (1) len = %d\n", len(hdlr.SecurityConfig.MayAccessApi[mdata["acct_type"]]))
					fmt.Printf("Check Privilages (2) data = %s\n", godebug.SVarI(hdlr.SecurityConfig.MayAccessApi[mdata["acct_type"]]))
					fmt.Printf("Check Privilages (3) newPath0 = [%s]\n", newPath0)
					//if len(hdlr.SecurityConfig.MayAccessApi[mdata["acct_type"]]) > 0 && (hdlr.SecurityConfig.MayAccessApi[mdata["acct_type"]][0] == "*" ||
					//	godebug.InArrayString(newPath0, hdlr.SecurityConfig.MayAccessApi[mdata["acct_type"]]) >= 0) { // if not found
					if CheckMayAccessApi(hdlr, rw, SandBoxPrefix, newPath0, mdata["auth"], mdata["acct_type"]) {
						fmt.Printf("At: %s\n", godebug.LF())
						goftlmux.AddValueToParams("$is_logged_in$", "y", 'i', goftlmux.FromAuth, ps)
						goftlmux.AddValueToParams("$is_full_login$", "n", 'i', goftlmux.FromAuth, ps)
						fmt.Printf("Security check - passed - At: %s\n", godebug.LF())
						hdlr.Next.ServeHTTP(www, req)
					} else {
						fmt.Printf("At: %s\n", godebug.LF())
						fmt.Printf("Security check - ***Failed*** - At: %s\n", godebug.LF())
						AnError(hdlr, www, req, 401, 1165, fmt.Sprintf(`Attempt to fetch resource requiring full login when not logged in, path=%s, mode=%s`, newPath0, mdata["auth"]))
						return
					}
				} else {
					fmt.Printf("At: %s\n", godebug.LF())
					AnError(hdlr, www, req, 401, 1166, fmt.Sprintf(`Attempt to fetch resource requiring login when not logged in, mode=[%s], resource=[%s]`, mdata["auth"], req.URL.Path))
					return
				}

			} else {
				fmt.Printf("At: %s\n", godebug.LF())
				fmt.Printf("\n*\n*\nforbidden?, %s\n*\n\n", godebug.LF())
				AnError(hdlr, www, req, 401, 1167, "Error(9052): Requested /api/ requires login and authentication")
				return
			}

			// xyzzyFFF -- re-fetch due to changes that could have been made in sub-calls - then not reflected as current.
			if mdata, ok = dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email)); !ok {
				AnError(hdlr, www, req, 400, 1168, fmt.Sprintf(`Unable to find account with email '%s`, email))
				return
			}

			// ----------------------------------------------------------------------------------------------------------------------------------
			// update Session and save! -- Merge rw.AddInfo -> Session and save
			for kk, vv := range rw.AddInfo {
				// switch kk {
				// case "Method", "URL", "username", "$user_id$", "$privs$", "$username$", "$auth_key$": // disalow some keys - XyzzyDisaloKeys
				if !ReservedIDs[kk] {
					mdata[kk] = vv
				}
			}
			dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)

			// ----------------------------------------------------------------------------------------------------------------------------------
			// must update rw.buffer - with new data !
			rv = ""
			plaintext = rw.GetBody()
			rw.EmptyBody()

			if dbEncr {
				fmt.Printf("\nBefore Encrypt (%s): AT %s\n", plaintext, godebug.LF())
			}

		default:
			plaintext = []byte(AnErrorRv(hdlr, www, req, 500, 33, fmt.Sprintf("invalid method passed - interal error, medhod = %s", newMethod)))
		}
	} else {
		plaintext = []byte(AnErrorRv(hdlr, www, req, 500, 34, "invalid parameter passed - interal error"))
	}

	rv, err = EncryptData(hdlr, www, req, encData.Salt, encData.Iter, encData.KeySize, tt, plaintext, key, dbCipher2, db_respHandlerCipher_1)
	if err != nil {
		return
	}

	//	{
	//
	//		// -------------------------------------------------------------------------------------------------------------------------
	//		// Inputs:
	//		//		encData.Salt					-- Supplied form session data / per-user -- check this
	//		//		encData.Iter					-- Default 1000
	//		//		encData.KeySize 				-- In Bits
	//		//		encData.KeySizeBytes			-- Derivable from KeySize/8			-- not sent in JSON data --
	//		//		tt 								-- used as adata/verified 			-- session key
	//		//  	plaintext						-- the JSON data to send back
	//		//		key								-- the encryption key
	//		//		cc.TagSizeBytes														-- not sent in JSON data --
	//		// Output:
	//		//		rv 								- the JSON encoded stirng -
	//		//
	//		// Misc Input
	//		// 		db_respHandlerCiperh_1			-- debug flag
	//		//
	//		// -------------------------------------------------------------------------------------------------------------------------
	//
	//		// 4. Encrypt into "rv"
	//		//     a. Create new return message
	//		//     b. Encrypt it
	//
	//		cc := &sjcl.SJCL_DataStruct{
	//			// InitilizationVector : "",		  //
	//			// CipherText:           "",		  //
	//			Salt:           encData.Salt,         // Communication Salt
	//			Version:        1,                    // Version of this message
	//			Iter:           encData.Iter,         // Number of iterations, normall 1000
	//			KeySize:        encData.KeySize,      // Key Size in Bits
	//			TagSize:        64,                   // In Bits
	//			Mode:           "ccm",                // Authentication Method (gcm might be metter but is not working yet)
	//			AdditionalData: []byte(tt),           // adata
	//			Cipher:         "aes",                // Encryption Method
	//			TagSizeBytes:   8,                    // Tag Size
	//			KeySizeBytes:   encData.KeySizeBytes, // Aes KeySize conv to bytes
	//			Status:         "success",            // Status of call
	//			Msg:            "",                   // Additional Error Message - empty
	//		}
	//
	//		ad := []byte(tt)
	//		var IV []byte
	//
	//		IV, _ = GenRandBytes(16) // 16 bytes of random Initialization Vector
	//		cc.InitilizationVector = IV
	//
	//		nlen := aesccm.CalculateNonceLengthFromMessageLength(len(plaintext))
	//		nonce := IV[0:nlen]
	//
	//		godebug.Printf(db_respHandlerCipher_1, "Nonce Length = %d\n", nlen)
	//
	//		cb, err := aes.NewCipher(key) // var cb cipher.Block
	//		if err != nil {
	//			AnError(hdlr, www, req, 400, 57, fmt.Sprintf("Error(0011): unable to setup AES:%s", err))
	//			return
	//		}
	//
	//		authmode, err := aesccm.NewCCM(cb, cc.TagSizeBytes, nlen) // var authmode cipher.AEAD, nlen is len(nonce)
	//		if err != nil {
	//			AnError(hdlr, www, req, 400, 58, fmt.Sprintf("Error(0012): unable to setup CCM:%s", err))
	//			return
	//		}
	//
	//		newCipterText := authmode.Seal(nil, nonce, []byte(plaintext), ad)
	//
	//		cc.CipherText = newCipterText
	//
	//		rv = lib.SVar(cc)
	//
	//		if dbCipher2 {
	//			fmt.Printf("// ////////////////////////////////////////////////////////////////////////////////////////////////////////\n")
	//			fmt.Printf("AT: %s \n", godebug.LF())
	//			fmt.Printf("Encrypted Return Value: rv -->>%s<<--\n", rv)
	//			fmt.Printf("// ////////////////////////////////////////////////////////////////////////////////////////////////////////\n")
	//		}
	//		if GetDebugFlag("DumpEncryptedReturnValue") {
	//			fmt.Printf("At: %s, Encrypted Return Value: %s\n", godebug.LF(), rv)
	//		}
	//	}

	io.WriteString(www, rv)
}

// ============================================================================================================================================
//
// Get a one-time-key for another user - requires that you be logged in as an "admin" role and have "MayGetOneTimeKey" privilage
//
func respHandlerAdminGetOneTimeKey(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1171, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	email_of_user := ps.ByNameDflt("email_of_user", "")
	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	tt := ps.ByNameDflt("t", "")
	if tt == "" || (InjectionTestMode && TestModeInject["invalid-tt-one-time-key-password"]) {
		AnError(hdlr, www, req, 400, 1172, `Invalid input data`)
		return
	}

	email, err := UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
	if err != nil {
		AnError(hdlr, www, req, 400, 1173, "Unable to find account with specified email.")
		return
	}

	admin_mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email))
	if !ok {
		AnError(hdlr, www, req, 400, 1174, "Unable to loade admin session data.  Please logout and logback in.")
		return
	}

	if admin_mdata["privs"] != "admin" {
		AnError(hdlr, www, req, 400, 1175, "Unable to identify user as an 'admin'.")
		return
	}

	// "admin": [ "MayChangeOtherPassword", "MayCreateAdminAccounts", "MayChangeOtherAttributes", "MayGetOneTimeKey" ]
	if godebug.InArrayString("MayGetOneTimeKey", hdlr.SecurityPrivilages["admin"]) < 0 {
		AnError(hdlr, www, req, 400, 1176, "Admin missing privilage 'MayGetOneTimeKey'.")
		return
	}

	mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email_of_user))
	if !ok {
		AnError(hdlr, www, req, 400, 1177, "Email is not valid for designated use - the one having the chagne made to.")
		return
	}

	DeviceID := mdata["DeviceID"] // xyzzyDeviceID

	// "Pre2Factor":               { "type":[ "string" ], "default":"p2f:" },
	OneTimeKey, err := DbGetString(hdlr, rw, SandBoxKey(hdlr.Pre2Factor, SandBoxPrefix, DeviceID))
	if err != nil {
		AnError(hdlr, www, req, 400, 1178, "Admin unable to find the one time key - account may be incorrect.")
		return
	}

	www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	www.Header().Set("Expires", "0")                                         // Proxies.

	www.WriteHeader(200)
	fmt.Fprintf(www, `{"status":"success","version":1,"OneTimeKey":%q}`, OneTimeKey)

	return

}

// ============================================================================================================================================
//
//
//	muxEnc.HandleFunc("/api/set_user_attrs", respHandlerSetUserAttrs).Method("GET", "POST")                   // ENC:
func respHandlerSetUserAttrs(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1179, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	tt := ps.ByNameDflt("t", "")
	if tt == "" {
		AnError(hdlr, www, req, 400, 1180, "Invalid input data.")
		return
	}

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")
	XAttrs := ps.ByNameDflt("XAttrs", "{}")
	if XAttrs == "{}" {
		fmt.Fprintf(www, `{"status":"success","msg":"nothing set"}`)
		return
	}

	email, err := UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
	if err != nil { // check user exists
		AnError(hdlr, www, req, 400, 1181, "Failed to find user. Invalid input email.")
		return
	}

	// parse it, remove unsettable attributes, limit to settable attributes
	newData, err := lib.JsonStringToString(XAttrs)
	if err != nil { // check user exists
		AnError(hdlr, www, req, 400, 1182, "Failed to parse attribute data. Invalid input.")
		logrus.Error(fmt.Sprintf("Failed to parse attribute data. Invalid input.  Code=939 error=%s", err))
		return
	}

	if mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email)); !ok {
		AnError(hdlr, www, req, 400, 1183, fmt.Sprintf(`Unable to find account with email '%s`, email))
		return
	} else {
		fmt.Printf("\n\nXXXX= %s\n\n", lib.SVarI(mdata))
		for _, vv := range []string{"RealName", "MidName", "LastName", "UserName", "PhoneNo"} {
			if v, ok := newData[vv]; ok {
				mdata[vv] = v
				delete(newData, vv)
			}
		}
		if v, ok := newData["XAttrs"]; ok {
			mdata["XAttrs"] = v
		}

		dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)
		hdlr.UpsertUserInfo(mdata["User_id"], mdata)
	}

	fmt.Fprintf(www, `{"status":"success"}`)
}

// ============================================================================================================================================
//
//
func respHandlerGetUserAttrs(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1184, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	tt := ps.ByNameDflt("t", "")
	if tt == "" {
		AnError(hdlr, www, req, 400, 1185, "Invalid input data.")
		return
	}

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	email, err := UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
	if err != nil { // check user exists
		AnError(hdlr, www, req, 400, 1186, "Failed to find user. Invalid input email.")
		return
	}

	// parse it, remove unsettable attributes, limit to settable attributes
	newData := make(map[string]string)

	sv := "{}"
	if mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email)); !ok {
		AnError(hdlr, www, req, 400, 1187, fmt.Sprintf(`Unable to find account with email '%s`, email))
		return
	} else {

		//	mdata["RealName"] = RealName                //
		//	mdata["FirstName"] = FirstName              //
		//	mdata["MidName"] = MidName                  //
		//	mdata["LastName"] = LastName                //
		//	mdata["UserName"] = UserName                //
		//	mdata["XAttrs"] = XAttrs                    //
		//xyzzyLoop
		if v, ok := mdata["RealName"]; ok {
			newData["RealName"] = v
		}
		if v, ok := mdata["MidName"]; ok {
			newData["MidName"] = v
		}
		if v, ok := mdata["LastName"]; ok {
			newData["LastName"] = v
		}
		if v, ok := mdata["UserName"]; ok {
			newData["UserName"] = v
		}
		if v, ok := mdata["PhoneNo"]; ok {
			newData["PhoneNo"] = v
		}
		if v, ok := mdata["XAttrs"]; ok {
			newData["XAttrs"] = v
		}
		if v, ok := mdata["email"]; ok {
			newData["email"] = v
		}
		sv = lib.SVar(newData)
	}

	fmt.Fprintf(www, `{"status":"success","attrs":%s}`, sv)
}

// ============================================================================================================================================
//
//
func respHandlerAdminSetUserAttrs(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1188, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	email_of_user := ps.ByNameDflt("email_of_user", "")
	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	tt := ps.ByNameDflt("t", "")
	if tt == "" || (InjectionTestMode && TestModeInject["invalid-tt-one-time-key-password"]) {
		AnError(hdlr, www, req, 400, 1189, `Invalid input data`)
		return
	}

	email, err := UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
	if err != nil {
		AnError(hdlr, www, req, 400, 1190, "Unable to find account with specified email.")
		return
	}

	admin_mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email))
	if !ok {
		AnError(hdlr, www, req, 400, 1191, "Unable to loade admin session data.  Please logout and logback in.")
		return
	}

	if admin_mdata["privs"] != "admin" {
		AnError(hdlr, www, req, 400, 1192, "Unable to identify user as an 'admin'.")
		return
	}

	// "admin": [ "MayChangeOtherPassword", "MayCreateAdminAccounts", "MayChangeOtherAttributes", "MayGetOneTimeKey" ]
	if godebug.InArrayString("MayGetSetAttrs", hdlr.SecurityPrivilages["admin"]) < 0 {
		AnError(hdlr, www, req, 400, 1193, "Admin missing privilage 'MayGetSetAttrs'.")
		return
	}

	mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email_of_user))
	if !ok {
		AnError(hdlr, www, req, 400, 1194, "Email is not valid for designated use - the one having the chagne made to.")
		return
	}

	XAttrs := ps.ByNameDflt("XAttrs", "{}")
	if XAttrs == "{}" {
		fmt.Fprintf(www, `{"status":"success","msg":"nothing set"}`)
		return
	}

	// parse it, remove unsettable attributes, limit to settable attributes
	newData, err := lib.JsonStringToString(XAttrs)
	if err != nil { // check user exists
		AnError(hdlr, www, req, 400, 1195, "Failed to parse attribute data. Invalid input.")
		logrus.Error(fmt.Sprintf("Failed to parse attribute data. Invalid input.  Code=939 error=%s", err))
		return
	}

	//	mdata["RealName"] = RealName                //
	//	mdata["FirstName"] = FirstName              //
	//	mdata["MidName"] = MidName                  //
	//	mdata["LastName"] = LastName                //
	//	mdata["UserName"] = UserName                //
	//	mdata["XAttrs"] = XAttrs                    //
	if v, ok := newData["RealName"]; ok {
		mdata["RealName"] = v
		delete(newData, "RealName")
	}
	if v, ok := newData["MidName"]; ok {
		mdata["MidName"] = v
		delete(newData, "MidName")
	}
	if v, ok := newData["LastName"]; ok {
		mdata["LastName"] = v
		delete(newData, "LastName")
	}
	if v, ok := newData["UserName"]; ok {
		mdata["UserName"] = v
		delete(newData, "UserName")
	}
	if v, ok := newData["PhoneNo"]; ok {
		mdata["PhoneNo"] = v
		delete(newData, "PhoneNo")
	}
	if v, ok := newData["XAttrs"]; ok {
		mdata["XAttrs"] = v
	}
	dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)
	hdlr.UpsertUserInfo(mdata["User_id"], mdata)

	fmt.Fprintf(www, `{"status":"success"}`)
}

// ============================================================================================================================================
//
//
func respHandlerAdminGetUserAttrs(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1196, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	email_of_user := ps.ByNameDflt("email_of_user", "")
	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	tt := ps.ByNameDflt("t", "")
	if tt == "" || (InjectionTestMode && TestModeInject["invalid-tt-one-time-key-password"]) {
		AnError(hdlr, www, req, 400, 1197, `Invalid input data`)
		return
	}

	email, err := UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
	if err != nil {
		AnError(hdlr, www, req, 400, 1198, "Unable to find account with specified email.")
		return
	}

	admin_mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email))
	if !ok {
		AnError(hdlr, www, req, 400, 1199, "Unable to loade admin session data.  Please logout and logback in.")
		return
	}

	if admin_mdata["privs"] != "admin" {
		AnError(hdlr, www, req, 400, 1200, "Unable to identify user as an 'admin'.")
		return
	}

	// "admin": [ "MayChangeOtherPassword", "MayCreateAdminAccounts", "MayChangeOtherAttributes", "MayGetOneTimeKey" ]
	if godebug.InArrayString("MayGetSetAttrs", hdlr.SecurityPrivilages["admin"]) < 0 {
		AnError(hdlr, www, req, 400, 1201, "Admin missing privilage 'MayGetSetAttrs'.")
		return
	}

	mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email_of_user))
	if !ok {
		AnError(hdlr, www, req, 400, 1202, "Email is not valid for designated use - the one having the chagne made to.")
		return
	}

	newData := make(map[string]string)

	sv := "{}"

	//	mdata["RealName"] = RealName                //
	//	mdata["FirstName"] = FirstName              //
	//	mdata["MidName"] = MidName                  //
	//	mdata["LastName"] = LastName                //
	//	mdata["UserName"] = UserName                //
	//	mdata["XAttrs"] = XAttrs                    //
	if v, ok := mdata["RealName"]; ok {
		newData["RealName"] = v
	}
	if v, ok := mdata["MidName"]; ok {
		newData["MidName"] = v
	}
	if v, ok := mdata["LastName"]; ok {
		newData["LastName"] = v
	}
	if v, ok := mdata["UserName"]; ok {
		newData["UserName"] = v
	}
	if v, ok := mdata["XAttrs"]; ok {
		newData["XAttrs"] = v
	}
	if v, ok := mdata["PhoneNo"]; ok {
		newData["PhoneNo"] = v
	}
	if v, ok := mdata["email"]; ok {
		newData["email"] = v
	}
	sv = lib.SVar(newData)

	fmt.Fprintf(www, `{"status":"success","attrs":%s}`, sv)
}

const base64GifPixel = "R0lGODlhAQABAIAAAP///wAAACwAAAAAAQABAAACAkQBADs="

func getUUIDAsString() (rv string) {
	id0x, _ := uuid.NewV4()
	rv = id0x.String()
	return
}

// ============================================================================================================================================
//
//	mux.HandleFunc("/api/1x1.gif", respHandler1x1Gif).Method("GET", "POST") // return the One Time key when Device ID is presented.
//
func respHandler1x1Gif(www http.ResponseWriter, req *http.Request) {

	var err error
	var fingerprint, email, SandBoxPrefix, tt, key, inm, id0, val string
	var ps goftlmux.Params
	var ok bool
	var mdata map[string]string

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		goto NeverLoggedIn
	}
	ps = rw.Ps

	godebug.Printf(db_etag, "Top of func 1x1.gif AT: %s\n", godebug.LF())
	// Fingerprint is a Username - but not unique
	// id0==Etag is a Password - and it is unique

	fingerprint = ps.ByNameDflt("fingerprint", "")
	SandBoxPrefix = ps.ByNameDflt("GOFTL_Sandbox", "")

	// PJS Temp
	id0 = getUUIDAsString()

	tt = ps.ByNameDflt("t", "")
	if tt == "" {
		godebug.Printf(db_etag, "AT: %s\n", godebug.LF())
		key = SandBoxKey("1xF:", SandBoxPrefix, fingerprint)
		email, err = DbGetString(hdlr, rw, key)
		if err != nil {
			goto NeverLoggedIn
		}
	} else {
		godebug.Printf(db_etag, "AT: %s\n", godebug.LF())
		email, err = UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
		if err != nil { // check user exists
			goto NeverLoggedIn
		}
	}

	// If have logged in previously then have email address at this point.

	if mdata, ok = dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email)); !ok {
		godebug.Printf(db_etag, "AT: %s\n", godebug.LF())
		goto NeverLoggedIn
	}

	godebug.Printf(db_etag, "AT: %s\n", godebug.LF())
	// etag := req.Header.Get("Etag")
	inm = req.Header.Get("If-None-Match")
	id0 = ""

	if inm != "" { // Ok supposedly we have seen this before
		id0 = inm
		godebug.Printf(db_etag, "AT: %s\n", godebug.LF())

		// xyzzy - at this point inm should match the mdata[user_etag] --
		if inm != mdata["user_etag"] {
			godebug.Printf(db_etag, "AT: %s\n", godebug.LF())
			goto NeverLoggedIn
		}

		val = fmt.Sprintf(`{"fingerprint":%q,"id0":%q,"email":%q}`, fingerprint, id0, email)

		// StayLoggedIn == 1x1:fingerprint -> email, + Cookie

		// save with fingerprint into Redis
		key = SandBoxKey("1x1:", SandBoxPrefix, fingerprint)
		err = DbSetExpire(hdlr, rw, key, val, 60) // Give them 1 minute to get the request issued.
		if err != nil {
			godebug.Printf(db_etag, "AT: %s\n", godebug.LF())
			goto NeverLoggedIn
		}
		godebug.Printf(db_etag, "AT: %s\n", godebug.LF())

		// save with id0 into Redis
		key = SandBoxKey("1x1:", SandBoxPrefix, id0)
		err = DbSetExpire(hdlr, rw, key, val, 60) // Give them 1 minute to get the request issued.
		if err != nil {
			godebug.Printf(db_etag, "AT: %s\n", godebug.LF())
			goto NeverLoggedIn
		}

	} else {
		id0 = getUUIDAsString()
	}

	godebug.Printf(db_etag, "AT: %s\n", godebug.LF())
	mdata["user_etag"] = id0
	dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)

NeverLoggedIn:

	godebug.Printf(db_etag, "AT: %s\n", godebug.LF())
	www.Header().Set("Etag", id0)

	www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	www.Header().Set("Expires", "0")                                         // Proxies.

	www.Header().Set("Content-Type", "image/gif")
	output, _ := base64.StdEncoding.DecodeString(base64GifPixel)
	io.WriteString(www, string(output))
}

const db_respHandlerCipher_1 = true
const dbRespHandlerSrpRegister = true // register and confirm of register
const dbRespHandlerSRPGetNg = true
const db_SetupRoles = true
const init_db1 = false
const db_etag = true

// ============================================================================================================================================
//
// mux.HandleFunc("/api/test1", respHandlerTest1).Method("GET")                                          // Test 1 - of Etag
//
// Use IamI cookie as 2fa/One-time-key for loggin to UUID/f account
//
var seq = 100

//func respHandlerTest1(www http.ResponseWriter, req *http.Request) {
//
//	godebug.Printf(db_etag, "\n\n ------------------ ====================== ----------------- start AT: %s\n", godebug.LF())
//	var key, val string
//	var ps goftlmux.Params
//	var ok bool
//
//	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
//	if !ok {
//		godebug.Printf(db_etag, "AT: %s\n", godebug.LF())
//	}
//	ps = rw.Ps
//
//	fingerprint := ps.ByNameDflt("fingerprint", "")
//	_ = fingerprint
//	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")
//	LoginAuthToken := ps.ByNameDflt("LoginAuthToken", "")
//
//	inm := req.Header.Get("If-None-Match")
//	//etag := req.Header.Get("Etag")
//	//_ = etag
//	id0 := ""
//	godebug.Printf(db_etag, "If-None-Match = %s AT: %s\n", inm, godebug.LF())
//
//	if inm != "" { // Ok supposedly we have seen this before
//		godebug.Printf(db_etag, "AT: %s\n", godebug.LF())
//
//		BrowserMark, err := DbGetString(hdlr, rw, SandBoxKey("1x1:", SandBoxPrefix, inm))
//		if err != nil || BrowserMark == "" {
//			godebug.Printf(db_etag, "err = %s BrowserMark = %s AT: %s\n", err, BrowserMark, godebug.LF())
//			// We do not have a record of this.
//			id0 = getUUIDAsString()
//			val = fmt.Sprintf(`var _v_ = {"status":"success","id":%q};`, id0)
//			etag := HashStrings.Sha256(val)
//			www.Header().Set("ETag", etag)
//			www.Header().Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0") // HTTP 1.1.
//			www.Header().Set("Content-Type", "application/javascript; charset=utf-8")       // For JS
//			expire := time.Now().AddDate(0, 0, hdlr.CookieExpireInXDays)                    // Years, Months, Days==2 // Xyzzy501 - should be a config - on how long to keep cookie
//			cookie := http.Cookie{Name: "IamI", Value: "new-login-redis-expire", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400, Secure: false, HttpOnly: true}
//			http.SetCookie(www, &cookie)
//			www.WriteHeader(200)
//			fmt.Fprintf(www, val)
//			// xyzzy - save data to Redis
//			key = SandBoxKey("1x1:", SandBoxPrefix, etag)
//			val = fmt.Sprintf(`{"cookie":%q}`, LoginAuthToken)
//			_ = DbSetExpire(hdlr, rw, key, val, 300) // Test - give them 5 min
//			return
//		}
//
//		// godebug.Printf(db_etag, "AT: %s\n", godebug.LF())
//		godebug.Printf(db_etag, "err = %s BrowserMark = %s AT: %s\n", err, BrowserMark, godebug.LF())
//
//		// Else - we have seen this before
//		id0 = inm
//		www.Header().Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0") // HTTP 1.1.
//		expire := time.Now().AddDate(0, 0, hdlr.CookieExpireInXDays)                    // Years, Months, Days==2 // Xyzzy501 - should be a config - on how long to keep cookie
//		cookie := http.Cookie{Name: "IamI", Value: id0 + "::" + fmt.Sprintf("%d", seq), Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400, Secure: false, HttpOnly: true}
//		http.SetCookie(www, &cookie)
//		seq++
//		// etag := req.Header.Get("ETag")
//		www.Header().Set("ETag", inm)
//		key = SandBoxKey("1x1:", SandBoxPrefix, inm)
//		val = fmt.Sprintf(`{"cookie":%q}`, LoginAuthToken)
//		_ = DbSetExpire(hdlr, rw, key, val, 300) // Test - give them 5 min
//		www.WriteHeader(304)
//		return
//
//	}
//
//	godebug.Printf(db_etag, "AT: %s\n", godebug.LF())
//	// We do not have a record of this.
//	id0 = getUUIDAsString()
//	val = fmt.Sprintf(`var _v_ = {"status":"success","id":%q};`, id0)
//	etag := HashStrings.Sha256(val)
//	www.Header().Set("ETag", etag)
//	www.Header().Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0") // HTTP 1.1.
//	www.Header().Set("Content-Type", "application/javascript; charset=utf-8")       // For JS
//	expire := time.Now().AddDate(0, 0, hdlr.CookieExpireInXDays)                    // Years, Months, Days==2 // Xyzzy501 - should be a config - on how long to keep cookie
//	cookie := http.Cookie{Name: "IamI", Value: "new-login-no-inm", Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400, Secure: false, HttpOnly: true}
//	http.SetCookie(www, &cookie)
//	www.WriteHeader(200)
//	fmt.Fprintf(www, val)
//	key = SandBoxKey("1x1:", SandBoxPrefix, etag)
//	val = fmt.Sprintf(`{"cookie":%q}`, LoginAuthToken)
//	_ = DbSetExpire(hdlr, rw, key, val, 300) // Test - give them 5 min
//	return
//
//}

// ============================================================================================================================================
// update Etag stuff with email // mark Etag as = if aMap["hasLoggedIn"] != "" {
func SetEmailFromIamI(hdlr *AesSrpType, rw *goftlmux.MidBuffer, SandBoxPrefix string, IamI, email string) (err error) {
	var s, Etag string

	Etag, err = DbGetString(hdlr, rw, SandBoxKey("1x1:", SandBoxPrefix, IamI))
	if err != nil {
		return
	}

	s, err = DbGetString(hdlr, rw, SandBoxKey("1x1:", SandBoxPrefix, Etag))
	if err != nil {
		return
	}

	aMap, err := lib.JsonStringToString(s)
	if err != nil {
		return
	}
	aMap["email"] = email         // Save the email address
	t := time.Now()               // Save when this got marked, could just be "Y" instead - time is for human
	tss := t.Format(time.RFC3339) //
	aMap["hasLoggedIn"] = tss

	err = DbSetString(hdlr, rw, SandBoxKey("1x1:", SandBoxPrefix, Etag), lib.SVar(aMap))

	return
}

// ============================================================================================================================================
// update Etag stuff with email // mark Etag as = if aMap["hasLoggedIn"] != "" {
func UpdateNotStayLoggedIn(hdlr *AesSrpType, rw *goftlmux.MidBuffer, SandBoxPrefix string, IamI, email string) (err error) {
	var s, Etag string

	Etag, err = DbGetString(hdlr, rw, SandBoxKey("1x1:", SandBoxPrefix, IamI))
	if err != nil {
		return
	}

	s, err = DbGetString(hdlr, rw, SandBoxKey("1x1:", SandBoxPrefix, Etag))
	if err != nil {
		return
	}

	aMap, err := lib.JsonStringToString(s)
	if err != nil {
		return
	}
	aMap["hasLoggedIn"] = ""

	err = DbSetString(hdlr, rw, SandBoxKey("1x1:", SandBoxPrefix, Etag), lib.SVar(aMap))

	return
}

//	etag, err := GetEtagFromIamI ( IamI )
// ============================================================================================================================================
func GetEmailFromIamI(hdlr *AesSrpType, rw *goftlmux.MidBuffer, SandBoxPrefix string, IamI string) (email, Etag string, err error) {
	var s string

	Etag, err = DbGetString(hdlr, rw, SandBoxKey("1x1:", SandBoxPrefix, IamI))
	if err != nil {
		return
	}

	s, err = DbGetString(hdlr, rw, SandBoxKey("1x1:", SandBoxPrefix, Etag))
	if err != nil {
		return
	}

	aMap, err := lib.JsonStringToString(s)
	if err != nil {
		return
	}

	email = aMap["email"]

	return
}

// ============================================================================================================================================
//mdata["LoginAuthCookie"] = SaveAsList(mdata["LoginAuthCookie"], cookieValue)
//cookie2 := http.Cookie{Name: "LoginHashCookie", Value: cookieHash, Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: LoginHashCookieLife, Secure: secure, HttpOnly: true}
func ValidateCookies(LoginAuthCookie, LoginHashCookie, validation_secret, email string) bool {
	tmp := HashStrings.Sha256(LoginAuthCookie + ":" + validation_secret)
	fmt.Printf("validation of cookie: tmp=[%s] lac=[%s] validation_secret=[%s]\n", tmp, LoginAuthCookie, validation_secret)
	if tmp == LoginHashCookie {
		return true
	}
	return false
}

// ============================================================================================================================================
//
// mux.HandleFunc("/api/markPage", respHandlerMarkPage).Method("GET")                                          // Setup page marker
//
// respHandlerMarkPage uses one method of maping a users browser to a unique ID.  This is based on
// the browser caching the file.   Each browser is given a unique flle to cache with a unique ID in
// it.  This ID can be retreived via a call to GetPageMarkerId() in the Javascript and must be
// returned to the server.
//
// As with other cached files the Sha256 of the file contents is used as the Etag/If-None-Match.
//
// This approach makes browser "incogneto" mode work properly.  If in "incogneto" mode you can not
// permanently cache the file and stay logged in.
//
// PageMarkJS is declared jsut following the function.
//
func respHandlerMarkPage(www http.ResponseWriter, req *http.Request) {

	godebug.Printf(db_etag, "\n\n ------------------ ====================== ----------------- start AT: %s\n", godebug.LF())
	var ps goftlmux.Params
	var ok bool

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		godebug.Printf(db_etag, "AT: %s\n", godebug.LF())
	}
	ps = rw.Ps

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")

	inm := req.Header.Get("If-None-Match")
	id0 := getUUIDAsString()
	secure := false
	if req.TLS != nil {
		secure = true
	}
	godebug.Printf(db_etag, "If-None-Match = %s id0 = %s secure = %v AT: %s\n", inm, id0, secure, godebug.LF())

	fx := func(cookieValue string) {
		jsCode := fmt.Sprintf(MarkPageJS, id0)
		etag := HashStrings.Sha256(jsCode)
		www.Header().Set("ETag", etag)
		www.Header().Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0") // HTTP 1.1.
		www.Header().Set("Content-Type", "application/javascript; charset=utf-8")       // For JS
		expire := time.Now().AddDate(0, 0, 30)                                          // Years, Months, Days==30 // xyzzy501 - should be a config - on how long to keep cookie
		cookie := http.Cookie{Name: "IamI", Value: id0, Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 30 * 86400, Secure: secure, HttpOnly: true}
		http.SetCookie(www, &cookie)
		www.WriteHeader(200)
		fmt.Fprintf(www, jsCode)
		// - save data to Redis ----------------------------------------------

		redis_etag_key := SandBoxKey("1x1:", SandBoxPrefix, etag)
		redis_etag_data := fmt.Sprintf(`{"etag":%q,"IamI":%q}`, etag, id0)
		_ = DbSetString(hdlr, rw, redis_etag_key, redis_etag_data)

		IamI_key := SandBoxKey("1x1:", SandBoxPrefix, id0)   // From Cookie to Etag - used in both setup and resume
		_ = DbSetExpire(hdlr, rw, IamI_key, etag, 370*86400) // Approx 370 days to do your first login
	}

	if inm != "" { // Ok supposedly we have seen this before
		godebug.Printf(db_etag, "AT: %s\n", godebug.LF())

		redis_etag_key := SandBoxKey("1x1:", SandBoxPrefix, inm)
		BrowserMark, err := DbGetString(hdlr, rw, redis_etag_key) // Pull, Parse, Update, Re-Write
		if err != nil || BrowserMark == "" {                      // We do not have a record of this.
			godebug.Printf(db_etag, "err = %s BrowserMark = %s AT: %s\n", err, BrowserMark, godebug.LF())
			fx("new-login-redis-expire")
			return
		}
		aMap, err := lib.JsonStringToString(BrowserMark)
		if err != nil {
			fx("new-login-redis-bad-data")
			return
		}

		IamI_old_key := SandBoxKey("1x1:", SandBoxPrefix, aMap["IamI"]) // At this point have the old IamI in aMap - delete it.
		DbDel(hdlr, rw, IamI_old_key)

		// godebug.Printf(db_etag, "AT: %s\n", godebug.LF())
		godebug.Printf(db_etag, "err = %s BrowserMark = %s AT: %s\n", err, BrowserMark, godebug.LF())

		// Else - we have seen this before
		www.Header().Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0") // HTTP 1.1.
		expire := time.Now().Add(90 * time.Second)
		cookie := http.Cookie{Name: "IamI", Value: id0, Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 90, Secure: secure, HttpOnly: true}
		http.SetCookie(www, &cookie)
		www.Header().Set("ETag", inm)

		// - save data to Redis ----------------------------------------------
		aMap["IamI"] = id0
		redis_etag_data := lib.SVar(aMap) // val = fmt.Sprintf(`{"etag":%q,"IamI":%q}`, inm, id0) //

		_ = DbSetExpire(hdlr, rw, redis_etag_key, redis_etag_data, 370*86400) // Approx 370 days -- A little over a year -- rewrite with new IamI key

		IamI_key := SandBoxKey("1x1:", SandBoxPrefix, id0) //
		if aMap["hasLoggedIn"] != "" {                     // If the user has logged in
			_ = DbSetExpire(hdlr, rw, IamI_key, inm, 90) // 90 seconds to use
		} else {
			_ = DbSetExpire(hdlr, rw, IamI_key, inm, 370*86400) // 370 days - never logged in
		}
		www.WriteHeader(304)
		return

	}

	godebug.Printf(db_etag, "AT: %s\n", godebug.LF())
	fx("new-login-never-seen-before")
	return

}

var MarkPageJS = `
var _v_ = %q;
function GetPageMarkerId(){ return _v_; }
`

/*
// An alternative implementation that is not dependent on ETag/If-None-Match+Cookies
var MarkPageJS = `
function GetPageMarkerId(){ return %q; }
`
*/

// ============================================================================================================================================
//
// muxEnc.HandleFunc("/api/setupStayLoggedIn", respHandlerSetupStayLoggedIn).Method("GET", "POST")           // ENC:	allow for stayLoggedIn user  -- new --
//
// Input:
//
// 	stayLoggedIn						t/f from webpage
//	t				Login required
//	IamI			Cookie				from pageMark
//	LoginAuthCookie	Cookie				from login/validate
//	salt			(optional)			if we need to register the LoginAuthCookie as an anonomous user.
//	v				(optional)			buildt from the 'fingerprint' as password
//
//		If need to register
//			salt, v, LoginAuthCookie for Username + Password (Salt/v) from fingerprint
//
//		(( IamI/C, salt/p, v/p, username/p, t(from login), email(from login) ))
//		1. Have IamI + Auth Cookies from Login + Key + t
//		2. send back "t-2", "key-2" for future with "enc-2" passwrod
//		3. Save into Redis 1x1:IamI -> ETag -> "email"+"enc-2"+"t-2"+"key-2"
//		*4. Create AuthCookie/hash(F)=salt/v
//		?5. Delete 1x1:IamI from Redis
//
func respHandlerSetupStayLoggedIn(www http.ResponseWriter, req *http.Request) {

	/* $func$ 40 */

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1203, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")
	StayLoggedInFlag := ps.ByNameDflt("stayLoggedIn", "") // passed
	StayLoggedInFlagBool, _ := lib.ParseBool(StayLoggedInFlag)
	salt := ps.ByNameDflt("salt", "") // optional may move down
	v := ps.ByNameDflt("v", "")       // optional may move down
	if v == "" {
		v = ps.ByNameDflt("verifier", "") // this is 'v'	-- alternate name 'verifier'
	}
	username := ps.ByNameDflt("LoginAuthCookie", "") // optional may move down
	tt := ps.ByNameDflt("t", "")
	if tt == "" {
		AnError(hdlr, www, req, 400, 1204, "Invalid input data.  Missing or invalid 't' parameter.")
		return
	}

	email, err := UserGetEmail(hdlr, rw, tt, SandBoxPrefix)
	if err != nil { // check user exists
		AnError(hdlr, www, req, 400, 1205, "Failed to find user. Invalid input email.")
		return
	}

	var mdata map[string]string
	var t_mdata map[string]string
	if mdata, ok = dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email)); !ok {
		AnError(hdlr, www, req, 400, 1206, fmt.Sprintf(`Unable to find account with email '%s`, email))
		return
	}
	_ = mdata

	IamI := ps.ByNameDflt("IamI", "")
	if IamI == "" {
		AnError(hdlr, www, req, 400, 1207, fmt.Sprintf(`Unable to find account with email '%s`, email))
		return
	}

	emailFromIamI, etag, err := GetEmailFromIamI(hdlr, rw, SandBoxPrefix, IamI)
	if IamI == "" {
		AnError(hdlr, www, req, 400, 1208, fmt.Sprintf(`Unable to find account with email '%s`, email))
		return
	}

	// xyzzy what if emailFromIamI != email?? -- is that an error? -- Do we ignore?

	fmt.Printf("AT: %s\n", godebug.LF())
	fmt.Printf("Data: tt=%s\n", tt)
	fmt.Printf("Data: email=%s\n", email)
	fmt.Printf("Data: IamI=%s\n", IamI)
	fmt.Printf("Data: emailFromIamI=%s\n", emailFromIamI)
	fmt.Printf("Data: StayLoggedInFlag=%s type=%T\n", StayLoggedInFlag, StayLoggedInFlag)
	fmt.Printf("Data: StayLoggedInFlagBool=%v\n", StayLoggedInFlagBool)
	fmt.Printf("Data: username=%v\n", username)
	fmt.Printf("Data: salt=%v\n", salt)
	fmt.Printf("Data: v=%v\n", v)

	// need to map from user(tt) -> Etag - to update it - so if StayLoggedInFlag != "y", then unmakr the users IamI
	// must be the "destination" user's email record that is updated?
	if !StayLoggedInFlagBool && emailFromIamI == email {
		fmt.Printf("AT: %s\n", godebug.LF())
		_ = UpdateNotStayLoggedIn(hdlr, rw, SandBoxPrefix, IamI, email)
		fmt.Fprintf(www, `{"status":"success","key2":%q,"enc2":%q}`, "x", "x")
		return
	}

	fmt.Printf("AT: %s\n", godebug.LF())
	// IamI -> Etag -> email stuff
	// update Etag stuff with email // mark Etag as = if aMap["hasLoggedIn"] != "" {
	err = SetEmailFromIamI(hdlr, rw, SandBoxPrefix, IamI, email)
	if err != nil {
		AnError(hdlr, www, req, 400, 1209, fmt.Sprintf(`Unable to save temporary login information for stay logged in, error='%s`, err))
		return
	}

	// xyzzy - if "username"/LoginAuthCookie is not a user then if - no salt/v - return error -- Else - update salt/v? or error if dup create
	haveUser := false
	t_mdata, haveUser = dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, username))

	if !haveUser && salt == "" && v == "" { // If don't have user ask browser to create salt/v and call again
		AnError(hdlr, www, req, 400, 1210, fmt.Sprintf(`Persistent login will require salt/v and a cookie`))
		return
	}

	if !haveUser {
		t_mdata = make(map[string]string)
	}

	// generate new "key-2"
	rn, _ := GenRandNumber(8)
	lmSalt := fmt.Sprintf("%x", rn) // string(hn)
	key2 := HashStrings.Sha256(lmSalt + ":" + IamI)
	t_mdata["key2"] = key2 // ///////////////////////////////////////////////////////////////////// boom
	rn, _ = GenRandNumber(8)
	lmSalt = fmt.Sprintf("%x", rn) // string(hn)
	enc2 := HashStrings.Sha256(lmSalt + ":" + IamI)
	t_mdata["enc2"] = enc2

	// 1. See key2 / enc2 generated on setupStayLoggedIn 				-- and saved at //xyzzyKey2Generated

	fmt.Printf("Data: key2=%v, %s\n", key2, godebug.LF())
	fmt.Printf("Data: enc2=%v\n", enc2)

	t_theKey := SandBoxKey("srp:U:", SandBoxPrefix, username) //
	fmt.Printf("--- register the new user now --- haveUser=%v AT: %s\n", haveUser, godebug.LF())
	if !haveUser {

		if !validSrpSalt(salt) {
			fmt.Printf("AT: %s\n", godebug.LF())
			AnError(hdlr, www, req, 400, 1211, "Invalid salt.")
			return
		}
		if !validSrpV(v) {
			fmt.Printf("AT: %s\n", godebug.LF())
			AnError(hdlr, www, req, 400, 1212, "Invalid 'v' verifier value.")
			return
		}

		// Create the "anon-user" for next time ---------------------------------------------------------------------------------------------
		// save new salt,v for this fingerprint
		SetSaltV(hdlr, www, req, t_mdata, salt, v) // Encrypt Salt,V if encryption is enabled
		t_mdata["acct_type"] = "anon-user"         //
		t_mdata["confirmed"] = "y"                 // mark as "email" is not confirmed by user yet
		t_mdata["disabled"] = "n"                  //
		t_mdata["disabled_reason"] = ""            //
		t_mdata["n_failed_login"] = "0"            //
		t := time.Now()                            //
		tss := t.Format(time.RFC3339)              //
		t_mdata["register_date_time"] = tss        //
		t_mdata["login_date_time"] = tss           //
		t_mdata["login_fail_time"] = ""            // if n_failed_login > threshold then this is the time when to resume login trials
		t_mdata["privs"] = "anon-user"             //
		t_mdata["num_login_times"] = "0"           //
		t_mdata["User_id"] = mdata["User_id"]      //	Same $user_id$ for both
		t_mdata["owner_email"] = email             // who is the person that owns this anon-user
		t_mdata["IamI"] = IamI                     //
		t_mdata["markPageEtag"] = etag             //
		//dataStore.RSetValue(hdlr, rw, t_theKey, t_mdata) //
		//DbExpire(hdlr, rw, t_theKey, 30*86400)           // expire this user
		// --------------------------------------------------------------------------------------------------------------------------------------
		fmt.Printf("Data: User_id = %s = %s, %s\n", t_mdata["User_id"], mdata["User_id"], godebug.LF())
	}

	fmt.Printf("The user should now be store under [%s] with key2/enc2 of %s %s, %s\n", t_theKey, t_mdata["key2"], t_mdata["enc2"], godebug.LF())

	dataStore.RSetValue(hdlr, rw, t_theKey, t_mdata) // may create user - must set enc2, key2
	DbExpire(hdlr, rw, t_theKey, 30*86400)           // expire this user

	// 1. See key2 / enc2 generated on setupStayLoggedIn 				-- and saved at //xyzzyKey2Generated

	fmt.Printf("AT: %s\n", godebug.LF())
	// check that we are on a real loggin, "$is_logged_in$":              true, "$is_full_login$":             true,
	is_logged_in := ps.ByNameDflt("$is_logged_in$", "")
	is_enc_logged_in := ps.ByNameDflt("$is_enc_logged_in$", "")
	is_full_login := ps.ByNameDflt("$is_full_login$", "")
	fmt.Printf("Data: is_logged_in=%v, is_full_login=%v is_enc_logged_in=%v\n", is_logged_in, is_full_login, is_enc_logged_in)
	if is_logged_in == "y" && (is_full_login == "y" || (hdlr.TwoFactorRequired == "n" && is_enc_logged_in == "y")) {
		s := fmt.Sprintf(`{"status":"success","key2":%q,"enc2":%q}`, key2, enc2) // xyzzy - need to send back: key2, enc2
		fmt.Printf("Successful Return of %s from /api/setupStayLoggedIn AT: %s\n", s, godebug.LF())
		fmt.Fprintf(www, `{"status":"success","key2":%q,"enc2":%q}`, key2, enc2) // xyzzy - need to send back: key2, enc2
		// 1. See key2 / enc2 generated on setupStayLoggedIn 				-- and saved at //xyzzyKey2Generated
		return
	}

	fmt.Printf("Failure at bottom AT: %s\n", godebug.LF())
	AnError(hdlr, www, req, 400, 1213, fmt.Sprintf(`Unable to support staying logged in.`))
}

// ============================================================================================================================================
//
//	muxEnc.HandleFunc("/api/resumeLogin", respHandlerResumeLogin).Method("GET", "POST")                       // ENC: for resumption of a previously logged in session -- new --
//
// Input:
//		IamI - cookie				- IamI -> Etag -> Email - if login has occured, tie back to an account
//		LoginAuthCookie - cookie	- username
//		LoginHashCookie - cookie	- cookie signature
//
// To use this first you login with LoginAuthCookie/fingerprint as your login.  Thnen you will get back data.
//
// xyzzy - defect - if visit but do not login, then must find key and perform -- aMap["hasLoggedIn"] = ""
// 		suggests that this just get passed the "StayLoggedIn" flag and it make the choices - call every time
//
func respHandlerResumeLogin(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 1214, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	stayLoggedIn := ps.ByNameDflt("stayLoggedIn", "")

	if stayLoggedIn == "false" {
		fmt.Fprintf(www, `{"status":"note-ok"}`)
		return
	}

	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")
	IamI := ps.ByNameDflt("IamI", "")
	LoginAuthCookie := ps.ByNameDflt("LoginAuthCookie", "")
	LoginHashCookie := ps.ByNameDflt("LoginHashCookie", "")

	tt := ps.ByNameDflt("t", "")
	if tt == "" {
		AnError(hdlr, www, req, 400, 1215, "Invalid input data.  Missing or invalid 't' parameter.")
		return
	}

	// go from IamI to email
	email, _, err := GetEmailFromIamI(hdlr, rw, SandBoxPrefix, IamI)
	if err != nil {
		AnError(hdlr, www, req, 400, 1216, "Failed to find user. Invalid input email.")
		return
	}

	fmt.Printf("Email from IamI = [%s]\n", email)

	// lookup mdata for email
	var mdata map[string]string           // data for supposed user (from super-cookie)
	var tmp_login_mdata map[string]string // data for the anon-user
	if mdata, ok = dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email)); !ok {
		AnError(hdlr, www, req, 400, 1217, fmt.Sprintf(`Unable to find account with email '%s`, email))
		return
	}

	t_theKey := SandBoxKey("srp:U:", SandBoxPrefix, LoginAuthCookie)
	if tmp_login_mdata, ok = dataStore.RGetValue(hdlr, rw, t_theKey); !ok {
		AnError(hdlr, www, req, 400, 1218, fmt.Sprintf(`Unable to find account with email '%s`, LoginAuthCookie))
		return
	}

	// -------------------------------------------------------------------------------------------------------
	// need to check login checks that users account is enabled, valid etc.
	// -------------------------------------------------------------------------------------------------------

	// If the user has not been confirmed then reject the user - not confirmed means no email confirm.
	if mdata["confirmed"] == "n" {
		AnError(hdlr, www, req, 401, 1219, `The account has not been confirmed.  Please confirm or register again and get a new confirmation email.`)
		return
	}

	// If the user is disabled then done - can not login - see admin
	if mdata["disabled"] == "y" {
		AnError(hdlr, www, req, 401, 1220, "The account has been disabled.  Please contact customer support (call them).")
		return
	}

	// If more than allowed number of failed login attempts or this is not a valid number in the database then - oops no login
	nf, err := strconv.ParseInt(mdata["n_failed_login"], 10, 64)

	if err != nil {
		mdata["n_failed_login"] = "0" //
		mdata["login_fail_time"] = "" // if n_failed_login > threshold then this is the time when to resume login trials
		dataStore.RSetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)
	} else if int(nf) > hdlr.FailedLoginThreshold {
		t := time.Now()
		if mdata["login_fail_time"] != "" {
			x, err := time.Parse(time.RFC3339, mdata["login_fail_time"])
			if err == nil && t.Before(x) {
				AnError(hdlr, www, req, 401, 1221, "The account has been disabled.  Please contact customer support (call them).")
				return
			}
		}
	}

	validation_secret := mdata["validation_secret"]
	fmt.Printf("validation_secret = [%s] AT: %s\n", validation_secret, godebug.LF())

	// validate LoginHashCookie/LoginAuthCookie
	if !ValidateCookies(LoginAuthCookie, LoginHashCookie, validation_secret, email) {
		fmt.Printf("LoginAuthCookie=[%s] LoginHashCookie=[%s] validation_secret=[%s] email=[%s] AT: %s\n", LoginAuthCookie, LoginHashCookie, validation_secret, email, godebug.LF())
		AnError(hdlr, www, req, 400, 1222, "Failed to find user. Invalid input email.")
		return
	}

	// ----------------------------------------------------------------------------------------------------------------------------------------------------------------------
	// pull session key
	// _ /*Password*/, _ /*tSalt*/, tKey, _ /*tIter*/, _ /*tKeySize*/, tEmail, _ /*Session*/ := GetKeyData(hdlr, rw, tt, SandBoxPrefix)
	// fmt.Printf("tEmail from old GetKeyData=%s tKey=%x\n", tEmail, tKey)
	tKey, tEmail, raw_session := GetKeyDataRaw(hdlr, rw, tt, SandBoxPrefix)
	_ = raw_session
	fmt.Printf("tEmail from new GetKeyData=%s tKey=%x\n", tEmail, tKey)

	x := fmt.Sprintf("%x", tKey)
	fmt.Printf("\nK <citical> <critical> old session key [%s] %T len=%d for tEmail=[%s] email=[%s]\n\n", x, x, len(x), tEmail, email)
	// fmt.Printf("\nK <citical> <critical> old session key [%s] %T len=%d for [%s]\n\n", tKey, tKey, len(tKey), tEmail)

	// xyzzy - validate that tEmail matches with email -- tEmail at this point matches with the UUID account not the uses email

	fmt.Printf("LoginAuthCookie=[%s] tEmail=[%s] AT: %s\n", LoginAuthCookie, tEmail, godebug.LF())
	if LoginAuthCookie != tEmail {
		fmt.Printf("Email is not a UUID at this point - that's bad\n")
		AnError(hdlr, www, req, 400, 1223, "Failed to find user. Invalid input email.")
		return
	}

	// ----------------------------------------------------------------------------------------------------------------------------------------------------------------------
	// save session key for next time.  // xyzzyNextCall set and save KEY to session in d.b.  - that way will get used on NEXT call.
	//if false {
	//	fmt.Printf("Switch to new encryption key\n")
	//	_ = UpdateSessionEncryptionKey(hdlr, rw, tt, SandBoxPrefix, tmp_login_mdata["key2"], raw_session)
	//}

	// Xyzzy - delete 1x1:IamI -- will timeout - good for testing

	// generate key3/enc3 and send // shouldn't we be sending key3/enc3 at same time???
	rn, _ := GenRandNumber(8)
	lmSalt := fmt.Sprintf("%x", rn) // string(hn)
	key3 := HashStrings.Sha256(lmSalt + ":" + IamI)
	rn, _ = GenRandNumber(8)
	lmSalt = fmt.Sprintf("%x", rn) // string(hn)
	enc3 := HashStrings.Sha256(lmSalt + ":" + IamI)

	fmt.Printf("key2=[%s] enc2=[%s] AT: %s\n", tmp_login_mdata["key2"], tmp_login_mdata["enc2"], godebug.LF())

	dataStore.RUpdValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, email), mdata)

	rv := make(map[string]string)
	rv["status"] = "success"
	rv["key2"] = tmp_login_mdata["key2"]
	rv["enc2"] = tmp_login_mdata["enc2"]
	rv["key3"] = key3
	rv["enc3"] = enc3
	rv["email"] = email
	rv["UserName"] = mdata["UserName"]
	rv["RealName"] = mdata["RealName"]
	rv["FirstName"] = mdata["FirstName"]
	rv["MidName"] = mdata["MidName"]
	rv["LastName"] = mdata["LastName"]
	rv["PhoneNo"] = mdata["PhoneNo"]

	// xyzzy missing attrs?  hh.DbUserCols = []string{"RealName", "Customer_id", "User_id", "UserName", "FirstName", "MidName", "LastName ", "User_id", "Customer_id", "XAttrs", "PhoneNo"}

	fmt.Fprintf(www, lib.SVar(rv))

	//	fmt.Fprintf(www, `{"status":"success", "key2":%q, "enc2":%q, "key3":%q, "enc3":%q }`, tmp_login_mdata["key2"], tmp_login_mdata["enc2"], key3, enc3)

	fmt.Printf("key3=[%s] enc3=[%s] AT: %s\n", key3, enc3, godebug.LF())
	tmp_login_mdata["key2"] = key3                           //
	tmp_login_mdata["enc2"] = enc3                           //
	dataStore.RSetValue(hdlr, rw, t_theKey, tmp_login_mdata) // may create user - must set enc2, key2
	DbExpire(hdlr, rw, t_theKey, 30*86400)                   // expire this user

	return

	/*

		Key is base 64 encoded - what?

		What is Pw?

			192.168.0.133:6379> get "ses:4dfb47cb-d64a-46d4-5656-61e585438336"
				{	"$auth$":"y",
					"$auth_key$":"d3f592dd-3363-46f5-7bbd-94f31458d7b2",
					"$privs$":"user",
					"$username$":"pschlump@gmail.com",
					"Email":"pschlump@gmail.com",
					"Iter":1000,
					"Key":"2nQ7WuPTWgK6iZwGHo53OSrHhmRqSeyN/TH69qrLfvw=",
					"Keysize":256,
					"Pw":"f8f24a520e9657aa883a6e2e13394c26a5ce678de1519f4105ec1cef3a92bcf6",
					"Salt":"YDkTYW9MijI=",
					"login_date_time":"2016-04-11T19:57:53-06:00",
					"login_fail_time":"",
					"n_failed_login":"0"
				}
			192.168.0.133:6379> ttl "ses:4dfb47cb-d64a-46d4-5656-61e585438336"
			(integer) 86244

		// From func GetKeyData(hdlr *AesSrpType, rw *goftlmux.MidBuffer, tt, SandBoxPrefix string) (pw string, salt string, key string, iter int, keysize int, email string, ss map[string]string) {
			skey, _ := base64.StdEncoding.DecodeString(rv.Key)
		// From func SaveKeyData(hdlr *AesSrpType, rw *goftlmux.MidBuffer, tt, SandBoxPrefix string, pw, salt, key string, iter int, keysize int) (err error) {
			skey := base64.StdEncoding.EncodeToString([]byte(key))

		----------------------
		key [da743b5ae3d35a02ba899c061e8e77392ac786646a49ec8dfd31faf6aacb7efc] len=32
		----------------------

		https://golang.org/pkg/encoding/hex/

		Fields to set when resumeSession:
			"$privs$":"user",
			"$username$":"pschlump@gmail.com",
			"Email":"pschlump@gmail.com",
			"Key":"2nQ7WuPTWgK6iZwGHo53OSrHhmRqSeyN/TH69qrLfvw=",
	*/

	// start encryption with new key
	// err = UpdateSessionEncryptionKey(hdlr, rw, tt, SandBoxPrefix, key2)
}

// --------------------------------------------------------------------------------------------------------------------------------------------------------------------
/*
func (hdlr *QRRedirectHandlerType) ApiKeyToCustUser(apiKey string) (CustomerId, UserId string, err error) {

	rows, e0 := hdlr.gCfg.Pg_client.Db.Query(`select "customer_id", "user_id" from "q_api_key" where "id" = $1`, apiKey)
	err = e0
	if err != nil {
		fmt.Printf("Error %s, %s\n", err, godebug.LF())
	}

	for nr := 0; rows.Next(); nr++ {
		if nr >= 1 {
			logrus.Errorf("Error too many rows, should be unique primary key\n")
			err = fmt.Errorf("Error too many rows, should be unique primary key\n")
			return
		}

		err = rows.Scan(&CustomerId, &UserId)
		if err != nil {
			logrus.Errorf("Error scanning values, %s\n", err)
			err = fmt.Errorf("Error on d.b./scanning values query %s\n", err)
			return
		}
	}

	return

}

hh.DbUserCols = []string{"RealName", "Customer_id", "User_id", "UserName", "FirstName", "MidName", "LastName", "User_id", "Customer_id", "XAttrs", "PhoneNo"}

	// xyzzy - do a select on user_id in qr_user table, if not found then insert, else update.
*/

// Ok -- this is a really rotten way to implement this - but it will work for the moment.
// All of this should be passed down-stack to TabServer2 - via a rewrite of the call and a .Next -- When this is fixed remember to remove the connect to the database.
func (hdlr *AesSrpType) UpsertUserInfo(userId string, mdata map[string]string) {

	found := false
	qry := `select "id" from "qr_user" where "id" = $1`
	vals := make([]interface{}, 0, len(hdlr.DbUserCols)+1)
	vals = append(vals, userId)
	rows, err := hdlr.gCfg.Pg_client.Db.Query(qry, vals...)
	fmt.Printf("qry (at top): %s, data:%s\n", qry, vals)
	if err != nil {
		fmt.Printf("Error %s, %s\n", err, godebug.LF())
		return
	}

	for nr := 0; rows.Next(); nr++ {
		if nr >= 1 {
			logrus.Errorf("Error too many rows, should be unique primary key\n")
			err = fmt.Errorf("Error too many rows, should be unique primary key\n")
			return
		}
		found = true
	}

	if found {
		setcols := ""
		com := ""
		vals := make([]interface{}, 0, len(hdlr.DbUserCols)+1)
		vals = append(vals, userId)
		n := 0
		for ii, vv := range hdlr.DbUserCols {
			ww := vv
			if ii < len(hdlr.DbUserColsDb) && hdlr.DbUserColsDb[ii] != "" {
				ww = hdlr.DbUserColsDb[ii]
			}
			if v, ok := mdata[vv]; ok {
				n++
				setcols = setcols + com + fmt.Sprintf("\"%s\" = $%d", ww, ii+2)
				vals = append(vals, v)
				com = ", "
			}
		}
		if n > 0 {
			qry := fmt.Sprintf(`update "qr_user" set %s where "id" = $1`, setcols)
			fmt.Printf("qry: %s, data:%s\n", qry, vals)
			_, err := hdlr.gCfg.Pg_client.Db.Query(qry, vals...)
			if err != nil {
				fmt.Printf("Error %s, %s\n", err, godebug.LF())
			}
		}
	} else {
		inscols := ""
		insdol := ""
		com := ""
		vals := make([]interface{}, 0, len(hdlr.DbUserCols))
		n := 0
		for ii, vv := range hdlr.DbUserCols {
			ww := vv
			if ii < len(hdlr.DbUserColsDb) && hdlr.DbUserColsDb[ii] != "" {
				ww = hdlr.DbUserColsDb[ii]
			}
			if v, ok := mdata[vv]; ok {
				n++
				inscols = inscols + com + fmt.Sprintf("\"%s\"", ww)
				insdol = insdol + com + fmt.Sprintf("$%d", ii+1)
				vals = append(vals, v)
				com = ", "
			}
		}
		if n > 0 {
			qry := fmt.Sprintf(`insert into "qr_user" ( %s ) values ( %s )`, inscols, insdol)
			fmt.Printf("qry: %s, data:%s\n", qry, vals)
			_, err := hdlr.gCfg.Pg_client.Db.Query(qry, vals...)
			if err != nil {
				fmt.Printf("Error %s, %s\n", err, godebug.LF())
			}
		}
	}

	return

}

// if mdata["auth"] == "" && mdata["acct_type"] == "anon-user" && hdlr.TwoFactorRequired == "n" && hdlr.IsValidAnonUserPath(newPath0) {
func (hdlr *AesSrpType) IsValidAnonUserPath(path string) bool {
	ok := hdlr.anonUserPaths[path]
	if !ok {
		fmt.Printf("\nAttempt to access [%s] but not in the set of valid AnonUserPath in the configuraiton, 4007\n\n", path)
	}
	fmt.Printf("\naccess [%s], 4007\n\n", path)
	return ok
}

func respHandlerTest2faReturn(www http.ResponseWriter, req *http.Request) {
	OneTimeKey := "123456784"

	www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	www.Header().Set("Expires", "0")                                         // Proxies.

	www.WriteHeader(200)
	www.Header().Set("Content-Type", "application/json") // For JS
	fmt.Fprintf(www, `{"status":"success","version":1,"OneTimeKey":%q}`, OneTimeKey)
}

// -------------------------------------------------------------------------------------------------------------------------
// Inputs:
//		encData.Salt					-- Supplied form session data / per-user -- check this
//		encData.Iter					-- Default 1000
//		encData.KeySize 				-- In Bits
//		encData.KeySizeBytes			-- Derivable from KeySize/8			-- not sent in JSON data --
//		tt 								-- used as adata/verified 			-- session key
//  	plaintext						-- the JSON data to send back
//		key								-- the encryption key
// Output:
//		rv 								- the JSON encoded stirng -
//
// Misc Input
// 		db_respHandlerCiperh_1			-- debug flag
//
// -------------------------------------------------------------------------------------------------------------------------

func EncryptData(hdlr *AesSrpType, www http.ResponseWriter, req *http.Request, Salt base64data.Base64Data, Iter int, KeySize int, tt string, plaintext, key []byte, debugFlag1, debugFlag2 bool) (rv string, err error) {

	// 4. Encrypt into "rv"
	//     a. Create new return message
	//     b. Encrypt it
	cc := &sjcl.SJCL_DataStruct{
		Salt:           Salt,        //encData.Salt,          // Communication Salt
		Version:        1,           // Version of this message
		Iter:           Iter,        // encData.Iter,         // Number of iterations, normall 1000
		KeySize:        KeySize,     // encData.KeySize,      // Key Size in Bits
		TagSize:        64,          // In Bits
		Mode:           "ccm",       // Authentication Method (gcm might be metter but is not working yet)
		AdditionalData: []byte(tt),  // adata
		Cipher:         "aes",       // Encryption Method
		TagSizeBytes:   8,           // Tag Size
		KeySizeBytes:   KeySize / 8, // encData.KeySizeBytes, // Aes KeySize conv to bytes
		Status:         "success",   // Status of call
		Msg:            "",          // Additional Error Message - empty
	}

	ad := []byte(tt)
	var IV []byte

	IV, _ = gosrp.GenRandBytes(16) // 16 bytes of random Initialization Vector
	cc.InitilizationVector = IV

	nlen := aesccm.CalculateNonceLengthFromMessageLength(len(plaintext))
	nonce := IV[0:nlen]

	godebug.Printf(debugFlag2, "Nonce Length = %d\n", nlen)

	cb, e1 := aes.NewCipher(key) // var cb cipher.Block
	if e1 != nil {
		AnError(hdlr, www, req, 400, 1224, fmt.Sprintf("Error(0011): unable to setup AES:%s", e1))
		err = e1
		return
	}

	authmode, e2 := aesccm.NewCCM(cb, cc.TagSizeBytes, nlen) // var authmode cipher.AEAD, nlen is len(nonce)
	if e2 != nil {
		AnError(hdlr, www, req, 400, 1225, fmt.Sprintf("Error(0012): unable to setup CCM:%s", e2))
		err = e2
		return
	}

	newCipterText := authmode.Seal(nil, nonce, plaintext, ad)

	cc.CipherText = newCipterText

	rv = lib.SVar(cc)

	if debugFlag1 {
		fmt.Printf("// ////////////////////////////////////////////////////////////////////////////////////////////////////////\n")
		fmt.Printf("AT: %s \n", godebug.LF())
		fmt.Printf("Encrypted Return Value: rv -->>%s<<--\n", rv)
		fmt.Printf("// ////////////////////////////////////////////////////////////////////////////////////////////////////////\n")
	}

	return
}

//
// Password - is the SRP "key" as a hex string.
//	ses:"t"	- is the session stored, Example:
//		get "ses:dd910a87-881c-42ee-414b-3f798ab14e61"
// 		{"$auth$":"y","$auth_key$":"5b8eac60-e076-44ed-7950-233dc08a168a","$privs$":"user","$username$":"pschlump@uwyo.edu","Email":"pschlump@uwyo.edu"
//		,"Pw":"833052a9508baa7807298c4882c7740d38316f56fd7f70cf0904907280e1ddc1","login_date_time":"2016-09-23T19:54:37-06:00","login_fail_time":"","n_failed_login":"0"}
//

func DecryptData(hdlr *AesSrpType, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, SandBoxPrefix, Password, tEmail, tSalt string, encData *sjcl.SJCL_DataStruct, tIter, tKeySize int, tKey string, Session map[string]interface{}, tt string, debugFlag1, debugFlag2 bool) (plaintext, key []byte, err error) {

	// encData.Salt.Debug_hex(db1, "salt")
	// encData.InitilizationVector.Debug_hex(db1, "Initilization Vector")
	if debugFlag1 {
		fmt.Printf("// ////////////////////////////////////////////////////////////////////////////////////////////////////////\n")
		fmt.Printf("AT: %s Pw[%s] (shared key should match app_js_config.enc_key\n", godebug.LF(), Password)
		fmt.Printf("   tSalt    = %x\n", tSalt)
		fmt.Printf("   tKey     = %x\n", tKey)
		fmt.Printf("   tIter    = %v\n", tIter)
		fmt.Printf("   tKeySize = %v\n", tKeySize)
		fmt.Printf("   tEmail   = %v\n", tEmail)
		fmt.Printf("   Session  = %v\n", Session)
		fmt.Printf("// ////////////////////////////////////////////////////////////////////////////////////////////////////////\n")
		fmt.Printf("tSalt [%x] encData.Salt [%x] tKey [%x] tIter %d %d tKeySize %d %d\n",
			tSalt, string(encData.Salt), tKey, tIter, encData.Iter, tKeySize, encData.KeySize)
	}
	if tSalt == "" || tKey == "" || tSalt != string(encData.Salt) || tIter != encData.Iter || tKeySize != encData.KeySize {
		if dbCipher2 {
			fmt.Printf("KEY GEN: password[%s] salt[%x] iter[%d] keysize[%d]\n", Password, encData.Salt, encData.Iter, encData.KeySizeBytes)
		}
		// Generete the "key" using the shared secret password and other parameters.
		key = pbkdf2.Key([]byte(Password), encData.Salt, encData.Iter, encData.KeySizeBytes, sha256.New)
		// debug_hex("key", key)
		SaveKeyData(hdlr, rw, tt, SandBoxPrefix, Password, string(encData.Salt), string(key), encData.Iter, encData.KeySize)
	} else {
		key = []byte(tKey)
	}

	if debugFlag2 {
		fmt.Printf("key is [%x], salt is [%x], %s\n", key, tSalt, godebug.LF())
		fmt.Printf("At: %s, Ps=%s, req=%s\n", godebug.LF(), rw.Ps.DumpParamDB(), lib.SVarI(req)) // XyzzyDumpData
	}

	cb, err := aes.NewCipher(key) // var cb cipher.Block
	if err != nil {
		AnError(hdlr, www, req, 400, 1226, fmt.Sprintf("Error(0053): unable to setup AES:%s", err))
		err = ErrEarlyExit
		return
	}

	nonce, nlen := sjcl.GetNonce(*encData)

	// b. Decrypt the "ct" - validate it.
	authmode, err := aesccm.NewCCM(cb, encData.TagSizeBytes, nlen) // var authmode cipher.AEAD
	if err != nil {
		AnError(hdlr, www, req, 400, 1227, fmt.Sprintf("Error(0054): unable to setup CCM:%s", err))
		err = ErrEarlyExit
		return
	}

	plaintext, err = authmode.Open(nil, nonce, encData.CipherText, encData.AdditionalData)
	if err != nil {
		AnError(hdlr, www, req, 400, 1228, fmt.Sprintf("Error(0055): decrypting or authenticating using CCM:%s", err))
		err = ErrEarlyExit
		return
	}

	// tt should match 1st part of AD - authenticated data
	// Session["one-time-key"] - should match 2nd part IFF one-time-auth is true
	// Split on ','
	OneTimeKey := "x" // "x" will not match to any hex number
	AdditionalData := string(encData.AdditionalData)
	fmt.Printf("Additional Data = %s, %T\n", AdditionalData, encData.AdditionalData)
	tt_ad := ""
	if strings.Index(AdditionalData, ",") >= 0 {
		tt_v := strings.Split(AdditionalData, ",")
		tt_ad = tt_v[0]
		if len(tt_v) > 1 {
			OneTimeKey = tt_v[1]
		}
		fmt.Printf("One Time Key from AD = [%s], %s\n", OneTimeKey, godebug.LF())
	} else {
		tt_ad = AdditionalData
	}

	// Verify that the AdditonalData[first part, t] matches with the past session 't' value.
	if tt_ad != tt {
		AnError(hdlr, www, req, 400, 1229, fmt.Sprintf("Error(0056): AdditionalData failed to match session key, Error:%s", err))
		err = ErrEarlyExit
		return
	}

	// xyzzy - Questionable
	if hdlr.TwoFactorRequired == "y" && Session["auth"] == "P" { // using one time key - then validated that OneTimeKey is a match
		AnError(hdlr, www, req, 400, 1230, "Error(0096): In 2FA mode, but did not validate 2nd factor.")
		err = ErrEarlyExit
		return
	}

	// xyzzy - Questionable
	if hdlr.TwoFactorRequired == "y" && Session["auth"] == "y" { // using one time key - then validated that OneTimeKey is a match

		savedKey, ok := Session["$saved_one_time_key_hashed$"]
		if !ok {
			AnError(hdlr, www, req, 400, 1231, fmt.Sprintf("Error(0056): AdditionalData failed to match session key - one time key not saved - can not match, %s", err))
			err = ErrEarlyExit
			return
		}

		if savedKey != OneTimeKey {
			AnError(hdlr, www, req, 400, 1232, fmt.Sprintf("Error(0056): AdditionalData failed to match session key - one time key did match, %s", err))
			err = ErrEarlyExit
			return
		}
	}

	return

}

var ErrEarlyExit = errors.New("Early Exit - return")

const dbDumpURL = true
const dbEncr = true
const db100 = true

/* vim: set noai ts=4 sw=4: */
