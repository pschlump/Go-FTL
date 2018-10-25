BasicAuth: Implement Basic Authentication Using a .htaccess File
================================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Basic Auth"
	,	"SubSectionGroup": "Authentication"
	,	"SubSectionTitle": "Basic Authentication"
	,	"SubSectionTooltip": "Basic Auth implemented with a flat file for hashed usernames/passwords"
	, 	"MultiSection":2
	}
```

This middleware implements HTTP basic auth with the authorization stored in a flat file.
If you need to use a database for the storage of usernames/passwords, then you should look
at one of the other two basic-auth middlware.   If you are looking for an example of how
to use a relational database, or how to use a non-relational database, the other basic-auth
middlware are recomended.

Basic auth should only be used in conjunction with TLS (https).  If you need to use
an authentication scheme with http, or you want a better authentication scheme
take a look at the auth_srp.go  middleware.  There are examples of using it with
jQuery and AngularJS 1.3 (2.0 of AngularJS coming soon).   

Also this is "basic auth" with the ugly browser popup of username/password and no
real error reporting to the user.  If you want something better, switch to the SRP/AES
solution.

Remember that rainbow tables can crack MD5 hashes in less than 30 seconds 95%
of the time.  So... this is only "basic" auth - with low security.

So what is "basic" auth really good for?  Simple answer.  If you need just a
touch of secruity - and no more.   Example:  You took a video of your children
and you want to send it to Grandma.  It is too big for her email so
you need to send a link.  So do a quick copy of it up to your server and set basic
auth on the directory/path.  Send her the link and the username and password.
This keeps googlebot and other nosy folks out of it - but it is not really
secure.  Then a couple of days later you delete the video.   Works like a
champ!

There is a command line tool in ./cli-tools/htaccess to maintain the .htaccess
file with the usernames and hashed passwords.

Configuration
-------------

For the paths that you want to protect with this turn on basic auth.  In the server configuration file:

``` JSON
	{ "BasicAuth": {
		"Paths": [ "/video/children", "/family/pictures" ],
		"Realm": "myserver.com"
	} },
``` 

With the "AuthName" you can set the name of the authorization file.  It defaults to .htaccess in the current directory.  

``` JSON
	{ "BasicAuth": {
		"Paths": [ "/video/children", "/family/pictures" ],
		"Realm": "myserver.com",
		"AuthName": "/etc/go-ftl-cfg/htaccess.conf"
	} },
``` 

If you use this middleware it will also ban fetching .htaccess or whatever you have set for AuthName as a file.

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "BasicAuth": {
					"Paths": [ "/private1", "/private2" ],
					"Realm": "zepher.com",
					"AuthName": "/Users/corwin/go/src/github.com/pschlump/Go-FTL/server/midlib/basicauth/htaccess.conf"
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

Tested On: Thu Dec 17 14:24:25 MST 2015, Version 0.5.8 of Go-FTL

Tested On: Sat Feb 27 07:30:27 MST 2016

### TODO

1. Add check that .htaccess becomes un-fetchable
