package main

import (
	"regexp"
)

type REGEX struct {
	BaseConfig
	Patterns   []string `json:"patterns"`
	MinLength  int      `json:"minlen"`
	MaxLength  int      `json:"maxlen"`
	regexpList []*regexp.Regexp
}

func (p *REGEX) Probe(header []byte) (result ProbeResult, address string) {
	if p.MinLength > 0 && len(header) < p.MinLength {
		return TRYAGAIN, ""
	}
	for _, re := range p.regexpList {
		if re.Match(header) {
			return MATCH, p.Address
		}
	}

	if p.MaxLength > 0 && len(header) >= p.MaxLength {
		return UNMATCH, ""
	}

	return TRYAGAIN, ""
}

func (p *REGEX) Compile() {
	for _, pattern := range p.Patterns {
		p.regexpList = append(p.regexpList, regexp.MustCompile(pattern))
	}
}
