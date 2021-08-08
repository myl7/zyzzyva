package msg

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
)

func Serialize(obj interface{}) (io.Reader, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, int32(len(b)))
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(b)
	if err != nil {
		return nil, err
	}

	return &buf, err
}
