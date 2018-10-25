Search Path for global-cfg.json and sql-cfg.json
================================================
``` JSON
	{
		"Section": "TabServer"
	,	"SubSection": "Config"
	,	"SubSectionGroup": "Config"
	,	"SubSectionTitle": "TabServer2 - search for configuration files"
	,	"SubSectionTooltip": "TabServer2 - search for configuration files"
	, 	"MultiSection":2
	}
```

TabServer searches for its configuration files using a search path.
They are named `sql-cfg[AppName].json` where `[AppName]` is the AppName in the configuration
and SearchPath is `~/cfg:./cfg:.` by default.  

`~` will be substituted with your home directory.  `~name/` is substituted for the home directory of the request user.

Examples:

``` gray-bar

		"AppName": "www.go-ftl.com",
		"AppRoot": "/Users/myuser/Projects/www.go-ftl.com/data/",

```

This will set the application name to `www.go-ftl.com` and search with a top level directory of:
`/Users/myuser/Projects/www.go-ftl.com/data/`.   It can find `sql-cfg-www.go-ftl.com.json`.

The search order is:

``` gray-bar

		"%{path_element%}/%{fileName%}-%{AppName%}-%{HostName%}%{fileExt%}",
		"%{path_element%}/%{fileName%}-%{AppName%}%{fileExt%}",
		"%{path_element%}/%{fileName%}-%{HostName%}%{fileExt%}",
		"%{path_element%}/%{fileName%}%{fileExt%}",

```

Where `path_element` is a path our of the `SearchPath` in the order supplied.

`fileName` is the `sql-cfg` section of the file name.

`AppName` is the specified application name.

`HostName` is the name of your computer.

`fileExt` is .json

You can create host-specific global-cfg.json files by putting them in your ~/cfg directory.
For example, you have `pschlump-dev1` and `pschlump-dev2` machines.  If you are not on one of
these then use the default file.

``` gray-bar

	~/cfg/global-cfg-pschlump-dev1.json
	~/cfg/global-cfg-pschlump-dev2.json
	~/cfg/global-cfg.json

```


xyzzy - Notes
------------------------

If the  AppRoot ends in /... then a recursive search for `sql-cfg.*.json` files will take place.

```
	, "AppRoot": 			"/Users/corwin/go/src/github.com/pschlump/Go-FTL/server/midlib/SessionRedis/..."
```

The "search order is" section above is out of date - code has additional file paths.

