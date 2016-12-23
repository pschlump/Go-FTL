Table API - the CDUD
====================
``` JSON
	{
		"Section": "TabServer"
	,	"SubSection": "API"
	,	"SubSectionGroup": "Config"
	,	"SubSectionTitle": "API"
	,	"SubSectionTooltip": "Details on the API suported in TabServer2"
	, 	"MultiSection":2
	}
```



These interfaces are created in crud.go, lines 68 to 82.

## Insert

HTTP, method POST

``` green-bar
	/api/table/NAME?col=Value&col2=value...
```

If you supply the ID as Primary key it will be used.
The ID/PK is specified in the `sql-cfg.json` file.

## Insert - with ID

HTTP, method POST

``` green-bar
	/api/table/NAME/ID?col=Value&col2=value...
```

If you supply the ID as Primary key it will be used.
The ID/PK is specified in the `sql-cfg.json` file.

## Update

HTTP, method PUT

``` green-bar
	/api/table/NAME?col=Value&col2=value...
```

The primary key can be a multi-part PK and must be supplied 
in the parameters.   It is possible to do non-PK updates with
this.  There is a flag in `sql-cfg.json` for this.
The ID/PK is specified in the `sql-cfg.json` file.

## Update - with ID

HTTP, method PUT

``` green-bar
	/api/table/NAME/ID?col=Value&col2=value...
```

If you supply the ID as Primary key it will be used.
The ID/PK is specified in the `sql-cfg.json` file.


## Delete

HTTP, method DELETE

``` green-bar
	/api/table/NAME?col=Value&col2=value...
```

The primary key can be a multi-part PK and must be supplied 
in the parameters.   It is possible to do non-PK updates with
this.  There is a flag in `sql-cfg.json` for this.
The ID/PK is specified in the `sql-cfg.json` file.

## Delete - with ID

HTTP, method DELETE

``` green-bar
	/api/table/NAME/ID?col=Value&col2=value...
```

If you supply the ID as Primary key it will be used.
The ID/PK is specified in the `sql-cfg.json` file.

## Select 

Select can be performed in both

``` green-bar
	/api/table/NAME/ID
```

and 

``` green-bar
	/api/table/NAME?col=Val1&col2=val2
```

format.   It is assumed that this is for a single row - primary or
unique key select.

Select provides a set of other options.  You can not name columns
with these options and use them in the `?col=val1` format.

``` green-bar
	?orderBy=[{"ColName":"abc","Dir":"asc"}]
```

is a JSON encoded array of columns with "asc" and "desc".
Note lower case on asc/desc.  Also the "Dir" is optional
and assumed to be "asc" if not specified.

``` green-bar
	?where={...}
```

a where clause - as a parse tree subset of a SELECT were
clause.  More on this later.

``` green-bar
	?limit=
	?offset=
```

Are integer values that will subset the query to the
specified range.  These are optional.




