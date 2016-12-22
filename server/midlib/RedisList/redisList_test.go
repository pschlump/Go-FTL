//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1274
//

package RedisList

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	// "github.com/mediocregopher/radix.v2/redis"
)

// -----------------------------------------------------------------------------------------------------------------------------------------------

func Test_RewriteServer(t *testing.T) {

	cfg.SetupRedisForTest("../test_redis.json")

	SetupRedisListTest()

	tests := []struct {
		url            string
		hdr            []lib.NameValue
		expectedOutput string
	}{
		{
			"http://example.com/foo?abc=def&$privs$=user&t=user",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			`[{"$key$"`,
		},
		{
			"http://example.com/def?$privs$=user&t=user",
			[]lib.NameValue{lib.NameValue{Name: "X-Test", Value: "A-Value"}},
			"",
		},
	}

	bot := mid.NewServer()
	ms := NewRedisListServer(bot, []string{"/foo"}, "srp:U:", []string{"bob,c1,c2", "dave,c1,c3,c5", "user,$key$,DeviceId,auth"})
	ms.gCfg = cfg.ServerGlobal

	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec, nil) // var wr http.ResponseWriter
		// lib.SetupTestCreateHeaders(wr, test.hdr)

		var req *http.Request

		req, err = http.NewRequest("GET", test.url, nil)
		if err != nil {
			t.Fatalf("Test %d: Could not create HTTP request: %v", ii, err)
		}
		goftlmux.ParseQueryParamsReg(wr, req, &wr.Ps) //
		lib.SetupTestMimicReq(req, "example.com")
		if db3 {
			fmt.Printf("{\"req\":%s,\n\"wr\":%s}\n", lib.SVarI(req), lib.SVarI(wr))
		}
		lib.SetupRequestHeaders(req, test.hdr)

		// bot.SetInfo("what gets returned")
		bot.SetInfo("")

		ms.ServeHTTP(wr, req)

		wr.FinalFlush()
		s := wr.GetBody()
		// fmt.Printf("Final body >%s<\n", s)

		if !strings.HasPrefix(string(s), test.expectedOutput) {
			t.Errorf("Error %2d, Invalid output: >%s<, expected >%s<\n", ii, s, test.expectedOutput)
		}

	}
}

/*
req ->{
	"Method": "GET",
	"URL": {
		"Scheme": "http",
		"Opaque": "",
		"User": null,
		"Host": "localhost:8204",
		"Path": "/api/process",
		"RawQuery": "path=foo\u0026name=example.com\u0026abc=def",
		"Fragment": ""
	},
	"Proto": "HTTP/1.1",
	"ProtoMajor": 1,
	"ProtoMinor": 1,
	"Header": {
		"X-Test": [
			"A-Value"
		]
	},
	"Body": null,
	"ContentLength": 0,
	"TransferEncoding": null,
	"Close": false,
	"Host": "example.com",
	"Form": null,
	"Form": null,
	"PostForm": null,
	"MultipartForm": null,
	"Trailer": null,
	"RemoteAddr": "1.2.2.2:52180",
	"RequestURI": "/api/process?path=foo\u0026name=example.com\u0026abc=def",
	"TLS": null
}<-
*/

const db3 = false

