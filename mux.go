package main

import (
	"io"
	"log"
	"net"
)

type MatchResult int

const (
	MATCH MatchResult = iota
	UNMATCH
	TRYAGAIN
)

const (
	BUFFSIZE = 1024
)

type Protocol interface {
	// address to proxy to
	Address() string

	// identify protocol from header
	Identify(header []byte) MatchResult
}

type Mux struct {
	Handlers       []Protocol
	defaultAddress string
}

// create a new Mux assignment
func NewMux() *Mux {
	return &Mux{}
}

// add a protocol to mux handler set
func (m *Mux) Handle(p Protocol) {
	m.Handlers = append(m.Handlers, p)
}

// match protocol to handler
// returns address to proxy to
func (m *Mux) Identify(header []byte) (matchResult MatchResult, address string) {
	matchResult, address = UNMATCH, ""

	if len(m.Handlers) < 1 {
		return
	}

	for _, handler := range m.Handlers {
		switch handler.Identify(header) {
		case MATCH:
			matchResult, address = MATCH, handler.Address()
			return
		case TRYAGAIN:
			matchResult = TRYAGAIN
		}
	}

	if matchResult == UNMATCH {
		address = m.defaultAddress
	}

	return
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

	address := ""
	header := make([]byte, BUFFSIZE)

	nBuffed, matchResult := 0, UNMATCH
	for nBuffed < BUFFSIZE {
		n, err := io.ReadAtLeast(conn, header[nBuffed:], 1)
		nBuffed += n
		if err != nil {
			return err
		}

		if matchResult, address = m.Identify(header[:nBuffed]); matchResult != TRYAGAIN {
			break
		}
	}

	if address == "" {
		log.Println("[INFO] none protocol matched and no default address")
		return nil
	}

	log.Printf("[INFO] proxy: from=%s to=%s\n", conn.RemoteAddr(), address)

	// connect to remote
	remote, err := net.Dial("tcp", address)
	if err != nil {
		log.Printf("[ERROR] remote: %s\n", err)
		return err
	}
	defer remote.Close()

	// write header we chopped back to remote
	remote.Write(header[:nBuffed])

	// proxy between us and remote server
	err = Shovel(conn, remote)
	if err != nil {
		return err
	}

	return nil
}
