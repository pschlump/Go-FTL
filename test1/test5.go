package main

import (
	"net/http"

	"github.com/r7kamura/entoverse"
)

func main() {
	// Please implement your host converter function.
	// This example always delegates HTTP requests to localhost:4000.
	hostConverter := func(originalHost string) string {
		return "localhost:4000"
	}

	// Creates an entoverse.Proxy object as an HTTP handler.
	proxy := entoverse.NewProxy(hostConverter)

	// Runs a reverse-proxy server on http://localhost:3000/
	http.ListenAndServe("localhost:3000", proxy)
}
