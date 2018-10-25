

-- "Copyright (C) Philip Schlump, 2009-2017." 

drop FUNCTION s_get_username_from_email(p_email_address varchar, p_ip_addr varchar, p_url varchar);

--	,"/api/session/get_username_from_email": { "g": "s_get_username_from_email", "p": [ "username", "$url$" ]
CREATE or REPLACE FUNCTION s_get_username_from_email(p_email_address varchar, p_ip_addr varchar, p_url varchar)
	RETURNS varchar AS $$

DECLARE
    l_id 				varchar (40);
	l_email_token 		varchar (40);
	l_data				varchar (800);
	l_fail				bool;
	l_token				varchar (40);
	l_bad_token			bool;

	l_customer_id		varchar (40);
	l_username 			varchar (100);
	l_username_is_email varchar (10);
BEGIN

	-- "Copyright (C) Philip Schlump, 2009-2017." 

	l_fail = false;
	l_data = '{"status":"success"}';

	if s_ip_ban(p_ip_addr) then
		l_data = '{ "status":"failed", "code":"009", "msg":"Invalid username or password." }';
		l_fail = true;
	end if;

	if not l_fail then 
		l_customer_id = s_get_customer_id_from_url ( p_url );
		l_username_is_email = s_get_config_item( 'username.is.email', l_customer_id , 'no' );
		if l_username_is_email = 'yes' then
			l_data = '{"status":"success","msg":"username is email address.","$sleep$":2,"username":'||to_json(s_nvl(p_email_address))||'}';
		else
			begin
				select "username"
					into l_username
					from "t_user"
					where "email_address" = p_email_address
					;
			exception
				when no_data_found then
					l_fail = true;
					l_data = '{"status":"error","msg":"email address is invalid.","$sleep$":2}';
			end;
			if not l_fail then
				l_data = '{"status":"success","$sleep$":2,"username":'||to_json(s_nvl(l_username))||'}';
			end if;
		end if;
	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

