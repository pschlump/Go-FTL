package sizlib

// (C) Copyright Philip Schlump, 2013-2018

// _ "github.com/mattn/go-oci8"			// OCI

import (
	// _ "../odbc" // _ "code.google.com/p/odbc"
	// _ "github.com/lib/pq"
	// _ "../pq" // _ "github.com/lib/pq"
	// _ "github.com/mattn/go-oci8"			// OCI
	// "database/sql"

	// "github.com/jackc/pgx" //  https://github.com/jackc/pgx

	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	_ "github.com/lib/pq"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/uuid"
	//	"encoding/json"
)

//:pgx:func extractConfig(auth string) pgx.ConnConfig {
//:pgx:	var config pgx.ConnConfig
//:pgx:	config.Host = "localhost"
//:pgx:	config.Database = "test"
//:pgx:	json.Unmarshal([]byte(auth), &config)
//:pgx:
//:pgx:	/*
//:pgx:		config.Host = os.Getenv("TODO_DB_HOST")
//:pgx:		if config.Host == "" {
//:pgx:			config.Host = "localhost"
//:pgx:		}
//:pgx:
//:pgx:		config.User = os.Getenv("TODO_DB_USER")
//:pgx:		if config.User == "" {
//:pgx:			config.User = os.Getenv("USER")
//:pgx:		}
//:pgx:
//:pgx:		config.Password = os.Getenv("TODO_DB_PASSWORD")
//:pgx:
//:pgx:		config.Database = os.Getenv("TODO_DB_DATABASE")
//:pgx:		if config.Database == "" {
//:pgx:			config.Database = "todo"
//:pgx:		}
//:pgx:	*/
//:pgx:
//:pgx:	return config
//:pgx:}

// -------------------------------------------------------------------------------------------------
// SET SCHEMA 'database_name'; -- Postgres way to set sechema to ...
// -------------------------------------------------------------------------------------------------
//:pgx:func ConnectToDb(auth string) *pgx.Conn {
//:pgx:	/*
//:pgx:		//db, err := sql.Open("odbc", "DSN=T1; UID=sa; PWD=f1ref0x12" )	// ODBC to Microsoft SQL Server
//:pgx:		//db, err := sql.Open("mymysql", "test/philip/f1ref0x12")		// mySQL
//:pgx:		// db, err := sql.Open("oci8", "scott/tiger@//192.168.0.101:1521/orcl")
//:pgx:		db, err := sql.Open("postgres", auth)
//:pgx:		if err != nil {
//:pgx:			panic(err)
//:pgx:		}
//:pgx:		db.SetMaxIdleConns(5)
//:pgx:	*/
//:pgx:	conn, err := pgx.Connect(extractConfig(auth))
//:pgx:	if err != nil {
//:pgx:		fmt.Fprintf(os.Stderr, "Unable to connection to database: %v\n", err)
//:pgx:		os.Exit(1)
//:pgx:	}
//:pgx:	return conn
//:pgx:}

func ConnectToDb(auth string) *sql.DB {
	db, err := sql.Open("postgres", auth) // ,"connectToPostgreSQL":"user=postgres password=f1ref0x2 dbname=test port=5432 host=192.168.0.181"
	// db, err := sql.Open("odbc", "DSN=T1; UID=sa; PWD=f1ref0x12" )	     // ODBC to Microsoft SQL Server
	// db, err := sql.Open("mymysql", "test/philip/f1ref0x12")		         // mySQL
	// db, err := sql.Open("oci8", "scott/tiger@//192.168.0.101:1521/orcl")  // Oracle
	if err != nil {
		panic(err)
	}
	db.SetMaxIdleConns(5)
	return db
}

type MyDb struct {
	//:pgx: Db     *pgx.Conn
	Db     *sql.DB
	DbType string
}

var DbBeginQuote = `"`
var DbEndQuote = `"`

