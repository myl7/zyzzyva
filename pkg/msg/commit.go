package msg

import (
	"crypto/rsa"
	"github.com/myl7/zyzzyva/pkg/conf"
)

type CC struct {
	SpecRes SpecRes
	SIdList []int
	SigList [][]byte
}

type Commit struct {
	CId int
	CC  CC
}

type CommitMsg struct {
	T         Type
	Commit    Commit
	CommitSig []byte
}

func (cm CommitMsg) getAllObj() []interface{} {
	sids := cm.Commit.CC.SIdList
	objs := make([]interface{}, 1+len(sids))
	objs[0] = cm.Commit
	for i := range sids {
		objs[i+1] = cm.Commit.CC.SpecRes
	}
	return objs
}

func (cm CommitMsg) getAllSig() [][]byte {
	sids := cm.Commit.CC.SIdList
	sigs := make([][]byte, 1+len(sids))
	sigs[0] = cm.CommitSig
	for i := range sids {
		sigs[i+1] = cm.Commit.CC.SigList[i]
	}
	return sigs
}

func (cm CommitMsg) GetAllPub() []*rsa.PublicKey {
	sids := cm.Commit.CC.SIdList
	pub := make([]*rsa.PublicKey, 1+len(sids))
	pub[0] = conf.Pub[cm.Commit.CId]
	for i := range sids {
		pub[i+1] = conf.Pub[sids[i]]
	}
	return pub
}
