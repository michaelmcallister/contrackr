package engine

import (
	"fmt"
	"io"
	"net"
	"time"

	log "github.com/golang/glog"
)

const (
	minimumPortScanned = 3
	trackerEntryTTL    = 1 * time.Minute
)

// Parser defines the contract for parsing arbitary bytes and extracting
// connections. It is seperated from the packet capturer for testability.
type Parser interface {
	Parse() (*Connection, error)
}

// BlockCloser defines the contract for blocking IP addresses on the host.
type BlockCloser interface {
	Block(*net.IP) error
	Close() error
}

// Add defines the contract for adding new connections to the tracker, and
// retrieving those that are considered port scanners.
type Adder interface {
	Add(*Connection)
	PortScanners() chan *TrackerEntry
}

// Engine contains the methods for running the connection tracker and blocker.
type Engine struct {
	c        chan *Connection
	capturer io.ReadCloser
	parser   Parser
	firewall BlockCloser
	tracker  Adder
}

// New accepts a deviceName (eg. eth0) and returns an instance of Engine, else
// error.
func New(deviceName string) (*Engine, error) {
	cap, err := newCapturer(deviceName)
	if err != nil {
		return nil, err
	}
	fw, err := newBlocker()
	if err != nil {
		return nil, err
	}
	return &Engine{
		capturer: cap,
		parser:   newPacketParser(cap),
		firewall: fw,
		tracker:  newTracker(trackerEntryTTL, minimumPortScanned),
	}, nil
}

// Run will monitor and block source IPs that attempt to port scan on the device.
func (e *Engine) Run() {
	e.c = make(chan *Connection)
	go func() {
		for {
			pkt, err := e.parser.Parse()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Warningf("Error parsing packet: %v", err)
				log.Warning("Continuing...")
			}
			e.c <- pkt
		}
	}()
	go func() {
		for v := range e.tracker.PortScanners() {
			var ports []int
			for k := range v.Ports {
				ports = append(ports, k)
			}
			log.Infof("Port scan detected: %s -> %s on ports %v", v.SrcIP, v.DstIP, ports)
			e.firewall.Block(v.SrcIP)
		}
	}()
	for pkt := range e.c {
		e.tracker.Add(pkt)
	}
}

// Close cleans up any dependencies.
func (e *Engine) Close() error {
	var closeErr error
	if err := e.firewall.Close(); err != nil {
		closeErr = fmt.Errorf("firewall %v", err)
	}
	if err := e.capturer.Close(); err != nil {
		closeErr = fmt.Errorf("capturer %v: %w", err, closeErr)
	}
	close(e.c)
	return closeErr
}
