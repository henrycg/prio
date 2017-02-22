package poly

import (
	"math/big"
	"runtime"
)

// If you are using shared libraries, use:
//      LDFLAGS: -lflint -lgmp

// #cgo CFLAGS: -Wall -std=c99 -pedantic
// #cgo LDFLAGS: -lflint -lmpfr -lgmp -L ${SRCDIR}/libs
// #include "poly_batch.h"
import "C"

// Precomputated data for interpolating or evaluating
// a polynomial in a field defined by modulus mod.
type BatchPre struct {
	mod     *big.Int
	nPoints int
	pre     C.precomp_t
}

// Representation of a polynomia.
type BatchPoly struct {
	nPoints int
	fpoly   C.fmpz_mod_poly_t
}

func destroyBatchPre(pre *BatchPre) {
	C.poly_batch_precomp_clear(&pre.pre)
}

func destroyBatchPoly(poly *BatchPoly) {
	C.poly_batch_clear(&poly.fpoly[0])
}

// Precompute values for interpolating/evaluating a polynomial
// at the specified x coordinates.
func NewBatch(mod *big.Int, xPointsIn []*big.Int) *BatchPre {
	n := len(xPointsIn)
	pre := new(BatchPre)
	pre.mod = mod
	pre.nPoints = n

	xPoints := make([]*C.char, n)
	for i := 0; i < n; i++ {
		xPoints[i] = bigToC(xPointsIn[i])
	}

	C.poly_batch_precomp_init(&pre.pre, bigToC(mod), C.int(n), &xPoints[0])
	runtime.SetFinalizer(pre, destroyBatchPre)
	return pre
}

func (pre *BatchPre) NPoints() int {
	return pre.nPoints
}

// Interpolate through the y coordinates given, using the pre-specified
// x-coordinates.
func (pre *BatchPre) Interp(yPointsIn []*big.Int) *BatchPoly {
	n := len(yPointsIn)
	yPoints := make([]*C.char, n)

	for i := 0; i < n; i++ {
		yPoints[i] = bigToC(yPointsIn[i])
	}

	poly := new(BatchPoly)
	poly.nPoints = n

	C.poly_batch_init(&poly.fpoly[0], &pre.pre)
	runtime.SetFinalizer(poly, destroyBatchPoly)

	C.poly_batch_interpolate(&poly.fpoly[0], &pre.pre, &yPoints[0])

	return poly
}

// Evaluate the polynomial at a specified point x.
func (poly *BatchPoly) EvalOnce(x *big.Int) *big.Int {
	return cToBig(C.poly_batch_evaluate_once(&poly.fpoly[0], bigToC(x)))
}

// Evaluate the polynomial many times using faster a multi-point evaluation
// algorithm.
func (poly *BatchPoly) Eval(xPointsIn []*big.Int) []*big.Int {
	n := len(xPointsIn)

	xPoints := make([]*C.char, n)
	for i := 0; i < n; i++ {
		xPoints[i] = bigToC(xPointsIn[i])
	}

	cstr := C.poly_batch_evaluate(&poly.fpoly[0], C.int(n), &xPoints[0])
	return cToBigArray(n, cstr)
}
