//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2017.
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1003
//

package cfg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/Sirupsen/logrus"
	_ "github.com/lib/pq"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/sizlib" //
	JsonX "github.com/pschlump/JSONx"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"       //
	"github.com/pschlump/radix.v2/pool" // Modified pool to have NewAuth for authorized connections
)

var ServerName = "Go-FTL"
var ServerType = "Go-FTL"
var Version = "0.5.9"
var BuildNo = "1811"

var ServersMutex sync.Mutex

var Wg sync.WaitGroup

// ----------------------------------------------------------------------------------------------------------------------------

type LoggingConfigType struct {
	FileOn  string
	RedisOn string
}

// ---- ServerGlobalConfigType -------------------------------------------------------------------------------------------------

type ServerGlobalConfigType struct {
	ServerName       string                         `gfJsonX:"server_name" gfDefault:"Go-FTL Server"`  //
	DebugFlags       []string                       `gfJsonX:"debug_flags"`                            //
	TraceFlags       []string                       `gfJsonX:"trace_flags"`                            //
	DefaultStatic    string                         `gfJsonX:"default_static" gfDefault:"./static"`    //
	RedisConnectHost string                         `gfJsonX:"RedisConnectHost" gfDefault:"127.0.0.1"` // Connection info for Redis Database
	RedisConnectPort string                         `gfJsonX:"RedisConnectPort" gfDefault:"6379"`      //
	RedisConnectAuth string                         `gfJsonX:"RedisConnectAuth"`                       //
	PGConn           string                         `gfJsonX:"PGConn"`                                 //
	DBType           string                         `gfJsonX:"DBType" gfDefault:"postgres"`            //
	DBName           string                         `gfJsonX:"DBName" gfDefault:"pschlump"`            //
	LoggingConfig    LoggingConfigType              `gfJsonX:"LoggingConfig"`                          //
	RedisPool        *pool.Pool                     `gfJsonX:"-"`                                      // Pooled Redis Client connection information
	mutex            sync.Mutex                     //                                                 // Lock for Redis
	Pg_client        *sizlib.MyDb                   `gfJsonX:"-"` //                                   // Client connection for PostgreSQL
	connected        string                         //                                                 // "ok" when connected to Redis, "err" if connection failed.  - 2-state flag.  (TODO: convert to a const/int)
	connected_rd     string                         //                                                 // "ok" when connected to relational database, "err" if connection failed.
	Config           map[string]PerServerConfigType //                                                 //Anything that did not match the above JSON names //
	pDebugFlags      map[string]bool
}

var ServerGlobal ServerGlobalConfigType

func ResolvLocalFile(fn string) (outFn string) {
	outFn = fn
	host := os.Getenv("HOST")
	home := os.Getenv("HOME")
	if host == "" || home == "" {
		return
	}
	lookFor := filepath.Join(home, "Local", host+"__"+fn)
	if sizlib.Exists(lookFor) {
		fmt.Printf("%sUsing [%s] for the [%s] file - local configuration file%s\n", MiscLib.ColorGreen, lookFor, fn, MiscLib.ColorReset)
		outFn = lookFor
	}
	return
}

func ReadGlobalConfigFile(fn string) {

	if false {

		data, err := ioutil.ReadFile(fn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to open %s, err=%s\n", fn, err)
			os.Exit(1)
		}
		err = json.Unmarshal(data, &ServerGlobal)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to parse %s, err=%s\n", fn, err)
			fmt.Fprintf(os.Stderr, "->%s<-\n", data)
			os.Exit(1)
		}

	} else {

		meta, err := JsonX.UnmarshalFile(fn, &ServerGlobal)
		_ = meta
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: Error returned from JsonX.UnmarshalFile: %s, %s\n", "Go-FTL", err, godebug.SVarI(meta))
			logrus.Errorf("Error: Invalid JsonX for %s Error:\n%s\n", fn, err)
			lib.IsErrFatal(err)
			panic("wow")
			os.Exit(1)
		}
	}

	if db_g2 {
		fmt.Fprintf(os.Stderr, "\n\n>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> File %s: Glboal Config: %s\n\n", fn, godebug.SVarI(ServerGlobal))
		fmt.Printf("\n\n>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> File %s: Glboal Config: %s\n\n", fn, godebug.SVarI(ServerGlobal))
	}
}

// ---- Configuration Files ----------------------------------------------------------------------------------------------------
type ListenToType struct {
	Protocal string // http or https
	Port     string // 3000
	Domain   string // localhost, www.test1.com etc. (the IP address)
	HasWild  bool   // True if *.test1.com
}

