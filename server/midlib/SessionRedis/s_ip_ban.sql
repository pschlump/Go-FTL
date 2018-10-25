





-- "Copyright (C) Philip Schlump, 2009-2017." 

drop FUNCTION s_ip_ban(p_ip_addr varchar);

CREATE or REPLACE FUNCTION s_ip_ban(p_ip_addr varchar)
	RETURNS boolean AS $$
DECLARE
	l_ip_ban 	boolean;
	l_junk		varchar (1);
begin

	-- "Copyright (C) Philip Schlump, 2009-2017." 

	select 'y' 
		into l_junk
		from "t_ip_ban"
		where "ip" = p_ip_addr
		;

	if not found then
		l_ip_ban = false;
	else
		l_ip_ban = true;
	end if;

	RETURN l_ip_ban;
END;
$$ LANGUAGE plpgsql;

