package main

import (
	"encoding/binary"
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

	// Map of active clients organized by Room ID
	// RoomID -> Map of IP Address -> net.UDPAddr
	rooms := make(map[uint32]map[string]*net.UDPAddr)
	var mu sync.Mutex

	// Network loop
	buffer := make([]byte, 4096)
	
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Read error: %v", err)
			continue
		}
		
		// Ensure packet is large enough to contain RoomID (4) + Sequence (4) = 8 bytes
		if n < 8 {
			continue
		}

		// Read the first 4 bytes to determine which room this packet belongs to
		roomID := binary.LittleEndian.Uint32(buffer[0:4])
		addrStr := addr.String()

		mu.Lock()
		// If this is a new room, initialize it
		if rooms[roomID] == nil {
			rooms[roomID] = make(map[string]*net.UDPAddr)
		}

		// If this is a new client in the room, add them
		if _, exists := rooms[roomID][addrStr]; !exists {
			fmt.Printf("New client %s joined Voice Room %d\n", addrStr, roomID)
			rooms[roomID][addrStr] = addr
		}

		// Forward the audio packet to everyone ELSE in this specific room
		for clientAddrStr, clientAddr := range rooms[roomID] {
			if clientAddrStr != addrStr {
				conn.WriteToUDP(buffer[:n], clientAddr)
			}
		}
		mu.Unlock()
	}
}
