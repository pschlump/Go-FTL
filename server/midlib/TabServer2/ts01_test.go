//
// TabServer2
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1225
//

package TabServer2

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	// "www.2c-why.com/sizlib-old"
	// _ "github.com/lib/pq"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux" //	"encoding/json"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/uuid"
)

const dbA = false

var db_setup = []struct {
	group  string
	run    string
	r_cmd  string
	r_data []string
	err_ok bool
}{
	{
		group:  "drop",
		run:    `DROP TABLE "test_stuff"`,
		err_ok: true,
	},
	{
		group:  "drop",
		run:    `delete from "log" where "error_level" = 100`,
		err_ok: true,
	},
	{
		group:  "drop",
		run:    `delete from "log" where "error_level" = 90`,
		err_ok: true,
	},
	{
		group: "create",
		run: `
CREATE TABLE "test_stuff" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "hostname"			char varying (40) not null 				
	, "item_name"			char varying (240) not null 				
	, "status"				char varying (40) 
	, "code"				int
	, "updated" 			timestamp 									
	, "created" 			timestamp default current_timestamp not null 
)`,
	},
	{
		group: "populate",
		run: `
insert into "test_stuff" ( "hostname", "item_name", "status", "code" ) values
	( 'dev2',      'backups', 'done', 01 ),
	( 'dev3',      'backups', 'i.p.', 02 ),
	( 'chantelle', 'backups', 'done', 01 ),
	( 'sasha',     'backups', 'pend', 03 ),
	( 'joyce',     'backups', 'done', 01 ),
	( 'corwin',    'backups', 'fail', 11 ),
	( 'mac1',      'backups', 'done', 01 ),
	( 'mac2',      'backups', 'fail', 08 ),
	( 'mac3',      'backups', 'done', 01 )
`,
	},
	{
		group:  "r.set",
		r_cmd:  "set",
		r_data: []string{"abc", "123"},
	},
}

func runSetPg(name, desc string) {
	var err error
	// Postgresql - get connection
	gCfg := cfg.ServerGlobal

	for ii, vv := range db_setup {
		if vv.group == name {
			_ /*Rows*/, err = gCfg.Pg_client.Db.Query(vv.run)
			if dbRunSetPg {
				fmt.Printf("AT: %s - cmd >%s<\n", godebug.LF(), vv.run)
			}
			if (!vv.err_ok) && err != nil {
				fmt.Printf("Error - on %s at item: %d, %s, Error: %s\n", desc, ii, godebug.LF(), err)
			}
		}
	}
}

