package mpc

import (
	"math/big"
	"testing"
)

func TestIntPowGood(t *testing.T) {
	nBits := 5
	pow := 8
	ckt := intPow_Circuit("pow", nBits, pow)
	vals := intPow_NewRandom(nBits, pow)

	if !ckt.Eval(vals) {
		t.Fail()
	}
}

func TestIntPowBad(t *testing.T) {
	nBits := 5
	pow := 8
	ckt := intPow_Circuit("pow", nBits, pow)
	vals := intPow_NewRandom(nBits, pow)

	vals[0] = big.NewInt(123123123123)

	if ckt.Eval(vals) {
		t.Fail()
	}
}
