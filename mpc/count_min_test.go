package mpc

import (
	"math/big"
	"testing"
)

func TestCountMinGood(t *testing.T) {
	nBuckets := 100
	nHashes := 5
	ckt := countMin_Circuit("blah", nHashes, nBuckets)
	vals := countMin_NewRandom(nHashes, nBuckets)

	if !ckt.Eval(vals) {
		t.Fail()
	}
}

func TestCountMinBad(t *testing.T) {
	nBuckets := 100
	nHashes := 5
	ckt := countMin_Circuit("CM", nHashes, nBuckets)
	vals := countMin_NewRandom(nHashes, nBuckets)

	vals[0] = big.NewInt(123123123123)

	if ckt.Eval(vals) {
		t.Fail()
	}
}
