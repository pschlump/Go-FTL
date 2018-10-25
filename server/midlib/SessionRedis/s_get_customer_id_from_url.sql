





-- "Copyright (C) Philip Schlump, 2009-2017." 

drop FUNCTION s_get_customer_id_from_url(p_url varchar);

CREATE or REPLACE FUNCTION s_get_customer_id_from_url(p_url varchar)
	RETURNS varchar AS $$
DECLARE
    l_customer_id 				varchar (40);
BEGIN

	-- "Copyright (C) Philip Schlump, 2009-2017." 

	select "customer_id"
		into l_customer_id
		from "t_host_to_customer"
		where "host_name" = p_url
		limit 1
		;
	if not found then
		l_customer_id = 'error-missing-url-to-customer';
	end if;

	RETURN l_customer_id;
END;
$$ LANGUAGE plpgsql;

