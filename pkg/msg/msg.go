package msg

const (
	TypeC2p int = iota
	TypeP2r
	TypeR2c
)

type Request struct {
	Data      []byte
	Timestamp int64
	ClientId  int
}

type Client2Primary struct {
	T      int
	Req    Request
	ReqSig []byte
}

func (c2p Client2Primary) getObjs() []interface{} {
	return []interface{}{c2p.Req}
}

func (c2p Client2Primary) getSigs() [][]byte {
	return [][]byte{c2p.ReqSig}
}

type OrderReq struct {
	View        int
	Seq         int
	HistoryHash []byte
	ReqHash     []byte
	Extra       []byte
}

type Primary2Replica struct {
	T           int
	OrderReq    OrderReq
	OrderReqSig []byte
	Req         Request
	ReqSig      []byte
}

func (p2r Primary2Replica) getObjs() []interface{} {
	return []interface{}{p2r.OrderReq, p2r.Req}
}

func (p2r Primary2Replica) getSigs() [][]byte {
	return [][]byte{p2r.OrderReqSig, p2r.ReqSig}
}

type SpecResponse struct {
	View        int
	Seq         int
	HistoryHash []byte
	ResHash     []byte
	ClientId    int
	Timestamp   int64
}

type Replica2Client struct {
	T           int
	SpecRes     SpecResponse
	SpecResSig  []byte
	ServerId    int
	Result      []byte
	OrderReq    OrderReq
	OrderReqSig []byte
}

func (r2c Replica2Client) getObjs() []interface{} {
	return []interface{}{r2c.SpecRes, r2c.OrderReq}
}

func (r2c Replica2Client) getSigs() [][]byte {
	return [][]byte{r2c.SpecResSig, r2c.OrderReqSig}
}
