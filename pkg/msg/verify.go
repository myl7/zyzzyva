package msg

import (
	"crypto/rsa"
	"github.com/myl7/zyzzyva/pkg/utils"
)

type Verifiable interface {
	getObjs() []interface{}
	getSigs() [][]byte
}

func VerifySig(v Verifiable, pub []*rsa.PublicKey) error {

	objs := v.getObjs()
	sigs := v.getSigs()

	for i := 0; i < len(objs); i++ {
		b, err := Serialize(objs[i])
		if err != nil {
			return err
		}

		d, err := utils.GenHash(b)
		if err != nil {
			return err
		}
		err = utils.VerifySig(d, sigs[i], pub[i])
		if err != nil {
			return err
		}
	}
	return nil
}
