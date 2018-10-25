
-- TODO: build an automatic test for this - put in Makefile

-- select s_validate_auth_token ( '$auth_token', 'http://localhost:9001' );
-- select s_validate_auth_token ( '87742649-11a7-4a0a-8931-2ab241f4d45d', 'http://localhost:9001' );
select s_validate_auth_token ( '87742649-11a7-4a0a-8931-2ab241f4d45d', 'http://localhost:9002' );

