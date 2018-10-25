

-- "Copyright (C) Philip Schlump, 2009-2017." 

drop FUNCTION s_login(p_username varchar, p_password varchar, p_ip_addr varchar, p_host varchar);

CREATE or REPLACE FUNCTION s_login(p_username varchar, p_password varchar, p_ip_addr varchar, p_host varchar)
	RETURNS varchar AS $$
DECLARE
    l_id 				varchar (40);
  	l_auth_token		varchar (40);
  	l_ip				varchar (40);
	l_email_confirmed	varchar (1);
	l_acct_state 		varchar (10);
	l_acct_expire		timestamp;
	l_n_login_fail		int;
	l_login_fail_delay	timestamp;
	l_last_login 		timestamp ;		 													-- 
    l_customer_id 		varchar (40);
    l_xsrf_token 		varchar (40);
    l_xsrf_mode 		varchar (100);
	l_94_days 		varchar(50);

    l_data 		varchar (8000);
    l_token 	varchar (40);
    l_ctoken 	varchar (40);
	l_fail 		boolean;
	l_bad_token	boolean;
    l_seq 		varchar (40);
    l_privs		varchar (400);
	l_config	varchar (7500);

BEGIN

	-- "Copyright (C) Philip Schlump, 2009-2017." 

	l_fail = false;
	l_id = null;
	l_data = '{ "status":"unknown"}';

	if s_ip_ban(p_ip_addr) then
		l_data = '{ "status":"error", "code":"400", "msg":"Invalid username or password." }';
		l_fail = true;
	end if;


	if not l_fail then
		select  "t_user"."id"
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
			  and "t_user"."password" = crypt(p_password, "t_user"."password")
			;

		if not found then
			l_data = '{ "status":"error", "code":"401", "msg":"Invalid username or password." }';
			l_fail = true;
		end if;
	end if;

	if l_privs is null then
		l_privs = '';
	end if;

	if not l_fail then
		if l_email_confirmed = 'n' then
			l_data = '{ "status":"error", "code":"402", "msg":"Before you login you have to confirm your email account.." }';
			l_fail = true;
		end if;
	end if;

	if not l_fail then
		if l_acct_state = 'locked' or l_acct_state = 'billing' or l_acct_state = 'closed' then
			l_data = '{ "status":"error", "code":"403", "msg":"Account is no longer valid.", "acct_state":"'||l_acct_state||'" }';
			l_fail = true;
		end if;
	end if;

	if not l_fail then
		if l_acct_state = 'pass-reset' then
			l_data = '{ "status":"error", "code":"404", "msg":"You must reset your password before you can login." }';
			l_fail = true;
		end if;
	end if;

	if not l_fail then
		if l_n_login_fail > 5 then
			if l_login_fail_delay + interval ' 120 seconds ' < now() then
				l_fail = false;
			else
				l_acct_state = 'temporary';
				l_data = '{ "status":"error", "code":"405", "msg":"Too many failed login attempts.  Please wate 120 seconds and try again." }';
				l_fail = true;
			end if;
		end if;
	end if;

	if not l_fail then
		if l_acct_expire < now() then
			l_data = '{ "status":"error", "code":"406", "msg":"Account is no longer valid.  Your trial period has ended." }';
			l_fail = true;
		end if;
	end if;
			
	l_94_days = s_get_config_item( 'acct.auth_token.expire' , l_customer_id, '94 days');

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
		update "t_user" set
				  "acct_state" = l_acct_state
				, "ip" = p_ip_addr
				, "last_login" = current_timestamp
				, "n_login_fail" = 0
				, "login_fail_delay" = null
			where "id" = l_id
			;
		l_data = '{ "status":"success"'
			||', "auth_token":'||to_json(l_auth_token)
			||', "seq":'||to_json(l_seq)
			||', "privs":'||to_json(l_privs)
			||', "user_id":'||to_json(l_id)
			||', "customer_id":'||to_json(l_customer_id)
			||', "config":'||to_json(l_config)
			||', "xsrf_token":'||to_json(l_xsrf_token)
			||', "$JWT-claims$":["auth_token"]'
			||', "$session$":{'
				||'"set":['
					|| '{"path":["user","$is_logged_in$"],"value":"y"}'
					||',{"path":["user","$xsrf_token$"],"value":'||to_json(l_xsrf_token)||'}'
				||']'
			||'}}';
		--insert into "t_output" ( msg ) values ( '222: l_data='||l_data );
	else
		if l_id is not null then
			update "t_user" set
				"acct_state" = l_acct_state,
				"ip" = p_ip_addr,
				"login_fail_delay" = current_timestamp,
				"n_login_fail" = "n_login_fail" + 1
			where "id" = l_id;
		end if;
	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

