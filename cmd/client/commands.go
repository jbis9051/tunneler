package main

import (
	"fmt"
	"io"
	"net"
	"tunneler/cmd/internal"
)

func handleConnectCommand(conn net.Conn, dest internal.AddrSpec, tunnelerAddr string) error {
	target, err := tunnelDial(&dest, tunnelerAddr)
	if err != nil {
		return err
	}
	defer target.Close()
	status := make([]byte, 1)
	if _, err := io.ReadFull(target, status); err != nil {
		return err
	}
	if status[0] != internal.SuccessResponse {
		if status[0] == internal.AddrTypeNotSupportedResponse {
			_ = sendReply(conn, nil, addrTypeNotSupported)
			return fmt.Errorf("tunnler reported address type not supported error")
		}
		if status[0] == internal.RuleFailureResponse {
			_ = sendReply(conn, nil, ruleFailure)
			return fmt.Errorf("tunnler reported rule failure")
		}
		_ = sendReply(conn, nil, serverFailure)
		return fmt.Errorf("tunnler reported unkown error")
	}
	// local := target.LocalAddr().(*net.TCPAddr)
	// bind := AddrSpec{IP: local.IP, Port: local.Port}
	if err := sendReply(conn, &internal.AddrSpec{IP: []byte{0, 0, 0, 0}}, successReply); err != nil { // the spec says we are supposed to send the BND ip and port but it works without it
		return fmt.Errorf("failed to send reply: %v", err)
	}
	errCh := make(chan error, 2)
	go proxy(target, conn, errCh)
	go proxy(conn, target, errCh)

	for i := 0; i < 2; i++ {
		e := <-errCh
		if e != nil {
			// return from this function closes target (and conn).
			return e
		}
	}
	return nil
}

func proxy(dst io.Writer, src io.Reader, errCh chan error) {
	_, err := io.Copy(dst, src)
	errCh <- err
}
