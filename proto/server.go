package proto

import (
	"errors"
	"log"
	"math/big"
	"sync"

	"github.com/henrycg/prio/config"
	"github.com/henrycg/prio/mpc"
)

func NewServer(cfg *config.Config, idx int) *Server {
	s := new(Server)
	s.ServerIdx = idx
	s.toProcess = make(chan *UploadArgs)
	s.pending = make(map[Uuid]*RequestStatus)
	s.cfg = cfg

	ns := cfg.NumServers()
	s.agg = make([]*mpc.Aggregator, ns)
	s.aggEpoch = make([]uint, ns)
	s.aggMutex = make([]sync.Mutex, ns)

	s.storedAgg = make(map[uint]*mpc.Aggregator)
	s.storedAggCount = make(map[uint]uint)

	s.nProcessed = make([]uint, ns)
	s.nProcessedMutex = make([]sync.Mutex, ns)
	s.nProcessedCond = make([]*sync.Cond, ns)

	s.pre = make([]*mpc.CheckerPrecomp, ns)
	for i := 0; i < ns; i++ {
		s.pre[i] = mpc.NewCheckerPrecomp(s.cfg)
	}

	for i := 0; i < ns; i++ {
		s.agg[i] = mpc.NewAggregator(s.cfg)
		s.nProcessedCond[i] = sync.NewCond(&s.nProcessedMutex[i])
	}

	s.randomX = make([]*big.Int, cfg.NumServers())
	s.randomXMutex = make([]sync.Mutex, ns)

	s.pool = make([]*checkerPool, cfg.NumServers())
	for leaderIdx := 0; leaderIdx < cfg.NumServers(); leaderIdx++ {
		s.pool[leaderIdx] = NewCheckerPool(s.cfg, s.ServerIdx, leaderIdx)
	}

	return s
}

func (s *Server) isLeader() bool {
	return s.ServerIdx == 0
}

func (s *Server) NewRequest(args *NewRequestArgs, reply *NewRequestReply) error {
	// Add request to queue
	r, err := decryptRequest(s.ServerIdx, &args.RequestID, &args.Ciphertext)
	if err != nil {
		log.Print("Could not decrypt insert args")
		return err
	}

	dstServer := HashToServer(s.cfg, args.RequestID)

	s.pendingMutex.RLock()
	exists := s.pending[args.RequestID] != nil
	s.pendingMutex.RUnlock()

	if exists {
		log.Print(s.pending[args.RequestID])
		log.Print("Error: Key collision! Ignoring bogus request.")
		return nil
	}

	status := new(RequestStatus)
	status.check = s.pool[dstServer].get()
	status.check.SetReq(r)
	if status.check == nil {
		log.Printf("Warning: ignoring invalid client request (%v)", args.RequestID)
		return nil
	}

	status.flag = NotStarted

	s.pendingMutex.Lock()
	s.pending[args.RequestID] = status
	s.pendingMutex.Unlock()

	return nil
}

func (s *Server) EvalCircuit(args *EvalCircuitArgs, reply *mpc.CorShare) error {
	leader := HashToServer(s.cfg, args.RequestID)

	s.pendingMutex.RLock()
	status, okay := s.pending[args.RequestID]
	s.pendingMutex.RUnlock()
	if !okay {
		return errors.New("Could not find specified request")
	}

	if status.flag != NotStarted {
		return errors.New("Request already processed")
	}

	s.randomXMutex[leader].Lock()
	if s.randomX[leader] != nil {
		s.pre[leader].SetCheckerPrecomp(s.randomX[leader])
	}
	s.randomX[leader] = nil
	s.randomXMutex[leader].Unlock()

	status.flag = Layer1
	status.check.CorShare(reply, s.pre[leader])

	//log.Print("Done evaluating ", args.RequestID)
	return nil
}

