//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1002
//
// MIT License LICENSE.txt
//

package goftlmux

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	logrus "github.com/pschlump/pslog" // "github.com/sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/RedisSessionData"
	"github.com/pschlump/godebug"
	"github.com/pschlump/json" //	Modifed from: "encoding/json"
)

//    "github.com/pschlump/Go-FTL/server/RedisSessionData"
// "github.com/pschlump/Go-FTL/server/base"

// Notes:
//	http://golang.org/src/net/http/server.go - ResponseWriter

// ------------------------------------------------------------------------------------------------------------
// Implement a compatible http.ResponseWriter that saves all the data until the end.
// Good side: you don't need to finish your headers before your data.
// Good side: you can post-process the data/status.
// Good side: length header can be manipulated after the data is generated.
// Bad side: This won't work with a streaming data interface at all.
// Bad side: Also it's all buffered in memory. -- Not that big a deal - and common to all proxy usages anyhow.
type MidBuffer struct {
	wr                  http.ResponseWriter
	bb                  bytes.Buffer                           // The body of the response
	StatusCode          int                                    // StatusCode like 200, 404 etc.
	Headers             http.Header                            // All the headers that will be writen when done
	Length              int64                                  //	Length of the response
	Error               error                                  // Most recent error, if StatusCode == 200, then ignore
	Prefix              string                                 // Tack onto response when you flush.
	Postfix             string                                 //
	IndentFlag          bool                                   //	If JSON/JsonX/XML searilize will searilize it with indentation
	Row                 map[string]interface{}                 //	Single Row Response -- or table header info
	Table               []map[string]interface{}               //	Table of Row Response
	State               StateType                              // Byte, Row, Table
	SearilizeFormat     SearilizeType                          // Byte, Row, Table
	StartTime           time.Time                              // Start time for deltaT and proxy timeout -- deltaT := time.Since(mb.StartTime).String()
	Modtime             time.Time                              // file modification time
	NRewrite            int                                    //	# of rewrites that have occured - may be a limit on this (prevents loops)
	RerunRequest        bool                                   //
	AddInfo             map[string]string                      // Way to pass info from middleware to middleware (Session)
	MapLock             sync.Mutex                             // Lock for accesing maps in this.
	Ps                  Params                                 // xyzzyParams - PJS - change request interface to pass/parse 'params' as modifed req
	Next                http.Handler                           // Required field for all chaining of middleware.
	Hdlr                interface{}                            // Handler to the "TOP" or nil
	ResolvedFn          string                                 // Single file name - resolved to local from fileserver
	DependentFNs        []string                               // Set of files that if any have chagned (mod-datetime) then should not cache and let lower levels re-generate
	IsProxyFile         bool                                   // File was fetched by a proxy to another server - it is NOT(local)
	SaveDataInCache     bool                                   // If true then save the data in the cache (gzip -> new data)
	DirTemplateFileName string                                 //
	TemplateLineNo      int                                    //
	IgnoreDirs          []string                               //
	OriginalURL         string                                 //
	G_Trx               interface{}                            //
	Extend              map[string]interface{}                 //
	RequestTrxId        string                                 //		// Id for the entire request //
	IsHijacked          bool                                   //
	ParsedHTML          interface{}                            //		New data
	ParsedCSS           interface{}                            //		New data
	PackedData          interface{}                            //		New data
	Dependencies        interface{}                            //		New data
	Session             *RedisSessionData.RedisSessionDataType //
}

// AddInfo         map[string]interface{}   // Way to pass info from middleware to middleware (Session)

// Verify meats interface
var _ http.ResponseWriter = (*MidBuffer)(nil) // Is A ResponseWriter
var _ http.Flusher = (*MidBuffer)(nil)        // Is A ...
var _ http.Hijacker = (*MidBuffer)(nil)       // Is A ...

type StateType int
type SearilizeType int

const (
	ByteBuffer StateType = iota
	RowBuffer
	TableBuffer
	IAmDone
)
const (
	SearilizeJSON SearilizeType = iota
	SearilizeXML
	SearilizeHTML
	SearilizeCSV
	SearilizeTemplate
)

