//
// user-redis - maintain uses in redis for basicredis middleware
//
// Copyright (C) Philip Schlump, 2013-2015.
// Version: 0.5.9
// BuildNo: 1811
//

//
// How to use
//
// To add a user:
// 		$ user-redis -a username -p password -r realm
//
// To Delete a user:
// 		$ user-redis -d username
//
// To modify a users password:
// 		$ user-redis -m username -p password -r realm
//
//
/*

	"HashUsername":  	 { "type":[ "bool" ], "required":false, "default":"false" },
	"HashUsernameSalt":  { "type":[ "string" ], "required":false, "default":"8H3QhT9uHElh+c5NfowHx1gLeDw6qBMSTLvoL87GcB4FwflM8v2cTs" },
*/

package main

import (
	"crypto/rand"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pschlump/radix.v2/redis"
	"github.com/pschlump/uuid"
	"golang.org/x/crypto/pbkdf2" // https://github.com/golang/crypto/blob/master/pbkdf2/pbkdf2.go
)

var RedisHost = flag.String("host", "127.0.0.1", "Redis host to connect to")                                                           // 0
var RedisPort = flag.String("port", "6379", "Redis port to connect to")                                                                // 1
var RedisAuth = flag.String("redisauth", "", "Redis auth token (leave empty if no auth)")                                              // 2
var RedisPrefix = flag.String("redisprefix", "BasicAuth", "Redis key prefix")                                                          // 3
var OptionAdd = flag.String("add", "", "To add user")                                                                                  // 4
var OptionDel = flag.String("delete", "", "To delete user")                                                                            // 5
var OptionMod = flag.String("modify", "", "To modify user")                                                                            // 6
var Realm = flag.String("realm", "", "Realm name")                                                                                     // 7
var Password = flag.String("password", "", "password")                                                                                 // 8
var HashUsername = flag.Bool("hashuser", false, "Hash the Username")                                                                   // 9
var HashUsernameSalt = flag.String("hashusersalt", "8H3QhT9uHElh+c5NfowHx1gLeDw6qBMSTLvoL87GcB4FwflM8v2cTs", "Salt for Username Hash") // 10
func init() {
	flag.StringVar(RedisHost, "H", "127.0.0.1", "Redis host to connect to")                                                   // 0
	flag.StringVar(RedisPort, "P", "6379", "Redis port to connect to")                                                        // 1
	flag.StringVar(RedisAuth, "A", "", "Redis auth token (leave empty if no auth)")                                           // 2
	flag.StringVar(RedisPrefix, "T", "BasicAuth", "Redis key prefix")                                                         // 3
	flag.StringVar(OptionAdd, "a", "", "To add user")                                                                         // 4
	flag.StringVar(OptionDel, "d", "", "To delete user")                                                                      // 5
	flag.StringVar(OptionMod, "m", "", "To modify user")                                                                      // 6
	flag.StringVar(Realm, "r", "", "Realm name")                                                                              // 7
	flag.StringVar(Password, "p", "", "password")                                                                             // 8
	flag.BoolVar(HashUsername, "h", false, "Hash the Username")                                                               // 9
	flag.StringVar(HashUsernameSalt, "s", "8H3QhT9uHElh+c5NfowHx1gLeDw6qBMSTLvoL87GcB4FwflM8v2cTs", "Salt for Username Hash") // 10
}

// ===============================================================================================================================================
var redis_client *redis.Client

func ConnectToRedis(redis_host, redis_port, redis_auth string) {
	var err error

	redis_client, err = redis.Dial("tcp", redis_host+":"+redis_port)
	if err != nil {
		fmt.Printf("user-redis: Failed to connect to redis-server.\n")
		os.Exit(1) // handle error
	}

	if redis_auth != "" { // New Redis AUTH section
		t, err := redis_client.Cmd("AUTH", redis_auth).Str()
		if err != nil {
			fmt.Printf("user-redis: Failed to authorize to use redis. %s\n", err)
			os.Exit(1)
		} else {
			fmt.Printf("user-redis: Connected and Authorized to redis-server. %s\n", t)
		}
	} else {
		fmt.Printf("user-redis: Connected to redis-server.\n")
	}
}

