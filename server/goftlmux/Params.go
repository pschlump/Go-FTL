package goftlmux

//
// Go Go Mux - Go Fast Mux / Router for HTTP requests
//
// (C) Philip Schlump, 2013-2018. All rights reserved.
//

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/pschlump/Go-FTL/server/common"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/json" //	"encoding/json"
)

type FromType int
type ParamType uint8

const (
	FromURL FromType = iota
	FromParams
	FromCookie
	FromBody
	FromBodyJson
	FromInject
	FromHeader
	FromOther
	FromDefault
	FromAuth
	FromSession
)

const MaxParams = 200

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Name     string
	Value    string
	From     FromType
	Type     ParamType
	UsedAt   int
	UsedFile string
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params struct {
	NParam       int              //
	Data         []Param          // has to be assided to array
	route_i      int              // What matched
	search       map[string]int   // has to be allocated
	search_ready bool             //
	allParam     [MaxParams]Param // The parameters for the current operation
	// parent       *MuxRouter     // // PJS Sun Nov 15 13:12:31 MST 2015
}

func InitParams(p *Params) {
	p.Data = p.allParam[:]
	p.search = make(map[string]int)
	p.NParam = 0
}

func FromTypeToString(ff FromType) string {
	switch ff {
	case FromURL:
		return "FromURL"
	case FromParams:
		return "FromParams"
	case FromCookie:
		return "FromCookie"
	case FromBody:
		return "FromBody"
	case FromBodyJson:
		return "FromBodyJson"
	case FromInject:
		return "FromInject"
	case FromHeader:
		return "FromHeader"
	case FromOther:
		return "FromOther"
	case FromDefault:
		return "FromDefault"
	case FromAuth:
		return "FromAuth"
	case FromSession:
		return "FromSession"
	default:
		return "Unk-FromType"
	}
}

func (ff FromType) String() string {
	return FromTypeToString(ff)
}

func (ff ParamType) String() string {
	switch ff {
	case 'i': // Injected Parameter
		return "-inject-"
	case 'I': // Injected Session Value Parameter
		return "-session-"
	case 's':
		return "pt:s"
	case 'J':
		return "pt:J"
	case 'a':
		return "pt:a"
	case 'b': // From the body
		return "-body-"
	case 'q': // From the URL/Query string
		return "-qry-"
	case 'r': // Re-maped parameter in TabServer2 - name changed from oritinal see: function RemapParams
		return "-rnm-"
	case ':': // URL name like /user/:username/table/:tablename - from the URL -- Created in gogomux, or goftlxmux
		return "-URL-"
	case 'c': // From a cookie
		return "-cookie-"
	case 'e': // Generated in AesSrp encryption by decrypting the body or URL
		return "-encrypted-"
	}
	return fmt.Sprintf("p?:%s", string(rune(ff)))
}

func (ps *Params) MakeStringMap(mdata map[string]string) {
	for _, v := range ps.Data {
		mdata[v.Name] = v.Value
	}
}

func (ps *Params) CreateSearch() {
	// fmt.Printf("CreateSearch - called\n")
	if ps.search_ready {
		return
	}

	for i, v := range ps.Data {
		ps.search[v.Name] = i
	}
	// fmt.Printf("CreateSearch - set to true\n")
	ps.search_ready = true
}

func (ps *Params) DumpParam() (rv string) {
	var Data2 []Param
	Data2 = ps.Data[0:ps.NParam]
	rv = godebug.SVar(Data2)
	return
}

func (ps *Params) DumpParamDB() (rv string) {
	var Data2 []Param
	Data2 = ps.Data[0:ps.NParam]
	rv = godebug.SVarI(Data2)
	return
}

func (ps *Params) DumpParamTable() (rv string) {
	rv = "\n"
	rv += fmt.Sprintf(" %34s  %-12s %11s %6s %s\n", "Name", "From", "Type", "UsedAt", "Value")
	rv += fmt.Sprintf(" %-35s %-12s %-11s %6s %s\n", "----------------------------------", "------------", "-----------", "------", "-----------------------------------------")
	for _, vv := range ps.Data[0:ps.NParam] {
		rv += fmt.Sprintf("%35s %12s %12s %-6d %s\n", vv.Name, vv.From, vv.Type, vv.UsedAt, vv.Value)
	}
	rv += "\n"
	return
}

