#!/bin/bash

auth_token=$( grep 'auth_token":' $1 | awk '{print $3}' | sed -e 's/.*:"//' -e 's/".*//' )

# echo auth_token = $auth_token

cat <<XXxx

select s_logout( '$auth_token', '1.1.1.1');

XXxx


