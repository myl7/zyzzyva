package conf

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var F = 1
var N = 3*F + 1
var M = 1

//goland:noinspection GoUnusedGlobalVariable
var KeySize = 2048
var ClientTimeout = 10 * time.Second

//goland:noinspection GoUnusedGlobalVariable
var CPInterval = 100

var Extra = []byte("extra")

var UdpMulticastAddr = "224.0.0.1:10001"
var UdpMulticastInterface = "enp5s0"

var IpPrefix = os.Getenv("IP_PREFIX")

func GetReqAddr(id int) string {
	i := strings.LastIndex(IpPrefix, ".")
	if i == -1 {
		panic(errors.New("invalid ip prefix"))
	}

	last, err := strconv.Atoi(IpPrefix[i+1:])
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%s.%d:10000", IpPrefix[:i], id+last)
}

func GetListenAddr(_ int) string {
	return ":10000"
}

var RandInputSize = 64

func GetRandInput() ([]byte, error) {
	in := make([]byte, RandInputSize)
	_, err := rand.Read(in)
	if err != nil {
		return nil, err
	}

	return in, nil
}
