package mpc

import (
	"math/big"

	"github.com/henrycg/prio/circuit"
	"github.com/henrycg/prio/utils"
)

type boolOp int

const (
	Op_OR  boolOp = iota
	Op_AND boolOp = iota
)

func bool_Circuit(name string) *circuit.Circuit {
	return circuit.UncheckedInput(name)
}

func bool_NewRandom() []*big.Int {
	v := (utils.RandInt(big.NewInt(2)).Cmp(big.NewInt(0)) == 1)
	return bool_New(v)
}

func bool_New(value bool) []*big.Int {
	vInt := int64(0)
	if value {
		vInt = 1
	}

	return []*big.Int{big.NewInt(vInt)}
}
