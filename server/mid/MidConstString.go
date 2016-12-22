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
// Return a constant string value.
// ----------------------------------------------------------------------------------------------------------------------------------------------------
// ct := h.Get("Content-Type")
// if rw.StatusCode == http.StatusOK && strings.HasPrefix(ct, "application/json") {
type ConstHandler struct {
	TheBody      string
	AHeader      string
	AHeaderValue string
}

// func NewConstHandler() *ConstHandler {
func NewConstHandler(body, n, v string) http.Handler {
	return &ConstHandler{TheBody: body, AHeader: n, AHeaderValue: v}
}

func (hdlr ConstHandler) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	www.Header().Set(hdlr.AHeader, hdlr.AHeaderValue)
	www.Write([]byte(hdlr.TheBody))
	www.WriteHeader(http.StatusOK)
	return
}
