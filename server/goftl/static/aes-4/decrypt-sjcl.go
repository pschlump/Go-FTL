package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	flags "github.com/jessevdk/go-flags"
	tr "github.com/pschlump/godebug"

	"golang.org/x/crypto/pbkdf2"

	"./base64data" // xyzzy-c
	"./sjcl"
)

var opts struct {
	Input          string `short:"i" long:"input"       description:"input file name"          default:"input.enc"`
	EncDec         string `short:"m" long:"mode"        description:"d=decrypt e=encrypt"      default:"d"`
	AdditionalData string `short:"a" long:"ad"          description:"additional data"          default:""`
	Password       string `short:"p" long:"password"    description:"password for decription"  default:""`
	IV             string `short:"v" long:"iv"          description:"initialization vector"    default:""`
	Salt           string `short:"s" long:"salt"        description:"salt"                     default:""`
	NIter          int    `short:"n" long:"niter"       description:"number of iterations"     default:"1000"` // iter
	KeySize        int    `short:"k" long:"ks"          description:"keysize"                  default:"128"`  // ks
}

func main() {

	fmt.Printf("GCM version\n")

	junk, err := flags.ParseArgs(&opts, os.Args)

	if err != nil || len(junk) != 1 {
		fmt.Printf("Invalid Command Line: %s\n", err)
		os.Exit(1)
	}

	if opts.EncDec == "d" {

		encData := sjcl.ReadSJCL(opts.Input)

		encData.Salt.Debug_hex(db1, "salt")
		encData.InitilizationVector.Debug_hex(db1, "Initilization Vector")

		key := pbkdf2.Key([]byte(opts.Password), encData.Salt, encData.Iter, encData.KeySizeBytes, sha256.New)
		debug_hex("key", key)

		cb, err := aes.NewCipher(key) // var cb cipher.Block
		if err != nil {
			log.Fatal("Error(0001): unable to setup AES:", err)
		}

		// -------------------------------------------------------------------------------------------------------------------------

		nonce, _ := getNonce(encData) // 12 chars constant for nlen

		// authmode, err := aesccm.NewCCM(cb, encData.TagSizeBytes, nlen) // var authmode cipher.AEAD
		// authmode, err := cipher.NewGCM(cb)
		authmode, err := cipher.NewGCMWithNonceSize(cb, encData.TagSizeBytes)
		if err != nil {
			log.Fatal("Error(0002): unable to setup CCM:", err)
		}

		debug("Additional Data", []byte(encData.AdditionalData))

		plaintext, err := authmode.Open(nil, nonce, encData.CipherText, encData.AdditionalData)
		if err != nil {
			log.Fatal("Error(0003): decrypting or authenticating using GCM:", err)
		}
		fmt.Printf("Decrypted Data: %q\n", plaintext)

	} else if opts.EncDec == "e" {

		cc := &sjcl.SJCL_DataStruct{
			// InitilizationVector : "",
			// Salt:                 "",
			// CipherText:           "",
			Version: 0.5.9
			Iter:           opts.NIter,
			KeySize:        opts.KeySize,
			TagSize:        64,
			Mode:           "gcm",
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

		// salt - generate? -S
		Salt, err := base64.StdEncoding.DecodeString(opts.Salt)
		if err != nil {
			fmt.Printf("Unable to decode Salt, should be base64 encoded, %s\n", err)
			os.Exit(1)
		}
		Salt = UpperByte(Salt)
		cc.Salt = Salt
		cc.Salt.Debug_hex(db1, "salt")

		// IV - generate? -V
		IV, err := base64.StdEncoding.DecodeString(opts.IV)
		if err != nil {
			fmt.Printf("Unable to decode initialization vector - from base64: err: %s\n", err)
			os.Exit(1)
		}
		IV = UpperByte(IV)
		fmt.Printf("IV in Hex: %x, %s, %s\n", IV, err, tr.LF())
		cc.InitilizationVector = IV
		cc.InitilizationVector.Debug_hex(db1, "IV")

		// The key length can be 32, 24, 16  bytes (OR in bits: 128, 192 or 256)
		// key := []byte("longer means more possible keys ") // xyzzy - generate key from password
		key := pbkdf2.Key([]byte(opts.Password), Salt, cc.Iter, cc.KeySizeBytes, sha256.New)
		if db1 {
			fmt.Printf("key (in hex) = %x, %s\n", key, tr.LF())
		}
		key = UpperByte(key)

		// nonce
		nlen := cc.TagSizeBytes
		nonce := IV[0:nlen]

		cb, err := aes.NewCipher(key) // var cb cipher.Block
		if err != nil {
			log.Fatal("Error(0011): unable to setup AES:", err)
		}

		// authmode, err := cipher.NewGCM(cb)
		authmode, err := cipher.NewGCMWithNonceSize(cb, cc.TagSizeBytes)
		if err != nil {
			log.Fatal("Error(0012): unable to setup GCM:", err)
		}

		newCipterText := authmode.Seal(nil, nonce, plaintext, ad)

		cc.CipherText = newCipterText

		fmt.Printf("%s\n", tr.SVarI(cc))

	} else {
		fmt.Printf("Invalid -m/--mode must be 'e' or 'd'\n")
		os.Exit(1)
	}

}

// xyzzy-c
func debug(name string, d []byte) {
	if db1 {
		log.Printf("%s: len=%d, 0x%x = %q = %v, %s", name, len(d), d, string(d), base64data.Base64Data(d).Int32Array(), tr.LF(2))
	}
}

// xyzzy-c
func debug_hex(name string, d []byte) {
	if db1 {
		log.Printf("%s: len=%d, 0x%x = %q = %x, %s", name, len(d), d, string(d), base64data.Base64Data(d).Uint32Array(), tr.LF(2))
	}
}

func getNonce(encData sjcl.SJCL_DataStruct) (nonce []byte, nlen int) {
	nonce = []byte(encData.InitilizationVector)
	nonce = nonce[0:encData.TagSizeBytes]
	nlen = encData.TagSizeBytes
	debug("nonce", nonce)
	return
}

func UpperByte(b []byte) (rv []byte) {
	// rv = []byte(strings.ToUpper(string(b)))
	rv = b
	return
}

const db1 = true
