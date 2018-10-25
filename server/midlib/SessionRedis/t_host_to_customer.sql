





-- "Copyright (C) Philip Schlump, 2009-2017." 

-- drop TABLE "t_host_to_customer" ;

CREATE TABLE "t_host_to_customer" (
	  "id"				uuid DEFAULT uuid_generate_v4() not null primary key	-- customer_id
	, "customer_id"		uuid not null
	, "host_no"			bigint DEFAULT nextval('t_host_id_seq'::regclass) NOT NULL 
	, "host_name"		text not null
	, "is_localhost"	char varying(3) not null default 'no'
	, "updated" 		timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 		timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);

create unique index "t_host_to_customer_u1" on "t_host_to_customer" ( "host_name" );



CREATE OR REPLACE function t_host_to_customer_upd()
RETURNS trigger AS 
$BODY$
BEGIN
  NEW.updated := current_timestamp;
  RETURN NEW;
END
$BODY$
LANGUAGE 'plpgsql';


CREATE TRIGGER t_host_to_customer_trig
BEFORE update ON "t_host_to_customer"
FOR EACH ROW
EXECUTE PROCEDURE t_host_to_customer_upd();



