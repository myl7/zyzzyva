package comu

import (
	"github.com/myl7/zyzzyva/pkg/conf"
	"github.com/myl7/zyzzyva/pkg/utils"
	"net"
)

func UdpSend(b []byte, tid int) {
	addr, err := net.ResolveUDPAddr("udp", conf.GetReqAddr(tid))
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}

	err = conn.SetWriteBuffer(1 * 1024 * 1024)
	if err != nil {
		panic(err)
	}

	_, err = conn.Write(b)
	if err != nil {
		panic(err)
	}
}

func UdpMulticast(b []byte) {
	conn, err := net.Dial("udp", conf.UdpMulticastAddr)
	if err != nil {
		panic(err)
	}

	_, err = conn.Write(b)
	if err != nil {
		panic(err)
	}
}

func UdpSendObj(obj interface{}, tid int) {
	UdpSend(utils.Serialize(obj), tid)
}

func UdpMulticastObj(obj interface{}) {
	UdpMulticast(utils.Serialize(obj))
}

func ListenMulticastUdp() *net.UDPConn {
	ifi, err := net.InterfaceByName(conf.UdpMulticastInterface)
	if err != nil {
		//goland:noinspection GoBoolExpressions
		if conf.UdpMulticastInterface == "" {
			ifi = nil
		} else {
			panic(err)
		}
	}

	addr, err := net.ResolveUDPAddr("udp", conf.UdpMulticastAddr)
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenMulticastUDP("udp", ifi, addr)
	if err != nil {
		panic(err)
	}

	err = conn.SetReadBuffer(1 * 1024 * 1024)
	if err != nil {
		panic(err)
	}

	return conn
}
