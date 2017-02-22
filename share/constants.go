package share

import (
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"
	"os"
	"path"

	"github.com/henrycg/prio/utils"
)

var CacheFileRoot string = "prio-roots-cache"

// A modulus p.
var IntModulus *big.Int

// A generator g of a subgroup of Z*_p.
var IntGen *big.Int

// The generator g generates a subgroup of
// order 2^IntGen2Order in Z*_p.
var IntGen2Order uint

// This is precomputed information about the
// roots of unity in Z*_p. We cache it on disk
// to avoid recomputing it every time we run
// the program.
type rootsPrecomp struct {
	Mod      *big.Int
	Gen      *big.Int
	TwoOrder uint
	Roots    []*big.Int
	RootsInv []*big.Int
}

var precomp *rootsPrecomp

func fromString(s string) *big.Int {
	out := new(big.Int)
	out.SetString(s, 16)
	return out
}

func GetRoots(n int) []*big.Int {
	return getRoots(n, false)
}

func GetRootsInv(n int) []*big.Int {
	return getRoots(n, true)
}

// Get a slice
//    (r^0, r^1, r^2, ..., )
// where r is an n-th root of unity.
func getRoots(n int, inverse bool) []*big.Int {
	out := make([]*big.Int, n)
	stepSize := (1 << IntGen2Order) / n
	for i := 0; i < n; i++ {
		if inverse {
			out[i] = precomp.RootsInv[(stepSize * i)]
		} else {
			out[i] = precomp.Roots[(stepSize * i)]
		}
	}
	return out
}

// Compute the roots of unity in the field defined by the given
// modulus, where g is a generator of order 2^twoOrder.
func computeRoots(mod, gen *big.Int, twoOrder uint) *rootsPrecomp {
	log.Printf("Computing roots...")
	out := new(rootsPrecomp)
	out.Mod = mod
	out.Gen = gen
	out.TwoOrder = twoOrder

	gInv := new(big.Int)
	gInv.ModInverse(gen, mod)

	// Roots holds the vector of powers
	//    (g^0, g^1, g^2, g^3, ..., )
	out.Roots = make([]*big.Int, 1<<twoOrder)
	out.RootsInv = make([]*big.Int, 1<<twoOrder)
	out.Roots[0] = utils.One
	out.RootsInv[0] = utils.One
	for i := 1; i < len(out.Roots); i++ {
		out.Roots[i] = new(big.Int)
		out.Roots[i].Mul(IntGen, out.Roots[i-1])
		out.Roots[i].Mod(out.Roots[i], mod)

		out.RootsInv[i] = new(big.Int)
		out.RootsInv[i].Mul(gInv, out.RootsInv[i-1])
		out.RootsInv[i].Mod(out.RootsInv[i], mod)
	}
	log.Printf("Done.")

	return out
}

func fileNameHash(mod, gen *big.Int, twoOrder uint) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("[%v]%v[%v]%v[%v]", mod.BitLen(), mod.Bytes(), gen.BitLen(), gen.Bytes(), twoOrder)))
	v := new(big.Int)
	v.SetBytes(hash[:])
	return path.Join(os.TempDir(), fmt.Sprintf("%v-%v", CacheFileRoot, v.Text(16)))
}

func saveRoots(p *rootsPrecomp) {
	fname := fileNameHash(p.Mod, p.Gen, p.TwoOrder)
	file, err := os.Create(fname)
	if err != nil {
		log.Fatalf("Error saving roots file: ", err)
	}
	encoder := gob.NewEncoder(file)
	encoder.Encode(p)
	file.Close()
}

func loadRoots(mod, gen *big.Int, twoOrder uint) *rootsPrecomp {
	log.Printf("Loading roots...")
	var out rootsPrecomp
	fname := fileNameHash(mod, gen, twoOrder)

	file, err := os.Open(fname)
	if err != nil {
		log.Printf("Could not open roots file: %v", err)
		return nil
	}
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&out)
	if err != nil {
		log.Printf("Error decoding: %v", err)
		return nil
	}
	file.Close()

	return &out
}

func init() {
	// This is a 265-bit modulus  --  2^18 | p - 1
	//IntModulus = fromString("2000000000000000000000000000000000000000000000000000000000000040001")
	//IntGen = fromString("1f4baf1f304cf6b688d71538d651e66357a4e30418ddd05c8d0b529fae93f55af")

	// This is a 102-bit modulus  --  2^19 | p-1
	//IntModulus = fromString("80000000000000000000080001")
	//IntGen = fromString("71a9f9595f292cfd55e4c5254e")

	// This is an 87-bit modulus  --  2^19 | p-1
	IntModulus = fromString("8000000000000000080001")
	IntGen = fromString("2597c14f48d5b65ed8dcca")

	// This is a 63-bit modulus   -- 2^19 | p-1
	//IntModulus = fromString("8000000000080001")
	//IntGen = fromString("22855fdf11374225")

	// This is a BOGUS 16-bit modulus
	//IntModulus = fromString("ffff")
	//IntGen = fromString("2")

	// This is a generator of order 2^19, so square it to get
	// one of order 2^18.
	IntGen.Mul(IntGen, IntGen)
	IntGen.Mod(IntGen, IntModulus)
	IntGen2Order = 18

	log.Printf("Working over field of %v bits", IntModulus.BitLen())
	precomp = loadRoots(IntModulus, IntGen, IntGen2Order)
	if precomp == nil {
		precomp = computeRoots(IntModulus, IntGen, IntGen2Order)
		saveRoots(precomp)
	}

}
