package main

import (
	"fmt"
	"io"
	"net"
)

func start(address string) (net.Listener, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := listener.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			return nil, err
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	err := handleSocksProtocol(conn)
	if err != nil {
		fmt.Printf("[ERR] connection failed: %v\n", err)
	}
}

func handleSocksProtocol(conn net.Conn) error {
	version := make([]byte, 1)
	if _, err := conn.Read(version); err != nil {
		return fmt.Errorf("failed to get version byte: %v", err)
	}
	if version[0] != socks5Version {
		return fmt.Errorf("unsupported socks5 version")
	}
	authenticated, err := authenticate(conn)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %v", err)

	}
	if !authenticated {
		conn.Write([]byte{socks5Version, byte(255)})
		return fmt.Errorf("user not authenticated")
	}
	conn.Write([]byte{socks5Version, byte(0)})

	err = handleRequest(conn)
	if err != nil {
		return fmt.Errorf("failed to handle request: %v", err)
	}
	return nil
}

func authenticate(conn net.Conn) (bool, error) {
	numMethods := make([]byte, 1)
	if _, err := conn.Read(numMethods); err != nil {
		return false, fmt.Errorf("failed to get NMETHODS byte: %v", err)
	}
	methods := make([]byte, numMethods[0])
	if _, err := io.ReadFull(conn, methods); err != nil {
		return false, fmt.Errorf("failed to get method bytes: %v", err)
	}
	for _, method := range methods {
		if method == byte(0) {
			return true, nil
		}
	}
	return false, nil
}
