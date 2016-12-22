package GenerateXSRF

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/pschlump/uuid"
)

// GenRandNumber3x Generates a random value that is 0x100000 or larger
func GenRandNumber3x() (buf string) {

	var n int64
	for {
		binary.Read(rand.Reader, binary.LittleEndian, &n)
		if n < 0 {
			n = -n
		}
		if n > 0x100000 {
			break
		}
	}
	n = n & 0xffffff
	// fmt.Printf("GenRandNumber=%d\n", n)
	buf = fmt.Sprintf("%06x", n)
	fmt.Printf("GenRandNumber3x buf=%s\n", buf)

	return
}

func getUUIDAsString() (rv string) {
	id0x, _ := uuid.NewV4()
	rv = id0x.String()
	return
}
