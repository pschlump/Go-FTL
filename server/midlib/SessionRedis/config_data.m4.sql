
-- "Copyright (C) Philip Schlump, 2009-2017." 

delete from "t_config" 
	where "customer_id" = '1'
;

insert into "t_config" ( "customer_id", "item_name", "value" ) values
	( '1', 'from.address', 'pschlump@gmail.com' )
,	( '1', 'debug.status.1', 'on' )
,	( '1', 'acct.auth_token.expire', '94 days' )
,	( '1', 'email.confirm.is.login', 'no' )
,	( '1', 'register.redirect.to', '/newly-registered.html' )
,	( '1', 'register.redirect.to.app', 'http://localhost:3000/newly-registered' )
,	( '1', 'recover.redirect.to', 'http://localhost:3000/recover-password-pt2' )
,	( '1', 'recover.redirect.to.old', '/recover-password.html' )
,	( '1', 'XSRF.token', 'per-user' )	-- or 'progressive-hashed' or 'off'
,	( '1', 'username.is.email', 'no' )	-- username is the email.  - email address value will be overwritten with 'username'
,	( '1', '2fa.required', 'no' )		-- yes/no - if yes then login is not finished unilt 2fa "pin" is provided.
;
insert into "t_config" ( "customer_id", "item_name", "value", "i_value" ) values
	( '1', 'ttl.user.auth_token', '', 180 )
;

delete from "t_customer"
	where "id" = '1'
;

insert into "t_customer" ( "id", "name" ) values ( '1', 'test-customer' );

delete from  "t_host_to_customer"
	where "customer_id" = '1'
;

insert into "t_host_to_customer" ( "customer_id", "host_name" ) values
	( '1', 'http://www.2c-why.com' )
,	( '1', 'http://auth.2c-why.com' )
,	( '1', 'http://www.2c-why.com/' )
,	( '1', 'http://auth.2c-why.com/' )
;

-- is_localhost set for testing APIs --
-- Like: validate auth_token key
insert into "t_host_to_customer" ( "customer_id", "host_name", "is_localhost" ) values
 	( '1', 'http://localhost:9001', 'yes' )
, 	( '1', 'http://localhost:9001/', 'yes' )
;


