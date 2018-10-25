#!/bin/sh

# "Copyright (C) Philip Schlump, 2009-2017." 

CUST_Name="$1"
CUST_ID="$2"
CUST_PORT="$3"
CN_EMAIL="$4"

cat >new_cust.$CUST_ID.sql <<XXxx

-- delete from "t_config" 
-- 	where "customer_id" = '$CUST_ID'
-- ;

insert into "t_config" ( "customer_id", "item_name", "value" ) values
	( '$CUST_ID', 'from.address', 'support.$CN_EMIAIL@2c-why.com' )
,	( '$CUST_ID', 'debug.status.1', 'on' )
,	( '$CUST_ID', 'acct.auth_token.expire', '94 days' )
,	( '$CUST_ID', 'email.confirm.is.login', 'no' )
,	( '$CUST_ID', 'register.redirect.to', '/newly-registered.html' )
,	( '$CUST_ID', 'recover.redirect.to', '/recover-password.html' )
,	( '$CUST_ID', 'XSRF.token', 'per-user' )	-- or 'progressive-hashed' or 'off'
,	( '$CUST_ID', 'username.is.email', 'no' )	-- username is the email.  - email address value will be overwritten with 'username'
,	( '$CUST_ID', '2fa.required', 'no' )		-- yes/no - if yes then login is not finished unilt 2fa "pin" is provided.
;
insert into "t_config" ( "customer_id", "item_name", "value", "i_value" ) values
	( '$CUST_ID', 'ttl.user.auth_token', '', 180 )
;

-- delete from "t_customer"
-- 	where "id" = '$CUST_ID'
-- ;

insert into "t_customer" ( "id", "name" ) values ( '$CUST_ID', '$CUST_Name' );

-- delete from  "t_host_to_customer"
-- 	where "customer_id" = '$CUST_ID'
-- ;

insert into "t_host_to_customer" ( "customer_id", "host_name" ) values
	( '$CUST_ID', 'http://www.2c-why.com' )
,	( '$CUST_ID', 'http://auth.2c-why.com' )
,	( '$CUST_ID', 'http://www.2c-why.com/' )
,	( '$CUST_ID', 'http://auth.2c-why.com/' )
;

insert into "t_host_to_customer" ( "customer_id", "host_name", "is_localhost" ) values
 	( '$CUST_ID', 'http://localhost:$CUST_PORT', 'yes' )
, 	( '$CUST_ID', 'http://localhost:$CUST_PORT/', 'yes' )
;

XXxx

mkdir -p out

d9 <new_cust.$CUST_ID.sql >out/new_cust.$CUST_ID.log

