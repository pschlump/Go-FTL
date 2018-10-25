BasicAuthPgSql: Basic Auth Using PostgreSQL
===========================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Basic Auth/Postgres"
	,	"SubSectionGroup": "Authentication"
	,	"SubSectionTitle": "Basic Authentication"
	,	"SubSectionTooltip": "Basic Auth implemented with data stored in PostgreSQL"
	, 	"MultiSection":2
	}
```

This middleware implements HTTP basic auth with the authorization stored in PostgreSQL.

The PG package used to access the database is:

	https://github.com/jackc/pgx

Pbkdf2 is used to help prevent cracking via rainbow tables.  Each hashed password
is strengthened by using salt and 5,000 iterations of Pbkdf2 with a sha256 hash.

Basic auth should only be used in conjunction with TLS (https).  If you need to use
an authentication scheme with http, or you want a better authentication scheme,
take a look at the aessrp.go  middleware.  There are examples of using it with
jQuery and AngularJS 1.3 (2.0 of AngularJS coming soon).   

Also this is "basic auth" with the ugly browser popup of username/password and no
real error reporting to the user.  If you want something better switch to the SRP/AES
solution.

Remember that rainbow tables can crack MD5 hashes in less than 30 seconds 95%
of the time.  So... this is only "basic" auth - with low security.

So what is "basic" auth really good for?  Simple answer.  If you need just a
touch of secruity - and no more.   Example:  You took a video of your children
 and you want to send it to Grandma.  It is too big for her email so
you need to send a link.  So quick copy it up to your server and set basic
auth on the directory/path.  Send her the link and the username and password.
This keeps googlebot and other nosy folks out of it - but it is not really
secure.  Then a couple of days later you delete the video.   Works like a
champ!

There is a command line tool in ../../../tools/user-pgsql/user-pgsql.go to maintain the data
in the PostgreSQL database.  You can create/update/delete users from the database.  Also the
tool is useful for verifying that you can connect to the database.

The database connection information is in the global-cfg.json file.

Configuration
-------------

For the paths that you want to protect with this turn on basic auth.  In the server configuration file:

``` JSON
	{ "BasicAuthPgSql": {
		"Paths": [ "/video/children", "/family/pictures" ],
		"Realm": "example.com"
	} },
``` 

SQL Configuration Script

The setup script to create the table in the database is in .../Go-FTL/server/midlib/basicpgsql/user-setup.sql.
You will need to modify this file and run this before using the middleware.  The realm in the "username" field
is "example.com".  That will need to match the realm you are using in your configuration.

``` SQL
	-- drop TABLE "basic_auth" ;
	CREATE TABLE "basic_auth" (
		  "username"				char varying (200) not null primary key
		, "salt"					char varying (100) not null
		, "password"				char varying (180) not null 
	);

	delete from "basic_auth" where "username" = 'example.com:testme';
	insert into "basic_auth" ( "username", "salt", "password" ) values ( 'example.com:testme', 'salt', 
		'9b6095510e3e1c0ea568c3faf29e545c364265d017b16614b1a2de3efe96bc6313cb9e1d221134a46fd5faa8499ebb8568a2ec489e32fa4c4adcd89c05394292'
	);

	\q
``` 
	
### Tested
		
Tested on : Thu Mar 10 16:25:37 MST 2016, Version 0.5.8 of Go-FTL with Version 9.4 of PostgreSQL.

