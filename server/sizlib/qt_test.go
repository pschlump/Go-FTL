package sizlib

import (
	"testing"
)

//func init () {
//}

func Test_Qt(t *testing.T) {
	var s string
	mdata := make(map[string]string, 40)
	mdata["aaa"] = "AAA"
	mdata["ccc"] = "CCC"

	s = Qt("abc", mdata)
	if s != "abc" {
		t.Fail()
	}

	s = Qt("abc%{aaa%}abc%{bbb%}abc%{ccc%}abc", mdata)
	if s != "abcAAAabcabcCCCabc" {
		t.Fail()
	}
}
