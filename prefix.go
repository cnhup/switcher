package main

import (
	"bytes"
	"errors"
)

type PREFIX struct {
	BaseConfig
	Patterns []string `json:"patterns"`
}

func (p *PREFIX) Probe(header []byte) (ProbeResult, string) {
	result := UNMATCH
	for _, pattern := range p.Patterns {
		if bytes.HasPrefix(header, []byte(pattern)) {
			return MATCH, p.Address
		}

		if len(header) < len(pattern) {
			result = TRYAGAIN
		}
	}

	return result, ""
}

func (p *PREFIX) Check() error {
	if len(p.Patterns) == 0 {
		return errors.New("at least one prefix pattern required")
	}

	for _, pattern := range p.Patterns {
		if len(pattern) == 0 {
			return errors.New("empty prefix pattern not allowed")
		}
	}

	return nil
}

type MatchTreeNode struct {
	ChildNodes map[byte]*MatchTreeNode
	Address    string
}

func (n *MatchTreeNode) Probe(header []byte) (ProbeResult, string) {
	nodes := n.ChildNodes
	for _, b := range header {
		if node, ok := nodes[b]; ok {
			if address := node.Address; address != "" {
				return MATCH, address
			}
			nodes = node.ChildNodes
		} else {
			return UNMATCH, ""
		}
	}

	return TRYAGAIN, ""
}

func createSubTree(data []byte) (root, leaf *MatchTreeNode) {
	root = new(MatchTreeNode)
	leaf = root
	for _, b := range data {
		nodes := make(map[byte]*MatchTreeNode)
		leaf.ChildNodes = nodes
		leaf = new(MatchTreeNode)
		nodes[b] = leaf
	}

	return
}

type MatchTree struct {
	Root *MatchTreeNode
}

func NewMatchTree() *MatchTree {
	t := new(MatchTree)
	t.Root = new(MatchTreeNode)
	return t
}

func (t *MatchTree) Add(p *PREFIX) {
	for _, patternStr := range p.Patterns {
		pattern := []byte(patternStr)
		node := t.Root
		for i, b := range pattern {
			nodes := node.ChildNodes
			if next_node, ok := nodes[b]; ok {
				node = next_node
				continue
			}

			if nodes == nil {
				nodes = make(map[byte]*MatchTreeNode)
				node.ChildNodes = nodes
			}

			root, leaf := createSubTree(pattern[i+1:])
			leaf.Address = p.Address
			nodes[b] = root

			break
		}
	}
}

func (t *MatchTree) Probe(header []byte) (ProbeResult, string) {
	return t.Root.Probe(header)
}
