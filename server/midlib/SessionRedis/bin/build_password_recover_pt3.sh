#!/bin/bash

# Called from Makeifle

auth_token=$( grep 'email_token":' $1 | sed -e 's/.*"email_token":"//' -e 's/".*//' )

# echo auth_token = $auth_token
# select s_password_reset_pt3 ( 'yanky444', 'yanky444', '8b0fd363-4773-4ef3-9ade-929359285d3c', '1.1.1.1', 'http://auth.2c-why.com' );

cat <<XXxx

select s_password_reset_pt3 ( 'yanky444', 'yanky444', '$auth_token', '1.1.1.1', 'http://auth.2c-why.com' );

XXxx


