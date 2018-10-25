Latency: Simulate latency for testing behavior on slow networks (i.e. mobile)
====================================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Add Latency"
	,	"SubSectionGroup": "Debugging"
	,	"SubSectionTitle": "Make requests slow to test latency in network."
	,	"SubSectionTooltip": "Use this  as a tool when testing your web application.  Slows it way down"
	, 	"MultiSection":2
	}
```

This is a simple middleware that allows slowing down requests.  It is intended to test slow networks like mobile and rural.

Good values to use are 50, for a slow rural network or Version.net mobile, 114 for an average mobile network, 240 for a busy
at 3 in the afternoon mobile network and 522 for my remote Wyoming land line.  By the way this is not an endorsement of
Verison.net in anyway - they claim  to have 50ms latency - but my tests indicate 148 is *much* more likely.

Configuration
-------------

If the `SlowDown` is not specified, then 500ms will be used (1/2 second).

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Latency": { 
					"Paths":   "/slowDownPath/",
					"SlowDown": 500
				} },
			...
	}
``` 

Full Example
------------

This full example slows down the results of every request by 114ms.  That is the average that I seed when I test on
ATT's mobile network in my remote Wyoming locaiton.


``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "Latency": { "LineNo":5, 
					"Paths":   "/",
					"SlowDown": 114
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

Tested On: Wed Jun 15 09:26:46 MDT 2016

