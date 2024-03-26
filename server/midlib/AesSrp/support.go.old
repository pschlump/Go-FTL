//
// Package aessrp implements encrypted authentication and encrypted REST.
// SRP-6a for login authenticaiton, followed by AES 256 bit encrypted RESTful calls.
//
// Copyright (C) Philip Schlump, 2013-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 0001
// 你好无聊的世界
//

package AesSrp

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	logrus "github.com/pschlump/pslog" // "github.com/sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"  //
	"github.com/pschlump/godebug"            //
	"github.com/pschlump/verhoeff_algorithm" //
	//
	"github.com/pschlump/json"   //	"encoding/json" - modified to allow dummy output of chanels
	"golang.org/x/crypto/pbkdf2" // "golang.org/x/crypto/pbkdf2"
)

// ============================================================================================================================================
// ============================================================================================================================================
// 1. Lookup user based on tt the session id - in Redis
// opts.Password, encData.Salt, encData.Iter, encData.KeySizeBytes = GetKeyData ( email, tt )
func GetKeyData(hdlr *AesSrpType, rw *goftlmux.MidBuffer, tt, SandBoxPrefix string) (pw string, salt string, key string, iter int, keysize int, email string, ss map[string]interface{}) {
	s, err := DbGetString(hdlr, rw, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt)) // .PreKey=="ses:"
	if err != nil {
		pw = ""
		fmt.Printf("Error on DbGetString : %s, %s\n", err, godebug.LF())
		return
	}
	if dbCipher {
		fmt.Printf("Gettting %s to %s\n", SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), s) // .PreKey=="ses:"
	}
	err = DbExpire(hdlr, rw, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), hdlr.SessionLife) // .PreKey=="ses:"

	type jsonDataType struct {
		Pw      string
		Salt    string
		Key     string
		Iter    int
		Keysize int
		Email   string
	}
	var rv jsonDataType
	rv.Salt, rv.Key = "", ""

	err = json.Unmarshal([]byte(s), &rv)
	if err != nil {
		pw = ""
		// xyzzyLogrus
		fmt.Printf("Error on Unmarshal(1) - Json Parse: %s, %s\n", err, godebug.LF())
		return
	}
	pw = rv.Pw
	ssalt, _ := base64.StdEncoding.DecodeString(rv.Salt)
	skey, _ := base64.StdEncoding.DecodeString(rv.Key)
	salt = string(ssalt)
	key = string(skey)
	iter = rv.Iter
	keysize = rv.Keysize
	email = rv.Email

	fmt.Printf("\n----------------------\nkey [%x] len=%d\n----------------------\n\n", skey, len(skey))

	err = json.Unmarshal([]byte(s), &ss)
	if err != nil {
		pw = ""
		// xyzzyLogrus
		fmt.Printf("Error on Unmarshal(2) - Json Parse: %s, %s\n", err, godebug.LF())
		return
	}
	delete(ss, "Pw")
	delete(ss, "Salt")
	delete(ss, "Key")
	delete(ss, "Iter")
	delete(ss, "Keysize")
	delete(ss, "Email")

	fmt.Printf("Success on regular GetKeyData -exit- at bottom\n")

	return
}

// ============================================================================================================================================
// ============================================================================================================================================
// 1. Lookup user based on tt the session id - in Redis
// opts.Password, encData.Salt, encData.Iter, encData.KeySizeBytes = GetKeyData ( email, tt )
func GetKeyDataRaw(hdlr *AesSrpType, rw *goftlmux.MidBuffer, tt, SandBoxPrefix string) (key string, email string, ss map[string]interface{}) {
	s, err := DbGetString(hdlr, rw, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt)) // .PreKey=="ses:"
	if err != nil {
		key = ""
		fmt.Printf("Error on DbGetString : %s, %s\n", err, godebug.LF())
		return
	}
	if dbCipher {
		fmt.Printf("Gettting %s to %s\n", SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), s) // .PreKey=="ses:"
	}
	err = DbExpire(hdlr, rw, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), hdlr.SessionLife) // .PreKey=="ses:"

	err = json.Unmarshal([]byte(s), &ss)
	if err != nil {
		key = ""
		// xyzzyLogrus
		fmt.Printf("Error on Unmarslal(1) - Json Parse: %s, %s\n", err, godebug.LF())
		return
	}

	fmt.Printf("GetKeyDataRaw ss=%s\n", godebug.SVarI(ss))

	email, _ = ss["Email"].(string)
	skey, _ := base64.StdEncoding.DecodeString(ss["Key"].(string))
	key = string(skey)

	fmt.Printf("\n----------------------\nkey [%x] len=%d\n----------------------\n\n", skey, len(skey))

	return
}

