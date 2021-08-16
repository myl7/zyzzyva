package msg

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
