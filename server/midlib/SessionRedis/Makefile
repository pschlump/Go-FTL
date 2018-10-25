
all:
	go build

doc:
	markdown-cli -i SessionRedis.md -o SessionRedis.html -c ../markdown-cfg.json

mkdocs:
	cat SessionRedis.md >>../tmp2
	echo "SessionRedis/SessionRedis.md" >>../m.list

pg: setup_make t_dual.sql \
	t_config.sql t_email_test_id_seq.sql t_email_id_seq.sql config_data.sql \
	t_ip_ban.sql t_customer.sql t_host_id_seq.sql t_host_to_customer.sql t_auth_token.sql t_user.sql \
	s_ip_ban.sql s_get_config_item.sql s_get_xsrf_mode.sql s_get_customer_id_from_url.sql s_nvl.sql \
	s_login.sql s_register_new_user.sql s_logout.sql \
	s_confirm_email.sql s_password_reset.sql s_change_password.sql status_db.sql s_stayLoggedIn.sql \
	s_validate_token.sql s_echo_builtin.sql s_missing_param_test.sql s_validate_auth_token.sql \
	s_validate_logged_in.sql cleanup.sql s_simulate_email.sql s_get_username_from_email.sql \
	t_pwned.sql s_check_pw_valid.sql
.PHONY: pg

setup_make:
	echo "" >all.sql

s_check_pw_valid.sql: s_check_pw_valid.m4.sql
	m4 -P setup.m4 s_check_pw_valid.m4.sql >s_check_pw_valid.sql
	echo "\i s_check_pw_valid.sql" >>all.sql

t_pwned.sql: t_pwned.m4.sql
	m4 -P setup.m4 t_pwned.m4.sql >t_pwned.sql
	echo "\i t_pwned.sql" >>all.sql

t_ip_ban.sql: t_ip_ban.m4.sql
	m4 -P setup.m4 t_ip_ban.m4.sql >t_ip_ban.sql
	echo "\i t_ip_ban.sql" >>all.sql

t_auth_token.sql: t_auth_token.m4.sql
	m4 -P setup.m4 t_auth_token.m4.sql >t_auth_token.sql
	echo "\i t_auth_token.sql" >>all.sql

t_user.sql: t_user.m4.sql
	m4 -P setup.m4 t_user.m4.sql >t_user.sql
	echo "\i t_user.sql" >>all.sql

s_nvl.sql: s_nvl.m4.sql
	m4 -P setup.m4 s_nvl.m4.sql >s_nvl.sql
	echo "\i s_nvl.sql" >>all.sql

s_login.sql: s_login.m4.sql t_ip_ban.sql t_auth_token.sql t_user.sql t_customer.sql config_data.sql
	m4 -P setup.m4 s_login.m4.sql >s_login.sql
	echo "\i s_login.sql" >>all.sql

t_host_id_seq.sql: t_host_id_seq.m4.sql
	m4 -P setup.m4 t_host_id_seq.m4.sql >t_host_id_seq.sql
	echo "\i t_host_id_seq.sql" >>all.sql

t_host_to_customer.sql: t_host_to_customer.m4.sql
	m4 -P setup.m4 t_host_to_customer.m4.sql >t_host_to_customer.sql
	echo "\i t_host_to_customer.sql" >>all.sql

s_register_new_user.sql: s_register_new_user.m4.sql t_user.sql config_data.sql t_ip_ban.sql
	m4 -P setup.m4 s_register_new_user.m4.sql >s_register_new_user.sql
	echo "\i s_register_new_user.sql" >>all.sql

s_logout.sql: s_logout.m4.sql t_auth_token.sql t_user.sql t_ip_ban.sql
	m4 -P setup.m4 s_logout.m4.sql >s_logout.sql
	echo "\i s_logout.sql" >>all.sql

s_confirm_email.sql: s_confirm_email.m4.sql t_ip_ban.sql t_user.sql t_config.sql
	m4 -P setup.m4 s_confirm_email.m4.sql >s_confirm_email.sql
	echo "\i s_confirm_email.sql" >>all.sql

