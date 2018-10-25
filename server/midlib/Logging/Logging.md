Logging: Output a Log Message for Every Request
===============================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Log Messages"
	,	"SubSectionGroup": "Logging"
	,	"SubSectionTitle": "Template format log messages for every requests"
	,	"SubSectionTooltip": "Add or remove loggin information using templates for log messages"
	, 	"MultiSection":2
	}
```

Limit serving of files to the specified set of extensions.  If the file is not one of the specified

Log all requests to the logger.

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

The format can substitute any of these items:

Item | Description
:---: | --- 
`IP` | IP address of remote client
`URI` | URI 
`delta_t` | How long the request has taken
`host` | Host name
`ERROR` | Error message that is returned by lower level middleware
`method` | Request Method
`now` | Current Time
`path` | Request Path
`port` | Port Number
`query` | Query String
`scheme` | HTTP or HTTPS
`start_time` | Start time of request
`status_code` | Numeric status code
`StatusCode` |  Numeric status code
`StatusText` | Numeric status converted to a description

Configuration
-------------

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Logging": { 
					"Paths":   "/api",
					"Format": "IP: {{.IP}} METHOD: {{.method}}"
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
				{ "Logging": { "LineNo":5, 
					"Paths":   "/api",
					"Format": "IP: {{.IP}} METHOD: {{.method}}"
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

Wed, Mar 2, 15:18:12 MST, 2016

