// From: http://r7kamura.github.io/2014/07/20/golang-reverse-proxy.html -- Websitein Japaneese

package main

import (
	"container/ring"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

func main() {
	sourceAddress := ":3000"

	ports := []string{
		":9293",
		":9292",
	}
	hostRing := ring.New(len(ports))
	for _, port := range ports {
		url, _ := url.Parse("http://127.0.0.1" + port)
		hostRing.Value = url
		hostRing = hostRing.Next()
	}

	mutex := sync.Mutex{}
	director := func(request *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		request.URL.Scheme = "http"
		request.URL.Host = hostRing.Value.(*url.URL).Host
		hostRing = hostRing.Next()
	}
	proxy := &httputil.ReverseProxy{Director: director}
	server := http.Server{
		Addr:    sourceAddress,
		Handler: proxy,
	}
	server.ListenAndServe()
}
