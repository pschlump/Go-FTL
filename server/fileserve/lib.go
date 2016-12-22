//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1012
//

package fileserve

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/tmplp"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/check-json-syntax/lib"
	"github.com/pschlump/godebug" //
	"github.com/pschlump/json"    //	Modifed from: "encoding/json"
	"github.com/russross/blackfriday"
)

// ============================================================================================================================================
func GetRwHdlrFromWWW_2(www http.ResponseWriter, req *http.Request) (rw *goftlmux.MidBuffer, ok bool) {

	rw, ok = www.(*goftlmux.MidBuffer)
	if !ok {
		//AnError(hdlr, www, req, 500, 5, fmt.Sprintf("hdlr not correct type in rw.!, %s\n", godebug.LF()))
		fmt.Printf("BAD BAD - did not have a rw for w !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!, %s\n", godebug.LF())
		return
	}

	//	hdlr, ok = rw.Hdlr.(*AesSrpType)
	//	if !ok {
	//		AnError(hdlr, www, req, 500, 5, fmt.Sprintf("hdlr not set in rw.!, %s\n", godebug.LF()))
	//		return
	//	}

	return
}

// ============================================================================================================================================

type InternalFuncTableType struct {
	InternalFunc     InternalFuncType `json:"-"`            // if true then if InternalFunc != nil, then call it.
	InternalFuncName string           `json:"InternalFunc"` // string name to lookkup InteralFunc by
}

var InternalFuncLookup map[string]InternalFuncTableType // xyzzyInit - Lookup of functions for ExtProcessType table

type ExtProcessType struct {
	ToExt            []string         `json:"ToExt"`        //	Set of output extensions
	FromExt          string           `json:"FromExt"`      // Input extension
	HasFs            string           `json:"HasFs"`        // File sep
	CommandToRun     string           `json:"CommandToRun"` // typescript {{.f_base}}.ts
	ReRun            bool             `json:"ReRun"`        // if false(default) then check timestamps on input->output, else just run every time		-- xyzzy not implemented yet?
	InternalFunc     InternalFuncType `json:"-"`            // if true then if InternalFunc != nil, then call it.
	InternalFuncName string           `json:"InternalFunc"` // string name to lookkup InteralFunc by
}

var CommandLocationMap map[string]string // xyzzyInit - Command set map of commands to full paths

var ExtProcessTable []*ExtProcessType

//
// Install SCSS
// Install SASS
//		http://sass-lang.com/install
// Install Less
// 		http://lesscss.org/
// Install UglifyJS
// Install TSC
//
// https://responsivedesign.is/articles/difference-between-sass-and-scss
//

func init() {
	CommandLocationMap = make(map[string]string)
	CommandLocationMap["tsc"] = "/usr/local/lib/node_modules/typescript/bin/tsc"
	CommandLocationMap["uglifyjs"] = "/usr/local/bin/uglifyjs"
	CommandLocationMap["css-pack"] = "/Users/corwin/bin/css-pack"
	CommandLocationMap["markdown-cli"] = "/Users/corwin/bin/markdown-cli"
	CommandLocationMap["cp"] = "/bin/cp"
	CommandLocationMap["make"] = "/usr/bin/make"
	CommandLocationMap["sass"] = "/usr/local/bin/sass"
	CommandLocationMap["scss"] = "/usr/local/bin/sass"
	CommandLocationMap["lessc"] = "/usr/local/bin/lessc"
	//	CommandLocationMap["xcat"] = "/usr/local/bin/xcat"

	if db_fileServer {
		fmt.Printf("CommandLocationMap = %s\n\n", lib.SVarI(CommandLocationMap))
	}

	ExtProcessTable = append(ExtProcessTable, &ExtProcessType{
		FromExt:      ".md",
		ToExt:        []string{".html"},
		CommandToRun: `{ "Cmd":"markdown-cli", "Params":[ "-i", "{{.inputFile}}", "-o", "{{.outputFile}}", "-c", "../markdown-cfg.json" ] }`,
	})
	ExtProcessTable = append(ExtProcessTable, &ExtProcessType{
		FromExt:      ".ts",
		ToExt:        []string{".js"},
		CommandToRun: `{ "Cmd":"tsc", "Params":[ "{{.inputFile}}", "--sourceMap" ] }`,
	})
	ExtProcessTable = append(ExtProcessTable, &ExtProcessType{
		FromExt:      ".ts",
		ToExt:        []string{".js.map"},
		CommandToRun: `{ "Cmd":"tsc", "Params":[ "{{.inputFile}}", "--sourceMap" ] }`,
	})
	ExtProcessTable = append(ExtProcessTable, &ExtProcessType{
		FromExt:      ".js",
		ToExt:        []string{".min.js", ".min.map"},
		CommandToRun: `{ "Cmd":"uglifyjs", "Params":[ "--input", "{{.inputFile}}", "--output", "{{.outputFile}}", "--source-map", "{{.base_file_name}}.min.map", "--comments" ] }`,
	})
	ExtProcessTable = append(ExtProcessTable, &ExtProcessType{
		FromExt:      ".css",
		ToExt:        []string{".min.css"},
		CommandToRun: `{ "Cmd":"css-pack", "Params":[ "-i", "{{.inputFile}}", "-o", "{{.outputFile}}" ] }`,
	})
	ExtProcessTable = append(ExtProcessTable, &ExtProcessType{
		FromExt:      ".in",
		ToExt:        []string{".out"},
		CommandToRun: `{ "Cmd":"cp", "Params":[ "{{.inputFile}}", "{{.outputFile}}" ] }`,
	})
	ExtProcessTable = append(ExtProcessTable, &ExtProcessType{
		FromExt:      ".jpg",
		ToExt:        []string{".brotilli"},
		CommandToRun: `{ "Cmd":"make", "Params":[ "{{.outputFile}}" ] }`,
	})
	ExtProcessTable = append(ExtProcessTable, &ExtProcessType{
		FromExt:      ".sass",
		ToExt:        []string{".css"},
		CommandToRun: `{ "Cmd":"sass", "Params":[ "--sourcemap=file", "{{.inputFile}}", "{{.outputFile}}" ] }`,
	})
	ExtProcessTable = append(ExtProcessTable, &ExtProcessType{
		FromExt:      ".scss",
		ToExt:        []string{".css"},
		CommandToRun: `{ "Cmd":"scss", "Params":[ "--scss", "--sourcemap=file", "{{.inputFile}}", "{{.outputFile}}" ] }`,
	})
	ExtProcessTable = append(ExtProcessTable, &ExtProcessType{
		FromExt:      ".less",
		ToExt:        []string{".min.css"},
		CommandToRun: `{ "Cmd":"lessc", "Params":[ "--clean-css", "{{.inputFile}}", "{{.outputFile}}" ] }`,
	})
	ExtProcessTable = append(ExtProcessTable, &ExtProcessType{
		FromExt:      ".markdown",
		ToExt:        []string{".html"},
		InternalFunc: ConvMakrdown,
	})
	//	ExtProcessTable = append(ExtProcessTable, &ExtProcessType{
	//		FromExt:      ".css",
	//		ToExt:        []string{".css"},
	//		HasFs:        "++",
	//		CommandToRun: `{ "Cmd":"xcat", "Params":[ "{{.inputFileList}}", "-o", "{{.outputFile}}" ] }`,
	//	})
	ExtProcessTable = append(ExtProcessTable, &ExtProcessType{
		FromExt:      ".js",
		ToExt:        []string{".js"},
		HasFs:        "++",
		CommandToRun: `{ "Cmd":"xcat", "Params":[ "{{.inputFileList}}", "-o", "{{.outputFile}}" ] }`,
	})
	if db_fileServer {
		fmt.Printf("ExtProcessTable = %s\n\n", lib.SVarI(ExtProcessTable))
	}
}

