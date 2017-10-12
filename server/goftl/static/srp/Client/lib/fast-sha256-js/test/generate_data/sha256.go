package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/pschlump/json" //	"encoding/json"
)

func main() {
	var rs [1070][2][]byte
	for i := range rs {
		m := make([]byte, i)
		if _, err := io.ReadFull(rand.Reader, m[:]); err != nil {
			panic(err)
		}
		h := sha256.Sum256(m)
		rs[i][0] = m
		rs[i][1] = h[:]
	}
	out, err := json.MarshalIndent(rs, "", "")
	if err != nil {
		panic(err)
	}
	fmt.Print("module.exports = ")
	fmt.Print(string(out))
	fmt.Println(";")
}
