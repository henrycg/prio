package mpc

import (
	"github.com/henrycg/prio/circuit"
	"github.com/henrycg/prio/config"
)

// Produce the arithmetic circuit for checking the validity of
// a client submission.
func configToCircuit(cfg *config.Config) *circuit.Circuit {
	nf := len(cfg.Fields)

	ckts := make([]*circuit.Circuit, nf)

	for f := 0; f < nf; f++ {
		field := &cfg.Fields[f]
		switch field.Type {
		default:
			panic("Unexpected type!")
		case config.TypeInt:
			ckts[f] = int_Circuit(field.Name, int(field.IntBits))
		case config.TypeIntPow:
			ckts[f] = intPow_Circuit(field.Name, int(field.IntBits), int(field.IntPow))
		case config.TypeIntUnsafe:
			ckts[f] = intUnsafe_Circuit(field.Name)
		case config.TypeBoolOr:
			ckts[f] = bool_Circuit(field.Name)
		case config.TypeBoolAnd:
			ckts[f] = bool_Circuit(field.Name)
		case config.TypeCountMin:
			ckts[f] = countMin_Circuit(field.Name, int(field.CountMinHashes), int(field.CountMinBuckets))
		case config.TypeLinReg:
			ckts[f] = linReg_Circuit(field)
		}
	}

	ckt := circuit.AndCircuits(ckts)
	return ckt
}
