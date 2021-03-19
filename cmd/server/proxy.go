package main

import (
	"fmt"
	"io"
	"net"
	"tunneler/cmd/internal"
)

func handleConnection(conn net.Conn) error {
	version := make([]byte, 1)
	if _, err := io.ReadFull(conn, version); err != nil {
		return fmt.Errorf("failed to read version byte: %v", err)
	}
	if version[0] != internal.TunnelerVersion {
		return fmt.Errorf("invalid version")
	}
	dest, err := internal.ParseDestination(conn)
	if err != nil {
		if err == internal.UnrecognizedAddrTypeError {
			_, _ = conn.Write([]byte{internal.AddrTypeNotSupportedResponse})
		}
		return err
	}
	fmt.Printf("%v\n", dest)
	target, err := net.Dial("tcp", dest.Address())
	if err != nil {
		_, _ = conn.Write([]byte{internal.ConnectionErrorResponse})
		return err
	}
	defer target.Close()
	if _, err := conn.Write([]byte{internal.SuccessResponse}); err != nil { // the spec says we are supposed to send the BND ip and port but it works without it
		return fmt.Errorf("failed to send reply: %v", err)
	}
	errCh := make(chan error, 2)
	go proxy(target, conn, errCh)
	go proxy(conn, target, errCh)

	for i := 0; i < 2; i++ {
		e := <-errCh
		if e != nil {
			return e
		}
	}
	return nil
}

func proxy(dst io.Writer, src io.Reader, errCh chan error) {
	_, err := io.Copy(dst, src)
	errCh <- err
}
