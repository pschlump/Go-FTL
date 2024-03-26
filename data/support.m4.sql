
m4_changequote(`[[[', `]]]')
m4_include(common.m4.sql)


CREATE TABLE "t_dual" (
	"x"	integer
);
create unique index "t_dual_u1" on "t_dual" ( "x" );
delete from "t_dual";
insert into "t_dual" ( "x" ) values ( 1 );


-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
--
-- for "__after_sync__" synthetic column in queries.
--
-- Add into TabServer2 - /api/list/syncPlul?table=a,b,c,d,e...
--
CREATE TABLE "t_sync_marker" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "note1"				char varying(50) not null
	, "created" 			timestamp default current_timestamp not null 						
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
);

insert into "t_sync_marker" ( "note1" ) values ( 'abc' );

m4_updTrig(t_sync_marker)

--
--	,"/api/sync_marker": { "g": "sync_marker_update", "p": [ "$ip$"]
--		, "LoginRequired":false
--		, "LineNo":"Line: __LINE__ File: __FILE__"
--		, "Method":["GET","POST"]
--		, "TableList":[ "t_sync_marker" ]
--		, "valid": {
--			 "$ip$": 		{ "required":true, "type":"string", "max_len":40, "min_len":4 }
--			}
--		}
--

CREATE or REPLACE FUNCTION sync_marker_update (p_ip_addr varchar)
	RETURNS varchar AS $$
DECLARE
	l_rv			varchar (40);
BEGIN
	update "t_sync_marker"
		set "note1" = p_ip_addr
		;
	RETURN '{"status":"success"}';
END;
$$ LANGUAGE plpgsql;




-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

-- 		theMux.HandleFunc("/api/saveEmailMessage", respHandlerSaveEmailMessage).Methods("GET", "POST")                                     //	insert/update on table

---- drop TABLE "t_email_tab" ;
CREATE TABLE "t_email_tab" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "email_group"			text
	, "person_name"			text
	, "email_addr"			text
	, "message_body"		text
	, "ip_addr"				text
	, "created" 			timestamp default current_timestamp not null 						
);




--	err := sizlib.Run1(db, `select save_email_in_db ( $1, $2, %3, $4, $5 ) as "x"`, group, person_name, email_addr, message_body, IPaddr)
--	,"/api/test/save_customer_config": { "g": "test_save_customer_config($1)"
CREATE or REPLACE FUNCTION save_email_in_db (p_group varchar, p_person_name varchar, p_email_addr varchar, p_message_body varchar, p_ip_addr varchar)
	RETURNS varchar AS $$
DECLARE
	l_rv			varchar (400);
BEGIN
	insert into "t_email_tab" (
		"email_group" , "person_name" , "email_addr" , "message_body" , "ip_addr"
	) values (
		p_group , p_person_name , p_email_addr , p_message_body , p_ip_addr 
	);
	RETURN '{"status":"success"}';
END;
$$ LANGUAGE plpgsql;



-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

-- 		theMux.HandleFunc("/api/grabFeedback", respHandlerGrabFeedback).Methods("GET")                                                     // 	insert/update on table
-- 		theMux.HandleFunc("/api/logit", respHandlerLogIt).Methods("GET", "POST")                                                           // 	insert/update on table

---- drop TABLE "t_log_info" ;
CREATE TABLE "t_log_info" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "info"				text
	, "url"					text
	, "body"				text
	, "method"				char varying(50)
	, "ip_addr"				text
	, "created" 			timestamp default current_timestamp not null 						
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
);




CREATE or REPLACE FUNCTION log_info (p_info varchar, p_url varchar, p_body varchar , p_method varchar , p_ip_addr varchar)
	RETURNS varchar AS $$
DECLARE
	l_rv			varchar (400);
BEGIN
	insert into "t_log_info" (
		"info", "url", "body", "method", "ip_addr"
	) values (
		p_info, p_url, p_body , p_method , p_ip_addr 
	);
	RETURN '{"status":"success"}';
END;
$$ LANGUAGE plpgsql;


m4_updTrig(t_log_info)

-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

-- 		theMux.HandleFunc("/api/serverVersion", respHandlerServerVersion).Methods("GET")                                                   //	select from table in stored procedure + init.


