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

package TabServer2

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"

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
// Should move to aesccm package
func GenRandBytes(nRandBytes int) (buf []byte, err error) {
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
}

const dbCipher = false

/* vim: set noai ts=4 sw=4: */
