package utils

import (
	"bufio"
	"bytes"
	"crypto/sha512"
	"io"
)

func GenHash(r io.Reader) ([]byte, error) {
	h := sha512.New()
	_, err := bufio.NewReader(r).WriteTo(h)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func VerifyHash(digest []byte, r io.Reader) (bool, error) {
	d, err := GenHash(r)
	if err != nil {
		return false, err
	}

	return bytes.Equal(digest, d), nil
}
