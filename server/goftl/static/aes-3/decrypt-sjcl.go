package main

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/pschlump/Go-FTL/server/lib"
	tr "github.com/pschlump/godebug"

	"github.com/pschlump/AesCCM"
	"github.com/pschlump/AesCCM/base64data"
	"github.com/pschlump/AesCCM/sjcl"
	"golang.org/x/crypto/pbkdf2"
)

const sjcl_version = "1.0"

// const sjcl_version = "1.1"

var opts struct {
	Input          string `short:"i" long:"input"       description:"input file name"          default:"input.enc"`
	Output         string `short:"o" long:"output"      description:"output file name"         default:""`
	EncDec         string `short:"m" long:"mode"        description:"d=decrypt e=encrypt"      default:"d"`
	AdditionalData string `short:"a" long:"ad"          description:"additional data"          default:""`
	Password       string `short:"p" long:"password"    description:"password for decription"  default:""`
	IV             string `short:"v" long:"iv"          description:"initialization vector"    default:""`
	Salt           string `short:"s" long:"salt"        description:"salt"                     default:""`
	NIter          int    `short:"n" long:"niter"       description:"number of iterations"     default:"1000"` // iter
	KeySize        int    `short:"k" long:"ks"          description:"keysize"                  default:"128"`  // ks
}

func main() {

	junk, err := flags.ParseArgs(&opts, os.Args)

	if err != nil || len(junk) != 1 {
		fmt.Printf("Invalid Command Line: %s\n", err)
		os.Exit(1)
	}

	if opts.EncDec == "d" || opts.EncDec == "dec" {

		encData := sjcl.ReadSJCL(opts.Input)

		encData.Salt.Debug_hex(db1, "salt")
		encData.InitilizationVector.Debug_hex(db1, "Initilization Vector")

		fmt.Printf("password[%s] salt[%x] iter[%d] keysize[%d]\n", opts.Password, encData.Salt, encData.Iter, encData.KeySizeBytes)
		key := pbkdf2.Key([]byte(opts.Password), encData.Salt, encData.Iter, encData.KeySizeBytes, sha256.New)
		debug_hex("key", key)

		cb, err := aes.NewCipher(key) // var cb cipher.Block
		if err != nil {
			log.Fatal("Error(0001): unable to setup AES:", err)
		}

		nonce, nlen := getNonce(encData)

		authmode, err := aesccm.NewCCM(cb, encData.TagSizeBytes, nlen) // var authmode cipher.AEAD
		if err != nil {
			log.Fatal("Error(0002): unable to setup CCM:", err)
		}

		debug("Additional Data", []byte(encData.AdditionalData))

		plaintext, err := authmode.Open(nil, nonce, encData.CipherText, encData.AdditionalData)
		if err != nil {
			log.Fatal("Error(0003): decrypting or authenticating using CCM:", err)
		}
		fmt.Printf("Decrypted Data: %q\n", plaintext)

		if opts.Output != "" {
			ioutil.WriteFile(opts.Output, plaintext, 0600)
		}

	} else if opts.EncDec == "e" || opts.EncDec == "enc" {

		cc := &sjcl.SJCL_DataStruct{
			// InitilizationVector : "",
			// Salt:                 "",
			// CipherText:           "",
			Version: 0.5.9
			Iter:           opts.NIter,
			KeySize:        opts.KeySize,
			TagSize:        64,
			Mode:           "ccm",
			AdditionalData: []byte(opts.AdditionalData),
			Cipher:         "aes",
			TagSizeBytes:   8,
			KeySizeBytes:   opts.KeySize / 8,
		}

		plaintext, err := ioutil.ReadFile(opts.Input)
		if err != nil {
			log.Fatal("Reading:", err)
		}
		ad := []byte(opts.AdditionalData)

		var Salt, IV []byte

		if opts.Salt != "" {
			// salt - generate? -S
			Salt, err = base64.StdEncoding.DecodeString(opts.Salt)
			if err != nil {
				fmt.Printf("Unable to decode Salt, should be base64 encoded, %s\n", err)
				os.Exit(1)
			}
		} else {
			Salt, err = GenRandBytes(8)
		}
		cc.Salt = Salt
		cc.Salt.Debug_hex(db1, "salt")

		if opts.IV != "" {
			// IV - generate? -V
			IV, err = base64.StdEncoding.DecodeString(opts.IV)
			if err != nil {
				fmt.Printf("Unable to decode initialization vector - from base64: err: %s\n", err)
				os.Exit(1)
			}
		} else {
			IV, err = GenRandBytes(16)
		}
		fmt.Printf("IV in Hex: %x, %s, %s\n", IV, err, tr.LF())
		cc.InitilizationVector = IV
		cc.InitilizationVector.Debug_hex(db1, "IV")

		// The key length can be 32, 24, 16  bytes (OR in bits: 128, 192 or 256)
		// key := []byte("longer means more possible keys ") // xyzzy - generate key from password
		fmt.Printf("password[%s] salt[%x] iter[%d] keysize[%d]\n", opts.Password, Salt, cc.Iter, cc.KeySizeBytes)
		key := pbkdf2.Key([]byte(opts.Password), Salt, cc.Iter, cc.KeySizeBytes, sha256.New)
		debug_hex("key", key)
		//if db1 {
		//	fmt.Printf("key (in hex) = %x, %s\n", key, tr.LF())
		//}

		// nonce
		// nlen := 13
		// nlen := NonceLength(len(plaintext))
		nlen := aesccm.NonceLengthFromMessageLength(len(plaintext))
		nonce := IV[0:nlen]

		cb, err := aes.NewCipher(key) // var cb cipher.Block
		if err != nil {
			log.Fatal("Error(0011): unable to setup AES:", err)
		}

		authmode, err := aesccm.NewCCM(cb, cc.TagSizeBytes, nlen) // var authmode cipher.AEAD, nlen is len(nonce)
		if err != nil {
			log.Fatal("Error(0012): unable to setup CCM:", err)
		}

		newCipterText := authmode.Seal(nil, nonce, plaintext, ad)

		cc.CipherText = newCipterText

		JSON := lib.SVarI(cc)
		fmt.Printf("%s\n", JSON)

		if opts.Output != "" {
			ioutil.WriteFile(opts.Output, []byte(JSON+"\n"), 0600)
		}

	} else {
		fmt.Printf("Invalid -m/--mode must be 'e' or 'd'\n")
		os.Exit(1)
	}

}

