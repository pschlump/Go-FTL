package main

// Updated to use new output format. ----  Wed Dec 26 12:50:04 MST 2018

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/sizlib"
	// "github.com/pschlump/godebug"
)

//
// Tool to generate default configurations from the table in the database
//
// -g <cfg> to read in the connection information - this can be the same configuration file that is used by Go-FTL server
//
// A list of tables to describe and generate a configuration for.
//

// Note: http://stackoverflow.com/questions/2146705/select-datatype-of-the-field-in-postgres
/*
select column_name,
case
    when domain_name is not null then domain_name
    when data_type='character varying' THEN 'varchar('||character_maximum_length||')'
    when data_type='numeric' THEN 'numeric('||numeric_precision||','||numeric_scale||')'
    else data_type
end as myType
from information_schema.columns
where table_name='test'
*/

var PGConn = flag.String("cfg", "global-cfg.json", "PotgresSQL connection info")  // 0
var CaseIns = flag.Bool("caseIns", true, "Case insenstivie match of table names") // 1
func init() {
	flag.StringVar(PGConn, "g", "global-cfg.json", "PotgresSQL connection info") // 0
}

func main() {

	flag.Parse()
	fns := flag.Args()

	if len(fns) == 0 {
		fmt.Printf("Must sepecify atleast 1 table to process\n")
		os.Exit(1)
	}

	cfg.SetupPgSqlForTest(*PGConn)

	fmt.Printf(`{
	"note:generated":"Tables: %s"

`, fns)

	for ii, vv := range fns {
		_ = ii

		Query := `SELECT column_name, data_type FROM information_schema.columns WHERE table_name = $1`
		if *CaseIns {
			Query = `SELECT column_name, data_type FROM information_schema.columns WHERE table_name ilike $1`
		}

		Rows, err := cfg.ServerGlobal.Pg_client.Db.Query(Query, vv)
		if err != nil {
			fmt.Printf("Error on getting columns for %s: %s\n", vv, err)
		} else {
			rv, _, _ := sizlib.RowsToInterface(Rows)
			Rows.Close()

			PkQuery := `SELECT a.attname, format_type(a.atttypid, a.atttypmod) AS data_type
FROM   pg_index i
JOIN   pg_attribute a ON a.attrelid = i.indrelid
                     AND a.attnum = ANY(i.indkey)
WHERE  i.indrelid = $1::regclass
AND    i.indisprimary
`
			PkRows, err := cfg.ServerGlobal.Pg_client.Db.Query(PkQuery, vv)
			if err != nil {
				fmt.Printf("Error getting PK: %s\n", err)
			}
			Pk, _, _ := sizlib.RowsToInterface(PkRows)
			PkRows.Close()

			// fmt.Printf("Data: %s\n", godebug.SVarI(rv))
			// fmt.Printf("PkData: %s\n", godebug.SVarI(Pk))

			fmt.Printf(`
	,"/api/table/%s": { "crud": [ "select", "insert", "update", "delete", "info" ]
		, "TableName": "%s"
		, "LineNo":"Line: __LINE__ File: __FILE__"
		, "p": [ ]
		, "LoginRequired":false
		, "Method":["GET","POST","PUT","DELETE"]
		, "cols": [
`, vv, vv)
			com := " "
			maxS := 0
			for _, aCol := range rv {
				colName := aCol["column_name"].(string)
				if len(colName) > maxS {
					maxS = len(colName)
				}
			}
			for jj, aCol := range rv {
				_ = jj
				colName := aCol["column_name"].(string)
				dt := "s"
				r_dt := aCol["data_type"].(string)
				switch r_dt {
				case "integer":
					dt = "i"
				case "text", "character varying":
				default:
					if strings.HasPrefix(r_dt, "timestamp ") {
						dt = "t"
					}
				}
				// fmt.Printf("col_no: %d is %s\n", jj, godebug.SVar(aCol))
				// xyzzy { "colName": "id" 				, "colType": "i",				   "insert":true, "autoGen": true, "isPk": true }
				if IsPkCol(colName, Pk) {
					fmt.Printf(`			%s { "colName": "%s"%s, "colType": "%s",	               "insert":true, "autoGen":true, "isPk": true }`+"\n", com, colName, nb(maxS-len(colName)+1), dt)
				} else {
					fmt.Printf(`			%s { "colName": "%s"%s, "colType": "%s",	"update":true, "insert":true		}`+"\n", com, colName, nb(maxS-len(colName)+1), dt)
				}
				com = ","
			}

			fmt.Printf(
				`			]
		}
`)

		}
	}
	fmt.Printf("}\n")

}

func IsPkCol(colName string, Pk []map[string]interface{}) bool {
	for _, vv := range Pk {
		cn := vv["attname"].(string)
		if cn == colName {
			return true
		}
	}
	return false
}

/*
	,"/api/table/log": { "crud": [ "select", "insert", "update", "delete", "info" ]
		, "Comment": "test (1) ability to query all data (2) ability to get back a single row via PK query"
		, "TableName": "log"
		, "LineNo":"__LINE__"
		, "nokey":true
		, "Method":["GET","POST","PUT","DELETE","HEAD"]
		, "cols": [
				  { "colName": "id" 				, "colType": "i",				   "insert":true, "autoGen": true, "isPk": true }
				, { "colName": "log_timestamp"		, "colType": "d",	"update":true, "insert":true		}
				, { "colName": "error_level"		, "colType": "i",	"update":true, "insert":true		}
				, { "colName": "message"			, "colType": "s",	"update":true, "insert":true		}
				, { "colName": "source"				, "colType": "s",	"update":true, "insert":true		}
			]
		}
*/

func nb(n int) (s string) {
	s = ""
	for i := 0; i < n; i++ {
		s += " "
	}
	return
}
