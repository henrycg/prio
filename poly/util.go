package poly

import (
	"math/big"
	"strings"
	"unsafe"
)

// #include <stdlib.h>
// #include "util.h"
import "C"

func bigToC(x *big.Int) *C.char {
	return C.CString(x.Text(16))
}

func cToBig(cstr *C.char) *big.Int {
	out := new(big.Int)
	gstr := C.GoString(cstr)
	C.free(unsafe.Pointer(cstr))
	out.SetString(gstr, 16)
	return out
}

func cToBigArray(n int, cstr *C.char) []*big.Int {
	out := make([]*big.Int, n)
	gstr := C.GoString(cstr)
	C.free(unsafe.Pointer(cstr))

	pieces := strings.Split(gstr, "\n")
	if len(pieces) != n+1 {
		panic("Dimension mismatch")
	}

	for i := 0; i < n; i++ {
		out[i] = new(big.Int)
		out[i].SetString(pieces[i], 16)
	}

	return out
}