// ============================================================================================================================================
func UserGetEmail(hdlr *AesSrpType, rw *goftlmux.MidBuffer, tt, SandBoxPrefix string) (email string, err error) {
	s, err := DbGetString(hdlr, rw, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt)) // .PreKey=="ses:"
	if err != nil {
		return
	}
	if dbCipher {
		fmt.Printf("Gettting %s to %s\n", SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), s) // .PreKey=="ses:"
	}
	err = DbExpire(hdlr, rw, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), hdlr.SessionLife) // .PreKey=="ses:"

	type jsonDataType struct {
		Email string
	}
	var rv jsonDataType

	err = json.Unmarshal([]byte(s), &rv)
	if err != nil {
		// xyzzyLogrus
		return
	}
	email = rv.Email

	return
}

// _ = UpdateSessionEncryptionKey(hdlr, rw, tt, SandBoxPrefix, tmp_login_mdata["key2"], raw_session)
func UpdateSessionEncryptionKey(hdlr *AesSrpType, rw *goftlmux.MidBuffer, tt, SandBoxPrefix string, keyInHex string, data map[string]interface{}) (err error) {
	key := SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt) // .PreKey=="ses:"

	//	s, err := DbGetString(hdlr, rw, key)              // .PreKey=="ses:"
	//	if err != nil {
	//		return
	//	}
	//	err = DbExpire(hdlr, rw, key, hdlr.SessionLife) // .PreKey=="ses:"
	//
	//	data, err := lib.JsonStringToString(s)
	//	if err != nil {
	//		return
	//	}

	keyBinary, err := hex.DecodeString(keyInHex)
	if err != nil {
		return
	}

	skey := base64.StdEncoding.EncodeToString(keyBinary)

	data["Key"] = skey

	ss := lib.SVar(data)
	DbSetExpire(hdlr, rw, key, ss, hdlr.SessionLife) // .PreKey=="ses:"

	return
}

// ============================================================================================================================================
func XxSaveUserExists(hdlr *AesSrpType, rw *goftlmux.MidBuffer, tt, SandBoxPrefix string, session map[string]interface{}) (err error) {
	_, err = DbGetString(hdlr, rw, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt)) // .PreKey=="ses:"
	if err != nil {
		return
	}
	return
}

// ============================================================================================================================================
// need to do a get/merge at this point to preserve any additional data in key. (Session Data like "email address")
func XxSaveSessionData(hdlr *AesSrpType, rw *goftlmux.MidBuffer, tt, SandBoxPrefix string, session map[string]interface{}) (err error) {
	data := make(map[string]interface{})
	s, err := DbGetString(hdlr, rw, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt)) // .PreKey=="ses:"
	if err == nil {
		err = json.Unmarshal([]byte(s), &data)
		if err != nil {
			fmt.Printf("Error %s - unable to unmarshal session data, key=%s rawdata=%s\n", err, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), s) // .PreKey=="ses:"
			return
		}
	}
	for name, val := range session {
		switch name {
		case "Pw", "Salt", "Key", "Iter", "Keysize":
		default:
			data[name] = val
		}
	}
	ss := lib.SVar(data)
	// ss := fmt.Sprintf(`{"Pw":%q,"Salt":%q,"Key":%q,"Iter":%d,"Keysize":%d}`, pw, ssalt, skey, iter, keysize)
	if dbCipher {
		fmt.Printf("Setting %s to %s\n", SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), ss) // .PreKey=="ses:"
	}
	DbSetExpire(hdlr, rw, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), ss, hdlr.SessionLife) // .PreKey=="ses:"
	return
}

// ============================================================================================================================================
func SaveLogoutData(hdlr *AesSrpType, rw *goftlmux.MidBuffer, tt, SandBoxPrefix string) (err error) {
	data := make(map[string]interface{})
	s, err := DbGetString(hdlr, rw, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt)) // .PreKey=="ses:"
	if err == nil {
		err = json.Unmarshal([]byte(s), &data)
		if err != nil {
			// xyzzyLogrus
			fmt.Printf("Error %s - unable to unmarshal session data, key=%s rawdata=%s\n", err, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), s) // .PreKey=="ses:"
			return
		}
	}
	data["Salt"] = "0"
	data["Key"] = "0"
	ss := lib.SVar(data)
	// ss := fmt.Sprintf(`{"Pw":%q,"Salt":%q,"Key":%q,"Iter":%d,"Keysize":%d}`, pw, ssalt, skey, iter, keysize)
	if dbCipher {
		fmt.Printf("Setting %s to %s\n", SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), ss) // .PreKey=="ses:"
	}
	DbSetExpire(hdlr, rw, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), ss, hdlr.SessionLife) // .PreKey=="ses:"
	return
}

