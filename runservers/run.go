// Program that starts up a set of Prio servers on your
// local machine. Convenient for development/testing.
package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/henrycg/prio/config"
)

func runServer(idx int, profile bool, memProfile bool, blockProfile bool, cfgFile string, debugNet bool, nothing bool) *exec.Cmd {
	args := []string{"./tserver", "-idx", strconv.Itoa(idx)}
	if profile {
		args = append(args, "-prof")
	}

	if memProfile {
		args = append(args, "-memprof")
	}

	if blockProfile {
		args = append(args, "-blockprof")
	}

	if debugNet {
		args = append(args, "-debug")
	}

	if nothing {
		args = append(args, "-nothing")
	}

	if len(cfgFile) > 0 {
		args = append(args, "-config", cfgFile)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Print("Error: ", err)
	}

	return cmd
}

func main() {

	prof := flag.Bool("prof", false, "Write CPU profile files")
	memprof := flag.Bool("memprof", false, "Write memory profile file")
	blockprof := flag.Bool("blockprof", false, "Write contention profile file")
	cfgFile := flag.String("config", "", "Configuration file to use")
	debugNet := flag.Bool("debug", false, "Print net debug info")
	nothing := flag.Bool("nothing", false, "Run dummy server")
	flag.Parse()

	cfg := config.LoadFile(*cfgFile)
	servers := make([]*exec.Cmd, cfg.NumServers())
	for i, _ := range cfg.Servers {
		servers[i] = runServer(i, *prof, *memprof, *blockprof, *cfgFile, *debugNet, *nothing)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	// On Ctrl-C, kill all of the server processes
	go func(toClean []*exec.Cmd) {
		<-c
		for i := 0; i < len(toClean); i++ {
			if toClean[i].ProcessState != nil && !toClean[i].ProcessState.Exited() {
				toClean[i].Process.Kill()
			}
		}
		os.Exit(1)
	}(servers)

	for {
		time.Sleep(1000000)
	}
}
