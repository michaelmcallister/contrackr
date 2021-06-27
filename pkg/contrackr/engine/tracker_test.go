package engine

import (
	"net"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestAdding(t *testing.T) {
	srcIP, dstIP := net.ParseIP("192.168.86.158"), net.ParseIP("192.168.86.191")

	// Let's move quickly in the tests.
	const evaluationTime = time.Millisecond
	testCases := []struct {
		desc               string
		minimumPortScanned int
		wait               time.Duration
		maxAge             time.Duration
		in                 []*Connection
		want               []*TrackerEntry
		wantConnections    int
	}{
		{
			desc:               "test 3 connections considered a port scanner",
			minimumPortScanned: 3,
			maxAge:             time.Minute,
			wantConnections:    3,
			in: []*Connection{
				{
					Src: &net.TCPAddr{
						IP:   srcIP,
						Port: 41832,
					},
					Dst: &net.TCPAddr{
						IP:   dstIP,
						Port: 1992,
					},
				},
				{
					Src: &net.TCPAddr{
						IP:   srcIP,
						Port: 41832,
					},
					Dst: &net.TCPAddr{
						IP:   dstIP,
						Port: 7,
					},
				},
				{
					Src: &net.TCPAddr{
						IP:   srcIP,
						Port: 41832,
					},
					Dst: &net.TCPAddr{
						IP:   dstIP,
						Port: 9,
					},
				},
			},
			want: []*TrackerEntry{
				{
					DstIP: &dstIP,
					SrcIP: &srcIP,
					Ports: map[int]int{1992: 1, 7: 1, 9: 1},
				},
			},
		},
		{
			desc:               "test entries removed after expiry",
			minimumPortScanned: 3,
			wait:               time.Second,
			maxAge:             time.Microsecond,
			in: []*Connection{
				{
					Src: &net.TCPAddr{
						IP:   srcIP,
						Port: 41832,
					},
					Dst: &net.TCPAddr{
						IP:   dstIP,
						Port: 1992,
					},
				},
				{
					Src: &net.TCPAddr{
						IP:   srcIP,
						Port: 41832,
					},
					Dst: &net.TCPAddr{
						IP:   dstIP,
						Port: 7,
					},
				},
				{
					Src: &net.TCPAddr{
						IP:   srcIP,
						Port: 41832,
					},
					Dst: &net.TCPAddr{
						IP:   dstIP,
						Port: 9,
					},
				},
			},
			want: nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			tkr := newTracker(tC.maxAge, evaluationTime, tC.minimumPortScanned)
			var got []*TrackerEntry
			go func() {
				defer tkr.Close()
				for _, c := range tC.in {
					tkr.Add(c)
					time.Sleep(tC.wait)
				}
			}()

			for i := range tkr.PortScanners() {
				if i != nil {
					got = append(got, i)
				}
			}

			if diff := cmp.Diff(tC.want, got, cmpopts.IgnoreUnexported(TrackerEntry{})); diff != "" {
				t.Errorf("PortScanners() mismatch (-want +got):\n%s", diff)
			}

			c := tkr.Connections()
			if c != tC.wantConnections {
				t.Errorf("tkr.Connections() = %d, want connections = %d", c, tC.wantConnections)
			}
		})
	}
}
