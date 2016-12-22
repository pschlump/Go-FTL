package AesSrp

//
// Package aessrp implements encrypted authentication and encrypted REST.
// SRP-6a for login authenticaiton, followed by AES 256 bit encrypted RESTful calls.
// A security model with roles is also implemented.
//
// Copyright (C) Philip Schlump, 2013-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 0001
// 你好无聊的世界
//

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/godebug"
)

// ----------------------------------------------------------------------------------------------------------------------------
func AnError(hdlr *AesSrpType, www http.ResponseWriter, req *http.Request, httpCode int, code int, msg string) {
	if hdlr.SendStatusOnError {
		www.WriteHeader(httpCode)
	}
	www.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(www, `{"status":"error","code":"%04d","msg":%q}`, code, msg)
	// fmt.Printf(`{"status":"error","code":"%04d","msg":%q,"LineFile":%q,"URI":%q}`, code, msg, godebug.LF(2), req.RequestURI)
	logrus.Errorf(`{"status":"error","code":"%04d","msg":%q,"LineFile":%q,"URI":%q}`, code, msg, godebug.LF(2), req.RequestURI)
}

func AnErrorRv(hdlr *AesSrpType, www http.ResponseWriter, req *http.Request, httpCode int, code int, msg string) (rv string) {
	if hdlr.SendStatusOnError {
		www.WriteHeader(httpCode)
	}
	rv = fmt.Sprintf(`{"status":"error","code":"%04d","msg":%q}`, code, msg)
	// fmt.Printf(`{"status":"error","code":"%04d","msg":%q,"LineFile":%q,"URI":%q}`, code, msg, godebug.LF(2), req.RequestURI)
	logrus.Errorf(`{"status":"error","code":"%04d","msg":%q,"LineFile":%q,"URI":%q}`, code, msg, godebug.LF(2), req.RequestURI)
	return
}

/* vim: set noai ts=4 sw=4: */
