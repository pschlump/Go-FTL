





-- "Copyright (C) Philip Schlump, 2009-2017." 


drop FUNCTION s_validate_token ( p_password varchar, p_auth_token varchar );

CREATE or REPLACE FUNCTION s_validate_token ( p_password varchar, p_auth_token varchar )
	RETURNS varchar AS $$
DECLARE
	l_data			varchar (100);
	l_id			varchar (40);
	l_fail 			boolean;
	l_user_id		varchar(40);
BEGIN
	l_fail = false;
	l_data = '{"status":"success"}';

	if not l_fail then
		select  "user_id"
			into  l_user_id
			from "t_auth_token" 
			where "auth_token" = p_auth_token
			  and "expire" >= current_timestamp
			;
		if not found then
			l_data = '{"status":"error", "code":"920", "msg":"Invalid Token"}';
			l_fail = true;
		end if;
	end if;

	if not l_fail then	
		select "id"
			into l_id
			from "t_user"
			where "id" = l_user_id
			  and "password" = crypt(p_password, "password")
			;
		if not found then
			l_fail = true;
			l_data = '{"status":"error", "code":"921", "msg":"Invalid Password"}';
		end if;
	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

