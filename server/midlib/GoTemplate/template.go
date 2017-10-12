//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1244
//

//
// GO Templates, Combine data with .tmpl template.
//

package GoTemplate

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/template"

	"www.2c-why.com/JsonX"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/Go-FTL/server/tmplp"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------
//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*GoTemplateType)
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
//	postInit := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
//
//		hh, ok := h.(*GoTemplateType)
//		if !ok {
//			logrus.Errorf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo)
//			fmt.Printf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo)
//			return mid.ErrInternalError
//		} else {
//			// xyzzy - check hh.TemplateRoot exists
//			n := 0
//			m := 0
//			for ii, vv := range hh.TemplateRoot {
//				if !lib.ExistsIsDir(vv) {
//					logrus.Warnf("Warning: at %d path %s is not a directory - will not have any templates, Line No: %d", ii, vv, hh.LineNo)
//				} else {
//					n++
//					fList := sizlib.FilesMatchingPattern(vv, ".tmpl")
//					if len(fList) > 0 {
//						m++
//					}
//				}
//			}
//			if n == 0 {
//				logrus.Errorf("Warning: did not find any directory with templates, Line No: %d", hh.LineNo)
//				fmt.Printf("Warning: did not find any directory with templates, Line No: %d\n", hh.LineNo)
//				fmt.Fprintf(os.Stderr, "%sWarning: did not find any directory with templates, Line No: %d%s\n", MiscLib.ColorRed, hh.LineNo, MiscLib.ColorReset)
//				return mid.ErrInternalError
//			}
//			if m == 0 {
//				logrus.Warnf("Warning: did not find any matching files in any directories, Line No: %d", hh.LineNo)
//				fmt.Printf("Warning: did not find any matching files in any directories, Line No: %d\n", hh.LineNo)
//				fmt.Fprintf(os.Stderr, "%sWarning: did not find any matching files in any directories, Line No: %d%s\n", MiscLib.ColorRed, hh.LineNo, MiscLib.ColorReset)
//			}
//		}
//
//		return nil
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} { return &GoTemplateType{} }
//
//	// /api/tmpl/showRpt.tmpl -> fetch data inside template?
//	// /api/tmpl/showRpt.tmpl?data=bob (data in row/table data)
//	cfg.RegInitItem2("GoTemplate", initNext, createEmptyType, postInit, `{
//		"Paths":                 { "type":[ "string", "filepath" ], "isarray":true, "default":"/" },
//		"TemplateParamName":     { "type":[ "string" ], "default":"template_name" },
//		"TemplateName":          { "type":[ "string" ], "default":"" },
//		"TemplateLibraryName":   { "type":[ "string" ], "isarray":true, "default":"" },
//		"TemplateRoot":          { "type":[ "string" ], "isarray":true, "default":"" },
//		"Root":                  { "type":[ "string" ], "isarray":true, "default":"" },
//		"LineNo":                { "type":[ "int" ], "default":"1" }
//		}`)
//}
//
//// normally identical
//func (hdlr *GoTemplateType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &GoTemplateType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("GoTemplate", CreateEmpty, `{
		"Paths":                 { "type":[ "string", "filepath" ], "isarray":true, "default":"/" },
		"TemplateParamName":     { "type":[ "string" ], "default":"template_name" },
		"TemplateName":          { "type":[ "string" ], "default":"" },
		"TemplateLibraryName":   { "type":[ "string" ], "isarray":true, "default":"" },
		"TemplateRoot":          { "type":[ "string" ], "isarray":true, "default":"" },
		"Root":                  { "type":[ "string" ], "isarray":true, "default":"" },
		"LineNo":                { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *GoTemplateType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *GoTemplateType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	// xyzzy - check hdlr.TemplateRoot exists
	n := 0
	m := 0
	for ii, vv := range hdlr.TemplateRoot {
		if !lib.ExistsIsDir(vv) {
			logrus.Warnf("Warning: at %d path %s is not a directory - will not have any templates, Line No: %d", ii, vv, hdlr.LineNo)
		} else {
			n++
			fList := sizlib.FilesMatchingPattern(vv, ".tmpl")
			if len(fList) > 0 {
				m++
			}
		}
	}
	if n == 0 {
		logrus.Errorf("Warning: did not find any directory with templates, Line No: %d", hdlr.LineNo)
		fmt.Printf("Warning: did not find any directory with templates, Line No: %d\n", hdlr.LineNo)
		fmt.Fprintf(os.Stderr, "%sWarning: did not find any directory with templates, Line No: %d%s\n", MiscLib.ColorRed, hdlr.LineNo, MiscLib.ColorReset)
		return mid.ErrInternalError
	}
	if m == 0 {
		logrus.Warnf("Warning: did not find any matching files in any directories, Line No: %d", hdlr.LineNo)
		fmt.Printf("Warning: did not find any matching files in any directories, Line No: %d\n", hdlr.LineNo)
		fmt.Fprintf(os.Stderr, "%sWarning: did not find any matching files in any directories, Line No: %d%s\n", MiscLib.ColorRed, hdlr.LineNo, MiscLib.ColorReset)
	}
	return
}

