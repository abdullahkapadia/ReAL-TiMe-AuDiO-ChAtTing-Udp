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

		fmt.Printf("Got message from %s: %s\n", clientAddr, string(buffer[:n]))
	}

}
