package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/MiscLib"
)

/*

TODO:

	------------------------------------------- phase 1 -------------------------------------------

	0. Connect to PostgreSQL on client (.150:12000) -- or -- .157:16020?
	0. Add in TabServer
	0. Get back result
	0. Validate it
	0. Run all sorts of resutls and validate them
	0. Add in direct-connect to d.b. and a validate results in d.b. (need a d.b. setup first - to get to initial state)
	0. Add in direct-connect to redis -- used in 1a --

	0. Metthods -- Test of TabServer, GoTemplate etc.
		1. .Put, .Post, .Delete, .Head, .Options etc.
		2. Body encode system for different things
		3. Add params as necessary

	------------------------------------------- phase 1a ------------------------------------------

	1. test caching

	------------------------------------------- phase 1b ------------------------------------------

	1. test proxy
	1. test rewrite
	1. test limit*
	1. test ban*
	1. test Latency
	1. test Logging

	------------------------------------------- phase 2 -------------------------------------------

	2. test of FileServer
	2. Cookies / Headers -- test of FileServer
		1. Set cookies - on request
		2. Set headers
		1. Validate cookies
		2. Validate headers
		1. Cookie Jar
		1. Handle SetCookie headers

	------------------------------------------- phase 3 -------------------------------------------

	3. Security Tests -- Client side drive for AesSrp code, Basci*, Un/Pw auth.
	3. Test LoginRequired

	------------------------------------------- phase 4 -------------------------------------------

	4. Multi-Requests ( N of ... at the same time )	-- Performance
	4. Delta T - timting of requests	-- Peformance -- Reliability -- Scalability

*/

type runTestType struct {
	Url             string                 // The url unless created by PrepIt function
	TheComment      string                 // What is this test all about
	CheckStatus     bool                   // if true then check return status
	ExpectedStatus  int                    // status expected, 200, 404 etc.
	Data            map[string]interface{} // data to pass to PrepIt, CheckIt functions -- or convert to parms for cli , or body, or cookies ?
	CheckRegExp     []string               // patterns to check that data is correct
	ExitIfTestFails bool                   // if test at this URL fails then quit testing (1st test, no connectivity)
	CheckBodyFn     string                 // file name for a body check ./ref/Name.out
	TestNo          int                    // what number test is this
	Run_it          bool                   // if false skip test
	DumpFile        bool                   // if false skip test
	OutputFile      string                 // file name to output to
}

type topRunTestType struct {
	Name    string
	Servers []string
	Tests   []runTestType
}

// headers - in
// cookies - in
// headers - out
// cookies - out
// * body - out
// * file to save body in - out
// * file to compare body to - out
// * grep[s] to apply to body out to test
// delta-T - out max time before failed response - out
// METHOD - get, post, put, delete, head, options - in
// Origin - in
// ? PrepIt - Fx - convert this plust "data" into the URL and other values.
// ? CheckIt - Fx - validate results - speical - out
//

//
// - What about paraalellized tests with "go" routines -
// - What about N iterations of a test  -
//

// add in POST, PUT etc.

// add in set of errors, n_errs that occured for output report, DetalT for the call.

var testSet1 topRunTestType

