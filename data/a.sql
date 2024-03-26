-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
--
-- for "__after_sync__" synthetic column in queries.
--
-- Add into TabServer2 - /api/list/syncPlul?table=a,b,c,d,e...
--
CREATE TABLE "t_sync_marker" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "note1"				char varying(50) not null
	, "created" 			timestamp default current_timestamp not null 						
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
);

insert into "t_sync_marker" ( "note1" ) values ( 'abc' );



CREATE OR REPLACE function t_sync_marker_upd()
RETURNS trigger AS 
$BODY$
BEGIN
  NEW.updated := current_timestamp;
  RETURN NEW;
END
$BODY$
LANGUAGE 'plpgsql';


CREATE TRIGGER t_sync_marker_trig
BEFORE update ON "t_sync_marker"
FOR EACH ROW
EXECUTE PROCEDURE t_sync_marker_upd();



--
--	,"/api/sync_marker": { "g": "sync_marker_update", "p": [ "$ip$"]
--		, "LoginRequired":false
--		, "LineNo":"Line: __LINE__ File: __FILE__"
--		, "Method":["GET","POST"]
--		, "TableList":[ "t_sync_marker" ]
--		, "valid": {
--			 "$ip$": 		{ "required":true, "type":"string", "max_len":40, "min_len":4 }
--			}
--		}
--

CREATE or REPLACE FUNCTION sync_marker_update (p_ip_addr varchar)
	RETURNS varchar AS $$
DECLARE
	l_rv			varchar (40);
BEGIN
	update "t_sync_marker"
		set "note1" = p_ip_addr
		;
	RETURN '{"status":"success"}';
END;
$$ LANGUAGE plpgsql;

