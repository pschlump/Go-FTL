package tr

// Copyright (C) Philip Schlump, 2013-2015.
/*

1. See if we can get pub/sub to publish a Start/End message on URI/SQL
2. See if we can build our "trace.go" package from "tab-server1.go" - w/o trace in it.
3. See if we can build a simple display tool
4. See if we can get monitoring to work again


Design:

	1. Store data in Redis
		"rest:trace:key" - a counter that indicates the current set for operations
	2. Store by set
		"rest:trace:uri:Key1" - do get/set on it and grow the set until "end":"yes" added or "end":"fail"
	3. Have a pub/sub for current events [ new "start":time, new "end":??? ]
	4. filter by user-name - set of filters on the server side
		You subscribe to a listener at port:8111
		You can get the current request or walk backward in time
		You can set the "keep" time
		You can turn on/off tracking
		"rest:trace:cfg" - config key

*/
import (
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/pschlump/Go-FTL/server/common"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/json" //	"encoding/json"
	"github.com/pschlump/radix.v2/pool"
	"github.com/pschlump/uuid"
)

// "github.com/garyburd/redigo/redis"

// ------------------------------------------------------------------------------------------------------------------------------------------
const ISO8601 = "2006-01-02T15:04:05.99999Z07:00"

var g_key_no int64 = 1000 // temporary, should be fetched from Redis as "rest:trace:key"
var g_seq_no int = 2000   // temporary, should be fetched from Redis as "rest:trace:key"

const default_ttl uint64 = (60 * 30) // in seconds, 60 seconds to min, 30 min

var ttl uint64 = default_ttl
var Ttl uint64 = default_ttl
var trace_on bool = false
var trace_debug bool = true
var stdout_on bool = false

type AvgTime struct {
	Avg float64
	N   int64
	Min float64
	Max float64
}

func (this *Trx) redisSetElapsedTime(theKey string, uri string, is_uri bool, elapsedTime time.Duration) {
	conn, err := this.redisPool.Get()
	if err != nil {
		return
	}
	defer this.redisPool.Put(conn)

	var l_theKey string
	if is_uri {
		x := strings.Split(uri, "?")
		l_theKey = theKey + x[0]
	} else {
		l_theKey = theKey + uri
	}
	dm := AvgTime{0.0, 0, 9999999999.0, 0.0}
	var dt float64
	dt = float64(elapsedTime) / 1000000.0
	//	,"/api/test/echo_request": {"Avg":3,"N":8,"Min":2,"Max":4}
	s, err := conn.Cmd("GET", l_theKey).Str()
	if err == nil {
		err = json.Unmarshal([]byte(s), &dm)
		if err != nil {
			dm.Avg = 0
			dm.N = 0
			dm.Min = 9999999999.0
			dm.Max = 0.0
		}
	}
	dm.Avg = ((dm.Avg * float64(dm.N)) + dt) / float64(dm.N+1)
	if dm.Min > dt {
		dm.Min = dt
	}
	if dm.Max < dt {
		dm.Max = dt
	}
	dm.N++
	s2 := SVar(dm)
	conn.Cmd("SET", l_theKey, s2)
}

func (this *Trx) TraceGetConfig() (uint64, bool) {
	return ttl, trace_on
}

func (this *Trx) TraceSetConfig(p_ttl uint64, p_trace_on bool, p_stdout_on bool) { // xyzzy - turn on/off stdout_on also
	conn, err := this.redisPool.Get()
	if err != nil {
		return
	}
	defer this.redisPool.Put(conn)

	ttl, trace_on, stdout_on = p_ttl, p_trace_on, p_stdout_on
	theKey := "rest:trace:cfg"
	theData := fmt.Sprintf(`{"ttl":%v,"trace_on":%v,"stdout_on":%v}`, ttl, trace_on, p_stdout_on)
	// fmt.Printf ( "set %s %s; %s\n", theKey, theData, LF() )
	err = conn.Cmd("SET", theKey, theData).Err
	if err != nil {
		fmt.Printf("Error with redis: %v\n", err)
	}
}

//func (this *Trx) SetRedisConn(r *pool.Pool) {

func (this *Trx) TraceInitConfig(r *pool.Pool) {
	this.redisPool = r

	conn, err := this.redisPool.Get()
	if err != nil {
		return
	}
	defer this.redisPool.Put(conn)

	theKey := "rest:trace:cfg"
	s, err := conn.Cmd("GET", theKey).Str()
	// fmt.Printf ( "get %s; %s; %v %v\n", theKey, LF(), s, err )
	if err != nil {
		ttl = default_ttl
		trace_on = true
		theData := fmt.Sprintf(`{"ttl":%v,"trace_on":%v,"stdout_on":%v}`, ttl, trace_on, stdout_on)
		// fmt.Printf ( "set %s %s; %s\n", theKey, theData, LF() )
		conn.Cmd("SET", theKey, theData)
	} else {
		var jsonData map[string]interface{}
		err = json.Unmarshal([]byte(s), &jsonData)
		if err == nil {
			ttl = uint64(jsonData["ttl"].(float64))
			trace_on = jsonData["trace_on"].(bool)
			// xyzzy stdout_on??
		} else {
			ttl = default_ttl
			trace_on = true
			// xyzzy stdout_on??
		}
	}
}