/*
func OldTest_RewriteServer(t *testing.T) {

	// xyzzy - need to pull this in from a config file for test.
	cfg.ServerGlobal = &cfg.ServerGlobalConfigType{}
	cfg.ServerGlobal.RedisConnectHost = "192.168.0.133"
	cfg.ServerGlobal.RedisConnectAuth = "lLJSmkccYJiVEwskr1RM4MWIaBM"
	cfg.ServerGlobal.ConnectToRedis()
	// xyzzy - to have data setup in Redis - do that in this code.

	tests := []struct {
		url         string
		hdr         []lib.NameValue
		expectedUrl string
	}{
		{
			"http://example.com/foo?abc=def&$privs$=user&t=user",
			[]lib.NameValue{lib.NameValue{"X-Test", "A-Value"}},
			"http://localhost:8204/api/process?path=foo&name=example.com&abc=def",
		},
		{
			"http://example.com/def?$privs$=user&t=user",
			[]lib.NameValue{lib.NameValue{"X-Test", "A-Value"}},
			"http://example.com/def",
		},
	}
	bot := mid.NewServer()
	ms := NewRedisListRawServer(bot, []string{"/foo"}, "srp:U:", []string{"user,$key$,DeviceId,auth"}, []string{})
	var err error
	lib.SetupTestCreateDirs()

	for ii, test := range tests {

		rec := httptest.NewRecorder()

		wr := goftlmux.NewMidBuffer(rec,nil) // var wr http.ResponseWriter
		// lib.SetupTestCreateHeaders(wr, test.hdr)

		var req *http.Request

		req, err = http.NewRequest("GET", test.url, nil)
		if err != nil {
			t.Fatalf("Test %d: Could not create HTTP request: %v", ii, err)
		}
		goftlmux.ParseQueryParamsReg(wr, req, &wr.Ps) //
		lib.SetupTestMimicReq(req, "example.com")
		if db3 {
			fmt.Printf("{\"req\":%s,\n\"wr\":%s}\n", lib.SVarI(req), lib.SVarI(wr))
		}
		lib.SetupRequestHeaders(req, test.hdr)

		ms.ServeHTTP(wr, req)

		fmt.Printf("wr.Table = %s\n", lib.SVarI(wr.Table))
		fmt.Printf("wr.State = %s\n", wr.State)

		if wr.Table != nil { // len of TableData > 0 &&
			// Tests to perform on final recorder data.
			if wr.State != goftlmux.TableBuffer {
				t.Errorf("Error %2d, Invalid data returned\n", ii)
			}
			// xyzzy - verify data
		}

		wr.FinalFlush()
		s := wr.GetBody()
		fmt.Printf("Final body >%s<\n", s)

		// xyzzy -has been hand checked and works-  Need auto-test

	}

}
*/

