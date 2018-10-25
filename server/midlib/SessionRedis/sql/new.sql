
-- ToDo
-- 1. Create a web-page as a part of who-cares that edits the set of monitored itesm
-- 2. Fix message on monitored items.  If "green" - then "ok" or "", if red then display message.
-- 3. Get messages from the config table - add this. -- Only use default if not specified.



-- ================================================================================================================================================================================================
--
--
-- ================================================================================================================================================================================================
--CREATE or REPLACE FUNCTION test_login_setup()
--	RETURNS varchar AS $$
--DECLARE
--	l_salt1 varchar(80);
--	l_salt2 varchar(80);
--	l_salt varchar(80);
--	l_password varchar(80);
--	data1 record;
--BEGIN
--	FOR data1 IN
--		select
--			  "id"
--		from "t_user"
--	LOOP
--		l_salt1 = uuid_generate_v4();
--		l_salt2 = uuid_generate_v4();
--		l_salt = l_salt1 || l_salt2;
--		l_password = sha256pw ( l_salt||'deadbeef'||l_salt );
--		update "t_user"
--			set "password" = l_password
--				, "salt" = l_salt
--			where "id" = data1."id"
--		;
--	END LOOP;
--	RETURN 'ok';
--END;
--$$ LANGUAGE plpgsql;
--
--select test_login_setup();

--drop FUNCTION test_login_setup();

drop FUNCTION test_login(p_username varchar, p_password varchar, p_ip_addr varchar, p_csrf_token varchar);
drop FUNCTION test_logout(p_auth_token varchar, p_ip_addr varchar);
drop FUNCTION test_register_new_user(p_username varchar, p_password varchar, p_ip_addr varchar, p_email varchar, p_real_name varchar, p_url varchar, p_csrf_token varchar);
drop FUNCTION test_monitor_it_happened(p_item_name varchar);
drop FUNCTION prep_info2 ( p_user_id varchar );
drop FUNCTION test_confirm_email ( p_auth_token varchar, p_ip_addr varchar );
drop FUNCTION test_change_password ( p_password varchar, p_again varchar, p_token varchar, p_ip_addr varchar );
drop FUNCTION test_password_reset ( p_username varchar, p_auth_token varchar, p_email varchar, p_ip_addr varchar, p_url varchar, p_top varchar );
drop FUNCTION status_db ( p_ip_addr varchar );


-- ================================================================================================================================================================================================
--
--
-- ================================================================================================================================================================================================
drop table "t_ip_ban" ;
CREATE TABLE "t_ip_ban" (
	  "ip"					char varying (40) not null primary key
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);
insert into "t_ip_ban" ( "ip" ) values ( '1.1.1.2' );

drop table "t_csrf_token" ;
CREATE TABLE "t_csrf_token" (
	  "token"				char varying (40) not null primary key
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);
insert into "t_csrf_token" ( "token" ) values ( '42' );
insert into "t_csrf_token" ( "token" ) values ( '44' );

drop table "t_csrf_token2" ;
CREATE TABLE "t_csrf_token2" (
	  "token"				char varying (40) not null primary key
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);
insert into "t_csrf_token2" ( "token" ) values ( '5544' );

drop TABLE "t_auth_token" ;
CREATE TABLE "t_auth_token" (
	  "auth_token"			char varying (40) not null primary key
	, "user_id"				char varying (40)
	, "expire"	 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);
create index "t_auth_token_p1" on "t_auth_token" ( "user_id" )

-- ================================================================================================================================================================================================
--
--
-- ================================================================================================================================================================================================

-- alter table "t_user" add column "customer_id"		char varying (40) default '1' ;

-- delete from "t_output";
-- drop FUNCTION test_login(p_username varchar, p_password varchar, p_ip_addr varchar, p_csrf_token varchar);

CREATE or REPLACE FUNCTION test_login(p_username varchar, p_password varchar, p_ip_addr varchar, p_csrf_token varchar, p_host varchar)
	RETURNS varchar AS $$
DECLARE
    l_id 				varchar (40);
	l_password			varchar (80);
	l_salt				varchar (400);
  	l_auth_token		varchar (40);
  	l_ip				varchar (40);
	l_email_confirmed	varchar (1);
	l_acct_state 		varchar (10);
	l_acct_expire		timestamp;
	l_n_login_fail		int;
	l_login_fail_delay	timestamp;
	l_last_login 		timestamp ;		 													-- 
    l_customer_id 		varchar (40);

    l_data 		varchar (8000);
    l_token 	varchar (40);
    l_ctoken 	varchar (40);
	l_fail 		boolean;
	l_junk		varchar (1);
	l_ip_ban 	boolean;
	l_bad_token	boolean;
    l_seq 		varchar (40);
    l_privs		varchar (400);
	l_XSRF_TOKEN varchar(40);
	l_config	varchar (7500);
BEGIN
	l_fail = false;
	l_id = null;
	l_data = '{ "status":"unknown"}';

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

--	if not l_fail then
--		select "token"
--			into l_token
--			from "t_csrf_token"
--			where "token" = p_csrf_token
--			;
--
--		if not found then
--			l_bad_token = true;
--		else
--			l_bad_token = false;
--
--			select "token"
--				into l_token
--				from "t_csrf_token"
--				order by created desc
--				limit 1
--				;
--
--			select "token"
--				into l_ctoken
--				from "t_csrf_token2"
--				order by created desc
--				limit 1
--				;
--
--		end if;
--
--		if l_bad_token then
--			l_data = '{ "status":"failed", "code":"010", "msg":"Invalid username or password." }';
--			l_fail = true;
--		end if;
--	end if;

	if not l_fail then
		select  "t_user"."id"
			, "t_user"."password"
			, "t_user"."salt"
			, "t_user"."auth_token"
			, "t_user"."ip"
			, "t_user"."email_confirmed"
			, "t_user"."acct_state"
			, "t_user"."acct_expire"
			, "t_user"."n_login_fail"
			, "t_user"."login_fail_delay"
			, "t_user"."last_login"
			, "t_user"."privs"
			, "t_user"."customer_id"
			, "t_customer"."config"
		into  l_id
			, l_password		
			, l_salt		
			, l_auth_token
			, l_ip		
			, l_email_confirmed	
			, l_acct_state 	
			, l_acct_expire	
			, l_n_login_fail		
			, l_login_fail_delay
			, l_last_login 	
			, l_privs 	
			, l_customer_id 	
			, l_config
			from "t_user" as "t_user" left join "t_customer" as "t_customer" on "t_customer"."id" = "t_user"."customer_id"
			where "username" = p_username
			;

		if not found then
			l_data = '{ "status":"failed", "code":"000", "msg":"Invalid username or password." }';
			l_fail = true;
		end if;
	end if;

	if l_privs is null then
		l_privs = '';
	end if;

	--if not l_fail then
	--	insert into "t_output" ( msg ) values ( '148: got a user with that name.' );
	--end if;

	if not l_fail then
		if l_email_confirmed = 'n' then
			l_data = '{ "status":"failed", "code":"001", "msg":"Before you login you have to confirm your email account.." }';
			l_fail = true;
		end if;
	end if;

	if not l_fail then
		if l_acct_state = 'locked' or l_acct_state = 'billing' or l_acct_state = 'closed' then
			l_data = '{ "status":"failed", "code":"002", "msg":"Account is no longer valid.", "acct_state":"'||l_acct_state||'" }';
			l_fail = true;
		end if;
	end if;

	if not l_fail then
		if l_acct_state = 'pass-reset' then
			l_data = '{ "status":"failed", "code":"012", "msg":"You must reset your password before you can login." }';
			l_fail = true;
		end if;
	end if;

	if not l_fail then
		if l_n_login_fail > 5 then
			if l_login_fail_delay + interval ' 120 seconds ' < now() then
				l_fail = false;
			else
				l_acct_state = 'temporary';
				l_data = '{ "status":"failed", "code":"003", "msg":"Too many failed login attempts.  Please wate 120 seconds and try again." }';
				l_fail = true;
			end if;
		end if;
	end if;

	if not l_fail then
		if l_acct_expire < now() then
			l_data = '{ "status":"failed", "code":"005", "msg":"Account is no longer valid.  Your trial period has ended." }';
			l_fail = true;
		end if;
	end if;
			
	if not l_fail then
		if l_password != sha256pw ( l_salt||p_password||l_salt ) then
			l_data = '{ "status":"failed", "code":"008", "msg":"Invalid username or password." }';
			l_fail = true;
		end if;
	end if;

	--if not l_fail then
	--	insert into "t_output" ( msg ) values ( '199: good at the end.' );
	--end if;

	if not l_fail then
		l_auth_token = uuid_generate_v4();
		l_seq = uuid_generate_v4();
		l_XSRF_TOKEN = uuid_generate_v4();
		update "t_user" set
				  "acct_state" = l_acct_state
				, "auth_token" = l_auth_token
				, "ip" = p_ip_addr
				, "last_login" = current_timestamp
				, "n_login_fail" = 0
				, "login_fail_delay" = null
			where "id" = l_id
			;
		insert into "t_auth_token" ( "auth_token", "user_id" ) values ( l_auth_token, l_id );		-- only used for debuging purposes - remove later
		l_data = '{ "status":"success", "auth_token":'||to_json(l_auth_token)||', "csrf_token":'||to_json(l_token)||', "cookie_csrf_token":'||to_json(l_ctoken)
			||', "seq":'||to_json(l_seq)
			||', "XSRF-TOKEN":'||to_json(l_XSRF_TOKEN)
			||', "privs":'||to_json(l_privs)
			||', "user_id":'||to_json(l_id)
			||', "customer_id":'||to_json(l_customer_id)
			||', "config":'||to_json(l_config)
			||'}' ;
		--insert into "t_output" ( msg ) values ( '222: l_data='||l_data );
	else
		if l_id is not null then
			update "t_user" set "acct_state" = l_acct_state, "auth_token" = '*', "ip" = p_ip_addr, "login_fail_delay" = current_timestamp, "n_login_fail" = "n_login_fail" + 1 where "id" = l_id;
		end if;
	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;