var ErrInputFileMissing = errors.New("Unable to find file that was avaialbe earlier in processing.")
var ErrOutputFileError = errors.New("Output error")

func ConvMakrdown(in, out string) (err error) {
	input, err := ioutil.ReadFile(in)
	if err != nil {
		return ErrInputFileMissing
	}
	var output []byte
	output = blackfriday.MarkdownCommon(input)
	err = ioutil.WriteFile(out, output, 0644)
	if err != nil {
		return ErrOutputFileError
	}
	return
}

/*
"CommandLocationMapFileNmae":    { "type":[ "string" ], "default":"command_locaiton_map.json" },
"ExtProcessTableFileNmae":       { "type":[ "string" ], "default":"ext_process_table.json" },

CommandLocationMap = {
	"cp": "/bin/cp",
	"css-pack": "/Users/corwin/bin/pack-css",
	"lessc": "/usr/bin/lessc",
	"make": "/usr/bin/make",
	"markdown-cli": "/Users/corwin/bin/markdown-cli",
	"sass": "/usr/bin/sass",
	"scss": "/usr/bin/sass",
	"tsc": "/usr/local/lib/node_modules/typescript/bin/tsc",
	"uglifyjs": "/usr/local/bin/uglifyjs"
}

ExtProcessTable = [
	{
		"ToExt": [
			".html"
		],
		"FromExt": ".md",
		"CommandToRun": "{ \"Cmd\":\"markdown-cli\", \"Params\":[ \"-i\", \"{{.inputFile}}\", \"-o\", \"{{.outputFile}}\", \"-c\", \"../markdown-cfg.json\" ] }",
		"ReRun": false
	},
	{
		"ToExt": [
			".js"
		],
		"FromExt": ".ts",
		"CommandToRun": "{ \"Cmd\":\"tsc\", \"Params\":[ \"{{.inputFile}}\" ] }",
		"ReRun": false
	},
	{
		"ToExt": [
			".min.js",
			".min.map"
		],
		"FromExt": ".js",
		"CommandToRun": "{ \"Cmd\":\"uglifyjs\", \"Params\":[ \"--input\", \"{{.inputFile}}\", \"--output\", \"{{.outputFile}}\", \"--source-map\", \"{{.base_file_name}}.min.map\", \"--comments\" ] }",
		"ReRun": false
	},
	{
		"ToExt": [
			".min.css"
		],
		"FromExt": ".css",
		"CommandToRun": "{ \"Cmd\":\"css-pack\", \"Params\":[ \"-i\", \"{{.inputFile}}\", \"-o\", \"{{.outputFile}}\" ] }",
		"ReRun": false
	},
	{
		"ToExt": [
			".out"
		],
		"FromExt": ".in",
		"CommandToRun": "{ \"Cmd\":\"cp\", \"Params\":[ \"{{.inputFile}}\", \"{{.outputFile}}\" ] }",
		"ReRun": false
	},
	{
		"ToExt": [
			".brotilli"
		],
		"FromExt": ".jpg",
		"CommandToRun": "{ \"Cmd\":\"make\", \"Params\":[ \"{{.outputFile}}\" ] }",
		"ReRun": false
	},
	{
		"ToExt": [
			".css"
		],
		"FromExt": ".sass",
		"CommandToRun": "{ \"Cmd\":\"sass\", \"Params\":[ \"--sourcemap=file\", \"{{.inputFile}}\", \"{{.outputFile}}\" ] }",
		"ReRun": false
	},
	{
		"ToExt": [
			".css"
		],
		"FromExt": ".scss",
		"CommandToRun": "{ \"Cmd\":\"scss\", \"Params\":[ \"--scss\", \"--sourcemap=file\", \"{{.inputFile}}\", \"{{.outputFile}}\" ] }",
		"ReRun": false
	},
	{
		"ToExt": [
			".min.css"
		],
		"FromExt": ".less",
		"CommandToRun": "{ \"Cmd\":\"lessc\", \"Params\":[ \"--clean-css\", \"{{.inputFile}}\", \"{{.outputFile}}\" ] }",
		"ReRun": false
	}
]

*/

