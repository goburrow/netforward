package netforward

import (
	"net"
	"sync"
	"testing"
)

func TestStream(t *testing.T) {
	var wg sync.WaitGroup

	// Echo server
	sln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer sln.Close()

	// Proxy server
	e1 := Endpoint{
		Network: "unix",
		Address: "/tmp/nftest.sock",
	}
	pln, err := e1.Listen()
	if err != nil {
		t.Fatal(err)
	}
	defer pln.Close()

	// Start the echo server
	wg.Add(1)
	go func() {
		defer wg.Done()
		c, err := sln.Accept()
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		// echo received data
		var buf [16]byte
		n, err := c.Read(buf[:])
		if err != nil {
			t.Fatal(err)
		}
		n, err = c.Write(buf[:n])
		if err != nil {
			t.Fatal(err)
		}
	}()

	// Forwarding
	wg.Add(1)
	go func() {
		defer wg.Done()
		e2 := Endpoint{
			Network: "tcp",
			Address: sln.Addr().String(),
		}
		err = Forward(&e2, pln)
		t.Log(err)
	}()

	c, err := net.Dial("unix", "/tmp/nftest.sock")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	n, err := c.Write([]byte("netforward"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 10 {
		t.Fatalf("unexpected bytes written: %d", err, n)
	}
	var buf [16]byte
	n, err = c.Read(buf[:])
	if err != nil {
		t.Fatal(err)
	}
	if n != 10 || "netforward" != string(buf[:n]) {
		t.Fatalf("unexpected receive: %s (%d bytes)", buf[:n], n)
	}

	sln.Close()
	pln.Close()

	wg.Wait()
}
