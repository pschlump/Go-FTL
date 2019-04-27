#!/bin/bash

NN=Rewrite01.json
OF=Rewrite01.out

go build

xx=$( ps -ef | grep goftl| grep -v grep | grep "$NN" | awk '{print $2}' )
if [ "X$xx" == "X" ] ; then	
	:
else
	kill $xx
fi
./goftl -c ./"$NN" > "$OF" 2>&1 &

