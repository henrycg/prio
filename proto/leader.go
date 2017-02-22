package proto

import (
	"crypto/tls"
	"log"
	"net/rpc"
	"time"

	"github.com/henrycg/prio/mpc"
	"github.com/henrycg/prio/share"
	"github.com/henrycg/prio/utils"
)

/*************************
 * Misc Functions
 */

func (l *Leader) evalCircuits() {
	//log.Print("Running evalCircuits()")
	for {
		//log.Print("Waiting...")
		upload := <-l.server.toProcess

		//log.Printf("Processing %v", uuid)
		l.beginEvaluation(upload)
	}
}

func drainErrors(n int, c chan error) bool {
	for i := 0; i < n; i++ {
		err := <-c
		if err != nil {
			log.Print("Error: ", err)
			// TODO: Implement graceful error handling.
			panic("An error")
			return false
		}
	}

	return true
}

/*************************
 * RPC Calls
 */

func (l *Leader) runNewRequest(upload *UploadArgs) {
	n := l.server.cfg.NumServers()
	c := make(chan error, n)

	newReqArgs := make([]NewRequestArgs, n)
	for s := 0; s < n; s++ {
		newReqArgs[s].RequestID = upload.PublicKey
		newReqArgs[s].Ciphertext = upload.Ciphertexts[s]
	}

	newReqReplies := make([]NewRequestReply, n)

	for i := 0; i < n; i++ {
		go func(j int) {
			c <- l.rpcClients[j].Call("Server.NewRequest", &newReqArgs[j], &newReqReplies[j])
		}(i)
	}

	drainErrors(n, c)
}

func (l *Leader) runEvalCircuit(uuid Uuid) []*mpc.CorShare {
	n := l.server.cfg.NumServers()
	c := make(chan error, n)

	var evalCircuitArgs EvalCircuitArgs
	evalCircuitArgs.RequestID = uuid

	evalReplies := make([]*mpc.CorShare, n)

	for i := 0; i < n; i++ {
		go func(j int) {
			c <- l.rpcClients[j].Call("Server.EvalCircuit", evalCircuitArgs, &evalReplies[j])
		}(i)
	}

	drainErrors(n, c)

	return evalReplies
}

func (l *Leader) runFinalCircuit(uuid Uuid, check *mpc.Checker, corShares []*mpc.CorShare) []*mpc.OutShare {
	n := l.server.cfg.NumServers()
	c := make(chan error, n)

	var finalCircuitArgs FinalCircuitArgs
	finalCircuitArgs.RequestID = uuid
	finalCircuitArgs.Cor = check.Cor(corShares)
	finalCircuitArgs.Key = utils.RandomPRGKey()

	finalReplies := make([]*mpc.OutShare, n)
	for i := 0; i < n; i++ {
		go func(j int) {
			c <- l.rpcClients[j].Call("Server.FinalCircuit", finalCircuitArgs, &finalReplies[j])
		}(i)
	}

	drainErrors(n, c)

	return finalReplies
}

func (l *Leader) runAccept(uuid Uuid, check *mpc.Checker, outShares []*mpc.OutShare) {
	n := l.server.cfg.NumServers()
	c := make(chan error, n)

	var acceptArgs AcceptArgs
	acceptArgs.RequestID = uuid
	acceptArgs.Accept = check.OutputIsValid(outShares)
	if !acceptArgs.Accept {
		log.Printf("Warning: rejecting request with ID %v", uuid)
	}

	if acceptArgs.Accept {
		l.lastRequestMutex.Lock()
		l.lastRequest++
		l.lastRequestMutex.Unlock()
	}

	acceptReplies := make([]AcceptReply, n)
	for i := 0; i < n; i++ {
		go func(j int) {
			c <- l.rpcClients[j].Call("Server.Accept", acceptArgs, &acceptReplies[j])
		}(i)
	}

	drainErrors(n, c)
}

// Main goroutine for processing a single client request.
func (l *Leader) beginEvaluation(upload *UploadArgs) {
	//tStart := time.Now().UnixNano()
	//log.Printf("Begin eval of: %v", uuid)

	// Wait until we are done publishing data from this batch
	// of requests to start working on the next batch.
	l.amPublishingMutex.RLock()

	l.runNewRequest(upload)
	uuid := upload.PublicKey

	l.server.pendingMutex.RLock()
	v, ok := l.server.pending[uuid]
	l.server.pendingMutex.RUnlock()
	check := v.check
	//log.Printf("Got pending mutex")
	if !ok {
		log.Fatal("Should never get here")
	}

	//log.Printf("Run eval circuit: %v", uuid)
	// Send eval request to all servers
	evalReplies := l.runEvalCircuit(uuid)

	//log.Printf("Run final circuit: %v", uuid)
	finalReplies := l.runFinalCircuit(uuid, check, evalReplies)

	//log.Printf("Run accept: %v", uuid)
	l.runAccept(uuid, check, finalReplies)

	//log.Printf("Done evaluating %v", uuid)
	l.amPublishingMutex.RUnlock()

}

