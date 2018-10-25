LimitPathTo: Limit Requests Based on File Extension
===================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Limit Path To"
	,	"SubSectionGroup": "Limit Requests"
	,	"SubSectionTitle": "Limit all requests to a set specified paths"
	,	"SubSectionTooltip": "Prevent access to non authorized directories"
	, 	"MultiSection":2
	}
```

Limit serving of files to the specified set of extensions.  If the file is not one of the specified

Limit serving of files to the specified set of paths.  If the file is not one of the specified
paths, then reject the request with a HTTP Not Found (404) error.

Configuration
-------------

You can provide a simple list of paths that when matched will be served. 
All other paths return a HTTP Not Found (404) error.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "LimitPathTo": { 
					"Paths":   "/api"
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "LimitPathTo": { 
					"Paths":  [ "/blog", "/api" ]
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Fri Feb 26 10:55:41 MST 2016

<! -- /Users/corwin/go/src/github.com/pschlump/Go-FTL/server/midlib/LimitPathTo/LimitPathTo.md -->

