package sizlib

// (C) Copyright Philip Schlump, 2013-2014

// _ "github.com/mattn/go-oci8"			// OCI

import (
	// _ "../odbc" // _ "code.google.com/p/odbc"
	// _ "github.com/lib/pq"
	// _ "../pq" // _ "github.com/lib/pq"
	// _ "github.com/mattn/go-oci8"			// OCI
	// "database/sql"

	// "github.com/jackc/pgx" //  https://github.com/jackc/pgx

	_ "github.com/lib/pq"

	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pschlump/godebug"
	"github.com/pschlump/json" //	"encoding/json"
)

// "github.com/nu7hatch/gouuid"

// UUID?? import "code.google.com/p/go-uuid/uuid"
//			import "github.com/nu7hatch/gouuid"
//			import "github.com/twinj/uuid"

// "database/sql/driver"
// "math/rand"
// "bytes"

// ISO format for date
const ISO8601 = "2006-01-02T15:04:05.99999Z07:00"

// ISO format for date
const ISO8601output = "2006-01-02T15:04:05.99999-0700"

// SVar convert a variable to it's JSON representation and return
func SVar(v interface{}) string {
	s, err := json.Marshal(v)
	// s, err := json.MarshalIndent ( v, "", "\t" )
	if err != nil {
		return fmt.Sprintf("Error:%s", err)
	}
	return string(s)
}

// SVarI convert a variable to it's JSON representation with indendted JSON
func SVarI(v interface{}) string {
	// s, err := json.Marshal ( v )
	s, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return fmt.Sprintf("Error:%s", err)
	}
	return string(s)
}

// moved IsUUID to uuid

// -------------------------------------------------------------------------------------------------
/*
type FileInfo interface {
	Name() string       // base name of the file
	Size() int64        // length in bytes for regular files; system-dependent for others
	Mode() FileMode     // file mode bits
	ModTime() time.Time // modification time
	IsDir() bool        // abbreviation for Mode().IsDir()
	Sys() interface{}   // underlying data source (can return nil)
}
*/
// Tested
// Exists reports whether the named file or directory exists.
// -------------------------------------------------------------------------------------------------
func ExistsGetUDate(name string) (bool, os.FileInfo) {
	fi, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
	}
	return true, fi
}

func ExistsIsDir(name string) bool {
	fi, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	if fi.IsDir() {
		return true
	}
	return false
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------
func DumpVar(v interface{}) {
	// s, err := json.Marshal ( v )
	s, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	} else {
		fmt.Printf("%s\n", s)
	}
}

// -------------------------------------------------------------------------------------------------
// xyzzy - str.
// Return the basename from a file path.  This is the last component with the directory path
// stripped off.  File extension removed.
// -------------------------------------------------------------------------------------------------
func Basename(fn string) (bn string) {
	i, j := strings.LastIndex(fn, "/"), strings.LastIndex(fn, path.Ext(fn)) // xyzzy windoz
	// fmt.Printf ( "i=%d j=%d\n", i, j )
	if i < 0 && j < 0 {
		bn = fn
	} else if i < 0 {
		bn = fn[0:j]
	} else {
		bn = fn[i+1 : j]
	}
	return
}

// -------------------------------------------------------------------------------------------------
// xyzzy - str.
// With file extension
// -------------------------------------------------------------------------------------------------
func BasenameExt(fn string) (bn string) {
	i, j := strings.LastIndex(fn, "/"), len(fn) // xyzzy windoz
	// fmt.Printf ( "i=%d j=%d\n", i, j )
	if i < 0 && j < 0 {
		bn = fn
	} else if i < 0 {
		bn = fn[0:j]
	} else {
		bn = fn[i+1 : j]
	}
	return
}

// -------------------------------------------------------------------------------------------------
// Get a list of filenames and directorys.
// xyzzy - fil.
// -------------------------------------------------------------------------------------------------
func GetFilenames(dir string) (filenames, dirs []string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, nil
	}
	for _, fstat := range files {
		if !strings.HasPrefix(string(fstat.Name()), ".") {
			if fstat.IsDir() {
				dirs = append(dirs, fstat.Name())
			} else {
				filenames = append(filenames, fstat.Name())
			}
		}
	}
	return
}

// -------------------------------------------------------------------------------------------------
// xyzzy - str.
// -------------------------------------------------------------------------------------------------
func InArray(lookFor string, inArr []string) bool {
	for _, v := range inArr {
		if lookFor == v {
			return true
		}
	}
	return false
}

