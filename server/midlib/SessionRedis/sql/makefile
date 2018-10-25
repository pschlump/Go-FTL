
all: tables-pg.sql auth.sql 



tables-pg.sql: tables-pg.m4.sql
	m4 -P tables-pg.m4.sql >tables-pg.sql
	
ang24.sql: ang24.m4.sql
	m4 -P ang24.m4.sql >ang24.sql

SafetyApp.trig.sql: SafetyApp.trig.m4.sql
	m4 -P SafetyApp.trig.m4.sql >SafetyApp.trig.sql

bughole.sql: bughole.m4.sql
	m4 -P bughole.m4.sql >bughole.sql
	
auth.sql: auth.m4.sql
	m4 -P auth.m4.sql >auth.sql

t_run_q.sql: t_run_q.m4.sql
	m4 -P t_run_q.m4.sql >t_run_q.sql