func (l *Leader) printTotalProcessed() {

	// The sum might be a little bit off because of races
	sum := uint(0)
	for i := 0; i < l.server.cfg.NumServers(); i++ {
		sum += l.server.nProcessed[i]
	}
	l.statCounter.Update(sum)

	log.Printf("Completed: %v", sum)
}

// Goroutine that mananges combining state of the
// different servers at the end of an epoch.
func (l *Leader) aggregate() {
	log.Print("Running aggregate()")
	n := l.server.cfg.NumServers()
	c := make(chan error, n)

	// Prevent other goroutines from processing client requests while
	// we are publishing.
	l.amPublishingMutex.Lock()

	var args AggregateArgs
	args.Serial = l.lastRequest
	args.Server = uint(l.server.ServerIdx)
	replies := make([]PublishReply, n)

	for i := 0; i < n; i++ {
		go func(j int) {
			c <- l.rpcClients[j].Call("Server.Aggregate", args, &replies[j])
		}(i)
	}

	drainErrors(n, c)
	l.amPublishingMutex.Unlock()

	l.server.nProcessedMutex[l.server.ServerIdx].Lock()
	sum := uint(0)
	for i := 0; i < n; i++ {
		sum += l.server.nProcessed[i]
	}
	log.Printf("NProcessed: %v", sum)
	l.server.nProcessedMutex[l.server.ServerIdx].Unlock()
}

func (l *Leader) changePoint() {
	log.Print("Changing MPC evaluation point")
	n := l.server.cfg.NumServers()
	c := make(chan error, n)

	// Prevent other goroutines from processing client requests while
	// we are changing the MPC evaluation point.
	l.amPublishingMutex.Lock()

	var args ChangePolyPointArgs
	args.Server = l.server.ServerIdx
	args.RandomX = utils.RandInt(share.IntModulus)
	replies := make([]int, n)

	for i := 0; i < n; i++ {
		go func(j int) {
			c <- l.rpcClients[j].Call("Server.ChangePolyPoint", args, &replies[j])
		}(i)
	}

	drainErrors(n, c)
	l.amPublishingMutex.Unlock()

	log.Printf("Changed evaluation point")
}

func (l *Leader) connectToServer(client **rpc.Client, serverAddr string, remoteIdx int, c chan error) {
	var err error
	certs := []tls.Certificate{utils.ServerCertificates[remoteIdx]}
	*client, err = utils.DialHTTPWithTLS("tcp", serverAddr, l.server.ServerIdx, certs)

	c <- err
}

func (l *Leader) openConnections() error {
	ns := l.server.cfg.NumServers()
	l.rpcClients = make([]*rpc.Client, ns)

	c := make(chan error, ns)
	servers := l.server.cfg.Servers
	for i := 0; i < ns; i++ {
		go l.connectToServer(&l.rpcClients[i], servers[i].Private, i, c)
	}

	// Wait for all connections
	drainErrors(ns, c)

	l.server.LeaderClient = l.rpcClients[0]

	return nil
}

func NewLeader(server *Server) *Leader {
	l := new(Leader)
	l.server = server

	return l
}

// Start up the server.
func (l *Leader) Run() {
	err := l.openConnections()
	if err != nil {
		log.Fatal("Leader could not connect to servers. ", err)
	}

	// Set servers to use a random point
	l.changePoint()

	//go utils.RunForever(l.evalCircuits, 100*time.Millisecond)
	for i := 0; i < l.server.cfg.MaxPendingReqs; i++ {
		go l.evalCircuits()
	}
	go utils.RunForever(l.aggregate, 180*time.Second)
	go utils.RunForever(l.changePoint, 180*time.Second)
	go l.statCounter.PrintEvery(5 * time.Second)
	go l.statCounter.PrintEvery(30 * time.Second)
	go l.statCounter.PrintEvery(60 * time.Second)
	utils.RunForever(l.printTotalProcessed, 3*time.Second)
}
