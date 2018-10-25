
-- CREATE or REPLACE FUNCTION s_get_xsrf_mode(p_customer_id varchar)
-- ,	( '1', 'XSRF.token', 'per-user' )	-- or 'progressive-hashed' or 'off'

select 'success-300' from dual
	where exists ( select s_get_xsrf_mode ( '1' ) )
;

select 'success-301' from dual
	where s_get_xsrf_mode ( '1' ) in ( 'off', 'pregressive-hashed', 'per-user' )
;

