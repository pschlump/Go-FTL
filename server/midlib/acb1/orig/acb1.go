//
// Go-FTL - Module
//
// Copyright (C) Philip Schlump, 2018-2019.
//

package Acb1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	JsonX "github.com/pschlump/JSONx"
	"github.com/pschlump/MiscLib"
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
		"AuthKey":  	     { "type":[ "string" ], "required":false, "default":"" },
		"RedisPrefix":  	 { "type":[ "string" ], "required":false, "default":"dip:" },
		"InputPath":  	     { "type":[ "string" ], "required":false, "default":"./image" },
		"OutputPath":  	     { "type":[ "string" ], "required":false, "default":"./qr-final" },
		"OutputURL":  	     { "type":[ "string" ], "required":false, "default":"/qr-final/" },
		"ArchiveURL":  	     { "type":[ "string" ], "required":false, "default":"/archive/" },
		"IsProd":  	         { "type":[ "string" ], "required":false, "default":"test" },
		"RedisQ":  	     	 { "type":[ "string" ], "required":false, "default":"geth:queue:" },
		"RedisGetQ":  	     { "type":[ "string" ], "required":false, "default":"get:queue:" },
		"GetEventURL": 	     { "type":[ "string" ], "required":false, "default":"http://www.2c-why.com/" },
		"RedisID": 	     	 { "type":[ "string" ], "required":false, "default":"doc:ID:" },
		"SingedOnceAddr":  	 { "type":[ "string" ], "required":false, "default":"" },
		"AppID":  	         { "type":[ "string" ], "required":false, "default":"100" },
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
//		"RedisPrefix":  	 { "type":[ "string" ], "required":false, "default":"dip:" },
//		"InputPath":  	     { "type":[ "string" ], "required":false, "default":"./image" },
//		"ArchiveURL":  	     { "type":[ "string" ], "required":false, "default":"/archive/" },
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
	Next           http.Handler                //
	Paths          []string                    //
	AuthKey        string                      // (acb)
	RedisPrefix    string                      //
	InputPath      string                      //
	OutputPath     string                      //
	OutputURL      string                      //
	ArchivePath    string                      //
	IsProd         string                      //
	RedisQ         string                      // Q that push to Geth is put on.
	RedisGetQ      string                      // Signal to outside world that data is ready. ((incomplete))
	GetEventURL    string                      // URL to do GET on to signal that data is ready. ((incomplete))
	RedisID        string                      // ID to increment for temp file names
	SingedOnceAddr string                      // Address of loaded proxy contract
	AppID          string                      // ID of this app
	validEvent     map[string]bool             // list of valid events for items (acb)
	LineNo         int                         //
	gCfg           *cfg.ServerGlobalConfigType //
}

// NewAcb1TypeServer will create a copy of the server for testing.
func NewAcb1TypeServer(n http.Handler, p []string, redisPrefix, realm string) *Acb1Type {
	return &Acb1Type{
		Next:        n,
		Paths:       p,
		RedisPrefix: redisPrefix,
	}
}

type imageTypeList struct {
	Pos   int
	Name  string
	found bool
}

type documentType struct {
	Title     string
	Desc      string
	Category  string
	Tags      string
	ImageList string
}

type metaDocumentType struct {
	Document      documentType
	DocumentID    string
	CategoryID    string
	OverallHash   []byte
	MerkleHash    []byte
	PdfHash       []byte
	ImageListHash []string
	LeafHash      []string
}

type dispatchType struct {
	handlerFunc func(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request)
}

var dispatch map[string]dispatchType

