LimitExtensionTo: Limit Requests Based on File Extension
========================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Limit Paths To"
	,	"SubSectionGroup": "Limit Requests"
	,	"SubSectionTitle": "Limit all requests to a set of paths"
	,	"SubSectionTooltip": "Prevent access to non authorized paths"
	, 	"MultiSection":2
	}
```

Limit serving of files to the specified set of extensions.  If the file is not one of the specified
extensions, then reject the request.

Configuration
-------------

You can provide a simple list of extensions that when matched will be served. 
All other extensions return a HTTP Not Found (404) error.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "LimitExtensionTo": { 
					"Paths":   "/",
					"Extensions": [ ".html" ]
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
				{ "LimitExtensionTo": { 
					"Paths":   "/",
					"Extensions": [ ".html", ".json", ".css", ".js" ]
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

Fri Feb 26 10:48:45 MST 2016


