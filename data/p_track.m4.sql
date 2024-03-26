
-- Tracking of where and when a person uses a single-page app

m4_changequote(`[[[', `]]]')
m4_include(common.m4.sql)

-- drop table "p_track" ;
create table "p_track" (
	  "id"						char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "app"						text		-- name of app
	, "locaiton_url"			text		-- the path in the url
	, "user_info"				text		-- any per-user identifier - if avail.
	, "updated" 				timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 				timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
	, "user_id"					char varying (40) not null 				-- fk to t_user
);


m4_updTrig(p_track)

