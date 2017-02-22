package main

import (
	"bytes"
	"crypto/tls"
	"encoding/gob"
	"flag"
	"log"
	"net/rpc"
	"os"

	"github.com/henrycg/prio/config"
	"github.com/henrycg/prio/mpc"
	"github.com/henrycg/prio/proto"
	"github.com/henrycg/prio/utils"
)

func tryUpload(client *rpc.Client, args *proto.UploadArgs) error {
	//log.Print("Start upload")
	var upRes proto.UploadReply
	err := client.Call("PublicServer.Upload", args, &upRes)
	if err != nil {
		return err
	}

	//log.Print("Got message! ", upRes) return nil
	return nil
}

func connectToServer(cfg *config.Config, serverIdx int) *rpc.Client {
	certs := make([]tls.Certificate, 1)
	certs[0] = utils.ServerCertificates[serverIdx]
	serverAddr := cfg.Servers[serverIdx].Public
	client, err := rpc.Dial("tcp", serverAddr)
	log.Printf("Connected!")
	if err != nil {
		log.Fatal("Could not connect:", err)
		return nil
	}

	return client
}

func uploadOnce(cfg *config.Config, clients []*rpc.Client, args *proto.UploadArgs) {
	leaderIdx := proto.HashToServer(cfg, args.PublicKey)
	tryUpload(clients[leaderIdx], args)
}

func makeArgs(cfg *config.Config, nReqs int, req [][]*mpc.ClientRequest) []*proto.UploadArgs {
	args := make([]*proto.UploadArgs, nReqs)
	c := make(chan int, nReqs)
	ns := cfg.NumServers()

	var err error
	for i := 0; i < nReqs; i++ {
		go func(j int) {
			if req == nil {
				args[j], err = proto.GenUploadArgs(cfg, -1, nil)
			} else {
				leaderIdx := j % ns
				args[j], err = proto.GenUploadArgs(cfg, leaderIdx, req[leaderIdx])
			}
			c <- 1

			if err != nil {
				log.Fatal("Oh no:", err)
			}
		}(i)
	}

	for i := 0; i < nReqs; i++ {
		if i%10 == 0 {
			log.Printf("Build request %v", i)
		}
		<-c
	}

	return args
}

func makeDummyArgs(cfg *config.Config, nReqs int, req [][]*mpc.ClientRequest) []*proto.UploadArgs {
	args := make([]*proto.UploadArgs, nReqs)
	c := make(chan int, nReqs)

	var err error
	for i := 0; i < nReqs; i++ {
		go func(j int) {
			if req == nil {
				args[j], err = proto.GenDummyUploadArgs(cfg, nil)
			} else {
				args[j], err = proto.GenDummyUploadArgs(cfg, req[0])
			}
			c <- 1

			if err != nil {
				log.Fatal("Oh no:", err)
			}
		}(i)
	}

	for i := 0; i < nReqs; i++ {
		if i%10 == 0 {
			log.Printf("Build request %v", i)
		}
		<-c
	}

	return args
}

func runDummyClient(cfg *config.Config, nReqs int, req [][]*mpc.ClientRequest) {
	args := makeDummyArgs(cfg, nReqs, req)

	c := make(chan int, nReqs)
	var client *rpc.Client

	client = connectToServer(cfg, 0)

	log.Print("Done generating args")

	for i := 0; i < nReqs; i++ {
		go func(argsIn *proto.UploadArgs) {
			var upRes proto.UploadReply
			err := client.Call("NothingServer.Upload", argsIn, &upRes)
			if err != nil {
				panic("Error!")
			}
			c <- 1
		}(args[i])
	}

	for i := 0; i < nReqs; i++ {
		<-c
		if i%100 == 0 {
			log.Print("Processed request ", i)
		}
	}
}

