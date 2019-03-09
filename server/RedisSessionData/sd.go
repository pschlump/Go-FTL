package RedisSessionData

import (
	"encoding/json"
	"fmt"
	"os"

	logrus "github.com/pschlump/pslog" // "github.com/sirupsen/logrus"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/radix.v2/redis"
)

/*

going down

	SessionRedis
		Pull out $session$ from Redis
		Create redisSessionDataType in gftlmux.MidBuffer -- set with Redis conneciton info
			if rw, ok := www.(*goftlmux.MidBuffer); ok {
		Parse raw data
		Set 'ps' with relavant fields - not logged in

	TabServer2/Post Proces - from 'login' / 'recover-password-pt2' / 'register-confirm'
		Set IsLoggedIn <- true,
		Set email_address, real_name, auth_token etc in data
		'delta' of any -> 'ps'
		Chagne parsed data
		Set dirty

	TabServer2/Post Proces - from 'logout'
		Set IsLoggedIn <- false,
		Chagne parsed data
		Set dirty
		Flush to Redis

	Some-Other-Thing
		ps.InjectRule -- Rule for this data -- Owner of this data
		ps.Inject ... ( is user-protected data, is regular data )
			If value is in Session - then set in session
			Mark dirty


going up
	SessionRedis - Look chagned values in 'ps' ( anything that is Injected - not Default - and in Session )
		if chagned then set into local data.
		set dirty
	SessionRedis - Look for dirty - if dirty
		Flush to Redis





*** Multiple logins on multiple auth_tokens at the same time to 1 account ***


*/

type RuleType struct {
	Temporary  bool `json:"temporary"`
	UserMaySet bool `json:"user_may_set"`
}

type RedisSessionParse struct {
	IsLoggedIn  bool                `json:"$is_logged_in$"`
	UserData    map[string]string   `json:"UserData"`
	RegularData map[string]string   `json:"RegularData"`
	Rules       map[string]RuleType `json:"Rules"`
}

type RedisSessionDataType struct {
	RawData     string
	SessionData RedisSessionParse
	isDirty     bool
	Prefix      string
	Key         string
	gc          func() (*redis.Client, error)
	pc          func(*redis.Client)
}

func NewRedisSesionDataType() *RedisSessionDataType {
	return &RedisSessionDataType{
		RawData: "",
		isDirty: false,
		// IsLoggedIn: false,
		Prefix: "",
		Key:    "",
	}
}

func (sd *RedisSessionDataType) IsDirty() bool {
	return sd.isDirty
	// return true
}

// func (p *Pool) Get() (*redis.Client, error) {
// rw.Session.FlushSessionToRedis()
func (sd *RedisSessionDataType) FlushSessionToRedis() {

	key := sd.Prefix + sd.Key
	ttl := 60 * 10
	if sd.SessionData.IsLoggedIn {
		ttl = 60 * 60 * 24 * 1 // xyzzy <<<<<<<<<<<<<<<<<<<<< should change // xyzzy -- change experation from 30 min to 94 days
	}

	sd.RawData = godebug.SVar(sd.SessionData)

	value := sd.RawData

	fmt.Printf("\n%s--------------------------------------------------------------------------------------%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
	fmt.Printf("%sSaving to redis -->%s<--%s\n", MiscLib.ColorRed, value, MiscLib.ColorReset)
	fmt.Printf("%s--------------------------------------------------------------------------------------%s\n\n", MiscLib.ColorRed, MiscLib.ColorReset)

	fmt.Fprintf(os.Stderr, "\n%s>>>>>>>>> Saving to redis -->%s<--%s\n\n", MiscLib.ColorRed, value, MiscLib.ColorReset)

	conn, err := sd.gc()
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}
	defer sd.pc(conn)

	// set to expire in 10 min - if not used -
	err = conn.Cmd("SETEX", key, ttl, value).Err
	if err != nil {
		if db4 {
			fmt.Printf("Note on redis - session not found - get(%s): redisPrefix[%s] %s, %s\n", key, sd.Prefix, err, godebug.LF())
		}
		return
	}

	return
}

// rw.Session.GetFreeRedisConn(func() (conn redis.Connection, err error) {

