package TabServer2

//
// R E S T s e r v e r - Server Component	(TabServer2)
//
// Copyright (C) Philip Schlump, 2012-2017 -- All rights reserved.
//
// Do not remove the following lines - used in auto-update.
// Version: 1.1.0
// BuildNo: 0391
// FileId: 0005
// File: TabServer2/crud.go
//

// xyzzy-JWT

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	logrus "github.com/pschlump/pslog" // "github.com/sirupsen/logrus"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/Go-FTL/server/tmplp"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/json" //	"encoding/json"
)

// Exists reports whether the named file or directory exists.
//func Exists(name string) bool {
//    if _, err := os.Stat(name); err != nil {
//		if os.IsNotExist(err) {
//			return false
//		   }
//		}
//	return true
//}

// ==============================================================================================================================================================================
// Redis Keys
// Do we need a "prefix" or a "table" type thing to limit access to a set of keys?
// ==============================================================================================================================================================================
func (hdlr *TabServer2Type) RespHandlerRedis(res http.ResponseWriter, req *http.Request, cfgTag string, ps *goftlmux.Params, rw *goftlmux.MidBuffer) {

	rw, _ /*top_hdlr*/, ps, err := GetRwPs(res, req)

	h := hdlr.SQLCfg[cfgTag] // get configuration						// Xyzzy - what if config not defined for this item at this point!!!!
	var rv string = ""
	var done bool = false
	var isError bool = false
	mdata := make(map[string]string)
	cookieList := make(map[string]string)

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	defer hdlr.gCfg.RedisPool.Put(conn)

	// Setting Heders
	res.Header().Set("Content-Type", "application/json") // set default reutnr type

	// Get Initial Data
	// m, fr := sizlib.UriToStringMap(req) // pull out the ?name=value params				// xyzzy - add 'fr'
	// tr.TraceUri(req, m)

	trx := mid.GetTrx(rw)
	trx.SetFrom(fmt.Sprintf("sql-cfg.json[%s]Line#:%s", cfgTag, h.LineNo))
	trx.SetFunc(1)
	// trx.TraceUriPs(req, ps)
	TraceUriPs(trx, req, ps)
	// trx.SetRedisUsed ( key )

	// InjectData(m, fr, h, res, req) // xyzzy - add 'fr'
	InjectDataPs(ps, h, res, req) // xyzzy - add 'fr'

	for _, kk := range h.P { // Not real clear to me why you would ever cobine CRUD and p:[] values
		mdata[kk] = ps.ByName(kk)
		if !(*ps).HasName(kk) {
			trx.AddNote(2, fmt.Sprintf("Warning(12050): Missing data for %s, Empty string used!", kk))
		}
	}

	key_tmpl := h.Query
	goftlmux.AddValueToParams("key", key_tmpl, 'i', goftlmux.FromOther, ps)
	mdata["key"] = key_tmpl

	// trx.SetData(m, fr)
	SetDataPs(trx, ps)

	RedisReferenced(godebug.FUNCNAME(), cfgTag, hdlr)

	// Validate Query Params
	trx.AddNote(1, "Validate Query Parameters")
	err = ValidateQueryParams(ps, h, req) // Validate them
	if err != nil {
		// trx.AddNote(1, "Validate Query Parameters Failed:"+fmt.Sprintf("%v", err))
		// rv = fmt.Sprintf(`{"status":"error","code":"3","msg":"Error(12048): Invalid Query Parameters (%s) sql-cfg.json[%s]",%s}`, err, cfgTag, godebug.LFj())
		done = true
		isError = true
		// trx.ErrorReturn ( 1, rv )
		ReturnErrorMessage(406, "Invalid Parameter", "12048",
			fmt.Sprintf(`Error(12048): Invalid Query Parameters (%s) sql-cfg.json[%s] %s`, sizlib.EscapeError(err), cfgTag, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
		rv = ""
	}

	// Authenticate User if Necessary
	if !done {
		if h.LoginRequired {
			trx.AddNote(1, "User Authentication is Requried")
			err = hdlr.ValidateUserTrx(h, trx, res, req, ps, rw)
			if err != nil {
				done = true
				isError = true
				ReturnErrorMessage(401, "Authentication Failed, Invalid API Key", "1", fmt.Sprintf(`Error(17004): Invalid API Key %s`, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
				rv = ""
			} else {
				trx.SetUserInfo(ps.ByName("username"), ps.ByName("$user_id$"), ps.ByName("auth_token"))
			}
		}
	}

	// CallBefore
	exit := false
	a_status := 200
	pptCode := PrePostContinue
	if !done {
		if len(h.CallBefore) > 0 {
			trx.AddNote(1, "Functions to call before running queries.  CallBefore is set.")
			for _, fx_name := range h.CallBefore {
				if !exit {
					trx.AddNote(1, fmt.Sprintf("CallBefore[%s]", fx_name))
					rv, pptCode, exit, a_status = hdlr.CallFunction("before", fx_name, res, req, cfgTag, rv, isError, cookieList, ps, trx)
				}
			}
		}
	}
	switch pptCode {
	case PrePostRVUpdatedSuccess, PrePostRVUpdatedFail, PrePostFatalSetStatus:
		fmt.Printf("*** exit=%v pptCode=%v from before operations has been signaled ***, rv=%s, %s\n", exit, pptCode, rv, godebug.LF())
		done = true
		isError = true
		ReturnErrorMessageRv(a_status, rv, "Preprocessing signaled error", "18008",
			fmt.Sprintf(`Error(18008): Preprocessing signaled error. sql-cfg.json[%s] %s`, cfgTag, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
	}

	//	if ! done {
	//		for i, v := range m {
	//			mdata[i] = v[0]
	//		}
	//	}

	if !done {
		rv = `{"status":"success"}`
		done = true
		key := sizlib.Qt(mdata["key"], mdata) // template the key
		switch req.Method {
		case "GET": // Fetch It
			// rv, err = redis.String(rr.RedisDo("GET", key))
			rv, err := conn.Cmd("GET", key).Str() // Get the value
			if err != nil {
				rv = fmt.Sprintf(`{"status":"error","msg":"key not found","raw":%q}`, err)
				trx.SetQryDone("", rv)
				ReturnErrorMessage(406, "Error(12810): Key not found", "12810",
					fmt.Sprintf(`Error(12810): Key not found (%s) %s, %s`, err, key, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
				return
			}
		case "POST": // Set It
			// _, err = rr.RedisDo("SET",key,sizlib.SVar(mdata))				// xyzzy - seems like a bad idea, should tell what to save
			// _, err = rr.RedisDo("SET", key, sizlib.SVar(mdata)) // xyzzy - seems like a bad idea, should tell what to save
			err := conn.Cmd("SET", key, sizlib.SVar(mdata)).Err //
			if err != nil {
				rv = fmt.Sprintf(`{"status":"error","msg":"key not found","raw":%q}`, err)
				trx.SetQryDone("", rv)
				ReturnErrorMessage(406, "Error(12810): Key not found", "12810",
					fmt.Sprintf(`Error(12810): Key not found (%s) %s, %s`, err, key, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
				return
			}
		case "PUT": // Do Update - Get, Merge, Set
			// rv, err := redis.String(rr.RedisDo("GET", key))
			rv, err := conn.Cmd("GET", key).Str() // Get the value
			if err != nil {
				rv = fmt.Sprintf(`{"status":"error","msg":"key not found","raw":%q}`, err)
				trx.SetQryDone("", rv)
				ReturnErrorMessage(406, "Error(12810): Key not found", "12810",
					fmt.Sprintf(`Error(12810): Key not found (%s) %s, %s`, err, key, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
				return
			} else {
				var old_data map[string]string
				err = json.Unmarshal([]byte(rv), &old_data) // Convert the JSON column to values
				if err == nil {
					mdata = sizlib.ExtendDataS(old_data, mdata) // merge
				}
				//rr.RedisDo("SET", key, sizlib.SVar(mdata))
				conn.Cmd("SET", key, sizlib.SVar(mdata))
			}
		case "DELETE":
			// _, err = rr.RedisDo("DEL", key)
			err := conn.Cmd("DEL", key).Err // Get the value
			if err != nil {
				rv = fmt.Sprintf(`{"status":"error","msg":"key not found","raw":%q}`, err)
				trx.SetQryDone("", rv)
				ReturnErrorMessage(406, "Error(12810): Key not found", "12810",
					fmt.Sprintf(`Error(12810): Key not found (%s) %s, %s`, err, key, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
				return
			}
		case "HEAD":
			rv = key
		}
		trx.SetQryDone("", rv)
	}

	exit = false
	a_status = 200
	pptCode = PrePostContinue
	if len(h.CallAfter) > 0 {
		trx.AddNote(1, "CallAfter fore is True - functions will be called.")
		for _, fx_name := range h.CallAfter {
			trx.AddNote(1, fmt.Sprintf("CallAfter fore[%s]", fx_name))
			if !exit {
				rv, pptCode, exit, a_status = hdlr.CallFunction("after", fx_name, res, req, cfgTag, rv, isError, cookieList, ps, trx)
			}
		}
	}
	switch pptCode {
	case PrePostRVUpdatedSuccess, PrePostRVUpdatedFail, PrePostFatalSetStatus:
		fmt.Printf("*** exit=%v pptCode=%v from before operations has been signaled ***, rv=%s, %s\n", exit, pptCode, rv, godebug.LF())
		done = true
		isError = true
		ReturnErrorMessageRv(a_status, rv, "Postprocessing signaled error", "18007",
			fmt.Sprintf(`Error(18007): Postprocessing signaled error. sql-cfg.json[%s] %s`, cfgTag, godebug.LF()), res, req, *ps, trx, hdlr) // status:error
	}

	if !isError { // xyzzy-new Thu Mar  5 09:06:21 MST 2015
		trx.SetRvBody(rv)
		// io.WriteString(res, sizlib.JsonP(rv, res, req))
		io.WriteString(res, rv)
	}

}

var getTmplCache map[string][]byte
var getTmplLock sync.Mutex

func init() {
	getTmplCache = make(map[string][]byte)
}

func (hdlr *TabServer2Type) getTemplate(name string) string {
	getTmplLock.Lock()
	defer getTmplLock.Unlock()
	var tv []byte
	var ok bool
	if tv, ok = getTmplCache[name]; !ok || string(tv) == "" {
		fn := hdlr.EmailTemplateDir + name + ".tmpl"
		tv, err := ioutil.ReadFile(fn)
		if err != nil {
			fmt.Printf("Unable to open %s - email template file, err=%s, %s\n", fn, err, godebug.LF())
			fmt.Fprintf(os.Stderr, "%sUnable to open %s - email template file, err=%s, %s%s\n", MiscLib.ColorRed, fn, err, godebug.LF(), MiscLib.ColorReset)
		}
		getTmplCache[name] = tv
		return string(tv)
	}
	return string(tv)
	//	return `template={{.template}}
	//username={{.username}}
	//real_name={{.real_name}}
	//email_token={{.email_token}}
	//app={{.app}}
	//url={{.url}}
	//from={{.from}}
	//`
}

// s1, b1, s2, b2, err := TemplateEmail ( ed.Email["template"], ed )
func (hdlr *TabServer2Type) TemplateEmail(template_name string, mdata map[string]string) (s1, b1, b2 string, err error) {
	// s1, b1, s2, b2 = "s1", "b1", "s2", "b2"
	fmt.Printf("TemlateEmail mdata=%s\n", godebug.SVarI(mdata))
	s1 = tmplp.ExecuteATemplateByName(hdlr.getTemplate(template_name), "email_subject", mdata)
	b1 = tmplp.ExecuteATemplateByName(hdlr.getTemplate(template_name), "body_html", mdata)
	b2 = tmplp.ExecuteATemplateByName(hdlr.getTemplate(template_name), "body_text", mdata)
	return
}

//func init() {
//	fmt.Printf("init2 in main\n")
//}

// ==============================================================================================================================================================================
func GetSI(s string, data map[string]interface{}) string {
	if x, ok := data[s]; ok {
		return x.(string)
	}
	return ""
}

// ==============================================================================================================================================================================
/*
	Issue00079: NoLog parameters
	Issue00018: Actual CC Processing with account + fake cards																	4 hr
			2. NoLog parameters
			3. Put in actual test params with account - see that it works

Try some of these numbers:
	4000 0000 0000 0002
	4026 0000 0000 0002
	5018 0000 0009
	5100 0000 0000 0008
	6011 0000 0000 0004

*/
//func ChargeCreditCard(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (string, bool, int) {
//	var exit bool = false
//	var a_status int = 200
//	fmt.Printf("************************************* Charge Credit Card Called *************************************\n")
//
//	if MiscLib.InArray("credit_card_test_mode", hdlr.DebugFlags) >= 0 {
//		ps.SetValue("cc_authorize", "test-mode")
//		goftlmux.AddValueToParams("cc_authorize", "'test-mode'", 's', goftlmux.FromOther, &ps)
//		fmt.Printf("************************************* Credit Card in test-mode *************************************, rv=%s\n", rv)
//		rv = `{"status":"success"}`
//		return rv, exit, a_status
//	}
//
//	auth := AuthorizeNet.AuthorizeNet{
//		Login:     hdlr.AuthorizeNetLogin, // "<YourLogin>",
//		Key:       hdlr.AuthorizeNetKey,   // "<YourKey>",
//		DupWindow: 120,
//		TestMode:  true,
//	}
//
//	card := AuthorizeNet.CardInfoType{
//		CreditCardNumber: ps.ByNameDflt("credit_card", "x"),                                         // 5555 5555 5555 5555
//		CVV:              ps.ByNameDflt("ccv", "x"),                                                 // 123
//		Month_Year:       ps.ByNameDflt("exp_month", "1") + "/" + ps.ByNameDflt("exp_year", "2017"), // 01/2017
//		Method:           ps.ByNameDflt("card_type", "visa"),                                        // METHOD_VISA,
//	}
//
//	// func (ps *Params) ByNameDflt(name string, dflt string) (rv string) {
//	// Extend ""ps"" user info with info from t_user, p_cart
//
//	data := AuthorizeNet.AuthorizeNetType{
//		InvoiceNumber:   ps.ByNameDflt("invoice_no", "y"),                // "123444",		// need to create the invoice number in p_cart when "address" is added
//		Amount:          ps.ByNameDflt("ex_total", "0.0"),                // "5.56",
//		Description:     ps.ByNameDflt("customer_transaction_name", "y"), // "My Test transaction",
//		FirstName:       ps.ByNameDflt("b_first_name", "y"),              // Address for shipping/billing pulled from p_cart
//		LastName:        ps.ByNameDflt("b_last_name", "y"),               //
//		Company:         ps.ByNameDflt("b_company", "y"),                 //
//		BillingAddress:  ps.ByNameDflt("b_line_1", "y"),                  //
//		BillingCity:     ps.ByNameDflt("b_city", "y"),                    //
//		BillingState:    ps.ByNameDflt("b_state", "y"),                   //
//		BillingZip:      ps.ByNameDflt("b_postal_code", "y"),             //
//		BillingCountry:  ps.ByNameDflt("b_country", "y"),                 //
//		Phone:           ps.ByNameDflt("b_phone", "y"),                   //
//		Email:           ps.ByNameDflt("email", "y"),                     // From auth_token -> t_user.email
//		CustomerId:      ps.ByNameDflt("user_id", "y"),                   // From auth_token -> t_user.id
//		CustomerIp:      ps.ByNameDflt("$IP$", "y"),                      //
//		ShipToFirstName: ps.ByNameDflt("s_first_name", "y"),              //
//		ShipToLastName:  ps.ByNameDflt("s_last_name", "y"),               //
//		ShipToCompany:   ps.ByNameDflt("s_company", "y"),                 //
//		ShipToAddress:   ps.ByNameDflt("s_line_1", "y"),                  //
//		ShipToCity:      ps.ByNameDflt("s_city", "y"),                    //
//		ShipToState:     ps.ByNameDflt("s_state", "y"),                   //
//		ShipToZip:       ps.ByNameDflt("s_postal_code", "y"),             //
//		ShipToCountry:   ps.ByNameDflt("s_country", "y"),                 //
//	}
//
//	// Authorize a payment
//	response := auth.Authorize(card, data, false)
//	if !response.IsApproved() {
//		fmt.Printf("CC Failed to Approve: %s\n", sizlib.SVar(response))
//		return fmt.Sprintf(`{"status":"error","msg":"Credit card failed to authorize, %s"}`, sizlib.EscapeDoubleQuote(response.ReasonText)), true, 402
//	}
//	fmt.Printf("%s\n", response)
//	fmt.Printf("Successful Authorization with id: %s ", response.TransId)
//
//	// Example of capture the preious authorization (using the transactionId)
//	response = auth.CapturePreauth(response.TransId, ps.ByNameDflt("ex_total", "0.0")) // 5.56
//	if !response.IsApproved() {
//		//log.Print("Capture failed : ")
//		//log.Print(response)
//		fmt.Printf("CC Failed to Preauth: %s\n", sizlib.SVar(response))
//		return fmt.Sprintf(`{"status":"error","msg":"Credit card failed to authorize/capture, %s"}`, sizlib.EscapeDoubleQuote(sizlib.SVar(response))), true, 402
//	}
//	// log.Print(response)
//	// log.Print("Successful Capture !")
//
//	// ps.SetValue("cc_authorize", sizlib.SVar(response))
//	goftlmux.AddValueToParams("cc_authorize", sizlib.SVar(response), 'J', goftlmux.FromOther, &ps)
//
//	return rv, exit, a_status
//}

/*
	,"/api/test/register_new_user": { "g": "test_register_new_user($1,$2,$3,$4,$5,$6,$7,$8,$9)", "p": [ "username", "password", "$ip$", "email", "real_name", "$url$", "csrf_token", "site", "name" ],

for ".G" - use "P" parameters and GetStoredProcedurePlist - so...
	PostgreSQL - ($1,$2...$n)
	ODBC		" "?,?,...       nof them
	Oracle		(:p1,:p2...:pN)

*/
func (hdlr *TabServer2Type) GetStoredProcedurePlist(np int) (rv string) {
	if hdlr.GetDbType() == DbType_postgres {
		rv = "("
		com := ""
		for i := 1; i <= np; i++ {
			rv += fmt.Sprintf("%s$%d", com, i)
			com = ","
		}
		rv += ")"
	} else if hdlr.GetDbType() == DbType_odbc {
		rv = " "
		com := ""
		for i := 0; i < np; i++ {
			rv += fmt.Sprintf("%s?", com)
			com = ","
		}
		rv += " "
	} else if hdlr.GetDbType() == DbType_Oracle {
		rv = "("
		com := ""
		for i := 0; i < np; i++ {
			rv += fmt.Sprintf("%s:p%d", com, i)
			com = ","
		}
		rv += ")"
	} else {
		panic("Error(00000): Not implemented yet.")
	}
	return
}

func MapKeys(kk map[string]interface{}) (rv string) {
	rv = ""
	com := ""
	for i := range kk {
		rv += com + i
		com = ","
	}
	if rv == "" {
		rv = " *** no keys *** "
	}
	return
}

func Get1stRowFromMap(kk map[string]interface{}) (rv interface{}) {
	rv = "{}"
	for _, v := range kk {
		rv = v
		return
	}
	return
}

func (hdlr *TabServer2Type) BindPlaceholder(n int) (rv string) {
	rv = "? "
	if hdlr.GetDbType() == DbType_postgres {
		rv = fmt.Sprintf("$%d ", n)
	} else if hdlr.GetDbType() == DbType_odbc {
		rv = "? "
	} else if hdlr.GetDbType() == DbType_mySQL {
		rv = "? "
	} else if hdlr.GetDbType() == DbType_MsSql {
		rv = "? "
	} else if hdlr.GetDbType() == DbType_Oracle {
		rv = fmt.Sprintf(":p%d ", n)
	}
	return
}

/*
create table "p_attr_meta" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "customer_id"			char varying (40) default '1'
	, "attr_type"			char varying (30) not null default 'str' check ( "attr_type" in ( 'int', 'int-r', 'str', 'str-r', 'float', 'float-r', 'date' , 'date-r', 'time', 'time-r', 'list-of', 'location', 'fraction' ) )
	, "table_name"			char varying (250) not null
	, "seq_no"				int
	, "attr_name"			char varying (250) not null
	, "fuzzy_search"		char varying(10) default 'n' not null
	, "attr_synonyms"		text not null
	, "default_value"		text 			-- Defaulit Value
	, "min_value"			text 			-- Min
	, "max_value"			text 			-- Max
	, "value_list"			text 			-- JSON list of values
	, "check_valid_data"	text 			-- RE used to validate
);
*/

type EAAttributeInfo struct {
	Attr_type        string
	Seq_no           int64
	Fuzzy_search     string
	Attr_synonyms    string
	Default_value    string
	Min_value        string
	Max_value        string
	Value_list       string
	Check_valid_data string
}

var EADataLoaded map[string]bool
var EA_Data map[string]map[string]EAAttributeInfo

func init() {
	EADataLoaded = make(map[string]bool)
	EA_Data = make(map[string]map[string]EAAttributeInfo)
}

func IfNullStr(x interface{}) string {
	if x == nil {
		return ""
	} else {
		return x.(string)
	}
}

func IfNullInt64(x interface{}) int64 {
	if x == nil {
		return 0
	} else {
		return x.(int64)
	}
}

func LoadEAData(table_name string, hdlr *TabServer2Type) {
	fmt.Printf("At, %s: Loading data for %s\n", godebug.LF(), table_name)
	d := sizlib.SelData(hdlr.gCfg.Pg_client.Db, "select * from p_attr_meta where table_name = $1", table_name)
	fmt.Printf("At, %s: data %s\n", godebug.LF(), sizlib.SVarI(d))
	for _, v := range d {
		fmt.Printf("At, %s\n", godebug.LF())
		if _, ok := EA_Data[table_name]; !ok {
			EA_Data[table_name] = make(map[string]EAAttributeInfo)
		}
		EA_Data[table_name][IfNullStr(v["attr_name"])] = EAAttributeInfo{
			Attr_type:        IfNullStr(v["attr_type"]),
			Seq_no:           IfNullInt64(v["seq_no"]),
			Fuzzy_search:     IfNullStr(v["fuzzy_search"]),
			Attr_synonyms:    IfNullStr(v["attr_synonyms"]),
			Default_value:    IfNullStr(v["default_value"]),
			Min_value:        IfNullStr(v["min_value"]),
			Max_value:        IfNullStr(v["max_value"]),
			Value_list:       IfNullStr(v["value_list"]),
			Check_valid_data: IfNullStr(v["check_valid_data"]),
		}
	}
	fmt.Printf("At, %s\n", godebug.LF())
	EADataLoaded[table_name] = true
}

// xyzzy synonyms for columns
// xyzzy count of values for columns in context
func EALookup(name string, table_name string, hdlr *TabServer2Type) (bool, *EAAttributeInfo) {
	fmt.Printf("At, %s\n", godebug.LF())
	if b, ok := EADataLoaded[table_name]; !ok || !b {
		fmt.Printf("At, %s\n", godebug.LF())
		LoadEAData(table_name, hdlr)
	}
	if x, ok := EA_Data[table_name][name]; ok {
		fmt.Printf("At, %s\n", godebug.LF())
		return ok, &x
	}
	fmt.Printf("At, %s\n", godebug.LF())
	return false, nil
}

// xyzzy5000
// qt_m["cat_col"] = fmt.Sprintf(`%s%s%s%s`, h.setWhereAlias, DbBeginQuote, h.Category_col, DbEndQuote)
// t, _ := GetDataListAsInList("s", vv, trx, h, bind)
// Used in categories - this is a postgresql specific function. -- Sets up an in list.
func GetDataListAsInList(ty string, wc WhereClause, trx *tr.Trx, h SQLOne, bind *[]interface{}) (string, error) {
	s := "("
	com := ""
	for _, vv := range wc.List {
		switch ty {
		case "i":
			s += com + fmt.Sprintf("%d", vv.Val1i)
		case "f":
			s += com + fmt.Sprintf("%f", vv.Val1f)
		case "u": // UUID/GUID
			s += com + "'" + vv.Val1s + "'"
		case "":
			fallthrough
		case "s":
			s += com + "'" + vv.Val1s + "'"
		/* d, t, e */
		default:
			trx.AddNote(1, fmt.Sprintf("Invalid Type: %s for column %s", ty, wc.Name))
			return "", errors.New(fmt.Sprintf("Error(12233): Invalid Type: %s for column %s", ty, wc.Name))
		}
		com = ","
	}
	s = s + ")"
	fmt.Printf("Category|Tag At, %s ->%s<-\n", godebug.LF(), s)
	return s, nil
}

// Used in categories - this is a postgresql specific function.
func GetDataListAsArray(ty string, wc WhereClause, trx *tr.Trx, h SQLOne, bind *[]interface{}) (string, error) {
	s := "{"
	com := ""
	for _, vv := range wc.List {
		switch ty {
		case "i":
			s += com + fmt.Sprintf("%d", vv.Val1i)
		case "f":
			s += com + fmt.Sprintf("%f", vv.Val1f)
		case "u": // UUID/GUID
			s += com + `"` + vv.Val1s + `"`
		case "":
			fallthrough
		case "s":
			s += com + `"` + vv.Val1s + `"`
		/* d, t, e */
		default:
			trx.AddNote(1, fmt.Sprintf("Invalid Type: %s for column %s", ty, wc.Name))
			return "", errors.New(fmt.Sprintf("Error(12233): Invalid Type: %s for column %s", ty, wc.Name))
		}
		com = ","
	}
	s = s + "}"
	fmt.Printf("Category|Tag At, %s ->%s<-\n", godebug.LF(), s)
	return s, nil
}

func GetColByType(attr_type string) string {
	switch attr_type {
	case "i":
		return "val1i"
	case "f":
		return "val1f"
	case "u": // UUID/GUID
		return "val1s"
	case "":
		fallthrough
	case "s":
		return "val1s"
	case "d", "t", "e":
		return "val1d"
	default:
		return "val1s"
	}
}

func AddBindValueByType(bind *[]interface{}, vv WhereClause, ty string) (pos int) {
	switch ty {
	case "i":
		return AddBindValue(bind, vv.Val1i)
	case "f":
		return AddBindValue(bind, vv.Val1f)
	case "u": // UUID/GUID
		return AddBindValue(bind, vv.Val1s)
	case "":
		fallthrough
	case "s":
		return AddBindValue(bind, vv.Val1s)
	case "d", "t", "e":
		return AddBindValue(bind, vv.Val1d)
	default:
		return AddBindValue(bind, vv.Val1s)
	}
}

// xyzzy - missing $customer_id$ if set/used
/*
http://tech.pro/tutorial/1142/building-faceted-search-with-postgresql  -- faceted-filters.html (saved)
http://www.postgresql.org/docs/8.3/static/textsearch-indexes.html
http://stackoverflow.com/questions/1540374/why-are-postgresql-text-search-gist-indexes-so-much-slower-than-gin-indexes
http://www.youlikeprogramming.com/2012/01/full-text-search-fts-in-postgresql-9-1/		-- trigger to update tsvecotr colum/index

http://blog.timothyandrew.net/blog/2013/06/24/recursive-postgres-queries/
http://stackoverflow.com/questions/11834579/postgresql-hierarchical-category-tree -- Category Example With Queries
https://gist.github.com/chanmix51/3225313 -- GIST example with get_chilren_of, get_parent_of, list and create table stuff

select * from p_get_children_of ( '{"guns","stock"}'::varchar[] );

, "key_word_col_name": "key_word"
, "key_word_list_col": "__keyword__"
, "key_word_tmpl": " %{kw_col%} @@ plainto_tsquery( %{kw_vals%} ) "

, "category_col_name": "category"
, "category_col": "category_id"
, "category_tmpl": " %{cat_col%} in ( select c1.\"id\" from p_get_children_of ( '%{cat_val%}'::varchar[] ) ) "

, "attr_table_name": "p_cart"
, "attr_col": "id"
, "attr_tmpl":" %{pk_attr_id%} in ( select a1.\"fk_id\" from \"p_attr\" as a1 where a1.\"attr_type\" = %{attr_type%} and a1.\"attr_name\" = %{attr_name%} and a1.\"%{ref_col%}\" %{attr_op%} %{attr_val%} )"

	Key_word_col_name string
	Key_word_list_col string
	Key_word_tmpl     string

	Category_col_name string
	Category_col      string
	Category_tmpl     string

	Attr_table_name string
	Attr_col        string
	Attr_tmpl       string

*/
func ExtendeAttributes(vv WhereClause, wc WhereClause, trx *tr.Trx, h SQLOne, bind *[]interface{}, hdlr *TabServer2Type) (processed bool, rv string, err error) {
	var bp int

	processed = false
	rv = " false "
	err = nil
	qt_m := make(map[string]string)
	if vv.Name == h.Key_word_col_name {
		fmt.Printf("%sAt, %s%s\n", MiscLib.ColorCyan, godebug.LF(), MiscLib.ColorReset)
		processed = true
		trx.AddNote(2, "Processing key_word extended attribute")
		// legit ops on keywords,
		//		kw = 'a'
		// https://www.postgresql.org/docs/9.5/static/textsearch-controls.html#TEXTSEARCH-PARSING-QUERIES
		// Looks like operator should be "@@"
		if !sizlib.InArray(vv.Op, []string{"==", "="}) {
			fmt.Printf("At, %s\n", godebug.LF())
			trx.AddNote(1, fmt.Sprintf("Error: Processing extended attributes, invalid op=%s %s%s", vv.Op, h.setWhereAlias, vv.Name))
			return false, "", errors.New(fmt.Sprintf("Error(14239): Error: Processing extended attributes, invalid op=%s %s%s", vv.Op, h.setWhereAlias, vv.Name))
		} else {
			fmt.Printf("At, %s\n", godebug.LF())
			bp = AddBindValue(bind, vv.Val1s)          // keywords are alwasy strings, or list of strings
			qt_m["kw_vals"] = hdlr.BindPlaceholder(bp) // %{kw_vals%}
			qt_m["kw_col"] = fmt.Sprintf(`%s%s%s%s`, h.setWhereAlias, DbBeginQuote, h.Key_word_list_col, DbEndQuote)
			rv = sizlib.Qt(h.Key_word_tmpl, qt_m)
		}
	} else if vv.Name == h.Category_col_name {
		fmt.Printf("%sAt, %s%s\n", MiscLib.ColorCyan, godebug.LF(), MiscLib.ColorReset)
		processed = true
		trx.AddNote(2, "Processing category extended attribute")
		// %{cat_vals%}
		// legit ops on keywords,
		//		cat in ( 'id1', 'id2' )
		if vv.Op != "in" {
			fmt.Printf("At, %s\n", godebug.LF())
			trx.AddNote(1, fmt.Sprintf("Error: Processing extended attributes, invalid op=%s %s%s", vv.Op, h.setWhereAlias, vv.Name))
			return false, "", errors.New(fmt.Sprintf("Error(14239): Error: Processing extended attributes, invalid op=%s %s%s", vv.Op, h.setWhereAlias, vv.Name))
		} else {
			fmt.Printf("At, %s\n", godebug.LF())
			s, err := GetDataListAsArray("s", vv, trx, h, bind)
			// xyzzy5000
			t, _ := GetDataListAsInList("s", vv, trx, h, bind)
			if err != nil {
				fmt.Printf("At, %s\n", godebug.LF())
				trx.AddNote(1, fmt.Sprintf("Error: Processing extended attributes, invalid data type=%s %s%s", "s", h.setWhereAlias, vv.Name))
				return false, "", errors.New(fmt.Sprintf("Error(14239): Error: Processing extended attributes, invalid data type=%s %s%s", "s", h.setWhereAlias, vv.Name))
			} else {
				qt_m["cat_vals"] = s
				qt_m["cat_in_vals"] = t
				qt_m["cat_col"] = fmt.Sprintf(`%s%s%s%s`, h.setWhereAlias, DbBeginQuote, h.Category_col, DbEndQuote)
				fmt.Printf("s = ->%s<- qt_m=%s At, %s\n", s, godebug.SVar(qt_m), godebug.LF())
				rv = sizlib.Qt(h.Category_tmpl, qt_m)
			}
		}
	} else if h.Attr_table_name != "" { // xyzzy - this should be some checkon name in attribute table names?
		fmt.Printf("At, %s\n", godebug.LF())
		fnd, x := EALookup(vv.Name, h.Attr_table_name, hdlr)
		if fnd {
			fmt.Printf("At, %s\n", godebug.LF())
			processed = true
			trx.AddNote(2, "Processing dynamic attribute extended attribute")
			qt_m["attr_col"] = fmt.Sprintf(`%s%s%s%s`, h.setWhereAlias, DbBeginQuote, h.Attr_col, DbEndQuote)
			qt_m["attr_type"] = x.Attr_type                // %{attr_type%} - from meta, x.Attr_type
			qt_m["attr_name"] = vv.Name                    // %{attr_name%} - vv.Name
			qt_m["ref_col"] = GetColByType(x.Attr_type)    // %{ref_col%} - derived from vv.Name + vv.Op - and fuzzy_search
			qt_m["attr_op"] = vv.Op                        // %{attr_op%} - vv.Op
			bp = AddBindValueByType(bind, vv, x.Attr_type) // keywords are alwasy strings, or list of strings
			qt_m["attr_vals"] = hdlr.BindPlaceholder(bp)   // %{attr_vals%} -- vv.Val1s, etc (based on type)
			rv = sizlib.Qt(h.Attr_tmpl, qt_m)
		}
	} else if vv.Name == h.Tag_col_name {
		fmt.Printf("%sAt, %s%s\n", MiscLib.ColorCyan, godebug.LF(), MiscLib.ColorReset)
		processed = true
		trx.AddNote(2, "Processing tag extended attribute")
		// %{cat_vals%}
		// legit ops on keywords,
		//		cat in ( 'id1', 'id2' )
		if vv.Op != "in" {
			fmt.Printf("At, %s\n", godebug.LF())
			trx.AddNote(1, fmt.Sprintf("Error: Processing extended attributes, invalid op=%s %s%s", vv.Op, h.setWhereAlias, vv.Name))
			return false, "", errors.New(fmt.Sprintf("Error(14239): Error: Processing extended attributes, invalid op=%s %s%s", vv.Op, h.setWhereAlias, vv.Name))
		} else {
			fmt.Printf("At, %s\n", godebug.LF())
			s, err := GetDataListAsArray("s", vv, trx, h, bind)
			// xyzzy5000
			t, _ := GetDataListAsInList("s", vv, trx, h, bind)
			if err != nil {
				fmt.Printf("At, %s\n", godebug.LF())
				trx.AddNote(1, fmt.Sprintf("Error: Processing extended attributes, invalid data type=%s %s%s", "s", h.setWhereAlias, vv.Name))
				return false, "", errors.New(fmt.Sprintf("Error(14239): Error: Processing extended attributes, invalid data type=%s %s%s", "s", h.setWhereAlias, vv.Name))
			} else {
				qt_m["tag_vals"] = s
				qt_m["tag_in_vals"] = t
				qt_m["tag_col"] = fmt.Sprintf(`%s%s%s%s`, h.setWhereAlias, DbBeginQuote, h.Tag_col, DbEndQuote)
				rv = sizlib.Qt(h.Tag_tmpl, qt_m)
				fmt.Printf("%ss = ->%s<- qt_m=%s rv= ->%s<- At, %s%s\n", MiscLib.ColorYellow, s, godebug.SVar(qt_m), rv, godebug.LF(), MiscLib.ColorReset)
			}
		}
	}
	return
}

// func respHandlerCreateNewDeviceId(www http.ResponseWriter, req *http.Request) {
// if rw, ok := www.(*goftlmux.MidBuffer); ok {
func GetRwPs(www http.ResponseWriter, req *http.Request) (rw *goftlmux.MidBuffer, top_hdlr interface{}, ps *goftlmux.Params, err error) {
	var ok bool
	rw, top_hdlr, ok = GetRwHdlrFromWWW(www, req) // xyzzy - hdlr alwasy is top - an error
	if rw != nil {
		ps = &rw.Ps
	}
	if !ok {
		err = fmt.Errorf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF(2))
		fmt.Fprintf(os.Stderr, "Failed to get rw/hdlr, AT: %s\n", godebug.LF())
		return
	}
	// fmt.Fprintf(os.Stderr, "Just before setting ps, %v, %v, %s\n", ps, rw.Ps, godebug.LF())
	return
}

// ============================================================================================================================================
func GetRwHdlrFromWWW(www http.ResponseWriter, req *http.Request) (rw *goftlmux.MidBuffer, top_hdlr interface{}, ok bool) {

	rw, ok = www.(*goftlmux.MidBuffer)
	if !ok {
		// AnError(hdlr, www, req, 500, 5, fmt.Sprintf("hdlr not correct type in rw.!, %s\n", godebug.LF()))
		fmt.Fprintf(os.Stderr, "Passed Wrong Thing, AT: %s\n", godebug.LF())
		return
	}

	top_hdlr = rw.Hdlr
	return
}

const debugCrud01 = false

/*
	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	defer hdlr.gCfg.RedisPool.Put(conn)
*/

type DbType int

const (
	DbType_postgres DbType = 1
	DbType_Oracle   DbType = 2
	DbType_MsSql    DbType = 3
	DbType_mySQL    DbType = 4
	DbType_odbc     DbType = 5
)

func (dd DbType) String() string {
	switch dd {
	case DbType_postgres:
		return "DbType_postgres"
	case DbType_Oracle:
		return "DbType_Oracle"
	case DbType_MsSql:
		return "DbType_MsSql"
	case DbType_mySQL:
		return "DbType_mySQL"
	case DbType_odbc:
		return "DbType_odbc"
	default:
		return fmt.Sprintf("-- unknown DbType %d --", dd)
	}
}

// return the type of the database --
func (hdlr *TabServer2Type) GetDbType() DbType {
	// from hdlr -> gCfg -> Global Config
	// conn := sizlib.ConnectToAnyDb(hdlr.DBType, hdlr.PGConn, hdlr.DBName)
	switch hdlr.gCfg.DBType {
	case "postgres":
		return DbType_postgres
	case "Oracle":
		return DbType_Oracle
	case "MsSQL":
		return DbType_MsSql
	case "odbc":
		return DbType_odbc
	case "mySQL", "mariadb":
		return DbType_mySQL
	}
	return DbType_postgres
}

func (hdlr *TabServer2Type) RemapParams(ps *goftlmux.Params, h SQLOne, trx *tr.Trx) {
	if db_RemapParams {
		fmt.Printf("AT: Comment:%s, %s\n", h.Comment, godebug.LF())
	}
	didRemap := false
	for _, vv := range h.ReMapParameter {
		if db_RemapParams {
			fmt.Printf("AT: %s\n", godebug.LF())
		}
		isRes := cfg.ReservedItems[vv.ToName]
		if vv.FromName != vv.ToName && !isRes {
			if ps.HasName(vv.FromName) {
				didRemap = true
				if db_RemapParams {
					fmt.Printf("AT: Remapping! To:%s From:%s isRes=%v, %s\n", vv.ToName, vv.FromName, isRes, godebug.LF())
					trx.AddNote(2, "Remaping Parameter Name "+vv.FromName+" to "+vv.ToName) // xyzzyTRX - announce remap to TRX for testing -- if trxTest --
				}
				s := ps.ByName(vv.FromName)
				goftlmux.AddValueToParams(vv.ToName, s, 'r', goftlmux.FromOther, ps)
			}
		}
	}
	if didRemap {
		if db_RemapParams {
			fmt.Printf("%sParams: AT %s - End of ReMap: %s%s\n", MiscLib.ColorYellow, godebug.LF(), ps.DumpParamTable(), MiscLib.ColorReset)
		}
	}
}

type SesItems struct {
	Path  []string
	Value string
}

type SesDataType struct {
	Set []SesItems
}

func ConvRawSesData(ss string) (rv SesDataType) {
	err := json.Unmarshal([]byte(ss), &rv)
	if err != nil {
		rv.Set = []SesItems{}
	}
	return
}

// Helper func:  Read input from specified file or stdin
func loadData(p string) ([]byte, error) {
	if p == "" {
		return nil, fmt.Errorf("No path specified")
	}

	var rdr io.Reader
	//	if p == "-" {
	//		rdr = os.Stdin
	//	} else if p == "+" {
	//		return []byte("{}"), nil
	//	} else {
	if f, err := os.Open(p); err == nil {
		rdr = f
		defer f.Close()
	} else {
		return nil, err
	}
	//	}
	return ioutil.ReadAll(rdr)
}

// Create, sign, and output a token.  This is a great, simple example of
// how to use this library to create and sign a token.
func SignToken(tokData []byte, keyFile string) (out string, err error) {

	// parse the JSON of the claims
	var claims jwt.MapClaims
	if err = json.Unmarshal(tokData, &claims); err != nil {
		err = fmt.Errorf("Couldn't parse claims JSON: %v", err)
		return
	}

	fmt.Fprintf(os.Stderr, "Siging: %s, AT: %s\n", tokData, godebug.LF())
	fmt.Fprintf(os.Stderr, "Claims: %s, AT: %s\n", godebug.SVarI(claims), godebug.LF())

	//-	// add command line claims
	//-	if len(flagClaims) > 0 {
	//-		for k, v := range flagClaims {
	//-			claims[k] = v
	//-		}
	//-	}

	// get the key
	var key interface{}
	key, err = loadData(keyFile)
	if err != nil {
		err = fmt.Errorf("Couldn't read key: %v", err)
		return
	}

	// get the signing alg
	// alg := jwt.GetSigningMethod(*flagAlg)
	alg := jwt.GetSigningMethod("RS256") // xyzzy - Param
	if alg == nil {
		err = fmt.Errorf("Couldn't find signing method: %v", "RS256") // xyzzy Param
		return
	}

	// create a new token
	token := jwt.NewWithClaims(alg, claims)

	//-	// add command line headers
	//-	if len(flagHead) > 0 {
	//-		for k, v := range flagHead {
	//-			token.Header[k] = v
	//-		}
	//-	}

	if isEs() {
		if k, ok := key.([]byte); !ok {
			err = fmt.Errorf("Couldn't convert key data to key")
			return
		} else {
			key, err = jwt.ParseECPrivateKeyFromPEM(k)
			if err != nil {
				return
			}
		}
	} else if isRs() {
		if k, ok := key.([]byte); !ok {
			err = fmt.Errorf("Couldn't convert key data to key")
			return
		} else {
			key, err = jwt.ParseRSAPrivateKeyFromPEM(k)
			if err != nil {
				return
			}
		}
	}

	if out, err = token.SignedString(key); err == nil {
		if db81 {
			fmt.Println(out)
		}
	} else {
		err = fmt.Errorf("Error signing token: %v", err)
	}

	return
}

func isEs() bool {
	// return strings.HasPrefix(*flagAlg, "ES")
	return false
}

func isRs() bool {
	// return strings.HasPrefix(*flagAlg, "RS")
	return true
}

const db81 = false
const db1 = false
const db2 = false
const db4 = false // Redis Sessions
const db_post = false
const db_get1 = false
const db_RemapParams = false
const db_trace_functions = false
const db_GenUpdateSet = false

const db_DumpInsert = true
const db_DumpDelete = true

// const hdlr.gCfg.DbOn("*", "TabServer2", "db-closure-1") = true
// const hdlr.gCfg.DbOn("*", "TabServer2", "db-closure-2") = true
// const hdlr.gCfg.DbOn("*", "TabServer2", "db-email") = true
// const hdlr.gCfg.DbOn("*", "TabServer2", "db-session") = true
//			if hdlr.gCfg.DbOn("*", "SessionRedis", "db1") {

/* vim: set noai ts=4 sw=4: */