// ============================================================================================================================================

// Problems
// 	1. Need a template-for config files - plus a search path for them.  (markdown-cfg.json for example )

// Required Problems
// 	1. Need pipes like |remove ".map.js" as ops in TemplateProcess
//  1. Need a "map" of [command] -> location of command in file system to "ls" -> /bin/ls" and config by user, if not found then do not run.
//	1. Concrurrency - need to limit the set of commands run ton 1 at a time in a single directory.
//		1. Commands not run in a Q
//		2. Latency

// Problems -- Later or solved.
//  *1. Logging and output functions need to have context to log to
//  *0. Template/Cookie thing - xyzzyCookie
//	*0. Parameter substitution - how
//	*0. Need to fix the "suffix" problem of .min.js -> .js
//	*0. Security - this runs outside commands - that is why I have chosen to copile in the commands that can be run. -- Map of runnable commands to take care of this.
//	?2. Limited utillity.  Make runs a topological sort to determine the order to run a set of commands to build something.   Just just looks at a single set of input->output.
//		That is rather limited.   One solution is to run "make" form a shell script to correct this limitation.
//   3. A tool that will walk the output/log and
//      1. report on any errors
//      2. pull out all build commands -- and then re-run them to "rebuild" a site

// Info/Note
// 	func (f *FileServerType) ServeFile(w http.ResponseWriter, r *http.Request, name string) {
//		abc.js->abc.def.js (Makefile)

func getThemeUserRoot(fcfg *FileServerType, www http.ResponseWriter, req *http.Request) (user, theme, t_root string) {
	user = ""
	if uu, err := req.Cookie(fcfg.UserCookieName); err == nil {
		user = uu.Value
	}

	theme = ""
	if uu, err := req.Cookie(fcfg.ThemeCookieName); err == nil {
		theme = uu.Value
	}

	t_root, _ = filepath.Abs(fcfg.ThemeRoot) // From Config

	if db6 || db9 {
		fmt.Printf(">>>>>>>>>>>>>>>>>>>> AT: %s theme=%s user=%s t_root=%s\n", godebug.LF(), theme, user, t_root)
		fmt.Printf(">>>>>>>>>>>>>>>>>>>> Called By: %s\n", godebug.LF(2))
	}

	return
}

func ResolveFnThemeUser(fcfg *FileServerType, www http.ResponseWriter, req *http.Request, urlIn string, g *FSConfig, rulNo int) (urlOut string, rootOut string, stat RuleStatus, err error) {
	if rw, ok := www.(*goftlmux.MidBuffer); ok {
		_ = rw

		if db1 || dbD {
			fmt.Printf("\n\n------------------------ ResolveFnThemeUser ---------------------- AT %s\n", godebug.LF())
		}

		rootOut = g.PreRule[rulNo].UseRoot
		urlOut = urlIn
		stat = PreNext

		user, theme, t_root := getThemeUserRoot(fcfg, www, req)

		chk2 := func(pth ...string) (foundIt bool, urlOut string) {
			if db4 {
				fmt.Printf("AT %s ---- doing chk2 urlIn >%s< pth >%s<\n", godebug.LF(), urlIn, pth)
			}
			xFn := filepath.Join(pth...)
			if db4 || db9 {
				fmt.Printf("AT %s -- xFn >%s<\n", godebug.LF(), xFn)
			}
			if ok, outFi := lib.ExistsGetFileInfo(xFn); ok { // If we have input and output, then we will check to see if need rebuild
				_ = outFi
				if db4 {
					fmt.Printf("()()() Success ()()() t_root=%s xFn=%s AT %s\n", t_root, xFn, godebug.LF())
				}
				stat = PreSuccess
				rootOut = t_root
				// urlOut = xFn
				if len(pth)-1 >= 1 {
					rootOut = filepath.Join(pth[0 : len(pth)-1]...)
				}
				urlOut = urlIn
				foundIt = true
				return
			}
			if db4 {
				fmt.Printf("DID NOT FIND FILE [%s] AT %s\n", xFn, godebug.LF())
			}
			return
		}

		if db4 {
			fmt.Printf("AT %s, root >%s< urlIn >%s<\n", godebug.LF(), t_root, urlIn)
		}
		if user != "" && theme != "" {
			foundIt, aUrl := chk2(t_root, user, theme, urlIn)
			if foundIt && db4 {
				fmt.Printf("AT %s, aUrl=%s\n", godebug.LF(), aUrl)
			}
		} else if theme != "" {
			foundIt, aUrl := chk2(t_root, theme, urlIn)
			if foundIt && db4 {
				fmt.Printf("AT %s, aUrl=%s\n", godebug.LF(), aUrl)
			}
		} else {
			foundIt, aUrl := chk2(rootOut, urlIn)
			if foundIt && db4 {
				fmt.Printf("AT %s, aUrl=%s\n", godebug.LF(), aUrl)
				fmt.Printf("AT %s -- indicates should have successulf file generation\n", godebug.LF())
			}
		}
	}
	return
}

