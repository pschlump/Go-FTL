//
// Go-FTL - Module
//
// Copyright (C) Philip Schlump, 2018-2019.
//

package Acb1

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"

	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/sizlib"
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
		"DisplayURL":	  	 { "type":[ "string" ], "required":false, "default":"http://www.2c-why.com/demo34" },
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

//
// http://t432z.com/dec/5c -> http://test.test.com
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
	DisplayURL  string                      // URL to display results - destination of QR redirect
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
	dispatch["/api/acb1/listBy"] = dispatchType{
		handlerFunc: listBy,
	}
	dispatch["/api/acb1/generateQrFor"] = dispatchType{
		handlerFunc: generateQrFor,
	}
	dispatch["/api/acb1/getTagId"] = dispatchType{
		handlerFunc: getTagId,
	}
	dispatch["/api/acb1/getInfo"] = dispatchType{
		handlerFunc: getInfo,
	}
	dispatch["/api/acb1/convToJson"] = dispatchType{
		handlerFunc: convToJson,
	}
	dispatch["/api/acb1/chainHash"] = dispatchType{
		handlerFunc: chainHash,
	}
	dispatch["/api/acb1/login_demo"] = dispatchType{
		handlerFunc: loginDemo,
	}

}

type bulkDataRow struct {
	Tag   string `json:"Tag"`    // RFIF etc. (unique)
	Note  string `json:"Notes"`  // User memo
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
		hdlr.validEvent["1"] = true
		hdlr.validEvent["2"] = true
		hdlr.validEvent["3"] = true
		hdlr.validEvent["4"] = true
		hdlr.validEvent["5"] = true
		hdlr.validEvent["6"] = true
	}
}

type DataToBeHashed struct {
	Tag      string
	PrevHash string `json:"prev_hash"`
	Created  string
	Note     string
}

type DataSetHashed []DataToBeHashed

func SerializeDataToBeHashed(dt DataToBeHashed) []byte {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, int32(len(dt.Tag)))
	buf.Write([]byte(dt.Tag))

	binary.Write(&buf, binary.BigEndian, int32(len(dt.PrevHash)))
	buf.Write([]byte(dt.PrevHash))

	binary.Write(&buf, binary.BigEndian, int32(len(dt.Created)))
	buf.Write([]byte(dt.Created))

	binary.Write(&buf, binary.BigEndian, int32(len(dt.Note)))
	buf.Write([]byte(dt.Note))

	return buf.Bytes()
}

func SerializeAnimal(dt DataSetHashed) []byte {
	var buf bytes.Buffer
	for _, row := range dt {
		rowS := SerializeDataToBeHashed(row)
		buf.Write(rowS)
	}
	return buf.Bytes()
}

func (hdlr *Acb1Type) InsertTrackAdd(tag, note string) (string, error) {
	hash_prev, max_ord_seq, qr_enc_id, err := hdlr.GetMostRecentHash(tag)
	godebug.DbPfb(db1, "%(Yellow) AT: %(LF) %(Red) hash_prev [%s] max_ord_seq [%s]\n", hash_prev, max_ord_seq)
	stmt := "insert into \"v1_trackAdd\" ( \"tag\", \"note\", \"prev_hash\" ) values ( $1, $2, $3 )"
	_, err = hdlr.gCfg.Pg_client.Db.Exec(stmt, tag, note, hash_prev)
	if err != nil {
		return qr_enc_id, err
	} else {
		fmt.Printf("Success: %s data[%s, %s, %s]\n", stmt, tag, note, hash_prev)
		fmt.Fprintf(os.Stderr, "Success: %s data[%s, %s, %s]\n", stmt, tag, note, hash_prev)
	}
	// 	1. Pull back data (including create date, hash)
	godebug.DbPfb(db1, "%(Yellow) AT: %(LF) hash_prev [%s] max_ord_seq [%s]\n", hash_prev, max_ord_seq)
	data, err := hdlr.GetAllRows(tag)
	// 	2. Put into data type
	godebug.DbPfb(db1, "%(Yellow) AT: %(LF) data ->%s<-\n", data)
	var set DataSetHashed
	err = json.Unmarshal([]byte(data), &set)
	// 	3. Serialize it
	godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
	serDat := SerializeAnimal(set)
	// 	4. Keccak256 hash it
	godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
	newHash := KeeackHash([]byte(serDat))
	// 	5. update it.
	godebug.DbPfb(db1, "%(Yellow) AT: %(LF) newHash [%x], qr_enc_id [%s]\n", newHash, qr_enc_id)
	hdlr.UpdateHashCurrentRow(tag, max_ord_seq, string(newHash))
	return qr_enc_id, nil
}

