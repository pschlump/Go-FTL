
-- stmt := "insert into \"v1_trackAdd\" ( \"tag\" ) values ( $1 )"


CREATE SEQUENCE v1_tracAdd_seq
INCREMENT 1
MINVALUE 1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- xyzzy110 - add note/hash/prev_hash
drop TABLE "v1_trackAdd" ;
CREATE TABLE "v1_trackAdd" (
  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
, "tag"					char varying (40) not null
, "state"				char varying (10) check ( "state" in ( 'new', 'hashed', 'error', 's1', 's2' ) ) default 'new' not null
, "premis_id"			text
, "premis_animal"		text
, "note"				text
, "hash"				char varying (100)
, "prev_hash"			char varying (100)
, "ord_seq"				bigint DEFAULT nextval('v1_tracAdd_seq'::regclass) NOT NULL 
, "qr_id"				char varying (40)		-- id of QR assigned to this.
, "updated" 			timestamp 									 						
, "created" 			timestamp default current_timestamp not null 					
);

-- create unique index "v1_trackAdd_u1" on "v1_trackAdd" ( "premis_id", "premis_animal" );




CREATE SEQUENCE v1_qr_avail_seq
INCREMENT 1
MINVALUE 1
MAXVALUE 9223372036854775807
START 1
CACHE 1;


drop TABLE "v1_avail_qr" ;
CREATE TABLE "v1_avail_qr" (
  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
, "file_name"			text
, "url_path"			text
, "qr_encoded_url_path"			text
, "qr_id"				text
, "qr_enc_id"			text
, "state"				char varying (10) check ( "state" in ( 'avail', 'used', 's1', 's2' ) ) default 'avail' not null
, "ord_seq"				bigint DEFAULT nextval('v1_qr_avail_seq'::regclass) NOT NULL 
, "updated" 			timestamp 									 						
, "created" 			timestamp default current_timestamp not null 					
);

create index "v1_avail_qr_p1" on "v1_avail_qr" ( "qr_id" );
create index "v1_avail_qr_p2" on "v1_avail_qr" ( "qr_enc_id" );

-- select t1.*,
-- 	t2.*
-- from "v1_trackAdd" as t1 left join "v1_avail_qr" on t1."qr_id" = t2."id"

-- insert into "v1_avail_qr" ( "file_name", "url_path", "qr_id", "qr_enc_id") values
-- 	( 'qr001.png', 'http://t5432z/q/1.png', '1', '1' )
-- ,	( 'qr002.png', 'http://t5432z/q/2.png', '2', '2' )
-- ,	( 'qr003.png', 'http://t5432z/q/3.png', '3', '3' )
-- ,	( 'qr004.png', 'http://t5432z/q/3.png', '4', '4' )
-- ;


