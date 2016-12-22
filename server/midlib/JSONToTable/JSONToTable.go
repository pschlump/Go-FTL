//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1256
//

//
// Convert an array of objects in JSON into table data in the buffer.
//

package JSONToTable

import (
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/json" //	Modifed from: "encoding/json"
)

// --------------------------------------------------------------------------------------------------------------------------
func init() {

	// normally identical
	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
		pCfg, ok := ppCfg.(*JSONToTableType)
		if ok {
			pCfg.SetNext(next)
			rv = pCfg
		} else {
			err = mid.FtlConfigError
			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
		}
		return
	}

	// normally identical
	createEmptyType := func() interface{} { return &JSONToTableType{} }

	postInitValidation := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
		fmt.Printf("In postInitValidation, h=%v\n", h)
		hh, ok := h.(*JSONToTableType)
		if !ok {
			fmt.Printf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo)
			return mid.ErrInternalError
		} else {
			if hh.ConvertRowTo1LongTable && hh.Convert1LongTableToRow {
				fmt.Printf("Error: Invalid Configuration JSONToTable, both ConvertRowTo1LongTable and Convert1LongTableToRow are true, only one can be true at a time, Line No:%d\n", hh.LineNo)
				return mid.ErrInvalidConfiguration
			}
		}
		return nil
	}

	// /api/tmpl/showRpt.tmpl -> fetch data inside template?
	// /api/tmpl/showRpt.tmpl?data=bob (data in row/table data)
	cfg.RegInitItem2("JSONToTable", initNext, createEmptyType, postInitValidation, `{
		"Paths":                   { "type":[ "string", "filepath" ], "isarray":true, "default":"/" },
		"ConvertRowTo1LongTable":  { "type":[ "bool" ], "default":"false" },
		"Convert1LongTableToRow":  { "type":[ "bool" ], "default":"false" },
		"LineNo":                  { "type":[ "int" ], "default":"1" }
		}`)
}

// normally identical
func (hdlr *JSONToTableType) SetNext(next http.Handler) {
	hdlr.Next = next
}