// ------------------------------------------------------------------------------------------------------------------------------------------
func SVar(v interface{}) string {
	s, err := json.Marshal(v)
	// s, err := json.MarshalIndent ( v, "", "\t" )
	if err != nil {
		return fmt.Sprintf("Error:%s", err)
	} else {
		return string(s)
	}
}

// ----------------------------------------------------------------------------------------------------------
type AnItemType int

const (
	Fx   AnItemType = iota
	User            // Implemented at the D.B. & Fucntion level
	URI
	Table
	Db
	Redis
)

// ----------------------------------------------------------------------------------------------------------
type OnOffIgnore int

const (
	On OnOffIgnore = iota
	Off
	Ignore
)

// ----------------------------------------------------------------------------------------------------------
type TraceItem struct {
	ItemType AnItemType  // func, user, URI, table
	OnOff    OnOffIgnore // Is item T=included or F=excluded,  i=ignored
	Data1    string
	Data2    string
}

// ----------------------------------------------------------------------------------------------------------
type TraceConfig struct {
	TraceFuncs map[int]TraceItem
}

// ----------------------------------------------------------------------------------------------------------
func TraceFunc(name string, args ...interface{}) {
}

// ----------------------------------------------------------------------------------------------------------
func TraceFuncExit(name string) {
}

// ----------------------------------------------------------------------------------------------------------
func (this *Trx) getRedisKeyInt64(theKey string, dflt int64) int64 {
	conn, err := this.redisPool.Get()
	if err != nil {
		return 0
	}
	defer this.redisPool.Put(conn)

	// ss, err := conn.Cmd("GET", theKey).Str()
	s, err := conn.Cmd("INCR", theKey).Int64()
	if err != nil {
		// fmt.Printf("Error(11006) error getting key, returning empty interface for:%s, err=%s\n", theKey, err)
		conn.Cmd("SET", theKey, dflt)
		return dflt
	}
	//s, err := strconv.ParseInt(ss, 10, 64)
	//if err != nil {
	//	fmt.Printf("Error(11006) error getting key, should be int, got:%s for %s, err=%s\n", ss, theKey, err)
	//	return 0
	//}
	return s
}

//func (this *Trx) getRedisIncrKey(theKey string) int64 {
//	conn, err := this.redisPool.Get()
//	if err != nil {
//		return 0
//	}
//	defer this.redisPool.Put(conn)
//
//	if db20 {
//		fmt.Fprintf(os.Stderr, "incr %s; %s\n", theKey, godebug.LF())
//	}
//	s, err := conn.Cmd("INCR", theKey).Int64()
//	if db20 {
//		fmt.Fprintf(os.Stderr, "aftr ss=%d; %s\n", s, godebug.LF())
//	}
//	if err != nil {
//		fmt.Printf("Error(11007) error incrementing key, returning empty interface for:%s, err=%s\n", theKey, err)
//	}
//	return s
//}

const db20 = false

func (this *Trx) GetKeySeqNo() (int64, int) {
	// g_key_no := this.getRedisKeyInt64("rest:trace:uri:key1", 1)
	return int64(this.Key), g_seq_no
}

func (this *Trx) IncrSeqNo() int {
	g_seq_no++
	return g_seq_no
}

//func (this *Trx) IncrKeyNo() int64 {
//	g_key_no = this.getRedisIncrKey("rest:trace:uri:key1")
//	return g_key_no
//}
func (this *Trx) StartSeqNo() int {
	g_seq_no = 1
	return g_seq_no
}

// ----------------------------------------------------------------------------------------------------------
type DBTiming struct {
	curTime     time.Time
	elapsedTime time.Duration
}

type DBQryHash struct {
	QryMap map[string]DBTiming
}

// ----------------------------------------------------------------------------------------------------------
var TimeDbCall DBQryHash

func init() {
	TimeDbCall.QryMap = make(map[string]DBTiming)
}
func TraceDbStartQry(s string) {
	var v DBTiming
	v.curTime = time.Now()
	TimeDbCall.QryMap[s] = v
}
func TraceDbEndQry(s string, v ...interface{}) time.Duration {
	startTime := TimeDbCall.QryMap[s].curTime
	finishTime := time.Now()
	elapsedTime := finishTime.Sub(startTime)
	return elapsedTime
}

