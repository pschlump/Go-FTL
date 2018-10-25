
-- CREATE or REPLACE FUNCTION s_ip_ban(p_ip_addr varchar)
-- select 'success-019' from dual where exists ( select 'ok' from "t_config" where "customer_id" = '1' and "item_name" = '2fa.required' );

select 'success-100' from dual where s_ip_ban ( '1.1.1.1' ) = false;
select 'success-101' from dual where s_ip_ban ( '1.1.1.2' ) = true;
