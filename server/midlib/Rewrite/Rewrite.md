Rewrite: Rewrite One Request to Another Location
================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Rewrite Request"
	,	"SubSectionGroup": "Redirect/Rewrite"
	,	"SubSectionTitle": "Rewrite"
	,	"SubSectionTooltip": "Rewrite of request URLs"
	}
```

Rewrite provides the ability to rewrite a URL with a new URL for later processing.

The rewrite uses a regular expression match for the URL.   The replacement allows substitution
of matched items into the resulting URL. 

If RestartAtTop is true, then the set of middleware is restarted from the very top with a re-parse
of parameters and rerunning of each of the middleware that preceded the Rewrite.  If it is false,
the processing continues with the next middleware.

A loop with RestartAtTop is limited to LoopLimit rewrites before it fails.  If RestartAtTop is 
true, then the rewritten URL should not match the regular expression.

Either way query parameters are re-parsed after the rewrite.

Configuration
-------------

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com", "http://localhost:8204/" ],
			"plugins":[
			...
				{ "Rewrite": { 
					"Paths":  [ "/api" ],
					"MatchReplace": [
						{ "Match": "http://(example.com)/(.*)\\?(.*)",
					      "Replace": "http://example.com/rw/process?${2}&name=${1}&${3}"
						}
					],
					"LoopLimit":     50, 
					"RestartAtTop":  true
				} },
			...
	}
``` 


Full Example
------------

``` JSON

	{
		"localhost-13004": { "LineNo":2,
			"listen_to":[ "http://localhost:13004" ],
			"plugins":[
				{ "DumpRequest": { "LineNo":6, "Msg":"Request Before Rewrite", "Paths":"/", "Final":"no" } },
				{ "Rewrite": { "LineNo":6, "Paths":"/",
						"MatchReplace": [
							{ "Match": "http://(localhost:[^/]*)/(.*)\\?(.*)",
							  "Replace": "http://localhost:13004/rw/${2}?rewriten_from=${1}&${3}"
							}
						]
				} },
				{ "DumpRequest": { "LineNo":10, "Msg":"Request After Rewrite", "Paths":"/", "Final":"no" } },
				{ "file_server": { "LineNo":11, "Root":"./www.test1.com", "Paths":"/"  } }
			]
		}
	}

``` 

Example with better regular expressions.  The previous regular expressions require a `?name=value` before 
matching.  This one is a little more realistic.

``` JSON
	{
		"working_test_ReactJS_15": { "LineNo":__LINE__,
			"listen_to":[ "http://localhost:16020", "http://dev2.test1.com:16020" ],
			"plugins":[
				{ "HTML5Path": { "LineNo":__LINE__,
					"Paths":["(/.*\\.html)/.*"]
				} },
				{ "DumpRequest": { "LineNo":__LINE__, "Msg":"Request at top At Top", "Paths":"/api/status", "Final":"yes" } },
				{ "GoTemplate": { "LineNo":__LINE__,
					"Paths":["/api/table/p_document"],
					"TemplateParamName":     "__template__",
					"TemplateName":          "search-docs.tmpl",
					"TemplateLibraryName":   "./tmpl/library.tmpl",
					"TemplateRoot":          "./tmpl"
				} },
				{ "Rewrite": { 
					"Paths":  [ "/api/comments" ],
					"MatchReplace": [
						{ "Match": "http://([^/]*)/api/comments(\\?)?(.*)",
						  "Replace": "http://${1}/api/table/comments${2}${3}"
						}
					],
					"RestartAtTop":  false
				} },
				{ "TabServer2": { "LineNo":__LINE__,
					"Paths":["/api/"],
					"AppName": "www.go-ftl.com",
					"AppRoot": "/Users/corwin/Projects/www.go-ftl.com_doc/_site/data/",
					"StatusMessage":"Version 0.0.4 Sun May 22 19:12:43 MDT 2016"
				} },
				{ "file_server": { "LineNo":__LINE__, "Root":"/Users/corwin/Projects/www.go-ftl.com_doc/_site", "Paths":"/"  } }
			]
		}
	}

```

### Tested

Tested On: Thu, Mar 10, 06:31:05 MST, 2016

Tested On: Sun, Mar 27, 11:48:58 MDT, 2016

### TODO

1. Match on the method also.  GET v.s. POST.   Allow replacement/alteration of method.