func UrlFileExt(fcfg *FileServerType, www http.ResponseWriter, req *http.Request, urlIn string, g *FSConfig, rulNo int) (urlOut string, rootOut string, stat RuleStatus, err error) {
	if rw, ok := www.(*goftlmux.MidBuffer); ok {
		_ = rw

		if db1 || db9 || dbD {
			fmt.Printf("\n\n------------------------ UrlFileExt ---------------------- AT %s\n", godebug.LF())
		}

		rootOut = g.PreRule[rulNo].UseRoot
		urlOut = urlIn
		stat = PreNext

		user, theme, t_root := getThemeUserRoot(fcfg, www, req)

		// check if input, and output exists first, if so then ...
		// if have input, not output || have input and delta T to output, then rebuild
		chk1 := func(inExt, wantExt string, ti int, tr *ExtProcessType, pth ...string) (foundIt bool) {
			if db1 {
				fmt.Printf("AT %s ---- doing chk1 inExt >%s< wantExt >%s< pth >%s<\n", godebug.LF(), inExt, wantExt, pth)
			}
			// name = filepath.Join(t_root, user, theme, urlIn, wantExt)
			outputFn := filepath.Join(pth...) + wantExt
			inputFn := filepath.Join(pth...) + inExt
			if db1 {
				fmt.Printf("!!!!!! AT %s -- outputFn >%s< inputFn >%s<\n", godebug.LF(), outputFn, inputFn)
			}
			if ok, outFi := lib.ExistsGetFileInfo(outputFn); ok { // If we have input and output, then we will check to see if need rebuild
				if db1 {
					fmt.Printf("AT %s\n", godebug.LF())
				}
				haveInput, inFi := lib.ExistsGetFileInfo(inputFn)
				if haveInput {
					if db1 {
						fmt.Printf("AT %s\n", godebug.LF())
					}
					// rw.ResolvedFn = name
					rw.DependentFNs = append(rw.DependentFNs, inputFn)
					runCmdIfNecessary(fcfg, www, req, inputFn, haveInput, inFi, inExt, outputFn, true, outFi, wantExt, ti, tr)
				}
				stat = PreSuccess
				rootOut = t_root
				// urlOut = "/" + filepath.Join(user, theme, urlIn, wantExt)
				return true
			} else if haveInput, inFi := lib.ExistsGetFileInfo(inputFn); haveInput { // have input, but no output
				if db1 {
					fmt.Printf("AT %s\n", godebug.LF())
				}
				// rw.ResolvedFn = name
				rw.DependentFNs = append(rw.DependentFNs, inputFn)
				runCmdIfNecessary(fcfg, www, req, inputFn, haveInput, inFi, inExt, outputFn, false, nil, wantExt, ti, tr)
				stat = PreSuccess
				rootOut = t_root
				// urlOut = "/" + filepath.Join(user, theme, urlIn, wantExt)
				return true
			}
			if db1 {
				fmt.Printf("AT %s\n", godebug.LF())
			}
			return false
		}

		chk2 := func(inExt, wantExt string, ti int, tr *ExtProcessType, pth ...string) (foundIt bool, urlOut string, outputFn string) {
			if db4 {
				fmt.Printf("AT %s ---- doing chk2 inExt >%s< wantExt >%s< pth >%s<\n", godebug.LF(), inExt, wantExt, pth)
			}
			// name = filepath.Join(t_root, user, theme, urlIn, wantExt)
			outputFn = filepath.Join(pth...) + wantExt
			inputFn := filepath.Join(pth...) + inExt
			if db4 {
				fmt.Printf("AT %s -- outputFn >%s< inputFn >%s<\n", godebug.LF(), outputFn, inputFn)
			}
			if ok, outFi := lib.ExistsGetFileInfo(outputFn); ok { // If we have input and output, then we will check to see if need rebuild
				if db4 {
					fmt.Printf("AT %s\n", godebug.LF())
				}
				haveInput, inFi := lib.ExistsGetFileInfo(inputFn)
				if haveInput {
					if db4 {
						fmt.Printf("AT %s\n", godebug.LF())
					}
					foundIt = checkRunCmdIfNecessary(fcfg, www, req, inputFn, haveInput, inFi, inExt, outputFn, true, outFi, wantExt, ti, tr)
				}
				stat = PreSuccess
				rootOut = t_root
				urlOut = "/" + filepath.Join(user, theme, urlIn, wantExt)
				return
			} else if haveInput, inFi := lib.ExistsGetFileInfo(inputFn); haveInput { // have input, but no output
				if db4 {
					fmt.Printf("AT %s\n", godebug.LF())
				}
				foundIt = checkRunCmdIfNecessary(fcfg, www, req, inputFn, haveInput, inFi, inExt, outputFn, false, nil, wantExt, ti, tr)
				stat = PreSuccess
				rootOut = t_root
				urlOut = "/" + filepath.Join(user, theme, urlIn, wantExt)
				return
			}
			if db4 {
				fmt.Printf("AT %s\n", godebug.LF())
			}
			return
		}

		// -- -- for each row in the table
		for ti, tr := range ExtProcessTable {

			if dbA {
				fmt.Printf("%sAT: %s%s\n", MiscLib.ColorMagenta, godebug.LF(), MiscLib.ColorReset)
			}

			if tr.HasFs != "" {
				if db4 {
					fmt.Printf("MultiFile Combine AT %s, ti=%d\n", godebug.LF(), ti)
				}

				if dbA {
					fmt.Printf("%sAT: %s%s\n", MiscLib.ColorMagenta, godebug.LF(), MiscLib.ColorReset)
				}

				for oi, oext := range tr.ToExt {

					if db4 {
						fmt.Printf("Out Extensions : AT %s, oi=%d\n", godebug.LF(), oi)
					}
					if dbA {
						fmt.Printf("%sAT: %s%s\n", MiscLib.ColorMagenta, godebug.LF(), MiscLib.ColorReset)
					}

					if strings.HasSuffix(urlIn, oext) && strings.Index(urlIn, tr.HasFs) >= 0 { // if the URL has an output extension that we need and file seps

						ss := strings.Split(urlIn, tr.HasFs) // split into the set of file names
						dirPath := ""
						runIt := false
						urlOfEach := make([]string, 0, len(ss))
						inExtEach := make([]string, 0, len(ss))
						outputFn := ""
						wantExt := oext
						for ssi, sss := range ss {

							if dbA {
								fmt.Printf("%sAT: %s%s\n", MiscLib.ColorMagenta, godebug.LF(), MiscLib.ColorReset)
							}
							if db4 {
								fmt.Printf("AT %s, matched suffix! and ssi=%d sss=%s, user >%s< theme >%s<\n", godebug.LF(), ssi, sss, user, theme)
							}

							// if file statrs with '/' then reltive to top, else relative to 1st file
							if ssi == 0 {
								dirPath = filepath.Dir(sss)
							} else if ssi > 0 && sss[0] != '/' && dirPath != "" { // means could be ./ ../ or [name]
								ss[ssi] = dirPath + "/" + sss
							}

							inExt := tr.FromExt
							inExtEach = append(inExtEach, inExt)
							turlIn := RmExtSpecified(sss, oext)
							if db4 {
								fmt.Printf("AT %s, ti=%d root >%s< turlIn >%s<\n", godebug.LF(), ti, t_root, turlIn)
							}
							if user != "" && theme != "" {
								foundIt, aUrl, outputFn0 := chk2(inExt, oext, ti, tr, t_root, user, theme, turlIn)
								if foundIt {
									if db4 {
										fmt.Printf("AT %s\n", godebug.LF())
									}
									runIt = true
									if outputFn == "" {
										outputFn = outputFn0
									}
								}
								urlOfEach = append(urlOfEach, aUrl)
							} else if theme != "" {
								foundIt, aUrl, outputFn0 := chk2(inExt, oext, ti, tr, t_root, theme, turlIn)
								if foundIt {
									if db4 {
										fmt.Printf("AT %s\n", godebug.LF())
									}
									runIt = true
									if outputFn == "" {
										outputFn = outputFn0
									}
								}
								urlOfEach = append(urlOfEach, aUrl)
							} else {
								if dbA {
									fmt.Printf("%sAT: %s%s\n", MiscLib.ColorMagenta, godebug.LF(), MiscLib.ColorReset)
								}
								foundIt, aUrl, outputFn0 := chk2(inExt, oext, ti, tr, rootOut, turlIn)
								if foundIt {
									if db4 {
										fmt.Printf("AT %s -- indicates should have successulf file generation\n", godebug.LF())
									}
									runIt = true
									if outputFn == "" {
										outputFn = outputFn0
									}
								}
								urlOfEach = append(urlOfEach, aUrl)
							}
							if db4 {
								fmt.Printf("runIt = %v urlOfEach = %s inExtEach = %s AT %s\n", runIt, urlOfEach, inExtEach, godebug.LF())
							}
						}

						if dbA {
							fmt.Printf("%sAT: %s%s\n", MiscLib.ColorMagenta, godebug.LF(), MiscLib.ColorReset)
						}
						if runIt {
							if dbA {
								fmt.Printf("%sAT: %s%s\n", MiscLib.ColorMagenta, godebug.LF(), MiscLib.ColorReset)
							}
							runCmd(fcfg, www, req, urlOfEach, inExtEach, outputFn, wantExt, ti, tr)
						}
					}
				}
				if dbA {
					fmt.Printf("%sAT: %s%s\n", MiscLib.ColorMagenta, godebug.LF(), MiscLib.ColorReset)
				}

			} else {

				if db1 {
					fmt.Printf("Regular process file -> file AT %s, ti=%d\n", godebug.LF(), ti)
				}

				// -- -- for each of the [out] extentions in the table
				for oi, oext := range tr.ToExt {
					_ = oi
					if db1 {
						fmt.Printf("AT %s, oi=%d\n", godebug.LF(), oi)
					}

					if strings.HasSuffix(urlIn, oext) { // if the URL has an output extension that we need

						if db1 {
							fmt.Printf("AT %s, matched suffix!, user >%s< theme >%s<\n", godebug.LF(), user, theme)
						}
						inExt := tr.FromExt
						turlIn := RmExtSpecified(urlIn, oext)
						if db1 {
							fmt.Printf("AT %s, ti=%d root >%s< turlIn >%s<\n", godebug.LF(), ti, t_root, turlIn)
						}
						if user != "" && theme != "" && chk1(inExt, oext, ti, tr, t_root, user, theme, turlIn) {
							if db1 {
								fmt.Printf("AT %s\n", godebug.LF())
							}
							return
						} else if theme != "" && chk1(inExt, oext, ti, tr, t_root, theme, turlIn) {
							if db1 {
								fmt.Printf("AT %s\n", godebug.LF())
							}
							return
						} else {
							//for jj := range fcfg.Root {
							//	x_root, _ := filepath.Abs(fcfg.Root[jj]) // From Config									// xyzzyRoot0 <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< fix this <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
							//	if chk1(inExt, oext, ti, tr, x_root, turlIn) {
							//		fmt.Printf("AT %s -- indicates should have successulf file generation\n", godebug.LF())
							//		return
							//	}
							//}
							// rootOut = g.PreRule[rulNo].UseRoot
							if chk1(inExt, oext, ti, tr, rootOut, turlIn) {
								if db1 {
									fmt.Printf("AT %s -- indicates should have successulf file generation\n", godebug.LF())
								}
								return
							}
						}
						if db1 {
							fmt.Printf("AT %s\n", godebug.LF())
						}

					}
				}
			}
		}

	}
	return
}

