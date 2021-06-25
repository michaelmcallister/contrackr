package engine

import (
	"fmt"

	log "github.com/golang/glog"
	"github.com/google/gopacket/pcap"
)

// captureBytes is the maximum bytes per packet to capture.
// It is intended to capture the header in all cases, and not much more else.
const captureBytes = 80

// bpfFilter is the BPF filter that is used to capture TCP SYN packets.
const bpfFilter = "tcp[tcpflags] == tcp-syn"

// interfaceExists returns true when devicename is found as an interface on the
// running system, else false.
func interfaceExists(devicename string) bool {
	ifs, err := pcap.FindAllDevs()
	if err != nil {
		log.Warning("Unable to find devices: ", err)
		return false
	}
	for _, iface := range ifs {
		//TODO(michaelmcallister): consider case insensitive device names?
		if iface.Name == devicename {
			return true
		}
	}
	return false
}

// PacketCapturer implements the io.ReadCloser interface.
type PacketCapturer struct {
	h *pcap.Handle
}

// newCapturer accepts a devicename that must exist as a network interface, and
// then returns an instance of PacketCapturer, else an error.
func newCapturer(devicename string) (*PacketCapturer, error) {
	if !interfaceExists(devicename) {
		return nil, fmt.Errorf("interface %q not found", devicename)
	}
	ih, err := pcap.NewInactiveHandle(devicename)
	defer ih.CleanUp()
	// TODO(michaelmcallister): consider refactoring to avoid the error checking
	// repetition.
	if err != nil {
		return nil, err
	}
	if err := ih.SetImmediateMode(true); err != nil {
		return nil, err
	}
	if err := ih.SetPromisc(false); err != nil {
		return nil, err
	}
	if err := ih.SetSnapLen(captureBytes); err != nil {
		return nil, err
	}
	h, err := ih.Activate()
	if err != nil {
		return nil, err
	}
	if err := h.SetDirection(pcap.DirectionIn); err != nil {
		return nil, err
	}
	if err := h.SetBPFFilter(bpfFilter); err != nil {
		return nil, err
	}
	return &PacketCapturer{h: h}, nil
}

// Read reads up to len(p) bytes of captured packets into p.
func (pc *PacketCapturer) Read(p []byte) (int, error) {
	b, _, err := pc.h.ZeroCopyReadPacketData()
	copy(p, b)
	return len(p), err
}

// Close closes the underlying pcap handle. It will always return a nil error.
// Attempting to read after closing is discouraged.
func (pc *PacketCapturer) Close() error {
	if pc != nil {
		pc.h.Close()
	}
	return nil
}
