BasicAuthRedis: Basic Auth using Redis
======================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Basic Auth/Redis"
	,	"SubSectionGroup": "Authentication"
	,	"SubSectionTitle": "Basic Authentication"
	,	"SubSectionTooltip": "Basic Auth implemented with data stored in Redis"
	, 	"MultiSection":2
	}
```

This middleware implements HTTP basic auth with the authorization stored in Redis.

The package used to access the Redis database is:

	https://github.com/garyburd/redigo/redis

Pbkdf2 is used to help prevent cracking via rainbow tables.  Each hashed password
is strengthened by using salt and 5,000 iterations of Pbkdf2 with a sha256 hash.

Basic auth should only be used in conjunction with TLS (https).  If you need to use
an authentication scheme with http, or you want a better authentication scheme,
take a look at the aessrp.go  middleware.  There are examples of using it with
jQuery and AngularJS 1.3 (2.0 of AngularJS coming soon).   

Also this is "basic auth" with the ugly browser popup of username/password and no
real error reporting to the user.  If you want something better switch to the SRP/AES
solution.

Remember that rainbow tables can crack MD5 hashes in less than 30 seconds 95%
of the time.  So... this is only "basic" auth - with low security.

So what is "basic" auth really good for?  Simple answer.  If you need just a
touch of secruity - and no more.   Example:  You took a video of your children
and you want to send it to Grandma.  It is too big for her email so
you need to send a link. So quick copy it up to your server and set basic
auth on the directory/path.  Send her the link and the username and password.
This keeps googlebot and other nosy folks out of it - but it is not really
secure.  Then a couple of days later you delete the video.   Works like a
champ!

There is a command line tool in ../../tools/user-redis/user-redis.go to maintain the data
in the Redis database.  You can create/update/delete users from the database.  Also the
tool is useful for verifying that you can connect to the database.

The database connection information is in the global-cfg.json file.

Configuration
-------------

For the paths that you want to protect with this turn on basic auth.  In the server configuration file:

``` JSON
	{ "BasicAuthRedis": {
		"Paths": [ "/video/children", "/family/pictures" ],
		"Realm": "example.com"
	} },
``` 

A sample setup for Redis is in: `redis-setup.redis`.  To run

``` Bash
	$ redis-cli <redis-setup.redis
``` 

To run this you must have valid connection info in ../test_redis.json.

### Tested
		
Tested on : Thu Mar 10 16:00:44 MST 2016, Version 0.5.8 of Go-FTL with Version 2.8 of Redis.

