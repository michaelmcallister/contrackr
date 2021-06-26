package engine

import (
	"fmt"
	"io"
	"net"
	"os"

	log "github.com/golang/glog"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// captureBytes is the maximum bytes per packet to capture.
// It is intended to capture the header in all cases, and not much more else.
const captureBytes = 80

// bpfFilter is the BPF filter that is used to capture TCP SYN packets.
// I did not know until writing this, but wireshark (libpcap?) does not support
// the syntactic sugar that IPv4 does, so we must inspect the packet data.
// We go 13 bytes into the TCP header + 40 bytes for the leading IPv6 header
// and we check to see if the SYN flag is set.
// TCP Header format is as such:
// -----------------------------------------------------------------
//     0                   1                   2                   3
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |          Source Port          |       Destination Port        |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                        Sequence Number                        |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                    Acknowledgment Number                      |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |  Data |           |U|A|P|R|S|F|                               |
// | Offset| Reserved  |R|C|S|S|Y|I|            Window             |
// |       |           |G|K|H|T|N|N|                               |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |           Checksum            |         Urgent Pointer        |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                    Options                    |    Padding    |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                             data                              |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
const bpfFilter = "tcp[tcpflags] == tcp-syn or ip6[13+40]&0x2!=0"

type Connection struct {
	Src *net.TCPAddr
	Dst *net.TCPAddr
}

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
	h   *pcap.Handle
	out chan *Connection
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
	return &PacketCapturer{h: h}, nil
}

// newCapturerOffline accepts a instance of os.File and attempts to read the
// packet data, returning an instance of PacketCapturer if successful, else
// error.
func newCapturerOffline(file *os.File) (*PacketCapturer, error) {
	h, err := pcap.OpenOfflineFile(file)
	if err != nil {
		return nil, err
	}
	if err := h.SetBPFFilter(bpfFilter); err != nil {
		return nil, err
	}
	return &PacketCapturer{h: h}, nil
}

// Parse will read from the supplied packet source and return a channel that
// will be populated with pointers to Connection that contain the Src and Dst
// IP:Port tuples of the inbound TCP packet. Packets that cannot be decoded,
// have no TCP header, or do not have SYN flag (or have the SYN + ACK flag set)
// will be silently dropped.
func (pc *PacketCapturer) Capture() chan *Connection {
	pc.out = make(chan *Connection)
	source := gopacket.NewPacketSource(pc.h, pc.h.LinkType())
	go func() {
		for {
			packet, err := source.NextPacket()
			if err == io.EOF {
				close(pc.out)
				break
			} else if err != nil {
				log.Warningf("error reading packet: %v", err)
				log.Warning("skipping...")
				continue
			}
			parsedTCP := &Connection{
				Src: &net.TCPAddr{},
				Dst: &net.TCPAddr{},
			}

			if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
				tcp, _ := tcpLayer.(*layers.TCP)
				// This shouldn't happen as the capturer isn't configured to
				// capture anything but SYN packets. The BPF Filter is applied
				// even on pcap files that may have been generated with
				// different filters.
				if !tcp.SYN || tcp.ACK {
					log.Warning("packet is not TCP with SYN flag, skipping...")
					continue
				}
				parsedTCP.Src.Port = int(tcp.SrcPort)
				parsedTCP.Dst.Port = int(tcp.DstPort)
			}
			if ipv6Layer := packet.Layer(layers.LayerTypeIPv6); ipv6Layer != nil {
				ip6, _ := ipv6Layer.(*layers.IPv6)
				parsedTCP.Src.IP = ip6.SrcIP
				parsedTCP.Dst.IP = ip6.DstIP
			}
			if ipv4Layer := packet.Layer(layers.LayerTypeIPv4); ipv4Layer != nil {
				ip4, _ := ipv4Layer.(*layers.IPv4)
				parsedTCP.Src.IP = ip4.SrcIP
				parsedTCP.Dst.IP = ip4.DstIP
			}

			if parsedTCP.Src.IP != nil && parsedTCP.Dst.Port != 0 {
				pc.out <- parsedTCP
			}

		}
	}()
	return pc.out
}

// Close closes the underlying pcap handle. It will always return a nil error.
// Attempting to read after closing is discouraged.
func (pc *PacketCapturer) Close() error {
	if pc != nil {
		pc.h.Close()
	}
	return nil
}