// func ExistsGetFileInfo(name string) (bool, os.FileInfo) {
// runCmdIfNecessary(inputFn, haveInput, inFi, inExt, outputFn, false, nil, wantExt, ti, tr)
func runCmdIfNecessary(
	fcfg *FileServerType, www http.ResponseWriter, req *http.Request,
	inputFn string, haveInput bool, inFi os.FileInfo, inExt string,
	outputFn string, haveOutput bool, outFi os.FileInfo, outExt string,
	ti int, tr *ExtProcessType) {

	if db1 {
		fmt.Printf("AT %s\n", godebug.LF())
	}
	// run commands if necessary to build output
	runIt := tr.ReRun
	if tr.ReRun {
	} else if haveInput && haveOutput && inFi != nil && outFi != nil {
		// if inFi.ModTime().After(outFi.ModTime()) { // or on mac and (inFi - outFi) duration < 1 second (HFS+ defect in timing of data)
		if CompareModTime(inFi.ModTime(), outFi.ModTime()) == NeedRebuild { // function used because this will need to be OS specific, Mac OS X HFS+ has only 1 second accuracy
			runIt = true
		}
	} else if haveInput {
		runIt = true
	}
	if runIt {
		ok, outString, errString := ExecuteCommands(tr.CommandToRun, inputFn, outputFn, inExt, outExt)
		if !ok {
			LogErrors(fcfg, www, req, errString+outString)
			OutputErrors(fcfg, www, req, errString+outString)
		} else {
			LogOutputAsSuccess(fcfg, www, req, outString)
		}
	}
}

