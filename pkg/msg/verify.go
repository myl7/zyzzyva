package msg

import (
	"crypto/rsa"
	"github.com/myl7/zyzzyva/pkg/utils"
)

type Verifiable interface {
	getObjs() []interface{}
	getSigs() [][]byte
}

func VerifySig(v Verifiable, pub []*rsa.PublicKey) bool {
	objs := v.getObjs()
	sigs := v.getSigs()

	for i := 0; i < len(objs); i++ {
		if !utils.VerifySigObj(objs[i], sigs[i], pub[i]) {
			return false
		}
	}
	return true
}
