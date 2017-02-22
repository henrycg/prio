package poly

import (
	"github.com/henrycg/prio/share"
	"math/big"
)

// #cgo CFLAGS: -Wall -std=c99 -pedantic
// #cgo LDFLAGS: -lflint -lgmp -L ${SRCDIR}/libs
// #include "fft.h"
import "C"

func FFT(yPointsIn []*big.Int) []*big.Int {
	return fft(yPointsIn, false)
}

func InverseFFT(yPointsIn []*big.Int) []*big.Int {
	return fft(yPointsIn, true)
}

func fft(yPointsIn []*big.Int, invert bool) []*big.Int {
	n := len(yPointsIn)
	var rootsIn []*big.Int
	if invert {
		rootsIn = share.GetRootsInv(n)
	} else {
		rootsIn = share.GetRoots(n)
	}

	yPoints := make([]*C.char, n)
	roots := make([]*C.char, n)

	for i := 0; i < n; i++ {
		yPoints[i] = bigToC(yPointsIn[i])
		roots[i] = bigToC(rootsIn[i])
	}

	cstr := C.fft_interpolate(bigToC(share.IntModulus), C.int(n),
		&roots[0], &yPoints[0], C.bool(invert))

	return cToBigArray(n, cstr)
}
