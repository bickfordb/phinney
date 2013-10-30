package phinney

import (
	"crypto/rand"
	"fmt"
	"io"
)

func GenUUID() string {
	xs := make([]byte, 16)
	io.ReadFull(rand.Reader, xs)
	return fmt.Sprintf("%x-%x-%x-%x-%x", xs[0:4], xs[4:6], xs[6:8], xs[8:10], xs[10:])
}
