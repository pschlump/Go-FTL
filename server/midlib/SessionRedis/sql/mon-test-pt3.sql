
update "t_monitor_stuff"
	set "timeout_event" = current_timestamp - interval ' 1 minute '
	where "id" = '000ef9f5-c82d-4866-8316-a9097cbcff97'
;

