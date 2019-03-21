//
// Go-FTL - Module - Support for Two Factor Auth (2FA)
//
// Copyright (C) Philip Schlump, 2009-2017.
//

package X2fa

/*
Note:
	https://api.etherscan.io/api?module=proxy&action=eth_blockNumber&apikey=YourApiKeyToken
		{"jsonrpc":"2.0","id":83,"result":"0x6d3d28"}
		{"jsonrpc":"2.0","id":83,"result":"0x6d3d29"}
		From: https://ethereum.stackexchange.com/questions/49726/how-do-i-get-the-latest-block-using-etherscan-api
		https://api.etherscan.io/api?module=block&action=getblockreward&blockno=0x6d3d29&apikey=YourApiKeyToken

	Even easier:
		https://api.blockcypher.com/v1/eth/main

			{
			  "name": "ETH.main",
			  "height": 7069494,
			  "hash": "f143602c86ad61a992cb4f2ec82714977c62efd8d138fc311e7ca7fcdb7476fa",
			  "time": "2019-02-01T05:00:09.8589251Z",
			  "latest_url": "https://api.blockcypher.com/v1/eth/main/blocks/f143602c86ad61a992cb4f2ec82714977c62efd8d138fc311e7ca7fcdb7476fa",
			  "previous_hash": "00dd3b5c702ec9938964dd95667e4ba5d256c47b4c399f77dc95abee6d44b387",
			  "previous_url": "https://api.blockcypher.com/v1/eth/main/blocks/00dd3b5c702ec9938964dd95667e4ba5d256c47b4c399f77dc95abee6d44b387",
			  "peer_count": 64,
			  "unconfirmed_count": 46,
			  "high_gas_price": 5838392818,
			  "medium_gas_price": 5838392818,
			  "low_gas_price": 5000000000,
			  "last_fork_height": 7034372,
			  "last_fork_hash": "8cfaabc0b64ce7df52474e45b258046529648fc1c7b69ebb622989e69c7fb33f"
			}

Notes:

	Oracle Project: working now.
		1. Put Oracle in ~/go/src/www.2c-why.com/RandomOracle

*/

// Oracle In: 1. Put Oracle in ~/go/src/www.2c-why.com/RandomOracle
// API:
//	http://.../GetValue?_ran_=

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/HashStrings"
	JsonX "github.com/pschlump/JSONx"
	"github.com/pschlump/godebug"
	logrus "github.com/pschlump/pslog" // "github.com/sirupsen/logrus"
	"github.com/pschlump/uuid"
)

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &X2faType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // Xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("X2fa", CreateEmpty, `{
		"Paths":        	{ "type":["string","filepath"], "isarray":true, "required":true },
		"RedisPrefix":  	{ "type":[ "string" ], "required":false, "default":"2fa:" },
		"TemplatePath":  	{ "type":[ "string" ], "required":false, "default":"./tmpl" },
		"QRPath":  	 		{ "type":[ "string" ], "required":false, "default":"/qr/" },
		"QRURLPath":  	 	{ "type":[ "string" ], "required":false, "default":"/qr/" },
		"DisplayURL":	  	{ "type":[ "string" ], "required":false, "default":"/2fa/2fa-app.html" },
		"TimeoutCodes":     { "type":[ "int" ], "default":"120" }
		"Server2faURL":	  	{ "type":[ "string" ], "required":false, "default":"http://t432z.com/2fa" },
		"AuthKey":  	 	{ "type":[ "string" ], "required":false, "default":"8181.2121" },
		"LineNo":       	{ "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *X2faType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	gCfg.ConnectToRedis()
	gCfg.ConnectToPostgreSQL()
	hdlr.gCfg = gCfg
	return
}

