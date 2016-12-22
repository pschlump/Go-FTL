//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1003
//

package cfg

import (
	"time"

	_ "github.com/lib/pq"

	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/sizlib" //
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/check-json-syntax/lib"
	"github.com/pschlump/godebug"       //
	"github.com/pschlump/json"          //	"encoding/json"
	"github.com/pschlump/radix.v2/pool" // Modified pool to have NewAuth for authorized connections
)

// "github.com/jackc/pgx" //

var ServerName = "Go-FTL"
var ServerType = "Go-FTL"
var Version = "0.5.9"
var BuildNo = "1811"

var ServersMutex sync.Mutex

var Wg sync.WaitGroup

// ----------------------------------------------------------------------------------------------------------------------------
type InitNextFx func(next http.Handler, gCfg *ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error)
type OneTimeInitFx func(h interface{}, cfgData map[string]interface{}, callNo int) error
type CreateEmptyFx func() interface{}

type NewInitType struct {
	Name            string        // Name of this (the directive this is called by
	FinalizeHandler InitNextFx    // Take the data and finaialize the hnadler
	OneTimeInit     OneTimeInitFx // One time Init
	CreateEmpty     CreateEmptyFx // Creates an empty configuration structure of the correct type and returns it.
	ValidJSON       string        // JSONP validaiton string for config for this item
	CfgData         interface{}   //
	CallNo          int           //
}

var NewInit []*NewInitType

func RegInitItem2(name string, nx InitNextFx, ce CreateEmptyFx, ot OneTimeInitFx, valid string) {
	NewInit = append(NewInit, &NewInitType{Name: name, FinalizeHandler: nx, ValidJSON: valid, OneTimeInit: ot, CreateEmpty: ce})
}

func LookupInitByName(name string) (p int) {
	p = -1
	for ii := range NewInit {
		if NewInit[ii].Name == name {
			return ii
		}
	}
	return
}

type LoggingConfigType struct {
	FileOn  string
	RedisOn string
}

// ---- ServerGlobalConfigType -------------------------------------------------------------------------------------------------
type ServerGlobalConfigType struct {
	ServerName       string                         `json:"server_name"`      //
	DebugFlags       []string                       `json:"debug_flags"`      //
	TraceFlags       []string                       `json:"trace_flags"`      //
	DefaultStatic    string                         `json:"default_static"`   //
	RedisConnectHost string                         `json:"RedisConnectHost"` // Connection infor for Redis Database
	RedisConnectPort string                         `json:"RedisConnectPort"` //
	RedisConnectAuth string                         `json:"RedisConnectAuth"` //
	PGConn           string                         `json:"PGConn"`           //
	DBType           string                         `json:"DBType"`           //
	DBName           string                         `json:"DBName"`           //
	LoggingConfig    LoggingConfigType              `json:"LoggingConfig"`    //
	RedisPool        *pool.Pool                     `json:"-"`                // Pooled Redis Client connectioninformation
	mutex            sync.Mutex                     //                        // Lock for redis
	Pg_client        *sizlib.MyDb                   `json:"-"` // Client connection for PostgreSQL
	connected        string                         //                        // "ok" when connected to redis, "err" if connection failed.  - 2-state flag.  (TODO: convert to a const/int)
	connected_rd     string                         //                        // "ok" when connected to relational database, "err" if connection failed.
	Config           map[string]PerServerConfigType //                        //	                       // Anything that did not match the abobve JSON names //
}

//Pg_client        *pgx.Conn                      `json:"-"`                // Client connection for PostgreSQL

var ServerGlobal *ServerGlobalConfigType

func NewServerGlobalConfigType() *ServerGlobalConfigType {
	return &ServerGlobalConfigType{
		ServerName:       ServerName + "(" + Version + " BuildNo " + BuildNo + ")",
		DefaultStatic:    "./static",
		Config:           make(map[string]PerServerConfigType),
		RedisConnectHost: "127.0.0.1",
		RedisConnectPort: "6379",
		RedisConnectAuth: "",
		// ACMEEmail:        "pschlump@gmail.com",      // xyzzy - That's me for the moment
		// ACMEServer:       "https://localhost:5672/", // xyzzy - Boulder running on local server for now :5672, :5673
		// LogDirectory:     "./log",
	}
}

