





-- "Copyright (C) Philip Schlump, 2009-2017." 

drop FUNCTION s_register_full(p_username varchar, p_password varchar, p_again varchar, p_ip_addr varchar, p_email varchar, p_real_name varchar, p_url varchar, p_app varchar );
drop FUNCTION s_register_full(p_username varchar, p_password varchar, p_again varchar, p_ip_addr varchar, p_email varchar, p_real_name varchar, p_url varchar, p_app varchar, p_method varchar );
CREATE or REPLACE FUNCTION s_register_full(p_username varchar, p_password varchar, p_again varchar, p_ip_addr varchar, p_email varchar, p_real_name varchar, p_url varchar, p_app varchar, p_method varchar )
	RETURNS varchar AS $$

DECLARE
    l_id 				varchar (40);
	l_email_token 		varchar (40);
	l_data				varchar (800);
	l_fail				bool;
	l_token				varchar (40);
	l_found				varchar (50);
	l_customer_id		varchar (40);
	l_bad_token			bool;
	l_from 				varchar (100);
	l_username_is_email varchar (10);
	l_have_email 		varchar (10);
	l_check_bad_pass	varchar (10);

    l_xsrf_token 		varchar (40);
    l_xsrf_mode 		varchar (100);
  	l_auth_token	varchar (40);
    l_seq 			varchar (40);
	l_redir 		varchar (400);
	l_94_days 		varchar(50);
	l_redirect_to		varchar (400);
	l_redirect_to_app	varchar (400);
    l_privs			varchar (400);
	l_config		varchar (7500);
BEGIN

	-- "Copyright (C) Philip Schlump, 2009-2019." 

	l_fail = false;
	l_data = '{"status":"success"}';

	l_customer_id = s_get_customer_id_from_url ( p_url );
	l_from = s_get_config_item( 'from.address', l_customer_id, 'pschlump@yahoo.com' );
	l_username_is_email = s_get_config_item( 'username.is.email', l_customer_id , 'no' );
	if l_username_is_email = 'yes' then
		p_username = p_email;
	end if;
	l_check_bad_pass = s_get_config_item( 'check.bad.password', l_customer_id, 'no' );

	l_id = uuid_generate_v4();
	l_email_token = uuid_generate_v4();

	begin
		select 'yes'
			into l_have_email
			from "t_user"
			where "email_address" = p_email
			;
	exception
		when no_data_found then
			l_have_email = 'no';
	end;
	if l_have_email = 'yes' then
		l_fail = true;
		l_data = '{"status":"error","msg":"Unable to create user with this username.  This email address is already used.","code":"600"}';
	end if;

	if not l_fail then
		if l_check_bad_pass = 'yes' then

			select 'found'
				into l_found
				from t_common_pass
				where password = p_password
			;

			IF FOUND THEN
				l_fail = true;
				l_data = '{"status":"error", "code":"129", "msg":"Invalid password in list of most common passwords, pick a different one."}';
			END IF;

		end if;
	end if;

	-- xyzzy - pull out to own function s_get_redirect_to
	select "value"
		into l_redirect_to
		from "t_config"
		where "item_name" = 'register.redirect.to'
		  and "customer_id" = l_customer_id
		;
		l_redirect_to = p_url || l_redirect_to;
	if not found then
		l_redirect_to = p_url || '/';
	end if;

	select "value"
		into l_redirect_to_app
		from "t_config"
		where "item_name" = 'register.redirect.to.app'
		  and "customer_id" = l_customer_id
		;
	if not found then
		l_redirect_to_app = '';
	end if;

	-- ,	( '1', 'XSRF.token', 'per-user' )	-- or 'progressive-hashed' or 'off'
    l_xsrf_mode = s_get_xsrf_mode(l_customer_id);
	if l_xsrf_mode = 'per-user' or l_xsrf_mode = 'progressive-hashed' then
		l_xsrf_token = uuid_generate_v4();
	else
		l_xsrf_token = 'n/a';
	end if;

	l_privs = '[]';

	if not l_fail then	
		select "t_customer"."config"
			into  l_config
			from "t_customer" as "t_customer" 
			where "id" = l_customer_id
			;
		if not found then
			l_fail = true;
			l_data = '{"status":"error","msg":"Invalid Customer_ID","code":"309"}';
		end if;
	end if;


	l_94_days = s_get_config_item( 'acct.auth_token.expire', l_customer_id, '94 days');

	
	if not l_fail then
		BEGIN
