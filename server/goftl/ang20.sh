#!/bin/bash

NN=ang20-2016-04-15.json
OF=ang20-2016-04-15.out

go build 2>&1 | color-cat -c red
# go build 

xx=$( ps -ef | grep goftl| grep -v grep | grep "$NN" | awk '{print $2}' )
if [ "X$xx" == "X" ] ; then	
	:
else
	kill $xx
fi
./goftl -c ./"$NN" > "$OF" &

