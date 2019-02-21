
-- Documenta Tracking API and tables
-- Pulled directory from the Ethereum Plumbing, LLC document tracking system

CREATE SEQUENCE dt_seq
  INCREMENT 1
  MINVALUE 1
  MAXVALUE 9223372036854775807
  START 1000
  CACHE 1;


drop TABLE "dt_document" cascade ;
CREATE TABLE "dt_document" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "origin"				char varying (50)
	, "title"				text
	, "desc"				text
	, "tags"				text
	, "eth_logged"			text
	, "eth_hash"			char varying (50)
	, "category_id"			char varying (40)
	, "qr_code"				text
	, "status"				char varying (80) default 'live'
	, "created_date" 		timestamp 									 						
	, "update_chain_id"		char varying (40)
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);

drop TABLE "dt_category" cascade ;
CREATE TABLE "dt_category" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "category"			title
);

create unique index "dt_category_u1" on "dt_category" ( "category" );

drop TABLE "dt_doc_image" cascade ;
CREATE TABLE "dt_doc_image" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "seq"	 				bigint DEFAULT nextval('dt_seq'::regclass) NOT NULL 
	, "file_name"			text
	, "merkle_hash"			char varying(50)
	, "merkle_leaf"			text
);
