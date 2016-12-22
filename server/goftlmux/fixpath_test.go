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
	"net/http"
	"testing"

	debug "github.com/pschlump/godebug"
)

var testRuns = []struct {
	param  string
	result string
}{
	{"", `["/"]`},
	{"/", `["/"]`},
	{"a", `["a"]`},
	{"aa", `["aa"]`},
	{"/a", `["/","a"]`},
	{"/aa", `["/","aa"]`},
	{"//a", `["/","a"]`},
	{"///a", `["/","a"]`},
	{"///a/", `["/","a"]`},
	{"///a//", `["/","a"]`},
	{"///a///", `["/","a"]`},
	{"aa", `["aa"]`},
	{"/aa", `["/","aa"]`},
	{"//aa", `["/","aa"]`},
	{"///aa", `["/","aa"]`},
	{"///aa/", `["/","aa"]`},
	{"///aa//", `["/","aa"]`},
	{"./aa", `["aa"]`},
	{"././aa", `["aa"]`},
	{"./././aa", `["aa"]`},
	{"/./aa", `["/","aa"]`},
	{"/././aa", `["/","aa"]`},
	{"/./././aa", `["/","aa"]`},
	{"/aa/bb", `["/","aa","bb"]`},
	{"/aa//bb/cc/dd", `["/","aa","bb","cc","dd"]`},
	{"/aa/bb///cc/dd", `["/","aa","bb","cc","dd"]`},
	{"/aa/bb/./cc//.//dd", `["/","aa","bb","cc","dd"]`},
	{"/aa/bb.html", `["/","aa","bb.html"]`},
	{"/aa//bb/cc/dd.php", `["/","aa","bb","cc","dd.php"]`},
	{"/aa//bb/cc/dd.php/", `["/","aa","bb","cc","dd.php"]`},
	{"/aa//bb/cc/dd.php//", `["/","aa","bb","cc","dd.php"]`},
	{"/aa//bb/cc/dd.php///", `["/","aa","bb","cc","dd.php"]`},
	{"/aa/bb///cc.php/dd", `["/","aa","bb","cc.php","dd"]`},
	{"/aa/bb/./...cc//.//dd", `["/","aa","bb","...cc","dd"]`},
	{"/aa/bb/./.cc//.//dd", `["/","aa","bb",".cc","dd"]`},
	{"/../a", `["/","a"]`},
	{"/../../a", `["/","a"]`},
	{"/../../../a", `["/","a"]`},
	{"/../../../../a", `["/","a"]`},
	{"../a", `["/","a"]`},
	{"../../a", `["/","a"]`},
	{"../../../a", `["/","a"]`},
	{"../../../../a", `["/","a"]`},
	{"../../a.html", `["/","a.html"]`},
	{"../../../a.html", `["/","a.html"]`},
	{"../../../../a.html", `["/","a.html"]`},
	{"../bb/cc/../../a.html", `["/","a.html"]`},
	{"../bb/cc/dd/../../a.html", `["/","bb","a.html"]`},
	{"./bb/cc/dd/../../a.html", `["bb","a.html"]`},
	{"bb/cc/dd/../../ee/a.html", `["bb","ee","a.html"]`},
	{"bb/cc/dd/../../ee/../a.html", `["bb","a.html"]`},
	{"bb/cc/dd/../../ee/../a.html/", `["bb","a.html"]`},
	{"bb/cc/dd/../../ee/../a.html//", `["bb","a.html"]`},
	{"/./../bb/cc/dd/../../ee/../a.html//", `["/","bb","a.html"]`},
	{"/./../.../cc/dd/../../ee/../a.html//", `["/","...","a.html"]`},
	{"/redis/planb/", `["/","redis","planb"]`},
	{"//////////aa", `["/","aa"]`},
	/*
	 */
}

var testMatchRuns = []struct {
	param  string
	result int
}{
	{"/api/table/liz/", 4},
	{"/api/table/liz/22", 3},
	{"/api/table/liz/", 4},
	{"/api/table/liz/:id", 3},
	{"/api/table/liz/1", 3},
	{"/api/table/liz/123", 3},
	{"/api/table/liz/4452-232323-2323232-232323", 3},
	{"/api/table/carbone/", 7},
	{"/index.html", 9},
	{"/api/js/jQuery-2.0.1.min.js", 9},
}

func TestFixPath(t *testing.T) {

	if false {
		fmt.Printf("Hello World\n")
	}

	// rv := make([]string, 0, 25)
	rv := make([]string, 25)

	for k, test := range testRuns {
		rv = rv[:25]
		n := FixPath(test.param, rv, 25)
		// fmt.Printf("Testing %s\n", test.param)
		rv = rv[:n]
		if debug.SVar(rv) != test.result {
			t.Errorf("Test %d - FixPath(%v) = %v, want %v", k, test.param, debug.SVar(rv), test.result)
		} // else if debug {
		//	fmt.Printf("Test %d - passed\n", k)
		//}
	}
}

