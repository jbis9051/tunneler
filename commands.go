package main

import (
	"fmt"
	"io"
	"net"
	"strings"
)

func handleConnectCommand(conn net.Conn, dest AddrSpec) error {
	target, err := net.Dial("tcp", dest.Address())
	if err != nil {
		msg := err.Error()
		resp := hostUnreachable
		if strings.Contains(msg, "refused") {
			resp = connectionRefused
		} else if strings.Contains(msg, "network is unreachable") {
			resp = networkUnreachable
		}
		_ = sendReply(conn, nil, resp)
		return err
	}
	defer target.Close()
	local := target.LocalAddr().(*net.TCPAddr)
	bind := AddrSpec{IP: local.IP, Port: local.Port}
	if err := sendReply(conn, &bind, successReply); err != nil {
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

// proxy is used to shuffle data from src to destination, and sends errors
// down a dedicated channel
func proxy(dst io.Writer, src io.Reader, errCh chan error) {
	_, err := io.Copy(dst, src)
	errCh <- err
}
