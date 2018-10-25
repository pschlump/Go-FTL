
# Copyright (C) Philip Schlump, 2014-2016.  All rights reserved.

all: upload-table.sql
	go build

upload-table.sql: upload-table.m4.sql common.m4.sql
	m4 -P upload-table.m4.sql >upload-table.sql

