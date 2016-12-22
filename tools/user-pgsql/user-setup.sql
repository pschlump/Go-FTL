
drop TABLE "basic_auth" ;
CREATE TABLE "basic_auth" (
	  "username"				char varying (200) not null primary key
	, "salt"					char varying (100) not null
	, "password"				char varying (180) not null 
);
insert into "basic_auth" ( "username", "salt", "password" ) values ( 'example.com:testme', 'salt', '4c205db6b361042ee973f0341433088922232dfb41d6b0721f8f91747bd0f71fc8ccefe250c3233c2c85a3e70e78d11cd98b8cf1d5f7a797f71dd2069a8fcc62' );

\q