func ConnectToAnyDb(db_type string, auth string, dbName string) *MyDb {
	mm := &MyDb{DbType: db_type}

	switch db_type {
	case "postgres":
		DbBeginQuote = `"`
		DbEndQuote = `"`
	case "oracle":
		os.Setenv("NLS_LANG", "")
		DbBeginQuote = `"`
		DbEndQuote = `"`
		db_type = "oci8"
	case "odbc":
		DbBeginQuote = `[`
		DbEndQuote = `]`
	default:
		panic("Invalid database type.")
	}

	db, err := sql.Open(db_type, auth)

	//db, err := sql.Open("odbc", "DSN=T1; UID=sa; PWD=f1ref0x12" )	// ODBC to Microsoft SQL Server
	//db, err := sql.Open("mymysql", "test/philip/f1ref0x12")		// mySQL
	//db, err := sql.Open("oci8", "scott/tiger@//192.168.0.101:1521/orcl")

	if err != nil {
		panic(err)
	}

	//:pgx: db := ConnectToDb(auth)
	mm.Db = db

	switch db_type {
	case "postgres":
		db.SetMaxIdleConns(5)
		// SET SCHEMA 'database_name'; -- Postgres way to set sechema to ...

	case "oci8":
		// set a default schema?? - or just use schema connected to?
		// No activity for now.

	case "odbc":
		err := Run1(db, "use "+dbName)
		if err != nil {
			fmt.Printf("Unable to set database, to %s, %s\n", dbName, err)
		}
	}

	return mm
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------
//:pgx:func GetColumns(rows *pgx.Rows) (columns []string, err error) {
//:pgx:	var fd []pgx.FieldDescription
//:pgx:	fd = rows.FieldDescriptions()
//:pgx:	columns = make([]string, 0, len(fd))
//:pgx:	for _, vv := range fd {
//:pgx:		columns = append(columns, vv.Name)
//:pgx:	}
//:pgx:	return
//:pgx:}

// -------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------
//:pgx:func RowsToInterface(rows *pgx.Rows) ([]map[string]interface{}, string, int) {
func RowsToInterface(rows *sql.Rows) ([]map[string]interface{}, string, int) {

	var finalResult []map[string]interface{}
	var oneRow map[string]interface{}
	var id string

	id = ""

	// Get column names
	columns, err := rows.Columns()
	//:pgx:columns, err := GetColumns(rows)
	if err != nil {
		panic(err.Error())
	}
	length := len(columns)

	// Make a slice for the values
	values := make([]interface{}, length)

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, length)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Fetch rows
	j := 0
	for rows.Next() {
		oneRow = make(map[string]interface{}, length)
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}

		// Print data
		for i, value := range values {
			// fmt.Printf ( "at top i=%d %T\n", i, value )
			switch value.(type) {
			case nil:
				// fmt.Println("n, %s", columns[i], ": NULL", godebug.LF())
				oneRow[columns[i]] = nil

			case []byte:
				// fmt.Printf("[]byte, len = %d, %s\n", len(value.([]byte)), godebug.LF())
				// if len==16 && odbc - then - convert from UniversalIdentifier to string (UUID convert?)
				if len(value.([]byte)) == 16 {
					// var u *uuid.UUID
					if uuid.IsUUID(fmt.Sprintf("%s", value.([]byte))) {
						u, err := uuid.Parse(value.([]byte))
						if err != nil {
							// fmt.Printf("Error: Invalid UUID parse, %s\n", godebug.LF())
							oneRow[columns[i]] = string(value.([]byte))
							if columns[i] == "id" && j == 0 {
								id = fmt.Sprintf("%s", value)
							}
						} else {
							if columns[i] == "id" && j == 0 {
								id = u.String()
							}
							oneRow[columns[i]] = u.String()
							// fmt.Printf(">>>>>>>>>>>>>>>>>> %s, %s\n", value, godebug.LF())
						}
					} else {
						if columns[i] == "id" && j == 0 {
							id = fmt.Sprintf("%s", value)
						}
						oneRow[columns[i]] = string(value.([]byte))
						// fmt.Printf(">>>>> 2 >>>>>>>>>>>>> %s, %s\n", value, godebug.LF())
					}
				} else {
					// Floats seem to end up at this point - xyzzy - instead of float64 -- so....  Need to check our column type info and see if 'f'  ---- xyzzy
					// fmt.Println("s", columns[i], ": ", string(value.([]byte)))
					if columns[i] == "id" && j == 0 {
						id = fmt.Sprintf("%s", value)
					}
					oneRow[columns[i]] = string(value.([]byte))
				}

			case int64:
				// fmt.Println("i, %s", columns[i], ": ", value, godebug.LF())
				// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )	// PJS-2014-03-06 - I suspect that this is a defect
				oneRow[columns[i]] = value

			case float64:
				// fmt.Println("f, %s", columns[i], ": ", value, godebug.LF())
				// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )
				// fmt.Printf ( "yes it is a float\n" )
				oneRow[columns[i]] = value

			case bool:
				// fmt.Println("b, %s", columns[i], ": ", value, godebug.LF())
				// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )		// PJS-2014-03-06
				// oneRow[columns[i]] = fmt.Sprintf ( "%t", value )		"true" or "false" as a value
				oneRow[columns[i]] = value

			case string:
				// fmt.Printf("string, %s\n", godebug.LF())
				if columns[i] == "id" && j == 0 {
					id = fmt.Sprintf("%s", value)
				}
				// fmt.Println("S", columns[i], ": ", value)
				oneRow[columns[i]] = fmt.Sprintf("%s", value)

			// Xyzzy - there is a timeNull structure in the driver - why is that not returned?  Maybee it is????
			// oneRow[columns[i]] = nil
			case time.Time:
				oneRow[columns[i]] = (value.(time.Time)).Format(ISO8601output)

			default:
				fmt.Printf("%s--- In default Case [%s] - %T %s\n", MiscLib.ColorRed, godebug.LF(), value, MiscLib.ColorReset)
				fmt.Fprintf(os.Stderr, "%s--- In default Case [%s] - %T %s\n", MiscLib.ColorRed, godebug.LF(), value, MiscLib.ColorReset)
				// fmt.Printf ( "default, yes it is a... , i=%d, %T\n", i, value, godebug.LF() )
				// fmt.Println("r", columns[i], ": ", value)
				if columns[i] == "id" && j == 0 {
					id = fmt.Sprintf("%v", value)
				}
				oneRow[columns[i]] = fmt.Sprintf("%v", value)
			}
			//fmt.Printf("\nType: %s\n", reflect.TypeOf(value))
		}
		// fmt.Println("-----------------------------------")
		finalResult = append(finalResult, oneRow)
		j++
	}
	return finalResult, id, j
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------
//:pgx:func SelQ(db *pgx.Conn, q string, data ...interface{}) (Rows *pgx.Rows, err error) {
func SelQ(db *sql.DB, q string, data ...interface{}) (Rows *sql.Rows, err error) {
	//godebug.TraceDb2("SelQ", q, data...)
	//godebug.TrIAmAt2(fmt.Sprintf("Query (%s) with data:", q))
	//godebug.DumpVar(data)
	if len(data) == 0 {
		Rows, err = db.Query(q)
	} else {
		Rows, err = db.Query(q, data...)
	}
	if err != nil {
		// tr.TraceDbError2("SelQ", q, err)
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("Database error (%v) at %s:%d, query=%s\n", err, file, line, q)
	}
	return
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------
//:pgx:func SelDataTrx(db *pgx.Conn, trx *tr.Trx, q string, data ...interface{}) []map[string]interface{} {
func SelDataTrx(db *sql.DB, trx *tr.Trx, q string, data ...interface{}) []map[string]interface{} {
	trx.SetQry(q, 2, data...)
	top_dir, err := SelData2(db, q, data...)
	if err == nil {
		trx.SetQryDone("", SVar(top_dir))
	} else {
		trx.SetQryDone(fmt.Sprintf("Error(10024): %v", err), "")
	}
	return top_dir
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------
//:pgx:func SelData2(db *pgx.Conn, q string, data ...interface{}) ([]map[string]interface{}, error) {
func SelData2(db *sql.DB, q string, data ...interface{}) ([]map[string]interface{}, error) {
	// 1 use "sel" to do the query
	// func sel ( res http.ResponseWriter, req *http.Request, db *pgx.Conn, q string, data ...interface{} ) ( Rows *sql.Rows, err error ) {
	Rows, err := SelQ(db, q, data...)

	if err != nil {
		fmt.Printf("Params: %s\n", SVar(data))
		// godebug.IAmAt2( fmt.Sprintf ( "Error (%s)", err ) )
		return make([]map[string]interface{}, 0, 1), err
	}

	rv, _, n := RowsToInterface(Rows)

	_ = n
	// tr.TraceDbEnd("SelData", q, n)
	return rv, err
}

// SelData seelct data from the database and return it.
//:pgx:func SelData(db *pgx.Conn, q string, data ...interface{}) []map[string]interface{} {
func SelData(db *sql.DB, q string, data ...interface{}) []map[string]interface{} {
	// 1 use "sel" to do the query
	// func sel ( res http.ResponseWriter, req *http.Request, db *pgx.Conn, q string, data ...interface{} ) ( Rows *sql.Rows, err error ) {
	// fmt.Printf("in SelData, %s\n", godebug.LF())

	Rows, err := SelQ(db, q, data...)

	if err != nil {
		fmt.Printf("Params: %s\n", SVar(data))
		return make([]map[string]interface{}, 0, 1)
	}

	rv, _, n := RowsToInterface(Rows)
	_ = n

	return rv
}

// -------------------------------------------------------------------------------------------------
// test: t-run1q.go, .sql, .out
// -------------------------------------------------------------------------------------------------
//:pgx:func Run1(db *pgx.Conn, q string, arg ...interface{}) error {
func Run1(db *sql.DB, q string, arg ...interface{}) error {
	//tr.TraceDb2 ( "Run1", q, arg... )
	//:pgx:	h := HashStr.HashStrToName(q) + q
	//:pgx:	ps, err := db.Prepare(h, q)
	//:pgx:	if err != nil {
	//:pgx:		//tr.TraceDbError2 ( "Run1.(Prepare)", q, err )
	//:pgx:		return err
	//:pgx:	}
	//:pgx:	_ = ps
	//:pgx:
	//:pgx:	// _, err = stmt.Exec(h, arg...)
	//:pgx:	_, err = db.Exec(h, arg...)
	//:pgx:	if err != nil {
	//:pgx:		//tr.TraceDbError2 ( "Run1.(Exec)", q, err )
	//:pgx:		return err
	//:pgx:	}
	//:pgx:
	//:pgx:	//tr.TraceDbEnd ( "Run1.(*Success*)", q, 0 )
	//:pgx:	return nil

	//tr.TraceDb2 ( "Run1", q, arg... )
	stmt, err := db.Prepare(q)
	if err != nil {
		//tr.TraceDbError2 ( "Run1.(Prepare)", q, err )
		return err
	}

	_, err = stmt.Exec(arg...)
	if err != nil {
		//tr.TraceDbError2 ( "Run1.(Exec)", q, err )
		return err
	}

	//tr.TraceDbEnd ( "Run1.(*Success*)", q, 0 )
	return nil
}

//:pgx:func Run2(db *pgx.Conn, q string, arg ...interface{}) (nr int64, err error) {
//:pgx:	nr = 0
//:pgx:	err = nil
//:pgx:
//:pgx:	h := HashStr.HashStrToName(q) + q
//:pgx:
//:pgx:	//tr.TraceDb2 ( "Run1", q, arg... )
//:pgx:	ps, err := db.Prepare(h, q)
//:pgx:	if err != nil {
//:pgx:		//tr.TraceDbError2 ( "Run1.(Prepare)", q, err )
//:pgx:		return
//:pgx:	}
//:pgx:	_ = ps
//:pgx:
//:pgx:	// R, err := stmt.Exec(h, arg...)
//:pgx:	R, err := db.Exec(h, arg...)
//:pgx:	if err != nil {
//:pgx:		//tr.TraceDbError2 ( "Run1.(Exec)", q, err )
//:pgx:		return
//:pgx:	}
//:pgx:
//:pgx:	//nr, err = R.RowsAffected()
//:pgx:	//if err != nil {
//:pgx:	//	//tr.TraceDbError2 ( "Run1.(Exec)", q, err )
//:pgx:	//	return
//:pgx:	//}
//:pgx:	nr = R.RowsAffected()
//:pgx:
//:pgx:	//tr.TraceDbEnd ( "Run1.(*Success*)", q, 0 )
//:pgx:	return
//:pgx:}
func Run2(db *sql.DB, q string, arg ...interface{}) (nr int64, err error) {
	nr = 0
	err = nil

	//tr.TraceDb2 ( "Run1", q, arg... )
	stmt, err := db.Prepare(q)
	if err != nil {
		//tr.TraceDbError2 ( "Run1.(Prepare)", q, err )
		return
	}

	R, err := stmt.Exec(arg...)
	if err != nil {
		//tr.TraceDbError2 ( "Run1.(Exec)", q, err )
		return
	}

	nr, err = R.RowsAffected()
	if err != nil {
		//tr.TraceDbError2 ( "Run1.(Exec)", q, err )
		return
	}

	//tr.TraceDbEnd ( "Run1.(*Success*)", q, 0 )
	return
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------
//:pgx:func Run1Thx(db *pgx.Conn, trx *tr.Trx, q string, data ...interface{}) error {
//:pgx:	h := HashStr.HashStrToName(q) + q
//:pgx:
//:pgx:	trx.SetQry(q, 2, data...)
//:pgx:	ps, err := db.Prepare(h, q)
//:pgx:	if err != nil {
//:pgx:		trx.SetQryDone(fmt.Sprintf("Error(10026): during Prepare, Run1Thx, %v", err), "")
//:pgx:		return err
//:pgx:	}
//:pgx:	_ = ps
//:pgx:
//:pgx:	// _, err = stmt.Exec(h, data...)
//:pgx:	_, err = db.Exec(h, data...)
//:pgx:	if err != nil {
//:pgx:		trx.SetQryDone(fmt.Sprintf("Error(10027): during Exec, Run1Thx, %v", err), "")
//:pgx:		return err
//:pgx:	}
//:pgx:
//:pgx:	trx.SetQryDone("success", "")
//:pgx:	return nil
//:pgx:}

func Run1Thx(db *sql.DB, trx *tr.Trx, q string, data ...interface{}) error {
	trx.SetQry(q, 2, data...)
	stmt, err := db.Prepare(q)
	if err != nil {
		trx.SetQryDone(fmt.Sprintf("Error(10026): during Prepare, Run1Thx, %v", err), "")
		return err
	}

	_, err = stmt.Exec(data...)
	if err != nil {
		trx.SetQryDone(fmt.Sprintf("Error(10027): during Exec, Run1Thx, %v", err), "")
		return err
	}

	trx.SetQryDone("success", "")
	return nil
}

func Run1IdThx(db *sql.DB, trx *tr.Trx, q string, data ...interface{}) (err error, id string) {
	trx.SetQry(q, 2, data...)
	err = db.QueryRow(q, data...).Scan(&id)
	if err != nil {
		trx.SetQryDone(fmt.Sprintf("Error(10026): during Prepare, Run1Thx, %v", err), "")
		return
	}

	// INSERT INTO persons (lastname,firstname) VALUES ('Smith', 'John') RETURNING id;
	// var id int
	// err := db.QueryRow("INSERT INTO user (name) VALUES ('John') RETURNING id").Scan(&id)
	// if err != nil {

	trx.SetQryDone("success", "")
	return nil, id
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------
//:pgx:func InsUpd(db *pgx.Conn, ins string, upd string, mdata map[string]string) {
func InsUpd(db *sql.DB, ins string, upd string, mdata map[string]string) {
	ins_q := Qt(ins, mdata)
	// fmt.Printf("     insUpd(ins) %s\n", ins_q)
	err := Run1(db, ins_q)
	if err != nil {
		// fmt.Printf("Error (1) in insUpd = %s\n", err)
		upd_q := Qt(upd, mdata)
		// fmt.Printf("     insUpd(upd) %s\n", upd_q)
		err = Run1(db, upd_q)
		if err != nil {
			fmt.Printf("Error (2) in insUpd = %s\n", err)
		}
	}
}

// -------------------------------------------------------------------------------------------------
// xyzzy-Rewrite
//	mdata["group_id"] = insSel ( "select \"id\" from \"img_group\" where \"group_name\" = '%{user_id%}'",
// -------------------------------------------------------------------------------------------------
//:pgx:func InsSel(db *pgx.Conn, sel string, ins string, mdata map[string]string) (id string) {
func InsSel(db *sql.DB, sel string, ins string, mdata map[string]string) (id string) {

	id = ""
	q := Qt(sel, mdata)

	Rows, err := db.Query(q)
	if err != nil {
		fmt.Printf("Error (237) on talking to database, %s\n", err)
		return
	} else {
		defer Rows.Close()
	}

	var x_id string
	n_row := 0
	for Rows.Next() {
		//  fmt.Printf ("Inside Rows Next\n" );
		n_row++
		err = Rows.Scan(&x_id)
		if err != nil {
			fmt.Printf("Error (249) on retreiving row from database, %s\n", err)
			return
		}
	}
	if n_row > 1 {
		fmt.Printf("Error (260) too many rows returned, n_rows=%d\n", n_row)
		return
	}
	if n_row == 1 {
		id = x_id
		return
	}

	y_id, _ := uuid.NewV4()
	id = y_id.String()
	mdata["id"] = id

	q = Qt(ins, mdata)

	Run1(db, q)
	return
}

// -------------------------------------------------------------------------------------------------
// Run a database query, return the rows.  Handle errors.
// -------------------------------------------------------------------------------------------------
//func Sel ( res http.ResponseWriter, req *http.Request, db *pgx.Conn, q string, data ...interface{} ) ( Rows *sql.Rows, err error ) {
//	godebug.TrIAmAt2( fmt.Sprintf ( "Query (%s) with data:", q ) )
//	godebug.DumpVar ( data )
//	Rows, err = db.Query(q, data... )
//	tr.TraceDb ( "Sel", q, data... )
//	if err != nil {
//
//		tr.TraceDbError ( "Sel", q, err )
//		_, file, line, _ := runtime.Caller(2)
//		fmt.Printf ( "Database error (%v) at %s:%d\n", err, file, line ) // Xyzzy - need to escape quotes and pass this back in JSON - what about '\' and '''' - encode those?
//
//		_, file, line, _ = runtime.Caller(1)
//		fmt.Printf ( "Database error (%v) at %s:%d\n", err, file, line ) // Xyzzy - need to escape quotes and pass this back in JSON - what about '\' and '''' - encode those?
//																			// Xyzzy - really should log this
//		detail := fmt.Sprintf ( "%v", err )
//		detail = strings.Replace(detail,"\"","\\\"",-1)
//		io.WriteString(res,JsonP(fmt.Sprintf("{\"status\":\"error\",\"code\":\"625\",\"msg\":\"Database error\",\"file\":\"%s\",\"line\":%d,\"detail\":\"%s\"}",file,line,detail),res,req))
//
//	}
//	return
//}

// -------------------------------------------------------------------------------------------------
// Rows to JSON -- Go from a set of "rows" returned by db.Query to a JSON string.
// -------------------------------------------------------------------------------------------------
//:pgx:func RowsToJson(rows *pgx.Rows) (string, string) {
func RowsToJson(rows *sql.Rows) (string, string) {

	var finalResult []map[string]interface{}
	var oneRow map[string]interface{}
	var id string

	id = ""

	// Get column names
	columns, err := rows.Columns()
	//:pgx:columns, err := GetColumns(rows)
	if err != nil {
		panic(err.Error())
	}
	length := len(columns)

	// Make a slice for the values
	values := make([]interface{}, length)

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, length)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Fetch rows
	j := 0
	for rows.Next() {
		oneRow = make(map[string]interface{}, length)
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}

		// Print data
		for i, value := range values {
			// fmt.Printf ( "at top i=%d %T\n", i, value )
			switch value.(type) {
			case nil:
				// fmt.Println("n", columns[i], ": NULL")
				oneRow[columns[i]] = nil

			case []byte:
				// Floats seem to end up at this point - xyzzy - instead of float64 -- so....  Need to check our column type info and see if 'f'  ---- xyzzy
				// fmt.Println("s", columns[i], ": ", string(value.([]byte)))
				if columns[i] == "id" && j == 0 {
					id = fmt.Sprintf("%s", value)
				}
				oneRow[columns[i]] = string(value.([]byte))

			case int64:
				// fmt.Println("i", columns[i], ": ", value)
				// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )	// PJS-2014-03-06 - I suspect that this is a defect
				oneRow[columns[i]] = value

			case float64:
				//fmt.Println("f", columns[i], ": ", value)
				// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )
				// fmt.Printf ( "yes it is a float\n" )
				oneRow[columns[i]] = value

			case bool:
				//fmt.Println("b", columns[i], ": ", value)
				// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )		// PJS-2014-03-06
				// oneRow[columns[i]] = fmt.Sprintf ( "%t", value )		"true" or "false" as a value
				oneRow[columns[i]] = value

			case string:
				if columns[i] == "id" && j == 0 {
					id = fmt.Sprintf("%s", value)
				}
				// fmt.Println("S", columns[i], ": ", value)
				oneRow[columns[i]] = fmt.Sprintf("%s", value)

			// Xyzzy - there is a timeNull structure in the driver - why is that not returned?  Maybee it is????
			case time.Time:
				//fmt.Printf("time.Time - %s, %s\n", columns[i], godebug.LF())
				//oneRow[columns[i]] = value
				oneRow[columns[i]] = (value.(time.Time)).Format(ISO8601output)

			default:
				// fmt.Printf ( "default, yes it is a... , i=%d, %T\n", i, value )
				// fmt.Println("r", columns[i], ": ", value)
				if columns[i] == "id" && j == 0 {
					id = fmt.Sprintf("%v", value)
				}
				oneRow[columns[i]] = fmt.Sprintf("%v", value)
			}
			//fmt.Printf("\nType: %s\n", reflect.TypeOf(value))
		}
		// fmt.Println("-----------------------------------")
		finalResult = append(finalResult, oneRow)
		j++
	}
	if j > 0 {
		s, err := json.MarshalIndent(finalResult, "", "\t")
		if err != nil {
			fmt.Printf("Unable to convert to JSON data, %v\n", err)
		}
		return string(s), id
	} else {
		return "[]", ""
	}
}

// -------------------------------------------------------------------------------------------------
// Rows to JSON -- Go from a set of "rows" returned by db.Query to a JSON string.
// -------------------------------------------------------------------------------------------------
//:pgx:func RowsToJsonFirstRow(rows *pgx.Rows) (string, string) {
func RowsToJsonFirstRow(rows *sql.Rows) (string, string) {

	// var finalResult   []map[string]interface{}
	var oneRow map[string]interface{}
	var id string

	id = ""

	// Get column names
	columns, err := rows.Columns()
	//:pgx:columns, err := GetColumns(rows)
	if err != nil {
		panic(err.Error())
	}
	length := len(columns)

	// Make a slice for the values
	values := make([]interface{}, length)

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, length)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Fetch rows
	j := 0
	for rows.Next() {
		oneRow = make(map[string]interface{}, length)
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}

		if j == 0 {

			// Print data
			for i, value := range values {
				// fmt.Printf ( "at top i=%d %T\n", i, value )
				switch value.(type) {
				case nil:
					// fmt.Println("n", columns[i], ": NULL")
					oneRow[columns[i]] = nil

				case []byte:
					// Floats seem to end up at this point - xyzzy - instead of float64 -- so....  Need to check our column type info and see if 'f'  ---- xyzzy
					// fmt.Println("s", columns[i], ": ", string(value.([]byte)))
					if columns[i] == "id" && j == 0 {
						id = fmt.Sprintf("%s", value)
					}
					oneRow[columns[i]] = string(value.([]byte))

				case int64:
					// fmt.Println("i", columns[i], ": ", value)
					// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )	// PJS-2014-03-06 - I suspect that this is a defect
					oneRow[columns[i]] = value

				case float64:
					//fmt.Println("f", columns[i], ": ", value)
					// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )
					// fmt.Printf ( "yes it is a float\n" )
					oneRow[columns[i]] = value

				case bool:
					//fmt.Println("b", columns[i], ": ", value)
					// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )		// PJS-2014-03-06
					// oneRow[columns[i]] = fmt.Sprintf ( "%t", value )		"true" or "false" as a value
					oneRow[columns[i]] = value

				case string:
					if columns[i] == "id" && j == 0 {
						id = fmt.Sprintf("%s", value)
					}
					// fmt.Println("S", columns[i], ": ", value)
					oneRow[columns[i]] = fmt.Sprintf("%s", value)

				// Xyzzy - there is a timeNull structure in the driver - why is that not returned?  Maybee it is????
				case time.Time:
					oneRow[columns[i]] = (value.(time.Time)).Format(ISO8601output)

				default:
					// fmt.Printf ( "default, yes it is a... , i=%d, %T\n", i, value )
					// fmt.Println("r", columns[i], ": ", value)
					if columns[i] == "id" && j == 0 {
						id = fmt.Sprintf("%v", value)
					}
					oneRow[columns[i]] = fmt.Sprintf("%v", value)
				}
				//fmt.Printf("\nType: %s\n", reflect.TypeOf(value))
			}
		}
		// fmt.Println("-----------------------------------")
		// finalResult = append ( finalResult, oneRow )
		j++
	}
	if j > 0 {
		s, err := json.MarshalIndent(oneRow, "", "\t")
		if err != nil {
			fmt.Printf("Unable to convert to JSON data, %v\n", err)
		}
		return string(s), id
	} else {
		return "{}", ""
	}
}
