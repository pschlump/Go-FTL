





-- "Copyright (C) Philip Schlump, 2009-2017." 

drop table "t_ip_ban" ;
CREATE TABLE "t_ip_ban" (
	  "ip"					uuid not null primary key
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);
insert into "t_ip_ban" ( "ip" ) values ( '1.1.1.2' );

