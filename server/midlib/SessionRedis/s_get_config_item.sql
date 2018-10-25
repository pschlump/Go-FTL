





-- "Copyright (C) Philip Schlump, 2009-2017." 

drop FUNCTION s_get_config_item(p_item_name varchar, p_customer_id varchar, p_dflt varchar);

CREATE or REPLACE FUNCTION s_get_config_item(p_item_name varchar, p_customer_id varchar, p_dflt varchar)
	RETURNS varchar AS $$
DECLARE
    l_mode 				varchar (1000);
BEGIN

	-- "Copyright (C) Philip Schlump, 2009-2017." 

	l_mode = p_dflt;

	select "value"
		into l_mode
		from "t_config"
		where "item_name" = p_item_name
		  and "customer_id" = p_customer_id
		limit 1
		;
	if not found then
		l_mode = p_dflt;
	end if;

	RETURN l_mode;
END;
$$ LANGUAGE plpgsql;





drop FUNCTION s_get_config_item_int(p_item_name varchar, p_customer_id varchar, p_dflt int);

CREATE or REPLACE FUNCTION s_get_config_item_int(p_item_name varchar, p_customer_id varchar, p_dflt int)
	RETURNS int AS $$
DECLARE
    l_mode 				int;
BEGIN

	-- "Copyright (C) Philip Schlump, 2009-2017." 

	l_mode = p_dflt;

	select "i_value"
		into l_mode
		from "t_config"
		where "item_name" = p_item_name
		  and "customer_id" = p_customer_id
		limit 1
		;
	if not found then
		l_mode = p_dflt;
	end if;

	RETURN l_mode;
END;
$$ LANGUAGE plpgsql;

