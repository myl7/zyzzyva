package server

import (
	"crypto/rsa"
	"crypto/sha512"
	"github.com/myl7/zyzzyva/pkg/comm"
	"github.com/myl7/zyzzyva/pkg/conf"
	"github.com/myl7/zyzzyva/pkg/msg"
	"github.com/myl7/zyzzyva/pkg/utils"
	"log"
	"sync"
)

func (s *Server) handleReq(rm msg.ReqMsg) {
	if !msg.VerifySig(rm, []*rsa.PublicKey{conf.Pub[rm.Req.CId]}) {
		log.Println("Failed to verify sig")
		return
	}

	if c, ok := s.respCache[rm.Req.CId]; ok && c.timestamp >= rm.Req.Timestamp {
		log.Println("Too early timestamp")
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
