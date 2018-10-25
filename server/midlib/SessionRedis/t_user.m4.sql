
-- "Copyright (C) Philip Schlump, 2009-2017." 

-- drop table "t_user" ;
alter table "t_user" add column  "user_attr" 			text default '{}'	;

CREATE TABLE "t_user" (
	  "id"					uuid DEFAULT uuid_generate_v4() not null primary key
	, "username"			text not null
	, "password"			text not null
  	, "ip" 					char varying (40) not null			 						
	, "real_name"			text not null
	, "email_address"		text not null
	, "email_confirmed"		char (1) default 'n' not null check ( "email_confirmed" in ( 'y', 'n' ) )
	, "acct_state"			char varying (10) not null default 'unknown' check ( "acct_state" in ( 'unknown', 'locked', 'ok', 'pass-reset', 'billing', 'closed', 'fixed', 'temporary' ) )
	, "default_priority"	int default 10 														-- 10 is standard priority ( higher is higher, lower is lower )
	, "acct_expire"			timestamp
	, "n_login_fail"		int default 0 not null
	, "login_fail_delay"	timestamp
	, "email_reset_key"		char varying (80) 
	, "email_reset_timeout" timestamp 		 													-- 
	, "password_set_date" 	timestamp 		 													-- 
	, "dflt_priority"		int default 10 not null												-- 10 is standard priority ( higher is higher, lower is lower )
	, "last_login" 			timestamp 		 													-- 
  	, "privs" 				text default '[]'													-- JSON of privilate set, empty means testing login
  	, "user_attr" 			text default '{}'													-- User specified attributes
	, "customer_id"			char varying (40) default '1'
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);

create unique index "t_user_u1" on "t_user" ( "auth_token" );
create unique index "t_user_u2" on "t_user" ( "email_address" );
create unique index "t_user_u3" on "t_user" ( "email_reset_key" );
create unique index "t_user_u4" on "t_user" ( "username" );

m4_updTrig(t_user)

