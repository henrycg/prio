package utils

import (
	"math/big"
	"testing"
)

func TestRandPerm(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	arr2 := RandPerm(arr)
	if len(arr2) != 10 {
		t.Fail()
	}
}

func TestInt(t *testing.T) {
	two := big.NewInt(2)

	var count int
	for i := 0; i < 10000; i++ {
		r := RandInt(two)
		if r.Sign() == 0 {
			count += 1
		}
	}

	if count < 4800 || count > 5200 {
		t.Fail()
	}
}