type UnusedParam struct {
	Match string
	IsRe  bool
	re    *regexp.Regexp
}

func SetupUnsedParam(NormalUnused []UnusedParam) {
	for ii, pp := range NormalUnused {
		if pp.IsRe {
			re, err := regexp.Compile(pp.Match)
			pp.re = re
			if err != nil {
				fmt.Fprintf(os.Stderr, "%sError: invalid regular expression for unused paremeters at [%d] in set -->>%s<<--, %s, %s%s\n", MiscLib.ColorRed, ii, pp.Match, err, godebug.LF(), MiscLib.ColorReset)
			}
			NormalUnused[ii] = pp
		}
	}
}

func (ps *Params) IsMatchUnused(NormalUnused []UnusedParam, aName string) bool {
	for _, xx := range NormalUnused {
		if xx.IsRe {
			if xx.re.MatchString(aName) {
				return true
			}
		} else {
			if xx.Match == aName {
				return true
			}
		}
	}
	return false
}

func (ps *Params) ReportUnexpectedUnused(NormalUnused []UnusedParam) {
	for _, vv := range ps.Data[0:ps.NParam] {
		if vv.UsedAt == 0 {
			if !ps.IsMatchUnused(NormalUnused, vv.Name) {
				fmt.Fprintf(os.Stderr, "%sUnsued parameter %s%s\n", MiscLib.ColorRed, vv.Name, MiscLib.ColorReset)
			}
		}
	}
}

func (ps *Params) DumpParamUsed(ign ...string) (rv string) {
	rv = "\n"
	rv += fmt.Sprintf(" %34s  %6s %s\n", "Name", "Line", "File Name")
	rv += fmt.Sprintf(" %-35s %6s %s\n", "----------------------------------", "-------", "-----------------------------------------")
	for _, vv := range ps.Data[0:ps.NParam] {
		if vv.UsedAt == 0 && !sizlib.InArray(vv.Name, ign) {
			rv += fmt.Sprintf("%35s  ***not used***\n", vv.Name)
		} else {
			rv += fmt.Sprintf("%35s  %7d %s\n", vv.Name, vv.UsedAt, vv.UsedFile)
		}
	}
	rv += "\n"
	return
}

func (ps *Params) DumpParamNVF() (rv []common.NameValueFrom) {
	for _, vv := range ps.Data[0:ps.NParam] {
		rv = append(rv, common.NameValueFrom{
			Name:      vv.Name,
			Value:     vv.Value,
			From:      vv.From.String(),
			ParamType: vv.Type.String(),
		})
		// rv += fmt.Sprintf("%35s %12s %12s %s\n", vv.Name, vv.From, vv.Type, vv.Value)
	}
	return
}

// ByName returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (ps *Params) ByName(name string) (rv string) {
	rv = ""
	// xyzzy100 Change this to use a map[string]int - build maps on setup.
	// fmt.Printf("Looking For: %s, ps = %s\n", name, godebug.SVarI(ps.Data[0:ps.NParam]))
	// fmt.Printf("ByName ------------------------\n")
	if ps.search_ready {
		// fmt.Printf("Is True ------------------------\n")
		if i, ok := ps.search[name]; ok {
			nn, ff := godebug.LINEnf(2)
			ps.Data[i].UsedAt = nn
			ps.Data[i].UsedFile = ff
			rv = ps.Data[i].Value
		}
		return
	}

	for i := 0; i < ps.NParam; i++ {
		if ps.Data[i].Name == name {
			nn, ff := godebug.LINEnf(2)
			if ps.Data[i].UsedAt != 0 {
				fmt.Printf("ByName - overwrite: Line:%d File:%s\n", ps.Data[i].UsedAt, ps.Data[i].UsedFile)
			}
			ps.Data[i].UsedAt = nn
			ps.Data[i].UsedFile = ff
			rv = ps.Data[i].Value
			return
		}
	}
	return
}

