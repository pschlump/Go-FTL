//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2018.
//

package CorpRegV01

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	template "text/template"
	"time"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/sizlib"
	JsonX "github.com/pschlump/JSONx"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/filelib"
	"github.com/pschlump/godebug"
	ms "github.com/pschlump/templatestrings"
)

// --------------------------------------------------------------------------------------------------------------------------
// TODO:
// 2. Add in a check for the template file existing!
// 3. Add in a check for Makefile - and a MakefileName for -f ption
// 4. Add in a Directory to run the Make in  (how?) -- run a sub-process with a "cd" in a shell script.
// 5. Change SetInEnv to be a template for each one, with Name={{.value}} for the template - so run template.
//     if have Name= then pull that of (string.Split) and set the name, else just set the name.
// --------------------------------------------------------------------------------------------------------------------------

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &CorpRegV01Type{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("CorpRegV01", CreateEmpty, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"Dir":           { "type":["string"], "default":"./output" },
		"Template":      { "type":["string"], "default":"./tmpl/default.tmpl" },
		"Fmt":           { "type":["string"], "default":"JSON" },
		"MakeTarget":    { "type":["string"], "isarray":true },
		"SetEnvNames":   { "type":["string"], "isarray":true },
		"ApiKey":        { "type":["string"], "default":"Y7Vqi7LHkqOnJcfvHugxjHO7f0"},
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *CorpRegV01Type) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *CorpRegV01Type) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	// If callNo == 0, then this is a 1st call -- it will count up.
	// fmt.Fprintf(os.Stderr, "%sCorpRegV01: %d%s\n", MiscLib.ColorCyan, callNo, MiscLib.ColorReset)
	return
}

var _ mid.GoFTLMiddleWare = (*CorpRegV01Type)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type CorpRegV01Type struct {
	Next        http.Handler //
	Paths       []string     //
	Dir         string       //
	Fmt         string       //
	MakeTarget  []string     //
	SetEnvNames []string     //
	ApiKey      string       //
	Template    string       // "" indicates just generate output
	seq         int          //
}

func NewCorpRegV01Server(n http.Handler, p []string, fmt string) *CorpRegV01Type {
	return &CorpRegV01Type{Next: n, Paths: p, seq: 0}
}

