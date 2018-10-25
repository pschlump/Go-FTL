Prefix: Allows configuration of a "prefix" before JSON responses
===============================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Handle Prefix"
	,	"SubSectionGroup": "Request Processing"
	,	"SubSectionTitle": "Handle Prefix requests"
	,	"SubSectionTooltip": "Transorm get reqeusts into Prefix if they have a callback parameter"
	, 	"MultiSection":2
	}
```

Adding a prefix (like Google or Facebook) helps to prevent the direct execution of JSON
code.  AngularJS supports `)]}',\n` as a prefix by default.

``` json

	where(1);{"json":"code"}

```

or

``` json

	)]};{"json":"code"}

```

This addresses [a known JSON security vulnerability](http://haacked.com/archive/2008/11/20/anatomy-of-a-subtle-json-vulnerability.aspx/).

Both server and the client must cooperate in order to eliminate these threats.
This implements the server side for mitigating this attack.
Angular comes pre-configured with strategies that address this issue, but for this to work backend server cooperation is required.
Other front end packages will use a different prefix.  You can set the prefix, but the default is for Angular.

JSON Vulnerability Protection
-----------------------------

A JSON vulnerability allows third party website to turn your JSON resource URL into JSONP request under some conditions.
To counter this your server can prefix all JSON requests with following string ")]}',\n".
The Client must automatically strip the prefix before processing it as JSON.

For example if your server needs to return:

``` json

	['one','two']

```

which is vulnerable to attack, your server can return:

``` json

	)]}',
	['one','two']

```


Configuration
-------------

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Prefix": { 
					"Paths":  "/api",
					"Prefix": ")]}',\n"
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
				{ "Prefix": { "LineNo":5, 
					"Paths":   "/api",
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

[//]: # (This may be the most platform independent comment)

Tested On: Tue Jun 21 08:26:53 MDT 2016