// ============================================================================================================================================
func SaveInitData(hdlr *AesSrpType, rw *goftlmux.MidBuffer, tt, SandBoxPrefix, pw, email, key, privs string) (err error) {
	data := make(map[string]interface{})
	s, err := DbGetString(hdlr, rw, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt)) // .PreKey=="ses:"
	if err == nil {
		err = json.Unmarshal([]byte(s), &data)
		if err != nil {
			// xyzzyLogrus
			fmt.Printf("Error %s - unable to unmarshal session data, key=%s rawdata=%s\n", err, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), s) // .PreKey=="ses:"
			return
		}
	}

	data["Pw"] = pw               //	This is the generated common key 'k'
	data["Email"] = email         //
	t := time.Now()               //
	tss := t.Format(time.RFC3339) //
	data["n_failed_login"] = "0"
	data["login_fail_time"] = ""  //
	data["login_date_time"] = tss //
	data["$auth$"] = "y"
	data["$username$"] = email
	data["$auth_key$"] = key
	data["$privs$"] = privs
	ss := lib.SVar(data)
	// ss := fmt.Sprintf(`{"Pw":%q,"Salt":%q,"Key":%q,"Iter":%d,"Keysize":%d}`, pw, ssalt, skey, iter, keysize)
	if dbCipher {
		fmt.Printf("Setting %s to %s\n", SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), ss) // .PreKey=="ses:"
	}
	DbSetExpire(hdlr, rw, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), ss, hdlr.SessionLife) // .PreKey=="ses:"
	return
}

// ============================================================================================================================================
func SaveInitFailedLogin(hdlr *AesSrpType, rw *goftlmux.MidBuffer, tt, SandBoxPrefix string) (err error) {
	data := make(map[string]interface{})
	s, err := DbGetString(hdlr, rw, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt)) // .PreKey=="ses:"
	if err == nil {
		err = json.Unmarshal([]byte(s), &data)
		if err != nil {
			// xyzzyLogrus
			fmt.Printf("Error %s - unable to unmarshal session data, key=%s rawdata=%s\n", err, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), s) // .PreKey=="ses:"
			return
		}
	}
	data["auth"] = "n"
	data["Pw"] = ""
	data["Email"] = ""
	ss := lib.SVar(data)
	// ss := fmt.Sprintf(`{"Pw":%q,"Salt":%q,"Key":%q,"Iter":%d,"Keysize":%d}`, pw, ssalt, skey, iter, keysize)
	if dbCipher {
		fmt.Printf("Setting %s to %s\n", SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), ss) // .PreKey=="ses:"
	}
	DbSetExpire(hdlr, rw, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), ss, hdlr.SessionLife) // .PreKey=="ses:"
	return
}

// ============================================================================================================================================
func SaveKeyData(hdlr *AesSrpType, rw *goftlmux.MidBuffer, tt, SandBoxPrefix string, pw, salt, key string, iter int, keysize int) (err error) {
	data := make(map[string]interface{})
	s, err := DbGetString(hdlr, rw, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt)) // .PreKey=="ses:"
	if err == nil {
		err = json.Unmarshal([]byte(s), &data)
		if err != nil {
			// xyzzyLogrus
			fmt.Printf("Error %s - unable to unmarshal session data, key=%s rawdata=%s\n", err, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), s) // .PreKey=="ses:"
			return
		}
	}
	ssalt := base64.StdEncoding.EncodeToString([]byte(salt))
	skey := base64.StdEncoding.EncodeToString([]byte(key))
	data["Pw"] = pw
	data["Salt"] = ssalt
	data["Key"] = skey
	data["Iter"] = iter
	data["Keysize"] = keysize
	ss := lib.SVar(data)
	// ss := fmt.Sprintf(`{"Pw":%q,"Salt":%q,"Key":%q,"Iter":%d,"Keysize":%d}`, pw, ssalt, skey, iter, keysize)
	if dbCipher || true {
		fmt.Printf("SaveKeyData: Setting %s to %s\n", SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), ss) // .PreKey=="ses:"
	}
	DbSetExpire(hdlr, rw, SandBoxKey(hdlr.PreKey, SandBoxPrefix, tt), ss, hdlr.SessionLife) // .PreKey=="ses:"
	return
}

