create table "t_config_data"  (
	  "name"					char varying (40) 
	, "value"					char varying (250) 
);
delete from "t_config_data";
insert into "t_config_data"  ( "name" , "value" ) values ( 'host-ip', '0.0.0.0' );
insert into "t_config_data"  ( "name" , "value" ) values ( 'i-am', 'dev' );
