





-- "Copyright (C) Philip Schlump, 2009-2017." 

-- drop TABLE "t_auth_token" ;

CREATE TABLE "t_auth_token" (
	  "auth_token"			uuid not null primary key
	, "user_id"				uuid not null 
	, "expire"	 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);

create index "t_auth_token_p1" on "t_auth_token" ( "user_id" );
create index "t_auth_token_p2" on "t_auth_token" ( "expire" );

