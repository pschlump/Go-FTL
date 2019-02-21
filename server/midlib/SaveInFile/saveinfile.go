//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2017
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1291
//

//
// A echo-like call, /api/saveinfile usually
//

package SaveInFile

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	template "text/template"
	"time"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	JsonX "github.com/pschlump/JSONx"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/filelib"
	"github.com/pschlump/godebug"
	ms "github.com/pschlump/templatestrings"
)

// --------------------------------------------------------------------------------------------------------------------------
// TODO: add in a "key" that is required. ----- See CorpRegV01 - has key
// 2. Add in a check for the template file existing!
// --------------------------------------------------------------------------------------------------------------------------

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &SaveInFileType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("SaveInFile", CreateEmpty, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"Dir":           { "type":["string"], "default":"./output" },
		"Template":      { "type":["string"], "default":"./tmpl/default.tmpl" },
		"Fmt":           { "type":["string"], "default":"JSON" },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *SaveInFileType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *SaveInFileType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	// If callNo == 0, then this is a 1st call -- it will count up.
	// fmt.Fprintf(os.Stderr, "%sSaveInFile: %d%s\n", MiscLib.ColorCyan, callNo, MiscLib.ColorReset)
	return
}

var _ mid.GoFTLMiddleWare = (*SaveInFileType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type SaveInFileType struct {
	Next     http.Handler //
	Paths    []string     //
	Dir      string       //
	Fmt      string       //
	Template string       // "" indicates just generate output
	seq      int          //
}

func NewSaveInFileServer(n http.Handler, p []string, fmt string) *SaveInFileType {
	return &SaveInFileType{Next: n, Paths: p, seq: 0}
}

func (hdlr *SaveInFileType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "SaveInFile", hdlr.Paths, pn, req.URL.Path)

			ps := &rw.Ps
			if !filelib.Exists(hdlr.Template) {
				fmt.Fprintf(os.Stdout, "%sSaveInFile: missing template file:%s%s\n", MiscLib.ColorRed, hdlr.Template, MiscLib.ColorReset)
				if db1 {
					fmt.Fprintf(os.Stderr, "%sSaveInFile: missing template file:%s%s\n", MiscLib.ColorRed, hdlr.Template, MiscLib.ColorReset)
				}
				www.WriteHeader(http.StatusInternalServerError)
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

			fmt.Fprintf(os.Stderr, "%sSuccess!, AT: %s%s\n", MiscLib.ColorCyan, godebug.LF(), MiscLib.ColorReset)
			type Resp struct {
				Status   string `json:"status"`
				Filename string `json:"FileName"`
			}
			rr := Resp{Status: "success", Filename: TemplateFn}
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
