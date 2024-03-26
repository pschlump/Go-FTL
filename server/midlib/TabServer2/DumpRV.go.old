package TabServer2

import (
	"fmt"
	"net/http"

	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
)

func DumpRV(www http.ResponseWriter, req *http.Request, cfgTag string, rv string, isError bool, cookieList map[string]string, ps *goftlmux.Params, trx *tr.Trx, hdlr *TabServer2Type) (rvOut string, pptFlag PrePostFlagType, exit bool, a_status int) {
	fmt.Printf("%srv ->%s<- AT:%s %s\n", MiscLib.ColorYellow, rv, godebug.LF(), MiscLib.ColorReset)
	return rv, PrePostContinue, false, 200
}

/* vim: set noai ts=4 sw=4: */
