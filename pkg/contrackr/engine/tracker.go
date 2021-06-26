package engine

import (
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/golang/glog"
)

// TrackerEntry contains the Src and Dst IPs, as well as a map of Dst Ports
// and how many times that port was scanned.
type TrackerEntry struct {
	DstIP  *net.IP
	SrcIP  *net.IP
	Ports  map[int]int
	expiry time.Time
}

// Tracker contains the methods for tracking new connections, and retrieving
// entries that constitute port scanning.
type Tracker struct {
	portScanners       chan *TrackerEntry
	minimumPortScanned int
	maxAge             time.Duration
	// protects everything below.
	l sync.Mutex
	m map[string]*TrackerEntry
}

// newTracker takes the maximum age each entry should be tracked for, and
// the minimum ports scanned before a src IP is considered a "port scanner"
// and returns an instance of Tracker.
func newTracker(maxAge, evaluationInterval time.Duration, minimumPortScanned int) (t *Tracker) {
	t = &Tracker{
		portScanners:       make(chan *TrackerEntry),
		minimumPortScanned: minimumPortScanned,
		maxAge:             maxAge,
		m:                  make(map[string]*TrackerEntry),
	}
	go func() {
		for now := range time.Tick(evaluationInterval) {
			t.l.Lock()
			for k, v := range t.m {
				if now.After(v.expiry) {
					log.Infof("removing %q because entry is expired", k)
					delete(t.m, k)
				}
			}
			t.l.Unlock()
		}
	}()
	return
}

// Add adds the connection v into the tracker. Connections are tracked in a
// Src IP + Dst IP tuple.
func (t *Tracker) Add(v *Connection) {
	t.l.Lock()
	// TODO(michaelmcallister): clarify if port scanning is *any* dst IP on
	// the interface, or a specific one. With the current implementation a
	// port scanner could scan up to 2 ports * N IP addresses on the interface.
	// If it's any Dst IP address, change the key to simply be the Src IP.
	key := fmt.Sprintf("[%s]>[%s]", v.Src.IP, v.Dst.IP)
	_, ok := t.m[key]
	if !ok {
		t.m[key] = &TrackerEntry{
			DstIP:  &v.Dst.IP,
			SrcIP:  &v.Src.IP,
			Ports:  make(map[int]int),
			expiry: time.Now().Add(t.maxAge),
		}
	}
	t.m[key].Ports[v.Dst.Port]++
	if len(t.m[key].Ports) >= t.minimumPortScanned {
		t.portScanners <- t.m[key]
	}
	t.l.Unlock()
}

// PortScanners returns a channel that callers can retrieve Entries that
// scan multiple ports.
func (t *Tracker) PortScanners() chan *TrackerEntry {
	return t.portScanners
}

func (t *Tracker) Close() {
	close(t.portScanners)
}
