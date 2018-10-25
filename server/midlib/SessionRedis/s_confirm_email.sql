




--
-- Description: Confirm a user from an email link.  This is in response to a "GET" request. 
--
-- The form/"POST" request is different.  "POST" skips the redirect because it is from a form in the applicaiton.
--
-- Questions:
-- 0. Perform login at this point in time?
-- 0. Is this different than confirm a user from a token cut/paste into app? -- Should the "app" just do a "POST" to this? v.s. a "GET" when link?
--		Should "$method$" be bassed to this? -- GET is link, POST is form.
-- 1. Do we need to store the xsrf_token in cookie/local-storage - so it will persis across browser restar/bad-network?
-- 1. If we redirect - how to pas xsrf_token back to redirected location (URL?/Cookie?)
-- 2. If we need to pass the "auth_token" and CORS how?
-- 3. $log_it_error$ -- post processing - send message to log 
-- 3. $log_it_info$ -- post processing - send message to log 
-- 3. $log_it_warn$ -- post processing - send message to log
--
-- "Copyright (C) Philip Schlump, 2009-2017." 
--

drop FUNCTION s_confirm_email ( p_email_auth_token varchar, p_ip_addr varchar, p_url varchar );
drop FUNCTION s_confirm_email ( p_email_auth_token varchar, p_ip_addr varchar, p_url varchar, p_method varchar );

CREATE or REPLACE FUNCTION s_confirm_email ( p_email_auth_token varchar, p_ip_addr varchar, p_url varchar, p_method varchar )
	RETURNS varchar AS $$
DECLARE
	l_data				varchar(800);
	l_customer_id		varchar(40);
	l_fail 				boolean;
	l_redirect_to		varchar (400);
	l_redirect_to_app	varchar (400);
    l_xsrf_token 		varchar (40);
    l_xsrf_mode 		varchar (100);

  	l_auth_token	varchar (40);
    l_seq 			varchar (40);
    l_id 			varchar (40);
    l_privs			varchar (400);
	l_config		varchar (7500);

	l_redir 		varchar (400);
	l_94_days 		varchar(50);
BEGIN

	-- "Copyright (C) Philip Schlump, 2009-2017." 

	l_fail = false;
	l_data = '{"status":"success"}';

	if p_email_auth_token is null or p_email_auth_token = '' then
		l_data = '{ "status":"error", "code":"200", "msg":"Invalid p_email_auth_token." }';
		l_fail = true;
	end if;

	if s_ip_ban(p_ip_addr) then
		l_data = '{ "status":"error", "code":"201", "msg":"Invalid username or password." }';
		l_fail = true;
	end if;

	-- xyzzy - could combine with user lookup - and skip a query
	-- xyzzy - pull out to own function s_url_to_customer_id	
	select "customer_id"
		into l_customer_id
		from "t_host_to_customer"
		where "host_name" = p_url
		limit 1
		;
	if not found then
		l_customer_id = 'error-missing-host-to-customer';
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


	if not l_fail then	
		select "t_user"."id"
			, "t_user"."privs"
			, "t_user"."customer_id"
			, "t_customer"."config"
			into l_id
			, l_privs 	
			, l_customer_id 	
			, l_config
			from "t_user" as "t_user" left join "t_customer" as "t_customer" on "t_customer"."id" = "t_user"."customer_id"
			where "email_reset_key" = p_email_auth_token
			;
		if not found then
			l_fail = true;
			l_data = '{"status":"error","msg":"Invalid Token or Expired Token","code":"202","email_auth_token":'||to_json(p_email_auth_token)||'}';
		end if;
	end if;
	
	l_94_days = s_get_config_item( 'acct.auth_token.expire', l_customer_id, '94 days');

	-- ,	( '1', 'XSRF.token', 'per-user' )	-- or 'progressive-hashed' or 'off'
    l_xsrf_mode = s_get_xsrf_mode(l_customer_id);
	if l_xsrf_mode = 'per-user' or l_xsrf_mode = 'progressive-hashed' then
		l_xsrf_token = uuid_generate_v4();
	else
		l_xsrf_token = 'n/a';
	end if;

	if not l_fail then	
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
		--	, current_timestamp + interval '94 days'
		update "t_user"
			set "email_confirmed" = 'y'
				, "ip" = p_ip_addr
				, "email_reset_key" = null
			where "id" = l_id
			;

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
