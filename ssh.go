package main

import (
	"bytes"
)

type SSH string

// address to proxy to
func (s SSH) Address() string {
	return string(s)
}

// identify header as one of SSH
func (s SSH) Identify(header []byte) MatchResult {
	// first 3 bytes of 1.0/2.0 is literal `SSH`
	if len(header) < 3 {
		return TRYAGAIN
	}

	if bytes.Compare(header[:3], []byte("SSH")) == 0 {
		return MATCH
	}

	return UNMATCH
}

func NewSSH(config BaseConfig) Protocol {
	return SSH(config.Address)
}