func (hdlr *X2faType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*X2faType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type X2faType struct {
	Next         http.Handler //
	Paths        []string     //
	RedisPrefix  string       //
	TemplatePath string       //	1stRedirct.tmpl
	QRPath       string       //
	QRURLPath    string       //
	DisplayURL   string       // URL of the 2fa Appliation
	TimeoutCodes int          // duration for a hash value stored in redis (2-min hash) - how long a "code" is valid for. In seconds.
	AuthKey      string       //
	Server2faURL string       // http://t432z.com

	LineNo int                         //
	gCfg   *cfg.ServerGlobalConfigType //
}

// NewX2faTypeServer will create a copy of the server for testing.
func NewX2faTypeServer(n http.Handler, p []string, redisPrefix, realm string) *X2faType {
	return &X2faType{
		Next:        n,
		Paths:       p,
		RedisPrefix: redisPrefix,
	}
}

type RedisData struct {
	Hash   string `json:"hash"`
	Fp     string `json:"fp"`
	T2faID string `json:"t_2fa_id"`
	UserID string `json:"user_id"`
	URL    string `json:"URL"`
}

type dispatchType struct {
	handlerFunc func(hdlr *X2faType, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string)
}

var dispatch map[string]dispatchType

func init() {
	dispatch = make(map[string]dispatchType)

	// xyzzy000 -need to concat base path, /api/x2fa on

	dispatch["/api/x2fa/test1"] = dispatchType{
		handlerFunc: func(hdlr *X2faType, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
			fmt.Printf("x2fa test1 called\n")
			fmt.Fprintf(os.Stderr, "x2fa test1 called\n")
			www.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(www, "<h1>x2fa test1 called</h1>\n")
		},
	}

	dispatch["/api/x2fa/getQRForSetup"] = dispatchType{
		// pull HTML/png - display when register. -- DIV with image+URL for QR, update qr on t432z.com
		// This is part of "registration" process - and should show up as a _immediate registration or as email-confim on registraiton
		// link.
		handlerFunc: getQRForSetup,
	}

	dispatch["/api/x2fa/pull-2-min-hash"] = dispatchType{
		// This is the hash that only lasts for 2 min - universal that is used in combination with
		// fingerprint and device-id (local-storage) to generate the 2fa 6 digit code.  Hash is generated if
		// not found in Redis - and has TTL in redis of 120.  Use (int)hdlr.TimeoutCodes for this.
		handlerFunc: get2MinHash,
	}
	dispatch["/api/x2fa/set-fp"] = dispatchType{
		// set the fingerpint for a particular user - Input Temporary Redis "ID" - use Redis to get user_id.
		handlerFunc: setFp,
	}
	dispatch["/api/x2fa/is-valid-2fa"] = dispatchType{
		// Return status=='success' if it is a valid 2fa - this will be disabled when not testing.  Requires a key to call.
		handlerFunc: IsValid2fa,
	}
	dispatch["/api/x2fa/gen-1-time-codes"] = dispatchType{
		// Return a database list of 1-time-codes for a user_id - will cause an Email to be sent to client.
		handlerFunc: gen1TimeCodes,
	}
	dispatch["/api/x2fa/n-1-time-codes"] = dispatchType{
		// Return JSON with count of # of 1 time codes left for user_id.
		handlerFunc: n1TimeCodes,
	}

}

func (hdlr *X2faType) UpdateQRMarkAsUsed(qrId string) error {
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

// err = hdlr.PullQRFromDB(rr.Tag)
func (hdlr *X2faType) PullQRFromDB(tag string) (qr_enc_id string, err error) {
	// Xyzzy - sould replace with stored proc. that updates state in same transaction.
	stmt := "select \"qr_enc_id\" from \"t_avail_qr\" where \"state\" = 'avail' limit 1"
	// insert into "t_avail_qr" ( "qr_id", "qr_enc_id", "url_path", "file_name", "qr_encoded_url_path" ) values
	// 	  ( '170', '4q', 'http://127.0.0.1:9019/qr/00170.4.png', './td_0008/q00170.4.png', 'http://t432z.com/q/4q' )
	rows, err := hdlr.gCfg.Pg_client.Db.Query(stmt)
	if err != nil {
		fmt.Printf("Database error %s, attempting to convert premis_id/animal_id to tag.\n", err)
		return "", err
	}
	godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
	for nr := 0; rows.Next(); nr++ {
		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
		if nr >= 1 {
			fmt.Printf("Error too many rows for a user, should be unique primary key\n")
			break
		}

		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
		var qr string
		err := rows.Scan(&qr)
		if err != nil {
			fmt.Printf("Error on d.b. query %s\n", err)
			return "", err
		}
		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")

		// Xyzzy - test fail to error report
		err = hdlr.UpdateQRMarkAsUsed(qr)
		if err != nil {
			fmt.Printf("Error on d.b. query %s\n", err)
			return "", err
		}

		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
		return qr, nil
	}
	return "", fmt.Errorf("Failed to get a QR code")
}

// ------------------------------------------------------------------------------------------------------------------------------------------
// DONE
// ------------------------------------------------------------------------------------------------------------------------------------------
// n1TimeCodes return JSON with count of # of 1 time codes left for user_id.
func n1TimeCodes(hdlr *X2faType, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("n1TimeCodes called\n")
	fmt.Fprintf(os.Stderr, "n1TimeCodes called\n")

	ps := &rw.Ps

	// Xyzzy - may need to use auth_token to convert from that to user_id.
	// Xyzzy - could use a sub-query inside select to do this from auth_token.

	user_id := ps.ByNameDflt("user_id", "")
	godebug.DbPfb(db1, "user_id: ->%s<-\n", user_id)

	stmt := "select count(1) as \"nOneTimeKeys\" from \"t_2fa_otk\" where \"user_id\" = $1 "

	Rows, err := hdlr.gCfg.Pg_client.Db.Query(stmt, user_id)
	if err != nil {
		fmt.Printf("Database error %s. stmt=%s\n", err, stmt)
		fmt.Fprintf(www, `{"status":"error","msg":"database error: [%v]"}`, err)
		return
	}

	defer Rows.Close()
	rowData, _, _ := sizlib.RowsToInterface(Rows)

	fmt.Fprintf(www, `{"status":"success","data":%s}`, godebug.SVarI(rowData))
	// fmt.Fprintf(www, `%s`, godebug.SVarI(rowData))
}

// ------------------------------------------------------------------------------------------------------------------------------------------
// DONE
// ------------------------------------------------------------------------------------------------------------------------------------------
/*
/api/2fa/getQRforSetup

display when register.
1. DIV with image+URL for QR
2. update qr on t432z.com
3. Set "ID" in redis with TTL of 1 hour.
*/
func getQRForSetup(hdlr *X2faType, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("getQRForSetup called\n")
	fmt.Fprintf(os.Stderr, "getQRForSetup called\n")

	ps := &rw.Ps

	user_id := ps.ByNameDflt("user_id", "")
	godebug.DbPfb(db1, "user_id: ->%s<-\n", user_id)

	// ----------------------------------------------------------------------------------------------------------------------------------------
	// ----------------------------------------------------------------------------------------------------------------------------------------
	// Cut Cut Cut
	// GenQR( user_id, TemplateDir, DisplayURL, RedisPrefix string, ConnToPg xyzzy400, ConnToRedis  xyzzy400)
	//		( html, ID, qrURL string, err error )
	// Also: 	qrId, QRImgUrl, err := hdlr.PullQRURLFromDB()
	//			key := fmt.Sprintf("%s%d", hdlr.RedisPrefix, ID)
	// ----------------------------------------------------------------------------------------------------------------------------------------
	// ----------------------------------------------------------------------------------------------------------------------------------------
	/*
		When we witch to using a template.

		<div class="getQRForSetup"><img src="{{.QRImgUrl}}"><div>Scan the QR code to setup your device or Enter {{.ID}} at <a href="/api/2fa/setup.html">/api/2fa/setup.html</a></div></div>"
	*/

	// ----------------------------------------------------------------------------------------------------------------------------------------
	// Generate ID
	ID := fmt.Sprintf("%d", rand.Intn(10000000)) // xyzzy201 - add in Checksum byte
	// xyzzy4141
	// RanHashBytes, err := UseRO.GenRandBytes(32)
	// Generate Random Hash
	RanHashBytes, err := GenRandBytes(32)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to generate random data.","LineFile":%q}`+"\n", err, godebug.LF()))
		fmt.Fprintf(www, `{"status":"error","msg":"Random Generation Failed","LineFile":%q}`, godebug.LF())
		return
	}
	RanHash := fmt.Sprintf("%x", RanHashBytes)
	// func GenRandNumber(nDigits int) (buf string, err error) {
	// func GenRandBytes(nRandBytes int) (buf []byte, err error) {

	// ----------------------------------------------------------------------------------------------------------------------------------------
	// get QR code from avail list
	qrId, QRImgUrl, err := hdlr.PullQRURLFromDB()
	godebug.DbPfb(db1, "%(Green) URL path: %s AT: %(LF)\n", QRImgUrl)

	// ----------------------------------------------------------------------------------------------------------------------------------------
	// update t432z.com URL shorter for this QR
	ran := fmt.Sprintf("%d", rand.Intn(1000000000))
	godebug.DbPfb(db1, "%(Cyan)AT: %(LF) ran [%v]\n", ran)

	theData := `{"data":"data written to system in file"}`
	// a432z.com - URL from config???
	status, body := DoGet("http://t432z.com/upd/", "url", hdlr.DisplayURL, "id", qrId, "data", theData, "_ran_", ran)
	if status != 200 {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to set QR Redirect","LineFile":%q}`+"\n", err, godebug.LF()))
		fmt.Fprintf(www, `{"status":"error","msg":"Unable to update QR code destination.","LineFile":%q}`, godebug.LF())
		return
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
		fmt.Fprintf(www, `{"status":"error","msg":"Failed to connect to Redis.","LineFile":%q}`, godebug.LF())
		return
	}

	key := fmt.Sprintf("%s%s", hdlr.RedisPrefix, ID)
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
	ttl := timeOutConst // 60 * 60 // 1 hour

	err = conn.Cmd("SETEX", key, ttl, val).Err
	if err != nil {
		if db4 {
			fmt.Printf("Error on redis - user not found - invalid relm - bad prefix - get(%s): %s\n", key, err)
		}
		fmt.Fprintf(www, `{"status":"error","msg":"Unable to set value in Redis.","LineFile":%q}`, godebug.LF())
		return
	}

	// ----------------------------------------------------------------------------------------------------------------------------------------
	// Setup OTK - 20 values for OTKs
	for i := 0; i < 20; i++ {
		rv, err := GenRandNumber(6)
		if err != nil {
			logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to generate random value.","LineFile":%q}`+"\n", err, godebug.LF()))
			fmt.Fprintf(www, `{"status":"error","LineFile":%q}`, godebug.LF())
			return
		}

		stmt := "insert into \"t_2fa_otk\" ( \"user_id\", \"one_time_key\" ) values ( $1, $2 )"
		_, err = hdlr.gCfg.Pg_client.Db.Exec(stmt, user_id, rv)
		if err != nil {
			logrus.Warn(fmt.Sprintf(`{"msg":"Error %s PG error.","LineFile":%q}`+"\n", err, godebug.LF()))
			fmt.Fprintf(www, `{"status":"error","LineFile":%q}`, godebug.LF())
			return
		}
	}

	// ----------------------------------------------------------------------------------------------------------------------------------------
	// Insert random hash -> PG to t_user
	stmt := "insert into \"t_2fa\" ( \"id\", \"user_id\", \"user_hash\" ) values ( $1, $2, $3 )"
	_, err = hdlr.gCfg.Pg_client.Db.Exec(stmt, t_2fa_ID, user_id, RanHash)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s PG error.","LineFile":%q}`+"\n", err, godebug.LF()))
		fmt.Fprintf(www, `{"status":"error","LineFile":%q}`, godebug.LF())
		return
	}

	// ----------------------------------------------------------------------------------------------------------------------------------------
	// Send back results.

	www.Header().Set("Content-Type", "text/html; charset=utf-8")

	buf := fmt.Sprintf(
		`<div class="getQRForSetup">
			<img src=%q>
			<div>
				Scan the QR code above to setup your mobile device or browse<br>
				on your mobile device to <a href="%s/msetup.html?id=%v">%s/msetup.html</a><br>
				and enter %v.
			</div>
		</div>`, QRImgUrl, hdlr.Server2faURL, ID, hdlr.Server2faURL, ID)

	fmt.Fprintf(www, buf)
}