type PerServerConfigType_old struct {
	Name           string                 `json:"-"`          // name for this server -- PerServerConfigType/cfg.go:109
	LineNo         int                    `json:"-"`          // Start line number for this config
	FileName       string                 `json:"-"`          // File name this config came from
	ListenTo       []string               `json:"listen_to"`  // URL to listen to
	Certs          []string               `json:"certs"`      // Certs if https, wss
	Port           []string               `json:"-"`          // port for URL
	ListenToParsed []ListenToType         `json:"-"`          // parsed url
	Plugins        interface{}            `json:"-"`          //
	ConfigData     map[string]interface{} `json:"ConfigData"` // Any other config info
}

type PerServerConfigType struct {
	Name           string                 `gfJsonX:"-"`          // name for this server -- PerServerConfigType/cfg.go:109
	LineNo         int                    `gfJsonX:"-"`          // Start line number for this config
	FileName       string                 `gfJsonX:"-"`          // File name this config came from
	ListenTo       []string               `gfJsonX:"listen_to"`  // URL to listen to
	Certs          []string               `gfJsonX:"certs"`      // Certs if https, wss
	Port           []string               `gfJsonX:"-"`          // port for URL
	ListenToParsed []ListenToType         `gfJsonX:"-"`          // parsed url
	Plugins        interface{}            `gfJsonX:"-"`          //
	ConfigData     map[string]interface{} `gfJsonX:"ConfigData"` // Any other config info
}

// ----------------------------------------------------------------------------------------------------------------------------
func (sgct *ServerGlobalConfigType) GetKeys(theKey string) []string {
	conn, err := sgct.RedisPool.Get()
	defer sgct.RedisPool.Put(conn)
	if err != nil {
		// goftlmux.G_Log.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return []string{}
	}
	kks, err := conn.Cmd("KEYS", theKey).List()
	if err != nil {
		fmt.Printf("Error(10095): %v, %s\n", err, godebug.LF())
		return []string{}
	}
	return kks
}

