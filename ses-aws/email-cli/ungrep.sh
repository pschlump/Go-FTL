#!/bin/bash

if grep $1 $2 >/dev/null 2>&1 ; then
	exit 1
else
	exit 0
fi

