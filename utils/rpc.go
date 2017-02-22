package utils

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"log"
	"net"
	"net/rpc"
)

/* For running RPC over TCP. */
func ListenAndServe(server *rpc.Server, address string) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Listener error:", err)
		return
	}

	defer l.Close()

	for {
		log.Printf("Listening at %v", address)
		conn, err := l.Accept()

		if err != nil {
			log.Printf("Listener error: %v", err)
			continue
		}

		defer conn.Close()
		go server.ServeConn(conn)
	}
}

/* For running RPC over TLS. */
func ListenAndServeTLS(server *rpc.Server, address string,
	keyIdx int, acceptCerts []tls.Certificate, debugNet bool) {
	var config tls.Config
	if len(acceptCerts) > 0 {
		config.ClientAuth = tls.RequireAnyClientCert
	}
	config.InsecureSkipVerify = true
	config.Certificates = []tls.Certificate{ServerCertificates[keyIdx]}

	l, err := tls.Listen("tcp", address, &config)
	if err != nil {
		log.Fatal("Listener error:", err)
		return
	}

	defer l.Close()

	var socketCounter int
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Listener error: %v", err)
			continue
		}
		socketCounter++

		go handleOneClientTLS(conn, server, debugNet)
	}
}

func handleOneClientTLS(conn net.Conn, server *rpc.Server, debugNet bool) {
	defer conn.Close()

	tlscon, ok := conn.(*tls.Conn)
	if !ok {
		log.Printf("Could not cast conn")
		return
	}

	err := tlscon.Handshake()
	if err != nil {
		log.Printf("Handshake failed: %v", err)
		return
	}

	state := tlscon.ConnectionState()
	log.Printf("Certs %v", state.PeerCertificates)

	log.Printf("Handshake OK")

	var dConn io.ReadWriteCloser
	if debugNet {
		dConn = NewDebugConn(conn, "")
	} else {
		dConn = conn
	}

	server.ServeConn(dConn)
}

func DialHTTPWithTLS(network, address string,
	client_idx int, acceptCerts []tls.Certificate) (*rpc.Client, error) {
	var config tls.Config
	config.InsecureSkipVerify = true

	if client_idx >= 0 {
		config.Certificates = []tls.Certificate{ServerCertificates[client_idx]}
	}

	conn, err := tls.Dial(network, address, &config)
	if err != nil {
		log.Printf("DialHTTP error: %v", err)
		return nil, err
	}

	state := conn.ConnectionState()
	log.Printf("State: %v", state.PeerCertificates)
	if len(acceptCerts) > 0 && !validateCert(acceptCerts, state.PeerCertificates[0]) {
		return nil, errors.New("Invalid certificate")
	}

	return rpc.NewClient(conn), nil
}

func validateCert(acceptCerts []tls.Certificate, present *x509.Certificate) bool {
	for i := 0; i < len(acceptCerts); i++ {
		if acceptCerts[i].Leaf == nil {
			certs, err := x509.ParseCertificates(acceptCerts[i].Certificate[0])
			if err != nil {
				log.Printf("Could not parse cert: %v", err)
				return false
			}

			acceptCerts[i].Leaf = certs[0]
		}

		if acceptCerts[i].Leaf.Equal(present) {
			return true
		}
	}
	return false
}
