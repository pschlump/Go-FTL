
--CREATE or REPLACE FUNCTION s_get_customer_id_from_url(p_url varchar)

select 'success-400' from dual where exists ( select s_get_customer_id_from_url('http://localhost:9001') );

