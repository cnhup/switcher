package main

import (
	"bytes"
)

type SSH struct {
	BaseConfig
}

// identify header as one of SSH
func (s SSH) Probe(header []byte) (result ProbeResult, address string) {
	// first 3 bytes of 1.0/2.0 is literal `SSH`
	if len(header) < 3 {
		return TRYAGAIN, ""
	}

	if bytes.Compare(header[:3], []byte("SSH")) == 0 {
		return MATCH, s.Address
	}

	return UNMATCH, ""
}
