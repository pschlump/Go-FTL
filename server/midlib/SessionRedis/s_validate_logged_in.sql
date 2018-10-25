





-- "PostgresQuery":	{ "type":["string"], "default":"select s_validate_logged_in( $1 )" },


-- "Copyright (C) Philip Schlump, 2009-2017." 

drop FUNCTION s_validate_logged_in ( p_auth_token varchar );

CREATE or REPLACE FUNCTION s_validate_logged_in ( p_auth_token varchar )
	RETURNS varchar AS $$
DECLARE
	l_data				varchar (100);
	l_user_id			varchar (40);
	l_username			varchar (200);
	l_fail				boolean;
BEGIN

	-- "Copyright (C) Philip Schlump, 2009-2017." 

	l_fail = false;
	l_data = '{"status":"error","code":"910","msg":"Error: unknown error."}';

	if not l_fail then
		select  "user_id"
			into  l_user_id
			from "t_auth_token" 
			where "auth_token" = p_auth_token
			  and "expire" >= current_timestamp
			;
		if not found then
			l_data = '{"status":"error", "code":"910", "msg":"Error: Not Authorized."}';
			l_fail = true;
		end if;
	end if;

	if not found then
		select  "username"
			into  l_username
			from "t_user" 
			where "id" = l_user_id
			  and "email_confirmed" = 'y'
			  and "acct_state" <> 'locked' and "acct_state" <> 'billing' and "acct_state" <> 'closed' and "acct_state" <> 'pass-reset'
			  and "acct_expire" >= now() 
			  and "n_login_fail" <= 5
			;

		if not found then
			l_data = '{ "status":"error", "code":"911", "msg":"Not logged in." }';
		else 
			--l_data = '{"status":"success",'
			--	||'"user_id":'||to_json(l_user_id)
			--	||'}';
			l_data = '{"status":"success"}';
		end if;
	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

