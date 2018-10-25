Cookie: Set/Delete Cookies
==========================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Set/Get Cookies"
	,	"SubSectionGroup": "Headers"
	,	"SubSectionTitle": "Set/Delete Cookies"
	,	"SubSectionTooltip": "Manipulation of cookies"
	, 	"MultiSection":2
	}
```

Create headers to set or delete cookies.

Configuration
-------------

Name and Value are required.  Other configuration options for the cookie are optional.  Normally Domain will
also need to be set.  If you want your cookie to be available to `www.example.com` and `cdn.example.com,` then use
`.example.com`.  

Use only one of `MaxAge` and `Expires`.  To delete a cookie set the value to an empty `Value`, `""` and `MaxAge` to `-1`.

In this example the path `/somepath` will get a cookie named `testcookie` with a value of `1234`.  The cookie 
expires in a very confusing `12001` seconds or in 2018 (not good, but this is an example).  This is not
a secure cookie.

Secure cookies can only be set when using HTTPS.


``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Cookie": { 
					"Paths":    "/somepath",
					"Name":     "testcookie",
					"Value":    "1234",
					"Domain":   "www.example.com",
					"Expires":  "Thu, 18 Dec 2018 12:00:00 UTC",
					"MaxAge":   "12001",
					"Secure":   false,
					"HttpOnly": false
				} },
			...
		]
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "Cookie": { 
					"Paths":    "/somepath",
					"Name":     "testcookie",
					"Value":    "1234",
					"Domain":   ".zepher.com",
					"Expires":  "Thu, 18 Dec 2018 12:00:00 UTC"
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

Thu, Mar 10, 13:11:43 MST, 2016

### TODO

Use template substitution on the cookie name and value.

Add a "Delete" flag that correctly sets the values for a delete with a single flag.