// 52.3 us
func OldBenchmarkFixPath(b *testing.B) {
	// noalloc = true
	rv := make([]string, 25)
	for n := 0; n < b.N; n++ {
		rv = rv[:25]
		// FixPath("/./../.../cc/dd/../../ee/../a.html//", &rv)
		FixPath("/cc/dd/a.html", rv, 25)
	}
}

var testCreateRoutes = []struct {
	route  string
	fx     Handle
	result int
}{
	{"/api/observedPage", emptyTestingHandle, 0},
	{"/api/getRecentObs", emptyTestingHandle, 0},
	{"/image/getUserRoot", emptyTestingHandle, 0},
	{"/image/getUserRoot-old-", emptyTestingHandle, 0},
	{"/api/assignPageTo", emptyTestingHandle, 0},
	{"/api/bobbob", emptyTestingHandle, 0},
	{"/api/deleteNote", emptyTestingHandle, 0},
	{"/api/deleteUrl", emptyTestingHandle, 0},
	{"/api/edit1TestSetMember", emptyTestingHandle, 0},
	{"/api/get-ip", emptyTestingHandle, 0},
	{"/api/getJs", emptyTestingHandle, 0},
	{"/api/getLogins", emptyTestingHandle, 0},
	{"/api/loginAs", emptyTestingHandle, 0},
	{"/api/markEditorDone", emptyTestingHandle, 0},
	{"/api/markToUpdate", emptyTestingHandle, 0},
	{"/api/observedPage2", emptyTestingHandle, 0},
	{"/api/perfTestDB1", emptyTestingHandle, 0},
	{"/api/perfTestDB4", emptyTestingHandle, 0},
	{"/api/ping_i_am_alive", emptyTestingHandle, 0},
	{"/api/pullDataFor", emptyTestingHandle, 0},
	{"/api/pullDataTopUser", emptyTestingHandle, 0},
	{"/api/pullListFor", emptyTestingHandle, 0},
	{"/api/pullListTopUser", emptyTestingHandle, 0},
	{"/api/releaseJob", emptyTestingHandle, 0},
	{"/api/saveJs", emptyTestingHandle, 0},
	{"/api/saveOneNote", emptyTestingHandle, 0},
	{"/api/set-ip", emptyTestingHandle, 0},
	{"/api/status_db", emptyTestingHandle, 0},
	{"/api/status_db2", emptyTestingHandle, 0},
	{"/api/table/get_monitor_data", emptyTestingHandle, 0},
	{"/api/table/img_group", emptyTestingHandle, 0},
	{"/api/table/img_set", emptyTestingHandle, 0},
	{"/api/table/t_available_test_systems", emptyTestingHandle, 0},
	{"/api/table/t_email_q", emptyTestingHandle, 0},
	{"/api/table/t_job", emptyTestingHandle, 0},
	{"/api/table/t_runSet", emptyTestingHandle, 0},
	{"/api/table/t_test_crud", emptyTestingHandle, 0},
	{"/api/table/t_test_crud2", emptyTestingHandle, 0},
	{"/api/table/t_test_crud3", emptyTestingHandle, 0},
	{"/api/table/tblActionPlan", emptyTestingHandle, 0},
	{"/api/table/tblCard", emptyTestingHandle, 0},
	{"/api/table/tblCategory", emptyTestingHandle, 0},
	{"/api/table/tblConfig", emptyTestingHandle, 0},
	{"/api/table/tblCrew", emptyTestingHandle, 0},
	{"/api/table/tblDepartment", emptyTestingHandle, 0},
	{"/api/table/tblLog", emptyTestingHandle, 0},
	{"/api/table/tblNotify", emptyTestingHandle, 0},
	{"/api/table/tblObservationType", emptyTestingHandle, 0},
	{"/api/table/tblPerson", emptyTestingHandle, 0},
	{"/api/table/tblSeverity", emptyTestingHandle, 0},
	{"/api/table/tblSite", emptyTestingHandle, 0},
	{"/api/test/change_password", emptyTestingHandle, 0},
	{"/api/test/confirm_email", emptyTestingHandle, 0},
	{"/api/test/extendlogin", emptyTestingHandle, 0},
	{"/api/test/getrun", emptyTestingHandle, 0},
	{"/api/test/login", emptyTestingHandle, 0},
	{"/api/test/logout", emptyTestingHandle, 0},
	{"/api/test/monitor_it_happened", emptyTestingHandle, 0},
	{"/api/test/password_reset", emptyTestingHandle, 0},
	{"/api/test/register_client", emptyTestingHandle, 0},
	{"/api/test/register_new_user", emptyTestingHandle, 0},
	{"/api/test/stayLoggedIn", emptyTestingHandle, 0},
	{"/api/upd1TestSetMember", emptyTestingHandle, 0},
	{"/api/updatePassword", emptyTestingHandle, 0},
	{"/api/visitedPage", emptyTestingHandle, 0},
	{"/image/createImageUser", emptyTestingHandle, 0},
	{"/image/delete_img_file", emptyTestingHandle, 0},
	{"/image/delete_img_set", emptyTestingHandle, 0},
	{"/image/insert_img_file", emptyTestingHandle, 0},
	{"/image/insert_img_set", emptyTestingHandle, 0},
	{"/image/setupTest", emptyTestingHandle, 0},
	{"/redis/test1", emptyTestingHandle, 0},
}

