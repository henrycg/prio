package mpc

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/henrycg/prio/circuit"
	"github.com/henrycg/prio/config"
)

// Holds aggregates statistics computed by the Prio instance.
type Aggregator struct {
	cfg    *config.Config
	ckt    *circuit.Circuit
	Values []*big.Int
	Names  []string
	n      int
}

func NewAggregator(cfg *config.Config) *Aggregator {
	agg := new(Aggregator)
	agg.cfg = cfg
	agg.ckt = configToCircuit(cfg)
	agg.n = len(agg.ckt.Outputs())
	agg.Values = make([]*big.Int, agg.n)
	agg.Names = make([]string, agg.n)

	for i := 0; i < agg.n; i++ {
		agg.Values[i] = new(big.Int)
		agg.Names[i] = agg.ckt.OutputName(i)
	}

	return agg
}

func (agg *Aggregator) Update(chk *Checker) {
	outs := chk.Outputs()

	for i := 0; i < agg.n; i++ {
		agg.Values[i].Add(agg.Values[i], outs[i].WireValue)
		agg.Values[i].Mod(agg.Values[i], agg.ckt.Modulus())
	}
}

func (agg *Aggregator) Reset() {
	for i := 0; i < agg.n; i++ {
		agg.Values[i].SetInt64(0)
	}
}

func (agg *Aggregator) Copy() *Aggregator {
	out := NewAggregator(agg.cfg)
	for i := 0; i < agg.n; i++ {
		out.Values[i].Set(agg.Values[i])
		out.Names[i] = agg.Names[i]
	}
	return out
}

func (agg *Aggregator) Combine(other *Aggregator) {
	for i := 0; i < agg.n; i++ {
		agg.Values[i].Add(agg.Values[i], other.Values[i])
		agg.Values[i].Mod(agg.Values[i], agg.ckt.Modulus())
	}
}

func (agg *Aggregator) String() string {
	var buf bytes.Buffer
	buf.WriteString("=== Aggregator ===\n")
	for i := 0; i < agg.n; i++ {
		buf.WriteString(fmt.Sprint(i, agg.Names[i], " => ", agg.Values[i], "\n"))
	}
	buf.WriteString("==================\n")

	return buf.String()
}