func (sd *RedisSessionDataType) GetFreeRedisConn(GetConn func() (*redis.Client, error), PutConn func(*redis.Client)) *RedisSessionDataType {
	sd.gc = GetConn
	sd.pc = PutConn
	return sd
}

func (sd *RedisSessionDataType) SetPrefixKey(Prefix, Key string) *RedisSessionDataType {
	sd.Prefix = Prefix
	sd.Key = Key
	return sd
}

func (sd *RedisSessionDataType) SetDirty(b bool) *RedisSessionDataType {
	sd.isDirty = b
	return sd
}

//					session = rw.Session.CreateDefaultSession()
func (sd *RedisSessionDataType) CreateDefaultSession() (ss string) {
	sd.isDirty = true
	return GetDefaultSession()
}

// session = rw.Session.GetSessionFromRedis()
func (sd *RedisSessionDataType) GetSessionFromRedis() (ss string) {

	key := sd.Prefix + sd.Key

	conn, err := sd.gc()
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}
	defer sd.pc(conn)

	data, err := conn.Cmd("GET", key).Str()
	if err != nil || data == "" {
		fmt.Fprintf(os.Stderr, "%s***\nSession not found - get(%s): redisPrefix[%s] %s, %s%s\n", MiscLib.ColorRed, key, sd.Prefix, err, godebug.LF(), MiscLib.ColorReset)
		data = GetDefaultSession()
	}
	sd.RawData = data

	err = json.Unmarshal([]byte(data), &sd.SessionData)
	if err != nil {
		fmt.Printf("Note on redis -- did not parse err(%s): data[%s] %s, %s\n", err, data, godebug.LF())
		return
	}

	return data
}

func (sd *RedisSessionDataType) Login() {
	sd.isDirty = true
	sd.SessionData.IsLoggedIn = true
	// xyzzy -- copy "User" data -> 'ps'
}

func (sd *RedisSessionDataType) Logout() {
	sd.isDirty = true
	sd.SessionData.IsLoggedIn = false
	// xyzzy -- should logout delete the session?
}

//			x := rw.Session.DumpData()
func (sd *RedisSessionDataType) DumpData() (rv string) {
	rv = godebug.SVarI(sd.SessionData)
	return
}

func (sd *RedisSessionDataType) SetData(grp, name, value string) {
	sd.isDirty = true
	if grp == "regular" || grp == "reg" {
		if sd.SessionData.RegularData == nil {
			sd.SessionData.RegularData = make(map[string]string)
		}
		sd.SessionData.RegularData[name] = value
	} else if grp == "user" {
		if sd.SessionData.UserData == nil {
			sd.SessionData.UserData = make(map[string]string)
		}
		sd.SessionData.UserData[name] = value
	} else {
		fmt.Fprintf(os.Stderr, "Error: grp should be 'regular' or 'user', found [%s], %s\n", grp, godebug.LF())
	}
}

func (sd *RedisSessionDataType) GetData(grp, name string) (dv string, err error) {
	var ok bool
	if db10 {
		fmt.Printf("%sGetData: grp[%s] name[%s], sd.SessionData[%s], AT:%s, Called By%s %s\n", MiscLib.ColorYellow,
			grp, name, godebug.SVarI(sd.SessionData), godebug.LF(), godebug.LF(2), MiscLib.ColorReset)
	}
	if grp == "regular" || grp == "reg" {
		dv, ok = sd.SessionData.RegularData[name]
		if db10 {
			fmt.Printf("\tdv[%s] ok[%v] AT: %s\n", dv, ok, godebug.LF())
		}
	} else if grp == "user" {
		dv, ok = sd.SessionData.UserData[name]
		if db10 {
			fmt.Printf("\tdv[%s] ok[%v] AT: %s\n", dv, ok, godebug.LF())
		}
	} else {
		if db10 {
			fmt.Printf("\tAT: %s\n", godebug.LF())
		}
		err = fmt.Errorf("GetData: error - invalid group %s - should be 'regular' or 'user'", grp)
		return
	}
	if db10 {
		fmt.Printf("\tAT: %s\n", godebug.LF())
	}
	if !ok {
		if db10 {
			fmt.Printf("\tAT: %s\n", godebug.LF())
		}
		err = fmt.Errorf("GetData: error - %s not found", name)
		return
	}
	if db10 {
		fmt.Printf("\tAT: %s\n", godebug.LF())
	}
	return
}

