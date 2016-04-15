package main

import (
	"bytes"
)

type MQTT string

func (s MQTT) Address() string {
	return string(s)
}

func (s MQTT) Identify(header []byte) bool {
	if header[0] != 0x10 {
		return false
	}

	i := 1
	for ; ; i++ {
		if header[i]&0x80 == 0 {
			break
		}

		if i == 4 {
			return false
		}
	}

	i++

	return bytes.Compare(header[i:i+8], []byte("\x00\x06MQIsdp")) == 0 || bytes.Compare(header[i:i+6], []byte("\x00\x04MQTT")) == 0

	if bytes.Compare(header, []byte("MQTT")) == 0 {
		return true
	}

	return false
}
