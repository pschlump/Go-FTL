package main

import (
	"fmt"
	"net"

	"github.com/pschlump/godebug"
)

func main() {
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		if false {
			fmt.Printf("Addrs: %s\n", godebug.SVarI(addrs))
		}
		for _, addr := range addrs {
			fmt.Printf("Addr: %s\n", addr)
		}
	} else {
		fmt.Printf("Error: %s\n", err)
	}
}
