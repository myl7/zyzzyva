package server

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha512"
	"github.com/myl7/zyzzyva/pkg/comm"
	"github.com/myl7/zyzzyva/pkg/conf"
	"github.com/myl7/zyzzyva/pkg/msg"
	"github.com/myl7/zyzzyva/pkg/utils"
)

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
		s.tentativeCP = struct {
			cp   msg.CP
			recv map[int]bool
		}{cp: cp, recv: map[int]bool{s.id: true}}

		go func() {
			cps := utils.GenSigObj(cp, conf.Priv[s.id])
			cpm := msg.CPMsg{
				T:     msg.TypeCP,
				SId:   s.id,
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