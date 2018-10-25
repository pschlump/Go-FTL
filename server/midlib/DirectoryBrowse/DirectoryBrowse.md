DirectoryBrowse: Use Template for Directory Browsing
====================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Templated Directories"
	,	"SubSectionGroup": "Authentication"
	,	"SubSectionTitle": "Templated directory browsing"
	,	"SubSectionTooltip": "Control layout and availabity of directory browsing with Go templates"
	, 	"MultiSection":2
	}
```

Implements templated directory browsing.  

You provide a template, (see example below), and place that in one of the directories specified by "Root" option.
If a *directory* is browsed inside the set of "Paths," then the template will be applied to the file names.

If the template fails to parse, or if no template is supplied, then this is logged to the log file.
An error will be returned.

If the tempalte root is not specified, then the root directory for serving files will be searched
for the specified template name.

This is implemeted inside the "file_serve" - this middlware just sets configuration for 
"file_serve".

Configuration
-------------

Specify template name and the location to find it.  The default template name is "index.tmpl".

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "DirectoryBrowse": { 
					"Paths": [ "/static", "/www" ],
					"TemplateName": "dir-template.tmpl",
					"TemplateRoot": [ "/static/tmpl" ]
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
				{ "DirectoryBrowse": { "LineNo":5, 
					"Paths":   "/static",
					"TemplateName": "index.tmpl"
				} },
				{ "DirectoryLimit": { "LineNo":5, 
					"Paths":   "/static",
					"Disalow": [ "/static/templates" ],
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


Example template, Put in index.tmpl

``` HTML
	{{define "content_type"}}text/html; charset=utf-8{{end}}
	{{define "page"}}<!DOCTYPE html>
	<html lang="en">
	<body>
		<ul>
		{{range $ii, $ee := .files}}
			<li><a href="{{$ee.name}}">{{$ee.name}}</a></li>
		{{end}}
		</ul>
	</body>
	</html>
	{{end}}
``` 

### Tested

Wed, Mar 2, 10:05:04 MST, 2016

