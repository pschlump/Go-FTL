
select * from "t_email_list" where "to_addr" = 'b@b.com';
select 'PASS' from "t_email_list" where "to_addr" = 'b@b.com' and "confirmed" = 'y' and "de_reg" = 'n' and "updated" is null
union
select 'FA'||'IL' from "t_dual" where not exists (
	select 'PASS' from "t_email_list" where "to_addr" = 'b@b.com' 
	);

