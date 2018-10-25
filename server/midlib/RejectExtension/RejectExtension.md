RejectExtension: Reject Requests Based on File Extension
========================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Ban Extensions"
	,	"SubSectionGroup": "Limit Requests"
	,	"SubSectionTitle": "Reject a set of extensions"
	,	"SubSectionTooltip": "Prevent to a set of file extensions by banning them"
	, 	"MultiSection":2
	}
```

Limit serving of files to the specified set of extensions.  If the file is not one of the specified

Based on file extension reject requests.  For example, you may want to prevent anybody
accessing any file ending in `*.cfg`.

Configuration
-------------

You can provide a simple list of extensions that when matched will return a HTTP Not Found (404) error.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RejectExtension": { 
					"Paths":   "/",
					"Extensions": [ ".cfg" ]
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
				{ "RejectExtension": { 
					"Paths":   "/",
					"Extensions": [ ".cfg", ".password_db" ]
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

Fri Feb 26 11:03:27 MST 2016
