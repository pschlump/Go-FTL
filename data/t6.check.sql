
select 'PASS' from "t_log_info" where "info" = 'some-info'
union
select 'FA'||'IL' from "t_dual" where not exists (
	select 'PASS' from "t_log_info" where "info" = 'some-info'
	);

