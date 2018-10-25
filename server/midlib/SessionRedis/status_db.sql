





-- "Copyright (C) Philip Schlump, 2009-2017." 

drop FUNCTION status_db ( p_ip_addr varchar );

CREATE or REPLACE FUNCTION status_db ( p_ip_addr varchar )
	RETURNS varchar AS $$
DECLARE
	l_data	varchar (200);
BEGIN
	l_data = '{"status":"success","ip":'||to_json(p_ip_addr)||'}';
	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

