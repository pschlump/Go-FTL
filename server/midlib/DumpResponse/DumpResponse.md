DumpResponse: Dump Request with Message to Output File - Development Tool
=========================================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Show Response"
	,	"SubSectionGroup": "Debugging"
	,	"SubSectionTitle": "Output responce in the middleware stack"
	,	"SubSectionTooltip": "Dump out the contents of the response at ths point in the middlware stack."
	, 	"MultiSection":2
	}
```

This is a simple middleware that allows the dumping of response to an output file. 

This is one of a set of tools for looking into the middleware stack.
These include:

Middleware | Description
|--- | --- 
`DumpResponse` | Look at output from a request.  It can be placed at different points in the stack. 
`DumpReq` |   Look at what is in the request.  It can be placed at different points in the stack.
`Status` |   Send back to the client what was in the request.  It returns for all matched paths so it is normally used only once for each path.
`Echo` |   Echo a message to standard output when you reach this point in the stack.
`Logging` |   Log what the request/response are at this point in the stack.
`Else` |   A catch all for handling requests that do not have any name resolution.  It will, by default, list all of the available sites on a server by name.

Configuration
-------------

If the `FileName` is not specified, then standard output will be used.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "DumpResponse": { 
					"Paths":   "/api",
					"FileName": "./log/out.log",
					"Msg": "At beginning of request",
					"SaveBodyFlag": true
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
				{ "DumpResponse": { "LineNo":5, 
					"Paths":   "/api",
					"Msg": "At beginning of request"
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

Tested On: Wed Mar  2 12:03:48 MST 2016

