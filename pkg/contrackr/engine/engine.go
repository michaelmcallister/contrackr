package engine

import (
	"fmt"
	"net"
	"time"

	log "github.com/golang/glog"
)

const (
	minimumPortScanned = 3
	// how long do entries get tracked for.
	trackerEntryTTL = 1 * time.Minute
	// how often do we evaluate our entries (ideally more often than entry TTL)
	evaluationInterval = 1 * time.Second
)

// CaptureCloser defines the contract for capturing packets from an interface.
type CaptureCloser interface {
	Capture() chan *Connection
	Close() error
}

// BlockCloser defines the contract for blocking IP addresses on the host.
type BlockCloser interface {
	Block(*net.IP) error
	Close() error
}

// Adder defines the contract for adding new connections to the tracker, and
// retrieving those that are considered port scanners.
type Adder interface {
	Add(*Connection)
	PortScanners() chan *TrackerEntry
}

// Engine contains the methods for running the connection tracker and blocker.
type Engine struct {
	capturer CaptureCloser
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
		firewall: fw,
		tracker:  newTracker(trackerEntryTTL, evaluationInterval, minimumPortScanned),
	}, nil
}

// Run will monitor and block source IPs that attempt to port scan on the device.
func (e *Engine) Run() {
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
	for pkt := range e.capturer.Capture() {
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
	return closeErr
}
