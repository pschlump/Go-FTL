

drop FUNCTION x_test0(p_username varchar);

CREATE or REPLACE FUNCTION x_test0(p_username varchar)
	RETURNS varchar AS $$
DECLARE
    l_data 		varchar (4000);
BEGIN

	l_data = 'bob';

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

drop FUNCTION x_test1(p_username varchar);

CREATE or REPLACE FUNCTION x_test1(p_username varchar)
	RETURNS varchar AS $$
DECLARE
    l_id 		varchar (40);
    l_data 		varchar (4000);
    l_token 	varchar (40);
BEGIN

	-- l_data = '{"status":"success"}';

	l_token = x_test0(p_username);

	l_data = '{"status":"success","x_test0":"'||l_token||'"}';

	RETURN l_data;
END;
$$ LANGUAGE plpgsql;

select x_test1('aaa');

