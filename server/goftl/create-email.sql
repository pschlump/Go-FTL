
CREATE TABLE "t_email_tab" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "email_group"			text
	, "person_name"			text
	, "email_addr"			text
	, "message_body"		text
	, "ip_addr"				text
	, "created" 			timestamp default current_timestamp not null 						
);




--	err := sizlib.Run1(db, `select save_email_in_db ( $1, $2, %3, $4, $5 ) as "x"`, group, person_name, email_addr, message_body, IPaddr)
--	,"/api/test/save_customer_config": { "g": "test_save_customer_config($1)"
CREATE or REPLACE FUNCTION save_email_in_db (p_group varchar, p_person_name varchar, p_email_addr varchar, p_message_body varchar, p_ip_addr varchar)
	RETURNS varchar AS $$
DECLARE
	l_rv			varchar (400);
BEGIN
	insert into "t_email_tab" (
		"email_group" , "person_name" , "email_addr" , "message_body" , "ip_addr"
	) values (
		p_group , p_person_name , p_email_addr , p_message_body , p_ip_addr 
	);
	RETURN '{"status":"success"}';
END;
$$ LANGUAGE plpgsql;





drop TABLE "t_reg_beta" ;
CREATE TABLE "t_reg_beta" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "prod_name"			text
	, "email_addr"			text
	, "ip_addr"				text
	, "state"				char varying ( 40 ) default 'ok'
	, "created" 			timestamp default current_timestamp not null 						
);

drop TABLE "t_log" ;
CREATE TABLE "t_log" (
	  "id"			char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "created_by"	char varying ( 100 )
	, "msg"			text
	, "created" 	timestamp default current_timestamp not null 						
);


CREATE or REPLACE FUNCTION save_reg_beta ( p_prod_name varchar, p_email_addr varchar, p_ip_addr varchar)
	RETURNS varchar AS $$
DECLARE
	l_rv				varchar (400);
	l_rx				varchar (400);
	l_password			varchar (40);
	l_auth_token		varchar (40);
	l_email_auth_token	varchar (40);
	l_fail 				boolean;
	l_permission		varchar (40);
BEGIN
	l_fail = false;
	l_password = uuid_generate_v4();
	l_rv = '{"status":"error"}';

	l_permission = uuid_generate_v4();
	insert into "t_permission" ( "permission_id" ) values ( l_permission );

	select "auth_token"
		into l_auth_token
		from "t_user"
		where "username" = p_email_addr
		;
	IF FOUND THEN
		l_rv = '{"status":"error","msg":"The email address is already in use.  Please try a different one or reset your password and login."}';
		l_fail = true;
	END IF;

	if not l_fail then
		insert into "t_reg_beta" ( "prod_name", "email_addr", "ip_addr") values ( p_prod_name, p_email_addr, p_ip_addr );

		l_rx = test_register_new_user(p_email_addr, l_password, p_ip_addr, p_email_addr, p_email_addr, 'http://www.2c-why.com/cp-tool/', '42', 'content-pusher', 'content-pusher');

		select "auth_token", "email_reset_key" 
			into l_auth_token, l_email_auth_token
			from "t_user"
			where "username" = p_email_addr
			;
		IF NOT FOUND THEN
			insert into "t_log" ( "created_by", "msg" ) values ( 'save_reg_beta', 'Error: failed to create user: '||l_rv );
			l_rv = '{"status":"error","msg":"Failed to create an account -- most unfortunate."}';
			l_fail = true;
		END IF;

		if not l_fail then
			l_rv = '{"status":"success","pw":"'||l_password||'","auth_token":"'||l_auth_token||'","email_auth_token":"'||l_email_auth_token||'","permission":"'||l_permission||'"}';
		end if;

	end if;

	return l_rv;
END;
$$ LANGUAGE plpgsql;

delete from "t_user" where "username" = 'a2@b.co';
delete from "t_reg_beta" where "email_addr" = 'a2@b.co';
select save_reg_beta ( 'content-pusher' , 'a2@b.co', '127.0.0.1' );





alter table "t_reg_beta" add column "state"		char varying (40) default 'ok' ;

CREATE or REPLACE FUNCTION cp_unsubscribe ( p_email_addr varchar, p_auth_token varchar)
	RETURNS varchar AS $$
DECLARE
	l_rv				varchar (400);
	l_junk				varchar (40);
	l_id				varchar (40);
	l_fail 				boolean;
BEGIN
	l_fail = false;
	l_rv = '{"status":"success"}';


	select "id"
		into l_id
		from "t_user"
		where "username" = p_email_addr
		  and "auth_token" = p_auth_token
		;
	IF NOT FOUND THEN
		l_rv = '{"status":"error","msg":"Failed to unsubscribe - not a valid login and email address - most unfortunate."}';
		l_fail = true;
	END IF;

	if not l_fail then
		l_junk = uuid_generate_v4();
		update "t_user"
			set "auth_token" = l_junk
			  , "acct_state" = 'closed'
			  , "email_addr" = 'closed..'||"email_addr"
			where "id" = l_id
			;
		update "t_reg_beta"
			set "state" = 'closed'
			  , "email_addr" = 'closed..'||"email_addr"
			where "user_id" = l_id
			;
	end if;

	return l_rv;
