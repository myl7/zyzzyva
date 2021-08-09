package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

func Serialize(obj interface{}) []byte {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, int32(len(b)))
	if err != nil {
		panic(err)
	}

	_, err = buf.Write(b)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}
