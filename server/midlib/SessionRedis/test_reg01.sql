
delete from "t_user" where "username" = 'test01';
 
select s_register_new_user('test01', '123456', '123456', '1.1.1.1', 'kermit.nosend.01@gmail.com', 'Kermit Frog', 'http://www.2c-why.com/', 'test-application' );

