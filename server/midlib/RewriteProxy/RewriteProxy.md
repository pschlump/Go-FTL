RewriteProxy: Rewrite Reqeust and Proxy It to a Different Server
================================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Rewrite and Proxy"
	,	"SubSectionGroup": "Proxy"
	,	"SubSectionTitle": "Rewrite request and proxy it"
	,	"SubSectionTooltip": "Combined rewrite of request and proxy request to a different server"
	, 	"MultiSection":2
	}
```

A single combed rewrite and proxy that allows access to a different server.  The intention is to
setup a forward proxy that can access other servers that are behind a fire wall.

Configuration
-------------

xyzzy - Need to document how this stuff works!
```
		"MatchRE":       { "type":[ "string" ], "default":"" },			-- apears to be not used
		"ReplaceRE":     { "type":[ "string" ], "default":"" },
		"AddGETParam":     { "type":[ "string" ], "default":"" },
		"Dest":          { "type":[ "string","url" ], "default":"" },
```

You can provide a simple list of extensions that when matched will return a HTTP Not Found (404) error.

`Dest` is the base url of the destination.  By default an URL `http://from.com/a/b?c=d` for a
destination of `http://192.168.0.44:20000/` will get transformed into
`http://192.168.0.44:20000/a/b?c=d`.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RewriteProxy": { 
					"Paths":   "/",
					"MatchRE": ".cfg" 
					"ReplaceRE": ".cfg" 
					"Dest": "http://localhost:8888/" 
				} },
			...
	}
``` 

Another Example:

```
			, { "RewriteProxy": { "LineNo":{{ __line_no__ }}
				, "Paths": [ "/RandomStatus^", "/RandomValue^" ]
				, "Dest": "http://127.0.0.1:10000/" 
			} },
```

An example with a rewrite.  Input is `http://.../Ran/RandomValue`, 
proxy expects `http://.../RandomValue`.

```
			, { "RewriteProxy": { "LineNo":{{ __line_no__ }}
				, "Paths": [ "/Ran/RandomStatus^", "/Ran/RandomValue^" ]
				, "MatchRE": [
						"/Ran/"
					]
				, "ReplaceRE": [
						"/"
					]
				, "Dest": "http://127.0.0.1:10000/" 
			} },
```

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "RewriteProxy": { 
					"Paths":   "/",
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
				} }
			]
		}
	}
``` 


### Tested

Tested On: Thu, Mar 10, 19:42:36 MST, 2016
Tested On: Sun Jul 29 09:43:56 MDT 2018 -- generally working.

1. Issue: can not use multiple paths, `[ "/q/", "/enc/" ]` with a single RewriteProxy.
2. Should be able to specify `"/enc$"` as end matches both "/enc" and "/enc/".
3. What about headers and cookies - are the passed.  Can they be added to - or removed.


