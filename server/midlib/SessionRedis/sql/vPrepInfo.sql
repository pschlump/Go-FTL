
-- <div class="itemTitle {{col.cssStat}}">{{col.name}}</div>
-- <div class="itemBody" ng-bind-html-unsafe="col.info"></div>
-- dn.runQuery ( stmt = ts0( 'select /*vPrepInfo.sql*/ * from "t_monitor_results" where "id" = \'%{it_work_id%}\' order by "seq" ', { "it_work_id":it_work_id } ), function ( err, result ) {
CREATE TABLE "t_monitor_results" (
	  "id"					char varying (40) DEFAULT uuid_generate_v4() not null 
	, "seq"					bigint DEFAULT nextval('t_host_id_seq'::regclass) NOT NULL 
	, "sev"					bigint DEFAULT 0 not null
	, "cssStat"				char varying (80) not null 				
	, "name"				char varying (80) not null 				
	, "info"				text
	, "updated" 			timestamp 									 						
	, "created" 			timestamp default current_timestamp not null 						--
);
create index "t_monitor_results_p1" on "t_monitor_results" ( "id", "seq" );


-- dn.runQuery ( stmt = ts0( 'select /*vPrepInfo.sql*/ prep_info ( \'%{user_id%}\' ) as "id" ', { "user_id":0 } ), function ( err, result ) {
CREATE or REPLACE FUNCTION prep_info ( p_user_id varchar ) RETURNS varchar AS $$
DECLARE
    work_id 	char varying(40);
	rec 		record;
	l_sev 		bigint;
	l_cssStat 	char varying(80);
	l_name 		char varying(80);
	l_info 		char varying(2000);
BEGIN

	work_id = uuid_generate_v4();

	FOR rec IN
		select "item_name", "event_to_raise", "delta_t", 'error' as "status"
			from "t_monitor_stuff"
			where "timeout_event" < now()
		union
		select "item_name", "event_to_raise", "delta_t", 'ok' as "status"
			from "t_monitor_stuff"
			where "timeout_event" >= now()
	LOOP
		l_sev = 0;
		l_cssStat = 'itemNormal';
		l_name = rec.item_name;
		l_info = 'On Time: '||rec.event_to_raise;
		if ( rec.status = 'error' ) then
			l_sev = l_sev + 1;
			l_cssStat = 'itemError';
			l_info = 'Missed Deadline: '||rec.event_to_raise;
		end if;
		insert into "t_monitor_results" ( "id", "sev", "cssStat", "name", "info" )
			values ( work_id, l_sev, l_cssStat, l_name, l_info );
	END LOOP;

	RETURN work_id;
END;
$$ LANGUAGE plpgsql;






delete from  "t_monitor_results" ;

select /*vPrepInfo.sql*/ prep_info ( '0' ) as "id" ;

select /*vPrepInfo.sql*/ * from "t_monitor_results" order by "sev" desc, "seq" asc;




-- update /*vTestInteval.sql*/ "t_monitor_stuff" set "timeout_event" = current_timestamp + CAST("delta_t" as Interval) where "item_name" = 'InternetUp' ;

;