func init() {
	dispatch = make(map[string]dispatchType)

	dispatch["/api/acb1/test1"] = dispatchType{
		handlerFunc: func(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request) {
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
		hdlr.validEvent["Update"] = true
	}
}

func (hdlr *Acb1Type) InsertTrackAdd(tag string) error {
	stmt := "insert into \"track_animal\" ( \"tag\" ) values ( $1 )"
	_, err := hdlr.gCfg.Pg_client.Db.Exec(stmt, tag)
	if err != nil {
		return err
	}
	return nil
}

func trackAdd(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request) {
	fmt.Printf("trackAdd called\n")
	fmt.Fprintf(os.Stderr, "trackAdd called\n")

	ps := &rw.Ps

	bulk := ps.ByNameDflt("bulk", "")
	var bulkData bulkDataType
	var err error

	if bulk == "" {
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

	if err != nil && hdlr.AuthKey != "" && bulkData.Auth == hdlr.AuthKey {
		err = fmt.Errorf("Invalid auth key")
	}

	var rv bulkRvType
	statusVal := "success"

	hdlr.SetupValidEvents()
	for _, rr := range bulkData.Row {
		if _, ok := hdlr.validEvent[rr.Event]; !ok {
			rv.Detail = append(rv.Detail, bulkRvListType{
				Tag:        rr.Tag,
				SiteId:     bulkData.SiteId,
				SubId:      rr.SubId,
				ItemStatus: "error",
				Msg:        "Invalid Event Type",
			})
			statusVal = "partial"
			err = nil
		} else {
			rv.Detail = append(rv.Detail, bulkRvListType{
				Tag:        rr.Tag,
				SiteId:     bulkData.SiteId,
				SubId:      rr.SubId,
				ItemStatus: "sucess",
			})
		}
	}

	for ii, rr := range bulkData.Row {
		if rv.Detail[ii].ItemStatus == "success" {
			if rr.Tag == "" && rr.SubId != "" {
				// xyzzy - pull out Tag id or error -- If error set ItemStatus to...
				// xyzzy - Call convSiteIDToTagId ( site_id, sub_id ) -> tagId, err
				// xyzzy - if error ...
			}
		}
		if rv.Detail[ii].ItemStatus == "success" {
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
		fmt.Fprintf(os.Stderr, "%sError[%s] AT: %s%s\n", MiscLib.ColorRed, err, godebug.LF(), MiscLib.ColorReset)
		fmt.Fprintf(os.Stdout, "%sError[%s] AT: %s%s\n", MiscLib.ColorRed, err, godebug.LF(), MiscLib.ColorReset)

		fmt.Fprintf(www, "%s", godebug.SVarI(bulkRvType{
			Status: "failed",
			Msg:    "error - falied to supply needed data for processing.",
		}))
		return
	}

	if statusVal != "success" {
		rv.Status = statusVal
		fmt.Fprintf(os.Stderr, "%sPartial Error[%s] AT: %s%s\n", MiscLib.ColorRed, err, godebug.SVarI(rv), MiscLib.ColorReset)
		fmt.Fprintf(os.Stdout, "%sPartial Error[%s] AT: %s%s\n", MiscLib.ColorRed, err, godebug.SVarI(rv), MiscLib.ColorReset)

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

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "Acb1", hdlr.Paths, pn, req.URL.Path)

			ps := &rw.Ps
			www.Header().Set("Content-Type", "application/json")

			fmt.Fprintf(os.Stderr, "%sAT: %s%s\n", MiscLib.ColorGreen, godebug.LF(), MiscLib.ColorReset)
			fmt.Fprintf(os.Stdout, "%sAT: %s%s\n", MiscLib.ColorGreen, godebug.LF(), MiscLib.ColorReset)

			if true {

				hdlr.SetupServer()

				fx, ok := dispatch[req.URL.Path]
				if !ok {
					fmt.Fprintf(os.Stderr, "%sInvalid Path[%s] AT: %s%s\n", MiscLib.ColorRed, req.URL.Path, godebug.LF(), MiscLib.ColorReset)
					fmt.Fprintf(os.Stdout, "%sInvalid Path[%s] AT: %s%s\n", MiscLib.ColorRed, req.URL.Path, godebug.LF(), MiscLib.ColorReset)

					fmt.Fprintf(www, "{\"status\":\"not-implemented-yet\"}")
					return
				}
				fx.handlerFunc(hdlr, rw, www, req)
				return

			} else {

				// ------------------------------------------------------------------------------------------------
				// 0. pull params - convert to internal format - validate
				// 		1. Check that all the images are in place - validate them.  If not then ?? error ??
				//		param["tags"] = tTags;
				//		param["title"] = tTitle;
				//		param["desc"] = tDesc;
				//		param["category_id"] = category_id;
				//		param["imageList"] = JSON.stringify ( imgListSelected );
				//		param["id"] = data.id; // In JSON format!!
				tags := ps.ByNameDflt("tags", "")
				title := ps.ByNameDflt("title", "")
				desc := ps.ByNameDflt("desc", "")
				categoryID := ps.ByNameDflt("category_id", "")
				category := ps.ByNameDflt("category", "")
				documentID := ps.ByNameDflt("id", "")
				imageListJSON := ps.ByNameDflt("image_list", "[]")

				TheDoc := documentType{
					Title:     title,
					Desc:      desc,
					Category:  category,
					Tags:      tags,
					ImageList: imageListJSON,
				}

				var imageList []imageTypeList
				err := json.Unmarshal([]byte(imageListJSON), &imageList)
				if err != nil {
					fmt.Fprintf(www, "{\"status\":\"error\",\"msg\":\"unable to parse list of images.\"}\n")
					return
				}

				fns, _ := GetFilenames(hdlr.InputPath)
				missingFn := []string{}
				for ii, need := range imageList {
					fmt.Fprintf(os.Stderr, "%sAT: ii=%d need=%+v %s%s\n", MiscLib.ColorGreen, ii, need, godebug.LF(), MiscLib.ColorReset)
					fmt.Fprintf(os.Stdout, "%sAT: ii=%d need=%+v %s%s\n", MiscLib.ColorGreen, ii, need, godebug.LF(), MiscLib.ColorReset)
					if MatchFn(need, fns) {
						imageList[ii].found = true
					} else {
						imageList[ii].found = false
						missingFn = append(missingFn, need.Name)
					}
				}
				if len(missingFn) > 0 {
					fmt.Fprintf(www, "{\"status\":\"error\",\"file\":%s,\"msg\":\"missing files in image directory.\"}\n", missingFn)
					return
				}

				fmt.Fprintf(os.Stderr, "%sAT: %s%s\n", MiscLib.ColorGreen, godebug.LF(), MiscLib.ColorReset)
				fmt.Fprintf(os.Stdout, "%sAT: %s%s\n", MiscLib.ColorGreen, godebug.LF(), MiscLib.ColorReset)

				// ------------------------------------------------------------------------------------------------
				// 2. Generate hashes of document (Overall, Merkle Leaf etc)
				sTheDoc := SearializeDocumentType(TheDoc)
				DocHash := Keccak256(sTheDoc)

				fmt.Fprintf(os.Stderr, "%sAT: %s%s\n", MiscLib.ColorGreen, godebug.LF(), MiscLib.ColorReset)
				fmt.Fprintf(os.Stdout, "%sAT: %s%s\n", MiscLib.ColorGreen, godebug.LF(), MiscLib.ColorReset)
				// ------------------------------------------------------------------------------------------------
				// 3. Use (implement) PDF generation to build a PDF of the images with the parametric data
				//		1. Use passed parametric data.
				//		2. GenPDF ( parametric, imageList, OutputPath, OutputFile ) ->
				pdfFn, err := hdlr.GeneratePDF(TheDoc, DocHash, documentID, categoryID, imageList, hdlr.InputPath, hdlr.OutputPath)
				if err != nil {
					fmt.Fprintf(www, "{\"status\":\"error\",\"msg\":\"failed to generate PDF.\",\"error\":%q}\n", err)
					return
				}

				// ------------------------------------------------------------------------------------------------
				// xyzzy002  Generate Meta document  - with hashes in it - parametric in it - date time stamp etc.
				//		*1. Pull from $uw class stuff - to generate hashes
				//		2. Put meta documenint in "to-S3" folder for dropbox clone to move.
				mHash := HashImages(imageList)
				pdfHash := HashFile(filepath.Join(hdlr.OutputPath, pdfFn))
				mDoc := metaDocumentType{
					Document:   TheDoc,
					DocumentID: documentID,
					CategoryID: categoryID,
					MerkleHash: mHash,
					PdfHash:    pdfHash,
				}
				documentBytes := SearializeMetaDocument(mDoc)
				documentHash := Keccak256(documentBytes)

				mDoc.OverallHash = documentHash

				// Rename PDF to be hash.pdf
				nFn := fmt.Sprintf("%x", pdfHash) + ".pdf"
				CopyFile(filepath.Join(hdlr.OutputPath, pdfFn), filepath.Join(hdlr.OutputPath, nFn))
				pdfFn = nFn

				fmt.Printf("pdfFn AT: %s pdfFn = [%x] pdfHash = [%s]\n", godebug.LF(), pdfFn, pdfHash)
				fmt.Fprintf(os.Stderr, "*** pdfFn AT: %s pdfFn = [%x] pdfHash = [%s]\n", godebug.LF(), pdfFn, pdfHash)

				// ------------------------------------------------------------------------------------------------
				// 5. Post to Geth/Geth-Q for hash push to BC. -- Backgorund job to run this form Q - ((Redis!))
				//		0. Setup Q in Redis
				//		1. Push data onto Q in Redis - pushr?
				// Mark as "Scheduled to go on-chain with hash"
				// function setDataOnce ( uint256 _app, uint256 _name, bytes32 _data ) public needMinPayment haveTicket payable {
				gethData := fmt.Sprintf(
					"{\"cmd\":\"call\",\"contract\":\"SignedOnce\",\"method\":\"setDataOnce\",\"plist\":[%s,1,\"0x%x\"],\"at\":\"0x%x\"}",
					hdlr.AppID, documentHash, hdlr.SingedOnceAddr)
				hdlr.RedisPushQ(hdlr.RedisQ, gethData)

				fmt.Fprintf(os.Stderr, "%sAT: %s%s\n", MiscLib.ColorGreen, godebug.LF(), MiscLib.ColorReset)
				fmt.Fprintf(os.Stdout, "%sAT: %s%s\n", MiscLib.ColorGreen, godebug.LF(), MiscLib.ColorReset)

				// ------------------------------------------------------------------------------------------------
				// Xyzzy-Later
				// 6. ?Make "get" request to notify that this is ready for S3 -- Or push to S3
				//		-- s3 config info to push to S3 --
				//		0. "DoGet" with random value _ran_ set.
				//		0. Make this a notify-url in config above.
				// Should this just be put on a Q in Redis??? - so it is not a "wait-for-net" thing.
				getData := fmt.Sprintf(
					"{\"cmd\":\"get\",\"url\":%q}",
					fmt.Sprintf("%s/?docHash=%s", hdlr.GetEventURL, documentHash))
				hdlr.RedisPushQ(hdlr.RedisGetQ, getData)

				// 0. Move images to final dir				- IsProd flag == "prod", then
				if hdlr.IsProd == "prod" {
					for _, img := range imageList {
						MoveFile(filepath.Join(hdlr.InputPath, img.Name), filepath.Join(hdlr.ArchivePath, img.Name))
					}
				}

				// xyzzy000 - Update row in d.b. to reflect data.
				pdfHashStr := fmt.Sprintf("%x", pdfHash)
				stmt := `update "dt_document"
				set "eth_hash" = $1
				where "id" = $2
				`
				// _ = stmt // $1 == pdfHash, $2 == documentID
				// rows, err := hdlr.gCfg.Pg_client.Db.Query("select \"salt\", \"password\", \"user_id\" from \"basic_auth\" where \"username\" = $1", key)
				_, err = hdlr.gCfg.Pg_client.Db.Exec(stmt, pdfHashStr, documentID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %s stmt=%s [%s %s] err=%s AT: %s%s\n", MiscLib.ColorRed, stmt, err, pdfHashStr, documentID, godebug.LF(), MiscLib.ColorReset)
					fmt.Fprintf(os.Stdout, "Error: %s stmt=%s [%s %s] err=%s AT: %s%s\n", MiscLib.ColorRed, stmt, err, pdfHashStr, documentID, godebug.LF(), MiscLib.ColorReset)
				}
				fmt.Fprintf(os.Stderr, "%s stmt=%s [%s %s] err=%s AT: %s%s\n", MiscLib.ColorYellow, stmt, err, pdfHashStr, documentID, godebug.SVarI(mDoc), godebug.LF(), MiscLib.ColorReset)
				fmt.Fprintf(os.Stdout, "%s stmt=%s [%s %s] err=%s AT: %s%s\n", MiscLib.ColorYellow, stmt, err, pdfHashStr, documentID, godebug.SVarI(mDoc), godebug.LF(), MiscLib.ColorReset)

				fmt.Fprintf(os.Stderr, "%sAT: %s%s\n", MiscLib.ColorGreen, godebug.LF(), MiscLib.ColorReset)
				fmt.Fprintf(os.Stdout, "%sAT: %s%s\n", MiscLib.ColorGreen, godebug.LF(), MiscLib.ColorReset)

				// ------------------------------------------------------------------------------------------------
				// 8. Return output file URL.
				fmt.Fprintf(www, "{\"status\":\"success\",\"pdfURL\":%q}", hdlr.OutputURL+pdfFn)
				return
			}

			fmt.Fprintf(www, "{\"status\":\"not-implemented-yet\"}")
		}
	}

	hdlr.Next.ServeHTTP(www, req)
}

// -----------------------------------------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------------------------------------

// MoveFile will move a file from one path to a new path - this is a rename.
func MoveFile(from, to string) (err error) {
	return os.Rename(from, to)
}

func (hdlr *Acb1Type) RedisPushQ(key, someData string) {
	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	err = conn.Cmd("RPUSH", key, someData).Err
	if err != nil {
		fmt.Printf("Error on redis  set(%s)=value (%s) error(%s) %s\n", key, someData, err, godebug.LF())
	}
}

func (hdlr *Acb1Type) GeneratePDF(
	TheDoc documentType,
	DocHash []byte,
	documentID, categoryID string,
	imageList []imageTypeList,
	InputPath, OutputPath string,
) (pdfFn string, err error) {
	pdf := NewReport()

	// pdf = Header(pdf, data[0])		// xyzzy - Generate a document with data in it. (Page 1)
	// pdf = Table(pdf, data[1:])
	/*
	   TheDoc := documentType{
	   	Title:     title,
	   	Desc:      desc,
	   	Category:  category,
	   	Tags:      tags,
	   	ImageList: imageListJSON,
	   }
	*/
	data := [][]string{
		{"Title", TheDoc.Title},
		{"Description", TheDoc.Desc},
		{"Category", TheDoc.Category},
		{"Tags", TheDoc.Tags},
	}
	pdf = Table(pdf, data)

	// pdf = InsertImage(pdf, "stats.png", 0)
	// pdf.ImageOptions("stats.png", 225, 10, 25, 25, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")

	for ii, fn := range imageList {
		pdf.AddPage()
		pdf = InsertImage(pdf, filepath.Join(InputPath, fn.Name), ii+1)
	}

	if pdf.Err() {
		err = fmt.Errorf("failed creating PDF report: %s\n", pdf.Error())
		return
	}

	// And finally, we write out our finished record to a file.
	id := hdlr.getID()
	fmt.Printf("id=%s before v3\n", id)
	id = zeroPad(id, 6)
	fmt.Printf("id=%s\n", id)
	pdfFn = id + ".pdf" // Real file name - from ID from Redis -- File will be renamed to hash of self.
	err = SavePDF(pdf, filepath.Join(OutputPath, pdfFn))
	if err != nil {
		err = fmt.Errorf("cannot save PDF: %s|n", err)
	}
	return
}

func zeroPad(sIn string, ln int) (sOut string) {
	sOut = "000000000000000000000" + sIn
	if (len(sOut) - ln) > 0 {
		sOut = sOut[len(sOut)-ln:]
	}
	return
}

func (hdlr *Acb1Type) getID() (id string) {
	key := hdlr.RedisID

	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return "1"
	}

	v, err := conn.Cmd("INCR", key).Int()
	if err != nil || v <= 0 {
		err = conn.Cmd("SET", key, "1").Err
		if err != nil {
			fmt.Printf("Error on redis - failed to create %s\n", key)
			return "1"
		}
	}

	id = fmt.Sprintf("%d", v)
	return

}

// MatchFn returns true if the 'need' file is in the list of avaiable files.
func MatchFn(need imageTypeList, fns []string) bool {
	fmt.Printf("need=%+v fns=%s\n", need, godebug.SVarI(fns))
	for _, fn := range fns {
		if fn == need.Name {
			return true
		}
	}
	return false
}

func HashImages(imageList []imageTypeList) []byte {
	Leaf := make([][]byte, 0, len(imageList))
	for _, img := range imageList {
		Leaf = append(Leaf, HashFile(img.Name))
	}
	return MerkleLeaves(Leaf)
}

func HashFile(fn string) []byte {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		//xyzzy deal with error
		return []byte{}
	}
	return Keccak256(data)
}

/* vim: set noai ts=4 sw=4: */