/*
-- list of setup and validated devices

CREATE TABLE "t_2fa" (
	  "id"					uuid DEFAULT uuid_generate_v4() not null primary key
	, "user_id"				char varying (40)
	, "user_hash"			text
	, "fp"					text
	, "updated" 			timestamp
	, "created" 			timestamp default current_timestamp not null
);

m4_updTrig(t_2fa)

-- list of user one time keys

CREATE TABLE "t_2fa_otk" (
	  "id"					uuid DEFAULT uuid_generate_v4() not null primary key
	, "user_id"				char varying (40)
	, "one_time_key"		text
	, "updated" 			timestamp
	, "created" 			timestamp default current_timestamp not null
);

m4_updTrig(t_2fa_otk)
*/

// ------------------------------------------------------------------------------------------------------------------------------------------
// DONE
// ------------------------------------------------------------------------------------------------------------------------------------------
// set the fingerpint for a particular user - Input Temporary Redis "ID" - use Redis to get user_id.
//
// Return URL, hash etc.
// xyzzy902
//
func setFp(hdlr *X2faType, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("setFp called\n")
	fmt.Fprintf(os.Stderr, "setFp called\n")

	ps := &rw.Ps

	id := ps.ByNameDflt("id", "")
	godebug.DbPfb(db1, "id: ->%s<-\n", id)

	fp := ps.ByNameDflt("fp", "")
	godebug.DbPfb(db1, "fp: ->%s<-\n", fp)

	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		fmt.Fprintf(www, `{"status":"error","LineFile":%q}`, godebug.LF())
		return
	}

	key := fmt.Sprintf("%s%s", hdlr.RedisPrefix, id)
	// val := fmt.Sprintf("{\"hash\":%q, \"t_2fa_id\":%q, \"fp\":%q, \"user_id\":%q}", RanHash, t_2fa_ID, "fingerprint-not-set-yet", user_id)
	v, err := conn.Cmd("GET", key).Str()
	if err != nil {
		if db4 {
			fmt.Printf("Error on redis - user not found - invalid relm - bad prefix - get(%s): %s\n", key, err)
		}
		fmt.Fprintf(www, `{"status":"error","LineFile":%q}`, godebug.LF())
		return
	}
	var rr RedisData
	err = json.Unmarshal([]byte(v), &rr)
	if rr.Fp == "fingerprint-not-set-yet" {
		rr.Fp = fp
	} else {
		fmt.Fprintf(www, `{"status":"error","LineFile":%q}`, godebug.LF())
		return
	}

	val := godebug.SVar(rr)
	ttl := timeOutConst // 60 * 60 // 1 hour

	err = conn.Cmd("SETEX", key, ttl, val).Err
	if err != nil {
		if db4 {
			fmt.Printf("Error on redis - get(%s): %s\n", key, err)
		}
		fmt.Fprintf(www, `{"status":"error","LineFile":%q}`, godebug.LF())
		return
	}

	stmt := "update \"t_2fa\" set \"fp\" = $1 where \"id\" = $2"
	_, err = hdlr.gCfg.Pg_client.Db.Exec(stmt, rr.Fp, rr.T2faID)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s PG error.","LineFile":%q}`+"\n", err, godebug.LF()))
		fmt.Fprintf(www, `{"status":"error","LineFile":%q}`, godebug.LF())
		return
	}

	// xyzzy - update t_user.acct_state = 'ok' - when 2fa setup.
	stmt = "update \"t_user\" set \"acct_state\" = 'ok' where \"id\" = $1"
	_, err = hdlr.gCfg.Pg_client.Db.Exec(stmt, rr.UserID)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s PG error.","LineFile":%q}`+"\n", err, godebug.LF()))
		fmt.Fprintf(www, `{"status":"error","LineFile":%q}`, godebug.LF())
		return
	}

	// xyzzy902
	fmt.Fprintf(www, `{"status":"success","hash":%q,"URL":%q}`, rr.Hash, rr.URL)
}

