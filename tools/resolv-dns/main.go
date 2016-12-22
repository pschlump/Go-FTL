package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/pschlump/Go-FTL/server/nameresolve"
	"github.com/pschlump/godebug"
)

func main() {

	flag.Parse()
	fns := flag.Args()

	for _, host := range fns {
		resolvOneHost(host)
	}
}

func resolvOneHost(host string) {

	// s := nameresolve.NoWild("*.test1.com")
	// host := "*.test1.com"

	IPs, err := net.LookupIP(nameresolve.NoWild(host))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Host: %s IPs: %s\n", host, godebug.SVarI(IPs))

}
