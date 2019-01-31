
-- list of setup and validated devices

CREATE TABLE "t_2fa" (
	  "id"					uuid DEFAULT uuid_generate_v4() not null primary key
	, "user_id"				char varying (40) 
	, "user_hash"			text 
	, "fp"					text 
	, "updated" 			timestamp 									 						
	, "created" 			timestamp default current_timestamp not null 					
);

-- m4_updTrig(t_2fa)

-- list of user one time keys

CREATE TABLE "t_2fa_otk" (
	  "id"					uuid DEFAULT uuid_generate_v4() not null primary key
	, "user_id"				char varying (40) 
	, "one_time_key"		text
	, "updated" 			timestamp 									 						
	, "created" 			timestamp default current_timestamp not null 					
);

-- m4_updTrig(t_2fa_otk)