func (sd *RedisSessionDataType) SetRule(name string, temp, user bool) {
	sd.isDirty = true
	if sd.SessionData.Rules == nil {
		sd.SessionData.Rules = make(map[string]RuleType)
	}
	sd.SessionData.Rules[name] = RuleType{Temporary: temp, UserMaySet: user}
}

func GetDefaultSession() string {
	return `{
	"$is_logged_in$": false
	,"UserData": { }
	,"RegularData": { }
	,"Rules": {
		"$is_logged_in$": { "temporary": false, "user_may_set": false } 
	}
}`
}

// GetFromRedis -- Use to pull data based on Prefix/Key -> RawData - then parse nto SessionData, set IsLoggedIn
// SaveToRedis -- if dirty then save back to redis
// IsLoggedIn -- Return true if user is currently logged in
// SetLoggedIn (t/f) -- Set the current state of is logged in
// GetValue -- ??
// GetSetValue ( user/regular, name, value )

/*
			// -----------------------------------------------------------------------------------------------------------------------------------
			// OLD!
			// -----------------------------------------------------------------------------------------------------------------------------------
			session := hdlr.RedisGetSession(www, rw, req, id)
			if session == "" {
				session = hdlr.RedisSetDefaultSession(www, rw, req, id)
			}
			goftlmux.AddValueToParams("$session$", session, 'i', goftlmux.FromInject, ps)

			// Xyzzy - if Logged In - then need to set
			// 		is_logged_in := ps.ByNameDflt("$is_logged_in$", "")
			// 		is_full_login := ps.ByNameDflt("$is_full_login$", "")
			if hdlr.SessionShowsLoggedIn(session, id) {
				goftlmux.AddValueToParams("$is_logged_in$", "yes", 'i', goftlmux.FromInject, ps)
				goftlmux.AddValueToParams("$is_full_login$", "no", 'i', goftlmux.FromInject, ps) // not a strong AesSrp login
			} else {
				goftlmux.AddValueToParams("$is_logged_in$", "no", 'i', goftlmux.FromInject, ps)
				goftlmux.AddValueToParams("$is_full_login$", "no", 'i', goftlmux.FromInject, ps) // not a strong AesSrp login
			}

//--	/ *
//--	   l_data = '{"status":"success","$send_email$":{'
//--	   		||'"template":"please_confirm_registration"'
//--	   		||',"username":'||to_json(p_username)
//--	   		||',"real_name":'||to_json(p_real_name)
//--	   		||',"email_token":'||to_json(l_email_token)
//--	   		||',"app":'||to_json(p_app)
//--	   		||',"name":'||to_json(p_name)
//--	   		||',"url":'||to_json(p_url)
//--	   		||',"from":'||to_json(l_from)
//--	   	||'},"$session$":{'
//--	   		||'"set":['
//--	   			||'{"path":["gen","auth"],"value":"y"}'
//--	   		||']'
//--	   	||'}}';
//--	* /
//--	func (hdlr *SessionRedis) SetSession(session, id, top, name, value string) {

//--		key := hdlr.RedisSessionPrefix + id

//--		mm := make(map[string]map[string]string)
//--		err := json.Unmarshal([]byte(session), &mm)
//--		if err != nil {
//--			logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to parse session [%s] [%s]. Reset to default!","LineFile":%q}`+"\n", err, id, session, godebug.LF()))
//--			hdlr.SetSessionDefault(id)
//--			return
//--		}
//--		_, ok := mm[top]
//--		if !ok {
//--			mm[top] = make(map[string]string)
//--		}
//--		mm[top][name] = value

//--		if db4 {
//--			fmt.Printf("SetSession: %s key= [%s], %s\n", godebug.LF(), key, godebug.LF())
//--		}

//--		conn, err := hdlr.gCfg.RedisPool.Get()
//--		defer hdlr.gCfg.RedisPool.Put(conn)
//--		if err != nil {
//--			logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
//--			return
//--		}

//--		session = godebug.SVar(mm)

//--		err = conn.Cmd("SET", key, session).Err
//--		if err != nil {
//--			if db4 {
//--				fmt.Printf("Note on redis - session not found - get(%s): redisPrefix[%s] %s, %s\n", key, hdlr.RedisSessionPrefix, err, godebug.LF())
//--			}
//--			return
//--		}

//--	}

func (hdlr *SessionRedis) SetSessionDefault(id string) {

	key := hdlr.RedisSessionPrefix + id

	value := `{
	"authReq": {
		  "user_id": ""
		, "username": ""
		, "email_address": ""
	}
	, "gen": {
		  "auth": "n"
		, "username": ""
		, "email_address": ""
		, "login_expire": ""
	}
}`

	if db4 {
		fmt.Printf("SetSessionDefault: %s key= [%s], %s\n", godebug.LF(), key, godebug.LF())
	}

	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	err = conn.Cmd("SET", key, value).Err
	if err != nil {
		if db4 {
			fmt.Printf("Note on redis - session not found - get(%s): redisPrefix[%s] %s, %s\n", key, hdlr.RedisSessionPrefix, err, godebug.LF())
		}
		return
	}

}

func (hdlr *SessionRedis) GetSession(session, id, top, name string) (value string) {
	mm := make(map[string]map[string]string)
	err := json.Unmarshal([]byte(session), &mm)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to parse session [%s] [%s]. Reset to default!","LineFile":%q}`+"\n", err, id, session, godebug.LF()))
		hdlr.SetSessionDefault(id)
		return
	}
	gen, ok := mm[top]
	if !ok {
		return
	}
	value, ok = gen[name]
	if !ok {
		return ""
	}
	return
}

func (hdlr *SessionRedis) SessionShowsLoggedIn(session, id string) (ok bool) {
	auth := hdlr.GetSession(session, id, "gen", "auth")
	return auth == "y"
}

//--	// if hdlr.ValidUser ( username, password ) {
//--	func (hdlr *SessionRedis) ValidUser(username string, password string) (ok bool) {


//--		key := "tmp_unpw:" + username + ":" + password

//--		if db4 {
//--			fmt.Printf("ValidUser: %s key= [%s], %s\n", godebug.LF(), key, godebug.LF())
//--		}

//--		conn, err := hdlr.gCfg.RedisPool.Get()
//--		defer hdlr.gCfg.RedisPool.Put(conn)
//--		if err != nil {
//--			logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
//--			return false
//--		}

//--		v, err := conn.Cmd("GET", key).Str()
//--		if err != nil {
//--			if db4 {
//--				fmt.Printf("Note on redis - user not found - get(%s): redisPrefix[%s] %s, %s\n", key, hdlr.RedisSessionPrefix, err, godebug.LF())
//--			}
//--			return false
//--		}

//--		if v == "" {
//--			return false
//--		}

//--		return true


//--	}

func (hdlr *SessionRedis) RedisGetSession(www http.ResponseWriter, rw *goftlmux.MidBuffer, req *http.Request, id string) (it string) {

	key := hdlr.RedisSessionPrefix + id

	if db4 {
		fmt.Printf("RedisGetSession: %s key= [%s], %s\n", godebug.LF(), key, godebug.LF())
	}

	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return ""
	}

	v, err := conn.Cmd("GET", key).Str()
	if err != nil {
		if db4 {
			fmt.Printf("Note on redis - session not found - get(%s): redisPrefix[%s] %s, %s\n", key, hdlr.RedisSessionPrefix, err, godebug.LF())
		}
		return ""
	}

	return v

}

func (hdlr *SessionRedis) RedisSetDefaultSession(www http.ResponseWriter, rw *goftlmux.MidBuffer, req *http.Request, id string) (it string) {

	key := hdlr.RedisSessionPrefix + id

	value := `{
	"authReq": {
		  "user_id": ""
		, "username": ""
		, "email_address": ""
	}
	, "gen": {
		  "auth": "n"
		, "username": ""
		, "email_address": ""
		, "login_expire": ""
	}
}`

	if db4 {
		fmt.Printf("RedisSetDefaultSession: %s key= [%s], %s\n", godebug.LF(), key, godebug.LF())
	}

	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	err = conn.Cmd("SET", key, value).Err
	if err != nil {
		if db4 {
			fmt.Printf("Note on redis - session not found - get(%s): redisPrefix[%s] %s, %s\n", key, hdlr.RedisSessionPrefix, err, godebug.LF())
		}
		return
	}

	return value

}


*/

const db4 = false
const db10 = false

/* vim: set noai ts=4 sw=4: */
