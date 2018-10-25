Echo: Output a Message When End Point Reached
=============================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Show Message"
	,	"SubSectionGroup": "Debugging"
	,	"SubSectionTitle": "Output"
	,	"SubSectionTooltip": "Output a message to the log"
	, 	"MultiSection":2
	}
```

This is a simple middleware that allows echoing of a message.

This can be used as an end-point to test other items.

This is one of a set of tools for looking into the middleware stack.
These include:

Middleware | Description
|--- | --- 
`DumpResponse` | Look at output from a request.  Can be placed at different points in the stack. 
`DumpReq` |   Look at what is in the request.  Can be placed at different points in the stack.
`Status` |   Send back to the client what was in the request.  It returns for all matched paths. So it is normally used only once for each path.
`Echo` |   Echo a message to standard output when you reach this point in the stack.
`Logging` |   Log what the requests/responses are at this point in the stack.
`Else` |   A catch all for handing requests that do not have any name resolution.  It will, by default, list all of the available sites on a server by name.


Configuration
-------------

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Echo": { 
					"Paths":   "/api/echo",
					"Msg": "Yes I reaced this point"
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
				{ "Echo": { "LineNo":5, 
					"Paths":   "/api/echo",
					"Msg": "Yes I reaced this point"
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

Wed Mar  2 15:11:25 MST 2016