// data, err := hdlr.GetAllRows(tag)
func (hdlr *Acb1Type) GetAllRows(tag string) (rowData string, err error) {
	stmt :=
		`select t1.*
			, t2."file_name"
			, t2."url_path"
			, t2."qr_id"
			, t2."qr_enc_id"
			, t2."state" as "qr_state"
		from "v1_trackAdd" as t1 left outer join "v1_avail_qr" as t2 on t1."qr_id" = t2."qr_enc_id"
		where "tag" = $1
		order by "ord_seq" desc
		`
	Rows, err := hdlr.gCfg.Pg_client.Db.Query(stmt, tag)
	if err != nil {
		fmt.Printf("Database error %s. stmt=%s data=[%s]\n", err, stmt, tag)
		// fmt.Fprintf(www, `{"status":"error","msg":"database error: [%v]"}`, tag)
		return
	}

	defer Rows.Close()
	finalData, _, _ := sizlib.RowsToInterface(Rows)

	return sizlib.SVar(finalData), nil
}

// hdlr.UpdateHashCurrentRow(tag, max_ord_seq, newHash)
func (hdlr *Acb1Type) UpdateHashCurrentRow(tag, max_ord_seq, newHash string) (err error) {
	stmt := "update \"v1_trackAdd\" set \"hash\" = $1 where \"tag\" = $2 and \"ord_seq\" > $3"
	godebug.DbPfb(db1, "%(Yellow) AT: %(LF) stmt=[%s] data=[%x, %s, %s]\n", stmt, newHash, tag, max_ord_seq)
	_, err = hdlr.gCfg.Pg_client.Db.Exec(stmt, fmt.Sprintf("%x", newHash), tag, max_ord_seq)
	if err != nil {
		godebug.DbPfb(db1, "%(Yellow) AT: %(LF) err=%s\n", err)
		return err
	} else {
		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
		fmt.Printf("Success: %s data[%x, %s, %s]\n", stmt, newHash, tag, max_ord_seq)
		fmt.Fprintf(os.Stderr, "Success: %s data[%x, %s, %s]\n", stmt, newHash, tag, max_ord_seq)
	}
	godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
	return nil
}

func (hdlr *Acb1Type) GetMostRecentHash(tag string) (hash, ord_seq, qr_enc_id string, err error) {
	stmt := "select \"hash\", \"ord_seq\", \"qr_id\" from \"v1_trackAdd\" where \"tag\" = $1 order by \"ord_seq\" desc"

	rows, err := hdlr.gCfg.Pg_client.Db.Query(stmt, tag)
	if err != nil {
		fmt.Printf("Database error %s, attempting to convert premis_id/animal_id to tag.\n", err)
		return "", "0", "", err
	}
	godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
	for nr := 0; rows.Next(); nr++ {
		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")

		var hash, max_ord_seq, latest_qr_enc_id string
		err := rows.Scan(&hash, &max_ord_seq, &latest_qr_enc_id)
		if err != nil {
			fmt.Printf("Error on d.b. query %s\n", err)
			return "", "0", "", err
		}
		godebug.DbPfb(db1, "%(Yellow) AT: %(LF) latest_qr_enc_id [%s]\n", latest_qr_enc_id)

		return hash, max_ord_seq, latest_qr_enc_id, nil
	}
	return "", "0", "", nil
}

