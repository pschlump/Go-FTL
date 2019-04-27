package TabServer2

//
// Go-FTL / TabServer2
//
// Copyright (C) Philip Schlump, 2012-2017. All rights reserved.
//

import (
	"fmt"
	mathRand "math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/json"         //	"encoding/json"
	logrus "github.com/pschlump/pslog" // "github.com/sirupsen/logrus"
)

var hdrl_QRUrl = "http://127.0.0.1:9999/kick"
var hdrl_EmailUrl = "http://127.0.0.1:9998/kick"

//
//  "SetupQRCodeRedirect", "KickQRGen", "KickEmail"
//
func KickQRGen(www http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rvOut string, pptFlag PrePostFlagType, exit bool, a_status int) {
	ran := fmt.Sprintf("%d", mathRand.Intn(1000000000))
	hdlr.DoGet(hdrl_QRUrl, "_ran_", ran)
	return rv, PrePostSuccessWriteRV, true, 200
}

func KickEmail(www http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rvOut string, pptFlag PrePostFlagType, exit bool, a_status int) {
	ran := fmt.Sprintf("%d", mathRand.Intn(1000000000))
	hdlr.DoGet(hdrl_EmailUrl, "_ran_", ran)
	return rv, PrePostSuccessWriteRV, true, 200
}

type SetupQRData struct {
	Status      string `json:"status"`
	Use2fa      string `json:"use_2fa"`
	RanHash     string `json:"hash"`
	X2faId      string `json:"2fa_unique_id"`
	URL2faSetup string `json:"URL_2fa_setup"`
	UserID      string `json:"user_id"`
	QRImgURL    string `json:"QRImgURL"`
	QrID        string `json:"qr_id"`
	QrEncID     string `json:"qr_enc_id"`
}

// -- done --
// var QRShortURL = "http://t432z.com" // xyzzy make this a config item. (see :97)
// /Users/corwin/go/src/github.com/pschlump/Go-FTL/server/midlib/TabServer2/PostFx.go
// /Users/corwin/go/src/github.com/pschlump/Go-FTL/server/midlib/TabServer2/X2faSetup.go
//	 	var QRShortURL = "http://t432z.com" // xyzzy make this a config item. (see :97)

