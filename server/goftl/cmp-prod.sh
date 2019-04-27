#!/bin/bash

	export GIT_COMMIT=`git rev-list -1 HEAD` && \
		echo "Version: ${GIT_COMMIT}" && \
		GOOS=linux go build -ldflags "-X main.GitCommit=${GIT_COMMIT}" -o $1 . && \
		echo "deploy: " ${GIT_COMMIT} `date` >>build-log.txt 
