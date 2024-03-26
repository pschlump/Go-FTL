
SVR=http://localhost:16040
KEY=8HbvBJR1qWkpMgGxzqj2sRvuXg

BUILD=p_track.sql p_document.sql p_issue.sql support.sql p_user.sql

all: $(BUILD)

p_track.sql: p_track.m4.sql
	m4 -P p_track.m4.sql >p_track.sql

p_document.sql: p_document.m4.sql
	m4 -P p_document.m4.sql >p_document.sql

p_issue.sql: p_issue.m4.sql
	m4 -P p_issue.m4.sql >p_issue.sql

p_user.sql: p_user.m4.sql
	m4 -P p_user.m4.sql >p_user.sql

support.sql: support.m4.sql
	m4 -P support.m4.sql >support.sql

test: test_setup test2

test_setup: $(BUILD)
	mkdir -p ./out
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <support.sql		>out/support.lst 2>&1
	./syn.sh out/support.lst
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <p_issue.sql		>out/p_issue.lst 2>&1
	./syn.sh out/p_issue.lst
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <p_document.sql	>out/p_document.lst 2>&1
	./syn.sh out/p_document.lst
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <p_track.sql		>out/p_track.lst 2>&1
	./syn.sh out/p_track.lst
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <docs.sql 		>out/docs.out 2>&1

# test 3 of the email list functions - at database level	
# 	CREATE or REPLACE FUNCTION reg_email_list ( p_to_addr varchar, p_ip_addr varchar )
# 	CREATE or REPLACE FUNCTION dereg_email_list ( p_to_addr varchar, p_ip_addr varchar )
# 	CREATE or REPLACE FUNCTION confirm_email_list ( p_to_addr varchar, p_ip_addr varchar )
test2:
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <test2.sql >out/test2.out
	./check.sh out/test2.out

# Load data for go-ftl documentation - key word search
load_data:
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <docs.sql >out/docs.out


# ##########################################################################################################################
# Test of /api end points
# ##########################################################################################################################

test3:
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <t3.setup.sql 
	wget -o out/t3.log -O out/t3.out '$(SVR)/api/registerEmail?to_addr=b@b.com&api_table_key=$(KEY)'
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <t3.check.sql  >./out/t3.chk1.out
	./check.sh out/t3.chk1.out
	@echo PASS

test3a:
	wget -o out/t3.log -O out/t3.out 'http://www.go-ftl.com/api/registerEmail?to_addr=b@b.com'

test4: test3
	wget -o out/t4.log -O out/t4.out '$(SVR)/api/deRegisterEmail?to_addr=b@b.com&api_table_key=$(KEY)'
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <t4.check.sql  >./out/t4.chk1.out
	./check.sh out/t4.chk1.out
	@echo PASS

test5: test4
	wget -o out/t5.log -O out/t5.out '$(SVR)/api/confirmEmailList?to_addr=b@b.com&api_table_key=$(KEY)'
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <t5.check.sql  >./out/t5.chk1.out
	./check.sh out/t5.chk1.out
	@echo PASS

test6:
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <t6.setup.sql 
	wget -o out/t6.log -O out/t6.out '$(SVR)/api/grabFeedback?info=some-info&body=some-body&api_table_key=$(KEY)'
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <t6.check.sql  >./out/t6.chk1.out
	./check.sh out/t6.chk1.out
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <t6.setup.sql 
	wget -o out/t6.log -O out/t6.out '$(SVR)/api/logit?info=some-info&body=some-body&api_table_key=$(KEY)'
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <t6.check.sql  >./out/t6.chk1.out
	./check.sh out/t6.chk1.out
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <t6.setup.sql 
	@echo PASS

test7: 
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <t7.setup.sql  >./out/t7.setup.out
	./chk-for-err.sh ./out/t7.setup.out
	wget -o out/t7.log -O out/t7.out '$(SVR)/api/saveEmailMessage?email_group=some-group&person_name=some-person-name&email_addr=b@b.com&message_body=some-message-body&api_table_key=$(KEY)'
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <t7.check.sql  >./out/t7.chk1.out
	./check.sh out/t7.chk1.out
	./chk-for-err.sh  out/t7.chk1.out
	psql -p 5433 -a -P pager=off -h 127.0.0.1 -U pschlump <t7.setup.sql 
	@echo PASS

#
# TODO
#
#	,"/api/serverVersion": { "g": "server_version", "p": [ "$ip$"]
#		, "LoginRequired":false
#		, "LineNo":"Line: __LINE__ File: __FILE__"
#		, "Method":["GET","POST"]
#		, "TableList":["t_version"]
#		, "valid": {
#			 "$ip$": 		{ "required":true, "type":"string", "max_len":40, "min_len":4 }
#			}
#		}

