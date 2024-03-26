#!/bin/bash

if grep '^ERROR:' $1 ; then
	echo "FAIL -- error in running script"
	exit 1
else
	exit 0
fi

