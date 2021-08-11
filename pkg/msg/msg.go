package msg

type Type int

const (
	TypeReq Type = iota
	TypeOrderReq
	TypeSpecRes
	TypeCP
)

type Msg struct {
	T Type
}
