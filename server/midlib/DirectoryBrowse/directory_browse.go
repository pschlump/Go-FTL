//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2018
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1232
//

//
// Package dumpit directory browsing.   The results of browsing to a direcotry can be fead through a Go template.
//

package DirectoryBrowse

import (
	"fmt"
	"net/http"

	JsonX "github.com/pschlump/JSONx"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
)

// --------------------------------------------------------------------------------------------------------------------------

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &DirectoryBrowseType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("DirectoryBrowse", CreateEmpty, `{
        "Paths":            { "type":[ "string", "filepath" ], "isarray":true, "required":true },
        "TemplateName":     { "type":[ "string" ], "default":"" },
        "TemplateRoot":     { "type":[ "[]string","filepath" ], "default":"" },
        "LineNo":           { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *DirectoryBrowseType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *DirectoryBrowseType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*DirectoryBrowseType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type DirectoryBrowseType struct {
	Next         http.Handler //
	Paths        []string     // thins directory browsing is enabled for -- Paths that are served with a directory index
	TemplateName string       // Template Name
	TemplateRoot []string     // Set of Root Directories for templates
	LineNo       int
}

// IgnoreDirectories []string     //

func NewDirectoryBrowseServer(n http.Handler, p []string, m string, r []string) *DirectoryBrowseType {
	return &DirectoryBrowseType{Next: n, Paths: p, TemplateName: m, TemplateRoot: r}
}

/*

../../fileserve/fs.go Line:220 func dirList(w http.ResponseWriter, f File) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

Directory Browse:		1 day
	1. Make template a library that allows setting of content type and body
		t, err := template.New("foo").Parse(`{{define "T"}}Hello, {{.}}!{{end}}`)
		err = t.ExecuteTemplate(out, "T", "<script>alert('you have been pwned')</script>")

			{{define "content_type"}}text/html; charset=en-us{{end}}"
			{{define "page"}}<html>...</html>{{end}}"

	2. Add in other info about file (length, mod date etc)
	3. read-if-modified libary
	4. pull template from cached in memory read-if-modified
	5. Process it - inside ../../fileserver/fs - Line:220 - func dirList(w http.ResponseWriter, f File) {
	6. Have default - for templates
	7. Have config to turn off directory browsing

1. Verify that data gets set
2. Wite the direcotry-no-middleware and test the same

*/

func (hdlr DirectoryBrowseType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "DirectoryBrowse", hdlr.Paths, pn, req.URL.Path)

			// fileRoot := rw.Root	// xyzzy - should be "root" for serving user
			fileRoot := []string{} // xyzzy - should be "root" for serving user

			// Search for the tempalte in directories -  How do I want this to happen?
			tmpl := "index.tmpl" // Default file name for template
			if hdlr.TemplateName != "" {
				tmpl = hdlr.TemplateName // Specified template name
			}

			templateRoot := hdlr.TemplateRoot
			if len(templateRoot) == 0 {
				templateRoot = fileRoot
			}

			templateFileName, ok := lib.SearchForFile(templateRoot, tmpl) // Search root for the template in the Root - use 1st one found
			if !ok {
				www.WriteHeader(http.StatusNotFound)
				logrus.Warn(fmt.Sprintf("Missing template file %s, Line No:%d", tmpl, hdlr.LineNo))
				return
			}

			rw.DirTemplateFileName = templateFileName
			rw.TemplateLineNo = hdlr.LineNo
		}
	}
	hdlr.Next.ServeHTTP(www, req)

}

/* vim: set noai ts=4 sw=4: */
