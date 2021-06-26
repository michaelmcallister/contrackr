package engine

import (
	"fmt"
	"net"

	"github.com/coreos/go-iptables/iptables"
)

const (
	inputChain = "INPUT"
	// we use our own chain to keep things seperated. It will be cleared and
	// removed on teardown.
	contrackrChain = "contrackr"
	// This is the default table, it contains the built-in chains INPUT
	// (for packet destined to local sockets) as per iptables(8).
	defaultTable = "filter"
	// IPTables jump target, use REJECT if you want the source to know that
	// they are blocked.
	blockAction = "DROP"
)

// BLocker contains the methods for Blocking IP addresses.
type Blocker struct {
	iptables, ip6tables *iptables.IPTables
}

var (
	// jumpRuleSpec dictates when and how we should pivot from the filter table
	// to our own.
	jumpRuleSpec = []string{"-m", "state", "--state", "NEW", "-j", contrackrChain}
)

// newBlocker returns and instance of Blocker.
func newBlocker() (*Blocker, error) {
	v4, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		return nil, err
	}
	v6, err := iptables.NewWithProtocol(iptables.ProtocolIPv6)
	if err != nil {
		return nil, err
	}
	for _, i := range []*iptables.IPTables{v4, v6} {
		// Create our own chain if it doesn't exist.
		ok, err := i.ChainExists(defaultTable, contrackrChain)
		if err != nil {
			return nil, err
		}
		if !ok {
			if err := i.NewChain(defaultTable, contrackrChain); err != nil {
				return nil, err
			}
		}
		// Add an entry to INPUT to jump to our chain.
		if err := i.Insert(defaultTable, inputChain, 1, jumpRuleSpec...); err != nil {
			return nil, err
		}
	}
	return &Blocker{iptables: v4, ip6tables: v6}, nil
}

// Block will take the IP Address v and add an entry to the host firewall.
func (b *Blocker) Block(v *net.IP) error {
	rule := []string{"-s", v.String(), "-j", blockAction}
	isV6 := v.To4() == nil
	if isV6 {
		return b.ip6tables.AppendUnique(defaultTable, contrackrChain, rule...)
	}
	return b.iptables.AppendUnique(defaultTable, contrackrChain, rule...)
}

// Close will cleanup the firewall rules that were created during instantiation.
func (b *Blocker) Close() error {
	var closeErr error
	for _, i := range []*iptables.IPTables{b.iptables, b.ip6tables} {
		if err := i.ClearAndDeleteChain(defaultTable, contrackrChain); err != nil {
			closeErr = fmt.Errorf("deleting chain: %v", err)
		}
		if err := i.DeleteIfExists(defaultTable, inputChain, jumpRuleSpec...); err != nil {
			closeErr = fmt.Errorf("deleting jump rule: %v: %w", err, closeErr)
		}
	}
	return closeErr
}