func (ps *Params) GetByName(name string) (rv string, found bool) {
	rv = ""
	found = false
	// xyzzy100 Change this to use a map[string]int - build maps on setup.
	// fmt.Printf("Looking For: %s, ps = %s\n", name, godebug.SVarI(ps.Data[0:ps.NParam]))
	// fmt.Printf("ByName ------------------------\n")
	if ps.search_ready {
		// fmt.Printf("Is True ------------------------\n")
		if i, ok := ps.search[name]; ok {
			nn, ff := godebug.LINEnf(2)
			ps.Data[i].UsedAt = nn
			ps.Data[i].UsedFile = ff
			rv = ps.Data[i].Value
			found = true
		}
		return
	}

	for i := 0; i < ps.NParam; i++ {
		if ps.Data[i].Name == name {
			nn, ff := godebug.LINEnf(2)
			ps.Data[i].UsedAt = nn
			ps.Data[i].UsedFile = ff
			rv = ps.Data[i].Value
			found = true
			return
		}
	}
	return
}

func (ps *Params) GetByNameAndType(name string, ft FromType) (rv string, found bool) {
	rv = ""
	found = false
	// xyzzy100 Change this to use a map[string]int - build maps on setup.
	// fmt.Printf("Looking For: %s, ps = %s\n", name, godebug.SVarI(ps.Data[0:ps.NParam]))
	// fmt.Printf("ByName ------------------------\n")
	if ps.search_ready {
		// fmt.Printf("Is True ------------------------\n")
		if i, ok := ps.search[name]; ok {
			trv := ps.Data[i].Value
			if ps.Data[i].From == ft {
				nn, ff := godebug.LINEnf(2)
				ps.Data[i].UsedAt = nn
				ps.Data[i].UsedFile = ff
				rv = trv
				found = true
			}
		}
		return
	}

	for i := 0; i < ps.NParam; i++ {
		if ps.Data[i].Name == name {
			trv := ps.Data[i].Value
			if ps.Data[i].From == ft {
				nn, ff := godebug.LINEnf(2)
				ps.Data[i].UsedAt = nn
				ps.Data[i].UsedFile = ff
				rv = trv
				found = true
			}
			return
		}
	}
	return
}

func (ps *Params) ByNameDflt(name string, dflt string) (rv string) {

	rv = dflt
	if ps.search_ready {
		// fmt.Printf("Is True ------------------------\n")
		if i, ok := ps.search[name]; ok {
			nn, ff := godebug.LINEnf(2)
			ps.Data[i].UsedAt = nn
			ps.Data[i].UsedFile = ff
			rv = ps.Data[i].Value
		}
		return
	}

	for i := 0; i < ps.NParam; i++ {
		if ps.Data[i].Name == name {
			nn, ff := godebug.LINEnf(2)
			ps.Data[i].UsedAt = nn
			ps.Data[i].UsedFile = ff
			rv = ps.Data[i].Value
			return
		}
	}
	// fmt.Printf("MoD, %s\n", godebug.LF())
	return
}

func (ps *Params) HasName(name string) (rv bool) {
	rv = false
	if ps.search_ready {
		if _, ok := ps.search[name]; ok {
			rv = true
		}
		return
	}
	for i := 0; i < ps.NParam; i++ {
		if ps.Data[i].Name == name {
			rv = true
			return
		}
	}
	return
}

func (ps *Params) SetValue(name string, val string) {
	x := ps.PositionOf(name)
	if x >= 0 {
		ps.Data[x].Value = val
	}
}

//func (ps *Params) SetValueType(name string, ty FromType, val string) {
//	x := ps.PositionOf(name)
//	if x >= 0 {
//		ps.Data[x].Value = val
//	}
//}

func (ps *Params) PositionOf(name string) (rv int) {
	rv = -1
	for i := 0; i < ps.NParam; i++ {
		if ps.Data[i].Name == name {
			rv = i
			return
		}
	}
	return
}

func (ps *Params) GetAllParam(skip ...string) (rv []common.NameValueFrom) {
	for _, vv := range ps.Data[0:ps.NParam] {
		if !sizlib.InArray(vv.Name, skip) {
			rv = append(rv, common.NameValueFrom{
				Name:      vv.Name,
				Value:     vv.Value,
				From:      vv.From.String(),
				ParamType: vv.Type.String(),
			})
		}
	}
	return
}