func ReadGlobalConfigFile(fn string) {
	// file, err := ioutil.ReadFile(fn)
	file, err := sizlib.ReadJSONDataWithComments(fn)
	lib.IsErrFatal(err)
	if db_g1 {
		fmt.Printf("File:%s data:%s\n", fn, file)
	}
	ServerGlobal = NewServerGlobalConfigType()

	err = json.Unmarshal(file, &ServerGlobal)
	if err != nil {
		es := jsonSyntaxErroLib.GenerateSyntaxError(string(file), err)
		fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorYellow, es, MiscLib.ColorReset)
		logrus.Errorf("Error: Invlaid JSON for %s %s Error:\n%s\n", fn, file, es)
		lib.IsErrFatal(err)
	}

	if db_g1 {
		fmt.Printf("\n\n>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> Glboal Config: %+v\n\n", ServerGlobal)
	}
}

// ---- Caching Config Type ----------------------------------------------------------------------------------------------------
type CacheConfigType struct {
	CacheForSeconds             int       // if 0, then not applicable, 1 = cache till end of 1 second up, 2..n do not refresh until timeout
	FetchedTime                 time.Time // When was data fetched
	ProxiedData                 bool      // Data is from a proxy, false implies source files local and can be re-checked
	CacheAndRecheckDependencies bool      // Cache it but re-checked dependencies
	OutputFile                  []string  // full path to output
	IntermediateFile            []string  // set of files that represent intermediate files
	InputFile                   []string  // set of files that represent input - timestamps can be checked
	CacheAndRevalidate          bool      // Cache - but re-generate source and see if SHA256 is same, if so then 304 else re-send
	Sha256Hash                  string    // Hash of output data
	IgnoreTotally               bool      // Not catchable at all
	CacheIfLargerThan           uint64    // Ignore if data size is less than this
	IgnoreIfLargerThan          uint64    // Ignore if data size is bigger than this
	CachePaths                  []string  // paths to Cache
	IgnorePaths                 []string  // paths to ignore
	IgnoreCookies               []string  // paths to ignore
	MatchUrl_Cookies            []string  // Add these cookies to URL before a lookup
	Prefetch                    bool      // Pre fetch indicates that catch pre-fetching should occurs on this item
	PrefetchCount               int       // Pre fetch this number of items
	PrefetchFreq                int       // Time for pre-fetch - how often
	StaleAfter                  int       // Delta-T for item in pre-fetch going stale (shelf-life)
	FlushFromCache              bool      // Indicates that a lower level knows that this should be flushed from the cache
}

// ---- Configuration Files ----------------------------------------------------------------------------------------------------
type ListenToType struct {
	Protocal string // http or https
	Port     string // 3000
	Domain   string // localhost, www.test1.com etc. (the IP address)
	HasWild  bool   // True if *.test1.com
}

type PerServerConfigType struct {
	Name           string                 `json:"-"`           // name for this server
	LineNo         int                    `json:"-"`           // Start line number for this config
	FileName       string                 `json:"-"`           // File name this config came from
	ListenTo       []string               `json:"listen_to"`   // URL to listen to
	Certs          []string               `json:"certs"`       // Certs if https, wss
	ACMEEmail      string                 `json:"ACME_email"`  // Let's Encrypt username (email address)
	ACMEServer     string                 `json:"ACME_server"` // Let's Encrypt server
	Port           []string               `json:"-"`           // port for URL
	ListenToParsed []ListenToType         `json:"-"`           // parsed url
	Plugins        interface{}            `json:"-"`           //
	ConfigData     map[string]interface{} `json:"ConfigData"`  // Any other config info
}

