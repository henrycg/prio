package utils

import "testing"

func TestNextPowerOfTwo(t *testing.T) {
	v := []int{0, 4, 3, 7, 12, 8, 16, 1}
	shouldBe := []int{1, 4, 4, 8, 16, 8, 16, 1}

	for i := range shouldBe {
		if NextPowerOfTwo(v[i]) != shouldBe[i] {
			t.Fatalf("Wanted %v, got %v", shouldBe[i], NextPowerOfTwo(v[i]))
		}
	}

}
