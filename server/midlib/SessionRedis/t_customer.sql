





-- "Copyright (C) Philip Schlump, 2009-2017." 

-- drop TABLE "t_customer" ;

CREATE TABLE "t_customer" (
	  "id"					uuid DEFAULT uuid_generate_v4() not null primary key
	, "name"				uuid not null
	, "ctype"				char varying (10) check ( "ctype" in ( 'regular', 'admin', 'test' ) ) default 'test' not null
	, "config"				text default '{}' not null
);

