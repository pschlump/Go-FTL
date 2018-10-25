-- 
-- Description: This is Recover Lost/Forgotten Password - Part1. -- Part2 is below.
--
-- Send an email to the user's account with a token to allow the reset of the password.
--
--
-- Note:Important! You must supply *one* of p_username, p_auth_token, p_email to this function.
--
-- "Copyright (C) Philip Schlump, 2009-2017." 
--

drop FUNCTION s_password_reset ( p_username varchar, p_auth_token varchar, p_email varchar, p_ip_addr varchar, p_url varchar, p_top varchar );
drop FUNCTION s_password_reset ( p_username varchar, p_auth_token varchar, p_email varchar, p_ip_addr varchar, p_url varchar );

CREATE or REPLACE FUNCTION s_password_reset ( p_username varchar, p_auth_token varchar, p_email varchar, p_ip_addr varchar, p_url varchar )
	RETURNS varchar AS $$
DECLARE
	l_data			varchar (400);
	l_id			varchar (40);
	l_token			varchar (40);
	l_customer_id	varchar (40);
	l_fail 			boolean;
	l_from 			varchar (100);
	l_to 			varchar (100);
	l_real_name		varchar (100);
	l_subject 		varchar (100);
	l_body 			varchar (400);
	l_user_id		varchar(40);
	l_log_id 		bigint;
BEGIN

	-- "Copyright (C) Philip Schlump, 2009-2017." 

	l_fail = false;
	l_data = '{"status":"error","code":"500"}';
	l_log_id = 1;

	select nextval('t_email_id_seq'::regclass) 
		into l_log_id;

	if ( p_username = '' or p_username is null ) and ( p_auth_token = '' or p_auth_token is null ) and ( p_email = '' or p_email is null ) then
		l_fail = true;
		l_data = '{"status":"error","code":"501","msg":"Must supply one of username/auth_token/email"}';
	end if;

	l_token = uuid_generate_v4();

	-- xyzzy - get customer id via URL

	if not l_fail then	
		select "customer_id"
			into l_customer_id
			from "t_host_to_customer"
			where "host_name" = p_url
			;
		if not found then
			l_fail = true;
			l_customer_id = '0';
			l_data = '{"status":"error","code":"502","url":"'||p_url||'"}';
		end if;
	end if;
	
	-- xyzzy - get from.address

	l_from = s_get_config_item('from.address', l_customer_id, 'x');
	if l_from = 'x' then 
		l_fail = true;
		l_from = 'pschlump@yahoo.com';
		l_data = '{"status":"error","code":"503"}';
	end if;

	if not l_fail then	
		if s_ip_ban(p_ip_addr) then
			l_fail = true;
			l_data = '{ "status":"error", "code":"504", "msg":"Invalid username or password." }';
		end if;
	end if;

	if not l_fail and p_auth_token is not null and p_auth_token <> '' then
		select  "user_id"
			into  l_user_id
			from "t_auth_token" 
			where "auth_token" = p_auth_token
			  and "expire" >= current_timestamp
			;
		if not found then
			l_data = '{"status":"error", "code":"505", "msg":"Error: Not Authorized."}';
			l_fail = true;
		end if;
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
		 	   or "id" = l_user_id
			;
		if not found then
			l_fail = true;
			l_data = '{"status":"error","code":"506","msg":"Invalid Username, Email or Token.  Unable to find user."}';
		end if;
	end if;
	
	-- ---------------------------------------------------------------------------------------
	-- Update user and create email.
	-- ---------------------------------------------------------------------------------------
	if not l_fail then	
		update "t_user"
			set "email_reset_key" = l_token
			  , "email_reset_timeout" = current_timestamp + interval '2 hours'
			where "id" = l_id
			;

		-- NOTE: chagne email on this !

		l_data = '{"status":"success"'
			||',"$send_email$":{'
				||'"template":"password_recovered"'
				||',"username":'||to_json(p_username)
				||',"email_token":'||to_json(l_token)
				||',"to":'||to_json(l_to)
				||',"email_addr":'||to_json(l_to)
				||',"real_name":'||to_json(l_real_name)
				||',"url":'||to_json(p_url)
				||',"from":'||to_json(l_from)
				||',"log_id":"'||to_json(l_log_id)||'"'
			||'}'
			||'}';

	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;







