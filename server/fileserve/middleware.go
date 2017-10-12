//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1011
//

package fileserve

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	JsonX "github.com/pschlump/JSONx"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical (not in this case)
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*FileServerType)
//		if ok {
//			pCfg.SetNext(next)
//			rv = pCfg
//
//			// Create the configuration for processing file extensions and handling themes/users.
//			// This has some defaults -
//			// If no default index.html is stupplied then index.html will be used.
//			gob := NewFSConfig("Go-FTL-File-Server")
//			if len(pCfg.IndexFileList) > 0 {
//				if dbD {
//					fmt.Printf("*** init time - found list of indexe files, %s ***, %s\n", pCfg.IndexFileList, godebug.LF())
//				}
//				for _, vv := range pCfg.IndexFileList {
//					if strings.HasPrefix(vv, "/") {
//						gob.IndexHTML(vv)
//					} else {
//						gob.IndexHTML("/" + vv)
//					}
//				}
//			} else {
//				if dbD {
//					fmt.Printf("*** init time - no default list, %s ***, %s\n", pCfg.IndexFileList, godebug.LF())
//				}
//				gob.IndexHTML("/index.html")
//			}
//			for _, rr := range pCfg.Root {
//				gob.AddPreRule([]PreRuleType{
//					{
//						IfMatch:       "/",        // Process things like .ts -> .js + .map.js files
//						UseRoot:       rr,         //
//						StatusOnMatch: PreSuccess, //
//						Fx:            UrlFileExt, //
//					},
//					{
//						IfMatch:       "/",                // user/theme search for files
//						UseRoot:       rr,                 //
//						StatusOnMatch: PreSuccess,         //
//						Fx:            ResolveFnThemeUser, //
//					},
//					{
//						IfMatch:       "",         // Just return files when found, root and index.yyy processing
//						UseRoot:       rr,         //
//						StatusOnMatch: PreSuccess, //
//					},
//				})
//			}
//			pCfg.Cfg = gob
//
//		} else {
//			err = mid.FtlConfigError
//		}
//		return
//	}
//
//	// ExtensionMapConfig.json - with config file - pulled in by PostInit
//
//	// func UrlFileExt(fcfg *FileServerType, www http.ResponseWriter, req *http.Request, urlIn string, g *FSConfig, rulNo int) (urlOut string, rootOut string, stat RuleStatus, err error) {
//
//	// normally identical
//	createEmptyType := func() interface{} { return &FileServerType{} }
//
//	// CommandLocationMap = {
//	// ExtProcessTable = [
//	postInit := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
//
//		hh, ok := h.(*FileServerType)
//		if !ok {
//			fmt.Printf("Error: Wrong data type passed to FileServeType - postInit\n")
//			return mid.ErrInternalError
//		} else {
//
//			for ii, vv := range hh.ExtProcessTable {
//				if mm, ok := InternalFuncLookup[vv.InternalFuncName]; ok {
//					vv.InternalFunc = mm.InternalFunc
//				} else {
//					fmt.Printf("Error: Failed to lookup %s as an internal function\n", vv.InternalFuncName) // error - did not lookup
//					return mid.ErrInternalError
//				}
//				hh.ExtProcessTable[ii] = vv
//			}
//
//			if hh.CommandLocationMap == nil || len(hh.CommandLocationMap) == 0 {
//				hh.CommandLocationMap = CommandLocationMap
//			}
//
//			for ii, vv := range hh.CommandLocationMap {
//				fmt.Printf("Checking %s %s for executable\n", ii, vv)
//				if !lib.ExistsIsExecutable(vv) {
//					fmt.Fprintf(os.Stderr, "%sWarning: file_server - %s => %s is not an executable on disk. %s\n%s", MiscLib.ColorRed, ii, vv, godebug.LF(), MiscLib.ColorReset)
//					logrus.Warnf("Initialize file_server - %s => %s is not an edexecutable on disk. %s", ii, vv, godebug.LF())
//					delete(hh.CommandLocationMap, ii)
//				}
//			}
//
//			fmt.Printf("Config: %s, %s\n", lib.SVarI(hh), godebug.LF())
//
//			// xyzzy - check that Root directories exist
//			n := 0
//			m := 0
//			for ii, vv := range hh.Root {
//				if !lib.ExistsIsDir(vv) {
//					logrus.Warnf("Warning: at %d path %s is not a directory - will not have any files, Line No: %d", ii, vv, hh.LineNo)
//				} else {
//					n++
//					fList := sizlib.FilesMatchingPattern(vv, ".html|.htm|.js|.jpg|.jpeg|.gif|.css|.png")
//					if len(fList) > 0 {
//						m++
//					}
//				}
//			}
//			if n == 0 {
//				logrus.Errorf("Warning: did not find any directory with files, may indicate empty directory. Line No: %d", hh.LineNo)
//				fmt.Printf("Warning: did not find any directory with files, may indicate empty directory. Line No: %d\n", hh.LineNo)
//				fmt.Fprintf(os.Stderr, "%sWarning: did not find any directory with files, may indicate empty directory. Line No: %d%s\n", MiscLib.ColorRed, hh.LineNo, MiscLib.ColorReset)
//				return mid.ErrInternalError
//			}
//			if m == 0 {
//				logrus.Warnf("Warning: did not find any matching files in any directories, may indicate empty directory. Line No: %d", hh.LineNo)
//				fmt.Printf("Warning: did not find any matching files in any directories, may indicate empty directory. Line No: %d\n", hh.LineNo)
//				fmt.Fprintf(os.Stderr, "%sWarning: did not find any matching files in any directories, may indicate empty directory. Line No: %d%s\n", MiscLib.ColorRed, hh.LineNo, MiscLib.ColorReset)
//			}
//
//		}
//
//		return nil
//	}
//
//	cfg.RegInitItem2("file_server", initNext, createEmptyType, postInit, `{
//		}`)
//}
//
//// normally identical
//func (hdlr *FileServerType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &FileServerType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("FileServer", CreateEmpty, `{
		"Paths":                         { "type":[ "string", "filepath" ], "isarray":true, "required":true },
		"Root":                          { "type":[ "string", "filepath" ], "isarray":true },
		"StripPrefix":                   { "type":[ "string", "filepath" ] },
		"IndexFileList":                 { "type":[ "string" ], "default":"index.html", "isarray":true },
		"ThemeRoot":                     { "type":[ "string" ], "default":"./theme/" },
	    "ThemeCookieName":               { "type":[ "string" ], "default":"theme" },
	    "UserCookieName":                { "type":[ "string" ], "default":"username" },
		"CommandLocationMap":            { "type":[ "struct" ] },
		"ExtProcessTable":               { "type":[ "struct" ] },
		"SampleNumber":                  { "type":[ "int" ], "default":"20" },
		"SamplePattern":                 { "type":[ "string" ], "default":"*.js|*.html|*.css|*.less|*.ts|*.scss|*.xml|*.txt|*.gif|*.jpg|*.png|*.svg" },
		"LineNo":                        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *FileServerType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init

	hdlr.Next = next

	// Create the configuration for processing file extensions and handling themes/users.
	// This has some defaults -
	// If no default index.html is stupplied then index.html will be used.
	gob := NewFSConfig("Go-FTL-File-Server")
	if len(hdlr.IndexFileList) > 0 {
		if dbD {
			fmt.Printf("*** init time - found list of indexe files, %s ***, %s\n", hdlr.IndexFileList, godebug.LF())
		}
		for _, vv := range hdlr.IndexFileList {
			if strings.HasPrefix(vv, "/") {
				gob.IndexHTML(vv)
			} else {
				gob.IndexHTML("/" + vv)
			}
		}
	} else {
		if dbD {
			fmt.Printf("*** init time - no default list, %s ***, %s\n", hdlr.IndexFileList, godebug.LF())
		}
		gob.IndexHTML("/index.html")
	}
	for _, rr := range hdlr.Root {
		gob.AddPreRule([]PreRuleType{
			{
				IfMatch:       "/",        // Process things like .ts -> .js + .map.js files
				UseRoot:       rr,         //
				StatusOnMatch: PreSuccess, //
				Fx:            UrlFileExt, //
			},
			{
				IfMatch:       "/",                // user/theme search for files
				UseRoot:       rr,                 //
				StatusOnMatch: PreSuccess,         //
				Fx:            ResolveFnThemeUser, //
			},
			{
				IfMatch:       "",         // Just return files when found, root and index.yyy processing
				UseRoot:       rr,         //
				StatusOnMatch: PreSuccess, //
			},
		})
	}
	hdlr.Cfg = gob

	return
}