func checkRunCmdIfNecessary(
	fcfg *FileServerType, www http.ResponseWriter, req *http.Request,
	inputFn string, haveInput bool, inFi os.FileInfo, inExt string,
	outputFn string, haveOutput bool, outFi os.FileInfo, outExt string,
	ti int, tr *ExtProcessType) (runIt bool) {

	if db1 {
		fmt.Printf("AT %s\n", godebug.LF())
	}
	// run commands if necessary to build output
	runIt = tr.ReRun
	if tr.ReRun {
	} else if haveInput && haveOutput {
		// if inFi.ModTime().After(outFi.ModTime()) { // or on mac and (inFi - outFi) duration < 1 second (HFS+ defect in timing of data)
		if CompareModTime(inFi.ModTime(), outFi.ModTime()) == NeedRebuild { // function used because this will need to be OS specific, Mac OS X HFS+ has only 1 second accuracy
			runIt = true
		}
	} else if haveInput {
		runIt = true
	}
	return
}

func runCmd(
	fcfg *FileServerType, www http.ResponseWriter, req *http.Request,
	inputFn []string, inExt []string,
	outputFn string, outExt string,
	ti int, tr *ExtProcessType) {

	if db1 {
		fmt.Printf("AT %s\n", godebug.LF())
	}
	ok, outString, errString := ExecuteCommandsMulti(tr.CommandToRun, inputFn, outputFn, inExt, outExt)
	if !ok {
		LogErrors(fcfg, www, req, errString+outString)
		OutputErrors(fcfg, www, req, errString+outString)
	} else {
		LogOutputAsSuccess(fcfg, www, req, outString)
	}
}

func LogErrors(fcfg *FileServerType, www http.ResponseWriter, req *http.Request, s string) {
	fmt.Printf("Errors - Loged: %s, %s\n", s, godebug.LF())
	nBuildErrLogged++
	// if rw, ok := www.(*goftlmux.MidBuffer); ok {
	logrus.Error(s)
	// }
}

var nErrLogged = 0
var nSuccessLogged = 0
var nBuildErrLogged = 0
var nBuildSuccessLogged = 0

func InitNLogged() {
	nErrLogged = 0
	nSuccessLogged = 0
	nBuildErrLogged = 0
	nBuildSuccessLogged = 0
}

func OutputErrors(fcfg *FileServerType, www http.ResponseWriter, req *http.Request, s string) {
	fmt.Printf("Output - Loged: %s, %s\n", s, godebug.LF())
	nBuildErrLogged++
	// if rw, ok := www.(*goftlmux.MidBuffer); ok {
	logrus.Info(s)
	// }
}

