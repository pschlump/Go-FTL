//
// user-pgsql - maintain uses in PostgreSQL for basicpgsql middleware
//
// Copyright (C) Philip Schlump, 2014-2015.
// Version: 0.5.9
// BuildNo: 1811
//

//
// How to use
//
// To add a user:
// 		$ user-pgsql -a username -p password -r realm
//
// To Delete a user:
// 		$ user-pgsql -d username
//
// To modify a users password:
// 		$ user-pgsql -m username -p password -r realm
//
//
// https://github.com/jackc/pgx/blob/master/examples/todo/main.go
//

package main

import (
	"crypto/rand"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"

	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/MiscLib" // "github.com/jackc/pgx"
	"golang.org/x/crypto/pbkdf2"  // https://github.com/golang/crypto/blob/master/pbkdf2/pbkdf2.go
)

//var opts struct {
//	PGConn           string `short:"C" long:"conn"         description:"Postgres connection info"                    default:""`
//	OptionAdd        string `short:"a" long:"add"          description:"To add user"                                 default:""`
//	OptionDel        string `short:"d" long:"delete"       description:"To delete user"                              default:""`
//	OptionMod        string `short:"m" long:"modify"       description:"To modify user"                              default:""`
//	Realm            string `short:"r" long:"realm"        description:"Realm name"                                  default:""`
//	Password         string `short:"p" long:"password"     description:"password"                                    default:""`
//	HashUsername     bool   `short:"h" long:"hashuser"     description:"Hash the Username"                           default:"false"`
//	HashUsernameSalt string `short:"s" long:"hashusersalt" description:"Salt for Username Hash"                      default:"8H3QhT9uHElh+c5NfowHx1gLeDw6qBMSTLvoL87GcB4FwflM8v2cTs"`
//}

// http://dba.stackexchange.com/questions/24370/how-to-use-aes-encryption-in-postgresql

var PGConn = flag.String("conn", "", "PotgresSQL connection info")                                                                     // 0
var DBName = flag.String("dbname", "test", "PotgresSQL database name")                                                                 // 8
var OptionAdd = flag.String("add", "", "To add user")                                                                                  // 1
var OptionDel = flag.String("delete", "", "To delete user")                                                                            // 2
var OptionMod = flag.String("modify", "", "To modify user")                                                                            // 3
var Realm = flag.String("realm", "", "Realm name")                                                                                     // 4
var Password = flag.String("password", "", "password")                                                                                 // 5
var HashUsername = flag.Bool("hashuser", false, "Hash the Username")                                                                   // 6
var HashUsernameSalt = flag.String("hashusersalt", "8H3QhT9uHElh+c5NfowHx1gLeDw6qBMSTLvoL87GcB4FwflM8v2cTs", "Salt for Username Hash") // 7
func init() {
	flag.StringVar(PGConn, "C", "", "PotgresSQL connection info")                                                             // 0
	flag.StringVar(PGConn, "N", "test", "PotgresSQL database name")                                                           // 8
	flag.StringVar(OptionAdd, "a", "", "To add user")                                                                         // 1
	flag.StringVar(OptionDel, "d", "", "To delete user")                                                                      // 2
	flag.StringVar(OptionMod, "m", "", "To modify user")                                                                      // 3
	flag.StringVar(Realm, "r", "", "Realm name")                                                                              // 4
	flag.StringVar(Password, "p", "", "password")                                                                             // 5
	flag.BoolVar(HashUsername, "h", false, "Hash the Username")                                                               // 6
	flag.StringVar(HashUsernameSalt, "s", "8H3QhT9uHElh+c5NfowHx1gLeDw6qBMSTLvoL87GcB4FwflM8v2cTs", "Salt for Username Hash") // 7
}

/*
// remember to parse...
func main() {

	// fns := flag.Args()

*/
// ===============================================================================================================================================
// var pg_client *pgx.Conn //
var pg_client *sizlib.MyDb // Client connection for PostgreSQL

const NIterations = 5000