func SaveEmailAuth(hdlr *AesSrpType, rw *goftlmux.MidBuffer, email, SandBoxPrefix, emailAuthToken string) {
	godebug.Db2Printf(db202, "SessionLife in Seconds=%d\n", hdlr.SessionLife)
	godebug.Db2Printf(db202, "Key1 =%s\n", SandBoxKey(hdlr.PreEau, SandBoxPrefix, email))
	godebug.Db2Printf(db202, "Key2 =%s\n", SandBoxKey(hdlr.PreEau, SandBoxPrefix, emailAuthToken))
	godebug.Db2Printf(db202, "Email =%s\n", email)
	godebug.Db2Printf(db202, "emailAuthToken =%s\n", emailAuthToken)

	DbSetExpire(hdlr, rw, SandBoxKey(hdlr.PreEau, SandBoxPrefix, email), emailAuthToken, hdlr.SessionLife) // PreEau=='eau:'
	DbSetExpire(hdlr, rw, SandBoxKey(hdlr.PreEau, SandBoxPrefix, emailAuthToken), email, hdlr.SessionLife) // PreEau=='eau:'
}

func GetEmailAuth(hdlr *AesSrpType, rw *goftlmux.MidBuffer, emailAuthToken, SandBoxPrefix string) (email string, ok bool) {
	var err error
	email, err = DbGetString(hdlr, rw, SandBoxKey(hdlr.PreEau, SandBoxPrefix, emailAuthToken))
	godebug.Db2Printf(db202, "Code 16: %s --->>> %s\n", SandBoxKey(hdlr.PreEau, SandBoxPrefix, emailAuthToken), err)
	if err != nil {
		return
	}
	ok = true
	return
}

func IsLoggedIn(hdlr *AesSrpType, rw *goftlmux.MidBuffer, ps goftlmux.Params) (rv bool) {
	rv = false
	SandBoxPrefix := ps.ByNameDflt("GOFTL_Sandbox", "")
	if pw, found := ps.GetByNameAndType("$auth_key$", goftlmux.FromAuth); found {
		if un, found := ps.GetByNameAndType("$username$", goftlmux.FromAuth); found {
			rkey, err := DbGetString(hdlr, rw, SandBoxKey(hdlr.PreAuth, SandBoxPrefix, un))
			if err != nil {
				return
			}
			if rkey == pw {
				return true
			}
		}
	}
	return
}

func SetLoggedIn(hdlr *AesSrpType, rw *goftlmux.MidBuffer, un, SandBoxPrefix, key string) {
	DbSetExpire(hdlr, rw, SandBoxKey(hdlr.PreAuth, SandBoxPrefix, un), key, hdlr.KeySessionLife)
}

func SaveCookieAuth(hdlr *AesSrpType, rw *goftlmux.MidBuffer, cookieValue, SandBoxPrefix, ip, email, hash, id, privs string) {
	DbSetExpire(hdlr, rw, SandBoxKey(hdlr.PreAuth, SandBoxPrefix, cookieValue), fmt.Sprintf(`{"ip":%q,"email":%q,"id":%q,"hash":%q, "privs":%q}`, ip, email, id, hash, privs), hdlr.CookieSessionLife)

}

// ----------------------------------------------------------------------------------------------------------------------------
// func Sha256(s string) (rv string) {
// 	rv = HashString.HashString(s)
// 	return
// }

// ----------------------------------------------------------------------------------------------------------------------------
// email, err = GetEmailFromRegKey ( regKey )
func GetEmailFromRegKey(regKey string) (email string, err error) {
	return
}

func SaveSupportMessage(rw *goftlmux.MidBuffer, fr, sub, bod string) {
	// fmt.Fprintf(MsgLog, `{"type":"email","from":%q,"subject":%q,"body":%q}`+"\n", fr, sub, bod)
	logrus.Warn(fmt.Sprintf(`{"type":"email","from":%q,"subject":%q,"body":%q}`+"\n", fr, sub, bod))
}

