package main

import (
	"fmt"

	// "github.com/garyburd/redigo/redis"
	// Path: /Users/corwin/go/src/www.2c-why.com/gosrp
	"pschlump/gosrp"
	"pschlump/gosrp/big" // "./big" // "math/big"
)

func main() {
	g := big.NewInt(2)
	N, _ := big.NewInt(0).SetString("AC6BDB41324A9A9BF166DE5E1389582FAF72B6651987EE07FC3192943DB56050A37329CBB4A099ED8193E0757767A13DD52312AB4B03310DCD7F48A9DA04FD50E8083969EDB767B0CF6095179A163AB3661A05FBD5FAAAE82918A9962F0B93B855F97993EC975EEAA80D740ADBF4FF747359D041D5C33EA71D281E446B14773BCA97B43A23FB801676BD207A436C6481F1D2B9078717461A5B9D32E688F87748544523B524B0D57D5EA77A2775D2ECFA032CFBDBF52FB3786160279004E57AE6AF874E7303CE53299CCC041C7BC308D82A5698F3A8D0C38271AE35F8E9DBFBB694B5C803D89F7AE435DE236D525F54759B65E372FCD68EF20FA7111F9E4AFF73", 16)

	I := "pschlump@gmail.com"
	p := "abc"
	s, _ := big.NewInt(0).SetString("6b6e1eda7efb668c36ebf95c107300a3", 16)
	ix_s := s.HexString() + I + p
	x_s := gosrp.Hashstring(s.HexString() + I + p)
	x, _ := big.NewInt(0).SetString(x_s, 16)
	v := big.NewInt(0).Exp(g, x, N) // v := pow(g, x, N)

	vRef, _ := big.NewInt(0).SetString("aa4495a557a7a5b047f5bffba993e456ffdc530476554d76641e75179b83dcecafa5b4fa9cd6fbdded13e68c736a0701f3a9765e536d875a6e9c6946d141305ed95ae48579d83a3ab06c79b0be0d276b9d8c39078c8b601608db3bb747b9ec70532ed614af1a5923f0e28ba93579a5e2d057ffb83b8b9b55aa354f8ed9d107fd628e1a746df35ef948815e24d7a4f505eb68f7bd05bef55c6c5ee2cf0c26d1c8be150d4479fa2e4816a74df4f2716e0f24077d3d589104f19a61576fd3d920421eec73bb52549f39cd777147abf727d9b77094aa037ba30851caeb1260186fae83f81b707bb566e4888f6a23c8d3c52de5a8ab2cac6274b5842109235d963299", 16)

	fmt.Printf("s=%s\n", s.HexString())
	fmt.Printf("x=%s\n", x.HexString())
	fmt.Printf("\tix_s [%s]\n", ix_s)
	fmt.Printf("\tx_s [%s]\n", x_s)
	fmt.Printf("v=%s\n", v.HexString())
	if v.HexString() == vRef.HexString() {
		fmt.Printf("PASS\n")
	} else {
		fmt.Printf("FAIL\n")
	}

}
