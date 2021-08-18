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

var F = envInt("F", intPtr(1))
var N = 3*F + 1
var M = envInt("M", intPtr(1))

func envInt(key string, val *int) int {
	v := os.Getenv(key)
	if v != "" {
		d, err := strconv.Atoi(v)
		if err != nil {
			panic(errors.New(fmt.Sprintf("Env %s is not int", key)))
		}
		return d
	} else if val != nil {
		return *val
	} else {
		panic(errors.New(fmt.Sprintf("Env %s not found", key)))
	}
}

func intPtr(d int) *int {
	return &d
}

//goland:noinspection GoUnusedGlobalVariable
var KeySize = 2048
var ClientTimeout = 10 * time.Second

var CPInterval = 5

var Extra = []byte("extra")

var UdpMulticastAddr = "239.255.0.1:10001"
var UdpMulticastInterfaces = []string{"enp5s0", "p2p1"}
var UdpBufSize = 1 * 1024 * 1024

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
