//
// Package aessrp implements encrypted authentication and encrypted REST.
// SRP-6a for login authenticaiton, followed by AES 256 bit encrypted RESTful calls.
//
// Copyright (C) Philip Schlump, 2013-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 0001
// 你好无聊的世界
//

package X2fa

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	mathRand "math/rand"
	"os"

	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug" //
)

// Generate a random number, 0..N, returned as a string with 6 to 8 non-zero digits.
func GenRandNumber(nDigits int) (buf string, err error) {

	var n int64
	for {
		binary.Read(rand.Reader, binary.LittleEndian, &n)
		if n < 0 {
			n = -n
		}
		if n > 1000000 {
			break
		}
		// fmt.Printf("Looping GenRandNumber=%d\n", n)
	}
	// fmt.Printf("Big Eenough GenRandNumber=%d\n", n)
	n = n % 100000000
	// fmt.Printf("GenRandNumber=%d\n", n)
	buf = fmt.Sprintf("%08d", n)
	// fmt.Printf("GenRandNumber buf=%s\n", buf)

	return
}

// ============================================================================================================================================
var FirstRequest bool = true
var TimeRemain int
var ThisEpoc int
var LastResut []byte

type RanData struct {
	Status string `json:"status"`
	Value  string `json:"value"`
	TTL    int    `json:"ttl"`
	Epoc   int    `json:"ep"`
}

// ============================================================================================================================================
// Should move to aesccm package
func GenRandBytes(nRandBytes int) (buf []byte, err error) {
	if LocalGen {
		if dbCipher {
			fmt.Printf("AT: %s\n", godebug.LF())
		}
		buf = make([]byte, nRandBytes)
		_, err = rand.Read(buf)
		if err != nil {
			fmt.Printf(`{"msg":"Error generaintg random numbers :%s"}\n`, err)
			return nil, err
		}
		// fmt.Printf("Value: %x\n", buf)
		return
	} else {

		URL := "http://www.2c-why.com/Ran/RandomValue"
		var status int
		var body string

		if FirstRequest {
			ran := fmt.Sprintf("%d", mathRand.Intn(1000000000))
			// status, body := DoGet("http://t432z.com/upd/", "url", hdlr.DisplayURL, "id", qrId, "data", theData, "_ran_", ran)
			status, body = DoGet(URL, "_ran_", ran)
		} else {
			status, body = DoGet(URL, "ep", fmt.Sprintf("%v", ThisEpoc)) // xyzzy Deal with TTL and timing to see if need to re-fetch.
			// xyzzy use TimeRemain, ThisEpoc, LastResult
		}

		if status != 200 {
			fmt.Printf("Unable to get RandomOracle - what to do, status = %v\n", status)
			fmt.Fprintf(os.Stderr, "Unable to get RandomOracle - what to do, status = %v\n", status)
			buf = make([]byte, nRandBytes)
			return
		}

		fmt.Fprintf(os.Stderr, "%sRandomValue%s ->%s<- AT:%s\n", MiscLib.ColorYellow, MiscLib.ColorReset, body, godebug.LF())

		// fmt.Fprintf(www, `{"status":"success","value":"%x","ttl":%d,"ep":%v}`, aValue, ttlCurrent, epoc_120)
		var pd RanData
		err = json.Unmarshal([]byte(body), &pd)
		if pd.Status != "success" {
			fmt.Printf("Unable to get RandomOracle - what to do, status = %v\n", status)
			fmt.Fprintf(os.Stderr, "Unable to get RandomOracle - what to do, status = %v\n", status)
			buf = make([]byte, nRandBytes)
			return
		}

		buf, err = hex.DecodeString(pd.Value)
		if err != nil {
			fmt.Printf("Unable to get RandomOracle - what to do, err = %v\n", err)
			fmt.Fprintf(os.Stderr, "Unable to get RandomOracle - what to do, err = %v\n", err)
			buf = make([]byte, nRandBytes)
			return
		}

		FirstRequest = false
		TimeRemain = pd.TTL
		ThisEpoc = pd.Epoc

		return
	}
}

const LocalGen = false
const dbCipher = false

/* vim: set noai ts=4 sw=4: */
