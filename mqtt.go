package main

import (
	"bytes"
)

type MQTT struct {
	BaseConfig
}

func (s *MQTT) Probe(header []byte) (result ProbeResult, address string) {
	if header[0] != 0x10 {
		return UNMATCH, ""
	}

	if len(header) < 13 {
		return TRYAGAIN, ""
	}

	i := 1
	for ; ; i++ {
		if header[i]&0x80 == 0 {
			break
		}

		if i == 4 {
			return UNMATCH, ""
		}
	}

	i++

	if bytes.Compare(header[i:i+8], []byte("\x00\x06MQIsdp")) == 0 || bytes.Compare(header[i:i+6], []byte("\x00\x04MQTT")) == 0 {
		return MATCH, s.Address
	}

	return UNMATCH, ""
}