func (rw *MidBuffer) SaveCurentBody(body string) {
	// xyzzy - goftlmux.bufferhtml - 107
	// xyzzy - 1. create a type in the buffer
	// xyzzy - 2. save "body" + Etag -> this with file name from ResolvedFn
	// xyzzy - may want to write body out to a file in ./cache - if it is a proxied file -- based on hash of data
}

// Return a new buffered http.ResponseWriter
func NewMidBuffer(w http.ResponseWriter, hdlr interface{}) (rv *MidBuffer) {
	// trx, id := cfg.TrNewTrx()
	rv = &MidBuffer{
		wr:              w,
		Headers:         make(http.Header),
		StatusCode:      http.StatusOK,
		StartTime:       time.Now(),
		Length:          0,             // default
		Error:           nil,           // default
		Prefix:          "",            // default
		Postfix:         "",            // default
		Row:             nil,           // default
		Table:           nil,           // default
		State:           ByteBuffer,    // default
		SearilizeFormat: SearilizeJSON, // default
		IndentFlag:      false,         // default
		Hdlr:            hdlr,          // The Handler, TOP - nil in most test code.
		//RequestTrxId:    id,            //
		//G_Trx:           trx,           // Ptr to the Trx tracer
	}
	// wr.RequestTrxId = trx.RequestId
	InitParams(&rv.Ps)
	return
}

// ---------------------------------------------------------------------------------------------------------------------------------------
func (st StateType) String() string {
	switch st {
	case ByteBuffer:
		return "BytesBuffer"
	case RowBuffer:
		return "RowBuffer"
	case TableBuffer:
		return "TableBuffer"
	case IAmDone:
		return "IAmDone"
	}
	return fmt.Sprintf("*** unknown %d value for StateType ***", st)
}

// ---------------------------------------------------------------------------------------------------------------------------------------
func (b *MidBuffer) DumpBuffer() {
	fmt.Printf("\n-------------------------------------------------------------------------------\n")
	fmt.Printf("Dump the write buffer, %s, Called From: %s\n", godebug.LF(), godebug.LF(2))
	fmt.Printf("-------------------------------------------------------------------------------\n")
	fmt.Printf("\tStatusCode = %d\n", b.StatusCode)
	fmt.Printf("\tLength = %d\n", b.Length)
	fmt.Printf("\tPre/Post-Fix = --[%s]-- --[%s]--\n", b.Prefix, b.Postfix)
	fmt.Printf("\tError = %s\n", b.Error)
	fmt.Printf("\tIndentFlag = %v\n", b.IndentFlag)
	fmt.Printf("\tState = %v\n", b.State)
	fmt.Printf("\tNRewrite = %v\n", b.NRewrite)
	fmt.Printf("\tResolvedFn = %s\n", b.ResolvedFn)
	fmt.Printf("\tDependentFNs = %s\n", SVar(b.DependentFNs))
	fmt.Printf("\tSearilizeFormat = %v\n", b.SearilizeFormat)
	fmt.Printf("\tDeltaT(calc) = %v\n", time.Since(b.StartTime).String())
	fmt.Printf("\tHeaders = %s\n", SVarI(b.Headers))
	fmt.Printf("\tAddInfo = %s\n", SVarI(b.AddInfo))
	fmt.Printf("\tPs = %s\n", b.Ps.DumpParamDB())
	/*
		bb              bytes.Buffer             // The body of the response
	*/
	fmt.Printf("\n")
}

// ---------------------------------------------------------------------------------------------------------------------------------------
// ---------------------------------------------------------------------------------------------------------------------------------------
func SVar(v interface{}) string {
	s, err := json.Marshal(v)
	// s, err := json.MarshalIndent ( v, "", "\t" )
	if err != nil {
		return fmt.Sprintf("Error:%s", err)
	} else {
		return string(s)
	}
}