// Setup REDIS with data on timeout after registration.
// Setup QR code on qr-short.
func SetupQRCodeRedirect(www http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rvOut string, pptFlag PrePostFlagType, exit bool, a_status int) {

	var ed SetupQRData
	var all map[string]interface{}

	err := json.Unmarshal([]byte(rv), &ed)
	if err != nil {
		return "", PrePostFatalSetStatus, true, 500
	}
	err = json.Unmarshal([]byte(rv), &all)
	if err != nil {
		return "", PrePostFatalSetStatus, true, 500
	}

	if ed.Status == "success" && ed.Use2fa == "yes" {

		// ----------------------------------------------------------------------------------------------------------------------------------------
		// Push ID + random hash to Redis w/ TTL
		// id0, _ := uuid.NewV4()
		// t_2fa_ID := id0.String()
		t_2fa_ID := ed.X2faId
		RanHash := ed.RanHash

		// ----------------------------------------------------------------------------------------------------------------------------------------
		// get QR code from avail list
		/*
					||', "QRImgURL":'||to_json(l_qr_data.url_path)
					||', "qr_id":'||to_json(l_qr_data.qr_id)
					||', "qr_enc_id":'||to_json(l_qr_data.qr_enc_id)
			var qrId, QRImgUrl string
			qrId, QRImgUrl, err = hdlr.PullQRURLFromDB()
			godebug.DbPfb(db102, "%(Green) URL path: %s AT: %(LF)\n", QRImgUrl)
			if err != nil {
				godebug.DbPfb(true, "%(Red) QR Failed to get id=[%s] url=[%s] err=%s AT: %(LF)\n", qrId, QRImgUrl, err)
				return rv, PrePostRVUpdatedFail, true, http.StatusExpectationFailed // 417
			}
			if qrId == "" {
				godebug.DbPfb(true, "%(Red) QR Failed to get id=[%s] url=[%s] err=%s AT: %(LF)\n", qrId, QRImgUrl, err)
				return rv, PrePostRVUpdatedFail, true, http.StatusExpectationFailed // 417
			}
		*/
		qrId := ed.QrEncID // xyzzy - cleanup
		// QRImgURL := ed.QRImgURL // xyzzy - cleanup

		// ------------------------------------------------------------------------------------------------------
		// QR Short - setup
		// ------------------------------------------------------------------------------------------------------
		// update t432z.com URL shorter for this QR
		ran := ed.RanHash // xyzzy - cleanup
		godebug.DbPfb(db102, "%(Cyan)AT: %(LF) ran ed.RanHash [%v]\n", ran)

		theData := url.QueryEscape(fmt.Sprintf(`{"t_2fa_ID":%q}`, t_2fa_ID))

		// a432z.com - URL from config???

		var qrShortURL_upd string
		if hdlr.QRShortURL[len(hdlr.QRShortURL)-1:] == "/" {
			qrShortURL_upd = fmt.Sprintf("%supd/", hdlr.QRShortURL)
		} else {
			qrShortURL_upd = fmt.Sprintf("%s/upd/", hdlr.QRShortURL)
		}

		var sep string
		if strings.Contains(hdlr.DisplayURL2fa, "?") {
			sep = "&"
		} else {
			sep = "?"
		}

		var http string
		if req.TLS == nil {
			http = "http"
		} else {
			http = "https"
		}

		// Single Encode url=
		// toURL := url.QueryEscape(fmt.Sprintf("%s%sid=%s&url=%s://%s/api/x2fa/set-fp", hdlr.DisplayURL2fa, sep, RanHash, http, req.Host))
		// Doulbe Encoe url=
		toURL := url.QueryEscape(fmt.Sprintf("%s%sid=%s&url=%s", hdlr.DisplayURL2fa, sep, RanHash,
			url.QueryEscape(fmt.Sprintf("%s://%s/api/x2fa/set-fp", http, req.Host)),
		))

		fmt.Fprintf(os.Stderr, "%s?url=%s&id=%s&data=%s&_ran_=%s\n", qrShortURL_upd, toURL, qrId, theData, ran)
		fmt.Fprintf(os.Stdout, "%s?url=%s&id=%s&data=%s&_ran_=%s\n", qrShortURL_upd, toURL, qrId, theData, ran)

		status, body := hdlr.DoGet(qrShortURL_upd, "url", toURL, "id", qrId, "data", theData, "_ran_", ran)
		if status != 200 {
			fmt.Fprintf(os.Stderr, "\n%sQR Redirect Setup: status=%d AT:%s%s\n\n", MiscLib.ColorRed, status, godebug.LF(), MiscLib.ColorReset)
			fmt.Fprintf(os.Stdout, "\n%sQR Redirect Setup: status=%d AT:%s%s\n\n", MiscLib.ColorRed, status, godebug.LF(), MiscLib.ColorReset)
			logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to set QR Redirect","LineFile":%q}`+"\n", err, godebug.LF()))
			// return "", "", "", 0, fmt.Errorf("Unable to set QR Redirect, Error [%s] AT: %s", err, godebug.LF())
			return rv, PrePostRVUpdatedFail, true, 401
		} else {
			godebug.DbPfb(db102, "%(Green) body from shortner : %s AT: %(LF)\n", body)
			if body != qrId {
				godebug.DbPfb(db102, "%(Red) DID NOT MATCH!!!!: got [%s] != expected [%s] AT: %(LF)\n", body, qrId)
			}
		}

		// ------------------------------------------------------------------------------------------------------
		// Redis
		// ------------------------------------------------------------------------------------------------------
		conn, err := hdlr.gCfg.RedisPool.Get()
		defer hdlr.gCfg.RedisPool.Put(conn)
		if err != nil {
			logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
			// return "", "", "", 0, fmt.Errorf("Failed to connect to Redis, Error [%s] AT: %s", err, godebug.LF())
			return rv, PrePostRVUpdatedFail, true, 401
		}

		// ID = fmt.Sprintf("%d", mathRand.Intn(10000000)) // xyzzy201 - add in Checksum byte
		ID := ed.RanHash
		key := fmt.Sprintf("%s%s", hdlr.RedisPrefix2fa, ID)
		var host string
		host = ""
		if req.TLS != nil {
			host = "https://" + req.Host
		} else {
			host = "http://" + req.Host
		}
		val := godebug.SVar(RedisData{
			Hash:   RanHash,
			Fp:     "fingerprint-not-set-yet",
			UserID: ed.UserID,
			T2faID: t_2fa_ID,
			URL:    host,
		})
		ttl := timeOutConst // 60 * 60 // 1 hour	// xyzzySetConfig - in seconds

		err = conn.Cmd("SETEX", key, ttl, val).Err
		if err != nil {
			if db4 {
				fmt.Printf("Error on redis - get(%s): %s\n", key, err)
			}
			// return "", "", "", 0, fmt.Errorf(`{"status":"error","msg":"Unable to set value in Redis.","LineFile":%q}`, godebug.LF())
			return rv, PrePostRVUpdatedFail, true, 401
		}

		delete(all, "x2fa_unique_id")

		// ----------------------------------------------------------------------------------------------------------------------------------------
		// Results back to user.
		// ----------------------------------------------------------------------------------------------------------------------------------------
		rv = godebug.SVar(all)
		fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())
		// return rv, PrePostRVUpdatedSuccess, true, 200
		return rv, PrePostSuccessWriteRV, true, 200
	} else if ed.Status == "error" {
		fmt.Printf("Error from underlying code - exit early - rv= ->%s<- at:%s\n", rv, godebug.LF())
		return rv, PrePostRVUpdatedFail, true, 401
	}

	fmt.Printf("rv= ->%s<- at:%s\n", rv, godebug.LF())
	return fmt.Sprintf(`{"status":"error","msg":"failed at %s"}`, godebug.LF()), PrePostRVUpdatedFail, true, 401
}