t_config.sql: t_config.m4.sql
	m4 -P setup.m4 t_config.m4.sql >t_config.sql
	echo "\i t_config.sql" >>all.sql

s_password_reset.sql: s_password_reset.m4.sql t_config.sql
	m4 -P setup.m4 s_password_reset.m4.sql >s_password_reset.sql
	echo "\i s_password_reset.sql" >>all.sql

s_change_password.sql: s_change_password.m4.sql t_config.sql t_user.sql t_ip_ban.sql t_host_to_customer.sql
	m4 -P setup.m4 s_change_password.m4.sql >s_change_password.sql
	echo "\i s_change_password.sql" >>all.sql

status_db.sql: status_db.m4.sql
	m4 -P setup.m4 status_db.m4.sql >status_db.sql
	echo "\i status_db.sql" >>all.sql

s_stayLoggedIn.sql: s_stayLoggedIn.m4.sql
	m4 -P setup.m4 s_stayLoggedIn.m4.sql >s_stayLoggedIn.sql
	echo "\i s_stayLoggedIn.sql" >>all.sql

t_dual.sql: t_dual.m4.sql
	m4 -P setup.m4 t_dual.m4.sql >t_dual.sql
	echo "\i t_dual.sql" >>all.sql

config_data.sql: config_data.m4.sql t_config.sql t_customer.sql
	m4 -P setup.m4 config_data.m4.sql >config_data.sql
	echo "\i t_user.sql" >>all.sql

t_customer.sql: t_customer.m4.sql
	m4 -P setup.m4 t_customer.m4.sql >t_customer.sql
	echo "\i t_customer.sql" >>all.sql

s_validate_token.sql: s_validate_token.m4.sql
	m4 -P setup.m4 s_validate_token.m4.sql >s_validate_token.sql
	echo "\i s_validate_token.sql" >>all.sql

# Used in testing for verifying how Go-FTL will translate $name$ items and checking config tables.
s_echo_builtin.sql: s_echo_builtin.m4.sql
	m4 -P setup.m4 s_echo_builtin.m4.sql >s_echo_builtin.sql
	echo "\i s_echo_builtin.sql" >>all.sql

s_missing_param_test.sql: s_missing_param_test.m4.sql
	m4 -P setup.m4 s_missing_param_test.m4.sql >s_missing_param_test.sql
	echo "\i s_missing_param_test.sql" >>all.sql

# FUNCTION s_validate_auth_token ( p_auth_token varchar, p_url varchar );
s_validate_auth_token.sql: s_validate_auth_token.m4.sql
	m4 -P setup.m4 s_validate_auth_token.m4.sql >s_validate_auth_token.sql
	echo "\i s_validate_auth_token.sql" >>all.sql

t_email_test_id_seq.sql: t_email_test_id_seq.m4.sql
	m4 -P setup.m4 t_email_test_id_seq.m4.sql >t_email_test_id_seq.sql
	echo "\i t_email_test_id_seq.sql" >>all.sql

s_validate_logged_in.sql: s_validate_logged_in.m4.sql
	m4 -P setup.m4 s_validate_logged_in.m4.sql >s_validate_logged_in.sql
	echo "\i s_validate_logged_in.sql" >>all.sql

cleanup.sql: cleanup.m4.sql
	m4 -P setup.m4 cleanup.m4.sql >cleanup.sql
	echo "\i cleanup.sql" >>all.sql

s_simulate_email.sql: s_simulate_email.m4.sql
	m4 -P setup.m4 s_simulate_email.m4.sql >s_simulate_email.sql
	echo "\i s_simulate_email.sql" >>all.sql

t_email_id_seq.sql: t_email_id_seq.m4.sql
	m4 -P setup.m4 t_email_id_seq.m4.sql >t_email_id_seq.sql
	echo "\i t_email_id_seq.sql" >>all.sql

