SELECT routines.routine_name
	, parameters.data_type
	, parameters.parameter_name
	, parameters.ordinal_position
FROM information_schema.routines
    JOIN information_schema.parameters ON routines.specific_name=parameters.specific_name
WHERE routines.specific_schema='public'
ORDER BY routines.routine_name, parameters.ordinal_position;