// err = hdlr.PullQRFromDB(rr.Tag)
func (hdlr *X2faType) PullQRURLFromDB() (qr_enc_id, qr_url string, err error) {
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

// ------------------------------------------------------------------------------------------------------------------------------------------
// DONE
// ------------------------------------------------------------------------------------------------------------------------------------------
// dispatch["/api/2fa/pull-2-min-hash"] = dispatchType{
// This is the hash that only lasts for 2 min - universal that is used in combination with
// fingerprint and device-id (local-storage) to generate the 2fa 6 digit code.  Hash is generated if
// not found in Redis - and has TTL in redis of 120.  Use (int)hdlr.TimeoutCodes for this.
//
// 2MinHash - return hash + TTL
//
// Xyzzy301 - This really should be on it's own channel for validate/update of hash values.
// Xyzzy301 - This really should have a go-routine with a time-loop that updates the 2Min hash on a time-loop.
//
func get2MinHash(hdlr *X2faType, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("get2MinHash called\n")
	fmt.Fprintf(os.Stderr, "get2MinHash called\n")

	// ------------------------------------------------------------------------------
	// Get Connection
	// ------------------------------------------------------------------------------
	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		fmt.Fprintf(www, `{"status":"error","LineFile":%q}`, godebug.LF())
		return
	}

	godebug.DbPfb(db1, "%(Cyan) (err may be nil) [%s] AT: %(LF)\n", err)

	RanHashBytes, ttl, _, err := GenRandBytesOracle()
	fmt.Fprintf(os.Stderr, "\n\nHashBytes [%x], AT:\n", RanHashBytes, godebug.LF())
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to generate random data.","LineFile":%q}`+"\n", err, godebug.LF()))
		fmt.Fprintf(www, `{"status":"error","LineFile":%q}`, godebug.LF())
		return
	}
	RanHash := fmt.Sprintf("%x", RanHashBytes)
	godebug.DbPfb(db1, "%(Cyan) RanHash=[%s] AT: %(LF)\n", RanHash)

	// ----------------------------------------------------------------------
	// SUCCESS return
	// ----------------------------------------------------------------------
	// ttl is wrong! xyzzy
	fmt.Fprintf(www, "{\"hash\":%q,\"ttl\":%d,\"status\":\"success\"}", RanHash, ttl)
	return

}