CREATE SEQUENCE t_version_seq
  INCREMENT 1
  MINVALUE 1
  MAXVALUE 9223372036854775807
  START 1
  CACHE 1;

-- drop TABLE "t_version" ;
CREATE TABLE "t_version" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "info"				text
	, "short"				char varying (20)
	, "seq"					integer DEFAULT nextval('t_version_seq')
	, "created" 			timestamp default current_timestamp not null 						
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
);


CREATE or REPLACE FUNCTION server_version (p_ip_addr varchar)
	RETURNS varchar AS $$
DECLARE
	l_rv			varchar (4000);
BEGIN
	select info
       	into l_rv
		from ( select max("seq") max_seq from "t_version") as t1
		, "t_version"
		where "seq" = t1.max_seq
		limit 1
	;
	RETURN l_rv;
END;
$$ LANGUAGE plpgsql;

-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

-- 		theMux.HandleFunc("/api/registerEmail", respHandlerRegEmail).Methods("GET", "POST")                                                //	insert/update on table
-- 		theMux.HandleFunc("/api/deRegisterEmail", respHandlerDeRegEmail).Methods("GET", "POST")                                            //	update on table



-- drop TABLE "t_email_list" ;
CREATE TABLE "t_email_list" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "to_addr"				text
	, "confirmed"			char (1) default 'n' not null
	, "de_reg"				char (1) default 'n' not null
	, "ip_addr"				text
	, "created" 			timestamp default current_timestamp not null 						
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
);

m4_updTrig(t_email_list)

CREATE or REPLACE FUNCTION reg_email_list ( p_to_addr varchar, p_ip_addr varchar )
	RETURNS varchar AS $$
DECLARE
	l_rv			varchar (400);
	l_id			varchar (40);
	l_de_reg		varchar (1);
	l_conf			varchar (1);
	l_done			varchar (1);
BEGIN
	l_rv = '{"status":"success"}';
	l_done = 'n';

	select "id", "de_reg", "confirmed"
		into l_id, l_de_reg, l_conf
		from "t_email_list"
		where "to_addr" = p_to_addr
		;
	IF FOUND THEN
		if l_de_reg = 'y' then
			update "t_email_list"
				set "de_reg" = 'n', "ip_addr" = p_ip_addr
				where "id" = l_id
					;
		elsif l_conf = 'n' then
			-- xyzzy - new confirm should be sent -
			l_rv = '{"status":"error","msg":"please confirm email","send_confirm":"y"}';
		else
			l_rv = '{"status":"error","msg":"already registered"}';
		end if;
		l_done = 'y';
	END IF;

	if l_done = 'n' then
		insert into "t_email_list" ( "to_addr", "ip_addr" ) values ( p_to_addr, p_ip_addr );
		l_rv = '{"status":"success","send_confirm":"y"}';
		l_done = 'y';
	end if;

	RETURN l_rv;
END;
$$ LANGUAGE plpgsql;

CREATE or REPLACE FUNCTION dereg_email_list ( p_to_addr varchar, p_ip_addr varchar )
	RETURNS varchar AS $$
BEGIN
	update "t_email_list"
		set "de_reg" = 'y', "ip_addr" = p_ip_addr
		where "to_addr" = p_to_addr
		;
	RETURN '{"status":"success"}';
END;
$$ LANGUAGE plpgsql;

CREATE or REPLACE FUNCTION confirm_email_list ( p_to_addr varchar, p_ip_addr varchar )
	RETURNS varchar AS $$
BEGIN
	update "t_email_list"
		set "de_reg" = 'n', "confirmed" = 'y', "ip_addr" = p_ip_addr
		where "to_addr" = p_to_addr
		;
	RETURN '{"status":"success"}';
END;
$$ LANGUAGE plpgsql;








-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
-- For data duplication to remote system.
-- API not complete yet.

-- drop TABLE "t_pull_cfg" ;
CREATE TABLE "t_pull_cfg" (
	  "id"				char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "id_key"			text
	, "legit_ip"		text
	, "pull_by" 		timestamp 
);

---- drop TABLE "t_pull_data" ;
CREATE TABLE "t_pull_data" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "table_name"			char varying (50)
	, "pull_at" 			timestamp 
);



