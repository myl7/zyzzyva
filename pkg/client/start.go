package client

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/myl7/zyzzyva/pkg/comm"
	"github.com/myl7/zyzzyva/pkg/conf"
	"github.com/myl7/zyzzyva/pkg/msg"
	"github.com/myl7/zyzzyva/pkg/utils"
	"log"
	"net"
	"sync"
	"time"
)

type Client struct {
	id       int
	primary  int
	listen   bool
	listenMu sync.Mutex
}

func NewClient(id int) *Client {
	return &Client{
		id: id,
	}
}

type specResKey struct {
	view        int
	seq         int
	historyHash string
	resHash     string
	clientId    int
	timestamp   int64
	result      string
}

func specRes2Key(srm msg.SpecResMsg) specResKey {
	return specResKey{
		view:        srm.SpecRes.View,
		seq:         srm.SpecRes.Seq,
		historyHash: base64.StdEncoding.EncodeToString(srm.SpecRes.HistoryHash),
		resHash:     base64.StdEncoding.EncodeToString(srm.SpecRes.ResHash),
		clientId:    srm.SpecRes.CId,
		timestamp:   srm.SpecRes.Timestamp,
		result:      base64.StdEncoding.EncodeToString(srm.Reply),
	}
}

func (c *Client) Run() {
	spec := make(chan msg.SpecResMsg, conf.N)

	go c.Listen(spec)

	time.Sleep(1 * time.Minute)

	for {
		time.Sleep(1 * time.Second)

		log.Println("Start")

		in, err := conf.GetRandInput()
		if err != nil {
			panic(err)
		}

		r := msg.Req{
			Data:      in,
			Timestamp: time.Now().UnixNano(),
			CId:       c.id,
		}
		rs := utils.GenSigObj(r, conf.Priv[c.id])
		rm := msg.ReqMsg{
			T:      msg.TypeReq,
			Req:    r,
			ReqSig: rs,
		}

		comm.UdpSendObj(rm, c.primary)

		c.listenMu.Lock()
		c.listen = true
		c.listenMu.Unlock()

		srKeyMap := make(map[specResKey]struct {
			n       int
			sidList []int
			sigList [][]byte
			sr      msg.SpecRes
			reply   []byte
		})

		func() {
			ctx, cancel := context.WithTimeout(context.Background(), conf.ClientTimeout)
			defer cancel()

			for {
				select {
				case <-ctx.Done():
					return
				case srm := <-spec:
					k := specRes2Key(srm)
					v := srKeyMap[k]
					v.n += 1
					v.sidList = append(v.sidList, srm.SId)
					v.sigList = append(v.sigList, srm.SpecResSig)
					v.sr = srm.SpecRes
					v.reply = srm.Reply
					srKeyMap[k] = v

					if v.n >= 3*conf.F+1 {
						return
					}
				}
			}
		}()

		c.listenMu.Lock()
		c.listen = false
		c.listenMu.Unlock()

		maxN := 0
		// var maxSr msg.SpecRes
		// var sigs [][]byte
		// var sids []int
		var reply []byte
		for _, v := range srKeyMap {
			if v.n > maxN {
				maxN = v.n
				// maxSr = v.sr
				// sigs = v.sigList
				// sids = v.sidList
				reply = v.reply
			}
		}
		if maxN >= 3*conf.F+1 {
			if utils.VerifyHash(reply, in) {
				log.Println("OK")
			} else {
				log.Fatalln("Failed")
			}
		} else if maxN >= 2*conf.F+1 {
			log.Println("Requires commit")
			if utils.VerifyHash(reply, in) {
				log.Println("OK")
			} else {
				log.Fatalln("Failed")
			}
			// cc := msg.CC{
			// 	SpecRes: maxSr,
			// 	SIdList: sids,
			// 	SigList: sigs,
			// }
			// commit := msg.Commit{
			// 	CId: c.id,
			// 	CC:  cc,
			// }
			// cm := msg.CommitMsg{
			// 	T:         msg.TypeCommit,
			// 	Commit:    commit,
			// 	CommitSig: utils.GenSigObj(c, conf.Priv[c.id]),
			// }
			// comm.UdpMulticastObj(cm)
		} else {
			log.Printf("Not complete: %d", maxN)
		}
	}
}

func (c *Client) Listen(spec chan<- msg.SpecResMsg) {
	l, err := net.ListenPacket("udp", conf.GetListenAddr(c.id))
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 1*1024*1024)

	for {
		n, _, err := l.ReadFrom(buf)
		if err != nil {
			panic(err)
		}

		c.listenMu.Lock()
		if !c.listen {
			continue
		}
		c.listenMu.Unlock()

		b := buf[:n]

		var m struct {
			T msg.Type
		}
		err = json.Unmarshal(b, &m)
		if err != nil {
			panic(err)
		}

		t := m.T
		switch t {
		case msg.TypeSpecRes:
			log.Println("Got srm")

			var srm msg.SpecResMsg
			err = json.Unmarshal(b, &srm)
			if err != nil {
				panic(err)
			}

			if !msg.VerifySig(srm, []*rsa.PublicKey{conf.Pub[srm.SId], conf.Pub[c.primary]}) {
				continue
			}

			spec <- srm
		default:
			panic(errors.New("unknown msg type"))
		}
	}
}
