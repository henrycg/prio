package share

import (
	"math/big"

	"github.com/henrycg/prio/utils"
)

// Compressed representation of secret-shared data.
type PRGHints struct {
	Key   utils.PRGKey
	Delta []*big.Int
}

// A server uses a ReplayPRG to recover the shared values
// that the client sent it (in the form of a PRGHints struct).
type ReplayPRG struct {
	serverIdx int
	leaderIdx int

	rand  *utils.BufPRGReader
	hints *PRGHints
	cur   int
}

// A client uses a GenPRG to split values into shares
// (one share per server) using a PRG to compress the
// shares.
type GenPRG struct {
	nServers  int
	leaderIdx int

	rand  []*utils.BufPRGReader
	delta []*big.Int
}

// Produce a new ReplayPRG object for the given server/leader combo.
func NewReplayPRG(serverIdx int, leaderIdx int) *ReplayPRG {
	out := new(ReplayPRG)
	out.leaderIdx = leaderIdx
	out.serverIdx = serverIdx

	return out
}

// Import the compressed secret-shared values from hints.
func (p *ReplayPRG) Import(hints *PRGHints) {
	p.hints = hints
	p.rand = utils.NewBufPRG(utils.NewPRG(&p.hints.Key))
	p.cur = 0
}

// Recover a secret-shared value that is shared in a field
// that uses modulus mod.
func (p *ReplayPRG) Get(mod *big.Int) *big.Int {
	out := p.rand.RandInt(mod)
	if p.IsLeader() {
		out.Add(out, p.hints.Delta[p.cur])
		out.Mod(out, mod)
	}
	p.cur++

	return out
}

func (p *ReplayPRG) IsLeader() bool {
	return p.serverIdx == p.leaderIdx
}

// Create a new GenPRG object for producing compressed secret-shared values.
func NewGenPRG(nServers int, leaderIdx int) *GenPRG {
	out := new(GenPRG)
	out.nServers = nServers
	out.leaderIdx = leaderIdx

	out.rand = make([]*utils.BufPRGReader, nServers)
	for i := 0; i < nServers; i++ {
		out.rand[i] = utils.NewBufPRG(utils.RandomPRG())
	}
	out.delta = make([]*big.Int, 0)

	return out
}

// Split value into shares using modulus mod.
func (g *GenPRG) Share(mod *big.Int, value *big.Int) []*big.Int {
	out := make([]*big.Int, g.nServers)
	delta := new(big.Int)
	for i := 0; i < g.nServers; i++ {
		out[i] = g.rand[i].RandInt(mod)
		delta.Add(delta, out[i])
	}
	delta.Sub(value, delta)
	delta.Mod(delta, mod)

	g.delta = append(g.delta, delta)

	out[g.leaderIdx].Add(out[g.leaderIdx], delta)
	out[g.leaderIdx].Mod(out[g.leaderIdx], mod)

	return out
}

// Generate the hints that serverIdx can use to recover the shares.
func (g *GenPRG) Hints(serverIdx int) *PRGHints {
	out := new(PRGHints)
	out.Key = g.rand[serverIdx].Key

	if serverIdx == g.leaderIdx {
		out.Delta = g.delta
	}
	return out
}