// ---------------------------------------------------------------------------------------------------------------------------------------
// ---------------------------------------------------------------------------------------------------------------------------------------
func SVarI(v interface{}) string {
	// s, err := json.Marshal ( v )
	s, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return fmt.Sprintf("Error:%s", err)
	} else {
		return string(s)
	}
}

// ---------------------------------------------------------------------------------------------------------------------------------------
// ResponseWriter Interface
// ---------------------------------------------------------------------------------------------------------------------------------------

// Return the headers - Required to make the interface work
func (b *MidBuffer) Header() http.Header {
	return b.Headers
}

// Implement http.ResponseWriter WriteHeader to just buffer the Status
func (b *MidBuffer) WriteHeader(StatusCode int) {
	if StatusCode == 301 {
		StatusCode = 307
	}
	b.StatusCode = StatusCode
}

func (b *MidBuffer) Write(buf []byte) (int, error) {
	// n, err := b.ResponseWriter.Write(buf)
	// fmt.Printf("Write [%s], %s\n", buf, godebug.LF())
	// b.NRewrite = 0
	if b.State == IAmDone {
		// fmt.Printf("xyzzy At %s -- Write afer final flush\n", godebug.LF())
		return 0, http.ErrWriteAfterFlush
	}

	// fmt.Printf("xyzzy At %s\n", godebug.LF())
	n, err := b.bb.Write(buf)
	if err == nil {
		// fmt.Printf("xyzzy At %s\n", godebug.LF())
		b.Length += int64(n)
	}
	// fmt.Printf("xyzzy At %s\n", godebug.LF())
	return n, err
}

// ---------------------------------------------------------------------------------------------------------------------------------------
// Hijacker Interface
// ---------------------------------------------------------------------------------------------------------------------------------------

func (b *MidBuffer) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := b.wr.(http.Hijacker); ok {
		b.IsHijacked = true
		return hj.Hijack()
	}
	return nil, nil, errors.New("I'm not a Hijacker")
}

// ---------------------------------------------------------------------------------------------------------------------------------------
// Flusher Interface
// ---------------------------------------------------------------------------------------------------------------------------------------

// Don't actualy flush - just ignore.
func (b *MidBuffer) Flush() {
}

// ---------------------------------------------------------------------------------------------------------------------------------------
//
// ---------------------------------------------------------------------------------------------------------------------------------------
func (b *MidBuffer) FinalFlush() {
	if b.IsHijacked {
		return
	}
	h := b.wr.Header()
	s := b.bb.Bytes()
	if db8 {
		fmt.Printf("In FinalFlush len(s)=>%d<= NRewrite=%d\n", len(s), b.NRewrite)
	}
	isHtml := false
	if len(b.Headers) > 0 {
		for key, val := range b.Headers {
			h[key] = val
			if key == "Content-Type" {
				for _, ss := range val {
					if strings.HasPrefix(ss, "text/html") {
						isHtml = true
					}
				}
			}
		}
	}
	if db8 {
		fmt.Printf("isHtml is %v ----------------------------------------------------------------------------- <<<<<<<<<<<<<<<<<<<<<<<<<<<<< \n", isHtml)
		if isHtml {
			fmt.Printf("In FinalFlush body=>%s<=, %s\n", s, godebug.LF())
		}
	}
	// ------------------------------------------- prefix / postfix --------------------------------
	b.Searilize()
	s = []byte(b.Prefix + string(s) + b.Postfix)
	l := len(s)
	b.Length = int64(l)
	if db8 {
		fmt.Printf("Length=%d\n", l)
	}
	h.Set("Content-Length", fmt.Sprintf("%d", l))
	if db8 {
		fmt.Printf("StatusCode = %d, %s\n", b.StatusCode, "260 Line No")
	}
	b.wr.WriteHeader(b.StatusCode)
	_, b.Error = b.wr.Write(s)
	if db8 {
		fmt.Printf("Error = %s, should be NIL\n", b.Error)
	}
	b.State = IAmDone
	return
}

func (b *MidBuffer) GetModtime() time.Time {
	return b.Modtime
}