-- ---------------------------------------------------------------------------------------------------------------------------------------------------------
--
-- Description: This is Recover Lost/Forgotten Password - Part2.
--
-- This is called from the "link" when it is clicked on - with the email_token
--
-- "Copyright (C) Philip Schlump, 2009-2017." 
--
drop FUNCTION s_password_reset_pt2 ( p_password varchar, p_again varchar, p_token varchar, p_ip_addr varchar, p_url varchar );
drop FUNCTION s_password_reset_pt2 (  p_token varchar, p_ip_addr varchar, p_url varchar );

CREATE or REPLACE FUNCTION s_password_reset_pt2 (  p_token varchar, p_ip_addr varchar, p_url varchar )
	RETURNS varchar AS $$
DECLARE
	l_data			varchar (800);
	l_id			varchar (40);
	l_customer_id	varchar (40);
	l_fail 			boolean;
	l_log_id 		bigint;
	l_redirect_to	varchar (400);
    l_seq 			varchar (40);
BEGIN

	-- "Copyright (C) Philip Schlump, 2009-2017." 

	l_fail = false;
	l_data = '{"status":"success"}';
	l_log_id = 1;

	select nextval('t_email_id_seq'::regclass) 
		into l_log_id;

	if p_token = '' || p_token is null then
		l_fail = true;
		l_data = '{"status":"error", "code":"510", "msg":"Invalid reseset token."}';
	end if;

	if s_ip_ban(p_ip_addr) then
		l_fail = true;
		l_data = '{ "status":"error", "code":"511", "msg":"Invalid username or password." }';
	end if;

	if not l_fail then	
		select "t_user"."id"
			into l_id
			from "t_user" as "t_user" 
			where "email_reset_key" = p_token
			  and "email_reset_timeout" <= current_timestamp 
			;
		IF FOUND THEN
			l_fail = true;
			l_data = '{"status":"error", "code":"512", "msg":"Your password-reset has expired - please start over."}';
		END IF;
	end if;
	
	if not l_fail then	
		select "t_user"."id"
				, "t_user"."customer_id"
			into l_id
				, l_customer_id
			from "t_user" as "t_user" left join "t_customer" as "t_customer" on "t_customer"."id" = "t_user"."customer_id"
			where "email_reset_key" = p_token
			  and "email_reset_timeout" > current_timestamp 
			;
		IF NOT FOUND THEN
			l_fail = true;
			l_data = '{"status":"error", "code":"513", "msg":"Invalid Token or Expired Token"}';
		END IF;
	end if;

	-- xyzzy
	select "value"
		into l_redirect_to
		from "t_config"
		where "item_name" = 'recover.redirect.to'
		  and "customer_id" = l_customer_id
		;
		l_redirect_to = p_url || l_redirect_to;
	if not found then
		l_redirect_to = p_url || '/';
	end if;

	if not l_fail then	

		--update "t_user" set
		--		 "ip" = p_ip_addr
		--	where "id" = l_id
		--	;

		l_data = '{ "status":"success"'
			||', "recovery_token":'||to_json(p_token)
			||', "$session$":{'
				||'"set":['
					||'{"path":["user","$is_logged_in$"],"value":"n"}'
				||']'
			||'}'
			||', "$redirect_to$":'||to_json(l_redirect_to)
			||', "$redirect_vars$":["recovery_token"]'
			||'}' ;
	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;





-- ---------------------------------------------------------------------------------------------------------------------------------------------------------
--
-- Description: This is Recover Lost/Forgotten Password - Part2.
--
-- This is part 3 - call - after the ??? -- email generated by FUNCTION s_password_reset 
-- For just a password-change - see function below.
--
-- "Copyright (C) Philip Schlump, 2009-2017." 
--
drop FUNCTION s_password_reset_pt3 ( p_password varchar, p_again varchar, p_token varchar, p_ip_addr varchar );
drop FUNCTION s_password_reset_pt3 ( p_password varchar, p_again varchar, p_token varchar, p_ip_addr varchar, p_url varchar );

