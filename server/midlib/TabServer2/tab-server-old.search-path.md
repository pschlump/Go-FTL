Search Path for global-cfg.json and sql-cfg.json
================================================


A search path is used, or can be set with the -S command line option.   The default search path is 

``` gray-bar
	~/cfg:./cfg:.
```

You can set a different one with -S.

``` bash
	$ ./tab-server1 -s '/usr/local/cfg:/etc/tab-server1/config' &
```

'~' is substituted for the home directory.  ~name/ is substiuted for the home direcotry of the
request user.

You can create host-sepcific global-cfg.json files by putting them in your ~/cfg diretory.
For example, you have pschlump-dev1 and pschlump-dev2 machines.  If you are not on one of
these then use the default file.

``` gray-bar
	~/cfg/global-cfg-pschlump-dev1.json
	~/cfg/global-cfg-pschlump-dev2.json
	~/cfg/global-cfg.json
```



