package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
)

func start(address string) (net.Listener, error) {
	cer, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	listener, err := tls.Listen("tcp", address, &tls.Config{Certificates: []tls.Certificate{cer}})
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

		go func(conn net.Conn) {
			defer conn.Close()
			err := handleConnection(conn)
			if err != nil {
				fmt.Printf("[ERR] connection failed: %v\n", err)
			}
		}(conn)
	}
}
