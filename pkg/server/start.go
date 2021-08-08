package server

import (
	"bufio"
	"bytes"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/myl7/zyzzyva/pkg/conf"
	"github.com/myl7/zyzzyva/pkg/msg"
	"github.com/myl7/zyzzyva/pkg/utils"
	"hash"
	"io/ioutil"
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

func (s *Server) Run() error {
	l, err := net.Listen("tcp", conf.GetListenAddr(s.id))
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		var n int32
		err = binary.Read(conn, binary.BigEndian, &n)
		if err != nil {
			return err
		}

		b := make([]byte, n)
		_, err = conn.Read(b)
		if err != nil {
			panic(err)
		}

		log.Printf("Got msg: %d", s.id)

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

			err := s.handleC2p(c2p)
			if err != nil {
				panic(err)
			}
		case msg.TypeP2r:
			log.Println("Got p2r")

			var p2r msg.Primary2Replica
			err = json.Unmarshal(b, &p2r)
			if err != nil {
				panic(err)
			}

			err := s.handleP2r(p2r)
			if err != nil {
				panic(err)
			}
		default:
			panic(errors.New("unknown msg type"))
		}
	}
}

func (s *Server) handleC2p(c2p msg.Client2Primary) error {
	err := msg.VerifySig(c2p, []*rsa.PublicKey{conf.Pub[c2p.Req.ClientId]})
	if err != nil {
		return nil
	}

	if c, ok := s.respCache[c2p.Req.ClientId]; ok && c.timestamp >= c2p.Req.Timestamp {
		return nil
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
	rb, err := msg.Serialize(c2p.Req)
	if err != nil {
		return err
	}

	rd, err := utils.GenHash(rb)
	if err != nil {
		return err
	}

	s.s.historyHash.Write(rd)

	or := msg.OrderReq{
		View:        s.s.view,
		Seq:         seq,
		HistoryHash: s.s.historyHash.Sum(nil),
		ReqHash:     rd,
		Extra:       conf.Extra,
	}
	orb, err := msg.Serialize(or)
	if err != nil {
		return err
	}

	ord, err := utils.GenHash(orb)
	if err != nil {
		return err
	}

	ors, err := utils.GenSig(ord, conf.Priv[s.id])
	if err != nil {
		return err
	}

	p2r := msg.Primary2Replica{
		T:           msg.TypeP2r,
		OrderReq:    or,
		OrderReqSig: ors,
		Req:         r,
		ReqSig:      rs,
	}

	p2rb, err := msg.Serialize(p2r)
	if err != nil {
		return err
	}

	p2rBin, err := ioutil.ReadAll(p2rb)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for i := 0; i < conf.N; i++ {
		if i == s.id {
			wg.Add(1)
			go func() {
				defer wg.Done()

				res, err := utils.GenHash(bytes.NewReader(p2r.Req.Data))
				if err != nil {
					panic(err)
				}

				resd, err := utils.GenHash(bytes.NewReader(res))
				if err != nil {
					panic(err)
				}

				sr := msg.SpecResponse{
					View:        s.s.view,
					Seq:         or.Seq,
					HistoryHash: s.s.historyHash.Sum(nil),
					ResHash:     resd,
					ClientId:    r.ClientId,
					Timestamp:   r.Timestamp,
				}

				srb, err := msg.Serialize(sr)
				if err != nil {
					panic(err)
				}

				srd, err := utils.GenHash(srb)
				if err != nil {
					panic(err)
				}

				srs, err := utils.GenSig(srd, conf.Priv[s.id])
				if err != nil {
					panic(err)
				}

				r2c := msg.Replica2Client{
					T:           msg.TypeR2c,
					SpecRes:     sr,
					SpecResSig:  srs,
					ServerId:    s.id,
					Result:      res,
					OrderReq:    or,
					OrderReqSig: ors,
				}

				r2cb, err := msg.Serialize(r2c)
				if err != nil {
					panic(err)
				}

				conn, err := net.Dial("tcp", conf.GetReqAddr(r.ClientId))
				if err != nil {
					panic(err)
				}

				_, err = bufio.NewReader(r2cb).WriteTo(conn)
				if err != nil {
					panic(err)
				}
			}()
		} else {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				conn, err := net.Dial("tcp", conf.GetReqAddr(i))
				if err != nil {
					panic(err)
				}

				_, err = bytes.NewReader(p2rBin).WriteTo(conn)
				if err != nil {
					panic(err)
				}
			}(i)
		}
	}
	wg.Wait()

	return nil
}

func (s *Server) handleP2r(p2r msg.Primary2Replica) error {
	err := msg.VerifySig(p2r, []*rsa.PublicKey{conf.Pub[s.s.view%conf.N], conf.Pub[p2r.Req.ClientId]})
	if err != nil {
		return nil
	}

	r := p2r.Req
	or := p2r.OrderReq
	ors := p2r.OrderReqSig

	rb, err := msg.Serialize(r)
	if err != nil {
		return err
	}

	rd, err := utils.GenHash(rb)
	if err != nil {
		return err
	}

	if !bytes.Equal(rd, or.ReqHash) {
		return nil
	}

	if or.Seq != s.s.nextSeq {
		return nil
	}

	hh := s.s.historyHash
	hh.Write(rd)
	if !bytes.Equal(hh.Sum(nil), or.HistoryHash) {
		return nil
	}

	s.s.history = append(s.s.history, r)
	s.s.historyHash = hh
	seq := s.s.nextSeq
	s.s.nextSeq += 1

	res, err := utils.GenHash(bytes.NewReader(p2r.Req.Data))
	if err != nil {
		return err
	}

	resd, err := utils.GenHash(bytes.NewReader(res))
	if err != nil {
		return err
	}

	sr := msg.SpecResponse{
		View:        s.s.view,
		Seq:         seq,
		HistoryHash: s.s.historyHash.Sum(nil),
		ResHash:     resd,
		ClientId:    r.ClientId,
		Timestamp:   r.Timestamp,
	}

	srb, err := msg.Serialize(sr)
	if err != nil {
		return err
	}

	srd, err := utils.GenHash(srb)
	if err != nil {
		return err
	}

	srs, err := utils.GenSig(srd, conf.Priv[s.id])
	if err != nil {
		return err
	}

	r2c := msg.Replica2Client{
		T:           msg.TypeR2c,
		SpecRes:     sr,
		SpecResSig:  srs,
		ServerId:    s.id,
		Result:      res,
		OrderReq:    or,
		OrderReqSig: ors,
	}

	r2cb, err := msg.Serialize(r2c)
	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp", conf.GetReqAddr(r.ClientId))
	if err != nil {
		return err
	}

	_, err = bufio.NewReader(r2cb).WriteTo(conn)
	if err != nil {
		return err
	}

	return nil
}
