TabServer2 - sql-cfg.json configuraiton file
============================================
``` JSON
	{
		"Section": "TabServer"
	,	"SubSection": "Config Examples"
	,	"SubSectionGroup": "Config"
	,	"SubSectionTitle": "Config Examples"
	,	"SubSectionTooltip": "A set of examples with explanation of what is being configured."
	, 	"MultiSection":2
	}
```



## Overview

This configuration file allows limits and configures access to the database without
changes to code.   Most RESTful access can be implemented using this file.   You can

1. Determine what tables are to be accessed via the RESTful interface.
2. Required authentication - or not - on tables and stored procedure calls.
3. Limit the columns that are returned.
4. Create Validation for all inputs.
5. Have validation that is required for individual operations.  For example,
you can require that a primary key be supplied, and that it is a UUID/GUID
for performing a delete on a specific table.
6. Configure REST calls that will access stored procedures in the database.
7. Set tables/rows that are to be cached in Redis.
8. Make changes on the fly with a running TabServer2 server or set of servers.
Make these changes without a server restart.

The configuration file can be watched for changes and the TabServer2 server can
be automatically notified when the file changes.   This is very convenient 
during development.

## Example 001

``` JSON
	,"/api/perfTestDB1": { "query": "select * from posts", "p": [ ], "LoginRequired":false
		, "LineNo":"__LINE__"
		, "valid": {
			 "callback": { }
			}
		}
```

"/api/perfTestDB1" is the RESTful GET call.  

"query" is the SQL query that will be returned.

"p":[] indicates that there a NO parameters that will be substituted as bind 
variables in this call.

"LoginRequired":false indicates that no authentication is required to make this call.

"LineNo":"__LINE__" indicates that the line number will be reported if there is an error.
Two items can be substituted, __LINE__ and __FILE__.   LineNo is a string so you can have
both.  For example:  `"LineNo":"File: __FILE__ LineNo: __LINE__"`.

"valid":{...} is the validation for all POST, GET, PUT, DELETE calls.
Since the method was not specified it defaults to just GET.  This is 
appropriate for a database *select* query operation.