func (hdlr *CorpRegV01Type) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "CorpRegV01", hdlr.Paths, pn, req.URL.Path)

			ps := &rw.Ps
			if !filelib.Exists(hdlr.Template) {
				fmt.Fprintf(os.Stdout, "%sCorpRegV01: missing template file:%s%s\n", MiscLib.ColorRed, hdlr.Template, MiscLib.ColorReset)
				if db1 {
					fmt.Fprintf(os.Stderr, "%sCorpRegV01: missing template file:%s%s\n", MiscLib.ColorRed, hdlr.Template, MiscLib.ColorReset)
				}
				www.WriteHeader(http.StatusInternalServerError)
				return
			}

			// see if this is a valid user
			apiKey := ps.ByNameDflt("ApiKey", "")
			if hdlr.ApiKey != apiKey {
				io.WriteString(www, `{"status":"error","code":"1","msg":"Error(10030): Invalid API Key"}`)
				return
			}

			allParams := ps.DumpParamNVF()
			if db1 {
				pwd, _ := os.Getwd()
				fmt.Fprintf(os.Stderr, "In: %s --->>>%s<<<---, %s\n", pwd, godebug.SVarI(allParams), godebug.LF())
			}

			data := make(map[string]interface{})
			data["seq"] = hdlr.seq
			hdlr.seq++
			t := time.Now()
			ts := t.Format(time.RFC3339)
			data["timestamp"] = ts

			for ii, vv := range allParams { // pull in argumetns
				data[vv.Name] = vv.Value
				data[fmt.Sprintf("_%d_", ii)] = vv.Value
				data[fmt.Sprintf("_%d_name_", ii)] = vv.Name
			}

			if db1 {
				fmt.Fprintf(os.Stderr, "data: --->>>%s<<<---, %s\n", godebug.SVarI(data), godebug.LF())
			}

			tmp := RunTemplate(hdlr.Template, "filename", data)
			TemplateFn := filepath.Clean(hdlr.Dir + filepath.Clean("/"+tmp))
			data["gen_filename"] = TemplateFn
			if db1 {
				fmt.Fprintf(os.Stderr, "TemplateFn: %sOutputFile: --->>>%s<<<---%s hdlr.Dir ->%s<- hdlr.Template ->%s<-, %s\n", MiscLib.ColorYellow, TemplateFn, MiscLib.ColorReset, hdlr.Dir, hdlr.Template, godebug.LF())
			}
			if filelib.Exists(TemplateFn) {
				if db1 {
					fmt.Fprintf(os.Stderr, "%sCan not overwrite existing file name:%s%s\n", MiscLib.ColorRed, TemplateFn, MiscLib.ColorReset)
				}
				www.WriteHeader(http.StatusInternalServerError)
				return
			}
			// 4. open specified file - use template for fn?
			// 5. write to file
			if "JSON" == hdlr.Fmt {
				if db1 {
					fmt.Fprintf(os.Stderr, "JSON: --->>>%s<<<---, %s\n", godebug.SVarI(data), godebug.LF())
				}
				ioutil.WriteFile(TemplateFn, []byte(godebug.SVarI(data)), 0644)
				fmt.Fprintf(os.Stderr, "%sSuccess!, AT: %s%s\n", MiscLib.ColorCyan, godebug.LF(), MiscLib.ColorReset)
			} else {
				fmt.Fprintf(os.Stderr, "%sAT: %s%s\n", MiscLib.ColorCyan, godebug.LF(), MiscLib.ColorReset)
				body := RunTemplate(hdlr.Template, "body", data)
				if db1 {
					fmt.Fprintf(os.Stderr, "Template Body: --->>>%s<<<---, %s\n", body, godebug.LF())
				}
				ioutil.WriteFile(TemplateFn, []byte(body), 0644)
				fmt.Fprintf(os.Stderr, "%sSuccess!, AT: %s%s\n", MiscLib.ColorCyan, godebug.LF(), MiscLib.ColorReset)
			}

			// -----------------------------------------------------------------------------------------------------------------------
			// Exec make at this point
			// -----------------------------------------------------------------------------------------------------------------------
			target := ps.ByName("target")
			if !sizlib.InArray(target, hdlr.MakeTarget) {
				fmt.Printf("Error(14017): Invalid command [%s] - can ony be one of %s\n", target, hdlr.MakeTarget)
				fmt.Fprintf(os.Stderr, "Error(14017): Invalid command [%s] - can ony be one of %s\n", target, hdlr.MakeTarget)
				fmt.Fprintf(os.Stderr, "%sFailed!, AT: %s%s\n", MiscLib.ColorRed, godebug.LF(), MiscLib.ColorReset)
				www.WriteHeader(http.StatusInternalServerError)
				return
			}

			// -----------------------------------------------------------------------------------------------------------------------
			for _, envItem := range hdlr.SetEnvNames {
				if aa, ok1 := data[envItem]; ok1 {
					if ss, ok2 := aa.(string); ok2 {
						os.Setenv(envItem, ss)
					} else {
						ss := fmt.Sprintf("%s", aa)
						os.Setenv(envItem, ss)
					}
				}
			}

			// -----------------------------------------------------------------------------------------------------------------------
			var status string
			out, err := exec.Command("/usr/bin/make", target).Output() // Run the command, get the output.
			if err != nil {                                            // If command running failed, report error go to next row
				status = "error"
				out = []byte(fmt.Sprintf("Make returned a non-0 status, error:%s", err))
				fmt.Printf("Debug(14019): Error: [%s], %s\n", err, godebug.LF())
				fmt.Fprintf(os.Stderr, "Debug(14019): Error: [%s], %s\n", err, godebug.LF())
			} else {
				status = "success"
				fmt.Printf("Debug(14019): Output: [%s], %s\n", out, godebug.LF())
				fmt.Fprintf(os.Stderr, "Debug(14019): Output: [%s], %s\n", out, godebug.LF())
			}
			// -----------------------------------------------------------------------------------------------------------------------

			fmt.Fprintf(os.Stderr, "%sSuccess!, AT: %s%s\n", MiscLib.ColorCyan, godebug.LF(), MiscLib.ColorReset)
			type Resp struct {
				Status   string `json:"status"`
				Filename string `json:"FileName"`
				Msg      string `json:"Msg"`
			}
			rr := Resp{Status: status, Filename: TemplateFn, Msg: fmt.Sprintf("%s", out)}
			www.Header().Set("Content-Type", "application/json")
			s := fmt.Sprintf("%s\n", lib.SVarI(rr))
			io.WriteString(www, s)
			www.WriteHeader(http.StatusOK)

			return
		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
		}
	}
	hdlr.Next.ServeHTTP(www, req)
}

