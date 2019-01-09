package Acb1

import (
	"encoding/hex"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/crypto/sha3"
)

// Exists returns true if the file or directory exists in the file system.
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// GetFilenames returns a slize of files and directories for the specified path.
func GetFilenames(dir string) (filenames, dirs []string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, nil
	}
	for _, fstat := range files {
		if !strings.HasPrefix(string(fstat.Name()), ".") {
			if fstat.IsDir() {
				dirs = append(dirs, fstat.Name())
			} else {
				filenames = append(filenames, fstat.Name())
			}
		}
	}
	return
}

// Keccak256 use the Ethereum Keccak hasing fucntions to return a hash from a list of values.
func Keccak256(data ...[]byte) []byte {
	d := sha3.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

// HashStringOf calcualtes the hash of the 'data' and returns it.
func HashStrngOf(data string) (h []byte) {
	h = Keccak256([]byte(data))
	return
}

// HashStringOfReturnHex calcualtes the hash of the 'data' and returns it.
func HashStrngOfReturnHex(data string) (s string) {
	h := Keccak256([]byte(data))
	s = hex.EncodeToString(h)
	return
}
