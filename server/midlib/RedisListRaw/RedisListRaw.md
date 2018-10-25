RedisListRaw: Return Data from Redis
====================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Redis Data Raw"
	,	"SubSectionGroup": "Data"
	,	"SubSectionTitle": "Access sets of keys in Redis"
	,	"SubSectionTooltip": "Provide limited access to data in Redis based on prefixes to a set of keys.  Return data in an unformated form so that other middlware can easliy access it."
	, 	"MultiSection":2
	}
```

Limit serving of files to the specified set of extensions.  If the file is not one of the specified

This allows for retrieving data from Redis that has a common prefix.

The data is returned as "raw" table data - it has not been converted into JSON or other text.   Pre-converted text can be had with RedisList.

Configuration
-------------

You can provide a simple list of IP addresses, either IPv4 or IPv6 addresses.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RedisListRaw": { 
					"Paths":           "/api",
					"Prefix":          "pf3:",
					"UserRoles":       [ "anon,$key$", "user,$key$,confirmed", "admin,$key$,confirmed,disabled", "root,name,confirmed,disabled,disabled_reason,login_date_time,login_fail_time,n_failed_login,num_login_times,privs,register_date_time" ]
					"UserRolesReject": [ "anon-user" ]
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
				{ "RedisListRaw": { "LineNo":5, 
					"Paths":   "/api",
					"Prefix":          "pf3:",
					"UserRoles":       [ "anon,$key$", "user,$key$,confirmed", "admin,$key$,confirmed,disabled", "root,name,confirmed,disabled,disabled_reason,login_date_time,login_fail_time,n_failed_login,num_login_times,privs,register_date_time" ]
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

Tested On: Sat Apr  9 13:08:03 MDT 2016

### TODO

Allow for other Redis types. - Currently only allows for name/value key pair.


