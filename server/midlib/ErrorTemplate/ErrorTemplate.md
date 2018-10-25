ErrorTemplate: Convert Errors to Pages
======================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Templated Errors"
	,	"SubSectionGroup": "Logging"
	,	"SubSectionTitle": "Use templates to control what gets logged"
	,	"SubSectionTooltip": "Extended loggin with additional attributes via a template stubstitution"
	, 	"MultiSection":2
	}
```

Map error codes from lower level calls onto template files.

The following items can be used in the template file, `{{.IP}}`. For example:

Item         | Description
|---        | --- 
`IP`         | Remote IP address
`URI`        | Remote URI
`delta_t`    | How long this has taken to process
`host`       | Host name
`ERROR`      | Text error message if any
`method`     | Request method
`now`        | Current time stamp
`path`       | Path from request
`port`       | Port request was made on
`query`      | The request query string
`scheme`     | http or https
`start_time` | Time request was started at
`StatusCode` | Status code, 200 ... 5xx
`StatusText` | Text description of status code

Configuration
-------------

You provide a list of errors that you want to have mapped, with a template, onto 
a page.  You can provide a directory where the templates are for custom error
templates.   If you do not, then the directory `./errorTemplates/` will be used.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "ErrorTemplate": { 
					"Paths":   "/",
					"Errors": [ "404", "500" ]
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
				{ "ErrorTemplate": { "LineNo":5, 
					"Paths":   "/",
					"Errors": [ "404", "500" ]
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

Tested On: Wed, Mar 30, 06:03:59 MDT, 2016

### TODO

1. Way to configure "application" or "home-page" for template.
2. Logging of errors.
3. Possibility of a "form" for errors to contact user when error is fixed.
4. Contact Support info.
5. ./errorTempaltes relative to "root" of application.
6. For users that are logged in - a different template that reflects name/time etc for logged in user.
7. Match "4xx" as an error to a 4xx.tmpl file and a 400 error so you don't have to have zillions of files.

