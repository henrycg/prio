package poly

import (
	"log"
	"math/big"
	"testing"

	"github.com/henrycg/prio/utils"
)

func evalSlow(mod *big.Int, coeffs []*big.Int, x *big.Int) *big.Int {
	y := new(big.Int)
	tmp := new(big.Int)
	for i := 0; i < len(coeffs); i++ {
		tmp.Exp(x, big.NewInt(int64(i)), mod)
		tmp.Mul(tmp, coeffs[i])
		tmp.Mod(tmp, mod)

		y.Add(y, tmp)
		y.Mod(y, mod)
	}

	return y
}

func TestFlint(t *testing.T) {
	mod := big.NewInt(13)

	xs := []*big.Int{big.NewInt(4), big.NewInt(5)}
	ys := []*big.Int{big.NewInt(3), big.NewInt(5)}

	pre := NewBatch(mod, xs)
	poly := pre.Interp(ys)

	y2 := poly.Eval([]*big.Int{big.NewInt(4), big.NewInt(6)})
	if y2[0].Cmp(big.NewInt(3)) != 0 {
		t.Fail()
	}
	if y2[1].Cmp(big.NewInt(7)) != 0 {
		t.Fail()
	}

	poly.EvalOnce(big.NewInt(10))
}

func TestFlintInterpZero(t *testing.T) {
	xs := []*big.Int{big.NewInt(1)}
	ys := []*big.Int{big.NewInt(4)}

	pre := NewBatch(big.NewInt(5), xs)
	poly := pre.Interp(ys)
	for i := 0; i < 13; i++ {
		out := poly.EvalOnce(big.NewInt(3))
		if out.Cmp(big.NewInt(4)) != 0 {
			t.Fail()
		}
	}
}

func TestFlintInterpOne(t *testing.T) {
	// f(x) = Ax + B

	const A = 3
	const B = 12
	const MOD = 13
	xs := []*big.Int{big.NewInt(0), big.NewInt(1)}
	ys := []*big.Int{big.NewInt(B), big.NewInt((A + B) % 13)}

	//log.Printf("f(x) = %vx + %v   (mod %v)", A, B, MOD)
	pre := NewBatch(big.NewInt(MOD), xs)
	poly := pre.Interp(ys)

	for i := int64(0); i < MOD; i++ {
		val := big.NewInt((i*A + B) % MOD)
		out := poly.EvalOnce(big.NewInt(i))
		//log.Printf("Wanted %v, Got %v", val, out)
		if out.Cmp(val) != 0 {
			t.Fail()
		}
	}
}

func TestFlintInterpTwo(t *testing.T) {
	// f(x) = Ax^2 + Bx + C

	const A = 7
	const B = 3
	const C = 12
	const MOD = 19

	f := func(x int64) int64 {
		return ((A * x * x) + (B * x) + C) % MOD
	}

	xs := []*big.Int{big.NewInt(2), big.NewInt(5), big.NewInt(10)}
	ys := []*big.Int{big.NewInt(f(2)), big.NewInt(f(5)), big.NewInt(f(10))}

	pre := NewBatch(big.NewInt(MOD), xs)
	poly := pre.Interp(ys)

	for i := int64(0); i < MOD; i++ {
		val := big.NewInt(f(i))
		out := poly.EvalOnce(big.NewInt(i))
		if out.Cmp(val) != 0 {
			t.Fail()
		}
	}
}

