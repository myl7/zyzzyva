package msg

type CP struct {
	Seq         int
	HistoryHash []byte
	StateHash   []byte
}

type CPMsg struct {
	T     Type
	SId   int
	CP    CP
	CPSig []byte
}

func (cm CPMsg) getAllObj() []interface{} {
	return []interface{}{cm.CP}
}

func (cm CPMsg) getAllSig() [][]byte {
	return [][]byte{cm.CPSig}
}