// ===================================================================================================================================================
// Run a template and get the results back as a stirng.
// Sample - used below.
//func ExecuteATemplate(tmpl string, data map[string]interface{}) string {
//	t := template.New("line-template")
//	t, err := t.Parse(tmpl)
//	if err != nil {
//		fmt.Printf("Error(): Invalid template: %s\n", err)
//		return tmpl
//	}
//
//	// Create an io.Writer to write to a string
//	var b bytes.Buffer
//	foo := bufio.NewWriter(&b)
//	err = t.ExecuteTemplate(foo, "line-template", data)
//	if err != nil {
//		fmt.Printf("Error(): Invalid template processing: %s\n", err)
//		return tmpl
//	}
//	foo.Flush()
//	s := b.String() // Fetch the data back from the buffer
//	return s
//}

// ===================================================================================================================================================
// Run a template and get the results back as a stirng.
// This is the primary template runner for sending email.
func RunTemplate(TemplateFn string, name_of string, g_data map[string]interface{}) string {

	rtFuncMap := template.FuncMap{
		"Center":      ms.CenterStr,    //
		"PadR":        ms.PadOnRight,   //
		"PadL":        ms.PadOnLeft,    //
		"PicTime":     ms.PicTime,      //
		"FTime":       ms.StrFTime,     //
		"PicFloat":    ms.PicFloat,     //
		"nvl":         ms.Nvl,          //
		"Concat":      ms.Concat,       //
		"title":       strings.Title,   // The name "title" is what the function will be called in the template text.
		"toLower":     strings.ToLower, //
		"toUpper":     strings.ToUpper, //
		"ifDef":       ms.IfDef,        //
		"ifIsDef":     ms.IfIsDef,      //
		"ifIsNotNull": ms.IfIsNotNull,  //
	}

	var b bytes.Buffer
	foo := bufio.NewWriter(&b)

	t, err := template.New("simple-tempalte").Funcs(rtFuncMap).ParseFiles(TemplateFn)
	// t, err := template.New("simple-tempalte").ParseFiles(TemplateFn)
	if err != nil {
		fmt.Printf("Error(12004): parsing/reading template, %s\n", err)
		return ""
	}

	err = t.ExecuteTemplate(foo, name_of, g_data)
	if err != nil {
		fmt.Fprintf(foo, "Error(12005): running template=%s, %s\n", name_of, err)
		return ""
	}

	foo.Flush()
	s := b.String() // Fetch the data back from the buffer

	if db1 {
		fmt.Fprintf(os.Stderr, "Template Output is: ----->%s<-----\n", s)
	}

	return s

}

const db1 = true

/* vim: set noai ts=4 sw=4: */
