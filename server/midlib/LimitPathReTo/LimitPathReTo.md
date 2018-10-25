LimitRePathTo: Limit Requests Based on File Extension
=====================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Limit File Extensions"
	,	"SubSectionGroup": "Limit Requests"
	,	"SubSectionTitle": "Limit all requests to a set of file extensions"
	,	"SubSectionTooltip": "Prevent access to non authorized file extensiosn by limiting to a set of valid extensions"
	, 	"MultiSection":2
	}
```

Limit serving of files to the specified set of regular expressions.  If the file is not one of the specified
paths, then reject the request.

Configuration
-------------

You can provide a simple list of extensions that when matched will be served. 
All other paths return  a HTTP Not Found (404) error.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "LimitRePathTo": { 
					"Paths": [ "^/.*\\.html" ]
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
				{ "LimitRePathTo": { 
					"Paths": [ "^/[a-z][a-z]/", ".html$" ]
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

Fri Feb 26 14:52:04 MST 2016


