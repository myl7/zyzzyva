package utils

import (
	"bytes"
	"crypto/sha512"
)

func GenHash(b []byte) []byte {
	d := sha512.Sum512(b)
	return d[:]
}

func VerifyHash(digest []byte, b []byte) bool {
	return bytes.Equal(digest, GenHash(b))
}

func GenHashObj(obj interface{}) []byte {
	return GenHash(Ser(obj))
}
