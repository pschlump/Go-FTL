#!/bin/bash

# check PLSQL listings for errors
# ignore errors about things that already exists.  Just means that we did create table twice.

if grep -v "ERROR:.*already exists" $1 | grep ERROR: ; then
	exit 1
else
	exit 0
fi

