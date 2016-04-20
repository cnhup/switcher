package main

import (
	"bytes"
	"errors"
)

type PREFIX struct {
	BaseConfig
	Patterns []string `json:"patterns"`
}

func (p *PREFIX) Probe(header []byte) (ProbeResult, string) {
	result := UNMATCH
	for _, pattern := range p.Patterns {
		if bytes.HasPrefix(header, []byte(pattern)) {
			return MATCH, p.Address
		}

		if len(header) < len(pattern) {
			result = TRYAGAIN
		}
	}

	return result, ""
}

func (p *PREFIX) Check() error {
	if len(p.Patterns) == 0 {
		return errors.New("at least one prefix pattern required")
	}

	for _, pattern := range p.Patterns {
		if len(pattern) == 0 {
			return errors.New("empty prefix pattern not allowed")
		}
	}

	return nil
}