func (s *Server) FinalCircuit(args *FinalCircuitArgs, reply *mpc.OutShare) error {

	s.pendingMutex.RLock()
	status, okay := s.pending[args.RequestID]
	s.pendingMutex.RUnlock()
	if !okay {
		return errors.New("Could not find specified request")
	}

	if status.flag != Layer1 {
		return errors.New("Request already processed")
	}
	status.flag = Finished

	status.check.OutShare(reply, args.Cor, args.Key)

	return nil
}

func (s *Server) Accept(args *AcceptArgs, reply *AcceptReply) error {
	s.pendingMutex.RLock()
	status, okay := s.pending[args.RequestID]
	s.pendingMutex.RUnlock()
	if !okay {
		return errors.New("Could not find specified request")
	}

	if status.flag != Finished {
		return errors.New("Request not yet processed")
	}

	s.pendingMutex.Lock()
	delete(s.pending, args.RequestID)
	s.pendingMutex.Unlock()

	l := HashToServer(s.cfg, args.RequestID)
	if args.Accept {
		s.aggMutex[l].Lock()
		s.agg[l].Update(status.check)
		s.aggMutex[l].Unlock()

		s.nProcessedCond[l].Signal()
		s.nProcessedMutex[l].Lock()
		s.nProcessed[l]++
		s.nProcessedMutex[l].Unlock()
		s.nProcessedCond[l].Signal()
	}

	//log.Printf("Done!")
	s.pool[l].put(status.check)

	return nil
}

func (s *Server) Aggregate(args *AggregateArgs, reply *AggregateReply) error {
	if args.Server >= uint(s.cfg.NumServers()) {
		return errors.New("Bogus server id")
	}

	// TODO: Check that number of aggregated values is large!
	// In other words, we want a large anonymity set!
	s.nProcessedMutex[args.Server].Lock()
	for s.nProcessed[args.Server] < args.Serial {
		s.nProcessedCond[args.Server].Wait()
	}

	s.aggMutex[args.Server].Lock()

	var pargs PublishArgs
	pargs.Server = uint(s.ServerIdx)
	pargs.Epoch = s.aggEpoch[args.Server]
	pargs.Agg = s.agg[args.Server].Copy()

	s.aggEpoch[args.Server]++

	// Wait until we have received an Accept() request
	// from all other servers.
	ready := true
	log.Printf("Epochs: %v", s.aggEpoch)
	for i := 0; i < s.cfg.NumServers(); i++ {
		if s.aggEpoch[i] < s.aggEpoch[args.Server] {
			ready = false
		}
	}

	if ready {
		var pubReply PublishReply
		err := s.LeaderClient.Call("Server.Publish", &pargs, &pubReply)
		if err != nil {
			log.Fatalf("Publish error: %v", err)
		}
	}

	// To be conservative, hold the lock until the publish operation
	// completes.
	s.agg[args.Server].Reset()
	s.aggMutex[args.Server].Unlock()

	s.nProcessedMutex[args.Server].Unlock()

	return nil
}

func (s *Server) Publish(args *PublishArgs, reply *PublishReply) error {
	if !s.isLeader() {
		return errors.New("Am not leader!")
	}

	s.storedAggMutex.Lock()
	_, okay := s.storedAgg[args.Epoch]
	if !okay {
		s.storedAgg[args.Epoch] = mpc.NewAggregator(s.cfg)
		s.storedAggCount[args.Epoch] = 0
	}

	s.storedAgg[args.Epoch].Combine(args.Agg)
	s.storedAggCount[args.Epoch]++

	if s.storedAggCount[args.Epoch] == uint(s.cfg.NumServers()) {
		log.Print(s.storedAgg[args.Epoch])
		delete(s.storedAgg, args.Epoch)
		delete(s.storedAggCount, args.Epoch)
	}

	s.storedAggMutex.Unlock()

	return nil
}

func (s *Server) ChangePolyPoint(args *ChangePolyPointArgs, reply *int) error {
	if args.Server >= s.cfg.NumServers() {
		return errors.New("Bogus server ID")
	}

	s.randomXMutex[args.Server].Lock()
	s.randomX[args.Server] = args.RandomX
	s.randomXMutex[args.Server].Unlock()
	return nil
}
