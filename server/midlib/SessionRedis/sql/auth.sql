

/*

May Edit Config - 1st real test of User v.s. Admin v.s. Root

*1. Add triggers to tables to deal with updated
*2. Add trigger to t_user to take contents and update "t_user_privs" -- On update, delte, then re-insert
*3. Add checks in S.P. for privs
4. Add code in front end to honor "May Edit Config" and "May Generate Reports"
5. Add "Root" priv for additional screens to do work.  Mod Menus based on Privs.

*/

drop table "t_privs_avail" ;
CREATE TABLE "t_privs_avail" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key	-- 
	, "auth_to"				char varying (80)													-- 
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);

drop table "t_group_of_privs" ;
CREATE TABLE "t_group_of_privs" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key	-- 
	, "group_name"			char varying (250)													-- Name of Group of Privs
	, "auth_to"				text																-- JSON array of privs
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);

-- D.B. Version of Privs - User_id + Set of Privs -- Updated by changes to other tables on user/privs  (Could be implemented as a materialized view?)
drop table "t_user_privs" ;
CREATE TABLE "t_user_privs" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key	-- 
	, "user_id"				char varying (40)													-- FK Join to t_user.id
	, "auth_to_item"		char varying (80)													-- list of privs - normalized to table
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);

-- "t_user":"privs"			-- JSON array of privs

-- Notes:
-- 	1. Changes to "preivs" in t_user - via trigger should update "t_user_privs" to keep consisitent.

insert into "t_privs_avail" ( "auth_to" ) values	
	( 'May Change Password' )
,	( 'May Create Accounts' )
,	( 'May Change Other Users Password' )
,	( 'May Delete Users' )
,	( 'May Generate Reports' )
,	( 'May Edit Config' )
,	( 'Root' )
;

insert into "t_group_of_privs" ( "group_name", "auth_to" ) values
	( 'Root', '{ "privs":[ "Root" ] }' )
,	( 'Admin', '{ "privs":[ "May Change Password", "May Create Accounts", "May Change Other Users Password", "May Delete Users", "May Edit Config" ] }' )
,	( 'User', '{ "privs":[ "May Generate Reports" ] }' )
;









CREATE OR REPLACE function t_user_privs_upd()
RETURNS trigger AS 
$BODY$
BEGIN
	NEW.updated := current_timestamp;
	RETURN NEW;
END
$BODY$
LANGUAGE 'plpgsql';


CREATE TRIGGER t_user_privs_trig
BEFORE update ON "t_user_privs"
FOR EACH ROW
EXECUTE PROCEDURE t_user_privs_upd();




CREATE OR REPLACE function t_privs_avail_upd()
RETURNS trigger AS 
$BODY$
BEGIN
	NEW.updated := current_timestamp;
	RETURN NEW;
END
$BODY$
LANGUAGE 'plpgsql';


CREATE TRIGGER t_privs_avail_trig
BEFORE update ON "t_privs_avail"
FOR EACH ROW
EXECUTE PROCEDURE t_privs_avail_upd();




CREATE OR REPLACE function t_group_of_privs_upd()
RETURNS trigger AS 
$BODY$
BEGIN
	NEW.updated := current_timestamp;
	RETURN NEW;
END
$BODY$
LANGUAGE 'plpgsql';


CREATE TRIGGER t_group_of_privs_trig
BEFORE update ON "t_group_of_privs"
FOR EACH ROW
EXECUTE PROCEDURE t_group_of_privs_upd();




-- delete from "t_output";

CREATE OR REPLACE function t_user_upd()
RETURNS trigger AS 
$BODY$
DECLARE	
	i char varying(80); 
	jo json;
BEGIN
	NEW.updated := current_timestamp;
	-- insert into "t_output" ( "msg" ) values ( 'Top: old='||OLD.privs||' new='||NEW.privs );
	if NEW.privs <> OLD.privs then
		delete from "t_user_privs" where "user_id" = NEW.id;
		-- insert into "t_output" ( "msg" ) values ( 'Delte' );
		jo := NEW.privs::json;
		FOR i IN SELECT * FROM json_array_elements(jo)
		LOOP
			-- insert into "t_output" ( "msg" ) values ( 'Insert' );
			insert into "t_user_privs" ( "user_id", "auth_to_item" ) values ( NEW.id, i );
		END LOOP;
	end if;
	RETURN NEW;
END
$BODY$
LANGUAGE 'plpgsql';

CREATE TRIGGER t_user_trig
BEFORE update ON "t_user"
FOR EACH ROW
EXECUTE PROCEDURE t_user_upd();

--update "t_user" set "privs" = '["abc","def"]' where "username" = 'goofy';
--
--delete from "t_output";
--update "t_user" set "privs" = '["abc","def","ghi"]' where "username" = 'goofy';
--
--select "msg" from "t_output" order by "seq";
--
--select * from "t_user_privs";
--



