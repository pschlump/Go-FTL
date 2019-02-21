
-- stmt := "insert into \"v1_trackAdd\" ( \"tag\" ) values ( $1 )"

drop TABLE "v1_trackAdd" ;
CREATE TABLE "v1_trackAdd" (
	  "id"				char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "tag"				char varying (40) not null
	, "state"			char varying (10) check ( "state" in ( 'new', 'hashed', 'error' ) ) default 'new' not null
);