func TestFlintInterpBig(t *testing.T) {
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

	f := func(xin int64) *big.Int {
		x := big.NewInt(xin)
		out := new(big.Int)
		tmp := new(big.Int)
		for i := 0; i < len(coeffs); i++ {
			tmp.Exp(x, big.NewInt(int64(i)), mod)
			tmp.Mul(tmp, coeffs[i])
			tmp.Mod(tmp, mod)

			out.Add(out, tmp)
			out.Mod(out, mod)
		}
		return out
	}

	xs := []*big.Int{
		big.NewInt(13),
		big.NewInt(1),
		big.NewInt(132),
		big.NewInt(8883),
		big.NewInt(93),
		big.NewInt(3),
		big.NewInt(77),
		big.NewInt(8989),
		big.NewInt(24),
		big.NewInt(3333),
		big.NewInt(4141),
		big.NewInt(77709),
		big.NewInt(7709),
		big.NewInt(709)}

	ys := []*big.Int{
		f(13),
		f(1),
		f(132),
		f(8883),
		f(93),
		f(3),
		f(77),
		f(8989),
		f(24),
		f(3333),
		f(4141),
		f(77709),
		f(7709),
		f(709)}

	pre := NewBatch(mod, xs)
	poly := pre.Interp(ys)

	for i := int64(0); i < 100; i++ {
		out := poly.EvalOnce(big.NewInt(i))
		if out.Cmp(f(i)) != 0 {
			t.Fail()
		}
	}
}

func testFlintInterpDegree(t *testing.T, degree int) {
	mod := big.NewInt(2541622467539)

	coeffs := make([]*big.Int, degree+1)
	for i := 0; i < degree+1; i++ {
		coeffs[i] = utils.RandInt(mod)
	}

	xs := make([]*big.Int, degree+1)
	ys := make([]*big.Int, degree+1)
	for i := 0; i < len(coeffs); i++ {
		x := big.NewInt(int64(2*i + 3))
		xs[i] = x
		ys[i] = evalSlow(mod, coeffs, x)
	}

	pre := NewBatch(mod, xs)
	poly := pre.Interp(ys)

	for i := 0; i < degree+1; i++ {
		x := big.NewInt(int64(4*i + 7))
		res1 := poly.EvalOnce(x)
		res2 := evalSlow(mod, coeffs, x)
		if res1.Cmp(res2) != 0 {
			log.Printf("%v != %v", res1, res2)
			log.Fatal("Failed")
		}
	}
}

func TestFlintInterpDegreeSpeed(t *testing.T) {
	degree := 40000
	mod := big.NewInt(2541622467539)
	xs := make([]*big.Int, degree+1)
	ys := make([]*big.Int, degree+1)

	for i := 0; i < len(xs); i++ {
		x := big.NewInt(int64(i))
		xs[i] = x
		ys[i] = utils.RandInt(mod)
	}

	pre := NewBatch(mod, xs)
	pre.Interp(ys)
}

func BenchmarkInterp(b *testing.B) {
	mod := utils.RandInt(big.NewInt(12315012494093))
	xs := make([]*big.Int, b.N)
	ys := make([]*big.Int, b.N)
	for i := 0; i < b.N; i++ {
		xs[i] = big.NewInt(int64(i))
		ys[i] = utils.RandInt(mod)
	}

	pre := NewBatch(mod, xs)
	b.ResetTimer()
	pre.Interp(ys)
}

func TestFlintInterp0(t *testing.T) {
	testFlintInterpDegree(t, 0)
}

func TestFlintInterp1(t *testing.T) {
	testFlintInterpDegree(t, 1)
}

func TestFlintInterp2(t *testing.T) {
	testFlintInterpDegree(t, 2)
}

func TestFlintInterp3(t *testing.T) {
	testFlintInterpDegree(t, 3)
}

func TestFlintInterp4(t *testing.T) {
	testFlintInterpDegree(t, 4)
}

func TestFlintInterp5(t *testing.T) {
	testFlintInterpDegree(t, 5)
}

func TestFlintInterp7(t *testing.T) {
	testFlintInterpDegree(t, 7)
}

func TestFlintInterp8(t *testing.T) {
	testFlintInterpDegree(t, 8)
}

func TestFlintInterp256(t *testing.T) {
	testFlintInterpDegree(t, 256)
}

func TestFlintInterp257(t *testing.T) {
	testFlintInterpDegree(t, 257)
}

/*
func TestFlintInterp2570(t *testing.T) {
  testFlintInterpDegree(t, 2570)
}

func TestFlintInterp1000(t *testing.T) {
  testFlintInterpDegree(t, 1000)
}
*/
