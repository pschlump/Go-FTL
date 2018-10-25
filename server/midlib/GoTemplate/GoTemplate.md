GoTemplate: Template using Go's Buit in Templates
=================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Go Templates"
	,	"SubSectionGroup": "Output Formatting"
	,	"SubSectionTitle": "Go Template"
	,	"SubSectionTooltip": "Use Go Templates to format data"
	, 	"MultiSection":2
	}
```

GoTemplate implements a middleware that combines templates with underlying data.

Basic usage of Go templates is also supported.  You can build a page with a header
template, a footer template and a body template.

A more powerful way to use this is to combine data with templates to render a
final text.  Examples of each of these will show how this can be used.

Configuration
-------------

Specify a path for templates and the location of the template library.

Parameter | Description
|--- | --- 
`TemplateParamName` | The name on the URL of the template that is to be rendered with this data.
`TemplateName` | The name of the template if __template__ has an empty value.
`TemplateLibraryName` | An array of file names or a single file that has the set of templates for rendering the page.
`TemplateRoot` | The path to search for the template libraries.  If this is not specified, then it will be searched for in `Root`.
`Root` | The root for the set of web pages.  It should be the same root as the `file_server` `Root`.


``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "GoTemplate": { 
					"Paths": [ "/data" ],
					"TemplateParamName":     "__template__",
					"TemplateName":          "render_body",
					"TemplateLibraryName":   "common_library.tmpl",
					"TemplateRoot":          "./tmpl",
					"Root":                  ""
				} },
			...
	}
``` 


Example 1: Simple Page Composition
----------------------------------

You have a website with a common header, footer and each body is different.

The Go-FTL configuration file is:

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "GoTemplate": { 
					"Paths": [ "/twww" ],
					"TemplateParamName":     "__template__",
					"TemplateName":          "body",
					"TemplateLibraryName":   [ "common_library.tmpl", "{{.__tempalte__}}.tmpl" ]
					"TemplateRoot":          "./tmpl",
					"Root":                  ""
				} },
				{ "Echo": { 
					"Paths": [ "/twww" ],
					"Msg": ""
				} }
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 

In ./tmpl/common_library.tmpl you have

``` HTML
	{{define "content_type"}}text/html; charset=utf-8{{end}}
	{{define "header"}}<!DOCTYPE html>
	<html lang="en">
	<body>
		<div> header </div>
	{{end}}
	{{define "footer"}}
		<div> footer </div>
	</body>
	</html>
	{{end}}
	{{define "body"}}
		{{header .}}
		<div> this is my default body - it is a good body 1 </div>
		<div> this is my default body - it is a good body 2 </div>
		{{footer .}}
	{{end}}
``` 

In ./tmpl/main.tmpl you have

``` HTML
	{{define "main"}}
		{{header .}}
		<div> this is my main body </div>
		{{footer .}}
	{{end}}
``` 

A request for `http://www.zepher.com:3210/twww?__template__=main` will do the following:

1. GoTemplate sees the url `/twww` and calls the next function down the stack.
2. Echo sees the url `/twww` and matches - It returns the Msg string as the results.  An empty string.
3. GoTemplate uses the returning data from Echo.  This is actually an empty string.   It reads in the template files in order, common_library.tmpl then substituting the parameter, main.tmpl.  It then calls the template "main" witch calls the "header" and "footer" templates to render.

The returned data is transformed into (with a couple of extra blank lines suppressed)

``` HTML
	<!DOCTYPE html>
	<html lang="en">
	<body>
		<div> header </div>
		<div> this is my main body </div>
		<div> footer </div>
	</body>
	</html>
``` 

If __template__ had not been specified, then the template "body" would have been called.  It acts as a default body in this case.

The `content_type` template is used to generate the content type for the page.  You can use this to generate XML or SVG, or to transform data
and return it in other mime types.

In this example you may want to use Rewrite first to generate the ugly URL: `http://www.zepher.com:3210/twww?__template__=main`

The documentation for this tool is generated in this fashion.  It is actually a little bit more complicated.  The files are in Markdown (.md) and processed from .md to
.html, then written into templates, .tmpl and combined with headers and footers.

Example 2: Page Composition with Data
-------------------------------------

Combining data with templates is incredibly powerful.  For this example we will combine some static data in a .json file with templates to render it.
You can also use this with the RedisListRaw to pull data out of Redis and combine it with templates to render it.   This turns the templates into a
simple report writer tool.  Complete access to a relational database is also available with the `TabServer2` middleware.  This has been tested with
PostgreSQL, MySQL, Oracle, and Microsoft MS-SQL.

The Go-FTL configuration file is:

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "GoTemplate": { 
					"Paths": [ "/data/" ],
					"TemplateParamName":     "__template__",
					"TemplateName":          "body",
					"TemplateLibraryName":   [ "data_library.tmpl" ]
					"TemplateRoot":          "./tmpl",
					"Root":                  ""
				} },
				{ "JSONToTable": { "LineNo":5, 
					"Paths":   "/data/",
					"ConvertRowTo1LongTable": true
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 

In ./tmpl/data_library.tmpl you have

``` HTML
	{{define "content_type"}}text/html; charset=utf-8{{end}}
	{{define "header"}}<!DOCTYPE html>
	<html lang="en">
	<body>
		<div> header </div>
		<ul>
	{{end}}
	{{define "data_render_body"}}
		{{range $ii, $ee := .data}}
			<li><a href="/{{$ee.abc}}.html"> {{$ee.abc}} id:{{$ee.myId}} </a></li>
		{{end}}
	{{end}}
	{{define "footer"}}
		</ul>
		<div> footer </div>
	</body>
	</html>
	{{end}}
```

With data served by the file server in ./data/some_data.json

``` JSON
	[
		{ "abc": "page-1", "myId": 101 },
		{ "abc": "page-2", "myId": 102 },
		{ "abc": "page-3", "myId": 103 }
	]
``` 

A request for `http://www.zepher.com:3210/data/some_data.json?__template__=data_render_body`
will do the following:

1. The request works its way down to the `file_server`.
2. JSONToTable converts the returning text to table data internally.
3. GoTemplate takes the table data and applies the templates.  `data_render_body` creates a header, then iterates over the set of rows, then adds the footer.

The url: `http://www.zepher.com:3210/data/some_data.json?__template__=data_render_body` will produce the following:

``` HTML
	<!DOCTYPE html>
	<html lang="en">
	<body>
		<div> header </div>
		<ul>
			<li><a href="/page-1.html"> page-1.html id:101 </a></li>
			<li><a href="/page-2.html"> page-1.html id:102 </a></li>
			<li><a href="/page-3.html"> page-1.html id:103 </a></li>
		</ul>
		<div> footer </div>
	</body>
	</html>
``` 

In this case any source of table data or a row of data can then be rendered into a final output form.

### Tested

Tested On: Wed Mar  2 10:01:28 MST 2016 - Unit Tests

Tested On: Wed Mar  3 12:40:48 MST 2016 - End to End Tests of Templates.

### TODO

TODO - Have links to Go templates and how to use them.

