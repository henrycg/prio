package share

import (
	"math/big"

	"github.com/henrycg/prio/utils"
)

// Split hte value secret into nPieces shares modulo mod.
func Share(mod *big.Int, nPieces int, secret *big.Int) []*big.Int {
	if nPieces == 0 {
		panic("Number of shares must be at least 1")
	} else if nPieces == 1 {
		return []*big.Int{secret}
	}

	out := make([]*big.Int, nPieces)

	acc := new(big.Int)
	for i := 0; i < nPieces-1; i++ {
		out[i] = utils.RandInt(mod)

		acc.Add(acc, out[i])
	}

	acc.Sub(secret, acc)
	acc.Mod(acc, mod)
	out[nPieces-1] = acc

	return out
}
