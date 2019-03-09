package sizlib

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/pschlump/godebug"
	"github.com/pschlump/json"
)

type SelCfgType struct {
	DbConnStr string
	Query     string
	Param     []string
}

func Test_SelQ_001(t *testing.T) {
	b := Exists("test_con.json")
	if b {

		data, err = ioutil.ReadFile("test_con.json")
		if err != nil {
			fmt.Printf("Skipping test - failed to read test_con.json\n")
			return
		}
		var SelCfg SelCfgType
		err = json.Unmarshal(data, &SelCfg)

		fmt.Printf("Input: %s\n", godebug.SVarI(SelCfg))

		// jDb := ConnectToAnyDb("postgres", SelCfg.DbConnStr, "pschlump")
		Db := ConnectToDb(SelCfg.DbConnStr)

		// rows, err := SelData2(db *sql.DB, q string, data ...interface{}) ([]map[string]interface{}, error) {

		var rows []map[string]interface{}
		if len(SelCfg.Param) > 0 {
			rows, err = SelData2(Db, SelCfg.Query, SelCfg.Param...)
		} else {
			rows, err = SelData2(Db, SelCfg.Query, SelCfg.Param...)
		}

		fmt.Printf("%s\n", godebug.SVarI(rows))

	}
}