func (hdlr *Acb1Type) UpdateQRMarkAsUsed(qrId string) error {
	stmt := "update \"v1_avail_qr\" set \"state\" = 'used' where \"qr_enc_id\" = $1"
	godebug.DbPfb(db1, "%(Yellow) AT: %(LF) - stmt [%s] data[%s]\n", stmt, qrId)
	_, err := hdlr.gCfg.Pg_client.Db.Exec(stmt, qrId)
	if err != nil {
		return err
	} else {
		fmt.Printf("Success: %s data[%s]\n", stmt, qrId)
		fmt.Fprintf(os.Stderr, "Success: %s data[%s]\n", stmt, qrId)
	}
	return nil
}

func (hdlr *Acb1Type) UpdateAnimalWithQR(tag, qrId string) error {
	stmt := "update \"v1_trackAdd\" set \"qr_id\" = $1 where \"tag\" = $2"
	_, err := hdlr.gCfg.Pg_client.Db.Exec(stmt, qrId, tag)
	if err != nil {
		return err
	} else {
		fmt.Printf("Success: %s data[%s, %s]\n", stmt, qrId, tag)
		fmt.Fprintf(os.Stderr, "Success: %s data[%s, %s]\n", stmt, qrId, tag)
	}
	return nil
}

// err = hdlr.PullQRFromDB(rr.Tag)
func (hdlr *Acb1Type) PullQRFromDB(tag string) (qr_enc_id string, err error) {
	// Xyzzy - sould replace with stored proc. that updates state in same transaction.
	stmt := "select \"qr_enc_id\" from \"v1_avail_qr\" where \"state\" = 'avail' limit 1"
	// insert into "v1_avail_qr" ( "qr_id", "qr_enc_id", "url_path", "file_name", "qr_encoded_url_path" ) values
	// 	  ( '170', '4q', 'http://127.0.0.1:9019/qr/00170.4.png', './td_0008/q00170.4.png', 'http://t432z.com/q/4q' )
	rows, err := hdlr.gCfg.Pg_client.Db.Query(stmt)
	if err != nil {
		fmt.Printf("Database error %s, attempting to convert premis_id/animal_id to tag.\n", err)
		return "", err
	}
	godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
	for nr := 0; rows.Next(); nr++ {
		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
		if nr >= 1 {
			fmt.Printf("Error too many rows for a user, should be unique primary key\n")
			break
		}

		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
		var qr string
		err := rows.Scan(&qr)
		if err != nil {
			fmt.Printf("Error on d.b. query %s\n", err)
			return "", err
		}
		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")

		// Xyzzy - test fail to error report
		err = hdlr.UpdateQRMarkAsUsed(qr)
		if err != nil {
			fmt.Printf("Error on d.b. query %s\n", err)
			return "", err
		}

		godebug.DbPfb(db1, "%(Yellow) AT: %(LF)\n")
		return qr, nil
	}
	return "", fmt.Errorf("Failed to get a QR code")
}

func FindTagId(hdlr *Acb1Type, premis_id, premis_animal string) (string, error) {
	stmt := "select \"tag\" from \"v1_trackAdd\" where \"premis_id\" = $1	and \"premis_animal\" = $2 limit 1"
	rows, err := hdlr.gCfg.Pg_client.Db.Query(stmt, premis_id, premis_animal)
	if err != nil {
		fmt.Printf("Database error %s, attempting to convert premis_id/animal_id to tag.\n", err)
		return "", err
	}

	for nr := 0; rows.Next(); nr++ {
		if nr >= 1 {
			fmt.Printf("Error too many rows for a user, should be unique primary key\n")
			break
		}

		var tag string
		err := rows.Scan(&tag)
		if err != nil {
			fmt.Printf("Error on d.b. query %s\n", err)
			return "", err
		}

		return tag, nil
	}
	return "", fmt.Errorf("Unable to use premis_id/animal_id to identify unique animal")
}

// --------------------------------------------------------------
// cowData := GetCowDisplayData(rr.Tag)	// xyzzy
// --------------------------------------------------------------
/*
{
	"Cow": "rfid: 111"
, 	"Ranch": "Ranch: RC"
,	"SlaughterHouse": "Wyoming Beef Lab"
,	"Aged": "Aged: 14 days"
,	"DryAged": ""
}
*/
type DataDisplayType struct {
	Cow            string
	Ranch          string
	SlaughterHouse string
	Aged           string
	DryAged        string
}

