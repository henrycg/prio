package share

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/henrycg/prio/utils"
)

func testShare(t *testing.T, ns int) {
	mod := big.NewInt(3123130983042421)
	r, err := rand.Int(rand.Reader, mod)
	if err != nil {
		t.Fatal("Randomness error")
	}

	shares := Share(mod, ns, r)

	if len(shares) != ns {
		t.Fatal("Wrong number of shares")
	}

	acc := big.NewInt(0)
	for i := 0; i < ns; i++ {
		acc.Add(acc, shares[i])
	}
	acc.Mod(acc, mod)

	if acc.Cmp(r) != 0 {
		t.Fatal("Wrong result")
	}
}

func TestShareOne(t *testing.T) {
	testShare(t, 1)
}

func TestShareMany(t *testing.T) {
	testShare(t, 127)
}

func BenchmarkShare(b *testing.B) {
	r := utils.RandInt(IntModulus)
	for i := 0; i < b.N; i++ {
		r.Mul(r, r)
		r.Mod(r, IntModulus)
	}
}

func BenchmarkMul(b *testing.B) {
	r1 := utils.RandInt(IntModulus)
	r2 := utils.RandInt(IntModulus)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r1.Mul(r1, r2)
		r1.Mod(r1, IntModulus)
	}
}

func BenchmarkExp(b *testing.B) {
	r1 := utils.RandInt(IntModulus)
	exps := make([]*big.Int, 50)
	for i := 0; i < len(exps); i++ {
		exps[i] = utils.RandInt(IntModulus)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r1.Exp(r1, exps[i%50], IntModulus)
	}
}
