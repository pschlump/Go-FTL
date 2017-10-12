package main

// From:  http://blog.giorgis.io/golang-aes-encryption
// http://stackoverflow.com/questions/14400729/cryptojs-aes-and-golang

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	tr "github.com/pschlump/godebug"
	"github.com/pschlump/json"   //	"encoding/json"
	"golang.org/x/crypto/pbkdf2" // https://github.com/golang/crypto/blob/master/pbkdf2/pbkdf2.go
)

// https://golang.org/src/crypto/cipher/gcm_test.go -- GCM example
// https://leanpub.com/gocrypto/read

// add CLI -
// -i input
// -o output
// -e Encrypt (plain text -> encrypted)
// -d Decrypt
// -p Password

var opts struct {
	Input    string `short:"i" long:"input"       description:"input file name"        default:""`
	Output   string `short:"o" long:"output"      description:"output fiel name"       default:""`
	Encrypt  bool   `short:"e" long:"encrypt"     description:"encrypt data"           default:"false"`
	Decrypt  bool   `short:"d" long:"decrypt"     description:"decrypt data"           default:"false"`
	Password string `short:"p" long:"password"    description:"password"               default:""`
	IV       string `short:"v" long:"iv"          description:"initialization vector"  default:""`
	Salt     string `short:"s" long:"salt"        description:"salt"                   default:""`
	NIter    int    `short:"n" long:"niter"       description:"number of iterations"   default:"1000"` // iter
	KeySize  int    `short:"k" long:"ks"          description:"keysize"                default:"256"`  // ks
}

// {"iv":"ClgRPpXdUl1v18x1MKHUVg==",
// "v":1,
// "iter":1000,
// "ks":256,
// "ts":64,
// "mode":"ccm",
// "adata":"",
// "cipher":"aes",
// "salt":"VX97AWo5WMA=",
// "ct":"RBp5FLMykk+mRZEWi15CcA=="}

type CypherTextStruct struct {
	Iv     string `json:"iv"`     // Initialization vector for AES (nounce)
	V      int    `json:"v"`      // Version == 1
	Iter   int    `json:"iter"`   // pbkdf2 number of iterations
	Ks     int    `json:"ks"`     // Key Size
	Ts     int    `json:"ts"`     // Authentication-Tag - Authenticated Strength == 64
	Mode   string `json:"mode"`   // ccm - not implemented yet
	AData  string `json:"adata"`  // Authenticated Data
	Cypher string `json:"cypher"` // "aes"
	Salt   string `json:"salt"`   // password salt - for pbkdf2
	Ct     string `json:"ct"`     // CypherText - the encrypted result
}

// this is .Ks/8 - should change
func KeySize_to_Bytes(ks int) (b_ks int) {
	switch ks {
	case 256:
		return 32
	case 192:
		return 24
	case 128:
		return 16
	}
	return 32
}

