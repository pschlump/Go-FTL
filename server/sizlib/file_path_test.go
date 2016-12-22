package sizlib

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

// func SearchPath(rawFileName string, searchPath string) (fullFileName string, ok bool) {
func Test_SearchPath(t *testing.T) {

	debug := false
	user := os.Getenv("USER")
	home := os.Getenv("HOME")

	os.Mkdir("./test", 0700)
	ioutil.WriteFile("./test/xx", []byte("xx"), 0600)

	fn, ok := SearchPath("dir.go", "~:./test:.")
	if debug {
		fmt.Printf("test 1 : fn=[%s]\n", fn)
	}
	if !ok {
		t.Errorf("Failed 1: dir.go not found in current directory\n")
	}
	if fn != "./dir.go" {
		t.Errorf("Failed 2: ./dir.go not found in current directory, found %s\n", fn)
	}

	fn, ok = SearchPath("xx.exe", "~:./test:.")
	if debug {
		fmt.Printf("test 2 : fn=[%s]\n", fn)
	}
	if ok {
		t.Errorf("Failed 3: xx.exe found in current directory\n")
	}
	if fn != "xx.exe" {
		t.Errorf("Failed 4: xx.exe incorrect file name not,  found %s\n", fn)
	}

	fn, ok = SearchPath("xx", "~:./test:.")
	if debug {
		fmt.Printf("test 3: fn=[%s]\n", fn)
	}
	if !ok {
		t.Errorf("Failed 5: xx incorrect file name not,  found %s\n", fn)
	}
	if fn != "./test/xx" {
		t.Errorf("Failed 6: xx incorrect file name not,  found %s\n", fn)
	}

	fn, ok = SearchPath("yy.zed", "~:./test:.")
	if debug {
		fmt.Printf("test 12 : fn=[%s]\n", fn)
	}
	if !ok {
		fmt.Printf("Failed 14: yy.zed\n")
		t.Fail()
	}
	if fn != "./test/yy-pschlump-dev2.zed" {
		fmt.Printf("Failed 15: yy.zed\n")
		t.Fail()
	}

	if debug {
		fmt.Printf("Test 5 - test user substitution\n")
	}
	mdata := make(map[string]string, 10)
	rs, has := SubstitueUserInFilePath("~"+user+"/cfg", mdata)

	if debug {
		fmt.Printf("rs=[%s] has=%v mdata=%s\n", rs, has, SVar(mdata))
	}

	if rs != "%{USER_"+user+"%}/cfg" {
		fmt.Printf("Failed 16: yy.zed\n")
		t.Fail()
	}
	if !has {
		fmt.Printf("Failed 17: yy.zed\n")
		t.Fail()
	}
	if _, ok := mdata["USER_"+user+""]; !ok {
		fmt.Printf("Failed 5: yy.zed\n")
		t.Fail()
	}

	fn, ok = SearchPath("~"+user+"/Desktop", "~:./test:.")
	if debug {
		fmt.Printf("test 6 : fn=[%s]\n", fn)
	}
	if !ok {
		fmt.Printf("Failed 18: ~" + user + "/Desktop not found\n")
		t.Fail()
	}
	if fn != home+"/Desktop" {
		fmt.Printf("Failed 19: Incorrect file name, got %s expecing %s/Desktop\n", fn, home)
		t.Fail()
	}

}
