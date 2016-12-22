//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1013
//

package fileserve

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/godebug"
)

// -----------------------------------------------------------------------------------------------------------------------------------------------
// should be moved to ../lib
func Test_FileServe_00(t *testing.T) {

	var s string

	s = RmExt("abc.md")
	if s != "abc" {
		t.Errorf("Error 0001, Expected >abc< got >%s<\n", s)
	}

	s = RmExt("abc")
	if s != "abc" {
		t.Errorf("Error 0002, Expected >abc< got >%s<\n", s)
	}

	s = RmExt("")
	if s != "" {
		t.Errorf("Error 0003, Expected >< got >%s<\n", s)
	}

	s = RmExt("abc.min.js")
	if s != "abc.min" {
		t.Errorf("Error 0004, Expected >abc.min< got >%s<\n", s)
	}

	s = RmExtSpecified("abc.min.js", ".js")
	if s != "abc.min" {
		t.Errorf("Error 0005, Expected >abc.min< got >%s<\n", s)
	}

	s = RmExtSpecified("abc.min.js", ".min.js")
	if s != "abc" {
		t.Errorf("Error 0006, Expected >abc< got >%s<\n", s)
	}

	s = RmExtSpecified("abc.min.js", "abc.min.js")
	if s != "" {
		t.Errorf("Error 0007, Expected >< got >%s<\n", s)
	}

	s = RmExtSpecified("abc.min.js", ".html")
	if s != "abc.min.js" {
		t.Errorf("Error 0008, Expected >abc.min.js< got >%s<\n", s)
	}
}

// -----------------------------------------------------------------------------------------------------------------------------------------------
// should be moved to ../lib
//func CompareModTime(in, out time.Time) bool {
// https://golang.org/src/time/sleep_test.go
func Test_FileServe_01(t *testing.T) {

	var b, shouldRebuild RebuildFlag

	t1 := time.Now()
	t2 := time.Now()

	b = CompareModTime(t1, t2)
	if shouldRebuild == NeedRebuild {
		t.Errorf("Error 0101, Expected >true< got >%v<\n", b)
	}

	lib.SetupTestCreateDirsFileServe()

	ok1, inFi := lib.ExistsGetFileInfo("./test/old.txt")
	ok2, outFi := lib.ExistsGetFileInfo("./test/new.txt")
	if !ok1 || !ok2 {
		t.Errorf("Error 0102, Test file missing\n")
	}

	shouldRebuild = CompareModTime(inFi.ModTime(), outFi.ModTime())
	if shouldRebuild == NeedRebuild {
		t.Errorf("Error 0103, Expected >NeedRebuild< got >%s<\n", shouldRebuild)
	}

	shouldRebuild = CompareModTime(outFi.ModTime(), inFi.ModTime())
	if shouldRebuild != NeedRebuild {
		t.Errorf("Error 0104, Expected >NOT NeedRebuild< got >%s<\n", shouldRebuild)
	}

}

// func runCmdIfNecessary(
// 	fcfg *FileServerType, www http.ResponseWriter, req *http.Request,
// 	inputFn string, haveInput bool, inFi os.FileInfo, inExt string,
// 	outputFn string, haveOutput bool, outFi os.FileInfo, outExt string,
// 	ti int, tr *ExtProcessType) {
// 		Create a ./test directory
// 		Create 2 files ./test/old.test and ./test/new.test

func Test_FileServe_02(t *testing.T) {

	lib.SetupTestCreateDirsFileServe()

	cmd := `{ "Cmd":"cp", "Params":[ "{{.inputFile}}", "{{.outputFile}}" ] }`
	ok, out, err := ExecuteCommands(cmd, "./test/rb.in", "./test/rb.out", ".in", ".out")
	if db1 {
		fmt.Printf("ok=%v out=%v err=%v, %s\n", ok, out, err, godebug.LF())
	}
	// ioutil.WriteFile("./test/rb.in", []byte(`rb.in`), 0644)
	// ioutil.WriteFile("./test/rb.out", []byte(`# Error - if this is found - output should be overwritten #`), 0644)

	data, err1 := ioutil.ReadFile("./test/rb.out")
	if err1 != nil {
		t.Errorf("Error 0201, File is missing after copy operaiton\n")
	}

	if string(data) != `rb.in` {
		t.Errorf("Error 0202, Wrong contents for ./test/rb.out, got >%s<, expected >rb.in<\n", data)
	}

}

/* vim: set noai ts=4 sw=4: */
