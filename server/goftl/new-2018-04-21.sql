
alter table "t_log" add column "log_timestamp"				timestamp ;
alter table "t_log" add column "error_level"				bigint ;
alter table "t_log" add column "message"				text;
alter table "t_log" add column "source"				text;

CREATE TABLE "t_log" (
	  "id"			char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "created_by"	char varying ( 100 )
	, "msg"			text
	, "log_timestamp"	timestamp
	, "error_level"		bigint
	, "message"			text
	, "source"			text
	, "created" 	timestamp default current_timestamp not null 						
);

