
-- drop TABLE "basic_auth" ;
CREATE TABLE "basic_auth" (
	  "user_id"				char varying (40) DEFAULT uuid_generate_v4() not null primary key		
	, "username"			char varying (200) not null 
	, "salt"				char varying (200) not null
	, "password"			char varying (280) not null 
);

create unique index "basic_auth_u1" on "basic_auth" ( "username" );

delete from "basic_auth" where "username" = 'example.com:testme';
insert into "basic_auth" ( "username", "salt", "password" ) values ( 'example.com:testme', 'salt', 
	'9b6095510e3e1c0ea568c3faf29e545c364265d017b16614b1a2de3efe96bc6313cb9e1d221134a46fd5faa8499ebb8568a2ec489e32fa4c4adcd89c05394292'
);

\q
