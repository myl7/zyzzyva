package server

import (
	"bytes"
	"crypto/rsa"
	"github.com/myl7/zyzzyva/pkg/conf"
	"github.com/myl7/zyzzyva/pkg/msg"
)

func (s *Server) handleCP(cpm msg.CPMsg) {
	if !msg.VerifySig(cpm, []*rsa.PublicKey{conf.Pub[cpm.SId]}) {
		return
	}

	if !bytes.Equal(cpm.CP.HistoryHash, s.tentativeCP.cp.HistoryHash) || cpm.CP.Seq != s.tentativeCP.cp.Seq || !bytes.Equal(cpm.CP.StateHash, []byte{}) {
		return
	}

	s.tentativeCP.recv[cpm.SId] = true

	n := 0
	for _, v := range s.tentativeCP.recv {
		if v {
			n++
		}
	}

	if n >= conf.F+1 {
		s.committedCP = s.tentativeCP.cp

		for i := range s.history {
			if bytes.Equal(s.historyHashes[i], cpm.CP.HistoryHash) {
				s.history = s.history[i+1:]
				s.historyHashes = s.historyHashes[i+1:]
			}
		}
	}
}
