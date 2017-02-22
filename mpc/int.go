package mpc

import (
	"log"
	"math/big"

	"github.com/henrycg/prio/circuit"
	"github.com/henrycg/prio/utils"
)

func bigToBits(nBits int, value *big.Int) []*big.Int {
	bits := make([]*big.Int, nBits)
	for i := 0; i < nBits; i++ {
		bits[i] = big.NewInt(int64(value.Bit(i)))
	}

	return bits
}

func int_Circuit(name string, nBits int) *circuit.Circuit {
	return circuit.NBits(nBits, name)
}

func int_NewRandom(nBits int) []*big.Int {
	max := big.NewInt(1)
	max.Lsh(max, uint(nBits))
	v := utils.RandInt(max)
	return int_New(nBits, v)
}

func int_New(nBits int, value *big.Int) []*big.Int {
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

	return bigToBits(nBits, value)
}