func runClient(cfg *config.Config, nReqs int, req [][]*mpc.ClientRequest) {
	args := makeArgs(cfg, nReqs, req)

	c := make(chan int, nReqs)
	n := cfg.NumServers()
	clients := make([]*rpc.Client, n)

	for i := 0; i < n; i++ {
		go func(j int) {
			clients[j] = connectToServer(cfg, j)
			c <- 1
		}(i)
	}

	for i := 0; i < n; i++ {
		<-c
	}

	log.Print("Done generating args")

	for i := 0; i < nReqs; i++ {
		go func(argsIn *proto.UploadArgs) {
			uploadOnce(cfg, clients, argsIn)
			c <- 1
		}(args[i])
	}

	for i := 0; i < nReqs; i++ {
		<-c
		if i%100 == 0 {
			log.Print("Processed request ", i)
		}
	}
}

func writeReq(cfg *config.Config, outFile string) {
	out := make([][]*mpc.ClientRequest, cfg.NumServers())
	ns := cfg.NumServers()

	c := make(chan int, ns)
	// Make one request per server
	for s := 0; s < ns; s++ {
		go func(i int) {
			out[i] = mpc.RandomRequest(cfg, i)
			c <- 1
		}(s)
	}

	for s := 0; s < cfg.NumServers(); s++ {
		<-c
	}

	file, err := os.Create(outFile)
	defer file.Close()

	if err != nil {
		log.Fatalf("Could not open file: %v", err)
	}

	enc := gob.NewEncoder(file)
	err = enc.Encode(out)
	if err != nil {
		log.Fatalf("Could encode: %v", err)
	}
}

func readReq(inFile string) [][]*mpc.ClientRequest {
	if len(inFile) == 0 {
		return nil
	}

	file, err := os.Open(inFile)
	defer file.Close()

	if err != nil {
		log.Fatalf("Could not open file: %v", err)
	}

	dec := gob.NewDecoder(file)

	var reqs [][]*mpc.ClientRequest
	err = dec.Decode(&reqs)
	if err != nil {
		log.Fatalf("Could decode: %v", err)
	}

	return reqs
}

func requestOnce(cfg *config.Config) {
	var network bytes.Buffer
	out := mpc.RandomRequest(cfg, 0)
	enc := gob.NewEncoder(&network)
	err := enc.Encode(out)
	if err != nil {
		log.Fatal("Encode error")
	}

	//log.Printf("Size: %v", network.Len())
	//log.Printf("Size+SNARK: %v", network.Len()+288+32)
}

func main() {
	prof := flag.Bool("prof", false, "Write pprof file")
	forever := flag.Bool("forever", false, "Repeat forever")
	once := flag.Bool("once", false, "Generate one request and throw it away")
	onek := flag.Bool("onek", false, "Generate 1k requests and throw them away")
	nothing := flag.Bool("nothing", false, "Send dummy request")
	nReqs := flag.Int("n", 100, "Number of client requests")
	outFile := flag.String("outfile", "", "Write req to file")
	inFile := flag.String("infile", "", "Read req to file")
	cfgFile := flag.String("config", "", "Configuration file to use")
	logFile := flag.String("log", "", "Write log to file instead of stdout/stderr")

	flag.Parse()

	if *prof {
		utils.StartProfiling("client.prof")
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

	if *nReqs < 0 {
		log.Fatal("Number of requests must be non-negative integer")
	}

	if len(*outFile) != 0 && len(*inFile) != 0 {
		log.Fatal("Must either write or read (not both)")
	}

	cfg := config.LoadFile(*cfgFile)

	if *once {
		tStart := utils.GetUtime()
		requestOnce(cfg)
		tFinish := utils.GetUtime()
		log.Printf("Finished in %0.06f sec", float64((tFinish-tStart))/1000000000.0)
		return
	}

	if *onek {
		tStart := utils.GetUtime()
		for i := 0; i < 1000; i++ {
			requestOnce(cfg)
		}
		tFinish := utils.GetUtime()
		log.Printf("Finished in %0.06f sec", float64((tFinish-tStart))/1000000000.0/1000.0)
		return
	}

	for {
		if len(*outFile) != 0 {
			writeReq(cfg, *outFile)
		} else {
			reqs := readReq(*inFile)
			if *nothing {
				runDummyClient(cfg, *nReqs, reqs)
			} else {
				runClient(cfg, *nReqs, reqs)
			}
		}

		if !*forever {
			break
		}
	}
}
