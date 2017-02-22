package proto

import (
	"sync"
	"time"
)

// A server that does nothing, just for performance
// comparison purposes.
type NothingServer struct {
	done      uint
	doneMutex sync.Mutex
	counter   stats
}

func NewNothingServer() *NothingServer {
	s := new(NothingServer)
	go s.counter.PrintEvery(10 * time.Second)
	return s
}

func (s *NothingServer) Upload(args *UploadArgs, reply *UploadReply) error {
	var uuid Uuid
	uuid = args.PublicKey
	_, err := decryptRequest(0, &uuid, &args.Ciphertexts[0])
	if err != nil {
		panic("Error")
	}

	s.doneMutex.Lock()
	s.done += 1
	s.counter.Update(s.done)
	s.doneMutex.Unlock()
	return nil
}
