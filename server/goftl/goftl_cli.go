//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1121
//

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/nameresolve"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/endless"
	"github.com/pschlump/godebug"
	"github.com/pschlump/mon-alive/lib"
	"github.com/pschlump/radix.v2/redis"
)

type HandlersStruct struct {
	Id   int
	Name string
	Hdlr http.Handler // Required field for all chaining of middleware.
}

// https://gobyexample.com/command-line-flags
var GlobalCfgFN = flag.String("globalCfgFile", "global-config.json", "Full path to global config file")
var CfgFN = flag.String("cfgFile", "ftl-config.json", "Full path to config file")
var Version = flag.Bool("version", false, "Report version of code and exit")

// Note is not used for anything, but it can make it easy to find the server with ps -ef | grep "note-text"
var Note = flag.String("note", "", "arbitrary note text, for use with ps -ef")

func init() {
	// example with short version for long flag
	flag.StringVar(GlobalCfgFN, "g", "global-config.json", "Full path to global config file")
	flag.StringVar(CfgFN, "c", "ftl-config.json", "Full path to config file")
}

var GitCommit string

//------------------------------------------------------------------------------------------------------------------------------
func RedisClient() (client *redis.Client, conFlag bool) {
	var err error
	fmt.Printf("AT: connect to redis with: %s %s\n", godebug.LF(), cfg.ServerGlobal.RedisConnectHost+":"+cfg.ServerGlobal.RedisConnectPort)
	client, err = redis.Dial("tcp", cfg.ServerGlobal.RedisConnectHost+":"+cfg.ServerGlobal.RedisConnectPort)
	if err != nil {
		// log.Fatal(err)
		fmt.Printf("Error on connect to redis:%s, fatal\n", err)
		fmt.Fprintf(os.Stderr, "%s\n\n\n-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------\nError on connect to redis:%s, fatal\n", MiscLib.ColorRed, err)
		fmt.Fprintf(os.Stderr, "Config Data: %s\n", godebug.SVarI(cfg.ServerGlobal))
		fmt.Fprintf(os.Stderr, "\n-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------\n\n\n%s", MiscLib.ColorReset)
		os.Exit(1)
	}
	if cfg.ServerGlobal.RedisConnectAuth != "" {
		err = client.Cmd("AUTH", cfg.ServerGlobal.RedisConnectAuth).Err
		if err != nil {
			fmt.Printf("Error on connect to Redis --- Invalid authentication:%s, fatal\n", err)
			fmt.Fprintf(os.Stderr, "%s\nError on connect to Redis --- Invalid authentication:%s, fatal%s\n\n", MiscLib.ColorRed, err, MiscLib.ColorReset)
			os.Exit(1)
		} else {
			conFlag = true
		}
	} else {
		conFlag = true
	}
	return
}

