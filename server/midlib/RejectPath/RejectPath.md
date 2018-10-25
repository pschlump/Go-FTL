RejectPath: Reject Requests Based on the Path
=============================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Ban Specified Path"
	,	"SubSectionGroup": "Limit Requests"
	,	"SubSectionTitle": "Ban Paths"
	,	"SubSectionTooltip": "Prevent access to a set of paths"
	, 	"MultiSection":2
	}
```

Limit serving of files to the specified set of extensions.  If the file is not one of the specified

If the path matches, then reject the requests.

Configuration
-------------

You can provide a simple list of paths to match.  Each match returns a HTTP Not Found (404) error.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RejectPath": { 
					"Paths":   "/SrcCode"
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
				{ "RejectPath": { 
					"Paths":   [ "/SrcCode", "/Tests" ]
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

Fri Feb 26 11:05:31 MST 2016


