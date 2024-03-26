// main.go
package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	sourceAddress := ":3000"
	destinationUrlString := "http://127.0.0.1:9292"
	destinationUrl, _ := url.Parse(destinationUrlString)
	proxyHandler := httputil.NewSingleHostReverseProxy(destinationUrl)
	server := http.Server{
		Addr:    sourceAddress,
		Handler: proxyHandler,
	}
	server.ListenAndServe()
}
