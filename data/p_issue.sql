

-- Tracking of where and when a person uses a single-page app







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



CREATE OR REPLACE function p_issue_upd()
RETURNS trigger AS 
$BODY$
BEGIN
  NEW.updated := current_timestamp;
  RETURN NEW;
END
$BODY$
LANGUAGE 'plpgsql';


CREATE TRIGGER p_issue_trig
BEFORE update ON "p_issue"
FOR EACH ROW
EXECUTE PROCEDURE p_issue_upd();



