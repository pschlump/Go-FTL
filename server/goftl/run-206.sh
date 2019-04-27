#!/bin/bash

NN=loc206.jsonx
OF=output/loc206.out

rm -f ./goftl
make 206

go build 2>&1 | color-cat -c red
# go build 

xx=$( ps -ef | grep goftl| grep -v grep | grep "$NN" | awk '{print $2}' )
if [ "X$xx" == "X" ] ; then	
	:
else
	kill $xx
fi
if [ -x ./goftl ] ; then
	./goftl -c ./"$NN" --note="$(pwd)/run-206.sh" > "$OF" &
fi

