package main

import (
	"bytes"
	"io"
	"log"
	"net"

	"github.com/hashicorp/yamux"
)

var session *yamux.Session

func main() {
	// Get a TCP connection
	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		panic(err)
	}

	session, err = yamux.Server(conn, nil)
	if err != nil {
		panic(err)
	}

	for {
		stream, err := session.Accept()
		println("Acceping stream")
		if err != nil {
			panic(err)
		}

		buf := make([]byte, 32)
		n, err := stream.Read(buf)
		if err != nil {
			log.Fatalln(err)
			continue
		}
		log.Printf("dialing %s", string(bytes.Trim(buf[:n], "\x00")))
		prox, err := net.Dial("tcp", string(bytes.Trim(buf[:n], "\x00")))
		buf = nil
		if err != nil {
			stream.Write([]byte{0x00})
			errBuff := make([]byte, 256)
			copy(errBuff, []byte(err.Error()))
			stream.Write(errBuff)
			continue
		}
		stream.Write([]byte{0x01})
		// Start proxying
		go proxy(prox, stream)
		go proxy(stream, prox)
	}
}

type closeWriter interface {
	CloseWrite() error
}

// proxy is used to suffle data from src to destination, and sends errors
// down a dedicated channel
func proxy(dst io.Writer, src io.Reader) {
	io.Copy(dst, src)
	if tcpConn, ok := dst.(closeWriter); ok {
		tcpConn.CloseWrite()
	}
}