s_ip_ban.sql: s_ip_ban.m4.sql
	m4 -P setup.m4 s_ip_ban.m4.sql >s_ip_ban.sql
	echo "\i s_ip_ban.sql" >>all.sql

s_get_config_item.sql: s_get_config_item.m4.sql
	m4 -P setup.m4 s_get_config_item.m4.sql >s_get_config_item.sql
	echo "\i s_get_config_item.sql" >>all.sql

s_get_xsrf_mode.sql: s_get_xsrf_mode.m4.sql
	m4 -P setup.m4 s_get_xsrf_mode.m4.sql >s_get_xsrf_mode.sql
	echo "\i s_get_xsrf_mode.sql" >>all.sql

s_get_customer_id_from_url.sql: s_get_customer_id_from_url.m4.sql
	m4 -P setup.m4 s_get_customer_id_from_url.m4.sql >s_get_customer_id_from_url.sql
	echo "\i s_get_customer_id_from_url.sql" >>all.sql

s_get_username_from_email.sql: s_get_username_from_email.m4.sql
	m4 -P setup.m4 s_get_username_from_email.m4.sql >s_get_username_from_email.sql
	echo "\i s_get_username_from_email.sql" >>all.sql


# Tests in makefile
#		test_start_server			-- start server
#		test_create_accont			-- create an account
#		test_login_t1				-- failed login - not verified
#		test_find_auth_tok			-- find token / log or d.b. lookup
#		test_email_verify			-- do the API to verify
#		test_login2					-- should succeede
#		test_change_password		-- verifie changed in d.b.
#		test_login3					-- fail old pw
#		test_login4					-- success new pw
#		test_logout					-- 
#		test_login5					-- success new pw
#		test_logout					-- 
#		-- test too many failed attemps locks account --
#		-- test account recovers after delta time --
#		-- test send of each of the email templates --
#		-- test SetSession -- Verify in Redis --
#		-- test test_cleanup - remove test account --
#		-- test banned ip --
#		-- test verify stored procedures in d.b. --


# Add a test with multiple logins to a single account

# TODO: Add tests for s_ip_ban.sql s_get_config_item.sql s_get_xsrf_mode.sql s_get_customer_id_from_url.sql 

test: test_config_000 test_status_db01 \
	test_ip_ban test_get_config_item test_get_xsrf_mode test_get_customer_id_from_url \
	test_reg01 test_login01 test_confirm_reg01 test_login02 test_logout01 test_change_pw_web01 test_login04 \
	test_login05 test_validate_tokens test_password_recover test_login06 test_login07 test_login08 \
	test_01_cleanup
	@echo PASS

test_config_000:
	@echo "test_config_000 - validation of configuraiton for a test"
	d9 <test_config.sql >./out/test_config.out
	@grep "success-001" out/test_config.out >/dev/null
	@grep "success-002" out/test_config.out >/dev/null
	@grep "success-003" out/test_config.out >/dev/null
	@grep "success-004" out/test_config.out >/dev/null
	@grep "success-005" out/test_config.out >/dev/null
	@grep "success-006" out/test_config.out >/dev/null
	@grep "success-007" out/test_config.out >/dev/null
	@grep "success-008" out/test_config.out >/dev/null
	@grep "success-009" out/test_config.out >/dev/null
	@grep "success-010" out/test_config.out >/dev/null
	@grep "success-011" out/test_config.out >/dev/null
	@grep "success-012" out/test_config.out >/dev/null
	@grep "success-013" out/test_config.out >/dev/null
	@grep "success-014" out/test_config.out >/dev/null
	@grep "success-015" out/test_config.out >/dev/null
	@grep "success-016" out/test_config.out >/dev/null
	@grep "success-017" out/test_config.out >/dev/null
	@grep "success-018" out/test_config.out >/dev/null
	@grep "success-019" out/test_config.out >/dev/null

# test that you can register for an account
test_reg01:
	@echo "Register a new user"
	d9 <test_reg01.sql >./out/test_reg01.out
	@grep "success" out/test_reg01.out >/dev/null

