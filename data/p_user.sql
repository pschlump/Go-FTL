






-- -------------------------------------------------------- -- --------------------------------------------------------
drop TABLE "qr_user";
CREATE TABLE "qr_user" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "customer_id"			char varying (40) default '1'
	, "UserName"			char varying (80) 
	, "RealName"			char varying (255) 
	, "email"				char varying (255) 
	, "acct_state"			char varying (10) default 'unknown' check ( "acct_state" in ( 'unknown', 'locked', 'ok', 'pass-reset', 'billing', 'closed', 'fixed', 'temporary' ) )
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);

create index "qr_user_p1" on "qr_user" ( "email" );
create unique index "qr_user_u4" on "qr_user" ( "UserName" );



CREATE OR REPLACE function qr_user_upd()
RETURNS trigger AS 
$BODY$
BEGIN
  NEW.updated := current_timestamp;
  RETURN NEW;
END
$BODY$
LANGUAGE 'plpgsql';


CREATE TRIGGER qr_user_trig
BEFORE update ON "qr_user"
FOR EACH ROW
EXECUTE PROCEDURE qr_user_upd();



