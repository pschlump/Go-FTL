CREATE TABLE "t_monitor_stuff" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "item_name"			char varying (240) not null 				
	, "event_to_raise"		char varying (240) not null 				
	, "delta_t"				char varying (240) not null 				
	, "timeout_event"		timestamp not null
	, "note"				text
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
);

create index "t_monitor_stuff_p1" on "t_monitor_stuff" ( "timeout_event" );
create unique index "t_monitor_stuff_u1" on "t_monitor_stuff" ( "item_name" );

delete from "t_monitor_stuff";
