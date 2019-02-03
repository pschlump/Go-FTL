package main

import (
	"fmt"
	"strconv"

	"github.com/pschlump/HashStrings"
)

func main() {
	// (Before sha256 String) ss= 9bf871f1144ef3ae16f72cb8c3dc9f3f30e00b5df49cf0ef774bf2bd1c2d6405:0c86a489bc238d8bda7849d78b0d62fa:c63be5867428fb28c05ae899f1358fd489617dfca6519f359bfb5b51fb6853d1
	ss := "9bf871f1144ef3ae16f72cb8c3dc9f3f30e00b5df49cf0ef774bf2bd1c2d6405:0c86a489bc238d8bda7849d78b0d62fa:c63be5867428fb28c05ae899f1358fd489617dfca6519f359bfb5b51fb6853d1"

	// (After sha256 Value - the Hash) val0= e1d9ff9112abf7208df37d669673c19e98c9d1ab9d6d69b2b2fdde2ea1ab05a8
	expectedVal0 := "e1d9ff9112abf7208df37d669673c19e98c9d1ab9d6d69b2b2fdde2ea1ab05a8"

	// val0 := HashStrings.Sha256(fmt.Sprintf("%s:%s:%s", user_hash, fp, current2MinHash))
	val0 := HashStrings.Sha256(ss)
	if val0 == expectedVal0 {
		fmt.Printf("PASS\n")
	} else {
		fmt.Printf("FAIL, expected [%s] got [%s]\n", expectedVal0, val0)
	}

	// val1 := fmt.Sprintf("%x", val0)
	val1 := string(val0)
	val2 := val1[len(val1)-6:]
	val, err := strconv.ParseInt(val2, 16, 64)
	if err != nil {
		fmt.Printf("Error on d.b. query %s\n", err)
	}
	val = val % 1000000

	expectedN := 208104
	if expectedN == int(val) {
		fmt.Printf("PASS - test2\n")
	} else {
		fmt.Printf("FAIL, expected [%d] got [%d]\n", expectedN, val)
	}
}
