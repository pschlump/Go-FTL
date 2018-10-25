Gzip: Compression of responses using gzip
=======================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Gzip Compression"
	,	"SubSectionGroup": "Performance"
	,	"SubSectionTitle": "Gzip output before returning it"
	,	"SubSectionTooltip": "Gzip compresses output before it is returned.  Interacts with caching so 'zip' process only happens if file changed"
	, 	"MultiSection":2
	}
```

Gzip allows for compression of return data.  It may pose a security risk if used in
combination with HTTPS.  The security risk is a timing attack. It is mitigated
by using caching that causes the gzip compression to only run when the file 
changes.

The default is to compress anything that is larger than 500 bytes.

Configuration
-------------

Gizp any data that is larger than 1,000 bytes and is from the /static directory.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Gzip": { 
					"Paths":   "/static",
					"MinLength": 1000
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
				{ "Gzip": { "LineNo":5, 
					"Paths":   "/static",
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

Sat Feb 27 08:04:50 MST 2016

Tue May  3 09:14:29 MDT 2016 -- After changes to work with caching.


