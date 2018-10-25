Else: Return a Page for a Failed Virtual Host Name or SNI Match
===============================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Server"
	,	"SubSectionGroup": "Misc"
	,	"SubSectionTitle": "Else"
	,	"SubSectionTooltip": "This may not be working yet.  Under Construction"
	, 	"MultiSection":2
	, 	"RmFromLeftMenu":true
	}
```

It is possible that a server will receive requests that do not match any name or IP address.
An example would be a new, or unconfigured DNS name, that maps to the current IP address of
the Go-FTL server.  If an `Else` middleware is used, then a list of resolvable names
will be displayed to the user and the user can click on one of the links.

There is only 1 `Else` for all named servers.  Usually it is placed at the bottom of the
configuration file.

Configuration
-------------

Not much configuration.  The only option is to have a message that displays before the
list of configured servers.

``` JSON
	{
		...
		...
		...
		...
		"elseServer": { 
			"listen_to":[ "*" ],
			"plugins":[
				{ "Else": { 
					"Paths":   "/",
					"Msg": "<h1> This is the Go-FTL server for: </h1>"
				} }
	}
``` 

### Tested

Tested On: Fri Mar 11 07:46:02 MST 2016

