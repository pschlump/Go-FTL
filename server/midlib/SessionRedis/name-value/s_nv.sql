





-- "Copyright (C) Philip Schlump, 2009-2017." 

drop table s_nv;
create table s_nv (
	  "id"					uuid DEFAULT uuid_generate_v4() not null primary key	
	, "name"				text not null
	, "value"				text
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
	, "updated" 			timestamp 
);


-----------------------------------------------------------------------------------------------------------------------------------

drop FUNCTION s_save_nv(p_name varchar, p_value varchar);

CREATE or REPLACE FUNCTION s_save_nv(p_name varchar, p_value varchar)
	RETURNS varchar AS $$
begin
	-- "Copyright (C) Philip Schlump, 2017." 

	insert into "s_nv" ( "name", "value" ) values ( p_name, p_value );

	RETURN '{"status":"success"}';
END;
$$ LANGUAGE plpgsql;


-- select s_save_nv('a','b');


-----------------------------------------------------------------------------------------------------------------------------------

drop FUNCTION s_upd_nv(p_name varchar, p_value varchar);

CREATE or REPLACE FUNCTION s_upd_nv(p_name varchar, p_value varchar)
	RETURNS varchar AS $$
DECLARE
    l_count 				int;
begin
	-- "Copyright (C) Philip Schlump, 2017." 

	WITH rows AS (
		update "s_nv"
			set "value" = p_value
			where "name" = p_name
			returning 1
		)
		SELECT count(*)
		INTO l_count
		FROM rows
	;

	if l_count is null or l_count < 1 then
		insert into "s_nv" ( "name", "value" ) values ( p_name, p_value );
	end if;

	RETURN '{"status":"success"}';
END;
$$ LANGUAGE plpgsql;

-----------------------------------------------------------------------------------------------------------------------------------

drop FUNCTION s_del_nv(p_name varchar);

CREATE or REPLACE FUNCTION s_del_nv(p_name varchar)
	RETURNS varchar AS $$
begin
	-- "Copyright (C) Philip Schlump, 2017." 

	delete from "s_nv" where "name" = p_name;

	RETURN '{"status":"success"}';
END;
$$ LANGUAGE plpgsql;


-----------------------------------------------------------------------------------------------------------------------------------

drop FUNCTION s_get_nv(p_name varchar);

CREATE or REPLACE FUNCTION s_get_nv(p_name varchar)
	RETURNS varchar AS $$
DECLARE
    l_data 				text;
begin
	-- "Copyright (C) Philip Schlump, 2017." 

	select "value"
		into l_data
		from "s_nv"
		where "name" = p_name
	;

	if not found then
		RETURN '{"status":"error", "msg":"not found"}';
	else 
		RETURN '{"status":"success", "data":'||to_json(l_data)||'}';
	end if;
END;
$$ LANGUAGE plpgsql;







-----------------------------------------------------------------------------------------------------------------------------------




CREATE OR REPLACE function s_nv_upd()
RETURNS trigger AS 
$BODY$
BEGIN
  NEW.updated := current_timestamp;
  RETURN NEW;
END
$BODY$
LANGUAGE 'plpgsql';


CREATE TRIGGER s_nv_trig
BEFORE update ON "s_nv"
FOR EACH ROW
EXECUTE PROCEDURE s_nv_upd();


