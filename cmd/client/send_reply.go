package main

import (
	"encoding/binary"
	"net"
	"tunneler/internal"
)

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
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, addrSpecParts.AddrPort)
	copy(msg[4+len(addrSpecParts.AddrBody):], portBytes)
	_, err = conn.Write(msg)
	return err
}
