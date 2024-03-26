Go-FTL Configuration
====================
``` JSON
	{
		"Section": "Overview"
	,	"SubSection": "Examples"
	,	"SubSectionGroup": "Server Config"
	,	"SubSectionTitle": "Configure Go-FTL"
	,	"SubSectionTooltip": "Go-FTL configuration examples"
	, 	"MultiSection":2
	}
```


The configuration file is in JSON.  It is very easy to configure Go-FTL.   This file has a set of progressivly
more complex examples in it.

## Example 1 - Simple - Listen for one machine

``` json

	{
		"demo1": { "LineNo":__LINE__,
			"listen_to":[ "http://localhost:8080", "http://dev2.test1.com:8080" ],
			"plugins":[
				{ "file_server": { "LineNo":__LINE__, "Root":"/www", "Paths":"/"  } }
			]
		}
	}

```

Listen on `localhost` port 8080 and on `http://dev2.test1.com:8080` for requests and serve them.   The directory with the files to server is ./www.


## Example 2 - Simple - Listen for one machine and gzip the data being sent back.

``` json

	{
		"http://localhost:8080/": { "LineNo":__LINE__,
			"listen_to":[ "http://localhost:8080", "http://dev2.test1.com:8080" ],
			"plugins":[
				{ "Gzip": { "LineNo":__LINE__, 
					"Paths":   "/www/static",
					"MinLength": 500
				} },
				{ "file_server": { "LineNo":__LINE__, "Root":"/www", "Paths":"/"  } }
			]
		}
	}

```

Listen on localhost port  8080 for requests and serve them.   The directory with the files to server is ./www.
Add the middleware "gzip" with a 500 byte minimum.  It will gzip any data returned (if the client accepts
gzip) that is over 500 bytes in size.

This shows how to pipe results from the `file_server` through another layer `gzip` before it is returned.

## Example 3 - Simple - Listen for both http and https requests.

``` json

	"demo_server": {
		"listento": [ "http://localhost:8080/", "https://localhost:8081" ],
		"certs": [ "/home/pschlump/certs/cert.pem", "/home/cpschlump/crts/key.pem" ],
		"root": "./www",
		"gzip": {
			"minsize": 500,
			"httponly": true
		}
	}

```

The "name" of the server is "demo_server".
List for http request on 8080 and for https on 8081.  Serve ./www.  Note the limitation on gzip as it
may be a security risk when combined with https.   The certificates are specified with the "certs" 
options. 

## Example 4 - Multiple name resolved servers.

``` json

	"demo_server": {
		"listento": [ "http://localhost:8080/", "https://localhost:8081" ],
		"certs": [ "/home/pschlump/certs/cert.pem", "/home/cpschlump/crts/key.pem" ],
		"root": "./www/demo_server",
		"gzip": {
			"minsize": 500,
			"httponly": true
		}
	}

	"test_server": {
		"listento": [ "http://test.2c-why.com/", "https://test.2c-why.com" ],
		"certs": [ "/home/pschlump/certs/test.2c-why.com/cert.pem", "/home/pschlump/crts/test.2c-why.com/key.pem" ],
		"root": "./www/test.2c-why.com",
		"gzip": {
			"minsize": 500,
			"httponly": true
		}
	}

```

Listen and server two different sets of pages.  The default ports are used for "test_server" with 80 for http
and 443 for https.