func LogOutputAsSuccess(fcfg *FileServerType, www http.ResponseWriter, req *http.Request, s string) {
	fmt.Printf("Success - Loged: %s\n", s)
	nBuildSuccessLogged++
	// if rw, ok := www.(*goftlmux.MidBuffer); ok {
	logrus.Info(s)
	// }
}

func RmExt(filename string) string {
	var extension = filepath.Ext(filename)
	var name = filename[0 : len(filename)-len(extension)]
	return name
}

// Remove the specified extension from the file name.  If the extension is the entire file name then return an empty stirng.
func RmExtSpecified(filename, extension string) string {
	if len(extension) < len(filename) {
		if strings.HasSuffix(filename, extension) {
			var name = filename[0 : len(filename)-len(extension)]
			return name
		}
		return filename
	}
	return ""
}

type ACommandToRun struct {
	Cmd    string
	Params []string
}

// ok, errString := ExecuteCommands(tr.CommandToRun)
// need to make this a "worker" and only run 1 at a time of this.,,  All others wait pending!
func ExecuteCommands(CommandToRun string, inputFile, outputFile, inExt, outExt string) (ok bool, outString string, errString string) {

	// fmt.Printf("AT %s\n", godebug.LF())
	ok, outString, errString = true, "", ""
	var jdata ACommandToRun
	err := json.Unmarshal([]byte(CommandToRun), &jdata)
	if err != nil {
		ok = false
		// fmt.Sprintf(errString, "Error: %s - in unmarshaling of CommandToRun: %s\n", err, CommandToRun)
		es := jsonSyntaxErroLib.GenerateSyntaxError(string(CommandToRun), err)
		fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorYellow, es, MiscLib.ColorReset)
		logrus.Errorf("Error: Invlaid JSON for %s Error:\n%s\n", CommandToRun, es)
		return
	}
	// fmt.Printf("AT %s\n", godebug.LF())
	Params := make([]string, 0, len(jdata.Params))
	data := make(map[string]string)
	data["inputFile"] = inputFile
	data["outputFile"] = outputFile
	data["base_file_name"] = inputFile
	if len(inExt) < len(inputFile) {
		data["base_file_name"] = inputFile[0 : len(inputFile)-len(inExt)]
	}
	data["inputDir"] = filepath.Dir(inputFile)
	data["inputBase"] = filepath.Base(inputFile)
	data["outputDir"] = filepath.Dir(outputFile)
	data["outputBase"] = filepath.Base(outputFile)
	data["inputExt"] = inExt
	data["outputExt"] = outExt
	for ii, vv := range jdata.Params {
		// fmt.Printf("AT %s\n", godebug.LF())
		data["param_no"] = fmt.Sprintf("%d", ii)
		if db1 {
			fmt.Printf("********************************** data = %s, %s\n", lib.SVarI(data), godebug.LF())
		}
		Params = append(Params, tmplp.ExecuteATemplate(vv, data))
	}

	cmd, run := CommandLocationMap[jdata.Cmd]
	if !run {
		ok = false
		errString = fmt.Sprintf("{\"status\":\"error\",\"msg\":\"Error(15023): Command is not authorized, %s.\"}", jdata.Cmd)
		return
	}

	if db2 {
		fmt.Printf("AT Cmd = %s Params=%s, %s\n", cmd, Params, godebug.LF())
	}

	key := GenLockKey(cmd, Params...)
	LockCmd(key)
	defer UnLockCmd(key)
	out, err := exec.Command(cmd, Params...).Output() // Run the command, get the output.
	if err != nil {                                   // If command running failed, report error go to next row
		// fmt.Printf("AT %s\n", godebug.LF())
		ok = false
		errString = fmt.Sprintf("{\"status\":\"error\",\"err\":%q,\"msg\":\"Error(14023): Unable to execute command.\"}", err)
		outString = string(out)
		return
	} else {
		// fmt.Printf("AT %s\n", godebug.LF())
		outString = fmt.Sprintf("{\"status\":\"success\",\"output\":%q}", string(out))
	}
	// fmt.Printf("AT %s\n", godebug.LF())

	return
}

