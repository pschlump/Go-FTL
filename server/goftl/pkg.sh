#!/bin/bash
# package and ship to .206
tar -czf ~/xx.tar.gz \
	./dot206.json \
	./run-206.sh \
	./linux.goftl.tar.gz \
	-C ../../tools/ubuntu-init.d/ \
		go-ftl-init.sh \
		install.sh
scp ~/xx.tar.gz pschlump@198.58.107.206:/home/pschlump