// ----------------------------------------------------------------------------------------------------------
func (this *Trx) TraceDb(name string, qry string, args ...interface{}) {
	if !trace_on {
		return
	}
	conn, err := this.redisPool.Get()
	if err != nil {
		return
	}
	defer this.redisPool.Put(conn)

	pc, file, line, _ := runtime.Caller(1)
	xfunc := runtime.FuncForPC(pc).Name()
	if stdout_on {
		fmt.Printf("\n\n>>Query Func:%s File:%s Line:%d<<: %s; args=%s\n", xfunc, file, line, qry, SVar(args))
	}
	// Key Start

	key_no, seq_no := this.GetKeySeqNo()
	// get the key - # + Incr on it
	theKey := fmt.Sprintf(`rest:trace:%d:seq:%d`, key_no, seq_no)
	// seq_no++			// seq_no = IncrSeqNo()
	seq_no = this.IncrSeqNo()
	theData := fmt.Sprintf(`{"query":%s,"data":%s}`, qry, SVar(args))
	// fmt.Printf ( "set %s %s; %s\n", theKey, theData, LF() )
	conn.Cmd("SET", theKey, theData)
	conn.Cmd("EXPIRE", theKey, ttl)
	conn.Cmd("INCR", "trace:cnt:QRY")
	TraceDbStartQry(qry)

}
func (this *Trx) TraceDb2(name string, qry string, args ...interface{}) {
	if !trace_on {
		return
	}
	conn, err := this.redisPool.Get()
	if err != nil {
		return
	}
	defer this.redisPool.Put(conn)

	pc, file, line, _ := runtime.Caller(2)
	xfunc := runtime.FuncForPC(pc).Name()
	if stdout_on {
		fmt.Printf("\n\n>>Query Func:%s File:%s Line:%d<<: %s; args=%s\n", xfunc, file, line, qry, SVar(args))
	}
	// Key Start
	key_no, seq_no := this.GetKeySeqNo()
	// get the key - # + Incr on it
	theKey := fmt.Sprintf(`rest:trace:%d:seq:%d`, key_no, seq_no)
	seq_no = this.IncrSeqNo()
	theData := fmt.Sprintf(`{"query":%s,"data":%s}`, qry, SVar(args))
	// fmt.Printf ( "set %s %s; %s\n", theKey, theData, LF() )
	conn.Cmd("SET", theKey, theData)
	conn.Cmd("EXPIRE", theKey, ttl)
	conn.Cmd("INCR", "trace:cnt:QRY")
	TraceDbStartQry(qry)
}
func (this *Trx) TraceDbData(name string, qry string, args ...interface{}) {
	if !trace_on {
		return
	}
	// Key End
}
func (this *Trx) TraceDbEnd(name string, qry string, n_rows int) {
	if !trace_on {
		return
	}
	conn, err := this.redisPool.Get()
	if err != nil {
		return
	}
	defer this.redisPool.Put(conn)

	// Key End
	key_no, seq_no := this.GetKeySeqNo()
	theKey := fmt.Sprintf(`rest:trace:%d:seq:%d`, key_no, seq_no)
	seq_no = this.IncrSeqNo()
	theData := fmt.Sprintf(`{"done":true,"status":"success","n_rows":%v}`, n_rows)
	// fmt.Printf ( "set %s %s; %s\n", theKey, theData, LF() )
	conn.Cmd("SET", theKey, theData)
	conn.Cmd("EXPIRE", theKey, ttl)
	dt := TraceDbEndQry(qry)
	this.redisSetElapsedTime("trace:time:QRY:", qry, false, dt)
}
func (this *Trx) TraceDbError(name string, qry string, err error) {
	if !trace_on {
		return
	}
	conn, err := this.redisPool.Get()
	if err != nil {
		return
	}
	defer this.redisPool.Put(conn)

	pc, file, line, _ := runtime.Caller(1)
	xfunc := runtime.FuncForPC(pc).Name()
	if stdout_on {
		fmt.Printf(">>Error Running (Func:%s File:%s Line:%d) qry=%s err=%s\n", xfunc, file, line, qry, err)
	}
	// Key End
	key_no, seq_no := this.GetKeySeqNo()
	theKey := fmt.Sprintf(`rest:trace:%d:seq:%d`, key_no, seq_no)
	seq_no = this.IncrSeqNo()
	theData := fmt.Sprintf(`{"done":true,"status":"error","msg":"%s"}`, err)
	// fmt.Printf ( "set %s %s; %s\n", theKey, theData, LF() )
	conn.Cmd("SET", theKey, theData)
	conn.Cmd("EXPIRE", theKey, ttl)
	dt := TraceDbEndQry(qry)
	this.redisSetElapsedTime("trace:time:QRY:", qry, false, dt)
}
func (this *Trx) TraceDbError2(name string, qry string, err error) {
	if !trace_on {
		return
	}
	conn, err := this.redisPool.Get()
	if err != nil {
		return
	}
	defer this.redisPool.Put(conn)

	pc, file, line, _ := runtime.Caller(2)
	xfunc := runtime.FuncForPC(pc).Name()
	if stdout_on {
		fmt.Printf(">>Error Running (Func:%s File:%s Line:%d) qry=%s err=%s\n", xfunc, file, line, qry, err)
	}
	// Key End
	key_no, seq_no := this.GetKeySeqNo()
	theKey := fmt.Sprintf(`rest:trace:%d:seq:%d`, key_no, seq_no)
	seq_no = this.IncrSeqNo()
	theData := fmt.Sprintf(`{"done":true,"status":"error","msg":"%s"}`, err)
	// fmt.Printf ( "set %s %s; %s\n", theKey, theData, LF() )
	conn.Cmd("SET", theKey, theData)
	conn.Cmd("EXPIRE", theKey, ttl)
	dt := TraceDbEndQry(qry)
	this.redisSetElapsedTime("trace:time:QRY:", qry, false, dt)
}
func (this *Trx) TraceDbRet(name string, args ...interface{}) {
	if !trace_on {
		return
	}
}
func (this *Trx) TraceDbRetJson(name string, data string) {
	if !trace_on {
		return
	}
}

