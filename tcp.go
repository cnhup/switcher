package main

type TCP string

// address to proxy to
func (t TCP) Address() string {
	return string(t)
}

// identify header as one of TCP
func (t TCP) Identify(header []byte) MatchResult {
	// this is a dummy protocol handler used for the default
	return UNMATCH
}
