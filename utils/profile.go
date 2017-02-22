package utils

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
)

// #include <time.h>
import "C"

const nanoSeconds = uint64(1000000000)

func GetUtime() uint64 {
	// Inspired by
	// https://gist.github.com/christopherobin/9247060
	var ts C.struct_timespec

	// From manpage:
	// Per-process CPU-time clock (measures CPU time consumed by all
	// threads in the process).
	err := C.clock_gettime(C.CLOCK_PROCESS_CPUTIME_ID, &ts)
	if err > 0 {
		log.Fatal("clock_getttime error code: ", err)
	}

	return (uint64(ts.tv_sec) * nanoSeconds) + uint64(ts.tv_nsec)
}

func PrintTime(tag string) {
	log.Printf("%v %v", tag, GetUtime())
}

func StartProfiling(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)

	// Stop on ^C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)
	go func() {
		for _ = range c {
			// sig is a ^C, handle it
			pprof.StopCPUProfile()
			os.Exit(0)
		}
	}()
}

func StopProfiling() {
	// Stop when process exits
	pprof.StopCPUProfile()
}

func writeMemProfile(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Writing memory profile")
	pprof.WriteHeapProfile(f)
	f.Close()
}

func StartMemProfiling(filename string) {
	// Stop on ^C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			// sig is a ^C, handle it
			writeMemProfile(filename)
			os.Exit(0)
		}
	}()
}

func writeBlockProfile(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Writing block profile")
	pprof.Lookup("block").WriteTo(f, 0)
	f.Close()
}

func StartBlockProfiling(filename string) {
	// Stop on ^C
	runtime.SetBlockProfileRate(1)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			// sig is a ^C, handle it
			writeBlockProfile(filename)
			os.Exit(0)
		}
	}()
}