-- ================================================================================================================================================================================================
--
--
-- ================================================================================================================================================================================================
CREATE or REPLACE FUNCTION test_csrf_tokens(p_ip_addr varchar, p_csrf_token varchar, p_cookie_csrf_token varchar)
	RETURNS varchar AS $$
DECLARE
    l_id 		varchar (40);
  	l_ip		varchar (40);

    l_data 		varchar (400);
    l_token 	varchar (40);
    l_ctoken 	varchar (40);
	l_fail 		boolean;
	l_junk		varchar (1);
	l_ip_ban 	boolean;
	l_bad_token	boolean;
BEGIN
	l_fail = false;
	l_id = null;

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
		l_data = '{ "status":"failed", "code":"009", "msg":"Invalid." }';
		l_fail = true;
	end if;

	if not l_fail then
		select "token"
			into l_token2
			from "t_csrf_token2"
			where "token" = p_cookie_csrf_token
			;

		if not found then
			l_bad_token = true;
		else
			l_bad_token = false;
		end if;

		if l_bad_token then
			l_data = '{ "status":"failed", "code":"014", "msg":"Invalid." }';
			l_fail = true;
		end if;
	end if;

	if not l_fail then
		select "token"
			into l_token
			from "t_csrf_token"
			where "token" = p_csrf_token
			;

		if not found then
			l_bad_token = true;
		else
			l_bad_token = false;

			select "token"
				into l_token
				from "t_csrf_token"
				order by created desc
				limit 1
				;

			select "token"
				into l_ctoken
				from "t_csrf_token2"
				order by created desc
				limit 1
				;

		end if;

		if l_bad_token then
			l_data = '{ "status":"failed", "code":"010", "msg":"Invalid." }';
			l_fail = true;
		end if;
	end if;

	if not l_fail then
		l_data = '{ "status":"success", "csrf_token":'||to_json(l_token)||', "cookie_csrf_token":'||to_json(l_ctoken)||'}';
	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;



-- select test_login('goofy', 'deadbeef', '1.1.1.1', '42');
-- select test_login('goofy', 'dXadbeef', '1.1.1.1', '42');
-- select test_login('goofy', 'dXadbeef', '1.1.1.2', '42');
-- select test_login('goofy', 'deadbeef', '1.1.1.1', '41');
-- 
-- update "t_user" 
-- 	set "email_confirmed" = 'n'
-- 	where "username" = 'uui'
-- ;
-- update "t_user" 
-- 	set "acct_state" = 'locked'
-- 	where "username" = 'ik5'
-- ;
-- update "t_user" 
-- 	set "acct_state" = 'pass-reset'
-- 	where "username" = 'ik5k'
-- ;
-- update "t_user" 
-- 	set "acct_expire" = current_timestamp - interval ' 1 second '
-- 	where "username" = 'aaaa'
-- ;
-- update "t_user" 
-- 	set "acct_expire" = current_timestamp + interval ' 10 second '
-- 	where "username" = 'uu1'
-- ;
-- 
-- 
-- select test_login('uui',  'deadbeef', '1.1.1.1', '42');
-- select test_login('ik5',  'deadbeef', '1.1.1.1', '42');
-- select test_login('ik5k', 'deadbeef', '1.1.1.1', '42');
-- select test_login('aaaa', 'deadbeef', '1.1.1.1', '42');
-- select test_login('uu1',  'deadbeef', '1.1.1.1', '42');
-- 
-- --		if l_n_login_fail > 5 then
-- select test_login('goofy', 'deadbeef', '1.1.1.1', '42');
-- select test_login('goofy', 'dXadbeef', '1.1.1.1', '42');
-- select test_login('goofy', 'dXadbeef', '1.1.1.1', '42');
-- select test_login('goofy', 'dXadbeef', '1.1.1.1', '42');
-- select test_login('goofy', 'dXadbeef', '1.1.1.1', '42');
-- select test_login('goofy', 'dXadbeef', '1.1.1.1', '42');
-- select test_login('goofy', 'dXadbeef', '1.1.1.1', '42');
-- select "login_fail" from "t_user" where "username" = 'goofy';
-- 
-- ---- Sleep for 3 minutes then Try ---
-- --		if l_n_login_fail > 5 then
-- select test_login('goofy', 'deadbeef', '1.1.1.1', '42');
-- -- select test_login('goofy', 'dXadbeef', '1.1.1.1', '42');
-- -- select test_login('goofy', 'dXadbeef', '1.1.1.1', '42');
-- -- select test_login('goofy', 'dXadbeef', '1.1.1.1', '42');
-- -- select test_login('goofy', 'dXadbeef', '1.1.1.1', '42');
-- -- select test_login('goofy', 'dXadbeef', '1.1.1.1', '42');
-- -- select test_login('goofy', 'dXadbeef', '1.1.1.1', '42');
-- select "n_login_fail" from "t_user" where "username" = 'goofy';
-- 






-- ================================================================================================================================================================================================
--
--
-- ================================================================================================================================================================================================
CREATE or REPLACE FUNCTION test_logout(p_auth_token varchar, p_ip_addr varchar)
RETURNS varchar AS $$
DECLARE
	l_auth_token 		varchar (40);
BEGIN
	l_auth_token = uuid_generate_v4();
	update "t_user" set "auth_token" = l_auth_token, "ip" = p_ip_addr, "n_login_fail" = 0, "login_fail_delay" = null where "auth_token" = p_auth_token;
	RETURN '{"status":"success"}';
END;
$$ LANGUAGE plpgsql;





-- ================================================================================================================================================================================================
--
--
-- ================================================================================================================================================================================================

-- xyzzy - need to add "site=" and "app=" to URL!

drop FUNCTION test_register_new_user(p_username varchar, p_password varchar, p_ip_addr varchar, p_email varchar, p_real_name varchar, p_url varchar, p_csrf_token varchar);

CREATE or REPLACE FUNCTION test_register_new_user(p_username varchar, p_password varchar, p_ip_addr varchar, p_email varchar, p_real_name varchar, p_url varchar, p_csrf_token varchar, p_app varchar, p_name varchar)
	RETURNS varchar AS $$

DECLARE
    l_id 				varchar (40);
	l_auth_token 		varchar (40);
	l_email_token 		varchar (40);
	l_user_id 			varchar (40);
	l_data				varchar (400);
	l_group_id			varchar (40);
	l_fail				bool;
	l_token				varchar (40);
	l_bad_token			bool;

	l_salt1 			varchar(80);
	l_salt2 			varchar(80);
	l_salt 				varchar(80);
	l_password 			varchar(80);

	l_subject			varchar (100);
	l_body				varchar (1000);

