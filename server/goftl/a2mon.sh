#!/bin/bash

NN=a2-aa-2016-05-14.json
OF=a2-aa-2016-05-14.out

go build 2>&1 | color-cat -c red

xx=$( ps -ef | grep goftl| grep -v grep | grep "$NN" | awk '{print $2}' )
if [ "X$xx" == "X" ] ; then	
	:
else
	kill $xx
fi
./goftl -c ./"$NN" > "$OF" &

