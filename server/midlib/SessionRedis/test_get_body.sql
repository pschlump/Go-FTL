
--SELECT routine_body 
--FROM information_schema.routines 
--WHERE specific_schema LIKE 'public'
--  and routine_name LIKE 's_echo_builtin'
--;

SELECT 'DROP '
    || CASE WHEN p.proisagg THEN 'AGGREGATE ' ELSE 'FUNCTION ' END
    || quote_ident(n.nspname) || '.' || quote_ident(p.proname) || '(' 
    || pg_catalog.pg_get_function_identity_arguments(p.oid) || ');' AS stmt
FROM   pg_catalog.pg_proc p
JOIN   pg_catalog.pg_namespace n ON n.oid = p.pronamespace
WHERE  n.nspname = 'public'                     -- schema name (optional)
AND    p.proname ILIKE 's_password%'            -- function name
-- AND pg_catalog.pg_function_is_visible(p.oid) -- function visible to user
ORDER  BY 1
;