func (ps *Params) ByPostion(pos int) (name string, val string, outRange bool) {
	// xyzzy101 Change this to use a map[string]int - build maps on setup.
	if pos >= 0 && pos < ps.NParam {
		return ps.Data[pos].Name, ps.Data[pos].Value, false
	}
	return "", "", true
}

// -------------------------------------------------------------------------------------------------
func AddValueToParams(Name string, Value string, Type ParamType, From FromType, ps *Params) (k int) {
	ps.search_ready = false
	j := ps.PositionOf(Name)
	k = ps.NParam
	// db("AddValueToParams","j=%d k=%d %s\n", j, k, godebug.LF())
	if j >= 0 {
		ps.Data[j].Value = Value
		ps.Data[j].Type = Type
		ps.Data[j].From = From
	} else {
		ps.Data[k].Value = Value
		ps.Data[k].Name = Name
		ps.Data[k].Type = Type
		ps.Data[k].From = From
		k++
		// xyzzy - check for more than MaxParams
	}
	ps.NParam = k
	// db(A"AddValueToParams","At end: NParam=%d %s\n", ps.NParam, godebug.SVar(ps.Data[0:ps.NParam]))
	return
}

// -------------------------------------------------------------------------------------------------
// func ParseBodyAsParams(w *MyResponseWriter, req *http.Request, ps *Params) int {
func ParseBodyAsParams(www *MidBuffer, req *http.Request, ps *Params) int {

	ct := req.Header.Get("Content-Type")
	if db9 {
		fmt.Printf("*************************************************************************** content type \n")
		fmt.Printf("content-type: %s, %s\n", ct, godebug.LF())
		fmt.Printf("*************************************************************************** content type \n")
	}
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" || req.Method == "DELETE" {
		fmt.Printf("AT %s\n", godebug.LF())
		if req.PostForm == nil {
			fmt.Printf("AT %s\n", godebug.LF())
			if strings.HasPrefix(ct, "application/json") {
				fmt.Printf("AT %s\n", godebug.LF())
				buf, err2 := ioutil.ReadAll(req.Body)
				if err2 != nil {
					fmt.Printf("Error(20008): Malformed body, RequestURI=%s err=%v\n", req.RequestURI, err2)
				}
				rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
				body, err2 := ioutil.ReadAll(req.Body)
				if err2 != nil {
					fmt.Printf("Error(20008): Malformed JSON body, RequestURI=%s err=%v\n", req.RequestURI, err2)
				}
				req.Body = rdr2
				fmt.Printf("THIS ONE                                           !!!!!!!!!!!!!!! body >%s< AT %s\n", body, godebug.LF())
				var jsonData map[string]interface{}
				err := json.Unmarshal(body, &jsonData)
				if err == nil {
					for Name, v := range jsonData {
						Value := ""
						switch v.(type) {
						case bool:
							Value = fmt.Sprintf("%v", v)
						case float64:
							Value = fmt.Sprintf("%v", v)
						case int64:
							Value = fmt.Sprintf("%v", v)
						case int32:
							Value = fmt.Sprintf("%v", v)
						case time.Time:
							Value = fmt.Sprintf("%v", v)
						case string:
							Value = fmt.Sprintf("%v", v)
						default:
							Value = fmt.Sprintf("%s", godebug.SVar(v))
						}
						AddValueToParams(Name, Value, 'b', FromBodyJson, ps)
					}
				} else {
					fmt.Printf("Error: in parsing JSON data >%s< Error: %s\n", body, err)
				}
			} else {
				fmt.Printf("AT %s\n", godebug.LF())
				err := req.ParseForm()
				if err != nil {
					fmt.Printf("Error(20010): Malformed body, RequestURI=%s err=%v\n", req.RequestURI, err)
				} else {
					for Name, v := range req.PostForm {
						if len(v) > 0 {
							AddValueToParams(Name, v[0], 'b', FromBody, ps)
						}
					}
				}
			}
		} else {
			fmt.Printf("AT %s\n", godebug.LF())
			for Name, v := range req.PostForm {
				if len(v) > 0 {
					AddValueToParams(Name, v[0], 'b', FromBody, ps)
				}
			}
		}
	}
	return 0
}