// ----------------------------------------------------------------------------------------------------------
func (this *Trx) TraceUri(req *http.Request, mm url.Values) {
	if !trace_on {
		return
	}
	conn, err := this.redisPool.Get()
	if err != nil {
		return
	}
	defer this.redisPool.Put(conn)

	// key_no := this.IncrKeyNo()
	key_no := this.Key
	// seq_no = 1
	seq_no := this.StartSeqNo()
	pc, file, line, _ := runtime.Caller(1)
	xfunc := runtime.FuncForPC(pc).Name()

	_, _, _ = xfunc, file, line
	//if stdout_on {
	//	fmt.Printf("\n\n>>Trace URI Func:%s File:%s Line:%d<<: %s\n", xfunc, file, line, req.RequestURI)
	//	fmt.Printf(">>Content Type<<: %s\n", req.Header.Get("Content-Type"))
	//	fmt.Printf(">>Data<<: %s\n", SVar(m))
	//}
	// Key Start
	theKey := fmt.Sprintf(`rest:trace:%d:seq:%d`, key_no, seq_no)
	seq_no = this.IncrSeqNo()
	theData := fmt.Sprintf(`{"uri":"%s","data":%s}`, req.RequestURI, SVar(mm)) // xxxxx
	// fmt.Printf ( "set %s %s; %s\n", theKey, theData, LF() )
	conn.Cmd("SET", theKey, theData, "EX", ttl)
	// conn.Cmd("EXPIRE", theKey, ttl)
	// 127.0.0.1:6379> publish pubsub "Ya, you think this will work"

	//	if _, ok := m["auth_token"]; ok {
	//		auth_token := m["auth_token"][0]
	//		// conn.Cmd ( "SET", "api:USER:"+auth_token, dt )
	//		conn.Cmd("EXPIRE", "api:USER:"+auth_token, 1*60*60) // Validate for 1 hour
	//	}

	conn.Cmd("PUBLISH", "trx:listen", fmt.Sprintf(`{"ClientTrxId":%q,"To":"rps://tracer/uri-start","maxKey":%d,"Path":"/uri-start","Scheme":"rps"}`, this.RequestId, key_no))
}

func (this *Trx) TraceUriRaw(req *http.Request) {
	if !trace_on {
		return
	}
	conn, err := this.redisPool.Get()
	if err != nil {
		return
	}
	defer this.redisPool.Put(conn)

	// key_no := this.IncrKeyNo()
	key_no := this.Key
	//seq_no = 1
	seq_no := this.StartSeqNo()
	pc, file, line, _ := runtime.Caller(1)
	xfunc := runtime.FuncForPC(pc).Name()
	if stdout_on {
		fmt.Printf("\n\n>>Trace URI Func:%s File:%s Line:%d<<: %s\n", xfunc, file, line, req.RequestURI)
		fmt.Printf(">>Content Type<<: %s\n", req.Header.Get("Content-Type"))
	}
	theKey := fmt.Sprintf(`rest:trace:%d:seq:%d`, key_no, seq_no)
	seq_no = this.IncrSeqNo()
	theData := fmt.Sprintf(`{"uri":"%s","data":{}}`, req.RequestURI)
	// fmt.Printf ( "set %s %s; %s\n", theKey, theData, LF() )
	conn.Cmd("SET", theKey, theData)
	conn.Cmd("EXPIRE", theKey, ttl)
	conn.Cmd("INCR", "trace:cnt:URI")
	// xyzzy - work on this
	conn.Cmd("PUBLISH", "trx:listen", fmt.Sprintf(`{"TrxId":%q,"cmd":"key","maxKey":%d,"op":"uri-start"}`, this.RequestId, key_no))
}

// ----------------------------------------------------------------------------------------------------------
// ps is convered ps.DumpParams() from the Params package // ps *goftlmux.Params
func (this *Trx) TraceUriPs(req *http.Request, ps string) {
	if !trace_on {
		return
	}
	conn, err := this.redisPool.Get()
	if err != nil {
		return
	}
	defer this.redisPool.Put(conn)

	// key_no := this.IncrKeyNo()
	key_no := this.Key
	// seq_no = 1
	seq_no := this.StartSeqNo()
	pc, file, line, _ := runtime.Caller(1)
	xfunc := runtime.FuncForPC(pc).Name()

	if stdout_on {
		fmt.Printf("\n\n>>Trace URI Func:%s File:%s Line:%d<<: %s\n", xfunc, file, line, req.RequestURI)
		fmt.Printf(">>Content Type<<: %s\n", req.Header.Get("Content-Type"))
		// fmt.Printf(">>Data<<: %s\n", ps.DumpParam())
	}
	// Key Start
	theKey := fmt.Sprintf(`rest:trace:%d:seq:%d`, key_no, seq_no)
	seq_no = this.IncrSeqNo()
	// theData := fmt.Sprintf(`{"uri":"%s","data":%s}`, req.RequestURI, ps.DumpParam())
	theData := fmt.Sprintf(`{"uri":"%s","data":%s}`, req.RequestURI, ps)
	// fmt.Printf ( "set %s %s; %s\n", theKey, theData, LF() )
	conn.Cmd("SET", theKey, theData)
	conn.Cmd("EXPIRE", theKey, ttl)
	// 127.0.0.1:6379> publish pubsub "Ya, you think this will work"

	// if ps.HasName("auth_token") {
	// 	auth_token := ps.ByName("auth_token")
	// 	conn.Cmd("EXPIRE", "api:USER:"+auth_token, 1*60*60) // Validate for 1 hour
	// }

	// xyzzy - work on this

	conn.Cmd("PUBLISH", "trx:listen", fmt.Sprintf(`{"TrxId":%q,"cmd":"key","maxKey":%d,"op":"uri-start"}`, this.RequestId, key_no))
}

