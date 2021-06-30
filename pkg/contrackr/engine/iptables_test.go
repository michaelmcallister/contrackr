package engine

import (
	"fmt"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// fakeIptables keeps a list of methods and the order they were executed in.
type fakeIptables struct {
	chainSetup       bool
	commandsExecuted []string
}

func (fi *fakeIptables) ChainExists(table, chain string) (bool, error) {
	fi.commandsExecuted = append(fi.commandsExecuted, fmt.Sprintf("ChainExists(%s, %s)", table, chain))
	return fi.chainSetup, nil
}

func (fi *fakeIptables) NewChain(table, chain string) error {
	fi.chainSetup = true
	fi.commandsExecuted = append(fi.commandsExecuted, fmt.Sprintf("NewChain(%s, %s)", table, chain))
	return nil
}

func (fi *fakeIptables) Insert(table, chain string, pos int, rulespec ...string) error {
	m := fmt.Sprintf("Insert(%s, %s, %d, %v)", table, chain, pos, rulespec)
	fi.commandsExecuted = append(fi.commandsExecuted, m)
	return nil
}

func (fi *fakeIptables) AppendUnique(table, chain string, rulespec ...string) error {
	m := fmt.Sprintf("AppendUnique(%s, %s, %v)", table, chain, rulespec)
	fi.commandsExecuted = append(fi.commandsExecuted, m)
	return nil
}

func (fi *fakeIptables) DeleteIfExists(table, chain string, rulespec ...string) error {
	m := fmt.Sprintf("DeleteIfExists(%s, %s, %v)", table, chain, rulespec)
	fi.commandsExecuted = append(fi.commandsExecuted, m)
	return nil
}

func (fi *fakeIptables) ClearAndDeleteChain(table, chain string) error {
	m := fmt.Sprintf("ClearAndDeleteChain(%s, %s)", table, chain)
	fi.commandsExecuted = append(fi.commandsExecuted, m)
	return nil
}

func TestBlockIpv4(t *testing.T) {
	v4 := &fakeIptables{}
	v6 := &fakeIptables{}
	b := &Blocker{ip4tables: v4, ip6tables: v6}
	ip := net.ParseIP("127.0.0.1")
	b.init()
	b.Block(&ip)
	b.Close()

	wantv4 := []string{
		"ChainExists(filter, contrackr)",
		"ChainExists(filter, contrackr)",
		"NewChain(filter, contrackr)",
		"Insert(filter, INPUT, 1, [-m state --state NEW -j contrackr])",
		"AppendUnique(filter, contrackr, [-s 127.0.0.1 -j DROP])",
		"ChainExists(filter, contrackr)",
		"DeleteIfExists(filter, INPUT, [-m state --state NEW -j contrackr])",
		"ClearAndDeleteChain(filter, contrackr)",
	}

	wantv6 := []string{
		"ChainExists(filter, contrackr)",
		"ChainExists(filter, contrackr)",
		"NewChain(filter, contrackr)",
		"Insert(filter, INPUT, 1, [-m state --state NEW -j contrackr])",
		"ChainExists(filter, contrackr)",
		"DeleteIfExists(filter, INPUT, [-m state --state NEW -j contrackr])",
		"ClearAndDeleteChain(filter, contrackr)",
	}

	if diff := cmp.Diff(wantv4, v4.commandsExecuted); diff != "" {
		t.Errorf("Block() mismatch (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(wantv6, v6.commandsExecuted); diff != "" {
		t.Errorf("Block() mismatch (-want +got):\n%s", diff)
	}
}

func TestBlockIpv6(t *testing.T) {
	v4 := &fakeIptables{}
	v6 := &fakeIptables{}
	b := &Blocker{ip4tables: v4, ip6tables: v6}
	ip := net.ParseIP("2001:4860:4860::8888")
	b.init()
	b.Block(&ip)
	b.Close()

	wantv4 := []string{
		"ChainExists(filter, contrackr)",
		"ChainExists(filter, contrackr)",
		"NewChain(filter, contrackr)",
		"Insert(filter, INPUT, 1, [-m state --state NEW -j contrackr])",
		"ChainExists(filter, contrackr)",
		"DeleteIfExists(filter, INPUT, [-m state --state NEW -j contrackr])",
		"ClearAndDeleteChain(filter, contrackr)",
	}

	wantv6 := []string{
		"ChainExists(filter, contrackr)",
		"ChainExists(filter, contrackr)",
		"NewChain(filter, contrackr)",
		"Insert(filter, INPUT, 1, [-m state --state NEW -j contrackr])",
		"AppendUnique(filter, contrackr, [-s 2001:4860:4860::8888 -j DROP])",
		"ChainExists(filter, contrackr)",
		"DeleteIfExists(filter, INPUT, [-m state --state NEW -j contrackr])",
		"ClearAndDeleteChain(filter, contrackr)",
	}

	if diff := cmp.Diff(wantv4, v4.commandsExecuted); diff != "" {
		t.Errorf("Block() mismatch (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(wantv6, v6.commandsExecuted); diff != "" {
		t.Errorf("Block() mismatch (-want +got):\n%s", diff)
	}
}
