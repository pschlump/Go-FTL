//
// Go-FTL / TabServer2
//
// Copyright (C) Philip Schlump, 2012-2017. All rights reserved.
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1011
//

package TabServer2

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/Go-FTL/server/tmplp"
	JsonX "github.com/pschlump/JSONx"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/SqlEr"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical - but not this time.
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*TabServer2Type)
//		if ok {
//			pCfg.SetNext(next)
//			rv = pCfg
//		} else {
//			err = mid.FtlConfigError
//		}
//		gCfg.ConnectToRedis()
//		gCfg.ConnectToPostgreSQL()
//		pCfg.gCfg = gCfg
//		return
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} { return &TabServer2Type{} }
//
//	postInit := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
//		// var err error
//
//		hh, ok := h.(*TabServer2Type)
//		if !ok {
//			fmt.Printf("Error: Wrong data type passed to FileServeType - postInit\n")
//			return mid.ErrInternalError
//		} else {
//
//			hh.MuxAuto = make(map[string]int)
//			hh.MuxAutoPass = 1
//
//			// xyzzy setup watchers for changes in files?
//
//			hh.db_func = make(map[string]bool, maxI(len(hh.DbFunctions), 1))
//			for _, vv := range hh.DbFunctions {
//				// db_func["PickInsertUpdateColumns"] = false
//				hh.db_func[vv] = true
//			}
//
//			t, err := os.Getwd()
//			if err != nil {
//				fmt.Fprintf(os.Stderr, "%sTabServer2: Error (14022):  Unable to get current working directory. LineNo:%d.%s\n", MiscLib.ColorRed, hh.LineNo, MiscLib.ColorReset)
//				fmt.Printf("TabServer2: Error (14022):  Unable to get current working directory. LineNo:%d.\n", hh.LineNo)
//				return mid.ErrInternalError
//			}
//			hh.pwd = t
//
//			fmt.Printf("\nTabServer2: --- start of TabServer2 config --- Running in [%s] LineNo:%d.\n", t, hh.LineNo)
//
//			// Convert from String LoginSystem -> Internal Type LoginSystemType
//			switch hh.LoginSystem {
//			case "LstNone":
//				hh.loginSystem = LstNone
//			case "LstAesSrp":
//				hh.loginSystem = LstAesSrp
//			case "LstUnPw":
//				hh.loginSystem = LstUnPw
//			case "LstBasic":
//				hh.loginSystem = LstBasic
//			default:
//				hh.loginSystem = LstAesSrp
//				// hh.loginSystem = LstNone
//				fmt.Fprintf(os.Stderr, "%sTabServer2: Info (15122):  Unable to convert LoginSystem [%s]. Should be one of 'LstNone', 'LstAesSrp', 'LstUnPw', 'LstBasic'.   AesSrp assumed.  LineNo:%d.%s\n", MiscLib.ColorYellow, hh.LoginSystem, hh.LineNo, MiscLib.ColorReset)
//				fmt.Printf("TabServer2: Info (15122):  Unable to convert LoginSystem [%s]. Should be one of 'LstNone', 'LstAesSrp', 'LstUnPw', 'LstBasic'.   AesSrp assumed.  LineNo:%d.\n", hh.LoginSystem, hh.LineNo)
//			}
//
//			if db3 {
//				sqlCfgFN, ok := sizlib.SearchPathApp(hh.SQLCfgFN, hh.AppName, hh.SearchPath)
//				fmt.Printf("sqlCfgFN = %s ok = %v, %s\n", sqlCfgFN, ok, godebug.LF())
//			}
//
//			n_config := 0
//			n_files_loaded := 0
//
//			if sqlCfgFN, ok := sizlib.SearchPathApp(hh.SQLCfgFN, hh.AppName, hh.SearchPath); ok {
//				fmt.Printf("TabServer2: sql config: %s, %s\n", sqlCfgFN, godebug.LF())
//				SQLCfg, err := readInSQLConfig(sqlCfgFN)
//				hh.SQLCfg = SQLCfg
//				if err != nil {
//					fmt.Printf("TabServer2: Error: %s\n", err)
//					SqlCfgFilesLoaded = append(SqlCfgFilesLoaded, SqlCfgLoaded{FileName: hh.pwd + sqlCfgFN[1:], ErrorMsg: fmt.Sprintf("%s", err)})
//				} else {
//					n_config++
//					n_files_loaded = 1
//					SqlCfgFilesLoaded = append(SqlCfgFilesLoaded, SqlCfgLoaded{FileName: hh.pwd + sqlCfgFN[1:], ErrorMsg: ""})
//				}
//			}
//
//			// ----------------------------------------------------------------------------------------------------------------------------------------------------
//			// Read in module based end-points
//			// ----------------------------------------------------------------------------------------------------------------------------------------------------
//			// Called from ~/Projects/w-watch/w-watch.go
//			// 		s := doGet(&client, "http://localhost:8090/api/reloadTableConfig")
//			// in: base.go -- respHandlerReloadTableConfig(res http.ResponseWriter, req *http.Request, ps goftlmux.Params) {
//			// ----------------------------------------------------------------------------------------------------------------------------------------------------
//			for _, TopPath := range hh.AppRoot {
//				var ignoreList []string
//				fmt.Printf("TabServer2: At Search for additional sql-cfg.json files, TopPath=%s from AppRoot=%s, %s\n", TopPath, godebug.SVar(hh.AppRoot), godebug.LF())
//				// fmt.Printf("At Search for additional sql-cfg.json files , %s\n", godebug.LF())
//				// opts__TopPath := sizlib.SubstitueUserInFilePathImmediate("/Users/corwin/Projects/who-cares/app") // xyzzy from CLI -W option //xyzzy - replace ~ with home dir.
//				opts__TopPath := sizlib.SubstitueUserInFilePathImmediate(TopPath)
//				// fmt.Printf("TabServer2: Path ->%s<- At, %s\n", opts__TopPath, godebug.LF())
//				// ignoreList = append(ignoreList, "/Users/corwin/Projects/who-cares/who-cares-server") // xyzzy from globa-cfg.json file
//				dirs := sizlib.FindDirsWithSQLCfg(opts__TopPath, ignoreList)
//				// fmt.Printf("TabServer2: dirs ->%s<- At, %s\n", sizlib.SVar(dirs), godebug.LF())
//				fList, ok := sizlib.SearchPathAppModule(hh.SQLCfgFN, hh.AppName, dirs)
//				// fmt.Printf("TabServer2: fList ->%s<- At, %s\n", sizlib.SVar(fList), godebug.LF())
//				if ok {
//					fmt.Fprintf(os.Stderr, "%sTabServer2: List of additional sql-cfg*.josn files found: %s server config line:%d AT, %s%s\n",
//						MiscLib.ColorGreen, sizlib.SVar(fList), hh.LineNo, godebug.LF(), MiscLib.ColorReset)
//					for _, v := range fList {
//						n_files_loaded++
//						fmt.Printf("TabServer2: Reading in additional SQLCfg: %s\n", v)
//						fmt.Fprintf(os.Stderr, "%sTabServer2: Reading in additional SQLCfg: %s%s\n", MiscLib.ColorGreen, v, MiscLib.ColorReset)
//						tSQLCfg, err := readInSQLConfig(v) // func readInSQLConfig(path string) map[string]SQLOne {
//						if err != nil {
//							fmt.Printf("TabServer2: Error: %s\n", err)
//							SqlCfgFilesLoaded = append(SqlCfgFilesLoaded, SqlCfgLoaded{FileName: hh.pwd + v[1:], ErrorMsg: fmt.Sprintf("%s", err)})
//						} else {
//							// 1. combine note: 'f' values - concatenate for each key - instead of overwrite -- xyzzyConcatNoteKey
//							//		Collect all the note: 'f's into a set of strings - then post-process them
//							preNote := make(map[string]string)
//							for ii, vv := range hh.SQLCfg {
//								if strings.HasPrefix(ii, "note:") {
//									preNote[ii] = vv.F
//								}
//							}
//							fmt.Printf("PreNote = %s\n", preNote)
//							if hh.SQLCfg == nil {
//								hh.SQLCfg = make(map[string]SQLOne)
//							}
//							for j, w := range tSQLCfg {
//								hh.SQLCfg[j] = w
//							}
//							for ii, vv := range hh.SQLCfg {
//								if strings.HasPrefix(ii, "note:") {
//									if old, ok := preNote[ii]; ok {
//										// preNote[ii] = vv.F
//										vv.F = old + "\n" + vv.F
//										hh.SQLCfg[ii] = vv
//									}
//								}
//							}
//							SqlCfgFilesLoaded = append(SqlCfgFilesLoaded, SqlCfgLoaded{FileName: hh.pwd + v[1:], ErrorMsg: ""})
//							n_config++
//						}
//					}
//				}
//			}
//
//			if n_files_loaded == 0 {
//				fmt.Fprintf(os.Stderr, "%sTabServer2: Error (14122):  Unable to find the %s file using %s path. AppName=%s LineNo:%d in server config file.%s\n", MiscLib.ColorRed, hh.SQLCfgFN, hh.SearchPath, hh.AppName, hh.LineNo, MiscLib.ColorReset)
//				fmt.Printf("TabServer2: Error (14122):  Unable to find the %s file using %s path. LineNo:%d in server config file.\n", hh.SQLCfgFN, hh.SearchPath, hh.LineNo)
//			}
//
//			// xyzzy - valid that the sql_cfg.json data is correct - check table/column info
//			if !hh.CheckSqlCfgValid() {
//				fmt.Printf("Early exit - sql_cfg.json is not valid\n")
//				n_config = -1
//			}
//
//			// xyzzy - put this back in -- loadAllCsrfTokens(hh)
//
//			hh.theMux = goftlmux.NewRouter()
//
//			initEndPoints(hh.theMux, hh)
//
//			hh.final, err = lib.ParseBool(hh.Final)
//
//			if n_config == 0 {
//				fmt.Printf("\n************************************************\n* Warning - no TabServer2 config files loaded\n ************************************************\n\n")
//				return mid.ErrInternalError
//			}
//		}
//
//		return nil
//	}
//
//	/*
//	   var opts struct {
//	   	GlobalCfgFN string `short:"g" long:"globaCfgFile"    description:"Full path to global config"          default:"global-cfg.json"`
//	   	SQLCfgFN    string `short:"s" long:"sqlCfgFile"      description:"Full path to SQL config"             default:"sql-cfg.json"`
//	   	Port        string `short:"p" long:"port"            description:"Port to listen on"                   default:"8090"` // Used from global-cfg.json
//	   	Search      string `short:"S" long:"searchPath"      description:"SearchPath to use for config files"  default:"./cfg:.:~/cfg"`
//	   	AppName     string `short:"A" long:"application"     description:"Application to run"                  default:""`
//	   	TopPath     string `short:"T" long:"topPath"         description:"search for sql-cfg.json files"       default:""`
//	   }
//	*/
//
//	cfg.RegInitItem2("TabServer2", initNext, createEmptyType, postInit, `{
//		}`)
//	// 1. g_schema // xyzzy - should pull from config "public"
//	// 2. xyzzyPath1 // xyzzyPath1 - should pull from config ./table_ddl
//}
//
//// normally identical
//func (hdlr *TabServer2Type) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &TabServer2Type{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("TabServer2", CreateEmpty, `{
		"Paths":                         { "type":[ "string", "filepath" ], "isarray":true, "required":true },
		"AppRoot":                       { "type":[ "string", "filepath" ], "isarray":true },
		"DbFunctions":                   { "type":[ "string", "filepath" ], "isarray":true },
		"WatchForConfigChanges":         { "type":[ "bool" ] },
		"SQLCfgFN":                      { "type":[ "string" ], "default":"sql-cfg.json" },
		"AppName":                       { "type":[ "string" ] },
		"SearchPath":                    { "type":[ "string" ], "default":"./cfg:.:~/cfg" },
		"Final":                         { "type":[ "string" ], "default":"no" },
		"DevAuthToken":                  { "type":[ "string" ], "default":"9abb4f75-f336-46d2-a3af-1115c3d49f14" },
		"DebugFlags":                    { "type":[ "string" ], "isarray":true },
		"AuthorizeNetLogin":             { "type":[ "string" ] },
		"AuthorizeNetKey":               { "type":[ "string" ] },
		"StatusMessage":                 { "type":[ "string" ] },
		"ApiTableKey":                   { "type":[ "string" ], "default":"324d4b9f-00dc-4ea9-7a6c-e5f125207759" },
		"RedisApiTableKey":              { "type":[ "string" ], "default":"" },
		"LogToFile":                     { "type":[ "string" ] },
		"LoginSystem":                   { "type":[ "string" ], "default":"LstAesSrp" },
		"ApiTable":                      { "type":[ "string" ], "default":"/api/table/" },
		"ApiList":                       { "type":[ "string" ], "default":"/api/list/" },
		"ApiStatus":                     { "type":[ "string" ], "default":"/api/list/" },
		"LimitPostJoinRows":             { "type":[ "int" ], "default":"-1" },
		"SendStatusOnError":             { "type":[ "bool" ], "default":"false" },
		"DbSchema":                      { "type":[ "string" ], "default":"public" },
		"DbCreateScript":                { "type":[ "string" ], "default":"./table_ddl/{{.TableName}}.jx" },
		"RedisSessionPrefix":            { "type":[ "string" ], "default":"session:" },
		"EmailConfigFileName":           { "type":[ "string" ], "default": "./email-config.json" },
		"EmailTemplateDir":              { "type":[ "string" ], "default": "./tmpl/" },
		"KeyFilePrivate":                { "type":[ "string" ] },
		"DisplayURL2fa":                 { "type":[ "string" ], "required":false, "default": "/2fa/2fa-app.html" },
		"RedisPrefix2fa":                { "type":[ "string" ], "required":false, "default": "2fa:" },
		"Server2faURL":                  { "type":[ "string" ], "required":false, "default":"http://t432z.com/2fa"  },
		"LineNo":                        { "type":[ "int" ], "default":"1" }
		}`)
}

// status, body := DoGet("http://t432z.com/upd/", "url", hdlr.DisplayURL2fa, "id", qrId, "data", theData, "_ran_", ran)
// key := fmt.Sprintf("%s%s", hdlr.RedisPrefix2fa, ID)

func (hdlr *TabServer2Type) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	gCfg.ConnectToRedis()
	gCfg.ConnectToPostgreSQL()
	hdlr.gCfg = gCfg
	return
}

func (hdlr *TabServer2Type) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	ConfigEmailAWS(hdlr, hdlr.EmailConfigFileName)
	hdlr.MuxAuto = make(map[string]int)
	hdlr.MuxAutoPass = 1

	// xyzzy setup watchers for changes in files?

	hdlr.db_func = make(map[string]bool, maxI(len(hdlr.DbFunctions), 1))
	for _, vv := range hdlr.DbFunctions {
		// db_func["PickInsertUpdateColumns"] = false
		hdlr.db_func[vv] = true
	}

	t, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sTabServer2: Error (14022):  Unable to get current working directory. LineNo:%d.%s\n", MiscLib.ColorRed, hdlr.LineNo, MiscLib.ColorReset)
		fmt.Printf("TabServer2: Error (14022):  Unable to get current working directory. LineNo:%d.\n", hdlr.LineNo)
		return mid.ErrInternalError
	}
	hdlr.pwd = t

	fmt.Printf("\nTabServer2: --- start of TabServer2 config --- Running in [%s] LineNo:%d.\n", t, hdlr.LineNo)

	// Convert from String LoginSystem -> Internal Type LoginSystemType
	switch hdlr.LoginSystem {
	case "LstNone":
		hdlr.loginSystem = LstNone
	case "LstAesSrp":
		hdlr.loginSystem = LstAesSrp
	case "LstUnPw":
		hdlr.loginSystem = LstUnPw
	case "LstBasic":
		hdlr.loginSystem = LstBasic
	default:
		hdlr.loginSystem = LstAesSrp
		// hdlr.loginSystem = LstNone
		fmt.Fprintf(os.Stderr, "%sTabServer2: Info (15122):  Unable to convert LoginSystem [%s]. Should be one of 'LstNone', 'LstAesSrp', 'LstUnPw', 'LstBasic'.   AesSrp assumed.  LineNo:%d.%s\n", MiscLib.ColorYellow, hdlr.LoginSystem, hdlr.LineNo, MiscLib.ColorReset)
		fmt.Printf("TabServer2: Info (15122):  Unable to convert LoginSystem [%s]. Should be one of 'LstNone', 'LstAesSrp', 'LstUnPw', 'LstBasic'.   AesSrp assumed.  LineNo:%d.\n", hdlr.LoginSystem, hdlr.LineNo)
	}

	if db3 {
		sqlCfgFN, ok := sizlib.SearchPathApp(hdlr.SQLCfgFN, hdlr.AppName, hdlr.SearchPath)
		fmt.Printf("sqlCfgFN = %s ok = %v, %s\n", sqlCfgFN, ok, godebug.LF())
	}

	n_config := 0
	n_files_loaded := 0

	if sqlCfgFN, ok := sizlib.SearchPathApp(hdlr.SQLCfgFN, hdlr.AppName, hdlr.SearchPath); ok {
		fmt.Printf("TabServer2: sql config: %s, %s\n", sqlCfgFN, godebug.LF())
		fmt.Fprintf(os.Stderr, "TabServer2: sql config: %s, %s\n", sqlCfgFN, godebug.LF())
		SQLCfg, err := readInSQLConfig(sqlCfgFN)
		hdlr.SQLCfg = SQLCfg
		if err != nil {
			fmt.Printf("TabServer2: Error: %s\n", err)
			SqlCfgFilesLoaded = append(SqlCfgFilesLoaded, SqlCfgLoaded{FileName: hdlr.pwd + sqlCfgFN[1:], ErrorMsg: fmt.Sprintf("%s", err)})
		} else {
			n_config++
			n_files_loaded = 1
			SqlCfgFilesLoaded = append(SqlCfgFilesLoaded, SqlCfgLoaded{FileName: hdlr.pwd + sqlCfgFN[1:], ErrorMsg: ""})
		}
	}

	// ----------------------------------------------------------------------------------------------------------------------------------------------------
	// Read in module based end-points
	// ----------------------------------------------------------------------------------------------------------------------------------------------------
	// Called from ~/Projects/w-watch/w-watch.go
	// 		s := doGet(&client, "http://localhost:8090/api/reloadTableConfig")
	// in: base.go -- respHandlerReloadTableConfig(res http.ResponseWriter, req *http.Request, ps goftlmux.Params) {
	// ----------------------------------------------------------------------------------------------------------------------------------------------------
	for _, TopPath := range hdlr.AppRoot {
		var ignoreList []string
		fmt.Printf("TabServer2: At Search for additional sql-cfg.json files, TopPath=%s from AppRoot=%s, %s\n", TopPath, godebug.SVar(hdlr.AppRoot), godebug.LF())
		// fmt.Printf("At Search for additional sql-cfg.json files , %s\n", godebug.LF())
		// opts__TopPath := sizlib.SubstitueUserInFilePathImmediate("/Users/corwin/Projects/who-cares/app") // xyzzy from CLI -W option //xyzzy - replace ~ with home dir.
		opts__TopPath := sizlib.SubstitueUserInFilePathImmediate(TopPath)
		// fmt.Printf("TabServer2: Path ->%s<- At, %s\n", opts__TopPath, godebug.LF())
		// ignoreList = append(ignoreList, "/Users/corwin/Projects/who-cares/who-cares-server") // xyzzy from globa-cfg.json file
		dirs := sizlib.FindDirsWithSQLCfg(opts__TopPath, ignoreList)
		// fmt.Printf("TabServer2: dirs ->%s<- At, %s\n", sizlib.SVar(dirs), godebug.LF())
		fList, ok := sizlib.SearchPathAppModule(hdlr.SQLCfgFN, hdlr.AppName, dirs)
		// fmt.Printf("TabServer2: fList ->%s<- At, %s\n", sizlib.SVar(fList), godebug.LF())
		if ok {
			fmt.Fprintf(os.Stdout, "%sTabServer2: List of additional sql-cfg*.josn files found: %s server config line:%d AT, %s%s\n",
				MiscLib.ColorGreen, sizlib.SVarI(fList), hdlr.LineNo, godebug.LF(), MiscLib.ColorReset)
			if hdlr.gCfg.DbOn("*", "LoginRequired", "db2") { // Not config yet - need "global" debug config
				// if db_list_cfg_fiels {
				fmt.Fprintf(os.Stderr, "%sTabServer2: List of additional sql-cfg*.josn files found: %s server config line:%d AT, %s%s\n",
					MiscLib.ColorGreen, sizlib.SVarI(fList), hdlr.LineNo, godebug.LF(), MiscLib.ColorReset)
			}
			for _, v := range fList {
				n_files_loaded++
				fmt.Printf("TabServer2: Reading in additional SQLCfg: %s\n", v)
				// fmt.Fprintf(os.Stderr, "TabServer2: Reading in additional SQLCfg: %s\n", v)
				// if hdlr.gCfg.DbOn("*", "TabServer2", "list-cfg-fiels") {
				if db_list_cfg_fiels {
					fmt.Fprintf(os.Stderr, "%sTabServer2: Reading in additional SQLCfg: %s%s\n", MiscLib.ColorGreen, v, MiscLib.ColorReset)
				}
				tSQLCfg, err := readInSQLConfig(v) // func readInSQLConfig(path string) map[string]SQLOne {
				if err != nil {
					fmt.Printf("TabServer2: Error: %s\n", err)
					SqlCfgFilesLoaded = append(SqlCfgFilesLoaded, SqlCfgLoaded{FileName: hdlr.pwd + v[1:], ErrorMsg: fmt.Sprintf("%s", err)})
				} else {
					// 1. combine note: 'f' values - concatenate for each key - instead of overwrite -- xyzzyConcatNoteKey
					//		Collect all the note: 'f's into a set of strings - then post-process them
					preNote := make(map[string]string)
					for ii, vv := range hdlr.SQLCfg {
						if strings.HasPrefix(ii, "note:") {
							preNote[ii] = vv.F
						}
					}
					fmt.Printf("PreNote = %s\n", preNote)
					if hdlr.SQLCfg == nil {
						hdlr.SQLCfg = make(map[string]SQLOne)
					}
					for j, w := range tSQLCfg {
						hdlr.SQLCfg[j] = w
					}
					for ii, vv := range hdlr.SQLCfg {
						if strings.HasPrefix(ii, "note:") {
							if old, ok := preNote[ii]; ok {
								// preNote[ii] = vv.F
								vv.F = old + "\n" + vv.F
								hdlr.SQLCfg[ii] = vv
							}
						}
					}
					SqlCfgFilesLoaded = append(SqlCfgFilesLoaded, SqlCfgLoaded{FileName: hdlr.pwd + v[1:], ErrorMsg: ""})
					n_config++
				}
			}
		}
	}

	if n_files_loaded == 0 {
		fmt.Fprintf(os.Stderr, "%sTabServer2: Error (14122):  Unable to find the %s file using %s path. AppName=%s LineNo:%d in server config file.%s\n", MiscLib.ColorRed, hdlr.SQLCfgFN, hdlr.SearchPath, hdlr.AppName, hdlr.LineNo, MiscLib.ColorReset)
		fmt.Printf("TabServer2: Error (14122):  Unable to find the %s file using %s path. LineNo:%d in server config file.\n", hdlr.SQLCfgFN, hdlr.SearchPath, hdlr.LineNo)
	}

	// xyzzy - valid that the sql_cfg.json data is correct - check table/column info
	if !hdlr.CheckSqlCfgValid() {
		fmt.Printf("Early exit - sql_cfg.json is not valid\n")
		n_config = -1
	}

	// xyzzy - put this back in -- loadAllCsrfTokens(hdlr)

	hdlr.theMux = goftlmux.NewRouter()
	// fmt.Fprintf(os.Stderr, "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! theMux hdlr=%p hdlr.theMux=%p\n", hdlr, hdlr.theMux)

	initEndPoints(hdlr.theMux, hdlr)

	hdlr.final, err = lib.ParseBool(hdlr.Final)

	if n_config == 0 {
		fmt.Printf("\n************************************************\n* Warning - no TabServer2 config files loaded\n ************************************************\n\n")
		return mid.ErrInternalError
	}
	return
}

var _ mid.GoFTLMiddleWare = (*TabServer2Type)(nil) // compile time validation that this matches with the GoFTLMiddleWare interface

// --------------------------------------------------------------------------------------------------------------------------

type TabServer2Type struct {
	Next                  http.Handler                // No Next, this is the bottom of the stack.
	Paths                 []string                    //
	AppRoot               []string                    // The patth where to start searhing for sql-cfg-<name>.json files -- Formerly TopPath
	DbFunctions           []string                    // Functions to turn on debugging output in.
	WatchForConfigChanges bool                        // If true then a 2nd process will be started to watch for changes in config files.
	SQLCfgFN              string                      // xyzzy
	AppName               string                      // xyzzy
	SearchPath            string                      // a search path like "~/cfg:./cfg" -- defaults to ./cfg:.:~/cfg
	Final                 string                      // If a path is matched with Paths, but not with the final routing, say /api/table/BADNAME, then if Final => 404 error
	DevAuthToken          string                      // Password for accessing ListSqlConfigFilesLoaded
	DebugFlags            []string                    // Debuging Flags for example, "credit_card_test_mode"
	AuthorizeNetLogin     string                      //
	AuthorizeNetKey       string                      //
	StatusMessage         string                      // Message printed out as a part of status - can be version number for the config file
	LogToFile             string                      // if "" then logging is off, else the path to log directory
	LoginSystem           string                      // "AesSrp" or "Basic" or "Un/Pw"
	ApiTable              string                      //
	ApiList               string                      //
	ApiStatus             string                      //
	ApiTableKey           string                      // If true (!= "") then this password will be requried to access /api/table calls.
	RedisApiTableKey      string                      // If used (!= "") then lookup the password in redis using this as the prefix for the key.
	LimitPostJoinRows     int                         // -1 indicates unlimited, default, 0 - is do not allow, N is maximum number of rows to post-join on get
	SendStatusOnError     bool                        // if true will send back errors as a status_code, else will send "status"=="error" in JSON, code 200
	DbSchema              string                      // // 1. g_schema // xyzzy - should pull from config "public"
	DbCreateScript        string                      // // 2. xyzzyPath1 // xyzzyPath1 - should pull from config ./table_ddl
	RedisSessionPrefix    string                      // -- use the same namespace as SessionRedis!
	EmailConfigFileName   string                      // name of file to take Email config from
	EmailTemplateDir      string                      //
	KeyFilePrivate        string                      // private key file for signing JWT tokens
	DisplayURL2fa         string                      // 2fa - see X2faSetup
	RedisPrefix2fa        string                      // 2fa - see X2faSetup
	Server2faURL          string                      // 2fa see X2faSetup
	LineNo                int                         //
	gCfg                  *cfg.ServerGlobalConfigType //
	MuxAuto               map[string]int              // formerly global		-- make private xyzzy --		// Config for automaic reload  - to delete routes removed
	MuxAutoPass           int                         // formerly global		-- make private xyzzy --		// Config for automaic reload  - to delete routes removed
	db_func               map[string]bool             //
	pwd                   string                      // Running path of the server
	theMux                *goftlmux.MuxRouter         //
	final                 bool                        // t/f converted version of .Final
	loginSystem           LoginSystemType             //
	SQLCfg                map[string]SQLOne
}

func NewTabServer2Server(n http.Handler, Path []string, AppRoot []string, gCfg *cfg.ServerGlobalConfigType) (rv *TabServer2Type) {

	tt, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sTabServer2: Error (14022):  Unable to get current working directory. LineNo:%s.%s\n", MiscLib.ColorRed, godebug.LF(2), MiscLib.ColorReset)
		fmt.Printf("TabServer2: Error (14022):  Unable to get current working directory. LineNo:%s.\n", godebug.LF(2))
		os.Exit(1)
	}

	rv = &TabServer2Type{
		Next:         n, // may be NIL!
		Paths:        Path,
		AppRoot:      AppRoot,
		MuxAuto:      make(map[string]int),
		MuxAutoPass:  1,
		loginSystem:  LstAesSrp,
		ApiTable:     "/api/table",
		ApiList:      "/api/list",
		LoginSystem:  "LstAesSrp",
		DevAuthToken: "9abb4f75-f336-46d2-a3af-1115c3d49f14",
		db_func:      make(map[string]bool),
		pwd:          tt,
		gCfg:         gCfg,
	}

	return
}

/*

-- Redis ------------------------------------------------------------------------------------------------------------------------------

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		rw.Log.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	// ... do stuff ..
	data, err := conn.Cmd("GET", aKey).Str() // Get the value

	hdlr.gCfg.RedisPool.Put(conn)

-- PostgreSQL -------------------------------------------------------------------------------------------------------------------------

	Rows, err := hdlr.gCfg.Pg_client.Query(Query, data...)

*/

func initEndPoints(theMux *goftlmux.MuxRouter, hdlr *TabServer2Type) {
	// pull all of these from the sql-*.json file - and then track changes to file. -- Will set up end points for all the one starting with '/'
	hdlr.MuxAuto = make(map[string]int, len(hdlr.SQLCfg))
	hdlr.MuxAutoPass = 1

	//api_table := "/api/table/"
	//api_list := "/api/list/"
	AddSlash := func(s string) string {
		if len(s) == 0 {
			return ""
		} else if s[len(s)-1] == '/' {
			return s
		}
		return s + "/"
	}
	api_table := AddSlash(hdlr.ApiTable)
	api_list := AddSlash(hdlr.ApiList)
	api_status := AddSlash(hdlr.ApiStatus)

	fmt.Printf("api_table [%s] api_list [%s] api_status [%s]\n", api_table, api_list, api_status)

	for key, val := range hdlr.SQLCfg {
		// xyzzy ------------------------------------------------------------------------------------------------------------------------------------------------
		// match agains hdlr.Paths at this point!
		// xyzzy ------------------------------------------------------------------------------------------------------------------------------------------------
		if pn := lib.PathsMatchN(hdlr.Paths, key); pn >= 0 {
			fmt.Printf("%sTabServer2: Matched: %s%s\n", MiscLib.ColorGreen, key, MiscLib.ColorReset)
			if len(key) > 0 && key[0:1] == "/" && (!strings.HasPrefix(key, api_table) || len(val.Crud) == 0) && !val.Redis {
				hdlr.MuxAuto[key] = hdlr.MuxAutoPass
				if debugCrud01 {
					fmt.Printf("Creating %s for %v\n", key, val.Method)
				}
				if len(val.Method) == 0 {
					theMux.HandleFunc(key, GetSqlCfgHandler2(key, hdlr)).Methods("GET").AppendFileName(":FromData(" + val.LineNo + ")")
				} else {
					theMux.HandleFunc(key, GetSqlCfgHandler2(key, hdlr)).Methods(val.Method...).AppendFileName(":FromData(" + val.LineNo + ")")
				}
			} else if len(key) > 0 && key[0:1] == "/" && val.Redis {
				hdlr.MuxAuto[key] = hdlr.MuxAutoPass
				if len(val.Method) == 0 {
					theMux.HandleFunc(key, GetRedisCfgHandler2(key, hdlr)).Methods("GET").AppendFileName(":FromData(" + val.LineNo + ")")
				} else {
					theMux.HandleFunc(key, GetRedisCfgHandler2(key, hdlr)).Methods(val.Method...).AppendFileName(":FromData(" + val.LineNo + ")")
				}
			}
		} else {
			fmt.Printf("%sTabServer2: Skipped: %s%s\n", MiscLib.ColorBlue, key, MiscLib.ColorReset)
		}
	}

	theMux.DebugMatch(true)
	// -------------------------- CRUD Handlers -------------------------------------------------------------------
	if pn := lib.PathsMatchN(hdlr.Paths, api_table); pn >= 0 {
		fmt.Printf("%sTabServer2: Matched: %s... builtins%s\n", MiscLib.ColorGreen, api_table, MiscLib.ColorReset)
		theMux.HandleFunc(api_table+"{name}/count", closure_respHandlerTableGetCount(hdlr)).Methods("GET")                     // Select count(*)
		theMux.HandleFunc(api_table+"{name}/desc", closure_respHandlerTableDesc(hdlr)).Methods("GET")                          // Describe
		theMux.HandleFunc(api_table+"{name}/{id}", closure_respHandlerTableGetPk1(hdlr)).Methods("GET").Comment("TableGetPk1") // Select - with single unique PK id - Not fond of positional param
		theMux.HandleFunc(api_table+"{name}/{id}", closure_respHandlerTableDelPk1(hdlr)).Methods("DELETE")                     // Delete - with single unique PK id - Not fond of positional param
		theMux.HandleFunc(api_table+"{name}/{id}", closure_respHandlerTablePutPk1(hdlr)).Methods("PUT")                        // Update
		theMux.HandleFunc(api_table+"{name}/{id}", closure_respHandlerTablePostPk1(hdlr)).Methods("POST")                      // Insert
		theMux.HandleFunc(api_table+"{name}", closure_respHandlerTableGet(hdlr)).Methods("GET").Comment("TableGet")            // Select
		theMux.HandleFunc(api_table+"{name}", closure_respHandlerTablePut(hdlr)).Methods("PUT")                                // Update
		theMux.HandleFunc(api_table+"{name}", closure_respHandlerTablePost(hdlr)).Methods("POST")                              // Insert
		theMux.HandleFunc(api_table+"{name}", closure_respHandlerTableDel(hdlr)).Methods("DELETE")                             // Delete
	} else {
		fmt.Printf("%sTabServer2: Skipped: %s%s\n", MiscLib.ColorBlue, api_table, MiscLib.ColorReset)
	}
	if pn := lib.PathsMatchN(hdlr.Paths, api_list); pn >= 0 {
		fmt.Printf("%sTabServer2: Matched: %s... builtins%s\n", MiscLib.ColorGreen, api_list, MiscLib.ColorReset)
		theMux.HandleFunc(api_list+"sql-cfg-files-loaded", closure_respHandlerListSQLCfgFilesLoaded(hdlr)).Methods("GET") //
		theMux.HandleFunc(api_list+"cfg-for", closure_respHandlerListCfgFor(hdlr)).Methods("GET")                         //
		theMux.HandleFunc(api_list+"end-points", closure_respHandlerListEndPoints(hdlr)).Methods("GET")                   //

		// -------------------------- From base.go --------------------------------------------------------------------
		theMux.HandleFunc(api_list+"swapLogFile/{seq}", closure_respHandlerSwapLogFile(hdlr)).Methods("GET", "POST")       // !!depricated!! -- goging to loggin component -- irrilevant
		theMux.HandleFunc(api_list+"reloadTableConfig", closure_respHandlerReloadTableConfig(hdlr)).Methods("GET", "POST") // research-and load sql-cfg*.* files
		theMux.HandleFunc(api_list+"builtin-routes", closure_respHandlerListBuiltinRoutes(hdlr)).Methods("GET")            // List back all the routes
		theMux.HandleFunc(api_list+"grabFeedback", respHandlerGrabFeedback).Methods("GET")                                 // !!depricated!! /api/grabFeeddback instead -DB! Feedback palced in log
		theMux.HandleFunc(api_list+"logit", respHandlerLogIt).Methods("GET", "POST")                                       // !!depricated!! /api/loggit instead - DB! Log information log files
		theMux.HandleFunc(api_list+"installed-themes", closure_respHandlerListInstalledThemes(hdlr)).Methods("GET")        // DB! Find the set of installed themes
		theMux.HandleFunc(api_list+"current-theme", closure_respHandlerListCurrentTheme(hdlr)).Methods("GET")              // DB! Find the currently set theme
	} else {
		fmt.Printf("%sTabServer2: Skipped: %s%s\n", MiscLib.ColorBlue, api_list, MiscLib.ColorReset)
	}
	if pn := lib.PathsMatchN(hdlr.Paths, api_status); pn >= 0 {
		fmt.Printf("%sTabServer2: Matched: %s... builtins%s\n", MiscLib.ColorGreen, api_status, MiscLib.ColorReset)
		theMux.HandleFunc(api_status+"tab-server2/status", closure_respHandlerStatus(hdlr)).Methods("GET", "POST", "HEAD", "PATCH", "PUT", "DELETE", "OPTIONS")
	} else {
		fmt.Printf("%sTabServer2: Skipped: %s%s\n", MiscLib.ColorBlue, api_status, MiscLib.ColorReset)
	}
}

func (hdlr *TabServer2Type) GetRedisKey(key string) (rv string, err error) {
	fmt.Fprintf(os.Stderr, "AT: %s, key=%s\n", godebug.LF(), key)
	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	// key := hdlr.RedisAuthTokenPrefix + auth_token

	val, err := conn.Cmd("GET", key).Str()
	if err != nil {
		return
	}
	rv = val
	return
}

// -----------------------------------------------------------------------------------------------------------------------------------------------------------------------------
func (hdlr *TabServer2Type) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "TabServer2", hdlr.Paths, pn, req.URL.Path)

			fmt.Printf("In TabServer2, hdlr.ApiTableKey=%s %s\n", hdlr.ApiTableKey, godebug.LF())

			// ----------------------------------------------------------------------------------------------------------------------------------------
			// xyzzy - Use redis hdlr.RedisApiTableKey != "" => check in redis using this as the prefix
			// RedisApiTableKey = "atk:" => lookup that key in redis, if "y" then it is a match.
			// ----------------------------------------------------------------------------------------------------------------------------------------
			if hdlr.RedisApiTableKey != "" { // defaults to "324d4b9f-00dc-4ea9-7a6c-e5f125207759" , so not normally null
				ps := &rw.Ps
				pwSupplied := ps.ByNameDflt("api_table_key", "")
				rv, err := hdlr.GetRedisKey(hdlr.RedisApiTableKey + pwSupplied)
				if hdlr.final || hdlr.Next == nil {
					if err != nil {
						trx.AddNote(1, "TabServer2: final - return - 406")
						logrus.Errorf("406 api_table_key did not match key in redis database AT: %s", godebug.LF())
						www.WriteHeader(http.StatusNotAcceptable) // 406
						return
					}
					if rv != "yes" {
						trx.AddNote(1, "TabServer2: final - return - 406")
						logrus.Errorf("406 api_table_key found in redis but is not 'yes', value=%s AT:%s", rv, godebug.LF())
						www.WriteHeader(http.StatusNotAcceptable) // 406
						return
					}
				} else {
					fmt.Printf("In TabServer2 - RedisApiTableKey used and passed - not final, %s\n", godebug.LF())
					hdlr.Next.ServeHTTP(www, req)
				}
			} else if hdlr.ApiTableKey != "" { // defaults to "324d4b9f-00dc-4ea9-7a6c-e5f125207759" , so not normally null
				ps := &rw.Ps
				pwSupplied := ps.ByNameDflt("api_table_key", "")
				fmt.Fprintf(os.Stderr, "%sApiTableKey[%s] is not \"\", so will match v.s. [%s], final=%v hdlr.Next==Nil = %v, %s%s\n",
					MiscLib.ColorGreen, hdlr.ApiTableKey, pwSupplied, hdlr.final, (hdlr.Next == nil), godebug.LF(), MiscLib.ColorReset)
				if hdlr.ApiTableKey != pwSupplied {
					if hdlr.final || hdlr.Next == nil {
						trx.AddNote(1, "TabServer2: final - return - 406")
						logrus.Errorf("406 api_table_key did not match required key: %s", godebug.LF())
						www.WriteHeader(http.StatusNotAcceptable) // 406
					} else {
						fmt.Printf("In TabServer2 - ApiTableKey used and passed - not final, %s\n", godebug.LF())
						hdlr.Next.ServeHTTP(www, req)
					}
					return
				}
				fmt.Printf("In TabServer2 - ApiTableKey used and matched, %s\n", godebug.LF())
			}

			fmt.Printf("In TabServer2 - RedisApiTableKey or ApiTableKey passed - will lookup path [%s], AT:%s\n", req.URL.Path, godebug.LF())

			found, err := hdlr.theMux.MatchAndServeHTTP(www, req)
			if found {
				fmt.Printf("TabServer2: matched and served, %s\n", godebug.LF())
			} else {
				fmt.Printf("TabServer2: did not match - not served, [%s] AT:%s\n", req.URL.Path, godebug.LF())
				fmt.Fprintf(os.Stderr, "%sTabServer2: input path [%s] did not match any available paths - not served, AT:%s%s\n",
					MiscLib.ColorRed, req.URL.Path, godebug.LF(), MiscLib.ColorReset)
				hdlr.theMux.DumpRouteData(fmt.Sprintf("%s-={Available Routes}=-%s", MiscLib.ColorGreen, MiscLib.ColorReset))
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrMuxError, MiscLib.ColorReset)
				fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, err, MiscLib.ColorReset)
				logrus.Errorf("Error: %s - %s, %s", mid.ErrMuxError, err, godebug.LF())
				trx.ErrorReturn(1, err)
				www.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !found {
				if !hdlr.final {
					trx.AddNote(1, "TabServer2: not final - so calling next")
					fmt.Printf("In TabServer2 - not final - so calling next, %s\n", godebug.LF())
					hdlr.Next.ServeHTTP(rw, req)
				} else {
					// fmt.Fprintf(rw, "%s\n", lib.SVarI(req))	// print out entire request to take a look at it
					trx.AddNote(1, "TabServer2: final - return - 404")
					fmt.Printf("In TabServer2 - final - so 404 , %s\n", godebug.LF())
					logrus.Errorf("404 att: %s", godebug.LF())
					www.WriteHeader(http.StatusNotFound) // 404
				}
			}
			return

		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
			www.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		if hdlr.Next != nil {
			hdlr.Next.ServeHTTP(www, req)
		} else {
			logrus.Errorf("404 att: %s", godebug.LF())
			www.WriteHeader(http.StatusNotFound)
		}
	}

}

type DbColumnsType struct {
	ColumnName string
	DBType     string
	TypeCode   string
	MinLen     int
	MaxLen     int
}

type DbTableType struct {
	TableName string
	DbColumns []DbColumnsType
}

func (hdlr *TabServer2Type) CheckSqlCfgValid() (isOk bool) {
	for endPointName, vv := range hdlr.SQLCfg {
		if vv.TableName != "" {
			fmt.Printf("\n-----------------------------------------------------------------------\n")
			fmt.Printf("Checking Table [%s] for endpoint: %s\n", vv.TableName, endPointName)
			// 1. Read in table - if err then table did not exist - check for spelling errors on table name
			// 	if error on (1) then return
			// 2. Check each column - that it exists - if not then err
			// xyzzyPostDb Checks -- xyzzy - at this point --
			//if false { // xyzzy - database connection is not setup at this point -- check has to be deferred to later
			//	TableInfo := hdlr.GetTableInformationSchema(vv.TableName)
			//	_ = TableInfo
			//}
			var TableName = vv.TableName
			var EndPoint = endPointName
			var TheCols = vv.Cols
			cfg.PostDbConnectChecks = append(cfg.PostDbConnectChecks, cfg.PostDbType{RunCheck: func(conn *sizlib.MyDb) bool {
				TableInfo, err := hdlr.GetTableInformationSchema(conn, TableName)
				if err != nil {
					return false
				}
				fmt.Printf("Doing Check for %s : %s\n", TableName, EndPoint)
				if !ValidateTableCols(TheCols, TableInfo) {
					return false
				}
				return true
			}})
		}
		if vv.G != "" {
			// godebug.Db2Printf(db84, "Top Loop Function [%s], params %s %s\n", vv.G, vv.P, godebug.LF())
			fmt.Printf("\n-----------------------------------------------------------------------\n")
			fmt.Printf("Checking Function/Procedure [%s] for endpoint: %s, params %s\n", vv.G, endPointName, vv.P)
			/*
				SELECT routines.routine_name, parameters.data_type, parameters.ordinal_position
				FROM information_schema.routines
					JOIN information_schema.parameters ON routines.specific_name=parameters.specific_name
				WHERE routines.specific_schema='public'
				ORDER BY routines.routine_name, parameters.ordinal_position;
			*/
			var FunctionName = vv.G
			var EndPoint = endPointName
			var TheParams = vv.P
			var Valid = vv.Valid
			cfg.PostDbConnectChecks = append(cfg.PostDbConnectChecks, cfg.PostDbType{RunCheck: func(conn *sizlib.MyDb) bool {
				FunctionInfo, err := hdlr.GetFunctionInformationSchema(conn, FunctionName, TheParams)
				if err != nil {
					return false
				}
				fmt.Printf("Doing Check for %s %s : %s\n", FunctionName, TheParams, EndPoint)
				chkExtra := make(map[string]bool)
				for name := range Valid {
					chkExtra[name] = false
				}
				for ii, pp := range TheParams {
					ok := false
					chkExtra[pp] = true
					for name := range Valid {
						if name == pp {
							ok = true
							break
						}
					}
					if !ok {
						godebug.Db2Printf(db84, "%sAt: %s Valid is wrong - endpoint[%s] missing[%s] at pos=%d in .P[], %s\n", MiscLib.ColorRed, godebug.LF(),
							EndPoint, pp, ii, MiscLib.ColorReset)
						fmt.Fprintf(os.Stderr, "%sMessage (41990): Valid is wrong - endpoint[%s] missing[%s] at pos=%d in .P[], %s\n", MiscLib.ColorRed,
							EndPoint, pp, ii, MiscLib.ColorReset)
						return false
					}
				}
				for key, val := range chkExtra {
					if !val && key != "callback" { // callback is used by JSONp!
						fmt.Fprintf(os.Stderr, "%sNote (41991): Valid has extra, unused field - endpoint[%s] extra[%s], %s\n", MiscLib.ColorCyan,
							EndPoint, key, MiscLib.ColorReset)
					}
				}
				if len(TheParams) != len(FunctionInfo.DbColumns) {
					godebug.Db2Printf(db84, "%sAt: %s Mismatch in number of params for function, expected %d(db) have %d, %s\n", MiscLib.ColorRed, godebug.LF(),
						len(TheParams), len(FunctionInfo.DbColumns), MiscLib.ColorReset)
					fmt.Fprintf(os.Stderr, "%sMessage (81407): EndPoint [%s] Function [%s] number of columns mismatch config(JsonX) expected %d PostresSQL has %d\n%s", MiscLib.ColorRed,
						EndPoint, FunctionName, len(TheParams), len(FunctionInfo.DbColumns), MiscLib.ColorReset)
					return false
				}
				//	if !ValidateFunctionCols(FunctionInfo, TheParams) {
				//		return false
				//	}
				return true
			}})
		}
	}
	isOk = true
	return
}

/*
1. g_schema // - should pull from config "public"
2. // xyzzyPath1 - should pull from config ./table_ddl
	DbSchema              string                      // // 1. g_schema // - should pull from config "public"
	DbCreateScript        string                      // // 2. - should pull from config ./table_ddl
// var g_schema string = "public"             // - should pull from config "public"
// var g_pathFmt string = "./table_ddl/%s.jx" // xyzzyPath1
*/

func searchSub(dt *SqlEr.SqlErType, lookFor string) (sub int, found bool) {
	for ii, vv := range dt.Sql.Tables {
		if vv.TableName == lookFor {
			sub = ii
			found = true
			return
		}
	}
	return
}

func (hdlr *TabServer2Type) CreateMissingTable(conn *sizlib.MyDb, tn string) (err error) {
	// fn := fmt.Sprintf(g_pathFmt, tn)
	mdata := make(map[string]string)
	mdata["TableName"] = tn
	fn := tmplp.ExecuteATemplate(hdlr.DbCreateScript, mdata)
	godebug.Db2Printf(db83, "%sAt: %s - check exists on fn:%s%s\n", MiscLib.ColorCyan, godebug.LF(), fn, MiscLib.ColorReset)
	// fn := fmt.Sprintf(hdlr.DbCreateScript, tn) // - replace with template call!!!
	// should be a single file - with multiple tables in it
	if sizlib.Exists(fn) {
		godebug.Db2Printf(db83, "%sAt: %s - found exists on fn:%s%s\n", MiscLib.ColorCyan, godebug.LF(), fn, MiscLib.ColorReset)
		se := SqlEr.ValidateInstallModel(fn)
		sub, ok := searchSub(se, tn)
		godebug.Db2Printf(db83, "%sAt: %s - searchSub = ok=%v, data=%s%s\n", MiscLib.ColorCyan, godebug.LF(), ok, JsonX.SVarI(sub), MiscLib.ColorReset)
		if ok {
			se.RunSliceOfSQLCommands(conn, se.Sql.Tables[sub].DDL) // [ array -lookup subscript- ]
		} else {
			// fmt.Printf("Did not find %s in table creation script\n", tn)
			fmt.Printf("%sAt: %s - Error did not find %s in table creation script%s\n", MiscLib.ColorRed, godebug.LF(), tn, MiscLib.ColorReset)
			err = fmt.Errorf("Unable to find table:%s in creation script.", tn)
		}
	} else {
		err = fmt.Errorf("Unable to find table:%s creation script missing", tn)
	}
	return
}

/*
	SELECT routines.routine_name, parameters.data_type, parameters.ordinal_position
	FROM information_schema.routines
		JOIN information_schema.parameters ON routines.specific_name=parameters.specific_name
	WHERE routines.specific_schema='public'
	ORDER BY routines.routine_name, parameters.ordinal_position;
*/
func (hdlr *TabServer2Type) GetFunctionInformationSchema(conn *sizlib.MyDb, FunctionName string, Params []string) (rv DbTableType, err error) {
	godebug.Db2Printf(db84, "Validateing Function [%s], params %s %s\n", FunctionName, Params, godebug.LF())

	// check that the function exists

	qry := `SELECT routines.routine_name
			FROM information_schema.routines
			WHERE routines.specific_schema = $1
			  and ( routines.routine_name = lower($2)
			     or routines.routine_name = $2
				  )
	`
	data := sizlib.SelData(conn.Db, qry, hdlr.DbSchema, FunctionName)
	if data == nil || len(data) == 0 {
		fmt.Fprintf(os.Stderr, "%sMessage (91532): Missing function:%s%s\n", MiscLib.ColorRed, FunctionName, MiscLib.ColorReset)
		fmt.Printf("Message(91532): Missing function:%s\n", FunctionName)
		//	err = hdlr.CreateMissingFunction(conn, FunctionName) // - attempt to create table now!
		//	if err != nil {
		//		err = fmt.Errorf("Missing Function:%s, unable to create", FunctionName)
		//	}
		return
	}
	fmt.Fprintf(os.Stderr, "%sFound function: %s%s\n", MiscLib.ColorGreen, FunctionName, MiscLib.ColorReset)
	fmt.Printf("Function [%s] found in schema [%s]\n", data[0]["routine_name"], hdlr.DbSchema)
	rv.TableName = FunctionName

	// get parametrs now
	qry = `SELECT routines.routine_name
				, parameters.data_type
				, parameters.parameter_name
				, parameters.ordinal_position
			FROM information_schema.routines
				JOIN information_schema.parameters ON routines.specific_name=parameters.specific_name
			WHERE routines.specific_schema = $1
			  and ( routines.routine_name = lower($2)
			     or routines.routine_name = $2
				  )
			ORDER BY routines.routine_name, parameters.ordinal_position;
	`
	cols := sizlib.SelData(conn.Db, qry, hdlr.DbSchema, FunctionName)
	fmt.Printf("params=%s\n", lib.SVarI(cols))
	for _, vv := range cols {
		rv.DbColumns = append(rv.DbColumns, DbColumnsType{
			ColumnName: vv["parameter_name"].(string),
			DBType:     vv["data_type"].(string),
			TypeCode:   GetTypeCode(vv["data_type"].(string)),
		})
	}
	godebug.Db2Printf(db84, "rv=%s\n", lib.SVarI(rv))
	return
}

func (hdlr *TabServer2Type) GetTableInformationSchema(conn *sizlib.MyDb, TableName string) (rv DbTableType, err error) {
	godebug.Db2Printf(db83, "Validateing [%s], %s\n", TableName, godebug.LF())
	qry := `SELECT * FROM information_schema.tables WHERE table_schema = $1 and table_name = $2`
	data := sizlib.SelData(conn.Db, qry, hdlr.DbSchema, TableName)
	if data == nil || len(data) == 0 {
		fmt.Fprintf(os.Stderr, "%sMessage(90532): Missing table:%s%s\n", MiscLib.ColorRed, TableName, MiscLib.ColorReset)
		fmt.Printf("Message(90532): Missing table:%s\n", TableName)
		err = hdlr.CreateMissingTable(conn, TableName) // - attempt to create table now!
		if err != nil {
			err = fmt.Errorf("Missing Table:%s, unable to create", TableName)
		}
		return
	}
	fmt.Fprintf(os.Stderr, "%sFound table: %s%s\n", MiscLib.ColorGreen, TableName, MiscLib.ColorReset)
	fmt.Printf("Table [%s] found in schema [%s]\n", data[0]["table_name"], data[0]["table_schema"])
	rv.TableName = TableName

	// get columns now
	qry = `SELECT * FROM information_schema.columns WHERE table_schema = $1 and table_name = $2`
	// cols := sizlib.SelData(conn.Db, qry, g_schema, TableName)
	cols := sizlib.SelData(conn.Db, qry, hdlr.DbSchema, TableName)

	fmt.Printf("data=%s\n", lib.SVarI(data))
	fmt.Printf("cols=%s\n", lib.SVarI(cols))
	for _, vv := range cols {
		rv.DbColumns = append(rv.DbColumns, DbColumnsType{
			ColumnName: vv["column_name"].(string),
			DBType:     vv["data_type"].(string),
			TypeCode:   GetTypeCode(vv["data_type"].(string)),
		})
	}
	godebug.Db2Printf(db83, "rv=%s\n", lib.SVarI(rv))
	return
}

func GetTypeCode(ty string) (rv string) {
	rv = "?"
	switch ty {
	case "character varying", "text":
		return "s"
	case "number":
		return "i"
	}
	if strings.HasPrefix(ty, "timestamp") {
		return "d"
	}
	return
}

func ValidateTableCols(TheCols []ColSpec, TableInfo DbTableType) (rv bool) {
	rv = true
	for _, vv := range TheCols {
		if pp := HaveColumn(vv.ColName, TableInfo); pp == -1 {
			fmt.Fprintf(os.Stderr, "%sMessage (90577): Missing Column [%s] in table [%s]%s\n", MiscLib.ColorRed, vv.ColName, TableInfo.TableName, MiscLib.ColorReset)
			fmt.Printf("Message (90577): Missing Column [%s] in table [%s]\n", vv.ColName, TableInfo.TableName)
			rv = false
		}
	}
	return
}

func HaveColumn(ColumnName string, TableInfo DbTableType) (rv int) {
	rv = -1
	for ii, vv := range TableInfo.DbColumns {
		if ColumnName == vv.ColumnName {
			rv = ii
			return
		}
	}
	return
}

const db3 = true
const db83 = true // table validation and columns
const db84 = true // function validation with parameters

const db_list_cfg_fiels = true

/* vim: set noai ts=4 sw=4: */
