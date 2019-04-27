#!/bin/bash

NN=doc-ang20-2016-05-19.json
OF=doc-ang20-2016-05-19.out

rm -f ./goftl

go build 2>&1 | color-cat -c red
# go build 

xx=$( ps -ef | grep goftl| grep -v grep | grep "$NN" | awk '{print $2}' )
if [ "X$xx" == "X" ] ; then	
	:
else
	kill $xx
fi
if [ -x ./goftl ] ; then
	./goftl -c ./"$NN" > "$OF" &
fi

