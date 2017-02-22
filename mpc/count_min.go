package mpc

import (
	"fmt"
	"log"
	"math/big"

	"github.com/henrycg/prio/circuit"
	"github.com/henrycg/prio/utils"
)

func bucketToIndex(nBuckets, hash, bucket int) int {
	return hash*nBuckets + bucket
}

func rowCircuit(nHashes, nBuckets, row int, bits *circuit.Circuit) *circuit.Circuit {
	ckt := circuit.Empty()

	var last *circuit.Gate
	inps := bits.InputGates()
	// Sum up the bits in every row
	for i := 0; i < nBuckets; i++ {
		g := circuit.NewGate()
		g.ParentL = inps[(nBuckets*row)+i]
		if i == 0 {
			g.GateType = circuit.Gate_AddConst
			g.Constant = circuit.NegOne
		} else {
			g.GateType = circuit.Gate_Add
			g.ParentR = last
		}

		if i == nBuckets-1 {
			ckt.AddZeroGate(g)
		} else {
			ckt.AddGate(g)
		}
		last = g
	}

	return ckt
}

func countMin_Circuit(name string, nHashes, nBuckets int) *circuit.Circuit {
	total := nHashes * nBuckets

	// Ensure that each value in the sketch is a 0/1 value
	cktsBits := make([]*circuit.Circuit, total)
	for i := 0; i < total; i++ {
		cktsBits[i] = circuit.OneBit(fmt.Sprintf("%v[%v]", name, i))
	}
	ckt := circuit.AndCircuits(cktsBits)

	// Ensure that the sum of every row is 1
	ckts := make([]*circuit.Circuit, nHashes+1)
	ckts[0] = ckt
	for h := 0; h < nHashes; h++ {
		ckts[h+1] = rowCircuit(nHashes, nBuckets, h, ckt)
	}

	return circuit.AndCircuits(ckts)
}

func countMin_NewRandom(nHashes, nBuckets int) []*big.Int {
	total := nHashes * nBuckets
	values := make([]bool, total)

	bigHashes := big.NewInt(int64(nBuckets))
	for h := 0; h < nHashes; h++ {
		idx := int(utils.RandInt(bigHashes).Int64())
		values[bucketToIndex(nBuckets, h, idx)] = true
	}

	return countMin_New(nHashes, nBuckets, values)
}

func countMin_New(nHashes, nBuckets int, values []bool) []*big.Int {
	if nBuckets < 1 || nHashes < 1 {
		log.Fatal("nBuckets and nHashes must have value >= 1")
	}

	total := nHashes * nBuckets
	if len(values) != total {
		log.Fatal("Malformed request")
	}

	out := make([]*big.Int, total)
	for i := 0; i < total; i++ {
		if values[i] {
			out[i] = utils.One
		} else {
			out[i] = utils.Zero
		}
	}
	return out
}
