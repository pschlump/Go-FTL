#!/bin/bash

if [ "$(whoami)" == "root" ] ; then
	:
else
	echo "Usage: !! run as root"
	exit 1
fi

cp go-ftl-init.sh /etc/init.d/go-ftl
cd /etc
ln -s /etc/init.d/go-ftl ./rc0.d/K98go-ftl
ln -s /etc/init.d/go-ftl ./rc1.d/K98go-ftl
ln -s /etc/init.d/go-ftl ./rc2.d/S98go-ftl
ln -s /etc/init.d/go-ftl ./rc3.d/S98go-ftl
ln -s /etc/init.d/go-ftl ./rc4.d/S98go-ftl
ln -s /etc/init.d/go-ftl ./rc5.d/S98go-ftl
ln -s /etc/init.d/go-ftl ./rc6.d/K98go-ftl

