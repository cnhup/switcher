package main

import (
	"errors"
	"io"
	"log"
	"net"
	"time"
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
	connectTimeout time.Duration
	detectTimeout  time.Duration
}

// create a new Mux assignment
func NewMux() *Mux {
	return &Mux{
		connectTimeout: 2 * time.Second,
		detectTimeout:  2 * time.Second,
	}
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
	for conn.SetReadDeadline(time.Now().Add(m.detectTimeout)); nBuffed < BUFFSIZE; {
		n, err := io.ReadAtLeast(conn, header[nBuffed:], 1)
		if err != nil {
			log.Printf("[ERROR] read error: %s\n", err)
			return err
		}

		nBuffed += n
		if matchResult, address = m.Identify(header[:nBuffed]); matchResult != TRYAGAIN {
			break
		}
	}

	if address == "" {
		err := errors.New("none protocol matched and no default address")
		log.Printf("[ERROR] %s\n", err)
		return err
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
