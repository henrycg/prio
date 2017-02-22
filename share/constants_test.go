package share

import (
	"math/big"
	"testing"
)

func TestGetRoots2(t *testing.T) {
	r := GetRoots(2)
	if len(r) != 2 {
		t.Fail()
	}

	if r[0].Cmp(big.NewInt(1)) != 0 {
		t.Fail()
	}

	v := new(big.Int)
	v.Set(r[1])
	v.Mul(v, v)
	v.Mod(v, IntModulus)
	if v.Cmp(big.NewInt(1)) != 0 {
		t.Fail()
	}
}

func TestGetRoots4(t *testing.T) {
	r := GetRoots(4)
	if len(r) != 4 {
		t.Fail()
	}

	if r[0].Cmp(big.NewInt(1)) != 0 {
		t.Fail()
	}

	v := new(big.Int)
	v.Set(r[1])
	v.Mul(v, v)
	v.Mul(v, v)
	v.Mul(v, v)
	v.Mod(v, IntModulus)
	if v.Cmp(big.NewInt(1)) != 0 {
		t.Fail()
	}
}
