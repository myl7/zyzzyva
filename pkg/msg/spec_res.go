package msg

type SpecRes struct {
	View        int
	Seq         int
	HistoryHash []byte
	ResHash     []byte
	CId         int
	Timestamp   int64
}

type SpecResMsg struct {
	T           Type
	SpecRes     SpecRes
	SpecResSig  []byte
	SId         int
	Reply       []byte
	OrderReq    OrderReq
	OrderReqSig []byte
}

func (srm SpecResMsg) getAllObj() []interface{} {
	return []interface{}{srm.SpecRes, srm.OrderReq}
}

func (srm SpecResMsg) getAllSig() [][]byte {
	return [][]byte{srm.SpecResSig, srm.OrderReqSig}
}
