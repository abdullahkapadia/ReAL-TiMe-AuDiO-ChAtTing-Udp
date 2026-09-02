package main

import (
	"fmt"
	"log"
	"net"
)

type conn struct {
	conn net.Conn
}

func main() {
	listener, err := net.ResolveUDPAddr("udp", "localhost:8080")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("the port on the udp server is on:- ", listener)

	conn, err := net.ListenUDP("udp", listener)

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)

		if err != nil {
			log.Printf("Read error: %v", err)
			continue
		}

		cc, wrerr := conn.WriteTo(buffer[0:n], clientAddr)

		if wrerr != nil {
			fmt.Printf("net.WriteTo() error: %s\n", wrerr)
		} else {
			fmt.Printf("Wrote %d bytes to socket\n", cc)
		}

		fmt.Printf("Got message from %s: %s\n", clientAddr, string(buffer[:n]))
	}


}
