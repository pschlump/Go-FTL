#!/bin/bash

if [ "z$1" = "z" ] ; then
	echo "$0: Usage: $0 Username"
	exit 1
fi

d9 >/tmp/,a <<XXxx
select 'EMAIL_'||'RESET_KEY', "email_reset_key"
from "t_user"
where "username" = '$1'
;
XXxx

auth_token=$( grep 'EMAIL_RESET_KEY' /tmp/,a | awk '{print $3}' )

#echo auth_token = $auth_token

echo $auth_token

#cat <<XXxx
#
#select s_confirm_email ( '$auth_token', '1.1.1.1' );
#
#XXxx

