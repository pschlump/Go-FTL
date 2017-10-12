//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1234
//

//
// Package Dump Request print out the request.
//

package DumpRequest

import (
	"fmt"
	"net/http"
	"os"

	"www.2c-why.com/JsonX"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------
//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*DumpRequestType)
//		if ok {
//			pCfg.SetNext(next)
//			rv = pCfg
//		} else {
//			err = mid.FtlConfigError
//			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
//		}
//		return
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} { return &DumpRequestType{} }
//
//	postInit := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
//
//		hh, ok := h.(*DumpRequestType)
//		if !ok {
//			// logrus.Warn(fmt.Sprintf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo))
//			fmt.Printf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo)
//			return mid.ErrInternalError
//		} else {
//			hh.outputFile = os.Stdout
//			if hh.FileName != "" {
//				var err error
//				hh.outputFile, err = lib.Fopen(hh.FileName, "a")
//				if err != nil {
//					fmt.Printf("Error: Unable to open %s for append, Error: %s Line No:%d\n", hh.FileName, err, hh.LineNo)
//					return mid.ErrInternalError
//				}
//			}
//			hh.final, _ = lib.ParseBool(hh.Final)
//		}
//
//		return nil
//	}
//
//	cfg.RegInitItem2("DumpRequest", initNext, createEmptyType, postInit, `{
//		"Paths":        { "type":["string","filepath"], "isarray":true, "default":"/" },
//		"Msg":          { "type":[ "string" ], "default":"" },
//		"Final":        { "type":[ "string" ], "default":"no" },
//		"FileName":     { "type":[ "string","filepath" ], "default":"" },
//		"LineNo":       { "type":[ "int" ], "default":"1" }
//		}`)
//}

// normally identical
func (hdlr *DumpRequestType) SetNext(next http.Handler) {
	hdlr.Next = next
}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &DumpRequestType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("DumpRequest", CreateEmpty, `{
		"Paths":        { "type":["string","filepath"], "isarray":true, "default":"/" },
		"Msg":          { "type":[ "string" ], "default":"" },
		"Final":        { "type":[ "string" ], "default":"no" },
		"FileName":     { "type":[ "string","filepath" ], "default":"" },
		"LineNo":       { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *DumpRequestType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *DumpRequestType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	hdlr.outputFile = os.Stdout
	if hdlr.FileName != "" {
		var err error
		hdlr.outputFile, err = lib.Fopen(hdlr.FileName, "a")
		if err != nil {
			fmt.Printf("Error: Unable to open %s for append, Error: %s Line No:%d\n", hdlr.FileName, err, hdlr.LineNo)
			return mid.ErrInternalError
		}
	}
	hdlr.final, _ = lib.ParseBool(hdlr.Final)
	return
}

var _ mid.GoFTLMiddleWare = (*DumpRequestType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type DumpRequestType struct {
	Next       http.Handler
	Paths      []string
	Msg        string
	Final      string
	FileName   string
	LineNo     int
	outputFile *os.File
	final      bool // t/f converted version of .Final
}

// Parameterized for testing? or just change the test
func NewDumpRequestServer(n http.Handler, p []string, m string, fn string) *DumpRequestType {
	return &DumpRequestType{Next: n, Paths: p, Msg: m, FileName: fn}
}

func (hdlr *DumpRequestType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "DumpRequest", hdlr.Paths, pn, req.URL.Path)

			// -----------------------------------------------------------------------------------------------------------------
			// -----------------------------------------------------------------------------------------------------------------
			// -----------------------------------------------------------------------------------------------------------------
			// -----------------------------------------------------------------------------------------------------------------
			//			if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" || req.Method == "DELETE" {
			//				ct := req.Header.Get("Content-Type")
			//				fmt.Printf("ct [%s] AT %s\n", ct, godebug.LF())
			//				if req.PostForm == nil {
			//					req.PostForm = url.Values{}
			//					fmt.Printf("AT %s\n", godebug.LF())
			//					if strings.HasPrefix(ct, "application/json") {
			//						fmt.Printf("AT %s\n", godebug.LF())
			//						body, err2 := ioutil.ReadAll(req.Body)
			//						if err2 != nil {
			//							fmt.Printf("Error(20008): Malformed JSON body, RequestURI=%s err=%v\n", req.RequestURI, err2)
			//						}
			//						fmt.Printf("THIS ONE                                           !!!!!!!!!!!!!!! body >%s< AT %s\n", body, godebug.LF())
			//						var jsonData map[string]interface{}
			//						err := json.Unmarshal(body, &jsonData)
			//						if err == nil {
			//							for Name, v := range jsonData {
			//								Value := ""
			//								switch v.(type) {
			//								case bool:
			//									Value = fmt.Sprintf("%v", v)
			//								case float64:
			//									Value = fmt.Sprintf("%v", v)
			//								case int64:
			//									Value = fmt.Sprintf("%v", v)
			//								case int32:
			//									Value = fmt.Sprintf("%v", v)
			//								case time.Time:
			//									Value = fmt.Sprintf("%v", v)
			//								case string:
			//									Value = fmt.Sprintf("%v", v)
			//								default:
			//									Value = fmt.Sprintf("%s", godebug.SVar(v))
			//								}
			//								req.PostForm.Add(Name, Value) // AddValueToParams(Name, Value, 'b', FromBodyJson, ps)
			//							}
			//						} else {
			//							fmt.Printf("Error: in parsing JSON data >%s< Error: %s\n", body, err)
			//						}
			//					} else {
			//						fmt.Printf("AT %s\n", godebug.LF())
			//						err := req.ParseForm()
			//						if err != nil {
			//							fmt.Printf("Error(20010): Malformed body, RequestURI=%s err=%v\n", req.RequestURI, err)
			//						}
			//					}
			//				}
			//			}
			// -----------------------------------------------------------------------------------------------------------------
			// -----------------------------------------------------------------------------------------------------------------
			// -----------------------------------------------------------------------------------------------------------------
			// -----------------------------------------------------------------------------------------------------------------

			fmt.Fprintf(hdlr.outputFile, "DumpRequest %s = %s\n", hdlr.Msg, lib.SVarI(req))

			if !hdlr.final {
				hdlr.Next.ServeHTTP(rw, req)
			} else {
				fmt.Fprintf(rw, "\n%s\nOriginalURL: %s, Method:%s, TrxCookie=%s, %s\n%s\n",
					"--------------------------------------------------------------------------------------------------------------------------------",
					rw.OriginalURL, req.Method, rw.RequestTrxId, godebug.LF(),
					"--------------------------------------------------------------------------------------------------------------------------------")
				fmt.Fprintf(rw, "\nParams + Cookies for (%s): %s AT %s\n", req.URL.Path, rw.Ps.DumpParamTable(), godebug.LF())
				fmt.Fprintf(rw, `%s`+"\n", lib.SVarI(req))
			}
		} else {
			logrus.Warn(fmt.Sprintf("Error: DumpRequest: Line No:%d  %s\n", hdlr.LineNo, mid.ErrNonMidBufferWriter))
			www.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		hdlr.Next.ServeHTTP(www, req)
	}
}

/* vim: set noai ts=4 sw=4: */
