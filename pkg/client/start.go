package client

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
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

func specResKeyFromR2C(r2c msg.Replica2Client) specResKey {
	return specResKey{
		view:        r2c.SpecRes.View,
		seq:         r2c.SpecRes.Seq,
		historyHash: base64.StdEncoding.EncodeToString(r2c.SpecRes.HistoryHash),
		resHash:     base64.StdEncoding.EncodeToString(r2c.SpecRes.ResHash),
		clientId:    r2c.SpecRes.ClientId,
		timestamp:   r2c.SpecRes.Timestamp,
		result:      base64.StdEncoding.EncodeToString(r2c.Result),
	}
}

func (c *Client) Run() error {
	l, err := net.Listen("tcp", conf.GetListenAddr(c.id))
	if err != nil {
		return err
	}

	spec := make(chan msg.Replica2Client, conf.N)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				panic(err)
			}

			c.listenMu.Lock()
			if !c.listen {
				continue
			}
			c.listenMu.Unlock()

			var n int32
			err = binary.Read(conn, binary.BigEndian, &n)
			if err != nil {
				panic(err)
			}

			b := make([]byte, n)
			_, err = conn.Read(b)
			if err != nil {
				panic(err)
			}

			var m struct {
				T int
			}
			err = json.Unmarshal(b, &m)
			if err != nil {
				panic(err)
			}

			t := m.T
			switch t {
			case msg.TypeR2c:
				log.Println("Got r2c")

				var r2c msg.Replica2Client
				err = json.Unmarshal(b, &r2c)
				if err != nil {
					panic(err)
				}

				err := msg.VerifySig(r2c, []*rsa.PublicKey{conf.Pub[r2c.ServerId], conf.Pub[c.primary]})
				if err != nil {
					continue
				}

				spec <- r2c
			default:
				panic(errors.New("unknown msg type"))
			}
		}
	}()

	for {
		time.Sleep(1 * time.Second)

		in, err := conf.GetRandInput()
		if err != nil {
			return err
		}

		r := msg.Request{
			Data:      in,
			Timestamp: time.Now().UnixNano(),
			ClientId:  c.id,
		}

		rb, err := msg.Serialize(r)
		if err != nil {
			return err
		}

		rd, err := utils.GenHash(rb)
		if err != nil {
			return err
		}

		rs, err := utils.GenSig(rd, conf.Priv[c.id])
		if err != nil {
			return err
		}

		c2p := msg.Client2Primary{
			T:      msg.TypeC2p,
			Req:    r,
			ReqSig: rs,
		}

		c2pb, err := msg.Serialize(c2p)
		if err != nil {
			return err
		}

		conn, err := net.Dial("tcp", conf.GetReqAddr(c.primary))
		if err != nil {
			return err
		}

		_, err = bufio.NewReader(c2pb).WriteTo(conn)
		if err != nil {
			return err
		}

		c.listenMu.Lock()
		c.listen = true
		c.listenMu.Unlock()

		srKeyMap := make(map[specResKey]struct {
			n   int
			r2c msg.Replica2Client
		})

		func() {
			ctx, cancel := context.WithTimeout(context.Background(), conf.ClientTimeout)
			defer cancel()

			for {
				select {
				case <-ctx.Done():
					return
				case r2c := <-spec:
					k := specResKeyFromR2C(r2c)
					v := srKeyMap[k]
					v.n += 1
					v.r2c = r2c
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
		var maxR2c msg.Replica2Client
		for _, v := range srKeyMap {
			if v.n > maxN {
				maxN = v.n
				maxR2c = v.r2c
			}
		}
		if maxN >= 3*conf.F+1 {
			out := maxR2c.Result
			ok, err := utils.VerifyHash(out, bytes.NewReader(in))
			if err != nil {
				panic(err)
			}

			if ok {
				log.Println("OK")
			} else {
				log.Fatalln("Failed")
			}
		} else {
			log.Printf("Not complete: %d", maxN)
		}
	}
}