// Should move to aesccm package
func GenRandBytes(nRandBytes int) (buf []byte, err error) {
	buf = make([]byte, nRandBytes)
	_, err = rand.Read(buf)
	if err != nil {
		fmt.Printf("Error generaintg random numbers :%s\n", err)
		return nil, err
	}
	// fmt.Printf("Value: %x\n", buf)
	return
}

func debug(name string, d []byte) {
	if db1 {
		fmt.Printf("%s: length(%s)=%d, %x = %q = %v, %s\n", name, name, len(d), d, string(d), base64data.Base64Data(d).Int32Array(), tr.LF(2))
	}
}

func debug_hex(name string, d []byte) {
	if db1 {
		fmt.Printf("%s: length(%s)=%d, %x = %q = %x, %s\n", name, name, len(d), d, string(d), base64data.Base64Data(d).Uint32Array(), tr.LF(2))
	}
}

// Should move to sjcl sub-package
func getNonce(encData sjcl.SJCL_DataStruct) (nonce []byte, nlen int) {
	nonce = []byte(encData.InitilizationVector)
	if db2 {
		fmt.Printf("tagsize: %d in bytes: %d\n", encData.TagSize, encData.TagSizeBytes)
	}
	nlen = aesccm.MaxNonceLength(len(encData.CipherText) - encData.TagSizeBytes)
	if db2 {
		fmt.Println("max nlen:%d\n", nlen)
	}
	if nlen > len(nonce) {
		nlen = len(nonce)
	} else {
		nonce = nonce[:nlen] // trim nonce to the first X bytes - that is all that is used ( this is SJCL specific )
	}
	if db2 {
		fmt.Println("nlen: %d\n", nlen)
	}
	debug_hex("nonce", nonce)
	return
}

const db1 = true
const db2 = false