func ReadConfigFile2(fn string) {
	// Note: best test for this is in the TabServer2 - test 0001 - checks that this works.

	file, err := sizlib.ReadJSONDataWithComments(fn)
	lib.IsErrFatal(err)

	////fmt.Printf("At: %s\n", lib.LF())
	RawConfig := make(map[string]interface{})
	err = json.Unmarshal(file, &RawConfig)
	if err != nil {
		es := jsonSyntaxErroLib.GenerateSyntaxError(string(file), err)
		fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorYellow, es, MiscLib.ColorReset)
		logrus.Errorf("Error: Invlaid JSON for %s %s Error:\n%s\n", fn, file, es)
		lib.IsErrFatal(err)
	}

	if ServerGlobal.Config == nil {
		ServerGlobal.Config = make(map[string]PerServerConfigType)
	}

	for name, v := range RawConfig {
		//fmt.Printf("At: %s\n", lib.LF())
		vv := v.(map[string]interface{}) // vv is a map[string]interface{}
		if db_g1 {
			fmt.Printf("Configuration for >%s< typeof vv = %T\n", name, vv)
		}
		// perServerConfig := PerServerConfigType{IndexFileList: []string{"index.html", "index.tmpl"}}
		perServerConfig := PerServerConfigType{}
		perServerConfig.Name = name
		LineNoF, ok := vv["LineNo"]
		LineNo := int(LineNoF.(float64))
		if !ok {
			if db_g1 {
				fmt.Printf("Missing LineNo from config\n")
			}
			LineNo = 1
		}

		// LineNo + FileName -----------------------------------------------------------------------------
		//fmt.Printf("At: %s\n", lib.LF())
		perServerConfig.LineNo = 1
		if tt, ok := vv["LineNo"]; ok {
			perServerConfig.LineNo = int(tt.(float64))
			LineNo = perServerConfig.LineNo
			delete(vv, "LineNo")
		}
		perServerConfig.FileName = fn
		if tt, ok := vv["FileName"]; ok {
			perServerConfig.FileName = tt.(string)
			delete(vv, "FileName")
		}

		// ConfigData -----------------------------------------------------------------------------
		if tt, ok := vv["ConfigData"]; ok {
			x, ok := tt.(map[string]interface{})
			if ok {
				perServerConfig.ConfigData = x
			} else {
				fmt.Printf("LineNo:%d Invalid type for ConfigData, got %T, need 'map[string]interface{}', %s\n", LineNo, tt, godebug.LF())
			}
		}

		// listen_to -----------------------------------------------------------------------------
		//fmt.Printf("At: %s\n", lib.LF())
		if tt, ok := vv["listen_to"]; ok {
			if db_g1 {
				fmt.Printf("\tlisten_to typeof = %T, %+v\n", tt, tt)
			}
			if ss, yep := tt.(string); yep {
				perServerConfig.ListenTo = append(perServerConfig.ListenTo, ss)
			} else {
				// xyzzy - check type as array
				for _, ww := range tt.([]interface{}) {
					perServerConfig.ListenTo = append(perServerConfig.ListenTo, ww.(string))
				}
			}
			if db_g1 {
				fmt.Printf("\tperServerConfig.ListenTo = %v\n", perServerConfig.ListenTo)
			}
			delete(vv, "listen_to")
		} else {
			if lib.IsProtocal(name) {
				perServerConfig.ListenTo = append(perServerConfig.ListenTo, name)
			} else {
				fmt.Printf("LineNo:%d A server must have a 'listen_to' value or it will not serve to anybody\n", LineNo)
			}
		}

		//fmt.Printf("At: %s\n", lib.LF())
		// certs -----------------------------------------------------------------------------
		if tt, ok := vv["certs"]; ok {
			if db_g1 {
				fmt.Printf("\tcerts typeof = %T, %+v\n", tt, tt)
			}
			// xyzzy - check type as array
			for _, ww := range tt.([]interface{}) {
				perServerConfig.Certs = append(perServerConfig.Certs, ww.(string))
			}
			if db_g1 {
				fmt.Printf("\tperServerConfig.Certs = %v\n", perServerConfig.Certs)
			}
			delete(vv, "certs")
		}

		// plugins -----------------------------------------------------------------------------
		//fmt.Printf("At: %s\n", lib.LF())
		if tt, ok := vv["plugins"]; ok {
			// Iterate over the array of plugins
			for ii, ww := range tt.([]interface{}) {
				// Get the name of this plugin
				//fmt.Printf("At: %s\n", lib.LF())
				wwt, ok := ww.(map[string]interface{})
				if !ok {
					fmt.Printf("Syntax Error: Line:%d Invalid data for plugin configuration (on %d'th plugin)\n", LineNo, ii)
				} else if lib.LenOfMap(wwt) != 1 {
					fmt.Printf("Syntax Error: Line:%d Invalid specification of options for a plugin (on %d'th plugin)\n", LineNo, ii)
				} else {
					//fmt.Printf("At: %s, wwt=%s\n", lib.LF(), lib.SVarI(wwt))
					nameOfPlugin := lib.FirstName(wwt)
					//fmt.Printf("At: %s, nameOfPlugin=%s\n", lib.LF(), nameOfPlugin)
					locInTab := LookupInitByName(nameOfPlugin)
					if db_g1 {
						fmt.Printf("nameOfPlugin: %s at %d in init table, %s\n", nameOfPlugin, locInTab, lib.LF())
					}
					if locInTab < 0 {
						//fmt.Printf("At: %s\n", lib.LF())
						fmt.Printf("Syntax Error: Line:%d Unknown plugin %s (on %d'th plugin)\n", LineNo, nameOfPlugin, ii)
					} else {
						//fmt.Printf("At: %s\n", lib.LF())
					}
				}
			}
		}
		// perServerConfig.Plugins = vv["plugins"].([]map[string]interface{})
		//fmt.Printf("At: %s\n", lib.LF())
		perServerConfig.Plugins = vv["plugins"]
		//fmt.Printf("At: %s\n", lib.LF())
		ServerGlobal.Config[name] = perServerConfig
		//fmt.Printf("At: %s\n", lib.LF())
	}
}

