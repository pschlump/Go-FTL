RejectRePath: Reject Requests Based on a Regular Expression Path Match
======================================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Ban Path Using Regular Expression"
	,	"SubSectionGroup": "Limit Requests"
	,	"SubSectionTitle": "Limit all based on regular expressions"
	,	"SubSectionTooltip": "Prevent access to paths based on a regular expression pattern match"
	, 	"MultiSection":2
	}
```

Limit serving of files to the specified set of extensions.  If the file is not one of the specified

If the path matches - using a regular expression - then reject the requests.

Configuration
-------------

You can provide a simple list of paths to match.  Each match returns a HTTP Not Found (404) error.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RejectRePath": { 
					"Paths":   [ "/.*/config" ]
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
				{ "RejectRePath": { 
					"Paths":   [ "^/.*/config$" ]
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

Fri Feb 26 11:26:41 MST 2016