var _ mid.GoFTLMiddleWare = (*JSONToTableType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type JSONToTableType struct {
	Next                   http.Handler //
	Paths                  []string     // Paths that match this
	ConvertRowTo1LongTable bool         //
	Convert1LongTableToRow bool         //
	LineNo                 int          //
}

// Parameterized for testing? or just change the test
func NewJSONToTableServer(n http.Handler, p []string, c, d bool) *JSONToTableType {
	return &JSONToTableType{Next: n, Paths: p, ConvertRowTo1LongTable: c, Convert1LongTableToRow: d}
}

func (hdlr *JSONToTableType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "JSONToTable", hdlr.Paths, pn, req.URL.Path)

			// fmt.Printf("AT: %s\n", godebug.LF())
			hdlr.Next.ServeHTTP(rw, req)

			if rw.StatusCode == 200 || rw.StatusCode == 0 {
				// fmt.Printf("AT: %s\n", godebug.LF())

				if rw.State == goftlmux.ByteBuffer {
					s := rw.GetBody() // peek at data, if { then hash, if [ then array, else err.
					// fmt.Printf("AT: %s, body is >>>%s<<<\n", godebug.LF(), s)

					jj := PeekAtJSONData(string(s), rw, hdlr.LineNo)
					// fmt.Printf("AT: %s body type is: %s\n", godebug.LF(), jj)
					switch jj {

					case IsJSONError:
						// fmt.Printf("AT: %s\n", godebug.LF())
						logrus.Errorf("Error in JSON processing, Empty data returned, Configuration Item Line No:%d", hdlr.LineNo)
						fallthrough
					case IsJSONEmpty:
						// fmt.Printf("AT: %s\n", godebug.LF())
						data := make(map[string]interface{})
						if hdlr.Convert1LongTableToRow {
							// fmt.Printf("AT: %s\n", godebug.LF())
							rw.ReplaceBody([]byte{})
							rw.WriteRow(data)
						} else if hdlr.ConvertRowTo1LongTable {
							// fmt.Printf("AT: %s\n", godebug.LF())
							tdata := make([]map[string]interface{}, 0, 1)
							// tdata[0] = data
							rw.ReplaceBody([]byte{})
							rw.WriteTable(tdata)
						} else {
							// fmt.Printf("AT: %s\n", godebug.LF())
							rw.ReplaceBody([]byte{})
							rw.WriteRow(data)
						}

					case IsJSONHash:
						// fmt.Printf("AT: %s\n", godebug.LF())
						data := make(map[string]interface{})

						err := json.Unmarshal([]byte(s), &data) // JSON to data
						if err != nil {
							// fmt.Printf("AT: %s\n", godebug.LF())
							// xyzzyLogrus
							logrus.Warnf("Unable to parse data in JSON, Configuration Item Line No:%d", hdlr.LineNo)
							if hdlr.Convert1LongTableToRow {
								// fmt.Printf("AT: %s\n", godebug.LF())
								rw.ReplaceBody([]byte{})
								rw.WriteRow(data)
							} else if hdlr.ConvertRowTo1LongTable {
								// fmt.Printf("AT: %s\n", godebug.LF())
								tdata := make([]map[string]interface{}, 0, 1)
								// tdata[0] = data
								rw.ReplaceBody([]byte{})
								rw.WriteTable(tdata)
							}
							www.WriteHeader(http.StatusInternalServerError)
						}

						if hdlr.Convert1LongTableToRow {
							// fmt.Printf("AT: %s\n", godebug.LF())
							rw.ReplaceBody([]byte{})
							rw.WriteRow(data)
						} else if hdlr.ConvertRowTo1LongTable {
							// fmt.Printf("AT: %s\n", godebug.LF())
							tdata := make([]map[string]interface{}, 1, 1)
							tdata[0] = data
							rw.ReplaceBody([]byte{})
							rw.WriteTable(tdata)
						} else {
							// fmt.Printf("AT: %s\n", godebug.LF())
							rw.ReplaceBody([]byte{})
							rw.WriteRow(data)
						}

					case IsJSONArray:
						// fmt.Printf("AT: %s\n", godebug.LF())
						fallthrough
					default:
						// fmt.Printf("AT: %s, s >>>%s<<<\n", godebug.LF(), s)
						data := make([]map[string]interface{}, 0, 20)

						err := json.Unmarshal([]byte(s), &data) // JSON to data
						if err != nil {
							// fmt.Printf("AT: %s\n", godebug.LF())
							logrus.Errorf("Unable to parse data in JSON, Configuration Item Line No:%d", hdlr.LineNo)
							www.WriteHeader(http.StatusInternalServerError)
						}

						if len(data) == 1 && hdlr.Convert1LongTableToRow {
							// fmt.Printf("AT: %s\n", godebug.LF())
							rw.ReplaceBody([]byte{})
							rw.WriteRow(data[0])
						} else if len(data) == 0 && hdlr.Convert1LongTableToRow {
							// fmt.Printf("AT: %s\n", godebug.LF())
							data := make(map[string]interface{})
							rw.ReplaceBody([]byte{})
							rw.WriteRow(data)
						} else {
							// fmt.Printf("AT: %s\n", godebug.LF())
							rw.ReplaceBody([]byte{})
							rw.WriteTable(data)
						}

						// fmt.Printf("AT: %s, data=%s\n", godebug.LF(), lib.SVar(data))
					}
				} else {
					logrus.Errorf("Unable to un-marshal JSON data, retuning data in JSON, Line No:%v", hdlr.LineNo)
				}
			}

		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		hdlr.Next.ServeHTTP(www, req)
	}

}

type JSONDataType int

const (
	IsJSONHash  JSONDataType = 1
	IsJSONArray JSONDataType = 2
	IsJSONOther JSONDataType = 3
	IsJSONError JSONDataType = 4
	IsJSONEmpty JSONDataType = 5
)

func (jdt JSONDataType) String() string {
	switch jdt {
	case IsJSONHash:
		return "IsJSONHash"
	case IsJSONArray:
		return "IsJSONArray"
	case IsJSONOther:
		return "IsJSONOther"
	case IsJSONError:
		return "IsJSONError"
	case IsJSONEmpty:
		return "IsJSONEmpty"
	}
	return fmt.Sprintf("*** Invalid JSONDataType = %d ***", jdt)
}

func PeekAtJSONData(s string, rw *goftlmux.MidBuffer, LineNo int) (jdt JSONDataType) {
	match, err := regexp.MatchString("^[ \t]*\\{", s)
	if err != nil {
		logrus.Warnf("Invalid regular expression, Error: %s Line No:%d", err, LineNo)
		return IsJSONError
	}
	if match {
		return IsJSONHash
	}
	match, err = regexp.MatchString("^[ \t]*\\[", s)
	if err != nil {
		logrus.Warnf("Invalid regular expression, Error: %s Line No:%d", err, LineNo)
		return IsJSONError
	}
	if match {
		return IsJSONArray
	}
	match, err = regexp.MatchString("^[ \t]*$", s)
	if err != nil {
		logrus.Warnf("Invalid regular expression, Error: %s Line No:%d", err, LineNo)
		return IsJSONError
	}
	if match {
		return IsJSONEmpty
	}
	return IsJSONOther
}

/* vim: set noai ts=4 sw=4: */
