//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2017
//

package tmplp

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/pschlump/Go-FTL/server/goftlmux" //
	ms "github.com/pschlump/templatestrings"     //
)

//	Modifed from: "encoding/json"

func GenDataFromReq(rw *goftlmux.MidBuffer, req *http.Request) (data map[string]string) {
	data = make(map[string]string)
	// process/pick from req what to allow for template stubstitution
	data["IP"] = GetRemoteIP(req)
	data["URI"] = req.RequestURI
	data["delta_t"] = time.Since(rw.StartTime).String() // Very slow - should bind this into a function iinstead
	data["host"] = req.Host
	data["ERROR"] = fmt.Sprintf("%s", rw.Error)
	data["method"] = req.Method
	data["now"] = time.Now().Format(time.RFC3339Nano) // Very slow - should bind this into a function iinstead
	data["path"] = req.URL.Path
	data["port"] = GetPort(req)
	data["query"] = req.URL.RawQuery
	data["scheme"] = func() (rv string) {
		rv = "http"
		if req.TLS != nil {
			rv = "https"
		}
		return
	}()
	data["start_time"] = rw.StartTime.Format(time.RFC3339Nano) // Very slow - should bind this into a function instead
	data["status_code"] = strconv.Itoa(rw.StatusCode)
	data["StatusCode"] = data["status_code"]
	data["StatusText"] = http.StatusText(rw.StatusCode) // convert status to name
	// xyzzy - add config
	return
}

func TemplateProcess(tmpl string, rw *goftlmux.MidBuffer, req *http.Request, data map[string]string) (rv string) {
	if data == nil {
		data = make(map[string]string)
	}
	// process/pick from req what to allow for template stubstitution
	data["IP"] = GetRemoteIP(req)
	data["URI"] = req.RequestURI
	data["delta_t"] = time.Since(rw.StartTime).String() // Very slow - should bind this into a function iinstead
	data["host"] = req.Host
	data["ERROR"] = fmt.Sprintf("%s", rw.Error)
	data["method"] = req.Method
	data["now"] = time.Now().Format(time.RFC3339Nano) // Very slow - should bind this into a function iinstead
	data["path"] = req.URL.Path
	data["port"] = GetPort(req)
	data["query"] = req.URL.RawQuery
	data["scheme"] = func() (rv string) {
		rv = "http"
		if req.TLS != nil {
			rv = "https"
		}
		return
	}()
	data["start_time"] = rw.StartTime.Format(time.RFC3339Nano) // Very slow - should bind this into a function instead
	data["status_code"] = strconv.Itoa(rw.StatusCode)
	data["StatusCode"] = data["status_code"]
	data["StatusText"] = http.StatusText(rw.StatusCode) // convert status to name
	// xyzzy - add config
	rv = ExecuteATemplate(tmpl, data)
	return
}

func ExecuteATemplate(tmpl string, data map[string]string) (rv string) {
	funcMapTmpl := template.FuncMap{
		"PadR":        ms.PadOnRight,
		"PadL":        ms.PadOnLeft,
		"PicTime":     ms.PicTime,
		"FTime":       ms.StrFTime,
		"PicFloat":    ms.PicFloat,
		"nvl":         ms.Nvl,
		"Concat":      ms.Concat,
		"title":       strings.Title, // The name "title" is what the function will be called in the template text.
		"ifDef":       ms.IfDef,
		"ifIsDef":     ms.IfIsDef,
		"ifIsNotNull": ms.IfIsNotNull,
		"dirname":     filepath.Dir, // xyzzyTemplateAdd - basename, dirname,
		"basename":    filepath.Base,
		"Title":       strings.Title,
		"TmpFileName": ms.GenTmpFileName("./", "", func(fn string) {
			data["TmpFileName"] = fn
		}),
	}
	t := template.New("line-template").Funcs(funcMapTmpl)
	t, err := t.Parse(tmpl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error(102): Invalid template: %s\n", err)
		return tmpl
	}

	// Create an io.Writer to write to a string
	var b bytes.Buffer
	foo := bufio.NewWriter(&b)
	err = t.ExecuteTemplate(foo, "line-template", data)
	// err = t.ExecuteTemplate(os.Stdout, "line-template", data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error(103): Invalid template processing: %s\n", err)
		return tmpl
	}
	foo.Flush()
	rv = b.String() // Fetch the data back from the buffer
	return
}

func ExecuteATemplateByName(tmpl, tmplName string, data map[string]string) (rv string) {
	funcMapTmpl := template.FuncMap{
		"PadR":        ms.PadOnRight,
		"PadL":        ms.PadOnLeft,
		"PicTime":     ms.PicTime,
		"FTime":       ms.StrFTime,
		"PicFloat":    ms.PicFloat,
		"nvl":         ms.Nvl,
		"Concat":      ms.Concat,
		"title":       strings.Title, // The name "title" is what the function will be called in the template text.
		"ifDef":       ms.IfDef,
		"ifIsDef":     ms.IfIsDef,
		"ifIsNotNull": ms.IfIsNotNull,
		"dirname":     filepath.Dir, // xyzzyTemplateAdd - basename, dirname,
		"basename":    filepath.Base,
		"TmpFileName": ms.GenTmpFileName("./", "", func(fn string) {
			data["TmpFileName"] = fn
		}),
	}
	t := template.New("line-template").Funcs(funcMapTmpl)
	t, err := t.Parse(tmpl)
	// fmt.Printf("---->>>>%s<<<<====, %s\n", tmpl, godebug.LF())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error(100): Invalid template: %s\n", err)
		return tmpl
	}

	// Create an io.Writer to write to a string
	var b bytes.Buffer
	foo := bufio.NewWriter(&b)
	err = t.ExecuteTemplate(foo, tmplName, data)
	// err = t.ExecuteTemplate(os.Stdout, "line-template", data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error(101): Invalid template processing: %s\n", err)
		return tmpl
	}
	foo.Flush()
	rv = b.String() // Fetch the data back from the buffer
	return
}

func GetRemoteIP(req *http.Request) string {
	if xForwardedHdr := req.Header.Get("X-Forwarded-For"); xForwardedHdr != "" {
		return xForwardedHdr
	}
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return host
}

func GetPort(req *http.Request) (rv string) {
	_, port, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return
	}
	rv = port
	return
}
