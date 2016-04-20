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
	var prefix_list []Protocol
	for _, c := range configs {
		switch p, service, err := createProtocol(c); {
		case err != nil:
			return err
		case service == "prefix":
			prefix_list = append(prefix_list, p)
		default:
			pm.Register(p)
		}
	}

	if len(prefix_list) > 0 {
		tree := NewMatchTree()

		for _, p := range prefix_list {
			prefix := p.(*PREFIX)
			tree.Add(prefix)
		}

		pm.Protocols = append([]Protocol{tree}, pm.Protocols...)
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
		service = "prefix"
		p = &PREFIX{ps.BaseConfig, []string{"SSH"}}
	case "http":
		service = "prefix"
		p = &PREFIX{ps.BaseConfig, []string{"GET ", "POST ", "PUT ", "DELETE ", "HEAD ", "OPTIONS "}}
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