// ----------------------------------------------------------------------------------------------------------------------------
func Save2FactorAuth(hdlr *AesSrpType, rw *goftlmux.MidBuffer, cookieValue, SandBoxPrefix, email string) {
	DbSetExpire(hdlr, rw, SandBoxKey(hdlr.Pre2Factor, SandBoxPrefix, cookieValue), email, hdlr.TwoFactorLife)
}

// ----------------------------------------------------------------------------------------------------------------------------
func GenBackupKeys(hdlr *AesSrpType, salt string, prefix string, www http.ResponseWriter, req *http.Request) (raw string, hash string) {
	com := ""
	raw = ""
	hash = ""
	for i := 0; i < 20; i++ {
		rn := GenerateRandomOneTimeKey(prefix)
		if db_GenBackupKeys {
			fmt.Printf("GenBackupKeys: rn [%s] %s\n", rn, godebug.LF())
		}
		hn := pbkdf2.Key([]byte(rn), []byte(salt), hdlr.BackupKeyIter, hdlr.BackupKeySizeBytes, sha256.New)
		raw = raw + com + rn
		hash = hash + com + fmt.Sprintf("%x", hn) // string(hn)
		com = ","
	}
	return
}

// ----------------------------------------------------------------------------------------------------------------------------
// Input
//	set 		the set of backup keys, hashed
//	salt		the users passwrod salt
//	to			key we are compareing to
// Output
//	hash 		new 'set' - after removing matched item if one found
//	found		true if a match was found
func CmpBackupKeys(hdlr *AesSrpType, salt string, set string, to string) (hash string, found bool, hv string) {
	v := strings.Split(set, ",")
	if len(v) == 0 {
		return
	}
	if to[0:1] != "9" {
		hash = set
		return
	}
	hash = ""
	com := ""
	// to_h := Sha256(salt + ":" + to[1:7])
	toHashedBuf := pbkdf2.Key([]byte(to[1:7]), []byte(salt), hdlr.BackupKeyIter, hdlr.BackupKeySizeBytes, sha256.New)
	toHashed := fmt.Sprintf("%x", toHashedBuf)
	for _, hn := range v {
		if 1 == subtle.ConstantTimeCompare([]byte(toHashed), []byte(hn)) { // if toHashed == hn {
			found = true
			hv = hn
		} else {
			hash = hash + com + hn
			com = ","
		}
	}
	return
}

// ============================================================================================================================================
func GetRwHdlrFromWWW(www http.ResponseWriter, req *http.Request) (rw *goftlmux.MidBuffer, hdlr *AesSrpType, ok bool) {

	rw, ok = www.(*goftlmux.MidBuffer)
	if !ok {
		AnError(hdlr, www, req, 500, 5, fmt.Sprintf("hdlr not correct type in rw.!, %s\n", godebug.LF()))
		return
	}

	hdlr, ok = rw.Hdlr.(*AesSrpType)
	if !ok {
		AnError(hdlr, www, req, 500, 5, fmt.Sprintf("hdlr not set in rw.!, %s\n", godebug.LF()))
		return
	}
	return
}

// Input username(email) fetch back the user informaiton - decrypt salt/v the verify value
func DbFetchUser(hdlr *AesSrpType, rw *goftlmux.MidBuffer, req *http.Request, username string, SandBoxPrefix string) (salt string, verif string, mdata map[string]string, err error) {

	ok := false
	mdata, ok = dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, username))
	if !ok {
		err = ErrNoSuchUser
		return
	}

	salt, verif = GetSalt(hdlr, rw, req, mdata)

	return
}

// Input username(email) fetch back the user information
func DbFetchUserMdata(hdlr *AesSrpType, rw *goftlmux.MidBuffer, username, SandBoxPrefix string) (mdata map[string]string, err error) {
	ok := false
	mdata, ok = dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, username))
	if !ok {
		err = ErrNoSuchUser
		return
	}
	return
}

// Validate an email address.   True if valid.
func ValidateEmail(email string) (rv bool) {
	rv = emailRe.MatchString(email)
	// fmt.Printf("ValidateEmail(%s) = %v, %s\n", email, rv, godebug.LF(2))
	return
}

// Validate a UUID as a string.  True if valid.
func ValidUUID(uuid string) (rv bool) {
	rv = isUUIDRe.MatchString(uuid)
	return
}

// Validate that the "salt" is likely to be long enough and that it is all hex digits.  Return true if valid.
func validSrpSalt(salt string) (rv bool) {
	// fmt.Printf("salt [%s] len(salt) = %d\n", salt, len(salt))
	if len(salt) <= 12 {
		return false
	}
	rv = hexRe.MatchString(salt)
	return
}