insert into "v1_avail_qr" ( "qr_id", "qr_enc_id", "url_path", "file_name", "qr_encoded_url_path" ) values
	  ( '170', '4q', 'http://127.0.0.1:9019/qr/00170.4.png', './td_0008/q00170.4.png', 'http://t432z.com/q/4q' )
	, ( '171', '4r', 'http://127.0.0.1:9019/qr/00171.4.png', './td_0008/q00171.4.png', 'http://t432z.com/q/4r' )
	, ( '172', '4s', 'http://127.0.0.1:9019/qr/00172.4.png', './td_0008/q00172.4.png', 'http://t432z.com/q/4s' )
	, ( '173', '4t', 'http://127.0.0.1:9019/qr/00173.4.png', './td_0008/q00173.4.png', 'http://t432z.com/q/4t' )
	, ( '174', '4u', 'http://127.0.0.1:9019/qr/00174.4.png', './td_0008/q00174.4.png', 'http://t432z.com/q/4u' )
	, ( '175', '4v', 'http://127.0.0.1:9019/qr/00175.4.png', './td_0008/q00175.4.png', 'http://t432z.com/q/4v' )
	, ( '176', '4w', 'http://127.0.0.1:9019/qr/00176.4.png', './td_0008/q00176.4.png', 'http://t432z.com/q/4w' )
	, ( '177', '4x', 'http://127.0.0.1:9019/qr/00177.4.png', './td_0008/q00177.4.png', 'http://t432z.com/q/4x' )
	, ( '178', '4y', 'http://127.0.0.1:9019/qr/00178.4.png', './td_0008/q00178.4.png', 'http://t432z.com/q/4y' )
	, ( '179', '4z', 'http://127.0.0.1:9019/qr/00179.4.png', './td_0008/q00179.4.png', 'http://t432z.com/q/4z' )
	, ( '180', '50', 'http://127.0.0.1:9019/qr/00180.4.png', './td_0008/q00180.4.png', 'http://t432z.com/q/50' )
	, ( '181', '51', 'http://127.0.0.1:9019/qr/00181.4.png', './td_0008/q00181.4.png', 'http://t432z.com/q/51' )
	, ( '182', '52', 'http://127.0.0.1:9019/qr/00182.4.png', './td_0008/q00182.4.png', 'http://t432z.com/q/52' )
	, ( '183', '53', 'http://127.0.0.1:9019/qr/00183.4.png', './td_0008/q00183.4.png', 'http://t432z.com/q/53' )
	, ( '184', '54', 'http://127.0.0.1:9019/qr/00184.4.png', './td_0008/q00184.4.png', 'http://t432z.com/q/54' )
	, ( '185', '55', 'http://127.0.0.1:9019/qr/00185.4.png', './td_0008/q00185.4.png', 'http://t432z.com/q/55' )
	, ( '186', '56', 'http://127.0.0.1:9019/qr/00186.4.png', './td_0008/q00186.4.png', 'http://t432z.com/q/56' )
	, ( '187', '57', 'http://127.0.0.1:9019/qr/00187.4.png', './td_0008/q00187.4.png', 'http://t432z.com/q/57' )
	, ( '188', '58', 'http://127.0.0.1:9019/qr/00188.4.png', './td_0008/q00188.4.png', 'http://t432z.com/q/58' )
	, ( '189', '59', 'http://127.0.0.1:9019/qr/00189.4.png', './td_0008/q00189.4.png', 'http://t432z.com/q/59' )
	, ( '190', '5a', 'http://127.0.0.1:9019/qr/00190.4.png', './td_0008/q00190.4.png', 'http://t432z.com/q/5a' )
	, ( '191', '5b', 'http://127.0.0.1:9019/qr/00191.4.png', './td_0008/q00191.4.png', 'http://t432z.com/q/5b' )
	, ( '192', '5c', 'http://127.0.0.1:9019/qr/00192.4.png', './td_0008/q00192.4.png', 'http://t432z.com/q/5c' )
	, ( '193', '5d', 'http://127.0.0.1:9019/qr/00193.4.png', './td_0008/q00193.4.png', 'http://t432z.com/q/5d' )
	, ( '194', '5e', 'http://127.0.0.1:9019/qr/00194.4.png', './td_0008/q00194.4.png', 'http://t432z.com/q/5e' )
	, ( '195', '5f', 'http://127.0.0.1:9019/qr/00195.4.png', './td_0008/q00195.4.png', 'http://t432z.com/q/5f' )
	, ( '196', '5g', 'http://127.0.0.1:9019/qr/00196.4.png', './td_0008/q00196.4.png', 'http://t432z.com/q/5g' )
	, ( '197', '5h', 'http://127.0.0.1:9019/qr/00197.4.png', './td_0008/q00197.4.png', 'http://t432z.com/q/5h' )
	, ( '198', '5i', 'http://127.0.0.1:9019/qr/00198.4.png', './td_0008/q00198.4.png', 'http://t432z.com/q/5i' )
	, ( '199', '5j', 'http://127.0.0.1:9019/qr/00199.4.png', './td_0008/q00199.4.png', 'http://t432z.com/q/5j' )
	, ( '200', '5k', 'http://127.0.0.1:9019/qr/00200.4.png', './td_0008/q00200.4.png', 'http://t432z.com/q/5k' )
	, ( '201', '5l', 'http://127.0.0.1:9019/qr/00201.4.png', './td_0008/q00201.4.png', 'http://t432z.com/q/5l' )
	, ( '202', '5m', 'http://127.0.0.1:9019/qr/00202.4.png', './td_0008/q00202.4.png', 'http://t432z.com/q/5m' )
	, ( '203', '5n', 'http://127.0.0.1:9019/qr/00203.4.png', './td_0008/q00203.4.png', 'http://t432z.com/q/5n' )
	, ( '204', '5o', 'http://127.0.0.1:9019/qr/00204.4.png', './td_0008/q00204.4.png', 'http://t432z.com/q/5o' )
;


