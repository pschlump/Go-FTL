package UploadFiles

//
// Copyright (C) Philip Schlump, 2013-2016.  All rights reserved.
//

/*

TODO
	+1. (client side) Programmatic call - so can upload image for processing via 2nd Ajax call.
		(or upload, then on completion do a second call?)
		(or upload, server side - do a call) -- Post process --

	2. Other Params -- user/project/product etc. -> mdata -- supply as config on middleware

	// xyzzy8 - check that directory exists
		// xyzzy8 - create if not exists

Demo/Test
	http://localhost:16040/dropzone.demo/test1.html

---------------------------------------------------------------------------------------------
Later
---------------------------------------------------------------------------------------------

	1. Document this
	1. Make part of Go-FTL

	3. If image - do a full read/save to sanitize image.
	3. If image - conversion - to .png for all? -- whatever we need for zxing processing.

	4. re-hash file name for 1st 2 char file sep
		Add up chars in user-file name - and mod 100 or mod 1000, or and 0xFF -- get a 2 char sum of file

*/

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"www.2c-why.com/JsonX"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/Go-FTL/server/urlpath"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*UploadHandlerType)
//		if ok {
//			pCfg.SetNext(next)
//			rv = pCfg
//		} else {
//			err = mid.FtlConfigError
//			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
//		}
//		gCfg.ConnectToRedis()
//		gCfg.ConnectToPostgreSQL()
//		pCfg.gCfg = gCfg
//		return
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} {
//		rv := &UploadHandlerType{}
//		return rv
//	}
//
//	postInitValidation := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
//		fmt.Printf("In postInitValidation, h=%v\n", h)
//		hh, ok := h.(*UploadHandlerType)
//		if !ok {
//			fmt.Printf("Error: Wrong data type passed, Line No:%d\n", hh.LineNo)
//			return mid.ErrInternalError
//		}
//		u, err := filepath.Abs(hh.UploadDirectory)
//		if err != nil {
//			fmt.Printf("Error: converting to absolute path, %s, Line No:%d\n", err, hh.LineNo)
//			return mid.ErrInternalError
//		}
//		// xyzzy8 - check that directory exists
//		// xyzzy8 - create if not exists
//		hh.UploadDirectory = u
//		t := int64(hh.MaxMemory) * 1024
//		hh.maxMemory = t
//		return nil
//	}
//
//	cfg.RegInitItem2("UploadFile", initNext, createEmptyType, postInitValidation, `{
//		}`)
//
//	// Another Example of "FileNameTmpl":        { "type":[ "string" ], "default":"%{upload_path%}/%{user_id%}/%{product_id%}/%{file_name%}" },
//}
//
//// normally identical
//func (hdlr *UploadHandlerType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &UploadHandlerType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("UploadFile", CreateEmpty, `{
		"Paths":        	{ "type":[ "string","filepath"], "isarray":true, "required":true },
        "UploadDirectory":  { "type":[ "string" ], "default":"./upload" },
        "MaxMemory":        { "type":[ "int" ], "default":"10240" },
        "FileNameTmpl":     { "type":[ "string" ], "default":"%{upload_path%}/%{uuid_of_file%}%{file_ext%}" },
		"LineNo":        	{ "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *UploadHandlerType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	gCfg.ConnectToRedis()
	gCfg.ConnectToPostgreSQL()
	hdlr.gCfg = gCfg
	return
}

func (hdlr *UploadHandlerType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	u, err := filepath.Abs(hdlr.UploadDirectory)
	if err != nil {
		fmt.Printf("Error: converting to absolute path, %s, Line No:%d\n", err, hdlr.LineNo)
		return mid.ErrInternalError
	}
	// xyzzy8 - check that directory exists
	// xyzzy8 - create if not exists
	hdlr.UploadDirectory = u
	t := int64(hdlr.MaxMemory) * 1024
	hdlr.maxMemory = t
	return
}

var _ mid.GoFTLMiddleWare = (*UploadHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type UploadHandlerType struct {
	Next            http.Handler                //
	Paths           []string                    //
	UploadDirectory string                      // Directory to place file in
	FileNameTmpl    string                      // Qt template for file name
	MaxMemory       int                         // Maximum mejory in K (*1024)
	LineNo          int                         //
	maxMemory       int64                       // Maximum mejory converted to bytes
	gCfg            *cfg.ServerGlobalConfigType //
}

func NewQRRedirectServer(n http.Handler, p []string, to string) *UploadHandlerType {
	rv := &UploadHandlerType{Next: n, Paths: p, UploadDirectory: to}
	return rv
}

type UploadReturnType struct {
	FileName string
	DBId     string
}

func (hdlr *UploadHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {

	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "UploadHandlerType", hdlr.Paths, pn, req.URL.Path)

			ps := &rw.Ps

			// New New ---------------------------------------------------------------------
			is_multipart := false
			is_raw := false
			var buf []byte
			var ft string

			ct := req.Header.Get("Content-Type")
			if debug_upload {
				fmt.Printf("Matched Upload, content-type=[%s], %s\n", ct, godebug.LF())
			}
			// RFC 2616, section 7.2.1 - empty type SHOULD be treated as application/octet-stream
			if ct == "" {
				ct = "application/octet-stream"
			}
			ct, _, _ = mime.ParseMediaType(ct)
			switch {
			case ct == "application/x-www-form-urlencoded":
				// look for "file" as base64 encoded data.

			case ct == "multipart/form-data":
				is_multipart = true

			case ct == "image/jpeg":
				// look for "file" as binary raw-data.
				buf, _ = ioutil.ReadAll(req.Body)
				fmt.Printf("len=%d\n", len(buf))
				is_raw = true
				ft = ".jpg"

			case ct == "image/png":
				// look for "file" as binary raw-data.
				buf, _ = ioutil.ReadAll(req.Body)
				fmt.Printf("len=%d\n", len(buf))
				is_raw = true
				ft = ".png"

			case ct == "image/gif":
				// look for "file" as binary raw-data.
				buf, _ = ioutil.ReadAll(req.Body)
				fmt.Printf("len=%d\n", len(buf))
				is_raw = true
				ft = ".gif"

			default:
				fmt.Printf("In upload, invalid content-type=%s, %s\n", ct, godebug.LF())
				err := fmt.Errorf("Invalid content type:%s", ct)
				http.Error(www, err.Error(), http.StatusBadRequest)
				return
			}
			if debug_upload {
				fmt.Fprintf(os.Stderr, "In upload ct=%s, %s\n", ct, godebug.LF())
			}

			mdata := make(map[string]string, 20) // The posts that match
			retData := make([]UploadReturnType, 0, 10)

			trx.SetFunc(1)

			id0 := lib.GetUUIDAsString()
			file_name := ps.ByNameDflt("file_name", "")
			file_name = urlpath.Clean(file_name)
			mdata["file_name"] = file_name
			mdata["file_ext"] = filepath.Ext(file_name)
			mdata["ex_id"] = ps.ByNameDflt("ex_id", "") // externa UUID generated by user on client side - saved in d.b.
			mdata["filename"] = ps.ByNameDflt("filename", "")
			mdata["file_type"] = ps.ByNameDflt("file_type", "") // mime type of file - used if not multi-part form.
			if is_raw {
				mdata["file_data"] = string(buf)
				mdata["file_type"] = ct
				mdata["file_ext"] = ft
			} else {
				mdata["file_data"] = ps.ByNameDflt("file_data", "")
			}
			mdata["uuid_of_file"] = id0
			mdata["id"] = id0

			mdata["user_id"] = ps.ByNameDflt("user_id", "")       // xyzzy4 from where? - config
			mdata["product_id"] = ps.ByNameDflt("product_id", "") // xyzzy4 from where? - config

			mdata["upload_path"] = hdlr.UploadDirectory
			pwd, _ := os.Getwd() // xyzzy4 make global config data in hdlr
			mdata["pwd"] = pwd

			if debug_upload {
				fmt.Printf("Data=[%s], AT:%s\n", lib.SVarI(mdata), godebug.LF())
				fmt.Printf("Params=[%s]\n", ps.DumpParamTable())
			}

			if is_multipart {

				godebug.Printf(debug_upload, "AT:%s\n", godebug.LF())

				//parse the multipart form in the request
				err := req.ParseMultipartForm(hdlr.maxMemory)
				if err != nil {
					godebug.Printf(debug_upload, "AT:%s\n", godebug.LF())
					fmt.Printf("Error (14350): Failed to parse multi-part form.\n") // xyzzy - logrus
					trx.AddNote(1, fmt.Sprintf("Error (14350): Failed to parse multi-part form.\n"))
					http.Error(www, err.Error(), http.StatusInternalServerError)
					return
				}

				//get a ref to the parsed multipart form
				m := req.MultipartForm

				godebug.Printf(debug_upload, "m.File=%s AT:%s\n", lib.SVarI(m), godebug.LF())
				//get the *fileheaders
				files := m.File["file"]
				for ii, vv := range files {
					godebug.Printf(debug_upload, "AT:%s\n", godebug.LF())
					file_name = vv.Filename
					file_name = urlpath.Clean(file_name)
					mdata["file_name"] = file_name
					id0 := lib.GetUUIDAsString()
					mdata["uuid_of_file"] = id0
					mdata["id"] = id0
					mdata["file_ext"] = filepath.Ext(file_name)
					ctf := vv.Header.Get("Content-Type") // content type for this file
					mdata["content-type"] = ctf

					if debug_upload {
						fmt.Printf("In multi-part parsing at [%d] id[%s] Data=[%s]\n", ii, id0, lib.SVarI(mdata))
					}

					//for each fileheader, get a handle to the actual file
					file, err := files[ii].Open()
					defer file.Close()
					if err != nil {
						fmt.Printf("Error (14351): Failed to open the file, dstfiel\n")
						trx.AddNote(1, fmt.Sprintf("Error (14351): Failed to open the file, dstfiel\n"))
						http.Error(www, err.Error(), http.StatusInternalServerError)
						return
					}
					trx.AddNote(1, "Successfully opened the file")
					ffile_name := sizlib.Qt(hdlr.FileNameTmpl, mdata)
					dst, err := os.OpenFile(ffile_name, os.O_WRONLY|os.O_CREATE, 0644)
					defer dst.Close()
					if err != nil {
						fmt.Printf("Error (14352): Failed to open the file, dstfiel\n")
						trx.AddNote(1, fmt.Sprintf("Error (14352): Failed to open the file, dstfiel\n"))
						http.Error(www, err.Error(), http.StatusInternalServerError)
						return
					}
					//copy the uploaded file to the destination file
					if _, err := io.Copy(dst, file); err != nil {
						fmt.Printf("Error (14356): Failed to open the file, dstfiel\n")
						trx.AddNote(1, fmt.Sprintf("Error (14356): Failed to open the file, dstfiel\n"))
						http.Error(www, err.Error(), http.StatusInternalServerError)
						return
					}

					h, w := 0, 0

					fsize := lib.GetFileSize(ffile_name)

					// use mime type of this part at this point - use to get h/w if image
					// https://gist.github.com/sergiotapia/7882944 -- read image png/jpeg get size h,w
					if ctf == "image/jpeg" || ctf == "image/gif" || ctf == "image/png" {
						w, h = lib.GetImageDimension(ffile_name)
					} else if ctf == "image/svg+xml" {
						// see note.5
					}

					mdata["height"] = fmt.Sprintf("%d", h)
					mdata["width"] = fmt.Sprintf("%d", w)
					mdata["file_size"] = fmt.Sprintf("%d", fsize)
					mdata["full_file_name"] = ffile_name

					hdlr.InsertIntoDB(mdata, www, req)

					retData = append(retData, UploadReturnType{FileName: mdata["file_name"], DBId: mdata["uuid_of_file"]})

				}

				godebug.Printf(debug_upload, "AT:%s\n", godebug.LF())
				fmt.Fprintf(www, "{ \"status\":\"success\", \"data\":%s }", lib.SVarI(retData))

			} else {

				ffile_name := sizlib.Qt(hdlr.FileNameTmpl, mdata)
				if debug_upload {
					fmt.Printf("Base 64 encoded, Full File Name [%s]\n", ffile_name)
				}
				trx.AddNote(1, fmt.Sprintf("Base 64 encoded, Full File Name [%s]\n", ffile_name))
				dst, err := os.OpenFile(ffile_name, os.O_WRONLY|os.O_CREATE, 0644)
				if err != nil {
					http.Error(www, err.Error(), http.StatusInternalServerError)
					// res.WriteHeader(http.StatusInternalServerError)
					fmt.Printf("Error (14357): Failed to open the file, dstfiel\n")
					trx.AddNote(1, fmt.Sprintf("Error (14357): Failed to open the file, dstfiel\n"))
					return
				}
				trx.AddNote(1, "Successfully opened the file")
				if is_raw {
					trx.AddNote(1, "raw data")
					dst.Write(buf)
				} else {
					trx.AddNote(1, "base64 data")
					data, err := base64.StdEncoding.DecodeString(ps.ByName("file_data"))
					if err != nil {
						fmt.Printf("Error (14358): Failed to decode the file\n")
						trx.AddNote(1, fmt.Sprintf("Error (14358): Failed to decode\n"))
					}
					dst.Write(data)
				}
				trx.AddNote(1, "Data Written")

				h, w := 0, 0

				fsize := lib.GetFileSize(ffile_name)

				ctf := mdata["file_ext"]
				mdata["content-type"] = ctf
				// use mime type of this part at this point - use to get h/w if image
				// https://gist.github.com/sergiotapia/7882944 -- read image png/jpeg get size h,w
				if ctf == ".jpeg" || ctf == ".jpg" || ctf == ".gif" || ctf == ".png" {
					w, h = lib.GetImageDimension(ffile_name)
				} else if ctf == ".svg" {
					// see note.5
				}

				mdata["height"] = fmt.Sprintf("%d", h)
				mdata["width"] = fmt.Sprintf("%d", w)
				mdata["file_size"] = fmt.Sprintf("%d", fsize)
				mdata["full_file_name"] = ffile_name

				hdlr.InsertIntoDB(mdata, www, req)

				retData = append(retData, UploadReturnType{FileName: mdata["file_name"], DBId: mdata["uuid_of_file"]})

				godebug.Printf(debug_upload, "AT:%s\n", godebug.LF())
				fmt.Fprintf(www, "{ \"status\":\"success\", \"data\":%s }", lib.SVarI(retData))
			}

			return
		}
		http.Error(www, "Not Authorized", http.StatusUnauthorized)
		return
	}

	hdlr.Next.ServeHTTP(www, req)
}

func (hdlr *UploadHandlerType) InsertIntoDB(mdata map[string]string, www http.ResponseWriter, req *http.Request) {

	_, err := hdlr.gCfg.Pg_client.Db.Exec(`insert into "p_uploaded_file" (
			  "id"
			, "raw_file_name"
			, "sha1_file_name"
			, "size_in_bytes"
			, "file_type"
			, "height"
			, "width"
		) values( $1, $2, $3, $4, $5, $6, $7 )`,
		mdata["uuid_of_file"],   // "id"
		mdata["file_name"],      // "raw_file_name"
		mdata["full_file_name"], // "sha1_file_name"
		mdata["file_size"],      // "size_in_bytes"
		mdata["content-type"],   // "file_type"
		mdata["height"],         // "height"
		mdata["width"],          // "width"
	)
	if err != nil {
		fmt.Printf("Database error %s, attempting to insert a q_track\n", err)
		logrus.Errorf("Database error %s at: %s", err, godebug.LF())
		www.WriteHeader(http.StatusInternalServerError)
		return
	}
}

var debug_upload = true

/* vim: set noai ts=4 sw=4: */
