#!/bin/bash
# package and ship to .206
tar -czf ~/x1.tar.gz \
	Makefile check.sh chk-for-err.sh common.m4.sql docs.sql note.1 p_document.m4.sql p_document.sql p_issue.m4.sql \
	p_issue.sql p_track.m4.sql p_track.sql pkg1.sh sql-cfg-new.json support.m4.sql support.sql syn.sh \
	t3.check.sql t3.setup.sql t4.check.sql t5.check.sql t6.check.sql t6.setup.sql t7.check.sql \
	t7.setup.sql test1.sql test2.sql
scp ~/x1.tar.gz pschlump@198.58.107.206:/home/pschlump
