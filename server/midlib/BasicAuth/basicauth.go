//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1224
//

//
// Package implements HTTP basic auth with the authorization stored in a file.
//
// Basic auth should only be used in conjunction with TLS (https).  If you need to use
// an authentication scheme with http or you want a better authentication scheme
// take a look at the auth_srp.go  middleware.  There are examples of using it with
// jQuery and AngularJS 1.3 (2.0 of AngularJS coming soon).   Also take a look at the
// auth_oauth20 middleware (also coming soon) if you need a 3rd party auth.
//
// Also this is "basic auth" with the ugly browser popup of username/password and no
// real error reporting to the user.  If you want something better switch to the SRP
// solution.
//
// Remember that rainbow tables can crack MD5 hashes in less than 30 seconds 95%
// of the time.  So... this only "basic" auth - with low security.
//
// So what is "basic" auth really good for?  Simple answer.  If you need just a
// touch of secruity - and no more.   Example:  You took a video of your children
// that is and you want to send it to Grandma.  It is too big for her email so
// you need to send a link - so quick copy it up to your server and set basic
// auth on the directory.  Send her the link and the username and password.
// This keeps googlebot and other nozzy folks out of it - but it is not really
// secure.  Then a couple of days later you delete the video.   Works like a
// champ!
//
// There is a command line tool in ./cli-tools/htaccess to maintain the .htaccess
// file.
//

package BasicAuth

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"www.2c-why.com/JsonX"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/uuid"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*BasicAuthType)
//		if ok {
//			pCfg.SetNext(next)
//			rv = pCfg
//		} else {
//			err = mid.FtlConfigError
//			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
//		}
//		return
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} { return &BasicAuthType{} }
//
//	cfg.RegInitItem2("BasicAuth", initNext, createEmptyType, nil, `{
//		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
//		"AuthName":      { "type":[ "string","filepath" ], "default":".htaccess" },
//		"Realm":         { "type":[ "string" ], "required":true },
//		"LineNo":        { "type":[ "int" ], "default":"1" }
//		}`)
//}
//
//// SetNext normally identical
//func (hdlr *BasicAuthType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &BasicAuthType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("BasicAuth", CreateEmpty, `{
		"Paths":         { "type":["string","filepath"], "isarray":true, "required":true },
		"AuthName":      { "type":[ "string","filepath" ], "default":".htaccess" },
		"Realm":         { "type":[ "string" ], "required":true },
		"LineNo":        { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *BasicAuthType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	return
}

func (hdlr *BasicAuthType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*BasicAuthType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

// BasicAuthType kesps the current configuration for the basic-auth middleware
type BasicAuthType struct {
	Next      http.Handler //
	Paths     []string     //
	AuthFile  string       // .htaccess
	Realm     string       //
	LineNo    int
	passwords map[string]string // passwords from file
	mutex     sync.Mutex        //
}

var loaded = false

func newBasicAuthServer(n http.Handler, p []string, authFile string, realm string) *BasicAuthType {
	return &BasicAuthType{Next: n, Paths: p, AuthFile: authFile, Realm: realm}
}

func (hdlr *BasicAuthType) ServeHTTP(www http.ResponseWriter, req *http.Request) {

	hdlr.CheckForLoad()

	_, fn := filepath.Split(filepath.Clean(req.URL.Path))
	if fn == hdlr.AuthFile {
		http.Error(www, "Not Authorized", http.StatusUnauthorized)
		return
	} else if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx1(rw)
			trx.PathMatched(1, "BasicAuth", hdlr.Paths, pn, req.URL.Path)

			auth := req.Header.Get("Authorization")
			if auth == "" {
				www.Header().Set("WWW-Authenticate", "Basic realm=\""+hdlr.Realm+"\"")
				http.Error(www, "Not Authorized", http.StatusUnauthorized)
				return
			}
			userPassword, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
			// fmt.Printf("At:%s userPassword >%s<\n", lib.LF(), userPassword)
			if err == nil {
				parts := strings.SplitN(string(userPassword), ":", 2)
				if len(parts) == 2 {
					un := parts[0]
					pw := parts[1]
					// fmt.Printf("At:%s un:pw >%s:%s<\n", lib.LF(), un, pw)
					// fmt.Printf("At:%s database: %+v\n", lib.LF(), hdlr.passwords)

					hdlr.mutex.Lock()
					pwh, ok := hdlr.passwords[un]
					hdlr.mutex.Unlock()

					// fmt.Printf("At:%s pw %s pwh %v ok %v\n", lib.LF(), pw, pwh, ok)

					if ok && subtle.ConstantTimeCompare([]byte(pw), []byte(pwh)) == 1 {
						// fmt.Printf("At: %s --------------- should be authoraized -- \n", lib.LF())
						goftlmux.AddValueToParams("$is_logged_in$", "y", 'i', goftlmux.FromAuth, &rw.Ps)
						goftlmux.AddValueToParams("$username$", un, 'i', goftlmux.FromAuth, &rw.Ps)
						id0, _ := uuid.NewV4()
						fAuthToken := id0.String()
						goftlmux.AddValueToParams("auth_token", fAuthToken, 'i', goftlmux.FromInject, &rw.Ps)
						hdlr.Next.ServeHTTP(www, req)
						return
					}
				}
			}
		}
		http.Error(www, "Not Authorized", http.StatusUnauthorized)
		return
	}

	hdlr.Next.ServeHTTP(www, req)
}

// CheckForLoad verifies that the data has been loaded -- It also handles locking to that in race conditions only one load occures.
func (hdlr *BasicAuthType) CheckForLoad() {
	hdlr.mutex.Lock()
	x := loaded
	hdlr.mutex.Unlock()
	// fmt.Printf("At:%s \n", lib.LF())
	if !x {
		// fmt.Printf("At:%s \n", lib.LF())
		hdlr.LoadFile()
		hdlr.mutex.Lock()
		loaded = true
		hdlr.mutex.Unlock()
	}
}

// LoadFile loads the data fro the .htaccess file
func (hdlr *BasicAuthType) LoadFile() {
	// fmt.Printf("At:%s AuthFile >%s<\n", lib.LF(), hdlr.AuthFile)
	fh, err := os.Open(hdlr.AuthFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error openng %q for read, %v", hdlr.AuthFile, err)
		os.Exit(1)
	}
	defer fh.Close()
	hdlr.passwords = make(map[string]string)

	//if err = parseHtpasswd(pm, fh); err != nil {
	//	return nil, fmt.Errorf("parsing htpasswd %q: %v", fh.Name(), err)
	//}
	//htpasswords[filename] = pm

	r := csv.NewReader(fh)
	r.Comma = ':'
	r.Comment = '#'

	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("At: %s, records=%+v\n", lib.LF(), records)

	// user:realm:MD5(user:realm:pass)
	for ii, vv := range records {
		if len(vv) == 3 {
			un := vv[0]
			relm := vv[1]
			pw := vv[2]
			// fmt.Printf("At: %s\n", lib.LF())
			if relm == hdlr.Realm {
				// fmt.Printf("At: %s\n", lib.LF())
				hdlr.passwords[un] = pw
			}
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Invalid number of columns on line %d in input\n", ii+1)
		}
	}

}

/* vim: set noai ts=4 sw=4: */
