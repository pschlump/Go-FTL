// Copyright (C) Philip Schlump, 2016-2017.

package main

// CMDs.
// {"cmd":"lm-update","data":...}
// {"cmd":"tracer"}
// {"cmd":"chat-msg","data":...}
// {"cmd":"trace-filter","data":...}
// {"cmd":"auth","un":...,"pw":...}				// Auth via server password check -> {"cmd":"authorized","jwt":...,"id":UUID}
// {"cmd":"auth","un":...,"v":...,"s":...}		// Auth via AesSrp -> {"cmd":"authorized"}
// {"cmd":"bye","id":...}						// loggout

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"www.2c-why.com/qr-today.com/MicroService/cfgLib"

	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/MicroServiceLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/mon-alive/lib"
	"github.com/pschlump/radix.v2/redis" // Modified pool to have NewAuth for authorized connections

	"github.com/pschlump/mon-alive/ListenLib"
)

func main() {

	dbFlag = make(map[string]bool)
	dbFlag["show-cfg"] = false

	flag.Parse()

	if *Help {
		fmt.Printf(`
./trace-svr [--help] [--debug=debug,strings|-D db,strgs] [--cfg=fn|-c fn] [--host=name|-h name] [--port=port|-p port] [--auth=password|-a password] &
	debug-strings are (comma separated, no spaces list)
		show-cfg			Dump out configuration file to verify it.
		db-startup			Show message that the tracer-bot is running at startup.
		echo-msg			Show the message data being "chat"ed
	cfg 		 configuration file, defaults to ./cfg.json
	host		 host to connect to for Redis, default 127.0.0.1 (localhost)
	port		 port to connect to for Redis, default 6379
	auth		 host to connect to for Redis, optional or Redis password.
`)
		os.Exit(0)
	}

	if *Cfg != "" {
		cfgLib.ReadConfigFile(*Cfg, &g_cfg)
		if dbFlag["show-cfg"] {
			fmt.Printf("read in config file, data=%s\n", lib.SVarI(g_cfg))
		}
		if g_cfg.RedisHost != nil {
			RedisHost = g_cfg.RedisHost
		}
		if g_cfg.RedisPort != nil {
			RedisPort = g_cfg.RedisPort
		}
		if g_cfg.RedisAuth != nil {
			RedisAuth = g_cfg.RedisAuth
		}
		for _, jj := range strings.Split(*Debug, ",") {
			dbFlag[jj] = true
		}
		if g_cfg.DebugFlags != nil {
			for _, jj := range strings.Split(*g_cfg.DebugFlags, ",") {
				dbFlag[jj] = true
			}
		}
	} else {
		fmt.Fprintf(os.Stderr, "Invalid configuration - must have a --cfg=./cfg.json file\n")
		os.Exit(3)
	}

	if dbFlag["show-cfg"] {
		fmt.Printf("Debug flags are %s\n", lib.SVarI(dbFlag))
	}

	// ---------------------------------------------------------------------------------------------------------------------------------------

	type commonConfig struct {
		Name string                     //
		conn *redis.Client              //
		mon  *MonAliveLib.MonIt         //
		ms   *MicroServiceLib.MsCfgType //
	}

	cc := commonConfig{
		Name: "ws:server:mon-alive",
	}

	// ---------------------------------------------------------------------------------------------------------------------------------------
	monClient, ok := cfgLib.RedisClient(*RedisHost, *RedisPort, *RedisAuth)
	if !ok {
		fmt.Fprintf(os.Stderr, "%s: failed to connect to redis\n", g_cfg.ServerName)
		os.Exit(1)
	}
	mon := MonAliveLib.NewMonIt(func() *redis.Client { return monClient }, func(conn *redis.Client) {})
	mon.SendPeriodicIAmAlive(g_cfg.ServerName)

	if dbFlag["db-startup"] {
		fmt.Printf("Servcie: [%s] Started\n", g_cfg.ServerName)
	}

	// ---------------------------------------------------------------------------------------------------------------------------------------
	// Run the single HUB for the chat
	hub := newHub()
	go hub.run()

	// ---------------------------------------------------------------------------------------------------------------------------------------
	// Listen for /ws and run the websocket server on it.
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		client := serveWs(hub, w, r)

		create_LiveMonitor := func(Verbose bool) func() error {

			connTmp, conFlag := cfgLib.RedisClient(*RedisHost, *RedisPort, *RedisAuth)
			if !conFlag {
				fmt.Printf("Did not connect to redis\n")
				os.Exit(1)
			}
			cc.conn = connTmp

			monTmp := MonAliveLib.NewMonIt(func() *redis.Client { return cc.conn }, func(conn *redis.Client) {})
			cc.mon = monTmp

			return func() error {

				ms := ListenLib.NewMsCfgType("trx:listen", "")

				ms.RedisConnectHost = *RedisHost
				ms.RedisConnectPort = *RedisPort
				ms.RedisConnectAuth = *RedisAuth

				ms.SetEventPattern("__keyevent@0__:expire*")

				ms.ConnectToRedis() // Create the redis connection pool, alternative is ms.SetRedisPool(pool) // ms . SetRedisPool(pool *pool.Pool)
				ms.SetRedisConnectInfo(*RedisHost, *RedisPort, *RedisAuth)
				ms.SetupListen()

				showStatus := func(dm map[string]interface{}) {
					// fmt.Printf("dm=%+v\n", dm)

					runIt := false

					cmd_r, ok0 := dm["cmd"]
					cmd, ok1 := cmd_r.(string)
					itemKey_r, ok2 := dm["val"]

					if ok0 && ok1 && ok2 && cmd == "expired" {

						itemKey, ok3 := itemKey_r.(string)

						if ok3 {
							runIt = cc.mon.IsMonitoredItem(itemKey)
						}

					} // check for this having a key name passed in.

					if ok0 && ok1 && cmd != "expired" { // cmd==timeout-call || cmd==at-top
						runIt = true
						// fmt.Printf("dm=%+v\n", dm)
					}

					if runIt {
						st, hasChanged := cc.mon.GetStatusOfItemVerbose(Verbose)
						if hasChanged {
							if db9 {
								fmt.Printf("For push to WebSocket: st=%s\n", godebug.SVarI(st))
							}

							sss := lib.SVarI(st)
							// {"cmd":"lm-update","data":...}
							client.clientBroadcast(`{"cmd":"lm-update","data":` + sss + "}")

						}
					}

				}

				var wg sync.WaitGroup

				ms.ListenForServer(showStatus, &wg)

				wg.Wait() // wait forever - server runs in loop. -- On "exit" message it will

				return nil
			}
		}

		go create_LiveMonitor(false)()

	})

	// If is is not /ws, then assume that it is a file and serve files.
	http.Handle("/", http.FileServer(http.Dir(*dir)))

	fmt.Fprintf(os.Stderr, "Listening on: %s\n", *addr)

	// Crank it up.  Star listening.
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

const db9 = false
