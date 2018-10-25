LoginRequired: Middleware After this Require Login
==================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Require Login"
	,	"SubSectionGroup": "Authentication"
	,	"SubSectionTitle": "Requrie Login"
	,	"SubSectionTooltip": "Require login before allowing access to the specified paths below this in the middleware stack"
	, 	"MultiSection":2
	}
```

Limit serving of files to the specified set of extensions.  If the file is not one of the specified

Each of the middleware after this in the processing stack will require a login via AesSrp.
This middleware also works with the BasicAuth, BasicAuthRedis, BasicAuthPgSQL.

This tests to verify if a successful login has been passed at a previous point in the
processing.  The top level of the processing reserves a set of parameters like `$is_logged_in$`.
During login, if the login is successful, then this parameter will be set to `y`.  That gets
checked by this middleware.

If "StrongLoginReq" is set to  "yes" then the parameter `$is_full_login$` is also checked to
be a `y`.  This is set to `y` when login has happened and if configured for it, two factor
authentication has taken place.

Why this works
--------------

At the top level the server (top) will remove the parameters $is_logged_in$ and $is_full_login$.  If the parameters
are found then they will get converted into "user_param::$is_logged_in$" and "user_param::$is_full_login$".
Then if login occurs it can set the params and this can see them.

Configuration
-------------

For the paths that you want to protect with this turn on basic auth, or use the AesSrp
authentication.  In the server configuration file:

``` JSON
	{ "LoginRequired": {
		"Paths": [ "/PrivateStuff" ],
		"StrongLoginReq":  "yes"
	} },
``` 


Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "LoginRequired": {
					"Paths": [ "/private1", "/private2" ],
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

As a part of the AesSrp login process.

