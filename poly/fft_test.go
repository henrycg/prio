package poly

import (
	//	"log"
	"math/big"
	"testing"
	//	"time"

	"github.com/henrycg/prio/share"
	//	"github.com/henrycg/prio/utils"
)

func TestFFTOne(t *testing.T) {
	coeffs := []*big.Int{big.NewInt(3)}
	ys := FFT(coeffs)

	if ys[0].Cmp(big.NewInt(3)) != 0 {
		t.Fail()
	}
}

func TestFFTSimple(t *testing.T) {
	mod := share.IntModulus
	coeffs := []*big.Int{big.NewInt(3), big.NewInt(8), big.NewInt(7), big.NewInt(9)}
	ys := FFT(coeffs)

	roots := share.GetRoots(4)
	shouldBe := new(big.Int)
	tmp := new(big.Int)
	for i := 0; i < 4; i++ {
		shouldBe.SetInt64(0)
		for j := 0; j < 4; j++ {
			tmp.Exp(roots[i], big.NewInt(int64(j)), mod)
			tmp.Mul(tmp, coeffs[j])
			tmp.Mod(tmp, mod)
			shouldBe.Add(shouldBe, tmp)
			shouldBe.Mod(shouldBe, mod)
		}

		if shouldBe.Cmp(ys[i]) != 0 {
			t.Fatalf("Wanted %v, got %v", shouldBe, ys[i])
		}
	}
}

func TestFFTInvert(t *testing.T) {
	coeffs := []*big.Int{big.NewInt(3), big.NewInt(8), big.NewInt(7), big.NewInt(9),
		big.NewInt(123), big.NewInt(123123987), big.NewInt(2), big.NewInt(0)}
	ys := FFT(coeffs)
	xs := InverseFFT(ys)

	for i := 0; i < len(coeffs); i++ {
		if xs[i].Cmp(coeffs[i]) != 0 {
			t.Fatalf("Wanted %v, got %v (%v)", coeffs[i], xs[i], i)
		}
	}
}

/*
func TestFFTCompare(t *testing.T) {
	n := 1 << 16

	coeffs := make([]*big.Int, n)
	xs := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		coeffs[i] = utils.RandInt(share.IntModulus)
		xs[i] = utils.RandInt(share.IntModulus)
	}

	start := time.Now()
	InverseFFT(coeffs)
	log.Printf("FFT took: %v", time.Since(start))

	roots := share.GetRoots(n)
	start = time.Now()
	batch := NewBatch(share.IntModulus, roots)
	poly := batch.Interp(coeffs)
	log.Printf("Precomp took: %v", time.Since(start))
	start = time.Now()
	poly.Eval(xs)
	log.Printf("Batch took: %v", time.Since(start))
}
*/
