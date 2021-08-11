package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
)

func GenSig(digest []byte, priv *rsa.PrivateKey) []byte {
	s, err := rsa.SignPSS(rand.Reader, priv, crypto.SHA512, digest, nil)
	if err != nil {
		panic(err)
	} else {
		return s
	}
}

func VerifySig(digest []byte, sig []byte, pub *rsa.PublicKey) bool {
	return rsa.VerifyPSS(pub, crypto.SHA512, digest, sig, nil) == nil
}

func GenSigObj(obj interface{}, priv *rsa.PrivateKey) []byte {
	return GenSig(GenHash(Ser(obj)), priv)
}

func VerifySigObj(obj interface{}, sig []byte, pub *rsa.PublicKey) bool {
	return VerifySig(GenHash(Ser(obj)), sig, pub)
}
