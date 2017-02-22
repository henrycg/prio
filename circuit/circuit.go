// Package circuit implements arithmetic circuits.
package circuit

import (
	"fmt"
	"math/big"

	"github.com/henrycg/prio/share"
	"github.com/henrycg/prio/utils"
)

// Supported gate types.
const (
	Gate_Input    = iota
	Gate_Add      = iota
	Gate_AddConst = iota
	Gate_Mul      = iota
	Gate_MulConst = iota
)

// Global constant set to -1 in the field.
var NegOne *big.Int

// Gate represents a gate in an arithmetic circuit, and possibly
// holds the value on its output wire.
//
// On inputs x and y, each bilinear gate computes output
// z, where
//    z = (A x + B)(A' y + B').
// If both parents are nil, it's an input gate.
type Gate struct {
	GateType int
	ParentL  *Gate
	ParentR  *Gate

	Constant *big.Int

	WireValue *big.Int
}

// Circuit represents an arithmetic circuit over a particular finite
// field (specified by the modulus mod). The gates listed in shouldBeZero
// should all zero on their output wires iff the circuit accepts an input.
type Circuit struct {
	mod          *big.Int
	gates        []*Gate
	outputs      []*Gate
	shouldBeZero []*Gate
	outputNames  []string
}

func (c *Circuit) Modulus() *big.Int {
	return c.mod
}

func (c *Circuit) Outputs() []*Gate {
	return c.outputs
}

func (c *Circuit) OutputNames() []string {
	return c.outputNames
}

func (c *Circuit) OutputName(i int) string {
	return c.outputNames[i]
}

func (c *Circuit) ShouldBeZero() []*Gate {
	return c.shouldBeZero
}

func (c *Circuit) AddGate(g *Gate) {
	c.gates = append(c.gates, g)
}

func (c *Circuit) AddZeroGate(g *Gate) {
	c.gates = append(c.gates, g)
	c.shouldBeZero = append(c.shouldBeZero, g)
}

func Empty() *Circuit {
	out := new(Circuit)
	out.mod = share.IntModulus
	out.gates = make([]*Gate, 0)
	out.outputs = make([]*Gate, 0)
	out.shouldBeZero = make([]*Gate, 0)
	out.outputNames = make([]string, 0)
	return out
}

func NewGate() *Gate {
	g := new(Gate)
	g.WireValue = new(big.Int)
	return g
}

func NewGates(n int) []*Gate {
	out := make([]*Gate, n)
	for i := 0; i < n; i++ {
		out[i] = NewGate()
	}
	return out
}

func AndCircuits(ckts []*Circuit) *Circuit {
	out := Empty()
	for i := 0; i < len(ckts); i++ {
		out.gates = append(out.gates, ckts[i].gates...)
		out.outputs = append(out.outputs, ckts[i].outputs...)
		out.shouldBeZero = append(out.shouldBeZero, ckts[i].shouldBeZero...)
		out.outputNames = append(out.outputNames, ckts[i].outputNames...)
	}

	return out
}

// Input -> Output circuit
func UncheckedInput(name string) *Circuit {
	gIn := NewGate()
	gIn.GateType = Gate_Input

	c := Empty()
	c.gates = append(c.gates, gIn)
	c.outputs = append(c.outputs, gIn)
	c.outputNames = append(c.outputNames, name)
	return c
}

// A circuit that takes a single input x and computes
// 		x*(x-1).
// If x represents a 0/1 value in F, then this circuit
// outputs zero.
func OneBit(name string) *Circuit {
	c := Empty()

	inp := NewGate()
	inp.GateType = Gate_Input

	subOne := NewGate()
	subOne.GateType = Gate_AddConst
	subOne.ParentL = inp
	subOne.Constant = NegOne

	check := NewGate()
	check.ParentL = inp
	check.ParentR = subOne
	check.GateType = Gate_Mul

	c.outputs = append(c.outputs, inp)
	c.gates = append(c.gates, inp, subOne, check)
	c.shouldBeZero = append(c.shouldBeZero, check)
	c.outputNames = append(c.outputNames, name)
	return c
}