func (b *MidBuffer) GetBody() (s []byte) {
	// fmt.Printf("xyzzy At %s\n", godebug.LF())
	// fmt.Printf("GetBody: top\n")
	if b.State != ByteBuffer { // If data has not been searilized, then...
		// fmt.Printf("xyzzy At %s\n", godebug.LF())
		// fmt.Printf("GetBody: searilize\n")
		b.searilizeInternal() // Searilize it
	}
	// fmt.Printf("xyzzy At %s\n", godebug.LF())
	s = b.bb.Bytes() // Pull out a copy of the data to return
	// fmt.Printf("GetBody: s=%s, %s\n", s, godebug.LF())
	if b.State != ByteBuffer {
		// fmt.Printf("xyzzy !!!! empty !!!! At %s\n", godebug.LF())
		b.bb.Truncate(0) // Empty the buffer
		b.Length = 0
	}
	return
}

func (b *MidBuffer) EmptyBody() {
	// b.bb = bytes.Buffer{}
	b.bb.Truncate(0)
	b.Length = 0
	b.State = ByteBuffer
}

func (b *MidBuffer) ReplaceBody(buf []byte) {
	b.bb.Truncate(0)
	b.Length = 0
	n, err := b.bb.Write(buf)
	if err != nil {
		b.Length = int64(n)
	}
}

func (b *MidBuffer) GetHeader() (h http.Header) {
	h = b.Headers
	return
}

func (b *MidBuffer) WriteRow(row map[string]interface{}) error {
	if b.State == IAmDone {
		return http.ErrWriteAfterFlush
	}
	// fmt.Printf("*** WriteRow ***, %s\n", godebug.LF())
	b.State = RowBuffer
	b.Row = row
	return nil
}

func (b *MidBuffer) WriteTable(d []map[string]interface{}) error {
	if b.State == IAmDone {
		return http.ErrWriteAfterFlush
	}
	// fmt.Printf("*** WriteTable ***, %s\n", godebug.LF())
	b.State = TableBuffer
	b.Table = d
	return nil
}

func (b *MidBuffer) searilizeInternal() {
	var err error
	var s []byte
	// SearilizeFormat SearilizeType            // Byte, Row, Table
	// SearilizeJSON     SearilizeType = 0
	if b.State == RowBuffer {
		if b.SearilizeFormat == SearilizeJSON {
			if b.IndentFlag {
				s, err = json.MarshalIndent(b.Row, "", "\t")
			} else {
				s, err = json.Marshal(b.Row)
			}
		} else if b.SearilizeFormat == SearilizeXML {
			s, err = xml.Marshal(b.Row)
		}
		b.Write(s)
	}
	if b.State == TableBuffer {
		if b.SearilizeFormat == SearilizeJSON {
			if b.IndentFlag {
				s, err = json.MarshalIndent(b.Table, "", "\t")
			} else {
				s, err = json.Marshal(b.Table)
			}
		} else if b.SearilizeFormat == SearilizeXML {
			s, err = xml.Marshal(b.Table)
		}
		b.Write(s)
	}
	b.Error = err
}

func (b *MidBuffer) Searilize() {
	b.searilizeInternal()
	b.State = ByteBuffer
}

const db8 = false

// ----------------------------------------------------------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------------------------------------------------------

type LogIt interface {
	Info(s string)
	Warn(s string)
	Error(s string)
	Fatal(s string)
	TraceEnabled() bool
}

type LogInfo struct {
	LoggingPath  string
	BaseFileName string
	TraceOn      bool
}

func (li *LogInfo) Info(s string) {
	logrus.Infof("Log: Info: %s\n", s)
}

func (li *LogInfo) Warn(s string) {
	logrus.Warnf("Log: Warn: %s\n", s)
}

func (li *LogInfo) Error(s string) {
	logrus.Errorf("Log: Error: %s\n", s)
}

func (li *LogInfo) Fatal(s string) {
	logrus.Fatalf("Log: Fatal: %s\n", s)
	os.Exit(1)
}

func (li *LogInfo) TraceEnabled() bool {
	return li.TraceOn
}

/* vim: set noai ts=4 sw=4: */
