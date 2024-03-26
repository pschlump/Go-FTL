--
--	,"/api/setup_test3": { "g": "setup_test3", "p": [ "$ip$"]
--		, "LoginRequired":false
--		, "LineNo":"Line: __LINE__ File: __FILE__"
--		, "Method":["GET","POST"]
--		, "TableList":[ "t_email_list", "t_log_info", "t_email_tab" ]
--		, "valid": {
--			 "$ip$": 		{ "required":true, "type":"string", "max_len":40, "min_len":4 }
--			}
--		}
--

CREATE or REPLACE FUNCTION setup_test3 (p_ip_addr varchar)
	RETURNS varchar AS $$
DECLARE
	l_rv			varchar (40);
BEGIN
	delete from "t_email_list" where "to_addr" = 'b@b.com';
	delete from "t_log_info" where "info" = 'some-info';
	delete from "t_email_tab" where "person_name" = 'some-person-name' and "email_addr" = 'b@b.com';
	RETURN '{"status":"success"}';
END;
$$ LANGUAGE plpgsql;

