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

import "net/http"

// ----------------------------------------------------------------------------------------------------------------------------------------------------
// Special handler that just returns 404 for everything.  If you get to this then you have failed to route/find-file etc.
// ----------------------------------------------------------------------------------------------------------------------------------------------------

type BotHandler struct {
}

// func NewBotHandler() *BotHandler {
func NewBotHandler() http.Handler {
	return &BotHandler{}
}

func (hdlr BotHandler) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	// if you get to the bottom and nobody handled it - then it is a 404 error
	www.WriteHeader(http.StatusNotFound)
	return
}
