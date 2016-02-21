package netforward

import (
	"io"
	"net"
	"os"
	"sync"
	"testing"
)

func echoStream(t *testing.T, ln net.Listener) {
	for {
		c, err := ln.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		go func() {
			defer c.Close()
			// echo received data
			var buf [16]byte
			_, err := io.CopyBuffer(c, c, buf[:])
			if err != nil {
				t.Error(err)
			}
		}()
	}
}

func verify(t *testing.T, network, addr string) {
	c, err := net.Dial(network, addr)
	if err != nil {
		t.Error(err)
		return
	}
	defer c.Close()

	verifyRW(t, c)
	verifyRW(t, c)
	verifyRW(t, c)
}

func verifyRW(t *testing.T, rw io.ReadWriter) {
	n, err := rw.Write([]byte("netforward"))
	if err != nil {
		t.Error(err)
		return
	}
	if n != 10 {
		t.Errorf("unexpected bytes written: %d", n)
		return
	}
	var buf [16]byte
	n, err = rw.Read(buf[:])
	if err != nil {
		t.Error(err)
		return
	}
	if n != 10 || "netforward" != string(buf[:n]) {
		t.Errorf("unexpected receive: %s (%d bytes)", buf[:n], n)
		return
	}
}

func testStream(t *testing.T, rnet, raddr, lnet, laddr string) {
	var wg sync.WaitGroup

	// Echo server
	sln, err := net.Listen(rnet, raddr)
	if err != nil {
		t.Fatal(err)
	}
	defer sln.Close()

	// Proxy server
	f := NetForwarder{
		Local: Endpoint{
			Network: lnet,
			Address: laddr,
		},
	}
	err = f.Listen()
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// Start the echo server
	wg.Add(1)
	go func() {
		defer wg.Done()
		echoStream(t, sln)
	}()

	// Forwarding
	wg.Add(1)
	go func() {
		defer wg.Done()
		e2 := Endpoint{
			Network: rnet,
			Address: sln.Addr().String(),
		}
		err = f.Forward(&e2)
		if err != nil {
			t.Log(err)
		}
	}()

	verify(t, lnet, laddr)

	sln.Close()
	f.Close()
	wg.Wait()
}

func TestTCPToTCP(t *testing.T) {
	testStream(t, "tcp", "127.0.0.1:0", "tcp", "127.0.0.1:17890")
}

func TestUDPToTCP(t *testing.T) {
	testStream(t, "tcp", "127.0.0.1:0", "udp", "127.0.0.1:17890")
}

func TestUnixToTCP(t *testing.T) {
	defer os.Remove("/tmp/nftest.sock")
	testStream(t, "tcp", "127.0.0.1:0", "unix", "/tmp/nftest.sock")
}

func TestTCPToUnix(t *testing.T) {
	defer os.Remove("/tmp/nftest.sock")
	testStream(t, "unix", "/tmp/nftest.sock", "tcp", "127.0.0.1:17890")
}

func TestUnixToUnix(t *testing.T) {
	defer os.Remove("/tmp/nftest.sock")
	defer os.Remove("/tmp/nftest2.sock")
	testStream(t, "unix", "/tmp/nftest2.sock", "unix", "/tmp/nftest.sock")
}

func echoPacket(t *testing.T, conn net.PacketConn) {
	var buf [16]byte
	for {
		nr, addr, err := conn.ReadFrom(buf[:])
		if err != nil {
			t.Log(err)
			return
		}
		nw, err := conn.WriteTo(buf[:nr], addr)
		if err != nil {
			t.Error(nw)
			continue
		}
		if nr != nw {
			t.Errorf("short write: %d. Expected: %d", nw, nr)
			continue
		}
	}
}

func testPacket(t *testing.T, rnet, raddr, lnet, laddr string) {
	var wg sync.WaitGroup

	// Echo server
	sln, err := net.ListenPacket(rnet, raddr)
	if err != nil {
		t.Fatal(err)
	}
	defer sln.Close()

	// Proxy server
	f := NetForwarder{
		Local: Endpoint{
			Network: lnet,
			Address: laddr,
		},
	}
	err = f.Listen()
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// Start the echo server
	wg.Add(1)
	go func() {
		defer wg.Done()
		echoPacket(t, sln)
	}()

	// Forwarding
	wg.Add(1)
	go func() {
		defer wg.Done()
		e2 := Endpoint{
			Network: rnet,
			Address: sln.LocalAddr().String(),
		}
		err = f.Forward(&e2)
		if err != nil {
			t.Log(err)
		}
	}()

	verify(t, lnet, laddr)

	sln.Close()
	f.Close()
	wg.Wait()
}

func TestUDPToUDP(t *testing.T) {
	testPacket(t, "udp", "127.0.0.1:0", "udp", "127.0.0.1:17890")
}

func TestTCPToUDP(t *testing.T) {
	testPacket(t, "udp", "127.0.0.1:0", "tcp", "127.0.0.1:17890")
}

func TestUnixgramToUDP(t *testing.T) {
	t.Skip("can not have multiple listener on a unix datagram socket")
	defer os.Remove("/tmp/nftest.sock")
	testPacket(t, "udp", "127.0.0.1:0", "unixgram", "/tmp/nftest.sock")
}

func TestUDPToUnixgram(t *testing.T) {
	t.Skip("can not have multiple listener on a unix datagram socket")
	defer os.Remove("/tmp/nftest.sock")
	testPacket(t, "unixgram", "/tmp/nftest.sock", "udp", "127.0.0.1:17890")
}

func TestUnixgramToUnixgram(t *testing.T) {
	t.Skip("can not have multiple listener on a unix datagram socket")
	defer os.Remove("/tmp/nftest.sock")
	defer os.Remove("/tmp/nftest2.sock")
	testPacket(t, "unixgram", "/tmp/nftest2.sock", "unixgram", "/tmp/nftest.sock")
}
