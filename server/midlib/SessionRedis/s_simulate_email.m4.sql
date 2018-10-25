
-- "Copyright (C) Philip Schlump, 2009-2017." 

drop FUNCTION s_simulate_email(p_tn varchar, p_email_token varchar, p_ip_addr varchar, p_url varchar, p_app varchar, p_kp varchar );

--	,"/api/session/simulate_email": { "g": "s_simulate_email", "p": [ "tn", "email_token", "$ip$", "$url$", "app" ], "nokey":true
CREATE or REPLACE FUNCTION s_simulate_email(p_tn varchar, p_email_token varchar, p_ip_addr varchar, p_url varchar, p_app varchar, p_kp varchar )
	RETURNS varchar AS $$

DECLARE
	l_email_token 		varchar (40);
	l_data				varchar (800);
	l_fail				bool;
	l_customer_id		varchar (40);
	l_from 				varchar (100);
	l_username 			varchar(100);
	l_real_name 		varchar(100);
	l_email 			varchar(100);
	l_log_id 			bigint;
BEGIN

	-- "Copyright (C) Philip Schlump, 2009-2017." 

	l_fail = false;
	l_data = '{"status":"success"}';
	l_log_id = 1;

	-- CREATE SEQUENCE t_email_id_seq
	--	, "host_no"			bigint DEFAULT nextval('t_host_id_seq'::regclass) NOT NULL 
	select nextval('t_email_id_seq'::regclass) 
		into l_log_id;

	if p_app is null or p_app = '' then
		p_app = 'test-app';
	end if;

	select "customer_id"
		into l_customer_id
		from "t_host_to_customer"
		where "host_name" = p_url
		limit 1
		;
	if not found then
		l_customer_id = 'error-missing-host-to-customer';
	end if;

	select "value"
		into l_from
		from "t_config"
		where "item_name" = 'from.address'
		  and "customer_id" = l_customer_id
		;
	if not found then
		l_from = 'pschlump@yahoo.com';
	end if;

	l_email_token = p_email_token;
	if p_email_token is null or p_email_token = '' then
		l_email_token = uuid_generate_v4();
	end if;

	if not ( p_tn = 'confirm_registration' ) then
		l_fail = true;
		l_data = '{"status":"error","code":"700","msg":"unknown email template:"'||p_tn||'}';
	end if;

	l_username = 'kermitfrog';
	l_real_name = 'Kermit The Frog';
	l_email = 'kermit.frog@the-green-pc.com';
	if p_kp = 'p' then
		l_username = 'mispiggy';
		l_real_name = 'Mis Piggy';
		l_email = 'mis_piggy.frog@the-green-pc.com';
	end if;

	if not l_fail then
		l_data = '{"status":"success"'
			||',"$send_email$":{'
				||'"template":'||to_json(p_tn)
				||',"username":'||to_json(l_username)
				||',"real_name":'||to_json(l_real_name)
				||',"email_token":'||to_json(l_email_token)
				||',"app":'||to_json(p_app)
				||',"url":'||to_json(p_url)
				||',"from":'||to_json(l_from)
				||',"email_address":'||to_json(l_email)
				||',"to":'||to_json(l_email)
				||',"log_id":"'||to_json(l_log_id)||'"'
			||'}}';
		-- xyzzy - add in stuff to send email
	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

