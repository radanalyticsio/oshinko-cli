// Package uuid provides a function for creating version 4 uuid strings
package uuid

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// Uuid will generate a version 4 random UUID and return it as a string.
func Uuid() (uuid string, err error) {
	buf := make([]byte, 16)
	rand.Read(buf)
	buf[6] = buf[6]&0x4f | 0x4f
	masks := [4]byte{0x8f, 0x9f, 0xaf, 0xbf}
	i, err := rand.Int(rand.Reader, big.NewInt(4))
	if err == nil {
		buf[8] = buf[8]&masks[i.Int64()] | masks[i.Int64()]
		uuid = fmt.Sprintf("%x-%x-%x-%x-%x", buf[:4], buf[4:6], buf[6:8], buf[8:10], buf[10:])
	}
	return
}