func runATestGET(reS *runTestType, tname, aServer string) (err []error) {

	if !reS.Run_it {
		return
	}

	// add in headers
	// add in cookies

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	dm := make(map[string]string)
	dm["Server"] = aServer
	rs := fmt.Sprintf("%d", r1.Intn(10000000))
	dm["random"] = rs // Used for cache busting

	lUrl := sizlib.Qt(reS.Url, dm)

	// do a "get"
	resp, e0 := http.Get(lUrl)
	if e0 != nil {
		err = append(err, e0)
		return
	}
	defer resp.Body.Close()

	// get back response -- copy to output
	//len, err := io.Copy(os.Stdout, resp.Body)
	//if err != nil {
	//	return err
	//}

	bodyBytes, e0 := ioutil.ReadAll(resp.Body)
	if e0 != nil {
		err = append(err, e0)
		return
	}
	body := string(bodyBytes)

	// validate status code
	if reS.CheckStatus && resp.StatusCode != reS.ExpectedStatus {
		err = append(err, fmt.Errorf("Invalid status - expected %d got %d, url:%s", reS.ExpectedStatus, resp.StatusCode, reS.Url))
		if reS.ExitIfTestFails {
			return
		}
	}

	// validate contents with "grep"/"re" or -- compare entire file
	for ii, vv := range reS.CheckRegExp {
		re, e1 := regexp.Compile(vv)
		if e1 != nil {
			err = append(err, fmt.Errorf("Error: regular expression [%s] did not compile, %s, %d in set, TestNo:%d\n", err, vv, ii, reS.TestNo))
		} else {

			if !re.MatchString(body) {
				err = append(err, fmt.Errorf("Failed to match regular expression [%s] with body ---[[[[%s]]]---, pos %d in set, TestNo:%d\n", vv, body, ii, reS.TestNo))
				if reS.ExitIfTestFails {
					return
				}
			}

		}
	}

	// validate entire body
	if reS.CheckBodyFn != "" {
		// xyzzy - if has %d fmt in it...
		cmpBody, e2 := ioutil.ReadFile(fmt.Sprintf(reS.CheckBodyFn, tname, reS.TestNo))
		if e2 != nil {
			err = append(err, fmt.Errorf("Error: unable to open/read [%s], error:%s, TestNo:%d\n", reS.CheckBodyFn, err, reS.TestNo))
		} else if string(cmpBody) != body {
			err = append(err, fmt.Errorf("Error: body did not match, expected ---[[[%s]]]---, got ---[[[%s]]]---, TestNo:%d\n", cmpBody, body, reS.TestNo))
			if reS.ExitIfTestFails {
				return
			}
		}
	}

	if reS.DumpFile {
		fmt.Printf("Dump Response --->>>%s<<<---\n", body)
	}

	if reS.OutputFile != "" {
		// xyzzy - if has %d fmt in it...
		ioutil.WriteFile(fmt.Sprintf(reS.OutputFile, tname, reS.TestNo), []byte(body), 0600)
	}

	return
}

// ---------------------------------------------------------------------------- main ----------------------------------------------------------------------------
// ---------------------------------------------------------------------------- main ----------------------------------------------------------------------------
// ---------------------------------------------------------------------------- main ----------------------------------------------------------------------------

var PGConn = flag.String("conn", "", "PotgresSQL connection info")     // 0
var DBName = flag.String("dbname", "test", "PotgresSQL database name") // 8
var TestSet = flag.String("testset", "./set1.json", "test set to run") // 1
func init() {
	flag.StringVar(PGConn, "C", "", "PotgresSQL connection info")   // 0
	flag.StringVar(PGConn, "N", "test", "PotgresSQL database name") // 8
	flag.StringVar(TestSet, "t", "./set1.json", "test set to run")  // 1
}

// var pg_client *sizlib.MyDb // Client connection for PostgreSQL

func main() {

	flag.Parse()

	// ConnectToPostgreSQL()

	sb, err := ioutil.ReadFile(*TestSet)
	err = json.Unmarshal(sb, &testSet1)
	if err != nil {
		fmt.Printf("Unable to open input data, %s\n", *TestSet)
		os.Exit(1)
	}

	n_err := 0
	n_test := 0
	n_skip := 0
	n_fail := 0
	var gErr []error
	for _, aServer := range testSet1.Servers {
		for jj, ww := range testSet1.Tests {
			ww.TestNo = jj
			if !ww.Run_it {
				n_skip++
			} else {
				n_test++
			}
			err := runATestGET(&ww, testSet1.Name, aServer)
			if len(err) > 0 {
				n_fail++
				n_err += len(err)
				gErr = append(gErr, err...)
			}
		}
	}
	if n_err > 0 {
		fmt.Printf("%s%d/%d completed - %s %d failed%s", MiscLib.ColorGreen, n_test-n_fail, n_test, MiscLib.ColorRed, n_fail, MiscLib.ColorReset)
		if n_skip > 0 {
			fmt.Printf("%s %d skipped%s", MiscLib.ColorYellow, n_skip, MiscLib.ColorReset)
		}
		fmt.Printf("\n")
		fmt.Printf("Error: %s\n", gErr)
	} else {
		fmt.Printf("%s%d/%d completed%s", MiscLib.ColorGreen, n_test, n_test, MiscLib.ColorReset)
		if n_skip > 0 {
			fmt.Printf("%s %d skipped%s", MiscLib.ColorYellow, n_skip, MiscLib.ColorReset)
		}
		fmt.Printf("\n")
		fmt.Printf("PASS\n")
	}
}
