Status: Echo a Request as JSON Data
===================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Show request as JSON"
	,	"SubSectionGroup": "Debugging"
	,	"SubSectionTitle": "Output request in JSON"
	,	"SubSectionTooltip": "Output request in JSON format to aid in debugging middleware stack"
	, 	"MultiSection":2
	}
```

Limit serving of files to the specified set of extensions.  If the file is not one of the specified

This is a simple middleware that allows echoing of a request as JSON data.

This can be used as an end-point to test other items or as an "I am Alive" synthetic request.

Status is also useful for debugging the middleware stack.

This is one of a set of tools for looking into the middleware stack.
These include:

Middleware | Description
|--- | --- 
`DumpResponse` | Look at output from a request.  Can be placed at different points in the stack. 
`DumpReq` |   Look at what is in the request.  Can be placed at different points in the stack.
`Status` |   Send back to the client what was in the request.  It returns for all matched paths so it is normally used only once for each path.
`Echo` |   Echo a message to standard output when you reach this point in the stack.
`Logging` |   Log what the request/response are at this point in the stack.
`Else` |   A catch all for handing requests that do not have any name resolution.  It will, by default, list all of the available sites on a server by name.


Configuration
-------------

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Status": { 
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
				{ "Status": { "LineNo":5, 
					"Paths":   "/api/status"
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

Wed Mar  2 14:44:20 MST 2016

