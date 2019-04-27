#!/bin/bash

# Example 
#
# $ ./run-test test001

mkdir -p testdata output

if [ "$1" == "test002" ] ; then	
	rm output
	ln -s /Users/pschlump/Output ./output
fi

NN=testdata/$1.jsonx
OF=output/$1.out

if [ -f $NN ] ; then
	:
else
	echo "Missing config file $NN"
	exit 1
fi

rm -f ./goftl

#echo AA $NN

go build 2>&1 | color-cat -c red
if [ -f goftl ] ; then
	:
else
	exit 1
fi

#echo BB $NN

xx=$( ps -ef | grep goftl | grep -v grep | grep "$NN" | awk '{print $2}' )
if [ "X$xx" == "X" ] ; then	
	:
else
	kill $xx
fi

#echo CC $NN

if [ -x ./goftl ] ; then
	./goftl -c ./"$NN" --note="$(pwd)/run-test.sh $1" > "$OF" &
fi

sleep 2
make verify-$1
make db-check-$1
