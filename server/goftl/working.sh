#!/bin/bash

NN=working-2016-03-18.json
OF=working-2016-03-18.out

go build 2>&1 | color-cat -c red
# go build 

xx=$( ps -ef | grep goftl| grep -v grep | grep "$NN" | awk '{print $2}' )
if [ "X$xx" == "X" ] ; then	
	:
else
	kill $xx
fi
./goftl -c ./"$NN" > "$OF" &

