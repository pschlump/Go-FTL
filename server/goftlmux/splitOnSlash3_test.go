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
)

var testRuns_SplitOnSlash3 = []struct {
	param string
	slash string
	hash  string
	nsl   int
}{
	{"", `[]`, ``, 1},
	{"/", `[]`, ``, 1},
	{"a", `["a"]`, ``, 1},
	{"aa", `["aa"]`, ``, 1},
	{"/a", `["a"]`, ``, 1},
	{"/aa", `["aa"]`, ``, 1},
	{"//a", `["a"]`, ``, 1},
	{"///a", `["a"]`, ``, 1},
	{"///a/", `["a"]`, ``, 1},
	{"///a//", `["a"]`, ``, 1},
	{"///a///", `["a"]`, ``, 1},
	{"aa", `["aa"]`, ``, 1},
	{"/aa", `["aa"]`, ``, 1},
	{"//aa", `["aa"]`, ``, 1},
	{"///aa", `["aa"]`, ``, 1},
	{"///aa/", `["aa"]`, ``, 1},
	{"///aa//", `["aa"]`, ``, 1},
	{"./aa", `["aa"]`, ``, 1},
	{"././aa", `["aa"]`, ``, 1},
	{"./././aa", `["aa"]`, ``, 1},
	{"/./aa", `["aa"]`, ``, 1},
	{"/././aa", `["aa"]`, ``, 1},
	{"/./././aa", `["aa"]`, ``, 1},
	{"/aa/bb", `["aa","bb"]`, ``, 2},
	{"/aa//bb/cc/dd", `["aa","bb","cc","dd"]`, ``, 4},
	{"/aa/bb///cc/dd", `["aa","bb","cc","dd"]`, ``, 4},
	{"/aa/bb/./cc//.//dd", `["aa","bb","cc","dd"]`, ``, 4},
	{"/aa/bb.html", `["aa","bb.html"]`, ``, 2},
	{"/aa//bb/cc/dd.php", `["aa","bb","cc","dd.php"]`, ``, 4},
	{"/aa//bb/cc/dd.php/", `["aa","bb","cc","dd.php"]`, ``, 4},
	{"/aa//bb/cc/dd.php//", `["aa","bb","cc","dd.php"]`, ``, 4},
	{"/aa//bb/cc/dd.php///", `["aa","bb","cc","dd.php"]`, ``, 4},
	{"/aa/bb///cc.php/dd", `["aa","bb","cc.php","dd"]`, ``, 4},
	{"/aa/bb/./...cc//.//dd", `["aa","bb","...cc","dd"]`, ``, 4},
	{"/aa/bb/./.cc//.//dd", `["aa","bb",".cc","dd"]`, ``, 4},
	{"/../a", `["a"]`, ``, 1},
	{"/../../a", `["a"]`, ``, 1},
	{"/../../../a", `["a"]`, ``, 1},
	{"/../../../../a", `["a"]`, ``, 1},
	{"../a", `["a"]`, ``, 1},
	{"../../a", `["a"]`, ``, 1},
	{"../../../a", `["a"]`, ``, 1},
	{"../../../../a", `["a"]`, ``, 1},
	{"../../a.html", `["a.html"]`, ``, 1},
	{"../../../a.html", `["a.html"]`, ``, 1},
	{"../../../../a.html", `["a.html"]`, ``, 1},
	{"../bb/cc/../../a.html", `["a.html"]`, ``, 1},
	{"../bb/cc/dd/../../a.html", `["bb","a.html"]`, ``, 2},
	{"./bb/cc/dd/../../a.html", `["bb","a.html"]`, ``, 2},
	{"bb/cc/dd/../../ee/a.html", `["bb","ee","a.html"]`, ``, 3},
	{"bb/cc/dd/../../ee/../a.html", `["bb","a.html"]`, ``, 2},
	{"bb/cc/dd/../../ee/../a.html/", `["bb","a.html"]`, ``, 2},
	{"bb/cc/dd/../../ee/../a.html//", `["bb","a.html"]`, ``, 2},
	{"/./../bb/cc/dd/../../ee/../a.html//", `["bb","a.html"]`, ``, 2},
	{"/./../.../cc/dd/../../ee/../a.html//", `["...","a.html"]`, ``, 2},
	{"/redis/planb/", `["redis","planb"]`, ``, 2},
}

func TestSplitOnSlash3_01(t *testing.T) {

	r, trr := setupHtx()

	for k, test := range testRuns_SplitOnSlash3 {
		r.SplitOnSlash3(trr, 0, test.param, false)
		s := arrFrom(trr.Slash[:], trr.NSl, trr.CurUrl, trr.Hash[:])
		// r.Hash[NSl-1] = h
		// r.Slash[NSl] = ln
		// r.NSl = NSl
		// fmt.Printf("s=%s\n", s)
		if s != test.slash {
			t.Errorf("Test %d - Url(%v) = ", k, test.param)
		}
		if trr.NSl != test.nsl {
			t.Errorf("Test %d - Url(%v) NSl = %d, expected %d ", k, test.param, trr.NSl, test.nsl)
		}
	}
}

func arrFrom(Slash []int, NSl int, url string, Hash []int) string {
	s := "["
	com := ""
	// fmt.Printf("NSl = %d\n", NSl)
	// fmt.Printf("url ->%s<- Hash=%s Slash=%s NSl=%d\n", url, debug.SVar(Hash[0:NSl]), debug.SVar(Slash[0:NSl+1]), NSl)
	for i := 0; i < NSl; i++ {
		if Slash[i]+1 > Slash[i+1] {
			s += com + "?"
			com = ","
		} else if Slash[i]+1 < len(url) && Slash[i+1]-1 < len(url) {
			s += com + fmt.Sprintf("%q", url[Slash[i]+1:Slash[i+1]])
			com = ","
		}
	}
	s += "]"
	return s
}

/*
// 52.3 us
func BenchmarkFixPath(b *testing.B) {
	// noalloc = true
	rv := make([]string, 25)
	for n := 0; n < b.N; n++ {
		rv = rv[:25]
		// FixPath("/./../.../cc/dd/../../ee/../a.html//", &rv)
		FixPath("/cc/dd/a.html", rv, 25)
	}
}
*/
