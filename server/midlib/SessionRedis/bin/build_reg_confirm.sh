#!/bin/bash

# Called from Makeifle
#	./bin/build_reg_confirm.sh ./out/reg_confirm.data.out >test_reg_confirm01.sql

auth_token=$( grep 'EMAIL_RESET_KEY' $1 | awk '{print $3}' )

# echo auth_token = $auth_token

cat <<XXxx

select s_confirm_email ( '$auth_token', '1.1.1.1', 'http://www.2c-why.com', 'GET'  );

XXxx


