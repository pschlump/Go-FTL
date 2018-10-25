--
-- Check that a password is valid , v.s. 300M known pwned passwsords
--


drop FUNCTION s_chk_passwd_not_pwned ( p_sha1_pw varchar );

CREATE or REPLACE FUNCTION s_chk_passwd_not_pwned ( p_sha1_pw varchar )
	RETURNS varchar AS $$
DECLARE
	l_data				varchar(80);
	l_junk				varchar(10);
BEGIN
	-- assume the worst
	l_data = '{"status":"error"}';

	select 'found' "found"
		into l_junk
		from "t_pwned"
		where "pw_hash" = ('\x' || p_sha1_pw)::bytea
		limit 1
		;
	if not found then
		l_data = '{"status":"success","msg":"password is not in ''pwned'' list."}';
	end if;

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;
