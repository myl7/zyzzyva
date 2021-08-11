package server

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"github.com/myl7/zyzzyva/pkg/comm"
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
	history     []msg.Req
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
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		s.Listen()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		s.ListenMulticast()
	}()

	wg.Wait()
}

func (s *Server) Listen() {
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
		case msg.TypeReq:
			log.Println("Got rm")

			var rm msg.ReqMsg
			err = json.Unmarshal(b, &rm)
			if err != nil {
				panic(err)
			}

			s.handleReq(rm)
		default:
			panic(errors.New("unknown msg type"))
		}
	}
}

func (s *Server) ListenMulticast() {
	l := comm.ListenMulticastUdp()
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
		case msg.TypeOrderReq:
			log.Println("Got orm")

			var orm msg.OrderReqMsg
			err = json.Unmarshal(b, &orm)
			if err != nil {
				panic(err)
			}

			s.handleOrderReq(orm)
		default:
			panic(errors.New("unknown msg type"))
		}
	}
}

func (s *Server) handleReq(rm msg.ReqMsg) {
	if !msg.VerifySig(rm, []*rsa.PublicKey{conf.Pub[rm.Req.CId]}) {
		return
	}

	if c, ok := s.respCache[rm.Req.CId]; ok && c.timestamp >= rm.Req.Timestamp {
		return
	} else {
		s.respCache[rm.Req.CId] = struct {
			state     int
			timestamp int64
		}{timestamp: rm.Req.Timestamp}
	}

	seq := s.s.nextSeq
	s.s.nextSeq += 1
	s.s.history = append(s.s.history, rm.Req)

	r := rm.Req
	rs := rm.ReqSig
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
	orm := msg.OrderReqMsg{
		T:           msg.TypeOrderReq,
		OrderReq:    or,
		OrderReqSig: ors,
		Req:         r,
		ReqSig:      rs,
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		rep := utils.GenHash(orm.Req.Data)
		repd := utils.GenHash(rep)
		sr := msg.SpecRes{
			View:        s.s.view,
			Seq:         or.Seq,
			HistoryHash: s.s.historyHash.Sum(nil),
			ResHash:     repd,
			CId:         r.CId,
			Timestamp:   r.Timestamp,
		}
		srs := utils.GenSigObj(sr, conf.Priv[s.id])
		srm := msg.SpecResMsg{
			T:           msg.TypeSpecRes,
			SpecRes:     sr,
			SpecResSig:  srs,
			SId:         s.id,
			Reply:       rep,
			OrderReq:    or,
			OrderReqSig: ors,
		}

		comm.UdpSendObj(srm, r.CId)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		comm.UdpMulticastObj(orm)
	}()

	wg.Wait()
}

func (s *Server) handleOrderReq(orm msg.OrderReqMsg) {
	if !msg.VerifySig(orm, []*rsa.PublicKey{conf.Pub[s.s.view%conf.N], conf.Pub[orm.Req.CId]}) {
		return
	}

	r := orm.Req
	or := orm.OrderReq
	ors := orm.OrderReqSig
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

	rep := utils.GenHash(orm.Req.Data)
	repd := utils.GenHash(rep)
	sr := msg.SpecRes{
		View:        s.s.view,
		Seq:         seq,
		HistoryHash: s.s.historyHash.Sum(nil),
		ResHash:     repd,
		CId:         r.CId,
		Timestamp:   r.Timestamp,
	}
	srs := utils.GenSigObj(sr, conf.Priv[s.id])
	srm := msg.SpecResMsg{
		T:           msg.TypeSpecRes,
		SpecRes:     sr,
		SpecResSig:  srs,
		SId:         s.id,
		Reply:       rep,
		OrderReq:    or,
		OrderReqSig: ors,
	}

	comm.UdpSendObj(srm, r.CId)
}
