

-- Tracking of where and when a person uses a single-page app

m4_changequote(`[[[', `]]]')
m4_include(common.m4.sql)

-- drop table "p_issue" ;
create table "p_issue" (
	  "id"						char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "title"					char varying (250)
	, "desc"					text		
	, "type_group"				char varying (50)		-- webpage / product name etc // [ Notification - ask for help ]
	, "locaiton_url"			text		
	, "info"					text					-- JSON of any related info
	, "assinged_to"				char varying (50)
	, "state_of"				char varying (50)
	, "owner_user_id"			char varying (40)  		-- fk to t_user
	, "assigned_user_id"		char varying (40)  		-- fk to t_user
	, "notify_flag"				char varying (15)		-- "please", "noted", "resp"
	, "updated" 				timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 				timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);

m4_updTrig(p_issue)

