package engine

import (
	"errors"
	"io"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// readBuffer is the maximum bytes that will be read into the buffer.
const readBuffer = 80

type Connection struct {
	Src *net.TCPAddr
	Dst *net.TCPAddr
}

type PacketParser struct {
	pc io.ReadCloser
}

func newPacketParser(pc io.ReadCloser) *PacketParser {
	return &PacketParser{pc: pc}
}

// Parse will read from the supplied packet capturer and return a populated
// instance of net.TCPAddr containing the dst port (local) and src IP (remote)
// of the inbound TCP packet. Packets that cannot be decoded,
// have no TCP header, or do not have SYN flag (or have the SYN + ACK flag set)
// will return an error.
func (pr *PacketParser) Parse() (*Connection, error) {
	buf := make([]byte, readBuffer)
	if _, err := pr.pc.Read(buf); err != nil {
		return nil, err
	}

	var eth layers.Ethernet
	var ip4 layers.IPv4
	var ip6 layers.IPv6
	var tcp layers.TCP
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &ip6, &tcp)
	decoded := []gopacket.LayerType{}
	if err := parser.DecodeLayers(buf, &decoded); err != nil {
		return nil, err
	}
	parsedTCP := &Connection{
		Src: &net.TCPAddr{},
		Dst: &net.TCPAddr{},
	}
	for _, layerType := range decoded {
		switch layerType {
		case layers.LayerTypeTCP:
			// this shouldn't happen as the capturer isn't configured to capture
			// anything but SYN packets, but contrackr is written to be modular
			// and it's possible a different implementation doesn't filter them.
			if !tcp.SYN || tcp.ACK {
				return nil, errors.New("packet is not TCP with SYN flag")
			}
			parsedTCP.Src.Port = int(tcp.SrcPort)
			parsedTCP.Dst.Port = int(tcp.DstPort)
		case layers.LayerTypeIPv6:
			parsedTCP.Src.IP = ip6.SrcIP
			parsedTCP.Dst.IP = ip6.DstIP
		case layers.LayerTypeIPv4:
			parsedTCP.Src.IP = ip4.SrcIP
			parsedTCP.Dst.IP = ip4.DstIP
		}
	}

	if parsedTCP == nil {
		return nil, errors.New("unable to parse packet")
	}
	return parsedTCP, nil
}
