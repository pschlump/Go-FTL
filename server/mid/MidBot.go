//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1197
//

package mid

import (
	"net/http"

	"github.com/pschlump/Go-FTL/server/cfg"
)

// ----------------------------------------------------------------------------------------------------------------------------------------------------
// Special handler that just returns 404 for everything.  If you get to this then you have failed to route/find-file etc.
// ----------------------------------------------------------------------------------------------------------------------------------------------------

type BotHandler struct {
}

// func NewBotHandler() http.Handler {
func NewBotHandler() GoFTLMiddleWare {
	return &BotHandler{}
}

func (hdlr *BotHandler) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	return
}

func (hdlr *BotHandler) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

func (hdlr *BotHandler) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	// if you get to the bottom and nobody handled it - then it is a 404 error
	www.WriteHeader(http.StatusNotFound)
	return
}
