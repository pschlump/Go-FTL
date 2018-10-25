





-- "Copyright (C) Philip Schlump, 2009-2017." 

drop FUNCTION s_echo_builtin(p_ip_addr varchar, p_url varchar, p_host varchar, p_top varchar);
drop FUNCTION s_echo_builtin(p_ip_addr varchar, p_url varchar, p_host varchar, p_top varchar, p_session varchar);

-- Anticipate that this will chagne to add in additional builtin stuff

-- Sample Output at level of Go-FTL server:
-- 	{
--		"status":"success", 
--		"ip":"127.0.0.1", 
--		"url":"http://localhost:9001", 
--		"host":"localhost:9001", 
--		"from":"pschlump@gmail.com", 
--		"ip_is_baned":"no", 
--		"customer_id":"1"
--	}

CREATE or REPLACE FUNCTION s_echo_builtin(p_ip_addr varchar, p_url varchar, p_host varchar, p_top varchar, p_session varchar)
	RETURNS varchar AS $$
DECLARE
    l_data 		varchar (8000);
	l_from 		varchar (100);
	l_junk		varchar (1);
	l_ip_ban_s	varchar (10);
    l_customer_id 		varchar (40);
    l_status 		varchar (40);
    l_code 		varchar (40);
BEGIN

	-- "Copyright (C) Philip Schlump, 2009-2017." 

    l_status = 'success';
    l_code = '';

	select "customer_id"
		into l_customer_id
		from "t_host_to_customer"
		where "host_name" = p_url
		limit 1
		;
	if not found then
		l_customer_id = 'error-missing-host-to-customer';
		l_status = 'error';
		l_code = ', "code":"300"';
	end if;
	if l_customer_id is null then
		l_customer_id = 'error-invalid-host-to-customer';
		l_status = 'error';
		l_code = ', "code":"301"';
	end if;

	select "value"
		into l_from
		from "t_config"
		where "item_name" = 'from.address'
		  and "customer_id" = l_customer_id
		;
	if not found then
		l_from = 'error-missing-config@gmail.com';
		l_status = 'error';
		l_code = ', "code":"302"';
	end if;
	if l_from is null then
		l_from = 'error-invalid-config@gmail.com';
		l_status = 'error';
		l_code = ', "code":"303"';
	end if;

	select 'y' 
		into l_junk
		from "t_ip_ban"
		where "ip" = p_ip_addr
		;

	if not found then
		l_ip_ban_s = 'no';
	else
		l_ip_ban_s = 'yes';
		l_status = 'error';
		l_code = ', "code":"304"';
	end if;

	l_data = '{ "status":'||to_json(l_status)
		||l_code
		||', "ip":'||to_json(p_ip_addr)
		||', "url":'||to_json(p_url)
		||', "top":'||to_json(p_top)
		||', "host":'||to_json(p_host)
		||', "from":'||to_json(l_from)
		||', "ip_is_baned":'||to_json(l_ip_ban_s)
		||', "customer_id":'||to_json(l_customer_id)
		||', "session":'||to_json(p_session)
		||' }' ;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;