// Return 200 if it is a valid 2fa - this will be disabled when not testing.
func IsValid2fa(hdlr *X2faType, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("IsValid2fa called\n")
	fmt.Fprintf(os.Stderr, "IsValid2fa called\n")

	ps := &rw.Ps

	Auth := ps.ByNameDflt("auth_key", "")
	godebug.DbPfb(db1, "auth_key: ->%s<- compare to hdlr.AuthKey ->%s<-\n", Auth, hdlr.AuthKey)

	val2fa := ps.ByNameDflt("val2fa", "")
	godebug.DbPfb(db1, "val2fa: ->%s<-\n", val2fa)

	user_id := ps.ByNameDflt("user_id", "")
	godebug.DbPfb(db1, "user_id: ->%s<-\n", user_id)

	auth_token := ps.ByNameDflt("auth_token", "")
	godebug.DbPfb(db1, "auth_token: ->%s<-\n", auth_token)

	// only run if hdlr.AuthKey is set to same as "auth_key". for this call.
	if hdlr.AuthKey != "" && Auth != hdlr.AuthKey {
		fmt.Fprintf(www, `{"status":"error","msg":"invalid auth_key","LineFile":%q}`, godebug.LF())
		return
	}

	var err error

	if auth_token != "" && user_id == "" {
		user_id, err = hdlr.GetUserIDFromAuthToken(auth_token)
		if err != nil {
			fmt.Fprintf(www, `{"status":"error","msg":"Unable to convert auth_token[%s] to user_id Error: %s","LineFile":%q}`, auth_token, err, godebug.LF())
			return
		}
	}

	// generate local copy based on user_id/auth_token - for all rows in t_2fa and any values in t_2fa_otk
	LocalVal2fa, err := hdlr.GetValidList(user_id)
	if err != nil {
		fmt.Fprintf(www, `{"status":"error","msg":"PG Database Error: %s","LineFile":%q}`, err, godebug.LF())
		return
	}
	godebug.DbPfb(db1, "%(Cyan) Local Values Array = %s AT: %(LF)\n", godebug.SVarI(LocalVal2fa))

	for _, v := range LocalVal2fa {
		if v == val2fa {
			stmt := "delete from \"t_2fa_otk\" where \"user_id\" = $1 and \"one_time_key\" = $2"
			_, err := hdlr.gCfg.Pg_client.Db.Query(stmt, user_id, v)
			if err != nil {
				fmt.Printf("Database error %s, attempting to convert premis_id/animal_id to tag.\n", err)
			}
			fmt.Fprintf(www, `{"status":"success"}`)
			return
		}
	}

	fmt.Fprintf(www, `{"status":"error","msg":"Two Factor Did Not Match","LineFile":%q}`, godebug.LF())
}

