package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"tunneler/internal"
)

const (
	ConnectCommand   = uint8(1)
	BindCommand      = uint8(2)
	AssociateCommand = uint8(3)
)

const (
	successReply uint8 = iota
	serverFailure
	ruleFailure
	networkUnreachable
	hostUnreachable
	connectionRefused
	ttlExpired
	commandNotSupported
	addrTypeNotSupported
)

func handleRequest(conn net.Conn, tunnelerAddr string) error {
	header := make([]byte, 3) // VER | CMD |  RSV
	if _, err := io.ReadFull(conn, header); err != nil {
		fmt.Printf("[ERR] socks: Failed to get NMETHODS byte: %v", err)
		return err
	}
	if header[0] != socks5Version { // VER
		return errors.New("[ERR] socks: Unsupported socks5 version")
	}
	dest, err := internal.ParseDestination(conn)
	fmt.Println(dest)
	if err != nil {
		if err == internal.UnrecognizedAddrTypeError {
			_ = sendReply(conn, nil, addrTypeNotSupported)
		}
		return err
	}
	switch header[1] { // CMD
	case ConnectCommand:
		return handleConnectCommand(conn, dest, tunnelerAddr)
	default:
		_ = sendReply(conn, nil, commandNotSupported)
		return fmt.Errorf("command not supported")
	}
}