var _ mid.GoFTLMiddleWare = (*GoTemplateType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type GoTemplateType struct {
	Next                http.Handler //
	Paths               []string     // Paths that match this
	TemplateParamName   string       // The name of the parameter that specifies the top level template for rendering this.
	TemplateName        string       // Default file name of template (also template name) if not specified on URL
	TemplateRoot        []string     // Set of directories to search for template in
	TemplateLibraryName []string     // set of files that will be read as libraries of templates for processing with this page -- xyzzy - TODO
	Root                []string     // lookup of template in Root if Root != ""
	LineNo              int          //
}

// Parameterized for testing? or just change the test
func NewGoTemplateServer(n http.Handler, p []string, tn string, tln []string) *GoTemplateType {
	return &GoTemplateType{Next: n, Paths: p, TemplateName: tn, TemplateLibraryName: tln}
}

func (hdlr *GoTemplateType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "GoTemplate", hdlr.Paths, pn, req.URL.Path)

			hdlr.Next.ServeHTTP(rw, req)

			var tmpl string
			if hdlr.TemplateParamName != "" {
				tmpl = rw.Ps.ByNameDflt(hdlr.TemplateParamName, hdlr.TemplateName)
			} else {
				tmpl = hdlr.TemplateName
			}

			if db22 {
				fmt.Printf("AT: %s, temlateFileName=%s\n", godebug.LF(), tmpl)
			}

			fileRoot := hdlr.Root

			templateRoot := hdlr.TemplateRoot
			if len(templateRoot) == 0 {
				templateRoot = fileRoot
			}

			if db22 {
				fmt.Printf("templateRoot = %s tmpl=%s, AT: %s\n", templateRoot, tmpl, godebug.LF())
			}

			templateFileName, ok := lib.SearchForFile(templateRoot, tmpl) // Search root for the template in the Root - use 1st one found
			if !ok {
				if db22 {
					fmt.Printf("AT: %s\n", godebug.LF())
				}
				logrus.Errorf("Template file %s was not found, URI=%s, retuing data in JSON, Line No:%d", tmpl, req.RequestURI, hdlr.LineNo)
				return
			}

			if templateFileName != "" {

				if db22 {
					fmt.Printf("AT: %s\n", godebug.LF())
				}
				// fmt.Printf("At: %s - template found, path >%s<\n", lib.LF(), a)
				tfn := make([]string, 0, len(hdlr.TemplateLibraryName)+1)
				tfn = append(tfn, hdlr.TemplateLibraryName...)
				tfn = append(tfn, templateFileName)

				if db23 {
					wd, _ := os.Getwd()
					fmt.Printf("pwd = %s tfn = %v, \n", tfn, wd)
				}

				funcMap := template.FuncMap{
					"json":      lib.SVarI, // Convert data to JSON format to put into JS variable
					"sqlEncode": sqlEncode, // Encode data for use in SQL with ' converted to ''
					"jsEsc":     jsEsc,     // Escape strings for use in JS - with ' converted to \'
					"jsEscDbl":  jsEscDbl,  // Escape strings for use in JS - with " converted to \"
				}

				compiledTemplate, err := template.New("file-template").Funcs(funcMap).ParseFiles(tfn...)
				if err != nil {
					www.WriteHeader(http.StatusInternalServerError)
					logrus.Errorf("Template parse error Error: %s, Line No:%d", err, rw.TemplateLineNo)
					return
				}

				if db22 {
					fmt.Printf("AT: %s\n", godebug.LF())
				}
				if rw.StatusCode == 200 || rw.StatusCode == 0 {

					if db22 {
						fmt.Printf("AT: %s\n", godebug.LF())
					}
					data := make(map[string]interface{})
					req_data := tmplp.GenDataFromReq(rw, req)
					data["requet"] = req_data
					meta := make(map[string]interface{})
					data["meta"] = meta
					if rw.State == goftlmux.RowBuffer {
						meta["rows"] = 1
						data["data"] = rw.Row
					} else if rw.State == goftlmux.TableBuffer {
						meta["rows"] = len(rw.Table)
						data["data"] = rw.Table
					} else {
						// xyzzy - JSON to data
						fmt.Printf("Error - json to data conversion ! AT: %s\n", godebug.LF())
						return
					}

					if db22 {
						fmt.Printf("data=%s, URI:%s, %s\n", lib.SVarI(data), req.RequestURI, godebug.LF())
					}

					// Check if template is defined, else chek for by name for single template
					definedTmpl := compiledTemplate.DefinedTemplates()

					ranOne := false

					if strings.Index(definedTmpl, "header") > 0 {
						// header, body, footer
						err = compiledTemplate.ExecuteTemplate(rw, "header", data)
						if err != nil {
							www.WriteHeader(http.StatusInternalServerError)
							return
						}
						ranOne = true
					}

					if strings.Index(definedTmpl, "body") > 0 {
						err = compiledTemplate.ExecuteTemplate(rw, "body", data)
						if err != nil {
							www.WriteHeader(http.StatusInternalServerError)
							return
						}
						ranOne = true
					} else if strings.Index(definedTmpl, hdlr.TemplateName) > 0 {
						err = compiledTemplate.ExecuteTemplate(rw, hdlr.TemplateName, data)
						if err != nil {
							www.WriteHeader(http.StatusInternalServerError)
							return
						}
						ranOne = true
					}

					if strings.Index(definedTmpl, "footer") > 0 {
						err = compiledTemplate.ExecuteTemplate(rw, "footer", data)
						if err != nil {
							www.WriteHeader(http.StatusInternalServerError)
							return
						}
						ranOne = true
					}

					if !ranOne {
						www.WriteHeader(http.StatusInternalServerError)
						logrus.Errorf("Template: Request[%s] did not find a template to run, looked for 'header', 'footer', 'body', '%s' in %s: Line No:%d", req.RequestURI, hdlr.TemplateName, definedTmpl, rw.TemplateLineNo)
						return
					}

					if db22 {
						fmt.Printf("AT: %s\n", godebug.LF())
					}
					rw.State = goftlmux.ByteBuffer

					if db22 {
						fmt.Printf("Should be written to buffer at this point\n")
					}

					// www.Header().Set("Content-Type", "text/html")

					ct := tmplp.ExecuteATemplate("content_type", req_data)
					www.Header().Set("Content-Type", ct)

					www.WriteHeader(http.StatusOK)

					/*
						s := rw.GetBody()
						// data := make(map[string]string) // Xyzzy template name, template libraries, data
						// parametric data?
						rv := lib.ExecuteATemplate(string(s), data)
						rw.ReplaceBody([]byte(rv))
					*/
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

func sqlEncode(s string) (rv string) {
	rv = strings.Replace(s, "'", "''", -1)
	return
}

func jsEsc(s string) (rv string) {
	fmt.Printf("s=%s\n", s)
	rv = strings.Replace(s, "'", `\'`, -1)
	return
}
func jsEscDbl(s string) (rv string) {
	rv = strings.Replace(s, `"`, `\"`, -1)
	return
}

const db22 = true
const db23 = true

/* vim: set noai ts=4 sw=4: */
