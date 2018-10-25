ZipIt: Return GZIP file for a set of paths
=======================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "ZipIt Send Gziped File"
	,	"SubSectionGroup": "Performance"
	,	"SubSectionTitle": "ZipIt Return specified Gzip Compressed File" 
	,	"SubSectionTooltip": "ZipIt returnes a specified file that is already gziped"
	, 	"MultiSection":2
	}
```

ZipIt allows you to return a zip-bom.  I am tired of getting poked on my static
site for Wordpress /admin directories.   The configuration examples is to return
a 10G gzip compressed file for all /adnim requests.

Amazing how the script-kiddies quit poking my sites after I started doing this.

Configuration
-------------

Gizp any data that is larger than 1,000 bytes and is from the /static directory.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "ZipIt": { 
					"Paths":   "/admin",
					"FileName": "10G.gzip"
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		... TBD ...
	}
``` 


### Tested

Thu Jul  6 11:57:11 MDT 2017


