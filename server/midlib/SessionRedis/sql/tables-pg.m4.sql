
-- -------------------------------------------------------- -- --------------------------------------------------------
-- Model
--
--   t_group ---< t_user ---< t_priv
--
-- A run needs
--	1. x_run_id (number/seq)
--	2. auth_key		- random hash for saving data
--	3. user_data 	- string
--	4. URL for data
--	5. user_auth_key - passwrod for running tests in user's code
--	6. host_id		- not transmitted to user
--	7. client_id	- ID of the Client that is running this
--
-- -------------------------------------------------------- -- --------------------------------------------------------
m4_changequote(`[[[', `]]]')






drop view t_available_test_systems;

drop table "t_activity" ;
drop table "t_client" ;
drop table "t_config" ;
drop table "t_group" ;
drop table "t_host" ;
drop table "t_host_can_run" ;
drop table "t_host_can_vm";
drop table "t_job" ;
drop table "t_priv" ;
drop table "t_project" ;
drop table "t_run" ;
drop table "t_runSet" ;
drop table "t_runSet_members" ;
drop table "t_run_useragent" ;
drop table "t_user" ;
drop table "t_userAgent" ;
drop table "t_run_result" ;
drop table "t_a_run";
drop table "t_a_run_coverage";
drop table "t_need_coverage";
drop table "t_email_q";
drop table "t_email_q2";
drop table "t_monitor_stuff";
drop table "t_status_stuff";
drop table "t_ssh_port_pool";
drop TABLE "t_ssh_login";
drop table "t_ssh_global";
drop table "t_it_work";
drop TABLE "t_knobs" ;
drop table "t_ua_log";
drop table "t_ms_test";
drop table "t_get_row";
drop table "t_post_results_back";
drop table "t_output" ;
drop table "t_dispatch_scripts";
drop table "t_a_run_periodic";
drop table "t_link_valid";
drop TABLE "t_customer" ;






-- -------------------------------------------------------- -- --------------------------------------------------------
-- Should add a trigger to this and catpure when stuff "DID" happen - make a nice display
-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE TABLE "t_monitor_stuff" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "item_name"			char varying (240) not null 				
	, "event_to_raise"		char varying (240) not null 				
	, "delta_t"				char varying (240) default '99 day' not null 				
	, "enabled"				char varying (1) default 'n' not null 				
	, "timeout_event"		timestamp not null
	, "note"				text
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
);

create index "t_monitor_stuff_p1" on "t_monitor_stuff" ( "timeout_event" );
create unique index "t_monitor_stuff_u1" on "t_monitor_stuff" ( "item_name" );

delete from "t_monitor_stuff";

insert into "t_monitor_stuff" ( "item_name", "event_to_raise", "delta_t", "timeout_event" )
	values ( 'DatabaseBackup', 'Database Backup', '2 days', current_timestamp );

insert into "t_monitor_stuff" ( "item_name", "event_to_raise", "delta_t", "timeout_event" )
	values ( 'SystemBackup', 'System Backup', '2 days', current_timestamp );

insert into "t_monitor_stuff" ( "item_name", "event_to_raise", "delta_t", "timeout_event" )
	values ( 'EmailSent-Q1', 'Send of email on normal chanel', '2 minutes', current_timestamp );

insert into "t_monitor_stuff" ( "item_name", "event_to_raise", "delta_t", "timeout_event" )
	values ( 'EmailSent-Q2', 'Send of email on backup chanel', '5 minutes', current_timestamp );

insert into "t_monitor_stuff" ( "item_name", "event_to_raise", "delta_t", "timeout_event" )
	values ( 'InternetUp', 'Check of DSL Modem Status', '2 minute', current_timestamp );

insert into "t_monitor_stuff" ( "item_name", "event_to_raise", "delta_t", "timeout_event" )
	values ( '99PctBrowsers', 'Check On 99% of Browsers', '2 days', current_timestamp );

-- delete from "t_monitor_stuff" where "item_name" = 'LoadManager';
insert into "t_monitor_stuff" ( "item_name", "event_to_raise", "delta_t", "timeout_event" )
	values ( 'LoadManager', 'Start/Stop VMs as needed based on usage', '4 minute', current_timestamp );

-- Add monthly cleanup ( 35 days )
-- Add Successful Login ( 1ce per day )
-- Add Successful test run ( 1ce per day )
-- Add Web page live ( 1ce per hour )
-- Add Marketing page viewd ( 1ce per day )








-- -------------------------------------------------------- -- --------------------------------------------------------
-- Table is a singelton - only 1 row
-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE TABLE "t_knobs" (
	  "id"					char varying (40) DEFAULT 'x' not null primary key
	, "start_vm_if_x_min"	bigint default 2 not null
	, "idle_for_x_min"		bigint default 30 not null
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
);

insert into "t_knobs" ( "id" ) values ( 'x' );;



-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE TABLE "t_status_stuff" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "hostname"			char varying (40) not null 				
	, "item_name"			char varying (240) not null 				
	, "status"				char varying (40) not null 				
	, "code"				char varying (40) not null 				
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
);

-- -------------------------------------------------------- -- --------------------------------------------------------
-- alter table "t_email_q" add column  "template_name"			text;
-- alter table "t_email_q" add column  "template_data"			text;

CREATE TABLE "t_email_q" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "user_id"				char varying (40) not null 				-- fk to t_user
	, "ip"					char varying (40) not null 				-- IP of sender, or 0.0.0.0 - from server
	, "auth_token"			char varying (40) not null 				-- A validated auth_token
	, "to"					char varying (255) not null
	, "from"				char varying (255) not null
	, "subject"				char varying (255) not null
	, "body"				text
	, "text_body"			text
	, "error"				text
	, "status"				char varying (10) check ( "status" in ( 'sent', 'pending', 'test', 'in-prog', 'error' ) )
	, "template_name"		text
	, "template_data"		text
	, "sent_at" 			timestamp 
	, "created" 			timestamp default current_timestamp not null 						--
);

create index "t_email_q_p1" on "t_email_q" ( "status", "created" );

-- -------------------------------------------------------- -- --------------------------------------------------------
-- For monitoring email that failed to send.  A 2nd daemon will pull from this and send using gmail account.
-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE TABLE "t_email_q2" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "user_id"				char varying (40) not null 				-- fk to t_user
	, "ip"					char varying (40) not null 				-- IP of sender, or 0.0.0.0 - from server
	, "auth_token"			char varying (40) not null 				-- A validated auth_token
	, "to"					char varying (255) not null
	, "from"				char varying (255) not null
	, "subject"				char varying (255) not null
	, "body"				text
	, "text_body"			text
	, "error"				text
	, "status"				char varying (10) check ( "status" in ( 'sent', 'pending', 'test', 'in-prog', 'error' ) )
	, "sent_at" 			timestamp 
	, "created" 			timestamp default current_timestamp not null 						--
);

create index "t_email_q2_p1" on "t_email_q2" ( "status", "created" );

-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE TABLE "t_ssh_port_pool" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "port"				char varying (10)					-- reservation for port# from the port pool.
	, "in_use"				char (1) default 'n' not null check ( "in_use" in ( 'y', 'n' ) )
	, "use_count"			int default 0 not null
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);
insert into "t_ssh_port_pool" ( "port" ) values ( '9000' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9001' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9002' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9003' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9004' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9005' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9006' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9007' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9008' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9009' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9010' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9011' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9012' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9013' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9014' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9015' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9016' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9017' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9018' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9019' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9020' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9021' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9022' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9023' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9024' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9025' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9026' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9027' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9028' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9029' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9030' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9031' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9032' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9033' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9034' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9035' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9036' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9037' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9038' );
insert into "t_ssh_port_pool" ( "port" ) values ( '9039' );

-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE TABLE "t_ssh_login" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "user_id"				char varying (40) 					-- Person making reservation
	, "port"				char varying (10)					-- reservation for port# from the port pool.
	, "expire_at" 			timestamp 									 	-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "updated" 			timestamp 									 	-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 	-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);
-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE TABLE "t_ssh_global" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "ip_of_this_site"		char varying (40)
	, "updated" 			timestamp 									 	-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 	-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);


-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE OR REPLACE FUNCTION sha1pw(bytea)
RETURNS character varying AS
$BODY$
BEGIN
RETURN ENCODE(DIGEST($1, 'sha1'), 'hex');
END;
$BODY$
LANGUAGE 'plpgsql';

-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE OR REPLACE FUNCTION sha256pw(bytea)
RETURNS character varying AS
$$
BEGIN
	RETURN ENCODE(DIGEST($1, 'sha256'), 'hex');
END;
$$ LANGUAGE 'plpgsql';

CREATE OR REPLACE FUNCTION sha256pw(p_txt varchar)
RETURNS character varying AS
$$
BEGIN
	RETURN sha256pw(p_txt::bytea);
END;
$$ LANGUAGE 'plpgsql';


-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE TABLE "t_config" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "name"				char varying (40) 
	, "type"				char varying (10)
	, "value"				char varying (255) 
	, "n_value"				int
	, "b_value"				char (1)
);

-- 0 = idle (awaiting (re-)run)
-- 1 = busy (being run by a client)
-- 2 = done (passed and/or reached max)
-- 3 = Waiting for User Input
-- 4 = Waiting for automatic event to occure
--insert into "t_config" ( "id", "name", "type", "n_value", "value" ) values ( '1', 't_run_useragent.status', 'ns', 0, 'idle' );
--insert into "t_config" ( "id", "name", "type", "n_value", "value" ) values ( '2', 't_run_useragent.status', 'ns', 1, 'busy' );
--insert into "t_config" ( "id", "name", "type", "n_value", "value" ) values ( '3', 't_run_useragent.status', 'ns', 2, 'done' );
--insert into "t_config" ( "id", "name", "type", "n_value", "value" ) values ( '4', 't_run_useragent.status', 'ns', 3, 'waiting-user' );
--insert into "t_config" ( "id", "name", "type", "n_value", "value" ) values ( '5', 't_run_useragent.status', 'ns', 4, 'waiting-auto' );

insert into "t_config" ( "name", "type", "value" ) values ( 'server.ip.001', 's', '0.0.0.0' );

-- alter table "t_user" add column "salt"				text ;
-- alter table "t_user" add column "customer_id"		char varying (40) default '1' ;
-- alter table "t_user" alter column "password" type char varying (80);
-- alter table "t_user" add column "privs"				text ;
-- ALTER TABLE "t_user" ALTER COLUMN "email_reset_key" TYPE char varying (80) ;
-- ALTER TABLE "t_user" ALTER COLUMN "auth_token" TYPE char varying (80) ;
-- ALTER TABLE "t_user" ALTER COLUMN "auth_token" TYPE char varying (80) ;
-- ALTER TABLE "t_a_run" ALTER COLUMN "auth_token" TYPE char varying (80) ;

-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE TABLE "t_user" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "group_id"			char varying (40) not null 											-- fk to t_group
	, "username"			char varying (40) not null
	, "password"			char varying (80) not null
	, "salt"				text
  	, "auth_token" 			char varying (80) not null			 								-- Hash of random-generated(uuid) token. To use as authentication to be allowed to
  	, "ip" 					char varying (40) not null			 						
	, "real_name"			char varying (255) not null
	, "email_address"		char varying (255) not null
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
	, "brTemplateName"		char varying (250)	default '%{browserFamily%} %{browserMajor%}.%{browserMinor%} <br> %{osFamily%} %{osMajor%}'
	, "brTemplateTitle"		char varying (250)	default '%{browserFamily%} %{browserMajor%}.%{browserMinor%}/%{osFamily%} %{osMajor%}'
	, "brHashTmpl"			char varying (250)	default '%{browserFamily_lc%}-%{browserMajor_lc%}-%{browserMinor_lc%}-%{osFamily_lc%}-%{osMajor_lc%}'
	, "brOptions"			char varying (250)	default 'browserFamily browserMajor browserMinor osFamily osMajor'
	, "customer_id"			char varying (40) default '1'
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);

create unique index "t_user_u1" on "t_user" ( "auth_token" );
-- create unique index "t_user_u2" on "t_user" ( "username", "password" );
create index "t_user_p1" on "t_user" ( "email_address" );
create unique index "t_user_u4" on "t_user" ( "username" );
create index "t_user_p2" on "t_user" ( "email_reset_key" );

insert into "t_user" ( "id", "group_id", "username", "real_name", "password", "email_address", "acct_state", "auth_token", "ip", "email_confirmed", "acct_expire", "privs" )
values ( '1', '1', 'goofy', 'Goofy Cat', sha1pw('Salt/Pepper/882211Fool!!'), 'pschlump@gmail.com', 'fixed', sha1pw('dfjkdfjldjfklsdjklsdjfklsdfkldjklfjdlkfjklsdjfklsdjf'), '0.0.0.0', 'y'
	, current_timestamp + interval '90000 days'
	, "{['who-cares-root','it-root'}"
 );
