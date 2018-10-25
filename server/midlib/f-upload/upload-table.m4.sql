

-- Copyright (C) Philip Schlump, 2014-2016.  All rights reserved.

m4_changequote(`[[[', `]]]')
m4_include(common.m4.sql)

drop table "p_uploaded_file" ;
create table "p_uploaded_file" (
	  "id"						char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "raw_file_name"			text				-- File name as supplied by user
	, "base_file_name"			text				-- File name striped of path as user supplied
	, "scrubed_file_name"		text				-- 
	, "sha1_file_name"			text				-- File name as stored on system - full -
	, "desc"					text				-- User data
	, "qr_text"					text				-- if QR image, then text
	, "size_in_bytes"			integer				-- actual stored size on disk
	, "file_type"				char varying (80) 	-- Extension on file / or mime type??
	, "user_id"					char varying (40) 	-- Person this file belongs to
	, "height"					integer				-- if image, then h,w else 0, svg not suppoted yet.
	, "width"					integer
	, "created"					timestamp default current_timestamp not null
	, "updated"					timestamp
	, "lifespan_of_file"		integer default 0	-- time to keep file, 0 is forever
	, "SKU"						char varying (50)
	, "fk_id"					char varying (40) 	-- product_id, option_id etc.
);

m4_updTrig(p_uploaded_file)

