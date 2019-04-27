#!/bin/bash

BUILD_NO=$(git rev-list --count HEAD)

for i in $* ; do

	ed $i <<XXxx
/BuildNo:/s/: [0-9][0-9][0-9]*/: 0$BUILD_NO/
w
/^var BuildNo =/s/ = "[0-9][0-9][0-9]*"/ = "0$BUILD_NO"/
w
q
XXxx

done




exit 0




date >,go
md5 *.go >>,go
echo "" >>,go

ed version.go <<XXxx
/^BuildNo:/+2,/<\/pre>/-1d
w
/^BuildNo:/+1r ,go
w
q
XXxx

