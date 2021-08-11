package msg

import (
	"crypto/rsa"
	"github.com/myl7/zyzzyva/pkg/utils"
)

type SignedMsg interface {
	getAllObj() []interface{}
	getAllSig() [][]byte
}

func VerifySig(v SignedMsg, pub []*rsa.PublicKey) bool {
	objs := v.getAllObj()
	sigs := v.getAllSig()

	for i := 0; i < len(objs); i++ {
		if !utils.VerifySigObj(objs[i], sigs[i], pub[i]) {
			return false
		}
	}
	return true
}
