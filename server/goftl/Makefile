
all: 
	rm -f goftl
	./cmp-local.sh

corp_reg_test1:
	echo " Ya " "gen_filename=${gen_filename}" >>,,,

corp_reg_test2:
	echo " Ya Yep Yes " >>,,,

old_all:
	( cd ./../midlib/aessrp ; make test_version )
	rm -f goftl
	go build
	./rs.sh

206: dot206.jsonx loc206.jsonx

dot206.jsonx: 206cfg.m4 dot206.m4.json 
	m4 -P 206cfg.m4 dot206.m4.json >dot206.jsonx

loc206.jsonx: 206loc.m4 dot206.m4.json 
	m4 -P 206loc.m4 dot206.m4.json >loc206.jsonx

deploy74: build_linux copyTo74

build_linux:
	./cmp-prod.sh go-ftl.linux

copyTo74:
	scp go-ftl.linux pschlump@192.154.97.74:/tmp

binaries:
	rm ./goftl
	go build	
	mkdir -p ~/Projects/www.go-ftl.com_doc/_site/download
	tar -czf ~/Projects/www.go-ftl.com_doc/_site/download/goftl.MacOS.10.10.5_ver0.5.9.tar.gz ./goftl

# rerun the server without doing a "go build" and recompiling it.
rr:
	./rs.sh

updBuildNo:
	( cd ../.. ; git commit -m "Set Build No on files." . )
	( cd .. ; ./goftl/updBuildNo.sh goftl/main.go bufferhtml/bufferhtml.go cfg/cfg.go cfg/input.go fileserve/fs.go fileserve/fs_ftl.go httpmux/server.go \
		inMemoryCache/inMemoryCache.go lib/lib.go mid/auth_srp.go mid/cache.go mid/go_template.go mid/m5.go mid/mid.go mid/simple_file_server.go \
		mid/simple_proxy.go mid/standard_file_server.go mid/strip_prefix.go mid/top.go midlib/LoginRequired/LoginRequired.go midlib/aessrp/raw_src/aessrp_ext.go \
		midlib/aessrp/raw_src/an_error.go midlib/aessrp/raw_src/redis_stuff.go midlib/banip/ip_ban.go midlib/basicauth/basicauth.go midlib/basicpgsql/basicpg.go \
		midlib/basicredis/basicredis.go midlib/directory_browse/directory_browse.go midlib/dumpit/dumpIt.go midlib/dumpreq/dumpreq.go \
		midlib/geo_ip_filter/geo_ip_filter.go midlib/gzip/gzip.go midlib/header/header.go midlib/jsonp/jsonp.go midlib/limit_ext/limit_ext.go \
		midlib/limit_re_path/limit_re_path.go midlib/limit_to/limit_to.go midlib/logging/logging.go midlib/redirect/redirect.go \
		midlib/reject_ext/reject_ext.go midlib/reject_path/reject_path.go midlib/reject_re_path/reject_re_path.go midlib/rewrite/rewrite.go \
		midlib/rewrite_proxy/rewrite_proxy.go midlib/status/status.go midlib/redisList/redisList.go midlib/redisList/redisList_test.go \
		)

misc_no_buildno:
	echo goftl/inc.go
	echo goftl/util_bsd.go
	echo goftl/util_linux.go
	echo goftl/util_windows.go





run-test001:
	./run-test.sh test001

verify-test001:
	echo Test verify-test001
	curl -b cookies.txt -c cookies.txt 'http://localhost:9001/t01.html' >output/test001.ref.out
	diff output/test001.ref.out ./ref
	@echo PASS-test001

test-01-test001:
	echo Test test-01-test001
	curl 'http://localhost:9001/Status' 

db-check-test001:
	echo Test db-check-test001
	curl -b cookies.txt -c cookies.txt 'http://localhost:9001/api/session/status_db1' >output/test002.ref.out
	echo diff output/test002.ref.out ./ref
	diff -w output/test002.ref.out ./ref/test002.ref-no-while.out
	@echo PASS-test002


