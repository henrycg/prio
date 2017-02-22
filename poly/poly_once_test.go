package poly

import (
	"log"
	"math/big"
	"testing"

	"github.com/henrycg/prio/utils"
)

func xPoints(n int) []*big.Int {
	out := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		out[i] = big.NewInt(int64(i))
	}
	return out
}

func TestAllocPreDeg(t *testing.T) {
	mod := big.NewInt(13)
	NewBatch(mod, xPoints(10))
}

func TestAllocEvalPointSmall(t *testing.T) {
	mod := big.NewInt(13)
	preX := NewBatch(mod, xPoints(10))
	preX.NewEvalPoint(big.NewInt(0))
}

func TestAllocEvalPointBig(t *testing.T) {
	mod := big.NewInt(13)
	preX := NewBatch(mod, xPoints(10))
	preX.NewEvalPoint(big.NewInt(12))
}

func TestOnceConstant(t *testing.T) {
	const MOD = 19
	y := big.NewInt(7)
	mod := big.NewInt(MOD)
	yValues := make([]*big.Int, 3)
	yValues[0] = y
	yValues[1] = y
	yValues[2] = y

	preX := NewBatch(mod, xPoints(3))
	prePoint := preX.NewEvalPoint(big.NewInt(5))
	out := prePoint.Eval(yValues)
	if out.Cmp(y) != 0 {
		t.Fail()
	}
}

func TestOnceLinear(t *testing.T) {
	const A = 2
	const B = 3
	const MOD = 13

	f := func(x int) *big.Int {
		return big.NewInt(int64((A*x + B) % MOD))
	}

	mod := big.NewInt(MOD)
	yValues := make([]*big.Int, 2)
	yValues[0] = f(0)
	yValues[1] = f(1)

	preX := NewBatch(mod, xPoints(2))
	prePoint := preX.NewEvalPoint(big.NewInt(7))
	out := prePoint.Eval(yValues)
	if out.Cmp(f(7)) != 0 {
		t.Fail()
	}

	prePoint1 := preX.NewEvalPoint(big.NewInt(1))
	if prePoint1.Eval(yValues).Cmp(f(1)) != 0 {
		t.Fail()
	}
}

func TestOnceTwo(t *testing.T) {
	// f(x) = Ax^2 + Bx + C

	const A = 7
	const B = 3
	const C = 12
	const MOD = 19

	f := func(x int64) *big.Int {
		return big.NewInt(((A * x * x) + (B * x) + C) % MOD)
	}

	yValues := []*big.Int{f(0), f(1), f(2)}
	preX := NewBatch(big.NewInt(MOD), xPoints(len(yValues)))

	for i := int64(0); i < MOD; i++ {
		prePoint := preX.NewEvalPoint(big.NewInt(i))
		out := prePoint.Eval(yValues)
		if out.Cmp(f(i)) != 0 {
			t.Fail()
		}
	}
}

func TestOnceBig(t *testing.T) {
	const degree = 13
	mod := big.NewInt(104287)

	coeffs := []*big.Int{
		big.NewInt(123),
		big.NewInt(8),
		big.NewInt(283),
		big.NewInt(87),
		big.NewInt(3),
		big.NewInt(15553),
		big.NewInt(9123),
		big.NewInt(0123),
		big.NewInt(2341),
		big.NewInt(111),
		big.NewInt(11),
		big.NewInt(90008),
		big.NewInt(104286),
	}

	yValues := make([]*big.Int, degree+1)
	for i := 0; i < degree+1; i++ {
		yValues[i] = evalSlow(mod, coeffs, big.NewInt(int64(i)))
	}

	preX := NewBatch(mod, xPoints(degree+1))
	prePoint := preX.NewEvalPoint(big.NewInt(1234))
	out := prePoint.Eval(yValues)
	if out.Cmp(evalSlow(mod, coeffs, big.NewInt(1234))) != 0 {
		t.Fail()
	}
}

func testOnceInterpDegree(t *testing.T, degree int) {
	mod := big.NewInt(2541622467539)

	coeffs := make([]*big.Int, degree+1)
	for i := 0; i < degree+1; i++ {
		coeffs[i] = utils.RandInt(mod)
	}

	yValues := make([]*big.Int, degree+1)
	for i := 0; i < len(coeffs); i++ {
		x := big.NewInt(int64(i))
		yValues[i] = evalSlow(mod, coeffs, x)
	}

	preX := NewBatch(mod, xPoints(degree+1))

	for i := 0; i < degree+1; i++ {
		x := big.NewInt(int64(40*i + 7))
		prePoint := preX.NewEvalPoint(x)
		res1 := prePoint.Eval(yValues)
		res2 := evalSlow(mod, coeffs, x)
		if res1.Cmp(res2) != 0 {
			log.Printf("%v != %v", res1, res2)
			log.Fatal("Failed")
		}
	}
}

func TestOnceInterp1(t *testing.T) {
	testOnceInterpDegree(t, 1)
}

func TestOnceInterp2(t *testing.T) {
	testOnceInterpDegree(t, 2)
}

func TestOnceInterp3(t *testing.T) {
	testOnceInterpDegree(t, 3)
}

func TestOnceInterp4(t *testing.T) {
	testOnceInterpDegree(t, 4)
}

func TestOnceInterp5(t *testing.T) {
	testOnceInterpDegree(t, 5)
}

func TestOnceInterp7(t *testing.T) {
	testOnceInterpDegree(t, 7)
}

func TestOnceInterp8(t *testing.T) {
	testOnceInterpDegree(t, 8)
}

func TestOnceInterp19(t *testing.T) {
	testOnceInterpDegree(t, 19)
}

func TestOnceInterp62(t *testing.T) {
	testOnceInterpDegree(t, 62)
}

func TestOnceInterp620(t *testing.T) {
	testOnceInterpDegree(t, 620)
}

func TestOnceInterp1234(t *testing.T) {
	testOnceInterpDegree(t, 1234)
}

func TestOnceInterp4345(t *testing.T) {
	mod := big.NewInt(2541622467539)
	NewBatch(mod, xPoints(4345))
}
