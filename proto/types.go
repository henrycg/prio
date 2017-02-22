package proto

import (
	"math/big"
	"net/rpc"
	"sync"

	"github.com/henrycg/prio/config"
	"github.com/henrycg/prio/mpc"
	"github.com/henrycg/prio/utils"
)

type Uuid [32]byte

// This is the object that we expose via the RPC.
// It hides the private server methods.
type PublicServer struct {
	server *Server
}

// This is the object that holds all of the
// state of the Prio server.
type Server struct {
	ServerIdx    int
	cfg          *config.Config
	LeaderClient *rpc.Client

	toProcess chan *UploadArgs

	// Indexed by sender public key... there should never
	// be collisions.
	pending      map[Uuid]*RequestStatus
	pendingMutex sync.RWMutex

	agg      []*mpc.Aggregator
	aggEpoch []uint
	aggMutex []sync.Mutex

	storedAgg      map[uint]*mpc.Aggregator
	storedAggCount map[uint]uint
	storedAggMutex sync.Mutex

	nProcessed      []uint
	nProcessedMutex []sync.Mutex
	nProcessedCond  []*sync.Cond

	verifPRGs []*utils.BufPRGReader

	pre          []*mpc.CheckerPrecomp
	randomX      []*big.Int
	randomXMutex []sync.Mutex

	pool []*checkerPool
}

// The leader coordinates the processing of each client submission.
// Every server is also a leader, but each client submission gets
// assigned to one particular leader.
type Leader struct {
	server     *Server
	rpcClients []*rpc.Client

	lastRequest      uint
	lastRequestMutex sync.Mutex

	amPublishingMutex sync.RWMutex

	statCounter stats
}

/*************************************
 * RPC Data structures
 */

type ServerCiphertext struct {
	Nonce      [24]byte // NaCl Box nonce
	Ciphertext []byte   // Encrypted upload payload
}

// Request from client to server to update the server
// DB state with fresh data from a client.
type UploadArgs struct {
	PublicKey   [32]byte // NaCl Box public key
	Ciphertexts []ServerCiphertext
}

type UploadReply struct {
}

type NewRequestArgs struct {
	RequestID  Uuid
	Ciphertext ServerCiphertext
}

type NewRequestReply struct {
}

type EvalCircuitArgs struct {
	RequestID Uuid
}

// First step of the MPC multiplication.
type EvalCircuitReply struct {
	CorShare *mpc.CorShare
}

// Second step of the MPC multiplication.
type FinalCircuitArgs struct {
	RequestID Uuid
	Cor       *mpc.Cor
	Key       *utils.PRGKey
}

// Decide whether to accept/reject client submission.
type AcceptArgs struct {
	RequestID Uuid
	Accept    bool
}

type AcceptReply struct {
}

type AggregateArgs struct {
	Server uint
	Serial uint
}

type AggregateReply struct {
}

type PublishArgs struct {
	Server uint
	Epoch  uint
	Agg    *mpc.Aggregator
}

type PublishReply struct {
}

type ChangePolyPointArgs struct {
	Server  int
	RandomX *big.Int
}

/*************************************/

type StatusFlag int

// Status of a client submission.
const (
	NotStarted    StatusFlag = iota
	OpenedTriples StatusFlag = iota
	Layer1        StatusFlag = iota
	Finished      StatusFlag = iota
)

type RequestStatus struct {
	check *mpc.Checker
	flag  StatusFlag
}