// -------------------------------------------------------------------------------------------------
func ParseBodyAsParamsReg(www http.ResponseWriter, req *http.Request, ps *Params) int {

	ct := req.Header.Get("Content-Type")
	if db4 {
		fmt.Printf("*************************************************************************** content type \n")
		fmt.Printf("content-type: %s, %s\n", ct, godebug.LF())
		fmt.Printf("*************************************************************************** content type \n")
	}
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" || req.Method == "DELETE" {
		if db4 {
			fmt.Printf("AT %s\n", godebug.LF())
		}
		if req.PostForm == nil {
			if db4 {
				fmt.Printf("AT %s\n", godebug.LF())
			}
			if strings.HasPrefix(ct, "application/json") {
				if db4 {
					fmt.Printf("AT %s\n", godebug.LF())
				}
				body := PeekAtBody(req)
				var jsonData map[string]interface{}
				err := json.Unmarshal(body, &jsonData)
				if db4 {
					fmt.Printf("AT %s\n", godebug.LF())
				}
				if err == nil {
					for Name, v := range jsonData {
						Value := ""
						switch v.(type) {
						case bool:
							Value = fmt.Sprintf("%v", v)
						case float64:
							Value = fmt.Sprintf("%v", v)
						case int64:
							Value = fmt.Sprintf("%v", v)
						case int32:
							Value = fmt.Sprintf("%v", v)
						case time.Time:
							Value = fmt.Sprintf("%v", v)
						case string:
							Value = fmt.Sprintf("%v", v)
						default:
							Value = fmt.Sprintf("%s", godebug.SVar(v))
						}
						AddValueToParams(Name, Value, 'b', FromBodyJson, ps)
					}
				} else {
					fmt.Printf("Error: in parsing JSON data >%s< Error: %s, %s\n", body, err, godebug.LF())
				}
				if db5 {
					fmt.Printf("Params Are: %s AT %s\n", ps.DumpParamDB(), godebug.LF())
				}
			} else {
				if db4 {
					fmt.Printf("AT %s\n", godebug.LF())
				}
				buf := PeekAtBody(req)
				err := req.ParseForm()
				if err != nil {
					fmt.Printf("Error(20010): Malformed body, RequestURI=%s err=%v\n", req.RequestURI, err)
				} else {
					for Name, v := range req.PostForm {
						if len(v) > 0 {
							AddValueToParams(Name, v[0], 'b', FromBody, ps)
						}
					}
				}
				req.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
				if db4 {
					fmt.Printf("AT %s\n", godebug.LF())
				}
			}
		} else {
			if db4 {
				fmt.Printf("AT %s\n", godebug.LF())
			}
			for Name, v := range req.PostForm {
				if len(v) > 0 {
					AddValueToParams(Name, v[0], 'b', FromBody, ps)
				}
			}
		}
	}
	return 0
}

func PeekAtBody(req *http.Request) []byte {
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" || req.Method == "DELETE" {
		bodyBytes, _ := ioutil.ReadAll(req.Body)
		req.Body.Close() //  must close
		if db44 {
			if len(bodyBytes) == 0 {
				fmt.Printf("%sLen Now 0 ! AT %s, %s, %s%s\n", MiscLib.ColorRed, godebug.LF(), godebug.LF(2), godebug.LF(3), MiscLib.ColorReset)
				fmt.Fprintf(os.Stdout, "%sLen Now 0 ! AT %s, %s, %s%s\n", MiscLib.ColorRed, godebug.LF(), godebug.LF(2), godebug.LF(3), MiscLib.ColorReset)
			}
		}
		if db45 {
			if len(bodyBytes) == 0 {
				fmt.Printf("%sBody ->%s<- AT: %s, %s%s\n", MiscLib.ColorRed, bodyBytes, godebug.LF(2), godebug.LF(3), MiscLib.ColorReset)
				fmt.Fprintf(os.Stdout, "%sBody ->%s<- AT: %s, %s%s\n", MiscLib.ColorRed, bodyBytes, godebug.LF(2), godebug.LF(3), MiscLib.ColorReset)
			} else {
				fmt.Printf("%sBody ->%s<- AT: %s, %s%s\n", MiscLib.ColorGreen, bodyBytes, godebug.LF(2), godebug.LF(3), MiscLib.ColorReset)
				fmt.Fprintf(os.Stdout, "%sBody ->%s<- AT: %s, %s%s\n", MiscLib.ColorGreen, bodyBytes, godebug.LF(2), godebug.LF(3), MiscLib.ColorReset)
			}
		}
		req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		return bodyBytes
	}
	return []byte{}
}

