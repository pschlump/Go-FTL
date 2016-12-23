TabServer2 - A Go-FTL middleware for building Restful Interfaces to a relational database
=========================================================================================
``` JSON
	{
		"Section": "TabServer"
	,	"SubSection": "Overview"
	,	"SubSectionGroup": "Config"
	,	"SubSectionTitle": "TabServer2 - search for configuration files"
	,	"SubSectionTooltip": "TabServer2 - search for configuration files"
	, 	"MultiSection":2
	}
```

TabServer2 is a Go-FTL middleware that allows the creation of RESTful interfaces to database tables.

A configuration file in JSON determines which tables will have interfaces and what security will be
applied.   Validation can be specified for the parameters.  Parameters are always substituted using
bind variables, never directly into the SQL statements.  This improves efficiency and prevents most
forms of SQL injection attacks.

Stored procedures in the database can be called.  This allows for the creation of business logic at
the level of the database.  Since most data related business logic requires multiple database 
queries the most efficient place to put it is inside the database stored procedures.  Also this
data-centric processing is test and developed quicker when it is in database stored procedures.

Data can be post-joined to produce more complex results.  For example an invoice can be returned
with its invoice details in a single RESTful call.

Searches and updates can use complex where clauses.  The where criteria can be supplied as a parse 
tree from the front end.  The set of columns that can be used in the where can be limited so that
only indexed columns are accessed in the where.

Column names can differ from the named parameters supplied.  By default a 1 to 1 match is assumed.

Every input can be validated.

In PostgreSQL complex keys like document keyword searches and hierarchal searches are directly supported.


