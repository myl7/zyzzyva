package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
)

func GenSig(digest []byte, priv *rsa.PrivateKey) ([]byte, error) {
	return rsa.SignPSS(rand.Reader, priv, crypto.SHA512, digest, nil)
}

func VerifySig(digest []byte, sig []byte, pub *rsa.PublicKey) error {
	return rsa.VerifyPSS(pub, crypto.SHA512, digest, sig, nil)
}