// ------------------------------------------------------------------------------------------------------------------
func main() {

	flag.Parse()

	if *Version {
		fmt.Printf("Version (Git Commit): %s\n", GitCommit)
		os.Exit(0)
	}

	globalCfgFN := cfg.ResolvLocalFile(*GlobalCfgFN)
	cfg.ReadGlobalConfigFile(globalCfgFN)

	haveConfig := false
	cfgFN := cfg.ResolvLocalFile(*CfgFN)
	if lib.Exists(cfgFN) {
		mid.ReadConfigFile2(cfgFN)
		haveConfig = true
	}

	fns := flag.Args()
	for _, s := range fns {
		ss := cfg.ResolvLocalFile(s)
		mid.ReadConfigFile2(ss)
		haveConfig = true
	}

	if !haveConfig {
		fmt.Fprintf(os.Stderr, "Error: no confuration file supplied\n")
		os.Exit(1)
	}

	monClient, err7 := RedisClient()
	fmt.Printf("err7=%v AT: %s\n", err7, godebug.LF())
	mon := MonAliveLib.NewMonIt(func() *redis.Client { return monClient }, func(conn *redis.Client) {})
	mon.SendPeriodicIAmAlive("Go-FTL")

	fmt.Printf("Successfully Initialized...\n")
	fmt.Printf("Config: %s\n", lib.SVarI(cfg.ServerGlobal))

	wg := &sync.WaitGroup{}

	Id := 1
	var HdlrSet []HandlersStruct
	bot := nameresolve.NewNameResolve()

	once := make(map[string]bool)

	for serverName, s_item := range cfg.ServerGlobal.Config { // Config map[string]PerServerConfigType // Anything that did not match the abobve JSON names //
		// PerServerConfigType.ConfigData -> s_item.ConfigData

		p1 := mid.NewBotHandler() // "bot" handler - for 404 routes // var p1 mid.GoFTLMiddleWare
		configArray, ok := s_item.Plugins.([]interface{})
		if !ok {
			fmt.Fprintf(os.Stderr, "%sError Plugins is not array - nothing will be confiured for %s, %s%s\n", MiscLib.ColorRed, serverName, godebug.LF(), MiscLib.ColorReset)
		} else {

			n_err := 0
			err_pos := -1
			missingPluginError := false
			for ii := len(configArray) - 1; ii >= 0; ii-- {
				vv, ok := configArray[ii].(map[string]interface{})
				fmt.Printf("configArray %s, %s\n", godebug.SVarI(vv), godebug.LF())
				if !ok {
					fmt.Printf("At: %s -- error Plugins is not array of map[string]interface ********************** \n", lib.LF())
				} else {
					pluginName := lib.FirstName(vv) // extract pluginName of plugin
					// should check if ok to do this
					if tt, ok := vv[pluginName]; !ok {
						fmt.Fprintf(os.Stderr, "ERROR: Missing plugin name [%s] at %s\n", pluginName, godebug.LF())
						// panic("fatal - missing plugin")
					} else if data, ok := tt.(map[string]interface{}); !ok {
						fmt.Fprintf(os.Stderr, "ERROR: Missing plugin name [%s] at %s / unable to type cast\n", pluginName, godebug.LF())
						// panic("fatal - missing plugin/2")
					} else {
						// data := vv[pluginName].(map[string]interface{})
						jj := mid.LookupInitByName3(pluginName)
						if jj >= 0 {
							// dataOfType := (cfg.NewInit[jj].CreateEmpty)()
							dataOfType := (mid.NewInit3[jj].CreateEmpty)(pluginName) // var dataOfType mid.GoFTLMiddleWare
							valiationStringJson := mid.NewInit3[jj].ValidJSON
							// Validate data into fd - vv - data source
							ok, dflt, msg := cfg.IsInputValid(pluginName, valiationStringJson, data)
							if !ok {
								err_pos = jj
								n_err++
								fmt.Fprintf(os.Stderr, "%sError (00010): Unable to initialize module '%s' in server '%s', Error:%s%s\n", MiscLib.ColorRed, pluginName, serverName, msg, MiscLib.ColorReset)
								fmt.Printf("Error (00010): Unable to initialize module '%s' in server '%s', Error:%s\n", pluginName, serverName, msg)
							} else {
								fmt.Printf("At: %s ----------- it is valid at this point ----------- \n", lib.LF())
								err := cfg.MapJsonToStruct(data, dflt, dataOfType) // xyzzy7
								if err != nil {
									fmt.Fprintf(os.Stderr, "%sError (00011): Unable to initialize module %s in server %s, %s%s\n", MiscLib.ColorRed, pluginName, serverName, err, MiscLib.ColorReset)
									fmt.Fprintf(os.Stdout, "Error (00011): Unable to initialize module %s in server %s, %s\n", pluginName, serverName, err)
								} else {
									fmt.Printf("%sAt: %s ---------------- struct set up at this point -----------%s\n", MiscLib.ColorCyan, lib.LF(), MiscLib.ColorReset)

									// -----------------------------------------------------------------------------------------------------------------------
									//finit := cfg.NewInit[jj].OneTimeInit	//xyzzy7
									//if finit != nil {
									//	// cfg.PerServerConfigType
									//	finit(dataOfType, s_item.ConfigData, cfg.NewInit[jj].CallNo) // xyzzy xyzzy
									//	cfg.NewInit[jj].CallNo++
									//}
									// p1, err = (cfg.NewInit[jj].FinalizeHandler)(p1, cfg.ServerGlobal, dataOfType, pluginName, ii)
									// dataOfType.OntTimeInit(dataOfType, s_item.ConfigData, mid.NewInit3[jj].CallNo)
									// -----------------------------------------------------------------------------------------------------------------------

									err = dataOfType.PreValidate(&cfg.ServerGlobal, s_item.ConfigData, pluginName, ii, mid.NewInit3[jj].CallNo)
									if err != nil {
										fmt.Fprintf(os.Stderr, "%sError: %s%s\n", MiscLib.ColorRed, err, MiscLib.ColorReset)
									}

									err = dataOfType.InitializeWithConfigData(p1, &cfg.ServerGlobal, pluginName, ii, mid.NewInit3[jj].CallNo)
									if err != nil {
										fmt.Fprintf(os.Stderr, "%sError: %s%s\n", MiscLib.ColorRed, err, MiscLib.ColorReset)
									}

									mid.NewInit3[jj].CallNo++
									p1 = dataOfType // Walk Forward
								}
							}
						} else {
							fmt.Fprintf(os.Stderr, "%sError: Unable to find middleware/plugin pluginName [%s]%s\n", MiscLib.ColorRed, pluginName, MiscLib.ColorReset)
							fmt.Fprintf(os.Stdout, "Error: Unable to find middleware/plugin pluginName [%s]\n", pluginName)
							missingPluginError = true
						}
					} // check for data and existing plugin.
				}
			}
			if missingPluginError {
				pluginList := mid.LookupPluginList()
				fmt.Fprintf(os.Stderr, "%sAvailable Plugins Are:\n%s%s\n", MiscLib.ColorRed, pluginList, MiscLib.ColorReset)
				fmt.Fprintf(os.Stdout, "Available Plugins Are:\n%s\n", pluginList)
			}
			if n_err > 0 {
				fmt.Fprintf(os.Stderr, "%sModules did not initialize properly - fatal error, serverName=%s, last error pos=%d\nSyntax error in JSON validation specification\nFatal error reported from: %s%s\n", MiscLib.ColorRed, serverName, err_pos, godebug.LF(), MiscLib.ColorReset)
				os.Exit(4)
			}
		}

		// Top Handler convers from standard HTTP request/Responce
		// Writer to the extende versions for this application.
		// Top hancler also needs to do "serverName" matching on all reuests
		p1 = mid.NewTopHandler(p1, &cfg.ServerGlobal, nil, "*top*", -1)

		fmt.Printf(">>>>>>>>>>>>>>>>>>>>>>>>>>> append to HdlrSet\n")
		// Save each of the handlers
		HdlrSet = append(HdlrSet, HandlersStruct{Id: Id, Name: serverName, Hdlr: p1})
		Id++

		fmt.Printf("\n\n ------------------------------------------------------------------------------------------------------------------- \n\n")
		fmt.Fprintf(os.Stderr, "\n\n ------------------------------------------------------------------------------------------------------------------- \n\n")

		// ----------------------------------------------------------------------------------------------------------
		// Set up the maping from IP addressses of local system to names that are to be name resolved.
		for _, listen := range s_item.ListenTo {
			fmt.Printf("bot.AddName(%s)\n", listen)
			fmt.Fprintf(os.Stderr, "bot.AddName(%s)\n", listen)
			bot.AddName(listen, p1, Id, "")
		}

		// cfg.ServerGlobal.Config[name] = s_item

		if db8 {
			fmt.Printf("At Bottom of loop, %s, %s\n", godebug.LF(), lib.SVarI(s_item))
		}
	}

	//

	//

	//

	//

	for name, s_item := range cfg.ServerGlobal.Config { // Config        map[string==UserName]PluginConfigType

		s_item.Port = make([]string, len(s_item.ListenTo), len(s_item.ListenTo))

		for jj, listen := range s_item.ListenTo {

			s_item.Port[jj] = "80"

			u, err := url.Parse(listen)
			_ = err
			if u.Scheme == "https" {
				s_item.Port[jj] = "443"
			} else {
				continue
			}

			fmt.Printf("bot.getTopHandler(%s)\n", listen)
			fmt.Fprintf(os.Stderr, "bot.getTopHandler(%s)\n", listen)

			p1, err := bot.GetRawTopHandler(listen)
			if err != nil {
				fmt.Printf("No server to listen to %s\n", listen)
				continue
			}

			hh, po, err := net.SplitHostPort(u.Host)
			if false {
				fmt.Printf("u Parsed=%s,%s,%s\n", hh, po, err)
			}
			if err == nil {
				if po != "" {
					s_item.Port[jj] = po
				}
			}

			if err != nil {
				fmt.Printf("Invalid URL: %s, no server configured to listen to this.\n", listen)
			} else {
				if strings.HasPrefix(u.Host, "*.") {
					u.Host = strings.TrimPrefix(u.Host, "*.") // this means that it is up to the internal p1 to do routing for it now.
				}
				fmt.Printf("Scheme=%s Host=%s\n", u.Scheme, u.Host) // host has domain:port in it			u.Scheme is https|http
				// uu.Host, uu.Port = net.SplitHostPort(u.Host)
				// log.Fatal(http.ListenAndServe(u.Host, http.FileServer(http.Dir(s_item.StaticDirs[0]))))

				// If localhost:PORT, then replace with listening to all local ports, specified by :PORT
				u.Scheme = strings.ToLower(u.Scheme)
				if strings.HasPrefix(u.Host, "localhost:") {
					u.Host = strings.TrimPrefix(u.Host, "localhost")
				}

				fmt.Printf("HTTPS testing now - scheme= >%s<\n", u.Scheme)
				fmt.Fprintf(os.Stderr, "%sHTTPS testing now - scheme= >%s< host >%s<, %s%s\n", MiscLib.ColorRed, u.Scheme, u.Host, godebug.LF(), MiscLib.ColorReset)
				fmt.Fprintf(os.Stdout, "%sHTTPS testing now - scheme= >%s< host >%s<, %s%s\n", MiscLib.ColorRed, u.Scheme, u.Host, godebug.LF(), MiscLib.ColorReset)
				fmt.Printf("All Config: %s\n", godebug.SVarI(s_item))
				if !strings.Contains(u.Host, ":") {
					u.Host += ":443"
				} else if len(u.Host) > 1 && u.Host[len(u.Host)-1] == ':' {
					u.Host += "443"
				}
				wg.Add(1)
				go func(s_item cfg.PerServerConfigType, name string) {
					fmt.Printf("Listen: %s, %s\n", name, godebug.LF())
					if _, ok := once[name]; ok {
						return
					}
					once[name] = true
					var err error
					defer wg.Done()
					//srv := http.Server{
					//	Addr:    u.Host, // need port 443? added
					//	Handler: p1,
					//}
					// h2.ConfigureServer(&srv, nil)
					tlsThisConfigServer := []endless.TLSConfig{}
					fmt.Fprintf(os.Stderr, "s_item.Certs >%s<, %s\n", godebug.SVar(s_item.Certs), godebug.LF())
					fmt.Fprintf(os.Stdout, "s_item.Certs >%s<, %s\n", godebug.SVar(s_item.Certs), godebug.LF())
					for i := 0; i < len(s_item.Certs); i += 2 {
						tlsThisConfigServer = append(tlsThisConfigServer, endless.TLSConfig{
							Certificate:        s_item.Certs[i],
							Key:                s_item.Certs[i+1],
							ProtocolMinVersion: tls.VersionTLS11, // may need to be 1.1
							Ciphers: []uint16{
								tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
								tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
								tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
								tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
								tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
								tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
								tls.TLS_RSA_WITH_AES_128_CBC_SHA,
								tls.TLS_RSA_WITH_AES_256_CBC_SHA,
								tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
								tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
								tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
								tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
							},
							PreferServerCipherSuites: false,
						})
					}
					fmt.Printf("%sSNI server start%s\n", MiscLib.ColorCyan, MiscLib.ColorReset)
					fmt.Fprintf(os.Stderr, "%sSNI server start%s\n", MiscLib.ColorCyan, MiscLib.ColorReset)
					// err = ListenAndServeTLSWithSNI(&srv, tlsThisConfigServer)
					err = endless.ListenAndServeTLSWithSNI(u.Host, p1, tlsThisConfigServer)
					log.Fatal(err)
				}(s_item, listen)
			}
		}

		cfg.ServerGlobal.Config[name] = s_item

		if db8 {
			fmt.Printf("At Bottom of loop, %s, %s\n", godebug.LF(), lib.SVarI(s_item))
		}

		// xyzzyAAA }

	}

	// bot.AddDefault("http:", "*", nil, 1000000)
	if db8 || true {
		fmt.Fprintf(os.Stdout, "Lookup table: %s, %s\n", lib.SVarI(bot), godebug.LF())
	}

	for host_port, _ := range bot.IpLookup {
		fmt.Fprintf(os.Stdout, "	host_port: %s\n", host_port)
	}

	for host_port, vvv := range bot.IpLookup {

		hasHttp := false
		for _, ip := range vvv {
			if ip.Proto == "http:" {
				hasHttp = true
				break
			}
		}

		if hasHttp {
			fmt.Printf("Will start host_port:%s, has http listener\n", host_port)

			wg.Add(1)
			// Run in parallel
			go func(host_port, names string) {
				defer wg.Done()
				fmt.Fprintf(os.Stderr, "Start Listener On --- host_port: %s Names:%s\n", host_port, names)
				//server := &http.Server{
				//	Addr:    host_port,
				//	Handler: bot,
				//}
				//err := server.ListenAndServe()
				err := endless.ListenAndServe(host_port, bot)
				fmt.Fprintf(os.Stderr, "%sFailed to start: %s%s\n", MiscLib.ColorRed, host_port, MiscLib.ColorReset)
				log.Fatal(err)
			}(host_port, getNamesForServer(vvv))
		}
	}

	wg.Wait()

}

func getNamesForServer(mmm map[string]*nameresolve.IpToHostPort) (rv string) {
	com := ""
	for aKey := range mmm {
		rv = rv + com + aKey
		com = ", "
	}
	return
}

const db8 = false
const most_recent2 = false // Aug 2016, 3

/* vim: set noai ts=4 sw=4: */
