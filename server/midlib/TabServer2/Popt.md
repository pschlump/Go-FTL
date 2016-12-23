Popt and Pname in sql-cfg.json
==============================


Note
----


2 additional fields were added to support extended equals "=" queries when .Query is specified.   These are
Popt and Pname.  They are arrays of names that can be pulled from the URL and substituted (bind-parameter) into
the where portion of the query.

The functionality is

	, Popt: [ "class_id" ]
	, Pname [ "cc1.\"id\" ]

with URL

	http://localhost:17050/api/getClassOutline?api_table_key=8HbvBJR1qWkpMgGxzqj2sRvuXg&user_id=1&class_id=c1

allows for optional a substitution of class_id with a value of 'c1' into the were.  The template

	where uc1."user_id" = $1 %{where_additional_params%}

must be specified and where must be "where " or "WHERE " and should not appear any other place in the
query (column names ending in "where" will muck this up).

Functionally it allows specifying additional "and" clauses on the URL by just listing the name=value on the
where clause.


Files Changed
	lib.go - added Popt, Pname columns
	crud.go - in 
		func (hdlr *TabServer2Type) RespHandlerSQL(res http.ResponseWriter, req *http.Request, cfgTag string, ps *goftlmux.Params, rw *goftlmux.MidBuffer) {
	near
		if strings.Index(HQuery, "%{where_additional_params%}") >= 0 {
	or places having addNames, addVals in that function.


Note: This may interact badly in combination with redis-caching of rows.  This has not been tested in that specific case.  

