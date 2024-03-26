Database - Features in Detail
=============================
``` JSON
	{
		"Section": "Overview"
	,	"SubSection": "Quickstart"
	,	"SubSectionGroup": "Database"
	,	"SubSectionTitle": "Go-FTL - Database Access"
	,	"SubSectionTooltip": "Go-FTL Database Access"
	, 	"MultiSection":2
	}
```

Go-FTL has extensive middleware to dedicated to database access.  PostgreSQL is the intended relational database however it is possible to use Oracle, Microsoft T-SQL, or MySQL.
Support for MySQL is still under development. 

The tab at the top, `TabServer` is all about configuration for database access.

A Quick Example.  You want to create a contact form:  First create the table in the database:

``` sql

	create table "p_issue" (
		  "id"						char varying (40) DEFAULT uuid_generate_v4() not null primary key
		, "title"					char varying (250)
		, "desc"					text		
		, "type_group"				char varying (50)		-- webpage / product name etc // [ Notification - ask for help ]
	);

```

Now create a TabServer configuration file with:

``` json

	{
		 "/api/table/contact": { "crud": [ "select", "insert", "update", "delete", "info" ]
			, "Comment": "Save Contact Requests"
			, "TableName": "p_issue"
			, "LineNo":"__LINE__, File:__FILE__"
			, "LoginRequired":false
			, "Method":["GET","POST","PUT","DELETE"]
			, "ReturnMeta":false
			, "ReturnAsHash":true
			, "cols": [
				  { "colName": "id"    		 , "colType": "s",	               "insert":true, "autoGen":true, "isPk":true 							}
				, { "colName": "title"	 	 , "colType": "s",	"update":true, "insert":true						, "DataColName":"subject"		}
				, { "colName": "desc"		 , "colType": "s",	"update":true, "insert":true														}
				, { "colName": "type_group"	 , "colType": "s",	"update":true, "insert":true, "DefaultData":"contact"					 			}
				]
			}
	}

```

We have turned off login on this with, `"LoginRequired":false`.  The API end point is `/api/table/contact`.  It will respond  to 

* GET requests to perform select
* POST requests to do insert
* PUT requests to do updates
* DELETE requests to do deletes

Add in a configuration section in the middleware to use TabServer and...

``` json

	{
		"working_test_AngularJS_20": { "LineNo":__LINE__,
			"listen_to":[ "http://localhost:16020", "http://dev2.test1.com:16020" ],
			"plugins":[
				{ "TabServer2": { "LineNo":__LINE__,
					"Paths":["/api/"],
					"AppName": "www.go-ftl.com",
					"AppRoot": "/Users/corwin/Projects/docs-Go-FTL/data/",
					"StatusMessage":"Version 0.0.4 Sun May 22 19:12:43 MDT 2016"
				} },
				{ "file_server": { "LineNo":__LINE__, "Root":"/Users/corwin/Projects/docs-Go-FTL", "Paths":"/"  } }
			]
		}
	}

```

You now have a working API.

![Database Output](./images/db1.png "Image showing output of API request.")

Things To Note
--------------

* Security is baked in.  Use the AesSrp module to provide strong authentication.
* A separate tracing package provides details of what happens with each request and how they get processed into queries.  It is incredibly useful for debugging your front end.
* You can also access Redis with the TabServer middleware.
* You can build complex data to return with joins.
* You can do full word searches and hierarchal/tree searches.
* Deletes can be configured to update the row with a flag as deleted data.
* You can call stored procedures as easily as performing queries.  This means that you can create complex business logic very quickly.
* You can cache tables or rows in Redis.
* The default format for data to be returned is JSON.  However this works with the GoTempalte middleware so that you can take a query and apply a template to the result.  This makes for some quick HTML results!
* All of this can be cached so that if the data has not changed then the cache can return a reasonable result.
* This is *NOT* an ORM.  This is a secure way to provide front end access and configuration based database access to your application.  This means that if you need to build a front end that updates 500,000 rows at a time it is easy to do and you don't need to end up with 500,000 update statements.
* You can (and should) validate the input at the level of the TabServer.

File: ./doc/Database_01.md




