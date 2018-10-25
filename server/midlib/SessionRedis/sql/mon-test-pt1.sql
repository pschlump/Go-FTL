-- cleanup.
delete from "t_monitor_stuff" where "id" = '000ef9f5-c82d-4866-8316-a9097cbcff97';
delete from "t_monitor_stuff" where "item_name" = 'bob';

-- Create new item
select t_mon_add ( '000ef9f5-c82d-4866-8316-a9097cbcff97', 'bob', '2 minute', 'on' );

-- Enable item
select t_mon_enable ( '000ef9f5-c82d-4866-8316-a9097cbcff97', 'on' );

update "t_monitor_stuff"
	set "timeout_event" = current_timestamp - interval ' 1 minute '
	where "id" = '000ef9f5-c82d-4866-8316-a9097cbcff97'
;

select * from prep_info5();

-- select ping_i_am_alive( 'bob', '127.0.0.1', 'hi there ' );
-- 
-- select * from prep_info5();
-- 
-- -- cleanup.
-- delete from "t_monitor_stuff" where "id" = '000ef9f5-c82d-4866-8316-a9097cbcff97';
-- delete from "t_monitor_stuff" where "item_name" = 'bob';
