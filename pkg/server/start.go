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
	"log"
	"net"
	"sync"
)

type Server struct {
	id            int
	history       []msg.Req
	historyHashes [][]byte
	maxCC         int
	view          int
	nextSeq       int
	committedCP   msg.CP
	tentativeCP   msg.CP
	respCache     map[int]struct {
		state     int
		timestamp int64
	}
}

func NewServer(id int) *Server {
	return &Server{
		id: id,
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

		go s.handle(b)
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

		go s.handle(b)
	}
}

func (s *Server) handle(b []byte) {
	t := utils.DeType(b)
	switch t {
	case msg.TypeReq:
		log.Println("Got rm")

		var rm msg.ReqMsg
		err := json.Unmarshal(b, &rm)
		if err != nil {
			panic(err)
		}

		s.handleReq(rm)
	case msg.TypeOrderReq:
		log.Println("Got orm")

		var orm msg.OrderReqMsg
		err := json.Unmarshal(b, &orm)
		if err != nil {
			panic(err)
		}

		s.handleOrderReq(orm)
	default:
		panic(errors.New("unknown msg type"))
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

	seq := s.nextSeq
	s.nextSeq += 1
	s.history = append(s.history, rm.Req)

	r := rm.Req
	rs := rm.ReqSig
	rd := utils.GenHashObj(r)

	hh := sha512.New()
	if s.historyHashes != nil {
		hh.Write(s.historyHashes[len(s.historyHashes)-1])
	}
	hh.Write(rd)
	s.historyHashes = append(s.historyHashes, hh.Sum(nil))

	or := msg.OrderReq{
		View:        s.view,
		Seq:         seq,
		HistoryHash: s.historyHashes[len(s.historyHashes)-1],
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
			View:        s.view,
			Seq:         or.Seq,
			HistoryHash: s.historyHashes[len(s.historyHashes)-1],
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
	if !msg.VerifySig(orm, []*rsa.PublicKey{conf.Pub[s.view%conf.N], conf.Pub[orm.Req.CId]}) {
		return
	}

	r := orm.Req
	or := orm.OrderReq
	ors := orm.OrderReqSig
	rd := utils.GenHashObj(r)

	if !bytes.Equal(rd, or.ReqHash) {
		return
	}

	if or.Seq != s.nextSeq {
		return
	}

	hh := sha512.New()
	if s.historyHashes != nil {
		hh.Write(s.historyHashes[len(s.historyHashes)-1])
	}
	hh.Write(rd)
	if !bytes.Equal(hh.Sum(nil), or.HistoryHash) {
		return
	}

	if len(s.history) >= 2*conf.CPInterval {
		return
	} else if len(s.history) == conf.CPInterval {
		cp := msg.CP{
			Seq:         s.nextSeq,
			HistoryHash: hh.Sum(nil),
			StateHash:   []byte{},
		}
		s.tentativeCP = cp

		go func() {
			cps := utils.GenSigObj(cp, conf.Priv[s.id])
			cpm := msg.CPMsg{
				T:     msg.TypeCP,
				CP:    cp,
				CPSig: cps,
			}

			comm.UdpMulticastObj(cpm)
		}()
	}

	s.history = append(s.history, r)
	s.historyHashes = append(s.historyHashes, hh.Sum(nil))
	seq := s.nextSeq
	s.nextSeq += 1

	rep := utils.GenHash(orm.Req.Data)
	repd := utils.GenHash(rep)
	sr := msg.SpecRes{
		View:        s.view,
		Seq:         seq,
		HistoryHash: s.historyHashes[len(s.historyHashes)-1],
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
