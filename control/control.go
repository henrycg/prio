package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path"

	"github.com/henrycg/prio/config"
)

type commandFunc func(serverIdx int) string
type fileFunc func(serverIdx int, host string) []string

func getHost(server config.ServerAddress) string {
	host, _, err := net.SplitHostPort(server.Public)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	return host
}

func runRemote(command string, argsIn []string) {
	args := append(sshOptions, argsIn...)
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal("Error: ", err)
	}
}

func runAt(hosts []string, do commandFunc) {
	n := len(hosts)
	c := make(chan int, n)
	for i := 0; i < n; i++ {
		go func(j int) {
			runRemote("ssh", []string{hosts[j], do(j)})
			c <- 0
		}(i)
	}

	for i := 0; i < n; i++ {
		<-c
	}
}

func runAtServers(cfg *config.Config, do commandFunc) {
	n := len(cfg.Servers)
	hosts := make([]string, n)
	for i := 0; i < n; i++ {
		hosts[i] = getHost(cfg.Servers[i])
	}

	runAt(hosts, do)
}

func runAtClients(cfg *config.Config, do commandFunc) {
	n := len(cfg.Clients)
	hosts := make([]string, n)
	for i := 0; i < n; i++ {
		hosts[i] = string(cfg.Clients[i])
	}

	runAt(hosts, do)
}

func copyToHosts(hosts []string, src, dst fileFunc) {
	n := len(hosts)
	c := make(chan int, n)
	for i := 0; i < n; i++ {
		go func(j int) {
			args := append(src(j, hosts[j]), dst(j, hosts[j])...)
			runRemote("scp", args)
			c <- 0
		}(i)
	}

	for i := 0; i < n; i++ {
		<-c
	}

	log.Printf("Done.")
}

func copyToServers(cfg *config.Config, src, dst fileFunc) {
	servers := make([]string, len(cfg.Servers))
	for i, s := range cfg.Servers {
		servers[i] = getHost(s)
	}

	copyToHosts(servers, src, dst)
}

func copyToClients(cfg *config.Config, src, dst fileFunc) {
	clients := make([]string, len(cfg.Clients))
	for i, c := range cfg.Clients {
		clients[i] = string(c)
	}

	copyToHosts(clients, src, dst)
}

func runKill(cfg *config.Config) {
	runAtServers(cfg, func(int) string {
		return "killall -s INT tserver"
	})
}

func runKillClients(cfg *config.Config) {
	runAtClients(cfg, func(int) string {
		return "killall -s INT tclient"
	})
}

func runStart(cfg *config.Config) {
	runAtServers(cfg, func(i int) string {
		return fmt.Sprintf("~/tserver -idx %v -config ~/config.conf -log /tmp/log-%v.log", i, i)
	})
}

func runStartClients(cfg *config.Config) {
	runAtClients(cfg, func(i int) string {
		return fmt.Sprintf("~/tclient -config ~/config.conf -log /tmp/client-%v.log -outfile /tmp/request.dat", i)
	})

	runAtClients(cfg, func(i int) string {
		return fmt.Sprintf("~/tclient -config ~/config.conf -log /tmp/client-%v.log -infile /tmp/request.dat -n 600 -forever", i)
	})
}

func runStartDummy(cfg *config.Config) {
	runAtServers(cfg, func(i int) string {
		return fmt.Sprintf("~/tserver -nothing -idx %v -config ~/config.conf -log /tmp/log-%v.log", i, i)
	})
}

func runStartDummyClients(cfg *config.Config) {
	runAtClients(cfg, func(i int) string {
		return fmt.Sprintf("~/tclient -nothing -config ~/config.conf -log /tmp/client-%v.log -outfile /tmp/request.dat", i)
	})

	runAtClients(cfg, func(i int) string {
		return fmt.Sprintf("~/tclient -nothing -config ~/config.conf -log /tmp/client-%v.log -infile /tmp/request.dat -n 5000 -forever", i)
	})
}

func runLogs(cfg *config.Config) {
	src := func(k int, h string) []string { return []string{fmt.Sprintf("%v:/tmp/log-%v.log", h, k)} }
	dst := func(k int, h string) []string { return []string{fmt.Sprintf("log-%v.log", k)} }

	copyToServers(cfg, src, dst)
}

func runRmlogs(cfg *config.Config) {
	runAtServers(cfg, func(i int) string {
		return fmt.Sprintf("rm /tmp/log-%v.log", i)
	})
}

func runCopy(cfg *config.Config) {
	src := func(k int, h string) []string {
		return []string{
			path.Join(LOCAL_GOBIN, "tserver"),
			os.Args[1]}
	}

	clientSrc := func(k int, h string) []string {
		return []string{
			path.Join(LOCAL_GOBIN, "tclient"),
			os.Args[1]}
	}

	dst := func(k int, h string) []string {
		return []string{fmt.Sprintf("%v:~", h)}
	}

	copyToServers(cfg, src, dst)
	copyToClients(cfg, clientSrc, dst)

	runAtClients(cfg, func(int) string {
		return fmt.Sprintf("mv %v config.conf", path.Join("~", path.Base(os.Args[1])))
	})

	runAtServers(cfg, func(int) string {
		return fmt.Sprintf("mv %v config.conf", path.Join("~", path.Base(os.Args[1])))
	})
}

func runCopyConfig(cfg *config.Config) {
	src := func(k int, h string) []string {
		return []string{os.Args[1]}
	}

	clientSrc := func(k int, h string) []string {
		return []string{os.Args[1]}
	}

	dst := func(k int, h string) []string {
		return []string{fmt.Sprintf("%v:~", h)}
	}

	copyToServers(cfg, src, dst)
	copyToClients(cfg, clientSrc, dst)

	runAtClients(cfg, func(int) string {
		return fmt.Sprintf("mv %v config.conf", path.Join("~", path.Base(os.Args[1])))
	})

	runAtServers(cfg, func(int) string {
		return fmt.Sprintf("mv %v config.conf", path.Join("~", path.Base(os.Args[1])))
	})
}