#	,"/api/session/login": { "g": "s_login", "p": [ "username", "password", "$ip$", "$host$" ], "nokey":true
# select s_login('test01', 'yanky444',  '1.1.1.1',  'http://www.2c-why.com/' );
login-test001:
	@echo Test Login-test001 pw=yanky444
	curl -X POST -b cookies.txt -c cookies.txt 'http://localhost:9001/api/session/login?username=test01&password=yanky444' 
	@echo

login-test002:
	@echo Test Login-test001 pw=yanky444
	curl -X POST -b cookies.txt -c cookies.txt 'http://localhost:9001/api/session/login?username=test01&password=yanky444' 
	@echo
	@echo
	@echo Test Login-test001 pw=test01
	curl -X POST -b cookies.txt -c cookies.txt 'http://localhost:9001/api/session/login?username=test01&password=test01' 
	@echo
	@echo
	@echo Test Login-test001 pw=bobob4
	curl -X POST -b cookies.txt -c cookies.txt 'http://localhost:9001/api/session/login?username=test01&password=bobob4' 
	@echo

# select s_register_new_user('test01', '123456', '123456', '1.1.1.1', 'kermit.nosend.01@gmail.com', 'Kermit Frog', 'http://www.2c-why.com/', 'test-application' );
#	,"/api/session/register_new_user": { "g": "s_register_new_user", "p": [ "username", "password", "again", "$ip$", "email", "real_name", "$url$", "app" ], "nokey":true
reg-test002:
	@echo Test reg-test002 pw=yanky444 "Register a new user"
	d9 <test_reg-test002.sql >./output/test_reg01.out
	curl -X POST -b cookies.txt -c cookies.txt 'http://localhost:9001/api/session/register_new_user?username=test03&password=test03&again=test03&email=kermit.01@the-green-pc.com&real_name=Kermit+The+Frog&app=reg-test002' 
	@echo

email-test004:
	@echo Test email-test004 tn='confirm_registration'
	curl -X POST -b cookies.txt -c cookies.txt 'http://localhost:9001/api/session/simulate_email?tn=confirm_registration&email_token=AAAbbbCCCdddEEE&app=email-test004&kp=p' 
	@echo

#s_change_password.m4.sql:				||'"template":"password_recovered"'		# 5
email-test005:
	@echo Test email-test005 tn='password_recovered'
	curl -X POST -b cookies.txt -c cookies.txt 'http://localhost:9001/api/session/simulate_email?tn=password_recovered&email_token=AAAbbbCCCdddEEE&app=email-test005&kp=p' 
	@echo

#s_change_password.m4.sql:				||'"template":"password_changed"'		# 6
email-test006:
	@echo Test email-test006 tn='password_changed'
	curl -X POST -b cookies.txt -c cookies.txt 'http://localhost:9001/api/session/simulate_email?tn=password_changed&email_token=AAAbbbCCCdddEEE&app=email-test006&kp=p' 
	@echo


validate-token:
	@echo Test validate token
	curl -X POST -b cookies.txt -c cookies.txt 'http://localhost:9001/api/session/validate_auth_token?auth_token=c9b0d60e-249a-4518-afd6-a209e86c40a0&_ran_=84e5a262-4df7-407a-5a22-7a47f25dd1b5'









test1: clean1
	go build
	( ( sleep 2 ; make test1a ) & )
	./cli2 -c 7-cfg.json

test1a: 
	( cd t1a ; make test1a )
	
clean1:
	( cd /Users/corwin/Projects/ftl0/cli2/cache ; rm -rf ??* )
	( cd /Users/corwin/Projects/ftl0/cli2; rm -f utf1.json )
 
test2:
	go build
	./cli2 -c 2-cfg.json

# test JSONP and Header ----------------
test3:
	go build
	./cli2 -c 5-cfg.json

test4: 
	-rm cli2
	go build
	./cli2 -c file-test.cfg.json

