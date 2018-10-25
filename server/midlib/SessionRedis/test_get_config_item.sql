
-- CREATE or REPLACE FUNCTION s_get_config_item(p_item_name varchar, p_customer_id varchar, p_dflt varchar)

select s_get_config_item ( 'bob', '1', 'success-200' );
select 'success-201' from dual where exists ( select s_get_config_item ( 'from.address', '1', 'fail-201' ) )

