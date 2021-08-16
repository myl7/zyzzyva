package utils

import (
	"encoding/json"
)

func Ser(obj interface{}) []byte {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return b
}
