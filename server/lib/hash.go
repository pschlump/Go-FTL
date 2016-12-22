package lib

import (
	"crypto/sha256"
	"fmt"
)

// import "github.com/pschlump/gosrp" // Path: /Users/corwin/go/src/www.2c-why.com/gosrp

// ----------------------------------------------------------------------------------------------------------------------------
// func Sha256(s string) (rv string) {
// 	rv = gosrp.Hashstring(s)
// 	return
// }

func Sha256(s string) (rv string) {
	rv = HashStrings(s)
	return
}

func HashStrings(a ...string) string {
	h := sha256.New()
	for _, z := range a {
		h.Write([]byte(z))
	}
	return fmt.Sprintf("%x", (h.Sum(nil)))
}

/* vim: set noai ts=4 sw=4: */
