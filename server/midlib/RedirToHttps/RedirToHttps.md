RedirectToHttps: Redirect One Request to Another Location
=========================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Redirect to HTTPS"
	,	"SubSectionGroup": "Redirect/Rewrite"
	,	"SubSectionTitle": "Redirect HTTP to HTTPS"
	,	"SubSectionTooltip": "Client side (307) response redirects to HTTPS"
	, 	"MultiSection":2
	}
```

Limit serving of files to the specified set of extensions.  If the file is not one of the specified

Redirect provides the ability to redirect a client to a new location on this or other servers.  If you do
not specify a HTTP status, then 307 temporary redirect will be used.   It is highly recommended that you
avoid 301 redirects.

Configuration
-------------

You can provide a simple list of paths that you want to redirect.  These will get 307 Temporary redirects.
This will take `/api.v2/getData` and redirect it to http://www.example.com/api/getData.
`{{.THE_REST}}` is defined to be any remaining content from the request URI after the Paths match.
 
``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RedirectToHttps": { 
					"Paths":  [ "/api.v2", "/v1.api" ],
					"To":  [ "http://www.example.com/api{{.THE_REST}}", "http://www.example.com/api{{.THE_REST}}" ],
					"Code": [ "MovedTemporary", "MovedPermanent" ],
					"TemplateFileName": "moved.tmpl"
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
				{ "RedirectToHttps": { "LineNo":5, 
					"To":  [ "http://www.zepher.com:3210/api{{.THE_REST}}", "http://www.zepher.com:3210/api{{.THE_REST}}" ],
					"To":  [ "/api", "/api" ]
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

Tested On: Sat Feb 27 18:26:02 MST 2016

1. Tested with simple redirect - Done
1. Test with template
1. Test with invalid configuration
1. Test with invalid template
1. Test with missing template


TODO
----

What happens with post/del etc.

