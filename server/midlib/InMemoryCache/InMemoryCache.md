InMemoryCache: Ban Certain IP Address
=====================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Caching"
	,	"SubSectionGroup": "Performance"
	,	"SubSectionTitle": "Caching of Pages"
	,	"SubSectionTooltip": "Implements in memory caching of hot resources and on disk caching for other pages"
	, 	"MultiSection":2
	}
```

This is primarily intended as an in-memory cache.  It will also, if configured, cache files to disk.
The cleanup time on disk cached items is by default 1 hour.

Configuration
-------------

Lots of configuration items.

Item | Default | Description
|--- | --- | ---
`Extensions`      | no-default  | The set of file extensions that will be cached.
`Duration`        |          60 | How long, in seconds, to cache in memory.
`IgnoreUrls`      | no-default  | Paths to be ignored - and not cached.  For example "/api/".
`SizeLimit`       |      500000 | Limit on size of items to be cached in memory.  Size in bytes.
`DiskCache`       | no-default  | Set of disk locations to place on-disk cached files.  Used round-robin.  If this item is empty then no disk caching will take place.
`DiskSize`        | 200000000   | Maximum amount of disk space to use for on-disk cached files.
`RedisPrefix`     |    "cache:" | The prefix used in Redis for data stored and updated by this middleware.
`DiskSizeLimit`   |    2000000  | The maximum size for disk-cached items.
`DiskCleanupFreq` |  3600       | How long to keep items in the disk cache.  They are discarded after this number of seconds.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "InMemoryCache": { 
					"Paths":   "/",
					"Extensions":       [ ".js", ".css", ".html" ],
					"Duration":         60,
					"IgnoreUrls":       [ "/api/" ],
					"SizeLimit":        500000,
					"DiskCache":        [ "./cache/" ],
					"DiskSize":         200000000,
					"RedisPrefix":      "cache:",
					"DiskSizeLimit":    2000000,
					"DiskCleanupFreq":  3600
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
				{ "InMemoryCache": { "LineNo":5, 
					"Paths":   "/api",
					"Extensions":       [ ".js", ".css", ".html" ],
					"IgnoreUrls":       [ "/api/" ],
					"DiskCache":        [ "./cache/" ],
					"DiskSize":         200000000
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

Fri, Mar 11, 09:22:36 MST, 2016

### TODO

1. Extensive testing with multiple components and the InMemoryCache at the same time.  For example verify that TabServer2 can/will correctly set cache timeout when used with this component.
2. Add the set of mime types to cache - instead of file extensions.
3. Make the file extensions consistent across the Go-FTL system.   In other places the extension `.js` is just `js`.

