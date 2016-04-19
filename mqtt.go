package main

import (
	"bytes"
)

type MQTT string

func (s MQTT) Address() string {
	return string(s)
}

func (s MQTT) Identify(header []byte) MatchResult {
	if header[0] != 0x10 {
		return UNMATCH
	}

	if len(header) < 13 {
		return TRYAGAIN
	}

	i := 1
	for ; ; i++ {
		if header[i]&0x80 == 0 {
			break
		}

		if i == 4 {
			return UNMATCH
		}
	}

	i++

	if bytes.Compare(header[i:i+8], []byte("\x00\x06MQIsdp")) == 0 || bytes.Compare(header[i:i+6], []byte("\x00\x04MQTT")) == 0 {
		return MATCH
	}

	return UNMATCH
}

type MQTTConfig BaseConfig

func (c *MQTTConfig) NewProtocol() Protocol {
	return MQTT(c.Address)
}
