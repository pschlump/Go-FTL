

select 'PASS' from "t_email_tab" where "person_name" = 'some-person-name' and "email_addr" = 'b@b.com'
union
select 'FA'||'IL' from "t_dual" where not exists (
	select 'PASS' from "t_email_tab" where "person_name" = 'some-person-name' and "email_addr" = 'b@b.com'
	);

