select t1."user_id", t1."auth_token", t2."username"
from "t_auth_token" as t1 join "t_user" as t2 on t1."user_id" = t2."id"
order by 3, 1, 2;
