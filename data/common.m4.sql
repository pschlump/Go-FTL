
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

