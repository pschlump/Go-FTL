Minify: Compress/Minify Files Before Serving Them
=================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Minify Files"
	,	"SubSectionGroup": "Performance"
	,	"SubSectionTitle": "Shrink the size of output"
	,	"SubSectionTooltip": "Shrink output using minification techniques.  Compress CSS, JavaScript, HTML, SVG, XML and JSON data."
	, 	"MultiSection":2
	}
```

Limit serving of files to the specified set of extensions.  If the file is not one of the specified

This provides on-the-fly compression and minimization of a number of different file types.  Currently all the files are
text based.

If used in combination with InMemoryCache the files will be cached.  The cache will automatically flush if the original
source file is changed.

Configuration
-------------

You can provide a simple list of IP addresses, either IPv4 or IPv6 addresses.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Minify": { 
					"Paths":   "/api",
					"FileTypes": [ "html", "css", "js", "svg", "json", "xml" ]
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
				{ "Minify": { "LineNo":5, 
					"Paths":   [ "/www/", "/static/" ],
					"FileTypes": [ "css", "js", "svg", "json", "xml" ]
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

Tesed On: Fri Mar 11 09:05:10 MST 2016

### TODO and Notes/Caveats

1. Using the node/npm UglifyJS middleware produces better results for minifying JavaScript than the internal Go code in this middleware.  Consider using that (accessible via the file_server middleware) instead of this.
2. Compression of images.
3. Compression of HTML will remove the `<body>` tag.  This can cause some client side JavaScript to break.

