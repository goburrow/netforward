package netforward

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"time"
)

// Listener gives a listener on a local address.
type Listener interface {
	Accept() (net.Conn, error)
}

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

func isPacketNetwork(network string) bool {
	switch network {
	case "udp", "udp4", "udp6", "ip", "ip4", "ip6", "unixgram":
		return true
	default:
		return false
	}
}

func Forward(remote Dialer, local Listener) error {
	for {
		conn, err := local.Accept()
		if err != nil {
			return err
		}
		go forward(remote, conn)
	}
}

func forward(remote Dialer, conn io.ReadWriteCloser) {
	defer conn.Close()

	remoteConn, err := remote.Dial()
	if err != nil {
		log.Printf("dial failed: %v", err)
		return
	}
	defer remoteConn.Close()

	// remote -> local
	go func() {
		_, err := io.Copy(remoteConn, conn)
		if err != nil {
			log.Printf("forward failed: %v", err)
		}
	}()

	// local -> remote
	_, err = io.Copy(conn, remoteConn)
	if err != nil {
		log.Printf("forward failed: %v", err)
	}
}
