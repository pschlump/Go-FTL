SELECT routines.routine_name
			FROM information_schema.routines
			WHERE routines.specific_schema = 'public'
			  and routines.routine_name = lower('fetchUserClass')
;