func (this *Trx) TraceUriRawEnd(req *http.Request, elapsedTime time.Duration) {
	if !trace_on {
		return
	}
	conn, err := this.redisPool.Get()
	if err != nil {
		return
	}
	defer this.redisPool.Put(conn)

	if stdout_on {
		fmt.Printf(">>Trace URI End<<: %s Duration: %v\n", req.RequestURI, elapsedTime)
	}
	// Key End
	key_no, seq_no := this.GetKeySeqNo()
	theKey := fmt.Sprintf(`rest:trace:%d:seq:%d`, key_no, seq_no)
	seq_no = this.IncrSeqNo()
	theData := fmt.Sprintf(`{"uridone":true}`)
	// fmt.Printf ( "set %s %s; %s\n", theKey, theData, LF() )
	conn.Cmd("SET", theKey, theData)
	conn.Cmd("EXPIRE", theKey, ttl)
	// conn.Cmd("PUBLISH", "trx:listen", fmt.Sprintf(`{"TrxId":%q,"cmd":"key","maxKey":%d,"op":"uri-end"}`, this.RequestId, key_no))
	conn.Cmd("PUBLISH", "trx:listen", fmt.Sprintf(`{"ClientTrxId":%q,"To":"rps://tracer/uri-end","maxKey":%d,"Path":"/uri-end","Scheme":"rps"}`, this.RequestId, key_no))
	// t = MGetKeys ( "trace:time:QRY:*" )
	// "elapsedTimeString": fmt.Sprintf ( "%f", r.elapsedTime.Seconds()),
	this.redisSetElapsedTime("trace:time:URI:", req.RequestURI, true, elapsedTime)
}

func (this *Trx) TraceUriRawEndHijacked(req *http.Request, elapsedTime time.Duration) {
	if !trace_on {
		return
	}
	conn, err := this.redisPool.Get()
	if err != nil {
		return
	}
	defer this.redisPool.Put(conn)

	if stdout_on {
		fmt.Printf(">>Trace URI End<<: %s Duration: %v\n", req.RequestURI, elapsedTime)
	}
	// Key End
	key_no, seq_no := this.GetKeySeqNo()
	theKey := fmt.Sprintf(`rest:trace:%d:seq:%d`, key_no, seq_no)
	seq_no = this.IncrSeqNo()
	theData := fmt.Sprintf(`{"uridone":true}`)
	// fmt.Printf ( "set %s %s; %s\n", theKey, theData, LF() )
	conn.Cmd("SET", theKey, theData)
	conn.Cmd("EXPIRE", theKey, ttl)
	conn.Cmd("PUBLISH", "trx:listen", fmt.Sprintf(`{"ClientTrxId":%q,"To":"rps://tracer/uri-end-hijacked","maxKey":%d,"Path":"/uri-end-hijacked","Scheme":"rps"}`, this.RequestId, key_no))
	// t = MGetKeys ( "trace:time:QRY:*" )
	// "elapsedTimeString": fmt.Sprintf ( "%f", r.elapsedTime.Seconds()),
	this.redisSetElapsedTime("trace:time:URI:", req.RequestURI, true, elapsedTime)
}

// ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

const (
	save_UriStart = 0
	save_Func     = 1
	save_At       = 2
	save_Qry      = 3
	save_QryDone  = 4
	save_UriEnd   = 5
)

var g_CfgSave []bool
var g_CfgSaveNames []string

func init() {
	g_CfgSave = []bool{
		false, // UriStart
		false, // Func
		false, // At
		false, // Qry
		true,  // QryDone
		true,  // UriEnd
	}
	g_CfgSaveNames = []string{
		"UriStart",
		"Func",
		"At",
		"Qry",
		"QryDone",
		"UriEnd",
	}
}

// moved to common/namevaluefrom.go
//type NameValueFrom struct {
//	Name  string
//	Value string
//	From  string
//}

type FuncData struct {
	Type     string
	File     string
	Line     string
	FuncName string
}

// { "SQL": "select...", "Data": [ ... ], "Time": ..., "RvData": [ ... ], "FuncName":... }
type QryData struct {
	SQL      string
	BindData []interface{}
	Time     string
	RvData   string
	QError   string
	File     string
	Line     string
	FuncName string
}

