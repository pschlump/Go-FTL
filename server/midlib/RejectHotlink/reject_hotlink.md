RejectHotlink: Reject requests based on invalid referer header
==============================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Prevent Hotlinking"
	,	"SubSectionGroup": "Limit Requests"
	,	"SubSectionTitle": "Stop Hotlinking"
	,	"SubSectionTooltip": "Prevent access to images and other files if a valid referer header is not set."
	, 	"MultiSection":2
	}
```

Limit serving of files to the specified set of extensions.  If the file is not one of the specified

For matching paths, if the file extension for the request matches then only allow the specified set of
`Referer` headers.   This is primarily used to prevent hot linking of images and JavaScript across sites.

Process:

If the path starts with one of the selected paths then:

If the host is in the list of ignored hosts then just pass this request on to the next handler.

If the request has one of the extensions then check the `referer` header. If the header is valid then pass this on.

If the tests fail to pass then either return an error (ReturnError is true) or return an empty clear 1px by 1px GIF image.

Configuration
-------------

You can provide a simple list of paths to match.  

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RejectHotlink": { 
					"Paths":           [ "/js/", "/css/", "/img/" ],
					"AllowedReferer":  [ "www.example.com", "example.com" ],
					"FileExtensions":  [ ".js", ".css", ".gif", ".png", ".ico", ".jpg", ".jpeg" ],
					"AlloweEmpty":     "false",
					"IgnoreHosts":     [ "localhost", "127.0.0.1" ],
					"ReturnError":     "yes"
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
				{ "RejectHotlink": { 
					"Paths":           [ "/js/", "/css/", "/img/" ],
					"AllowedReferer":  [ "www.zepher.com", "zepher.com", "cdn0.zepher.com", "cdn1.zepher.com", "img.zepher.com" ],
					"FileExtensions":  [ ".js", ".css", ".gif", ".png", ".ico", ".jpg", ".jpeg", ".otf", ".eot", ".svg", ".xml", ".ttf", ".woff", ".woff2", ".less", ".sccs", ".csv", ".pdf" ],
					"AlloweEmpty":     "false",
					"IgnoreHosts":     [ "localhost", "127.0.0.1", "[::1]", "::1" ],
					"ReturnError":     "no"
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

Fri Apr 22 12:46:06 MDT 2016 -- Tested only as a part of an entire server.  The automated test is still in the works.



