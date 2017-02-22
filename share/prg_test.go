package share

import (
	"math/big"
	"testing"
)

func TestPRG(t *testing.T) {
	mod := big.NewInt(3123130983042421)

	ns := 13
	leader := 3
	gen := NewGenPRG(ns, leader)

	v := big.NewInt(123131)
	shares := gen.Share(mod, v)

	res := new(big.Int)
	for i := 0; i < ns; i++ {
		res.Add(res, shares[i])
	}
	res.Mod(res, mod)
	if res.Cmp(v) != 0 {
		t.Fail()
	}

	for i := 0; i < ns; i++ {
		hints := gen.Hints(i)
		replay := NewReplayPRG(i, leader)
		replay.Import(hints)
		r := replay.Get(mod)
		if shares[i].Cmp(r) != 0 {
			t.Fail()
		}
	}
}
