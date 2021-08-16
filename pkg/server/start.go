package server

import (
	"encoding/json"
	"errors"
	"github.com/myl7/zyzzyva/pkg/comm"
	"github.com/myl7/zyzzyva/pkg/conf"
	"github.com/myl7/zyzzyva/pkg/msg"
	"github.com/myl7/zyzzyva/pkg/utils"
	"log"
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
	tentativeCP   struct {
		cp   msg.CP
		recv map[int]bool
	}
	respCache map[int]struct {
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

		s.listen()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		s.listenMulticast()
	}()

	wg.Wait()
}

func (s *Server) listen() {
	comm.UdpListen(conf.GetListenAddr(s.id), s.handle)
}

func (s *Server) listenMulticast() {
	comm.UdpListenMulticast(s.handle)
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
	case msg.TypeCP:
		log.Println("Got cpm")

		var cpm msg.CPMsg
		err := json.Unmarshal(b, &cpm)
		if err != nil {
			panic(err)
		}

		s.handleCP(cpm)
	case msg.TypeCommit:
		log.Println("Got cm")

		var cm msg.CommitMsg
		err := json.Unmarshal(b, &cm)
		if err != nil {
			panic(err)
		}

		s.handleCommit(cm)
	default:
		panic(errors.New("unknown msg type"))
	}
}
