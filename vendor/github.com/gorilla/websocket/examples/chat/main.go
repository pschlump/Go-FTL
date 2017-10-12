// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"
)

// PJS - mod addr
// var addr = flag.String("addr", ":8080", "http service address")
var addr = flag.String("addr", ":9876", "http service address")
var dir = flag.String("dir", "./static", "static file server ")

var debug map[string]bool

func init() {
	debug = make(map[string]bool)
	debug["echo-msg"] = true
}

func main() {
	flag.Parse()

	// Run the single HUB for the chat
	hub := newHub()
	go hub.run()

	// Listen for /ws and run the websocket server on it.
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	// If is is not /ws, then assume that it is a file and serve files.
	http.Handle("/", http.FileServer(http.Dir(*dir)))

	// Crank it up.  Star listening.
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