insert into "t_user" ( "id", "group_id", "username", "real_name", "password", "email_address", "acct_state", "auth_token", "ip", "email_confirmed", "acct_expire", "privs" )
values ( '2', '1', 'rodbrown', 'Mr Brown', sha1pw('Salt/Pepper/dEAdbEEf01'), 'rod@vanaire.com', 'fixed', sha1pw('dfj11f1l1jfklsdjklsdjfklsdfkldjklfjdlkfjklsdjfklsdjf'), '0.0.0.0', 'y'
	, current_timestamp + interval '90000 days'
	, "{['who-cares-root','it-root'}"
 );


-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE TABLE "t_group" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "primary_user_id"		char varying (40) not null 											-- fk to t_user for the contact/primary of group
	, "group_name"			char varying (40) not null											-- To join a group you will need group_name and password
	, "password"			char varying (50) not null
	, "group_type"			char varying (30) default 'Business' not null check ( "group_type" in ( 'Business', 'Root', 'OSS', 'Other' ) )
	, "billed"				char (1) default 'y' not null
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);

insert into "t_group" ( "id", "group_name", "password", "billed", "primary_user_id" ) values ( '1', 'Root', sha1pw('Salt/Pepper/Test Fool 342323'), 'n', '1' );
insert into "t_group" ( "id", "group_name", "password", "billed", "primary_user_id" ) values ( '2', 'Temporary', sha1pw('Salt/Pepper/Temporary 3333'), 'n', '1' );

-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE TABLE "t_priv" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "user_id"				char varying (40) not null 				-- fk to t_user
	, "priv_name"			char varying (240) not null 			-- privilage
);

insert into "t_priv" ( "id", "user_id", "priv_name" ) values ( '1', '1', 'May Create Admin' );
insert into "t_priv" ( "id", "user_id", "priv_name" ) values ( '2', '1', 'May Create Users' );
insert into "t_priv" ( "id", "user_id", "priv_name" ) values ( '3', '1', 'May Send Email' );
insert into "t_priv" ( "id", "user_id", "priv_name" ) values ( '4', '1', 'May Send Email to Self' );



-- -------------------------------------------------------- -- --------------------------------------------------------
-- This is a log of activities done by the users - there should never be an update.
-- This is used in billing the user.
-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE TABLE "t_activity" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "user_id"				char varying (40) not null 				-- fk to t_user
	, "act_name"			char varying (40) not null 			
	, "updated" 			timestamp  						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);


-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE SEQUENCE t_host_id_seq
  INCREMENT 1
  MINVALUE 1
  MAXVALUE 9223372036854775807
  START 1
  CACHE 1;

-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE TABLE "t_host" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "host_seq"			bigint DEFAULT nextval('t_host_id_seq'::regclass) NOT NULL 
	, "url_of_it"			char varying (255) not null 
	, "hostname"			char varying (255) not null 
  	, "host_type"			char varying (40) default 'VM' not null 	-- 'host', 'VM', '???'
	, "control_method"		char varying (40) default 'it'
	, "comm_method" 		char varying(10) not null default 'post' 	-- How communication occures with client, 'post', 'jsonp', 'socket.io', 'server'
	, "is_running_now"		char (1) default 'n' not null 
	, "can_rdc"				char (1) default 'n' not null 
	, "rdc_port"			char (6) 
	, "rdc_in_use"			char (1) default 'n' not null 
	, "rdc_user_id"			char (40)						-- User ID of person that is currently using this VMs Desktop 
	, "ip_addr"				char (40)						-- Location of server if not a local system
	, "last_run_at" 		timestamp  						-- Time of last "ping" or data received back from "it"
	, "hosted_at"			char (100) default 'inhouse'	-- Locaiton of hosing (www.macincloud.com, www.linode.com, digitalocean.com, aws.amazon.com etc) 
	, "connection_method"	char (10) default 'rdc'			-- rdc, vnc etc.
	, "n_test_run"			bigint default 0										
	, "osFamily"			char varying (40)
	, "osMajor"				char varying (40)
	, "osMinor"				char varying (40)
	, "osOptions"			char varying (40)
	, "isMobile"			char varying (1) default 'n'
	, "hasOrientation"		char varying (1) default 'n'
	, "currentOrientation"	char varying (4) default '0'
	, "updated" 			timestamp  						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 		
);


-- -------------------------------------------------------- -- --------------------------------------------------------
-- reall the list of browsers that a host can run.
-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE TABLE "t_host_can_run" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "host_id"				char varying (40) not null 
	, "useragent_id"		char varying (40) not null 
	, "vendor_version_no"	char varying (40) 
	, "config_data" 		char varying(255)  
	, "client_name" 		char varying(255)  
	, "client_id"			char varying (40) 
	, "is_running_now"		char (1) default 'n' not null 
	, "running_at"			timestamp
	, "updated" 			timestamp  						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 		
);
CREATE INDEX "t_host_can_run_u1" on "t_host_can_run" ( "host_id", "useragent_id" );

CREATE TABLE "t_host_can_vm" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "host_id"				char varying (40) not null 
	, "client_name" 		char varying(255)  
	, "osFamily"			char varying (40)
	, "osMajor"				char varying (40)
	, "osMinor"				char varying (40)
	, "osOptions"			char varying (40)
	, "isMobile"			char varying (1) default 'n'
	, "hasOrientation"		char varying (1) default 'n'
	, "currentOrientation"	char varying (4) default '0'
	, "is_running_now"		char (1) default 'n' not null 
	, "updated" 			timestamp  						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 		
);
CREATE INDEX "t_host_can_vm_p1" on "t_host_vm_run" ( "client_name" );



-- -------------------------------------------------------- -- --------------------------------------------------------
-- This table lists the "client" that is running on a "host" - it is kind of a stand-alone table since it is a client app.
CREATE TABLE "t_client" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
  	, "host_id" 			char varying (40) not null 										-- From starting the host up.  On command line of host.
	, "name" 				char varying(255) not null 						-- Freeform client name.
	, "useragent_id" 		char varying(40) not null 						-- Key to config.userAgents property.		-- change to t_userAgent table fk
	, "useragent" 			char varying(255) not null 						-- Raw User-Agent string.
	, "ip" 					char varying(40) not null 			 			-- Raw IP string as extractred by WebRequest::getIP
	, "is_running_now"		char (1) default 'n' not null 
	, "last_run_at" 		timestamp  									
	, "n_test_run"			bigint default 0										
	, "updated" 			timestamp  									
	, "created" 			timestamp default current_timestamp not null 		
);

CREATE INDEX "t_clients_useragent_updated_p1" ON "t_client" ("useragent_id", "updated"); 		-- Usage: HomePage, SwarmstateAction.
CREATE INDEX "t_clients_updated_p1" ON "t_client" ("updated"); 									-- Usage: CleanupAction.
CREATE INDEX "t_clients_name_ua_created_p1" ON "t_client" ("name", "useragent_id", "created");	-- Usage: ClientAction, ScoresAction, BrowserInfo and Client.

insert into "t_client" (
	  "id"					
  	, "host_id" 		
	, "name" 		
	, "useragent_id" 		
	, "useragent" 		
	, "ip" 			
) values ( 
	  '300'
  	, '400'
	, 'Ubuntu Linux 12.04:Grape:h01_c001'
	, '200'
	, 'bla bla bla'
	, '192.168.0.50'
);
insert into "t_client" (
	  "id"					
  	, "host_id" 		
	, "name" 		
	, "useragent_id" 		
	, "useragent" 		
	, "ip" 			
) values ( 
	  '301'
  	, '400'
	, 'Ubuntu Linux 12.04:Grape:h01_c001'
	, '201'
	, 'bla bla bla'
	, '192.168.0.51'
);
insert into "t_client" (
	  "id"					
  	, "host_id" 		
	, "name" 		
	, "useragent_id" 		
	, "useragent" 		
	, "ip" 			
) values ( 
	  '302'
  	, '400'
	, 'Windows 8:Grape:h01_c003'
	, '202'
	, 'bla bla bla'
	, '192.168.0.51'
);

-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE TABLE "t_userAgent" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "name"				char varying (40) not null
	, "title"				char varying (80) not null
	, "browserFamily"		char varying (40)
	, "browserMajor"		char varying (40)
	, "browserMinor"		char varying (40)
	, "browserOptions"		char varying (40)
	, "osFamily"			char varying (40)
	, "osMajor"				char varying (40)
	, "osMinor"				char varying (40)
	, "osOptions"			char varying (40)
	, "created" 			timestamp default current_timestamp not null 		
	, "n_test_run"			bigint default 0										
);

insert into "t_host" ( "id", "hostname", "url_of_it", "control_method", "is_running_now" )
	values ( '1', 'grape', 'http://192.168.0.167:9002', '*it*', 'y' );

insert into "t_host_can_run" ( "host_id", "useragent_id", "client_name", "is_running_now" )
	values ( '1', '200', 'h01_c001', 'y' );
insert into "t_host_can_run" ( "host_id", "useragent_id", "client_name", "is_running_now" )
	values ( '1', '201', 'h01_c002', 'y' );
insert into "t_host_can_run" ( "host_id", "useragent_id", "client_name", "is_running_now" )
	values ( '1', '202', 'h01_c003', 'y' );
insert into "t_host_can_run" ( "host_id", "useragent_id", "client_name", "is_running_now" )
	values ( '1', '203', 'h01_c004', 'y' );
insert into "t_host_can_run" ( "host_id", "useragent_id", "client_name", "is_running_now" )
	values ( '1', '204', 'h01_c005', 'y' );
insert into "t_host_can_run" ( "host_id", "useragent_id", "client_name", "is_running_now" )
	values ( '1', '205', 'h01_c006', 'y' );
insert into "t_host_can_run" ( "host_id", "useragent_id", "client_name", "is_running_now" )
	values ( '1', '206', 'h01_c007', 'y' );
insert into "t_host_can_run" ( "host_id", "useragent_id", "client_name", "is_running_now" )
	values ( '1', '207', 'h01_c008', 'y' );

insert into "t_userAgent" ( "id", "title", "name", "browserFamily", "browserMajor", "browserMinor", "osFamily", "osMajor", "osMinor" )
	values ( '200', 'Chrome 28/Linux', 'Chrome 28', 'chrome', '28', '0', 'linux', '3', '35' );
insert into "t_userAgent" ( "id", "title", "name", "browserFamily", "browserMajor", "browserMinor", "osFamily", "osMajor", "osMinor" )
	values ( '201', 'Chrome 12/Linux', 'Opera 12', 'opera', '12', '2', 'linux', '3', '35' );
insert into "t_userAgent" ( "id", "title", "name", "browserFamily", "browserMajor", "browserMinor", "osFamily", "osMajor", "osMinor" )
	values ( '202', 'IE 10/Win8', 'IE 10', 'IE', '10', '3', 'windows', '8', '1' );
insert into "t_userAgent" ( "id", "title", "name", "browserFamily", "browserMajor", "browserMinor", "osFamily", "osMajor", "osMinor" )
	values ( '203', 'FireFox 18/Mac OS X', 'FireFox 18', 'firefox', '18', '3', 'mac_os_x', '7', '2' );
insert into "t_userAgent" ( "id", "title", "name", "browserFamily", "browserMajor", "browserMinor", "osFamily", "osMajor", "osMinor" )
	values ( '204', 'iOS 6/Safari', 'Safari 6', 'safari', '6', '0', 'ios', '7', '0' );
insert into "t_userAgent" ( "id", "title", "name", "browserFamily", "browserMajor", "browserMinor", "osFamily", "osMajor", "osMinor" )
	values ( '205', 'Chrom/Android 2.3', 'Chrome 2.3', 'chrome', '2', '3', 'android', '2', '3' );
insert into "t_userAgent" ( "id", "title", "name", "browserFamily", "browserMajor", "browserMinor", "osFamily", "osMajor", "osMinor" )
	values ( '206', 'IE 9/Win 7', 'IE 9', 'ie', '9', '0', 'windows', '7', '4' );
insert into "t_userAgent" ( "id", "title", "name", "browserFamily", "browserMajor", "browserMinor", "osFamily", "osMajor", "osMinor" )
	values ( '207', 'Chromium 2.2', 'Chromium 2.2', 'chromium', '2', '2', 'chromium', '2', '2' );


