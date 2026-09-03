package main

import (
	"fmt"
	"log"
	"net"
	"sync"
)

func main() {
	fmt.Println("Starting UDP SFU Voice Server (Router)...")
	listener, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", listener)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Println("Listening on", listener)

	// Map of active clients in the global "room"
	clients := make(map[string]*net.UDPAddr)
	var mu sync.Mutex

	// Network loop
	buffer := make([]byte, 4096)
	
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Read error: %v", err)
			continue
		}
		
		addrStr := addr.String()

		mu.Lock()
		// If this is a new client, add them to the room
		if _, exists := clients[addrStr]; !exists {
			fmt.Printf("New client joined the voice room: %s\n", addrStr)
			clients[addrStr] = addr
		}

		// Forward the audio packet to everyone ELSE in the room
		for clientAddrStr, clientAddr := range clients {
			if clientAddrStr != addrStr {
				// We don't want to send the user's own voice back to them
				conn.WriteToUDP(buffer[:n], clientAddr)
			}
		}
		mu.Unlock()
	}
}
