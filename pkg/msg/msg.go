package msg

type Type int

const (
	TypeReq Type = iota
	TypeOrderReq
	TypeSpecRes
)

type Msg struct {
	T Type
}
