#!/bin/bash

HEAD=/usr/bin/head
BS="\`"

for i in $( cat m.list ) ; do
	L1=$( $HEAD -n 1 "$i" )
	N=$( echo "$L1" | sed -e 's/:.*//' )
	echo "$BS$N$BS | $L1"
done

# `BasicAuth` | BasicAuth: Implement basic authentication usings a .htaccess file