# Test that you can not login unil you email confirm account
test_login01: 
	@echo "Test Should not be able to login - have not confirmed email"
	d9 <test_login01.sql >./out/test_login01.out
	@grep "error.*code.*[0-9][0-9][0-9]" out/test_login01.out >/dev/null

# lookup auth_token and use it to confirm the registration
test_confirm_reg01: 
	@echo "Conform the email - and validate that we get success."
	d9 <extract_reg_confirm.sql >./out/reg_confirm.data.out
	./bin/build_reg_confirm.sh ./out/reg_confirm.data.out >test_reg_confirm01.sql
	d9 <test_reg_confirm01.sql >./out/test_confirm_reg01.out
	@grep "success" out/test_confirm_reg01.out >/dev/null

# Test that you can login after account is confirmed
test_login02: 
	@echo "Test Should be able to login successfully now that you have confirmed email"
	d9 <test_login02.sql >./out/test_login02.out
	@grep "success" out/test_login02.out >/dev/null

# Logout user - destroy current auth_token (by replacing with new random and not reporing it)
test_logout01: 
	./bin/build_logout.sh ./out/test_login02.out >test_logout01.sql
	d9 <test_logout01.sql >./out/test_logout01.out
	@grep "success" out/test_logout01.out >/dev/null

# just check status of DB is up.
test_status_db01:
	d9 <test_status_db01.sql >./out/test_status_db01.out
	@grep "success" out/test_status_db01.out >/dev/null

#s_change_password.sql: s_change_password.m4.sql t_config.sql
# change a user's password after login
test_change_pw_web01: 
	@echo "Test ability to chagne a password - as if we were logged into the website - using old password"
	d9 <test_login02.sql >./out/test_login02.out
	@grep "success" out/test_login02.out >/dev/null
	./bin/build_test_change_pw_web01.sh ./out/test_login02.out >test_change_pw_web01.sql
	d9 <test_change_pw_web01.sql >./out/test_change_pw_web01.out
	@grep "success" out/test_change_pw_web01.out >/dev/null

# Test that you can not login with old password
test_login04: 
	@echo "Test can not login using old password"
	d9 <test_login01.sql >./out/test_login04.out
	@grep "error.*code.*[0-9][0-9][0-9]" out/test_login04.out >/dev/null

# Test that you can login with new password
test_login05: 
	@echo "Test can login using new password"
	d9 <test_login05.sql >./out/test_login05.out
	@grep "success" out/test_login05.out >/dev/null

# verify that old auth_token is in fact invalid!
test_validate_tokens: 
	./bin/build_validate_token.sh ./out/test_login02.out >test_validate_tokens_pt1.sql
	d9 <./test_validate_tokens_pt1.sql >out/test_validate_tokens_pt1.out
	@grep "error.*code.*920" out/test_validate_tokens_pt1.out >/dev/null
	./bin/build_validate_token.sh ./out/test_login05.out >test_validate_tokens_pt2.sql
	d9 <./test_validate_tokens_pt2.sql >out/test_validate_tokens_pt2.out
	@grep "success"  out/test_validate_tokens_pt2.out >/dev/null

# test recovery of lost password
#s_password_reset.sql: s_password_reset.m4.sql t_config.sql
#s_change_password.sql: s_change_password.m4.sql t_config.sql
# test_password_recover:  test_login05
test_password_recover: 
	d9 <./test_password_recover_pt1.sql >out/test_password_recover_pt1.out
	@grep "success"  out/test_password_recover_pt1.out >/dev/null
	./bin/build_password_recover_pt2.sh ./out/test_password_recover_pt1.out >test_password_recover_pt2.sql
	d9 <./test_password_recover_pt2.sql >out/test_password_recover_pt2.out
	@grep "success"  out/test_password_recover_pt2.out >/dev/null
	./bin/build_password_recover_pt3.sh ./out/test_password_recover_pt1.out >test_password_recover_pt3.sql
	d9 <./test_password_recover_pt3.sql >out/test_password_recover_pt3.out
	@grep "success"  out/test_password_recover_pt3.out >/dev/null