func runSetRedis(name, desc string) {
	var err error
	gCfg := cfg.ServerGlobal

	// Redis - get connection
	conn, err := gCfg.RedisPool.Get()
	if err != nil {
		// rw.Log.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}
	defer gCfg.RedisPool.Put(conn)

	// Redis - insert keys (overwrite if exists)
	for ii, vv := range db_setup {
		if vv.group == name {

			var tmp []interface{}
			for _, ww := range vv.r_data {
				tmp = append(tmp, ww)
			}

			err := conn.Cmd(vv.r_cmd, tmp...).Err
			if dbRunSetRedis {
				fmt.Printf("AT: %s - cmd >%s<\n", godebug.LF(), vv.run)
			}
			if (!vv.err_ok) && err != nil {
				fmt.Printf("Error - on %s at item: %d, %s, Error: %s, running %s %s\n", desc, ii, godebug.LF(), err, vv.r_cmd, vv.r_data)
			}
		}
	}
}

const dbRunSetPg = false
const dbRunSetRedis = false

func SetupTabServer2TestEnvironment() {
	var err error
	// Postgresql - get connection
	gCfg := cfg.ServerGlobal

	// Verify connection to database
	_ /*Rows*/, err = gCfg.Pg_client.Db.Query(`select 42`)
	if err != nil {
		fmt.Printf("Info on setup at %s, Error: %s\n", godebug.LF(), err)
		return
	}

	runSetPg("drop", "setup")
	runSetPg("create", "create")
	runSetPg("populate", "insert data")

	// Redis - get connection
	// Redis - insert keys (overwrite if exists)
	runSetRedis("r.set", "redis setup")

	fmt.Printf("Success: Tables, Data and Redis setup for test\n")

}
func TeardownTabServer2TestEnvironment() {
	runSetPg("drop", "post-test-postgre-cleanup")
	runSetRedis("r.del", "post-test-redis-cleanup")
	fmt.Printf("Success: Tables, Data and Redis cleaned up - post test\n\n")
}

// -----------------------------------------------------------------------------------------------------------------------------------------------
func Test_0001_TabServer2(t *testing.T) {

	if !cfg.SetupRedisForTest("../test_redis.json") {
		return
	}

	cur_pwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sTabServer2: Error (14022):  Unable to get current working directory. %s\n", MiscLib.ColorRed, MiscLib.ColorReset)
		fmt.Printf("TabServer2: Error (14022):  Unable to get current working directory.\n")
		// return mid.ErrInternalError
		return
	}

	tests := []struct {
		RunIt                bool   // if true then run tst
		url                  string //
		method               string //
		expectedCode         int    //
		nr_ret               int    //
		dump_data            bool   //
		chk_data             bool   //
		chk_hash             bool   //
		row_n                int    //
		col_name             string //
		col_type             string //
		i_value              int    //
		s_value              string //
		a_s_value            bool   //
		expectedReturnStatus string //
		query                string // xyzzy - add a "query" to check count of # of items in d.b. directly that these should return -- run if non-empty and nr_ret == -2
	}{
		// 0
		{
			RunIt:        true,
			url:          "http://example.com/api/table/test7",
			expectedCode: http.StatusOK,
			nr_ret:       -1,
			dump_data:    false,
		},
		// 1
		{
			RunIt:        true,
			url:          "http://example.com/api/table/log",
			expectedCode: http.StatusOK,
			nr_ret:       13189,
			dump_data:    false,
		},
		// 2
		{
			RunIt:        true,
			url:          "http://example.com/api/table/log?id=123",
			expectedCode: http.StatusOK,
			nr_ret:       1,
			dump_data:    true,
		},
		// 3
		{
			RunIt:        true,
			url:          "http://example.com/api/table/log/count?id=123",
			expectedCode: http.StatusOK,
			nr_ret:       1,
			dump_data:    false,
			chk_data:     true,
			row_n:        0,
			col_name:     "nRows",
			col_type:     "int",
			i_value:      1,
		},
		// 4
		{
			RunIt:        true,
			url:          "http://example.com/api/table/log/count",
			expectedCode: http.StatusOK,
			nr_ret:       1,
			dump_data:    false,
			chk_data:     true,
			row_n:        0,
			col_name:     "nRows",
			col_type:     "int",
			i_value:      13189,
		},
		// 5
		{
			RunIt:        true,
			url:          "http://example.com/api/table/log_test_0001?other_id=123",
			expectedCode: http.StatusOK,
			nr_ret:       1,
		},
		// 6 theMux.HandleFunc(api_list+"builtin-routes", closure_respHandlerListBuiltinRoutes(hdlr)).Methods("GET")                                               // List back all the routes
		{
			RunIt:                true,
			url:                  "http://example.com/api/list/builtin-routes",
			expectedCode:         http.StatusOK,
			nr_ret:               -1,
			dump_data:            true,
			expectedReturnStatus: "success",
		},
		// 7 theMux.HandleFunc(api_list+"sql-cfg-files-loaded", closure_respHandlerListSQLCfgFilesLoaded(hdlr)).Methods("GET") //
		{
			RunIt:                true,
			url:                  "http://example.com/api/list/sql-cfg-files-loaded",
			expectedCode:         http.StatusOK,
			nr_ret:               -1,
			dump_data:            true,
			expectedReturnStatus: "error",
		},
		// 8 theMux.HandleFunc(api_list+"cfg-for", closure_respHandlerListCfgFor(hdlr)).Methods("GET")                         //
		{
			RunIt:                true,
			url:                  "http://example.com/api/list/cfg-for",
			expectedCode:         http.StatusOK,
			expectedReturnStatus: "error",
			nr_ret:               -1,
			dump_data:            true,
		},
		// 9 theMux.HandleFunc(api_list+"cfg-for", closure_respHandlerListCfgFor(hdlr)).Methods("GET")                         //
		//   item := ps.ByName("item")
		{
			RunIt:                true,
			url:                  "http://example.com/api/list/cfg-for?item=/api/status_db&dev_auth_token=9abb4f75-f336-46d2-a3af-1115c3d49f14",
			expectedCode:         http.StatusOK,
			expectedReturnStatus: "success",
			nr_ret:               -1,
			dump_data:            true,
		},
		// 10 theMux.HandleFunc(api_list+"end-points", closure_respHandlerListEndPoints(hdlr)).Methods("GET")                   //
		{
			RunIt:        true,
			url:          "http://example.com/api/list/end-points?dev_auth_token=9abb4f75-f336-46d2-a3af-1115c3d49f14",
			expectedCode: http.StatusOK,
			nr_ret:       9,
			dump_data:    true,
		},
		// 11 theMux.HandleFunc(api_list+"logit", respHandlerLogIt).Methods("GET", "POST")                                   // DB! Log information log files
		{
			RunIt:                true,
			url:                  "http://example.com/api/list/logit",
			expectedCode:         http.StatusOK,
			nr_ret:               -1,
			dump_data:            true,
			expectedReturnStatus: "success",
		},
		// 12 theMux.HandleFunc(api_table+"{name}/{id}", closure_respHandlerTableGetPk1(hdlr)).Methods("GET")                   // Select - with single unique PK id - Not fond of positional param
		{
			RunIt:        true,
			url:          "http://example.com/api/table/log/123",
			expectedCode: http.StatusOK,
			nr_ret:       1,
		},
		// 13 theMux.HandleFunc(api_table+"{name}", closure_respHandlerTablePost(hdlr)).Methods("POST")                         // Insert
		{
			RunIt:                true,
			url:                  "http://example.com/api/table/log?id=900000&error_level=100&message=bob+bob+bob",
			method:               "POST",
			expectedCode:         http.StatusOK,
			nr_ret:               -1,
			expectedReturnStatus: "success",
			dump_data:            true,
		},
		// 14 theMux.HandleFunc(api_table+"{name}", closure_respHandlerTablePut(hdlr)).Methods("PUT")                           // Update
		{
			RunIt:                true,
			url:                  "http://example.com/api/table/log?id=900000&error_level=90",
			method:               "PUT",
			expectedCode:         http.StatusOK,
			nr_ret:               -1,
			expectedReturnStatus: "success",
			dump_data:            true,
		},
		// 15 theMux.HandleFunc(api_table+"{name}/{id}", closure_respHandlerTableGetPk1(hdlr)).Methods("GET")                   // Select - with single unique PK id - Not fond of positional param
		{
			RunIt:        true,
			url:          "http://example.com/api/table/log/900000",
			expectedCode: http.StatusOK,
			nr_ret:       1,
			// xyzzy must check data at this point!
		},
		// 16 theMux.HandleFunc(api_table+"{name}", closure_respHandlerTableDel(hdlr)).Methods("DELETE")                        // Delete
		{
			RunIt:                true,
			url:                  "http://example.com/api/table/log?id=900000",
			method:               "DELETE",
			expectedCode:         http.StatusOK,
			nr_ret:               -1,
			expectedReturnStatus: "success",
			dump_data:            true,
		},
		// 17 theMux.HandleFunc(api_table+"{name}/{id}", closure_respHandlerTableGetPk1(hdlr)).Methods("GET")                   // Select - with single unique PK id - Not fond of positional param
		{
			RunIt:        true,
			url:          "http://example.com/api/table/log/900000",
			expectedCode: http.StatusOK,
			nr_ret:       0,
		},
		// 18 theMux.HandleFunc(api_table+"{name}/{id}", closure_respHandlerTablePostPk1(hdlr)).Methods("POST")                 // Insert
		{
			RunIt:                true,
			url:                  "http://example.com/api/table/log/900001?error_level=100&message=bob+bob+bob",
			method:               "POST",
			expectedCode:         http.StatusOK,
			nr_ret:               -1,
			expectedReturnStatus: "success",
			dump_data:            true,
			chk_hash:             true,
			col_name:             "id",
			col_type:             "string",
			a_s_value:            true,
		},
		// 19 theMux.HandleFunc(api_table+"{name}/{id}", closure_respHandlerTablePutPk1(hdlr)).Methods("PUT")                   // Update
		{
			RunIt:                true,
			url:                  "http://example.com/api/table/log/900001?error_level=90",
			method:               "PUT",
			expectedCode:         http.StatusOK,
			nr_ret:               -1,
			expectedReturnStatus: "success",
			dump_data:            true,
		},
		// 20 theMux.HandleFunc(api_table+"{name}/{id}", closure_respHandlerTableDelPk1(hdlr)).Methods("DELETE")                // Delete - with single unique PK id - Not fond of positional param
		{
			RunIt:                true,
			url:                  "http://example.com/api/table/log?id=900001",
			method:               "DELETE",
			expectedCode:         http.StatusOK,
			nr_ret:               -1,
			expectedReturnStatus: "success",
			dump_data:            true,
		},
		// 21 theMux.HandleFunc(api_table+"{name}", closure_respHandlerTableDesc(hdlr)).Methods("HEAD")                         // Describe
		{
			RunIt:                true,
			url:                  "http://example.com/api/table/log/desc",
			expectedCode:         http.StatusOK,
			nr_ret:               -1,
			expectedReturnStatus: "success",
			dump_data:            true,
		},
		// 22 theMux.HandleFunc(api_list+"installed-themes", closure_respHandlerListInstalledThemes(hdlr)).Methods("GET")        // DB! Find the set of installed themes
		{
			RunIt:                true,
			url:                  "http://example.com/api/list/installed-themes",
			expectedCode:         http.StatusOK,
			expectedReturnStatus: "success",
			nr_ret:               -1,
			dump_data:            true,
		},
		// 23 theMux.HandleFunc(api_list+"current-theme", closure_respHandlerListCurrentTheme(hdlr)).Methods("GET")              // DB! Find the currently set theme
		{
			RunIt:                true,
			url:                  "http://example.com/api/list/current-theme",
			expectedCode:         http.StatusOK,
			expectedReturnStatus: "success",
			nr_ret:               -1,
			dump_data:            true,
		},
		// 24 theMux.HandleFunc(api_list+"reloadTableConfig", closure_respHandlerReloadTableConfig(hdlr)).Methods("GET", "POST") // research-and load sql-cfg*.* files
		// /Users/corwin/go/src/github.com/pschlump/TabServer2/test/test01_sql_cfg.json
		{
			RunIt:                true,
			url:                  "http://example.com/api/list/reloadTableConfig?dev_auth_token=9abb4f75-f336-46d2-a3af-1115c3d49f14&fn=" + cur_pwd + "/test/test01_sql_cfg.json",
			expectedCode:         http.StatusOK,
			expectedReturnStatus: "success",
			nr_ret:               -1,
			dump_data:            true,
		},
		//
		// 25 1. Insert using "autoGen" true - and returingin an id
		//    done *** theMux.HandleFunc(api_table+"{name}", closure_respHandlerTablePost(hdlr)).Methods("POST")                         // Insert
		//
		{
			RunIt:                true,
			url:                  "http://example.com/api/table/log_test_0002?id=900000&error_level=100&message=bob+bob+bob",
			method:               "POST",
			expectedCode:         http.StatusOK,
			nr_ret:               -1,
			expectedReturnStatus: "success",
			dump_data:            true,
		},
		// 26
		{
			RunIt:                true,
			url:                  "http://example.com/api/table/test7a",
			expectedCode:         http.StatusOK,
			nr_ret:               -1,
			dump_data:            true,
			expectedReturnStatus: "success",
		},
		// 27 theMux.HandleFunc(api_list+"sql-cfg-files-loaded", closure_respHandlerListSQLCfgFilesLoaded(hdlr)).Methods("GET") //
		{
			RunIt:                true,
			url:                  "http://example.com/api/list/sql-cfg-files-loaded?dev_auth_token=9abb4f75-f336-46d2-a3af-1115c3d49f14",
			expectedCode:         http.StatusOK,
			nr_ret:               1,
			dump_data:            true,
			expectedReturnStatus: "success",
		},
	}

	// 1. Connect to D.B. - Redis and PostgreSQL
	// 2. If fail to connect then report this - and make this a passing test -
	if !cfg.SetupPgSqlForTest("./test/pg_and_redis.json") {
		fmt.Printf("Failed to connect to Postgres - no tests run\n")
		return
	}
	if !cfg.SetupRedisForTest("./test/pg_and_redis.json") {
		fmt.Printf("Failed to connect to Redis - no tests run\n")
		return
	}

	// 3. Setup test table(s) and data
	SetupTabServer2TestEnvironment()

	bot := mid.NewServer()

	ms := NewTabServer2Server(bot, []string{"/api/table/", "/api/list/", "/api/"}, []string{"./test/"}, cfg.ServerGlobal)

	// xyzzy 5. - Configure more -- setup config for running test SQLOne stuff
	hh := ms

	// if sqlCfgFN, ok := sizlib.SearchPathApp(hh.SQLCfgFN, hh.AppName, hh.SearchPath); ok {
	hh.SQLCfgFN = "./test/test01_sql_cfg.json" // -- move inside testing loop use: sql_one_cfg_file string
	hh.AppName = "./test"
	hh.SearchPath = "./test"

	// xyzzy setup watchers for changes in files?

	hh.db_func = make(map[string]bool, maxI(len(hh.DbFunctions), 1))
	for _, vv := range hh.DbFunctions {
		// db_func["PickInsertUpdateColumns"] = false
		hh.db_func[vv] = true
	}

	hh.pwd = cur_pwd
	n_err, prev_err := 0, 0

	// Convert from String LoginSystem -> Internal Type LoginSystemType
	switch hh.LoginSystem {
	case "LstNone":
		hh.loginSystem = LstNone
	case "LstAesSrp":
		hh.loginSystem = LstAesSrp
	case "LstUnPw":
		hh.loginSystem = LstUnPw
	case "LstBasic":
		hh.loginSystem = LstBasic
	default:
		hh.loginSystem = LstAesSrp
		// hh.loginSystem = LstNone
		fmt.Fprintf(os.Stderr, "%sTabServer2: Info (14122):  Unable to convert LoginSystem [%s]. Should be one of 'LstNone', 'LstAesSrp', 'LstUnPw', 'LstBasic'.   AesSrp assumed.  LineNo:%d.%s\n", MiscLib.ColorYellow, hh.LoginSystem, hh.LineNo, MiscLib.ColorReset)
		fmt.Printf("TabServer2: Info (14122):  Unable to convert LoginSystem [%s]. Should be one of 'LstNone', 'LstAesSrp', 'LstUnPw', 'LstBasic'.   AesSrp assumed.  LineNo:%d.\n", hh.LoginSystem, hh.LineNo)
	}

	if db3 {
		sqlCfgFN, ok := sizlib.SearchPathApp(hh.SQLCfgFN, hh.AppName, hh.SearchPath)
		fmt.Printf("sqlCfgFN = %s ok = %v, %s\n", sqlCfgFN, ok, godebug.LF())
	}

	if sqlCfgFN, ok := sizlib.SearchPathApp(hh.SQLCfgFN, hh.AppName, hh.SearchPath); ok {
		fmt.Printf("TabServer2: sql config: %s\n", sqlCfgFN)
		SQLCfg, err := readInSQLConfig(sqlCfgFN)
		hh.SQLCfg = SQLCfg
		if err != nil {
			fmt.Printf("TabServer2: Error: %s\n", err)
			SqlCfgFilesLoaded = append(SqlCfgFilesLoaded, SqlCfgLoaded{FileName: hh.pwd + sqlCfgFN[1:], ErrorMsg: fmt.Sprintf("%s", err)})
			t.Errorf("%sError Unable to load config file%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
			n_err++
		} else {
			SqlCfgFilesLoaded = append(SqlCfgFilesLoaded, SqlCfgLoaded{FileName: hh.pwd + sqlCfgFN[1:], ErrorMsg: ""})
		}
	} else {
		fmt.Fprintf(os.Stderr, "%sTabServer2: Error (14122):  Unable to find the %s file using %s path. LineNo:%d.%s\n", MiscLib.ColorRed, hh.SQLCfgFN, hh.SearchPath, hh.LineNo, MiscLib.ColorReset)
		fmt.Printf("TabServer2: Error (14122):  Unable to find the %s file using %s path. LineNo:%d.\n", hh.SQLCfgFN, hh.SearchPath, hh.LineNo)
		// return mid.ErrInternalError
		return
	}

	// xyzzy - check use of "LineNo" and validate that
	// xyzzy - 0. That we got a set of items more than 5 from config file
	// xyzzy - 1. Wrong line numbers are possible via a constant
	// xyzzy - 2. Some line numbes are larger than others
	// xyzzy - 3. File... Line... works
	// xyzzy - 4. File has file name in it and correct
	// xyzzy - 5. That certain items have "Comment" as a non-empty item

	if n_err != 0 {
		return
	}

	hh.theMux = goftlmux.NewRouter()

	initEndPoints(hh.theMux, hh)

	hh.final = true

	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		if test.RunIt {

			prev_err = n_err

			fmt.Printf(`


+==============================================================================================================================================
| Start of test %d 
+==============================================================================================================================================

`, ii)
			fmt.Printf("URL: %s\n", test.url)

			rec := httptest.NewRecorder()
			wr := goftlmux.NewMidBuffer(rec, nil)

			id := "test-01-StatusHandler"
			trx := tr.NewTrx(cfg.ServerGlobal.RedisPool)
			trx.TrxIdSeen(id, test.url, "GET")
			wr.RequestTrxId = id

			wr.G_Trx = trx

			var req *http.Request

			req, err = http.NewRequest("GET", test.url, nil)
			if err != nil {
				t.Fatalf("Test %d: Could not create HTTP request: %v", ii, err)
				n_err++
			}

			// ----------------------------------------------------------------------------------- should be a func  ---------------------------------------------------
			// ----------------------------------------------------------------------------------- share with mid.go ---------------------------------------------------
			id0, _ := uuid.NewV4()
			id := id0.String()

			trx := tr.NewTrx()
			// Per request ID that is used by Trx tracing package.
			trx.RequestId = id
			wr.RequestTrxId = id

			trx.SetRedisConn(cfg.ServerGlobal.RedisPool) // Set to connect to redis

			wr.G_Trx = trx
			// ---------------------------------------------------------------------------------------------------------------------------------------------------------

			// Parse Query
			goftlmux.ParseQueryParamsReg(wr, req, &wr.Ps) //
			fmt.Printf("Params: %s\n", wr.Ps.DumpParamTable())

			lib.SetupTestMimicReq(req, "example.com")
			if dbA {
				fmt.Printf("{\"req\":%s,\n\"wr\":%s}\n", lib.SVarI(req), lib.SVarI(wr))
			}

			if test.method != "" {
				fmt.Printf("%sSetting Method to %s%s\n", MiscLib.ColorGreen, test.method, MiscLib.ColorReset)
				req.Method = test.method
			}

			ms.ServeHTTP(wr, req)

			code := wr.StatusCode
			// Tests to perform on final recorder data.
			if code != test.expectedCode {
				t.Errorf("Error %2d, reject error got: %d, expected %d\n", ii, wr.StatusCode, test.expectedCode)
				n_err++
			}

			wr.FinalFlush()

			b := string(rec.Body.Bytes())
			if test.dump_data {
				fmt.Printf("%sRaw Data: %s%s\n", MiscLib.ColorYellow, b, MiscLib.ColorReset)
			}
			b_parsed, err := lib.JsonStringToArrayOfData(b)
			if err != nil {
				a_parsed, err1 := lib.JsonStringToData(b)
				if err1 == nil {
					if test.expectedReturnStatus != "" {
						if a_parsed["status"] == test.expectedReturnStatus {
						} else if a_parsed["Status"] == test.expectedReturnStatus {
							t.Errorf("Error Status returned with upper case!")
						} else {
							t.Errorf("Error %2d, not a successful call, --->>>%s<<<---\n", ii, b)
							n_err++
						}
					}
					if test.chk_hash {
						tt, ok := a_parsed[test.col_name]
						if !ok {
							t.Errorf("Error %2d, expect to access filed %s, failed\n", ii, test.col_name)
							n_err++
						} else {
							switch test.col_type {
							case "int":
								rr, ok := tt.(float64)
								if !ok {
									t.Errorf("Error %2d, expect to type-cast column %s to (int) failed\n", ii, test.col_name)
									n_err++
								} else {
									rri := int(rr)
									if rri != test.i_value {
										t.Errorf("Error %2d, expected %d, got %v\n", ii, test.i_value, rr)
										n_err++
									}
								}
							case "string":
								// fmt.Printf("AtAt: %s\n", godebug.LF())
								rrs, ok := tt.(string)
								if !ok {
									t.Errorf("Error %2d, expect to type-cast column %s to (int) failed\n", ii, test.col_name)
									n_err++
								} else {
									// fmt.Printf("AtAt: %s\n", godebug.LF())
									if test.a_s_value {
										// fmt.Printf("AtAt: %s\n", godebug.LF())
										if len(rrs) == 0 {
											t.Errorf("Error %2d, expected a non-empty string, got [%s] for [%s]\n", ii, rrs, test.col_name)
											n_err++
										}
									} else if rrs != test.s_value {
										t.Errorf("Error %2d, expected %s, got [%s] for [%s]\n", ii, test.s_value, rrs, test.col_name)
										n_err++
									}
								}
							// Xyzzy - float??
							// Xyzzy - date??
							default:
								panic("not supported type")
							}
						}
					}
				} else {
					c_parsed, err2 := lib.JsonStringToArrayOfString(b)
					if err2 == nil {
						l := len(c_parsed)
						if test.nr_ret != -1 && l != test.nr_ret {
							t.Errorf("Error %2d, expect %d rows, got %d\n", ii, test.nr_ret, l)
							n_err++
						}
					} else {
						t.Errorf("Error %2d, unable to parse JSON return value, --->>>%s<<<--- error, %v\n", ii, b, err2)
						n_err++
					}
				}
			} else {
				if test.expectedReturnStatus != "" {
					t.Errorf("Error %2d, to have a 'status' in a hash but got array instead\n", ii)
					n_err++
				}
				l := len(b_parsed)
				if test.nr_ret >= 0 && l != test.nr_ret {
					t.Errorf("Error %2d, expect %d rows, got %d\n", ii, test.nr_ret, l)
					n_err++
				}
			}

			if test.chk_data {
				ok := len(b_parsed) > test.row_n
				if !ok {
					t.Errorf("Error %2d, expect to access row %d, failed\n", ii, test.row_n)
					n_err++
				} else {
					ss := b_parsed[test.row_n]
					tt, ok := ss[test.col_name]
					if !ok {
						t.Errorf("Error %2d, expect to access column %s, failed\n", ii, test.col_name)
						n_err++
					} else {
						switch test.col_type {
						case "int":
							rr, ok := tt.(float64)
							if !ok {
								t.Errorf("Error %2d, expect to type-cast column %s to (int) failed\n", ii, test.col_name)
								n_err++
							} else {
								rri := int(rr)
								if rri != test.i_value {
									t.Errorf("Error %2d, expected %d, got %v\n", ii, test.i_value, rr)
									n_err++
								}
							}
						case "string":
							// fmt.Printf("AtAt: %s\n", godebug.LF())
							rrs, ok := tt.(string)
							if !ok {
								t.Errorf("Error %2d, expect to type-cast column %s to (int) failed\n", ii, test.col_name)
								n_err++
							} else {
								// fmt.Printf("AtAt: %s\n", godebug.LF())
								if test.a_s_value {
									// fmt.Printf("AtAt: %s\n", godebug.LF())
									if len(rrs) == 0 {
										t.Errorf("Error %2d, expected a non-empty string, got [%s] for [%s]\n", ii, rrs, test.col_name)
										n_err++
									}
								} else if rrs != test.s_value {
									t.Errorf("Error %2d, expected %s, got [%s] for [%s]\n", ii, test.s_value, rrs, test.col_name)
									n_err++
								}
							}
						// Xyzzy - float??
						// Xyzzy - date??
						default:
							panic("not supported type")
						}
					}
				}
			}

			if n_err != prev_err {
				fmt.Printf("%sErrors Occured in **This** Test %d%s\n", MiscLib.ColorRed, n_err-prev_err, MiscLib.ColorReset)
			} else {
				fmt.Printf("%sSuccessful PASS%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
			}
		}
	}

	fmt.Printf("\n\nTests Completed -------------------------------------------------------------------------------------\n\n")

	// 3. Cleanup test table(s) and data
	TeardownTabServer2TestEnvironment()

}

// const db6 = false
// const db8 = false

/* Individual tests
TODO:
	*** done *** func (hdlr *TabServer2Type) RemapParams(ps *goftlmux.Params, h SQLOne, ptr *tr.Trx) {
	*** done *** func (hdlr *TabServer2Type) GetDbType() DbType {

	// -------------------------- CRUD Handlers -------------------------------------------------------------------
	*** done *** theMux.HandleFunc(api_table+"{name}/count", closure_respHandlerTableGetCount(hdlr)).Methods("GET")            // Select count(*)
	*** done *** theMux.HandleFunc(api_table+"{name}/{id}", closure_respHandlerTableGetPk1(hdlr)).Methods("GET")               // Select - with single unique PK id - Not fond of positional param
	*** done *** theMux.HandleFunc(api_table+"{name}/{id}", closure_respHandlerTableDelPk1(hdlr)).Methods("DELETE")            // Delete - with single unique PK id - Not fond of positional param
	*** done *** theMux.HandleFunc(api_table+"{name}/{id}", closure_respHandlerTablePutPk1(hdlr)).Methods("PUT")                   // Update
	*** done *** theMux.HandleFunc(api_table+"{name}/{id}", closure_respHandlerTablePostPk1(hdlr)).Methods("POST")                 // Insert
	*** done *** theMux.HandleFunc(api_table+"{name}", closure_respHandlerTableGet(hdlr)).Methods("GET")                        // Select
	*** done *** theMux.HandleFunc(api_table+"{name}", closure_respHandlerTablePut(hdlr)).Methods("PUT")                           // Update
	*** done *** theMux.HandleFunc(api_table+"{name}", closure_respHandlerTablePost(hdlr)).Methods("POST")                         // Insert
	*** done *** theMux.HandleFunc(api_table+"{name}", closure_respHandlerTableDel(hdlr)).Methods("DELETE")                        // Delete
0.	not-implemented !! theMux.HandleFunc(api_table+"{name}", closure_respHandlerTableDesc(hdlr)).Methods("HEAD")                         // Describe
		1. Desc of table in d.b.?
		2. Desc of config of table in config file?
		3. Desc-able columns?
	***done*** theMux.HandleFunc(api_list+"sql-cfg-files-loaded", closure_respHandlerListSQLCfgFilesLoaded(hdlr)).Methods("GET") //
	***done*** theMux.HandleFunc(api_list+"cfg-for", closure_respHandlerListCfgFor(hdlr)).Methods("GET")                         //
	***done*** theMux.HandleFunc(api_list+"end-points", closure_respHandlerListEndPoints(hdlr)).Methods("GET")                   //

	// -------------------------- From base.go --------------------------------------------------------------------
	*** done *** theMux.HandleFunc(api_list+"tab-server2/status", closure_respHandlerStatus(hdlr)).Methods("GET", "POST", "HEAD", "PATCH", "PUT", "DELETE", "OPTIONS") //
	***done***  theMux.HandleFunc(api_list+"reloadTableConfig", closure_respHandlerReloadTableConfig(hdlr)).Methods("GET", "POST")                                    // research-and load sql-cfg*.* files
	***done*** theMux.HandleFunc(api_list+"builtin-routes", closure_respHandlerListBuiltinRoutes(hdlr)).Methods("GET")                                    // List back all the routes
	***done*** theMux.HandleFunc(api_list+"logit", respHandlerLogIt).Methods("GET", "POST")                                                               // DB! Log information log files

	***done*** theMux.HandleFunc(api_list+"installed-themes", closure_respHandlerListInstalledThemes(hdlr)).Methods("GET")                                           // DB! Find the set of installed themes
	***done*** theMux.HandleFunc(api_list+"current-theme", closure_respHandlerListCurrentTheme(hdlr)).Methods("GET")                                                 // DB! Find the currently set theme

On Insert tests to perform
	**done** 1. Insert using "autoGen" true - and returingin an id
	**done** 2. Verify that Id that is retuned is the sam as passed.
		, "DebugFlag": [ "dump_insert_params", "db_insert" ]

crud.go: -- Implement desc
		// xyzzyDOIT - Don't you think that you should DO somethin at this point
*/

/* vim: set noai ts=4 sw=4: */
