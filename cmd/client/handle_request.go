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

func handleRequest(conn net.Conn) error {
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
		return handleConnectCommand(conn, dest)
	default:
		_ = sendReply(conn, nil, commandNotSupported)
		return fmt.Errorf("command not supported")
	}
}

func sendReply(conn net.Conn, addr *internal.AddrSpec, reply uint8) error {
	var addrType uint8
	var addrBody []byte
	var addrPort uint16

	switch {
	case addr == nil:
		addrType = internal.Ipv4Address
		addrBody = []byte{0, 0, 0, 0}
		addrPort = 0

	case addr.Domain != "":
		addrType = internal.DomainAddress
		addrBody = append([]byte{byte(len(addr.Domain))}, []byte(addr.Domain)...)
		addrPort = uint16(addr.Port)

	case addr.IP.To4() != nil:
		addrType = internal.Ipv4Address
		addrBody = addr.IP.To4()
		addrPort = uint16(addr.Port)

	case addr.IP.To16() != nil:
		addrType = internal.Ipv6Address
		addrBody = addr.IP.To16()
		addrPort = uint16(addr.Port)

	default:
		return fmt.Errorf("failed to format address: %v", addr)
	}

	msg := make([]byte, 6+len(addrBody))
	msg[0] = socks5Version
	msg[1] = reply
	msg[2] = 0 // Reserved
	msg[3] = addrType
	copy(msg[4:], addrBody)
	msg[4+len(addrBody)] = byte(addrPort >> 8)
	msg[4+len(addrBody)+1] = byte(addrPort & 0xff)

	_, err := conn.Write(msg)
	return err
}