// ----------------------------------------------------------------------------------------------------------------------------
func (hdlr *ServerGlobalConfigType) GetKeys(theKey string) []string {
	conn, err := hdlr.RedisPool.Get()
	defer hdlr.RedisPool.Put(conn)
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
func (hdlr *ServerGlobalConfigType) ConnectToRedis() bool {
	// Note: best test for this is in the TabServer2 - test 0001 - checks that this works.
	var err error

	hdlr.mutex.Lock()
	defer hdlr.mutex.Unlock()

	if hdlr.connected == "ok" {
		return true
	}

	dflt := func(a string, d string) (rv string) {
		rv = a
		if rv == "" {
			rv = d
		}
		return
	}

	redis_host := dflt(hdlr.RedisConnectHost, "127.0.0.1")
	redis_port := dflt(hdlr.RedisConnectPort, "6379")
	redis_auth := hdlr.RedisConnectAuth

	if redis_auth == "" { // If Redis AUTH section
		hdlr.RedisPool, err = pool.New("tcp", redis_host+":"+redis_port, 20)
	} else {
		hdlr.RedisPool, err = pool.NewAuth("tcp", redis_host+":"+redis_port, 20, redis_auth)
	}
	if err != nil {
		hdlr.connected = "err"
		fmt.Fprintf(os.Stderr, "%sError: Failed to connect to redis-server.%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
		fmt.Printf("Error: Failed to connect to redis-server.\n")
		// goftlmux.G_Log.Info("Error: Failed to connect to redis-server.\n")
		logrus.Fatalf("Error: Failed to connect to redis-server.\n")
		return false
	} else {
		if db11 {
			fmt.Fprintf(os.Stderr, "%sSuccess: Connected to redis-server.%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
		}
		hdlr.connected = "ok"
	}

	return true
}

func (hdlr *ServerGlobalConfigType) ConnectToPostgreSQL() bool {
	// Note: best test for this is in the TabServer2 - test 0001 - checks that this works.
	var err error

	hdlr.mutex.Lock()
	defer hdlr.mutex.Unlock()

	if hdlr.connected_rd == "ok" {
		return true
	}

	// fmt.Printf("AT: %s\n", godebug.LF())
	//:pgx:conn, err := pgx.Connect(hdlr.extractConfig())
	//:pgx:if err != nil {
	//:pgx: conn = "PGConn": "127.0.0.1:5433:pschlump:f1ref0x2:pschlump"
	conn := sizlib.ConnectToAnyDb(hdlr.DBType, hdlr.PGConn, hdlr.DBName)
	if conn == nil {
		fmt.Fprintf(os.Stdout, "Unable to connection to database: %v\n", err)
		fmt.Fprintf(os.Stderr, "%sUnable to connection to database: %v%s\n", MiscLib.ColorRed, err, MiscLib.ColorReset)
		hdlr.connected_rd = "err"
		return false
	}

	if db11 {
		fmt.Fprintf(os.Stderr, "%sSuccess: Connected to PostgreSQL-server.%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
	}

	hdlr.Pg_client = conn
	hdlr.connected_rd = "ok"

	// xyzzyPostDb Checks -- xyzzy - at this point --
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

//:pgx:func (hdlr *ServerGlobalConfigType) extractConfig() (config pgx.ConnConfig) {
//:pgx:
//:pgx:	dflt := func(s, t string) (r string) {
//:pgx:		r = s
//:pgx:		if s == "" {
//:pgx:			r = t
//:pgx:		}
//:pgx:		return
//:pgx:	}
//:pgx:
//:pgx:	// host:user:pass:db
//:pgx:	t := strings.Split(hdlr.PGConn, ":")
//:pgx:	if len(t) != 5 {
//:pgx:		fmt.Printf("Invalid confuration should have Postgres Connect string of host:port:user:pass:db\n")
//:pgx:		fmt.Printf("  Host default 127.0.0.1\n")
//:pgx:		fmt.Printf("  Port default 5432\n")
//:pgx:		fmt.Printf("  User default test\n")
//:pgx:		fmt.Printf("  Password default password\n")
//:pgx:		fmt.Printf("  Database default test\n")
//:pgx:		fmt.Fprintf(os.Stderr, "%sInvalid confuration should have Postgres Connect string of host:port:user:pass:db%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
//:pgx:		os.Exit(1)
//:pgx:	}
//:pgx:	config.Host = dflt(t[0], "127.0.0.1")
//:pgx:	p := 5432
//:pgx:	if t[1] != "" {
//:pgx:		x, err := strconv.ParseInt(t[1], 10, 32)
//:pgx:		if err != nil {
//:pgx:			fmt.Printf("invalid port in connection string: %s\n", err)
//:pgx:			p = 5432
//:pgx:		} else {
//:pgx:			p = int(x)
//:pgx:		}
//:pgx:	}
//:pgx:	config.Port = uint16(p)
//:pgx:	config.User = dflt(t[2], "test")
//:pgx:	config.Password = dflt(t[3], "password")
//:pgx:	config.Database = dflt(t[4], "test")
//:pgx:	return
//:pgx:}

const db11 = true

var redis_conn_setup = false
var pg_conn_setup = false

func SetupRedisForTest(redis_cfg_file string) bool {

	if redis_conn_setup {
		return true
	}
	redis_conn_setup = true

	if ServerGlobal == nil {
		ServerGlobal = NewServerGlobalConfigType()
	}

	s, err := sizlib.ReadJSONDataWithComments(redis_cfg_file)
	lib.IsErrFatal(err)

	err = json.Unmarshal(s, &ServerGlobal)
	if err != nil {
		fmt.Printf("Unable to connect to Redis - Test will not be run! error %s\n", err)
		es := jsonSyntaxErroLib.GenerateSyntaxError(string(s), err)
		fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorYellow, es, MiscLib.ColorReset)
		return false
	}

	return ServerGlobal.ConnectToRedis()
}

type PostDbType struct {
	RunCheck func(conn *sizlib.MyDb) bool
}

var PostDbConnectChecks []PostDbType

func SetupPgSqlForTest(test_cfg string) bool {

	if pg_conn_setup {
		return true
	}
	pg_conn_setup = true

	if ServerGlobal == nil {
		ServerGlobal = NewServerGlobalConfigType()
	}

	s, err := sizlib.ReadJSONDataWithComments(test_cfg)
	lib.IsErrFatal(err)

	err = json.Unmarshal(s, &ServerGlobal)
	if err != nil {
		fmt.Printf("Error: Unable to connect to PostgreSQL - Test will not be run! error %s\n", err)
		es := jsonSyntaxErroLib.GenerateSyntaxError(string(s), err)
		fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorYellow, es, MiscLib.ColorReset)
		return false
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

func SetupEmptyForTest() bool {
	if ServerGlobal == nil {
		ServerGlobal = NewServerGlobalConfigType()
	}
	return true
}

const db_g1 = false

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
}

// trx, id := cfg.TrNewTrx()

//func TrNewTrx() (ptr interface{}, id string) {
//	trx := tr.NewTrx()
//	// wr.RequestTrxId = trx.RequestId
//	id = trx.RequestId
//	ptr = trx
//	return
//}

/* vim: set noai ts=4 sw=4: */
