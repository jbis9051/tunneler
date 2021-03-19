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

type AddrSpecParts struct {
	AddrType uint8
	AddrBody []byte
	AddrPort uint16
}

func (addr AddrSpec) Address() string {
	if len(addr.IP) != 0 {
		return net.JoinHostPort(addr.IP.String(), strconv.Itoa(addr.Port))
	}
	return net.JoinHostPort(addr.Domain, strconv.Itoa(addr.Port))
}

func (addr *AddrSpec) ToParts() (AddrSpecParts, error) {
	switch {
	case addr == nil:
		return AddrSpecParts{AddrType: Ipv4Address, AddrBody: []byte{0, 0, 0, 0}, AddrPort: 0}, nil
	case addr.Domain != "":
		return AddrSpecParts{AddrType: DomainAddress, AddrBody: append([]byte{byte(len(addr.Domain))}, []byte(addr.Domain)...), AddrPort: uint16(addr.Port)}, nil

	case addr.IP.To4() != nil:
		return AddrSpecParts{AddrType: Ipv4Address, AddrBody: addr.IP.To4(), AddrPort: uint16(addr.Port)}, nil

	case addr.IP.To16() != nil:
		return AddrSpecParts{AddrType: Ipv6Address, AddrBody: addr.IP.To16(), AddrPort: uint16(addr.Port)}, nil

	default:
		return AddrSpecParts{}, fmt.Errorf("failed to format address: %v", addr)
	}
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
