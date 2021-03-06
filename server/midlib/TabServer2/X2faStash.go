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
	Msg    string `json:"msg"`
}

/*
(rvOut string, pptFlag PrePostFlagType, exit bool, a_status int) {
		return "", PrePostFatalSetStatus, true, 500
		return rv, PrePostFatalSetStatus, true, 500
	return rv, PrePostContinue, exit, a_status
*/

func X2faStash(www http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rvOut string, pptFlag PrePostFlagType, exit bool, a_status int) {
	// 	1. Take 'rv' and pull out of it
	// 		1.a. the "username"
	//	2. Stash 'rv' in x2fa:${UN}:${auth_tok_2part} // ttl=4min
	//  3. Lookfor number of x2fa:${UN} - int, if found increment, else set to 1 // ttl=4min (update ttl to 4min)
	//  4. Reuturn # from x2fa:{UN} // ttl=4min
	//  5. { "auth_tok_2part": ${auth_tok_2part}, "nt": ${n} }
	fmt.Printf("%sAT:%s at top rv = -->>%s<<-- %s\n", MiscLib.ColorBlue, MiscLib.ColorReset, rv, godebug.LF())
	fmt.Fprintf(os.Stderr, "\n%s ++++++++++++++++++++++++++++++++ X2faStash ++++++++++++++++++++++++++++++++ %s\n", MiscLib.ColorRed, MiscLib.ColorReset)
	fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at top rv = -->>%s<<-- %s\n\n", MiscLib.ColorRed, MiscLib.ColorReset, rv, godebug.LF())

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
		return "", PrePostFatalSetStatus, true, 500
	}
	err = json.Unmarshal([]byte(rv), &all)
	if err != nil {
		return "", PrePostFatalSetStatus, true, 500
	}

	fmt.Printf("raw=%s ed=%s at:%s\n", rv, godebug.SVar(ed), godebug.LF())

	if ed.Status == "success" && ed.Use2fa == "yes" {

		id, err := uuid.NewV4()
		if err != nil {
			fmt.Fprintf(www, `{"status":"failed","msg":"Error [%s]","LineFile":%q}`, err, godebug.LF())
			return "", PrePostFatalSetStatus, true, 500
		}
		auth_tok_2part := id.String()
		_ = auth_tok_2part

		conn, err := hdlr.gCfg.RedisPool.Get()
		defer hdlr.gCfg.RedisPool.Put(conn)
		if err != nil {
			rv = fmt.Sprintf(`{"status":"failed","msg":"Error [%s]","LineFile":%q}`, err, godebug.LF())
			return "", PrePostFatalSetStatus, true, 500
		}

		all["auth_tok_2part"] = auth_tok_2part
		rv = godebug.SVarI(all)

		key := fmt.Sprintf("%sStash:User:%s", hdlr.RedisPrefix2fa, auth_tok_2part)
		fmt.Fprintf(os.Stderr, "KEY(1): %s\n", key)
		fmt.Fprintf(os.Stdout, "KEY(1): %s for UserID=%s\n", key, ed.UserID)
		err = conn.Cmd("SETEX", key, 60*4, rv).Err
		if err != nil {
			rv = fmt.Sprintf(`{"status":"failed","msg":"Error [%s]","LineFile":%q}`, err, godebug.LF())
			return "", PrePostFatalSetStatus, true, 500
		}

		key = fmt.Sprintf("%sStash:Count:%s", hdlr.RedisPrefix2fa, ed.UserID)
		fmt.Fprintf(os.Stderr, "KEY(2): %s containing user id.\n", key)
		fmt.Fprintf(os.Stdout, "KEY(2): %s containing user id.\n", key)
		val, err := conn.Cmd("GET", key).Str()
		if err != nil || val == "" {
			err = conn.Cmd("SETEX", key, 60*4, "1").Err
			if err != nil {
				rv = fmt.Sprintf(`{"status":"failed","msg":"Error [%s]","LineFile":%q}`, err, godebug.LF())
				return "", PrePostFatalSetStatus, true, 500
			}
			val = "0"
		} else {
			err = conn.Cmd("EXPIRE", key, 60*4).Err
			if err != nil {
				rv = fmt.Sprintf(`{"status":"failed","msg":"Error [%s]","LineFile":%q}`, err, godebug.LF())
				return "", PrePostFatalSetStatus, true, 500
			}
			err = conn.Cmd("INCR", key).Err
			if err != nil {
				rv = fmt.Sprintf(`{"status":"failed","msg":"Error [%s]","LineFile":%q}`, err, godebug.LF())
				return "", PrePostFatalSetStatus, true, 500
			}
		}
		nVal, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			nVal = 0
		}
		nVal++

		rv = fmt.Sprintf(`{"status":"success","auth_tok_2part":%q,"nVal":%d}`, auth_tok_2part, nVal)
		// return rv, PrePostContinue, false, 200
		fmt.Fprintf(os.Stderr, "%s **** AT **** :%s at bottom rv = -->>%s<<-- %s\n\n", MiscLib.ColorRed, MiscLib.ColorReset, rv, godebug.LF())
		return rv, PrePostSuccessWriteRV, false, 200

	}

	rv = fmt.Sprintf(`{"status":"failed","msg":"Error [%s]","LineFile":%q}`, ed.Msg, godebug.LF())
	fmt.Fprintf(www, `{"status":"failed","msg":"Error [%s]","LineFile":%q}`, ed.Msg, godebug.LF())
	return "", PrePostFatalSetStatus, true, 200

}

// Part1of2 is X2faStash
func X2faSetupPt2of2(www http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rvOut string, pptFlag PrePostFlagType, exit bool, a_status int) {

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
		return "", PrePostFatalSetStatus, true, 500
	}

	// get the key to pull back the data.
	auth_tok_2part := ps.ByNameDflt("auth_tok_2part", "")

	key := fmt.Sprintf("%sStash:User:%s", hdlr.RedisPrefix2fa, auth_tok_2part)
	rv, err = conn.Cmd("GET", key).Str()
	if err != nil || rv == "" {
		rv = fmt.Sprintf(`{"status":"failed","msg":"Login has timed out - please try again.","LineFile":%q}`, godebug.LF())
		return "", PrePostFatalSetStatus, true, 500
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
		return "", PrePostFatalSetStatus, true, 500
	}
	err = json.Unmarshal([]byte(rv), &all)
	if err != nil {
		return "", PrePostFatalSetStatus, true, 500
	}

	if ed.Status == "success" && ed.Use2fa == "yes" {

		rv = godebug.SVar(all)
		godebug.DbPfb(db1x2fa, "%(Green) SHOULD BE SUCCESS, Pull back from Redis, rv = %s AT: %(LF), Parent = %s, p2 = %s\n", rv, godebug.LF(2), godebug.LF(3))
		return rv, PrePostContinue, false, 200
	}

	return "", PrePostFatalSetStatus, true, 500
}
