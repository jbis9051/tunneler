package main

import (
	"crypto/tls"
	"net"
	"tunneler/cmd/internal"
)

func tunnelDial(dest *internal.AddrSpec, tunnelerAddr string) (net.Conn, error) {
	conn, err := tls.Dial("tcp", tunnelerAddr, &tls.Config{InsecureSkipVerify: true, ServerName: "test123.com"})
	if err != nil {
		return nil, err
	}

	addrSpecParts, err := dest.ToParts()

	if err != nil {
		return nil, err
	}

	msg := make([]byte, 4+len(addrSpecParts.AddrBody))
	msg[0] = internal.TunnelerVersion
	msg[1] = addrSpecParts.AddrType
	copy(msg[2:], addrSpecParts.AddrBody)
	msg[2+len(addrSpecParts.AddrBody)] = byte(addrSpecParts.AddrPort >> 8)
	msg[2+len(addrSpecParts.AddrBody)+1] = byte(addrSpecParts.AddrPort & 0xff)

	_, err = conn.Write(msg)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