END;
$$ LANGUAGE plpgsql;


CREATE or REPLACE FUNCTION cp_validate_auth_token ( p_email_addr varchar, p_auth_token varchar)
	RETURNS varchar AS $$
DECLARE
	l_rv				varchar (400);
	l_id				varchar (40);
	l_fail 				boolean;
BEGIN
	l_fail = false;
	l_rv = '{"status":"success"}';

	select "id"
		into l_id
		from "t_user"
		where "username" = p_email_addr
		  and "auth_token" = p_auth_token
		;
	IF NOT FOUND THEN
		l_rv = '{"status":"error","msg":"Failed to validate - not a valid login and email address - most unfortunate."}';
		l_fail = true;
	END IF;

	return l_rv;
END;
$$ LANGUAGE plpgsql;



drop TABLE "t_permission" ;
CREATE TABLE "t_permission" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "permission_id"		char varying (40)
	, "timeout"	 			timestamp default current_timestamp + interval '1 hour' not null 	-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);
create index "t_permission_p1" on "t_permission" ( "permission_id", "timeout" );
create index "t_permission_p2" on "t_permission" ( "created" );

CREATE or REPLACE FUNCTION cp_get_permission ( )
	RETURNS varchar AS $$
DECLARE
	l_rv				varchar (400);
	l_id				varchar (40);
BEGIN
	l_id = uuid_generate_v4();
	l_rv = '{"status":"success","permission":"'||l_id||'"}';
	insert into "t_permission" ( "permission_id" ) values ( l_id );
	return l_rv;
END;
$$ LANGUAGE plpgsql;


CREATE or REPLACE FUNCTION cp_validate_auth_token_cli ( p_email_addr varchar, p_auth_token varchar, p_permission varchar )
	RETURNS varchar AS $$
DECLARE
	l_rv				varchar (400);
	l_id				varchar (40);
	l_fail 				boolean;
BEGIN
	l_fail = false;
	l_rv = '{"status":"success"}';

	delete from "t_permission" where "created" < current_timestamp - interval '1 days';
	select "id"
		into l_id
		from "t_permission"
		where "permission_id" = p_permission
		  and "timeout" < current_timestamp 
		;
	IF NOT FOUND THEN
		l_rv = '{"status":"error","msg":"Failed to validate - not a valid temporary permission - most unfortunate."}';
		l_fail = true;
	END IF;

	if not l_fail then
		select "id"
			into l_id
			from "t_user"
			where "username" = p_email_addr
			  and "auth_token" = p_auth_token
			;
		IF NOT FOUND THEN
			l_rv = '{"status":"error","msg":"Failed to validate - not a valid login and email address - most unfortunate."}';
			l_fail = true;
		END IF;
	end if;

	return l_rv;
END;
$$ LANGUAGE plpgsql;

select cp_get_permission();





-- http://localhost:8204/?auth_token=8842f657-d0e0-4faa-8033-0a90100ee678&email_addr=philip6%40gmail.com#/

select cp_validate_auth_token ( 'philip6@gmail.com', '8842f657-d0e0-4faa-8033-0a90100ee678' ) as "x";

CREATE or REPLACE FUNCTION cp_password_reset ( p_email varchar, p_ip_addr varchar ) RETURNS varchar AS $$
DECLARE
	l_data			varchar (400);
	l_id			varchar (40);
	l_token			varchar (40);
	l_fail 			boolean;
	l_ip_ban 		boolean;
	l_junk			varchar (10);
BEGIN
	l_fail = false;
	l_data = '{"status":"failed"}';
	l_token = uuid_generate_v4();

	-- check banned ip
	select 'y' 
		into l_junk
		from "t_ip_ban"
		where "ip" = p_ip_addr
		;

	if not found then
		l_ip_ban = false;
	else
		l_ip_ban = true;
	end if;

	if l_ip_ban then
		l_data = '{ "status":"failed", "code":"009", "msg":"Invalid username or password." }';
		l_fail = true;
	end if;

	-- ---------------------------------------------------------------------------------------
	-- Fetch Data	
	-- ---------------------------------------------------------------------------------------
	if not l_fail then	
		select "id"
			into l_id
			from "t_user"
			where "email_address" = p_email
			;
		IF NOT FOUND THEN
			l_fail = true;
			l_data = '{"status":"error","msg":"Invalid Email. Unable to find user."}';
		END IF;
	end if;
	
	-- ---------------------------------------------------------------------------------------
	-- Update user and create email.
	-- ---------------------------------------------------------------------------------------
	if not l_fail then	
		update "t_user"
			set "email_reset_key" = l_token
			  , "email_reset_timeout" = current_timestamp + interval '2 hours'
			where "id" = l_id;
		l_data = '{"status":"success","email_reset_key":"'||l_token||'"}';
	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

select cp_password_reset ( 'p115@gmail.com', '127.0.0.1' );

