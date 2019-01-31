all: users auth_token redis_2fa_data t2fa

users:
	@echo
	@echo User List
	@echo
	d199 <list-user_id.sql
	@echo

auth_token:
	@echo
	@echo Auth Tokens
	@echo
	d199 <autk_keys.sql
	@echo

t2fa:
	@echo
	@echo t_2fa auth ids
	@echo
	d199 <t2fa.sql
	@echo

redis_2fa_data:
	@echo
	@echo Redis Keys
	@echo
	redis-cli -h 192.168.0.199 <redis_2fa.redis | tee ,r1
	@echo
	ed ,r1 <ed.1 >,ed.1 
	redis-cli -h 192.168.0.199 <,r1
	@echo

