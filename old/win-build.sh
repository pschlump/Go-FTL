#!/bin/bash

export GOPATH=~/go

cd server/goftl
go build
cd ../..

cd tools/htaccess
go build
cd ../..

cd tools/user-pgsql
go build
cd ../..

cd tools/user-redis
go build
cd ../..

tar -czf /C/dragon1share/Win7.tar.gz \
	-C ./server/goftl ./goftl \
	-C ../../tools/htaccess ./htaccess \
	-C ../../tools/user-pgsql ./user-pgsql \
	-C ../../tools/user-redis ./user-redis 

