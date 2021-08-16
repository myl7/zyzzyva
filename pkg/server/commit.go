package server

import (
	"github.com/myl7/zyzzyva/pkg/msg"
	"log"
)

func (s *Server) handleCommit(cm msg.CommitMsg) {
	if !msg.VerifySig(cm, cm.GetAllPub()) {
		log.Println("Failed to verify sig")
		return
	}

	log.Fatalln("Unimplemented")
}