func main() {

	var Salt []byte

	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		fmt.Printf("Invalid Command Line: %s\n", err)
		os.Exit(1)
	}

	if !opts.Encrypt && !opts.Decrypt {
		fmt.Printf("Musth either encrypt -e or decrypt -d\n")
		os.Exit(1)
	}

	Salt, err = base64.StdEncoding.DecodeString(opts.Salt)
	if err != nil {
		fmt.Printf("Unable to decode Salt, should be base64 encoded, %s\n", err)
		os.Exit(1)
	}

	if false {
		/*

		   # Password: "aaa"
		   # Key: D74C5574 35F891C7 0AA51E10 2C3E6D46 2E3E4B65 219C6705 8DB9E2BC 7D1C1B67
		   # Salt: 513BE253 4AB590FA
		   # CCM Initialization Vector: CD8E1F0F 3F67D9AD 7F3F6591 37618755
		   # Plain Text: the quick brown fox jumps over the lazy dog
		   # Authenticated Data: very authentic

		   # {"iv":"zY4fDz9n2a1/P2WRN2GHVQ==",
		   # "v":1,
		   # "iter":1000,
		   # "ks":256,
		   # "ts":64,
		   # "mode":"ccm",
		   # "adata":"dmVyeSBhdXRoZW50aWM=",
		   # "cipher":"aes",
		   # "salt":"UTviU0q1kPo=",
		   # "ct":"zl7ghVA1acPB59Y5omXl7OsZpvcHl2T+TLSp9mpKYztFs36Q1wq192YAW9tzDJx9irH3"}

		*/
		Salt, err = base64.StdEncoding.DecodeString("UTviU0q1kPo=")
		fmt.Printf("Salt in Hex: %x, %s, %s\n", Salt, err, tr.LF())

		IV, err := base64.StdEncoding.DecodeString("zY4fDz9n2a1/P2WRN2GHVQ==")
		fmt.Printf("IV in Hex: %x, %s, %s\n", IV, err, tr.LF())
	}

	// ---------------------------------------------------- good so far --------------------------------------------------
	// ---------------------------------------------------- good so far --------------------------------------------------
	// ---------------------------------------------------- good so far --------------------------------------------------
	// ---------------------------------------------------- good so far --------------------------------------------------
	// ---------------------------------------------------- good so far --------------------------------------------------
	// ---------------------------------------------------- good so far --------------------------------------------------

	if opts.Encrypt {

		// The key length can be 32, 24, 16  bytes (OR in bits: 128, 192 or 256)
		// key := []byte("longer means more possible keys ") // xyzzy - generate key from password
		key := pbkdf2.Key([]byte(opts.Password), Salt, opts.NIter, KeySize_to_Bytes(opts.KeySize), sha256.New)
		if db1 {
			fmt.Printf("key (in hex) = %x, %s\n", key, tr.LF())
		}

		plaintext, err := ioutil.ReadFile(opts.Input)
		if err != nil {
			fmt.Printf("Unable to read input file: %s\n", err)
			os.Exit(1)
		}
		if ciphertext, iv, err := encrypt(key, plaintext); err != nil {
			log.Fatal(err)
		} else {
			c := &CypherTextStruct{
				Iv:     base64.StdEncoding.EncodeToString(iv),
				V:      1,
				Iter:   opts.NIter,
				Ks:     opts.KeySize,
				Ts:     64,
				Mode:   "ccm",
				AData:  "",
				Cypher: "aes",
				Salt:   opts.Salt,
				Ct:     base64.StdEncoding.EncodeToString(ciphertext),
			}
			fmt.Printf("%s\n", tr.SVarI(c))
		}
	}

	if opts.Decrypt {
		c := &CypherTextStruct{
			Iv:     "",
			V:      1,
			Iter:   opts.NIter,
			Ks:     opts.KeySize,
			Ts:     64,
			Mode:   "ccm",
			AData:  "",
			Cypher: "aes",
			Salt:   "",
			Ct:     "",
		}
		raw, err := ioutil.ReadFile(opts.Input)
		if err != nil {
			fmt.Printf("Unable to read input file: %s\n", err)
			os.Exit(1)
		}
		err = json.Unmarshal(raw, &c)
		if err != nil {
			fmt.Printf("Error(10012): %v, %s, Config File:%s\n", err, tr.LF(), opts.Input)
			os.Exit(1)
		}

		Salt, err = base64.StdEncoding.DecodeString(c.Salt)
		fmt.Printf("Salt in Hex: %x, %s, %s\n", Salt, err, tr.LF())

		IV, err := base64.StdEncoding.DecodeString(c.Iv)
		if err != nil {
			fmt.Printf("Unable to decode initialization vector - from base64: err: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("IV in Hex: %x, %s, %s\n", IV, err, tr.LF())

		// The key length can be 32, 24, 16  bytes (OR in bits: 128, 192 or 256)
		// key := []byte("longer means more possible keys ") // xyzzy - generate key from password
		key := pbkdf2.Key([]byte(opts.Password), Salt, c.Iter, KeySize_to_Bytes(c.Ks), sha256.New)
		if db1 {
			fmt.Printf("key (in hex) = %x, %s\n", key, tr.LF())
		}

		cc, err := base64.StdEncoding.DecodeString(c.Ct)
		if err != nil {
			fmt.Printf("Unable to decode ciphertext - from base64: err: %s\n", err)
			os.Exit(1)
		}

		if plaintext, err := decrypt(key, IV, cc); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("%s\n", plaintext)
		}
	}
}

func encrypt(key, text []byte) (ciphertext []byte, iv []byte, err error) {

	// var ciphertext, plaintext []byte
	var block cipher.Block

	if block, err = aes.NewCipher(key); err != nil {
		return
	}

	ciphertext = make([]byte, aes.BlockSize+len(string(text)))

	// iv =  initialization vector
	iv = ciphertext[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], text)

	return
}

func decrypt(key, iv, ciphertext []byte) (plaintext []byte, err error) {

	var block cipher.Block

	if block, err = aes.NewCipher(key); err != nil {
		return
	}

	if len(ciphertext) < aes.BlockSize {
		err = errors.New("ciphertext too short")
		return
	}

	//iv := ciphertext[:aes.BlockSize]
	//ciphertext = ciphertext[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(ciphertext, ciphertext)

	plaintext = ciphertext

	return
}

// GenerateNonce creates a new random nonce.
func GenerateNonce() (*[NonceSize]byte, error) {
	nonce := new([NonceSize]byte)
	_, err := io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		return nil, err
	}

	return nonce, nil
}

const db1 = true