-- p_username - user to be changed
-- p_user_id -- Auth-User making the change



-- ==============================================================================================================================================================================
--
-- p_username - user getting password changed. (Some User)
-- p_user_id - user_id of the user that is requesting the change (Admin)
--
-- ==============================================================================================================================================================================
CREATE or REPLACE FUNCTION test_change_others_password ( p_username varchar, p_user_id varchar, p_password varchar, p_again varchar,  p_ip_addr varchar ) RETURNS varchar AS $$
DECLARE
	l_data			varchar (400);
	l_id			varchar (40);
	l_username		varchar (40);
	l_privs			varchar (400);
	l_junk			varchar (40);
	l_customer_id	varchar (40);
	l_token			varchar (40);
	l_ctoken		varchar (40);
	l_fail 			boolean;
	l_salt1 		varchar(80);
	l_salt2 		varchar(80);
	l_salt 			varchar(80);
	l_ip_ban 		boolean;
	l_auth_token 	varchar(40);
	l_XSRF_TOKEN 	varchar(40);
BEGIN
	l_fail = false;
	l_data = '{"status":"success"}';

	select "id"
		into l_id
		from "t_user_auth"
		where ( "auth_to_item" = 'May Change Other Users Password' 
		   or "auth_to_item" = 'Root' )
		and "user_id" = p_user_id
		limit 1
		;
	IF NOT FOUND THEN
		l_fail = true;
		l_data = '{"status":"error", "code":"141", "msg":"Not authorized to change other users passwords."}';
	END IF;

	if not l_fail then	
		if p_password != p_again then
			l_fail = true;
			l_data = '{"status":"error", "code":"101", "msg":"Passwords did not match."}';
		end if;
	end if;

	if not l_fail then	
		-- check banned ip
		select 'y' 
			into l_junk
			from "t_ip_ban"
			where "ip" = p_ip_addr
			;

		if not found then
			l_ip_ban = false;
		else
			l_ip_ban = true;
		end if;

		if l_ip_ban then
			l_data = '{ "status":"failed", "code":"009", "msg":"Invalid username or password." }';
			l_fail = true;
		end if;
	end if;
	
	if not l_fail then	
		select "id"
				, "privs"
				, "customer_id"
				, "username"
			into l_id
				, l_privs
				, l_customer_id
				, l_username
			from "t_user"
			where "username" = p_username
			;
		IF NOT FOUND THEN
			l_fail = true;
			l_data = '{"status":"error", "code":"100", "msg":"Invalid Token or Expired Token"}';
		END IF;
	end if;

	if not l_fail then

		select "token"
			into l_token
			from "t_csrf_token"
			order by created desc
			limit 1
			;
		IF NOT FOUND THEN
			l_fail = true;
			l_data = '{"status":"error", "code":"102", "msg":"Token Error 1"}';
		END IF;

		select "token"
			into l_ctoken
			from "t_csrf_token2"
			order by created desc
			limit 1
			;
		IF NOT FOUND THEN
			l_fail = true;
			l_data = '{"status":"error", "code":"103", "msg":"Token Error 2"}';
		END IF;

	end if;

	if not l_fail then	
		l_salt1 = uuid_generate_v4();
		l_salt2 = uuid_generate_v4();
		l_salt = l_salt1 || l_salt2;
		l_auth_token = uuid_generate_v4();
		l_XSRF_TOKEN = uuid_generate_v4();
		update "t_user" set
				  "acct_state" = 'ok'
				, "auth_token" = l_auth_token
				, "ip" = p_ip_addr
				, "last_login" = current_timestamp
				, "n_login_fail" = 0
				, "login_fail_delay" = null
			  , "password" = sha256pw( l_salt||p_password||l_salt )
			  , "salt" = l_salt
			  , "email_reset_key" = null
			  , "email_reset_timeout" = null
			where "username" = p_username
			;
		l_data = '{ "status":"success", "auth_token":'||to_json(l_auth_token)||', "csrf_token":'||to_json(l_token)||', "cookie_csrf_token":'||to_json(l_ctoken)
			||', "XSRF-TOKEN":'||to_json(l_XSRF_TOKEN)
			||', "privs":'||to_json(l_privs)
			||', "user_id":'||to_json(l_id)
			||', "customer_id":'||to_json(l_customer_id)
			||', "username":'||to_json(l_username)
			||'}' ;
	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

--update "t_user"
--	set "email_reset_key" = 'aaa'
--	, "email_reset_timeout" = current_timestamp + interval '1 day'
--	where "id" = 'u0'
--;
--
--select test_change_password ( 'deadbeef', 'deadbeef', 'aaa', '127.0.0.1' );


