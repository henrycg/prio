package mpc

import (
	"math/big"
	"testing"
)

func TestIntGood(t *testing.T) {
	nBits := 5
	ckt := int_Circuit("int", nBits)
	vals := int_NewRandom(nBits)

	if !ckt.Eval(vals) {
		t.Fail()
	}
}

func TestIntBad(t *testing.T) {
	nBits := 5
	ckt := int_Circuit("int", nBits)
	vals := int_NewRandom(nBits)

	vals[0] = big.NewInt(123123123123)

	if ckt.Eval(vals) {
		t.Fail()
	}
}
