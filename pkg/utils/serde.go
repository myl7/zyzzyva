package utils

import (
	"encoding/json"
	"github.com/myl7/zyzzyva/pkg/msg"
)

func Ser(obj interface{}) []byte {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return b
}

func DeType(b []byte) msg.Type {
	var m msg.Msg
	err := json.Unmarshal(b, &m)
	if err != nil {
		panic(err)
	}

	return m.T
}