BEGIN
	l_fail = false;
	l_data = '{"status":"success"}';

	l_group_id = '2';							-- xyzzy

	l_id = uuid_generate_v4();
	l_auth_token = uuid_generate_v4();
	l_user_id = uuid_generate_v4();
	l_email_token = uuid_generate_v4();

	l_salt1 = uuid_generate_v4();
	l_salt2 = uuid_generate_v4();
	l_salt = l_salt1 || l_salt2;
	l_password = sha256pw ( l_salt||p_password||l_salt );

	select "token"
		into l_token
		from "t_csrf_token"
		where "token" = p_csrf_token
		;

	if not found then
		l_bad_token = true;
	else
		l_bad_token = false;
	end if;

	if l_bad_token then
		l_data = '{ "status":"failed", "code":"010", "msg":"Invalid csrf token." }';
		l_fail = true;
	end if;

	if not l_fail then
		BEGIN
			insert into "t_user" ( "id", "group_id", "username", "password", "auth_token", "ip", "real_name", "email_address", "acct_state", "acct_expire", "email_confirmed", "salt", "email_reset_key" ) 
				values ( l_id, l_group_id, p_username, l_password, l_auth_token, p_ip_addr, p_real_name, p_email, 'temporary', current_timestamp + interval '30 days', 'n', l_salt, l_email_token )
			;
		EXCEPTION WHEN unique_violation THEN
			l_fail = true;
			l_data = '{"status":"error","msg":"Unable to create user with this username.  Please choose a different username (try your email address).","code":"020"}';
			-- Report Error
		END;
	end if;

	if not l_fail then

		-- l_name = regexp_replace(encode(p_name,'hex'),'(..)',E'%\\1','g');
		-- select regexp_replace(encode('héllo there','hex'),'(..)',E'%\\1','g');

		l_subject = 'Welcome!  Click below to complete registration.';
		l_body = 
			  'Hello, '||p_real_name||'<br>'
			||'<br>'
			||'Please follow the link below or cut and paste it into a browser to complete yor registration and validate your email address.<br>'
			||'<br>'
			||'         <a href="'||p_url||'/confirm-email.html?token='||l_email_token||'&app='||p_app||'&name='||p_name||'"> '||p_url||'/confirm-email.html?token='||l_email_token||'&app='||p_app||'&name='||p_name||' </a><br>'
			||'<br>'
			||'A temporary password has been created for you.   You can change it under the configuration menu.<br>'
			||'Your temporary account is good for 30 days.  Welcome!<br>'
			||'<br>'
			;

		BEGIN
			insert into "t_email_q" ( "user_id", "ip", "auth_token", "to", "from", "subject", "body", "status" )
				values ( l_id, p_ip_addr, l_email_token, p_email, 'registration@2c-why.com', l_subject, l_body, 'pending' )
			;
		EXCEPTION WHEN others THEN
			l_fail = true;
			l_data = '{"status":"error","msg":"This is really sad.  Our email system is broken.  Give us a call a 720-209-7888.","code":"021"}';
			-- Report Error
		END;
	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

-- delete from "t_email_q";
-- delete from "t_user" where "username" = 't100';
--
-- select test_register_new_user(
-- 	't100',
-- 	'fredredfred',
-- 	'127.0.0.1',
-- 	'pschlump@gmail.com',
-- 	't100',
-- 	'http://localhost:8090',
-- 	'42'
-- ) as "x";

-- -- Newer
--

-- [ "username", "password", "$ip$", "email", "real_name", "$url$", "csrf_token", "site", "name" ], "nokey":true

-- select test_register_new_user('goofy2','dogmeat','127.0.0.1','pschlump@gmail.com','goofy person', 'url', '42', 'image.html', ' image manager ' ) as "x";


-- select * from "t_email_q";
-- delete from "t_email_q";

--update "t_user"
--	set "acct_state" = 'ok'
--		, "email_confirmed" = 'y'
--	where "email_reset_key" in ( select "auth_token" from "t_email_q" )
--;


-- =====================================================================================================================================================================
-- Depricated!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
-- =====================================================================================================================================================================
--CREATE or REPLACE FUNCTION test_monitor_it_happened(p_item_name varchar)
--	RETURNS varchar AS $$
--
--DECLARE
--    l_id 				varchar (40);
--	l_data				varchar (400);
--begin
--	l_data = '{"status":"success"}';
--
--	select 'found' 
--		into l_id
--		from "t_monitor_stuff" 
--		where "item_name" = p_item_name
--	;
--		
--	if not found then
--		insert into "t_monitor_stuff" ( "item_name", "event_to_raise", "delta_t", "timeout_event", "note" )
--			values ( p_item_name, 'none', '100 years', current_timestamp, '$generated automatically$' );
--	end if;
--
--	update "t_monitor_stuff" 
--		set "timeout_event" = current_timestamp + CAST("delta_t" as Interval) 
--		where "item_name" = p_item_name
--	;
--
--	RETURN l_data;
--END;
--$$ LANGUAGE plpgsql;
--
-- select test_monitor_it_happened('test bob');
-- select * from "t_monitor_stuff" ;
-- -- wait for 10 sec
-- select test_monitor_it_happened('test bob');
-- select * from "t_monitor_stuff" ;
-- -- Verify that event time has increasd.
-- delete from "t_monitor_stuff" where "item_name" = 'test bob';










-- =====================================================================================================================================================================
--
--
-- =====================================================================================================================================================================
CREATE or REPLACE FUNCTION prep_info2 ( p_user_id varchar ) RETURNS varchar AS $$
DECLARE
    work_id 	char varying(40);
	rec 		record;
	l_sev 		bigint;
	l_cssStat 	char varying(80);
	l_name 		char varying(80);
	l_info 		char varying(2000);
    l_junk 		char varying(40);
BEGIN

	work_id = uuid_generate_v4();

	FOR rec IN
		select "item_name", "event_to_raise", "delta_t", 'error' as "status"
			from "t_monitor_stuff"
			where "timeout_event" < now()
			  and "enabled" = 'y'
		union
		select "item_name", "event_to_raise", "delta_t", 'ok' as "status"
			from "t_monitor_stuff"
			where "timeout_event" >= now()
			  and "enabled" = 'y'
	LOOP
		l_sev = 0;
		l_cssStat = 'itemNormal';
		l_name = rec.item_name;
		l_info = 'On Time: '||rec.event_to_raise;
		l_info = '';
		if ( rec.status = 'error' ) then
			l_sev = l_sev + 1;
			l_cssStat = 'itemError';
			l_info = 'Missed Deadline: '||rec.event_to_raise;
		end if;

		insert into "t_monitor_results" ( "id", "sev", "cssStat", "name", "info" )
			values ( work_id, l_sev, l_cssStat, l_name, l_info );

	END LOOP;

	RETURN work_id;
END;
$$ LANGUAGE plpgsql;





-- =====================================================================================================================================================================
--
--
-- =====================================================================================================================================================================

drop type "prep_info3rv" cascade ;
CREATE TYPE "prep_info3rv" as (
	 "id"       character varying(40)      
	,"seq"      bigint                   
	,"sev"      bigint                  
	,"cssStat"  character varying(80)  
	,"name"     character varying(80) 
	,"info"     text                 
	,"updated"  timestamp 
	,"created"  timestamp 
);

CREATE or REPLACE function prep_info5( ) returns setof "prep_info3rv" as
$$
declare
	r "prep_info3rv"%rowtype;
	l_it_work_id	varchar(40);
begin
	select prep_info2 ( '1' ) into l_it_work_id; 												-- process new data
	delete from "t_monitor_results" where "created" < current_timestamp - interval '8 days'; 	-- clean up old data

	for r in select "id" ,"seq" ,"sev" ,"cssStat" ,"name" ,"info" ,"updated" ,"created"
			from "t_monitor_results"
			where "id" = l_it_work_id
			order by "sev" desc, "seq" asc 
	loop
		return next r;
	end loop;
	return;
end;
$$ LANGUAGE plpgsql;



-- ================================================================================================================================================================================================
--
--
-- ================================================================================================================================================================================================
CREATE or REPLACE FUNCTION test_confirm_email ( p_auth_token varchar, p_ip_addr varchar ) RETURNS varchar AS $$
DECLARE
	l_data			varchar (400);
	l_junk			varchar (40);
	l_fail 			boolean;
	l_ip_ban 		boolean;
