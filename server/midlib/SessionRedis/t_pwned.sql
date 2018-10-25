




-- 
-- 	select 'found' "found"
-- 		into l_junk
-- 		from "t_pwned"
-- 		where "pw_hash" = p_sha1_pw
-- 		limit 1
-- 		;
-- 	if not found then
-- 		l_data = '{"status":"success","msg":"password is not in ''pwned'' list."}';

create table "t_pwned" (
	"pw_hash" bytea not null primary key
);


