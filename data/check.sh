#!/bin/bash

# check for FAIL in file, if found then return FALSE for Make

if grep FAIL $1 >/dev/null ; then
	echo FAIL
	exit 1
else
	exit 0
fi

