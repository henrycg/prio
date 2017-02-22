package circuit

import (
	"math/big"
	"testing"
)

func TestOneBit(t *testing.T) {
	c := OneBit("")
	if c.Eval([]*big.Int{big.NewInt(0)}) != true {
		t.Fail()
	}

	if c.Eval([]*big.Int{big.NewInt(1)}) != true {
		t.Fail()
	}

	if c.Eval([]*big.Int{big.NewInt(2)}) != false {
		t.Fail()
	}

	if c.Eval([]*big.Int{big.NewInt(1231232)}) != false {
		t.Fail()
	}
}
