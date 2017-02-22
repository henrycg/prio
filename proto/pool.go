package proto

import (
	"github.com/henrycg/prio/config"
	"github.com/henrycg/prio/mpc"
)

// A pool of checker objects. We maintain this pool ourself
// to avoid initializing a new big Checker object for each
// client request. This also lets us control how many checkers
// each server creates, which gives us some rough control over
// how much memory the server uses.
type checkerPool struct {
	cfg       *config.Config
	serverIdx int
	leaderIdx int
	buffer    chan *mpc.Checker
}

func NewCheckerPool(cfg *config.Config, serverIdx int, leaderIdx int) *checkerPool {
	out := new(checkerPool)
	out.cfg = cfg
	out.serverIdx = serverIdx
	out.leaderIdx = leaderIdx

	out.buffer = make(chan *mpc.Checker, cfg.MaxPendingReqs)
	for i := 0; i < cfg.MaxPendingReqs; i++ {
		out.buffer <- mpc.NewChecker(cfg, serverIdx, leaderIdx)
	}
	return out
}

func (p *checkerPool) get() *mpc.Checker {
	select {
	case out := <-p.buffer:
		return out
		//	default:
		//		return mpc.NewChecker(p.cfg, p.serverIdx, p.leaderIdx)
	}
}

func (p *checkerPool) put(check *mpc.Checker) {
	select {
	case p.buffer <- check:
	default:
		// Do nothing
	}
}
