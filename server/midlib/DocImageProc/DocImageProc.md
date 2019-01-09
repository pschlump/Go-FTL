HostToCustomerId: Convert host name to customer id
==================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Basic Auth/Redis"
	,	"SubSectionGroup": "Authentication"
	,	"SubSectionTitle": "Convert host to customer id"
	,	"SubSectionTooltip": "Support for multiple customers in a single database"
	, 	"MultiSection":2
	}
```

This middleware uses a lookup in Redis to convert from host names to `customer_id` and
injects this as the `$customer_id$` parameter.

Configuration
-------------

For the paths that you want to protect with this turn on basic auth.  In the server configuration file:

``` JSON
	{ "HostToCustomerId": {
		"Paths": [ "/" ]
	} },
``` 

A sample setup for Redis is in: `redis-setup.redis`.  To run

``` Bash
	$ redis-cli <redis-setup.redis
``` 

### Tested
		

