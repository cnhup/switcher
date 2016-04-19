package main

import (
	"errors"
	"io"
	"log"
	"net"
	"time"
)

const (
	BUFFSIZE = 1024
)

type Mux struct {
	defaultAddress string
	connectTimeout time.Duration
	probeTimeout   time.Duration
	pm             ProtocolManager
}

// create a new Mux assignment
func NewMux() *Mux {
	return &Mux{
		connectTimeout: 2 * time.Second,
		probeTimeout:   2 * time.Second,
	}
}

// create a server on given address and handle incoming connections
func (m *Mux) ListenAndServe(addr string) error {
	server, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	for {
		conn, err := server.Accept()
		if err != nil {
			return err
		}

		go m.Serve(conn)
	}

	return nil
}

// serve takes an incomming connection, applies configured protocol
// handlers and proxies the connection based on result
func (m *Mux) Serve(conn net.Conn) error {
	defer conn.Close()

	probe := m.pm.Probe
	address := ""
	header := make([]byte, BUFFSIZE)

	nBuffed, matchResult := 0, UNMATCH
	for conn.SetReadDeadline(time.Now().Add(m.probeTimeout)); nBuffed < BUFFSIZE; {
		n, err := io.ReadAtLeast(conn, header[nBuffed:], 1)
		if err != nil {
			log.Printf("[ERROR] read error: %s\n", err)
			return err
		}

		nBuffed += n
		if matchResult, address = probe(header[:nBuffed]); matchResult != TRYAGAIN {
			break
		}
	}

	if address == "" {
		address = m.defaultAddress
		if address == "" {
			err := errors.New("none protocol matched and no default address")
			log.Printf("[ERROR] %s\n", err)
			return err
		}
	}

	log.Printf("[INFO] proxy: from=%s to=%s\n", conn.RemoteAddr(), address)

	// connect to remote
	remote, err := net.DialTimeout("tcp", address, m.connectTimeout)
	if err != nil {
		log.Printf("[ERROR] remote: %s\n", err)
		return err
	}
	defer remote.Close()

	// cancel read timeout
	conn.SetReadDeadline(time.Time{})

	// write header we chopped back to remote
	remote.Write(header[:nBuffed])

	// proxy between us and remote server
	err = Shovel(conn, remote)
	if err != nil {
		return err
	}

	return nil
}