// ----------------------------------------------------------------------------------------------------------------------------
func (sgct *ServerGlobalConfigType) ConnectToRedis() bool {
	// Note: best test for this is in the TabServer2 - test 0001 - checks that this works.
	var err error

	sgct.mutex.Lock()
	defer sgct.mutex.Unlock()

	if sgct.connected == "ok" {
		return true
	}

	dflt := func(a string, d string) (rv string) {
		rv = a
		if rv == "" {
			rv = d
		}
		return
	}

	redis_host := dflt(sgct.RedisConnectHost, "127.0.0.1")
	redis_port := dflt(sgct.RedisConnectPort, "6379")
	redis_auth := sgct.RedisConnectAuth

	if redis_auth == "" { // If Redis AUTH section
		sgct.RedisPool, err = pool.New("tcp", redis_host+":"+redis_port, 20)
	} else {
		sgct.RedisPool, err = pool.NewAuth("tcp", redis_host+":"+redis_port, 20, redis_auth)
	}
	if err != nil {
		sgct.connected = "err"
		fmt.Fprintf(os.Stderr, "%sError: Failed to connect to redis-server.%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
		fmt.Printf("Error: Failed to connect to redis-server.\n")
		// goftlmux.G_Log.Info("Error: Failed to connect to redis-server.\n")
		logrus.Fatalf("Error: Failed to connect to redis-server.\n")
		return false
	} else {
		if db11 {
			fmt.Fprintf(os.Stderr, "%sSuccess: Connected to redis-server.%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
		}
		sgct.connected = "ok"
	}

	return true
}

// ----------------------------------------------------------------------------------------------------------------

type PostDbType struct {
	RunCheck func(conn *sizlib.MyDb) bool
}

var PostDbConnectChecks []PostDbType

// ConnectToPostgreSQL will take global connection information and connect to the database.
func (sgct *ServerGlobalConfigType) ConnectToPostgreSQL() bool {
	// Note: best test for this is in the TabServer2 - test 0001 - checks that this works.
	var err error

	sgct.mutex.Lock()
	defer sgct.mutex.Unlock()

	if sgct.connected_rd == "ok" {
		return true
	}

	conn := sizlib.ConnectToAnyDb(sgct.DBType, sgct.PGConn, sgct.DBName)
	if conn == nil {
		fmt.Fprintf(os.Stdout, "Unable to establish connection to database: %v\n", err)
		fmt.Fprintf(os.Stderr, "%sUnable to establish connection to database: %v%s\n", MiscLib.ColorRed, err, MiscLib.ColorReset)
		sgct.connected_rd = "err"
		return false
	}

	if db11 {
		fmt.Fprintf(os.Stderr, "%sSuccess: Connected to PostgreSQL-server.%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
	}

	sgct.Pg_client = conn
	sgct.connected_rd = "ok"

	ok := true
	for _, vv := range PostDbConnectChecks {
		tok := vv.RunCheck(conn)
		if !tok {
			ok = false
		}
	}
	if !ok {
		fmt.Fprintf(os.Stderr, "%sWarning: Failed DB Check - invalid configuration - some endpoints will not work%s\n", MiscLib.ColorYellow, MiscLib.ColorReset)
		fmt.Printf("Warning: Failed DB Check - invalid configuration - some endpoints will not work\n")
	} else {
		fmt.Fprintf(os.Stderr, "%sDB Check - valid configuration - tables/columns match database%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
		fmt.Printf("DB Check - valid configuration - tables/columns match database\n")
	}

	return true

}

// ----------------------------------------------------------------------------------------------------------------

var redis_conn_setup = false

func SetupRedisForTest(test_cfg string) bool {

	if redis_conn_setup {
		return true
	}
	redis_conn_setup = true

	meta, err := JsonX.UnmarshalFile(test_cfg, &ServerGlobal)
	_ = meta
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: Error returned from JsonX.UnmarshalFile: %s\n", "Go-FTL", err)
		logrus.Errorf("Error: Invalid JsonX for %s Error:\n%s\n", test_cfg, err)
		lib.IsErrFatal(err)
		os.Exit(1)
	}

	return ServerGlobal.ConnectToRedis()
}

// ----------------------------------------------------------------------------------------------------------------

var pg_conn_setup = false

func SetupPgSqlForTest(test_cfg string) bool {

	if pg_conn_setup {
		return true
	}
	pg_conn_setup = true

	meta, err := JsonX.UnmarshalFile(test_cfg, &ServerGlobal)
	_ = meta
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: Error returned from JsonX.UnmarshalFile: %s\n", "Go-FTL", err)
		logrus.Errorf("Error: Invalid JsonX for %s Error:\n%s\n", test_cfg, err)
		lib.IsErrFatal(err)
		os.Exit(1)
	}

	x := ServerGlobal.ConnectToPostgreSQL()
	if x {
		key := "*"
		// rows, err := ServerGlobal.Pg_client.Db.Query("select \"salt\", \"password\" from \"basic_auth\" where \"username\" = $1", key)
		rows, err := ServerGlobal.Pg_client.Db.Query("select 'a' \"salt\", 'b' \"password\" where $1 <> $1", key)
		if err != nil {
			fmt.Printf("Error: Database error %s, attempting to connect to database\n", err)
			return false
		}
		for nr := 0; rows.Next(); nr++ {
			fmt.Printf("Error: Database error got data back when we should not get data back. Error=%s\n", err)
			_ = nr
		}
	}
	return true
}

// ----------------------------------------------------------------------------------------------------------------

func SetupEmptyForTest() bool {
	return true
}

// ----------------------------------------------------------------------------------------------------------------

var ReservedItems = map[string]bool{
	"$auth_key$":                  true,
	"$email$":                     true,
	"$is_logged_in$":              true,
	"$is_enc_logged_in$":          true,
	"$$host_name$$":               true,
	"$is_full_login$":             true,
	"$privs$":                     true,
	"$saved_one_time_key_hashed$": true,
	"$user_id$":                   true,
	"$customer_id$":               true,
	"$username$":                  true,
	"$session$":                   true,
	"$ip_sha256$":                 true,
}

//------------------------------------------------------------------------------------------------------------------------

// Return true if the debuging flag is enabled for this set of module/server/flag
//
// Example: pass "godebug.FILE()" for the module, "localhost" for the server and "db_CORS_login" for the flag.
//   the enabled flag is "CORS/localhost:.*/db_CORS.*
//	This should result in a "true" return.
// 1. Get MiddlwareName form FileName
// 2. Match of Flag
//
// ExampleFlag:
// 		http://localhost:8088/SessionRedis/db1
//		Server: http://localhost:8088
//		Module: SessionRedis
//		Flag: db1
func (sgct *ServerGlobalConfigType) DbSetup() {
	if len(sgct.DebugFlags) > 0 && (sgct.pDebugFlags == nil || len(sgct.pDebugFlags) == 0) {
		fmt.Printf("\nCreateing the parsed debug flags\n======================================================\n")
		sgct.pDebugFlags = make(map[string]bool)
		for _, db := range sgct.DebugFlags {
			sgct.pDebugFlags[db] = true
		}
	}
}

func (sgct *ServerGlobalConfigType) DbOn(server, module, flag string) bool {
	var ServerGlobal *ServerGlobalConfigType
	if sgct == nil {
		sgct = ServerGlobal
		if sgct == nil {
			fmt.Printf("Call of DbOn before globals setup - assuming %v, CallAT: %s, change db_DbOn global constant to set in ./cfg/cfg.go\n", db_DbOn, godebug.LF(2))
			// fmt.Fprintf(os.Stderr, "Call of DbOn before globals setup - assuming %v, CallAT: %s\n", db_DbOn, godebug.LF(2))
			return db_DbOn
		}
	}
	sgct.DbSetup()
	return (sgct.pDebugFlags[server+"/"+module+"/"+flag])
}

const db_DbOn = false
const db_g1 = false
const db_g2 = false
const db11 = true

/* vim: set noai ts=4 sw=4: */