# check can not login with old passwords: 123456
test_login06: 
	d9 <test_login01.sql >./out/test_login06.out
	@grep "error.*code.*[0-9][0-9][0-9]" out/test_login06.out >/dev/null

# check can not login with old passwords: bobob4
test_login07: 
	d9 <test_login05.sql >./out/test_login07.out
	@grep "error.*code.*[0-9][0-9][0-9]" out/test_login07.out >/dev/null

# check can login with new - reset - password: 
test_login08: 
	d9 <test_login08.sql >./out/test_login08.out
	@grep "success" out/test_login08.out >/dev/null

test_get_username_from_email01:
	d9 <test_get_username_from_email.sql >./out/test_get_username_from_email.out
	@grep "success" out/test_get_username_from_email.out >/dev/null
	@grep "success.*username.*test01" out/test_get_username_from_email.out >/dev/null


test_01_cleanup:
	@echo "nothing - removed"

test_02_cleanup:
	d9 <test_01_cleanup.sql >./out/test_01_cleanup.out
	@echo "test01 user removed"




test02: test_reg01 test_confirm_reg01 test_login02a test_login02b

# Test that you can login after account is confirmed
test_login02a: 
	@echo "Test Should be able to login successfully now that you have confirmed email"
	d9 <test_login02.sql >./out/test_login02a.out
	@grep "success" out/test_login02a.out >/dev/null

test_login02b: 
	@echo "Test Should be able to login successfully now that you have confirmed email"
	d9 <test_login02.sql >./out/test_login02b.out
	@grep "success" out/test_login02b.out >/dev/null





test_ip_ban:
	@echo "Test IP Ban function"
	d9 <test_ip_ban.sql >./out/test_ip_ban.out
	@grep "success-100" out/test_ip_ban.out >/dev/null
	@grep "success-101" out/test_ip_ban.out >/dev/null
	
test_get_config_item:
	@echo "Test Getting of config item"
	d9 <test_get_config_item.sql >./out/test_get_config_item.out
	@grep "success-200" out/test_get_config_item.out >/dev/null
	@grep "success-201" out/test_get_config_item.out >/dev/null

test_get_xsrf_mode:
	@echo "Test Getting of config item"
	d9 <test_get_xsrf_mode.sql >./out/test_get_xsrf_mode.out
	@grep "success-300" out/test_get_xsrf_mode.out >/dev/null
	@grep "success-301" out/test_get_xsrf_mode.out >/dev/null


test_get_customer_id_from_url:
	@echo "Test Getting of customer id based on URL"
	d9 <test_get_customer_id_from_url.sql >./out/test_get_customer_id_from_url.out
	@grep "success-400" out/test_get_customer_id_from_url.out >/dev/null






# TODO: Add tests for
#	-- add in email-to-username as a part of auto test.
#	recover email from username















# test substitution and sending of email for each of the templates
#
# ./emalicli/emailcli -t TemplateName -p Params -o Out 
#		generate template substitute for email
# ./emalicli/emailcli -t TemplateName -p Params --sendit
#		contact amazon and send the email (-S)
#	-c cfg.JsonX
#		tmpl ./tmpl directory
#		un/pw auth info for send.
#



# TODO

# Extra tests -----------------------------------------------------------------------------------------------------------------------

# s_stayLoggedIn.sql: s_stayLoggedIn.m4.sql

# verify that when we recover a password this creates a new login session and that the session has a correct token

# check will block if too many bad password logins

# check will recover-block if correct password supplied

# check invalid IP address is blocked.

# ,"/api/session/change_others_password": { "g": "s_change_others_password", "p": [ "username", "$user_id$", "password", "again", "$ip$" ]

# implement group/auth system

# tests of group/auth system

# ------------------------------------------------------------------------------------------------------------------------------------