func SetupRedisListTest() {

	aKey := "test:Key:"
	aValue := ""

	conn, err := cfg.ServerGlobal.RedisPool.Get()
	if err != nil {
		fmt.Printf("Test unable to get redis connection from pool\n")
		return
	}
	defer cfg.ServerGlobal.RedisPool.Put(conn)

	aValue = "{\"$saved_one_time_key_hashed$\":\"9251836\",\"DeviceId\":\"14101190\",\"auth\":\"P\",\"backup_one_time_keys\":\"b09f6c8f907e49e047d5ef8864051c24,c0c0a3f8ca13d289fac8572d4ef3c2ca,ca9fd61a901f41716f70ef225379e75d,7998a0e2a5fcacffbc73e72bcf87d975,2f2c3eb2048cf2d789e35819d6ff7cb5,70259801824ca7bcf99f7a87b0504595,ce4b8a3434b93cf8054e00ea72dbb0c0,67ad454a0ecd7ad65fa092cf1e8e632a,e9030683f1304c70afc9f87a92503230,6e4c133f416a59d71a31341ae883340f,15f9ce236d08e505ca264b7276c5c215,5dceca3e75ae405f84a3f2285c81eda4,b39e13ef388095a23162b4dd77296957,4d2a84036e809d1729442ba3af0a87d4,c1e5d70fd4fe00630e1c2b16cd68eb97,720d5d1551df60d9c6d1d944a0a01760,29d038f2a84049e5c31f1a530e11af06,a8c24bc9e36970494173dcef764825fe,cf571fd425e05ef661a24891234436d2\",\"confirmed\":\"y\",\"disabled\":\"n\",\"disabled_reason\":\"\",\"fingerprint\":\"c1167a8fed52653bc57fcf2a8dc990e9\",\"login_date_time\":\"2015-12-18T21:10:37-07:00\",\"login_fail_time\":\"\",\"n_failed_login\":\"0\",\"num_login_times\":\"11\",\"privs\":\"user\",\"register_date_time\":\"2015-12-18T21:10:37-07:00\",\"salt\":\"18a281c26d14f1291604f522328da97e\",\"v\":\"3b127b41533578ba1dc7d757044623fa83c9e612f8b6a620f75bee16ab00cca8968b6d246252027da02541114b2e87da4e52f1da8c6550e586ff13c4e49c6da71228a21b280e7253e90f8d54c505dd35194d6931f8e6abe881a7e73c0743a3972baee91e321a9249ce573758570f54f1e2639b51d5d2e9862c8406e50be6255613026c5fd2b6c0f4bfdb7b5c5a80ea5d0ad1e375f3677778e6b79d2b83f4642ef73ef67c7f62bd226ca36ff07c349a64e0bb049d1a268fc040ed96b092a5118b6565a10af8d2b53200625c988f65f8d07501701bad56083d01657eaf58a3319d17de856706c3d550a7fb5923dc2315bcd2ee009af73e0d0101aa35366f100498\"}"
	err = conn.Cmd("SET", aKey+"1", aValue).Err
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	aValue = "{\"$saved_one_time_key_hashed$\":\"9251836\",\"DeviceId\":\"14101190\",\"auth\":\"P\",\"backup_one_time_keys\":\"b09f6c8f907e49e047d5ef8864051c24,c0c0a3f8ca13d289fac8572d4ef3c2ca,ca9fd61a901f41716f70ef225379e75d,7998a0e2a5fcacffbc73e72bcf87d975,2f2c3eb2048cf2d789e35819d6ff7cb5,70259801824ca7bcf99f7a87b0504595,ce4b8a3434b93cf8054e00ea72dbb0c0,67ad454a0ecd7ad65fa092cf1e8e632a,e9030683f1304c70afc9f87a92503230,6e4c133f416a59d71a31341ae883340f,15f9ce236d08e505ca264b7276c5c215,5dceca3e75ae405f84a3f2285c81eda4,b39e13ef388095a23162b4dd77296957,4d2a84036e809d1729442ba3af0a87d4,c1e5d70fd4fe00630e1c2b16cd68eb97,720d5d1551df60d9c6d1d944a0a01760,29d038f2a84049e5c31f1a530e11af06,a8c24bc9e36970494173dcef764825fe,cf571fd425e05ef661a24891234436d2\",\"confirmed\":\"y\",\"disabled\":\"n\",\"disabled_reason\":\"\",\"fingerprint\":\"c1167a8fed52653bc57fcf2a8dc990e9\",\"login_date_time\":\"2015-12-18T21:10:37-07:00\",\"login_fail_time\":\"\",\"n_failed_login\":\"0\",\"num_login_times\":\"11\",\"privs\":\"user\",\"register_date_time\":\"2015-12-18T21:10:37-07:00\",\"salt\":\"18a281c26d14f1291604f522328da97e\",\"v\":\"3b127b41533578ba1dc7d757044623fa83c9e612f8b6a620f75bee16ab00cca8968b6d246252027da02541114b2e87da4e52f1da8c6550e586ff13c4e49c6da71228a21b280e7253e90f8d54c505dd35194d6931f8e6abe881a7e73c0743a3972baee91e321a9249ce573758570f54f1e2639b51d5d2e9862c8406e50be6255613026c5fd2b6c0f4bfdb7b5c5a80ea5d0ad1e375f3677778e6b79d2b83f4642ef73ef67c7f62bd226ca36ff07c349a64e0bb049d1a268fc040ed96b092a5118b6565a10af8d2b53200625c988f65f8d07501701bad56083d01657eaf58a3319d17de856706c3d550a7fb5923dc2315bcd2ee009af73e0d0101aa35366f100498\"}"
	err = conn.Cmd("SET", aKey+"2", aValue).Err
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

}

/* vim: set noai ts=4 sw=4: */
