//
// Test connection test to PoggreSQL
//
// This is the first thing that should be run.
//
// Example:
//	$ ./con-to-pg -C 'user=postgres password=f1ref0x2 dbname=test port=5432 host=127.0.0.1'
//
// TODO:
// 1. -g global-config.json file - read that for connection string/database-type etc.
// 2. -n Database - to set a specific database for non-PG
// 3. -d postgres|Oracle|T-SQL|ocbc etc. -- database type
//
// 4. Improve error reporting on ConnectToAnyDb and Run1
//

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/MiscLib"
)

var db *sql.DB

var PGConn = flag.String("conn", "", "PotgresSQL connection info") // 0
func init() {
	flag.StringVar(PGConn, "C", "", "PotgresSQL connection info") // 0
}

func main() {

	var err error
	var dbName string = "" // Not used for PostgreSQL???

	flag.Parse()

	auth := *PGConn

	db_x := sizlib.ConnectToAnyDb("postgres", auth, dbName)
	if db_x == nil {
		fmt.Fprintf(os.Stderr, "%sUnable to connection to database: %v%s\n", MiscLib.ColorRed, err, MiscLib.ColorReset)
		os.Exit(1)
	}
	db = db_x.Db

	data, err := sizlib.SelData2(db, "select \"id\" as \"x\" from \"xyzzy\" limit 1")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sUnable to connection to database/failed on table select: %v%s\n", MiscLib.ColorRed, err, MiscLib.ColorReset)
		os.Exit(1)
	}

	fmt.Printf("Data=%s\n", lib.SVarI(data))

	fmt.Printf("%sPASS Success!!! Connected to database%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
	os.Exit(0)

}
