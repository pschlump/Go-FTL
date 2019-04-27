#!/bin/bash

# kill $( ps -ef | grep goftl| grep -v grep | grep bb3 | awk '{print $2}' )
xx=$( ps -ef | grep goftl| grep -v grep | grep 'bb3.json' | awk '{print $2}' )
if [ "X$xx" == "X" ] ; then
	:
else
	kill $xx
fi
# ./goftl -c ./bb0.json &
./goftl -c ./bb3.json 2>&1 | tee bb3.out &


