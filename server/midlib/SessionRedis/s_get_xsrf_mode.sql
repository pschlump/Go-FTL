






-- "Copyright (C) Philip Schlump, 2009-2017." 

drop FUNCTION s_get_xsrf_mode(p_customer_id varchar);

CREATE or REPLACE FUNCTION s_get_xsrf_mode(p_customer_id varchar)
	RETURNS varchar AS $$
DECLARE
    l_mode 				varchar (1000);
BEGIN

	-- "Copyright (C) Philip Schlump, 2009-2017." 

	l_mode = s_get_config_item( 'XSRF.token', p_customer_id, 'off');

	RETURN l_mode;
END;
$$ LANGUAGE plpgsql;