// -------------------------------------------------------------------------------------------------
func ParseCookiesAsParams(www *MidBuffer, req *http.Request, ps *Params) int {

	Ck := req.Cookies()
	for _, v := range Ck {
		AddValueToParams(v.Name, v.Value, 'c', FromCookie, ps)
	}
	return 0
}

// -------------------------------------------------------------------------------------------------
func ParseCookiesAsParamsReg(www http.ResponseWriter, req *http.Request, ps *Params) int {

	Ck := req.Cookies()
	for _, v := range Ck {
		AddValueToParams(v.Name, v.Value, 'c', FromCookie, ps)
	}
	return 0
}

// -------------------------------------------------------------------------------------------------
var ApacheLogFile *os.File

const ApacheFormatPattern = "%s %v %s %s %s %v %d %v\n"

const benchmar = false

/*
func itoaPos(n int, buffer *bytes.Buffer, padLen int, pad uint8) {
	i := 0
	var s [10]uint8
	for {
		s[i] = uint8((n % 10) + '0')
		i++
		n /= 10
		if n == 0 {
			break
		}
	}
	for ; i < padLen; i++ {
		s[i] = pad
	}

	for j := i - 1; j >= 0; j-- {
		buffer.WriteByte(s[j])
	}
}
*/
var mutexCurTime = &sync.Mutex{}
var timeFormatted string

func SetCurTime(s string) {
	mutexCurTime.Lock()
	timeFormatted = s
	mutexCurTime.Unlock()
}

func GetCurTime() (s string) {
	mutexCurTime.Lock()
	s = timeFormatted
	mutexCurTime.Unlock()
	return
}

func init() {
	// Once a sec update of time formated string
	onceASec := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-onceASec.C:
				// do stuff
				finishTime := time.Now()
				curTimeUTC := finishTime.UTC()
				SetCurTime(curTimeUTC.Format("02/Jan/2006 03:04:05"))
				// fmt.Printf("cur-time: %s\n", GetCurTime())
			case <-quit:
				onceASec.Stop()
				return
			}
		}
	}()
}

func ApacheLogingBefore(www *MidBuffer, req *http.Request, ps *Params) int {
	if ApacheLogFile == nil {
		return 0
	}
	www.StartTime = time.Now()
	return 0
}

func ApacheLogingAfter(www *MidBuffer, req *http.Request, ps *Params) int {
	if ApacheLogFile == nil {
		return 0
	}
	ip := req.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}

	finishTime := time.Now()
	elapsedTime := finishTime.Sub(www.StartTime)
	// The next line is a real problem.  Taking 450us to convert a data is just ICKY!  I have a new
	// version of Format that reduces this to about 300us.  What is needed is a real fast formatting
	// tool for dates that reduces this to a reasonable 30us.
	// Sadly enough - by not exposing the interals of the time.Time type - fixing this will require
	// a major re-write of the entire time type.   That is oging to take some days to do.
	var timeFormatted string

	//
	// OLD: finishTimeUTC := finishTime.UTC()
	// OLD: timeFormatted = finishTimeUTC.Format("02/Jan/2006 03:04:05") // 450+ us to do a time format and 1 alloc
	//
	// This entire thing could be replaced with a goroutein that runs 1ce a second, and makes a
	// variable with the date-time in it.  that would be one "Format" per second on a different
	// thread.  Then just use a lock, unlock process to access the variable - simple enough.
	//
	timeFormatted = GetCurTime()

	fmt.Fprintf(ApacheLogFile, ApacheFormatPattern, ip, timeFormatted, req.Method, req.RequestURI, req.Proto, www.StatusCode, www.Length, elapsedTime.Seconds())

	return 0
}

