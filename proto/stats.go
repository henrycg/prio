package proto

import (
	"log"
	"time"
)

// Print statistics about client requets.
type stats struct {
	total uint
}

func (s *stats) Update(total uint) {
	s.total = total
}

// Make sure to run this inside its own goroutine, since
// it never returns.
func (s *stats) PrintEvery(d time.Duration) {
	var startCount uint
	for {
		startCount = s.total
		time.Sleep(d)
		delta := s.total - startCount

		log.Printf("Rate is: %5f req/sec [%v sec]", float64(delta)/float64(d.Seconds()), d.Seconds())
	}
}
