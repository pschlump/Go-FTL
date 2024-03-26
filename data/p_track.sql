
-- Tracking of where and when a person uses a single-page app







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




CREATE OR REPLACE function p_track_upd()
RETURNS trigger AS 
$BODY$
BEGIN
  NEW.updated := current_timestamp;
  RETURN NEW;
END
$BODY$
LANGUAGE 'plpgsql';


CREATE TRIGGER p_track_trig
BEFORE update ON "p_track"
FOR EACH ROW
EXECUTE PROCEDURE p_track_upd();



