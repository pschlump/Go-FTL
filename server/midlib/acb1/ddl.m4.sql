
-- stmt := "insert into \"v1_trackAdd\" ( \"tag\" ) values ( $1 )"
m4_changequote(`[[[', `]]]')

CREATE SEQUENCE v1_tracAdd_seq
INCREMENT 1
MINVALUE 1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

drop TABLE "v1_trackAdd" ;
CREATE TABLE "v1_trackAdd" (
"id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
, "tag"					char varying (40) not null
, "state"				char varying (10) check ( "state" in ( 'new', 'hashed', 'error', 's1', 's2' ) ) default 'new' not null
, "premis_id"			text
, "premis_animal"		text
, "ord_seq"				bigint DEFAULT nextval('v1_tracAdd_seq'::regclass) NOT NULL 
, "qr_id"				char varying (40)		-- id of QR assigned to this.
, "updated" 			timestamp 									 						
, "created" 			timestamp default current_timestamp not null 					
);

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
, "qr_id"				text
, "qr_enc_id"			text
, "state"				char varying (10) check ( "state" in ( 'avail', 'used', 's1', 's2' ) ) default 'avail' not null
, "ord_seq"				bigint DEFAULT nextval('v1_qr_avail_seq'::regclass) NOT NULL 
, "updated" 			timestamp 									 						
, "created" 			timestamp default current_timestamp not null 					
);


-- select t1.*,
-- 	t2.*
-- from "v1_trackAdd" as t1 left join "v1_avail_qr" on t1."qr_id" = t2."id"

insert into "v1_avail_qr" ( "file_name", "url_path", "qr_id", "qr_enc_id") values
	( 'qr001.png', 'http://t5432z/q/1.png', '1', '1' )
,	( 'qr002.png', 'http://t5432z/q/2.png', '2', '2' )
,	( 'qr003.png', 'http://t5432z/q/3.png', '3', '3' )
,	( 'qr004.png', 'http://t5432z/q/3.png', '4', '4' )
;

/*
	fmt.Printf("generateQrFor called\n")
	fmt.Fprintf(os.Stderr, "generateQrFor called\n")

	// TODO
	// pick out a QR to use
	stmt := "select * from \"v1_avail_qr\" where \"state\" = 'avail'"
	// update the QR row
	stmt = "update \"v1_avail_qr\" where \"state\" = 'avail'"
	// do get/post request to update where to on QR code

	fmt.Fprintf(www, `{"status":"success"}`)

func generateQrFor(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("generateQrFor called\n")
	fmt.Fprintf(os.Stderr, "generateQrFor called\n")

	stmt := "select v1_next_avail_qr as \"x\""

	// TODO - call function, return x

	fmt.Fprintf(www, `{"status":"success"}`)
}

func getTagId(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("getTagId called\n")
	fmt.Fprintf(os.Stderr, "getTagId called\n")

	// TODO - convert a premis_id/sub_id -> tag id and return

	fmt.Fprintf(www, `{"status":"success"}`)
}

func getInfo(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("getInfo called\n")
	fmt.Fprintf(os.Stderr, "getInfo called\n")

	// TODO - get all the info on a cow

	fmt.Fprintf(www, `{"status":"success"}`)
}

func convToJson(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("convToJson called\n")
	fmt.Fprintf(os.Stderr, "convToJson called\n")

	// TODO -- get all the info on a cow and convert to JSON and return

	fmt.Fprintf(www, `{"status":"success"}`)
}

func chainHash(hdlr *Acb1Type, rw *goftlmux.MidBuffer, www http.ResponseWriter, req *http.Request, mdata map[string]string) {
	fmt.Printf("chainHash called\n")
	fmt.Fprintf(os.Stderr, "chainHash called\n")

*/

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


m4_updTrig(v1_trackAdd)
m4_updTrig(v1_avail_qr)

