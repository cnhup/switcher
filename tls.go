package main

type TLS struct {
	BaseConfig
}

func (t *TLS) Probe(header []byte) (result ProbeResult, address string) {
	if len(header) < 3 {
		return TRYAGAIN, ""
	}

	if header[0] == 0x16 && header[1] == 3 && header[2] <= 3 {
		return MATCH, t.Address
	}

	return UNMATCH, ""
}
