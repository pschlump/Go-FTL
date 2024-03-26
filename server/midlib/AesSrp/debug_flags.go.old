//
// Package aessrp implements encrypted authentication and encrypted REST.
// SRP-6a for login authenticaiton, followed by AES 256 bit encrypted RESTful calls.
//
// Copyright (C) Philip Schlump, 2013-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 0001
// 你好无聊的世界
//

package AesSrp

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/godebug" //
)

var DebugFlags = map[string]bool{
	"DumpUnencryptedRequest":   false,
	"DumpEncryptedReturnValue": false,
	"DumpEncryptedRequest":     false,
}
var DebugMutex sync.Mutex // Lock for map

// use data from global-cfg.json to set local debug flags
func SetDebugFlagsFromGlobal(gCfg *cfg.ServerGlobalConfigType) {
	for _, vv := range gCfg.DebugFlags {
		SetDebugFlag(vv, true)
	}
}

// Get T/F for a siggle named debug flag
func GetDebugFlag(name string) (rv bool) {
	DebugMutex.Lock()
	defer DebugMutex.Unlock()
	rv = DebugFlags[name]
	return
}

// Set a debug flag to t/f - if not exists then just set new flag.
func SetDebugFlag(name string, to bool) {
	DebugMutex.Lock()
	defer DebugMutex.Unlock()
	DebugFlags[name] = to
}

// Resp Handler - to set or toggle debug flags, initial set is to true, after that it is toggle.
func respHandlerSetDebugFlags(www http.ResponseWriter, req *http.Request) {

	rw, hdlr, ok := GetRwHdlrFromWWW(www, req)
	if !ok {
		AnError(hdlr, www, req, 500, 5, fmt.Sprintf("Fatal Error - did not get passed a goftlmux.MidBuffer - AT: %s\n", godebug.LF()))
		return
	}
	ps := rw.Ps

	fmt.Printf("Params Are: %s AT %s\n", ps.DumpParam(), godebug.LF())

	name := ps.ByNameDflt("name", "")

	if name != "" {

		DebugMutex.Lock()
		defer DebugMutex.Unlock()

		if v, ok := DebugFlags[name]; !ok {
			DebugFlags[name] = true
		} else {
			DebugFlags[name] = !v
		}
	}

	fmt.Fprintf(www, `{"status":"success","name":%q}`, name)
}

/* vim: set noai ts=4 sw=4: */
