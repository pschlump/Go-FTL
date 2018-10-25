
all: s_nv.sql

s_nv.sql: s_nv.m4.sql
	m4 -P setup.m4 s_nv.m4.sql >s_nv.sql

test: test_setup test_001 test_002 test_003 test_004 test_005 test_006
	echo PASS

test_setup:
	-rm test/save* test/upd* test/get* test/list* test/del*

test_001:
	( cd test ; wget 'http://localhost:9001/api/store/saveNVKey?name=aaa&value=aaa-value' )
	grep success test/save*

test_002:
	( cd test ; wget 'http://localhost:9001/api/store/updNVKey?name=aaa&value=aaa-value1' )
	grep success test/upd*

test_003:
	( cd test ; wget 'http://localhost:9001/api/store/updNVKey?name=bbb&value=aaa-value1' )
	grep success test/upd*

test_004:
	( cd test ; wget 'http://localhost:9001/api/store/getNVKey?name=bbb' )
	grep success test/get*

test_005:
	( cd test ; wget 'http://localhost:9001/api/store/listNVKey' )
	grep success test/list*

test_006:
	( cd test ; wget 'http://localhost:9001/api/store/delNVKey?name=bbb' )
	( cd test ; wget 'http://localhost:9001/api/store/delNVKey?name=aaa' )
	( cd test ; wget 'http://localhost:9001/api/store/delNVKey?name=b' )
	( cd test ; wget 'http://localhost:9001/api/store/delNVKey?name=a' )
	grep success test/del*


