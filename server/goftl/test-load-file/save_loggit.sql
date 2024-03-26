
-- ,"/api/save-loggit": { "g": "save_loggit", "p": [ "subject", "body", "status", "$ip$", "$host$" ], "nokey":true

CREATE or REPLACE FUNCTION save_loggit ( p_subject varchar, p_body varchar, p_status varchar, p_ip_addr varchar, p_host varchar ) RETURNS varchar AS $$
DECLARE
    l_body char varying(2100);
    l_subject char varying(200);
    l_status char varying(40);
BEGIN

	if p_status is null then
		l_status = '';
	else 
		l_status = p_status;
	end if;
	if p_body is null then
		l_body = '';
	else 
		l_body = p_body;
	end if;
	if p_subject is null then
		l_subject = '';
	else 
		l_subject = p_subject;
	end if;

	l_body = 'IP:'||p_ip_addr||' HOST:'||p_host||' Body:'||l_body;

	insert into "feedback" (
			  "subject"	
			, "body"
			, "status"				
		) values (
			  l_subject
			, l_body
			, l_status
		);

	RETURN '{"status":"success"}';

END;
$$ LANGUAGE plpgsql;

-- quick test of function

-- select save_loggit ( 'a', 'b', 'c', '127.0.0.1', 'host' );

