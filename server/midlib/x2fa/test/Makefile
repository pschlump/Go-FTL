all: users auth_token redis_2fa_data

users:
	@echo
	@echo User List
	@echo
	d199 <list-user_id.sql

auth_token:
	@echo
	@echo Auth Tokens
	@echo
	d199 <autk_keys.sql

redis_2fa_data:
	@echo
	@echo Redis Keys
	@echo
	redis-cli -h 192.168.0.199 <redis_2fa.redis | tee ,r1
	ed ,r1 <ed.1
	redis-cli -h 192.168.0.199 <,r1

