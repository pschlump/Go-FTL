
delete from "t_email_list" where "to_addr" = 'a@b.com';

select reg_email_list ( 'a@b.com'::varchar, '127.0.0.1'::varchar ) as x;
select * from "t_email_list" where "to_addr" = 'a@b.com';
select 'PASS' from "t_email_list" where "to_addr" = 'a@b.com' and "confirmed" = 'n' and "de_reg" = 'n' and "updated" is null
union
select 'FA'||'IL' from "t_email_list" where "to_addr" = 'a@b.com' and ( "confirmed" <> 'n' or "de_reg" <> 'n' or "updated" is not null )
union
select 'FA'||'IL' from "t_dual" where not exists (
	select 'PASS' from "t_email_list" where "to_addr" = 'a@b.com' 
	);

select reg_email_list ( 'a@b.com', '127.0.0.1' ) as x;
select * from "t_email_list" where "to_addr" = 'a@b.com';
select 'PASS' from "t_email_list" where "to_addr" = 'a@b.com' and "confirmed" = 'n' and "de_reg" = 'n' and "updated" is null
union
select 'FA'||'IL' from "t_dual" where not exists (
	select 'PASS' from "t_email_list" where "to_addr" = 'a@b.com' and "confirmed" = 'n' and "de_reg" = 'n' and "updated" is null
	);

select dereg_email_list ( 'a@b.com', '127.0.0.1' ) as x;
select * from "t_email_list" where "to_addr" = 'a@b.com';
select 'PASS' from "t_email_list" where "to_addr" = 'a@b.com' and "confirmed" = 'n' and "de_reg" = 'y' and "updated" is not null
union
select 'FA'||'IL' from "t_dual" where not exists (
	select 'PASS' from "t_email_list" where "to_addr" = 'a@b.com' and "confirmed" = 'n' and "de_reg" = 'y' and "updated" is not null
	);

select confirm_email_list ( 'a@b.com', '127.0.0.1' ) as x;
select * from "t_email_list" where "to_addr" = 'a@b.com';
select 'PASS' from "t_email_list" where "to_addr" = 'a@b.com' and "confirmed" = 'y' and "de_reg" = 'n' and "updated" is not null
union
select 'FA'||'IL' from "t_dual" where not exists (
	select 'PASS' from "t_email_list" where "to_addr" = 'a@b.com' and "confirmed" = 'y' and "de_reg" = 'n' and "updated" is not null
	);