//

//

//

//

/*
type RedisData struct {
	Hash   string `json:"hash"`
	Fp     string `json:"fp"`
	T2faID string `json:"t_2fa_id"`
	UserID string `json:"user_id"`
	URL    string `json:"URL"`
}

// xyzzy - config item
const timeOutConst = (60 * 60 * 24) + 5

// xyzzy-2fa - X2faSetup
// type PrePostFlagType int - to replace 'bool' type.
func X2faSetup(www http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rvOut string, pptFlag PrePostFlagType, exit bool, a_status int) {

	fmt.Printf("%sX2faSetup! AT:%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())
	fmt.Fprintf(os.Stderr, "\n%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())

	// func SignToken(tokData []byte, keyFile string) (out string, err error) {
	//	hdlr.KeyFilePrivate        string                      // private key file for signing JWT tokens
	// https://github.com/dgrijalva/jwt-go.git
	/ *
	   {
	       "auth_token": "46155d84-1de9-418f-b22b-314b8d228ec1",
	       "config": "{}",
	       "customer_id": "1",
	       "jwt_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRoX3Rva2VuIjoiNDYxNTVkODQtMWRlOS00MThmLWIyMmItMzE0YjhkMjI4ZWMxIn0.M80E47h-ntyVpvGU7cdDFfsATUUQ8vW95NH-HtJSCxMVOFhxT_ovQo7sUCf0cQ_ALDnLq_Aoa757ZQMRDRf7bi2-L3j59_FliFvrM53Gnhe5b2ga8AiGpdVbNHGsJHPu-ZLu0zY9n4MPpYXWGrzii4Nn7kuR_0STzDEIt83NUwOcGRowoZZGTiwdqFq5Buma021BwsCfC6TStPm5tfrOB7R8kpNlvtm7s87HZ4mGJoKE-eMBUnmEsEhinQXGculbelAZ4jL8yt6z0MOagOQNNdchX1S827IUQ99chSCWuM52aXC_gb6aydNUMprvYZkIR0kVm43nw4hXhTZP27ghmw",
	       "privs": "[]",
	       "ranch_id": "971663ca-b4d5-484b-8210-f60cba218669",
	       "redir_to_app": "http://localhost:3000/newly-registered",
	       "seq": "605b7cae-4363-4b37-888a-39f6ae3d6d2b",
	       "status": "success",
	       "use_2fa": "yes",
	       "user_id": "f7ab4a0d-c53d-44a7-b869-49bb81b8919a",
	       "xsrf_token": "05266ceb-ee79-4c91-9418-c4f3a6b267fd"
	   }
	* /

	var ed RedirectToData
	var all map[string]interface{}

	err := json.Unmarshal([]byte(rv), &ed)
	if err != nil {
		return "", PrePostFatalSetStatus, true, 500
	}
	err = json.Unmarshal([]byte(rv), &all)
	if err != nil {
		return "", PrePostFatalSetStatus, true, 500
	}

	if ed.Status == "success" && ed.Use2fa == "yes" {

		fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())

		html, QRImgURL, ID, qr_ttl, err := GetQRForSetup(hdlr, www, req, ps, ed.UserID)
		if err != nil {
			fmt.Fprintf(www, `{"status":"error","msg":"Error [%s]","LineFile":%q}`, err, godebug.LF())
			return "{\"status\":\"error\"}", PrePostRVUpdatedFail, true, 200 // xyzzy - better error return
		}

		all["html_2fa"] = html
		all["QRImgURL"] = QRImgURL
		all["X2fa_Temp_ID"] = ID
		all["qr_ttl"] = fmt.Sprintf("%v", qr_ttl)

		// delete(all, "user_id")

		rv = godebug.SVar(all)
		fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())
		// return rv, PrePostRVUpdatedSuccess, true, 200
		return rv, PrePostSuccessWriteRV, true, 200
	} else if ed.Status == "error" {
		fmt.Printf("Error from underlying code - exit early - rv= ->%s<- at:%s\n", rv, godebug.LF())
		return rv, PrePostRVUpdatedFail, true, 401
	}

	fmt.Printf("rv= ->%s<- at:%s\n", rv, godebug.LF())
	return fmt.Sprintf(`{"status":"error","msg":"failed at %s"}`, godebug.LF()), PrePostRVUpdatedFail, true, 401
}

// xyzzy-2fa - X2faValidateToken
// rv - return value string - JSON
// rexit, if true, then will return with error from parent
// rstatus - status to return with
// func X2faValidateToken(www http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rrv string, rexit bool, rstatus int) {
func X2faValidateToken(www http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rvOut string, pptFlag PrePostFlagType, exit bool, a_status int) {

	fmt.Printf("%sAT:%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())
	fmt.Fprintf(os.Stderr, "\n\n%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())


	type RedirectToData struct {
		Status string `json:"status"`
		UserID string `json:"user_id"`
	}

	var ed RedirectToData
	var all map[string]interface{}

	err := json.Unmarshal([]byte(rv), &ed)
	if err != nil {
		return "", PrePostFatalSetStatus, true, 500
	}
	err = json.Unmarshal([]byte(rv), &all)
	if err != nil {
		return "", PrePostFatalSetStatus, true, 500
	}

	if ed.Status == "success" { // this means UN/PW are ok, is not a blocked IP address etc.  Account not expired etc.

		fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())

		// all["html_2fa"] = html
		all["2fa"] = "is *NOT* valid"

		user_id := ed.UserID

		delete(all, "user_id")

		rv = godebug.SVar(all)
		fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())

		fmt.Printf("IsValid2fa called\n")
		fmt.Fprintf(os.Stderr, "IsValid2fa called\n")

		val2fa := ps.ByNameDflt("val2fa", "")
		godebug.DbPfb(db1x2fa, "val2fa: ->%s<-\n", val2fa)

		var err error
		godebug.DbPfb(db1x2fa, "%(Cyan) user_id = %q AT: %(LF)\n", user_id)

		// generate local copy based on user_id/auth_token - for all rows in t_2fa and any values in t_2fa_otk
		LocalVal2fa, err := hdlr.GetValidList(user_id)
		if err != nil {
			rv = fmt.Sprintf(`{"status":"error","msg":"PG Database Error: %s","LineFile":%q}`, err, godebug.LF())
			fmt.Fprintf(os.Stderr, `{"status":"error","msg":"PG Database Error: %s","LineFile":%q}`+"\n", err, godebug.LF())
			return rv, PrePostRVUpdatedFail, false, 200
		}
		godebug.DbPfb(db1x2fa, "%(Cyan) Local Values Array = %s AT: %(LF)\n", godebug.SVarI(LocalVal2fa))

		for _, v := range LocalVal2fa {
			if v == val2fa {
				stmt := "delete from \"t_2fa_otk\" where \"user_id\" = $1 and \"one_time_key\" = $2"
				_, err := hdlr.gCfg.Pg_client.Db.Query(stmt, user_id, v)
				if err != nil {
					fmt.Printf("Database error %s, attempting to convert premis_id/animal_id to tag.\n", err)
				}
				all["2fa"] = "is valid. Yea!"
				rv = godebug.SVar(all)
				godebug.DbPfb(db1x2fa, "%(Green) SHOULD BE SUCCESS rv = %s AT: %(LF), Parent = %s, p2 = %s\n", rv, godebug.LF(2), godebug.LF(3))
				// return rv, PrePostRVUpdatedSuccess, false, 200
				return rv, PrePostSuccessWriteRV, false, 200
			}
		}

		rv = fmt.Sprintf(`{"status":"error","msg":"Two Factor Did Not Match","LineFile":%q}`, godebug.LF())
		fmt.Fprintf(www, rv)
		return rv, PrePostRVUpdatedFail, true, 200
	}

	rv = fmt.Sprintf(`{"status":"error","msg":"Two Factor Did Not Match","LineFile":%q}`, godebug.LF())
	return rv, PrePostRVUpdatedFail, true, 200
}

const db1x2fa = true

func (hdlr *TabServer2Type) Get2MinHashFunc() (hash string, ttlLeft int, err error) {
	fmt.Printf("Get2MinHashFunc called\n")
	fmt.Fprintf(os.Stderr, "Get2MinHashFunc called\n")

	// ------------------------------------------------------------------------------
	// Get Connection
	// ------------------------------------------------------------------------------
	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return "", 0, err
	}

	godebug.DbPfb(db1, "%(Cyan) (err may be nil) [%s] AT: %(LF)\n", err)

	// ------------------------------------------------------------------------------
	// Construct key for "GET" then Get data and TTL
	// ------------------------------------------------------------------------------
	RanHashBytes, ttl, _, e0 := GenRandBytesOracle()
	fmt.Fprintf(os.Stderr, "\n\nHashBytes [%x]\n", RanHashBytes)
	if e0 != nil {
		err = e0
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to generate random data.","LineFile":%q}`+"\n", err, godebug.LF()))
		// fmt.Fprintf(www, `{"status":"error","LineFile":%q}`, godebug.LF())
		return
	}
	hash = fmt.Sprintf("%x", RanHashBytes)
	godebug.DbPfb(db1, "%(Cyan) hash(returned)=[%s] AT: %(LF)\n", hash)

	// ----------------------------------------------------------------------
	// SUCCESS return
	// ----------------------------------------------------------------------
	ttlLeft = ttl
	return
}

// ============================================================================================================================================
// Should move to aesccm package
func GenRandBytesOracle() (buf []byte, ttl, epoc int, err error) {
	URL := "http://www.2c-why.com/Ran/RandomValue"
	var status int
	var body string

	if FirstRequest {
		ran := fmt.Sprintf("%d", mathRand.Intn(1000000000))
		// status, body := hdlr.DoGet("http://t432z.com/upd/", "url", hdlr.DisplayURL, "id", qrId, "data", theData, "_ran_", ran)
		status, body = hdlr.DoGet(URL, "_ran_", ran) // Get Random from Oralce
	} else {
		status, body = hdlr.DoGet(URL, "ep", fmt.Sprintf("%v", ThisEpoc)) // xyzzy Deal with TTL and timing to see if need to re-fetch. // random from oracle
		// xyzzy use TimeRemain, ThisEpoc, LastResult
	}

	if status != 200 {
		fmt.Printf("Unable to get RandomOracle - what to do, status = %v\n", status)
		fmt.Fprintf(os.Stderr, "Unable to get RandomOracle - what to do, status = %v\n", status)
		buf = make([]byte, 32)
		return
	}

	fmt.Fprintf(os.Stderr, "%sRandomValue%s ->%s<- AT:%s\n", MiscLib.ColorYellow, MiscLib.ColorReset, body, godebug.LF())

	// fmt.Fprintf(www, `{"status":"success","value":"%x","ttl":%d,"ep":%v}`, aValue, ttlCurrent, epoc_120)
	var pd RanData
	err = json.Unmarshal([]byte(body), &pd)
	if pd.Status != "success" {
		fmt.Printf("Unable to get RandomOracle - what to do, status = %v\n", status)
		fmt.Fprintf(os.Stderr, "Unable to get RandomOracle - what to do, status = %v\n", status)
		buf = make([]byte, 32)
		return
	}

	buf, err = hex.DecodeString(pd.Value)
	if err != nil {
		fmt.Printf("Unable to get RandomOracle - what to do, err = %v\n", err)
		fmt.Fprintf(os.Stderr, "Unable to get RandomOracle - what to do, err = %v\n", err)
		buf = make([]byte, 32)
		return
	}

	FirstRequest = false

	TimeRemain = pd.TTL
	ThisEpoc = pd.Epoc

	ttl = pd.TTL
	epoc = pd.Epoc

	return
}

// ------------------------------------------------------------------------------------------------------------------------
// ------------------------------------------------------------------------------------------------------------------------
// GetValidList get list of convened to string int values for valid 2fa
func (hdlr *TabServer2Type) GetValidList(user_id string) (list []string, err error) {

	stmt := `
select 'current' as "ty", "user_hash", "fp", 'x' as "one_time_key"
	from "t_2fa"
	where "user_id" = $1
	  and "fp" is not null
union
	select 'otk' as "ty", 'x' as "user_hash", 'x' as "fp", "one_time_key"
	from "t_2fa_otk"
	where "user_id" = $1
order by 1, 2
`
	rows, err := hdlr.gCfg.Pg_client.Db.Query(stmt, user_id)
	if err != nil {
		fmt.Printf("Database error %s, attempting to convert premis_id/animal_id to tag.\n", err)
		return
	}
	defer rows.Close()
	current2MinHash, _, err := hdlr.Get2MinHashFunc()
	godebug.DbPfb(db1, "%(Yellow) AT: %(LF), current2MinHash=%s\n", current2MinHash)
	for nr := 0; rows.Next(); nr++ {

		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
		var ty, user_hash, fp, one_time_key string
		err = rows.Scan(&ty, &user_hash, &fp, &one_time_key)
		if err != nil {
			fmt.Printf("Error on d.b. query %s\n", err)
			return
		}
		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")

		if ty == "okt" {
			list = append(list, one_time_key)
		} else {
			val0 := HashStrings.Sha256(fmt.Sprintf("%s:%s:%s", user_hash, fp, current2MinHash))
			// val1 := fmt.Sprintf("%x", val0)
			val1 := string(val0)
			val2 := val1[len(val1)-6:]
			val, err := strconv.ParseInt(val2, 16, 64)
			if err != nil {
				fmt.Printf("Error on d.b. query %s\n", err)
				continue
			}
			val = val % 1000000
			list = append(list, fmt.Sprintf("%06d", val))
		}

	}
	return
}

// ------------------------------------------------------------------------------------------------------------------------
// ------------------------------------------------------------------------------------------------------------------------
func (hdlr *TabServer2Type) GetUserIDFromAuthToken(auth_token string) (user_id string, err error) {
	// 		user_id, err = hdlr.GetUserIDFromAuthToken ( auth_token );

	stmt := `select "user_id" from "t_auth_token" where "auth_token" = $1`
	rows, err := hdlr.gCfg.Pg_client.Db.Query(stmt, user_id)
	if err != nil {
		fmt.Printf("Database error %s, attempting to convert premis_id/animal_id to tag.\n", err)
		return
	}
	defer rows.Close()
	for nr := 0; rows.Next(); nr++ {

		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
		err = rows.Scan(&user_id)
		if err != nil {
			fmt.Printf("Error on d.b. query %s\n", err)
			return
		}
		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
		return
	}
	fmt.Printf("Error on d.b. query -got 0 rows\n")
	return "", nil
}

// ============================================================================================================================================
var FirstRequest bool = true
var TimeRemain int
var ThisEpoc int
var LastResut []byte

type RanData struct {
	Status string `json:"status"`
	Value  string `json:"value"`
	TTL    int    `json:"ttl"`
	Epoc   int    `json:"ep"`
}

func GetQRForSetup(hdlr *TabServer2Type, www http.ResponseWriter, req *http.Request, ps *goftlmux.Params, user_id string) (html, QRImgURL, ID string, qr_ttl int, err error) {
	fmt.Printf("getQRForSetup called -- TabServer2Type\n")
	fmt.Fprintf(os.Stderr, "getQRForSetup called -- TabServer2Type\n")

	godebug.DbPfb(db1, "user_id: ->%s<-\n", user_id)

	// ----------------------------------------------------------------------------------------------------------------------------------------
	// Generate ID
	ID = fmt.Sprintf("%d", mathRand.Intn(10000000)) // xyzzy201 - add in Checksum byte
	// Generate Random Hash
	RanHashBytes, err := GenRandBytes(32)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to generate random data.","LineFile":%q}`+"\n", err, godebug.LF()))
		return "", "", "", 0, fmt.Errorf("Random Generation Failed")
	}
	RanHash := fmt.Sprintf("%x", RanHashBytes)
	// func GenRandNumber(nDigits int) (buf string, err error) {
	// func GenRandBytes(nRandBytes int) (buf []byte, err error) {

	// ----------------------------------------------------------------------------------------------------------------------------------------
	// get QR code from avail list
	var qrId, QRImgUrl string
	qrId, QRImgUrl, err = hdlr.PullQRURLFromDB()
	godebug.DbPfb(db1, "%(Green) URL path: %s AT: %(LF)\n", QRImgUrl)

	// ----------------------------------------------------------------------------------------------------------------------------------------
	// update t432z.com URL shorter for this QR
	ran := fmt.Sprintf("%d", mathRand.Intn(1000000000))
	godebug.DbPfb(db1, "%(Cyan)AT: %(LF) ran [%v]\n", ran)

	theData := `{"data":"data written to system in file"}`
	// a432z.com - URL from config???
	fmt.Printf("http://t432z.com/upd/?url=%s&id=%s&data=%s&_ran_=%s\n", hdlr.DisplayURL2fa, qrId, theData, ran)
	fmt.Fprintf(os.Stderr, "http://t432z.com/upd/?url=%s&id=%s&data=%s&_ran_=%s\n", hdlr.DisplayURL2fa, qrId, theData, ran)
	status, body := hdlr.DoGet("http://t432z.com/upd/", "url", hdlr.DisplayURL2fa, "id", qrId, "data", theData, "_ran_", ran)
	if status != 200 {
		fmt.Printf("status=%d\n", status)
		fmt.Fprintf(os.Stderr, "status=%d\n", status)
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to set QR Redirect","LineFile":%q}`+"\n", err, godebug.LF()))
		return "", "", "", 0, fmt.Errorf("Unable to set QR Redirect, Error [%s] AT: %s", err, godebug.LF())
	} else {
		godebug.DbPfb(db1, "%(Green) body from shortner : %s AT: %(LF)\n", body)
	}

	// ----------------------------------------------------------------------------------------------------------------------------------------
	// Push ID + random hash to Redis w/ TTL
	id0, _ := uuid.NewV4()
	t_2fa_ID := id0.String()
	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return "", "", "", 0, fmt.Errorf("Failed to connect to Redis, Error [%s] AT: %s", err, godebug.LF())
	}

	key := fmt.Sprintf("%s%s", hdlr.RedisPrefix2fa, ID)
	var host string
	host = ""
	if req.TLS != nil {
		host = "https://" + req.Host
	} else {
		host = "http://" + req.Host
	}
	val := godebug.SVar(RedisData{
		Hash:   RanHash,
		Fp:     "fingerprint-not-set-yet",
		UserID: user_id,
		T2faID: t_2fa_ID,
		URL:    host,
	})
	ttl := timeOutConst // 60 * 60 // 1 hour	// xyzzySetConfig - in seconds
	qr_ttl = ttl

	err = conn.Cmd("SETEX", key, ttl, val).Err
	if err != nil {
		if db4 {
			fmt.Printf("Error on redis - get(%s): %s\n", key, err)
		}
		return "", "", "", 0, fmt.Errorf(`{"status":"error","msg":"Unable to set value in Redis.","LineFile":%q}`, godebug.LF())
	}

	// ----------------------------------------------------------------------------------------------------------------------------------------
	// Setup OTK - 20 values for OTKs
	for i := 0; i < 20; i++ {
		rv, err := GenRandNumber(6)
		if err != nil {
			logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to generate random value.","LineFile":%q}`+"\n", err, godebug.LF()))
			return "", "", "", 0, fmt.Errorf("Unabgle to generate randome value AT: %s", err, godebug.LF())
		}

		stmt := "insert into \"t_2fa_otk\" ( \"user_id\", \"one_time_key\" ) values ( $1, $2 )"
		_, err = hdlr.gCfg.Pg_client.Db.Exec(stmt, user_id, rv)
		if err != nil {
			logrus.Warn(fmt.Sprintf(`{"msg":"Error %s PG error.","LineFile":%q}`+"\n", err, godebug.LF()))
			return "", "", "", 0, fmt.Errorf("PG error %s AT: %s", err, godebug.LF())
		}
	}

	// ----------------------------------------------------------------------------------------------------------------------------------------
	// Insert random hash -> PG to t_user
	stmt := "insert into \"t_2fa\" ( \"id\", \"user_id\", \"user_hash\" ) values ( $1, $2, $3 )"
	_, err = hdlr.gCfg.Pg_client.Db.Exec(stmt, t_2fa_ID, user_id, RanHash)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s PG error.","LineFile":%q}`+"\n", err, godebug.LF()))
		return "", "", "", 0, fmt.Errorf("PG error %s AT: %s", err, godebug.LF())
	}

	// ----------------------------------------------------------------------------------------------------------------------------------------
	// Send back results.
	// ----------------------------------------------------------------------------------------------------------------------------------------
	// Use temlate to send back results.
	// xyzzyTemplate -template- $(project)/tmpl/qr
	// /Users/corwin/go/src/github.com/pschlump/Go-FTL/server/goftl/tmpl/qr_success_resp.tmpl
	if false {
		html = fmt.Sprintf(
			`<div class="getQRForSetup">
			<img src=%q>
			<div>
				Scan the QR code above to setup your mobile device or browse<br>
				on your mobile device to <a href="%s/msetup.html?id=%v">%s/msetup.html</a><br>
				and enter %v.
			</div>
		</div>`, QRImgUrl, hdlr.Server2faURL, ID, hdlr.Server2faURL, ID)
	} else {
		mdata := make(map[string]interface{})

		mdata["QRImgUrlQuoted"] = fmt.Sprintf("%q", QRImgUrl)
		mdata["ID"] = ID // Decimal Int
		mdata["hdlr_Server2faURL"] = hdlr.Server2faURL
		mdata["hdlr_CustomerURL_UrlEncoded"] = url.QueryEscape(hdlr.CustomerURL)
		mdata["QR_valid_for"] = qr_ttl // Seconds as an integer
		mdata["qr_ttl"] = "seconds"

		html = RunTemplate(hdlr.QrSuccessTemplateFn, "qr_success", mdata)
	}

	return
}

// qrId, QRImgUrl, err = hdlr.PullQRURLFromDB()
// err = hdlr.PullQRFromDB(rr.Tag)
func (hdlr *TabServer2Type) PullQRURLFromDB() (qr_enc_id, qr_url string, err error) {
	// Xyzzy - sould replace with stored proc. that updates state in same transaction.
	stmt := "select \"qr_enc_id\", \"url_path\" from \"t_avail_qr\" where \"state\" = 'avail' limit 1"
	// insert into "t_avail_qr" ( "qr_id", "qr_enc_id", "url_path", "file_name", "qr_encoded_url_path" ) values
	// 	  ( '170', '4q', 'http://127.0.0.1:9019/qr/00170.4.png', './td_0008/q00170.4.png', 'http://t432z.com/q/4q' )
	rows, err := hdlr.gCfg.Pg_client.Db.Query(stmt)
	if err != nil {
		fmt.Printf("Database error %s, attempting to convert premis_id/animal_id to tag.\n", err)
		return "", "", err
	}
	defer rows.Close()
	godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
	for nr := 0; rows.Next(); nr++ {
		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
		if nr >= 1 {
			fmt.Printf("Error too many rows for a user, should be unique primary key\n")
			break
		}

		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
		var qr string
		err := rows.Scan(&qr, &qr_url)
		if err != nil {
			fmt.Printf("Error on d.b. query %s\n", err)
			return "", "", err
		}
		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")

		// Xyzzy - test fail to error report
		err = hdlr.UpdateQRMarkAsUsed(qr)
		if err != nil {
			fmt.Printf("Error on d.b. query %s\n", err)
			return "", "", err
		}

		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
		return qr, qr_url, nil
	}
	return "", "", fmt.Errorf("Failed to get a QR code")
}

// Modified to send Header!
//
// ---------------------------------------------
// // Xyzzy101 - Setup QR Redirect
// ---------------------------------------------
//
// 	export QR_SHORT_AUTH_TOKEN="w4h0wvtb1zk4uf8Xv.Ns9Q7j8"
// 	wget -o out/,list1 -O out/,list2 \
// 		--header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" \
// 		"http://t432z.com/upd/?url=http://test.test.com&id=5c"
//
// 	-- 1. hdlr.DoGet - change to create a header
// 	-- 2. Example Call to set this
// * /
func DoGet(uri string, args ...string) (status int, rv string) {

	sep := "?"
	var qq bytes.Buffer
	qq.WriteString(uri)
	for ii := 0; ii < len(args); ii += 2 {
		// q = q + sep + name + "=" + value;
		qq.WriteString(sep)
		qq.WriteString(url.QueryEscape(args[ii]))
		qq.WriteString("=")
		if ii < len(args) {
			qq.WriteString(url.QueryEscape(args[ii+1]))
		}
		sep = "&"
	}
	url_q := qq.String()

	// res, err := http.Get(url_q)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url_q, nil)
	req.Header.Add("User-Agent", "Go-FTL-x2fa")
	req.Header.Add("X-Qr-Auth", "w4h0wvtb1zk4uf8Xv.Ns9Q7j8") // Xyzzy - set from config?
	res, err := client.Do(req)

	if err != nil {
		return 500, ""
	} else {
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return 500, ""
		}
		status = res.StatusCode
		if status == 200 {
			rv = string(body)
		}
		return
	}
}

func (hdlr *TabServer2Type) UpdateQRMarkAsUsed(qrId string) error {
	stmt := "update \"t_avail_qr\" set \"state\" = 'used' where \"qr_enc_id\" = $1"
	godebug.DbPfb(db1, "%(Yellow) AT: %(LF) - stmt [%s] data[%s]\n", stmt, qrId)
	_, err := hdlr.gCfg.Pg_client.Db.Exec(stmt, qrId)
	if err != nil {
		return err
	} else {
		fmt.Printf("Success: %s data[%s]\n", stmt, qrId)
		fmt.Fprintf(os.Stderr, "Success: %s data[%s]\n", stmt, qrId)
	}
	return nil
}

// ===================================================================================================================================================
// Run a template and get the results back as a stirng.
// This is the primary template runner for sending email.
func RunTemplate(TemplateFn string, name_of string, g_data map[string]interface{}) string {

	rtFuncMap := template.FuncMap{
		"Center":      ms.CenterStr,   //
		"PadR":        ms.PadOnRight,  //
		"PadL":        ms.PadOnLeft,   //
		"PicTime":     ms.PicTime,     //
		"FTime":       ms.StrFTime,    //
		"PicFloat":    ms.PicFloat,    //
		"nvl":         ms.Nvl,         //
		"Concat":      ms.Concat,      //
		"title":       strings.Title,  // The name "title" is what the function will be called in the template text.
		"ifDef":       ms.IfDef,       //
		"ifIsDef":     ms.IfIsDef,     //
		"ifIsNotNull": ms.IfIsNotNull, //
		// "g":           global_g,       //
		// "set":         global_set,     //
	}

	var b bytes.Buffer
	foo := bufio.NewWriter(&b)

	t, err := template.New("simple-tempalte").Funcs(rtFuncMap).ParseFiles(TemplateFn)
	// t, err := template.New("simple-tempalte").ParseFiles(TemplateFn)
	if err != nil {
		fmt.Printf("Error(12004): parsing/reading template, %s\n", err)
		return ""
	}

	err = t.ExecuteTemplate(foo, name_of, g_data)
	if err != nil {
		fmt.Fprintf(foo, "Error(12005): running template=%s, %s\n", name_of, err)
		return ""
	}

	foo.Flush()
	s := b.String() // Fetch the data back from the buffer

	// LogIt()
	fmt.Fprintf(os.Stdout, "Template Output is: ----->%s<-----\n", s)

	return s

}

// status, body := DoGet("http://t432z.com/upd/", "url", hdlr.DisplayURL2fa, "id", qrId, "data", theData, "_ran_", ran)
// key := fmt.Sprintf("%s%s", hdlr.RedisPrefix2fa, ID)
*/

const db102 = true

/* vim: set noai ts=4 sw=4: */
