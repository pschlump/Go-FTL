RejectIPAddress: Ban Certain IP Address
=======================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Ban IP Addresses"
	,	"SubSectionGroup": "Limit Requests"
	,	"SubSectionTitle": "Ban requests based on IP"
	,	"SubSectionTooltip": "Prevent access to site based on IP address"
	, 	"MultiSection":2
	}
```

Limit serving of files to the specified set of extensions.  If the file is not one of the specified

Allows for the banning of specific IP addresses.  If a matching IP address is found, then a
HTTP Status Forbidden (403) error will be returned.

Planned:  Adding ability to match ranges and sets of IP addresses. 

Also you can block based on geographic location using geoIPFilter.

Configuration
-------------

You can provide a simple list of IP addresses, either IPv4 or IPv6 addresses.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RejectIPAddress": { 
					"Paths":   "/api",
					"IPAddrs": [ "206.22.41.8", "206.22.41.9" ]
				} },
			...
	}
``` 

or you can provide a Redis prefix where a successful lookup will result in a
HTTP Status Forbidden (403) error.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RejectIPAddress": { 
					"Paths":            "/api",
					"RedisPrefix": 		"reject-ip|"
				} },
			...
	}
``` 

If both IPAddrs and RedisPrefix are provided, then an error will be logged and the RedisPrefix will be used.  
To apply to all paths use a "Paths" of "/".

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "RejectIPAddress": { "LineNo":5, 
					"Paths":   "/api",
					"IPAddrs": [ "206.22.41.8", "206.22.41.9" ]
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

Thu Feb 25 12:37:05 MST 2016

### TODO

Add IP Ranges/Patterns: see /Users/corwin/Projects/IP/ip.go

