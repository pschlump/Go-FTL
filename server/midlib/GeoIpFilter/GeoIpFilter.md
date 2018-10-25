GeoIpFilter: Filer Requests Based on Geographic Mapping of IP Address
=====================================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Geographic Fitler"
	,	"SubSectionGroup": "Limit Requests"
	,	"SubSectionTitle": "Geographic Filtering"
	,	"SubSectionTooltip": "Use IP address to filter to a set of geograpic regions"
	, 	"MultiSection":2
	}
```

Limit serving of content to geographic regions based on mapping of IP addresses to these regions.
This works on a per-country basis most of the time.  The data is not 100% accurate.

The data is based on the freely available GetLite2 database.  You need to download your own copy
of this data - the data that is in the ./cfg directory is terribly out of date and should only
be used for testing of this middleware.

Also note: The data changes periodically.   Hopefully one day this module will automatically
update the data - but for the moment you have to update it by hand.

Configuration
-------------

You can provide a simple list of IP addresses, either IPv4 or IPv6 addresses.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "GeoIpFilter": { 
					"Paths":   "/",
					"Action":  "reject",
					"CountryCodes":  [ "JP", "VN" ],
					"DBFileName":    "./cfg/GeoLite2-Country.mmdb",
					"PageIfBlocked": "not-avaiable-in-your-country.html"
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
				{ "GeoIpFilter": { "LineNo":5, 
					"Paths":   "/",
					"Action":  "reject",
					"CountryCodes":  [ "JP", "VN", "CN" ],
					"DBFileName":    "./cfg/GeoLite2-Country.mmdb",
					"PageIfBlocked": "not-avaiable-in-your-country.html"
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

Fri, Mar 11, 09:15:38 MST, 2016

### TODO

1. Add automatic update of underlying data.
1. Improve data quality.


