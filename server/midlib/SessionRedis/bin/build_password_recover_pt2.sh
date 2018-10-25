#!/bin/bash

# Called from Makeifle

auth_token=$( grep 'email_token":' $1 | sed -e 's/.*"email_token":"//' -e 's/".*//' )

# echo auth_token = $auth_token
# CREATE or REPLACE FUNCTION s_password_reset_pt2 (  p_token varchar, p_ip_addr varchar, p_url varchar )

cat <<XXxx

select s_password_reset_pt2 ( '$auth_token', '1.1.1.1', 'http://auth.2c-why.com' );

XXxx


