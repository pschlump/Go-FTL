#!/bin/bash

cd
rm -f x1.tar.gz

PSCH=go/src/github.com/pschlump
TOP=go/src/github.com/pschlump/Go-FTL
BLD=go/src/github.com/pschlump/Go-FTL/server/goftl
BL2=go/src/github.com/pschlump/Go-FTL/tools

if [ -d $TOP ] ; then
	:
else
	mkdir -p go/src/github.com/pschlump
	cd go/src/github.com/pschlump
	git clone https://github.com/pschlump/Go-FTL.git
fi

cd
cd $PSCH
for i in * ; do
	cd $i
	git pull
	cd
	cd $PSCH
done

#cd
#cd $TOP
#git pull
#cd ../TabServer2
#git pull
#cd ../mon-alive
#git pull

cd
cd $BLD
go get
go build

cd
cd $BL2/htaccess
go build

cd
cd $BL2/user-pgsql
go build

cd
cd $BL2/user-redis
go build

cd
tar -czf ~/x1.tar.gz \
	-C $BLD ./goftl  \
	-C /home/pschlump/$BL2/htaccess ./htaccess \
	-C /home/pschlump/$BL2/user-pgsql ./user-pgsql \
	-C /home/pschlump/$BL2/user-redis ./user-redis

