

select 'success'||'-'||'001' from dual where exists ( select 'ok' from "t_config" where "customer_id" = '1' );
select 'success'||'-'||'002' from dual where exists ( select 'ok' from "t_customer" where "id" = '1' );
select 'success'||'-'||'003' from dual where exists ( select 'ok' from "t_host_to_customer" where "customer_id" = '1' );
select 'success'||'-'||'004' from dual where exists ( select 'ok' from "t_host_to_customer" where "customer_id" = '1' and "host_name" = 'http://www.2c-why.com' );
select 'success'||'-'||'005' from dual where exists ( select 'ok' from "t_host_to_customer" where "customer_id" = '1' and "host_name" = 'http://www.2c-why.com/' );
select 'success'||'-'||'006' from dual where exists ( select 'ok' from "t_host_to_customer" where "customer_id" = '1' and "host_name" = 'http://auth.2c-why.com' );
select 'success'||'-'||'007' from dual where exists ( select 'ok' from "t_host_to_customer" where "customer_id" = '1' and "host_name" = 'http://auth.2c-why.com/' );
select 'success'||'-'||'008' from dual where exists ( select 'ok' from "t_host_to_customer" where "customer_id" = '1' and "is_localhost" = 'yes' );
select 'success'||'-'||'009' from dual where exists ( select 'ok' from "t_host_to_customer" where "customer_id" = '1' and "host_name" = 'http://localhost:9001' );
select 'success'||'-'||'010' from dual where exists ( select 'ok' from "t_host_to_customer" where "customer_id" = '1' and "host_name" = 'http://localhost:9001/' );

select 'success'||'-'||'011' from dual where exists ( select 'ok' from "t_config" where "customer_id" = '1' and "item_name" = 'from.address' );
select 'success'||'-'||'012' from dual where exists ( select 'ok' from "t_config" where "customer_id" = '1' and "item_name" = 'debug.status.1' );
select 'success'||'-'||'013' from dual where exists ( select 'ok' from "t_config" where "customer_id" = '1' and "item_name" = 'acct.auth_token.expire' );
select 'success'||'-'||'014' from dual where exists ( select 'ok' from "t_config" where "customer_id" = '1' and "item_name" = 'email.confirm.is.login' );
select 'success'||'-'||'015' from dual where exists ( select 'ok' from "t_config" where "customer_id" = '1' and "item_name" = 'register.redirect.to' );
select 'success'||'-'||'016' from dual where exists ( select 'ok' from "t_config" where "customer_id" = '1' and "item_name" = 'recover.redirect.to' );
select 'success'||'-'||'017' from dual where exists ( select 'ok' from "t_config" where "customer_id" = '1' and "item_name" = 'XSRF.token' );
select 'success'||'-'||'018' from dual where exists ( select 'ok' from "t_config" where "customer_id" = '1' and "item_name" = 'username.is.email' );
select 'success'||'-'||'019' from dual where exists ( select 'ok' from "t_config" where "customer_id" = '1' and "item_name" = '2fa.required' );

