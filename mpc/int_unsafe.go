package mpc

import (
	"log"
	"math/big"

	"github.com/henrycg/prio/circuit"
	"github.com/henrycg/prio/utils"
)

func intUnsafe_Circuit(name string) *circuit.Circuit {
	return circuit.UncheckedInput(name)
}

func intUnsafe_NewRandom(nBits int) []*big.Int {
	max := big.NewInt(1)
	max.Lsh(max, uint(nBits))
	v := utils.RandInt(max)
	return intUnsafe_New(nBits, v)
}

func intUnsafe_New(nBits int, value *big.Int) []*big.Int {
	if nBits < 1 {
		log.Fatal("nBits must have value >= 1")
	}

	if value.Sign() == -1 {
		log.Fatal("Value must be non-negative")
	}

	vLen := value.BitLen()
	if vLen > nBits {
		log.Fatal("Value is too long")
	}

	return []*big.Int{value}
}
