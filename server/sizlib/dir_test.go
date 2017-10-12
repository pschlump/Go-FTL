package sizlib

import (
	"testing"

	"github.com/pschlump/HashStr"
)

//func init() {
//}

func Test_Exists(t *testing.T) {
	b := Exists("dir.go")
	if !b {
		t.Fail()
	}
	b = Exists("no-dir.go")
	if b {
		t.Fail()
	}
}

func Test_RmExt(t *testing.T) {
	if RmExt("/a/b/c.jpg") != "/a/b/c" {
		t.Fail()
	}
	if RmExt("/a/b/c") != "/a/b/c" {
		t.Fail()
	}
	if RmExt("abc.jpg") != "abc" {
		t.Fail()
	}
	if RmExt("abc") != "abc" {
		t.Fail()
	}
}

// func FnToCssClass ( filename string ) string {
//func Test_FnToCssClass(t *testing.T) {
//	if FnToCssClass("goo boo loo * too") != "goo-boo-loo---too" {
//		t.Fail()
//	}
//	if FnToCssClass("abc-def-ghi") != "abc-def-ghi" {
//		t.Fail()
//	}
//}

func Test_EscapeDoubleQuote(t *testing.T) {
	if EscapeDoubleQuote(`abc"def`) != "abc\\\"def" {
		t.Fail()
	}
	if EscapeDoubleQuote(`abcdef`) != "abcdef" {
		t.Fail()
	}
	if EscapeDoubleQuote(`abc'def`) != "abc'def" {
		t.Fail()
	}
}

//func HashStr(s []byte) (n int) {
//func HashStrToName(s string) (s string) {
func Test_HashFunctions(t *testing.T) {
	// n := HashStr.HashStrToName("abc")
	// fmt.Printf("n=%d\n", n)
	s := HashStr.HashStrToName("select 12 from dual")
	// fmt.Printf("s=>%s< %x\n", s, s)
	w := HashStr.HashStrToName("binky dinky")
	if s == w {
		t.Errorf("Hashes matched")
	}
}

func Test_SVar(t *testing.T) {
	s := SVar(map[string]string{"abc": "def"})
	if s != `{"abc":"def"}` {
		t.Errorf("SVar did not return correct JSON, got [%s]", s)
	}
}

// func JsonStringToData(s string) (theJSON map[string]interface{}, err error) {
//func Test_JsonStringToData(t *testing.T) {
//	a, err := JsonStringToData(`{"abc":"def"}`)
//	if err != nil {
//		t.Errorf("JsonStringToData: failed to parse valid JOSN")
//	}
//	v, ok := a["abc"]
//	if !ok {
//		t.Errorf("JsonStringToData: failed to parse valid JOSN")
//	}
//	vs, ok := v.(string)
//	if !ok {
//		t.Errorf("JsonStringToData: failed to parse valid JOSN")
//	}
//	if vs != "def" {
//		t.Errorf("JsonStringToData: failed to parse valid JOSN")
//	}
//}

//func ReadJSONDataWithComments(path string) (file []byte, err error) {
func Test_ReadJSONDataWithComments(t *testing.T) {
	file, err := ReadJSONDataWithComments("./test/j_com.json")
	if err != nil {
		t.Errorf("ReadJSONDataWithCommetns 1:")
	}
	// fmt.Printf("--->>>%s<<<---\n", file)
	exp := `{
 "LineNo":2

, "FileName":./test/j_com.json
}
`
	if string(file) != exp {
		t.Errorf("ReadJSONDataWithCommetns 2:")
	}
}

//func GetIpFromRemoteAddr(RemoteAddr string) (rv string) {
//func Test_GetIpFromRemoteAddr(t *testing.T) {
//	ip := GetIpFromRemoteAddr("192.168.0.33:1111")
//	if ip != "192.168.0.33" {
//		t.Errorf("GetIpFromRemoteAddr")
//	}
//}