// ===============================================================================================================================================
func main() {

	flag.Parse()

	ConnectToPostgreSQL()

	genKey := func(un string) (key string) {
		key = *Realm + ":" + un
		if *HashUsername {
			key = fmt.Sprintf("%x", pbkdf2.Key([]byte(key), []byte(*HashUsernameSalt), NIterations, 64, sha256.New))
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
		value := salt + ":" + dk

		// select
		rows, err := pg_client.Db.Query("select \"salt\", \"password\" from \"basic_auth\" where \"username\" = $1", key)
		if err != nil {
			fmt.Printf("Database error %s, attempting to validate user %s\n", err, un)
			return
		}

		nr := 0
		for ; rows.Next(); nr++ {
		}
		if nr != 0 {
			fmt.Printf("Error: Attempt to add when %s already exists in file\n", *OptionAdd)
			os.Exit(2)
		}

		// insert
		_, err = pg_client.Db.Exec("insert into \"basic_auth\" ( \"username\", \"salt\", \"password\" ) values( $1, $2, $3 )", key, salt, value)
		if err != nil {
			fmt.Printf("Error: Attempt to add user %s\n", err)
			os.Exit(2)
		}

	} else if *OptionDel != "" {
		if *Realm == "" {
			Usage()
		}

		un := *OptionDel
		key := genKey(un)

		// delete
		_, err := pg_client.Db.Exec("delete from \"basic_auth\" where \"username\" = $1", key)
		if err != nil {
			fmt.Printf("Error: Attempt to add user %s\n", err)
			os.Exit(2)
		}

	} else if *OptionMod != "" {
		if *Realm == "" || *Password == "" {
			Usage()
		}

		un := *OptionMod
		salt := genSalt()
		key := genKey(un)
		dk := fmt.Sprintf("%x", pbkdf2.Key([]byte(*Password), []byte(salt), NIterations, 64, sha256.New))
		value := salt + ":" + dk

		// select - if not found then, insert, else update
		rows, err := pg_client.Db.Query("select \"salt\", \"password\" from \"basic_auth\" where \"username\" = $1", key)
		if err != nil {
			fmt.Printf("Database error %s, attempting to validate user %s\n", err, un)
			return
		}

		nr := 0
		for ; rows.Next(); nr++ {
		}
		if nr == 0 {

			_, err = pg_client.Db.Exec("insert into \"basic_auth\" ( \"username\", \"salt\", \"password\" ) values( $1, $2, $3 )", key, salt, value)
			if err != nil {
				fmt.Printf("Error: Attempt to add user %s\n", err)
				os.Exit(2)
			}

		} else {
			_, err = pg_client.Db.Exec("update \"basic_auth\" set \"salt\" = $2, \"password\" = $3 where \"username\" = $1", key, salt, value)
			if err != nil {
				fmt.Printf("Error: Attempt to add user %s\n", err)
				os.Exit(2)
			}
		}

	} else {
		fmt.Printf("Error: Invalid combination of options\n")
		Usage()
	}

}

//func usage() {
//	fmt.Printf(`Usage: user-pgsql -C "connstring" -a user -p passowrd -r realm
//     user-pgsql -C "connstring" -d user -r realm
//     user-pgsql -C "connstring" -m user -p password r realm
//Connstring is
//	host:port:user:password:database
//	all as a single string with colons between fields.
//`)
//	os.Exit(2)
//}

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, `Connstring is
	host:port:user:password:database 
	all as a single string with colons between fields.
`)
	flag.PrintDefaults()
	os.Exit(2)
}

func ConnectToPostgreSQL() bool {
	//:pgx	conn, err := pgx.Connect(extractConfig(*PGConn))
	//:pgx	if err != nil {
	//:pgx		fmt.Fprintf(os.Stderr, "Unable to connection to database: %v\n", err)
	//:pgx		os.Exit(1)
	//:pgx	}

	//"PGConn": "user=pschlump password=f1ref0x2 sslmode=disable dbname=pschlump port=5433 host=127.0.0.1",
	//"DBName": "pschlump",
	// fmt.Printf("AT: %s\n", godebug.LF())
	conn := sizlib.ConnectToAnyDb("postgres", *PGConn, *DBName)
	if conn == nil {
		fmt.Fprintf(os.Stdout, "Unable to connection to database: %v\n", *DBName)
		fmt.Fprintf(os.Stderr, "%sUnable to connection to database: %v%s\n", MiscLib.ColorRed, *DBName, MiscLib.ColorReset)
		return false
	}

	if db11 {
		fmt.Fprintf(os.Stderr, "Success: Connected to PostgreSQL-server.\n")
	}

	pg_client = conn
	return true
}

//:pgxfunc extractConfig(PGConn string) (config pgx.ConnConfig) {
//:pgx
//:pgx	dflt := func(s, t string) (r string) {
//:pgx		r = s
//:pgx		if s == "" {
//:pgx			r = t
//:pgx		}
//:pgx		return
//:pgx	}
//:pgx
//:pgx	// host:user:pass:db
//:pgx	t := strings.Split(PGConn, ":")
//:pgx	if len(t) != 5 {
//:pgx		fmt.Printf("Invalid confuration should have Postgres Connect string of host:port:user:pass:db\n")
//:pgx		os.Exit(1)
//:pgx	}
//:pgx	config.Host = dflt(t[0], "127.0.0.1")
//:pgx	// tPort = dflt(t[1], "5432")	// xyzzy
//:pgx	p := 5432
//:pgx	if t[1] != "" {
//:pgx		x, err := strconv.ParseInt(t[1], 10, 32)
//:pgx		if err != nil {
//:pgx			fmt.Printf("invalid port in connection string: %s\n", err)
//:pgx			p = 5432
//:pgx		} else {
//:pgx			p = int(x)
//:pgx		}
//:pgx	}
//:pgx	config.Port = uint16(p)
//:pgx	config.User = dflt(t[2], "test")
//:pgx	config.Password = dflt(t[3], "password")
//:pgx	config.Database = dflt(t[4], "test")
//:pgx	return
//:pgx}

func genSalt() (s string) {
	s = ""
	nRandBytes := 50
	buf := make([]byte, nRandBytes)
	_, err := rand.Read(buf)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	s = fmt.Sprintf("%x\n", buf)
	return
}

const db11 = true

/* vim: set noai ts=4 sw=4: */