//		this.Note = append ( this.Note, ANote{ Txt:note, File:file, Line:fmt.Sprintf("%d",line), FuncName:xfunc ) } )
type ANote struct {
	Txt      string
	File     string
	Line     string
	FuncName string
}

type APathMatch struct {
	AtDepth        int
	Match          string
	Url            string
	MiddlewareName string
	ErrorReturn    string
}

type Trx struct {
	Key                   int
	Data                  []common.NameValueFrom
	Func                  []FuncData
	From                  string
	Qry                   []QryData
	RvBody                string
	ClientIp              string
	RequestTime           string
	Method, Uri, Protocol string
	Status                int
	ResponseBytes         int64
	ElapsedTime           string // In Seconds
	ElapsedTimeMs         string // In Milesconds
	startTime             time.Time
	TableList             []string
	Note                  []ANote
	Username              string
	User_id               string
	Auth_token            string
	HasBeenSaved          bool
	IAm                   string
	MiddlewareStk         []APathMatch
	redisPool             *pool.Pool
	RequestId             string
	Header                http.Header
}

func NewTrx(r *pool.Pool) *Trx {

	id0, _ := uuid.NewV4()
	id := id0.String()

	runAtTime := time.Now()
	timeFormatted := runAtTime.Format(ISO8601)

	this := &Trx{
		redisPool:     r,
		Data:          nil,
		Func:          nil,
		From:          "",
		Qry:           nil,
		RvBody:        "",
		ClientIp:      "",
		RequestTime:   timeFormatted,
		Method:        "",
		Uri:           "",
		Protocol:      "",
		Status:        0,
		ResponseBytes: 0,
		ElapsedTime:   "",
		TableList:     nil,
		Note:          nil,
		Username:      "",
		User_id:       "",
		Auth_token:    "",
		IAm:           "unspec",
		MiddlewareStk: nil,
		HasBeenSaved:  false,
		RequestId:     id,
		Header:        make(http.Header),
	}

	// key_no, _ := this.GetKeySeqNo()
	key_no := this.getRedisKeyInt64("rest:trace:uri:key1", 1)
	this.Key = int(key_no)

	return this
}

//2. Mod "tr.go" to have a AddTrxId - call that will
//	1. create trx-id:ID - with { url, method } -- TTL for 30 min
//	2. create trx- List with ID in it

func (this *Trx) TrxIdSeen(TrxId, URL, Method string) {
	if !trace_on {
		return
	}
	conn, err := this.redisPool.Get()
	if err != nil {
		return
	}
	defer this.redisPool.Put(conn)

	this.RequestId = TrxId
	conn.Cmd("SET", "trx-id:"+TrxId, fmt.Sprintf(`{"url":%q,"method":%q}`, URL, Method), "EX", ttl) // conn.Cmd("EXPIRE", theKey, ttl)
	prev, cur := TopOfTheHour()
	conn.Cmd("EXPIRE", "trx-"+prev, ttl)
	conn.Cmd("SADD", "trx-"+cur, fmt.Sprintf(`{"Key":%d,"TrxId":%q,"url":%q,"method":%q,"RequestTime":"%s"}`, this.Key, TrxId, URL, Method, this.RequestTime))
	conn.Cmd("EXPIRE", "trx-"+cur, ttl)
}

func (this *Trx) MatchedPath(depth int, note string) {
}

func (this *Trx) NextNoMatch(depth int, note string) {
}

func (this *Trx) SetSavePoint(s string, on bool) {
	i := MiscLib.InArray(s, g_CfgSaveNames)
	if i >= 0 {
		g_CfgSave[i] = on
	}
}

func (this *Trx) SetHeader(h http.Header) {
	this.Header = h
}

func (this *Trx) SavePointOn(x int) bool {
	if x < 0 || x > 5 {
		return false
	}
	b := g_CfgSave[x]
	return b
}

func (this *Trx) SaveIt() {
	conn, err := this.redisPool.Get()
	if err != nil {
		return
	}
	defer this.redisPool.Put(conn)

	this.HasBeenSaved = true
	this.IAm = "DbTrace"
	//key_no, _ := this.GetKeySeqNo()
	//this.Key = int(key_no)
	theData := SVar(this)
	theKey := fmt.Sprintf(`trx:%06d`, this.Key)
	conn.Cmd("SET", theKey, theData, "EX", ttl)
	conn.Cmd("INCR", "trace:cnt:QRY")
}

// trx.UriStart ( ... )								-- opt-save --
func (this *Trx) UriStart(uri string) {
	if !trace_on {
		return
	}
	this.Uri = uri
	if this.SavePointOn(save_UriStart) {
		this.SaveIt()
	}
}

func (this *Trx) SetTraceDebug(db bool) {
	trace_on = true
	trace_debug = db
}

func (this *Trx) SetTablesUsed(tableList []string) {
	if !trace_on {
		return
	}
	this.TableList = tableList
}

