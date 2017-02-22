package mpc

import (
	"fmt"
	"math/big"

	"github.com/henrycg/prio/circuit"
	"github.com/henrycg/prio/config"
	"github.com/henrycg/prio/share"
	"github.com/henrycg/prio/utils"
)

func linReg_Circuit(field *config.Field) *circuit.Circuit {
	nTerms := len(field.LinRegBits)
	// Check x_i's
	xCkts := make([]*circuit.Circuit, nTerms)
	for t := 0; t < nTerms; t++ {
		name := fmt.Sprintf("%v-bits[%v]", field.Name, t)
		xCkts[t] = circuit.NBits(field.LinRegBits[t], name)
	}

	// Check x_i * x_j
	prodCkts := make([]*circuit.Circuit, 0)
	prodMulCkts := make([]*circuit.Circuit, 0)
	for i := 0; i < nTerms; i++ {
		for j := 0; j < nTerms; j++ {
			if i >= j {
				name := fmt.Sprintf("%v-prod[%v*%v]", field.Name, i, j)
				prod := circuit.UncheckedInput(name)

				x_i := xCkts[i].Outputs()[0]
				x_j := xCkts[j].Outputs()[0]
				prodMulCkts = append(prodMulCkts, circuit.CheckMul(x_i, x_j, prod.Outputs()[0]))
				prodCkts = append(prodCkts, prod)
			}
		}
	}

	ckts := make([]*circuit.Circuit, 0)
	ckts = append(ckts, xCkts...)
	ckts = append(ckts, prodCkts...)
	ckts = append(ckts, prodMulCkts...)

	return circuit.AndCircuits(ckts)
}

func linReg_NewRandom(field *config.Field) []*big.Int {
	nTerms := len(field.LinRegBits)
	max := new(big.Int)
	values := make([]*big.Int, nTerms)
	for t := 0; t < nTerms; t++ {
		max.SetUint64(1)
		max.Lsh(max, uint(field.LinRegBits[t]))
		values[t] = utils.RandInt(max)
	}

	return linReg_New(field, values)
}

func linReg_New(field *config.Field, values []*big.Int) []*big.Int {

	nTerms := len(field.LinRegBits)
	out := make([]*big.Int, 0)

	if len(values) != nTerms {
		panic("Invalid data input")
	}

	// Output x_i's in bits
	for t := 0; t < nTerms; t++ {
		out = append(out, bigToBits(field.LinRegBits[t], values[t])...)
	}

	// Compute  (x_i * x_j) for all (i,j)
	for i := 0; i < nTerms; i++ {
		for j := 0; j < nTerms; j++ {
			if i >= j {
				v := new(big.Int)
				v.Mul(values[i], values[j])
				v.Mod(v, share.IntModulus)
				out = append(out, v)
			}
		}
	}

	return out
}
