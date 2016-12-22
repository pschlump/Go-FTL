//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1187
//

package main

import "fmt"

func encodeMethod(mt string) (s []byte) {
	if len(mt) < 3 {
		return []byte("z")
	}
	var c byte = (((mt[0] << 1) ^ mt[1] ^ mt[2]) + ' ') & 0x7F
	s = append(s, c)
	// fmt.Printf("C = %x S >%s< >%x<\n", c, s, s)
	return
}

func main() {
	for ii, vv := range []string{"GET", "PUT", "POST", "DELETE", "OPTIONS", "HEAD", "PATCH", "CONNECT"} {
		x := encodeMethod(vv)
		fmt.Printf("%d: %x = %s\n", ii, x, x)
	}
}
