package netforward

import (
	"io"
	"log"
	"net"
)

func Forward(remote Dialer, local net.Listener) error {
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
		buf := getBuffer()
		defer releaseBuffer(buf)
		_, err := io.CopyBuffer(remoteConn, conn, buf)
		if err != nil {
			log.Printf("forward failed: %v", err)
		}
	}()

	// local -> remote
	buf := getBuffer()
	defer releaseBuffer(buf)
	_, err = io.CopyBuffer(conn, remoteConn, buf)
	if err != nil {
		log.Printf("forward failed: %v", err)
	}
}