-- 170 4q http://127.0.0.1:9019/qr/00170.4.png ./td_0008/q00170.4.png http://t432z.com/q/4q
-- 171 4r http://127.0.0.1:9019/qr/00171.4.png ./td_0008/q00171.4.png http://t432z.com/q/4r
-- 172 4s http://127.0.0.1:9019/qr/00172.4.png ./td_0008/q00172.4.png http://t432z.com/q/4s
-- 173 4t http://127.0.0.1:9019/qr/00173.4.png ./td_0008/q00173.4.png http://t432z.com/q/4t
-- 174 4u http://127.0.0.1:9019/qr/00174.4.png ./td_0008/q00174.4.png http://t432z.com/q/4u
-- 175 4v http://127.0.0.1:9019/qr/00175.4.png ./td_0008/q00175.4.png http://t432z.com/q/4v
-- 176 4w http://127.0.0.1:9019/qr/00176.4.png ./td_0008/q00176.4.png http://t432z.com/q/4w
-- 177 4x http://127.0.0.1:9019/qr/00177.4.png ./td_0008/q00177.4.png http://t432z.com/q/4x
-- 178 4y http://127.0.0.1:9019/qr/00178.4.png ./td_0008/q00178.4.png http://t432z.com/q/4y
-- 179 4z http://127.0.0.1:9019/qr/00179.4.png ./td_0008/q00179.4.png http://t432z.com/q/4z
-- 180 50 http://127.0.0.1:9019/qr/00180.4.png ./td_0008/q00180.4.png http://t432z.com/q/50
-- 181 51 http://127.0.0.1:9019/qr/00181.4.png ./td_0008/q00181.4.png http://t432z.com/q/51
-- 182 52 http://127.0.0.1:9019/qr/00182.4.png ./td_0008/q00182.4.png http://t432z.com/q/52
-- 183 53 http://127.0.0.1:9019/qr/00183.4.png ./td_0008/q00183.4.png http://t432z.com/q/53
-- 184 54 http://127.0.0.1:9019/qr/00184.4.png ./td_0008/q00184.4.png http://t432z.com/q/54
-- 185 55 http://127.0.0.1:9019/qr/00185.4.png ./td_0008/q00185.4.png http://t432z.com/q/55
-- 186 56 http://127.0.0.1:9019/qr/00186.4.png ./td_0008/q00186.4.png http://t432z.com/q/56
-- 187 57 http://127.0.0.1:9019/qr/00187.4.png ./td_0008/q00187.4.png http://t432z.com/q/57
-- 188 58 http://127.0.0.1:9019/qr/00188.4.png ./td_0008/q00188.4.png http://t432z.com/q/58
-- 189 59 http://127.0.0.1:9019/qr/00189.4.png ./td_0008/q00189.4.png http://t432z.com/q/59
-- 190 5a http://127.0.0.1:9019/qr/00190.4.png ./td_0008/q00190.4.png http://t432z.com/q/5a
-- 191 5b http://127.0.0.1:9019/qr/00191.4.png ./td_0008/q00191.4.png http://t432z.com/q/5b
-- 192 5c http://127.0.0.1:9019/qr/00192.4.png ./td_0008/q00192.4.png http://t432z.com/q/5c
-- 193 5d http://127.0.0.1:9019/qr/00193.4.png ./td_0008/q00193.4.png http://t432z.com/q/5d
-- 194 5e http://127.0.0.1:9019/qr/00194.4.png ./td_0008/q00194.4.png http://t432z.com/q/5e
-- 195 5f http://127.0.0.1:9019/qr/00195.4.png ./td_0008/q00195.4.png http://t432z.com/q/5f
-- 196 5g http://127.0.0.1:9019/qr/00196.4.png ./td_0008/q00196.4.png http://t432z.com/q/5g
-- 197 5h http://127.0.0.1:9019/qr/00197.4.png ./td_0008/q00197.4.png http://t432z.com/q/5h
-- 198 5i http://127.0.0.1:9019/qr/00198.4.png ./td_0008/q00198.4.png http://t432z.com/q/5i
-- 199 5j http://127.0.0.1:9019/qr/00199.4.png ./td_0008/q00199.4.png http://t432z.com/q/5j
-- 200 5k http://127.0.0.1:9019/qr/00200.4.png ./td_0008/q00200.4.png http://t432z.com/q/5k
-- 201 5l http://127.0.0.1:9019/qr/00201.4.png ./td_0008/q00201.4.png http://t432z.com/q/5l
-- 202 5m http://127.0.0.1:9019/qr/00202.4.png ./td_0008/q00202.4.png http://t432z.com/q/5m
-- 203 5n http://127.0.0.1:9019/qr/00203.4.png ./td_0008/q00203.4.png http://t432z.com/q/5n
-- 204 5o http://127.0.0.1:9019/qr/00204.4.png ./td_0008/q00204.4.png http://t432z.com/q/5o







-- pick out a QR to use
-- update the QR row
-- do get/post request to update where to on QR code
drop FUNCTION v1_next_avail_qr ;
CREATE OR REPLACE FUNCTION v1_next_avail_qr ()
	RETURNS varchar AS 
$body$
DECLARE
    l_id char varying(40);
    l_file_name char varying(240);
    l_url_path char varying(240);
    l_qr_id char varying(30);
    l_qr_enc_id char varying(30);
	l_data				varchar (800);
	l_fail				bool;
