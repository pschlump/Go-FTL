
m4_changequote(`[[[', `]]]')
m4_include(common.m4.sql)

drop table "p_document" ;
create table "p_document" (
	  "id"						char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "doc"						text
	, "file_name"				text
	, "link"					text
	, "title"					text
	, "desc"					text
	, "group"					char varying (50) default 'all'
	, "updated" 				timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 				timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
	, "__keyword__"				tsvector
);

CREATE TRIGGER tsvectorupdate_on_p_document BEFORE INSERT OR UPDATE ON "p_document"
	FOR EACH ROW
	EXECUTE PROCEDURE tsvector_update_trigger('__keyword__', 'pg_catalog.english', 'title', 'desc', 'doc');

CREATE INDEX "p_document_keyword_p1" ON "p_document" USING gin("__keyword__");

m4_updTrig(p_document)

