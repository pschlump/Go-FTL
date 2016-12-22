#!/bin/bash

# make it so that the executable can bind to port 80 on linux - as a non-root process.

setcap 'cap_net_bind_service=+ep' go-ftl

