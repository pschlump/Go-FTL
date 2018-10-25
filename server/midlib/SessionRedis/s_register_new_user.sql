





-- "Copyright (C) Philip Schlump, 2009-2017." 

drop FUNCTION s_register_new_user(p_username varchar, p_password varchar, p_again varchar, p_ip_addr varchar, p_email varchar, p_real_name varchar, p_url varchar, p_app varchar );
CREATE or REPLACE FUNCTION s_register_new_user(p_username varchar, p_password varchar, p_again varchar, p_ip_addr varchar, p_email varchar, p_real_name varchar, p_url varchar, p_app varchar )
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
BEGIN

	-- "Copyright (C) Philip Schlump, 2009-2017." 

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

	if not l_fail then
		BEGIN
			insert into "t_user" ( "id", "username", "password", "ip", "real_name", "email_address"
					, "acct_state", "acct_expire", "email_confirmed", "email_reset_key" ) 
				values ( l_id, p_username, crypt(p_password,gen_salt('bf',8)), p_ip_addr, p_real_name, p_email
					, 'temporary', current_timestamp + interval '30 days', 'n', l_email_token )
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

	if not l_fail then
		l_data = '{"status":"success"'
			||',"$send_email$":{'
				||'"template":"confirm_registration"'
				||',"username":'||to_json(s_nvl(p_username))
				||',"real_name":'||to_json(s_nvl(p_real_name))
				||',"email_token":'||to_json(s_nvl(l_email_token))
				||',"app":'||to_json(s_nvl(p_app))
				||',"url":'||to_json(s_nvl(p_url))
				||',"from":'||to_json(s_nvl(l_from))
				||',"email_address":'||to_json(s_nvl(p_email))
				||',"to":'||to_json(s_nvl(p_email))
			||'},"$session$":{'
				||'"set":['
					||'{"path":["user","$is_logged_in$"],"value":"n"}'
				||']'
			||'}}';
	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

