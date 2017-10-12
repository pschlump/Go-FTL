package main

import "flag"

type CfgType struct {
	RedisHost  *string `json:"RedisHost"`
	RedisPort  *string `json:"RedisPort"`
	RedisAuth  *string `json:"RedisAuth"`
	DebugFlags *string `json:"DebugFlags"`
	ServerName string  `json:"ServerName"`
	// ReplyTo    *string `json:"ReplyTo"`
}

var RedisHost = flag.String("host", "127.0.0.1", "PotgresSQL connection info") // 0
var RedisPort = flag.String("port", "6379", "PotgresSQL connection info")      // 1
var RedisAuth = flag.String("auth", "", "PotgresSQL connection info")          // 2
var Cfg = flag.String("cfg", "cfg.json", "PotgresSQL connection info")         // 3
var Debug = flag.String("debug", "", "debug flags")                            // 4
var Help = flag.Bool("help", false, "get help")                                // 5
var addr = flag.String("addr", ":9876", "http service address")                // 6
var dir = flag.String("dir", "./static", "static file server ")                // 7
func init() {
	flag.StringVar(RedisHost, "H", "127.0.0.1", "PotgresSQL connection info") // 0
	flag.StringVar(RedisPort, "P", "6379", "PotgresSQL connection info")      // 1
	flag.StringVar(RedisAuth, "A", "", "PotgresSQL connection info")          // 2
	flag.StringVar(Cfg, "c", "cfg.json", "PotgresSQL connection info")        // 3
	flag.StringVar(Debug, "D", "", "debug flags")                             // 4
	flag.StringVar(addr, "a", ":9876", "http service address")                // 6
	flag.StringVar(dir, "d", "./static", "static file server ")               // 7
}

var dbFlag map[string]bool
var g_cfg CfgType