select 
	  t3."osFamily" as "osNameClass"
	, t3."browserFamily" as "browserNameClass"
	, t3."osMajor" as "osMajorClass"
	, min(t3."osMinor") as "osMinorClass"
	, t3."browserMajor" as "browserMajorClass"
	, min(t3."browserMinor") as "browserMinorClass"
	, t3."name" as "browserName"
	, t3."name" as "title"
	, count(1) as "n_clients"
	, sum(t3."n_test_run") as "n_runs"
from "t_userAgent" t3
	, "t_host_can_run" t2
	, "t_host" t1
where t1."id" = t2."host_id"
  and t2."useragent_id" = t3."id"
group by t2."useragent_id"
	,  t3."osFamily" 
	, t3."browserFamily" 
	, t3."osMajor" 
	, t3."browserMajor" 
	, t3."name" 
order by 10 desc
;

-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE TABLE "t_runSet" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "user_id"				char varying (40) not null 				-- fk to t_user
	, "name"				char varying (40) not null
	, "created" 			timestamp default current_timestamp not null 		
);
create table "t_runSet_members" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "runset_id"			char varying (40) not null 									-- FK to t_runSet
	, "seq"					bigint DEFAULT nextval('t_run_id_seq'::regclass) NOT NULL 	-- ID for passing to client as a number
	, "useragent_id" 		char varying(40) not null 									-- FK t_userAgent table fk
	, "created" 			timestamp default current_timestamp not null 		
);

create index "t_runSet_membes_p1"  on "t_runSet_members" ( "runset_id", "seq" );

insert into "t_runSet" ( "id", "user_id", "name" ) values ( '100', '1', 'Standard Browsers' );
	insert into "t_runSet_members" ( "runset_id", "useragent_id" ) values ( '100', '200' );
	insert into "t_runSet_members" ( "runset_id", "useragent_id" ) values ( '100', '202' );
	insert into "t_runSet_members" ( "runset_id", "useragent_id" ) values ( '100', '203' );
	insert into "t_runSet_members" ( "runset_id", "useragent_id" ) values ( '100', '204' );
	insert into "t_runSet_members" ( "runset_id", "useragent_id" ) values ( '100', '205' );
	insert into "t_runSet_members" ( "runset_id", "useragent_id" ) values ( '100', '207' );
insert into "t_runSet" ( "id", "user_id", "name" ) values ( '101', '1', 'Linux Test' );
	insert into "t_runSet_members" ( "runset_id", "useragent_id" ) values ( '101', '200' );
	insert into "t_runSet_members" ( "runset_id", "useragent_id" ) values ( '101', '201' );


-- -------------------------------------------------------- -- --------------------------------------------------------
-- Insertions handled by the AddjobAction class.
--	, "project_id" 			char varying(40)  not null 							-- FK to t_projects.id field.
CREATE TABLE "t_job" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "user_id"				char varying (40) not null 				-- fk to t_user
	, "name" 				char varying(255)  not null 						-- Job name (can contain HTML). (the run_name)
	, "runset_id" 			char varying(40)  not null 							-- FK to t_runset.id field.
   	, "user_data" 			char varying(255) 		 							-- user supplied data
  	, "url" 				char varying(255) not null	 						-- Run url
	, "test_type" 			char varying(40)  									-- Interctive, Library etc.
	, "priority"			int default 10 not null								-- 10 is standard priority ( higher is higher, lower is lower )
	, "created" 			timestamp default current_timestamp not null 		
);

--CREATE INDEX "t_jobs_project_created_p1" ON "t_job" ("project_id", "created"); 				-- Usage: ProjectAction.
--CREATE INDEX "t_jobs_project_created_p2" ON "t_job" ("project_id", "created", "priority"); 	-- Usage: ProjectAction.

insert into "t_job" (
	  "id"				
	, "user_id"		
	, "name" 				
	, "runset_id" 		
   	, "user_data" 	
  	, "url" 				
	, "test_type" 		
	, "priority"	
	, "created" 
) values (
	'2858e27b-a098-4af5-c5b5-e5d9300427b2'
	,'1'
	,'Samp-Test-001'
	,'101'
	,null
	,'http://localhost/testswarm/qunit-example1.html'
	,'User Interface'
	,'10'
	,'2013-07-16 12:58:53.117021'
);


-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE SEQUENCE t_run_id_seq
  INCREMENT 1
  MINVALUE 1
  MAXVALUE 9223372036854775807
  START 1
  CACHE 1;

-- -------------------------------------------------------- -- --------------------------------------------------------
-- A run needs
--	1. x_run_id (number/seq)
--	2. auth_key		- random hash for saving data
--	3. user_data 	- string
--	4. URL for data
--	5. user_auth_key - passwrod for running tests in user's code
--	6. host_id		- not transmitted to user
--	7. client_id	- ID of the Client that is running this
create table "t_a_run" ( 
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "user_id"				char varying (40) not null 									-- fk to t_user
	, "c_run_id"			bigint DEFAULT nextval('t_run_id_seq'::regclass) NOT NULL 		-- ID for passing to client as a number
	, "x_run_id"			bigint DEFAULT nextval('t_run_id_seq'::regclass) NOT NULL 		-- ID for passing to client as a number
  	, "job_id" 				char varying(40) not null 									-- FK to t_job.id field.
  	, "client_id" 			char varying(40)  											-- FK to t_client.id field.
  	, "client_hash" 		char varying(255)  											-- key for matching with a client
  	, "host_id" 			char varying(40)  											--	
  	, "useragent_id" 		char varying(40) not null 						 			-- Key to config.userAgents property.
  	, "max" 				int not null default 1 										-- Addjob runMax
  	, "n_runs" 				int not null default 0 										-- Addjob runMax
	, "run_timeout"			int not null default 240									-- amoutn of time before timout if running
	, "io_timeout"			int not null default 240									-- amoutn of time before timout if waiting for user input
	, "event_timeout"		int not null default 240									-- amoutn of time before timout if waiting for event to occure
	, "width"				int not null default 1000									-- screen resolution to use
	, "height"				int not null default 600									-- screen resolution to use
  	, "completed" 			int not null default 0 										-- Number of times this run has run to completion for this user agent.
  	, "status" 				char varying(40) null default 'init'
								check ( "status" in (
									  'init'											-- Create State, waiting to be picked
									, 'picked'											-- Ok it has been picked to be run
									, 'busy'											-- Ping has occured
									, 'finished'										-- Successful test completed
									, 'timeout-max-execution'							-- Took too long to run
									, 'user-input-wait'									-- Waiting for userinput 
									, 'sys-input-wait'									-- Waiting for screen capture or other system event
									, 'event-wait'										-- Unspecifed event wait
									, 'timeout-client-lost'								-- Fialed 
									, 'timeout-waiting-for-input'						-- Fialed 
									, 'timeout-waiting-for-event'						-- Fialed 
								) )
  	, "run_name" 			char varying(255) not null						 			-- Run name
  	, "url" 				char varying(255) not null	 								-- Run url
  	, "user_data" 			char varying(255) 		 									-- user supplied data
	, "priority"			int default 10 not null										-- 10 is standard priority ( higher is higher, lower is lower )
  	, "ex_time" 			int not null default 0							 			-- Exeuction time (relevant for library tests)
  	, "total" 				int not null default 0							 			-- Total number of tests that were run.
  	, "fail" 				int not null default 0 										-- Number of failed tests.
  	, "error" 				int not null default 0 										-- Number of errors.
  	, "report_html" 		text 														-- HTML snapshot of the test results page - gzipped.
  	, "console_log" 		text 														-- What was on the console
  	, "auth_token" 			char varying (80) not null			 						-- Hash(sha256) of user_id and other stuff. To use as authentication to be allowed to
																						-- store runresults in this row. This protects from bad insertions.
  	, "run_picked" 			timestamp  										
  	, "updated" 			timestamp  										
  	, "created" 			timestamp default current_timestamp not null 
);

create index "t_a_run_p1" on "t_a_run" ( "x_run_id", "auth_token" );
-- create index "t_a_run_p2" on "t_a_run" ( "status", "priority", "created" );
create index "t_a_run_p3" on "t_a_run" ( "user_id", "job_id", "c_run_id", "updated" );
-- create index "t_a_run_p4" on "t_a_run" ( "client_id", "host_id", "x_run_id" );
-- create index "t_a_run_p5" on "t_a_run" ( "client_hash", "priority", "created" );
create index "t_a_run_p6" on "t_a_run" ( "status", "client_id", "client_hash", "priority", "created" );

-- -------------------------------------------------------- -- --------------------------------------------------------
create table "t_a_run_periodic" ( 
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "run_id"				char varying (40) 											-- FK to "t_a_run"
	, "next_run_after"		timestamp
  	, "run_started" 		timestamp  										
  	, "run_finished" 		timestamp  										
  	, "status" 				char varying(40) null default 'available'
								check ( "status" in (
									  'available'										-- Create State, waiting to be picked
									, 'busy'											-- Run has been stared
									, 'finished'										-- Successful test completed
								) )
	, "interval_forward"	char varying (40) default '1 day'
								check ( "interval_forward" in (
									  '1 hour'
									, '1 day'	
									, '1 week'	
									, '1 month'	
								) )
);
create index "t_a_run_periodic_p1" on "t_a_run_periodic" ( "next_run_after", "status" );

-- -------------------------------------------------------- -- --------------------------------------------------------
create table "t_a_run_coverage" ( 
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "user_id"				char varying (40) not null 									-- fk to t_user
	, "a_run_id"			char varying (40) not null 									-- FK to t_a_run
  	, "at" 					int not null 												-- line number
  	, "file_name"			char varying (100) not null 								-- 
);

create table "t_need_coverage" ( 
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "user_id"				char varying (40) not null 									-- fk to t_user
  	, "at" 					int not null 												-- line number
  	, "file_name"			char varying (100) not null 								-- 
);

-- -------------------------------------------------------- -- --------------------------------------------------------
-- http://test.2c-why.com/link-to-test?token=jfkdfjkdfjkdjf - put in iframe and it will run report and return results.
-- May need to add last time run - if to many runs
-- May need to cache retults
-- May track number of runs
-- -------------------------------------------------------- -- --------------------------------------------------------
create table "t_link_to_test" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "url_of_test"			char varying (255) not null				-- the URL that the user will click on to run the test
	, "auth_token"			char varying (40) not null 				-- A validated auth_token
	, "user_id"				char varying (40) not null 				-- User that this test belongs to
	, "job_id"				char varying (40) not null 				-- Job to be run
	, "n_runs"				int default 0 not null
);




-- -------------------------------------------------------- -- --------------------------------------------------------
create table "t_it_work"  (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "host_id"				char varying (40) not null 
	, "status"				char varying (40) default 'init'
								check ( "status" in (
									  'init'											-- Create State, waiting to be picked
									, 'picked'											-- Ok it has been picked to be run
									, 'busy'											-- Ping has occured
									, 'finished'										-- Successful test completed
									, 'timeout'	
								) )
	, "hostname"			char varying (80) 
	, "cmd"					char varying (80) 
	, "params"				text														-- JSON data - should really use the extention
	, "data"				text														-- JSON data - Results of run when status == 'finished'
	, "xcmd"				char varying (80) 
	, "run_id"				char varying (40)											-- "t_a_run"."id" if this is from the UI/script running
	, "seq"					bigint DEFAULT nextval('t_run_id_seq'::regclass) NOT NULL 	-- ID for passing to client as a number
  	, "updated" 			timestamp  										
  	, "created" 			timestamp default current_timestamp not null 
);

create index "t_it_work_p1" on "t_it_work" ( "status", "hostname", "seq" );


-- -------------------------------------------------------- -- --------------------------------------------------------
-- Currently used only for storing data about the current IP address of the Garage DC.
-- -------------------------------------------------------- -- --------------------------------------------------------
create table "t_config_data"  (
	  "name"					char varying (40) 
	, "value"					char varying (250) 
);
delete from "t_config_data";
insert into "t_config_data"  ( "name" , "value" ) values ( 'host-ip', '0.0.0.0' );
insert into "t_config_data"  ( "name" , "value" ) values ( 'i-am', 'dev' );
-- insert into "t_config_data"  ( "name" , "value" ) values ( 'i-am', 'prod' );



