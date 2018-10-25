Header: Set/Delete Headers
==========================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Headers"
	,	"SubSectionGroup": "Headers"
	,	"SubSectionTitle": "Set/Delete Headers"
	,	"SubSectionTooltip": "Manipulation of response heades"
	, 	"MultiSection":2
	}
```

Create headers to set or delete cookies.

See a header in the response to a request, or delete a header if it exists.

Configuration
-------------

Create a header.  If you want to set a cookie it is probably better to use `Cookie` middleware instead.
If the header `Name` starts with '-' then delete the header if it exists.

Create a header

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Header": { 
					"Paths":    "/somepath",
					"Name":     "X-Header",
					"Value":    "1234"
				} },
			...
	}
``` 

Delete a header

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Header": { 
					"Paths":    "/somepath",
					"Name":     "-X-bad-header"
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
				{ "Header": { 
					"Paths":    "/somepath",
					"Name":     "X-Test-Header1",
					"Value":    "1234"
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

Tested On: Sat Feb 27 08:02:47 MST 2016

### TODO

1. Use template to allow substitution of header name and values.

