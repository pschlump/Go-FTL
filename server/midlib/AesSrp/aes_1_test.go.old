//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1229
//

package AesSrp

//
// NOTE! **** Requries Redis to run test ****
//
// Running this test requries a redis server to be running with the follwing information - The configuration for the
// server should be in ../test_redis.json
//
// To setup the data for this test run
//
//   $ redis-cli ... <./redis-setup.redis
//

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/tr"
)

// rr "github.com/pschlump/rediswrap" //

// -----------------------------------------------------------------------------------------------------------------------------------------------
func Test_BasicAuthServer(t *testing.T) {

	if !cfg.SetupRedisForTest("../test_redis.json") {
		return
	}

	/*
	   {"$saved_one_time_key_hashed$":"9632962",
	   	"DeviceID":"51748634",
	   	"auth":"P",
	   	"backup_one_time_keys":"de95375edfb9116de3bff4037696448f,
	   	115e6522c787b20e2125c39d2be7c67e,
	   	a5a5fdd8ed3573f38be10dbe6af32922,
	   	be5fa8d0b12ea6a6ee336b07866110dd,
	   	ff3c826d0c2ba7fa00b68ee575a2f2dc,
	   	bf6e1e2af8bf9ad5d7edad88196cf002,
	   	c1d1c4c995fa059c1ebb0b96b7c9ed23,
	   	af14814ff7048752410852660e1f252d,
	   	fae03fc58d2cdb40f760d4929d3769be,
	   	4060963a88e9cbbd0fb201160e217405,
	   	06494e1649271f5794361e4250770fe3,
	   	ae6fd6fa91540d0a95e95370e3d2dd8f,
	   	52ae5aff451f420186b9043b5bc5e49f,
	   	20cf42f7fb0a0f266dc2f3c79dc79127,
	   	f5cbb5ac411c1c817b29381cf974b0cc,
	   	83ed37b55586fcc86e59bb642dfe06e9,
	   	fcd699c2ee5a681a067a0f05c4e78072,
	   	e4a2bfd95a68da9d7984e4936a228863,
	   	c121939071b3f5b7f4bac55f8241296a",
	   	"confirmed":"y",
	   	"disabled":"n",
	   	"disabled_reason":"",
	   	"fingerprint":"0",
	   	"login_date_time":"2015-12-19T21:12:10-07:00",
	   	"login_fail_time":"",
	   	"n_failed_login":"0",
	   	"num_login_times":"19",
	   	"privs":"user",
	   	"register_date_time":"2015-12-19T21:12:10-07:00",
	   	"salt":"42ce852b31aa2beb5e2f89872f944d4b",
	   	"v":"51a5be478b590a312b5ee9c76c12a360c8d5952721d27a03cb9c6e55097a831883cfbe2cdbafa4aa47d6f62971ff6ce9c6b05886a5d783b9c0fc56c5a14ef795693cc2d3e2a66d12cd82d5fd83579949ee48111ac4e31cd0f3a60d661042c9f2334f0894dff979ca0e1f8cc06624b24370b05d8a65454454ea3f0a93fa05362dac528bdfe8d8eef084349791fe4387f9b0e05e3093656b7c5db4a9b81e9d18161c7e69cf0538261c90a90b287cea73deb1b0ea22cc22c1926de39da71c7751ae27155444341156736e069d21e8d0467256d48790bdecb9ed5990eb417452b4e8b41cdd320116b3a417d840e99abaee8ba91ca1014e865bb267da89503ace0316"}
	*/

	tests := []struct {
		RunTest             bool
		url                 string
		expectedCode        int
		isSuccessful        bool
		username            string
		password            string
		UserNameForRegister bool
		FailureCode         string
		RedisOk             func() bool
		StatusIs            string
		CodeIs              string
	}{
		// If run twice then 2nd one will re-send confiermation email (Verify that it sets the 'salt' and 'v' values form 2nd try)
		{
			RunTest:      true,
			url:          "http://localhost:3118/api/srp_register?email=t1@example.com&salt=42ce852b31aa2beb5e2f89872f944d4b&v=51a5be478b590a312b5ee9c76c12a360c8d5952721d27a03cb9c6e55097a831883cfbe2cdbafa4aa47d6f62971ff6ce9c6b05886a5d783b9c0fc56c5a14ef795693cc2d3e2a66d12cd82d5fd83579949ee48111ac4e31cd0f3a60d661042c9f2334f0894dff979ca0e1f8cc06624b24370b05d8a65454454ea3f0a93fa05362dac528bdfe8d8eef084349791fe4387f9b0e05e3093656b7c5db4a9b81e9d18161c7e69cf0538261c90a90b287cea73deb1b0ea22cc22c1926de39da71c7751ae27155444341156736e069d21e8d0467256d48790bdecb9ed5990eb417452b4e8b41cdd320116b3a417d840e99abaee8ba91ca1014e865bb267da89503ace0316&_ran_=2323232323232323232",
			isSuccessful: true,
			username:     "t1@example.com",
			password:     "bobbob",
			expectedCode: http.StatusOK,
			// Setup of test - delete "srp:U:t1@example.com"
			// Test at the end - verify that we have user - and it is not confirmed.
			RedisOk: func() bool {
				// Verify t1.example.com has been creatd
				// var ServerGlobal *ServerGlobalConfigType
				conn, _ := cfg.ServerGlobal.RedisPool.Get()
				defer cfg.ServerGlobal.RedisPool.Put(conn)
				v, err := conn.Cmd("GET", "srp:U:t1@example.com").Str()
				if err != nil {
					fmt.Printf("Error in ReisOk, test1, --->%s<---\n", err)
					return false
				}
				if len(v) == 0 {
					fmt.Printf("Error in ReisOk, test1, -- length 0\n")
					return false
				}
				return true
			},
		},
		{
			RunTest:      true,
			url:          "http://localhost:3118/api/srp_register?email=t1@example.com&salt=42ce852b31aa2beb5e2f89872f944d4b&v=51a5be478b590a312b5ee9c76c12a360c8d5952721d27a03cb9c6e55097a831883cfbe2cdbafa4aa47d6f62971ff6ce9c6b05886a5d783b9c0fc56c5a14ef795693cc2d3e2a66d12cd82d5fd83579949ee48111ac4e31cd0f3a60d661042c9f2334f0894dff979ca0e1f8cc06624b24370b05d8a65454454ea3f0a93fa05362dac528bdfe8d8eef084349791fe4387f9b0e05e3093656b7c5db4a9b81e9d18161c7e69cf0538261c90a90b287cea73deb1b0ea22cc22c1926de39da71c7751ae27155444341156736e069d21e8d0467256d48790bdecb9ed5990eb417452b4e8b41cdd320116b3a417d840e99abaee8ba91ca1014e865bb267da89503ace0316&_ran_=2323232323232323232",
			isSuccessful: true,
			username:     "t1@example.com",
			password:     "bobbob",
			expectedCode: http.StatusOK,
			FailureCode:  "9001", // xyzzy - should return a failure code, check for that.
			RedisOk:      func() bool { return true },
			StatusIs:     "error",
			CodeIs:       "9001",
		},
		{ // Salt is too small - will error
			RunTest:      false,
			url:          "http://localhost:3118/api/srp_register?email=t3@example.com&salt=1&v=51a5be478b590a312b5ee9c76c12a360c8d5952721d27a03cb9c6e55097a831883cfbe2cdbafa4aa47d6f62971ff6ce9c6b05886a5d783b9c0fc56c5a14ef795693cc2d3e2a66d12cd82d5fd83579949ee48111ac4e31cd0f3a60d661042c9f2334f0894dff979ca0e1f8cc06624b24370b05d8a65454454ea3f0a93fa05362dac528bdfe8d8eef084349791fe4387f9b0e05e3093656b7c5db4a9b81e9d18161c7e69cf0538261c90a90b287cea73deb1b0ea22cc22c1926de39da71c7751ae27155444341156736e069d21e8d0467256d48790bdecb9ed5990eb417452b4e8b41cdd320116b3a417d840e99abaee8ba91ca1014e865bb267da89503ace0316&_ran_=2323232323232323232",
			isSuccessful: true,
			username:     "t1@example.com",
			password:     "bobbob",
			expectedCode: http.StatusOK,
			RedisOk:      func() bool { return true },
			// Setup of test - delete "srp:U:t1@example.com"
			// Test at the end - verify that we have user - and it is not confirmed.
			StatusIs: "error",
			CodeIs:   "9004",
		},
		// do DeviceID and UserName - with config and w/o config
		{
			RunTest:      true,
			url:          "http://localhost:3118/api/srp_register?DeviceID=81234567&salt=42ce852b31aa2beb5e2f89872f944d4b&v=51a5be478b590a312b5ee9c76c12a360c8d5952721d27a03cb9c6e55097a831883cfbe2cdbafa4aa47d6f62971ff6ce9c6b05886a5d783b9c0fc56c5a14ef795693cc2d3e2a66d12cd82d5fd83579949ee48111ac4e31cd0f3a60d661042c9f2334f0894dff979ca0e1f8cc06624b24370b05d8a65454454ea3f0a93fa05362dac528bdfe8d8eef084349791fe4387f9b0e05e3093656b7c5db4a9b81e9d18161c7e69cf0538261c90a90b287cea73deb1b0ea22cc22c1926de39da71c7751ae27155444341156736e069d21e8d0467256d48790bdecb9ed5990eb417452b4e8b41cdd320116b3a417d840e99abaee8ba91ca1014e865bb267da89503ace0316&_ran_=2323232323232323232",
			isSuccessful: true,
			username:     "t1@example.com",
			password:     "bobbob",
			expectedCode: http.StatusOK,
			// Setup of test - delete "srp:U:t1@example.com"
			// Test at the end - verify that we have user - and it is not confirmed.
			RedisOk: func() bool {
				// Verify t1.example.com has been creatd
				conn, _ := cfg.ServerGlobal.RedisPool.Get()
				defer cfg.ServerGlobal.RedisPool.Put(conn)
				v, err := conn.Cmd("GET", "srp:U:81234567").Str()
				if err != nil {
					fmt.Printf("Error in ReisOk, test1, --->%s<---\n", err)
					return false
				}
				if len(v) == 0 {
					fmt.Printf("Error in ReisOk, test1, -- length 0\n")
					return false
				}
				return true
			},
		},
		{
			RunTest:             true,
			url:                 "http://localhost:3118/api/srp_register?email=t1@example.com&UserName=fredFred&salt=42ce852b31aa2beb5e2f89872f944d4b&v=51a5be478b590a312b5ee9c76c12a360c8d5952721d27a03cb9c6e55097a831883cfbe2cdbafa4aa47d6f62971ff6ce9c6b05886a5d783b9c0fc56c5a14ef795693cc2d3e2a66d12cd82d5fd83579949ee48111ac4e31cd0f3a60d661042c9f2334f0894dff979ca0e1f8cc06624b24370b05d8a65454454ea3f0a93fa05362dac528bdfe8d8eef084349791fe4387f9b0e05e3093656b7c5db4a9b81e9d18161c7e69cf0538261c90a90b287cea73deb1b0ea22cc22c1926de39da71c7751ae27155444341156736e069d21e8d0467256d48790bdecb9ed5990eb417452b4e8b41cdd320116b3a417d840e99abaee8ba91ca1014e865bb267da89503ace0316&_ran_=2323232323232323232",
			isSuccessful:        true,
			username:            "t1@example.com",
			password:            "bobbob",
			expectedCode:        http.StatusOK,
			UserNameForRegister: true,
			RedisOk: func() bool {
				// Verify t1.example.com has been creatd
				conn, _ := cfg.ServerGlobal.RedisPool.Get()
				defer cfg.ServerGlobal.RedisPool.Put(conn)
				v, err := conn.Cmd("GET", "srp:U:fredFred").Str()
				if err != nil {
					fmt.Printf("Error in ReisOk, test1, --->%s<---\n", err)
					return false
				}
				if len(v) == 0 {
					fmt.Printf("Error in ReisOk, test1, -- length 0\n")
					return false
				}
				return true
			},
		},
	}

	bot := mid.NewServer()

	ms := NewAesSrpServer(bot, []string{"/api"}, []string{}, cfg.ServerGlobal)

	var err error
	lib.SetupTestCreateDirs()

	// ------------------------------------------------------------------------------------------------------------------------------------
	// Clean up of Redis before test starts
	// ------------------------------------------------------------------------------------------------------------------------------------
	{
		key := []string{
			"srp:U:t2@example.com",
			"srp:U:fredFred",
			"srp:U:t1@example.com",
			"srp:U:81234567",
		}
		conn, _ := cfg.ServerGlobal.RedisPool.Get()
		defer cfg.ServerGlobal.RedisPool.Put(conn)
		for ii, kk := range key {
			err := conn.Cmd("DEL", kk).Err
			if err != nil {
				fmt.Printf("On %d Redis Error %s\n", ii, err)
			}
		}
	}

	for ii, test := range tests {

		if test.RunTest {
			if db8 {
				fmt.Printf("\nTest %d -------------------------------------------------------\n", ii)
			}
			ms.UserNameForRegister = test.UserNameForRegister

			rec := httptest.NewRecorder()
			wr := goftlmux.NewMidBuffer(rec, nil)

			id := "test-01-BasicAuthServer"
			trx := tr.NewTrx(cfg.ServerGlobal.RedisPool)
			trx.TrxIdSeen(id, test.url, "GET")
			wr.RequestTrxId = id

			wr.G_Trx = trx

			var req *http.Request

			req, err = http.NewRequest("GET", test.url, nil)
			if err != nil {
				t.Fatalf("Test %d: Could not create HTTP request: %v", ii, err)
			}
			goftlmux.ParseQueryParamsReg(wr, req, &wr.Ps) //
			lib.SetupTestMimicReq(req, "localhost:3118")
			if dbA {
				fmt.Printf("{\"req\":%s,\n\"wr\":%s}\n", lib.SVarI(req), lib.SVarI(wr))
			}

			ms.ServeHTTP(wr, req)

			code := wr.StatusCode
			// Tests to perform on final recorder data.
			if code != test.expectedCode {
				t.Errorf("Error %2d, reject error got: %d, expected %d\n", ii, wr.StatusCode, test.expectedCode)
			}

			if test.isSuccessful { // registration occured
			} else { // No registraiton - verify reason returned
			}

			wr.FinalFlush()
			body := wr.GetBody()

			// StatusIs:     "error",
			// CodeIs:       "9001",
			rvMap := make(map[string]string)
			rvMap, err := lib.JsonStringToString(string(body))
			if err != nil {
				t.Errorf("Error %2d, Unable to parse JSON return data, >>>%s<<<\n", ii, body)
			}

			if rvMap["status"] == "success" {
				if !test.RedisOk() {
					t.Errorf("Error %2d, Redis Data is NOT Ok\n", ii)
				}
			} else {
				if rvMap["status"] != test.StatusIs {
					t.Errorf("Error %2d, Invalid error status, expected %s got %s\n", ii, test.StatusIs, rvMap["status"])
				}
				if rvMap["code"] != test.CodeIs {
					t.Errorf("Error %2d, Invalid error code , expected %s got %s\n", ii, test.CodeIs, rvMap["code"])
				}
			}

			fmt.Printf("\n")
		}
	}

}

const db8 = false
const dbA = false

/* vim: set noai ts=4 sw=4: */
