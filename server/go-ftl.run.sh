#!/bin/bash

cd "$2"

if [ "$1" == "start" ] ; then
	./go-ftl >,log_$3 2>&1
fi