func ExecuteCommandsMulti(CommandToRun string, inputFile []string, outputFile string, inExt []string, outExt string) (ok bool, outString string, errString string) {

	// fmt.Printf("AT %s\n", godebug.LF())
	ok, outString, errString = true, "", ""
	var jdata ACommandToRun
	err := json.Unmarshal([]byte(CommandToRun), &jdata)
	if err != nil {
		ok = false
		// fmt.Sprintf(errString, "Error: %s - in unmarshaling of CommandToRun: %s\n", err, CommandToRun)
		es := jsonSyntaxErroLib.GenerateSyntaxError(string(CommandToRun), err)
		fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorYellow, es, MiscLib.ColorReset)
		logrus.Errorf("Error: Invlaid JSON for %s Error:\n%s\n", CommandToRun, es)
		return
	}
	// fmt.Printf("AT %s\n", godebug.LF())
	Params := make([]string, 0, len(jdata.Params))
	data := make(map[string]string)
	data["inputFileList"] = strings.Join(inputFile, " ")
	data["outputFile"] = outputFile
	//data["base_file_name"] = inputFile		// xyzzy - may need to work on this
	//if len(inExt) < len(inputFile) {
	//	data["base_file_name"] = inputFile[0 : len(inputFile)-len(inExt)]
	//}
	data["inputDir"] = filepath.Dir(inputFile[0])   // xyzzy - may need to work on this
	data["inputBase"] = filepath.Base(inputFile[0]) // xyzzy - may need to work on this
	data["outputDir"] = filepath.Dir(outputFile)
	data["outputBase"] = filepath.Base(outputFile)
	data["inputExt"] = strings.Join(inExt, " ")
	data["outputExt"] = outExt
	for ii, vv := range jdata.Params {
		// fmt.Printf("AT %s\n", godebug.LF())
		data["param_no"] = fmt.Sprintf("%d", ii)
		if db1 {
			fmt.Printf("********************************** data = %s, %s\n", lib.SVarI(data), godebug.LF())
		}
		Params = append(Params, tmplp.ExecuteATemplate(vv, data))
	}

	cmd, run := CommandLocationMap[jdata.Cmd]
	if !run {
		ok = false
		errString = fmt.Sprintf("{\"status\":\"error\",\"msg\":\"Error(15023): Command is not authorized, %s.\"}", jdata.Cmd)
		return
	}

	if db2 {
		fmt.Printf("AT Cmd = %s Params=%s, %s\n", cmd, Params, godebug.LF())
	}

	key := GenLockKey(cmd, Params...)
	LockCmd(key)
	defer UnLockCmd(key)
	out, err := exec.Command(cmd, Params...).Output() // Run the command, get the output.
	if err != nil {                                   // If command running failed, report error go to next row
		// fmt.Printf("AT %s\n", godebug.LF())
		ok = false
		errString = fmt.Sprintf("{\"status\":\"error\",\"err\":%q,\"msg\":\"Error(14023): Unable to execute command.\"}", err)
		outString = string(out)
		return
	} else {
		// fmt.Printf("AT %s\n", godebug.LF())
		outString = fmt.Sprintf("{\"status\":\"success\",\"output\":%q}", string(out))
	}
	// fmt.Printf("AT %s\n", godebug.LF())

	return
}

type RebuildFlag int

const (
	NeedRebuild       RebuildFlag = 1
	TimestampsInOrder RebuildFlag = 2
)

func (nr RebuildFlag) String() string {
	switch nr {
	case NeedRebuild:
		return "NeedRebuild"
	case TimestampsInOrder:
		return "TimestampsInOrder"
	}
	return fmt.Sprintf("*** invalid RebuildFlag %d ***", nr)
}

// return true if modification time for in is after out
func CompareModTime(in, out time.Time) RebuildFlag {
	if in.After(out) { // or on mac and (inFi - outFi) duration < 1 second (HFS+ defect in timing of data)
		// fmt.Printf("After is true, %s\n", godebug.LF())
		return NeedRebuild
	}
	/*
		// I am not certain that this is correct -
		fmt.Printf("fallthrough, %s\n", godebug.LF())
		if runtime.GOOS == "darwin" { // if on mac
			fmt.Printf("on mac, %s\n", godebug.LF())
			duration := in.Sub(out)
			deltaT := duration.Nanoseconds()
			fmt.Printf("deltaT = %v nanoseconds, %s\n", deltaT, godebug.LF())
			if deltaT > -1000000000 {
				return NeedRebuild
			}
		}
	*/
	return TimestampsInOrder
}

type ExLockType struct {
	// Key   string
	mutex sync.Mutex //
}

/*
	hdlr.mutex.Lock()
	hdlr.mutex.Unlock()
*/

var cmdLockMutex sync.RWMutex //
var cmdLock map[string]ExLockType

func init() {
	cmdLock = make(map[string]ExLockType)
}

func GenLockKey(s0 string, s ...string) (rv string) {
	rv = s0
	com := "!"
	for _, vv := range s {
		rv += com + vv
	}
	return
}

func LockCmd(aKey string) {
	return
	// fmt.Fprintf(os.Stderr, "LockCmd[%s], %s\n", aKey, godebug.LF())
	cmdLockMutex.RLock()
	lk, ok := cmdLock[aKey]
	cmdLockMutex.RUnlock()
	if !ok {
		// fmt.Fprintf(os.Stderr, "LockCmd[%s], -- !ok - build one %s\n", aKey, godebug.LF())
		// lk = ExLockType{Key: aKey}
		lk = ExLockType{}
		cmdLockMutex.Lock()
		lk.mutex.Lock()
		cmdLock[aKey] = lk
		cmdLockMutex.Unlock()
		return
	}
	// fmt.Fprintf(os.Stderr, "LockCmd[%s], -- is locked %s\n", aKey, godebug.LF())
	lk.mutex.Lock()
}

func UnLockCmd(aKey string) {
	return
	// fmt.Fprintf(os.Stderr, "UnLockCmd[%s], %s\n", aKey, godebug.LF())
	cmdLockMutex.RLock()
	lk, ok := cmdLock[aKey]
	cmdLockMutex.RUnlock()
	if !ok {
		// fmt.Fprintf(os.Stderr, "UnLockCmd[%s], -- %s!ok%s - *** early return *** %s\n", aKey, MiscLib.ColorRed, MiscLib.ColorReset, godebug.LF())
		return
	}
	// fmt.Fprintf(os.Stderr, "UnLockCmd[%s], -- is *UN*locked %s\n", aKey, godebug.LF())
	lk.mutex.Unlock()
}

const db_fileServer = false
const db1 = false
const db2 = false
const db4 = false
const db5 = false
const db6 = false
const db7 = false // dump request beofre making it.
const db8 = false // New directory template/name search stuff
const db9 = false // New  Windows one
const dbA = false // New  Windows one

/* vim: set noai ts=4 sw=4: */
