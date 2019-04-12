package main

import "strings"

type LoadBalanceDef struct {
	Name string
	Desc string
	Fn   string
}

type LoadBalancesDef struct {
	LoadBalances map[string]LoadBalanceDef
}

func (this *LoadBalancesDef) New(name string) string {
	return this.LoadBalances[name].Fn
}

func (this *LoadBalancesDef) Add(addLines []string, info *TCDInfo) error {
	desc, lines := comment(addLines)
	if lines[0] != "loadBalance {" {
		return NotMatch
	}
	lines = body(lines)
	for ; len(lines) > 0; lines = lines[1:] {
		desc, lines = comment(lines)
		nc := strings.SplitN(lines[0], " ", 2)
		ldDef := LoadBalanceDef{
			Name: nc[0], Fn: nc[1], Desc: desc,
		}
		this.LoadBalances[nc[0]] = ldDef
	}
	return nil
}

func NewLoadBalance() *LoadBalancesDef {
	return &LoadBalancesDef{
		LoadBalances: map[string]LoadBalanceDef{
			"round": {
				Name: "round",
				Fn:   "registry.NewRoundLoadBalance",
			},
			"none": {
				Name: "none",
				Fn:   "registry.NewNoneLoadBalance",
			},
		},
	}
}
