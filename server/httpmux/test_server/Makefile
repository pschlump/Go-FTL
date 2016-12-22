
all: samp-server reg-server

samp-server: samp-server.go ../../httpmux/server.go
	go build samp-server.go

reg-server: reg-server.go
	go build reg-server.go

