package engine

import (
	"net"
	"sync"
	"testing"
)

// fakeBlocker implements the BlockCloser interface.
type fakeBlocker struct {
	blockCalled chan bool
}

// Block will track whether it has been called, returning a nil error always.
func (fb *fakeBlocker) Block(_ *net.IP) error {
	fb.blockCalled <- true
	return nil
}

// Close always returns nil.
func (fb *fakeBlocker) Close() error {
	return nil
}

// fakeTracker implements the Adder interface.
type fakeTracker struct {
	tracking int
	tc       chan *TrackerEntry
}

// Add is a no-op.
func (ft *fakeTracker) Add(_ *Connection) {
	ft.tracking++
}

// PortScanners returns the channel that callers can recieve as soon as a
// "port scanner" is identified.
func (ft *fakeTracker) PortScanners() chan *TrackerEntry {
	return ft.tc
}

// Count returns how many connections are being tracked. This instance tracks
// them forever, until manually cleared.
func (ft *fakeTracker) Connections() int {
	return ft.tracking
}

// Close closes the underlying channel.
func (ft *fakeTracker) Close() {
	close(ft.tc)
}

// fakeCapturer implements the CaptureCloser interface.
type fakeCapturer struct {
	captureChan chan *Connection
}

// Capture returns the channel that contains TCP SYN connections as soon as
// they are parsed.
func (fc *fakeCapturer) Capture() chan *Connection {
	return fc.captureChan
}

// Close closes the capture channel.
func (fc *fakeCapturer) Close() error {
	close(fc.captureChan)
	return nil
}

func TestEngineBlocksPortScans(t *testing.T) {
	// Entries added to this channel are considered "port scanners"
	portscanners := make(chan *TrackerEntry)
	fakeBlocker := &fakeBlocker{blockCalled: make(chan bool)}
	fakeEngine := &Engine{
		capturer: &fakeCapturer{captureChan: make(chan *Connection)},
		firewall: fakeBlocker,
		tracker:  &fakeTracker{tc: portscanners},
	}
	var wg sync.WaitGroup

	// Run the engine.
	wg.Add(1)
	go func() {
		defer wg.Done()
		fakeEngine.Run()
	}()

	// Send an entry to the portscanner channel.
	srcIP, dstIP := net.ParseIP("192.168.86.158"), net.ParseIP("192.168.86.191")
	portscanners <- &TrackerEntry{
		DstIP: &srcIP,
		SrcIP: &dstIP,
		Ports: map[int]int{1992: 1, 7: 1, 9: 1},
	}

	ok := <-fakeBlocker.blockCalled
	if ok {
		fakeEngine.Close()
	}
	wg.Wait()

	// Test that the engine correctly sent a block request for this entry.
	if !ok {
		t.Errorf("Expected Block() method to be called but wasn't")
	}
}