BEGIN

	l_fail = false;
	l_data = '{"status":"success"}';

	select "id", "file_name", "url_path", "qr_id", "qr_enc_id"
		into l_id, l_file_name, l_url_path, l_qr_id, l_qr_enc_id 
		from "v1_avail_qr"
		where "state" = 'avail'
		order by "ord_seq"
		limit 1
		;
	if not found then
		l_fail = true;
		l_data = '{"status":"error","code":"100","msg":"unable to generate QR code"}';
	end if;

	update "v1_avail_qr"
		set "state" = 'used'
		where "id" = l_id
		;
	
	if not l_fail then
		l_data = '{"status":"success"'
				||'"id":'||to_json(l_id)
				||',"file_name":'||to_json(l_file_name)
				||',"url_path":'||to_json(l_url_path)
				||',"qr_id":'||to_json(l_qr_id)
			||'}';
	end if;

	RETURN l_data;
END;
$body$
LANGUAGE plpgsql;

-- select v1_next_avail_qr() as "x";






-- TODO - conv from site_id/sub_id -> tag
drop FUNCTION v1_getTagId p_site_id varchar, p_sub_id varchar );
CREATE OR REPLACE FUNCTION v1_getTagId ( p_site_id varchar, p_sub_id varchar )
	RETURNS varchar AS 
$body$
DECLARE
    l_id 				char varying(40);
	l_data				varchar (800);
	l_tag				varchar (80);
	l_fail				bool;
BEGIN

	l_fail = false;
	l_data = '{"status":"success"}';

	select "tag"
		into l_tag
		from "v1_trackAdd"
		where "premis_id" = p_site_id
		  and "premis_animal" = p_sub_id
		limit 1
		;
	if not found then
		l_fail = true;
		l_data = '{"status":"error","code":"101","msg":"unable to convert to tag"}';
	end if;

	if not l_fail then
		l_data = '{"status":"success"'
				||'"tag":'||to_json(l_tag)
			||'}';
	end if;

	RETURN l_data;
END;
$body$
LANGUAGE plpgsql;


-- update "v1_trackAdd" set "premis_id" = '500', "premis_animal" = '3' where "tag" = '34000000000001';
-- update "v1_trackAdd" set "premis_id" = '500', "premis_animal" = '4' where "tag" = '34000000000002';
-- select v1_getTagId ( '500', '3' );



/*
func getInfo(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("getInfo called\n")
	fmt.Fprintf(os.Stderr, "getInfo called\n")


	fmt.Fprintf(www, `{"status":"success"}`)
}
*/
-- pull back all info on a cow. (Differnt from JSON below?)
drop FUNCTION v1_getInfo ;
CREATE OR REPLACE FUNCTION v1_getInfo ( p_tag varchar )
	RETURNS varchar AS 
$body$
DECLARE
    l_id 				char varying(40);
	l_data				varchar (8000);
	l_tag				varchar (80);
	-- xyzzy
	l_fail				bool;
BEGIN

	l_fail = false;
	l_data = '{"status":"success"}';

	// TODO
	-- xyzzy

	RETURN l_data;
END;
$body$
LANGUAGE plpgsql;





/*
func convToJson(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("convToJson called\n")
	fmt.Fprintf(os.Stderr, "convToJson called\n")

	// TODO -- get all the info on a cow and convert to JSON and return

	fmt.Fprintf(www, `{"status":"success"}`)
}
*/
drop FUNCTION v1_convToJson ;
CREATE OR REPLACE FUNCTION v1_convToJson ()
	RETURNS varchar AS 
$body$
DECLARE
    l_id 				char varying(40);
	l_data				varchar (800);
	l_fail				bool;
BEGIN

	l_fail = false;
	l_data = '{"status":"success"}';

	// TODO

	RETURN l_data;
END;
$body$
LANGUAGE plpgsql;


















CREATE OR REPLACE function v1_trackAdd_upd()
RETURNS trigger AS 
$BODY$
BEGIN
NEW.updated := current_timestamp;
RETURN NEW;
END
$BODY$
LANGUAGE 'plpgsql';


CREATE TRIGGER v1_trackAdd_trig
BEFORE update ON "v1_trackAdd"
FOR EACH ROW
EXECUTE PROCEDURE v1_trackAdd_upd();




CREATE OR REPLACE function v1_avail_qr_upd()
RETURNS trigger AS 
$BODY$
BEGIN
NEW.updated := current_timestamp;
RETURN NEW;
END
$BODY$
LANGUAGE 'plpgsql';


CREATE TRIGGER v1_avail_qr_trig
BEFORE update ON "v1_avail_qr"
FOR EACH ROW
EXECUTE PROCEDURE v1_avail_qr_upd();



