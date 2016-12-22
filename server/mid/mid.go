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
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/tmplp"
	"github.com/pschlump/Go-FTL/server/tr"
	"github.com/pschlump/Go-FTL/server/urlpath"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/uuid"
)

type GoFTLMiddleWare interface {
	SetNext(next http.Handler)
	ServeHTTP(www http.ResponseWriter, req *http.Request)
}

// var Mutex = &sync.Mutex{}

// ----------------------------------------------------------------------------------------------------------------------------------------------------
// The "top" handler is a special handler that sets up the calls below it and handles errors at the very end.
// ----------------------------------------------------------------------------------------------------------------------------------------------------

type TopHandler struct {
	Format string
	Root   []string
	Next   http.Handler // Required field for all chaining of middleware.
}

func NewTopHandler(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) http.Handler {
	cfg.ServerGlobal.ConnectToRedis()
	return &TopHandler{Next: next, Format: "Error: {{.Error}} From: TopHandler "}
}

func (hdlr TopHandler) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	rw := goftlmux.NewMidBuffer(www, hdlr) // can get rid of memory alloc at this point - just declare and init on stack.

	fmt.Fprintf(os.Stderr, "\n%s---------------------------- Top Handler ----------------------------\n%s\n", MiscLib.ColorCyan, MiscLib.ColorReset)

	fmt.Printf("req.URL.RawQuery [%s] req.URL.Path [%s], %s\n", req.URL.RawQuery, req.URL.Path, godebug.LF())

	var id string
	var trx_id_found = false
	Ck := req.Cookies()
	for _, v := range Ck {
		if v.Name == "X-Go-FTL-Trx-Id" {
			trx_id_found = true
			id = v.Value
			break
		}
	}

	if !trx_id_found {
		id0, _ := uuid.NewV4()
		id = id0.String()
	}

	trx := tr.NewTrx(cfg.ServerGlobal.RedisPool)  // Set to connect to redis // Per request ID that is used by Trx tracing package. and by Socket.IO for directing trafic
	trx.TrxIdSeen(id, req.RequestURI, req.Method) // trx.RequestId = id
	rw.RequestTrxId = id

	rw.G_Trx = trx

	trx.SetTraceDebug(true) // xyzzy - Config Option

	fmt.Fprintf(os.Stderr, "%sAt top of request: %s %s\n", MiscLib.ColorMagenta, req.RequestURI, MiscLib.ColorReset)

	trx.TraceUri(req, url.Values{}) // xyzzy - convert to pass of Parsed Params in Table form -- xyzzyTrx below

	trx.AddNote(1, fmt.Sprintf("Start of Request, IP=%s URI=%s, request-id=%s", req.RemoteAddr, req.RequestURI, id))

	if !trx_id_found {
		expire := time.Now().AddDate(10, 0, 2) // good for 10 year 2 days
		cookie := http.Cookie{Name: "X-Go-FTL-Trx-Id", Value: id, Path: "/", Expires: expire, RawExpires: expire.Format(time.UnixDate), MaxAge: 86400 * 366 * 10, Secure: false, HttpOnly: false}
		http.SetCookie(www, &cookie)
	}

	a := ""
	if req.URL.RawQuery != "" {
		a = "?"
	}
	rw.OriginalURL = req.URL.Path + a + req.URL.RawQuery

	fmt.Printf("\n%s\nOriginalURL: %s, Method:%s, TrxCookie=%s, %s\n%s\n",
		"--------------------------------------------------------------------------------------------------------------------------------",
		rw.OriginalURL, req.Method, id, godebug.LF(),
		"--------------------------------------------------------------------------------------------------------------------------------")

	for {
		//0. clean_up of URLs ->
		//	1. if /foo/bob - is a "dir" then add "/" to end to get /foo/bob/
		//	2. if /abc//def - then /abc/def
		req.URL.Path = urlpath.Clean(req.URL.Path) // -- Any trailing slash is removed, doulbe // are remove

		// XyzzyParams - Make following sections optional via some sort of middleware system.
		goftlmux.ParseQueryParamsReg(www, req, &rw.Ps) //
		goftlmux.MethodParamReg(www, req, &rw.Ps)      // 15ns
		// fmt.Printf("Immeditly before ParseBodyAsParamsReg\n")
		goftlmux.ParseBodyAsParamsReg(www, req, &rw.Ps) // 27ns
		// fmt.Printf("Params Are: %s AT %s\n", rw.Ps.DumpParam(), godebug.LF())
		goftlmux.ParseCookiesAsParamsReg(www, req, &rw.Ps) // 28ns
		// 0. dump of params make it a decent printout - in a little table. -- With text values for Type and From
		// fmt.Printf("Params Are After Cookies: %s AT %s\n", rw.Ps.DumpParamDB(), godebug.LF())
		// xyzzyTrx - set params
		fmt.Printf("\nParams + Cookies for (%s): %s AT %s\n", req.URL.Path, rw.Ps.DumpParamTable(), godebug.LF())

		// Delete reserved items: is_logged_in := ps.ByNameDflt("$is_logged_in$", "")
		goftlmux.RenameReservedItems(www, req, &rw.Ps, cfg.ReservedItems)

		// func (ps *Params) DumpParamNVF() (rv []common.NameValueFrom) {
		dump := rw.Ps.DumpParamNVF()
		trx.SetDataPs(dump)

		rw.RerunRequest = false
		hdlr.Next.ServeHTTP(rw, req)
		if rw.StatusCode == 0 {
			rw.StatusCode = http.StatusOK
		}
		if rw.StatusCode != http.StatusOK { // 200
			// xyzzyTrx -- If rw.RerunRequest is true note that - and exit due to a non-200 code
			break
		}
		if rw.RerunRequest == false { // If no re-run request
			// xyzzyTrx
			break
		}
	}

	fmt.Printf("Mid: StatusCode = %d, req.URL.Path=%s, IsHijacked=%v\n", rw.StatusCode, req.URL.Path, rw.IsHijacked)
	uu := rw.Ps.DumpParamUsed()
	fmt.Printf("\nParams + Cookies for (Shows Used) (%s): %sAT %s\n", req.URL.Path, uu, godebug.LF())
	// trx.AddNote(1, fmt.Sprintf("End of Request, UseParameters=%s", uu, id))
	dump := rw.Ps.DumpParamNVF()
	trx.UpdateDataPs(dump)

	dumpBody := false
	if !rw.IsHijacked {
		if rw.Error != nil && rw.StatusCode == 200 {
			rw.StatusCode = http.StatusInternalServerError
		}
		if rw.StatusCode != 200 && rw.StatusCode != 304 {
			if false {
				logrus.Errorf(tmplp.TemplateProcess(hdlr.Format, rw, req, make(map[string]string)))
			}
			logrus.Errorf("This one -- error code -- %d, %s", rw.StatusCode, godebug.LF())
		}
		if rw.StatusCode == 200 {
			ct := rw.Header().Get("Content-Type")
			// fmt.Printf("ct=%s At: %s\n", ct, godebug.LF())
			if ct == "application/json" { // xyzzy starts with?
				// fmt.Printf("At: %s\n", godebug.LF())
				dumpBody = true
			}
		}
	}

	oldbody := rw.GetBody()
	trx.RvBody = string(oldbody)

	// add in headers for Trx -- espeically content type - xyzzy201412
	trx.SetHeader(rw.Header())

	rw.FinalFlush()

	if dumpBody {
		// fmt.Printf("At: %s\n", godebug.LF())
		body := rw.GetBody()
		fmt.Printf("\nBody: %s\n\n", body)
	}
	if !rw.IsHijacked {
		// fmt.Printf("At: %s\n", godebug.LF())
		fmt.Printf("*** Final Flush ***, %s\n\n=== End %s\n\n", godebug.LF(), strings.Repeat("=== ", 30))
	}

	// ---------------------------------------------------------------------------------------------------------------------------
	ip := req.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	finishTime := time.Now()
	// elapsedTime := finishTime.Since(rw.StartTime)
	elapsedTime := finishTime.Sub(rw.StartTime)

	trx.UriSaveData(ip, rw.StartTime, req.Method, req.RequestURI, req.Proto, rw.StatusCode, rw.Length, elapsedTime, req)

	if !rw.IsHijacked {
		trx.TraceUriRawEnd(req, elapsedTime)
	} else {
		trx.TraceUriRawEndHijacked(req, elapsedTime)
	}
}

func GetTrx(www http.ResponseWriter) (ptr *tr.Trx) {
	if rw, ok := www.(*goftlmux.MidBuffer); ok {
		if rw.G_Trx == nil {
			panic(fmt.Sprintf("Should have has a *tr.Trx - it was NIL, %s\n", godebug.LF(2)))
		}
		if ptr, ok = rw.G_Trx.(*tr.Trx); ok {
			return ptr
		}
	}
	panic(fmt.Sprintf("Should have has a *goftlmux.MidBuffger - got passed a %T, %s\n", www, godebug.LF(2)))
}

func GetTrx1(rw *goftlmux.MidBuffer) (ptr *tr.Trx) {
	if rw.G_Trx == nil {
		panic(fmt.Sprintf("Should have has a *tr.Trx - it was NIL, %s\n", godebug.LF(2)))
	}
	if ptr, ok := rw.G_Trx.(*tr.Trx); ok {
		return ptr
	}
	panic(fmt.Sprintf("Invalid Type - Should have has a *tr.Trx, %s\n", godebug.LF(2)))
}

/* vim: set noai ts=4 sw=4: */
