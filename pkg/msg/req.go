package msg

type Req struct {
	Data      []byte
	Timestamp int64
	CId       int
}

type ReqMsg struct {
	T      Type
	Req    Req
	ReqSig []byte
}

func (rm ReqMsg) getAllObj() []interface{} {
	return []interface{}{rm.Req}
}

func (rm ReqMsg) getAllSig() [][]byte {
	return [][]byte{rm.ReqSig}
}