"valid":{"callback":{}"}"  The empty callback option indicates that
this can be called via JSONp.   This is by default a non-required
field.  Requests will only have access to fields that are specified
in one of the validation sections.   You can have method-specific
validation fields.  If "callback" had not been specified then JSONP
would not be allowed.   "validDel", "validGet", "validPut", "validPost"
override this section and match with the "DELETE", "GET", "PUT" and "POST"
methods.  There is no validation on "HEAD" requests.

"TableName" was NOT specified.  This means that TraceRest will not
know what table was accessed when this operation is called.






## Exampel 002

``` JSON
	,"/api/saveJs": { "Fx": "e_js_save"
		, "LineNo":"__LINE__"
		, "LineNo":"171"
		, "p": [ "user", "id", "desc", "title" ]
		, "Method":["POST"]
		, "valid": {
			 "user": { "required":true, "type":"string", "min_len":4, "max_len":100 }
			,"id": { "required":true, "type":"string", "max_len":40 }
			,"desc": { "required":false, "type":"string", "max_len":400 }
			,"title": { "required":false, "type":"string", "max_len":400 }
			,"auth_token": { "required":true, "type":"uuid" }
			}
		}
```

"f" specifies a query that calls a stored procedure - no return value is expected from the
stored procedure.

"p":["user","id","desc","title"] specifies the parameters that will be bound to the stored 
procedure.

"Method":["POST"]  Specifies that this call will only respond to "POST" calls.  This is
appropriate for a stored procedure that saves/updates data in the database.   Using a "GET" 
call (other that in a known development/test environment) when calling for a database
change is not an advised activity.

"LineNo":"171" hard codes the line number to 171 for some reason.  Usually better to use __LINE__ or __LINE__ and __FILE__.
 
"user":{"required":true,"type":"string","min_len":4,"max_len":100}  Sets the validation
for this to be a string, with a minimum length of for and not to exceed 100 chars.  It is
a required field.

"id":{"required":true,"type":"string","max_len":40}  Sets a required field.  It has
a maximum length of 40.

"desc":{"required":false,"type":"string","max_len":400} Sets an optional field.   An empty
string will be used if the field is not supplied.  Maximum length is 400.

"auth_token":{"required":true,"type":"uuid"} Is required for all authorization required requests.
Since "noauth":true has not been specified, this is an authorization required operation.
"auth_token" will be required and validated and the type of the "auth_token" is "uuid" or "u".





## Example 003

``` JSON
	,"/api/test/change_password": { "g": "change_password", "p": [ "password", "again", "token","$ip$" ]
		, "LineNo":"__LINE__"
		, "Method":["POST"]
		, "TableList":["t_user","t_ip_ban"]
		, "valid": {
			 "password": 	{ "required":true, "type":"string", "max_len":80, "min_len":4 }
			,"again":	 	{ "required":true, "type":"string", "max_len":80, "min_len":4 }
			,"token": 		{ "required":true, "type":"string", "max_len":40, "min_len":2 }
			,"$ip$": 		{ "required":true, "type":"string", "max_len":40, "min_len":4 }
			}
		, "CallAfter": ["CacheEUser"]
		}
```

"g" Call stored procedure named "change_password"

"p":["password","again","token","$ip$"]  call stored procedure with "password", "again", "token", and "$ip$".
"$ip$" is an injected value.  It is the IP address of the client making the change.    For all of the
injected values see the secion on parameter injection.

"TableList":["t_user","t_ip_ban"]  This stored procedure access the tables "t_user" and "t_ip_ban".

"LineNo":"__LINE__" indicates the line number for any errors.

"Method":["POST"]  Specifies that this call will only respond to "POST" calls.  

"valid" - all 4 values are required.  Length is specified for all of them.

"CallAfter":["CacheEUser"]  A function named CacheEUser is called in the  GO code after this stored
procedure runs.   See the section on *pre/post function calls* for the parameters and details on
this.




## Exampe 004 - Expose a table for selects and updates.  Require authentation.

``` JSON
	,"/api/table/t_email_track": { "crud": [ "select", "insert", "update", "delete", "info" ]
		, "LineNo":"__LINE__"
		, "p": [ ]
		, "TableName":"t_email_track"
		, "Method":["GET","POST","PUT","DELETE","HEAD"]
		, "cols": [
				  { "colName": "id" 			, "colType":"u"	,"isPk":true, "insert":true									, "autoGen":true					}
				, { "colName": "user_id" 		, "colType":"s"				, "insert":true	, "update":true														}
				, { "colName": "auth_token" 	, "colType":"s"				, "insert":true	, "update":false													}
				, { "colName": "to" 			, "colType":"s"				, "insert":true	, "update":true														}
				, { "colName": "from" 			, "colType":"s"				, "insert":true	, "update":true														}
				, { "colName": "subject" 		, "colType":"s"				, "insert":true	, "update":true														}
				, { "colName": "body" 			, "colType":"s"				, "insert":true	, "update":true														}
				, { "colName": "error" 			, "colType":"s"				, "insert":true	, "update":true														}
				, { "colName": "status" 		, "colType":"s"				, "insert":true	, "update":true														}
				, { "colName": "ip"	 			, "colType":"s"				, "insert":true	, "update":true, "DataColName": "$ip$"								}
			]
		, "validDel": {
				 "id":		 			{ "required":false, "type":"string", "max_len":40, "min_len":2 }
				,"auth_token": 			{ "required":false, "type":"string", "max_len":40, "min_len":2 }
				,"$ip$": 				{ "required":true, "type":"string", "max_len":40, "min_len":4 }
			}
		, "validPost": {
				 "to": 					{ "required":false, "type":"s", "min_len":2, "max_len":250 }
				,"subject": 			{ "required":false, "type":"s", "min_len":2, "max_len":250 }
				,"from": 				{ "required":false, "type":"s", "min_len":2, "max_len":250 }
				,"auth_token": 			{ "required":false, "type":"s", "max_len":40, "min_len":2 }
				,"user_id": 			{ "required":false, "type":"s", "max_len":40, "min_len":2 }
				,"$ip$": 				{ "required":true, "type":"s", "max_len":40, "min_len":4 }
			}
		, "validPut": {
				 "to": 					{ "required":false, "type":"s", "min_len":2, "max_len":250 }
				,"subject": 			{ "required":false, "type":"s", "min_len":2, "max_len":250 }
				,"from": 				{ "required":false, "type":"s", "min_len":2, "max_len":250 }
				,"auth_token": 			{ "required":false, "type":"s", "max_len":40, "min_len":2 }
				,"$ip$": 				{ "required":true, "type":"s", "max_len":40, "min_len":4 }
			}
		, "orderBy": [ { "colName": "4" } ]
		}
```

"/api/table/..." indicates that this is to be handed as a table request.  This means that the select/update/delete/insert
queries will be generated automatically.

"crud" sets the operations that are allowed on this table.

"crud":["select","insert","update","delete","info"]  Allows all operations on this table.  "info" is for HEAD requests.

"Method":["GET","POST","PUT","DELETE","HEAD"]  This sets the methods that are allowed.

"cols" sets the coluns that will be allowed/returned in operations on this table.  The table may have other columns
that are not visible to the outside world.  For example, creation_date, update_date columns are often not returned.

{"colName":"id","colType":"u","isPk":true,"insert":true,"autoGen":true}  Specifies the column named "id".  Its column type
is set with "colType" to "u" - this is a UUID.  It is marked as primary key (or unique key) with the "isPk":true.  "insert":true
says that this field can be inserted.  "autoGen":true says that TabServer2 should generate a UUID for this if it is not
supplied during an insert operaiton. The UUID will be returned.  UUIDs are generated by the TabServer2 server rather than
by the database.  This is so that as little load as possible is placed on the databse.  The *CLIENT* can generate the
UUID/GUID and send it with an insert.  

"update":false on "colName":"auth_token" indicates that this field can ot be updated after it has been created.

{"colName":"ip","colType":"s","insert":true,"update":true,"DataColName":"$ip$"}  The column name in the database is "ip".  The
parameter that is used in the REST call is "$ip$".  This is an injected value for the client IP address.

Validation is performed based on the method of the operation.

"orderBy":[{"colName":"4"}]  On GET/select requests the default sort order is column 4.  



## Exampel 005 -- Expose a table for selects.

This is a good example of what will be generated using the "app-discovery.html" tool.  A no-authentation, select
only configuration.  "app-discovery.html"  allows for quick discovery of tables in the database and generation of
the necessary configuration for them.

``` JSON
	,"/api/table/t_available_test_systems": { "crud": [ "select" ]
		, "TableName": "t_available_test_systems"
		, "LineNo":"__LINE__"
		, "Method":["GET"]
		, "cols": [
				  { "colName": "osNameClass" 			, "colType": "s" 			}
				, { "colName": "browserNameClass" 		, "colType": "s" 			}
				, { "colName": "osMajorClass" 			, "colType": "s" 			}
				, { "colName": "osMinorClass" 			, "colType": "s" 			}
				, { "colName": "browserMajorClass" 		, "colType": "s" 			}
				, { "colName": "browserMinorClass" 		, "colType": "s" 			}
				, { "colName": "browserName" 			, "colType": "s" 			}
				, { "colName": "title"					, "colType": "s" 			}
				, { "colName": "n_clients"	 			, "colType": "s" 			}
				, { "colName": "n_runs"	 				, "colType": "s" 			}
				, { "colName": "useragent_id"	 		, "colType": "s" 	, "NoSort":true		}
				, { "colName": "is_running_now"	 		, "colType": "s" 			}
			]
	}
```

"TableName":"t_available_test_systems"  name of the table being exposed.

"crud":["select"] The only operation allowed.

"Method":["GET"]  The only operation allowed.

"cols" the set of columns that will be retuned in the select.

No validation is specified.  Where clauses on any column in the "cols" are allowed.
Sorting by any of the columns, except "useragent_id", is allowed.



## Exampel 006 -- Validation of int fields.

``` JSON
	,"/api/saveOneNote": { "Fx": "e_save_data_func", "p": [ "url", "top", "left" ]
		, "LineNo":"__LINE__"
		, "valid": {
			 "url": { "required":true, "type":"s" }
			,"top": { "required":false, "type":"i", "min": -4000, "max": 4000 }
			,"left": { "required":false, "type":"i", "min": -4000, "max": 4000 }
			,"auth_token": { "required":true, "type":"u" }
			}
		}
```

"top":{"required":false,"type":"i","min":-4000,"max":4000}
the type is specified to be an integer, with the "type":"i".  Minimum and maximum integer values are set.
"min","max" can also be set for "f"/float data.

"Fx" is the name of a stored procedure to be called.   Data returned from the stored procedure is 
logged and discarded.  You can have a "Query" called after this.  If you need data retuned then use
the "G" instead.



# Exampel 007 -- Confirmation of return values 

``` JSON
	,"/api/table/tblDepartment": { "crud": [ "select", "insert", "update", "delete", "info" ]
		, "TableName": "tblDepartment"
		, "LineNo":"__LINE__"
		, "ReturnGetPKAsHash": true
		, "Method":["GET","POST","PUT","DELETE"]
		, "deleteViaUpdate": { "colType":"i", "colName":"isDeleted", "Absent":"1", "Present":"0" }
		, "CustomerIdPart": { "colType":"u", "colName":"customer_id" }
		, "cols": [
				  { "colName": "id" 				, "colType": "s" 			, "autoGen": true , "isPk": true }
				, { "colName": "name"				, "colType": "s",	"update":true, "insert":true		}
				, { "colName": "description"		, "colType": "s",	"update":true, "insert":true		}
				, { "colName": "isDeleted"			, "colType": "i",	"update":true, "insert":true		}
			]
		, "orderBy": [ { "colName": "2" } ]
		}
```

"ReturnGetPKAsHash":true Sets the return value when the /api/table/tblDepartment/&lt;PK&gt; call is made.   The default
is to return an array 1 long with the data as a has in the array.   Setting this to true just returns the hash
without the array.

"deleteViaUpdate":{"colType":"i","colName":"isDeleted","Absent":"1","Present":"0"}  Sets that deletes are 
handed via performing an update on the row.  The data tyep for the update is an integer, "i".  The column
name is `isDeleted`.   The row is deleted when `isDelted` is set to 1, and not-delted when set to 0.  You can
use strings, booleans and integers for deleted flags.

"CustomerIdPart":{"colType":"u","colName":"customer_id"}  This table is partitioned by customer_id.   This
allows for multi-customer in a singe table data.   The customer id is a UUID/GUID in this case.   If you are
using Oracle the partitioning should be an integer that is generated.  If you are using Microsoft SQL Server
you should not use UUID/GUID for partitioning.  


# Exampel 008 -- Post Join

``` JSON
	,"/api/table/p_cart:GET": { "crud": [ "select" ]
		, "TableName": "p_cart"
		, "LineNo":"__LINE__"
		, "Method":["GET"]
		, "CustomerIdPart": { "colType":"s", "colName":"customer_id" }
		, "cols": [
				  { "colName": "id" 				, "colType": "s" 			, "autoGen": true , "isPk": true }
				, { "colName": "state"				, "colType": "s",	}
				, { "colName": "marked"				, "colType": "s",	}
				, { "colName": "user_id"			, "colType": "s",	}
				, { "colName": "cart_name"			, "colType": "s",	}
				, { "colName": "pagecookie"			, "colType": "s" 	}
				, { "colName": "total"				, "colType": "f" 	}
			]
		, "orderBy": [ { "colName": "1" } ]
		, "PostJoin": [
					{ "ColName": "id", "ColType":"s", "p":[ "id" ], "SetCol": "cartItemsOld"
						, "Query": "SELECT \"p_cart_item\".\"id\" as \"item_id\", \"p_cart_item\".\"product_id\", \"p_cart_item\".\"product_inventory_id\", \"p_cart_item\".\"n_in_cart\", \"p_cart_item\".\"state\", \"p_product\".\"prod_name\", \"p_product\".\"desc\", \"p_product\".\"state\", \"p_product\".\"SKU\" , \"p_cart_item\".\"total\", \"p_cart_item\".\"ex_total\", \"p_cart_item\".\"options\" FROM \"p_cart_item\" as \"p_cart_item\" left join \"p_product\" as \"p_product\" on \"p_cart_item\".\"product_id\" = \"p_product\".\"id\" WHERE \"p_cart_item\".\"cart_id\" = $1 "
					}
					, { "ColName": "id", "ColType":"s", "p":[ "id" ], "SetCol": "imageList"
						, "Query": "SELECT \"p_image_list\".\"id\", \"p_image_list\".\"image_id\", \"p_image_list\".\"seq_no\", \"p_image\".\"file_name\", \"p_image\".\"base_file_name\", \"p_image\".\"h_size\", \"p_image\".\"w_size\", \"p_image\".\"f_size\", \"p_image\".\"img_type\", \"p_cart_item\".\"product_id\" FROM \"p_cart_item\" as \"p_cart_item\" left join \"p_image_list\" as \"p_image_list\" on ( \"p_cart_item\".\"product_id\" = \"p_image_list\".\"fk_id\" ) left join \"p_image\" as \"p_image\" on ( \"p_image_list\".\"image_id\" = \"p_image\".\"id\" ) WHERE \"p_cart_item\".\"cart_id\" = $1 and \"p_image_list\".\"id\" is not null "
					}
					, { "ColName": "id", "ColType":"s", "p":[ "id" ], "SetCol": "cartItems"
						, "Query": "SELECT * from get_cart_items ( $1 )"
					}
				]
		}
	,"/api/table/p_cart": { "crud": [ "insert", "update", "delete" ]
		, "TableName": "p_cart"
		, "LineNo":"__LINE__"
		, "LoginRequired":true
		, "Method":["POST","PUT","DELETE","HEAD"]
		, "CustomerIdPart": { "colType":"s", "colName":"customer_id" }
		, "cols": [
				  { "colName": "id" 				, "colType": "s" 			, "autoGen": true , "isPk": true }
				, { "colName": "state"				, "colType": "s",	"update":true, "insert":true		}
				, { "colName": "marked"				, "colType": "s",	"update":true, "insert":true		}
				, { "colName": "user_id"			, "colType": "s",	"update":true, "insert":true		}
				, { "colName": "cart_name"			, "colType": "s",	"update":true, "insert":true		}
			]
		}
```

# Exampel 009 -- Templates for Queries

``` JSON
	,"/api/table/tblActionPlan:GET": { "crud": [ "select" ]
		, "TableName": "tblActionPlan"
		, "TableList":[ "tblActionPlan", "tblPerson" ]
		, "LineNo":"__LINE__"
		, "ReturnGetPKAsHash": true
		, "Method":["GET"]
		, "deleteViaUpdate": { "colType":"i", "colName":"isDeleted", "ColAlias":"tblActionPlan", "Absent":"1", "Present":"0" }
		, "CustomerIdPart": { "colType":"s", "colName":"customer_id", "ColAlias":"tblActionPlan" }
		, "cols": [
				  { "colName": "id" 					, "colType": "s" 			, "autoGen": true , "isPk": true }
				, { "colName": "cardId" 				, "colType": "s", 	}
				, { "colName": "sequence" 				, "colType": "i", 	}
				, { "colName": "actionPlan" 			, "colType": "s", 	}
				, { "colName": "dateEntered" 			, "colType": "d" 	}
				, { "colName": "targetCompletion" 		, "colType": "d", 	}
				, { "colName": "responsiblePersonId" 	, "colType": "s", 	}
				, { "colName": "notes" 					, "colType": "s", 	}
				, { "colName": "actionCompleted" 		, "colType": "d", 	}
				, { "colName": "isDeleted"	 			, "colType": "i" 	}
				, { "colName": "firstName" 				, "colType": "s" 	}
				, { "colName": "lastName" 				, "colType": "s" 	}
				, { "colName": "email" 					, "colType": "s"	}
				, { "colName": "phone" 					, "colType": "s" 	}
			]
		, "orderBy": [ { "colName": "3" } ]
		, "SetWhereAlias":"tblActionPlan"
		, "SelectPK1Tmpl": " SELECT \"tblActionPlan\".\"id\" ,\"tblActionPlan\".\"cardId\" ,\"tblActionPlan\".\"sequence\" ,\"tblActionPlan\".\"actionPlan\" ,\"tblActionPlan\".\"dateEntered\" ,\"tblActionPlan\".\"targetCompletion\" ,\"tblActionPlan\".\"responsiblePersonId\" ,\"tblActionPlan\".\"responsiblePersonId\" ,\"tblActionPlan\".\"notes\" ,\"tblActionPlan\".\"actionCompleted\" ,\"tblActionPlan\".\"isDeleted\" ,\"tblPerson\".\"firstName\" ,\"tblPerson\".\"lastName\" ,\"tblPerson\".\"email\" ,\"tblPerson\".\"phone\" FROM \"tblActionPlan\" as \"tblActionPlan\" left join \"tblPerson\" as \"tblPerson\" on \"tblActionPlan\".\"responsiblePersonId\" = \"tblPerson\".\"id\" %{where_where%} %{where%} %{order_by_order_by%} %{order_by%} %{limit_limit%} %{limit%} %{offset_offset%} %{offset%}"
		, "SelectTmpl": " SELECT \"tblActionPlan\".\"id\" ,\"tblActionPlan\".\"cardId\" ,\"tblActionPlan\".\"sequence\" ,\"tblActionPlan\".\"actionPlan\" ,\"tblActionPlan\".\"dateEntered\" ,\"tblActionPlan\".\"targetCompletion\" ,\"tblActionPlan\".\"responsiblePersonId\" ,\"tblActionPlan\".\"notes\" ,\"tblActionPlan\".\"actionCompleted\" ,\"tblActionPlan\".\"isDeleted\" ,\"tblPerson\".\"firstName\" ,\"tblPerson\".\"lastName\" ,\"tblPerson\".\"email\" ,\"tblPerson\".\"phone\" FROM \"tblActionPlan\" as \"tblActionPlan\" left join \"tblPerson\" as \"tblPerson\" on \"tblActionPlan\".\"responsiblePersonId\" = \"tblPerson\".\"id\" %{where_where%} %{where%} %{order_by_order_by%} %{order_by%} %{limit_limit%} %{limit%} %{offset_offset%} %{offset%}"
		}
	,"/api/table/tblActionPlan": { "crud": [ "insert", "update", "delete" ]
		, "TableName": "tblActionPlan"
		, "LineNo":"__LINE__"
		, "ReturnGetPKAsHash": true
		, "Method":["POST","PUT","DELETE","HEAD"]
		, "deleteViaUpdate": { "colType":"i", "colName":"isDeleted", "Absent":"1", "Present":"0" }
		, "CustomerIdPart": { "colType":"s", "colName":"customer_id" }
		, "cols": [
				  { "colName": "id" 					, "colType": "s" 			, "autoGen": true , "isPk": true }
				, { "colName": "cardId" 				, "colType": "s", 	"update":true, "insert":true		}
				, { "colName": "sequence" 				, "colType": "i", 	"update":true, "insert":true		}
				, { "colName": "actionPlan" 			, "colType": "s", 	"update":true, "insert":true		}
				, { "colName": "dateEntered" 			, "colType": "d" 	}
				, { "colName": "targetCompletion" 		, "colType": "d", 	"update":true, "insert":true		}
				, { "colName": "responsiblePersonId" 	, "colType": "s", 	"update":true, "insert":true		}
				, { "colName": "notes" 					, "colType": "s", 	"update":true, "insert":true		}
				, { "colName": "actionCompleted" 		, "colType": "d", 	"update":true, "insert":true		}
				, { "colName": "isDeleted"	 			, "colType": "i" 	}
			]
		, "orderBy": [ { "colName": "3" } ]
		}
```

# Exampel 010 -- Using PostgreSQL keyword search facility

``` JSON
	{
	  "note:comment": { "f": "(C) Philip Schlump, 2009-2015." }
	, "note:version": { "f": "v1.0.2" }

		,"/api/table/x_product:GET": { "crud": [ "select" ]
			, "TableName": "x_product"
			, "LineNo":"__FILE__ : __LINE__"
			, "ReturnGetPKAsHash": true
			, "ReturnMeta": true
			, "TableList":[ "x_product", "x_attr", "x_product_inventory", "x_product_options", "x_product_options_meta" ]
			, "Method":["GET"]
			, "valid": {
				 "$customer_id$": { "required":true, "type":"s" }
				,"callback": { }
				}
			, "CustomerIdPart": { "colType":"s", "colName":"customer_id" }
			, "cols": [
					  { "colName": "id" 				, "colType": "s" 			, "autoGen": true , "isPk": true }
					, { "colName": "customer_id"		, "colType": "s" }
					, { "colName": "prod_name"			, "colType": "s" }
					, { "colName": "desc"				, "colType": "s" }
					, { "colName": "cart_tmpl"			, "colType": "s" }
					, { "colName": "state"				, "colType": "s" }
					, { "colName": "limit_per_cust"		, "colType": "s" }
					, { "colName": "min_count_of"		, "colType": "i" }
					, { "colName": "max_count_of"		, "colType": "i" }
					, { "colName": "category_id"		, "colType": "s" }
					, { "colName": "valid_attr_id"		, "colType": "s" }
					, { "colName": "inventory_order"	, "colType": "s" }
					, { "colName": "min_inv_level"		, "colType": "i" }
					, { "colName": "price_model"		, "colType": "s" }
					, { "colName": "price_01"			, "colType": "f" }
					, { "colName": "price_02"			, "colType": "f" }
					, { "colName": "price_03"			, "colType": "f" }
					, { "colName": "price_04"			, "colType": "f" }
					, { "colName": "price"				, "colType": "s" }
					, { "colName": "SKU"				, "colType": "s" }
					, { "colName": "product_type"		, "colType": "s" }
					, { "colName": "is_default"			, "colType": "s" }
					, { "colName": "group"				, "colType": "s" }
					, { "colName": "taxable_item"		, "colType": "s" }
					, { "colName": "start_date"			, "colType": "d" }
					, { "colName": "end_date"			, "colType": "d" }
					, { "colName": "prod_start_date"	, "colType": "d" }
					, { "colName": "prod_end_date"		, "colType": "d" }
				]
			, "orderBy": [ { "colName": "3" } , { "colName": "2" } , { "colName": "4" } ]

			, "SelectPK1Tmpl": "SELECT \"id\", \"customer_id\", \"prod_name\", \"desc\", \"cart_tmpl\", \"state\", \"limit_per_cust\", \"min_count_of\", \"max_count_of\", \"category_id\", \"valid_attr_id\", \"inventory_order\", \"min_inv_level\", \"price_model\", \"price_01\", \"price_02\", \"price_03\", \"price_04\", p_price_for_product_3x ( \"id\", \"price_model\", \"price_01\", \"price_02\", \"price_03\", \"price_04\", \"start_date\", \"end_date\" ) as \"price\", \"SKU\", \"taxable_item\" FROM \"x_product\" %{where_where%} %{where%} %{order_by_order_by%} %{order_by%} %{limit_limit%} %{limit%} %{offset_offset%} %{offset%}"
			, "SelectTmpl": "SELECT \"id\", \"customer_id\", \"prod_name\", \"desc\", \"cart_tmpl\", \"state\", \"limit_per_cust\", \"min_count_of\", \"max_count_of\", \"category_id\", \"valid_attr_id\", \"inventory_order\", \"min_inv_level\", \"price_model\", \"price_01\", \"price_02\", \"price_03\", \"price_04\", p_price_for_product_3x ( \"id\", \"price_model\", \"price_01\", \"price_02\", \"price_03\", \"price_04\", \"start_date\", \"end_date\" ) as \"price\", \"SKU\", \"taxable_item\" FROM \"x_product\" %{where_where%} %{where%} %{order_by_order_by%} %{order_by%} %{limit_limit%} %{limit%} %{offset_offset%} %{offset%}"

			, "key_word_col_name": "key_word"
			, "key_word_list_col": "__keyword__"
			, "key_word_tmpl": " %{kw_col%} @@ plainto_tsquery( %{kw_vals%} ) "

			, "category_col_name": "category"
			, "category_col": "category_id"
			, "category_tmpl": " %{cat_col%} in ( select \"id\" from p_get_children_of ( '%{cat_vals%}'::varchar[] ) ) "

			, "attr_table_name": "x_product"
			, "attr_col": "id"
			, "attr_tmpl":" %{attr_col%} in ( select a1.\"fk_id\" from \"x_attr\" as a1 where a1.\"attr_type\" = '%{attr_type%}' and a1.\"attr_name\" = '%{attr_name%}' and a1.\"%{ref_col%}\" %{attr_op%} %{attr_vals%} )"

			, "PostJoin": [
						{ "ColName": "id", "ColType":"s", "p":[ "id" ], "SetCol": "x_attrs"
							, "Query": "SELECT \"id\", \"attr_type\", \"attr_name\", \"fk_id\", \"val1s\", \"val2s\", \"val1i\", \"val2i\", \"val1f\", \"val2f\", \"val1d\", \"val2d\" FROM \"x_attr\" WHERE \"fk_id\" = $1 "
						}
						, { "ColName": "id", "ColType":"s", "p":[ "id" ], "SetCol": "productInventory"
							, "Query": "SELECT \"id\", \"seq_no\", \"is_countable\", \"count_of\", \"reservation_count_of\", \"location_of\", \"start_date\", \"end_date\", \"weight\", \"box_size_h\", \"box_size_w\", \"box_size_d\", \"SKU\" FROM \"x_product_inventory\" WHERE \"product_id\" = $1 "
						}
						, { "ColName": "id", "ColType":"s", "p":[ "id" ], "SetCol": "imageList"
							, "Query": "SELECT \"x_image_list\".\"id\", \"x_image_list\".\"image_id\", \"x_image_list\".\"seq_no\", \"p_image\".\"file_name\", \"p_image\".\"base_file_name\", \"p_image\".\"h_size\", \"p_image\".\"w_size\", \"p_image\".\"f_size\", \"p_image\".\"img_type\"  FROM \"x_image_list\" as \"x_image_list\" left join \"p_image\" as \"p_image\" on \"x_image_list\".\"image_id\" = \"p_image\".\"id\" WHERE \"x_image_list\".\"fk_id\" = $1 "
						}
						, { "ColName": "id", "ColType":"s", "p":[ "id" ], "SetCol": "optionsList"
							, "Query": "SELECT \"x_product_options\".\"id\" as \"product_options_id\" , \"x_product_options\".\"group\" , \"x_product_options_meta\".\"display_order\" , \"x_product_options\".\"seq_no\" , \"x_product_options\".\"price_01\" , \"x_product_options\".\"price_02\" , \"x_product_options\".\"price_03\" , \"x_product_options\".\"price_04\" , \"x_product_options\".\"start_date\" , \"x_product_options\".\"end_date\" , \"x_product_options\".\"SKU\" , \"x_product_options_meta\".\"id\" as \"product_options_meta_id\" , \"x_product_options_meta\".\"option_type\" , \"x_product_options_meta\".\"required\" , \"x_product_options_meta\".\"count_of_option\" , \"x_product_options_meta\".\"price_model\" , \"p_image\".\"file_name\" , \"p_image\".\"base_file_name\" , \"x_product_options\".\"option_title\" FROM \"x_product_options\" as \"x_product_options\" left join \"x_product_options_meta\" as \"x_product_options_meta\" on ( \"x_product_options_meta\".\"group\" = \"x_product_options\".\"group\" ) left join \"p_image_list\" as \"p_image_list\" on ( \"p_image_list\".\"fk_id\" = \"x_product_options\".\"id\" ) left join \"p_image\" as \"p_image\" on ( \"p_image_list\".\"image_id\" = \"p_image\".\"id\" and \"p_image\".\"img_type\" = 'other' ) WHERE \"x_product_options\".\"product_id\" = $1 ORDER BY 3 asc, 4 asc "
						}
					]
			}
	}
```




