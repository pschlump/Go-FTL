GenError: return an error for testing of front end
==================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Show Response"
	,	"SubSectionGroup": "Debugging"
	,	"SubSectionTitle": "Generate An Error to test Front End"
	,	"SubSectionTooltip": "Sometimes you just need an error returned"
	, 	"MultiSection":2
	}
```

This is a simple middleware that return an error for a particular API location.  It is allows for testing of error responses in the front end.


Configuration
-------------

If the `FileName` is not specified, then standard output will be used.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "GenError": { 
					"Paths":   "/api/my-406",
					"StatusCode": 406
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
				{ "GenError": { "LineNo":5, 
					"Paths":   "/api/my-406",
					"StatusCode": 406
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

Tested On: Tue Jul 19 09:55:15 MDT 2016

