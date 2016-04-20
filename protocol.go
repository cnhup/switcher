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
		if p, _, err := createProtocol(c); err != nil {
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

func createProtocol(data json.RawMessage) (p Protocol, service string, err error) {
	var ps protocolService
	if err = json.Unmarshal(data, &ps); err != nil {
		return
	}

	if ps.Address == "" || ps.Service == "" {
		err = errors.New("service and addr are required for protocol")
		return
	}

	switch service = ps.Service; service {
	case "mqtt":
		p = &MQTT{BaseConfig: ps.BaseConfig}
	case "ssh":
		p = &SSH{BaseConfig: ps.BaseConfig}
	case "regex":
		re := new(REGEX)
		if err = json.Unmarshal(data, re); err != nil {
			return
		}
		if err = re.Check(); err != nil {
			return
		}
		p = re
	case "prefix":
		prefix := new(PREFIX)
		if err = json.Unmarshal(data, prefix); err != nil {
			return
		}
		if err = prefix.Check(); err != nil {
			return
		}
		p = prefix
	default:
		err = errors.New("invalid protocol: " + service)
	}

	return
}
