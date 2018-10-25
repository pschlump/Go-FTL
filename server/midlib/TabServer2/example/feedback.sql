
-- -------------------------------------------------------- -- --------------------------------------------------------
-- feedback table - example.
-- -------------------------------------------------------- -- --------------------------------------------------------
-- drop TABLE "feedback" ;
CREATE TABLE "feedback" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "subject"				text
	, "body"				text
	, "status"				char varying (40) 
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);


CREATE OR REPLACE function feedback_upd()
RETURNS trigger AS 
$BODY$
BEGIN
	NEW.updated := current_timestamp;
	RETURN NEW;
END
$BODY$
LANGUAGE 'plpgsql';


CREATE TRIGGER feedback_trig
BEFORE update ON "feedback"
FOR EACH ROW
EXECUTE PROCEDURE feedback_upd();



-- create a row of sample data for this table
insert into "feedback" ( "subject", "body", "status" ) values
	( 'Feedback test data 1', 'Just some feedback - it is awsome!', 'test-data' ),
	( 'Feedback test data 2', 'Just some feedback - still awsome!', 'test-data' )
;

