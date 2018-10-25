JSONp: Implement JSONp requests
===============================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Handle JSONp"
	,	"SubSectionGroup": "Request Processing"
	,	"SubSectionTitle": "Handle JSONp requests"
	,	"SubSectionTooltip": "Transorm get reqeusts into JSONp if they have a callback parameter"
	, 	"MultiSection":2
	}
```

JSONP allows for remotely accessing an API that is cross domain.  This implements
JSONP for an existing API.  For example if "callback=Func9999" is provide on the URL and the JSON
returned is {"josn":"code"}, will be wrapped in:

	Func9999({"json":"code"});

This converts the original JSON to a JavaScript callback function.   This can be used 
from jQuery with a request type of "jsonp".

Configuration
-------------

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "JSONp": { 
					"Paths":   "/api",
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
				{ "JSONp": { "LineNo":5, 
					"Paths":   "/api/status",
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

[//]: # (This may be the most platform independent comment)

Wed Mar  2 10:36:09 MST 2016


