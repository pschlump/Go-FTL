JSONToTable: Convert JSON to Internal Table Data
================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Data to JSON"
	,	"SubSectionGroup": "Output Formatting"
	,	"SubSectionTitle": "Convert raw tabular data into a JSON format before returing it."
	,	"SubSectionTooltip": "Format data into JSON"
	, 	"MultiSection":2
	}
```

Convert data in JSON format into internal table data in the response buffer.

By itself this is not very useful.  However when combined with a template
it allows for JSON data to be read from a file and then formatted into a
final set of data.

Configuration
-------------

A number of options are planned. (See TODO below.)

ConvertRowTo1LongTable:  If this is true, then
a single row of data will be converted into an array 1 long.   If the data is empty,
then an empty array will be returned.

Convert1LongTableToRow: If this is true, then
a table that is 1 row long, (or 0), will be converted to a hash.

Both flags can not be true at the same time.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "JSONToTable": { 
					"Paths":   "/api",
					"ConvertRowTo1LongTable": true,
					"Convert1LongTableToRow": false
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
				{ "GoTemplate": { "LineNo":5, 
					"Paths":   "/config/initialSetupData.json",
					"TemplateName": "initialSetupData.tmpl",
					"TemplateRoot": "/tmpl/"
				} },
				{ "JSONToTable": { "LineNo":5, 
					"Paths":   "/config/initialSetupData.json",
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


### Tested

Tested On: Fri, Mar 11, 12:15:38 MST, 2016

