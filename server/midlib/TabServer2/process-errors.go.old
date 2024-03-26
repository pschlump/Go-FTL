package TabServer2

import (
	"fmt"
	"net/http"
	"os"

	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/uuid"
)

// type FuncMapType func(res http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rvOut string, pptFlag PrePostFlagType, exitProcessing bool, status int)

type ErrorLog struct {
	LineFile1   string        //
	LineFile2   string        //
	Msg         string        //
	Code        string        //
	Status      string        //
	SQLStmt     string        //
	Data        []interface{} //
	CallType    string        //
	UniqueLogID string        //
	RvDiscarded string        //
}

// Really needs to be an Interface{ ProcessErros, SetupLog } that allows caputre of log messages
// and plug in of different error handeling.  Use a global with ability to set your own
// xyzzy - TODO -

var logFile *os.File

func (hdlr *TabServer2Type) ProcessErrors(
	www http.ResponseWriter, // Ability to create responce and record info from request
	req *http.Request, //       Ability to create responce and record info from request
	rvIn string, //             Current rv for pre-post processing.
	errorCode int, //           Unique "code" for error
	msg string, //              Error message
	status int, // 	            current status
	stmt string, //             if SQL then the SQL Statment
	data []interface{}, //      The set of data for the SQL bind parameters
	ct string, //               current operation, 'g', 'f', etc.
) (rv string) {
	rv = rvIn
	id, _ := uuid.NewV4()
	idStr := id.String()
	e := ErrorLog{
		LineFile1:   godebug.LF(2),
		LineFile2:   godebug.LF(3),
		Msg:         msg,
		Code:        fmt.Sprintf("%d", errorCode),
		Status:      fmt.Sprintf("%d", status),
		SQLStmt:     stmt,
		Data:        data,
		CallType:    ct,
		UniqueLogID: idStr,
		RvDiscarded: rvIn,
	}
	if logFile != nil {
		fmt.Fprintf(logFile, "%s\n", godebug.SVarI(e))
	} else {
		fmt.Fprintf(os.Stdout, "%s\n", godebug.SVarI(e))
	}
	// xyzzy - should be able to configure errors to stderr?
	fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, godebug.SVar(e), MiscLib.ColorReset)

	if hdlr.StatusForAllErrors == "yes" {
		// xyzzy - www.SetStatus(status)
		if status == 200 || status == 304 {
			www.WriteHeader(http.StatusInternalServerError) // 501
		} else {
			www.WriteHeader(status)
		}
	} else {
		fmt.Fprintf(www, `{"status":"error","msg":%q,"code":%d,"uniqueLogId":%q}`, msg, errorCode, idStr)
	}
	return
}
