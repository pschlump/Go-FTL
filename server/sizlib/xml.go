package sizlib

// (C) Copyright Philip Schlump, 2013-2014

// _ "github.com/mattn/go-oci8"			// OCI

// _ "../odbc" // _ "code.google.com/p/odbc"
// _ "github.com/lib/pq"
// _ "../pq" // _ "github.com/lib/pq"
// _ "github.com/mattn/go-oci8"			// OCI
// "database/sql"

// "github.com/jackc/pgx" //  https://github.com/jackc/pgx
import (
	_ "github.com/lib/pq"

	"encoding/xml"
	"fmt"
)

// XVar convert a variable to it's XML representation
func XVar(v interface{}) string {
	s, err := xml.Marshal(v)
	// s, err := json.MarshalIndent ( v, "", "\t" )
	if err != nil {
		return fmt.Sprintf("Error:%s", err)
	} else {
		return string(s)
	}
}

// XVarI convert a variable ot it's XML representaiton with indented XML
func XVarI(v interface{}) string {
	// s, err := json.Marshal ( v )
	s, err := xml.MarshalIndent(v, "", "\t")
	if err != nil {
		return fmt.Sprintf("Error:%s", err)
	} else {
		return string(s)
	}
}
