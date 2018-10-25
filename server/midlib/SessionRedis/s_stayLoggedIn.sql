





-- "Copyright (C) Philip Schlump, 2009-2017." 

drop FUNCTION s_stayLoggedIn() ;

CREATE or REPLACE FUNCTION s_stayLoggedIn()
	RETURNS varchar AS $$
DECLARE
BEGIN
	RETURN '{ "status":"success"}';
END;
$$ LANGUAGE plpgsql;

