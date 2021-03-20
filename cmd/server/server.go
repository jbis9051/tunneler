package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
)

func start(address string) (net.Listener, error) {
	certPath, certBool := os.LookupEnv("CERTPATH")
	keypath, keyBool := os.LookupEnv("KEYPATH")
	if !certBool || !keyBool {
		return nil, fmt.Errorf("CERTPATH or KEYPATH unset")
	}
	cer, err := tls.LoadX509KeyPair(certPath, keypath)
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
