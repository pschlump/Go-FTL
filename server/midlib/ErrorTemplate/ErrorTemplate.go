//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1282
//

//
// Map errors onto error template files
//

package ErrorTemplate

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	JsonX "github.com/pschlump/JSONx"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/tmplp"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*ErrorTemplateType)
//		if ok {
//			pCfg.SetNext(next)
//			rv = pCfg
//		} else {
//			err = mid.FtlConfigError
//			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
//		}
//		pCfg.gCfg = gCfg
//		return
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} { return &ErrorTemplateType{} }
//
//	postInitValidation := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
//		// fmt.Printf("In postInitValidation, h=%v\n", h)
//		hh, ok := h.(*ErrorTemplateType)
//		if !ok {
//			fmt.Printf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo)
//			return mid.ErrInternalError
//		} else {
//			hh.errorsTemplate = make(map[string]string)
//			for _, vv := range hh.Errors {
//				for _, ww := range hh.TemplatePath {
//					fn := ww + "/" + vv + ".tmpl"
//					if lib.Exists(fn) {
//						s, err := ioutil.ReadFile(fn)
//						if err == nil {
//							hh.errorsTemplate[vv] = string(s)
//						} else {
//							fmt.Printf("Error: Unable to read template file %s\n", err)
//							return mid.ErrInvalidConfiguration
//						}
//						break
//					}
//				}
//			}
//		}
//		return nil
//	}
//
//	cfg.RegInitItem2("ErrorTemplate", initNext, createEmptyType, postInitValidation, `{
//		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
//		"Errors":        { "type":[ "string" ], "isarray":true },
//		"TemplatePath":  { "type":[ "string" ], "isarray":true, "default":"./errorTemplates/" },
//		"LineNo":        { "type":[ "int" ], "default":"1" }
//		}`)
//
//}
//
//// normally identical
//func (hdlr *ErrorTemplateType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &ErrorTemplateType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("ErrorTemplate", CreateEmpty, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"Errors":        { "type":[ "string" ], "isarray":true },
		"TemplatePath":  { "type":[ "string" ], "isarray":true, "default":"./errorTemplates/" },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *ErrorTemplateType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *ErrorTemplateType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	hdlr.errorsTemplate = make(map[string]string)
	for _, vv := range hdlr.Errors {
		for _, ww := range hdlr.TemplatePath {
			fn := ww + "/" + vv + ".tmpl"
			if lib.Exists(fn) {
				s, err := ioutil.ReadFile(fn)
				if err == nil {
					hdlr.errorsTemplate[vv] = string(s)
				} else {
					fmt.Printf("Error: Unable to read template file %s\n", err)
					return mid.ErrInvalidConfiguration
				}
				break
			}
		}
	}
	return
}

var _ mid.GoFTLMiddleWare = (*ErrorTemplateType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type ErrorTemplateType struct {
	Next           http.Handler
	Paths          []string
	Errors         []string
	TemplatePath   []string
	LineNo         int
	gCfg           *cfg.ServerGlobalConfigType //
	errorsTemplate map[string]string
}

func NewErrorTemplateServer(n http.Handler, p []string, e []string) *ErrorTemplateType {

	hh := &ErrorTemplateType{Next: n, Paths: p, errorsTemplate: make(map[string]string), Errors: e, TemplatePath: []string{"./errorTemplates"}}

	for _, vv := range hh.Errors {
		for _, ww := range hh.TemplatePath {
			fn := ww + "/" + vv + ".tmpl"
			if lib.Exists(fn) {
				s, err := ioutil.ReadFile(fn)
				if err == nil {
					hh.errorsTemplate[vv] = string(s)
				} else {
					fmt.Printf("Error: Unable to read template file %s\n", err)
					os.Exit(1)
				}
				break
			}
		}
	}

	return hh
}

func (hdlr *ErrorTemplateType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "ErrorTemplate", hdlr.Paths, pn, req.URL.Path)

			hdlr.Next.ServeHTTP(www, req)
			if rw.StatusCode == 0 {
				godebug.Printf(db1, "AT: %s\n", godebug.LF())
				rw.StatusCode = 200
			} else if rw.StatusCode != 200 {
				godebug.Printf(db1, "AT: %s\n", godebug.LF())
				StatusCode := fmt.Sprintf("%d", rw.StatusCode)
				if tt, ok := hdlr.errorsTemplate[StatusCode]; ok {
					godebug.Printf(db1, "AT: %s\n", godebug.LF())
					// data := make(map[string]string)
					// data["StatusCode"] = StatusCode
					// data["StatusText"] = http.StatusText(rw.StatusCode) // convert status to name
					s := tmplp.TemplateProcess(tt, rw, req, nil)
					fmt.Fprintf(rw, "%s", s)
					godebug.Printf(db1, "Template: %s, %s\n", s, godebug.LF())
				}
			}
		}
		return
	}
	hdlr.Next.ServeHTTP(www, req)

}

const db1 = false

/* vim: set noai ts=4 sw=4: */