// -------------------------------------------------------------------------------------------------
func ParseQueryParams(www *MidBuffer, req *http.Request, ps *Params) int {
	// u, err := url.ParseRequestURI(req.RequestURI)
	if req.URL.RawQuery == "" {
		return 0
	}
	m, err := url.ParseQuery(req.URL.RawQuery)
	// db("ParseQueryParams","Parsing Raw Query ->%s<-, m=%s\n", req.URL.RawQuery, godebug.SVar(m))
	if err != nil {
		fmt.Printf("Unable to parse URL query, %s\n", err)
	}
	for Name, v := range m {
		vv := ""
		if len(v) == 1 {
			vv = v[0]
		} else {
			vv = godebug.SVar(v)
		}
		AddValueToParams(Name, vv, 'q', FromParams, ps)
	}
	return 0
}

func ParseQueryParamsReg(www http.ResponseWriter, req *http.Request, ps *Params) int {
	// u, err := url.ParseRequestURI(req.RequestURI)
	if req.URL.RawQuery == "" {
		return 0
	}
	m, err := url.ParseQuery(req.URL.RawQuery)
	// db("ParseQueryParams","Parsing Raw Query ->%s<-, m=%s\n", req.URL.RawQuery, godebug.SVar(m))
	if err != nil {
		fmt.Printf("Unable to parse URL query, %s\n", err)
	}
	for Name, v := range m {
		vv := ""
		if len(v) == 1 {
			vv = v[0]
		} else {
			vv = godebug.SVar(v)
		}
		AddValueToParams(Name, vv, 'q', FromParams, ps)
	}
	return 0
}

// -------------------------------------------------------------------------------------------------
func PrefixWith(www *MidBuffer, req *http.Request, ps *Params) int {
	// Prefix for AngularJS
	www.Write([]byte(")]}"))
	// Other Common Prefixes are:
	//		while(1);
	//		for(;;);
	//		//							Comment
	//		while(true);
	return 0
}

// -------------------------------------------------------------------------------------------------
func MethodParam(www *MidBuffer, req *http.Request, ps *Params) int {
	// fmt.Printf("MethodParam, Params Are: %s, %s\n", ps.DumpParam(), godebug.LF())
	// fmt.Printf("%s\n", godebug.LF())
	if ps.HasName("METHOD") {
		x := ps.ByName("METHOD")
		// fmt.Printf("x=%s %s\n", x, godebug.LF())
		if b, ok := validMethod[x]; ok && b {
			// fmt.Printf("A Valid Method: b=%v ok=%v %s\n", ok, b, godebug.LF())
			req.Method = x
		}
	}
	return 0
}

func MethodParamReg(www http.ResponseWriter, req *http.Request, ps *Params) int {
	//fmt.Printf("MethodParam\n")
	//fmt.Printf("%s\n", godebug.LF())
	if ps.HasName("METHOD") {
		//fmt.Printf("%s\n", godebug.LF())
		x := ps.ByName("METHOD")
		if b, ok := validMethod[x]; ok && b {
			//fmt.Printf("%s\n", godebug.LF())
			req.Method = x
		}
	}
	return 0
}

func RenameReservedItems(www http.ResponseWriter, req *http.Request, ps *Params, ri map[string]bool) {
	for i := 0; i < ps.NParam; i++ {
		if ri[ps.Data[i].Name] {
			ps.Data[i].Name = "user_param::" + ps.Data[i].Name
		}
	}
}

/*
	HTTP/1.1 401 Unauthorized
	{
		"status": "Error"
		, "msg": "No access token provided."
		, "code": "10002"
		, "details": "bla bla bla"
	}
req.Header.Add("If-None-Match", `W/"wyzzy"`)
https://developer.github.com/guides/traversing-with-pagination/

req.Header.Add("Link", `W/"wyzzy"`)
Link: <https://api.github.com/search/code?q=addClass+user%3Amozilla&page=2>; rel="next",
  <https://api.github.com/search/code?q=addClass+user%3Amozilla&page=34>; rel="last"
*/

const db4 = false // Parse Body
const db5 = false // Dump params to log (stdout) in human format.
const db9 = false
const db44 = false
const db45 = false

/* vim: set noai ts=4 sw=4: */
