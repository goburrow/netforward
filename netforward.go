package netforward

import (
	"crypto/tls"
	"errors"
	"log"
	"net"
	"time"
)

var (
	errMustListen = errors.New("must listen first")
)

// Dialer dials to a remote address.
type Dialer interface {
	Dial() (net.Conn, error)
}

// Forwarder forwards packets sending to the local network to the remote network.
type Forwarder interface {
	Forward(remote Dialer) error
}

// Endpoint is a network endpoint which can dial to or listen from.
type Endpoint struct {
	Network string
	Address string

	TLS *tls.Config

	// Remote
	Timeout time.Duration
}

// Dial dials to the given network endpoint.
func (e *Endpoint) Dial() (net.Conn, error) {
	dialer := net.Dialer{Timeout: e.Timeout}
	if e.TLS != nil {
		return tls.DialWithDialer(&dialer, e.Network, e.Address, e.TLS)
	}
	return dialer.Dial(e.Network, e.Address)
}

// Listen returns a listener of a local address.
// Network must be a stream type.
func (e *Endpoint) Listen() (net.Listener, error) {
	if e.TLS != nil {
		return tls.Listen(e.Network, e.Address, e.TLS)
	}
	return net.Listen(e.Network, e.Address)
}

// ListenPacket returns a listener of a local address.
// Network must be a packet oriented type.
func (e *Endpoint) ListenPacket() (net.PacketConn, error) {
	if e.TLS != nil {
		// not supported
		log.Println("DTLS is not supported yet")
	}
	return net.ListenPacket(e.Network, e.Address)
}

type NetForwarder struct {
	Local Endpoint

	// Stream
	ln net.Listener
	// Packet
	packetConn net.PacketConn
}

func (f *NetForwarder) Listen() error {
	var err error
	if isPacketNetwork(f.Local.Network) {
		f.packetConn, err = f.Local.ListenPacket()
	} else {
		f.ln, err = f.Local.Listen()
	}
	return err
}

func (f *NetForwarder) Forward(remote Dialer) error {
	if f.ln != nil {
		return Forward(remote, f.ln)
	}
	if f.packetConn != nil {
		return ForwardPacket(remote, f.packetConn)
	}
	return errMustListen
}

func (f *NetForwarder) Close() error {
	if f.ln != nil {
		return f.ln.Close()
	}
	if f.packetConn != nil {
		return f.packetConn.Close()
	}
	return nil
}

func isPacketNetwork(network string) bool {
	switch network {
	case "udp", "udp4", "udp6", "ip", "ip4", "ip6", "unixgram":
		return true
	default:
		return false
	}
}
