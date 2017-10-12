//
// Go-FTL - in memory / disk cache.
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1251
//

package InMemoryCache

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"www.2c-why.com/JsonX"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/fileserve"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid" // Path: /Users/corwin/go/src/www.2c-why.com/gosrp
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
)

// --------------------------------------------------------------------------------------------------------------------------
//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*InMemoryCacheType)
//		if ok {
//			pCfg.SetNext(next)
//			rv = pCfg
//		} else {
//			err = mid.FtlConfigError
//			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
//		}
//		pCfg.cache = NewTimedInMemoryCache()
//		PeriodicCleanup(pCfg.cache, pCfg.Duration)
//		if len(pCfg.DiskCache) > 0 {
//			pCfg.diskSize = make([]int64, len(pCfg.DiskCache), len(pCfg.DiskCache))
//			for ii, vv := range pCfg.DiskCache {
//				pCfg.diskSize[ii] = 0 // unlimited size
//				if !lib.Exists(vv) {
//					os.Mkdir(vv, 0755)
//				}
//				pCfg.diskCache = append(pCfg.diskCache, lib.FilepathAbs(vv))
//			}
//			for ii, vv := range pCfg.DiskSize {
//				var u int64
//				u, err = ConvertMGTPToValue(vv)
//				if err != nil {
//					fmt.Fprintf(os.Stderr, "%sWarning: %s - unlimited size used for %s\n%s", MiscLib.ColorRed, err, vv, MiscLib.ColorReset)
//				}
//				pCfg.diskSize[ii] = u
//			}
//			n_ex := 0 // count number of locations
//			disk := make([]string, 0, len(pCfg.DiskCache))
//			SizS := make([]string, 0, len(pCfg.DiskCache))
//			SizI := make([]int64, 0, len(pCfg.DiskCache))
//			for ii, vv := range pCfg.DiskCache {
//				if !lib.Exists(vv) {
//					err = os.MkdirAll(vv, 0700)
//					if err != nil {
//						fmt.Fprintf(os.Stderr, "%sError: %s - unable to create %s\n%s", MiscLib.ColorRed, err, vv, MiscLib.ColorReset)
//						err = mid.FtlConfigError
//					} else {
//						disk = append(disk, vv)
//						SizS = append(SizS, pCfg.DiskSize[ii])
//						SizI = append(SizI, pCfg.diskSize[ii])
//						n_ex++
//					}
//				} else {
//					disk = append(disk, vv)
//					SizS = append(SizS, pCfg.DiskSize[ii])
//					SizI = append(SizI, pCfg.diskSize[ii])
//					n_ex++
//				}
//				os.Remove(vv + "/test.txt")
//				err = ioutil.WriteFile(vv+"/test.txt", []byte("test data\n"), 0600)
//				if err != nil {
//					fmt.Fprintf(os.Stderr, "%sError: %s - unable to create test file in %s\n%s", MiscLib.ColorRed, err, vv, MiscLib.ColorReset)
//					err = mid.FtlConfigError
//					return
//				}
//				os.Remove(vv + "/test.txt")
//			}
//			if n_ex == 0 && len(pCfg.DiskCache) > 0 { // if we have 0 left and we are supposed to cache on disk
//				err = mid.FtlConfigError
//				logrus.Errorf("Unable to initialize InMemoryCacheType - no place to cache files on disk. %s", godebug.LF())
//				return
//			}
//			pCfg.DiskCache = disk
//			pCfg.DiskSize = SizS
//			pCfg.diskSize = SizI
//		}
//		pCfg.PeriodicCleanupDiskFiles()
//		pCfg.gCfg = gCfg
//		return
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} { return &InMemoryCacheType{} }
//
//	cfg.RegInitItem2("CacheData", initNext, createEmptyType, nil, `{
//		}`)
//
//}
//
//// normally identical
//func (hdlr *InMemoryCacheType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &InMemoryCacheType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("InMemoryCache", CreateEmpty, `{
		"Paths":            { "type":["string","filepath"], "isarray":true, "default":"/" },
		"Extensions":       { "type":[ "string" ], "isarray":true },
		"Duration":         { "type":[ "int" ], "default":"60" },
		"IgnoreUrls":       { "type":[ "string" ], "isarray":true },
		"SizeLimit":        { "type":[ "int" ], "default":"500000" },
		"DiskCache":        { "type":[ "string" ], "isarray":true },
		"DiskSize":         { "type":[ "string" ], "isarray":true },
		"RedisPrefix":      { "type":[ "string" ], "default":"cache:" },
		"DiskSizeLimit":    { "type":[ "int" ], "default":"2000000" },
		"DiskCleanupFreq":  { "type":[ "int" ], "defualt":"3600" },
		"LineNo":           { "type":[ "int" ], "default":"1" }
		}`)
}

func (hdlr *InMemoryCacheType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	hdlr.cache = NewTimedInMemoryCache()
	PeriodicCleanup(hdlr.cache, hdlr.Duration)
	if len(hdlr.DiskCache) > 0 {
		hdlr.diskSize = make([]int64, len(hdlr.DiskCache), len(hdlr.DiskCache))
		for ii, vv := range hdlr.DiskCache {
			hdlr.diskSize[ii] = 0 // unlimited size
			if !lib.Exists(vv) {
				os.Mkdir(vv, 0755)
			}
			hdlr.diskCache = append(hdlr.diskCache, lib.FilepathAbs(vv))
		}
		for ii, vv := range hdlr.DiskSize {
			var u int64
			u, err = ConvertMGTPToValue(vv)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%sWarning: %s - unlimited size used for %s\n%s", MiscLib.ColorRed, err, vv, MiscLib.ColorReset)
			}
			hdlr.diskSize[ii] = u
		}
		n_ex := 0 // count number of locations
		disk := make([]string, 0, len(hdlr.DiskCache))
		SizS := make([]string, 0, len(hdlr.DiskCache))
		SizI := make([]int64, 0, len(hdlr.DiskCache))
		for ii, vv := range hdlr.DiskCache {
			if !lib.Exists(vv) {
				err = os.MkdirAll(vv, 0700)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%sError: %s - unable to create %s\n%s", MiscLib.ColorRed, err, vv, MiscLib.ColorReset)
					err = mid.FtlConfigError
				} else {
					disk = append(disk, vv)
					SizS = append(SizS, hdlr.DiskSize[ii])
					SizI = append(SizI, hdlr.diskSize[ii])
					n_ex++
				}
			} else {
				disk = append(disk, vv)
				SizS = append(SizS, hdlr.DiskSize[ii])
				SizI = append(SizI, hdlr.diskSize[ii])
				n_ex++
			}
			os.Remove(vv + "/test.txt")
			err = ioutil.WriteFile(vv+"/test.txt", []byte("test data\n"), 0600)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%sError: %s - unable to create test file in %s\n%s", MiscLib.ColorRed, err, vv, MiscLib.ColorReset)
				err = mid.FtlConfigError
				return
			}
			os.Remove(vv + "/test.txt")
		}
		if n_ex == 0 && len(hdlr.DiskCache) > 0 { // if we have 0 left and we are supposed to cache on disk
			err = mid.FtlConfigError
			logrus.Errorf("Unable to initialize InMemoryCacheType - no place to cache files on disk. %s", godebug.LF())
			return
		}
		hdlr.DiskCache = disk
		hdlr.DiskSize = SizS
		hdlr.diskSize = SizI
	}
	hdlr.PeriodicCleanupDiskFiles()
	hdlr.gCfg = gCfg
	return
}

func (hdlr *InMemoryCacheType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	return
}

var _ mid.GoFTLMiddleWare = (*InMemoryCacheType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type InMemoryCacheType struct {
	Next            http.Handler                //
	Paths           []string                    //
	Extensions      []string                    // Legit extensions to cache
	Duration        int                         // Time in seconds that data is kept in memory
	SizeLimit       int                         // Size limit on how bit to cache
	IgnoreUrls      []string                    // URLs that are to be specifically ignored
	DiskCache       []string                    // Save stuff on disk at locations
	DiskSize        []string                    // Sizes as strings 1G, 1M converted to values
	DiskSizeLimit   int                         // Size limit per-file on disk - how big of individual file to cache
	DiskCleanupFreq int                         // clean up on disk files - default is hourly
	LineNo          int                         //
	RedisPrefix     string                      //
	cache           *TimedInMemoryCache         // Meta data and data cached in memory
	hits            int                         //
	skips           int                         // Indicates that the underlying file changed and memory cached returned a false
	misses          int                         //
	d_hits          int                         // Hits and misses on disk caching
	d_skips         int                         // Indicates that the underlying file changed and cached returned a false
	d_dep_skips     int                         // Indicates that one of the underlying dependencies changed and cached returned a false
	d_misses        int                         // Did not find in cache
	d_read_error    int                         // d_misses incremented, indicates unable to read file
	d_file_missing  int                         // ..., indicates missing file
	d_meta_corrupt  int                         // ..., corrupted meta data - unable to parse JSON
	d_file_corrupt  int                         // ..., checksum failed to match meta data
	diskSize        []int64                     // Sizes in bytes for each cache, 0 unlimited, -1 do not use
	diskCache       []string                    // disk catch converted to absolute paths
	diskToUse       int                         // with of the disks to use for the next save - round robin allocation
	gCfg            *cfg.ServerGlobalConfigType //
}

// Parameterized for testing? or just change the test
func NewInMemoryCacheServer(n http.Handler, p []string, e []string, d int, sl int) *InMemoryCacheType {
	var err error
	x := &InMemoryCacheType{Next: n, Paths: p, Extensions: e, Duration: d, SizeLimit: sl}
	x.cache = NewTimedInMemoryCache()
	PeriodicCleanup(x.cache, d)
	x.PeriodicCleanupDiskFiles()
	x.gCfg = cfg.ServerGlobal
	x.DiskCache = []string{"./cache"} // Save stuff on disk at locations
	x.DiskSize = []string{"1M"}       // Sizes as strings 1G, 1M converted to values
	x.DiskSizeLimit = 10 * 1024 * 1024
	x.diskSize = make([]int64, len(x.DiskCache), len(x.DiskCache))
	for ii, vv := range x.DiskSize {
		var u int64
		u, err = ConvertMGTPToValue(vv)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%sError: %s - unlimited size used\n%s", MiscLib.ColorRed, err, MiscLib.ColorReset)
			break
		}
		x.diskSize[ii] = u
	}
	return x
}

func (hdlr *InMemoryCacheType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if req.Method != "GET" { // only for GET requests
			if cache_db2 {
				fmt.Printf("%s - Ignored, not a GET\n", req.URL.Path)
			}
			hdlr.Next.ServeHTTP(www, req)
		} else if lib.PathsMatchIgnore(hdlr.IgnoreUrls, req.URL.Path) {
			if cache_db2 {
				fmt.Printf("%s - Ignored\n", req.URL.Path)
			}
			hdlr.Next.ServeHTTP(www, req)
		} else if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "TabServer2", hdlr.Paths, pn, req.URL.Path)

			if cache_db2 {
				fmt.Printf("%s - Check if in cache\n", req.URL.Path)
			}
			url, hUrl := lib.GenURL(www, req)

			if hdlr.FoundInMemoryCache(url, hUrl, www, req, rw) {

				if db_SendCacheHeader {
					www.Header().Set("X-In-Memory-Cache", "memory")
				}

				return

			} else if hdlr.FoundInDiskCache(url, hUrl, www, req, rw) {

				// // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
				// Disk cache check at this point
				// // // // // // // // // // // // // // // // // // // // // // // // // // // // // //

				// Found on disk - pop into memory cache -
				dt, _, may := HeadersAllowCaching(www.Header(), hdlr.Duration)
				bod := rw.GetBody()
				if rw.Length <= int64(hdlr.SizeLimit) && may {
					if cache_db2 {
						fmt.Printf("Caching File At: %s, hits=%d misses=%d\n", godebug.LF(), hdlr.hits, hdlr.misses)
						fmt.Printf("Caching Bod: %s\n", bod)
					}
					howLong := hdlr.Duration
					if dt < int64(hdlr.Duration) {
						howLong = int(dt)
					}
					hdlr.cache.SaveContent(url, hUrl, howLong, www.Header(), bod, rw.GetModtime(), rw.ResolvedFn)
				}
				if db_SendCacheHeader {
					www.Header().Set("X-In-Memory-Cache", "on-disk-1")
				}
				return

			} else {

				hdlr.misses++
				if cache_db2 {
					fmt.Printf("At: %s, hits=%d misses=%d\n", godebug.LF(), hdlr.hits, hdlr.misses)
				}
				hdlr.Next.ServeHTTP(rw, req)
				fmt.Printf("\n>>>>>>>>>>>>>>> just after status = %d, hdr %+v, %s\n", rw.StatusCode, www.Header(), godebug.LF())
				if rw.StatusCode == 200 || rw.StatusCode == 0 {
					dt, rdt, may := HeadersAllowCaching(www.Header(), hdlr.Duration)
					bod := rw.GetBody()
					if rw.Length <= int64(hdlr.SizeLimit) && may {
						if cache_db2 {
							fmt.Printf("***** Caching File At: %s, hits=%d misses=%d\n", godebug.LF(), hdlr.hits, hdlr.misses)
							fmt.Printf("***** Caching Bod: %s\n", bod)
						}
						howLong := hdlr.Duration
						if dt < int64(hdlr.Duration) {
							howLong = int(dt)
						}
						hdlr.cache.SaveContent(url, hUrl, howLong, www.Header(), bod, rw.GetModtime(), rw.ResolvedFn)
					}
					if rw.Length <= int64(hdlr.DiskSizeLimit) && may {
						hdlr.SaveInDiskCache(www, req, bod, rdt, rw) // Save File
					}
				} else if rw.StatusCode == 304 {
					// This means that the browser has the same version that is on disk.  However we don't have it in the
					// local in memory cache - do that now.
					dt, _, may := HeadersAllowCaching(www.Header(), hdlr.Duration)
					// xyzzy304 if "must-revalidate" cache - then -- regenerate data - and if not changed send 304
					if may {
						if rw.ResolvedFn != "" {
							bod, err := ioutil.ReadFile(rw.ResolvedFn) // xyzzy - check size limit before read
							if err != nil {
								// xyzzy - log error - IO error on read of cached file
							} else if rw.Length <= int64(hdlr.SizeLimit) {
								if cache_db2 {
									fmt.Printf("***** Caching 304 File At: %s, hits=%d misses=%d\n", godebug.LF(), hdlr.hits, hdlr.misses)
									fmt.Printf("***** Caching 304 Bod: %s\n", bod)
								}
								howLong := hdlr.Duration
								if dt < int64(hdlr.Duration) {
									howLong = int(dt)
								}
								hdlr.cache.SaveContent(url, hUrl, howLong, www.Header(), bod, rw.GetModtime(), rw.ResolvedFn)
								// We don't save it in the disk cache because it is a local file and it is already on disk.
								// This may change when we handle FN.js to FN.min.js and FN.min.map stuff.
							}
						}
						// else if not local - we may want to do a request for the data so that we can cache it.
						// this shall be a worker-q async request so that we get the data.
					}
				}

				return
			}

		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		if cache_db2 {
			fmt.Printf("At: %s\n", godebug.LF())
		}
		hdlr.Next.ServeHTTP(www, req)
	}

}

// -----------------------------------------------------------------------------------------------------------------------------------------------------------------
func (hdlr *InMemoryCacheType) FoundInMemoryCache(url, hUrl string, www http.ResponseWriter, req *http.Request, rw *goftlmux.MidBuffer) (found bool) {

	fmt.Printf("In check memory cache, %s\n", godebug.LF())

	// buf, id, when, modtime, rhdr, ResolvedFn, err, found := hdlr.cache.HaveContent(url, hUrl, hdlr.Duration, req.Header)
	to, err, found := hdlr.cache.HaveContent(url, hUrl, hdlr.Duration, req.Header)

	if err != nil && found {
		fmt.Printf("************** In check memory cache, found=%v err=%v ResolvedFN = %v %s\n", found, err, to.ResolvedFn, godebug.LF())
	}

	if err != nil {
		fmt.Printf("in_memory_cache error: %s\n", err)
		if cache_db2 {
			fmt.Printf("At: %s\n", godebug.LF())
		}
	} else if found {

		fmt.Printf("In Memory MetaData Pulled Back is id=%s modtime=%v %s\n", to.Id, to.ModTime, godebug.LF())
		fmt.Printf("meta_data.ResolvedFn -->>%s<<--, %s\n", to.ResolvedFn, godebug.LF())

		// check if from disk originally at this point
		// xyzzyDependencieCheck - check to see if on-disk served and if so if modified or dependencies failed.
		if to.ResolvedFn != "" {
			fmt.Printf("At: %s\n", godebug.LF())
			fnFound, fnInfo := lib.ExistsGetUDate(to.ResolvedFn)
			if !fnFound {
				fmt.Printf("At: %s\n", godebug.LF())
				// File has been deleted, return false, delete from cache
				hdlr.skips++
				hdlr.cache.DelContent(url, hUrl)
				return false
			}
			fmt.Printf("At: %s\n", godebug.LF())
			// compare modtime
			new_modTime := fnInfo.ModTime()
			fmt.Printf("Fn mod time %v     to.When: %v, %s\n", new_modTime, to.When, godebug.LF())
			if new_modTime.After(to.When) {
				fmt.Printf("At: %s\n", godebug.LF())
				hdlr.skips++
				hdlr.cache.DelContent(url, hUrl)
				return false
			}
			fmt.Printf("At: %s\n", godebug.LF())
		}

		hdlr.hits++
		if cache_db1 {
			fmt.Printf("%s - Found if in cache\n", req.URL.Path)
			rw.Header().Set("X-From-Memory-Cache", "true-"+to.Id) // for debuging purposes.
		}
		// If have ETag && ETag == Id, then 304, else return data
		if fileserve.CheckLastModified(rw, req, to.ModTime) {
			fmt.Printf("%s - Lat Modified, Returining 304, hits=%d misses=%d\n", req.URL.Path, hdlr.hits, hdlr.misses)
			www.WriteHeader(304)
			return true
		}
		// xyzzy - need to look at hash of content and see if modified content? -- Mod time did not match, but content not changed, return 304
		rangeReq, done := fileserve.CheckETag(rw, req, to.ModTime)
		if done {
			if cache_db2 {
				fmt.Printf("At: %s, hits=%d misses=%d\n", godebug.LF(), hdlr.hits, hdlr.misses)
			}
			return true
		}
		_ = rangeReq // need to support range requests
		// copy in headers
		for ii, vv := range to.Hdr {
			if ii != "Etag" {
				for jj := range vv {
					rw.Header().Set(ii, vv[jj])
				}
			}
			// xyzzyCookie - still need to figure out the SetCookie headers - and CookieJar
		}
		rw.Header().Set("Etag", to.Id)
		rw.Write(to.Buf)
		if cache_db2 {
			fmt.Printf("At: %s, hits=%d misses=%d\n", godebug.LF(), hdlr.hits, hdlr.misses)
		}
		return true
	}
	return
}

// -----------------------------------------------------------------------------------------------------------------------------------------------------------------

type MetaData struct {
	Id           string
	Etag         string
	FileName     string
	Hdr          http.Header
	RequestURL   string
	SentURI      string // xyzzy - will need to chagne this
	ModTime      time.Time
	ResolvedFn   string
	DependentFNs []string
	FileSource   string    // local, proxy
	TimeServed   time.Time // for marked files
}

const MaxRetention int64 = 60 * 60 * 24 * 365 * 10 // 10 years

// howLongToSave is in seconds into the futrue - need to convert this for ZADD to a meaningful time.
// Add in the current time since beginning of epoc, then if current time is larger(in-loop) can
// clean up file.
func (hdlr *InMemoryCacheType) SaveInDiskCache(www http.ResponseWriter, req *http.Request, bod []byte, howLongToSave int64, rw *goftlmux.MidBuffer) {
	var err error

	if len(hdlr.DiskCache) == 0 {
		return
	}

	if howLongToSave <= 0 || howLongToSave > MaxRetention {
		howLongToSave = MaxRetention
	}

	now := time.Now()
	epoch := now.Unix()
	saveTill := howLongToSave + epoch

	id := lib.GenSHA(bod)
	FileName, err := hdlr.GetFileName(www, id, bod)
	if err != nil {
		fmt.Printf("Error %s\n", err) // no space available - no save.
		return
	}
	AbsFileName := lib.FilepathAbs(FileName)

	url, hUrl := lib.GenURL(www, req)

	err = ioutil.WriteFile(FileName, bod, 0400)
	if err != nil { // If No Error from save to redis or file system
		fmt.Printf("Error %s - failed to save file\n", err)
		return
	}

	_, fnInfo := lib.ExistsGetUDate(FileName)

	meta_data := MetaData{
		Id:           id,
		Etag:         id,
		FileName:     FileName,
		Hdr:          www.Header(),
		RequestURL:   url,
		SentURI:      req.RequestURI,   // xyzzy - will chagne with 3rd party stuff
		ModTime:      fnInfo.ModTime(), // ModTime:    time.Now(), -- maybee??
		ResolvedFn:   rw.ResolvedFn,    // Actual original file name on disk
		DependentFNs: rw.DependentFNs,  // Set of files this was built from
	}

	s_meta_data := lib.SVarI(meta_data)

	fmt.Printf("MetaData Saved is -->>%s<<--, %s\n", s_meta_data, godebug.LF())

	key := hdlr.RedisPrefix + hUrl

	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return
	}

	err = conn.Cmd("SET", key, s_meta_data).Err // - what to set in redis-
	if err == nil {                             // If No Error from save to redis or file system
		err = conn.Cmd("EXPIRE", key, howLongToSave-1).Err // - what to set in redis-
		if err != nil {                                    // If No Error from save to redis or file system
			fmt.Printf("Error %s - redis error\n", err)
		}
		err = conn.Cmd("ZADD", hdlr.RedisPrefix+"fct", fmt.Sprintf("%d", saveTill), AbsFileName).Err // - what to set in redis-
		if err != nil {                                                                              // If No Error from save to redis or file system
			fmt.Printf("Error %s - redis error\n", err)
		}
	}

}

var ErrNoSpaceAvailable = errors.New("No space available on specified volumes")

func (hdlr *InMemoryCacheType) GetFileName(www http.ResponseWriter, id string, bod []byte) (fn string, err error) {
	ctypes := lib.GetCTypes(www, bod)
	//ctypes, haveType := www.Header()["Content-Type"]
	//if !haveType {
	//	ctype := http.DetectContentType(bod)
	//	www.Header().Set("Content-Type", ctype)
	//	ctypes = append(ctypes, ctype)
	//}
	for i := 0; i < len(hdlr.DiskCache); i++ {
		n := hdlr.diskToUse % len(hdlr.DiskCache)
		hdlr.diskToUse++
		if hdlr.SpaceAvailable(hdlr.DiskCache[n], n, www) {
			dir := hdlr.DiskCache[n]
			g := dir + "/" + id[0:3]
			if !lib.Exists(g) {
				os.Mkdir(g, 0755)
			}
			fn = g + "/" + id + "." + lib.GetExtenstionBasedOnMimeType(ctypes[0])
			return
		}
	}
	err = ErrNoSpaceAvailable
	return
}

func (hdlr *InMemoryCacheType) SpaceAvailable(path string, nth int, www http.ResponseWriter) bool {
	// TODO: Space - implement space check // path is where to check, // check space on disk
	// nth - is internal check
	if rw, ok := www.(*goftlmux.MidBuffer); ok {
		if nth < len(hdlr.diskSize) {
			need := int64(rw.Length)
			if hdlr.diskSize[nth] == 0 || need < hdlr.diskSize[nth] {
				// check space on disk
				return true
			}
		}
	}
	// check space on disk
	return false
}

func (hdlr *InMemoryCacheType) FoundInDiskCache(url, hUrl string, www http.ResponseWriter, req *http.Request, rw *goftlmux.MidBuffer) bool {

	if len(hdlr.DiskCache) == 0 {
		return false
	}

	// url, hUrl := lib.GenURL(www, req)

	//		1. See if in redis based on URL, if not		-- redis GET
	//			return false

	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return false
	}

	key := hdlr.RedisPrefix + hUrl
	meta, err := conn.Cmd("GET", key).Str()
	if err != nil {
		hdlr.d_misses++
		return false
	}
	var meta_data MetaData
	err = json.Unmarshal([]byte(meta), &meta_data)
	if err != nil {
		hdlr.d_misses++
		hdlr.d_meta_corrupt++
		conn.Cmd("DEL", key)
		conn.Cmd("ZADD", hdlr.RedisPrefix+"fct", 0, url) // - what to set in redis-
		return false
	}

	fmt.Printf("MetaData Pulled Back is -->>%+v<<--, %s\n", meta_data, godebug.LF())
	fmt.Printf("meta_data.ResolvedFn -->>%s<<--, %s\n", meta_data.ResolvedFn, godebug.LF())

	// xyzzyDependencieCheck - check to see if on-disk served and if so if modified or dependencies failed.
	if meta_data.ResolvedFn != "" {
		fmt.Printf("At: %s\n", godebug.LF())
		fnFound, fnInfo := lib.ExistsGetUDate(meta_data.ResolvedFn)
		if !fnFound {
			// 2. xyzzyMultiServer - This is the spot to go looking for it across server farm.
			fmt.Printf("At: %s\n", godebug.LF())
			if fnFound, fnInfo = meta_data.FoundOnTheFarm(); !fnFound {
				fmt.Printf("At: %s\n", godebug.LF())
				// File has been deleted, return false, delete from cache
				hdlr.d_skips++
				conn.Cmd("DEL", key)
				conn.Cmd("ZADD", hdlr.RedisPrefix+"fct", 0, url) // - what to set in redis-
				return false
			}
		}
		fmt.Printf("At: %s\n", godebug.LF())
		// compare modtime
		new_modTime := fnInfo.ModTime()
		if new_modTime.After(meta_data.ModTime) {
			fmt.Printf("At: %s\n", godebug.LF())
			hdlr.d_skips++
			conn.Cmd("DEL", key)
			conn.Cmd("ZADD", hdlr.RedisPrefix+"fct", 0, url) // - what to set in redis-
			return false
		}
		fmt.Printf("At: %s\n", godebug.LF())
	}

	fmt.Printf("At: %s\n", godebug.LF())
	rw.ResolvedFn = meta_data.ResolvedFn

	etag := www.Header().Get("Etag")
	rangeReq := req.Header.Get("Range")
	if meta_data.Id == etag && rangeReq == "" {
		fmt.Printf("%s - DISK Hash based Etag, Returining 304, hits=%d misses=%d\n", req.URL.Path, hdlr.d_hits, hdlr.d_misses)
		www.WriteHeader(304)
		hdlr.d_misses++
		return true
	}

	//		2. Use meta data to find file, if file not found
	//			return false
	fn := meta_data.FileName
	var buf []byte
	if !lib.Exists(fn) {
		hdlr.d_misses++
		hdlr.d_file_missing++
		//rr.RedisDo("DEL", key)
		//rr.RedisDo("ZADD", hdlr.RedisPrefix+"fct", 0, url) // - what to set in redis-
		return false
	} else {
		buf, err = ioutil.ReadFile(fn)
		if err != nil {
			hdlr.d_misses++
			hdlr.d_read_error++
			conn.Cmd("DEL", key)
			conn.Cmd("ZADD", hdlr.RedisPrefix+"fct", 0, url) // - what to set in redis-
			return false
		}
		// validate file
		old_id := meta_data.Id
		new_id := lib.GenSHA(buf)
		if new_id != old_id {
			hdlr.d_misses++
			hdlr.d_file_corrupt++
			conn.Cmd("DEL", key)
			conn.Cmd("ZADD", hdlr.RedisPrefix+"fct", 0, url) // - what to set in redis-
			return false                                     // file exists but it is wrong - bad data - error in read - truncated etc.
		}
	}

	//		3. Combine meta + data
	//			return true
	hdlr.d_hits++
	if cache_db1 {
		fmt.Printf("%s - Found in disk cache\n", req.URL.Path)
		www.Header().Set("X-From-Disk-Cache", "true-"+meta_data.Id) // for debuging purposes.
	}

	// If have ETag && ETag == Id, then 304, else return data
	//if fileserve.CheckLastModified(rw, req, modtime) {
	//	fmt.Printf("%s - Lat Modified, Returining 304, hits=%d misses=%d\n", req.URL.Path, hdlr.hits, hdlr.misses)
	//	www.WriteHeader(304)
	//	return
	//}

	// xyzzy - need to look at hash of content and see if modified content? -- Mod time did not match, but content not changed, return 304
	//rangeReq, done := fileserve.CheckETag(rw, req, modtime)
	//if done {
	//	if cache_db2 {
	//		fmt.Printf("At: %s, hits=%d misses=%d\n", godebug.LF(), hdlr.hits, hdlr.misses)
	//	}
	//	return
	//}
	//_ = rangeReq // need to support range requests

	// copy in headers
	for ii, vv := range meta_data.Hdr {
		if ii != "Etag" {
			for jj := range vv {
				www.Header().Set(ii, vv[jj])
			}
		}
		// xyzzy - still need to figure out the SetCookie headers - and CookieJar
	}
	www.Header().Set("Etag", meta_data.Id)
	www.Write(buf)
	if cache_db2 {
		fmt.Printf("At: %s, hits=%d misses=%d\n", godebug.LF(), hdlr.d_hits, hdlr.d_misses)
	}
	return true

}

// if fnFound, fnInfo = FoundOnTheFarm ( meta_data ); !fnFound {
func (meta_data *MetaData) FoundOnTheFarm() (foundIt bool, fnInfo os.FileInfo) {
	foundIt = false
	// if in multi-farm mode -
	// 		do a GET from most recent - if fetch rate high - proxy style - if not found then
	// 		do a GET from owner - if not found then return false
	// 		if found then save - take a look at rate - if fetch rate is low, then schedule file for delete.
	return
}

//	DiskCleanupFreq int					// clean up on disk files - default is hourly
func (hdlr *InMemoryCacheType) PeriodicCleanupDiskFiles() {
	d := hdlr.DiskCleanupFreq
	if d == 0 {
		return // turned off
	}
	ticker := time.NewTicker(time.Duration(d) * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				if cache_db3 {
					fmt.Printf("\nDisk Cleanup start -------------------------------------------------------- \n\n")
				}
				// do something
				list, et := hdlr.GetExpiredFiles()
				fmt.Printf("list = %s, et = %s\n", list, et)
				for _, fn := range list {
					if hdlr.InCachePath(fn) {
						err := os.Remove(fn)
						if err != nil {
							fmt.Printf("Error: uable to remove cached file %s, error:%s\n", fn, err)
						}
					}
				}
				if cache_db3 {
					fmt.Printf("Disk Cleanup Finished\n\n")
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

//------------------------------------------------------------------------------------------------
// comapre abs(cahce) path with fn prefix and if match then return true
func (hdlr *InMemoryCacheType) InCachePath(fn string) bool {
	// hdlr.diskCache = append ( hdlr.diskCache , filepath.Abs(vv) )
	for _, tPth := range hdlr.diskCache {
		if strings.HasPrefix(fn, tPth) {
			return true
		}
	}
	return false
}

//------------------------------------------------------------------------------------------------
// Fetch the list of expired files and delete this from the list.
func (hdlr *InMemoryCacheType) GetExpiredFiles() (kks []string, curTimeUnix string) {
	var err error
	theKey := hdlr.RedisPrefix + "fct"
	now := time.Now()
	epoch := now.Unix()
	curTimeUnix = fmt.Sprintf("%d", epoch)

	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Info(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return []string{}, curTimeUnix
	}

	// ZRANGEBYSCORE cache:fct -inf 1765388241
	kks, err = conn.Cmd("ZRANGEBYSCORE", theKey, "-inf", curTimeUnix).List()
	if err != nil {
		return []string{}, curTimeUnix
	}

	// ZREMRANGEBYSCORE myzset -inf (2)
	conn.Cmd("ZREMRANGEBYSCORE", theKey, "-inf", curTimeUnix)

	return
}

// -----------------------------------------------------------------------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------------------------------------------------------------------
type TimedInMemoryData struct {
	Buf        []byte      // body data
	Id         string      // Self SHA256 hash
	ModTime    time.Time   // When modified
	When       time.Time   // When fetched last
	Hdr        http.Header // Headers to return
	Hits       int         // how many times fetched
	ResolvedFn string      // If a file on disk, then this is the single file name
}

type TimedInMemoryCache struct {
	Cache   map[string]*TimedInMemoryData
	MapLock sync.Mutex // Lock for accesing maps in this.
}

func NewTimedInMemoryCache() (rv *TimedInMemoryCache) {
	return &TimedInMemoryCache{
		Cache: make(map[string]*TimedInMemoryData),
	}
}

func (ti *TimedInMemoryCache) HaveContent(uri, hUrl string, d int, hdr http.Header) (to TimedInMemoryData, err error, found bool) {
	ti.MapLock.Lock()
	defer ti.MapLock.Unlock()
	found = false

	data, found := ti.Cache[hUrl]
	if !found {
		return
	} else {
		to = *data
		data.Hits++
		data.When = time.Now()
		ti.Cache[hUrl] = data
		found = true
	}
	return
}

func (ti *TimedInMemoryCache) SaveContent(uri, hUrl string, d int, hdr http.Header, bod []byte, modtime time.Time, ResolvedFn string) {
	ti.MapLock.Lock()
	defer ti.MapLock.Unlock()

	fmt.Printf("********************* saving %s %s\n", uri, hUrl)

	id := lib.GenSHA(bod)
	ti.Cache[hUrl] = &TimedInMemoryData{Buf: bod, Id: id, When: time.Now(), ModTime: modtime, Hdr: hdr, ResolvedFn: ResolvedFn}
}

func (ti *TimedInMemoryCache) DelContent(url, hUrl string) {
	ti.MapLock.Lock()
	defer ti.MapLock.Unlock()

	delete(ti.Cache, hUrl)
}

// Duration 'd' is in seconds - indicates how often the cleanup process will happen.  d==0, no cleanup.
func PeriodicCleanup(ti *TimedInMemoryCache, d int) {
	if d == 0 {
		return
	}
	ticker := time.NewTicker(time.Duration(d) * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				if cache_db3 {
					fmt.Printf("\nCleanup start -------------------------------------------------------- \n\n")
				}
				ti.MapLock.Lock()
				for uri, data := range ti.Cache {
					if data.When.Add(time.Duration(d) * time.Second).Before(time.Now()) {
						if cache_db3 {
							fmt.Printf("Discarding/Delete [%s] Hits %d\n", uri, data.Hits)
						}
						delete(ti.Cache, uri)
					} else {
						if cache_db3 {
							fmt.Printf("Keeping [%s] Hits %d\n", uri, data.Hits)
						}
					}
				}
				ti.MapLock.Unlock()
				if cache_db3 {
					fmt.Printf("Cleanup Finished\n\n")
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

/*
	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Warn(fmt.Sprintf(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF()))
		return false
	}

	v, err := conn.Cmd("GET", key).Str()

	hdlr.gCfg.RedisPool.Put(conn)
*/

const cache_db1 = true
const cache_db2 = true
const cache_db3 = true
const db_SendCacheHeader = true

/* vim: set noai ts=4 sw=4: */
