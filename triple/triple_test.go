package triple

import (
	"log"
	"math/big"
	"testing"

	"github.com/henrycg/prio/share"
	"github.com/henrycg/prio/utils"
)

func TestAddMul(t *testing.T) {
	n := 1024 * 1024

	tAdd := uint64(0)
	for i := 0; i < n; i++ {
		a := utils.RandInt(share.IntModulus)
		b := utils.RandInt(share.IntModulus)
		c := new(big.Int)
		t := utils.GetUtime()
		c.Add(a, b)
		c.Mod(c, share.IntModulus)
		tAdd += t - utils.GetUtime()
	}

	tMul := uint64(0)
	for i := 0; i < n; i++ {
		a := utils.RandInt(share.IntModulus)
		b := utils.RandInt(share.IntModulus)
		c := new(big.Int)
		t := utils.GetUtime()
		c.Mul(a, b)
		c.Mod(c, share.IntModulus)
		tMul += t - utils.GetUtime()
	}

	log.Printf("Add: %v", float64(tAdd)/float64(n))
	log.Printf("Mul: %v", float64(tMul)/float64(n))
}

func testGen(tst *testing.T, nServers int, nTriples int) {
	mod := share.IntModulus

	ts := NewTriple(mod, nServers)
	t := EmptyTriple()

	for s := 0; s < nServers; s++ {
		t.A.Add(t.A, ts[s].ShareA)
		t.B.Add(t.B, ts[s].ShareB)
		t.C.Add(t.C, ts[s].ShareC)
	}

	t.A.Mod(t.A, mod)
	t.B.Mod(t.B, mod)
	t.C.Mod(t.C, mod)

	prod := new(big.Int)
	prod.Mul(t.A, t.B)
	prod.Mod(prod, mod)
	prod.Sub(prod, t.C)
	prod.Mod(prod, mod)
	if prod.Sign() != 0 {
		tst.Fatal("Product is wrong")
	}
}

func TestBatchOneOne(t *testing.T) {
	testGen(t, 1, 1)
}

func TestBatchOneTriple(t *testing.T) {
	testGen(t, 14, 1)
}

func TestBatchOneServer(t *testing.T) {
	testGen(t, 1, 123)
}

func TestBatchMany(t *testing.T) {
	testGen(t, 13, 123)
}

func TestBatchMany2(t *testing.T) {
	testGen(t, 23, 22)
}

func TestBatchManyMany(t *testing.T) {
	testGen(t, 8, 220)
}
