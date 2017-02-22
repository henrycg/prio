package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"time"
)

import (
	"github.com/henrycg/prio/config"
	"github.com/henrycg/prio/proto"
	"github.com/henrycg/prio/utils"
)

func main() {
	prof := flag.Bool("prof", false, "Write CPU prof file")
	memprof := flag.Bool("memprof", false, "Write memory prof file")
	blockprof := flag.Bool("blockprof", false, "Write contention prof file")
	debugNet := flag.Bool("debug", false, "Print network debugging info")
	nothing := flag.Bool("nothing", false, "Run dummy server that does nothing")
	cfgFile := flag.String("config", "", "Configuration file to use")
	logFile := flag.String("log", "", "Write log to file instead of stdout/stderr")
	idx := flag.Int("idx", -1, "Server index")
	flag.Parse()

	cfg := config.LoadFile(*cfgFile)

	if *idx < 0 || *idx >= cfg.NumServers() {
		log.Fatal("Invalid index: ", *idx)
		return
	}

	log.SetPrefix(fmt.Sprintf("[Server %v] ", *idx))

	if *prof {
		utils.StartProfiling(fmt.Sprintf("server-%v.prof", *idx))
		defer utils.StopProfiling()
	}

	if len(*logFile) > 0 {
		f, err := os.Create(*logFile)
		if err != nil {
			log.Fatal("Could not open file: ", err)
		}

		defer f.Close()
		log.SetOutput(f)
	}

	if *memprof {
		utils.StartMemProfiling(fmt.Sprintf("server-%v.mprof", *idx))
	}

	if *blockprof {
		utils.StartBlockProfiling(fmt.Sprintf("server-%v.bprof", *idx))
	}

	// Allow connections from any client on this RPC server.
	_, pubPort, err := net.SplitHostPort(cfg.Servers[*idx].Public)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	pubAddr := "0.0.0.0:" + pubPort

	pubRPC := new(rpc.Server)
	if *nothing {
		log.Printf("Public nothing RPC %d is listening at %s", *idx, pubAddr)
		noServ := proto.NewNothingServer()
		pubRPC.Register(noServ)
		log.Printf("Public nothing RPC %d is listening at %s", *idx, pubAddr)
		utils.ListenAndServe(pubRPC, pubAddr)
	} else {
		privServer := proto.NewServer(cfg, *idx)
		pubServer := proto.NewPublicServer(privServer)
		l := proto.NewLeader(privServer)
		go func() {
			// Wait for servers to start up
			time.Sleep(3 * time.Second)
			l.Run()
		}()

		pubRPC.Register(pubServer)
		go utils.ListenAndServe(pubRPC, pubAddr)
		log.Printf("Public RPC %d is listening at %s", *idx, pubAddr)

		// This RPC server is for server-to-server communication only
		// and only TLS-authenticated accepts RPC requests from other servers.
		privRPC := new(rpc.Server)
		privRPC.Register(privServer)

		_, privPort, err := net.SplitHostPort(cfg.Servers[*idx].Private)
		if err != nil {
			log.Fatal("Error: ", err)
		}
		privAddr := "0.0.0.0:" + privPort
		log.Printf("Private RPC %d is listening at %s", *idx, privAddr)

		// TODO: Go's RPC system doesn't allow the receiver of a call to figure
		// out who the client is who made the call. Thus, server 0 can't tell
		// whether it was server 1 or 2 who made a particular call. A real-world
		// implementation would need to worry about this, but we are running
		// everything over TLS already so adding this extra separation will
		// not cause a performance hit.
		utils.ListenAndServeTLS(privRPC, privAddr, *idx, utils.ServerCertificates, *debugNet)
	}
}