func (hdlr *X2faType) Get2MinHashFunc() (hash string, ttlLeft int, err error) {
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

// ------------------------------------------------------------------------------------------------------------------------
// ------------------------------------------------------------------------------------------------------------------------
// GetValidList get list of convened to string int values for valid 2fa
func (hdlr *X2faType) GetValidList(user_id string) (list []string, err error) {

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
func (hdlr *X2faType) GetPerDeviceHash(user_id, t2faId string) (user_hash, fp string, err error) {
	//	user_hash, err := hdlr.GetPerDeviceHash(user_id, t2faId)

	stmt := `select "user_hash", "fp" from "t_2fa" where "id" = $1 and "user_id" = $2`
	rows, err := hdlr.gCfg.Pg_client.Db.Query(stmt, t2faId, user_id)
	if err != nil {
		fmt.Printf("Database error %s, attempting to convert premis_id/animal_id to tag.\n", err)
		return
	}
	defer rows.Close()
	for nr := 0; rows.Next(); nr++ {

		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
		err = rows.Scan(&user_hash, &fp)
		if err != nil {
			return "", "", fmt.Errorf("Error [%s] user_id=[%s] t_2fa_id=[%s].", err, user_id, t2faId)
		}
		return
	}
	return "", "", fmt.Errorf("Failed to get user_hash/fp from t_tfa for user_id=[%s] t_2fa_id=[%s].", user_id, t2faId)
}

// ------------------------------------------------------------------------------------------------------------------------
// ------------------------------------------------------------------------------------------------------------------------
func (hdlr *X2faType) GetUserIDFromAuthToken(auth_token string) (user_id string, err error) {
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

// Return a database list of 1-time-codes for a user_id - will cause an Email to be sent to client.
func gen1TimeCodes(hdlr *X2faType, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("gen1TimeCodes called\n")
	fmt.Fprintf(os.Stderr, "gen1TimeCodes called\n")

	ps := &rw.Ps

	auth_key := ps.ByNameDflt("auth_key", "")
	godebug.DbPfb(db1, "auth_key: ->%s<- compare to hdlr.AuthKey ->%s<-\n", auth_key, hdlr.AuthKey)

	user_id := ps.ByNameDflt("user_id", "")
	godebug.DbPfb(db1, "user_id: ->%s<-\n", user_id)

	t2faId := ps.ByNameDflt("t2faId", "")
	godebug.DbPfb(db1, "t2faId: ->%s<-\n", t2faId)

	auth_token := ps.ByNameDflt("auth_token", "")
	godebug.DbPfb(db1, "auth_token: ->%s<-\n", auth_token)

	// only run if hdlr.AuthKey is set to same as "auth_key". for this call.
	if hdlr.AuthKey != "" && auth_key != hdlr.AuthKey {
		fmt.Fprintf(www, `{"status":"error","msg":"invalid auth_key","LineFile":%q,"auth_key":%q}`, godebug.LF(), auth_key)
		return
	}

	var err error

	if auth_token != "" && user_id == "" {
		user_id, err = hdlr.GetUserIDFromAuthToken(auth_token)
		if err != nil {
			fmt.Fprintf(www, `{"status":"error","msg":"Unable to convert auth_token[%s] to user_id Error: %s","LineFile":%q}`, auth_token, err, godebug.LF())
			return
		}
	}

	current2MinHash, ttlLeft, err := hdlr.Get2MinHashFunc()
	if err != nil {
		fmt.Fprintf(www, `{"status":"error","msg":"Unable to get 2Min hash Error: %s","LineFile":%q}`, err, godebug.LF())
		return
	}

	user_hash, fp, err := hdlr.GetPerDeviceHash(user_id, t2faId)
	if err != nil {
		fmt.Fprintf(www, `{"status":"error","msg":"Unable to get per-device hash Error: %s","LineFile":%q}`, err, godebug.LF())
		return
	}

	// xyzzy - convert to a function for this.
	val0 := HashStrings.Sha256(fmt.Sprintf("%s:%s:%s", user_hash, fp, current2MinHash))
	// val1 := fmt.Sprintf("%x", val0)
	val1 := string(val0)
	val2 := val1[len(val1)-6:]
	val, err := strconv.ParseInt(val2, 16, 64)
	if err != nil {
		fmt.Fprintf(www, `{"status":"error","msg":"Failed to parse[%s] as base 16 int Error: %s","LineFile":%q}`, val2, err, godebug.LF())
		return
	}
	val = val % 1000000
	val_s := fmt.Sprintf("%06d", val) // list = append(list, fmt.Sprintf("%06d", val))

	fmt.Fprintf(www, `{"status":"success","val":%q,"ttl":%d}`, val_s, ttlLeft)
}

func (hdlr *X2faType) ServeHTTP(www http.ResponseWriter, req *http.Request) {

	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			www.Header().Set("Content-Type", "application/json")

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "X2fa", hdlr.Paths, pn, req.URL.Path)

			ps := &rw.Ps
			data := ps.ByNameDflt("Data", "{}")
			var mdata map[string]string
			err := json.Unmarshal([]byte(data), &mdata)
			if err != nil {
				fmt.Fprintf(www, "{\"status\":\"error\",\"msg\":%q}", err)
				return
			}

			godebug.DbPfb(db1, "%(Green) (err may be nil) [%s] AT: %(LF)\n", err)

			fx, ok := dispatch[req.URL.Path]
			if !ok {
				godebug.DbPfb(db1, "%(Red)Error Path Invalid [%s] AT: %(LF)\n", req.URL.Path)

				fmt.Fprintf(www, `{"status":"not-implemented-yet","data":%q,"LineFile":%q}`, req.URL.Path, godebug.LF())
				return
			}
			fx.handlerFunc(hdlr, rw, www, req, mdata)
			return

			fmt.Fprintf(www, `{"status":"not-implemented-yet","LineFile":%q}`, godebug.LF())
		}
	}

	hdlr.Next.ServeHTTP(www, req)
}

// Modified to send Header!
/*
---------------------------------------------
// Xyzzy101 - Setup QR Redirect
---------------------------------------------

	export QR_SHORT_AUTH_TOKEN="w4h0wvtb1zk4uf8Xv.Ns9Q7j8"
	wget -o out/,list1 -O out/,list2 \
		--header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" \
		"http://t432z.com/upd/?url=http://test.test.com&id=5c"

	-- 1. DoGet - change to create a header
	-- 2. Example Call to set this
*/
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

const timeOutConst = (60 * 60 * 24) + 5
const db1 = true
const db4 = true

/* vim: set noai ts=4 sw=4: */
