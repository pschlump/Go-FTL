
-- "Copyright (C) Philip Schlump, 2009-2017." 

drop FUNCTION s_change_password ( p_password varchar, p_again varchar, p_old_password varchar, p_auth_token varchar, p_ip_addr varchar, p_url varchar );

CREATE or REPLACE FUNCTION s_change_password ( p_password varchar, p_again varchar, p_old_password varchar, p_auth_token varchar, p_ip_addr varchar, p_url varchar )
	RETURNS varchar AS $$
DECLARE
	l_data			varchar (800);
	l_username		varchar (40);
	l_privs			varchar (400);
	l_id			varchar (40);
	l_customer_id	varchar (40);
	l_real_name		varchar (200);
	l_to			varchar (200);
	l_from 			varchar (100);
	l_found 		varchar (10);
	l_fail 			boolean;
	l_auth_token 	varchar(40);
	l_user_id		varchar(40);
	l_log_id 		bigint;
	l_94_days 		varchar(50);
    l_xsrf_token 		varchar (40);
    l_xsrf_mode 		varchar (100);
	l_check_bad_pass	varchar (10);
BEGIN

	-- "Copyright (C) Philip Schlump, 2009-2017." 

	l_fail = false;
	l_data = '{"status":"success"}';
	l_log_id = 1;

	select nextval('t_email_id_seq'::regclass) 
		into l_log_id;

	if p_password != p_again then
		l_fail = true;
		l_data = '{"status":"error", "code":"103", "msg":"Passwords did not match."}';
	end if;

	if s_ip_ban(p_ip_addr) then
		l_data = '{"status":"error", "code":"102", "msg":"Invalid username or password."}';
		l_fail = true;
	end if;

	if not l_fail then
		select  "user_id"
			into  l_user_id
			from "t_auth_token" 
			where "auth_token" = p_auth_token
			  and "expire" >= current_timestamp
			;
		if not found then
			l_data = '{"status":"error", "code":"101", "msg":"Error: Not Authorized."}';
			l_fail = true;
		end if;
	end if;

	if not l_fail then	
		select "id"
				, "privs"
				, "customer_id"
				, "username"
				, "real_name"
				, "email_address"
			into l_id
				, l_privs
				, l_customer_id
				, l_username
				, l_real_name
				, l_to
			from "t_user"
			where "id" = l_user_id
			  and "password" = crypt(p_old_password, "password")
			;
		IF NOT FOUND THEN
			l_fail = true;
			l_data = '{"status":"error", "code":"100", "msg":"Invalid Token or Expired Token"}';
		END IF;
	end if;

	l_check_bad_pass = s_get_config_item( 'check.bad.password', l_customer_id, 'no' );
	l_from = s_get_config_item( 'from.address' , l_customer_id, 'pschlump@yahoo.com' );
	l_94_days = s_get_config_item( 'acct.auth_token.expire' , l_customer_id, '94 days');

    l_xsrf_mode = s_get_xsrf_mode(l_customer_id);
	if l_xsrf_mode = 'per-user' or l_xsrf_mode = 'progressive-hashed' then
		l_xsrf_token = uuid_generate_v4();
	else
		l_xsrf_token = 'n/a';
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

	if not l_fail then	

		delete from "t_auth_token" where "user_id" = l_id;	-- delete all otehr auth tokens - invalidates current logins.
		l_auth_token = uuid_generate_v4();
		insert into "t_auth_token" (
			  "auth_token"	
			, "user_id"	
			, "expire"
		) values (
			  l_auth_token
			, l_id
			, current_timestamp + l_94_days::interval
		);
		--	, current_timestamp + interval '94 days'

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
			||', "auth_token":'||to_json(l_auth_token)
			||', "privs":'||to_json(l_privs)
			||', "user_id":'||to_json(l_id)
			||', "customer_id":'||to_json(l_customer_id)
			||', "username":'||to_json(l_username)
			||', "xsrf_token":'||to_json(l_xsrf_token)
			||', "$JWT-claims$":["auth_token"]'
			||',"$send_email$":{'
				||'"template":"password_changed"'
				||',"username":'||to_json(l_username)
				||',"real_name":'||to_json(l_real_name)
				||',"url":'||to_json(p_url)
				||',"from":'||to_json(l_from)
				||',"to":'||to_json(l_to)
				||',"log_id":"'||to_json(l_log_id)||'"'
			||'}'
			||', "$session$":{'
				||'"set":['
					||' {"path":["user","$xsrf_token$"],"value":'||to_json(l_xsrf_token)||'}'
				||']'
			||'}'
			||'}' ;

	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

