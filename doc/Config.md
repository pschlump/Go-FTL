Config - Server Configuraiton
=============================
``` JSON
	{
		"Section": "Overview"
	,	"SubSection": "Global Config"
	,	"SubSectionGroup": "Server Config"
	,	"SubSectionTitle": "Go-FTL - Name based server configuation"
	,	"SubSectionTooltip": "Go-FTL Configure a set of servers"
	, 	"MultiSection":2
	}
```


Go-FTL configuration is in JSON files.  When you run the server by default it will look for two files, `ftl-config.json` and `global-config.json`.

`global-config.json` has global configuration in it like the name/type of the server being run and the connection information for how to authenticate
with Redis and PostgreSQL.  You can set a different global configuration file with the `-g` or `--globalCfgFile` command line option.

`ftl-config.json` is the per-server configuration file.  This has a section in it for each named server that will be run.  You can set the file name
with `-c` or `--cdgFile` option.

For Example, the a `glboal-config.json` file that I use has:

``` json

	{
		"debug_flags": [ "server" ],
		"trace_flags": [ "*" ],
		"server_name": "Go-FTL (v 0.5.9)",
		"RedisConnectHost":  "192.168.0.133",
		"RedisConnectAuth":  "lLJSwkwww3e24wAbr4RM4MWIaBM",
		"PGConn": "user=pschlump password=803728230121123 sslmode=disable dbname=pschlump port=5433 host=127.0.0.1",
		"DBType": "postgres",
		"DBName": "pschlump",
		"LoggingConfig": {
			"FileOn": "yes",
			"RedisOn": "yes"
		}
	}

```

Extensive examples of how to configure `ftl-config.json` are in the next section.  Some middleware components will have additional
configuration files, however most of the configuration is in `ftl-config.json`.


TODO
----

The plan is to allow the sections in `ftl-config.json` to be changed on the fly with a web interface.  That is still under development.



File: ./doc/Config.md 
