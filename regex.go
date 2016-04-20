package main

import (
	"errors"
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

func (p *REGEX) Check() error {
	if len(p.Patterns) == 0 {
		return errors.New("at least one prefix pattern required")
	}

	for _, pattern := range p.Patterns {
		if len(pattern) == 0 {
			return errors.New("empty prefix pattern not allowed")
		}

		p.regexpList = append(p.regexpList, regexp.MustCompile(pattern))
	}

	return nil
}
