package triple

import (
	"github.com/henrycg/prio/utils"
	"math/big"
)

// A Beaver multiplication triple for the MPC multiplication.
type Triple struct {
	A *big.Int
	B *big.Int
	C *big.Int
}

// A share of a Beaver multiplication triple
type Share struct {
	ShareA *big.Int
	ShareB *big.Int
	ShareC *big.Int
}

func EmptyTriple() *Triple {
	t := new(Triple)
	t.A = new(big.Int)
	t.B = new(big.Int)
	t.C = new(big.Int)
	return t
}

func EmptyShare() *Share {
	s := new(Share)
	s.ShareA = new(big.Int)
	s.ShareB = new(big.Int)
	s.ShareC = new(big.Int)
	return s
}

func IsValid(mod *big.Int, shares []Share) bool {
	t := EmptyTriple()

	n := len(shares)
	for i := 0; i < n; i++ {
		t.A.Add(t.A, shares[i].ShareA)
		t.B.Add(t.B, shares[i].ShareB)
		t.C.Add(t.C, shares[i].ShareC)

		t.A.Mod(t.A, mod)
		t.B.Mod(t.B, mod)
		t.C.Mod(t.C, mod)
	}

	// Check if C == A*B
	t.A.Mul(t.A, t.B)
	t.A.Mod(t.A, mod)

	return t.A.Cmp(t.C) == 0
}

func NewTriple(mod *big.Int, nServers int) []*Share {
	out := make([]*Share, nServers)
	for s := 0; s < nServers; s++ {
		out[s] = EmptyShare()
	}

	t := EmptyTriple()
	for s := 0; s < nServers; s++ {
		out[s].ShareA = utils.RandInt(mod)
		out[s].ShareB = utils.RandInt(mod)
		out[s].ShareC = utils.RandInt(mod)

		t.A.Add(t.A, out[s].ShareA)
		t.B.Add(t.B, out[s].ShareB)
		t.C.Add(t.C, out[s].ShareC)
	}

	t.A.Mod(t.A, mod)
	t.B.Mod(t.B, mod)
	t.C.Mod(t.C, mod)

	// We want A*B = C. Tweak the last share so that
	// this relation holds.

	// prod now holds the product A*B
	prod := new(big.Int)
	prod.Mul(t.A, t.B)

	// Compute prod - C. This is the delta we need
	// to have A*B = C.
	prod.Sub(prod, t.C)
	prod.Mod(prod, mod)

	out[0].ShareC.Add(out[0].ShareC, prod)
	out[0].ShareC.Mod(out[0].ShareC, mod)

	return out
}
