//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2017
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1197
//

package mid

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	JsonX "github.com/pschlump/JSONx"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/sizlib" //
	"github.com/pschlump/Go-FTL/server/tmplp"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/Go-FTL/server/urlpath"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/check-json-syntax/lib"
	"github.com/pschlump/godebug"
	// "github.com/pschlump/json" //	"encoding/json"
	"github.com/pschlump/uuid"
	// Modified pool to have NewAuth for authorized connections
)

type GoFTLMiddleWare interface {
	// InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (rv GoFTLMiddleWare, err error)
	InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error)
	PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error)
	ServeHTTP(www http.ResponseWriter, req *http.Request)
}

// SetNext(next http.Handler)
// OnteTimeInit(cfgData map[string]interface{}, callNo int) error
//							err = dataOfType.InitalizeWithUserValues(cfg.ServerGlobal, pluginName, ii, cfg.NewInit3[jj].CallNo)

// -------------------------------------------------------------------------------------------------------------------------------------------------
// new interface based
// -------------------------------------------------------------------------------------------------------------------------------------------------
type CreateEmptyFx3 func(name string) GoFTLMiddleWare

type GoFtlHttpServer struct {
	Name        string         // Name of this (the directive this is called by
	ValidJSON   string         // JSONP/JsonX validaiton string for config for this item
	CreateEmpty CreateEmptyFx3 //
	CallNo      int            //
}

var NewInit3 []GoFtlHttpServer

//type GoFtlServerInteface interface {
//	InitNext(next http.Handler, gCfg *ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error)
//	OnteTimeInit(cfgData map[string]interface{}, callNo int) error
//	// FinializeHandler(next http.Handler, gCfg *ServerGlobalConfigType, serverName string, pNo int) (rv http.Handler, err error)
//}

// /Users/corwin/go/src/github.com/pschlump/Go-FTL/server/cfg/cfg.go
func RegInitItem3(name string, fx CreateEmptyFx3, valid string) {
	NewInit3 = append(NewInit3, GoFtlHttpServer{Name: name, CreateEmpty: fx, ValidJSON: valid})
}

func LookupInitByName3(name string) (p int) {
	p = -1
	for ii := range NewInit3 {
		if NewInit3[ii].Name == name {
			return ii
		}
	}
	return
}

//				pluginList := mid.LookupPluginList()
func LookupPluginList() (rv string) {
	var Items []string
	for ii := range NewInit3 {
		Items = append(Items, NewInit3[ii].Name)
	}
	sort.Strings(Items)
	for _, name := range Items {
		rv += "\t" + name + "\n"
	}
	return
}

func ReadConfigFile2(fn string) {
	// Note: best test for this is in the TabServer2 - test 0001 - checks that this works.

	xyzzyJsonX := false
	RawConfig := make(map[string]interface{})

	if xyzzyJsonX {

		// xyzzy - chagne to use options and turn on `json:` as an option.

		////fmt.Printf("At: %s\n", lib.LF())
		meta, err := JsonX.UnmarshalFile(fn, &RawConfig)
		_ = meta
		lib.IsErrFatal(err) // all errors are fatal, print, exit if error

		// xyzzyJsonX - print out errors, eval if fatal!

	} else {

		file, err := sizlib.ReadJSONDataWithComments(fn)
		lib.IsErrFatal(err)

		////fmt.Printf("At: %s\n", lib.LF())
		err = json.Unmarshal(file, &RawConfig)
		if err != nil {
			es := jsonSyntaxErroLib.GenerateSyntaxError(string(file), err)
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorYellow, es, MiscLib.ColorReset)
			logrus.Errorf("Error: Invlaid JSON for %s %s Error:\n%s\n", fn, file, es)
			lib.IsErrFatal(err) // all errors are fatal, print, exit if error
		}
	}

	if cfg.ServerGlobal.Config == nil {
		cfg.ServerGlobal.Config = make(map[string]cfg.PerServerConfigType)
	}

	for name, v := range RawConfig {

		vv := v.(map[string]interface{}) // vv is a map[string]interface{}
		if db_g1 {
			fmt.Printf("Configuration for >%s< typeof vv = %T\n", name, vv)
		}
		// perServerConfig := PerServerConfigType{IndexFileList: []string{"index.html", "index.tmpl"}}
		perServerConfig := cfg.PerServerConfigType{}
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
					locInTab := LookupInitByName3(nameOfPlugin)
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
		cfg.ServerGlobal.Config[name] = perServerConfig
		//fmt.Printf("At: %s\n", lib.LF())
	}
}

// ----------------------------------------------------------------------------------------------------------------------------------------------------
// The "top" handler is a special handler that sets up the calls below it and handles errors at the very end.
// ----------------------------------------------------------------------------------------------------------------------------------------------------

type TopHandler struct {
	Format string
	Root   []string
	Next   http.Handler // Required field for all chaining of middleware.
}

