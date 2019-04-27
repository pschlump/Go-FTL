#!/bin/bash

xx=$( ps -ef | grep goftl| grep -v grep | grep 'bb5.json' | awk '{print $2}' )
if [ "X$xx" == "X" ] ; then
	:
else
	kill $xx
fi
./goftl -c ./bb5.json 2>&1 | tee bb5.out &