CREATE or REPLACE FUNCTION s_password_reset_pt3 ( p_password varchar, p_again varchar, p_token varchar, p_ip_addr varchar, p_url varchar )
	RETURNS varchar AS $$
DECLARE
	l_data			varchar (800);
	l_username		varchar (40);
	l_privs			varchar (400);
	l_id			varchar (40);
	l_customer_id	varchar (40);
	l_real_name		varchar (200);
	l_to			varchar (200);
	l_fail 			boolean;
	l_auth_token 	varchar(40);
	l_log_id 		bigint;
    l_seq 			varchar (40);
	l_config		varchar (7500);
	l_xsrf_token	varchar(40);
	l_94_days 		varchar(50);
BEGIN

	-- "Copyright (C) Philip Schlump, 2009-2017." 

	l_fail = false;
	l_data = '{"status":"success"}';
	l_log_id = 1;

	select nextval('t_email_id_seq'::regclass) 
		into l_log_id;

	if p_password != p_again then
		l_fail = true;
		l_data = '{"status":"error", "code":"521", "msg":"Passwords did not match."}';
	end if;

	if p_token = '' || p_token is null then
		l_fail = true;
		l_data = '{"status":"error", "code":"522", "msg":"Invalid reseset token."}';
	end if;

	if s_ip_ban(p_ip_addr) then
		l_data = '{ "status":"failed", "code":"523", "msg":"Invalid username or password." }';
		l_fail = true;
	end if;
	
	l_xsrf_token = uuid_generate_v4();

	if not l_fail then	
		select "t_user"."id"
				, "t_user"."privs"
				, "t_user"."customer_id"
				, "t_user"."username"
				, "t_user"."real_name"
				, "t_user"."email_address"
				, "t_customer"."config"
			into l_id
				, l_privs
				, l_customer_id
				, l_username
				, l_real_name
				, l_to
				, l_config
			from "t_user" as "t_user" left join "t_customer" as "t_customer" on "t_customer"."id" = "t_user"."customer_id"
			where "email_reset_key" = p_token
			  and "email_reset_timeout" > current_timestamp 
			;
		IF NOT FOUND THEN
			l_fail = true;
			l_data = '{"status":"error", "code":"524", "msg":"Invalid Token or Expired Token"}';
		END IF;
	end if;

	l_94_days = s_get_config_item( 'acct.auth_token.expire', l_customer_id, '94 days');

	if not l_fail then	
		delete from "t_auth_token" where "user_id" = l_id;	-- delete all otehr auth tokens - invalidates current logins.
		l_auth_token = uuid_generate_v4();
		l_seq = uuid_generate_v4();
		insert into "t_auth_token" (
			  "auth_token"	
			, "user_id"	
			, "expire"
		) values (
			  l_auth_token
			, l_id
			, current_timestamp + l_94_days::interval
		);
		update "t_user" set
				  "acct_state" = 'ok'
				, "ip" = p_ip_addr
				, "last_login" = current_timestamp
				, "n_login_fail" = 0
				, "login_fail_delay" = null
			    , "password" = crypt(p_password,gen_salt('bf',8))
			  	, "email_reset_key" = null
			  	, "email_reset_timeout" = null
			where "id" = l_id
			;

		l_data = '{ "status":"success"'
			||', "user_id":'||to_json(l_id)
			||', "customer_id":'||to_json(l_customer_id)
			||', "username":'||to_json(l_username)
			||', "auth_token":'||to_json(l_auth_token)
			||', "seq":'||to_json(l_seq)
			||', "privs":'||to_json(l_privs)
			||', "config":'||to_json(l_config)
			||', "xsrf_token":'||to_json(l_xsrf_token)
			||', "$JWT-claims$":["auth_token"]'
			||', "$session$":{'
				||'"set":['
					||'{"path":["user","$is_logged_in$"],"value":"y"}'
					||',{"path":["user","$xsrf_token$"],"value":'||to_json(l_xsrf_token)||'}'
				||']'
			||'}'
			||'}' ;
	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

