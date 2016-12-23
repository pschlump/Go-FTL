Where Clause on Select, Update, Delete
====================
``` JSON
	{
		"Section": "TabServer"
	,	"SubSection": "Where Clause"
	,	"SubSectionGroup": "Config"
	,	"SubSectionTitle": "where clause"
	,	"SubSectionTooltip": "Details on the API where clause"
	, 	"MultiSection":2
	}
```



``` green-bar
	?where={...}
```

This is a JSON encoded string with a parse tree in it.
Each of the constants from this will be used as a bind 
variable in the select.

``` green-bar
	?where={"op":"and","List":[...]}
```

A list of and-ed together clauses in the where.

```
	?where={"op":"and","List":[{"op":"between","name":"DateColumName","Val1d":"2006-02-08T14:22:23","Val2d":"***2nd ISO Date/DateTime"},{...}]}
```

Op's are: `between`, `not between`, `<`, `>`, `==`, `!=`, `<>`, `>=`, `<=`, `like`, `not like`, `in`, `not in`

"name" is the column name to do the operation on.

Values are: Val1d, Val2d - dates or date/time in ISO format, YYYY-MM-DDTHH24:MI:SS.sssssss
Val1s, Val2s - strings.  Val1b, Val2b - boolean.  Val1i, Val2i as integers.
Val1f, val2f as floats.   

``` green-bar
	?where={"op":"and","List":[{"op":"in","name":"ColName","List":[{"Val1s":"abc"},{"Val1s":"def"}] }] }
```

becomes

``` green-bar
	where ColName in ( $1, $2 )
```

with values of $1 == "abc" and $2 == "def".  It is not generally a good idea to use floating point data and
in lists in combination.  Val1? are used for values, so Val1s, Val1d, Val1i, Val1b, Val1f.

Expressions can be inclued as r-values.  To get `where ColName = ( 12 + 14 )` you can

``` green-bar
	?where={"op":"and","List":[{"op":"=","name":"ColName","Expr":[{"op":"+","Expr":[{"Val1i":12},{"Val1i":14}] }] }] }
```

It is my intent to build a SQL to Expression translator - SOON!  That would allow you to put in a string form
of the "where" clause and get the JSON parse tree that is equivalent.  This would be a purely development tool.