func (hdlr *Acb1Type) GetCowDisplayData(tag string) (dataJsonString string, err error) {
	var dd DataDisplayType
	dd.Cow = "rfid: " + tag
	dd.Ranch = "Wyoming: Demo Ranch"
	dd.SlaughterHouse = "Wyoming: Beef Lab"
	dd.Aged = "Aged Beef"
	dd.DryAged = "Dry Aged"

	// xyzzy
	// xyzzy
	// xyzzy107 - test QR setup on t432z.com

	dataJsonString = godebug.SVarI(dd)
	return
}

/*
// Setup QR Redirect

	export QR_SHORT_AUTH_TOKEN="w4h0wvtb1zk4uf8Xv.Ns9Q7j8"
	wget -o out/,list1 -O out/,list2 \
		--header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" \
		"http://t432z.com/upd/?url=http://test.test.com&id=5c"

	-- 1. DoGet - change to create a header
	-- 2. Example Call to set this
*/
func trackAdd(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("trackAdd called\n")
	fmt.Fprintf(os.Stderr, "trackAdd called\n")

	ps := &rw.Ps

	bulk := ps.ByNameDflt("bulk", "")
	godebug.DbPfb(db1, "bulk: ->%s<-\n", bulk)
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
			Note:  ps.ByNameDflt("Note", ""),
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
				// Xyzzy100 - pull out Tag id or error -- If error set ItemStatus to...
				// Xyzzy - Call convSiteIDToTagId ( site_id, sub_id ) -> tagId, err
				// Xyzzy - if error ...
				rr.Tag, err = FindTagId(hdlr, bulkData.SiteId, rr.SubId)
			}
		}
		qrId := ""
		if rv.Detail[ii].ItemStatus == "success" {
			godebug.DbPfb(db1, "%(Cyan)AT: %(LF)\n")
			// xyzzy104 - premis_id/animal_id etc.  // xyzzy - other params to pass! --
			qrId, err = hdlr.InsertTrackAdd(rr.Tag, rr.Note)
			if err != nil {
				statusVal = "partial"
				rv.Detail[ii].ItemStatus = "error"
				rv.Detail[ii].Msg = fmt.Sprintf("%s", err)
				err = nil
			}
		}
		godebug.DbPfb(db1, "%(Cyan)AT: %(LF) qrId [%s]\n", qrId)
		if rv.Detail[ii].ItemStatus == "success" && qrId == "" {
			godebug.DbPfb(db1, "%(Cyan)AT: %(LF)\n")
			qrId, err = hdlr.PullQRFromDB(rr.Tag)
			if err != nil {
				statusVal = "partial"
				rv.Detail[ii].ItemStatus = "error"
				rv.Detail[ii].Msg = fmt.Sprintf("%s", err)
				err = nil
			}
			// pull out/update preped - QR from d.b.
			// get the next avail QR code
			//  	1. pull from d.b.
			// 	 	2. update d.b. to mark as used.
			// 	 	(3 below). update row about animal to show use of QR.
		}
		if rv.Detail[ii].ItemStatus == "success" {
			// test QR setup on t432z.com - update the redirect for QR code
			ran := fmt.Sprintf("%d", rand.Intn(1000000000))
			godebug.DbPfb(db1, "%(Cyan)AT: %(LF) ran [%v]\n", ran)
			cowData, err := hdlr.GetCowDisplayData(rr.Tag)
			if err != nil {
				statusVal = "partial"
				rv.Detail[ii].ItemStatus = "error"
				rv.Detail[ii].Msg = fmt.Sprintf("Failed to set QR Redirect for - failed to get data for %s, error %s", qrId, err)
				err = nil
			} else {
				godebug.DbPfb(db1, "%(Cyan)AT: %(LF) ran [%v]\n", ran)
				// t432z.com - URL from config???
				status, body := DoGet("http://t432z.com/upd/", "url", hdlr.DisplayURL, "id", qrId, "data", cowData, "_ran_", ran)
				if status != 200 {
					statusVal = "partial"
					rv.Detail[ii].ItemStatus = "error"
					rv.Detail[ii].Msg = fmt.Sprintf("Failed to set QR Redirect for %s", qrId)
					err = nil
				} else {
					godebug.DbPfb(db1, "%(Cyan)AT: %(LF)\n")
					fmt.Printf("body ->%s<-\n", body)
				}
			}
		}
		if rv.Detail[ii].ItemStatus == "success" {
			godebug.DbPfb(db1, "%(Cyan)AT: %(LF)\n")
			// pull out/update preped - QR from d.b.
			// 	 	2. update d.b. to mark as used.
			// 	 	3. update row about animal to show use of QR.
			err = hdlr.UpdateAnimalWithQR(rr.Tag, qrId)
			if err != nil {
				statusVal = "partial"
				rv.Detail[ii].ItemStatus = "error"
				rv.Detail[ii].Msg = fmt.Sprintf("%s", err)
				err = nil
			}
		}
		if rv.Detail[ii].ItemStatus == "success" {
			godebug.DbPfb(db1, "%(Green)AT: %(LF)\n")
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

/*
List Query
select t1.*
	, t2."file_name"
	, t2."url_path"
	, t2."qr_id"
	, t2."qr_enc_id"
	, t2."state" as "qr_state"
from "v1_trackAdd" as t1 left outer join "v1_avail_qr" as t2 on t1."qr_id" = t2."qr_enc_id"
;
*/
func listBy(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("listBy called\n")
	fmt.Fprintf(os.Stderr, "listBy called\n")

	stmt :=
		`select t1.*
			, t2."file_name"
			, t2."url_path"
			, t2."qr_id"
			, t2."qr_enc_id"
			, t2."state" as "qr_state"
		from "v1_trackAdd" as t1 left outer join "v1_avail_qr" as t2 on t1."qr_id" = t2."qr_enc_id"
		`
	_ = stmt

	ps := &rw.Ps

	typ := ps.ByNameDflt("typ", "cow")
	dat := ""
	switch typ {
	case "cow":
		stmt += "where t1.\"tag\" = $1\norder by t1.\"tag\" asc\n"
		dat = ps.ByNameDflt("tag", "$err$")
	case "ranch":
		stmt += "where t1.\"ranch_name\" = $1\norder by t1.\"ranch_name\" asc\n"
		dat = ps.ByNameDflt("ranch", "$err$")
	case "locaiton":
		stmt += "where t1.\"location\" = $1\n"
		dat = ps.ByNameDflt("location", "$err$")
	case "premis_id", "site_id":
		stmt += "where t1.\"premis_id\" = $1\n"
		dat = ps.ByNameDflt("premis_id", "$err$")
	}
	if dat == "$err$" {
		fmt.Printf("Missing data\n")
		fmt.Fprintf(www, `{"status":"error","msg":"database error: [%s]"}`, "missing data")
		return
	}

	Rows, err := hdlr.gCfg.Pg_client.Db.Query(stmt, dat)
	if err != nil {
		fmt.Printf("Database error %s. stmt=%s data=[%s]\n", err, stmt, dat)
		fmt.Fprintf(www, `{"status":"error","msg":"database error: [%v]"}`, err)
		return
	}

	defer Rows.Close()
	rowData, _, _ := sizlib.RowsToInterface(Rows)

	fmt.Fprintf(www, `{"status":"success","data":%s}`, godebug.SVarI(rowData))
}

func generateQrFor(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("generateQrFor called\n")
	fmt.Fprintf(os.Stderr, "generateQrFor called\n")

	// -- xyzzy - change to just pull back QR infor for tag - select.
	stmt := "select v1_next_avail_qr as \"x\""
	_ = stmt

	// TODO - call function, return x

	fmt.Fprintf(www, `{"status":"success"}`)
}

func getTagId(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("getTagId called\n")
	fmt.Fprintf(os.Stderr, "getTagId called\n")

	// TODO - convert a premis_id/sub_id -> tag id and return

	fmt.Fprintf(www, `{"status":"success"}`)
}

// getInfo will get all the info on a cow.
// Example: http://127.0.0.1:9019/api/acb1/getInfo?api_table_key=kip.philip&tag=5234321412419
func getInfo(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("getInfo called\n")
	fmt.Fprintf(os.Stderr, "getInfo called\n")

	ps := &rw.Ps

	tag := ps.ByNameDflt("tag", "")

	stmt :=
		`select t1.*
			, t2."file_name"
			, t2."url_path"
			, t2."qr_id"
			, t2."qr_enc_id"
			, t2."state" as "qr_state"
		from "v1_trackAdd" as t1 left outer join "v1_avail_qr" as t2 on t1."qr_id" = t2."qr_enc_id"
		where "tag" = $1
		order by "ord_seq" desc
		`
	Rows, err := hdlr.gCfg.Pg_client.Db.Query(stmt, tag)
	if err != nil {
		fmt.Printf("Database error %s. stmt=%s data=[%s]\n", err, stmt, tag)
		fmt.Fprintf(www, `{"status":"error","msg":"database error: [%v]"}`, tag)
		return
	}

	defer Rows.Close()
	rowData, _, _ := sizlib.RowsToInterface(Rows)

	fmt.Fprintf(www, `{"status":"success","data":%s}`, godebug.SVarI(rowData))
}

func convToJson(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("convToJson called\n")
	fmt.Fprintf(os.Stderr, "convToJson called\n")

	// TODO -- get all the info on a cow and convert to JSON and return

	fmt.Fprintf(www, `{"status":"success"}`)
}

func chainHash(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("chainHash called\n")
	fmt.Fprintf(os.Stderr, "chainHash called\n")

	fmt.Fprintf(www, `{"status":"success"}`)
}

func loginDemo(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("loginDemo called\n")
	fmt.Fprintf(os.Stderr, "loginDemo called\n")

	ps := &rw.Ps
	// <input class="form-control" name="username" type="text">
	// <input class="form-control" name="password" type="text">
	// <input class="form-control" name="val2fa" type="text">
	Un := ps.ByNameDflt("username", "")
	Pw := ps.ByNameDflt("password", "")
	X2fa := ps.ByNameDflt("val2fa", "")
	if Un == "philip" && Pw == "1234" && X2fa == "1234" {
		fmt.Fprintf(www, `{"status":"success","auth_token":"Xa-1jf2a-8891421-22a","jwt_token":"782389822718139823213873827827283782738238273827827382378237823827382738273823"}`)
	} else {
		fmt.Fprintf(www, `{"status":"failed"}`)
	}
}

//func listBy(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
//	fmt.Printf("listBy called\n")
//	fmt.Fprintf(os.Stderr, "listBy called\n")
//
//	fmt.Fprintf(www, `{"status":"success"}`)
//}

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

// Modified to send Header!
/*
---------------------------------------------
// Xyzzy101 - Setup QR Redirect
---------------------------------------------

	export QR_SHORT_AUTH_TOKEN="w4h0wvtb1zk4uf8Xv.Ns9Q7j8"
	wget -o out/,list1 -O out/,list2 \
		--header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" \
		"http://t432z.com/upd/?url=http://test.test.com&id=5c"

	-- 1. DoGet - change to create a header
	-- 2. Example Call to set this
*/
func DoGet(uri string, args ...string) (status int, rv string) {

	sep := "?"
	var qq bytes.Buffer
	qq.WriteString(uri)
	for ii := 0; ii < len(args); ii += 2 {
		// q = q + sep + name + "=" + value;
		qq.WriteString(sep)
		qq.WriteString(url.QueryEscape(args[ii]))
		qq.WriteString("=")
		if ii < len(args) {
			qq.WriteString(url.QueryEscape(args[ii+1]))
		}
		sep = "&"
	}
	url_q := qq.String()

	// res, err := http.Get(url_q)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url_q, nil)
	req.Header.Add("User-Agent", "Go-FTL-acb1")
	req.Header.Add("X-Qr-Auth", "w4h0wvtb1zk4uf8Xv.Ns9Q7j8") // Xyzzy - set from config?
	res, err := client.Do(req)

	if err != nil {
		return 500, ""
	} else {
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return 500, ""
		}
		status = res.StatusCode
		if status == 200 {
			rv = string(body)
		}
		return
	}
}

func KeeackHash(b []byte) []byte {
	d := sha3.NewKeccak256()
	d.Write(b)
	return d.Sum(nil)
}

const db1 = true

/* vim: set noai ts=4 sw=4: */
