package poly

import (
	"math/big"
	"runtime"
)

// #cgo CFLAGS: -Wall -std=c99 -pedantic
// #cgo LDFLAGS: -lgmp -L${SRCDIR}/libs
// #include "poly_once.h"
import "C"

// When we need to interpolate a polynomial through points
//    (x1, y1),  (x2, y2),  (x3, y3), ...
// and the x's are fixed AND we later want to evaluate this
// polynomial at a fixed point x*, you want this data struct.
type PreX struct {
	batchPre *BatchPre
	pre      C.precomp_x_t
}

func destroyPreX(pre *PreX) {
	C.precomp_x_clear(&pre.pre)
}

// Precompute data necessary to interpolate through fixed
// points (specified by batchPre) and evaluate at a point x.
func (batchPre *BatchPre) NewEvalPoint(x *big.Int) *PreX {
	out := new(PreX)
	out.batchPre = batchPre
	C.precomp_x_init(&out.pre, &batchPre.pre, bigToC(x))
	runtime.SetFinalizer(out, destroyPreX)
	return out
}

// Run the combined interpolation and evaluation routine.
func (preEval *PreX) Eval(yValues []*big.Int) *big.Int {
	strs := make([]*C.char, len(yValues))
	for i := 0; i < len(yValues); i++ {
		strs[i] = bigToC(yValues[i])
	}

	out := C.precomp_x_eval(&preEval.pre, &strs[0])
	return cToBig(out)
}
