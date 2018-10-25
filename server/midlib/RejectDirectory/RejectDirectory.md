RejectDirectory: Prevent Browsing of a Set of Directories
=========================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Ban Directory"
	,	"SubSectionGroup": "Limit Requests"
	,	"SubSectionTitle": "Prevent access to a set of directories"
	,	"SubSectionTooltip": "Limit all access to a set of directories"
	, 	"MultiSection":2
	}
```

Limit serving of files to the specified set of extensions.  If the file is not one of the specified


RejectDirectory allows for a set of directories to be un-browsable.   Files from the directories
can still be served - but the directories themselves would not be browsable.

If you do not want anything served from the directory, then use "LimitRePath".

This is implemeted inside the "file_serve." - This middlware just sets configuration for 
"file_serve".

Configuration
-------------

Specify a path and a set of specific directory to not be browsable.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RejectDirectory": { 
					"Paths": [ "/static" ],
					"Disalow": [ "/static/templates" ]
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
				{ "RejectDirectory": { "LineNo":5, 
					"Paths":   "/static",
					"Disalow": [ "/static/templates" ],
				} },
				{ "DirectoryBrowse": { "LineNo":5, 
					"Paths":   "/static",
					"TemplateName": "index.tmpl"
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

Wed, Mar 2, 10:01:28 MST, 2016