-- -------------------------------------------------------- -- --------------------------------------------------------
-- dn.runQuery ( stmt = ts0( 'insert /*l575*/ into "t_ua_log" ( "ua", "ip", "user_no_id" ) values ( \'%{ua%}\', \'%{ip%}\', \'%{client_id%}\' )"
-- -------------------------------------------------------- -- --------------------------------------------------------
create table "t_ua_log"  (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "ua"					text
	, "ip"					char varying (40) 
	, "user_no_id"			char varying (40) 
	, "seq"					bigint DEFAULT nextval('t_run_id_seq'::regclass) NOT NULL 	-- ID for passing to client as a number
  	, "updated" 			timestamp  										
  	, "created" 			timestamp default current_timestamp not null 
);





-- -------------------------------------------------------- -- --------------------------------------------------------
-- dn.runQuery ( stmt = ts0( 'insert into t_link_valid ( "user_id", "url" ) values ( \'%{user_id%}\', \'%{url%}\' )'
-- -------------------------------------------------------- -- --------------------------------------------------------
create table "t_link_valid"  (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "user_id"				char varying (40) 
	, "url"					char varying (2000)
  	, "updated" 			timestamp  										
  	, "created" 			timestamp default current_timestamp not null 
);







-- -------------------------------------------------------- -- --------------------------------------------------------
-- -------------------------------------------------------- -- --------------------------------------------------------
drop table "t_dispatch_scripts";
create table "t_dispatch_scripts" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "dev_or_prod"			char varying (5) default 'prod'
	, "script_to_run"		char varying (80)
	, "hosted_at"			char varying (80)
	, "hostname"			char varying (80)
	, "host_type"			char varying (80)
	, "ip_of_host"			char varying (80)
	, "mon_url"				char varying (80)
	, "osFamily"			char varying (80)
	, "osMajor"				char varying (80)
	, "osMinor"				char varying (80)
	, "script_desc"			char varying (80)
	, "url_of_host"			char varying (80)
);
create unique index "t_dispatch_scripts_u1" on "t_dispatch_scritps" ( "dev_or_prod", "script_to_run" );

insert into "t_dispatch_scripts" ( "dev_or_prod", "script_to_run", "hosted_at", "hostname", "host_type", "ip_of_host", "mon_url", "osFamily", "osMajor", "osMinor", "script_desc", "url_of_host" ) values
( 'prod',  'chk-link', 'linode.com', 'peach', 'linode.com:ubuntu', '198.58.107.206', 'http://test.2c-why.com:3050/who-cares.html', 'Linux', '12', '04', 'Check Links in HTML', 'http://test.2c-why.com:3050/who-cares.html' );

insert into "t_dispatch_scripts" ( "dev_or_prod", "script_to_run", "hosted_at", "hostname", "host_type", "ip_of_host", "mon_url", "osFamily", "osMajor", "osMinor", "script_desc", "url_of_host" ) values
( 'prod',  'check-html', 'Garage-DC', 'grape', 'test-system:ubuntu', '172.16.0.167', 'http://192.168.0.151:3050/who-cares.html', 'Linux', '12', '04', 'Check HTML for Validity', 'http://172.16.0.167/it.html' );

insert into "t_dispatch_scripts" ( "dev_or_prod", "script_to_run", "hosted_at", "hostname", "host_type", "ip_of_host", "mon_url", "osFamily", "osMajor", "osMinor", "script_desc", "url_of_host" ) values
( 'prod',  'check-css', 'Garage-DC', 'grape', 'test-system:ubuntu', '172.16.0.167', 'http://192.168.0.151:3050/who-cares.html', 'Linux', '12', '04', 'Check CSS for Validity', 'http://172.16.0.167/it.html' );

insert into "t_dispatch_scripts" ( "dev_or_prod", "script_to_run", "hosted_at", "hostname", "host_type", "ip_of_host", "mon_url", "osFamily", "osMajor", "osMinor", "script_desc", "url_of_host" ) values
( 'prod',  'check-js', 'Garage-DC', 'grape', 'test-system:ubuntu', '172.16.0.167', 'http://192.168.0.151:3050/who-cares.html', 'Linux', '12', '04', 'Check JS for common errors', 'http://172.16.0.167/it.html' );



insert into "t_dispatch_scripts" ( "dev_or_prod", "script_to_run", "hosted_at", "hostname", "host_type", "ip_of_host", "mon_url", "osFamily", "osMajor", "osMinor", "script_desc", "url_of_host" ) values
( 'dev',  'chk-link', 'Garage-DC', 'pschlump-dev1', 'ubuntu', '192.168.0.151', 'http://192.168.0.151:3050/who-cares.html', 'Linux', '12', '04', 'Check Links in HTML', 'http://192.168.0.151:3050/who-cares.html' );

insert into "t_dispatch_scripts" ( "dev_or_prod", "script_to_run", "hosted_at", "hostname", "host_type", "ip_of_host", "mon_url", "osFamily", "osMajor", "osMinor", "script_desc", "url_of_host" ) values
( 'dev',  'check-html', 'Garage-DC', 'grape', 'test-system:ubuntu', '172.16.0.167', 'http://192.168.0.151:3050/who-cares.html', 'Linux', '12', '04', 'Check HTML for Validity', 'http://172.16.0.167/it.html' );

insert into "t_dispatch_scripts" ( "dev_or_prod", "script_to_run", "hosted_at", "hostname", "host_type", "ip_of_host", "mon_url", "osFamily", "osMajor", "osMinor", "script_desc", "url_of_host" ) values
( 'dev',  'check-css', 'Garage-DC', 'grape', 'test-system:ubuntu', '172.16.0.167', 'http://192.168.0.151:3050/who-cares.html', 'Linux', '12', '04', 'Check CSS for Validity', 'http://172.16.0.167/it.html' );

insert into "t_dispatch_scripts" ( "dev_or_prod", "script_to_run", "hosted_at", "hostname", "host_type", "ip_of_host", "mon_url", "osFamily", "osMajor", "osMinor", "script_desc", "url_of_host" ) values
( 'dev',  'check-js', 'Garage-DC', 'grape', 'test-system:ubuntu', '172.16.0.167', 'http://192.168.0.151:3050/who-cares.html', 'Linux', '12', '04', 'Check JS for common errors', 'http://172.16.0.167/it.html' );

insert into "t_dispatch_scripts" ( "dev_or_prod", "script_to_run", "hosted_at", "hostname", "host_type", "ip_of_host", "mon_url", "osFamily", "osMajor", "osMinor", "script_desc", "url_of_host" ) values
( 'dev',  'some-win-only-thing', 'Garage-DC', 'grape', 'test-system:win7', '172.16.0.167:5022', 'http://192.168.0.151:3050/who-cares.html', 'Windows', '07', '00', 'Check some silly thing on windows', 'http://172.16.0.167:5022/it.html' );



-- -------------------------------------------------------- -- --------------------------------------------------------
--		dn.runQuery ( stmt = ts0( 'select /*vWorkToDo.sql*/ it_find_work ( \'%{hostname%}\' ) as "id" ', { "hostname":hostname } ), function ( err, result ) {
-- Doc on JSON in pgsql
--     http://www.postgresql.org/docs/devel/static/functions-json.html 
--
--	/* xyzzy - this is the spot - should check in t_a_run for "client_hash" = 'linux-script'
--
--		"t_it_work"."cmd" = "t_a_run"."user_data"
--		"t_it_work"."params" = "t_a_run"."url"
--			+ need the "t_a_run".id - so can send back compltion data
--
--		user_data has to have
--			"url"				-- the URL of the site to be checked (check-link)
--			"script"			-- the script to be run
--
--		http://www.postgresql.org/docs/9.1/static/plpgsql-control-structures.html	
--
--	*/
-- -------------------------------------------------------- -- --------------------------------------------------------

-- function with exception block to be called later
CREATE OR REPLACE FUNCTION f_insert_test_insert(
    id integer,
    col1 double precision,
    col2 text
)
RETURNS void AS
$body$
BEGIN
    INSERT INTO insert_test
    VALUES ($1, $2, $3)
    ;
EXCEPTION
    WHEN unique_violation
    THEN NULL;
END;
$body$
LANGUAGE plpgsql;


CREATE or REPLACE FUNCTION create_user ( p_username varchar, p_password varchar, p_name varchar, p_email varchar, p_group_id varchar
	, p_acct_state varchar, p_ip_addr varchar, p_n_days varchar, p_email_confirmed varchar, p_privs varchar ) RETURNS varchar AS $$
DECLARE
    l_salt char varying(80);
    l_user_id char varying(40);
    l_days char varying(40);
    l_auth_token char varying(40);
    l_expire timestamp;
    l_intval interval;
BEGIN
	l_salt = uuid_generate_v4();
	l_user_id = uuid_generate_v4();
	l_days = p_n_days||' days';
	-- l_intval = interval '90000 days';
	l_intval = interval l_days;
	l_expire = current_timestamp + l_intval;
	l_auth_token = uuid_generate_v4();
// -- sha256 - and salt update. xyzzy
	BEGIN
		insert into "t_user" ( 
			  "id"
			, "group_id"
			, "username"
			, "real_name"
			, "password"
			, "email_address"
			, "acct_state"
			, "auth_token"
			, "ip"
			, "email_confirmed"
			, "acct_expire"
			, "privs"
			, "salt"
		) values ( 
			  l_user_id
			, p_group_id
			, p_username
			, p_name									-- name in real life
			, sha256pw((l_salt||p_password)::bytea)
			, p_email
			, p_acct_state
			, sha256pw(l_auth_token::bytea)
			, p_ip_addr									-- IP Address
			, p_email_confirmed
			, l_expire
			, p_privs
			, l_salt
		 );
	EXCEPTION
		WHEN unique_violation THEN 
			l_user_id = 'error';
	END;
	return l_user_id;
END;
$$ LANGUAGE plpgsql;

select create_user (
	  'fool2'					-- Fool2 
	, '882211Fool!!2'			-- Fool2 password
	, 'A Fool'
	, 'pschlump@gmail.com'
	, '1'						-- group_id
	, 'fixed'					-- acct_state
	, '0.0.0.0'					-- Account IP address register from
	, '90000'					-- Days till expire
	, 'y'						-- Email Confirmed (normally 'n')
	, '{"who_cares":["root"],"it":["root"], "image":["root"]}'
);






CREATE or REPLACE FUNCTION is_valid_user ( p_username varchar, p_password varchar ) RETURNS varchar AS $$
DECLARE
    l_password char varying(80);
    l_salt char varying(80);
    l_id char varying(40);
BEGIN
	l_id = 'test';
	-- insert into "t_output" ( "msg" ) values ( 'at top' );
	select "password", "salt", "id"
		into l_password
			, l_salt
			, l_id
		from "t_user"
		where "username" = p_username
		;
	IF NOT FOUND THEN
		l_id = 'nope';
		-- insert into "t_output" ( "msg" ) values ( 'not found' );
	ELSE
		-- insert into "t_output" ( "msg" ) values ( 'found it' );
		-- insert into "t_output" ( "msg" ) values ( 'l_salt='||l_salt );
		-- insert into "t_output" ( "msg" ) values ( 'l_password='||l_password );
		-- insert into "t_output" ( "msg" ) values ( 'sha256='||sha256pw((l_salt||l_password)::bytea) );
		-- insert into "t_output" ( "msg" ) values ( 'p_password='||p_password );
		-- See:  https://crackstation.net/hashing-security.htm 
		IF sha256pw((l_salt||p_password)::bytea) != l_password THEN
			l_id = 'nope';
			-- insert into "t_output" ( "msg" ) values ( 'no match' );
		-- ELSE
			-- insert into "t_output" ( "msg" ) values ( 'GOT IT!' );
		END IF;
	END IF;

	RETURN l_id;
END;
$$ LANGUAGE plpgsql;


-- delete from "t_output";
select is_valid_user ( 'fool2', '882211Fool!!2' );
-- select "msg", "seq" from "t_output" order by "seq";

-- delete from "t_output";
select is_valid_user ( 'fool2', '882211Fool!!x' );
-- select "msg", "seq" from "t_output" order by "seq";



CREATE or REPLACE FUNCTION it_find_work ( p_hostname varchar, p_user_id varchar ) RETURNS varchar AS $$
DECLARE
    work_id char varying(40);
    p_run_id char varying(40);
	data1 record;
	p_host_id varchar(40);
	p_client_id varchar(40);
	p_useragent_id varchar(40);
	l_hostname varchar(40);
	p_host_url varchar(200);
	p_url_of_host varchar(200);
	p_ip_of_host varchar(40);
	p_host_type varchar(40);
	p_hosted_at varchar(100);
	p_osFamily varchar(40);
	p_osMajor varchar(40);
	p_osMinor varchar(40);
	p_script_to_run varchar(200);
	p_mon_url varchar(200);
	p_auth_token varchar(40);
	p_script_desc varchar(200);
    l_cmd char varying(80);
	l_i_am varchar(40);
BEGIN

	p_host_id = 'x';
	p_client_id = 'x';
	p_useragent_id = 'x';

	select "value"
		into l_i_am
		from "t_config_data"
		where "name" = 'i-am'
	;
	IF NOT FOUND THEN
		l_i_am = 'prod';
	END IF;

	-- insert into "t_output" ( "msg" ) values ( 'it_find_work:800:before hostname=' );
	-- insert into "t_output" ( "msg" ) values ( p_hostname );

	update "t_monitor_stuff"
		set
		  "timeout_event" = current_timestamp + CAST("delta_t" as Interval) 
		where "item_name" = 'It-On:'||p_hostname
		;

	-- insert into "t_output" ( "msg" ) values ( 'it_find_work:860:hostname='||p_hostname );


	FOR data1 IN
		select
			  "id"
			, "url"
			, "user_data"
			, "auth_token"
		from "t_a_run"
		where "status" = 'init'
		 and "client_hash" like 'script-%'
		order by "priority" desc, "created" 
		for update
	LOOP
		-- insert into "t_output" ( "msg" ) values ( 'it_find_work:889:loop top' );

		-- Worded in dev and prod.
		-- http://pgxn.org/dist/json_accessors/doc/json_accessors.html 
		-- https://github.com/theirix/json_accessors 
			-- 1st install the tool for doing installs
			-- $ sudo easy_install pgxnclient
		-- $ pgxn --pg_config <postgresql_install_dir>/bin/pg_config install json_accessors	
		-- $ pgxn install json_accessors	
		-- Doc:  http://pgxn.org/dist/json_accessors/doc/json_accessors.html 
		-- #db# CREATE EXTENSION json_accessors; 

		--  \pset pager off 

		-- 9.3: p_script_to_run = json_extract_path ( data1.user_data, 'script' );
		select json_get_text(data1.user_data, 'script')	
			into p_script_to_run
			;

		select
				  "hosted_at"
				, "hostname"
				, "host_type"
				, "ip_of_host"
				, "mon_url"
				, "osFamily"
				, "osMajor"
				, "osMinor"
				, "script_desc"
				, "url_of_host"
			into 
				  p_hosted_at 			--  'Garage-DC';
				, l_hostname 			--  'grape';
				, p_host_type 			--  'test-system:win7';
				, p_ip_of_host 			--  '172.16.0.167:5022';
				, p_mon_url 			--  'http://192.168.0.151:3050/who-cares.html';
				, p_osFamily 			--  'Windows';
				, p_osMajor 			--  '07';
				, p_osMinor 			--  '00';
				, p_script_desc 		--  'Check some silly thing on windows';
				, p_url_of_host 		--  'http://172.16.0.167:5022/it.html';
			from "t_dispatch_scripts" 
			where script_to_run = p_script_to_run
				and dev_or_prod = l_i_am
			;
		IF NOT FOUND THEN
			RAISE EXCEPTION 'invalid script name, script_to_run="%", i_am="%"', p_script_to_run, l_i_am;
		END IF;

		if ( p_host_id = 'x' ) then
			select "id"
				into p_host_id
				from "t_host"
				where "hostname" = l_hostname
				;
			IF NOT FOUND THEN
				p_host_id = uuid_generate_v4();
				-- -------------------------------------------------------- -- --------------------------------------------------------
				-- Data for systems - this will need to be more robust later when we need to distribute activity across boxes.
				-- -------------------------------------------------------- -- --------------------------------------------------------
				insert into "t_host" ( "id", "url_of_it", "hostname", "host_type", "is_running_now", "ip_addr", "hosted_at", "osFamily", "osMajor", "osMinor" )
					values ( p_host_id, p_url_of_host, l_hostname, p_host_type, 'y', p_ip_of_host, p_hosted_at, p_osFamily, p_osMajor, p_osMinor );
			END IF;
		end if;
		if ( p_useragent_id = 'x' ) then
			select "id"
				into p_useragent_id
				from "t_userAgent"
				where "name" = p_script_to_run
				;
			IF NOT FOUND THEN
				p_useragent_id = uuid_generate_v4();
				insert into "t_userAgent" ( "id", "name", "title", "osFamily", "osMajor", "osMinor" )
					values ( p_useragent_id, p_script_to_run, p_script_desc, p_osFamily, p_osMajor, p_osMinor );
			END IF;
		end if;
		if ( p_client_id = 'x' ) then
			select "id"
				into p_client_id
				from "t_client"
				where "name" = p_script_to_run
				;
			IF NOT FOUND THEN
				p_client_id = uuid_generate_v4();
				insert into "t_client" ( "id", "host_id", "name", "useragent_id", "useragent", "ip" )
					values ( p_client_id, p_host_id, p_script_to_run, p_useragent_id, p_script_to_run,  p_ip_of_host );
			END IF;
		end if;

		update "t_a_run" 
			set "status" = 'picked'
			  , "client_id" = p_client_id
			  , "host_id" = p_host_id
			  , "useragent_id" = p_useragent_id
			  , "run_picked" = current_timestamp
			where "id" = data1."id"
			;
		/* after 30 minutes idle a host can be stoped. */
		update "t_host"
			set "last_run_at" = current_timestamp
			,	"n_test_run" = "n_test_run" + 1
			where "id" = p_host_id
			;

		/* after 15 minutes idle a client can be stoped. */
		update "t_client"
			set "last_run_at" = current_timestamp
			,	"n_test_run" = "n_test_run" + 1
			where "id" = p_client_id
			;

		insert into "t_it_work" (
			  "host_id"			
			, "hostname"	
			, "cmd"		
			, "params"				
			, "data"			
			, "run_id"			
			, "xcmd"
		) values (
			  p_host_id
			, l_hostname
			, 'run-cmd'
			, data1."user_data"
			, '{"run_id":"'||data1."id"||'","url":"'||data1.url||'","auth_token":"'||data1.auth_token||'"}'				-- xyzzy - URL, Params
			, data1."id"
			, p_script_to_run
		);
		--	, "xcmd"		-- Looks to be not used.
	END LOOP;

	-- insert into "t_output" ( "msg" ) values ( 'it_find_work:1033:after loop' );

	select id, run_id, cmd
	into work_id, p_run_id, l_cmd
		from "t_it_work" 
		where "status" = 'init'
		 and "hostname" = p_hostname
		order by "seq"
		for update 
		limit 1 
		;

	-- insert into "t_output" ( "msg" ) values ( 'it_find_work:1045:did select for work' );

	IF NOT FOUND THEN
		work_id = 0;
	ELSE
		if l_cmd = 'api/self-terminate' then
			delete from "t_monitor_stuff"
				where "item_name" = 'It-On:'||p_hostname
				;
		end if;
	END IF;

	update "t_it_work"
		set "status" = 'picked'
		where id = work_id
		;

	-- insert into "t_output" ( "msg" ) values ( 'it_find_work:1062:marked work as picked' );

	if ( p_run_id is not null ) then
		update "t_a_run"
			set "status" = 'busy'
			where id = p_run_id
			;
	end if;

	RETURN work_id;
END;
$$ LANGUAGE plpgsql;

-- select /*vWorkToDo.sql*/ it_find_work ( 'grape', '0' ) as "id";
-- select msg from t_output;




-- -------------------------------------------------------- -- --------------------------------------------------------
--
-- Called on completion of work by *it*
-- From approx line:3775 in monitor.m4.js
--
-- -------------------------------------------------------- -- --------------------------------------------------------
--
--	//dn.runQuery ( stmt = ts0( 'update "t_it_work" set "status" = \'finished\', "data" = \'%{data%}\'  where "id" = \'%{id%}\' '
--	//, { "id":id, "data":JSON.stringify(req.params) } )
--	dn.runQuery ( stmt = ts0( 'select /*l:3775*/ it_work_done ( \'%{data%}\', \'%{id%}\' ) '
--	, { "id":id, "data":JSON.stringify(req.params) } )
--
-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE or REPLACE FUNCTION it_work_done ( p_data varchar, p_it_work_id varchar ) RETURNS varchar AS $$
DECLARE
    l_run_id char varying(40);
	l_status varchar(40);
	l_total varchar(40);
	l_fail varchar(40);
	l_error varchar(40);
	l_auth_token varchar(40);
	l_report_html varchar(2000);
	l_itotal int;
	l_ifail int;
	l_ierror int;
BEGIN

	update "t_it_work" set "status" = 'finished', "data" = p_data  where "id" = p_it_work_id;

	select "run_id"
		into l_run_id
		from "t_it_work"
	 	where "id" = p_it_work_id
	;
		
	IF NOT FOUND THEN
		l_run_id = 0;
	ELSE
		if l_run_id is null then
			l_run_id = 0;
		else
			-- 9.3: l_status = json_extract_path ( p_data, 'status' );
			select json_get_text( p_data, 'status')	into l_status ;
			if ( l_status = 'success' ) then
				l_status = 'finished';
			end if;
			-- 9.3: l_total = json_extract_path ( p_data, 'total' );
			select json_get_text( p_data, 'total')	into l_total ;
			-- 9.3: l_fail = json_extract_path ( p_data, 'fail' );
			select json_get_text( p_data, 'fail')	into l_fail ;
			-- 9.3: l_error = json_extract_path ( p_data, 'error' );
			select json_get_text( p_data, 'error')	into l_error ;
			-- 9.3: l_report_html = json_extract_path ( p_data, 'report_html' );
			select json_get_text( p_data, 'report_html')	into l_report_html ;
			-- 9.3: l_auth_token = json_extract_path ( p_data, 'auth_token' );
			select json_get_text( p_data, 'auth_key')	into l_auth_token ;

			l_itotal = cast(l_total as int);
			l_ifail = cast(l_fail as int);
			l_ierror = cast(l_error as int);

			-- insert into "t_output" ( "msg" ) values ( 'l_itotal='||l_itotal );
			-- insert into "t_output" ( "msg" ) values ( 'l_run_id='||l_run_id );
			-- insert into "t_output" ( "msg" ) values ( 'l_auth_token='||l_auth_token );
			-- insert into "t_output" ( "msg" ) values ( 'l_status='||l_status );

			update "t_a_run"
				set
					  "completed" = "completed" + 1 
					, "status" = l_status
					, "total" = l_itotal
					, "fail" = l_ifail
					, "error" = l_ierror
					, "report_html" =  l_report_html
				where "id" = l_run_id
				  and "auth_token" = l_auth_token
			;
		end if;
	END IF;

	RETURN l_run_id;
END;
$$ LANGUAGE plpgsql;












-- -------------------------------------------------------- -- --------------------------------------------------------
--
-- Tables used for the transfer/sync of data from/to the master database on Linode - with local data centers.
--
-- 1. start with test data table t_ms_test
-- 2. Add in monetering data
-- 3. Add in test system data / request test / udpate results
--
-- -------------------------------------------------------- -- --------------------------------------------------------
--create table "t_ms_test" (
--	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
--	, "seq"					bigint DEFAULT nextval('t_run_id_seq'::regclass) NOT NULL 	-- ID for passing to client as a number
--	, "data"				char varying (100)
--  	, "updated" 			timestamp  										
--  	, "created" 			timestamp default current_timestamp not null 
--);
--
--create table "t_get_row" (
--	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
--	, "seq"					bigint DEFAULT nextval('t_run_id_seq'::regclass) NOT NULL 	-- ID for passing to client as a number
--	, "table_name"			char varying(50) not null
--	, "when_data_ready"		timestamp default current_timestamp not null
--	, "row_id"				char varying(40) not null
--	, "row_hash_code"		char varying(40) not null
--	, "row_data"			text									-- Insert statment to run
--	, "transfer_started"	char varying(1) default 'n'
--	, "transfer_checked"	char varying(1) default 'n'				-- n, y, E - for errors
--	, "transfer_set"		char varying(40)
--	, "error_msg"			char varying(255)
--	, "for_dc"				char varying(20) default 'garage'		-- Destination data center
--);
--
--create table "t_post_results_back" (
--	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
--	, "seq"					bigint DEFAULT nextval('t_run_id_seq'::regclass) NOT NULL 	-- ID for passing to client as a number
--	, "table_name"			char varying(50) not null
--	, "when_data_ready"		timestamp default current_timestamp not null
--	, "row_id"				char varying(40) not null
--	, "row_hash_code"		char varying(40) not null
--	, "row_data"			text									-- Insert or Update statment to run
--	, "transfer_started"	char varying(1) default 'n' not null
--	, "transfer_checked"	char varying(1) default 'n' not null
--	, "transfer_set"		char varying(40)
--	, "error_msg"			char varying(255)
--	, "call_func_on_data"	char varying(1) default 'n' not null
--	, "func_name"			char varying(255)
--	, "from_dc"				char varying(20) default 'garage'		-- Name of the DC that the data originates is from
--);
--
--



-- -------------------------------------------------------- -- --------------------------------------------------------
--old-create or replace view t_available_test_systems
--old-as
--old-select 
--old-	  lower(replace(t3."osFamily",' ','_') as "osNameClass"
--old-	, t3."browserFamily" as "browserNameClass"
--old-	, t3."osMajor" as "osMajorClass"
--old-	, min(t3."osMinor") as "osMinorClass"
--old-	, t3."browserMajor" as "browserMajorClass"
--old-	, min(t3."browserMinor") as "browserMinorClass"
--old-	, t3."name" as "browserName"
--old-	, t3."name" as "title"
--old-	, count(1) as "n_clients"
--old-	, sum(t1."n_test_run") as "n_runs"
--old-	, t3."id" as "useragent_id"
--old-	, t2."is_running_now"
--old-from "t_userAgent" t3
--old-	, "t_host_can_run" t2
--old-	, "t_host" t1
--old-where t1."id" = t2."host_id"
--old-  and t2."useragent_id" = t3."id"
--old-group by t2."useragent_id"
--old-	,  t3."osFamily" 
--old-	, t3."browserFamily" 
--old-	, t3."osMajor" 
--old-	, t3."browserMajor" 
--old-	, t3."name" 
--old-	, t3."id" 
--old-;
create or replace view t_available_test_systems
as
select 
	  t3."osFamily" as "osNameClass"
	, t3."browserFamily" as "browserNameClass"
	, t3."osMajor" as "osMajorClass"
	, min(t3."osMinor") as "osMinorClass"
	, t3."browserMajor" as "browserMajorClass"
	, min(t3."browserMinor") as "browserMinorClass"
	, t3."name" as "browserName"
	, t3."name" as "title"
	, count(1) as "n_clients"
	, sum(t1."n_test_run") as "n_runs"
	, t3."id" as "useragent_id"
	, t2."is_running_now"
from "t_userAgent" t3
	, "t_host_can_run" t2
	, "t_host" t1
where t1."id" = t2."host_id"
  and t2."useragent_id" = t3."id"
group by t2."useragent_id"
	,  t3."osFamily" 
	, t3."browserFamily" 
	, t3."osMajor" 
	, t3."browserMajor" 
	, t3."name" 
	, t3."id" 
	, t2."is_running_now"
order by t2."is_running_now" desc , 1, 2
;


-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE or REPLACE FUNCTION generate_run(p_job_id varchar, p_user_id varchar ) RETURNS integer AS $$
DECLARE
    common_seq INTEGER;
    auth_token char varying(40);
    l_test_type char varying(40);
    l_url char varying(1000);
	l_user_data varchar(2000);
	l_useragent_id char varying(40);
	l_host_id varchar(40);
	l_client_id varchar(40);
	l_hostname varchar(40);
	l_host_url varchar(200);
	l_url_of_host varchar(200);
	l_ip_of_host varchar(40);
	l_host_type varchar(40);
	l_hosted_at varchar(100);
	l_osFamily varchar(40);
	l_osMajor varchar(40);
	l_osMinor varchar(40);
	l_script_to_run varchar(200);
	l_mon_url varchar(200);
	l_auth_token varchar(40);
	l_script_desc varchar(200);
    l_cmd char varying(80);
	l_i_am varchar(40);
BEGIN

	SELECT nextval('t_run_id_seq'::regclass) INTO common_seq;

	IF NOT FOUND THEN
		RAISE EXCEPTION 'sequence % broken', myname;
	END IF;

// -- sha256 - and salt update. xyzzy
	auth_token = sha256pw( convert_to('Salt/Pepper/Rosmerry/'||p_job_id||p_user_id||common_seq, 'UTF8' ) );

	select t1."test_type", t1."url"
		into l_test_type, l_url
		from "t_job" as t1
		where t1."user_id" = p_user_id
		  and t1."id" = p_job_id
		;
	IF NOT FOUND THEN
		RAISE EXCEPTION 'invalid p_job_id=% or p_user_id=%', p_job_id, p_user_id;
	END IF;

	l_script_to_run = 'chk-link';
	l_host_id = 'x';
	l_client_id = 'x';
	l_useragent_id = 'x';

	select "value"
		into l_i_am
		from "t_config_data"
		where "name" = 'i-am'
	;
	IF NOT FOUND THEN
		l_i_am = 'prod';
	END IF;

	if ( l_test_type = 'Check Links/HTML/CSS' ) then	

		select
				  "hosted_at"
				, "hostname"
				, "host_type"
				, "ip_of_host"
				, "mon_url"
				, "osFamily"
				, "osMajor"
				, "osMinor"
				, "script_desc"
				, "url_of_host"
			into 
				  l_hosted_at 			--  'Garage-DC';
				, l_hostname 			--  'grape';
				, l_host_type 			--  'test-system:win7';
				, l_ip_of_host 			--  '172.16.0.167:5022';
				, l_mon_url 			--  'http://192.168.0.151:3050/who-cares.html';
				, l_osFamily 			--  'Windows';
				, l_osMajor 			--  '07';
				, l_osMinor 			--  '00';
				, l_script_desc 		--  'Check some silly thing on windows';
				, l_url_of_host 		--  'http://172.16.0.167:5022/it.html';
			from "t_dispatch_scripts" 
			where script_to_run = l_script_to_run
				and dev_or_prod = l_i_am
			;
		IF NOT FOUND THEN
			RAISE EXCEPTION 'invalid script name (E1415), script_to_run="%", i_am="%"', l_script_to_run, l_i_am;
		END IF;

		if ( l_host_id = 'x' ) then
			select "id"
				into l_host_id
				from "t_host"
				where "hostname" = l_hostname
				;
			IF NOT FOUND THEN
				l_host_id = uuid_generate_v4();
				-- -------------------------------------------------------- -- --------------------------------------------------------
				-- Data for systems - this will need to be more robust later when we need to distribute activity across boxes.
				-- -------------------------------------------------------- -- --------------------------------------------------------
				insert into "t_host" ( "id", "url_of_it", "hostname", "host_type", "is_running_now", "ip_addr", "hosted_at", "osFamily", "osMajor", "osMinor" )
					values ( l_host_id, l_url_of_host, l_hostname, l_host_type, 'y', l_ip_of_host, l_hosted_at, l_osFamily, l_osMajor, l_osMinor );
			END IF;
		end if;
		if ( l_useragent_id = 'x' ) then
			select "id"
				into l_useragent_id
				from "t_userAgent"
				where "name" = l_script_to_run
				;
			IF NOT FOUND THEN
				l_useragent_id = uuid_generate_v4();
				insert into "t_userAgent" ( "id", "name", "title", "osFamily", "osMajor", "osMinor" )
					values ( l_useragent_id, l_script_to_run, l_script_desc, l_osFamily, l_osMajor, l_osMinor );
			END IF;
		end if;
		if ( l_client_id = 'x' ) then
			select "id"
				into l_client_id
				from "t_client"
				where "name" = l_script_to_run
				;
			IF NOT FOUND THEN
				l_client_id = uuid_generate_v4();
				insert into "t_client" ( "id", "host_id", "name", "useragent_id", "useragent", "ip" )
					values ( l_client_id, l_host_id, l_script_to_run, l_useragent_id, l_script_to_run,  l_ip_of_host );
			END IF;
		end if;

		l_user_data = '{"script":"chk-link","auth_token":"'||auth_token||'","-a":"'||auth_token||'","-b":"'||l_url||'","-u":"'||l_url||'","note":"t002-note"}';

		insert into "t_a_run" (
			  "user_id"				
			, "c_run_id"		
			, "job_id" 		
			, "client_id" 		
			, "client_hash" 		
			, "host_id" 	
			, "useragent_id" 
			, "run_name" 
			, "url" 	
			, "user_data"
			, "auth_token"
			, "priority" 
		) select
			  p_user_id						-- user_id
			, common_seq					-- c_run_id
			, p_job_id						-- job_id
			, l_client_id					-- t4."id"				-- t_client.id == client_id	-- xyzzy can get this.
			, 'script-check-link'			-- client_hash
			, l_host_id						-- t4."host_id" 			-- xyzzy - can get this.
			, l_useragent_id 				-- useragent_id
			, t1."name"						-- run_name
			, t1."url"						-- url to check -- url
			, l_user_data					-- user_data - needs to be JSON, eg: {"script":"chk-link","auth_token":"inky dinky and bob","-a":"inky dinky and bob","-b":"http://www.excenent-answers.com/","-u":"http://www.excelent-answers.com/index.html","note":"t001-test"} 
			, auth_token
			, t1."priority"
		from "t_job" as t1
		where t1."user_id" = p_user_id
		  and t1."id" = p_job_id
		;

	else

		insert into "t_a_run" (
			  "user_id"				
			, "c_run_id"		
			, "job_id" 		
			, "client_id" 		
			, "client_hash" 		
			, "host_id" 	
			, "useragent_id" 
			, "run_name" 
			, "url" 	
			, "user_data"
			, "auth_token"
			, "priority" 
		) select
			  p_user_id
			, common_seq
			, p_job_id
			, 'available'					-- t4."id"				-- t_client.id == client_id
			, lower(t5."browserFamily" || '-' || t5."browserMajor" || '-' || t5."browserMinor" || '-' || t5."osFamily" || '-' || t5."osMajor" || '-' || t5."osMinor")
			, 'available'					-- t4."host_id" 
			, t3."useragent_id" 
			, t1."name"
			, t1."url"
			, t1."user_data"
			, auth_token
			, t1."priority"
		from "t_job" as t1
			, "t_runSet_members" t3
			, "t_userAgent" t5
		where t1."user_id" = p_user_id
		  and t1."id" = p_job_id
		  and t1."runset_id" = t3."runset_id"
		  and t3."useragent_id" = t5."id"
		;

	end if;

	RETURN common_seq;
END;
$$ LANGUAGE plpgsql;













-- -------------------------------------------------------- -- --------------------------------------------------------
create table "t_output" (
	  "seq"	 				bigint DEFAULT nextval('t_host_id_seq'::regclass) NOT NULL 
	, "msg"					varchar(255)
	, "created" 			timestamp default current_timestamp not null 						--
);
create index "t_output_p1" on "t_output" ( "created" );

-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE or REPLACE FUNCTION daily_cleanup( p_who_did_it varchar ) RETURNS varchar AS $$
BEGIN
	delete from "t_output" where "created" < current_timestamp - interval '1 day';
	delete from "t_link_valid" where "created" < current_timestamp - interval '30 day';
	delete from "t_monitor_results" where "created" < current_timestamp - interval '2 day';
	RETURN p_who_did_it;
END;
$$ LANGUAGE plpgsql;








-- -------------------------------------------------------- -- --------------------------------------------------------
-- To find when not enough browsers are running.
--		select min(created) from "t_a_run" where status = 'init' and created < current_time - interval '2 min';
-- Also calculate the average wait for a test system from this data.
--
-- For scripts/linux - a client-hash of "linux-script" or "linux-script-link-check"
-- For scripts/linux - a client-hash of "linux-jshint"
-- For scripts/linux - a client-hash of "linux-html-lint"
-- For scripts/linux - a client-hash of "linux-css-lint"
-- -------------------------------------------------------- -- --------------------------------------------------------
-- old --CREATE or REPLACE FUNCTION find_work(p_useragent_id varchar , p_client_id varchar , p_host_id varchar ) RETURNS varchar AS $$
-- old --DECLARE
-- old --    run_id char varying(40);
-- old --	browserFamily char varying(40);
-- old --	browserMajor char varying(40);
-- old --	browserMinor char varying(40);
-- old --	osFamily char varying(40);
-- old --	osMajor char varying(40);
-- old --	osMinor char varying(40);
-- old --	l_client_hash char varying(250);
-- old --	--	client_hash = ts0 ( "%{browserFamily%}-%{browserMajor%}-%{browserMinor%}-%{osFamily%}-%{osMajor%}-%{osMinor%}", ua_data );
-- old --BEGIN
-- old --
-- old --	select  "browserFamily", "browserMajor", "browserMinor", "osFamily", "osMajor", "osMinor"
-- old --	into  browserFamily, browserMajor, browserMinor, osFamily, osMajor, osMinor
-- old --		from "t_userAgent"
-- old --		where "id" = p_useragent_id
-- old --		;
-- old --	  -- uuid_generate_v4() not null primary key
-- old --
-- old --	IF NOT FOUND THEN
-- old --		RAISE EXCEPTION 'invalid useragent_id=%', p_useragent_id;
-- old --	END IF;
-- old --
-- old --	-- if ( lower(osFamily) = 'ubuntu' ) then
-- old --	 -- 	osFamily = 'linux';
-- old --	-- end if;
-- old --
-- old --	l_client_hash = lower(browserFamily || '-' || browserMajor || '-' || browserMinor || '-' || osFamily || '-' || osMajor || '-' || osMinor);
-- old --
-- old --	-- insert into "t_output" ( "msg" ) values ( 'client_hash='||l_client_hash );
-- old --
-- old --	select id
-- old --	into run_id
-- old --		from "t_a_run" 
-- old --		where "status" = 'init'
-- old --		 and "client_hash" = l_client_hash
-- old --		order by "priority" desc, "created" 
-- old --		for update 
-- old --		limit 1 
-- old --		;
-- old --
-- old --	IF NOT FOUND THEN
-- old --		-- RAISE EXCEPTION 'nothing to run for %', client_hash;
-- old --		run_id = '0';
-- old --	END IF;
-- old --
-- old --	if ( run_id <> '0' ) then
-- old --
-- old --		update "t_a_run" 
-- old --			set "status" = 'picked'
-- old --			  , "client_id" = p_client_id
-- old --			  , "host_id" = p_host_id
-- old --			  , "useragent_id" = p_useragent_id
-- old --			  , "run_picked" = current_timestamp
-- old --			where "id" = run_id
-- old --			;
-- old --
-- old --		/* xyzzy - we need to track # of runs on a per-user basis */
-- old --
-- old --		/* after 30 minutes idle a host can be stoped. */
-- old --		update "t_host"
-- old --			set "last_run_at" = current_timestamp
-- old --			,	"n_test_run" = "n_test_run" + 1
-- old --			where "id" = p_host_id
-- old --			;
-- old --
-- old --		/* after 15 minutes idle a client can be stoped. */
-- old --		update "t_client"
-- old --			set "last_run_at" = current_timestamp
-- old --			,	"n_test_run" = "n_test_run" + 1
-- old --			where "id" = p_client_id
-- old --			;
-- old --
-- old --	end if;
-- old --
-- old --	RETURN run_id;
-- old --END;
-- old --$$ LANGUAGE plpgsql;


--
--	// ===============================================================================================================================================================
--	// Pulls back all the info about the useragent - generates the client_hash;
--	// ===============================================================================================================================================================
--	function get_useragent_info(callback){
--		console.log('get_useragent_info');
--		dn.runQuery ( stmt = ts0( 
--					 'select /*v08*/ * '
--					+'from "t_useragent" '
--					+'where "id" = \'%{useragent_id%}\' '
--				, { "client_id":client_id, "client_hash":client_hash } )
--			, function ( err, result ) {		// xyzzy need to limit?
--				console.log ( "stmt="+stmt );
--				if ( err === null ) {
--					if ( result.rows.length ) {
--						ua_data = result.rows[0];
--						client_hash = ts0 ( "%{browserFamily%}-%{browserMajor%}-%{browserMinor%}-%{osFamily%}-%{osMajor%}-%{osMinor%}", ua_data );
--						found_useragent = true;
--					} else {
--						found_useragent = false;
--					}
--				} else {
--					g_error = true;
--					g_rv = { "status":"error", "msg":"error looking up the useragent." };
--				}
--				callback(null,'find-work');
--		});
--	}
--
--	// ===============================================================================================================================================================
--	// Xyzzy - this really neds to be a store procedure so that it is in a single transation.
--	function find_work(callback){
--		console.log('find_work');
--		dn.runQuery ( stmt = ts0( 
--					 'select /*v08*/ * '
--					+'from "t_a_run" '
--					+'where "status" = \'init\' '
--					+ ' and "client_id" = \'available\' '						// , ' and "client_id" = \'%{client_id%}\' '			// xyzzy - this is wrong.  Need to pick based on useragent and host
--					+ ' and "client_hash" = \'%{client_hash%}\' '					// Xyzzy - "for update" - do we need a transaction
--					+'order by "priority" desc, "created" '
--					// +'for update '
--					// +'limit 1 '
--				, { "client_id":client_id, "client_hash":client_hash } )
--			, function ( err, result ) {		// xyzzy need to limit?
--				console.log ( "stmt="+stmt );
--				if ( err === null ) {
--					if ( result.rows.length ) {
--						t_a_run = result.rows;
--						found_work = true;
--					} else {
--						found_work = false;
--					}
--				} else {
--					g_error = true;
--					g_rv = { "status":"error", "msg":"error looking for work." };
--				}
--				callback(null,'find-work');
--		});
--	}
--
--	// ===============================================================================================================================================================
--	function mark_as_picked(callback){
--		console.log('mark_as_picked');
--		if ( t_a_run.length > 0 ) {
--			var id = t_a_run[0].id;
--			dn.runQuery ( stmt = ts0( 'update "t_a_run" set "status" = \'picked\', "client_id" = \'%{client_id%}\', "host_id" = \'%{host_id%}\' where "id" = \'%{id%}\' '
--			, { "id":id, "client_id":client_id, "host_id":host_id } )
--				, function ( err, result ) {
--					console.log ( "stmt="+stmt );
--					if ( err === null ) {
--						
--					} else {
--						g_error = true;
--						g_rv = { "status":"error", "msg":"failed to update run to a status of started." };
--					}
--					callback(null,'mark-as-piced');
--				});
--		} else {
--			callback(null,'mark-as-piced(no work)');
--		}
--	}


delete from "t_output";

--    "status": "success",
--    "client_id": "e360540f-8069-4439-8c43-34113977ea00",
--    "useragent_id": "6b281f24-591d-4ce6-fdce-8ce44b7f897d",
--    "browserCSSClass": "swarm-browser-mobile_safari swarm-browser-mobile_safari-6 swarm-os swarm-os-ios",
--    "browserDisplayName": "mobile_safari 6.0 <br> ios ",
--    "browserDisplayTitle": "mobile_safari 6.0/ios "

-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE or REPLACE FUNCTION find_work(p_useragent_id varchar , p_client_id varchar , p_host_id varchar, p_client_name varchar ) RETURNS varchar AS $$
DECLARE
    run_id char varying(40);
	browserFamily char varying(40);
	browserMajor char varying(40);
	browserMinor char varying(40);
	osFamily char varying(40);
	osMajor char varying(40);
	osMinor char varying(40);
	l_client_hash char varying(250);
	found_error char varying(40);
	--	client_hash = ts0 ( "%{browserFamily%}-%{browserMajor%}-%{browserMinor%}-%{osFamily%}-%{osMajor%}-%{osMinor%}", ua_data );
BEGIN

	found_error = 'no';

	select  "browserFamily", "browserMajor", "browserMinor", "osFamily", "osMajor", "osMinor"
	into  browserFamily, browserMajor, browserMinor, osFamily, osMajor, osMinor
		from "t_userAgent"
		where "id" = p_useragent_id
		;

	-- uuid_generate_v4() not null primary key

	IF NOT FOUND THEN
		run_id = '0';
		found_error = 'yes';
		-- insert into "t_output" ( "msg" ) values ( 'Invalid useragent_id='||p_useragent_id );
	END IF;

	if ( found_error = 'no' ) then 

		update "t_host_can_run"
			set "is_running_now" = 'n'
			where ( "running_at" < current_timestamp - interval '5 minutes' 
			   or "running_at" is null
				)
			and "useragent_id" <> p_useragent_id
			;

		update "t_host_can_run"
			set "is_running_now" = 'y'
			  , "running_at" = current_timestamp
			where "useragent_id" = p_useragent_id
			;

		osFamily = lower(osFamily);
		osFamily = replace(osFamily,' ','_');
		browserFamily = lower(browserFamily);
		browserFamily = replace(browserFamily,' ','_');

		l_client_hash = browserFamily || '-' || browserMajor || '-' || browserMinor || '-' || osFamily || '-' || osMajor || '-' || osMinor;

		-- insert into "t_output" ( "msg" ) values ( 'l_client_hash='||l_client_hash||' client_id='||p_client_id );

		select id
		into run_id
			from "t_a_run" 
			where "status" = 'init'
			 and "client_hash" = l_client_hash
			order by "priority" desc, "created" 
			for update 
			limit 1 
			;

		IF NOT FOUND THEN
			-- RAISE EXCEPTION 'nothing to run for %', client_hash;
			run_id = '0';
		END IF;

		if ( run_id <> '0' ) then

			update "t_a_run"
				set "status" = 'picked'
					, "client_id" = p_client_id
					, "host_id" = p_host_id
				where "id" = run_id;

		end if;

	end if;

	RETURN run_id;
END;
$$ LANGUAGE plpgsql;

















m4_define([[[m4_updTrig]]],[[[

CREATE OR REPLACE function $1_upd()
RETURNS trigger AS 
$BODY$
BEGIN
  NEW.updated := current_timestamp;
  RETURN NEW;
END
$BODY$
LANGUAGE 'plpgsql';


CREATE TRIGGER $1_trig
BEFORE update ON "$1"
FOR EACH ROW
EXECUTE PROCEDURE $1_upd();

]]])


-- triggers for t_user in auth.sql - should be run after this

m4_updTrig(t_group)
m4_updTrig(t_activity)
m4_updTrig(t_project)
m4_updTrig(t_host)
m4_updTrig(t_host_can_run)
m4_updTrig(t_host_can_vm)
m4_updTrig(t_client)
m4_updTrig(t_a_run)
m4_updTrig(t_monitor_stuff)
m4_updTrig(t_status_stuff)
m4_updTrig(t_ssh_port_pool)
m4_updTrig(t_ssh_login)
m4_updTrig(t_ssh_global)
m4_updTrig(t_it_work)
m4_updTrig(t_knobs)
m4_updTrig(t_ua_log)
m4_updTrig(t_link_valid)







--
--
--CREATE OR REPLACE function t_run_result_upd()
--RETURNS trigger AS 
--$BODY$
--BEGIN
--  NEW.updated := current_timestamp;
--  RETURN NEW;
--END
--$BODY$
--LANGUAGE 'plpgsql';
--
--
--CREATE TRIGGER t_run_results_trig
--BEFORE update ON "t_run_result"
--FOR EACH ROW
--EXECUTE PROCEDURE t_run_result_upd();
--
--










-- -------------------------------------------------------- -- --------------------------------------------------------
--CREATE TABLE "t_host" (
--	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
--	, "host_id"				bigint DEFAULT nextval('t_host_id_seq'::regclass) NOT NULL 
--	, "url_of_it"			char varying (255) not null 
--	, "hostname"			char varying (255) not null 
--  	, "host_type"			char varying (40) default 'VM' not null 	-- 'host', 'VM', '???'
--	, "control_method"		char varying (40) default 'it'
--	, "comm_method" 		char varying(10) not null default 'post' 	-- How communication occures with client, 'post', 'jsonp', 'socket.io', 'server'
--	, "is_running_now"		char (1) default 'n' not null 
--	, "can_rdc"				char (1) default 'n' not null 
--	, "rdc_port"			char (6) 
--	, "rdc_in_use"			char (1) default 'n' not null 
--	, "rdc_user_id"			char (40)						-- User ID of person that is currently using this VMs Desktop 
--	, "ip_addr"				char (40)						-- Location of server if not a local system
--	, "last_run_at" 		timestamp  						-- Time of last "ping" or data received back from "it"
--	, "hosted_at"			char (100) default 'inhouse'	-- Locaiton of hosing (www.macincloud.com, www.linode.com, digitalocean.com, aws.amazon.com etc) 
--	, "connection_method"	char (10) default 'rdc'			-- rdc, vnc etc.
--	, "updated" 			timestamp  						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
--	, "created" 			timestamp default current_timestamp not null 		
--);

-- -------------------------------------------------------- -- --------------------------------------------------------
CREATE or REPLACE FUNCTION i_am_alive(p_hostname varchar, p_user_id varchar, p_ip_addr varchar, p_host_type varchar ) RETURNS varchar AS $$
DECLARE
    l_host_id char varying(40);
	insert_flag boolean;

    l_id char varying(40);
	monetering_flag boolean;
BEGIN

	-- ----------------------------------------------------- t_monitor_stuff -------------------------------------------------------

	monetering_flag = false;

	select  "id"
	into  l_id
		from "t_monitor_stuff"
		where "item_name" = 'It-On:'||p_hostname
		;

	IF NOT FOUND THEN
		monetering_flag = true;
	END IF;

	if ( monetering_flag ) then

		insert into "t_monitor_stuff" ( "item_name", "event_to_raise", "delta_t", "timeout_event" )
			values ( 'It-On:'||p_hostname, 'System '||p_hostname||' alive and running *it*', '4 minute', current_timestamp );

	else

		update "t_monitor_stuff"
			set
			  "timeout_event" = current_timestamp + CAST("delta_t" as Interval) 
			where "item_name" = 'It-On:'||p_hostname
		;

	end if;

	-- ----------------------------------------------------- t_host -------------------------------------------------------

	insert_flag = false;
	
	select  "id"
	into  l_host_id
		from "t_host"
		where "hostname" = p_hostname
		;

	IF NOT FOUND THEN
		l_host_id = uuid_generate_v4();
		insert_flag = true;
	END IF;

	if ( insert_flag ) then

		insert into "t_host" (
			  "id"		
			, "url_of_it"		
			, "hostname"	
			, "host_type"	
			, "comm_method" 
			, "is_running_now"	
			, "ip_addr"	
			, "last_run_at" 
		) values (
			  l_host_id
			, 'n/a'					-- It's a Post-It controled system.
			, p_hostname
			, p_host_type
			, 'post'
			, 'y'
			, p_ip_addr
			, current_timestamp
		);

	else

		update "t_host"
			set
			  "is_running_now" = 'y'
			, "ip_addr"	= p_ip_addr
			, "last_run_at" = current_timestamp
		where "id" = l_host_id
		;

	end if;

	RETURN l_host_id;
END;
$$ LANGUAGE plpgsql;




-- alter TABLE "t_host_can_run" add "running_at" timestamp;

--CREATE TABLE "t_host_can_run" (
--	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
--	, "host_id"				char varying (40) not null 
--	, "useragent_id"		char varying (40) not null 
--	, "vendor_version_no"	char varying (40) 
--	, "config_data" 		char varying(255)  
--	, "client_name" 		char varying(255)  
--	, "client_id"			char varying (40) 
--	, "is_running_now"		char (1) default 'n' not null 
--	, "updated" 			timestamp  						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
--	, "created" 			timestamp default current_timestamp not null 		
--);

--drop table "t_host_can_vm" ;
--CREATE TABLE "t_host_can_vm" (
--	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
--	, "host_id"				char varying (40) not null 
--	, "client_name" 		char varying(255)  
--	, "osFamily"			char varying (40)
--	, "osMajor"				char varying (40)
--	, "osMinor"				char varying (40)
--	, "is_running_now"		char (1) default 'n' not null 
--	, "updated" 			timestamp  						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
--	, "created" 			timestamp default current_timestamp not null 		
--);





-- dn.runQuery ( stmt = ts0( 'select /*vTestICanRun.sql*/ i_can_run ( \'%{host_id%}\', \'%{hostname%}\', \'%{user_id%}\', \'%{xyzzy%}\', \'%{xyzzy%}\' ) as "id" '
CREATE or REPLACE FUNCTION i_can_run(p_host_id varchar, p_hostname varchar, p_user_id varchar, p_ip_addr varchar, p_host_type varchar, p_browserFamily varchar, p_browserMajor varchar, p_browserMinor varchar, p_osFamily varchar, p_osMajor varchar, p_osMinor varchar, p_client varchar, p_clientOsFamily varchar, p_clientOsMajor varchar, p_clientOsMinor varchar ) RETURNS varchar AS $$
DECLARE
    l_host_id char varying(40);
	insert_flag boolean;

    can_run_id char varying(40);
    can_run_flag boolean;

	l_useragent_id varchar(40);
    useragent_flag boolean;
	
	host_can_vm_id varchar(40);
    client_can_vm_flag boolean;

	junk bigint;
BEGIN

	insert_flag = false;
	
	select  "id"
	into  l_host_id
		from "t_host"
		where "hostname" = p_hostname
		;

	IF NOT FOUND THEN
		l_host_id = uuid_generate_v4();
		insert_flag = true;
	END IF;

	if ( insert_flag ) then

		insert into "t_host" (
			  "id"		
			, "url_of_it"		
			, "hostname"	
			, "host_type"	
			, "comm_method" 
			, "is_running_now"	
			, "ip_addr"	
			, "last_run_at" 
		) values (
			  l_host_id
			, 'n/a'					-- It's a Post-It controled system.
			, p_hostname
			, p_host_type
			, 'post'
			, 'y'
			, p_ip_addr
			, current_timestamp
		);

	else

		update "t_host"
			set
			  "is_running_now" = 'y'
			, "ip_addr"	= p_ip_addr
			, "last_run_at" = current_timestamp
		where "id" = l_host_id
		;

	end if;

	if ( ( p_host_type = 'VM' ) or ( p_host_type = 'host' ) ) then

		--CREATE TABLE "t_userAgent" (
		--	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
		--	, "name"				char varying (40) not null
		--	, "title"				char varying (80) not null
		--	, "browserFamily"		char varying (40)
		--	, "browserMajor"		char varying (40)
		--	, "browserMinor"		char varying (40)
		--	, "browserOptions"		char varying (40)
		--	, "osFamily"			char varying (40)
		--	, "osMajor"				char varying (40)
		--	, "osMinor"				char varying (40)
		--	, "osOptions"			char varying (40)
		--	, "created" 			timestamp default current_timestamp not null 		
		--	, "n_test_run"			bigint default 0										
		--);

		select  1, "id"
		into  junk, l_useragent_id
			from "t_userAgent"
			where
				"browserFamily" = p_browserFamily
			and "browserMajor" = p_browserMajor
			and "browserMinor" = p_browserMinor
			and	"osFamily" = p_osFamily
			and "osMajor" = p_osMajor
			and "osMinor" = p_osMinor
		union 
			select  2, "id"
				from "t_userAgent"
				where
					"browserFamily" = p_browserFamily
				and "browserMajor" = p_browserMajor
				and "browserMinor" = p_browserMinor
				and	"osFamily" = p_osFamily
				and "osMajor" = p_osMajor
		order by 1 asc
		limit 1
		;

		IF NOT FOUND THEN
			l_useragent_id = uuid_generate_v4();
			useragent_flag = true;
		END IF;

		if ( useragent_flag ) then

			-- values ( '200', 'Chrome 28/Linux', 'Chrome 28', 'chrome', '28', '0', 'linux', '3', '35' );
			insert into "t_userAgent" (
				  "id"			
				, "name"			
				, "title"		
				, "browserFamily"	
				, "browserMajor"
				, "browserMinor"
				, "osFamily"
				, "osMajor"
				, "osMinor"	
			) values (
				  l_useragent_id
				, p_browserFamily||' '||p_browserMajor||'/'||p_osFamily									-- xyzzy - should pass these in and use JS to upper 1st char for appearance
				, p_browserFamily||' '||p_browserMajor
				, p_browserFamily		
				, p_browserMajor	
				, p_browserMinor
				, p_osFamily	
				, p_osMajor
				, p_osMinor
			);

		end if;




		--CREATE TABLE "t_host_can_run" (
		--	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
		--	, "host_id"				char varying (40) not null 
		--	, "useragent_id"		char varying (40) not null 
		--	, "vendor_version_no"	char varying (40) 
		--	, "config_data" 		char varying(255)  
		--	, "client_name" 		char varying(255)  
		--	, "client_id"			char varying (40) 
		--	, "is_running_now"		char (1) default 'n' not null 
		--	, "updated" 			timestamp  						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
		--	, "created" 			timestamp default current_timestamp not null 		
		--);

		can_run_flag = false;
		
		select  "id"
		into  can_run_id
			from "t_host_can_run"
			where "host_id" = l_host_id
			  and "useragent_id" = l_useragent_id
			;

		IF NOT FOUND THEN
			can_run_id = uuid_generate_v4();
			can_run_flag = true;
		END IF;

		if ( can_run_flag ) then

			insert into "t_host_can_run" (
				  "id"		
				, "host_id"	
				, "useragent_id"
				, "client_name"
			) values (
				  can_run_id
				, l_host_id
				, l_useragent_id
				, p_browserFamily||' '||p_browserMajor||'/'||p_osFamily									-- xyzzy - should pass these in and use JS to upper 1st char for appearance
			);

		else

			update "t_host_can_run"
				set
				  "is_running_now" = 'n'
			where "host_id" = l_host_id
			  and "useragent_id" = l_useragent_id
			;

		end if;

	end if;

	if ( p_host_type = 'host' ) then

		--	CREATE TABLE "t_host_can_vm" (
		--		  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
		--		, "host_id"				char varying (40) not null 
		--		, "client_name" 		char varying(255)  
		--		, "is_running_now"		char (1) default 'n' not null 
		--		, "updated" 			timestamp  						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
		--		, "created" 			timestamp default current_timestamp not null 		
		--	);

		client_can_vm_flag = false;

		select  "id"
		into  host_can_vm_id
			from "t_host_can_vm"
			where "client_name" = p_client
			;

		IF NOT FOUND THEN
			host_can_vm_id = uuid_generate_v4();
			client_can_vm_flag = true;
		END IF;

		if ( client_can_vm_flag ) then

			insert into "t_host_can_vm" (
				  "id"			
				, "host_id"			
				, "client_name" 
				, "osFamily"
				, "osMajor"
				, "osMinor"	
				, "is_running_now"		
			) values ( 
				   host_can_vm_id
				,  l_host_id
				,  p_client
				,  p_clientOsFamily
				,  p_clientOsMajor
				,  p_clientOsMinor	
				, 'n'
			);

		end if;

	end if;

	RETURN l_host_id;
END;
$$ LANGUAGE plpgsql;



drop TABLE "t_customer" ;
CREATE TABLE "t_customer" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "name"				char varying (40) not null
	, "ctype"				char varying (10) check ( "ctype" in ( 'regular', 'admin', 'test' ) ) default 'test' not null
	, "config"				text default '{}' not null
);

-- "config" to replace...
-- CREATE TABLE "tblObservationType"(		: name, description
-- CREATE TABLE "tblMine"(			-> Site : name, description, division, abbreviation 		
-- CREATE TABLE "tblDepartment"(			: name, description
-- CREATE TABLE "tblCrew"(					: name, description
-- CREATE TABLE "tblCategory"(				: code, name, description - as hash
-- CREATE TABLE "tblConfig"(				: name { values }
-- CREATE TABLE "newImportance"(			: name { values: numeric-import, description }

drop TABLE "t_host_to_customer" ;
CREATE TABLE "t_host_to_customer" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "host"				char varying (240) not null
	, "customer_id"			char varying (40) not null
);

insert into "t_host_to_customer" ( "host", "customer_id" ) values
	  ( 'localhost:8090', '1' )
	, ( '127.0.0.1:8090', '1' )
	, ( '192.168.0.151:8090', '1' )
	, ( 'mine1.2c-why.com:80', '2' )
	, ( 'mine1.2c-why.com', '2' )
;

-- insert into "t_customer" ( "id", "name", "ctype", "config" ) values
-- 	  ( '1', 'dev-test-1', 'test', '{}' )
-- ;
insert into "t_customer" ( "id", "name", "ctype", "config" ) values ( '1', 'dev-test-1', 'test', '{"Category":[{"id":1,"value":"Saftey Solgan","code":"SL","flag":"s"},{"id":3,"value":"Process Improvement","code":"PI","flag":"."},{"id":5,"value":"Information/Note","code":"IN","flag":"i"},{"id":8,"value":"At Risk","code":"AR","flag":"."},{"id":9,"value":"Incedent/Injury","code":"II","flag":"."}],"Application Name":"Saftey Observation","Report Base URL":"http://localhost:8099/","Crew":[{"id":1,"value":"A","description":"5:00AM to 2:00PM"},{"id":2,"value":"B","description":"2:00PM to 10:00PM"},{"id":3,"value":"C","description":"10:00PM to 5:00AM"},{"id":4,"value":"M","description":""},{"id":5,"value":"J","description":""}],"Department":[{"id":2,"value":"Office","flag":""},{"id":3,"value":"Operations","flag":""},{"id":4,"value":"Prep Plant","flag":""},{"id":5,"value":"Purchasing","flag":""},{"id":6,"value":"Training","flag":""},{"id":7,"value":"IT Services","flag":""},{"id":8,"value":"All","flag":"*"},{"id":1,"value":"Maintance","flag":""}],"Base URL":"http://localhost:8099/","Site":[{"id":2,"value":"North Mine","description":"","business_unit":"Operations","abreviation":"AD"},{"id":3,"value":"South Mine","description":"","business_unit":"Operations","abreviation":"OP"},{"id":1,"value":"Admin Building","description":"","business_unit":"Admin","abreviation":"OP"}],"Severity":[{"id":1,"value":"Severe","color":"yellow"},{"id":2,"value":"Important","color":"green"},{"id":3,"value":"Note","color":"green"}],"Reviewable":[{"id":0,"value":"No"},{"id":1,"value":"Yes"}],"Complete":[{"id":0,"value":"No"},{"id":1,"value":"Yes"}],"ObservationType":[{"id":0,"value":"0","description":"0"},{"id":1,"value":"1","description":"1"}]}' );


