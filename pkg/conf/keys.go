package conf

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

var Priv []*rsa.PrivateKey
var Pub []*rsa.PublicKey

var KeyUrl = "https://share.myl.moe/?/zyzzyva/keys/"

func InitKeys(id int) {
	Priv = make([]*rsa.PrivateKey, N+M)

	r, err := http.Get(KeyUrl + fmt.Sprintf("%d.txt", id))
	if err != nil {
		panic(err)
	}

	rb, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	rs := string(rb)
	rs = strings.TrimSpace(rs)
	b, err := base64.StdEncoding.DecodeString(rs)
	if err != nil {
		panic(err)
	}

	priv, err := x509.ParsePKCS1PrivateKey(b)
	if err != nil {
		panic(err)
	}

	Priv[id] = priv

	r, err = http.Get(KeyUrl + "pub.txt")
	if err != nil {
		panic(err)
	}

	rb, err = ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	rs = string(rb)
	lines := strings.Split(rs, "\n")
	if len(lines) < N+M {
		panic(errors.New("pubkey not enough"))
	}

	for i := 0; i < N+M; i++ {
		line := strings.TrimSpace(lines[i])
		b, err := base64.StdEncoding.DecodeString(line)
		if err != nil {
			panic(err)
		}

		pub, err := x509.ParsePKCS1PublicKey(b)
		if err != nil {
			panic(err)
		}

		Pub = append(Pub, pub)
	}
}
