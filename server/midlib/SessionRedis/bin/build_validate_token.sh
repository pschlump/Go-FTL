#!/bin/bash

auth_token=$( grep 'auth_token":' $1 | awk '{print $3}' | sed -e 's/.*:"//' -e 's/".*//' )

# echo auth_token = $auth_token
# select s_login('test01', 'bobob4',  '1.1.1.1',  'http://www.2c-why.com/' );

cat <<XXxx

select s_validate_token( 'bobob4', '$auth_token' );

XXxx


