#!/bin/bash

auth_token=$( grep 'auth_token":' $1 | awk '{print $3}' | sed -e 's/.*:"//' -e 's/".*//' )

# echo auth_token = $auth_token

cat <<XXxx

select s_change_password ( 'bobob4', 'bobob4', '123456', '$auth_token', '1.1.1.1', 'http://auth.2c-why.com' );

XXxx

