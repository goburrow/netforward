package netforward

import (
	"io"
	"log"
	"net"
	"sync"
)

type packetConn struct {
	conn net.PacketConn
	addr net.Addr
}

func (p *packetConn) Write(b []byte) (int, error) {
	return p.conn.WriteTo(b, p.addr)
}

type syncConns struct {
	mu    sync.RWMutex
	conns map[string]net.Conn
}

func (f *syncConns) get(addr net.Addr) (net.Conn, bool) {
	var key string
	if addr != nil {
		key = addr.String()
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	conn, ok := f.conns[key]
	return conn, ok
}

func (f *syncConns) set(addr net.Addr, conn net.Conn) {
	var key string
	if addr != nil {
		key = addr.String()
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.conns[key] = conn
}

func (f *syncConns) remove(addr net.Addr) {
	var key string
	if addr != nil {
		key = addr.String()
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.conns, key)
}

func ForwardPacket(remote Dialer, local net.PacketConn) error {
	conns := syncConns{
		conns: make(map[string]net.Conn),
	}
	buf := getBuffer()
	defer releaseBuffer(buf)
	for {
		nr, addr, err := local.ReadFrom(buf)
		if err != nil {
			return err
		}
		remoteConn, ok := conns.get(addr)
		if !ok {
			remoteConn, err = remote.Dial()
			if err != nil {
				log.Printf("%s: dial failed: %v", addr, err)
				continue
			}
			conns.set(addr, remoteConn)
			// unix datagram does not have a network addr
			if addr != nil {
				go forwardPacket(remoteConn, local, addr, &conns)
			}
		}
		nw, err := remoteConn.Write(buf[:nr])
		if err != nil {
			log.Printf("%s: write failed: %v", addr, err)
			continue
		}
		if nr != nw {
			log.Printf("%s: %v", addr, io.ErrShortWrite)
		}
	}
}

func forwardPacket(remote net.Conn, local net.PacketConn, addr net.Addr, conns *syncConns) {
	defer remote.Close()
	defer conns.remove(addr)

	buf := getBuffer()
	defer releaseBuffer(buf)

	_, err := io.CopyBuffer(&packetConn{local, addr}, remote, buf)
	if err != nil {
		log.Printf("%s: %v", addr, err)
	}
}