BEGIN
	l_fail = false;
	l_data = '{"status":"success"}';

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
	
			-- where "auth_token" = p_auth_token
	if not l_fail then	
		select '1'
			into l_junk
			from "t_user"
			where "email_reset_key" = p_auth_token
			;
		IF NOT FOUND THEN
			l_fail = true;
			l_data = '{"status":"error","msg":"Invalid Token or Expired Token"}';
		END IF;
	end if;
	
	if not l_fail then	
		update "t_user"
			set "email_confirmed" = 'y'
				, "ip" = p_ip_addr
				, "email_reset_key" = null
			where "email_reset_key" = p_auth_token
			;
	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;


--select test_confirm_email ( "auth_token", '1.1.1.1' )
--	from "t_email_q"




-- ================================================================================================================================================================================================
--
--
-- ================================================================================================================================================================================================
CREATE or REPLACE FUNCTION test_change_password ( p_password varchar, p_again varchar, p_token varchar, p_ip_addr varchar ) RETURNS varchar AS $$
DECLARE
	l_data			varchar (400);
	l_username		varchar (40);
	l_privs			varchar (400);
	l_junk			varchar (40);
	l_id			varchar (40);
	l_customer_id	varchar (40);
	l_token					varchar (40);
	l_ctoken				varchar (40);
		-- |', "csrf_token":'||to_json(l_token)||', "cookie_csrf_token":'||to_json(l_ctoken)
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

	if p_password != p_again then
		l_fail = true;
		l_data = '{"status":"error", "code":"101", "msg":"Passwords did not match."}';
	end if;

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
			where "email_reset_key" = p_token
			  and "email_reset_timeout" > current_timestamp 
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

	
--xyzzy - call caching like login
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
			where "email_reset_key" = p_token
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



-- ================================================================================================================================================================================================
--
--
-- ================================================================================================================================================================================================
CREATE or REPLACE FUNCTION test_password_reset ( p_username varchar, p_auth_token varchar, p_email varchar, p_ip_addr varchar, p_url varchar, p_top varchar ) RETURNS varchar AS $$
DECLARE
	l_data			varchar (400);
	l_id			varchar (40);
	l_token			varchar (40);
	l_fail 			boolean;
	l_from 			varchar (100);
	l_to 			varchar (100);
	l_real_name		varchar (100);
	l_subject 		varchar (100);
	l_body 			varchar (400);
	l_ip_ban 		boolean;
	l_junk			varchar (10);
BEGIN
	l_fail = false;
	l_data = '{"status":"success"}';
	l_token = uuid_generate_v4();
	l_from = 'noreply@2c-why.com';
	l_subject = 'Request to reset paswsword for 2c-why.com';

	-- Check csrf_token

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

	-- ---------------------------------------------------------------------------------------
	-- Fetch Data	
	-- ---------------------------------------------------------------------------------------
	if not l_fail then	
		select "real_name", "id", "email_address"
			into l_real_name, l_id, l_to
			from "t_user"
			where "username" = p_username
			   or "email_address" = p_email
		 	   or "auth_token" = p_auth_token
			;
		IF NOT FOUND THEN
			l_fail = true;
			l_data = '{"status":"error","msg":"Invalid Username, Email or Token.  Unable to find user."}';
		ELSE
			l_body = 
				'Hello, '||l_real_name||'<br>'
				||'<br>'
				||'Please follow the link below or cut and paste it into a browser to reset your password.<br>'
				||'<br>'
				||'         <a href="'||p_url||p_top||'?token='||l_token||'#/recover-password"> '||p_url||p_top||'?token='||l_token||'#/recover-password?token='||l_token||' </a><br>'
				||'<br>'
				;
		END IF;
	end if;
	
	-- ---------------------------------------------------------------------------------------
	-- Update user and create email.
	-- ---------------------------------------------------------------------------------------
	if not l_fail then	
		update "t_user"
			set "email_reset_key" = l_token
			  , "email_reset_timeout" = current_timestamp + interval '2 hours'
			where "id" = l_id;
		insert into "t_email_q" ( "user_id", "ip", "auth_token", "to", "from", "subject", "body", "status" )
			values ( l_id, p_ip_addr, p_auth_token, l_to, l_from, l_subject, l_body, 'pending' )
			;
	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

--delete from "t_email_q";
--select "id", "email_reset_key", "email_reset_timeout" from "t_user" where "username" = 'kc';
--select test_password_reset ( 'kc', '', '', '1.1.1.1', 'http://localhost:8090/', '/image.html' );
--select * from "t_email_q";
--delete from "t_email_q";
--select "id", "email_reset_key", "email_reset_timeout" from "t_user" where "username" = 'kc';

--select "password" from "t_user" where "username" = 'kc';
--select test_change_password ( 'liveBeef', 'liveBeef', 'cbe8782b-e86d-4a7a-9446-e37043043bd4', '1.1.1.1' );
--select "password" from "t_user" where "username" = 'kc';
--select "id", "email_reset_key", "email_reset_timeout" from "t_user" where "username" = 'kc';



CREATE or REPLACE FUNCTION status_db ( p_ip_addr varchar ) RETURNS varchar AS $$
DECLARE
	l_data			varchar (400);
BEGIN
	l_data = '{"status":"success","ip":'||to_json(p_ip_addr)||'}';
	RETURN l_data;
END;
$$ LANGUAGE plpgsql;



--CREATE TABLE "img_file" (
--	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key		-- file_id
--	, "img_set_id"			char varying (40)
--	, "n_acc"				bigint
--	, "file_name"			text
--	, "base_file_name"		text
--	, "width"				int
--	, "height"				int
--	, "description"			text
--	, "title"				text
--	, "attrs"				text
--	, "status"				char varying ( 10 )						-- One of 'sized', 'cmp-orig', 'orig'
--	, "upload_time"			timestamp
--	, "defaults_data"		text
--	, "use_cdn"				char varying ( 1 ) default 'n'											-- 
--	, "cdn_push_time"		timestamp
--	, "img_seq"				float
--	, "user_dir"			text
--	, "group_dir"			text
--	, "img_set_dir"			text
--	, "updated" 			timestamp 									 	-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
--	, "created" 			timestamp default current_timestamp not null 	-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
--);
-- // text := `'<table class="goofy"><tr><td>'||"title"||'</td><td><img class="bob" src=''/image/'||"user_dir"||'/'||"group_dir"||'/'||"img_set_dir"||'/'||"base_file_name"||'.jpg?req_h=46&req_w=0''></td></tr></table>' as "text"`
-- , filter_img_file(title,width,height,file_name,'/image/'||user_dir||'/'||group_dir||'/'||img_set_dir||'/'||base_file_name||'.jpg',base_file_name) as "text"













-- Note on E'\\"' strings - http://stackoverflow.com/questions/5785219/escaping-quotes-inside-text-when-dumping-postgres-sql 


-- delete from "t_output";

-- drop FUNCTION filter_img_file ( p_title varchar, p_width bigint, p_height bigint, p_file_name0 varchar, p_file_name varchar, p_base_file_name varchar );

CREATE or REPLACE FUNCTION filter_img_file2 ( p_title varchar, p_width bigint, p_height bigint, p_file_name0 varchar, p_file_name varchar, p_base_file_name varchar, p_cssClass varchar ) RETURNS varchar AS $$
DECLARE
	l_data			varchar (4000);
	x_height		int;
	l_useCss		varchar(50);
	l_cssClass		varchar(350);
BEGIN
	x_height = 46;
	if ( p_title is null ) then
		-- p_title = '';
		p_title = p_base_file_name;
	end if;
	if ( p_title = '' ) then
		p_title = p_base_file_name;
	end if;
	if ( p_width is null ) then
		p_width = 0;
	end if;
	if ( p_height is null ) then
		p_height = 0;
	end if;
	if ( p_file_name0 is null ) then
		p_file_name0 = '';
	end if;
	if ( p_file_name is null ) then
		p_file_name = '';
	end if;
	if ( p_base_file_name is null ) then
		p_base_file_name = '';
	end if;
	l_useCss = 'true';
	if ( p_cssClass is null ) then
		p_cssClass = '';
		l_useCss = 'false';
	end if;

	-- insert into "t_output" ( msg ) values ( 'A, '||p_file_name);

	-- and ( t2."img_set_id" = '22f06a74-3a47-4e82-be98-0213f5fe105f' and t2."base_file_name" = t1."base_file_name" and t2."height" = 46 and t2."status" != 'orig' )
	select "cssClass"
		into l_cssClass
		from "img_file"
		where "height" = 46
		  and "base_file_name" = p_base_file_name
	;
	if NOT FOUND then
		-- insert into "t_output" ( msg ) values ( 'B' );
		l_cssClass = '';
		l_useCss = 'false';
	else
		l_useCss = 'true';
		-- insert into "t_output" ( msg ) values ( 'G, '||l_cssClass );
	end if;
	if ( l_cssClass is null ) then
		-- insert into "t_output" ( msg ) values ( 'H' );
		l_cssClass = '';
		l_useCss = 'false';
	end if;

	-- 1. Must fix quote marks in strings
	l_data = '{"status":"success"'
		||',"height":'||p_height::varchar||''
		||',"width":'||p_width::varchar||''
		||',"use_background_img":'||l_useCss||''
		||',"styleForImage":"'||replace(l_cssClass,'"','\"')||'"'
		||',"file_name":"'||replace(p_file_name,'"','\"')||'?req_h='||x_height||'&req_w=0"'
		||',"r_name":"'||replace(p_file_name,'"','\"')||'?req_h='||x_height||'&req_w=0"'
		||',"title":"'||replace(p_title,'"','\"')||'"'
		||',"base_file_name":"'||replace(p_base_file_name,'"','\"')||'"'
		||'}'
		;
		-- insert into "t_output" ( msg ) values ( 'C' );
		-- ||',"orig_file_name":"'||replace(p_file_name0,'"','\"')||'"'
	RETURN l_data;
