package main

import (
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
	msg[4+len(addrSpecParts.AddrBody)] = byte(addrSpecParts.AddrPort >> 8)
	msg[4+len(addrSpecParts.AddrBody)+1] = byte(addrSpecParts.AddrPort & 0xff)

	_, err = conn.Write(msg)
	return err
}
