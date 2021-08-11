package msg

type OrderReq struct {
	View        int
	Seq         int
	HistoryHash []byte
	ReqHash     []byte
	Extra       []byte
}

type OrderReqMsg struct {
	T           Type
	OrderReq    OrderReq
	OrderReqSig []byte
	Req         Req
	ReqSig      []byte
}

func (orm OrderReqMsg) getAllObj() []interface{} {
	return []interface{}{orm.OrderReq, orm.Req}
}

func (orm OrderReqMsg) getAllSig() [][]byte {
	return [][]byte{orm.OrderReqSig, orm.ReqSig}
}