func (this *Trx) AddNote(depth int, note string) {
	if !trace_on {
		return
	}
	note = strings.Replace(note, "%", "%%", -1)
	pc, file, line, ok := runtime.Caller(depth)
	if ok {
		xfunc := runtime.FuncForPC(pc).Name()
		line_s := fmt.Sprintf("%d", line)
		this.Note = append(this.Note, ANote{Txt: note, File: file, Line: line_s, FuncName: xfunc})
		if trace_debug {
			fmt.Printf("Trx: { \"Note\":%q, \"File\":%q, \"Line\":%q, \"FuncName\":%q }\n", note, file, line_s, xfunc)
		}
	} else {
		this.Note = append(this.Note, ANote{Txt: note})
		if trace_debug {
			fmt.Printf("Trx: { \"Note\":%q }\n", note)
		}
	}
}

func (this *Trx) PathMatched(depth int, middlewareName string, path []string, pn int, url string) {
	if len(path) == 0 && pn >= 0 {
		this.MiddlewareStk = append(this.MiddlewareStk, APathMatch{Url: url, MiddlewareName: middlewareName, Match: fmt.Sprintf("***error PathMatch: pn=%d when 0 paths ***", pn), AtDepth: depth})
	} else if len(path) >= pn {
		this.MiddlewareStk = append(this.MiddlewareStk, APathMatch{Url: url, MiddlewareName: middlewareName, Match: path[pn], AtDepth: depth})
	} else {
		this.MiddlewareStk = append(this.MiddlewareStk, APathMatch{Url: url, MiddlewareName: middlewareName, Match: fmt.Sprintf("***error PathMatch: pn=%d when only %d paths ***", pn, len(path)), AtDepth: depth})
	}
}

func (this *Trx) ErrorReturn(depth int, err error) {
	this.MiddlewareStk = append(this.MiddlewareStk, APathMatch{ErrorReturn: fmt.Sprintf("%s", err), AtDepth: depth})
}

func (this *Trx) SetDataPs(data []common.NameValueFrom) {
	if !trace_on {
		return
	}
	for ii := 0; ii < len(data); ii++ {
		haveIt := false
		for _, ww := range this.Data {
			if ww.Name == data[ii].Name {
				haveIt = true
			}
		}
		if !haveIt {
			this.Data = append(this.Data, data[ii])
		}
	}
}

func (this *Trx) UpdateDataPs(data []common.NameValueFrom) {
	n := len(data)
	this.Data = make([]common.NameValueFrom, 0, n)
	this.SetDataPs(data)
}

// func (this *Trx) SetDataPs(ps *goftlmux.Params) {
/*
func SetDataPs(trx *tr.Trx, ps *goftlmux.Params) {
	if ps != nil {
		for i := 0; i < ps.NParam; i++ {
			nm := ps.Data[i].Name
			vl := ps.Data[i].Value
			ff := goftlmux.FromTypeToString(ps.Data[i].From)
			haveIt := false
			for jj, ww := range trx.Data {
				if ww.Name == nm {
					haveIt = true
				}
			}
			if !haveIt {
				trx.Data = append(trx.Data, common.NameValueFrom{Name: nm, Value: vl, From: ff})
			}
		}
	}
}
*/

func (this *Trx) AddData(i string, m url.Values, fr map[string]string) {
	if !trace_on {
		return
	}
	this.Data = append(this.Data, common.NameValueFrom{Name: i, Value: m[i][0], From: fr[i]})
	fmt.Printf("i= ->%s<- ->%v<-\n", i, i)
	fmt.Printf("this.Data=%s\n", SVar(this.Data))

}

// trx.Func ( name, file, lineno ) 					-- opt-save --
func (this *Trx) SetFunc(depth int) {
	if !trace_on {
		return
	}
	// xyzzy - see if this is a function we should be tracing? ( or decented of )
	pc, file, line, ok := runtime.Caller(depth)
	if ok {
		xfunc := runtime.FuncForPC(pc).Name()
		this.Func = append(this.Func, FuncData{"Func", file, fmt.Sprintf("%d", line), xfunc})
	} else {
		this.Func = append(this.Func, FuncData{"Func", "?", "-1", "?"})
	}
	if this.SavePointOn(save_Func) {
		this.SaveIt()
	}
}

// trx.At ( file, lineno )								-- opt-save --
func (this *Trx) At(depth int) {
	if !trace_on {
		return
	}
	// xyzzy - see if this is a function we should be tracing? ( or decented of )
	pc, file, line, ok := runtime.Caller(depth)
	if ok {
		xfunc := runtime.FuncForPC(pc).Name()
		this.Func = append(this.Func, FuncData{"At", file, fmt.Sprintf("%d", line), xfunc})
	} else {
		this.Func = append(this.Func, FuncData{"At", "?", "-1", "?"})
	}
	if this.SavePointOn(save_At) {
		this.SaveIt()
	}
}

func (this *Trx) SetFuncRet(depth int) {
	if !trace_on {
		return
	}
	// xyzzy - see if this is a function we should be tracing? ( or decented of )
	pc, file, line, ok := runtime.Caller(depth)
	if ok {
		xfunc := runtime.FuncForPC(pc).Name()
		this.Func = append(this.Func, FuncData{"Ret", file, fmt.Sprintf("%d", line), xfunc})
	} else {
		this.Func = append(this.Func, FuncData{"Ret", "?", "-1", "?"})
	}
	if this.SavePointOn(save_At) {
		this.SaveIt()
	}
}

