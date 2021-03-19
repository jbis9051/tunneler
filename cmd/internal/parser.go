package internal

import (
	"fmt"
	"io"
	"net"
	"strconv"
)

const (
	Ipv4Address   = uint8(1)
	DomainAddress = uint8(3)
	Ipv6Address   = uint8(4)
)

var (
	UnrecognizedAddrTypeError = fmt.Errorf("unrecognized address type")
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

func ParseDestination(conn net.Conn) (AddrSpec, error) {
	addrType := make([]byte, 1) // atyp
	if _, err := io.ReadFull(conn, addrType); err != nil {
		return AddrSpec{}, err
	}
	switch addrType[0] {
	case Ipv4Address:
		address := make([]byte, net.IPv4len)
		if _, err := io.ReadFull(conn, address); err != nil {
			return AddrSpec{}, err
		}
		port, err := parsePort(conn)
		if err != nil {
			return AddrSpec{IP: address}, UnrecognizedAddrTypeError
		}
		return AddrSpec{IP: address, Port: port}, nil

	case Ipv6Address:
		address := make([]byte, net.IPv6len)
		if _, err := io.ReadFull(conn, address); err != nil {
			return AddrSpec{}, err
		}
		port, err := parsePort(conn)
		if err != nil {
			return AddrSpec{IP: address}, UnrecognizedAddrTypeError
		}
		return AddrSpec{IP: address, Port: port}, nil

	case DomainAddress:
		domainLength := make([]byte, 1)
		if _, err := io.ReadFull(conn, domainLength); err != nil {
			return AddrSpec{}, err
		}
		domain := make([]byte, int(domainLength[0]))
		if _, err := io.ReadFull(conn, domain); err != nil {
			return AddrSpec{}, err
		}
		ip, err := net.ResolveIPAddr("ip", string(domain))
		if err != nil {
			return AddrSpec{Domain: string(domain)}, UnrecognizedAddrTypeError
		}
		port, err := parsePort(conn)
		if err != nil {
			return AddrSpec{Domain: string(domain), IP: ip.IP}, UnrecognizedAddrTypeError
		}
		return AddrSpec{Domain: string(domain), IP: ip.IP, Port: port}, nil
	default:
		return AddrSpec{}, UnrecognizedAddrTypeError
	}
}

func parsePort(conn net.Conn) (int, error) {
	port := make([]byte, 2)
	if _, err := io.ReadFull(conn, port); err != nil {
		return 0, err
	}
	return (int(port[0]) << 8) | int(port[1]), nil // this does some math shit to convert binary to decimal...somehow
}
