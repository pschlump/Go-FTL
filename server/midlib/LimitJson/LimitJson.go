//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1254
//

//

package LimitJson

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"www.2c-why.com/JsonX"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &LimitJsonHandlerType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("LimitJson", CreateEmpty, `{
		"Paths":            { "type":["string","filepath"], "isarray":true, "required":true },
		"Allowed":       	{ "type":[ "struct" ] },
		"OnErrorDiscard":   { "type":[ "string" ], "default":"yes" },
		"LineNo":           { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *LimitJsonHandlerType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *LimitJsonHandlerType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*LimitJsonHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type LimitJsonType struct {
	Path         []string
	ItemsAllowed []string
	ItemsRemoved []string
}

type LimitJsonHandlerType struct {
	Next           http.Handler    //
	Paths          []string        // Paths that this will work for
	Allowed        []LimitJsonType // Limit to only these json items
	OnErrorDiscard string          //
	LineNo         int             //
}

func NewLimitJsonServer(n http.Handler, p []string) *LimitJsonHandlerType {
	return &LimitJsonHandlerType{
		Next:  n,
		Paths: p,
	}
}

func (hdlr *LimitJsonHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "LimitJson", hdlr.Paths, pn, req.URL.Path)

			hdlr.Next.ServeHTTP(rw, req)
			h := www.Header()
			ct := h.Get("Content-Type")
			trx.AddNote(1, fmt.Sprintf("Content-Type == %s StatusCode = %d", ct, rw.StatusCode))

			if db1 {
				fmt.Printf("%sLimitJson: hdlr.Allowed = %s, %s%s\n", MiscLib.ColorCyan, godebug.SVarI(hdlr.Allowed), godebug.LF(), MiscLib.ColorReset)
			}

			for ii := range hdlr.Allowed {
				if pn := lib.PathsMatchN(hdlr.Allowed[ii].Path, req.URL.Path); pn >= 0 {
					if db1 {
						fmt.Printf("\tLimitJson: found match at %d, %s\n", pn, godebug.LF())
					}
					if rw.StatusCode == http.StatusOK && strings.HasPrefix(ct, "application/json") {
						if db1 {
							fmt.Printf("\tLimitJson: is JSON %s\n", godebug.LF())
						}
						mdata := make(map[string]interface{})
						body := rw.GetBody()
						if db1 {
							fmt.Printf("\tLimitJson: body -->>%s<<-- %s\n", body, godebug.LF())
						}
						err := json.Unmarshal(body, &mdata)
						nItems := 0
						if err != nil {
							if db1 {
								fmt.Printf("\tLimitJson: Failed to parse, data=%s err=%s, %s\n", body, err, godebug.LF())
							}
							if hdlr.OnErrorDiscard == "yes" {
								fmt.Fprintf(os.Stderr, "%sData Discarded - due to syntax error%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
								fmt.Fprintf(os.Stdout, "%sData Discarded - due to syntax error%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
								www.WriteHeader(http.StatusInternalServerError)
								rw.ReplaceBody([]byte("{}"))
								rw.SaveDataInCache = false
								return
							}
						} else {
							if db1 {
								fmt.Printf("\tLimitJson: will proces %s\n", godebug.LF())
							}
							for _, key := range hdlr.Allowed[ii].ItemsRemoved {
								if db1 {
									fmt.Printf("\tLimitJson: delete of item [%s], %s\n", key, godebug.LF())
								}
								delete(mdata, key)
								nItems++
							}
							if len(hdlr.Allowed[ii].ItemsAllowed) > 0 {
								if db1 {
									fmt.Printf("\tLimitJson: items allowed processing, %s\n", godebug.LF())
								}
								for key := range mdata {
									if !lib.InArray(key, hdlr.Allowed[ii].ItemsAllowed) {
										if db1 {
											fmt.Printf("\tLimitJson: delete of item [%s], %s\n", key, godebug.LF())
										}
										delete(mdata, key)
										nItems++
									}
								}
							}
							if nItems > 0 {
								newData := godebug.SVar(mdata)
								rw.ReplaceBody([]byte(newData))
								if db1 {
									fmt.Printf("\tLimitJson: newData -->>%s<<-- %s\n", newData, godebug.LF())
								}
								rw.SaveDataInCache = false
							}
						}
					}
				}
			}

		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		hdlr.Next.ServeHTTP(www, req)
	}

}

const db1 = false

/* vim: set noai ts=4 sw=4: */
