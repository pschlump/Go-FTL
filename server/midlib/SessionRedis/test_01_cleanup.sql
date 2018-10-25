
delete from "t_auth_token" where "user_id" in ( select "id" from "t_user" where "username" = 'test01' );
delete from "t_user" where "username" = 'test01';

