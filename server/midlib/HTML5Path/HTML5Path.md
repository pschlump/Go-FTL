HTML5Path: Redirect 404 Errors to Index.html for AngularJS Router
=================================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "HTML5 Paths"
	,	"SubSectionGroup": "Redirect/Rewrite"
	,	"SubSectionTitle": "Redirect requests to a .html file"
	,	"SubSectionTooltip": "Angular 1.x, 2.x and other HTML5 single pages applications uses multiple URLs that all need to direct to a single .html page."
	, 	"MultiSection":2
	}
```

AngularJS 2.0 and AngularJS 1.x have an interesting default routing.  They change the current path.
For example, `http://myapp.com/app.html` becomes `http://myapp.com/app.html/dashboard`  and then `http://myapp.com/app.html/productList`.
When a person bookmarks or refreshes one of these URLs the server has no clue what a "/app.html/dashboard" is and returns
a 404 error.  

One possible solution is to map all 404 errors to `app.html`.   This is *icky* because it breaks all 404 handling.  You end up
returning `app.html` for `/image/nonexistent.jpg` and the browser is not happy with you at all (and it shouldn't be!)

The solution in this middleware is more nuanced. If the lower levels return a 404 and this is a `GET` request
then if one of the Paths regular expression matches use a regular expression to replace the selected portion of
the URL and retry that.

What should happen is that all of these should be mapped to the single page application.  By default this
is `&lt;some-name&gt;.html`.  You can change this with the `ReplaceWith` option.

After the file server returns a 404 you can limit the set of paths with the `LimitTo` set of options.
If `LimitTo` is not specified, then all 404 errors will be returned as index.html.

At best this should be considered *experimental*. I am still working on what to do with `/` maping to `/index.html`.
Anticipate changes in this middleware in the near future.

Configuration
-------------

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "HTML5Path": { "LineNo":__LINE__,
					"Paths":["(/.*\\.html)/.*"]
				} },
			...
	}
```

Full Example
------------

This example is the server confgiuration that I used for my Angular 2.0 rc 1 documentation development (The page 
you are currently reading)  This also includes the configuration for TabServer2.

``` JSON
	{
		"working_test_AngularJS_20": { "LineNo":__LINE__,
			"listen_to":[ "http://localhost:16020", "http://dev2.test1.com:16020" ],
			"plugins":[
				{ "HTML5Path": { "LineNo":__LINE__,
					"Paths":["(/.*\\.html)/.*"]
				} },
				{ "TabServer2": { "LineNo":__LINE__,
					"Paths":["/api/"],
					"AppName": "docs.go-ftl.com",
					"AppRoot": "/Users/corwin/Projects/docs-Go-FTL/data/",
					"StatusMessage":"Version 0.0.4 Sun May 22 19:12:43 MDT 2016"
				} },
				{ "file_server": { "LineNo":__LINE__, "Root":"/Users/corwin/Projects/docs-Go-FTL", "Paths":"/"  } }
			]
		}
	}
```

### Tested

Tested On: Wed Jun  1 13:05:11 MDT 2016 (Note - Tested by using it in an AngularJS 2.0 application)  An automated test is in-the-works.

### TODO

1. A better name for this middleware.   As soon as I can figure out what to call it I will change this.




