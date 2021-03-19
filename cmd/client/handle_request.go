package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"tunneler/cmd/internal"
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

func sendReply(conn net.Conn, addr *internal.AddrSpec, reply uint8) error {
	addrSpecParts, err := addr.ToParts()

	if err != nil {
		return err
	}

	msg := make([]byte, 6+len(addrSpecParts.AddrBody))
	msg[0] = socks5Version
	msg[1] = reply
	msg[2] = 0 // Reserved
	msg[3] = addrSpecParts.AddrType
	copy(msg[4:], addrSpecParts.AddrBody)
	msg[4+len(addrSpecParts.AddrBody)] = byte(addrSpecParts.AddrPort >> 8)
	msg[4+len(addrSpecParts.AddrBody)+1] = byte(addrSpecParts.AddrPort & 0xff)

	_, err = conn.Write(msg)
	return err
}
