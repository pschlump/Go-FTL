Geth: Interface to Geth / Ethereum
=======================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Geth Interface"
	,	"SubSectionGroup": "Interface"
	,	"SubSectionTitle": "Geth - Interace for Ethereum"
	,	"SubSectionTooltip": "xyzzy"
	, 	"MultiSection":2
	}
```

Geth xyzzy - overview

Example HTTP calls
------------------

```

http://www.example.com/geth/ContractInfo
	-> Version of Contract 	-> ( v1.0.0 )
	-> Address
	-> Date Loaded
	-> Owner Info
	-> In Contract Registry

http://www.example.com/geth/call?contract=ManageCorpReg&method=CreateNewCorp?Name=Sam&TokenName=SAMTOKEN&AbrevTok=ST&Qty=12000&Decimal=0
	-> ID returned
	-> Address of new contract
	-> Contract Name

http://www.example.com/geth/call?contract=SamToken&method=ListAccounts?
	-> { data: [ ... ] }

```

1. What about Q'ed stuff - may take 30 sec for block in Eth, why not just Q some stuff and run later
	1. Get a "Status" / Log of what has made it to on chain.
	2. Keep "password" for "unlock" local to config 
2. This means that the front end is just a Web/App - now the back end handls all the "ETH" stuff

Configuration
-------------

Gizp any data that is larger than 1,000 bytes and is from the /static directory.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Geth": { 
					"Paths":   "/geth",
					"ContractList": [ "ManageCorpReg" ],
					"ABIPath": [ "./abi" ]
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
				{ "Geth": { "LineNo":5, 
					"Paths":   "/static",
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

Mon Mar 12 10:51:50 MDT 2018 -- Original design



