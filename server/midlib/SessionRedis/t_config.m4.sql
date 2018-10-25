
-- "Copyright (C) Philip Schlump, 2009-2017." 

drop TABLE "t_config" ;

CREATE TABLE "t_config" (
	  "id"				uuid DEFAULT uuid_generate_v4() not null primary key	-- customer_id
	, "customer_id"		uuid not null 
	, "item_name"		char varying (80) not null 
	, "value"			text 
	, "i_value"			int
	, "updated" 		timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 		timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);

create unique index "t_config_p1" on "t_config" ( "customer_id", "item_name" );

m4_updTrig(t_config)

