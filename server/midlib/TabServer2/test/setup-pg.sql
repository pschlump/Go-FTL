
-- -------------------------------------------------------- -- --------------------------------------------------------
drop TABLE "test_stuff" ;
CREATE TABLE "test_stuff" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null primary key
	, "hostname"			char varying (40) not null 				
	, "item_name"			char varying (240) not null 				
	, "status"				char varying (40) 
	, "code"				int
	, "updated" 			timestamp 									 						-- Project update timestamp (YYYYMMDDHHMMSS timestamp).
	, "created" 			timestamp default current_timestamp not null 						-- Project creation timestamp (YYYYMMDDHHMMSS timestamp).
);

insert into "test_stuff" ( "hostname", "item_name", "status", "code" ) values
	( 'dev2',      'backups', 'done', 01 ),
	( 'dev3',      'backups', 'i.p.', 02 ),
	( 'chantelle', 'backups', 'done', 01 ),
	( 'sasha',     'backups', 'pend', 03 ),
	( 'joyce',     'backups', 'done', 01 ),
	( 'corwin',    'backups', 'fail', 11 ),
	( 'mac1',      'backups', 'done', 01 ),
	( 'mac2',      'backups', 'fail', 08 ),
	( 'mac3',      'backups', 'done', 01 )
;

