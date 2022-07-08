package model

import (
	"bytes"
	"fmt"
	"net"
)

type Handshake struct {
	Address string
}

func readError(stream net.Conn) error {
	respBuf := make([]byte, 1)
	_, err := stream.Read(respBuf)
	if err != nil {
		return err
	}

	if respBuf[0] == 0x00 {
		errBuff := make([]byte, 256)
		n, err := stream.Read(errBuff)
		if err != nil {
			return err
		}
		return fmt.Errorf(string(bytes.Trim(errBuff[:n], "\x00")))
	} else if respBuf[0] != 0x01 {
		return fmt.Errorf("corrupted stream")
	}
	return nil
}

func writeConnectionAddress(stream net.Conn, address string) error {
	addrBuf := make([]byte, 32)
	copy(addrBuf, []byte(address))
	_, err := stream.Write(addrBuf)
	return err
}
