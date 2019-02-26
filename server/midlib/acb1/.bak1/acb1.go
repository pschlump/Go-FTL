//
// Go-FTL - Module
//
// Copyright (C) Philip Schlump, 2018-2019.
//

package Acb1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	JsonX "github.com/pschlump/JSONx"
	"github.com/pschlump/godebug"
)

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &Acb1Type{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // Xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("Acb1", CreateEmpty, `{
		"Paths":        	 { "type":["string","filepath"], "isarray":true, "required":true },
		"AuthKey":  	     { "type":[ "string" ], "required":false, "default":"kip.philip" },
		"InputPath":  	     { "type":[ "string" ], "required":false, "default":"./image" },
		"OutputPath":  	     { "type":[ "string" ], "required":false, "default":"./qr-final" },
		"OutputURL":  	     { "type":[ "string" ], "required":false, "default":"/qr-final/" },
		"RedisPrefix":  	 { "type":[ "string" ], "required":false, "default":"dip:" },
		"LineNo":       	 { "type":[ "int" ], "default":"1" }
		}`)
}

//
// Used by /api/acb1/generateQrFor
// 		OutputURL -	/qr-final
// 		OutputPath - path for generation of .png/.svg QR Codes
//
// AuthKey - key used to auth calls to this.
//
// Not Used Yet -- or -- will be removed from old code:
//		"IsProd":  	         { "type":[ "string" ], "required":false, "default":"test" },
//		"RedisQ":  	     	 { "type":[ "string" ], "required":false, "default":"geth:queue:" },
//		"RedisGetQ":  	     { "type":[ "string" ], "required":false, "default":"get:queue:" },
//		"GetEventURL": 	     { "type":[ "string" ], "required":false, "default":"http://www.2c-why.com/" },
//		"RedisID": 	     	 { "type":[ "string" ], "required":false, "default":"doc:ID:" },
//		"SingedOnceAddr":  	 { "type":[ "string" ], "required":false, "default":"" },
//		"AppID":  	         { "type":[ "string" ], "required":false, "default":"100" },
//
//

func (hdlr *Acb1Type) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	gCfg.ConnectToRedis()
	gCfg.ConnectToPostgreSQL()
	hdlr.gCfg = gCfg
	return
}

func (hdlr *Acb1Type) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*Acb1Type)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type Acb1Type struct {
	Next        http.Handler                //
	Paths       []string                    //
	AuthKey     string                      // (acb)
	RedisPrefix string                      //
	InputPath   string                      //
	OutputPath  string                      //
	OutputURL   string                      //
	validEvent  map[string]bool             // list of valid events for items (acb)
	LineNo      int                         //
	gCfg        *cfg.ServerGlobalConfigType //
}

// NewAcb1TypeServer will create a copy of the server for testing.
func NewAcb1TypeServer(n http.Handler, p []string, redisPrefix, realm string) *Acb1Type {
	return &Acb1Type{
		Next:        n,
		Paths:       p,
		RedisPrefix: redisPrefix,
	}
}

type dispatchType struct {
	handlerFunc func(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string)
}

var dispatch map[string]dispatchType

func init() {
	dispatch = make(map[string]dispatchType)

	dispatch["/api/acb1/test1"] = dispatchType{
		handlerFunc: func(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
			fmt.Printf("test1 called\n")
			fmt.Fprintf(os.Stderr, "test1 called\n")
		},
	}
	dispatch["/api/acb1/track_add"] = dispatchType{
		handlerFunc: trackAdd,
	}

}

type bulkDataRow struct {
	Tag   string `json:"Tag"`    // RFIF etc. (unique)
	SubId string `json:"Sub_id"` // Used with Site_id to pull out Tag
	Event string `json:"Event"`  // One of standard set of vevents - will be validated.
	Data  string `json:"Data"`   // Additional Data in JSON format
	Date  string `json:"Date"`   // Date/Time ISO format date/time stamp.
}
type bulkDataType struct {
	SiteId string `json:"Site_id"`
	Auth   string `json:"Auth"` // AuthKey for this.
	Row    []bulkDataRow
}

// Set of results - per tag id
type bulkRvListType struct {
	Tag        string `json:"Tag"` // RFIF etc. (unique)
	SiteId     string `json:"Site_id"`
	SubId      string `json:"Sub_id"`     // Used with SiteId to pull out Tag
	ItemStatus string `json:"ItemStatus"` // Error for this
	Msg        string `json:"Msg"`        // Used with SiteId to pull out Tag
}
type bulkRvType struct {
	Status string           `json:"status"` // status of success or "partial", or "error"
	Msg    string           `json:"msg"`    // msg - if not "", then all failed.
	Detail []bulkRvListType `json:"detail"`
}

func (hdlr *Acb1Type) SetupValidEvents() {
	if hdlr.validEvent == nil {
		hdlr.validEvent = make(map[string]bool)
		hdlr.validEvent["Init"] = true
		hdlr.validEvent["Create-Event"] = true
		hdlr.validEvent["Delete-Event"] = true
		hdlr.validEvent["Update"] = true
	}
}

func (hdlr *Acb1Type) InsertTrackAdd(tag string) error {
	stmt := "insert into \"v1_trackAdd\" ( \"tag\" ) values ( $1 )"
	_, err := hdlr.gCfg.Pg_client.Db.Exec(stmt, tag)
	if err != nil {
		return err
	} else {
		fmt.Printf("Success: %s data[%s]\n", stmt, tag)
		fmt.Fprintf(os.Stderr, "Success: %s data[%s]\n", stmt, tag)
	}
	return nil
}

func trackAdd(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("trackAdd called\n")
	fmt.Fprintf(os.Stderr, "trackAdd called\n")

	ps := &rw.Ps

	bulk := ps.ByNameDflt("bulk", "")
	godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
	fmt.Printf("bulk: ->%s<-\n", bulk)
	var bulkData bulkDataType
	var err error

	godebug.DbPfb(db1, "%(Yellow)Partial Error [%s] AT: %(LF)\n", err)
	if bulk != "" {
		err = json.Unmarshal([]byte(bulk), &bulkData)
	} else {
		err = nil
		bulkData.Auth = ps.ByNameDflt("auth", "")
		bulkData.SiteId = ps.ByNameDflt("Site_id", "")
		bulkData.Row = append(bulkData.Row, bulkDataRow{
			Tag:   ps.ByNameDflt("Tag", ""),
			SubId: ps.ByNameDflt("Sub_id", ""),
			Event: ps.ByNameDflt("Event", ""),
			Data:  ps.ByNameDflt("Data", ""),
			Date:  ps.ByNameDflt("Date", ""),
		})
	}
	if err != nil {
		godebug.DbPfb(db1, "%(Red)Error [%s] AT: %(LF)\n", err)

		fmt.Fprintf(www, "%s", godebug.SVarI(bulkRvType{
			Status: "failed",
			Msg:    fmt.Sprintf("error - falied to supply needed data for processing [%s].", err),
		}))
		return
	}

	godebug.DbPfb(db1, "%(Yellow)Partial Error [%s] AT: %(LF)\n", err)
	if hdlr.AuthKey != "" && bulkData.Auth != hdlr.AuthKey {
		err = fmt.Errorf("Invalid auth key")
	}
	if err != nil {
		godebug.DbPfb(db1, "%(Red)Error [%s] AT: %(LF)\n", err)

		fmt.Fprintf(www, "%s", godebug.SVarI(bulkRvType{
			Status: "failed",
			Msg:    fmt.Sprintf("error - falied to supply needed data for processing [%s].", err),
		}))
		return
	}

	godebug.DbPfb(db1, "%(Yellow)Partial Error [%s] AT: %(LF)\n", err)
	var rv bulkRvType
	statusVal := "success"

	fmt.Printf("Processing ->%s<-\n", godebug.SVarI(bulkData))

	hdlr.SetupValidEvents()
	for _, rr := range bulkData.Row {
		if _, ok := hdlr.validEvent[rr.Event]; !ok {
			rv.Detail = append(rv.Detail, bulkRvListType{
				Tag:        rr.Tag,
				SiteId:     bulkData.SiteId,
				SubId:      rr.SubId,
				ItemStatus: "error",
				Msg:        fmt.Sprintf("Invalid Event Type [%s]", rr.Event),
			})
			statusVal = "partial"
			err = nil
		} else {
			rv.Detail = append(rv.Detail, bulkRvListType{
				Tag:        rr.Tag,
				SiteId:     bulkData.SiteId,
				SubId:      rr.SubId,
				ItemStatus: "success",
			})
		}
	}

	godebug.DbPfb(db1, "%(Yellow)AT: %(LF)\n")
	fmt.Fprintf(os.Stdout, "rv = %s\n", godebug.SVarI(rv))
	for ii, rr := range bulkData.Row {
		godebug.DbPfb(db1, "%(Yellow)AT: %(LF)\n")
		if rv.Detail[ii].ItemStatus == "success" {
			godebug.DbPfb(db1, "%(Yellow)AT: %(LF)\n")
			if rr.Tag == "" && rr.SubId != "" {
				// xyzzy100 - pull out Tag id or error -- If error set ItemStatus to...
				// xyzzy - Call convSiteIDToTagId ( site_id, sub_id ) -> tagId, err
				// xyzzy - if error ...
			}
		}
		if rv.Detail[ii].ItemStatus == "success" {
			godebug.DbPfb(db1, "%(Yellow)AT: %(LF)\n")
			err = hdlr.InsertTrackAdd(rr.Tag) // xyzzy - other params to pass!
			if err != nil {
				statusVal = "partial"
				rv.Detail[ii].ItemStatus = "error"
				rv.Detail[ii].Msg = fmt.Sprintf("%s", err)
				err = nil
			}
		}
	}

	if err != nil {
		godebug.DbPfb(db1, "%(Red)Error [%s] AT: %(LF)\n", err)

		fmt.Fprintf(www, "%s", godebug.SVarI(bulkRvType{
			Status: "failed",
			Msg:    "error - falied to supply needed data for processing.",
		}))
		return
	}

	if statusVal != "success" {
		rv.Status = statusVal
		godebug.DbPfb(db1, "%(Yellow)Partial Error [%s] AT: %(LF)\n", err)

		fmt.Fprintf(www, "%s", godebug.SVarI(rv))
		return
	}

	fmt.Fprintf(www, "%s", godebug.SVarI(bulkRvType{
		Status: "success",
	}))
}

func (hdlr *Acb1Type) ServeHTTP(www http.ResponseWriter, req *http.Request) {

	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			hdlr.SetupServer()
			www.Header().Set("Content-Type", "application/json")

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "Acb1", hdlr.Paths, pn, req.URL.Path)

			ps := &rw.Ps
			data := ps.ByNameDflt("Data", "{}")
			var mdata map[string]string
			err := json.Unmarshal([]byte(data), &mdata)
			if err != nil {
				fmt.Fprintf(www, "{\"status\":\"error\",\"msg\":%q}", err)
				return
			}

			godebug.DbPfb(db1, "%(Yellow)Partial Error [%s] AT: %(LF)\n", err)

			fx, ok := dispatch[req.URL.Path]
			if !ok {
				godebug.DbPfb(db1, "%(Red)Error Path Invalid [%s] AT: %(LF)\n", req.URL.Path)

				fmt.Fprintf(www, "{\"status\":\"not-implemented-yet\"}")
				return
			}
			fx.handlerFunc(hdlr, rw, www, req, mdata)
			return

			fmt.Fprintf(www, "{\"status\":\"not-implemented-yet\"}")
		}
	}

	hdlr.Next.ServeHTTP(www, req)
}

const db1 = false

/* vim: set noai ts=4 sw=4: */