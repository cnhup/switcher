package main

import (
	"encoding/json"
	"errors"
)

type ProbeResult int

const (
	MATCH ProbeResult = iota
	UNMATCH
	TRYAGAIN
)

type BaseConfig struct {
	Address string `json:"addr"`
}

type Protocol interface {
	Probe(header []byte) (result ProbeResult, address string)
}

type ProtocolManager struct {
	Protocols []Protocol
}

func (pm *ProtocolManager) Probe(header []byte) (ProbeResult, string) {
	result, address := UNMATCH, ""
	for _, p := range pm.Protocols {
		switch result, address = p.Probe(header); result {
		case MATCH:
			return MATCH, address
		case TRYAGAIN:
			result = TRYAGAIN
		}
	}

	return result, address
}

func (pm *ProtocolManager) Register(p Protocol) {
	pm.Protocols = append(pm.Protocols, p)
}

func (pm *ProtocolManager) LoadProtocols(configs []json.RawMessage) error {
	for _, c := range configs {
		if p, err := createProtocol(c); err != nil {
			return err
		} else {
			pm.Register(p)
		}
	}

	return nil
}

type protocolService struct {
	BaseConfig
	Service string `json:"service"`
}

func createProtocol(data json.RawMessage) (Protocol, error) {
	var ps protocolService
	if err := json.Unmarshal(data, &ps); err != nil {
		return nil, err
	}

	if ps.Address == "" || ps.Service == "" {
		return nil, errors.New("service and addr are required for protocol")
	}

	switch service := ps.Service; service {
	case "mqtt":
		return &MQTT{BaseConfig: ps.BaseConfig}, nil
	case "ssh":
		return &SSH{BaseConfig: ps.BaseConfig}, nil
	case "regex":
		var p REGEX
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, err
		}
		if len(p.Patterns) == 0 {
			return nil, errors.New("at least one regex pattern required")
		}
		for _, pattern := range p.Patterns {
			if pattern == "" {
				return nil, errors.New("empty regex pattern not allowed")
			}
		}

		p.Compile()

		return &p, nil

	default:
		return nil, errors.New("invalid protocol: " + service)
	}
}
