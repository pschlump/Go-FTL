package goftlmux

//
// Go Go Mux - Go Fast Mux / Router for HTTP requests
//
// (C) Philip Schlump, 2013-2014.
// Version: 0.4.3
// BuildNo: 804
//
//

import (
	"fmt"
	"testing"

	debug "github.com/pschlump/godebug"
)

func Test_CmpUrlToCleanRoute(t *testing.T) {

	r, trr := setupHtx()
	//	r := htx
	//	var procData MuxRouterProcessing
	//	InitMuxRouterProcessing(r, &procData)

	url := "/abc/:def/ghi"
	r.SplitOnSlash3(trr, 0, url, false)

	if false {
		fmt.Printf("r.Slash=%s NSl=%d %s\n", debug.SVar(trr.Slash[0:trr.NSl+1]), trr.NSl, debug.LF())
	}

	b := r.CmpUrlToCleanRoute(trr, "T:T", "/abc/:/ghi")
	if false {
		fmt.Printf("b=%v, %s\n", b, debug.LF())
	}
	if !b {
		t.Errorf("Not Found\n")
	}

}

// 36 ns
func OldBenchmark_CmpUrlToCleanRoute(b *testing.B) {
	//r := htx
	//var procData MuxRouterProcessing
	//InitMuxRouterProcessing(r, &procData)
	r, trr := setupHtx()

	url := "/abc/:def/ghi"
	r.SplitOnSlash3(trr, 0, url, false)

	for n := 0; n < b.N; n++ {
		b := r.CmpUrlToCleanRoute(trr, "T:T", "/abc/:/ghi")
		_ = b
	}
}
