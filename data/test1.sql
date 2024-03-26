
CREATE SEQUENCE t_version_seq
  INCREMENT 1
  MINVALUE 1
  MAXVALUE 9223372036854775807
  START 1
  CACHE 1;

drop TABLE "t_version" ;
CREATE TABLE "t_version" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "info"				text
	, "short"				char varying (20)
	, "seq"					integer DEFAULT nextval('t_version_seq')
	, "created" 			timestamp default current_timestamp not null 						
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
);

insert into "t_version" ( "info" ) values ( 'abc' );
insert into "t_version" ( "info" ) values ( 'def' );
insert into "t_version" ( "info" ) values ( 'good' );

	select info
		from ( select max("seq") max_seq from "t_version") as t1
		, "t_version"
		where "seq" = t1.max_seq
		limit 1
	;