func InArrayN(lookFor string, inArr []string) int {
	for i, v := range inArr {
		if lookFor == v {
			return i
		}
	}
	return -1
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------
func FilterArray(re string, inArr []string) (outArr []string) {
	var validID = regexp.MustCompile(re)

	outArr = make([]string, 0, len(inArr))
	for k := range inArr {
		if validID.MatchString(inArr[k]) {
			outArr = append(outArr, inArr[k])
		}
	}
	// fmt.Printf ( "output = %v\n", outArr )
	return
}

// -------------------------------------------------------------------------------------------------
// list of image files in a directory
// -------------------------------------------------------------------------------------------------
func GetImageFiles(dir string) []string {
	filenames, _ := GetFilenames(dir)
	// xyzzy - case sensitive!
	imgFn := FilterArray(".jpg$|.jpeg$|.png$|.gif$", filenames)
	return imgFn
}

func FilesMatchingPattern(dir, pattern string) []string {
	filenames, _ := GetFilenames(dir)
	// xyzzy - case sensitive!
	imgFn := FilterArray(pattern, filenames)
	return imgFn
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------
func GetFile(fn string) string {
	file, err := ioutil.ReadFile(fn)
	if err != nil {
		fmt.Printf("Error(10103): File (%s) missing or unreadable error: %v\n", fn, err)
		return ""
	}
	return string(file)
}

// -------------------------------------------------------------------------------------------------
// Tested
// -------------------------------------------------------------------------------------------------
func RmExt(filename string) string {
	var extension = filepath.Ext(filename)
	var name = filename[0 : len(filename)-len(extension)]
	return name
}

// -------------------------------------------------------------------------------------------------
// xyzzy - str.
// Tested
// -------------------------------------------------------------------------------------------------
func FnToCSSClass(filename string) string {
	re2 := regexp.MustCompile("[^-a-zA-Z0-9_]")
	s := re2.ReplaceAllLiteralString(filename, "-")
	return s
}

// -------------------------------------------------------------------------------------------------
// time.After - see
// -------------------------------------------------------------------------------------------------
func getMtimeOfFile(fn string) (tm time.Time, err error) {
	info, err := os.Stat(fn)
	if err != nil {
		return
	}
	tm = info.ModTime()
	return
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------
func NeedRebuild(parent string, optsRebuild bool, child string, child1 string, child2 string) bool {
	if optsRebuild {
		return true
	}
	pt, err := getMtimeOfFile(parent)
	if err != nil {
		return false
	}

	ct, err := getMtimeOfFile(child)
	if err != nil {
		return true
	}
	t1 := pt.After(ct)
	if t1 {
		return true
	}

	if child1 != "" {
		ct, err = getMtimeOfFile(child1)
		if err != nil {
			return true
		}
		t1 = pt.After(ct)
		if t1 {
			return true
		}
	}

	if child2 != "" {
		ct, err = getMtimeOfFile(child2)
		if err != nil {
			return true
		}
		t1 = pt.After(ct)
		if t1 {
			return true
		}
	}

	return false
}

// -------------------------------------------------------------------------------------------------
// mTime on file
// Compare mTims' to see if need to act - return bool
// Make directory path

// read JSON file to map
// -------------------------------------------------------------------------------------------------
func ReadJSONPath(pth string) map[string]string {
	var d string
	c := make(map[string]string, 40)
	if pth[len(pth)-1] == '/' {
		d, _ = filepath.Split(pth[0 : len(pth)-1])
	} else {
		d, _ = filepath.Split(pth)
		pth = pth + "/"
	}
	if d == "" {
		if Exists("cfg.json") {
			c = readJSONFile("cfg.json")
		}
		return c
	}
	c = ReadJSONPath(d)
	if Exists(pth + "cfg.json") {
		m := readJSONFile(pth + "cfg.json")
		v := ExtendStringMap(c, m)
		return v
	}
	return c
}

// -------------------------------------------------------------------------------------------------
// not used
// -------------------------------------------------------------------------------------------------
func doPath(pth string) {
	var d, f string
	if string(pth[len(pth)-1]) == "/" {
		d, f = filepath.Split(pth[0 : len(pth)-1])
	} else {
		d, f = filepath.Split(pth)
		pth = pth + "/"
	}
	fmt.Printf("d=|%s| f=|%s|\n", d, f)
	if d == "" {
		fmt.Printf("doPath: %s\n", "cfg.json")
	} else {
		doPath(d)
	}
	fmt.Printf("doPath: %s\n", pth+"cfg.json")
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------
func readJSONFile(fn string) map[string]string {
	jdata := make(map[string]string, 40) // The posts that match
	file, e := ioutil.ReadFile(fn)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	json.Unmarshal(file, &jdata)
	return jdata
}

// -------------------------------------------------------------------------------------------------
// Tested
// Exists reports whether the named file or directory exists.
// -------------------------------------------------------------------------------------------------
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func URIToStringMap(req *http.Request) (m url.Values, fr map[string]string) {

	ct := req.Header.Get("Content-Type")

	fr = make(map[string]string)

	// db_uriToString := false

	// if ( db_uriToString ) { fmt.Printf ( "PJS Apr 9: %s Content Type:%v\n", godebug.LF(), ct ) }
	//if db_uriToString {
	//	fmt.Printf("PJS Sep 20: %s Content Type:%v\n", godebug.LF(), ct)
	//}

	u, _ := url.ParseRequestURI(req.RequestURI)
	m, _ = url.ParseQuery(u.RawQuery)
	for i := range m {
		fr[i] = "Query-String"
	}

	// xyzzy - add in cookies??		req.Cookies() -> []string
	// if ( db_uriToString ) { fmt.Printf ( "Cookies are: %s\n", sizlib.SVar( req.Cookies() ) ) }
	Ck := req.Cookies()
	for _, v := range Ck {
		if _, ok := m[v.Name]; !ok {
			m[v.Name] = make([]string, 1)
			m[v.Name][0] = v.Value
			// fmt.Printf ( "Name=%s Value=%s\n", v.Name, v.Value )
			fr[v.Name] = "Cookie"
		}
	}

	// fmt.Printf ( "Checking to see if post\n" )

	// add in POST parmeters
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" || req.Method == "DELETE" {
		//if db_uriToString {
		//	fmt.Printf("It's a POST/PUT/PATCH/DELETE, req.PostForm=%v, ct=%s\n", req.PostForm, ct)
		//}
		if req.PostForm == nil {
			// if ( db_uriToString ) { fmt.Printf ( "ParseForm has !!!not!!! been  called\n" ) }
			if strings.HasPrefix(ct, "application/json") {
				body, err2 := ioutil.ReadAll(req.Body)
				if err2 != nil {
					fmt.Printf("err=%v\n", err2)
				}
				// if ( db_uriToString) { fmt.Printf("body=%s\n",string(body)) }
				// fmt.Printf("request body=%s\n",string(body))
				var jsonData map[string]interface{}
				err := json.Unmarshal(body, &jsonData)
				if err == nil {
					for i, v := range jsonData {
						m[i] = make([]string, 1)
						switch v.(type) {
						case bool:
							m[i][0] = fmt.Sprintf("%v", v)
						case float64:
							m[i][0] = fmt.Sprintf("%v", v)
						case int64:
							m[i][0] = fmt.Sprintf("%v", v)
						case int32:
							m[i][0] = fmt.Sprintf("%v", v)
						case time.Time:
							m[i][0] = fmt.Sprintf("%v", v)
						case string:
							m[i][0] = fmt.Sprintf("%v", v)
						default:
							m[i][0] = fmt.Sprintf("%s", SVar(v))
						}
						fr[i] = "JSON-Encoded-Body/Post"
					}
				}
			} else {
				err := req.ParseForm()
				//if db_uriToString {
				//	fmt.Printf("Form data is now: %s\n", SVar(req.PostForm))
				//}
				if err != nil {
					fmt.Printf("Error - parse form just threw an error , why? %v\n", err)
				} else {
					for i, v := range req.PostForm {
						if len(v) > 0 {
							m[i] = make([]string, 1)
							m[i][0] = v[0]
							fr[i] = "URL-Encoded-Body(1-a)/Post"
						}
					}
				}
			}
		} else {
			for i, v := range req.PostForm {
				if len(v) > 0 {
					m[i] = make([]string, 1)
					m[i][0] = v[0]
					fr[i] = "URL-Encoded-Body(2)/Post"
				}
			}
		}
	}

	//if db_uriToString {
	//	fmt.Printf(">>m=%s\n", SVar(m))
	//}

	return
}

//------------------------------------------------------------------------------------------------
// Copy 'a', then copy 'b' over 'a'
// Tests:  t-extendData.go
//------------------------------------------------------------------------------------------------
// jDataDefaults = lowerCaseNames ( jDataDefaults )
func LowerCaseNames(a map[string]interface{}) (rv map[string]interface{}) {
	rv = make(map[string]interface{})
	for i, v := range a {
		rv[strings.ToLower(i)] = v
	}
	return
}

func ExtendData(a map[string]interface{}, b map[string]interface{}) (rv map[string]interface{}) {
	rv = make(map[string]interface{})
	for i, v := range a {
		rv[i] = v
	}
	for i, v := range b {
		rv[i] = v
	}
	return
}

// Copy 'a', if same key in 'b', then copy data from b, prefering data from 'b'
func LeftData(a map[string]interface{}, b map[string]interface{}) (rv map[string]interface{}) {
	rv = make(map[string]interface{})
	for i, v := range a {
		rv[i] = v
	}
	for i, v := range b {
		if _, ok := a[i]; ok {
			rv[i] = v
		}
	}
	return
}

// Keep the data that has common keys between 'a' and 'b', prefering data from 'b'
// not used at the moment.
func IntersectData(a map[string]interface{}, b map[string]interface{}) (rv map[string]interface{}) {
	rv = make(map[string]interface{})
	for i, v := range a {
		if _, ok := b[i]; ok {
			rv[i] = v
		}
	}
	for i, v := range b {
		if _, ok := a[i]; ok {
			rv[i] = v
		}
	}
	return
}

func ExtendDataS(a map[string]string, b map[string]string) (rv map[string]string) {
	rv = make(map[string]string)
	for i, v := range a {
		rv[i] = v
	}
	for i, v := range b {
		rv[i] = v
	}
	return
}

//------------------------------------------------------------------------------------------------
func EscapeDoubleQuote(s string) string {
	return strings.Replace(s, `"`, "\\\"", -1)
}

func EscapeError(err error) string {
	s := fmt.Sprintf("%v", err)
	return strings.Replace(s, `"`, "\\\"", -1)
}

func HexSha1(s string) string {
	h := sha1.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

//------------------------------------------------------------------------------------------------
func TypeOf(v []interface{}) {
	for i := range v {
		fmt.Printf("Type of %d = %T\n", i, v[i])
	}
}

// ===============================================================================================================================================================================================
var isIntStringRe *regexp.Regexp
var trueValues map[string]bool

func init() {
	isIntStringRe = regexp.MustCompile("[0-9][0-9]*")

	trueValues = make(map[string]bool)
	trueValues["t"] = true
	trueValues["T"] = true
	trueValues["yes"] = true
	trueValues["Yes"] = true
	trueValues["YES"] = true
	trueValues["1"] = true
	trueValues["true"] = true
	trueValues["True"] = true
	trueValues["TRUE"] = true
	trueValues["on"] = true
	trueValues["On"] = true
	trueValues["ON"] = true
}
func IsIntString(s string) bool {
	return isIntStringRe.MatchString(s)
}

func ParseBool(s string) (b bool) {
	_, b = trueValues[s]
	return
	//if InArray(s, []string{"t", "T", "yes", "Yes", "YES", "1", "true", "True", "TRUE", "on", "On", "ON"}) {
	//	return true
	//}
	//return false
}

//------------------------------------------------------------------------------------------------
// func HasKeys ( v map[string]Validation ) bool {
func HasKeys(v map[string]interface{}) bool {
	for _, _ = range v {
		return true
	}
	return false
}

// ====================================================================================================================================================================================
// 1. Split path into array
// 2. Use templates/substitution for
//     	~ == HOME			$HOME from environment
// This will search for [path] / file - [hostname] . ext, then...
// This will search for [path] / file  . ext, then...
// ====================================================================================================================================================================================

var hasUserPat *regexp.Regexp
var replUserPat *regexp.Regexp
var homeDir string

func init() {
	ps := string(os.PathSeparator)
	if ps != "/" {
		ps = ps + ps
	}

	hasUserPat = regexp.MustCompile("~([a-zA-Z][^" + ps + "]*)" + ps)
	replUserPat = regexp.MustCompile("(~[a-zA-Z][^" + ps + "]*)")

	homeDir = os.Getenv("HOME")
}

func SubstitueUserInFilePath(s string, mdata map[string]string) (rs string, has bool) {
	has = false
	x := hasUserPat.FindStringSubmatch(s)
	// fmt.Printf("x=%s\n", SVar(x))
	rs = s
	if len(x) > 1 {
		has = true
		p := x[1]
		ud, err := user.Lookup(p)
		if err != nil {
			fmt.Printf("Error (13922): unable to lookup %s as a username, error=%s\n", p, err)
		} else {
			mdata["USER_"+ud.Username] = ud.HomeDir
			rs = replUserPat.ReplaceAllLiteralString(rs, "%{USER_"+ud.Username+"%}")
		}
	} else if strings.HasPrefix(rs, "~") {
		// fmt.Printf("Before last substitue rs [%s]\n", rs)
		rs = strings.Replace(rs, "~", "%{HOME%}", 1)
		// fmt.Printf("At bottom rs [%s]\n", rs)
	}
	return
}

func SubstitueUserInFilePathImmediate(s string) (rs string) {
	x := hasUserPat.FindStringSubmatch(s)
	// fmt.Printf("x=%s\n", SVar(x))
	rs = s
	if len(x) > 1 {
		p := x[1]
		ud, err := user.Lookup(p)
		if err != nil {
			fmt.Printf("Error (13922): unable to lookup %s as a username, error=%s\n", p, err)
		} else {
			rs = replUserPat.ReplaceAllLiteralString(rs, ud.HomeDir)
		}
	} else if strings.HasPrefix(rs, "~") {
		// fmt.Printf("Before last substitue rs [%s]\n", rs)
		rs = strings.Replace(rs, "~", homeDir, 1)
		// fmt.Printf("At bottom rs [%s]\n", rs)
	}
	return
}

// 1. Match ~name
// 2. get "name" out
// 3. user.Lookup ( name )
// 4. Replace ~name with %{USER_name%}, set mdata["USER_name"]
// 5. DO .Qt

func SearchPath(rawFileName string, searchPath string) (fullFileName string, ok bool) {

	hostname, err := os.Hostname()
	if err != nil {
		fmt.Printf("Error(10020): Unable to get the hostname (%v)\n", err)
		os.Exit(1)
	}

	mdata := make(map[string]string, 30)
	mdata["HostName"] = hostname
	mdata["IS_WINDOWS"] = ""
	ps := string(os.PathSeparator)
	if ps != "/" {
		mdata["IS_WINDOWS"] = ""
	} else {
		mdata["IS_WINDOWS"] = "ms"
	}
	mdata["HOME"] = os.Getenv("HOME")
	mdata["FILENAMERAW"] = rawFileName
	mdata["FILENAME"] = RmExt(rawFileName)
	mdata["FILEEXT"] = filepath.Ext(rawFileName)
	if ps != "/" {
		ps = ps + ps
	}
	mdata["OS_SEP"] = ps

	sp := strings.Split(searchPath, string(os.PathListSeparator))
	ok = false
	for _, p := range sp {
		mdata["CUR_PATH"] = p

		tmpl := "%{CUR_PATH%}%{OS_SEP%}%{FILENAME%}-%{HostName%}%{FILEEXT%}"
		fullFileName = Qt(tmpl, mdata)
		fullFileName, _ = SubstitueUserInFilePath(fullFileName, mdata)
		fullFileName = Qt(fullFileName, mdata)
		// fmt.Printf("1: %s\n", fullFileName)
		if Exists(fullFileName) {
			ok = true
			return
		}

		tmpl = "%{CUR_PATH%}%{OS_SEP%}%{FILENAME%}%{FILEEXT%}"
		fullFileName = Qt(tmpl, mdata)
		fullFileName, _ = SubstitueUserInFilePath(fullFileName, mdata)
		fullFileName = Qt(fullFileName, mdata)
		// fmt.Printf("2: %s\n", fullFileName)
		if Exists(fullFileName) {
			ok = true
			return
		}

		tmpl = "%{CUR_PATH%}%{OS_SEP%}%{FILENAMERAW%}"
		fullFileName = Qt(tmpl, mdata)
		fullFileName, _ = SubstitueUserInFilePath(fullFileName, mdata)
		fullFileName = Qt(fullFileName, mdata)
		// fmt.Printf("3: %s\n", fullFileName)
		if Exists(fullFileName) {
			ok = true
			return
		}

	}
	fullFileName = rawFileName
	fullFileName, _ = SubstitueUserInFilePath(fullFileName, mdata)
	fullFileName = Qt(fullFileName, mdata)
	ok = Exists(fullFileName)
	return
}

// Perform a search for files and return the full name for each file.
//
// rawFileName: sql-cfg.json
// appName: "store"
//
func SearchPathApp(rawFileName string, appName string, searchPath string) (fullFileName string, ok bool) {

	hostname, err := os.Hostname()
	if err != nil {
		fmt.Printf("Error(10020): Unable to get the hostname (%v)\n", err)
		os.Exit(1)
	}

	mdata := make(map[string]string, 30)
	mdata["HostName"] = hostname
	mdata["AppName"] = appName
	if dbInit1 {
		fmt.Printf("-- HostName [%s] AppName [%s] --\n", hostname, appName)
	}
	mdata["IS_WINDOWS"] = ""
	ps := string(os.PathSeparator)
	if ps != "/" {
		mdata["IS_WINDOWS"] = ""
	} else {
		mdata["IS_WINDOWS"] = "ms"
	}
	mdata["HOME"] = os.Getenv("HOME")
	mdata["FILENAMERAW"] = rawFileName
	mdata["FILENAME"] = RmExt(rawFileName)
	mdata["FILEEXT"] = filepath.Ext(rawFileName)
	if ps != "/" {
		ps = ps + ps
	}
	mdata["OS_SEP"] = ps

	sp := strings.Split(searchPath, string(os.PathListSeparator))
	ok = false
	tmplArr := []string{
		"%{CUR_PATH%}%{OS_SEP%}%{FILENAME%}-%{AppName%}-%{HostName%}%{FILEEXT%}",
		"%{CUR_PATH%}%{OS_SEP%}%{FILENAME%}-%{AppName%}%{FILEEXT%}",
		"%{CUR_PATH%}%{OS_SEP%}%{FILENAME%}-%{HostName%}%{FILEEXT%}",
		"%{CUR_PATH%}%{OS_SEP%}%{FILENAME%}%{FILEEXT%}",
		"%{CUR_PATH%}%{OS_SEP%}%{FILENAMERAW%}",
	}
	for _, p := range sp {
		mdata["CUR_PATH"] = p

		for _, tmpl := range tmplArr {
			fullFileName = Qt(tmpl, mdata)
			fullFileName, _ = SubstitueUserInFilePath(fullFileName, mdata)
			fullFileName = Qt(fullFileName, mdata)
			if dbInit1 {
				fmt.Printf("-- Test to see if -->>%s<<-- exits -- \n", fullFileName)
			}
			if Exists(fullFileName) {
				ok = true
				return
			}
		}

	}
	fullFileName = rawFileName
	fullFileName, _ = SubstitueUserInFilePath(fullFileName, mdata)
	fullFileName = Qt(fullFileName, mdata)
	ok = Exists(fullFileName)
	return
}

const dbInit1 = false

//var db_uriToString = false
//
//func SizlibSetDebugFlag(s string, v bool) {
//	switch s {
//	case "uriToString":
//		db_uriToString = v
//	}
//}

func FindFiles(pth string, ignoreDirs []string) (rv []string) {
	// fmt.Printf("pth=->%s<-, checking vs %s, %s\n", pth, SVar(ignoreDirs), godebug.LF())
	if InArray(pth, ignoreDirs) {
		return
	}
	fns, dirs := GetFilenames(pth)
	fns = FilterArray("^sql-cfg.*\\.json$", fns)
	for i, v := range fns {
		fns[i] = pth + "/" + v
	}
	rv = append(rv, fns...)
	for _, v := range dirs {
		trv := FindFiles(pth+"/"+v, ignoreDirs)
		rv = append(rv, trv...)
	}
	return
}

func FindDirsWithSQLCfg(pth string, ignoreDirs []string) (rv []string) {
	fns, dirs := GetFilenames(pth)
	_ = dirs
	fns = FilterArray("^sql-cfg.*\\.json$", fns)
	for i, v := range fns {
		fns[i] = pth + "/" + v
	}
	fmt.Printf("fns=%s\n", fns)
	for _, vv := range fns {
		ww := Dirname(vv)
		if InArray(ww, ignoreDirs) {
		} else if !InArray(ww, rv) {
			rv = append(rv, ww)
		}
	}
	return

	//	rv = append(rv, fns...)
	//	for _, v := range dirs {
	//		trv := FindFiles(pth+"/"+v, ignoreDirs)
	//		for _, w := range trv {
	//			ww := Dirname(w)
	//			if !InArray(ww, rv) {
	//				rv = append(rv, ww)
	//			}
	//		}
	//	}
	//	return

}

// Return the directory part of a file name
func Dirname(fn string) (bn string) {
	fn = filepath.Clean(fn)
	i := strings.LastIndex(fn, "/")
	// fmt.Printf("i=%d\n", i)
	bn = fn
	if i > 0 {
		bn = fn[0:i]
	}
	fmt.Printf("Dirname Input[%s] Output[%s], %s\n", fn, bn, godebug.LF())
	return
}

func SearchPathAppModule(rawFileName string, appName string, searchPath []string) (fullFileName []string, ok bool) {

	hostname, err := os.Hostname()
	if err != nil {
		fmt.Printf("Error(10020): Unable to get the hostname (%v)\n", err)
		os.Exit(1)
	}

	mdata := make(map[string]string, 30)
	mdata["HostName"] = hostname
	mdata["AppName"] = appName
	mdata["IS_WINDOWS"] = ""
	ps := string(os.PathSeparator)
	if ps != "/" {
		mdata["IS_WINDOWS"] = ""
	} else {
		mdata["IS_WINDOWS"] = "ms"
	}
	mdata["HOME"] = os.Getenv("HOME")
	mdata["FILENAMERAW"] = rawFileName
	mdata["FILENAME"] = RmExt(rawFileName)
	mdata["FILEEXT"] = filepath.Ext(rawFileName)
	if ps != "/" {
		ps = ps + ps
	}
	mdata["OS_SEP"] = ps

	ok = false
	tmplArr := []string{
		"%{CUR_PATH%}%{OS_SEP%}%{FILENAME%}-%{AppName%}-%{ModuleName%}-%{HostName%}%{FILEEXT%}",
		"%{CUR_PATH%}%{OS_SEP%}%{FILENAME%}-%{AppName%}-%{ModuleName%}%{FILEEXT%}",
		"%{CUR_PATH%}%{OS_SEP%}%{FILENAME%}-%{ModuleName%}-%{HostName%}%{FILEEXT%}",
		"%{CUR_PATH%}%{OS_SEP%}%{FILENAME%}-%{ModuleName%}%{FILEEXT%}",
		"%{CUR_PATH%}%{OS_SEP%}%{FILENAME%}-%{AppName%}-%{HostName%}%{FILEEXT%}",
		"%{CUR_PATH%}%{OS_SEP%}%{FILENAME%}-%{AppName%}%{FILEEXT%}",
		"%{CUR_PATH%}%{OS_SEP%}%{FILENAME%}-%{HostName%}%{FILEEXT%}",
		"%{CUR_PATH%}%{OS_SEP%}%{FILENAME%}%{FILEEXT%}",
		"%{CUR_PATH%}%{OS_SEP%}%{FILENAMERAW%}",
	}
	for _, p := range searchPath {
		mdata["CUR_PATH"] = p
		mdata["ModuleName"] = Basename(p)
		fmt.Printf("ModuleName: ->%s<- for %s, %s\n", Basename(p), p, godebug.LF())

		for _, tmpl := range tmplArr {
			aName := Qt(tmpl, mdata)
			aName, _ = SubstitueUserInFilePath(aName, mdata)
			aName = Qt(aName, mdata)
			fmt.Printf("aName: ->%s<- Checking to see if file exists, %s\n", aName, godebug.LF())
			if Exists(aName) {
				if !InArray(aName, fullFileName) {
					fullFileName = append(fullFileName, aName)
					ok = true
				}
			}
		}

	}
	return
}

// see: mdata["url"] = url.QueryEscape(xurl)
//func UrlEncoded(str string) string {
//	u, err := url.Parse(str)
//	if err != nil {
//		return str
//	}
//	return u.String()
//}

// xyzzy - should use umask, or 0640 for Winderz

var invalidMode = errors.New("Invalid Mode")

func Fopen(fn string, mode string) (file *os.File, err error) {
	file = nil
	if mode == "r" {
		file, err = os.Open(fn) // For read access.
	} else if mode == "w" {
		file, err = os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	} else if mode == "a" {
		file, err = os.OpenFile(fn, os.O_RDWR|os.O_APPEND, 0660)
		if err != nil {
			file, err = os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		}
	} else {
		err = invalidMode
	}
	return
}

func DbValueFromRow(rvdata []map[string]interface{}, row int, col string, def string) (rv string) {
	xCol, xOk := rvdata[row][col] //
	rv = def
	if xOk && xCol != nil {
		// rv = rvdata[row][col].(string)
		rv = xCol.(string)
	}
	return
}
func DbValueIsNull(rvdata []map[string]interface{}, row int, col string) (rv bool) {
	xCol, xOk := rvdata[row][col] //
	rv = true
	if xOk && xCol != nil {
		// rv = rvdata[row][col].(string)
		rv = false
	}
	return
}

// GetIPFromRemoteAddr get client ip address from request
//
// Example: ip,_,_ := net.SplitHostPort(r.RemoteAddr)
func GetIPFromRemoteAddr(RemoteAddr string) (rv string) {
	rv, _, _ = net.SplitHostPort(RemoteAddr)
	//n := strings.LastIndex(RemoteAddr, ":")
	//rv = RemoteAddr
	//if n > 0 && n < len(RemoteAddr) {
	//	rv = RemoteAddr[:n]
	//}
	return
}

// ===============================================================================================================================================
var ln *regexp.Regexp
var fi *regexp.Regexp
var cm *regexp.Regexp
var en *regexp.Regexp

func init() {
	ln = regexp.MustCompile("__LINE__")
	fi = regexp.MustCompile("__FILE__")
	en = regexp.MustCompile("__ENV__:[a-zA-Z][a-zA-Z_0-9]*")
	cm = regexp.MustCompile("////.*$")
}

// ReadJSONDataWithComments read in the file and handle __LINE__, __FILE__ and comments starting with 4 slashes.
func ReadJSONDataWithComments(path string) (file []byte, err error) {
	file, err = ioutil.ReadFile(path)
	if err != nil {
		// fmt.Printf("Error(10014): Error Reading/Opening %v, %s, Config File:%s\n", err, godebug.LF(), path)
		// fmt.Fprintf(os.Stderr, "%sError(10014): Error Reading/Opening %v, %s, Config File:%s%s\n", MiscLib.ColorRed, err, godebug.LF(), path, MiscLib.ColorReset)
		return
	}

	data := strings.Replace(string(file), "\t", " ", -1)
	lines := strings.Split(data, "\n")
	//ln := regexp.MustCompile("__LINE__")
	//fi := regexp.MustCompile("__FILE__")
	//cm := regexp.MustCompile("//.*$")
	for lineNo, aLine := range lines {
		aLine = ln.ReplaceAllString(aLine, fmt.Sprintf("%d", lineNo+1))
		aLine = fi.ReplaceAllString(aLine, path)
		aLine = cm.ReplaceAllString(aLine, "")
		if en.MatchString(aLine) { // pick up and replace environment variables - put passwords in env not in config files
			fmt.Printf("matched __ENV__:Name, %s\n", godebug.LF())
			ss := en.FindAllString(aLine, 1)
			// fmt.Printf("ss = %s\n", ss)
			s := ss[0] // the matched, no need to check array because inside MatchString already
			// fmt.Printf("s(raw) = %s\n", s)
			s = s[8:] // remove __ENV__:
			// fmt.Printf("env name = [%s]\n", s)
			v := os.Getenv(s)
			// fmt.Printf("v = [%s]\n", v)
			if v == "" {
				fmt.Fprintf(os.Stderr, "Fatal: Invalid environemtn variable setting: %s - returned empty string - not allowed\n", s)
				os.Exit(1)
			}
			aLine = en.ReplaceAllString(aLine, v)
			fmt.Printf("final line = [%s]\n", aLine)
		}
		lines[lineNo] = aLine
	}
	file = []byte(strings.Join(lines, "\n"))

	// fmt.Printf("Results >%s<\n", file)

	return file, nil
}

// JsonStringToData convert from a string to a map[string]interface{} - parse JSON
func JSONStringToData(s string) (theJSON map[string]interface{}, err error) {
	err = json.Unmarshal([]byte(s), &theJSON)
	if err != nil {
		theJSON = make(map[string]interface{})
	}
	return
}

/*
Renamed Functions:
	dir.go|244 col 6| func FnToCssClass should be FnToCSSClass
	dir.go|314 col 6| func ReadJsonPath should be ReadJSONPath
	dir.go|361 col 6| func readJsonFile should be readJSONFile
	dir.go|386 col 6| func UriToStringMap should be URIToStringMap
	dir.go|1020 col 6| func GetIpFromRemoteAddr should be GetIPFromRemoteAddr
	dir.go|1069 col 6| func JsonStringToData should be JSONStringToData
*/

/* vim: set noai ts=4 sw=4: */
