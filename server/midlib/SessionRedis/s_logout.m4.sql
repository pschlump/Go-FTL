
-- "Copyright (C) Philip Schlump, 2009-2017." 

drop FUNCTION s_logout(p_auth_token varchar, p_ip_addr varchar);

CREATE or REPLACE FUNCTION s_logout(p_auth_token varchar, p_ip_addr varchar)
	RETURNS varchar AS $$
DECLARE
	l_user_id		varchar(40);
	l_fail 			boolean;
BEGIN

	-- "Copyright (C) Philip Schlump, 2009-2017." 

	if not l_fail then
		select  "user_id"
			into  l_user_id
			from "t_auth_token" 
			where "auth_token" = p_auth_token
			;
		if not found then
			l_fail = true;
		end if;
	end if;

	-- Cleanup expired auth tokens
	delete from "t_auth_token"
		where "expire" < current_timestamp - interval '1 days'
		;

	-- Delete this users auth token
	delete from "t_auth_token"
		where "auth_token" = p_auth_token
		;

	if not l_fail then
		update "t_user"
			set "ip" = p_ip_addr
			  , "n_login_fail" = 0
			  , "login_fail_delay" = null
			where "id" = l_user_id
			;
	end if;

	RETURN '{"status":"success"}';
END;
$$ LANGUAGE plpgsql;

