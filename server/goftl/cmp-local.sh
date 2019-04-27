#!/bin/bash

	export GIT_COMMIT=`git rev-list -1 HEAD` && \
		echo "Version: ${GIT_COMMIT}" && \
		go build -ldflags "-X main.GitCommit=${GIT_COMMIT}" && \
		echo "local:  " ${GIT_COMMIT} `date` >>build-log.txt 
