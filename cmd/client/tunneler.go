package main

import (
	"crypto/tls"
	"encoding/binary"
	"net"
	"tunneler/internal"
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
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, addrSpecParts.AddrPort)
	copy(msg[2+len(addrSpecParts.AddrBody):], portBytes)
	_, err = conn.Write(msg)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
