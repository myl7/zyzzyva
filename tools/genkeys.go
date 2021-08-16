package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"path"
)

func main() {
	n := flag.Int("n", 0, "Key num to be generated")
	d := flag.String("dir", ".", "Key dir to store output")
	flag.Parse()

	var pubSs []string
	for i := 0; i < *n; i++ {
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}

		pub := &priv.PublicKey
		pubB := x509.MarshalPKCS1PublicKey(pub)
		pubS := base64.StdEncoding.EncodeToString(pubB)
		pubSs = append(pubSs, pubS)
		privB := x509.MarshalPKCS1PrivateKey(priv)
		privS := base64.StdEncoding.EncodeToString(privB)

		f, err := os.Create(path.Join(*d, fmt.Sprintf("%d.txt", i)))
		if err != nil {
			panic(err)
		}

		_, err = fmt.Fprintln(f, privS)
		if err != nil {
			panic(err)
		}

		err = f.Close()
		if err != nil {
			panic(err)
		}
	}

	if len(pubSs) == 0 {
		os.Exit(0)
	}

	f, err := os.Create(path.Join(*d, "pub.txt"))
	if err != nil {
		panic(err)
	}

	for i := range pubSs {
		_, err = fmt.Fprintln(f, pubSs[i])
		if err != nil {
			panic(err)
		}
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}
}
