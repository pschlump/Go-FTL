
drop FUNCTION s_nvl(p_x varchar);

CREATE or REPLACE FUNCTION s_nvl(p_x varchar)
	RETURNS varchar AS $$
BEGIN
	if p_x is null then
		RETURN '';
	else 
		RETURN p_x;
	end if;
END;
$$ LANGUAGE plpgsql;