// Check that the srp 'v' value is long enough and all hex digits.
func validSrpV(v string) (rv bool) {
	// fmt.Printf("v [%s] len(v) = %d\n", v, len(v))
	if len(v) <= 200 {
		return false
	}
	//if ok, _ := regexp.MatchString("^[0-9a-fA-F]*$", v); !ok { // xyzzy2016 - performance - pre build regexp and save
	//	return false
	//}
	rv = hexRe.MatchString(v)
	return true
}

func SandBoxKey(pre, sandbox, key string) (rKey string) {
	rKey = pre + key
	if !SandBoxMode {
		return
	}
	if sandbox != "" {
		rKey = pre + sandbox + ":" + key
	}
	return
}

func (hdlr *AesSrpType) CookieEmailMatch(rw *goftlmux.MidBuffer, email, cookie, SandBoxPrefix string) bool {
	return true // <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< This !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// lookup cookie
	// if not found then false
	// if found then if mdata["owner_email"] == email - return true
	// else false
	if mdata, ok := dataStore.RGetValue(hdlr, rw, SandBoxKey("srp:U:", SandBoxPrefix, cookie)); !ok {
		if mdata["owner_email"] == email {
			return true
		}
	}
	return false
}

func SaveAsList(orig string, item string) (rv string) {
	if orig == "" {
		rv = item
	} else {
		ss := strings.Split(orig, ",")
		for _, vv := range ss {
			if vv == item {
				rv = orig
				return
			}
		}
		rv = orig + "," + item
	}
	return
}

func GenerateRandomDeviceID() (DeviceID string) {
	w, _ := GenRandNumber(7) //
	DeviceID = verhoeff_algorithm.GenerateVerhoeffString(w)
	return
}

// mdata["validation_secret"] = GenerateValidationSecret() //
func GenerateValidationSecret() (secret string) {
	w, _ := GenRandNumber(11) //
	// secret = verhoeff_algorithm.GenerateVerhoeffString(w)
	secret = w
	return
}

func GenerateRandomOneTimeKey(initialDigit string) (OneTimeKey string) {
	w, _ := GenRandNumber(8) //
	w = initialDigit + w[0:6]
	OneTimeKey = verhoeff_algorithm.GenerateVerhoeffString(w)
	return
}

func GenerateEmailAuthKey() (EmailAuthKey string) {
	w, _ := GenRandNumber(9) //
	EmailAuthKey = verhoeff_algorithm.GenerateVerhoeffString(w)
	return
}

func CheckMayAccessApi(hdlr *AesSrpType, rw *goftlmux.MidBuffer, SandBoxPrefix string, thePath, auth, acct_type string) (ok bool) {
	ok = false
	lookupKey := ""
	if _, ok := hdlr.SecurityConfig.MayAccessApi[auth+":"+acct_type]; ok {
		lookupKey = auth + ":" + acct_type
	} else if _, ok := hdlr.SecurityConfig.MayAccessApi[acct_type]; ok {
		lookupKey = acct_type
	} else if _, ok := hdlr.SecurityConfig.MayAccessApi[auth+":*"]; ok {
		lookupKey = auth + ":*"
	} else {
		fmt.Printf("At: %s\n", godebug.LF())
		return false
	}

	fmt.Printf("lookupKey=[%s] auth=%s acct_type=%s, data=%s, %s\n", lookupKey, auth, acct_type, godebug.SVarI(hdlr.SecurityConfig.MayAccessApi[lookupKey]), godebug.LF())

	if len(hdlr.SecurityConfig.MayAccessApi[lookupKey]) == 0 && auth == "y" { // If you haven't specified a list at all and your are looged in then - yes - all is allowed
		fmt.Printf("At: %s\n", godebug.LF())
		return true
	}
	fmt.Printf("At: %s\n", godebug.LF())
	if len(hdlr.SecurityConfig.MayAccessApi[lookupKey]) > 0 && (hdlr.SecurityConfig.MayAccessApi[lookupKey][0] == "*" ||
		godebug.InArrayString(thePath, hdlr.SecurityConfig.MayAccessApi[lookupKey]) >= 0) {
		fmt.Printf("At: %s\n", godebug.LF())
		ok = true
	}
	fmt.Printf("At: %s\n", godebug.LF())
	return
}

const db_GenBackupKeys = false
const db202 = false

/* vim: set noai ts=4 sw=4: */