// A circuit that takes N inputs x1, ..., xN, which
// should all be 0/1 values in F.
//
// The circuit first computes y_i = (x_i)*(x_i - 1) and
// requires that all y_i's are zero.
//
// The output of the circuit is the sum:
//		\sum_i (2^i * x_i)
func NBits(bits int, name string) *Circuit {
	ckts := make([]*Circuit, bits)
	for b := 0; b < bits; b++ {
		ckts[b] = OneBit(fmt.Sprintf("%s[%v]", name, b))
	}
	bigckt := AndCircuits(ckts)

	muls := make([]*Gate, bits)
	adds := make([]*Gate, bits)
	for i := 0; i < bits; i++ {
		muls[i] = NewGate()
		muls[i].GateType = Gate_MulConst
		muls[i].Constant = big.NewInt(1 << uint(i))
		muls[i].ParentL = bigckt.InputGates()[i]

		adds[i] = NewGate()
		adds[i].ParentL = muls[i]
		if i == 0 {
			adds[i].GateType = Gate_AddConst
			adds[i].Constant = utils.Zero
		} else {
			adds[i].GateType = Gate_Add
			adds[i].ParentR = adds[i-1]
		}
	}

	bigckt.gates = append(bigckt.gates, muls...)
	bigckt.gates = append(bigckt.gates, adds...)
	bigckt.outputs = []*Gate{adds[bits-1]}
	bigckt.outputNames = []string{name}
	return bigckt
}

func MulByNegOne(parent *Gate) *Gate {
	inv := NewGate()
	inv.GateType = Gate_MulConst
	inv.Constant = NegOne
	inv.ParentL = parent
	return inv
}

// A circuit that checks that
//  (left * right == output)
func CheckMul(left, right, prod *Gate) *Circuit {
	mul := NewGate()
	mul.GateType = Gate_Mul
	mul.ParentL = left
	mul.ParentR = right

	inv := MulByNegOne(mul)

	add := NewGate()
	add.GateType = Gate_Add
	add.ParentL = inv
	add.ParentR = prod

	ckt := Empty()
	ckt.gates = []*Gate{mul, inv, add}
	ckt.shouldBeZero = []*Gate{add}

	return ckt
}

// Evaluate a circuit on the given input values. Return true
// iff the circuit accepts the input.
func (c *Circuit) Eval(inputs []*big.Int) bool {

	// Evaluate interal wires
	inpCount := 0
	for _, gate := range c.gates {
		switch gate.GateType {
		case Gate_Input:
			gate.WireValue.Set(inputs[inpCount])
			inpCount += 1
		case Gate_Add:
			gate.WireValue.Add(gate.ParentL.WireValue, gate.ParentR.WireValue)
			gate.WireValue.Mod(gate.WireValue, c.mod)
		case Gate_AddConst:
			gate.WireValue.Add(gate.ParentL.WireValue, gate.Constant)
			gate.WireValue.Mod(gate.WireValue, c.mod)
		case Gate_Mul:
			gate.WireValue.Mul(gate.ParentL.WireValue, gate.ParentR.WireValue)
			gate.WireValue.Mod(gate.WireValue, c.mod)
		case Gate_MulConst:
			gate.WireValue.Mul(gate.ParentL.WireValue, gate.Constant)
			gate.WireValue.Mod(gate.WireValue, c.mod)
		default:
			panic("Unknown gate type")
		}
	}

	out := true
	for _, gate := range c.shouldBeZero {
		out = out && (gate.WireValue.Sign() == 0)
	}

	return out
}

// Split the wire values into shares.
func (c *Circuit) ShareWires(prg *share.GenPRG) {
	for _, gate := range c.gates {
		switch gate.GateType {
		case Gate_Input:
			prg.Share(c.mod, gate.WireValue)
		case Gate_Mul:
			prg.Share(c.mod, gate.WireValue)
		}
	}
}

func (c *Circuit) gatesOfType(t int) []*Gate {
	gates := make([]*Gate, 0)
	for _, gate := range c.gates {
		switch gate.GateType {
		case t:
			gates = append(gates, gate)
		}
	}
	return gates
}

func (c *Circuit) MulGates() []*Gate {
	return c.gatesOfType(Gate_Mul)
}

func (c *Circuit) InputGates() []*Gate {
	return c.gatesOfType(Gate_Input)
}

// Import shared wire values from a ReplayPRG.
func (c *Circuit) ImportWires(prg *share.ReplayPRG) {
	for _, gate := range c.gates {
		switch gate.GateType {
		case Gate_Input:
			gate.WireValue = prg.Get(c.mod)
		case Gate_Add:
			gate.WireValue.Add(gate.ParentL.WireValue, gate.ParentR.WireValue)
			gate.WireValue.Mod(gate.WireValue, c.mod)
		case Gate_AddConst:
			// Only add constant if is leader
			toAdd := utils.Zero
			if prg.IsLeader() {
				toAdd = gate.Constant
			}
			gate.WireValue.Add(gate.ParentL.WireValue, toAdd)
			gate.WireValue.Mod(gate.WireValue, c.mod)
		case Gate_Mul:
			gate.WireValue = prg.Get(c.mod)
		case Gate_MulConst:
			gate.WireValue.Mul(gate.ParentL.WireValue, gate.Constant)
			gate.WireValue.Mod(gate.WireValue, c.mod)
		default:
			panic("Unknown gate type")
		}
	}
}

func init() {
	NegOne = new(big.Int)
	NegOne.Sub(share.IntModulus, utils.One)
}
