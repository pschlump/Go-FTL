#!/bin/bash

XXX=$(pwd)

for i in $( cat ,c ) ; do
	cd $i
	MD=$( ls *.md )
	cat >>Makefile <<XXy

mkdocs:
	cat $MD  >>../tmp2
	echo "$i/$MD" >>../m.list

XXy
	cd $XXX
done

