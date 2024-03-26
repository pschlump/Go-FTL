#!/bin/bash

NN=dot206.json
OF=,log

#rm -f ./go-ftl
#make 206

#go build 2>&1 | color-cat -c red
# go build 

xx=$( ps -ef | grep go-ftl| grep -v grep | grep "$NN" | awk '{print $2}' )
if [ "X$xx" == "X" ] ; then	
	:
else
	kill $xx
fi
if [ -x ./go-ftl ] ; then
	./go-ftl -c ./"$NN" > "$OF" &
fi