END;
$$ LANGUAGE plpgsql;




select
					t1."id"
					, filter_img_file2(t1."title",t1."width",t1."height",t1."file_name",'/image/'||t1."user_dir"||'/'||t1."group_dir"||'/'||t1."img_set_dir"||'/'||t1."base_file_name"||t1."ext",t1."base_file_name",t1."cssClass") as "text"
					, 'leaf' as "type"
					, false as "children"
					, t1."img_seq"
					, t1."user_dir"
					, t1."group_dir"
					, t1."img_set_dir"
,"height", "width", "title"
				from "img_file" as t1
				where t1."img_set_id" = '22f06a74-3a47-4e82-be98-0213f5fe105f' and t1."status" = 'orig' 
				order by t1."img_seq", t1."title", t1."base_file_name"
;

-- and t1."id" = 'b2cc459c-0d80-494f-b054-626729201f91'
-- select "msg" from "t_output";



drop TABLE "t_test_crud" ;
CREATE TABLE "t_test_crud" (
	  "id"				char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "name"			char varying (40) 
	, "value"			char varying (240)
);

insert into "t_test_crud" ( "id", "name", "value" ) values ( '46771e3a-68dc-40ad-7c57-aaa36f677489', 't1', 'hay bob' );
insert into "t_test_crud" ( "id", "name", "value" ) values ( 'c651123c-a3b9-4c51-46f6-f39121917051', 't2', 'hay bob' );

drop TABLE "t_test_crud2" ;
CREATE TABLE "t_test_crud2" (
	  "id"				char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "name"			char varying (40) 
	, "value1"			bigint
	, "value2"			float
	, "value3"			timestamp
	, "value4" 			char varying(80)
);

insert into "t_test_crud2" ( "name", "value1", "value2", "value3", "value4" ) values 
	( 't1', 121212, 123456.789, current_timestamp, 'the quick brown fox jumps over the lazy dog' )
;

-- ----------------------------------------------------- t_monitor_stuff -------------------------------------------------------
-- To create a new monitored item:
-- 1. Insert item into t_monitor_stuff -the table is in  ./tables-pg.m4.sql
-- 2. OR: Call the i_am_alive function once
--		select ping_i_am_alive('rpt-run','127.0.0.1','')	
-- 3. Then enable the item in the GUI
-- -----------------------------------------------------------------------------------------------------------------------------
drop TABLE "t_monitor_cfg" ;
--CREATE TABLE "t_monitor_cfg" (
--	  "id"				char varying (40) DEFAULT uuid_generate_v4() not null primary key
--	, "item_name"		char varying (250) 
--	, "minutes"			char varying (40)
--);
--delete from "t_monitor_cfg";
--insert into "t_monitor_cfg" ( "item_name", "minutes" ) values
--	( 'w-watch:cfg-json.sql',	            '4 minute' 	)
--,	( 'notif-js:postgres',		            '4 minute' 	)
--,	( 'tab-server1:8090',		            '4 minute' 	)
--,	( 'socket-app:8094',		            '4 minute' 	)
--,	( 'node:app.js:3050',		            '4 minute' 	)
--,	( 'Chantelle(.159)-Backup',	            '1 day' 	)
--,	( 'Postgres-DB',		            	'1 minute' 	)
--,	( 'Redis-DB',			            	'1 minute' 	)
--,	( 'DatabaseBackup',    		            '2 days' 	)
--,	( 'SystemBackup',      	            	'2 days' 	)
--,	( 'EmailSent-Q1',      		            '2 minutes' )
--,	( 'InternetUp',        	            	'2 minute' 	)
--,	( '99PctBrowsers',     		            '2 days' 	)
--,	( 'ws:http://www.crs-studio.com',		'15 minute' )
--,	( 'ws:http://blog.crs-studio.com',		'15 minute' )
--,	( 'ws:http://blog.2c-why.com',			'15 minute' )
--,	( 'rpt-run',	            			'4 minute' 	)
--;
-- ,	( 'EmailSent-Q2',      		'5 minutes' )
-- ,	( 'LoadManager',       		'4 minute' 	)
-- delete from "t_monitor_cfg" where "item_name" = 'EmailSent-Q2';
-- delete from "t_monitor_cfg" where "item_name" = 'LoadManager';


drop TABLE if exists "t_ignore_stuff" ;
--CREATE TABLE "t_ignore_stuff" (
--	  "id"				char varying (40) DEFAULT uuid_generate_v4() not null primary key
--	, "item_name"		char varying (250) 
--	, "minutes"			char varying (40)
--);

delete from "t_output";
CREATE or REPLACE FUNCTION ping_i_am_alive(p_item varchar, p_ip varchar, p_note varchar) RETURNS varchar AS $$
DECLARE
    l_id char varying(40);
    l_enabled char varying(40);
    l_minutes char varying(40);
    l_data char varying(400);
BEGIN

	insert into "t_output" ( msg ) values ( 'ping_i_am_alive called, p_item='||p_item );

	if p_ip is null then
		p_ip = '';
	end if;
	if p_note is null then
		p_note = '';
	end if;

	l_data = '{"status":"success","msg":"not monitored"}';

	select  "delta_t", "enabled"
	into  l_minutes, l_enabled
		from "t_monitor_stuff"
		where "item_name" = p_item
		;

	IF NOT FOUND THEN

		insert into "t_output" ( msg ) values ( '     doing insert to t_monitor_stuff' );

		l_minutes = '99 day';

		insert into "t_monitor_stuff" ( "item_name", "event_to_raise", "delta_t", "timeout_event", "note" )
			values ( p_item
				, 'System '||p_item||' has timed out.'
				, l_minutes
				, current_timestamp + CAST ( l_minutes as Interval )
				, p_note || 'IP:' || p_ip
		);

	ELSE

		IF l_enabled = 'n' THEN

			l_data = '{"status":"success","msg":"ignored"}';
			insert into "t_output" ( msg ) values ( '      ignored in t_monitor_stuff' );

		ELSE

			l_data = '{"status":"success","msg":"monitored"}';
			insert into "t_output" ( msg ) values ( '      doing update of t_monitor_stuff' );

			update "t_monitor_stuff"
				set
				  "timeout_event" = current_timestamp + CAST("delta_t" as Interval) 
				, "note" = p_note || 'IP:' || p_ip
				where "item_name" = p_item
			;

		END IF;

	END IF;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

-- select ping_i_am_alive( 'tab-server1:8090', '127.0.0.1', 'hi there ' );
-- select ping_i_am_alive( 'w-watch:cfg-json.sql', '127.0.0.1', 'hi there ' );