func NewTopHandler(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) GoFTLMiddleWare {
	cfg.ServerGlobal.ConnectToRedis()
	return &TopHandler{Next: next, Format: "Error: {{.Error}} From: TopHandler "}
}

func (hdlr *TopHandler) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *TopHandler) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var NormalUnused []goftlmux.UnusedParam

func init() {
	NormalUnused = []goftlmux.UnusedParam{
		{
			Match: "^_.*_$",
			IsRe:  true,
		},
		{
			Match: "^\\$.*\\$$",
			IsRe:  true,
		},
		{
			Match: "X-Go-FTL-Trx-Id",
		},
	}
	goftlmux.SetupUnsedParam(NormalUnused)
}

func (hdlr *TopHandler) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	rw := goftlmux.NewMidBuffer(www, hdlr) // can get rid of memory alloc at this point - just declare and init on stack.

	if dbTop {
		fmt.Fprintf(os.Stderr, "\n%s---------------------------- Top Handler ----------------------------\n%s\n", MiscLib.ColorCyan, MiscLib.ColorReset)
	}

	fmt.Printf("req.URL.RawQuery [%s] req.URL.Path [%s], %s\n", req.URL.RawQuery, req.URL.Path, godebug.LF())

	var id string
	var trx_id_found = false
	Ck := req.Cookies()
	for _, v := range Ck {
		if v.Name == "X-Go-FTL-Trx-Id" {
			trx_id_found = true
			id = v.Value
			break
		}
	}

	if !trx_id_found {
		id0, _ := uuid.NewV4()
		id = id0.String()
	}

	trx := tr.NewTrx(cfg.ServerGlobal.RedisPool)  // Set to connect to redis // Per request ID that is used by Trx tracing package. and by Socket.IO for directing trafic
	trx.TrxIdSeen(id, req.RequestURI, req.Method) // trx.RequestId = id
	rw.RequestTrxId = id

	rw.G_Trx = trx

	trx.SetTraceDebug(true) // xyzzy - Config Option

	fmt.Fprintf(os.Stderr, "\n%sStart Request: %s %s\n", MiscLib.ColorMagenta, req.RequestURI, MiscLib.ColorReset)

	trx.TraceUri(req, url.Values{}) // xyzzy - convert to pass of Parsed Params in Table form -- xyzzyTrx below

	trx.AddNote(1, fmt.Sprintf("Start Request, IP=%s URI=%s, request-id=%s", req.RemoteAddr, req.RequestURI, id))

	if !trx_id_found {
		expire := time.Now().AddDate(10, 0, 2) // good for 10 year 2 days
		cookie := http.Cookie{Name: "X-Go-FTL-Trx-Id", Value: id, Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400 * 366 * 10, Secure: false, HttpOnly: false}
		http.SetCookie(www, &cookie)
	}

	a := ""
	if req.URL.RawQuery != "" {
		a = "?"
	}
	rw.OriginalURL = req.URL.Path + a + req.URL.RawQuery

	fmt.Printf("%s\n%s\nOriginalURL: %s, Method:%s, TrxCookie=%s, %s\n%s%s\n", MiscLib.ColorYellow,
		"--------------------------------------------------------------------------------------------------------------------------------",
		rw.OriginalURL, req.Method, id, godebug.LF(),
		"--------------------------------------------------------------------------------------------------------------------------------",
		MiscLib.ColorReset)

	for {
		//0. clean_up of URLs ->
		//	1. if /foo/bob - is a "dir" then add "/" to end to get /foo/bob/
		//	2. if /abc//def - then /abc/def
		req.URL.Path = urlpath.Clean(req.URL.Path) // -- Any trailing slash is removed, doulbe // are remove

		// fmt.Printf("Params Are: %s AT %s\n", rw.Ps.DumpParam(), godebug.LF())
		goftlmux.ParseCookiesAsParamsReg(www, req, &rw.Ps) // 28ns
		// XyzzyParams - Make following sections optional via some sort of middleware system.
		goftlmux.ParseQueryParamsReg(www, req, &rw.Ps) //
		goftlmux.MethodParamReg(www, req, &rw.Ps)      // 15ns
		// fmt.Printf("Immeditly before ParseBodyAsParamsReg\n")
		goftlmux.ParseBodyAsParamsReg(www, req, &rw.Ps) // 27ns
		// 0. dump of params make it a decent printout - in a little table. -- With text values for Type and From
		// fmt.Printf("Params Are After Cookies: %s AT %s\n", rw.Ps.DumpParamDB(), godebug.LF())
		// xyzzyTrx - set params
		fmt.Printf("\nParams + Cookies for (%s): %s AT %s\n", req.URL.Path, rw.Ps.DumpParamTable(), godebug.LF())

		// Delete reserved items: is_logged_in := ps.ByNameDflt("$is_logged_in$", "")
		goftlmux.RenameReservedItems(www, req, &rw.Ps, cfg.ReservedItems)

		// func (ps *Params) DumpParamNVF() (rv []common.NameValueFrom) {
		dump := rw.Ps.DumpParamNVF()
		trx.SetDataPs(dump)

		rw.RerunRequest = false
		hdlr.Next.ServeHTTP(rw, req)
		if rw.StatusCode == 0 {
			rw.StatusCode = http.StatusOK
		}
		if rw.StatusCode != http.StatusOK { // 200
			// xyzzyTrx -- If rw.RerunRequest is true note that - and exit due to a non-200 code
			break
		}
		if rw.RerunRequest == false { // If no re-run request
			// xyzzyTrx
			break
		}
	}

	fmt.Printf("Mid: StatusCode = %d, req.URL.Path=%s, IsHijacked=%v\n", rw.StatusCode, req.URL.Path, rw.IsHijacked)
	fmt.Printf("\nFinal! Params + Cookies for (%s): %s AT %s\n", req.URL.Path, rw.Ps.DumpParamTable(), godebug.LF())
	uu := rw.Ps.DumpParamUsed("api_table_key")
	fmt.Printf("\nParams + Cookies for (Shows Used) (%s): %sAT %s\n", req.URL.Path, uu, godebug.LF())
	// trx.AddNote(1, fmt.Sprintf("End of Request, UseParameters=%s", uu, id))
	dump := rw.Ps.DumpParamNVF()
	trx.UpdateDataPs(dump)

	rw.Ps.ReportUnexpectedUnused(NormalUnused)
	fmt.Printf("%sAfter At: %s\n%s%s", MiscLib.ColorCyan, godebug.LF(), rw.Ps.DumpParamTable(), MiscLib.ColorReset)

	dumpBody := false
	if !rw.IsHijacked {
		if rw.Error != nil && rw.StatusCode == 200 {
			rw.StatusCode = http.StatusInternalServerError
		}
		if rw.StatusCode != 200 && rw.StatusCode != 304 {
			if false {
				logrus.Errorf(tmplp.TemplateProcess(hdlr.Format, rw, req, make(map[string]string)))
			}
			logrus.Errorf("This one -- error code -- %d, %s", rw.StatusCode, godebug.LF())
		}
		if rw.StatusCode == 200 {
			ct := rw.Header().Get("Content-Type")
			// fmt.Printf("ct=%s At: %s\n", ct, godebug.LF())
			if ct == "application/json" { // xyzzy starts with?
				// fmt.Printf("At: %s\n", godebug.LF())
				dumpBody = true
			}
		}
	}

	oldbody := rw.GetBody()
	trx.RvBody = string(oldbody)

	// add in headers for Trx -- espeically content type - xyzzy201412
	trx.SetHeader(rw.Header())

	rw.FinalFlush()

	if dumpBody {
		// fmt.Printf("At: %s\n", godebug.LF())
		body := rw.GetBody()
		fmt.Printf("\nBody: %s\n\n", body)
	}
	if !rw.IsHijacked {
		// fmt.Printf("At: %s\n", godebug.LF())
		fmt.Printf("*** Final Flush ***, %s\n\n=== End %s\n\n", godebug.LF(), strings.Repeat("=== ", 30))
	}

	// ---------------------------------------------------------------------------------------------------------------------------
	ip := req.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	finishTime := time.Now()
	// elapsedTime := finishTime.Since(rw.StartTime)
	elapsedTime := finishTime.Sub(rw.StartTime)

	trx.UriSaveData(ip, rw.StartTime, req.Method, req.RequestURI, req.Proto, rw.StatusCode, rw.Length, elapsedTime, req)

	if !rw.IsHijacked {
		trx.TraceUriRawEnd(req, elapsedTime)
	} else {
		trx.TraceUriRawEndHijacked(req, elapsedTime)
	}
}

func GetTrx(www http.ResponseWriter) (ptr *tr.Trx) {
	if rw, ok := www.(*goftlmux.MidBuffer); ok {
		if rw.G_Trx == nil {
			panic(fmt.Sprintf("Should have has a *tr.Trx - it was NIL, %s\n", godebug.LF(2)))
		}
		if ptr, ok = rw.G_Trx.(*tr.Trx); ok {
			return ptr
		}
	}
	panic(fmt.Sprintf("Should have has a *goftlmux.MidBuffger - got passed a %T, %s\n", www, godebug.LF(2)))
}

func GetTrx1(rw *goftlmux.MidBuffer) (ptr *tr.Trx) {
	if rw.G_Trx == nil {
		panic(fmt.Sprintf("Should have has a *tr.Trx - it was NIL, %s\n", godebug.LF(2)))
	}
	if ptr, ok := rw.G_Trx.(*tr.Trx); ok {
		return ptr
	}
	panic(fmt.Sprintf("Invalid Type - Should have has a *tr.Trx, %s\n", godebug.LF(2)))
}

const db_g1 = false
const dbTop = false

/* vim: set noai ts=4 sw=4: */