const NIterations = 5000

// ===============================================================================================================================================
func main() {

	flag.Parse()
	// fns := flag.Args()

	ConnectToRedis(*RedisHost, *RedisPort, *RedisAuth)

	//fmt.Printf("******************* successful connection ****************************\n")
	//os.Exit(0)

	genKey := func(un string) (key string) {
		if *HashUsername {
			em := fmt.Sprintf("%x", pbkdf2.Key([]byte(*Realm+":"+un), []byte(*HashUsernameSalt), NIterations, 64, sha256.New))
			key = *RedisPrefix + ":" + em
		} else {
			key = *RedisPrefix + ":" + *Realm + ":" + un
		}
		return
	}

	if *OptionAdd != "" {
		if *Realm == "" || *Password == "" {
			Usage()
		}

		un := *OptionAdd
		salt := genSalt()
		key := genKey(un)
		dk := fmt.Sprintf("%x", pbkdf2.Key([]byte(*Password), []byte(salt), NIterations, 64, sha256.New))
		id0, _ := uuid.NewV4()
		user_id := id0.String()
		value := salt + ":" + dk + ":" + user_id

		err := redis_client.Cmd("GET", key).Err
		if err != nil {
			fmt.Printf("user-redis: Error: Attempt to add when %s already exists in file\n", *OptionAdd)
			os.Exit(2)
		}

		redis_client.Cmd("SET", key, value)
	} else if *OptionDel != "" {
		if *Realm == "" {
			Usage()
		}

		un := *OptionDel
		key := genKey(un)
		err := redis_client.Cmd("GET", key).Err
		if err != nil {
			fmt.Printf("user-redis: Error: Attempt to delete non-existend user %s\n", *OptionDel)
			os.Exit(2)
		}

		redis_client.Cmd("DEL", key)
	} else if *OptionMod != "" {
		if *Realm == "" || *Password == "" {
			Usage()
		}

		un := *OptionMod
		key := genKey(un)
		id0, _ := uuid.NewV4()
		user_id := id0.String()
		old, err := redis_client.Cmd("GET", key).Str()
		if err == nil {
			t := strings.Split(old, ":")
			if len(t) > 2 {
				user_id = t[2]
			}
		}

		salt := genSalt()
		dk := fmt.Sprintf("%x", pbkdf2.Key([]byte(*Password), []byte(salt), NIterations, 64, sha256.New))
		value := salt + ":" + dk + ":" + user_id

		redis_client.Cmd("SET", key, value)
	} else {
		fmt.Printf("user-redis: Error: Invalid combination of options\n")
		Usage()
	}

}

// RedisHost   string `short:"H" long:"host"         description:"Redis host to connect to"                    default:"127.0.0.1"`
// RedisPort   string `short:"P" long:"port"         description:"Redis port to connect to"                    default:"6379"`
// RedisAuth   string `short:"A" long:"redisauth"    description:"Redis auth token (leave empty if no auth)"   default:""`
// RedisPrefix string `short:"T" long:"redisprefix"  description:"Redis key prefix"                            default:"BasicAuth"`
//func usage() {
//	fmt.Printf(`Usage: user-redis -a user -p passowrd -r realm
//     user-redis -d user -r realm
//     user-redis -m user -p password r realm
//You can use
//	-H host			host running Redis server
//	-P port			to specify a non standard Redis port
//	-A authStr		authorization token, not used if empty
//	-T prefix		String added to beginning of Redis key, default BasicAuth - matches with default in middleware.
//`)
//	os.Exit(2)
//}

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, `You can use
	-H host			host running Redis server	
	-P port			to specify a non standard Redis port	
	-A authStr		authorization token, not used if empty 
	-T prefix		String added to beginning of Redis key, default BasicAuth - matches with default in middleware.
`)
	flag.PrintDefaults()
	os.Exit(2)
}

func genSalt() (s string) {
	s = ""
	nRandBytes := 50
	buf := make([]byte, nRandBytes)
	_, err := rand.Read(buf)
	if err != nil {
		fmt.Println("user-redis: Error:", err)
		return
	}
	s = fmt.Sprintf("%x\n", buf)
	return
}

/* vim: set noai ts=4 sw=4: */
