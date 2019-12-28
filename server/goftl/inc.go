//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1122
//

package main

import (
	_ "github.com/pschlump/Go-FTL/server/fileserve"
	_ "github.com/pschlump/Go-FTL/server/midlib/AesSrp"
	_ "github.com/pschlump/Go-FTL/server/midlib/BasicAuth"
	_ "github.com/pschlump/Go-FTL/server/midlib/BasicAuthPgSQL"
	_ "github.com/pschlump/Go-FTL/server/midlib/BasicAuthRedis"
	_ "github.com/pschlump/Go-FTL/server/midlib/CORS"
	_ "github.com/pschlump/Go-FTL/server/midlib/Cookie"
	_ "github.com/pschlump/Go-FTL/server/midlib/CorpRegV01"
	_ "github.com/pschlump/Go-FTL/server/midlib/DirectoryBrowse"
	_ "github.com/pschlump/Go-FTL/server/midlib/DocImageProc" // acb
	_ "github.com/pschlump/Go-FTL/server/midlib/DumpRequest"
	_ "github.com/pschlump/Go-FTL/server/midlib/DumpResponse"
	_ "github.com/pschlump/Go-FTL/server/midlib/Echo"
	_ "github.com/pschlump/Go-FTL/server/midlib/ErrorReturn"
	_ "github.com/pschlump/Go-FTL/server/midlib/GenError"
	_ "github.com/pschlump/Go-FTL/server/midlib/GeoIpFilter" // move this out to github.com/pschlump/Go-FTL-GeoIpFilter
	_ "github.com/pschlump/Go-FTL/server/midlib/GoTemplate"
	_ "github.com/pschlump/Go-FTL/server/midlib/Gzip"
	_ "github.com/pschlump/Go-FTL/server/midlib/HTML5Path"
	_ "github.com/pschlump/Go-FTL/server/midlib/HardcodeRewrite"
	_ "github.com/pschlump/Go-FTL/server/midlib/Header"
	_ "github.com/pschlump/Go-FTL/server/midlib/HostToCustomerId"
	_ "github.com/pschlump/Go-FTL/server/midlib/InMemoryCache"
	_ "github.com/pschlump/Go-FTL/server/midlib/JSONToTable"
	_ "github.com/pschlump/Go-FTL/server/midlib/JSONp"
	_ "github.com/pschlump/Go-FTL/server/midlib/Latency"
	_ "github.com/pschlump/Go-FTL/server/midlib/LimitBandwidth"
	_ "github.com/pschlump/Go-FTL/server/midlib/LimitExtensionTo"
	_ "github.com/pschlump/Go-FTL/server/midlib/LimitJson"
	_ "github.com/pschlump/Go-FTL/server/midlib/LimitPathReTo"
	_ "github.com/pschlump/Go-FTL/server/midlib/LimitPathTo"
	_ "github.com/pschlump/Go-FTL/server/midlib/Logging"
	_ "github.com/pschlump/Go-FTL/server/midlib/LoginRequired"
	_ "github.com/pschlump/Go-FTL/server/midlib/Minify"
	_ "github.com/pschlump/Go-FTL/server/midlib/Monitor"
	_ "github.com/pschlump/Go-FTL/server/midlib/OriginToCustomerId"
	_ "github.com/pschlump/Go-FTL/server/midlib/Prefix"
	_ "github.com/pschlump/Go-FTL/server/midlib/Redirect"
	_ "github.com/pschlump/Go-FTL/server/midlib/RedisList"
	_ "github.com/pschlump/Go-FTL/server/midlib/RedisListRaw"
	_ "github.com/pschlump/Go-FTL/server/midlib/RejectDirectory"
	_ "github.com/pschlump/Go-FTL/server/midlib/RejectExtension"
	_ "github.com/pschlump/Go-FTL/server/midlib/RejectHotlink"
	_ "github.com/pschlump/Go-FTL/server/midlib/RejectIpAddress"
	_ "github.com/pschlump/Go-FTL/server/midlib/RejectPath"
	_ "github.com/pschlump/Go-FTL/server/midlib/RejectRePath"
	_ "github.com/pschlump/Go-FTL/server/midlib/Rewrite"
	_ "github.com/pschlump/Go-FTL/server/midlib/RewriteProxy"
	_ "github.com/pschlump/Go-FTL/server/midlib/SaveInFile"
	_ "github.com/pschlump/Go-FTL/server/midlib/SessionRedis"
	_ "github.com/pschlump/Go-FTL/server/midlib/SocketIO"
	_ "github.com/pschlump/Go-FTL/server/midlib/Status"
	_ "github.com/pschlump/Go-FTL/server/midlib/TabServer2"
	_ "github.com/pschlump/Go-FTL/server/midlib/X2fa"
	_ "github.com/pschlump/Go-FTL/server/midlib/ZipIt"
	_ "github.com/pschlump/Go-FTL/server/midlib/acb1"     // acb
	_ "github.com/pschlump/Go-FTL/server/midlib/f-upload" // _ "github.com/pschlump/m6"
	_ "github.com/pschlump/mon-alive/middleware"          // Modules that are not in the midlib deirectory
)
