package TabServer2

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/json" //	"encoding/json"
	"github.com/pschlump/uuid"
)

type RedirectToData struct {
	Status string `json:"status"`
	UserID string `json:"user_id"`
	Use2fa string `json:"use_2fa"`
}

func X2faStash(www http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rrv string, rexit bool, rstatus int) {
	// 	1. Take 'rv' and pull out of it
	// 		1.a. the "username"
	//	2. Stash 'rv' in x2fa:${UN}:${auth_tok_2part} // ttl=4min
	//  3. Lookfor number of x2fa:${UN} - int, if found increment, else set to 1 // ttl=4min (update ttl to 4min)
	//  4. Reuturn # from x2fa:{UN} // ttl=4min
	//  5. { "auth_tok_2part": ${auth_tok_2part}, "nt": ${n} }
	fmt.Printf("%sAT:%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())
	fmt.Fprintf(os.Stderr, "\n\n%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())
	fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorRed, MiscLib.ColorReset, rv, godebug.LF())
	fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n\n\n", MiscLib.ColorGreen, MiscLib.ColorReset, rv, godebug.LF())

	// func SignToken(tokData []byte, keyFile string) (out string, err error) {
	//	hdlr.KeyFilePrivate        string                      // private key file for signing JWT tokens
	// https://github.com/dgrijalva/jwt-go.git
	/*
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
	*/

	var ed RedirectToData
	var all map[string]interface{}

	err := json.Unmarshal([]byte(rv), &ed)
	if err != nil {
		return rv, false, 200
	}
	err = json.Unmarshal([]byte(rv), &all)
	if err != nil {
		return rv, false, 200
	}

	if ed.Status == "success" && ed.Use2fa == "yes" {

		id, err := uuid.NewV4()
		if err != nil {
			fmt.Fprintf(www, `{"status":"failed","msg":"Error [%s]","LineFile":%q}`, err, godebug.LF())
			return "{\"status\":\"failed\"}", true, 200
		}
		auth_tok_2part := id.String()
		_ = auth_tok_2part

		conn, err := hdlr.gCfg.RedisPool.Get()
		defer hdlr.gCfg.RedisPool.Put(conn)
		if err != nil {
			rv = fmt.Sprintf(`{"status":"failed","msg":"Error [%s]","LineFile":%q}`, err, godebug.LF())
			return rv, true, 200
		}

		all["auth_tok_2part"] = auth_tok_2part
		rv = godebug.SVarI(all)

		// key := fmt.Sprintf("2faStash:%s:%s", ed.UserID, auth_tok_2part)
		key := fmt.Sprintf("2faStash:User:%s", auth_tok_2part) // xyzzy - 2faStash: to be a "key" in config
		err = conn.Cmd("SETEX", key, 60*4, rv).Err
		if err != nil {
			rv = fmt.Sprintf(`{"status":"failed","msg":"Error [%s]","LineFile":%q}`, err, godebug.LF())
			return rv, true, 200
		}

		key = fmt.Sprintf("2faStash:%s", ed.UserID)
		val, err := conn.Cmd("GET", key).Str()
		if err != nil || val == "" {
			err = conn.Cmd("SETEX", key, 60*4, "1").Err
			if err != nil {
				rv = fmt.Sprintf(`{"status":"failed","msg":"Error [%s]","LineFile":%q}`, err, godebug.LF())
				return rv, true, 200
			}
			val = "0"
		} else {
			err = conn.Cmd("TTL", key, 60*4).Err
			if err != nil {
				rv = fmt.Sprintf(`{"status":"failed","msg":"Error [%s]","LineFile":%q}`, err, godebug.LF())
				return rv, true, 200
			}
			err = conn.Cmd("INCR", key).Err
			if err != nil {
				rv = fmt.Sprintf(`{"status":"failed","msg":"Error [%s]","LineFile":%q}`, err, godebug.LF())
				return rv, true, 200
			}
		}
		nVal, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			nVal = 0
		}
		nVal++

		rv = fmt.Sprintf(`{"status":"success","auth_tok_2part":%q,"nVal":%d}`, auth_tok_2part, nVal)
		return rv, true, 200

	}

	rv = fmt.Sprintf(`{"status":"failed","msg":"Error [%s]","LineFile":%q}`, "login not successful", godebug.LF())
	return rv, true, 200

}

// Part1of2 is X2faStash
func X2faSetupPt2of2(www http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (string, bool, int) {

	fmt.Printf("%sAT:%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())
	fmt.Fprintf(os.Stderr, "\n\n%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())
	fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorRed, MiscLib.ColorReset, rv, godebug.LF())
	fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n\n\n", MiscLib.ColorGreen, MiscLib.ColorReset, rv, godebug.LF())

	// func SignToken(tokData []byte, keyFile string) (out string, err error) {
	//	hdlr.KeyFilePrivate        string                      // private key file for signing JWT tokens
	// https://github.com/dgrijalva/jwt-go.git
	/*
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
	*/

	// pull back data from Redis - xyzzy - TODO
	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		rv = fmt.Sprintf(`{"status":"failed","msg":"Error [%s]","LineFile":%q}`, err, godebug.LF())
		return rv, true, 200
	}

	// get the key to pull back the data.
	auth_tok_2part := ps.ByNameDflt("auth_tok_2part", "")

	key := fmt.Sprintf("2faStash:User:%s", auth_tok_2part)
	rv, err = conn.Cmd("GET", key).Str()
	if err != nil {
		rv = fmt.Sprintf(`{"status":"failed","msg":"Login has timed out - please try again.","LineFile":%q}`, godebug.LF())
		return rv, true, 200
	}

	// ---------------------------------------------------------------------------------------------------------
	// Save - pull back # of login attempts - not relativant to this code but will be used in the 2fa stuff.
	// ---------------------------------------------------------------------------------------------------------
	// key = fmt.Sprintf("2faStash:%s", ed.UserID)
	// val, err := conn.Cmd("GET", key).Str()
	// if err != nil {
	// 	rv = fmt.Sprintf(`{"status":"failed","msg":"Error [%s]","LineFile":%q}`, err, godebug.LF())
	// 	return rv, true, 200
	// }
	// convert key to value!

	var ed RedirectToData
	var all map[string]interface{}

	err = json.Unmarshal([]byte(rv), &ed)
	if err != nil {
		return rv, false, 200
	}
	err = json.Unmarshal([]byte(rv), &all)
	if err != nil {
		return rv, false, 200
	}

	if ed.Status == "success" && ed.Use2fa == "yes" {

		fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())

		//xyzzy
		// xyzzy *
		// 	xyzzy ***
		// 	xyzzy ***
		// xyzzy *
		//xyzzy
		html, QRImgURL, ID, err := GetQRForSetup(hdlr, www, req, ps, ed.UserID)
		if err != nil {
			fmt.Fprintf(www, `{"status":"failed","msg":"Error [%s]","LineFile":%q}`, err, godebug.LF())
			return "{\"status\":\"failed\"}", true, 200 // xyzzy - better error return
		}

		all["html_2fa"] = html
		all["QRImgURL"] = QRImgURL
		all["X2fa_Temp_ID"] = ID

		delete(all, "user_id")

		rv = godebug.SVar(all)
		fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())
		return rv, true, 200
	}

	return rv, false, 200
}