var StaticRoutes = []string{
	"/api/observedPage",
	"/api/getRecentObs",
	"/image/getUserRoot",
	"/image/getUserRoot-old-",
	"/api/assignPageTo",
	"/api/bobbob",
	"/api/deleteNote",
	"/api/deleteUrl",
	"/api/edit1TestSetMember",
	"/api/get-ip",
	"/api/getJs",
	"/api/getLogins",
	"/api/loginAs",
	"/api/markEditorDone",
	"/api/markToUpdate",
	"/api/observedPage2",
	"/api/perfTestDB1",
	"/api/perfTestDB4",
	"/api/ping_i_am_alive",
	"/api/pullDataFor",
	"/api/pullDataTopUser",
	"/api/pullListFor",
	"/api/pullListTopUser",
	"/api/releaseJob",
	"/api/saveJs",
	"/api/saveOneNote",
	"/api/set-ip",
	"/api/status_db",
	"/api/status_db2",
	"/api/table/get_monitor_data",
	"/api/table/img_group",
	"/api/table/img_set",
	"/api/table/t_available_test_systems",
	"/api/table/t_email_q",
	"/api/table/t_job",
	"/api/table/t_runSet",
	"/api/table/t_test_crud",
	"/api/table/t_test_crud2",
	"/api/table/t_test_crud3",
	"/api/table/tblActionPlan",
	"/api/table/tblCard",
	"/api/table/tblCategory",
	"/api/table/tblConfig",
	"/api/table/tblCrew",
	"/api/table/tblDepartment",
	"/api/table/tblLog",
	"/api/table/tblNotify",
	"/api/table/tblObservationType",
	"/api/table/tblPerson",
	"/api/table/tblSeverity",
	"/api/table/tblSite",
	"/api/test/change_password",
	"/api/test/confirm_email",
	"/api/test/extendlogin",
	"/api/test/getrun",
	"/api/test/login",
	"/api/test/logout",
	"/api/test/monitor_it_happened",
	"/api/test/password_reset",
	"/api/test/register_client",
	"/api/test/register_new_user",
	"/api/test/stayLoggedIn",
	"/api/upd1TestSetMember",
	"/api/updatePassword",
	"/api/visitedPage",
	"/image/createImageUser",
	"/image/delete_img_file",
	"/image/delete_img_set",
	"/image/insert_img_file",
	"/image/insert_img_set",
	"/image/setupTest",
}

/*
	route  string
	fx     Handle
	result int


var rr *Router

func OldBenchmarkCreateRoutes01(b *testing.B) {
	rr = NewRouter()
	for i, v := range testCreateRoutes {
		testCreateRoutes[i].result = i
		rr.CreateRoute( v.route, v.fx)
	}
	for n := 0; n < b.N; n++ {
		_ = rr.MatchRoute("GET", "/api/test/stayLoggedIn/")
	}
}

func OldBenchmarkCreateRoutes02(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = rr.MatchRoute("GET", "/api/test/stayLoggedIn/")
	}
}
*/

func OldBenchmarkHash01(b *testing.B) {
	h := 0
	l := 0
	for n := 0; n < b.N; n++ {
		//for _, v := range testCreateRoutes {
		v := testCreateRoutes[5]
		l = len(v.route)
		if l > 15 {
			h = (l + int(v.route[14]) + int(v.route[l-6]) + int(v.route[9]))
		} else {
			h = (l + int(v.route[l-6]) + int(v.route[9]))
		}
		_ = h
		//}
	}
}

// 127 us per Alloc call (avg)
type TestAlloSpeed struct {
	Fx  []int
	nFx int
}

func Alloc30b() *TestAlloSpeed {
	return &TestAlloSpeed{nFx: 10, Fx: make([]int, 0, 10)}
}

func OldBenchmarkMemAlloc(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = Alloc30b()
	}
}

var cmp string = "/abcdefghijklmnopqrstuvwxyz!"

func OldBenchmarkHash02(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if "/abcdefghijklmnopqrstuvwxyz" == cmp {
		}
	}
}

var arrived = 0

func emptyTestingHandle(w http.ResponseWriter, r *http.Request) {
	arrived = 1
}

func serveFileTestingHandle(w http.ResponseWriter, r *http.Request) {
	arrived = 2
}
