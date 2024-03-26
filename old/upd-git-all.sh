#!/bin/bash

cd ..
XXX=$(pwd)

for i in * ; do
	if [ -d $i ] ; then
		if [ -d .git ] ; then
			cd $i
			git pull
		fi
	fi
	cd $XXX
done

