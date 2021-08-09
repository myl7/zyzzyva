package server

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"github.com/myl7/zyzzyva/pkg/comu"
	"github.com/myl7/zyzzyva/pkg/conf"
	"github.com/myl7/zyzzyva/pkg/msg"
	"github.com/myl7/zyzzyva/pkg/utils"
	"hash"
	"log"
	"net"
	"sync"
)

type Server struct {
	id          int
	s           state
	committedCP checkpoint
	tentativeCP checkpoint
	respCache   map[int]struct {
		state     int
		timestamp int64
	}
}

type state struct {
	history     []msg.Request
	historyHash hash.Hash
	maxCC       int
	view        int
	nextSeq     int
}

type checkpoint struct {
	seq   int
	state state
}

func NewServer(id int) *Server {
	s := state{
		historyHash: sha512.New(),
	}
	return &Server{
		id: id,
		s:  s,
		respCache: make(map[int]struct {
			state     int
			timestamp int64
		}),
	}
}

func (s *Server) Run() {
	l, err := net.ListenPacket("udp", conf.GetListenAddr(s.id))
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 1*1024*1024)

	for {
		n, _, err := l.ReadFrom(buf)
		if err != nil {
			panic(err)
		}

		b := buf[:n]

		var m struct {
			T int
		}
		err = json.Unmarshal(b, &m)
		if err != nil {
			panic(err)
		}

		t := m.T
		switch t {
		case msg.TypeC2p:
			log.Println("Got c2p")

			var c2p msg.Client2Primary
			err = json.Unmarshal(b, &c2p)
			if err != nil {
				panic(err)
			}

			s.handleC2p(c2p)
		case msg.TypeP2r:
			log.Println("Got p2r")

			var p2r msg.Primary2Replica
			err = json.Unmarshal(b, &p2r)
			if err != nil {
				panic(err)
			}

			s.handleP2r(p2r)
		default:
			panic(errors.New("unknown msg type"))
		}
	}
}

func (s *Server) handleC2p(c2p msg.Client2Primary) {
	if !msg.VerifySig(c2p, []*rsa.PublicKey{conf.Pub[c2p.Req.ClientId]}) {
		return
	}

	if c, ok := s.respCache[c2p.Req.ClientId]; ok && c.timestamp >= c2p.Req.Timestamp {
		return
	} else {
		s.respCache[c2p.Req.ClientId] = struct {
			state     int
			timestamp int64
		}{timestamp: c2p.Req.Timestamp}
	}

	seq := s.s.nextSeq
	s.s.nextSeq += 1
	s.s.history = append(s.s.history, c2p.Req)

	r := c2p.Req
	rs := c2p.ReqSig
	rd := utils.GenHashObj(r)

	s.s.historyHash.Write(rd)

	or := msg.OrderReq{
		View:        s.s.view,
		Seq:         seq,
		HistoryHash: s.s.historyHash.Sum(nil),
		ReqHash:     rd,
		Extra:       conf.Extra,
	}
	ors := utils.GenSigObj(or, conf.Priv[s.id])
	p2r := msg.Primary2Replica{
		T:           msg.TypeP2r,
		OrderReq:    or,
		OrderReqSig: ors,
		Req:         r,
		ReqSig:      rs,
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		res := utils.GenHash(p2r.Req.Data)
		resd := utils.GenHash(res)
		sr := msg.SpecResponse{
			View:        s.s.view,
			Seq:         or.Seq,
			HistoryHash: s.s.historyHash.Sum(nil),
			ResHash:     resd,
			ClientId:    r.ClientId,
			Timestamp:   r.Timestamp,
		}
		srs := utils.GenSigObj(sr, conf.Priv[s.id])
		r2c := msg.Replica2Client{
			T:           msg.TypeR2c,
			SpecRes:     sr,
			SpecResSig:  srs,
			ServerId:    s.id,
			Result:      res,
			OrderReq:    or,
			OrderReqSig: ors,
		}

		comu.UdpSendObj(r2c, r.ClientId)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		comu.UdpBroadcastObj(p2r)
	}()

	wg.Wait()
}

func (s *Server) handleP2r(p2r msg.Primary2Replica) {
	if !msg.VerifySig(p2r, []*rsa.PublicKey{conf.Pub[s.s.view%conf.N], conf.Pub[p2r.Req.ClientId]}) {
		return
	}

	r := p2r.Req
	or := p2r.OrderReq
	ors := p2r.OrderReqSig
	rd := utils.GenHashObj(r)

	if !bytes.Equal(rd, or.ReqHash) {
		return
	}

	if or.Seq != s.s.nextSeq {
		return
	}

	hh := s.s.historyHash
	hh.Write(rd)
	if !bytes.Equal(hh.Sum(nil), or.HistoryHash) {
		return
	}

	s.s.history = append(s.s.history, r)
	s.s.historyHash = hh
	seq := s.s.nextSeq
	s.s.nextSeq += 1

	res := utils.GenHash(p2r.Req.Data)
	resd := utils.GenHash(res)
	sr := msg.SpecResponse{
		View:        s.s.view,
		Seq:         seq,
		HistoryHash: s.s.historyHash.Sum(nil),
		ResHash:     resd,
		ClientId:    r.ClientId,
		Timestamp:   r.Timestamp,
	}
	srs := utils.GenSigObj(sr, conf.Priv[s.id])
	r2c := msg.Replica2Client{
		T:           msg.TypeR2c,
		SpecRes:     sr,
		SpecResSig:  srs,
		ServerId:    s.id,
		Result:      res,
		OrderReq:    or,
		OrderReqSig: ors,
	}

	comu.UdpSendObj(r2c, r.ClientId)
}
