package main

import (
	"net/http"
	"net/http/httputil"
)

func main() {
	sourceAddress := ":3000"
	director := func(request *http.Request) {
		request.URL.Scheme = "http"
		request.URL.Host = ":9292"
	}
	proxy := &httputil.ReverseProxy{Director: director}
	server := http.Server{
		Addr:    sourceAddress,
		Handler: proxy,
	}
	server.ListenAndServe()
}