--o			insert into "t_user" ( "id", "username", "password", "ip", "real_name", "email_address"
--o					, "acct_state", "acct_expire", "email_confirmed", "email_reset_key" ) 
--o				values ( l_id, p_username, crypt(p_password,gen_salt('bf',8)), p_ip_addr, p_real_name, p_email
--o					, 'temporary', current_timestamp + interval '30 days', 'n', l_email_token )
--o			;
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
			insert into "t_user" ( "id", "username", "password", "ip", "real_name", "email_address"
					, "acct_state", "acct_expire", "email_confirmed", "email_reset_key" ) 
				values ( l_id, p_username, crypt(p_password,gen_salt('bf',8)), p_ip_addr, p_real_name, p_email
					, 'temporary', current_timestamp + interval '30 days', 'y', null )
			;
		EXCEPTION WHEN unique_violation THEN
			l_fail = true;
			if l_username_is_email = 'yes' then
				l_data = '{"status":"error","msg":"Unable to create user with this email address.  Please choose a different email address.","code":"601"}';
			else
				l_data = '{"status":"error","msg":"Unable to create user with this username.  Please choose a different username (try your email address).","code":"602"}';
			end if;
		END;
	end if;

-- 	if not l_fail then	
-- 		l_auth_token = uuid_generate_v4();
-- 		l_seq = uuid_generate_v4();
-- 		insert into "t_auth_token" (
-- 			  "auth_token"	
-- 			, "user_id"	
-- 			, "expire"
-- 		) values (
-- 			  l_auth_token
-- 			, l_id
-- 			, current_timestamp + l_94_days::interval
-- 		);
-- 		--	, current_timestamp + interval '94 days'
-- 		update "t_user"
-- 			set "email_confirmed" = 'y'
-- 				, "ip" = p_ip_addr
-- 				, "email_reset_key" = null
-- 			where "id" = l_id
-- 			;
-- 
-- 		if p_method = 'GET' then
-- 			l_redir = ', "$redirect_to$":'||to_json(l_redirect_to)
-- 				||', "$redirect_vars$":["auth_token","xsrf_token","redir_to_app"]'
-- 				;
-- 			-- variables - how 
-- 			-- variables - where in code // xyzzy - Redirect to line 4990 in crud.go
-- 		else -- if p_method = "POST" then
-- 			l_redir = '';
-- 		end if;
-- 		l_data = '{"status":"success"'
-- 			||', "auth_token":'||to_json(l_auth_token)
-- 			||', "seq":'||to_json(l_seq)
-- 			||', "privs":'||to_json(l_privs)
-- 			||', "user_id":'||to_json(l_id)
-- 			||', "redir_to_app":'||to_json(l_redirect_to_app)
-- 			||', "customer_id":'||to_json(l_customer_id)
-- 			||', "config":'||to_json(l_config)
-- 			||', "xsrf_token":'||to_json(l_xsrf_token)
-- 			||', "$JWT-claims$":["auth_token"]'
-- 			||', "$session$":{'
-- 				||'"set":['
-- 					||'{"path":["user","$is_logged_in$"],"value":"y"}'
-- 					||',{"path":["user","$xsrf_token$"],"value":'||to_json(l_xsrf_token)||'}'
-- 				||']'
-- 			||'}'
-- 			||l_redir
-- 			||'}';
-- 	end if;
	if not l_fail then
--o		l_data = '{"status":"success"'
--o			||',"$send_email$":{'
--o				||'"template":"confirm_registration"'
--o				||',"username":'||to_json(s_nvl(p_username))
--o				||',"real_name":'||to_json(s_nvl(p_real_name))
--o				||',"email_token":'||to_json(s_nvl(l_email_token))
--o				||',"app":'||to_json(s_nvl(p_app))
--o				||',"url":'||to_json(s_nvl(p_url))
--o				||',"from":'||to_json(s_nvl(l_from))
--o				||',"email_address":'||to_json(s_nvl(p_email))
--o				||',"to":'||to_json(s_nvl(p_email))
--o			||'},"$session$":{'
--o				||'"set":['
--o					||'{"path":["user","$is_logged_in$"],"value":"n"}'
--o				||']'
--o			||'}}';
 		if p_method = 'GET' then
 			l_redir = ', "$redirect_to$":'||to_json(l_redirect_to)
 				||', "$redirect_vars$":["auth_token","xsrf_token","redir_to_app"]'
 				;
 			-- variables - how 
 			-- variables - where in code // xyzzy - Redirect to line 4990 in crud.go
 		else -- if p_method = "POST" then
 			l_redir = '';
 		end if;
 		l_data = '{"status":"success"'
 			||', "auth_token":'||to_json(l_auth_token)
 			||', "seq":'||to_json(l_seq)
			||', "privs":'||to_json(l_privs)
 			||', "user_id":'||to_json(l_id)
			||', "redir_to_app":'||to_json(l_redirect_to_app)
 			||', "customer_id":'||to_json(l_customer_id)
			||', "config":'||to_json(l_config)
			||', "xsrf_token":'||to_json(l_xsrf_token)
 			||', "$JWT-claims$":["auth_token"]'
 			||', "$session$":{'
 				||'"set":['
 					||'{"path":["user","$is_logged_in$"],"value":"y"}'
 					||',{"path":["user","$xsrf_token$"],"value":'||to_json(l_xsrf_token)||'}'
 				||']'
 			||'}'
 			||l_redir
 			||'}';
	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

