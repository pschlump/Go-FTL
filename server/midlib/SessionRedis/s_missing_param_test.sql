






-- "Copyright (C) Philip Schlump, 2009-2017." 

-- ,"/api/session/missing_param_test": { "g": "s_missing_param_test", "p": [ "abc" ]
drop FUNCTION s_missing_param_test ( p_abc varchar );

CREATE or REPLACE FUNCTION s_missing_param_test ( p_abc varchar )
	RETURNS varchar AS $$
DECLARE
	l_data	varchar (800);
BEGIN
	l_data = '{"status":"success","abc":'||to_json(p_abc)||'}';
	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

