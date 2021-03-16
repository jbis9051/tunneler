package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
)

const (
	ConnectCommand   = uint8(1)
	BindCommand      = uint8(2)
	AssociateCommand = uint8(3)

	ipv4Address   = uint8(1)
	domainAddress = uint8(3)
	ipv6Address   = uint8(4)
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

var (
	unrecognizedAddrType = fmt.Errorf("unrecognized address type")
)

type AddrSpec struct {
	Domain string
	IP     net.IP
	Port   int
}

func (a AddrSpec) Address() string {
	if 0 != len(a.IP) {
		return net.JoinHostPort(a.IP.String(), strconv.Itoa(a.Port))
	}
	return net.JoinHostPort(a.Domain, strconv.Itoa(a.Port))
}

func handleRequest(bufConn *bufio.Reader, conn net.Conn) error {
	header := make([]byte, 3) // VER | CMD |  RSV
	if _, err := io.ReadFull(bufConn, header); err != nil {
		fmt.Printf("[ERR] socks: Failed to get NMETHODS byte: %v", err)
		return err
	}
	if header[0] != socks5Version { // VER
		return errors.New("[ERR] socks: Unsupported socks5 version")
	}
	dest, err := parseDestination(bufConn)
	fmt.Println(dest)
	if err != nil {
		if err == unrecognizedAddrType {
			_ = sendReply(conn, nil, addrTypeNotSupported)
			return err
		}
	}
	switch header[1] { // CMD
	case ConnectCommand:
		return handleConnectCommand(bufConn, conn, dest)
	default:
		_ = sendReply(conn, nil, commandNotSupported)
		return fmt.Errorf("command not supported")
	}
}

func sendReply(conn net.Conn, addr *AddrSpec, reply uint8) error {
	var addrType uint8
	var addrBody []byte
	var addrPort uint16

	switch {
	case addr == nil:
		addrType = ipv4Address
		addrBody = []byte{0, 0, 0, 0}
		addrPort = 0

	case addr.Domain != "":
		addrType = domainAddress
		addrBody = append([]byte{byte(len(addr.Domain))}, []byte(addr.Domain)...)
		addrPort = uint16(addr.Port)

	case addr.IP.To4() != nil:
		addrType = ipv4Address
		addrBody = addr.IP.To4()
		addrPort = uint16(addr.Port)

	case addr.IP.To16() != nil:
		addrType = ipv6Address
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

func parseDestination(bufConn *bufio.Reader) (AddrSpec, error) {
	addrType := make([]byte, 1) // atyp
	if _, err := io.ReadFull(bufConn, addrType); err != nil {
		return AddrSpec{}, err
	}
	switch addrType[0] {
	case ipv4Address:
		address := make([]byte, net.IPv4len)
		if _, err := io.ReadFull(bufConn, address); err != nil {
			return AddrSpec{}, err
		}
		port, err := parsePort(bufConn)
		if err != nil {
			return AddrSpec{IP: address}, unrecognizedAddrType
		}
		return AddrSpec{IP: address, Port: port}, nil

	case ipv6Address:
		address := make([]byte, net.IPv6len)
		if _, err := io.ReadFull(bufConn, address); err != nil {
			return AddrSpec{}, err
		}
		port, err := parsePort(bufConn)
		if err != nil {
			return AddrSpec{IP: address}, unrecognizedAddrType
		}
		return AddrSpec{IP: address, Port: port}, nil

	case domainAddress:
		domainLength := make([]byte, 1)
		if _, err := io.ReadFull(bufConn, domainLength); err != nil {
			return AddrSpec{}, err
		}
		domain := make([]byte, int(domainLength[0]))
		if _, err := io.ReadFull(bufConn, domain); err != nil {
			return AddrSpec{}, err
		}
		ip, err := net.ResolveIPAddr("ip", string(domain))
		if err != nil {
			return AddrSpec{Domain: string(domain)}, unrecognizedAddrType
		}
		port, err := parsePort(bufConn)
		if err != nil {
			return AddrSpec{Domain: string(domain), IP: ip.IP}, unrecognizedAddrType
		}
		return AddrSpec{Domain: string(domain), IP: ip.IP, Port: port}, nil
	default:
		return AddrSpec{}, unrecognizedAddrType
	}
}

func parsePort(bufConn *bufio.Reader) (int, error) {
	port := make([]byte, 2)
	if _, err := io.ReadFull(bufConn, port); err != nil {
		return 0, err
	}
	return (int(port[0]) << 8) | int(port[1]), nil // this does some math shit to convert binary to decimal...somehow
}
