
--
-- Validate and update a auth_token.  Calling this will, if successful, validate an 
-- auth_token.   Also if the token is timing out it will update the timeout time
-- and return a new - xsrf_token for this session.
--
--	,"/api/session/validate_auth_token": { "g": "s_validate_auth_token", "p": [ "auth_token" ]
--
-- "Copyright (C) Philip Schlump, 2009-2017." 
--

drop FUNCTION s_validate_auth_token ( p_auth_token varchar, p_url varchar );
CREATE or REPLACE FUNCTION s_validate_auth_token ( p_auth_token varchar, p_url varchar )
	RETURNS varchar AS $$
DECLARE
	l_data				varchar (800);
	l_username			varchar (200);
	l_user_id			varchar (40);
	l_email_confirmed	varchar (1);
	l_fail 				boolean;
	l_is_localhost		varchar (10);
	l_xsrf_token 		varchar (40);
	l_xsrf_mode 		varchar (100);
	l_customer_id		varchar (40);
	l_ttl_timeout 		int;
	l_94_days 			varchar(50);
BEGIN

	-- "Copyright (C) Philip Schlump, 2009-2017." 

	l_data = '{"status":"error","msg":"Error: unknown error."}';
	l_fail = false;

	select "is_localhost"
		into l_is_localhost
		from "t_host_to_customer"
		where "host_name" = p_url
		;
	if not found then
		l_data = '{"status":"error", "code":"900", "msg":"Error: Not Authorized."}';
		l_fail = true;
	end if;
	if l_is_localhost = 'no' then
		l_data = '{"status":"error", "code":"901", "msg":"Error: Not Authorized."}';
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
			l_data = '{"status":"error", "code":"902", "msg":"Error: Not Authorized."}';
			l_fail = true;
		end if;
	end if;

	l_customer_id = s_get_customer_id_from_url ( p_url );
	l_ttl_timeout = s_get_config_item_int('ttl.user.auth_token', l_customer_id, 180);
	l_94_days = s_get_config_item( 'acct.auth_token.expire' , l_customer_id, '94 days');
	l_xsrf_mode = s_get_xsrf_mode(l_customer_id);
	if l_xsrf_mode = 'per-user' or l_xsrf_mode = 'progressive-hashed' then
		l_xsrf_token = uuid_generate_v4();
	else
		l_xsrf_token = 'n/a';
	end if;

	if not l_fail then

		update "t_auth_token" 
			set "expire" = current_timestamp + l_94_days::interval
			where "auth_token" = p_auth_token
			;
		-- set "expire" = current_timestamp + interval '94 days'

		select  "username", "email_confirmed"
			into  l_username, l_email_confirmed	
			from "t_user" 
			where "id" = l_user_id
			;

		if not found then
			l_data = '{ "status":"failed", "code":"903", "msg":"Invalid auth_token." }';
		else 
			l_data = '{"status":"success"'
				||',"username":'||to_json(l_username)
				||',"xsrf_token":'||to_json(l_xsrf_token)
				||',"email_confirmed":'||to_json(l_email_confirmed)
				||',"ttl":'||to_json(l_ttl_timeout)
				||',"user_id":'||to_json(l_user_id)
				||', "$session$":{'
					||'"set":['
						||'{"path":["user","$is_logged_in$"],"value":"y"}'
						||',{"path":["user","$xsrf_token$"],"value":'||to_json(l_xsrf_token)||'}'
					||']'
				||'}}';
		end if;
	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