func (hdlr *FileServerType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {

	// If callNo == 0, then this is a 1st call -- it will count up.
	// fmt.Fprintf(os.Stderr, "%sFileServer: %d%s\n", MiscLib.ColorCyan, callNo, MiscLib.ColorReset)

	for ii, vv := range hdlr.ExtProcessTable {
		if mm, ok := InternalFuncLookup[vv.InternalFuncName]; ok {
			vv.InternalFunc = mm.InternalFunc
		} else {
			fmt.Printf("Error: Failed to lookup %s as an internal function\n", vv.InternalFuncName) // error - did not lookup
			return mid.ErrInternalError
		}
		hdlr.ExtProcessTable[ii] = vv
	}

	if hdlr.CommandLocationMap == nil || len(hdlr.CommandLocationMap) == 0 {
		hdlr.CommandLocationMap = CommandLocationMap
	}

	for ii, vv := range hdlr.CommandLocationMap {
		fmt.Printf("Checking %s %s for executable\n", ii, vv)
		if !lib.ExistsIsExecutable(vv) {
			fmt.Fprintf(os.Stderr, "%sWarning: file_server - %s => %s is not an executable on disk. %s\n%s", MiscLib.ColorRed, ii, vv, godebug.LF(), MiscLib.ColorReset)
			logrus.Warnf("Initialize file_server - %s => %s is not an edexecutable on disk. %s", ii, vv, godebug.LF())
			delete(hdlr.CommandLocationMap, ii)
		}
	}

	fmt.Printf("Config: %s, %s\n", lib.SVarI(hdlr), godebug.LF())

	// xyzzy - check that Root directories exist
	n := 0
	m := 0
	for ii, vv := range hdlr.Root {
		if !lib.ExistsIsDir(vv) {
			logrus.Warnf("Warning: at %d path %s is not a directory - will not have any files, Line No: %d", ii, vv, hdlr.LineNo)
		} else {
			n++
			fList := sizlib.FilesMatchingPattern(vv, ".html|.htm|.js|.jpg|.jpeg|.gif|.css|.png")
			if len(fList) > 0 {
				m++
			}
		}
	}
	if n == 0 {
		logrus.Errorf("Warning: did not find any directory with files, may indicate empty directory. Line No: %d", hdlr.LineNo)
		fmt.Printf("Warning: did not find any directory with files, may indicate empty directory. Line No: %d\n", hdlr.LineNo)
		fmt.Fprintf(os.Stderr, "%sWarning: did not find any directory with files, may indicate empty directory. Line No: %d%s\n", MiscLib.ColorRed, hdlr.LineNo, MiscLib.ColorReset)
		return mid.ErrInternalError
	}
	if m == 0 {
		logrus.Warnf("Warning: did not find any matching files in any directories, may indicate empty directory. Line No: %d", hdlr.LineNo)
		fmt.Printf("Warning: did not find any matching files in any directories, may indicate empty directory. Line No: %d\n", hdlr.LineNo)
		fmt.Fprintf(os.Stderr, "%sWarning: did not find any matching files in any directories, may indicate empty directory. Line No: %d%s\n", MiscLib.ColorRed, hdlr.LineNo, MiscLib.ColorReset)
	}

	return
}

var _ mid.GoFTLMiddleWare = (*FileServerType)(nil) // compile time validation that this matches with the GoFTLMiddleWare interface

// --------------------------------------------------------------------------------------------------------------------------

type FileServerType struct {
	Next               http.Handler // No Next, this is the bottom of the stack.
	Paths              []string
	Root               []string
	StripPrefix        string
	IndexFileList      []string
	ThemeRoot          string
	ThemeCookieName    string
	UserCookieName     string
	LineNo             int
	Cfg                *FSConfig         // Runtime configuration for servring and finding files.
	SampleNumber       int               // xyzzy - not implemented yet
	SamplePattern      string            // xyzzy - not implemented yet
	CommandLocationMap map[string]string // xyzzyInit - Command set map of commands to full paths
	ExtProcessTable    []*ExtProcessType
}

func NewFileServer(n http.Handler, g *FSConfig, Path []string, IndexFileList []string, Root []string) *FileServerType {
	if Path == nil || len(Path) == 0 {
		Path = []string{"/"}
	}
	if Root == nil || len(Root) == 0 {
		Root = []string{"./www"}
	}
	if IndexFileList == nil || len(IndexFileList) == 0 {
		IndexFileList = []string{"index.html"}
	}
	if g == nil {
		gob := NewFSConfig("Go-FTL-File-Server")
		if len(IndexFileList) > 0 {
			for _, vv := range IndexFileList {
				gob.IndexHTML("/" + vv)
			}
		} else {
			gob.IndexHTML("/index.html")
		}
		for _, rr := range Root {
			gob.AddPreRule([]PreRuleType{
				PreRuleType{
					IfMatch:       "/",
					UseRoot:       rr,
					StatusOnMatch: PreSuccess,
					Fx:            UrlFileExt,
				},
				PreRuleType{
					IfMatch:       "/",
					UseRoot:       rr,
					StatusOnMatch: PreSuccess,
					Fx:            ResolveFnThemeUser,
				},
				PreRuleType{
					IfMatch:       "",
					UseRoot:       rr,
					StatusOnMatch: PreSuccess,
				},
			})
		}
		g = gob
	}
	// xyzzy - CommandLocationMap map[string]string // xyzzyInit - Command set map of commands to full paths
	// xyzzy - ExtProcessTable    []*ExtProcessType
	return &FileServerType{
		Next:            n, // may be NIL!
		Paths:           Path,
		Root:            Root,
		IndexFileList:   IndexFileList,
		ThemeRoot:       "./theme/",
		ThemeCookieName: "theme",
		UserCookieName:  "username",
		Cfg:             g,
		LineNo:          1,
	}
}

/* vim: set noai ts=4 sw=4: */