-- Notify Python/Twisted Example: http://www.divillo.com/ 
-- Notify Doc: http://www.postgresql.org/docs/9.0/static/sql-notify.html 
-- Trigger Doc: http://www.postgresql.org/docs/9.2/static/plpgsql-trigger.html 
CREATE or REPLACE FUNCTION notify_trigger() RETURNS trigger AS $$
DECLARE
BEGIN
	-- TG_TABLE_NAME is the name of the table who's trigger called this function
	-- TG_OP is the operation that triggered this function: INSERT, UPDATE or DELETE.
	execute 'NOTIFY w_watchers, ''' || TG_TABLE_NAME || ':' || TG_OP || ':' || NEW."id" || '''';
	return new;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER "t_test_crud2_n_ud" BEFORE update or delete on "t_test_crud2" for each row execute procedure notify_trigger();
CREATE TRIGGER "t_test_crud2_n_i"  AFTER  insert           on "t_test_crud2" for each row execute procedure notify_trigger();

CREATE or REPLACE FUNCTION notify_trigger_table() RETURNS trigger AS $$
DECLARE
BEGIN
	-- TG_TABLE_NAME is the name of the table who's trigger called this function
	-- TG_OP is the operation that triggered this function: INSERT, UPDATE or DELETE.
	execute 'NOTIFY w_watchers, ''' || TG_TABLE_NAME || ':' || TG_OP || '''';
	return new;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER "t_test_crud_tab" BEFORE insert or update or delete on "t_test_crud" execute procedure notify_trigger_table();

drop TABLE "t_test_crud3" ;
CREATE TABLE "t_test_crud3" (
	  "id1"				char varying (40) DEFAULT uuid_generate_v4() not null
	, "id2"				char varying (40) DEFAULT uuid_generate_v4() not null
	, "name"			char varying (40) 
);

create unique index "t_test_crud3_u1" on "t_test_crud3" ( "id1", "id2" )

insert into "t_test_crud3" ( "name" ) values ( 'Hi Bob' );

CREATE or REPLACE FUNCTION notify_trigger_0001() RETURNS trigger AS $$
DECLARE
BEGIN
	execute 'NOTIFY w_watchers, ''' || TG_TABLE_NAME || ':' || TG_OP || ':' || NEW."id1" || '/' || NEW."id2" || '''';
	return new;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER "t_test_crud3_n_ud" BEFORE update or delete on "t_test_crud3" for each row execute procedure notify_trigger_0001();
CREATE TRIGGER "t_test_crud3_n_i"  AFTER  insert           on "t_test_crud3" for each row execute procedure notify_trigger_0001();




CREATE or REPLACE FUNCTION get_ip(p_ip varchar) RETURNS varchar AS $$
DECLARE
    l_data 		varchar (400);
BEGIN
	l_data = '{ "status":"success"}';
	update "t_config_data"
		set "value" = p_ip
		where "name" = 'host-ip'
	;
	-- xyzzy - check for errors and return error if so.
	RETURN l_data;
END;
$$ LANGUAGE plpgsql;




-- xyzyz - what to do?
-- ,"/api/releaseJob": { "g": "generate_job($1,$2)"
-- ,"/api/releaseJob": { "g": "releaseJob($1,$2)"

-- drop FUNCTION updatePassword(p_oldPassword varchar, p_newPassword varchar, p_newPasword2 varchar, p_username varchar, p_ip varchar, p_auth_token varchar);
CREATE or REPLACE FUNCTION updatePassword(p_oldPassword varchar, p_newPassword varchar, p_newPassword2 varchar, p_username varchar, p_ip varchar, p_auth_token varchar) RETURNS varchar AS $$
DECLARE
    l_data 			varchar (400);
    l_privs 		varchar (400);
	l_junk			varchar (1);
	l_id			varchar (40);
	l_customer_id	varchar (40);
	l_fail 			boolean;
	l_ip_ban 		boolean;
	l_salt 			varchar(80);
	l_password 		varchar(80);
  	l_auth_token	varchar (40);
BEGIN
	l_fail = false;
	l_data = '{ "status":"unknown"}';

	select 'y' 
		into l_junk
		from "t_ip_ban"
		where "ip" = p_ip
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

	if not l_fail then
		if ( p_newPassword != p_newPassword2 ) then
			l_data = '{ "status":"error", "code":"081", "msg":"new password and password again did not match." }';
			l_fail = true;
		end if;
	end if;

	if not l_fail then

		select "id", "privs", "customer_id"
			into l_id
				, l_privs
				, l_customer_id
			from "t_user"
			where "auth_token" = p_auth_token
			;

		if not found then
			l_fail = true;
			l_data = '{ "status":"failed", "code":"029", "msg":"Invalid auth_token." }';
		end if;

	end if;

	if not l_fail then

		l_auth_token = uuid_generate_v4();
		l_salt = uuid_generate_v4();
		l_password = sha256pw ( l_salt||p_newPassword||l_salt );

		update "t_user" set 
			  "password" = l_password
			, "salt" = l_salt
			, "auth_token" = l_auth_token
			, "password_set_date" = current_timestamp
			, "ip" = p_ip
			, "acct_state" = 'ok'
			, "n_login_fail" = 0
			where "id" = l_id
			;

		l_data = '{ "status":"success", "auth_token":'||to_json(l_auth_token)
			||', "user_id":'||to_json(l_id)
			||', "privs":'||to_json(l_privs)
			||', "customer_id":'||to_json(l_customer_id)
			||'}';

	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;


-- send_result ( res, next, null, { "status":"error", "msg":"Unfortunately the existing password you entered was not correct.  Your password was not updated.", "line":3138 } );
-- send_result ( res, next, null, { "status":"error", "msg":"Unfortunately the password you entered was in our list of banned passwords.  You will need to pick a different one or generate one.", "line":3143 } );
-- send_result ( res, next, null, { "status":"error", "msg":"Unfortunately the existing password you entered was not correct.  Your password was not updated.", "line":3146 } );
-- send_result ( res, next, null, { "status":"error", "msg":"invalid auth_token, not authorized to use this service.", "line":3153 } );




-- Original 574 lines of Node.js
-- New 174 lines of code.
CREATE or REPLACE FUNCTION test_register_client(
	  p_host_id varchar
	, p_name varchar
	, p_ip varchar
	, p_method varchar
	, p_ua varchar
	, p_ua_family varchar
	, p_ua_major varchar
	, p_ua_minor varchar
	, p_ua_patch varchar
	, p_os varchar
	, p_os_family varchar
	, p_os_major varchar
	, p_os_minor varchar
	, p_os_patch varchar
	, p_is_mobile varchar
) RETURNS varchar AS $$
DECLARE
    l_data 				varchar(400);
	l_junk				varchar(1);
	l_fail 				boolean;
	l_ip_ban 			boolean;
	l_useragent_id 		varchar(40);
	l_client_id 		varchar(40);
	l_host_can_run_id	varchar(40);
	l_name 				varchar(200); 
BEGIN
	l_fail = false;
	l_data = '{ "status":"unknown"}';

	select 'y' 
		into l_junk
		from "t_ip_ban"
		where "ip" = p_ip
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

	if not l_fail then

		select "id", "name"
		into l_useragent_id, l_name
		from "t_userAgent" 
		where "browserFamily" ~* p_ua_family
		  and "browserMajor"  =  p_ua_major
		  and "browserMinor"  =  p_ua_minor
		  and "osFamily"      ~* p_os_family
		  and "osMajor"       =  p_os_major
		  and "osMinor"       =  p_os_minor
		limit 1 
		;

		if not found then

			l_useragent_id = uuid_generate_v4();
			l_name = p_ua_family || ' ' || p_ua_major;

			insert into "t_userAgent" ( 
				  "id" 
				, "name" 
				, "title" 
				, "browserFamily" 
				, "browserMajor" 
				, "browserMinor" 
				, "osFamily" 
				, "osMajor" 
				, "osMinor" 
			 ) values ( 
				  l_useragent_id
				, l_name
				, l_name
				, p_ua_family
				, p_ua_major
				, p_ua_minor
				, p_os_family
				, p_os_major
				, p_os_minor
			);
 
		end if;

		select "id" 
			into l_client_id
			from "t_client" 
			where "host_id" = p_host_id
			  and "useragent_id"  = l_useragent_id
			  and "name"  = l_name
		;

		if not found then

			l_client_id = uuid_generate_v4();

			insert into "t_client" ( 
				  "id"
				, "useragent"
				, "useragent_id"
				, "ip"
				, "host_id"
				, "name" 
			) values ( 
				  l_client_id
				, p_ua
				, l_useragent_id
				, p_ip
				, p_host_id
				, p_name
			);

		end if;

		select "id" 
			into l_host_can_run_id
			from "t_host_can_run" 
			where "host_id" = p_host_id
			   and "useragent_id"  = l_useragent_id
		;

		if not found then

			l_host_can_run_id = uuid_generate_v4();

			insert into "t_host_can_run" ( 
				"id", 
				"host_id", 
				"useragent_id", 
				"client_id" 
			) values ( 
				  l_host_can_run_id
				, p_host_id
				, l_useragent_id
				, l_client_id
			);

		end if;

		p_ua_family = lower(p_ua_family);
		p_os_family = lower(p_os_family);

		l_data = '{ "status":"success"'
			||', "client_id":'||to_json(l_client_id)
			||', "useragent_id":'||to_json(l_useragent_id)
			||', "browserCSSClass":'||to_json( 'swarm-browser-'||p_ua_family||' swarm-browser-'||p_ua_family||'-'||p_ua_major||' '||'swarm-os swarm-os-'||p_os_family )
			||', "browserDisplayName":'||to_json(p_ua_family || ' ' || p_ua_major ||'.'||p_ua_minor||'<br>'||p_os_family)
			||', "browserDisplayTitle":'||to_json(p_ua_family || ' ' || p_ua_major ||'.'||p_ua_minor||'/'||p_os_family)
			||'}'
			;

	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;














CREATE or REPLACE FUNCTION test_getrun(
	  p_client_id varchar
	, p_host_id varchar
	, p_useragent_id varchar
	, p_client_name varchar
	, p_name varchar
	, p_ip varchar
	, p_ua varchar
	, p_ua_family varchar
	, p_ua_major varchar
	, p_ua_minor varchar
	, p_ua_patch varchar
	, p_os varchar
	, p_os_family varchar
	, p_os_major varchar
	, p_os_minor varchar
	, p_os_patch varchar
	, p_is_mobile varchar
) RETURNS varchar AS $$
DECLARE
    l_data 				varchar(400);
	l_junk				varchar(1);
	l_fail 				boolean;
	l_ip_ban 			boolean;
	l_useragent_id 		varchar(40);
	l_client_id 		varchar(40);
	l_host_can_run_id	varchar(40);
	l_name 				varchar(200); 
	l_job_id			varchar(40);
	l_url				varchar(400);
	l_run_name			varchar(100);
	l_height			int;
	l_width				int;
	l_id				varchar(40);
	l_auth_token		varchar(40);
	l_user_data			varchar(255);
	l_run_id			varchar(40);
BEGIN
	l_fail = false;
	l_data = '{ "status":"unknown"}';

	select 'y' 
		into l_junk
		from "t_ip_ban"
		where "ip" = p_ip
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

	if not l_fail then

		select "id", "name"
			into l_useragent_id, l_name
			from "t_userAgent" 
			where "browserFamily" ~* p_ua_family
			  and "browserMajor"  =  p_ua_major
			  and "browserMinor"  =  p_ua_minor
			  and "osFamily"      ~* p_os_family
			  and "osMajor"       =  p_os_major
			  and "osMinor"       =  p_os_minor
		limit 1 
		;

		if not found then

			l_useragent_id = uuid_generate_v4();
			l_name = p_ua_family || ' ' || p_ua_major;

			insert into "t_userAgent" ( 
				  "id" 
				, "name" 
				, "title" 
				, "browserFamily" 
				, "browserMajor" 
				, "browserMinor" 
				, "osFamily" 
				, "osMajor" 
				, "osMinor" 
			 ) values ( 
				  l_useragent_id
				, l_name
				, l_name
				, p_ua_family
				, p_ua_major
				, p_ua_minor
				, p_os_family
				, p_os_major
				, p_os_minor
			);

		else

			l_useragent_id = p_useragent_id; 

		end if;

		select find_work( l_useragent_id, p_client_id, p_host_id, p_client_name ) as "run_id" 
			into l_run_id
			;

		select job_id
			, url
			, run_name
			, height
			, width
			, id
			, auth_token
			, user_data
			into
				  l_job_id
				, l_url
				, l_run_name
				, l_height
				, l_width
				, l_id
				, l_auth_token
				, l_user_data
			from "t_a_run" 
			where "id" = l_run_id
		;

		if l_height is null then
			l_height = -1;
		end if;
		if l_width is null then
			l_width = -1;
		end if;

		if found then
			 l_data = '{"getrun":{"runInfo":{"id":'||to_json(l_job_id)||
						',"url":'||to_json(l_url)||
						',"desc":'||to_json(l_run_name)||
						',"height":'||to_json(l_height)||
						',"width":'||to_json(l_width)||
						',"resultsId":'||to_json(l_id)||
						',"resultsStoreToken":'||to_json(l_auth_token)||
						',"user_data":'||to_json(l_user_data)||
						'}}}';
		else
			-- g_rv.noWork = "no-work";
			l_data = '{ "status":"success", "noWork":"no-work" }';
		end if;


--		l_data = '{ "status":"success"'
--			||', "client_id":'||to_json(l_client_id)
--			||', "useragent_id":'||to_json(l_useragent_id)
--			||', "browserCSSClass":'||to_json( 'swarm-browser-'||p_ua_family||' swarm-browser-'||p_ua_family||'-'||p_ua_major||' '||'swarm-os swarm-os-'||p_os_family )
--			||', "browserDisplayName":'||to_json(p_ua_family || ' ' || p_ua_major ||'.'||p_ua_minor||'<br>'||p_os_family)
--			||', "browserDisplayTitle":'||to_json(p_ua_family || ' ' || p_ua_major ||'.'||p_ua_minor||'/'||p_os_family)
--			||'}'
--			;

	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;



--select test_getrun(
--	  '1'
--	, '1'
--	, '1'
--	, '1'
--	, '1'
--	, '1'
--	, '1'
--	, '1'
--	, '1'
--	, '1'
--	, '1'
--	, '1'
--	, '1'
--	, '1'
--	, '1'
--	, '1'
--	, '1'
--) ;




-- theMux.HandleFunc("/api/edit1TestSetMember"				, respHandlerEdit1TestSetMember).Methods("POST")
-- /api/edit1TestSetMember
--	{ runset_id:"101", runSet_members:[207, 200, 201] }

CREATE type idList as (id varchar);

--		stmt0 = 
--			'delete from "t_runSet_members" '
--			+'where "runset_id" = \'%{runset_id%}\' '
--			+'  and "useragent_id" not in ( %{useragent_list%} ) '
--			;
--		dn.runQuery ( stmt = ts0( stmt0, hh = { "runset_id":runset_id, "useragent_list": useragent_list, "user_id": user_id } )
--			, function ( err, result ) {
--				console.log ( "stmt="+stmt );
--				if ( err === null ) {
--			
--					// Insert into XXX select ... where in ( List ) and not exists ( Query )
--					stmt0 =
--						'insert into "t_runSet_members" ( "runset_id", "useragent_id" ) '
--						+'select \'%{runset_id%}\', t1."useragent_id" '
--						+'from "t_available_test_systems" t1 '
--						+'where t1."useragent_id" in ( %{useragent_list%} ) '
--						+'  and not exists ( '
--						+      'select 1 as "found" '
--						+      'from "t_runSet_members" t3 '
--						+      'where t3."runset_id" = \'%{runset_id%}\' '
--						+	   '  and t3."useragent_id" = t1."useragent_id" '
--						+'  ) '
--						;
--


--insert into "t_runSet_members" ( "id", "runset_id", "seq", "useragent_id", "created" ) values ( 'ce84b629-a7be-4dd9-9efb-4f877cf05418', '100', '1', '200', '2014-04-29T19:35:38.70578Z' );
--insert into "t_runSet_members" ( "id", "runset_id", "seq", "useragent_id", "created" ) values ( 'f41ba715-11a9-4eba-bfa3-c782c2a1fb65', '100', '2', '202', '2014-04-29T19:35:38.71414Z' );
--insert into "t_runSet_members" ( "id", "runset_id", "seq", "useragent_id", "created" ) values ( '0e7999d7-8a0c-425f-b2b6-71f211cb93de', '100', '3', '203', '2014-04-29T19:35:38.72231Z' );
--insert into "t_runSet_members" ( "id", "runset_id", "seq", "useragent_id", "created" ) values ( 'd1d08ed9-9837-4b1d-b4c5-5ecda5f732fd', '100', '4', '204', '2014-04-29T19:35:38.73065Z' );
--insert into "t_runSet_members" ( "id", "runset_id", "seq", "useragent_id", "created" ) values ( '3f223dcd-1d43-4787-9c32-af61349b74c7', '100', '5', '205', '2014-04-29T19:35:38.73904Z' );
--insert into "t_runSet_members" ( "id", "runset_id", "seq", "useragent_id", "created" ) values ( '3b4e2ff1-0bd4-4eb1-90e4-722c53efdacd', '100', '6', '207', '2014-04-29T19:35:38.7474Z' );
--
--select * from "t_runSet_members";

CREATE OR REPLACE FUNCTION to_string(anyarray, sep text, nullstr text DEFAULT '') 
RETURNS text AS $$
SELECT array_to_string(ARRAY(SELECT coalesce(v::text, $3) 
                                FROM unnest($1) g(v)),
                       $2)
$$ LANGUAGE sql;

CREATE OR REPLACE FUNCTION to_array(text, sep text, nullstr text DEFAULT '') 
RETURNS text[] AS $$ 
  SELECT ARRAY(SELECT CASE 
                           WHEN v = $3 THEN NULL::text 
                           ELSE v END 
                  FROM unnest(string_to_array($1,$2)) g(v)) 
$$ LANGUAGE sql;


CREATE or REPLACE FUNCTION edit1TestSetMember(p_runset_id varchar, p_runSet_members varchar ) RETURNS varchar AS $$
DECLARE
    l_data 		varchar (400);
	l_junk		varchar (1);
	l_fail 		boolean;
    r 			idList%rowtype;
	l_tmp		varchar(40)[];
	i			int;
BEGIN
	l_fail = false;
	l_data = '{ "status":"success"}';

	i = 0;
	FOR r IN select * from json_array_elements(p_runSet_members::json)
	LOOP
		-- can do some processing here
		l_tmp[i] = replace(r.id,'"','');
		i = i + 1;
	END LOOP;

	delete from "t_runSet_members"
		where "useragent_id" not in(SELECT(UNNEST(l_tmp)))  
		  and "runset_id" = p_runset_id
	;
	insert into "t_runSet_members" ( "runset_id", "useragent_id" ) 
		select p_runset_id, t1."useragent_id" 
			from "t_available_test_systems" t1 
			where t1."useragent_id" in(SELECT(UNNEST(l_tmp)))  
			  and not exists ( 
				select 1 as "found" 
					from "t_runSet_members" t3 
					where t3."runset_id" = p_runset_id
					  and t3."useragent_id" = t1."useragent_id" 
			  )
	;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

--select * from "t_runSet_members";
--
--select edit1TestSetMember('100', '["207","204"]' );
--
--select "useragent_id"
--from "t_runSet_members"
--	where "runset_id" = '100'
--;
--
--select edit1TestSetMember('100', '["207","204","201"]' );
--
--select "useragent_id"
--from "t_runSet_members"
--	where "runset_id" = '100'
--;
--

CREATE or REPLACE FUNCTION updatePasswordInternal( p_newPassword varchar,  p_username varchar  ) RETURNS varchar AS $$
DECLARE
    l_data 			varchar (400);
    l_privs 		varchar (400);
	l_junk			varchar (1);
	l_id			varchar (40);
	l_customer_id	varchar (40);
	l_fail 			boolean;
	l_ip_ban 		boolean;
	l_salt 			varchar(80);
	l_password 		varchar(80);
  	l_auth_token	varchar (40);
BEGIN
	l_fail = false;
	l_data = '{ "status":"unknown"}';

	select "id", "privs", "customer_id"
		into l_id
			, l_privs
			, l_customer_id
		from "t_user"
		where "username" = p_username
		;

	if not found then
		l_fail = true;
		l_data = '{ "status":"failed", "code":"029", "msg":"Invalid auth_token." }';
	end if;

	if not l_fail then

		l_auth_token = uuid_generate_v4();
		l_salt = uuid_generate_v4();
		l_password = sha256pw ( l_salt||p_newPassword||l_salt );

		update "t_user" set 
			  "password" = l_password
			, "salt" = l_salt
			, "auth_token" = l_auth_token
			, "password_set_date" = current_timestamp
			, "acct_state" = 'ok'
			, "n_login_fail" = 0
			where "id" = l_id
			;

		l_data = '{ "status":"success", "auth_token":'||to_json(l_auth_token)
			||', "user_id":'||to_json(l_id)
			||', "privs":'||to_json(l_privs)
			||', "customer_id":'||to_json(l_customer_id)
			||'}';

	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

-- select updatePasswordInternal( 'deadbeef', 'goofy' );





CREATE or REPLACE FUNCTION test_stayLoggedIn() RETURNS varchar AS $$
DECLARE
BEGIN
	RETURN '{ "status":"success"}';
END;
$$ LANGUAGE plpgsql;

CREATE TABLE dual (
	  x	text
);
delete from dual;
insert into dual ( x ) values ( 'y' );



-- alter table "t_rpt_q" add column  "url_html"			text;
-- alter table "t_rpt_q" add column  "url_pdf"				text;

-- Table Depricated ---------------------------------!! ! ! ! 
-- See t_rpt_run_q
-- drop TABLE "t_rpt_q" ;
drop TABLE "t_rpt_q" ;
--CREATE TABLE "t_rpt_q" (
--	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
--  	, "status" 				char varying (40) null default 'init'
--								check ( "status" in (
--									  'init'										-- Create State, waiting to be picked
--									, 'started-run'									-- 
--									, 'generated'									-- 
--									, 'done'										-- 
--									, 'error'										-- 
--								) )
--	, "cli"					text
--	, "rv"					text
--	, "url_html"			text
--	, "url_pdf"				text
--	, "start_at" 			timestamp 
--	, "done_at" 			timestamp 
--	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
--	, "created" 			timestamp default current_timestamp not null 						--
--);
--
--insert into "t_rpt_q" ( "cli" ) values 
--	 ( '{"cmd":["ls",".."]}' )
--	,( '{"cmd":["cat","rpt-qry.go"]}' )
--;



--	,"/api/test/en_q_report": { "g": "test_en_q_report($1)"
--	,"/api/test/report_status": { "g": "test_report_status($1)"

-- Sample CLI:  {"cmd":["go-sql","-i","rpt-daily.rpt","-c","{\"auth_token\":\"10bd00e5-6258-436e-a581-7cbd5363885b\",\"dest\":\"Email\",\"to\":\"g@h.com\",\"subject\":\" Saftey Observation - 2014-08-17 04:50:36\",\"site\":\"\",\"from\":\"2014-01-05 05:00:00\",\"thru\":\"2014-08-12 05:00:00\"}","",""," Limit 42 ","FromDate","ToDate","MineName"]}
-- Test CLI:  {"cmd":["go-sql","-i","shipping-label.rpt","-c","{\"auth_token\":\"10bd00e5-6258-436e-a581-7cbd5363885b\",\"dest\":\"Email\",\"to\":\"pschlump@gmail.com\",\"invoice_id\":\"000000-0000-0000-000000000000"}","","","","","",""]}

drop FUNCTION test_en_q_report(p_cli varchar);
--CREATE or REPLACE FUNCTION test_en_q_report(p_cli varchar)
--RETURNS varchar AS $$
--DECLARE
--	l_id			varchar (40);
--BEGIN
--	l_id = uuid_generate_v4();
--	insert into "t_rpt_q" (
--		  "id"
--		, "cli"
--	) values (
--		  l_id
--		, p_cli
--	);
--	RETURN '{"status":"success","rid":'||to_json(l_id)||'}';
--END;
--$$ LANGUAGE plpgsql;



-- drop FUNCTION test_report_status(p_cli varchar);

drop FUNCTION test_report_status(p_id varchar);
--CREATE or REPLACE FUNCTION test_report_status(p_id varchar)
--RETURNS varchar AS $$
--DECLARE
--	l_rv			varchar (400);
--	l_status		varchar (40);
--BEGIN
--	select "status", "rv"
--		into l_status, l_rv
--		from "t_rpt_q"
--		where "id" = p_id
--		;
--	if l_rv is null then
--		l_rv = '';
--	end if;
--	RETURN '{"status":'||to_json(l_status)||',"rv":'||to_json(l_rv)||'}';
--END;
--$$ LANGUAGE plpgsql;


--	,"/api/test/save_customer_config": { "g": "test_save_customer_config($1)"
drop FUNCTION test_save_customer_config(p_customer_id varchar, p_cfg varchar);
CREATE or REPLACE FUNCTION test_save_customer_config(p_customer_id varchar, p_cfg varchar)
RETURNS varchar AS $$
DECLARE
	l_rv			varchar (400);
BEGIN
	update "t_customer"
		set "config" = p_cfg
		where "id" = p_customer_id
	;
	RETURN '{"status":"success"}';
END;
$$ LANGUAGE plpgsql;

