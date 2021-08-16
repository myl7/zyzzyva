package msg

import "encoding/json"

type Type int

const (
	TypeReq Type = iota
	TypeOrderReq
	TypeSpecRes
	TypeCP
	TypeCommit
)

type Msg struct {
	T Type
}

func DeType(b []byte) Type {
	var m Msg
	err := json.Unmarshal(b, &m)
	if err != nil {
		panic(err)
	}

	return m.T
}