// trx.From ( ... )
func (this *Trx) SetFrom(from string) {
	if !trace_on {
		return
	}
	this.From = from
	//if this.SavePointOn(save_UriStart) {
	//	this.SaveIt()
	//}
}

// trx.From ( ... )
func (this *Trx) SetUserInfo(username string, user_id string, auth_token string) {
	if !trace_on {
		return
	}
	this.Username = username
	this.User_id = user_id
	this.Auth_token = auth_token
}

// trx.Qry ( Sql, Data, FuncName, FileName, LineNo ) 	-- opt-save --
func (this *Trx) SetQry(sql string, depth int, data ...interface{}) {
	if !trace_on {
		return
	}
	xfunc := ""
	pc, file, line, ok := runtime.Caller(depth)
	if ok {
		xfunc = runtime.FuncForPC(pc).Name()
	} else {
		xfunc = "?"
	}
	this.Qry = append(this.Qry, QryData{SQL: sql, BindData: data, Time: "", RvData: "", QError: "", File: file, Line: fmt.Sprintf("%d", line), FuncName: xfunc})
	this.startTime = time.Now()
	if this.SavePointOn(save_Qry) {
		this.SaveIt()
	}
}

// trx.QryDone ( RvData )								-- opt-save --
func (this *Trx) SetQryDone(qError string, rvData string) {
	if !trace_on {
		return
	}
	j := len(this.Qry) - 1
	// fmt.Printf ( "j = %v !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n", j )
	if j >= 0 {
		finishTime := time.Now()
		elapsedTime := finishTime.Sub(this.startTime)
		// fmt.Printf ( "Adding in data and time!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n" )
		this.Qry[j].Time = fmt.Sprintf("%v", elapsedTime)
		this.Qry[j].RvData = rvData
		this.Qry[j].QError = qError
	}
	if this.SavePointOn(save_QryDone) {
		this.SaveIt()
	}
}

func (this *Trx) SetCacheData(sql string, depth int, rvData string, data ...interface{}) {
	if !trace_on {
		return
	}
	xfunc := ""
	pc, file, line, ok := runtime.Caller(depth)
	if ok {
		xfunc = runtime.FuncForPC(pc).Name()
	} else {
		xfunc = "?"
	}
	this.Qry = append(this.Qry, QryData{SQL: sql, BindData: data, Time: "", RvData: rvData, QError: "cache", File: file, Line: fmt.Sprintf("%d", line), FuncName: xfunc})
	if this.SavePointOn(save_QryDone) {
		this.SaveIt()
	}
}

// trx.RvBody ( ... )
func (this *Trx) SetRvBody(rvBody string) {
	if !trace_on {
		return
	}
	this.RvBody = rvBody
	//if this.SavePointOn(save_UriStart) {
	//	this.SaveIt()
	//}
}

// trx.RvHdr ( ... )			// xyzzy - set
// trx.RvCookie ( ... )			// xyzzy - set

func (this *Trx) UriSaveData(clientIP string, runAtTime time.Time, method string, uri string, protocal string, status int, bodyLen int64, elapsedTime time.Duration, r *http.Request) {
	if !trace_on {
		return
	}
	this.ClientIp = clientIP
	// timeFormatted := runAtTime.time.Format("02/Jan/2006 03:04:05")
	timeFormatted := runAtTime.Format(ISO8601)
	this.RequestTime = timeFormatted
	this.Method = method
	this.Uri = uri
	this.Protocol = protocal
	this.Status = status
	this.ResponseBytes = bodyLen
	this.ElapsedTime = fmt.Sprintf("%v", elapsedTime.Seconds())
	this.ElapsedTimeMs = fmt.Sprintf("%v", elapsedTime.Seconds()*1000.0)
	/* xyzzy extra fields from 'r' */
	/* what about other URI fields ?
	fmt.Printf ( "\treq.URL.Scheme=%s\n", r.URL.Scheme)
	fmt.Printf ( "\treq.URL.Host=%s\n", r.URL.Host)
	fmt.Printf ( "\treq.URL.Path=%s\n", r.URL.Path)
	fmt.Printf ( "\treq.URL.RawQuery=%s\n", r.URL.RawQuery)
	fmt.Printf ( "\treq.URL.Fragment=%s\n", r.URL.Fragment)
	*/
	/* what about Headers */
	/* what about Cookies */
	if this.SavePointOn(save_UriEnd) {
		this.SaveIt()
	}
}

func TopOfTheHour() (prev, cur string) {

	nw := time.Now()
	hr := nw.Hour()
	mn := nw.Minute()
	if mn > 0 && mn < 30 {
		mn = 0
	} else {
		mn = 1
	}

	xx := (hr * 2) + mn
	cur = fmt.Sprintf("%02d", xx)

	if hr >= 1 {
		xx = (hr * 2) + (mn - 1)
	} else {
		xx = (23 * 2) + (mn - 1)
	}
	prev = fmt.Sprintf("%02d", xx)

	fmt.Printf("cur [%s] prev [%s]\n", cur, prev)

	//prev = "00"
	//cur = "01"
	return
}

/* vim: set noai ts=4 sw=4: */
