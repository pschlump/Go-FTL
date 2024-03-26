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
//

package AesSrp

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/pschlump/godebug" //
	"golang.org/x/crypto/pbkdf2"  // "golang.org/x/crypto/pbkdf2"
)

// Beginning of encrypting salt/v in Redis
// Check www.hdrl.PasswordSV - if not "", then use that to decrypt
// PasswordSV - encpyt before saving
func GetSalt(hdlr *AesSrpType, www http.ResponseWriter, req *http.Request, mdata map[string]string) (salt string, v string) {
	ok := false
	salt, ok = mdata["salt"]
	if !ok {
		fmt.Printf("Error - attempt to get salt/v from map that did not contain it. - AT: %s\n", godebug.LF())
	}
	v, ok = mdata["v"]
	if !ok {
		fmt.Printf("Error - attempt to get salt/v from map that did not contain it. - AT: %s\n", godebug.LF())
	}

	// PasswordSV - encpyt before saving
	PasswordSV := hdlr.PasswordSV
	if PasswordSV == "" {
		fmt.Printf("Should be non-encrypted keys (regular), %s\n", godebug.LF())
	} else {
		fmt.Printf("Shal be encrypted keys, %s\n", godebug.LF())
		db_genKey(hdlr)
		salt = db_decrypt(hdlr.passwordSVKey, salt)
		v = db_decrypt(hdlr.passwordSVKey, v)
	}

	return
}

// Beginning of encrypting salt/v in Redis
// Check www.hdrl.PasswordSV - if not "", then use that to decrypt
// PasswordSV - encpyt before saving
func SetSaltV(hdlr *AesSrpType, www http.ResponseWriter, req *http.Request, mdata map[string]string, salt string, v string) {
	// PasswordSV - encpyt before saving
	PasswordSV := hdlr.PasswordSV
	if PasswordSV == "" {
		fmt.Printf("Should be non-encrypted keys (regular), %s\n", godebug.LF())
	} else {
		fmt.Printf("Shal be encrypted keys, %s\n", godebug.LF())
		db_genKey(hdlr)
		salt = db_encrypt(hdlr.passwordSVKey, salt)
		v = db_encrypt(hdlr.passwordSVKey, v)
	}

	mdata["salt"] = salt
	mdata["v"] = v
}

//package main
//
//import (
//	"crypto/aes"
//	"crypto/cipher"
//	"crypto/rand"
//	"encoding/base64"
//	"fmt"
//	"io"
//)
//
//func main() {
//	originalText := "encrypt this golang"
//	fmt.Println(originalText)
//
//	key := []byte("example key 1234")
//
//	// encrypt value to base64
//	cryptoText := encrypt(key, originalText)
//	fmt.Println(cryptoText)
//
//	// encrypt base64 crypto to original value
//	text := decrypt(key, cryptoText)
//	fmt.Printf(text)
//}

func db_genKey(hdlr *AesSrpType) {
	if hdlr.PasswordSV == "" {
		fmt.Printf("Should be non-encrypted keys (regular), - no key generation - %s\n", godebug.LF())
	} else if len(hdlr.passwordSVKey) == 0 {
		fmt.Printf("Should be encrypted keys, %s\n", godebug.LF())
		Salt := []byte("abcdefghijklmnopqrstuvvwxyz") // Xyzzy - shoud use better salt than this and set it in config
		hdlr.passwordSVKey = pbkdf2.Key([]byte(hdlr.PasswordSV), Salt, 1000, 16, sha256.New)
	}

}

// encrypt string to base64 crypto using AES
func db_encrypt(key []byte, text string) string {

	// key := []byte(keyText)
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	return base64.URLEncoding.EncodeToString(ciphertext)
}

// decrypt from base64 to decrypted string
func db_decrypt(key []byte, cryptoText string) string {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s", ciphertext)
}

/* vim: set noai ts=4 sw=4: */
